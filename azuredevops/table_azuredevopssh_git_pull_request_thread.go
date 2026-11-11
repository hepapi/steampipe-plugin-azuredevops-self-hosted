package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsGitPullRequestThread(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_git_pull_request_thread",
		Description: "Retrieve information about pull request comment threads.",
		List: &plugin.ListConfig{
			ParentHydrate: listGitPullRequests,
			Hydrate:       listGitPullRequestThreads,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "pull_request_id", Require: plugin.Required},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "pull_request_id",
				Description: "The ID of the pull request the thread belongs to.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "repository_id",
				Description: "The ID of the repository containing the target branch of the pull request.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "The ID of the project this pull request belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "thread_id",
				Description: "The comment thread id.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "status",
				Description: "The status of the comment thread.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_deleted",
				Description: "True if the thread is deleted (all comments deleted).",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "published_date",
				Description: "The time this thread was published.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("PublishedDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "last_updated_date",
				Description: "The time this thread was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("LastUpdatedDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "comments",
				Description: "A list of the comments in this thread.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "identities",
				Description: "Set of identities related to this thread.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "properties",
				Description: "Optional properties associated with the thread as a collection of key-value pairs.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "thread_context",
				Description: "The thread context such as position in left/right file.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "pull_request_thread_context",
				Description: "Extended context information unique to pull requests.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "url",
				Description: "The URL of the thread.",
				Type:        proto.ColumnType_STRING,
			},
		}),
	}
}

// Thread enriches a comment thread with the pull request and repository context
// it belongs to, so the child comment table can call repository-scoped endpoints.
type Thread struct {
	git.GitPullRequestCommentThread
	ProjectId     string
	RepositoryId  string
	PullRequestId int
	ThreadId      int
}

func listGitPullRequestThreads(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	pr, ok := h.Item.(PullRequest)
	if !ok {
		return nil, nil
	}
	if pr.PullRequestId == nil {
		return nil, nil
	}
	// Only fetch threads for the pull request this query is scoped to, so a
	// pull_request_id-qualified query does not fan out an API call for every PR
	// in the organization.
	if pullRequestID := d.EqualsQuals["pull_request_id"].GetInt64Value(); pullRequestID != int64(*pr.PullRequestId) {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_pull_request_thread.listGitPullRequestThreads", "connection_error", err)
		return nil, err
	}
	client, err := git.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_pull_request_thread.listGitPullRequestThreads", "client_error", err)
		return nil, err
	}

	input := git.GetThreadsArgs{
		RepositoryId:  types.String(pr.RepositoryId),
		PullRequestId: pr.PullRequestId,
		Project:       types.String(pr.ProjectId),
	}

	threads, err := client.GetThreads(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_pull_request_thread.listGitPullRequestThreads", "api_error", err)
		return nil, err
	}

	for _, thread := range *threads {
		d.StreamListItem(ctx, Thread{thread, pr.ProjectId, pr.RepositoryId, *pr.PullRequestId, *thread.Id})

		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}
