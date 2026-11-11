package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/test"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsTestResult(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_test_result",
		Description: "Retrieve test results for one Azure DevOps test run.",
		List: &plugin.ListConfig{
			Hydrate: listTestResults,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Required},
				{Name: "run_id", Require: plugin.Required},
				{Name: "id", Require: plugin.Optional},
				{Name: "outcome", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "run_id", "project_id"}),
			Hydrate:    getTestResult,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The test result ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "run_id",
				Description: "ID of the test run.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this result was read from.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "test_case_title",
				Description: "Name of the test.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "outcome",
				Description: "Outcome of the test result.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "state",
				Description: "State of the test result.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "priority",
				Description: "Priority of the executed test.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "duration_in_ms",
				Description: "Duration of test execution in milliseconds.",
				Type:        proto.ColumnType_DOUBLE,
			},
			{
				Name:        "failure_type",
				Description: "Failure type of the test result.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "error_message",
				Description: "Error message from test execution.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "automated_test_name",
				Description: "Fully qualified name of the executed test.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "computer_name",
				Description: "Machine where the test executed.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "started_date",
				Description: "Time test execution started.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("StartedDate.Time"),
			},
			{
				Name:        "completed_date",
				Description: "Time test execution completed.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CompletedDate.Time"),
			},
			{
				Name:        "url",
				Description: "REST URL of the test result.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "test_case",
				Description: "Reference to the executed test case.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "test_point",
				Description: "Reference to the executed test point.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "owner",
				Description: "Owner of the test result.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "run_by",
				Description: "Identity that executed the test.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("TestCaseTitle"),
			},
		}),
	}
}

// TestResultRow is one test result plus the project and run it was read from.
type TestResultRow struct {
	test.TestCaseResult
	ProjectId string
	RunId     int
}

func listTestResults(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	runID, ok := qualInt(d, "run_id")
	if projectID == "" || !ok {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_result.listTestResults", "connection_error", err)
		return nil, err
	}
	client, err := test.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_result.listTestResults", "client_error", err)
		return nil, err
	}

	run := int(runID)
	skip := 0
	resultID, hasResultID := qualInt(d, "id")
	var outcomes *[]test.TestOutcome
	if outcome := qualString(d, "outcome"); outcome != "" {
		value := test.TestOutcome(outcome)
		outcomes = &[]test.TestOutcome{value}
	}
	for {
		top := 1000
		if remaining := int(d.RowsRemaining(ctx)); remaining < top {
			top = remaining
		}
		if top <= 0 {
			return nil, nil
		}
		results, err := client.GetTestResults(ctx, test.GetTestResultsArgs{
			Project:  types.String(projectID),
			RunId:    &run,
			Skip:     &skip,
			Top:      &top,
			Outcomes: outcomes,
		})
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_test_result.listTestResults", "project_id", projectID, "run_id", run, "api_error", err)
			return nil, err
		}
		if results == nil || len(*results) == 0 {
			return nil, nil
		}
		for _, result := range *results {
			if hasResultID && (result.Id == nil || int64(*result.Id) != resultID) {
				continue
			}
			d.StreamListItem(ctx, TestResultRow{TestCaseResult: result, ProjectId: projectID, RunId: run})
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		if len(*results) < top || hasResultID {
			return nil, nil
		}
		skip += len(*results)
	}
}

func getTestResult(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	runID, okRun := qualInt(d, "run_id")
	resultID, okResult := qualInt(d, "id")
	if projectID == "" || !okRun || !okResult {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_result.getTestResult", "connection_error", err)
		return nil, err
	}
	client, err := test.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_result.getTestResult", "client_error", err)
		return nil, err
	}

	run := int(runID)
	result := int(resultID)
	item, err := client.GetTestResultById(ctx, test.GetTestResultByIdArgs{
		Project:          types.String(projectID),
		RunId:            &run,
		TestCaseResultId: &result,
	})
	if err != nil {
		if absentFromCollection(d, err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_test_result.getTestResult", "project_id", projectID, "run_id", run, "id", result, "api_error", err)
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return &TestResultRow{TestCaseResult: *item, ProjectId: projectID, RunId: run}, nil
}
