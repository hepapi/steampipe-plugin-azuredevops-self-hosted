package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/test"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsTestRun(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_test_run",
		Description: "Retrieve test runs for Azure DevOps projects.",
		List: &plugin.ListConfig{
			Hydrate: listTestRuns,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Optional},
				{Name: "id", Require: plugin.Optional},
				{Name: "is_automated", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "project_id"}),
			Hydrate:    getTestRun,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The test run ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "Name of the test run.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this test run was read from.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "state",
				Description: "State of the test run.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_automated",
				Description: "True when the test run is automated.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "total_tests",
				Description: "Total number of tests in the run.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "passed_tests",
				Description: "Number of passed tests in the run.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "incomplete_tests",
				Description: "Number of incomplete tests in the run.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "unanalyzed_tests",
				Description: "Number of failed tests that have not been analyzed.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "not_applicable_tests",
				Description: "Number of not-applicable tests in the run.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "created_date",
				Description: "Time the test run was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedDate.Time"),
			},
			{
				Name:        "started_date",
				Description: "Time the test run started.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("StartedDate.Time"),
			},
			{
				Name:        "completed_date",
				Description: "Time the test run completed.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CompletedDate.Time"),
			},
			{
				Name:        "url",
				Description: "REST URL of the test run.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "web_access_url",
				Description: "Web access URL of the test run.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "plan",
				Description: "Test plan associated with the run.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "build",
				Description: "Build associated with the run.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "owner",
				Description: "Owner of the test run.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "pipeline_reference",
				Description: "Pipeline this test run belongs to.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "run_statistics",
				Description: "Run summary by outcome.",
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

// TestRunRow is one test run plus the project it was read from.
type TestRunRow struct {
	test.TestRun
	ProjectId string
}

func listTestRuns(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_run.listTestRuns", "connection_error", err)
		return nil, err
	}
	client, err := test.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_run.listTestRuns", "client_error", err)
		return nil, err
	}

	if projectID := qualString(d, "project_id"); projectID != "" {
		return nil, listTestRunsForProject(ctx, d, client, projectID)
	}
	return nil, forEachProject(ctx, d, func(projectID string) error {
		return listTestRunsForProject(ctx, d, client, projectID)
	})
}

func listTestRunsForProject(ctx context.Context, d *plugin.QueryData, client test.Client, projectID string) error {
	details := true
	skip := 0
	runID, hasRunID := qualInt(d, "id")
	for {
		top := 100
		if remaining := int(d.RowsRemaining(ctx)); remaining < top {
			top = remaining
		}
		if top <= 0 {
			return nil
		}
		input := test.GetTestRunsArgs{
			Project:           types.String(projectID),
			IncludeRunDetails: &details,
			Skip:              &skip,
			Top:               &top,
		}
		if automated, ok := qualBool(d, "is_automated"); ok {
			input.Automated = &automated
		}
		runs, err := client.GetTestRuns(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_test_run.listTestRuns", "project_id", projectID, "api_error", err)
			return err
		}
		if runs == nil || len(*runs) == 0 {
			return nil
		}
		for _, run := range *runs {
			if hasRunID && (run.Id == nil || int64(*run.Id) != runID) {
				continue
			}
			d.StreamListItem(ctx, TestRunRow{TestRun: run, ProjectId: projectID})
			if d.RowsRemaining(ctx) == 0 {
				return nil
			}
		}
		if len(*runs) < top || hasRunID {
			return nil
		}
		skip += len(*runs)
	}
}

func getTestRun(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	id, ok := qualInt(d, "id")
	if projectID == "" || !ok {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_run.getTestRun", "connection_error", err)
		return nil, err
	}
	client, err := test.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_run.getTestRun", "client_error", err)
		return nil, err
	}

	runID := int(id)
	details := true
	run, err := client.GetTestRunById(ctx, test.GetTestRunByIdArgs{
		Project:        types.String(projectID),
		RunId:          &runID,
		IncludeDetails: &details,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_run.getTestRun", "project_id", projectID, "id", runID, "api_error", err)
		return nil, err
	}
	if run == nil {
		return nil, nil
	}
	return &TestRunRow{TestRun: *run, ProjectId: projectID}, nil
}
