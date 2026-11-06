package azuredevops

import (
	"context"
	"strconv"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsTestCase(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_test_case",
		Description: "Retrieve test cases in one Azure DevOps test suite.",
		List: &plugin.ListConfig{
			Hydrate: listTestCases,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Required},
				{Name: "plan_id", Require: plugin.Required},
				{Name: "suite_id", Require: plugin.Required},
				{Name: "id", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "suite_id", "plan_id", "project_id"}),
			Hydrate:    getTestCase,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "Work item ID of the test case.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("WorkItem.Id"),
			},
			{
				Name:        "name",
				Description: "Name of the test case.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("WorkItem.Name"),
			},
			{
				Name:        "project_id",
				Description: "ID of the project this test case was read from.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "plan_id",
				Description: "ID of the test plan.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "suite_id",
				Description: "ID of the test suite.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "order",
				Description: "Order of the test case in the suite.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "point_assignments",
				Description: "Point assignments for the test case.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "work_item",
				Description: "Work item details of the test case.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "links",
				Description: "REST reference links for the test case.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Links"),
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("WorkItem.Name"),
			},
		}),
	}
}

// TestCaseRow is one suite test case plus the project, plan, and suite it was read from.
type TestCaseRow struct {
	testplan.TestCase
	ProjectId string
	PlanId    int
	SuiteId   int
}

func listTestCases(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	planID, okPlan := qualInt(d, "plan_id")
	suiteID, okSuite := qualInt(d, "suite_id")
	if projectID == "" || !okPlan || !okSuite {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_case.listTestCases", "connection_error", err)
		return nil, err
	}
	client := testplan.NewClient(ctx, connection)

	plan := int(planID)
	suite := int(suiteID)
	identity := true
	input := testplan.GetTestCaseListArgs{
		Project:           types.String(projectID),
		PlanId:            &plan,
		SuiteId:           &suite,
		ReturnIdentityRef: &identity,
	}
	if id, ok := qualInt(d, "id"); ok {
		input.TestIds = types.String(strconv.FormatInt(id, 10))
	}
	for {
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
		cases, err := client.GetTestCaseList(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_test_case.listTestCases", "project_id", projectID, "plan_id", plan, "suite_id", suite, "api_error", err)
			return nil, err
		}
		if cases == nil {
			return nil, nil
		}
		for _, testCase := range cases.Value {
			d.StreamListItem(ctx, TestCaseRow{TestCase: testCase, ProjectId: projectID, PlanId: plan, SuiteId: suite})
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		if cases.ContinuationToken == "" {
			return nil, nil
		}
		input.ContinuationToken = types.String(cases.ContinuationToken)
	}
}

func getTestCase(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	planID, okPlan := qualInt(d, "plan_id")
	suiteID, okSuite := qualInt(d, "suite_id")
	caseID, okCase := qualInt(d, "id")
	if projectID == "" || !okPlan || !okSuite || !okCase {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_case.getTestCase", "connection_error", err)
		return nil, err
	}
	client := testplan.NewClient(ctx, connection)

	plan := int(planID)
	suite := int(suiteID)
	testCaseID := strconv.FormatInt(caseID, 10)
	cases, err := client.GetTestCase(ctx, testplan.GetTestCaseArgs{
		Project:    types.String(projectID),
		PlanId:     &plan,
		SuiteId:    &suite,
		TestCaseId: &testCaseID,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_test_case.getTestCase", "project_id", projectID, "plan_id", plan, "suite_id", suite, "id", testCaseID, "api_error", err)
		return nil, err
	}
	if cases == nil || len(*cases) == 0 {
		return nil, nil
	}
	return &TestCaseRow{TestCase: (*cases)[0], ProjectId: projectID, PlanId: plan, SuiteId: suite}, nil
}
