package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsGitPullRequest(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_git_pull_request",
		Description: "Retrieve information about your pull requests.",
		List: &plugin.ListConfig{
			ParentHydrate: listGitRepositories,
			Hydrate:       listGitPullRequests,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "repository_id", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "pull_request_id",
				Description: "The ID of the pull request.",
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
				Name:        "title",
				Description: "The title of the pull request.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the pull request.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "source_ref_name",
				Description: "The name of the source branch of the pull request.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "target_ref_name",
				Description: "The name of the target branch of the pull request.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the pull request (Active, Abandoned, Completed, or NotSet).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "merge_status",
				Description: "The current status of the pull request merge.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_draft",
				Description: "True if the pull request is a draft or work-in-progress pull request.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "has_multiple_merge_bases",
				Description: "True if the pull request has multiple merge bases.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "supports_iterations",
				Description: "True if this pull request supports multiple iterations.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "created_by",
				Description: "The identity of the user who created the pull request.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "creation_date",
				Description: "The date when the pull request was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreationDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "closed_by",
				Description: "The identity of the user who closed the pull request.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "closed_date",
				Description: "The date when the pull request was closed (completed, abandoned, or merged externally).",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("ClosedDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "completion_queue_time",
				Description: "The most recent date at which the pull request entered the queue to be completed.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CompletionQueueTime").Transform(azureDevOpsToTime),
			},
			{
				Name:        "merge_failure_type",
				Description: "The type of failure (if any) of the pull request merge.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "merge_failure_message",
				Description: "If set, the reason the pull request merge failed.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "last_merge_commit",
				Description: "The commit of the most recent pull request merge.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "last_merge_source_commit",
				Description: "The commit at the head of the source branch at the time of the last pull request merge.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "last_merge_target_commit",
				Description: "The commit at the head of the target branch at the time of the last pull request merge.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "commits",
				Description: "The commits contained in the pull request.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "reviewers",
				Description: "A list of reviewers on the pull request along with the state of their votes.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "labels",
				Description: "The labels associated with the pull request.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "work_item_refs",
				Description: "Any work item references associated with this pull request.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "fork_source",
				Description: "Information about the pull request source if it is a fork.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "completion_options",
				Description: "Options which affect how the pull request will be merged when it is completed.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "merge_options",
				Description: "Options used when the pull request merge runs.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "remote_url",
				Description: "The remote URL of the pull request.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "The URL of the pull request.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "artifact_id",
				Description: "A string which uniquely identifies this pull request.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ArtifactId"),
			},
			{
				Name:        "repository",
				Description: "The repository containing the target branch of the pull request.",
				Type:        proto.ColumnType_JSON,
			},
		}),
	}
}

// PullRequest enriches a git.GitPullRequest with the repository and project id
// it belongs to. This is required for child tables (threads, comments,
// reviewers) which must call repository-scoped git endpoints.
type PullRequest struct {
	git.GitPullRequest
	ProjectId    string
	RepositoryId string
}

func listGitPullRequests(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	repo, ok := h.Item.(git.GitRepository)
	if !ok {
		return nil, nil
	}
	repository_id := d.EqualsQuals["repository_id"].GetStringValue()

	// Check if the provided repository_id is not matching with the parentHydrate
	if repository_id != "" && repository_id != repo.Id.String() {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_pull_request.listGitPullRequests", "connection_error", err)
		return nil, err
	}
	client, err := git.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_pull_request.listGitPullRequests", "client_error", err)
		return nil, err
	}

	input := git.GetPullRequestsArgs{
		RepositoryId: types.String(repo.Id.String()),
		SearchCriteria: &git.GitPullRequestSearchCriteria{
			IncludeLinks: types.Bool(true),
		},
	}
	// Respect a limit if present.
	if d.QueryContext.Limit != nil {
		input.Top = types.Int(int(*d.QueryContext.Limit))
	}

	pullRequests, err := client.GetPullRequests(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_pull_request.listGitPullRequests", "api_error", err)
		return nil, err
	}

	projectId := ""
	if repo.Project != nil && repo.Project.Id != nil {
		projectId = repo.Project.Id.String()
	}

	for _, pr := range *pullRequests {
		d.StreamListItem(ctx, PullRequest{pr, projectId, repo.Id.String()})

		// Context can be cancelled due to manual cancellation or the limit has been hit
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}