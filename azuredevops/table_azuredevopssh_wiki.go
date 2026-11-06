package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/wiki"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsWiki(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_wiki",
		Description: "Retrieve Azure DevOps wikis in the collection or in one project.",
		List: &plugin.ListConfig{
			Hydrate: listWikis,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "id", Require: plugin.Required},
				{Name: "project_id", Require: plugin.Optional},
			},
			Hydrate: getWiki,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "Wiki identifier.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "Wiki name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "Wiki type: projectWiki or codeWiki.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "Project that owns the wiki.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "repository_id",
				Description: "Git repository that backs the wiki.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "mapped_path",
				Description: "Repository folder published as the wiki. Project wikis use /.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "remote_url",
				Description: "Web URL of the wiki.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "REST URL of the wiki.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_disabled",
				Description: "True when the wiki repository is disabled.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "versions",
				Description: "Versions published for the wiki.",
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

func listWikis(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := newWikiClient(ctx, d, "azuredevopssh_wiki.listWikis")
	if err != nil {
		return nil, err
	}
	wikis, err := client.GetAllWikis(ctx, wiki.GetAllWikisArgs{Project: optionalProject(qualString(d, "project_id"))})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_wiki.listWikis", "api_error", err)
		return nil, err
	}
	if wikis == nil {
		return nil, nil
	}
	for _, item := range *wikis {
		d.StreamListItem(ctx, item)
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func getWiki(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	id := qualString(d, "id")
	if id == "" {
		return nil, nil
	}
	client, err := newWikiClient(ctx, d, "azuredevopssh_wiki.getWiki")
	if err != nil {
		return nil, err
	}
	item, err := client.GetWiki(ctx, wiki.GetWikiArgs{
		WikiIdentifier: &id,
		Project:        optionalProject(qualString(d, "project_id")),
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_wiki.getWiki", "api_error", err)
		return nil, err
	}
	return item, nil
}

func newWikiClient(ctx context.Context, d *plugin.QueryData, hydrate string) (wiki.Client, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "connection_error", err)
		return nil, err
	}
	client, err := wiki.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "client_error", err)
		return nil, err
	}
	return client, nil
}
