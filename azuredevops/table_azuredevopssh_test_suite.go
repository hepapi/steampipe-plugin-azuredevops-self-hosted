package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsTestSuite(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_test_suite",
		Description: "Retrieve test suites for one Azure DevOps test plan.",
		List: &plugin.ListConfig{
			Hydrate: listTestSuites,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Required},
				{Name: "plan_id", Require: plugin.Required},
				{Name: "id", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "plan_id", "project_id"}),
			Hydrate:    getTestSuite,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The test suite ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "Name of the test suite.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this suite belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "plan_id",
				Description: "ID of the test plan this suite belongs to.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "suite_type",
				Description: "Type of the test suite.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "parent_suite_id",
				Description: "ID of the parent suite.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("ParentSuite.Id"),
			},
			{
				Name:        "requirement_id",
				Description: "Requirement work item ID for a requirement-based suite.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "revision",
				Description: "Revision of the test suite.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "has_children",
				Description: "True when the suite has child suites.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "query_string",
				Description: "WIQL query for a dynamic suite.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "last_updated_date",
				Description: "Time the suite was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("LastUpdatedDate.Time"),
			},
			{
				Name:        "parent_suite",
				Description: "Parent suite reference.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "plan",
				Description: "Test plan this suite belongs to.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "last_updated_by",
				Description: "Identity that last updated the suite.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "links",
				Description: "REST reference links for the suite.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Links"),
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

// TestSuiteRow is one test suite plus the project and plan it was read from.
type TestSuiteRow struct {
	testplan.TestSuite
	ProjectId string
	PlanId    int
}

func listTestSuites(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	planID, ok := qualInt(d, "plan_id")
	if projectID == "" || !ok {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_suite.listTestSuites", "connection_error", err)
		return nil, err
	}
	client := testplan.NewClient(ctx, connection)

	plan := int(planID)
	expand := testplan.SuiteExpandValues.None
	asTree := false
	input := testplan.GetTestSuitesForPlanArgs{
		Project:    types.String(projectID),
		PlanId:     &plan,
		Expand:     &expand,
		AsTreeView: &asTree,
	}
	suiteID, hasSuiteID := qualInt(d, "id")
	for {
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
		suites, err := client.GetTestSuitesForPlan(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_test_suite.listTestSuites", "project_id", projectID, "plan_id", plan, "api_error", err)
			return nil, err
		}
		if suites == nil {
			return nil, nil
		}
		for _, suite := range suites.Value {
			if hasSuiteID && (suite.Id == nil || int64(*suite.Id) != suiteID) {
				continue
			}
			d.StreamListItem(ctx, TestSuiteRow{TestSuite: suite, ProjectId: projectID, PlanId: plan})
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		if suites.ContinuationToken == "" {
			return nil, nil
		}
		input.ContinuationToken = types.String(suites.ContinuationToken)
	}
}

func getTestSuite(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	planID, okPlan := qualInt(d, "plan_id")
	suiteID, okSuite := qualInt(d, "id")
	if projectID == "" || !okPlan || !okSuite {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_suite.getTestSuite", "connection_error", err)
		return nil, err
	}
	client := testplan.NewClient(ctx, connection)

	plan := int(planID)
	suite := int(suiteID)
	result, err := client.GetTestSuiteById(ctx, testplan.GetTestSuiteByIdArgs{
		Project: types.String(projectID),
		PlanId:  &plan,
		SuiteId: &suite,
	})
	if err != nil {
		if absentFromCollection(d, err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_test_suite.getTestSuite", "project_id", projectID, "plan_id", plan, "id", suite, "api_error", err)
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &TestSuiteRow{TestSuite: *result, ProjectId: projectID, PlanId: plan}, nil
}
