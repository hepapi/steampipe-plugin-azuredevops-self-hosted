package azuredevops

import (
	"context"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// packageVersionRow carries the feed and package that own this version.
type packageVersionRow struct {
	feed.PackageVersion
	FeedId    string
	PackageId string
	ProjectId string
}

func tableAzureDevOpsPackageVersion(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_package_version",
		Description: "Retrieve versions of a package in an Azure Artifacts feed.",
		List: &plugin.ListConfig{
			Hydrate: listPackageVersions,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "feed_id", Require: plugin.Required},
				{Name: "package_id", Require: plugin.Required},
				{Name: "project_id", Require: plugin.Optional},
				{Name: "id", Require: plugin.Optional},
				{Name: "is_deleted", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "id", Require: plugin.Required},
				{Name: "feed_id", Require: plugin.Required},
				{Name: "package_id", Require: plugin.Required},
				{Name: "project_id", Require: plugin.Optional},
			},
			Hydrate: getPackageVersion,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "Package version identifier.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "feed_id",
				Description: "Feed that contains the package.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "package_id",
				Description: "Package this version belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "Project that owns the feed. Empty for a collection feed.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "version",
				Description: "Display version.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "normalized_version",
				Description: "Version normalized for the package protocol.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_latest",
				Description: "True when this is the latest version by protocol sort order.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_deleted",
				Description: "True when this version has been deleted.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_listed",
				Description: "True when this version is listed. Applies to NuGet and Cargo.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "publish_date",
				Description: "When the version was published.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("PublishDate.Time"),
			},
			{
				Name:        "deleted_date",
				Description: "When the version was deleted.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("DeletedDate.Time"),
			},
			{
				Name:        "author",
				Description: "Author of the package version.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "Package version description.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "URL of the package version.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "tags",
				Description: "Tags on the package version.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "views",
				Description: "Feed views that contain this version.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "dependencies",
				Description: "Dependencies of this package version.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Version"),
			},
		}),
	}
}

func listPackageVersions(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	feedID := qualString(d, "feed_id")
	packageID := qualString(d, "package_id")
	if feedID == "" || packageID == "" {
		return nil, fmt.Errorf("azuredevops_package_version requires feed_id and package_id")
	}
	projectID := qualString(d, "project_id")
	id := qualString(d, "id")

	client, err := newFeedClient(ctx, d, "azuredevops_package_version.listPackageVersions")
	if err != nil {
		return nil, err
	}

	args := feed.GetPackageVersionsArgs{
		FeedId:      &feedID,
		PackageId:   &packageID,
		Project:     optionalProject(projectID),
		IncludeUrls: boolPtr(true),
	}
	if deleted, ok := qualBool(d, "is_deleted"); ok {
		args.IsDeleted = &deleted
	}

	versions, err := client.GetPackageVersions(ctx, args)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_package_version.listPackageVersions", "api_error", err)
		return nil, err
	}
	if versions == nil {
		return nil, nil
	}

	for _, item := range *versions {
		if id != "" && (item.Id == nil || item.Id.String() != id) {
			continue
		}
		d.StreamListItem(ctx, packageVersionRow{
			PackageVersion: item,
			FeedId:         feedID,
			PackageId:      packageID,
			ProjectId:      projectID,
		})
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func getPackageVersion(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	id := qualString(d, "id")
	feedID := qualString(d, "feed_id")
	packageID := qualString(d, "package_id")
	if id == "" || feedID == "" || packageID == "" {
		return nil, nil
	}
	projectID := qualString(d, "project_id")

	client, err := newFeedClient(ctx, d, "azuredevops_package_version.getPackageVersion")
	if err != nil {
		return nil, err
	}

	item, err := client.GetPackageVersion(ctx, feed.GetPackageVersionArgs{
		FeedId:           &feedID,
		PackageId:        &packageID,
		PackageVersionId: &id,
		Project:          optionalProject(projectID),
		IncludeUrls:      boolPtr(true),
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_package_version.getPackageVersion", "api_error", err)
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return packageVersionRow{
		PackageVersion: *item,
		FeedId:         feedID,
		PackageId:      packageID,
		ProjectId:      projectID,
	}, nil
}
