package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsAgentPool(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_agent_pool",
		Description: "Retrieve Azure DevOps agent pools.",
		List: &plugin.ListConfig{
			Hydrate: listAgentPools,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "name", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getAgentPool,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "Agent pool identifier.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "Agent pool name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "pool_type",
				Description: "Pool type, such as automation or deployment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "size",
				Description: "Number of agents in the pool.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "is_hosted",
				Description: "True when the pool is hosted by the service.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_legacy",
				Description: "True when the pool is a legacy pool.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "auto_provision",
				Description: "True when a queue is provisioned for each collection.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "auto_update",
				Description: "True when agents in the pool may update themselves.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "created_on",
				Description: "When the pool was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedOn.Time"),
			},
			{
				Name:        "created_by",
				Description: "Display name of the pool creator.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("CreatedBy.DisplayName"),
			},
			{
				Name:        "owner",
				Description: "Display name of the pool owner.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Owner.DisplayName"),
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

func listAgentPools(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := newTaskAgentClient(ctx, d, "azuredevops_agent_pool.listAgentPools")
	if err != nil {
		return nil, err
	}
	args := taskagent.GetAgentPoolsArgs{}
	if name := qualString(d, "name"); name != "" {
		args.PoolName = &name
	}
	pools, err := client.GetAgentPools(ctx, args)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_agent_pool.listAgentPools", "api_error", err)
		return nil, err
	}
	if pools == nil {
		return nil, nil
	}
	for _, item := range *pools {
		d.StreamListItem(ctx, item)
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func getAgentPool(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	id, ok := qualInt(d, "id")
	if !ok {
		return nil, nil
	}
	client, err := newTaskAgentClient(ctx, d, "azuredevops_agent_pool.getAgentPool")
	if err != nil {
		return nil, err
	}
	poolID := int(id)
	item, err := client.GetAgentPool(ctx, taskagent.GetAgentPoolArgs{PoolId: &poolID})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_agent_pool.getAgentPool", "api_error", err)
		return nil, err
	}
	return item, nil
}

func newTaskAgentClient(ctx context.Context, d *plugin.QueryData, hydrate string) (taskagent.Client, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "connection_error", err)
		return nil, err
	}
	client, err := taskagent.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "client_error", err)
		return nil, err
	}
	return client, nil
}
