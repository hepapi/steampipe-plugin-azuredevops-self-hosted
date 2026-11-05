package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsGitCommit(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_git_commit",
		Description: "Retrieve commit information for Git repositories.",
		List: &plugin.ListConfig{
			ParentHydrate: listGitRepositories,
			Hydrate:       listGitCommits,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "repository_id", Require: plugin.Required},
				{Name: "commit_id", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "repository_id",
				Description: "The ID of the repository the commit belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "The ID of the project this repository belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "commit_id",
				Description: "ID (SHA-1) of the commit.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "comment",
				Description: "Comment or message of the commit.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "comment_truncated",
				Description: "True if the comment is truncated from the full git commit comment message.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "commit_too_many_changes",
				Description: "True if the commit contains too many changes to be displayed.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "author_name",
				Description: "The name of the commit author.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Author.Name"),
			},
			{
				Name:        "author_email",
				Description: "The email of the commit author.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Author.Email"),
			},
			{
				Name:        "author_date",
				Description: "The date the commit was authored.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Author.Date").Transform(azureDevOpsToTime),
			},
			{
				Name:        "committer",
				Description: "The committer of the commit.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "parents",
				Description: "An enumeration of the parent commit IDs for this commit.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "push",
				Description: "The push associated with this commit.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "remote_url",
				Description: "Remote URL path to the commit.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "REST URL for this resource.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "statuses",
				Description: "A list of status metadata from services and extensions that may associate additional information to the commit.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "work_items",
				Description: "A list of work items associated with this commit.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "change_counts",
				Description: "Counts of the types of changes (edits, deletes, etc.) included with the commit.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "links",
				Description: "A collection of related REST reference links.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Links"),
			},
		}),
	}
}

// RepositoryCommit enriches a GitCommitRef with its owning repository context.
type RepositoryCommit struct {
	git.GitCommitRef
	RepositoryId string
	ProjectId    string
}

func listGitCommits(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	repo, ok := h.Item.(git.GitRepository)
	if !ok {
		return nil, nil
	}
	repository_id := d.EqualsQuals["repository_id"].GetStringValue()

	// Check if the provided repository_id is not matching with the parentHydrate
	if repository_id != "" && repository_id != repo.Id.String() {
		return nil, nil
	}

	// A specific commit can be requested directly by id.
	commit_id := d.EqualsQuals["commit_id"].GetStringValue()
	criteria := &git.GitQueryCommitsCriteria{
		IncludePushData:  types.Bool(true),
		IncludeWorkItems: types.Bool(true),
	}
	if commit_id != "" {
		criteria.Ids = &[]string{commit_id}
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_git_commit.listGitCommits", "connection_error", err)
		return nil, err
	}
	client, err := git.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_git_commit.listGitCommits", "client_error", err)
		return nil, err
	}

	input := git.GetCommitsArgs{
		RepositoryId:   types.String(repo.Id.String()),
		SearchCriteria: criteria,
		Skip:           types.Int(0),
		Top:            types.Int(100),
	}
	// Respect a limit if present.
	if d.QueryContext.Limit != nil {
		input.Top = types.Int(int(*d.QueryContext.Limit))
	}

	projectId := ""
	if repo.Project != nil && repo.Project.Id != nil {
		projectId = repo.Project.Id.String()
	}

	for {
		commits, err := client.GetCommits(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevops_git_commit.listGitCommits", "api_error", err)
			return nil, err
		}
		if commits == nil || len(*commits) == 0 {
			return nil, nil
		}

		for _, commit := range *commits {
			d.StreamListItem(ctx, RepositoryCommit{commit, repo.Id.String(), projectId})

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		// Respect a limit if present.
		if d.QueryContext.Limit != nil {
			return nil, nil
		}

		if len(*commits) < *input.Top {
			return nil, nil
		}

		input.Skip = types.Int(*input.Skip + *input.Top)
	}
}