package azuredevops

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelinesapproval"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsPipelineApproval(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_pipeline_approval",
		Description: "Retrieve one Azure DevOps YAML pipeline approval by id.",
		List: &plugin.ListConfig{
			Hydrate: listPipelineApprovals,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "project_id", Require: plugin.Required},
				{Name: "id", Require: plugin.Required},
				{Name: "status", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "project_id"}),
			Hydrate:    getPipelineApproval,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The approval ID.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this approval belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "Overall status of the approval.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "execution_order",
				Description: "Order in which approvers become actionable.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "instructions",
				Description: "Instructions shown to approvers.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "min_required_approvers",
				Description: "Minimum number of approvers required.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "created_on",
				Description: "Time the approval was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedOn.Time"),
			},
			{
				Name:        "last_modified_on",
				Description: "Time the approval was last modified.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("LastModifiedOn.Time"),
			},
			{
				Name:        "steps",
				Description: "Approval steps.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "permissions",
				Description: "Current user permissions on the approval.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "blocked_approvers",
				Description: "Identities that are not allowed to approve.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Status"),
			},
		}),
	}
}

// PipelineApproval is a YAML pipeline approval plus the project it was listed from.
type PipelineApproval struct {
	pipelinesapproval.Approval
	ProjectId string
}

func listPipelineApprovals(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	id := qualString(d, "id")
	if projectID == "" || id == "" {
		return nil, nil
	}
	approvalID, err := uuid.Parse(id)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_pipeline_approval.listPipelineApprovals", "id", id, "parse_error", err)
		return nil, err
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_pipeline_approval.listPipelineApprovals", "connection_error", err)
		return nil, err
	}
	client, err := pipelinesapproval.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_pipeline_approval.listPipelineApprovals", "client_error", err)
		return nil, err
	}

	expand := pipelinesapproval.ApprovalDetailsExpandParameterValues.Steps
	approvals, err := client.QueryApprovals(ctx, pipelinesapproval.QueryApprovalsArgs{
		Project:     types.String(projectID),
		ApprovalIds: &[]uuid.UUID{approvalID},
		Expand:      &expand,
	})
	if err != nil {
		if approvalNotFound(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_pipeline_approval.listPipelineApprovals", "project_id", projectID, "id", id, "api_error", err)
		return nil, err
	}
	if approvals == nil {
		return nil, nil
	}

	statusQual := qualString(d, "status")
	for _, approval := range *approvals {
		if statusQual != "" && (approval.Status == nil || !strings.EqualFold(string(*approval.Status), statusQual)) {
			continue
		}
		d.StreamListItem(ctx, PipelineApproval{
			Approval:  approval,
			ProjectId: projectID,
		})
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func getPipelineApproval(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	id := qualString(d, "id")
	if projectID == "" || id == "" {
		return nil, nil
	}
	approvalID, err := uuid.Parse(id)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_pipeline_approval.getPipelineApproval", "id", id, "parse_error", err)
		return nil, err
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_pipeline_approval.getPipelineApproval", "connection_error", err)
		return nil, err
	}
	client, err := pipelinesapproval.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_pipeline_approval.getPipelineApproval", "client_error", err)
		return nil, err
	}

	expand := pipelinesapproval.ApprovalDetailsExpandParameterValues.Steps
	approval, err := client.GetApproval(ctx, pipelinesapproval.GetApprovalArgs{
		Project:    types.String(projectID),
		ApprovalId: &approvalID,
		Expand:     &expand,
	})
	if err != nil {
		if approvalNotFound(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_pipeline_approval.getPipelineApproval", "project_id", projectID, "id", id, "api_error", err)
		return nil, err
	}
	if approval == nil {
		return nil, nil
	}
	return &PipelineApproval{
		Approval:  *approval,
		ProjectId: projectID,
	}, nil
}

// approvalNotFound reports the Azure DevOps Server response for an approval
// id that does not exist. Querying by id is required on Server 2022, and a
// missing id is an empty result.
func approvalNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "No approvals found") || strings.Contains(msg, "ApprovalNotFoundException")
}
