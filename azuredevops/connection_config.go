package azuredevops

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

type azureDevOpsConfig struct {
	OrganizationURL     *string `hcl:"organization_url"`
	ServerURL           *string `hcl:"server_url"`
	PersonalAccessToken *string `hcl:"personal_access_token"`
	// Collections limits a server-root connection to these collection names.
	// "*" discovers every collection. Ignored when the connection is a single collection URL and ExcludedCollections is also empty.
	Collections []string `hcl:"collections,optional"`
	// ExcludedCollections drops these collection names after discovery or after Collections is applied.
	ExcludedCollections []string `hcl:"excluded_collections,optional"`
}

func ConfigInstance() interface{} {
	return &azureDevOpsConfig{}
}

func GetConfig(connection *plugin.Connection) azureDevOpsConfig {
	if connection == nil || connection.Config == nil {
		return azureDevOpsConfig{}
	}

	config, _ := connection.Config.(azureDevOpsConfig)

	return config
}

func getConnection(ctx context.Context, d *plugin.QueryData) (*azuredevops.Connection, error) {
	config := GetConfig(connectionConfig(d))
	if message := qualString(d, "collection_error"); message != "" {
		return nil, errors.New(message)
	}

	serviceURL := configuredOrganizationURL(config)
	token := personalAccessToken(config)
	if collectionURL := qualString(d, "collection_url"); collectionURL != "" {
		serviceURL = collectionURL
	} else if d != nil && d.Table != nil && isServerScopedTable(d.Table.Name) && resolveCollectionPlan(config).Multi {
		serviceURL = getServerURL(config)
	}

	if serviceURL != "" && token != "" {
		connection := azuredevops.NewPatConnection(serviceURL, token)
		return connection, nil
	}

	return nil, errors.New("'organization_url' and 'personal_access_token' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
}

func connectionConfig(d *plugin.QueryData) *plugin.Connection {
	if d == nil {
		return nil
	}
	return d.Connection
}

func configuredOrganizationURL(config azureDevOpsConfig) string {
	value := os.Getenv("AZDO_ORG_SERVICE_URL")
	if config.OrganizationURL != nil {
		value = *config.OrganizationURL
	}
	return value
}

func personalAccessToken(config azureDevOpsConfig) string {
	value := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	if config.PersonalAccessToken != nil {
		value = *config.PersonalAccessToken
	}
	return value
}

// getServerURL returns the Azure DevOps server root URL used to enumerate
// project collections. For self-hosted deployments this is the server root
// (e.g. https://azuredevops.example.com/), while organization_url points at a
// specific collection (e.g. https://azuredevops.example.com/DefaultCollection).
//
// Resolution order:
//  1. The explicit `server_url` config value.
//  2. Derived from `organization_url` by stripping the last path segment.
//  3. The `organization_url` itself (SaaS, where there is no separate server).
func getServerURL(config azureDevOpsConfig) string {
	if config.ServerURL != nil && *config.ServerURL != "" {
		return strings.TrimRight(*config.ServerURL, "/")
	}

	serviceURL := configuredOrganizationURL(config)

	u, err := url.Parse(serviceURL)
	if err != nil {
		return serviceURL
	}

	// Strip the last path segment (the collection/org name) to get the server root.
	path := strings.Trim(u.Path, "/")
	if path != "" {
		segments := strings.Split(path, "/")
		u.Path = "/" + strings.Join(segments[:len(segments)-1], "/")
	}

	return strings.TrimRight(u.String(), "/")
}

// getServerConnection builds a connection to the server root (used only for
// enumerating project collections), as opposed to getConnection which targets
// the organization/collection URL.
func getServerConnection(ctx context.Context, d *plugin.QueryData) (*azuredevops.Connection, error) {
	config := GetConfig(connectionConfig(d))

	serverURL := getServerURL(config)
	token := personalAccessToken(config)

	if serverURL != "" && token != "" {
		connection := azuredevops.NewPatConnection(serverURL, token)
		return connection, nil
	}

	return nil, errors.New("'server_url' (or 'organization_url') and 'personal_access_token' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
}
