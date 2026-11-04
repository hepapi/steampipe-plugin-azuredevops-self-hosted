package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsProjectCollection(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_project_collection",
		Description: "Retrieve information about your project collections (self-hosted Azure DevOps Server).",
		List: &plugin.ListConfig{
			Hydrate: listProjectCollections,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getProjectCollection,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The collection identifier.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The collection name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "The collection REST URL.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "avatar_url",
				Description: "The collection avatar URL.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The project collection description.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getProjectCollection,
			},
			{
				Name:        "state",
				Description: "The project collection state.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getProjectCollection,
			},
			{
				Name:        "process_customization_type",
				Description: "Process customization type on this collection (Xml or Inherited).",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getProjectCollection,
			},
			{
				Name:        "links",
				Description: "The links to other objects related to this object.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getProjectCollection,
			},

			/// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
		}),
	}
}

func listProjectCollections(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getServerConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_project_collection.listProjectCollections", "connection_error", err)
		return nil, err
	}

	client, err := core.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_project_collection.listProjectCollections", "client_error", err)
		return nil, err
	}

	// Limiting the results
	maxLimit := 1000
	if d.QueryContext.Limit != nil {
		limit := int(*d.QueryContext.Limit)
		if limit < maxLimit {
			maxLimit = limit
		}
	}

	input := core.GetProjectCollectionsArgs{
		Top: types.Int(maxLimit),
	}

	collections, err := client.GetProjectCollections(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_project_collection.listProjectCollections", "api_error", err)
		return nil, err
	}

	for _, collection := range *collections {
		d.StreamListItem(ctx, collection)

		// Context can be cancelled due to manual cancellation or the limit has been hit
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}

func getProjectCollection(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	var id string
	if h.Item != nil {
		id = h.Item.(core.TeamProjectCollectionReference).Id.String()
	} else {
		id = d.EqualsQuals["id"].GetStringValue()
	}

	// Check if id is empty
	if id == "" {
		return nil, nil
	}

	connection, err := getServerConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_project_collection.getProjectCollection", "connection_error", err)
		return nil, err
	}

	client, err := core.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_project_collection.getProjectCollection", "client_error", err)
		return nil, err
	}

	input := core.GetProjectCollectionArgs{
		CollectionId: types.String(id),
	}

	collection, err := client.GetProjectCollection(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_project_collection.getProjectCollection", "api_error", err)
		return nil, err
	}

	return collection, nil
}
