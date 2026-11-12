package azuredevops

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

const (
	projectCollectionCacheKey = "azuredevopssh_project_collections"
	projectCollectionPageSize = 1000
	projectCollectionMaxPages = 50
)

type cachedProjectCollections struct {
	Items []projectCollection
	Err   string
}

// collectionMatrix returns one item per collection for collection-scoped tables.
// A nil result keeps the historical single-URL call. Discovery failures are
// returned as a matrix item because this callback cannot return an error.
func collectionMatrix(ctx context.Context, d *plugin.QueryData) []map[string]interface{} {
	var config azureDevOpsConfig
	if d != nil {
		config = GetConfig(d.Connection)
	}
	plan := resolveCollectionPlan(config)
	if !plan.Multi {
		return nil
	}

	discovered, err := discoverProjectCollections(ctx, d, plan.ServerURL)
	if err != nil {
		return collectionMatrixError(err)
	}
	selected, err := selectCollections(plan.ServerURL, discovered, plan.Allow, plan.Deny)
	if err != nil {
		return collectionMatrixError(err)
	}

	// An organization qual limits the list to that collection. The SDK then
	// keeps a matrix item only when its organization value equals the qual,
	// so the returned item uses the query text.
	org := strings.TrimSpace(qualString(d, "organization"))
	if org != "" {
		matched := filterCollectionsByOrganization(selected, org)
		if len(matched) == 0 {
			names := make([]string, 0, len(selected))
			for _, collection := range selected {
				names = append(names, collection.Name)
			}
			discoveredList := strings.Join(names, ", ")
			if discoveredList == "" {
				discoveredList = "(none)"
			}
			return collectionMatrixError(fmt.Errorf("collection %q was not found on %s; discovered collections: %s", org, plan.ServerURL, discoveredList))
		}
		selected = matched
	}

	items := make([]map[string]interface{}, 0, len(selected))
	for _, collection := range selected {
		name := collection.Name
		if org != "" {
			name = org
		}
		items = append(items, map[string]interface{}{
			"organization":   name,
			"collection_url": collection.URL,
		})
	}
	return items
}

// filterCollectionsByOrganization keeps collections whose name matches name,
// ignoring case and surrounding space. An empty name returns the input unchanged.
func filterCollectionsByOrganization(collections []projectCollection, name string) []projectCollection {
	name = strings.TrimSpace(name)
	if name == "" {
		return collections
	}
	matched := make([]projectCollection, 0, 1)
	for _, collection := range collections {
		if strings.EqualFold(collection.Name, name) {
			matched = append(matched, collection)
		}
	}
	return matched
}

func collectionMatrixError(err error) []map[string]interface{} {
	return []map[string]interface{}{
		{"collection_error": err.Error()},
	}
}

func discoverProjectCollections(ctx context.Context, d *plugin.QueryData, server string) ([]projectCollection, error) {
	if d != nil && d.ConnectionCache != nil {
		if cached, ok := d.ConnectionCache.Get(ctx, projectCollectionCacheKey); ok {
			if stored, ok := cached.(cachedProjectCollections); ok {
				if stored.Err != "" {
					return nil, errors.New(stored.Err)
				}
				return stored.Items, nil
			}
		}
	}

	items, err := fetchProjectCollections(ctx, d, server)
	if d != nil && d.ConnectionCache != nil {
		stored := cachedProjectCollections{Items: items}
		if err != nil {
			stored = cachedProjectCollections{Err: err.Error()}
		}
		_ = d.ConnectionCache.SetWithTTL(ctx, projectCollectionCacheKey, stored, 10*time.Minute)
	}
	return items, err
}

func fetchProjectCollections(ctx context.Context, d *plugin.QueryData, server string) ([]projectCollection, error) {
	connection, err := getServerConnection(ctx, d)
	if err != nil {
		return nil, err
	}
	client, err := core.NewClient(ctx, connection)
	if err != nil {
		return nil, err
	}

	var found []projectCollection
	skip := 0
	for page := 0; page < projectCollectionMaxPages; page++ {
		top := projectCollectionPageSize
		collections, err := client.GetProjectCollections(ctx, core.GetProjectCollectionsArgs{
			Top:  &top,
			Skip: &skip,
		})
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh.discoverProjectCollections", "api_error", err)
			return nil, err
		}
		if collections == nil || len(*collections) == 0 {
			break
		}
		for _, collection := range *collections {
			if collection.Name == nil {
				continue
			}
			name := strings.TrimSpace(*collection.Name)
			if name == "" {
				continue
			}
			found = append(found, projectCollection{
				Name: name,
				URL:  collectionBaseURL(server, name),
			})
		}
		if len(*collections) < projectCollectionPageSize {
			break
		}
		skip += len(*collections)
	}
	return found, nil
}

func attachCollectionMatrix(p *plugin.Plugin) {
	for name, table := range p.TableMap {
		if table == nil || isServerScopedTable(name) {
			continue
		}
		table.GetMatrixItemFunc = collectionMatrix
		ensureOrganizationKeyColumn(table)
	}
}

// ensureOrganizationKeyColumn makes organization an optional list qual so
// Steampipe pushes `where organization = ...` before the matrix is built.
func ensureOrganizationKeyColumn(table *plugin.Table) {
	if table.List == nil {
		return
	}
	for _, column := range table.List.KeyColumns {
		if column != nil && column.Name == "organization" {
			return
		}
	}
	table.List.KeyColumns = append(table.List.KeyColumns, &plugin.KeyColumn{
		Name:    "organization",
		Require: plugin.Optional,
	})
}
