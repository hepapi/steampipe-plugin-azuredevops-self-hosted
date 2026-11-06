package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

type agentRow struct {
	taskagent.TaskAgent
	PoolId int
}

func tableAzureDevOpsAgent(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_agent",
		Description: "Retrieve agents in an Azure DevOps agent pool.",
		List: &plugin.ListConfig{
			Hydrate: listAgents,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "pool_id", Require: plugin.Optional},
				{Name: "name", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "id", Require: plugin.Required},
				{Name: "pool_id", Require: plugin.Required},
			},
			Hydrate: getAgent,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "Agent identifier.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "pool_id",
				Description: "Agent pool that contains the agent. Omit it to read agents from each pool until the row limit is reached.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "Agent name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "version",
				Description: "Agent version.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "Agent status, such as online or offline.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "enabled",
				Description: "True when the agent can run jobs.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "os_description",
				Description: "Operating system of the agent host.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "provisioning_state",
				Description: "Provisioning state of the agent.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "created_on",
				Description: "When the agent was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedOn.Time"),
			},
			{
				Name:        "status_changed_on",
				Description: "When the agent status last changed.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("StatusChangedOn.Time"),
			},
			{
				Name:        "system_capabilities",
				Description: "System capabilities reported by the agent.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "user_capabilities",
				Description: "User-defined capabilities on the agent.",
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

func listAgents(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := newTaskAgentClient(ctx, d, "azuredevops_agent.listAgents")
	if err != nil {
		return nil, err
	}
	name := qualString(d, "name")
	if poolID, ok := qualInt(d, "pool_id"); ok {
		return nil, streamAgents(ctx, d, client, int(poolID), name)
	}

	pools, err := client.GetAgentPools(ctx, taskagent.GetAgentPoolsArgs{})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_agent.listAgents", "api_error", err)
		return nil, err
	}
	if pools == nil {
		return nil, nil
	}
	for _, pool := range *pools {
		if pool.Id == nil {
			continue
		}
		if err := streamAgents(ctx, d, client, *pool.Id, name); err != nil {
			return nil, err
		}
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func streamAgents(ctx context.Context, d *plugin.QueryData, client taskagent.Client, poolID int, name string) error {
	if d.RowsRemaining(ctx) == 0 {
		return nil
	}
	args := taskagent.GetAgentsArgs{
		PoolId:              &poolID,
		IncludeCapabilities: boolPtr(true),
	}
	if name != "" {
		args.AgentName = &name
	}
	agents, err := client.GetAgents(ctx, args)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_agent.listAgents", "api_error", err)
		return err
	}
	if agents == nil {
		return nil
	}
	for _, item := range *agents {
		d.StreamListItem(ctx, agentRow{TaskAgent: item, PoolId: poolID})
		if d.RowsRemaining(ctx) == 0 {
			return nil
		}
	}
	return nil
}

func getAgent(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	id, ok := qualInt(d, "id")
	poolID, poolOK := qualInt(d, "pool_id")
	if !ok || !poolOK {
		return nil, nil
	}
	client, err := newTaskAgentClient(ctx, d, "azuredevops_agent.getAgent")
	if err != nil {
		return nil, err
	}
	agentID := int(id)
	pool := int(poolID)
	item, err := client.GetAgent(ctx, taskagent.GetAgentArgs{
		PoolId:              &pool,
		AgentId:             &agentID,
		IncludeCapabilities: boolPtr(true),
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_agent.getAgent", "api_error", err)
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return agentRow{TaskAgent: *item, PoolId: pool}, nil
}
