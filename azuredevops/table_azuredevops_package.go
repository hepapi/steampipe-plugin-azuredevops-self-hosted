package azuredevops

import (
	"context"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

const packagePageSize = 100

// packageRow carries the feed scope. name_query is the search text, which is
// not the same as the package name, so it has to travel with the row.
type packageRow struct {
	feed.Package
	FeedId    string
	ProjectId string
	NameQuery string
}

func tableAzureDevOpsPackage(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_package",
		Description: "Retrieve packages in an Azure Artifacts feed.",
		List: &plugin.ListConfig{
			Hydrate: listPackages,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "feed_id", Require: plugin.Required},
				{Name: "project_id", Require: plugin.Optional},
				{Name: "id", Require: plugin.Optional},
				{Name: "protocol_type", Require: plugin.Optional},
				{Name: "name_query", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "id", Require: plugin.Required},
				{Name: "feed_id", Require: plugin.Required},
				{Name: "project_id", Require: plugin.Optional},
			},
			Hydrate: getPackage,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "Package identifier.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "feed_id",
				Description: "Feed that contains the package.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "Project that owns the feed. Empty for a collection feed.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "Package name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "normalized_name",
				Description: "Normalized package name for its protocol.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "protocol_type",
				Description: "Package protocol, such as NuGet, npm, Maven, PyPI, or Universal.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name_query",
				Description: "Substring matched against package names. This is the search text, not the package name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "URL of the package.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_cached",
				Description: "True when the package was cached from an upstream.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "versions",
				Description: "Versions included with the package. The list returns the latest version unless the get includes every version.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
		}),
	}
}

func listPackages(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	feedID := qualString(d, "feed_id")
	if feedID == "" {
		return nil, fmt.Errorf("azuredevops_package requires feed_id")
	}
	projectID := qualString(d, "project_id")
	protocolType := qualString(d, "protocol_type")
	nameQuery := qualString(d, "name_query")
	id := qualString(d, "id")

	client, err := newFeedClient(ctx, d, "azuredevops_package.listPackages")
	if err != nil {
		return nil, err
	}

	skip := 0
	for {
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
		top := packagePageSize
		args := feed.GetPackagesArgs{
			FeedId:      &feedID,
			Project:     optionalProject(projectID),
			Top:         &top,
			Skip:        &skip,
			IncludeUrls: boolPtr(true),
		}
		if protocolType != "" {
			args.ProtocolType = &protocolType
		}
		if nameQuery != "" {
			args.PackageNameQuery = &nameQuery
		}

		packages, err := client.GetPackages(ctx, args)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevops_package.listPackages", "api_error", err)
			return nil, err
		}
		if packages == nil || len(*packages) == 0 {
			return nil, nil
		}

		for _, item := range *packages {
			if id != "" && (item.Id == nil || item.Id.String() != id) {
				continue
			}
			d.StreamListItem(ctx, packageRow{
				Package:   item,
				FeedId:    feedID,
				ProjectId: projectID,
				NameQuery: nameQuery,
			})
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		if len(*packages) < packagePageSize {
			return nil, nil
		}
		skip += len(*packages)
	}
}

func getPackage(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	id := qualString(d, "id")
	feedID := qualString(d, "feed_id")
	if id == "" || feedID == "" {
		return nil, nil
	}
	projectID := qualString(d, "project_id")

	client, err := newFeedClient(ctx, d, "azuredevops_package.getPackage")
	if err != nil {
		return nil, err
	}

	item, err := client.GetPackage(ctx, feed.GetPackageArgs{
		FeedId:             &feedID,
		PackageId:          &id,
		Project:            optionalProject(projectID),
		IncludeUrls:        boolPtr(true),
		IncludeAllVersions: boolPtr(true),
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_package.getPackage", "api_error", err)
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return packageRow{Package: *item, FeedId: feedID, ProjectId: projectID}, nil
}

func newFeedClient(ctx context.Context, d *plugin.QueryData, hydrate string) (feed.Client, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "connection_error", err)
		return nil, err
	}
	client, err := feed.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "client_error", err)
		return nil, err
	}
	return client, nil
}
