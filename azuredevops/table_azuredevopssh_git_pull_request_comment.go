package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsGitPullRequestComment(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_git_pull_request_comment",
		Description: "Retrieve comments made on pull request comment threads.",
		List: &plugin.ListConfig{
			ParentHydrate: listGitPullRequestThreads,
			Hydrate:       listGitPullRequestComments,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "pull_request_id", Require: plugin.Required},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "pull_request_id",
				Description: "The ID of the pull request the comment belongs to.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "repository_id",
				Description: "The ID of the repository containing the target branch of the pull request.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "thread_id",
				Description: "The ID of the thread this comment belongs to.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "comment_id",
				Description: "The comment ID. IDs start at 1 and are unique to a pull request.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("Id"),
			},
			{
				Name:        "content",
				Description: "The comment content.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "comment_type",
				Description: "The comment type at the time of creation.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_deleted",
				Description: "Whether or not this comment was soft-deleted.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "parent_comment_id",
				Description: "The ID of the parent comment. This is used for replies.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "author",
				Description: "The author of the comment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "published_date",
				Description: "The date the comment was first published.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("PublishedDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "last_content_updated_date",
				Description: "The date the comment's content was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("LastContentUpdatedDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "last_updated_date",
				Description: "The date the comment was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("LastUpdatedDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "users_liked",
				Description: "A list of the users who have liked this comment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "links",
				Description: "Links to other related objects.",
				Type:        proto.ColumnType_JSON,
			},
		}),
	}
}

func listGitPullRequestComments(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	thread, ok := h.Item.(Thread)
	if !ok {
		return nil, nil
	}
	// Only fetch comments on threads belonging to the pull request this query
	// is scoped to, so a pull_request_id-qualified query does not fan out an API
	// call for every thread in the organization.
	if pullRequestID := d.EqualsQuals["pull_request_id"].GetInt64Value(); pullRequestID != int64(thread.PullRequestId) {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_pull_request_comment.listGitPullRequestComments", "connection_error", err)
		return nil, err
	}
	client, err := git.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_pull_request_comment.listGitPullRequestComments", "client_error", err)
		return nil, err
	}

	input := git.GetCommentsArgs{
		RepositoryId:  types.String(thread.RepositoryId),
		PullRequestId: types.Int(thread.PullRequestId),
		ThreadId:      types.Int(thread.ThreadId),
		Project:       types.String(thread.ProjectId),
	}

	comments, err := client.GetComments(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_pull_request_comment.listGitPullRequestComments", "api_error", err)
		return nil, err
	}

	for _, comment := range *comments {
		d.StreamListItem(ctx, PullRequestComment{comment, thread.RepositoryId, thread.PullRequestId, thread.ThreadId})

		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}

// PullRequestComment enriches a git.Comment with the pull request context it
// belongs to.
type PullRequestComment struct {
	git.Comment
	RepositoryId  string
	PullRequestId int
	ThreadId      int
}