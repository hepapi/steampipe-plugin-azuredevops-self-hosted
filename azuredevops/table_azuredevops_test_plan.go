package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsTestPlan(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_test_plan",
		Description: "Retrieve test plans for Azure DevOps projects.",
		List: &plugin.ListConfig{
			Hydrate: listTestPlans,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Optional},
				{Name: "id", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "project_id"}),
			Hydrate:    getTestPlan,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The test plan ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "Name of the test plan.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this test plan belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "state",
				Description: "State of the test plan.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "area_path",
				Description: "Area path of the test plan.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "iteration",
				Description: "Iteration path of the test plan.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "revision",
				Description: "Revision of the test plan.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "description",
				Description: "Description of the test plan.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "start_date",
				Description: "Start date of the test plan.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("StartDate.Time"),
			},
			{
				Name:        "end_date",
				Description: "End date of the test plan.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("EndDate.Time"),
			},
			{
				Name:        "updated_date",
				Description: "Time the test plan was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("UpdatedDate.Time"),
			},
			{
				Name:        "owner",
				Description: "Owner of the test plan.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "updated_by",
				Description: "Identity that last updated the test plan.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "root_suite",
				Description: "Root test suite of the test plan.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "links",
				Description: "REST reference links for the test plan.",
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

// TestPlanRow is one test plan plus the project it was read from.
type TestPlanRow struct {
	testplan.TestPlan
	ProjectId string
}

func listTestPlans(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_test_plan.listTestPlans", "connection_error", err)
		return nil, err
	}
	client := testplan.NewClient(ctx, connection)

	if projectID := qualString(d, "project_id"); projectID != "" {
		return nil, listTestPlansForProject(ctx, d, client, projectID)
	}
	return nil, forEachProject(ctx, d, func(projectID string) error {
		return listTestPlansForProject(ctx, d, client, projectID)
	})
}

func listTestPlansForProject(ctx context.Context, d *plugin.QueryData, client testplan.Client, projectID string) error {
	details := true
	input := testplan.GetTestPlansArgs{
		Project:            types.String(projectID),
		IncludePlanDetails: &details,
	}
	planID, hasPlanID := qualInt(d, "id")
	for {
		if d.RowsRemaining(ctx) == 0 {
			return nil
		}
		plans, err := client.GetTestPlans(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevops_test_plan.listTestPlans", "project_id", projectID, "api_error", err)
			return err
		}
		if plans == nil {
			return nil
		}
		for _, plan := range plans.Value {
			if hasPlanID && (plan.Id == nil || int64(*plan.Id) != planID) {
				continue
			}
			d.StreamListItem(ctx, TestPlanRow{TestPlan: plan, ProjectId: projectID})
			if d.RowsRemaining(ctx) == 0 {
				return nil
			}
		}
		if plans.ContinuationToken == "" {
			return nil
		}
		input.ContinuationToken = types.String(plans.ContinuationToken)
	}
}

func getTestPlan(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	id, ok := qualInt(d, "id")
	if projectID == "" || !ok {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_test_plan.getTestPlan", "connection_error", err)
		return nil, err
	}
	client := testplan.NewClient(ctx, connection)

	planID := int(id)
	plan, err := client.GetTestPlanById(ctx, testplan.GetTestPlanByIdArgs{
		Project: types.String(projectID),
		PlanId:  &planID,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_test_plan.getTestPlan", "project_id", projectID, "id", planID, "api_error", err)
		return nil, err
	}
	if plan == nil {
		return nil, nil
	}
	return &TestPlanRow{TestPlan: *plan, ProjectId: projectID}, nil
}
