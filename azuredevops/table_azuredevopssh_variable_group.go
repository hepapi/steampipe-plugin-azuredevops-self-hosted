package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// variableGroupRow carries the project the group was read from. The variables
// map is the API payload, which includes secret values when the token can read them.
type variableGroupRow struct {
	taskagent.VariableGroup
	ProjectId string
}

func tableAzureDevOpsVariableGroup(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_variable_group",
		Description: "Retrieve Azure DevOps variable groups.",
		List: &plugin.ListConfig{
			Hydrate: listVariableGroups,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Optional},
				{Name: "name", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "id", Require: plugin.Required},
				{Name: "project_id", Require: plugin.Required},
			},
			Hydrate: getVariableGroup,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "Variable group identifier.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "Project the group was read from. Omit it to visit every project.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "Variable group name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "Variable group description.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "Variable group type.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_shared",
				Description: "True when the group is shared with other projects.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "created_on",
				Description: "When the group was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedOn.Time"),
			},
			{
				Name:        "modified_on",
				Description: "When the group was last modified.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("ModifiedOn.Time"),
			},
			{
				Name:        "created_by",
				Description: "Display name of the user who created the group.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("CreatedBy.DisplayName"),
			},
			{
				Name:        "variables",
				Description: "Variables in the group. Secret values are included when the connection can read them.",
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

func listVariableGroups(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := newTaskAgentClient(ctx, d, "azuredevopssh_variable_group.listVariableGroups")
	if err != nil {
		return nil, err
	}
	name := qualString(d, "name")
	if projectID := qualString(d, "project_id"); projectID != "" {
		return nil, streamVariableGroups(ctx, d, client, projectID, name)
	}
	return nil, forEachProject(ctx, d, func(projectID string) error {
		return streamVariableGroups(ctx, d, client, projectID, name)
	})
}

func streamVariableGroups(ctx context.Context, d *plugin.QueryData, client taskagent.Client, projectID, name string) error {
	if d.RowsRemaining(ctx) == 0 {
		return nil
	}
	top := int(d.RowsRemaining(ctx))
	if top > 100 {
		top = 100
	}
	args := taskagent.GetVariableGroupsArgs{Project: &projectID, Top: &top}
	if name != "" {
		args.GroupName = &name
	}
	groups, err := client.GetVariableGroups(ctx, args)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_variable_group.listVariableGroups", "api_error", err)
		return err
	}
	if groups == nil {
		return nil
	}
	for _, item := range *groups {
		d.StreamListItem(ctx, variableGroupRow{VariableGroup: item, ProjectId: projectID})
		if d.RowsRemaining(ctx) == 0 {
			return nil
		}
	}
	return nil
}

func getVariableGroup(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	id, ok := qualInt(d, "id")
	projectID := qualString(d, "project_id")
	if !ok || projectID == "" {
		return nil, nil
	}
	client, err := newTaskAgentClient(ctx, d, "azuredevopssh_variable_group.getVariableGroup")
	if err != nil {
		return nil, err
	}
	groupID := int(id)
	groups, err := client.GetVariableGroupsById(ctx, taskagent.GetVariableGroupsByIdArgs{
		Project:  &projectID,
		GroupIds: &[]int{groupID},
	})
	if err != nil {
		if absentFromCollection(d, err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_variable_group.getVariableGroup", "api_error", err)
		return nil, err
	}
	if groups == nil || len(*groups) == 0 {
		return nil, nil
	}
	return variableGroupRow{VariableGroup: (*groups)[0], ProjectId: projectID}, nil
}
