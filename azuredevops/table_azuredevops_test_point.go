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

func tableAzureDevOpsTestPoint(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_test_point",
		Description: "Retrieve test points in one Azure DevOps test suite.",
		List: &plugin.ListConfig{
			Hydrate: listTestPoints,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Required},
				{Name: "plan_id", Require: plugin.Required},
				{Name: "suite_id", Require: plugin.Required},
				{Name: "id", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "suite_id", "plan_id", "project_id"}),
			Hydrate:    getTestPoint,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The test point ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this test point was read from.",
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
				Name:        "is_active",
				Description: "True when the test point is active.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_automated",
				Description: "True when the test point is for an automated test case.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "comment",
				Description: "Comment on the test point.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "configuration_id",
				Description: "ID of the configuration assigned to the test point.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Configuration.Id"),
			},
			{
				Name:        "configuration_name",
				Description: "Name of the configuration assigned to the test point.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Configuration.Name"),
			},
			{
				Name:        "test_case_id",
				Description: "ID of the test case.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("TestCaseReference.Id"),
			},
			{
				Name:        "test_case_name",
				Description: "Name of the test case.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("TestCaseReference.Name"),
			},
			{
				Name:        "outcome",
				Description: "Latest outcome of the test point.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Results.Outcome"),
			},
			{
				Name:        "last_updated_date",
				Description: "Time the test point was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("LastUpdatedDate.Time"),
			},
			{
				Name:        "tester",
				Description: "Tester assigned to the test point.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "results",
				Description: "Latest result summary for the test point.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "test_case_reference",
				Description: "Reference to the test case.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "links",
				Description: "REST reference links for the test point.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Links"),
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("TestCaseReference.Name"),
			},
		}),
	}
}

// TestPointRow is one test point plus the project, plan, and suite it was read from.
type TestPointRow struct {
	testplan.TestPoint
	ProjectId string
	PlanId    int
	SuiteId   int
}

func listTestPoints(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	planID, okPlan := qualInt(d, "plan_id")
	suiteID, okSuite := qualInt(d, "suite_id")
	if projectID == "" || !okPlan || !okSuite {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_test_point.listTestPoints", "connection_error", err)
		return nil, err
	}
	client := testplan.NewClient(ctx, connection)

	plan := int(planID)
	suite := int(suiteID)
	details := true
	identity := true
	input := testplan.GetPointsListArgs{
		Project:             types.String(projectID),
		PlanId:              &plan,
		SuiteId:             &suite,
		IncludePointDetails: &details,
		ReturnIdentityRef:   &identity,
	}
	if id, ok := qualInt(d, "id"); ok {
		input.TestPointIds = types.String(strconv.FormatInt(id, 10))
	}
	for {
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
		points, err := client.GetPointsList(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevops_test_point.listTestPoints", "project_id", projectID, "plan_id", plan, "suite_id", suite, "api_error", err)
			return nil, err
		}
		if points == nil {
			return nil, nil
		}
		for _, point := range points.Value {
			d.StreamListItem(ctx, TestPointRow{TestPoint: point, ProjectId: projectID, PlanId: plan, SuiteId: suite})
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		if points.ContinuationToken == "" {
			return nil, nil
		}
		input.ContinuationToken = types.String(points.ContinuationToken)
	}
}

func getTestPoint(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	planID, okPlan := qualInt(d, "plan_id")
	suiteID, okSuite := qualInt(d, "suite_id")
	pointID, okPoint := qualInt(d, "id")
	if projectID == "" || !okPlan || !okSuite || !okPoint {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_test_point.getTestPoint", "connection_error", err)
		return nil, err
	}
	client := testplan.NewClient(ctx, connection)

	plan := int(planID)
	suite := int(suiteID)
	point := strconv.FormatInt(pointID, 10)
	details := true
	points, err := client.GetPoints(ctx, testplan.GetPointsArgs{
		Project:             types.String(projectID),
		PlanId:              &plan,
		SuiteId:             &suite,
		PointId:             &point,
		IncludePointDetails: &details,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_test_point.getTestPoint", "project_id", projectID, "plan_id", plan, "suite_id", suite, "id", point, "api_error", err)
		return nil, err
	}
	if points == nil || len(*points) == 0 {
		return nil, nil
	}
	return &TestPointRow{TestPoint: (*points)[0], ProjectId: projectID, PlanId: plan, SuiteId: suite}, nil
}
