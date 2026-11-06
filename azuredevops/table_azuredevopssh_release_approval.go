package azuredevops

import (
	"context"
	"strconv"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsReleaseApproval(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_release_approval",
		Description: "Retrieve approvals on Azure DevOps classic releases.",
		List: &plugin.ListConfig{
			ParentHydrate: listProjects,
			Hydrate:       listReleaseApprovals,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "project_id", Require: plugin.Optional},
				{Name: "id", Require: plugin.Optional},
				{Name: "status", Require: plugin.Optional},
				{Name: "approval_type", Require: plugin.Optional},
				{Name: "release_id", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The approval ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this approval belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "Status of the approval.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "approval_type",
				Description: "When the approval runs (preDeploy or postDeploy).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "attempt",
				Description: "Deployment attempt this approval belongs to.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "rank",
				Description: "Order of the approval. Equal ranks run in parallel.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "is_automated",
				Description: "True when the approval is automated.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "comments",
				Description: "Comments recorded on the approval.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "release_id",
				Description: "ID of the release this approval belongs to.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Release.Id"),
			},
			{
				Name:        "release_definition_id",
				Description: "ID of the release definition.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("ReleaseDefinition.Id"),
			},
			{
				Name:        "release_environment_id",
				Description: "ID of the release environment this approval belongs to.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("ReleaseEnvironment.Id"),
			},
			{
				Name:        "url",
				Description: "REST URL of the approval.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "created_on",
				Description: "Time the approval was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedOn.Time"),
			},
			{
				Name:        "modified_on",
				Description: "Time the approval was last modified.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("ModifiedOn.Time"),
			},
			{
				Name:        "approver",
				Description: "Identity that should approve.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "approved_by",
				Description: "Identity that approved.",
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

// ReleaseApprovalRow is a classic release approval plus the project it was listed from.
type ReleaseApprovalRow struct {
	release.ReleaseApproval
	ProjectId string
}

func listReleaseApprovals(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	project, ok := h.Item.(core.TeamProjectReference)
	if !ok || project.Id == nil {
		return nil, nil
	}
	projectID := project.Id.String()
	if qual := qualString(d, "project_id"); qual != "" && !strings.EqualFold(qual, projectID) {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_release_approval.listReleaseApprovals", "connection_error", err)
		return nil, err
	}
	client, err := release.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_release_approval.listReleaseApprovals", "client_error", err)
		return nil, err
	}

	statuses := []release.ApprovalStatus{
		release.ApprovalStatusValues.Pending,
		release.ApprovalStatusValues.Approved,
		release.ApprovalStatusValues.Rejected,
		release.ApprovalStatusValues.Canceled,
		release.ApprovalStatusValues.Skipped,
		release.ApprovalStatusValues.Reassigned,
	}
	if status := qualString(d, "status"); status != "" {
		statuses = []release.ApprovalStatus{release.ApprovalStatus(status)}
	}

	top := 50
	if d.QueryContext.Limit != nil && *d.QueryContext.Limit > 0 && int(*d.QueryContext.Limit) < top {
		top = int(*d.QueryContext.Limit)
	}

	approvalID, hasApprovalID := qualInt(d, "id")
	for _, status := range statuses {
		statusFilter := status
		// IncludeMyGroupApprovals is only valid together with a release id,
		// an assigned-to filter, and status pending. Leave it unset so a
		// status inventory is accepted.
		input := release.GetApprovalsArgs{
			Project:      types.String(projectID),
			StatusFilter: &statusFilter,
			Top:          &top,
		}
		if releaseID, ok := qualInt(d, "release_id"); ok {
			id := int(releaseID)
			input.ReleaseIdsFilter = &[]int{id}
		}
		if approvalType := qualString(d, "approval_type"); approvalType != "" {
			typeFilter := release.ApprovalType(approvalType)
			input.TypeFilter = &typeFilter
		}

		for {
			approvals, err := client.GetApprovals(ctx, input)
			if err != nil {
				plugin.Logger(ctx).Error("azuredevopssh_release_approval.listReleaseApprovals", "project_id", projectID, "status", string(status), "api_error", err)
				return nil, err
			}
			for _, approval := range approvals.Value {
				if hasApprovalID && (approval.Id == nil || int64(*approval.Id) != approvalID) {
					continue
				}
				d.StreamListItem(ctx, ReleaseApprovalRow{
					ReleaseApproval: approval,
					ProjectId:       projectID,
				})
				if d.RowsRemaining(ctx) == 0 {
					return nil, nil
				}
			}
			if approvals.ContinuationToken == "" {
				break
			}
			token, err := strconv.Atoi(approvals.ContinuationToken)
			if err != nil {
				plugin.Logger(ctx).Error("azuredevopssh_release_approval.listReleaseApprovals", "continuation_token_error", err)
				return nil, err
			}
			input.ContinuationToken = &token
		}
	}
	return nil, nil
}
