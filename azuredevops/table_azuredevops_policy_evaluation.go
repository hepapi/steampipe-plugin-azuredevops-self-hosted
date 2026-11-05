package azuredevops

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

const policyEvaluationPageSize = 100

func tableAzureDevOpsPolicyEvaluation(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_policy_evaluation",
		Description: "Retrieve policy evaluation records for a pull request or other policy artifact.",
		List: &plugin.ListConfig{
			Hydrate: listPolicyEvaluations,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "project_id", Require: plugin.Required},
				{Name: "artifact_id", Require: plugin.Required},
				{Name: "status", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "project_id"}),
			Hydrate:    getPolicyEvaluation,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "GUID which uniquely identifies this evaluation record.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("EvaluationId"),
			},
			{
				Name:        "project_id",
				Description: "ID of the project this evaluation belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "artifact_id",
				Description: "Identifier of the target of this policy evaluation, such as a pull request artifact id.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "Status of the policy (queued, running, approved, rejected, notApplicable, or broken).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "started_date",
				Description: "Time when this policy was first evaluated on this artifact.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("StartedDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "completed_date",
				Description: "Time when this policy finished evaluating on this artifact.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CompletedDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "configuration_id",
				Description: "ID of the policy configuration being evaluated.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Configuration.Id"),
			},
			{
				Name:        "policy_type_id",
				Description: "ID of the policy type being evaluated.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Configuration.Type.Id"),
			},
			{
				Name:        "policy_type_display_name",
				Description: "Display name of the policy type being evaluated.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Configuration.Type.DisplayName"),
			},
			{
				Name:        "is_enabled",
				Description: "Indicates whether the evaluated policy configuration is enabled.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("Configuration.IsEnabled"),
			},
			{
				Name:        "is_blocking",
				Description: "Indicates whether the evaluated policy configuration is blocking.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("Configuration.IsBlocking"),
			},
			{
				Name:        "configuration",
				Description: "Full configuration data for the policy which is being evaluated.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "context",
				Description: "Internal context data of this policy evaluation.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "links",
				Description: "Links to other related objects.",
				Type:        proto.ColumnType_JSON,
			},

			/// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Configuration.Type.DisplayName"),
			},
		}),
	}
}

// PolicyEvaluation is a policy evaluation record plus the project it was queried in.
type PolicyEvaluation struct {
	policy.PolicyEvaluationRecord
	ProjectId string
}

func listPolicyEvaluations(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	artifactID := qualString(d, "artifact_id")
	if projectID == "" || artifactID == "" {
		return nil, nil
	}
	statusQual := qualString(d, "status")

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_policy_evaluation.listPolicyEvaluations", "connection_error", err)
		return nil, err
	}
	client, err := policy.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_policy_evaluation.listPolicyEvaluations", "client_error", err)
		return nil, err
	}

	// Always request not-applicable records so a status qual can see them.
	// Callers who do not want those rows can filter status <> 'notApplicable'.
	skip := 0
	for {
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}

		// Keep full pages when status is filtered in-process. Shrinking the page
		// to the remaining row limit would skip later matches.
		top := policyEvaluationPageSize
		if statusQual == "" && d.QueryContext.Limit != nil {
			remaining := int(d.RowsRemaining(ctx))
			if remaining > 0 && remaining < top {
				top = remaining
			}
		}

		input := policy.GetPolicyEvaluationsArgs{
			Project:              types.String(projectID),
			ArtifactId:           types.String(artifactID),
			IncludeNotApplicable: types.Bool(true),
			Top:                  types.Int(top),
			Skip:                 types.Int(skip),
		}

		records, err := client.GetPolicyEvaluations(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevops_policy_evaluation.listPolicyEvaluations", "project_id", projectID, "artifact_id", artifactID, "api_error", err)
			return nil, err
		}
		if records == nil || len(*records) == 0 {
			return nil, nil
		}

		for _, record := range *records {
			if statusQual != "" && (record.Status == nil || !strings.EqualFold(string(*record.Status), statusQual)) {
				continue
			}
			d.StreamListItem(ctx, PolicyEvaluation{
				PolicyEvaluationRecord: record,
				ProjectId:              projectID,
			})
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if len(*records) < top {
			return nil, nil
		}
		skip += len(*records)
	}
}

func getPolicyEvaluation(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	id := qualString(d, "id")
	if projectID == "" || id == "" {
		return nil, nil
	}
	evaluationID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid policy evaluation id %q: %w", id, err)
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_policy_evaluation.getPolicyEvaluation", "connection_error", err)
		return nil, err
	}
	client, err := policy.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_policy_evaluation.getPolicyEvaluation", "client_error", err)
		return nil, err
	}

	record, err := client.GetPolicyEvaluation(ctx, policy.GetPolicyEvaluationArgs{
		Project:      types.String(projectID),
		EvaluationId: &evaluationID,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_policy_evaluation.getPolicyEvaluation", "project_id", projectID, "id", id, "api_error", err)
		return nil, err
	}
	if record == nil {
		return nil, nil
	}

	return PolicyEvaluation{PolicyEvaluationRecord: *record, ProjectId: projectID}, nil
}
