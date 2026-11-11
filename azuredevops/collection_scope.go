package azuredevops

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

const collectionWildcard = "*"

// serverScopedTables are registered on the Azure DevOps Server deployment,
// not on a project collection. Queries against them run once at the server root.
var serverScopedTables = map[string]struct{}{
	"azuredevopssh_project_collection": {},
	"azuredevopssh_agent_pool":         {},
	"azuredevopssh_agent":              {},
}

func isServerScopedTable(name string) bool {
	_, ok := serverScopedTables[name]
	return ok
}

// collectionPlan decides which collections a connection queries.
// Multi is false for a single collection URL when neither list is set.
type collectionPlan struct {
	Multi     bool
	ServerURL string
	Allow     []string
	Deny      []string
}

func resolveCollectionPlan(config azureDevOpsConfig) collectionPlan {
	allow := normalizeCollectionNames(config.Collections)
	deny := normalizeCollectionNames(config.ExcludedCollections)
	if collectionListIsWildcard(allow) {
		allow = nil
		return collectionPlan{
			Multi:     true,
			ServerURL: getServerURL(config),
			Deny:      deny,
		}
	}

	single := len(allow) == 0 && len(deny) == 0 && urlHasCollectionPath(configuredOrganizationURL(config))
	return collectionPlan{
		Multi:     !single,
		ServerURL: getServerURL(config),
		Allow:     allow,
		Deny:      deny,
	}
}

func collectionListIsWildcard(names []string) bool {
	for _, name := range names {
		if name == collectionWildcard {
			return true
		}
	}
	return false
}

func normalizeCollectionNames(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	var normalized []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		normalized = append(normalized, name)
	}
	return normalized
}

func urlHasCollectionPath(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return strings.Trim(raw, "/") != "" && strings.Contains(strings.Trim(raw, "/"), "/")
	}
	return strings.Trim(parsed.Path, "/") != ""
}

type projectCollection struct {
	Name string
	URL  string
}

func collectionBaseURL(server, name string) string {
	return strings.TrimRight(server, "/") + "/" + strings.Trim(name, "/")
}

// selectCollections applies an allow-list and then a deny-list.
// A name present in both lists is excluded. A name in either list that was
// not discovered is an error. An empty result is an error.
func selectCollections(server string, discovered []projectCollection, allow, deny []string) ([]projectCollection, error) {
	byName := make(map[string]projectCollection, len(discovered))
	names := make([]string, 0, len(discovered))
	ordered := make([]projectCollection, 0, len(discovered))
	for _, collection := range discovered {
		key := strings.ToLower(collection.Name)
		if _, exists := byName[key]; exists {
			continue
		}
		byName[key] = collection
		names = append(names, collection.Name)
		ordered = append(ordered, collection)
	}
	discoveredList := strings.Join(names, ", ")
	if discoveredList == "" {
		discoveredList = "(none)"
	}

	selected := ordered
	if len(allow) > 0 {
		selected = make([]projectCollection, 0, len(allow))
		seen := make(map[string]struct{}, len(allow))
		for _, name := range allow {
			key := strings.ToLower(name)
			if _, duplicate := seen[key]; duplicate {
				continue
			}
			seen[key] = struct{}{}
			collection, ok := byName[key]
			if !ok {
				return nil, fmt.Errorf("collection %q was not found on %s; discovered collections: %s", name, server, discoveredList)
			}
			selected = append(selected, collection)
		}
	} else if len(ordered) == 0 {
		return nil, fmt.Errorf("no project collections were discovered on %s", server)
	}

	if len(deny) > 0 {
		denied := make(map[string]struct{}, len(deny))
		for _, name := range deny {
			key := strings.ToLower(name)
			if _, ok := byName[key]; !ok {
				return nil, fmt.Errorf("excluded collection %q was not found on %s; discovered collections: %s", name, server, discoveredList)
			}
			denied[key] = struct{}{}
		}
		filtered := make([]projectCollection, 0, len(selected))
		for _, collection := range selected {
			if _, skip := denied[strings.ToLower(collection.Name)]; skip {
				continue
			}
			filtered = append(filtered, collection)
		}
		selected = filtered
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("no project collections left to query on %s", server)
	}
	return selected, nil
}

// absentFromCollection reports a Get miss in one collection while the query
// is scanning more than one. Single-collection calls keep the API error.
func absentFromCollection(d *plugin.QueryData, err error) bool {
	if err == nil || qualString(d, "collection_url") == "" {
		return false
	}
	return isNotFound(err)
}

func isNotFound(err error) bool {
	var wrapped *azuredevops.WrappedError
	if errors.As(err, &wrapped) && notFoundStatus(wrapped) {
		return true
	}
	var value azuredevops.WrappedError
	if errors.As(err, &value) && notFoundStatus(&value) {
		return true
	}
	return false
}

func notFoundStatus(err *azuredevops.WrappedError) bool {
	return err != nil && err.StatusCode != nil && *err.StatusCode == http.StatusNotFound
}
