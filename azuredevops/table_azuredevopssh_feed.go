package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// feedRow carries the project scope used to read the feed. Collection feeds
// leave ProjectId empty.
type feedRow struct {
	feed.Feed
	ProjectId string
}

func tableAzureDevOpsFeed(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_feed",
		Description: "Retrieve Azure Artifacts feeds in the collection or in one project.",
		List: &plugin.ListConfig{
			Hydrate: listFeeds,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "id", Require: plugin.Required},
				{Name: "project_id", Require: plugin.Optional},
			},
			Hydrate: getFeed,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "Feed identifier.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "Feed name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "fully_qualified_name",
				Description: "Full feed name, including the view when the feed was read through a view.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "fully_qualified_id",
				Description: "Feed id, including the view id when the feed was read through a view.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "Feed description.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "Project that owns the feed. Empty for a collection feed. Set this to list project feeds.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_name",
				Description: "Name of the project that owns the feed.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Project.Name"),
			},
			{
				Name:        "upstream_enabled",
				Description: "True when the feed can consume upstream sources.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "badges_enabled",
				Description: "True when the feed can generate package badges.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "hide_deleted_package_versions",
				Description: "True when deleted package versions are hidden.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_read_only",
				Description: "True when packages in the feed are immutable.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "url",
				Description: "URL of the feed.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "capabilities",
				Description: "Feed capability flags.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "upstream_sources",
				Description: "Sources this feed fetches packages from.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "deleted_date",
				Description: "When the feed was deleted.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("DeletedDate.Time"),
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

func listFeeds(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, projectID, err := feedClient(ctx, d, "azuredevopssh_feed.listFeeds")
	if err != nil {
		return nil, err
	}

	feeds, err := client.GetFeeds(ctx, feed.GetFeedsArgs{
		Project:     optionalProject(projectID),
		IncludeUrls: boolPtr(true),
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_feed.listFeeds", "api_error", err)
		return nil, err
	}
	return nil, streamFeeds(ctx, d, projectID, feeds)
}

func getFeed(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	id := qualString(d, "id")
	if id == "" {
		return nil, nil
	}
	client, projectID, err := feedClient(ctx, d, "azuredevopssh_feed.getFeed")
	if err != nil {
		return nil, err
	}

	item, err := client.GetFeed(ctx, feed.GetFeedArgs{
		FeedId:  &id,
		Project: optionalProject(projectID),
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_feed.getFeed", "api_error", err)
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return newFeedRow(projectID, *item), nil
}

func feedClient(ctx context.Context, d *plugin.QueryData, hydrate string) (feed.Client, string, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "connection_error", err)
		return nil, "", err
	}
	client, err := feed.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "client_error", err)
		return nil, "", err
	}
	return client, qualString(d, "project_id"), nil
}

func streamFeeds(ctx context.Context, d *plugin.QueryData, projectID string, feeds *[]feed.Feed) error {
	if feeds == nil {
		return nil
	}
	for _, item := range *feeds {
		d.StreamListItem(ctx, newFeedRow(projectID, item))
		if d.RowsRemaining(ctx) == 0 {
			return nil
		}
	}
	return nil
}

func newFeedRow(queriedProjectID string, item feed.Feed) feedRow {
	projectID := queriedProjectID
	if item.Project != nil && item.Project.Id != nil {
		projectID = item.Project.Id.String()
	}
	return feedRow{Feed: item, ProjectId: projectID}
}

func optionalProject(projectID string) *string {
	if projectID == "" {
		return nil
	}
	return &projectID
}

func boolPtr(value bool) *bool {
	return &value
}
