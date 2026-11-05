package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsPipelineRun(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_pipeline_run",
		Description: "Retrieve runs of a single Azure DevOps pipeline.",
		List: &plugin.ListConfig{
			Hydrate: listPipelineRuns,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Required},
				{Name: "pipeline_id", Require: plugin.Required},
				{Name: "id", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "pipeline_id", "project_id"}),
			Hydrate:    getPipelineRun,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The run ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "Name of the run.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "ID of the project the pipeline belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "pipeline_id",
				Description: "ID of the pipeline.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "state",
				Description: "State of the run.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "result",
				Description: "Result of the run.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "REST URL of the run.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "created_date",
				Description: "Time the run was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedDate.Time"),
			},
			{
				Name:        "finished_date",
				Description: "Time the run finished.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("FinishedDate.Time"),
			},
			{
				Name:        "variables",
				Description: "Variables supplied to the run.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "resources",
				Description: "Resources used by the run.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "template_parameters",
				Description: "Template parameters supplied to the run.",
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

// PipelineRun is one pipeline run plus the project and pipeline it was read from.
type PipelineRun struct {
	pipelines.Run
	ProjectId  string
	PipelineId int
}

func listPipelineRuns(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	pipelineID, ok := qualInt(d, "pipeline_id")
	if projectID == "" || !ok {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_pipeline_run.listPipelineRuns", "connection_error", err)
		return nil, err
	}
	client := pipelines.NewClient(ctx, connection)
	id := int(pipelineID)
	runs, err := client.ListRuns(ctx, pipelines.ListRunsArgs{
		Project:    types.String(projectID),
		PipelineId: &id,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_pipeline_run.listPipelineRuns", "project_id", projectID, "pipeline_id", id, "api_error", err)
		return nil, err
	}
	if runs == nil {
		return nil, nil
	}

	runID, hasRunID := qualInt(d, "id")
	for _, run := range *runs {
		if hasRunID && (run.Id == nil || int64(*run.Id) != runID) {
			continue
		}
		d.StreamListItem(ctx, PipelineRun{
			Run:        run,
			ProjectId:  projectID,
			PipelineId: id,
		})
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func getPipelineRun(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	pipelineID, okPipeline := qualInt(d, "pipeline_id")
	runID, okRun := qualInt(d, "id")
	if projectID == "" || !okPipeline || !okRun {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_pipeline_run.getPipelineRun", "connection_error", err)
		return nil, err
	}
	client := pipelines.NewClient(ctx, connection)
	pipeline := int(pipelineID)
	run := int(runID)
	result, err := client.GetRun(ctx, pipelines.GetRunArgs{
		Project:    types.String(projectID),
		PipelineId: &pipeline,
		RunId:      &run,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_pipeline_run.getPipelineRun", "project_id", projectID, "pipeline_id", pipeline, "id", run, "api_error", err)
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &PipelineRun{
		Run:        *result,
		ProjectId:  projectID,
		PipelineId: pipeline,
	}, nil
}
