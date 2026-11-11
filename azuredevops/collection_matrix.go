package azuredevops

import (
	"context"
	"errors"
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

	items := make([]map[string]interface{}, 0, len(selected))
	for _, collection := range selected {
		items = append(items, map[string]interface{}{
			"organization":   collection.Name,
			"collection_url": collection.URL,
		})
	}
	return items
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
	}
}
