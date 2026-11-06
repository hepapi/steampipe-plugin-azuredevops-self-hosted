package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelineschecks"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsPipelineCheck(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_pipeline_check",
		Description: "Retrieve check configurations on one Azure DevOps pipeline resource.",
		List: &plugin.ListConfig{
			Hydrate: listPipelineChecks,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Required},
				{Name: "resource_type", Require: plugin.Required},
				{Name: "resource_id", Require: plugin.Required},
				{Name: "id", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The check configuration ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this check belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "resource_id",
				Description: "ID of the resource the check is configured on.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Resource.Id"),
			},
			{
				Name:        "resource_type",
				Description: "Type of the resource, such as endpoint, environment, queue, or variablegroup.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Resource.Type"),
			},
			{
				Name:        "resource_name",
				Description: "Name of the resource the check is configured on.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Resource.Name"),
			},
			{
				Name:        "type_id",
				Description: "ID of the check type.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Type.Id"),
			},
			{
				Name:        "type_name",
				Description: "Name of the check type.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Type.Name"),
			},
			{
				Name:        "url",
				Description: "REST URL of the check configuration.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "timeout",
				Description: "Timeout of the check, in minutes.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "created_on",
				Description: "Time the check was configured.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedOn.Time"),
			},
			{
				Name:        "modified_on",
				Description: "Time the check was last modified.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("ModifiedOn.Time"),
			},
			{
				Name:        "created_by",
				Description: "Identity that configured the check.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "modified_by",
				Description: "Identity that last modified the check.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "issue",
				Description: "Issue connected to the check configuration.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Type.Name"),
			},
		}),
	}
}

// PipelineCheck is a check configuration plus the project it was read from.
type PipelineCheck struct {
	pipelineschecks.CheckConfiguration
	ProjectId string
}

func listPipelineChecks(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	resourceType := qualString(d, "resource_type")
	resourceID := qualString(d, "resource_id")
	if projectID == "" || resourceType == "" || resourceID == "" {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_pipeline_check.listPipelineChecks", "connection_error", err)
		return nil, err
	}
	client, err := pipelineschecks.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_pipeline_check.listPipelineChecks", "client_error", err)
		return nil, err
	}

	checks, err := client.GetCheckConfigurationsOnResource(ctx, pipelineschecks.GetCheckConfigurationsOnResourceArgs{
		Project:      types.String(projectID),
		ResourceType: &resourceType,
		ResourceId:   &resourceID,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_pipeline_check.listPipelineChecks", "project_id", projectID, "resource_type", resourceType, "resource_id", resourceID, "api_error", err)
		return nil, err
	}
	if checks == nil {
		return nil, nil
	}

	checkID, hasCheckID := qualInt(d, "id")
	for _, check := range *checks {
		if hasCheckID && (check.Id == nil || int64(*check.Id) != checkID) {
			continue
		}
		d.StreamListItem(ctx, PipelineCheck{
			CheckConfiguration: check,
			ProjectId:          projectID,
		})
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}
