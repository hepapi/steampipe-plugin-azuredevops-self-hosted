package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func tableAzureDevOpsGitPullRequestReviewer(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_git_pull_request_reviewer",
		Description: "Retrieve reviewers and their votes on pull requests.",
		List: &plugin.ListConfig{
			ParentHydrate: listGitPullRequests,
			Hydrate:       listGitPullRequestReviewers,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "pull_request_id", Require: plugin.Required},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "pull_request_id",
				Description: "The ID of the pull request the reviewer was assigned to.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "repository_id",
				Description: "The ID of the repository containing the target branch of the pull request.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "reviewer_id",
				Description: "The ID of the reviewer identity.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "display_name",
				Description: "The display name of the reviewer.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "unique_name",
				Description: "The unique name of the reviewer.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "vote",
				Description: "Vote on the pull request: 10 - approved, 5 - approved with suggestions, 0 - no vote, -5 - waiting for author, -10 - rejected.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "is_required",
				Description: "True if this is a required reviewer for the pull request.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_flagged",
				Description: "True if this reviewer is flagged for attention on the pull request.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "has_declined",
				Description: "True if this reviewer has declined to review the pull request.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "descriptor",
				Description: "The descriptor that uniquely identifies the reviewer across accounts and organizations.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "directory_alias",
				Description: "The directory alias of the reviewer.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_aad_identity",
				Description: "True if the reviewer is an Azure Active Directory identity.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_container",
				Description: "True if the reviewer is a group or container identity.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_deleted_in_origin",
				Description: "True if the reviewer identity has been deleted at its origin.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "inactive",
				Description: "True if the reviewer identity is inactive.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "reviewer_url",
				Description: "The URL to retrieve information about this identity.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "The URL of the reviewer identity.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "image_url",
				Description: "The URL of the reviewer's avatar image.",
				Type:        proto.ColumnType_STRING,
			},
		}),
	}
}

func listGitPullRequestReviewers(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	pr, ok := h.Item.(PullRequest)
	if !ok {
		return nil, nil
	}
	if pr.PullRequestId == nil {
		return nil, nil
	}
	// Only return reviewers for the pull request this query is scoped to, so a
	// pull_request_id-qualified query does not scan reviewers for every PR in
	// the organization.
	if pullRequestID := d.EqualsQuals["pull_request_id"].GetInt64Value(); pullRequestID != int64(*pr.PullRequestId) {
		return nil, nil
	}
	if pr.Reviewers == nil {
		return nil, nil
	}

	for _, reviewer := range *pr.Reviewers {
		d.StreamListItem(ctx, PullRequestReviewer{reviewer, pr.RepositoryId, *pr.PullRequestId})

		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}

// PullRequestReviewer enriches an IdentityRefWithVote with the pull request
// context it belongs to.
type PullRequestReviewer struct {
	git.IdentityRefWithVote
	RepositoryId  string
	PullRequestId int
}