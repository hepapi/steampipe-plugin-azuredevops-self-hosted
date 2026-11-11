package azuredevops

import (
	"context"
	"net/url"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/memoize"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// qualString returns the equals-qual value for column, or "" when the qual
// was not supplied. EqualsQuals only contains columns present in the query,
// so a missing entry is nil.
func qualString(d *plugin.QueryData, column string) string {
	if d == nil || d.EqualsQuals[column] == nil {
		return ""
	}
	return d.EqualsQuals[column].GetStringValue()
}

// qualInt reports whether an equals qual was supplied for an integer column.
func qualInt(d *plugin.QueryData, column string) (int64, bool) {
	if d == nil || d.EqualsQuals[column] == nil {
		return 0, false
	}
	return d.EqualsQuals[column].GetInt64Value(), true
}

// qualBool reports whether an equals qual was supplied for a boolean column.
func qualBool(d *plugin.QueryData, column string) (bool, bool) {
	if d == nil || d.EqualsQuals[column] == nil {
		return false, false
	}
	return d.EqualsQuals[column].GetBoolValue(), true
}

func commonColumns(c []*plugin.Column) []*plugin.Column {
	return append([]*plugin.Column{
		{
			Name:        "organization",
			Description: "The name of the organization.",
			Type:        proto.ColumnType_STRING,
			Hydrate:     getOrganization,
			Transform:   transform.FromValue(),
		},
	}, c...)
}

// if the caching is required other than per connection, build a cache key for the call and use it in Memoize.
var getOrganizationMemoized = plugin.HydrateFunc(getOrganizationUncached).Memoize(memoize.WithCacheKeyFunction(getOrganizationCacheKey))

// declare a wrapper hydrate function to call the memoized function
// - this is required when a memoized function is used for a column definition
func getOrganization(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	return getOrganizationMemoized(ctx, d, h)
}

// Build a cache key for the call to getOrganizationCacheKey.
func getOrganizationCacheKey(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	if name := matrixOrganization(d); name != "" {
		return "getOrganization|" + name, nil
	}
	return "getOrganization", nil
}

func getOrganizationUncached(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	if name := matrixOrganization(d); name != "" {
		return name, nil
	}
	var config azureDevOpsConfig
	if d != nil {
		config = GetConfig(d.Connection)
	}
	return organizationName(configuredOrganizationURL(config)), nil
}

// matrixOrganization is the collection name for the current matrix item.
// Server-scoped tables have no matrix item and keep the host-level name.
func matrixOrganization(d *plugin.QueryData) string {
	if qualString(d, "collection_url") == "" {
		return ""
	}
	return qualString(d, "organization")
}

// organizationName derives the organization (SaaS) or collection (self-hosted)
// name from the organization_url. For self-hosted URLs such as
// https://azuredevops.example.com/DefaultCollection it returns the last path segment
// (DefaultCollection). For SaaS URLs such as https://dev.azure.com/org it
// returns the last path segment (org). If there is no path segment, it falls
// back to the host.
func organizationName(organizationURL string) string {
	organizationURL = strings.TrimRight(organizationURL, "/")

	u, err := url.Parse(organizationURL)
	if err != nil {
		return organizationURL
	}

	// Take the last non-empty path segment (collection or org name).
	path := strings.Trim(u.Path, "/")
	if path != "" {
		segments := strings.Split(path, "/")
		if last := segments[len(segments)-1]; last != "" {
			return last
		}
	}

	// Fall back to the host when there is no path segment.
	return u.Host
}
