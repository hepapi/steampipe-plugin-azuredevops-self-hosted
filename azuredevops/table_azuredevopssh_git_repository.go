package azuredevops

import (
	"context"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsGitRepository(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_git_repository",
		Description: "Retrieve information about your repositories.",
		List: &plugin.ListConfig{
			Hydrate: listGitRepositories,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "project_id", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getRepository,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The repository id.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The repository name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "default_branch",
				Description: "The repository default branch.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "default_branch_object",
				Description: "Details of the repository default branch, including its current commit and ahead/behind counts.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getRepositoryDefaultBranch,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "is_disabled",
				Description: "True if the repository is disabled.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_fork",
				Description: "True if the repository was created as a fork.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "project_id",
				Description: "The project Id.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Project.Id"),
			},
			{
				Name:        "project_name",
				Description: "The name of the project this repository belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Project.Name"),
			},
			{
				Name:        "remote_url",
				Description: "The repository remote url.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "size",
				Description: "Compressed size (bytes) of the repository.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "ssh_url",
				Description: "The repository ssh url.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "The repository url.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "web_url",
				Description: "The repository web url.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "links",
				Description: "The class to represent a collection of REST reference links.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "parent_repository",
				Description: "The parent repository.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "project",
				Description: "The project this repository belongs to.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "valid_remote_urls",
				Description: "The repository valid remote urls.",
				Type:        proto.ColumnType_JSON,
			},

			/// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
		}),
	}
}

func listGitRepositories(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository.listGitRepositories", "connection_error", err)
		return nil, err
	}
	client, err := git.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository.listGitRepositories", "client_error", err)
		return nil, err
	}

	input := git.GetRepositoriesArgs{
		IncludeLinks:   types.Bool(true),
		IncludeAllUrls: types.Bool(true),
	}
	if d.EqualsQuals["project_id"] != nil {
		input.Project = types.String(d.EqualsQuals["project_id"].GetStringValue())
	}

	repositories, err := client.GetRepositories(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository.listGitRepositories", "api_error", err)
		return nil, err
	}

	for _, repository := range *repositories {
		d.StreamListItem(ctx, repository)

		// Context can be cancelled due to manual cancellation or the limit has been hit
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}

func getRepository(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	repositoryId := d.EqualsQuals["id"].GetStringValue()

	// Check if repositoryId is empty
	if repositoryId == "" {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository.getRepository", "connection_error", err)
		return nil, err
	}
	client, err := git.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository.getRepository", "client_error", err)
		return nil, err
	}

	input := git.GetRepositoryArgs{
		RepositoryId: types.String(repositoryId),
	}

	repo, err := client.GetRepository(ctx, input)
	if err != nil {
		if absentFromCollection(d, err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_git_repository.getRepository", "api_error", err)
		return nil, err
	}

	return repo, nil
}

// getRepositoryDefaultBranch returns the branch stats object for the repository
// default branch. This is a separate per-row hydrate so that the expensive
// branch lookup only runs when the column is requested. Returns nil (null
// column) when the repository has no default branch or the default branch
// cannot be resolved (e.g. an empty/uninitialized repository).
func getRepositoryDefaultBranch(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	repo, ok := h.Item.(git.GitRepository)
	if !ok {
		return nil, nil
	}
	if repo.DefaultBranch == nil || *repo.DefaultBranch == "" {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository.getRepositoryDefaultBranch", "connection_error", err)
		return nil, err
	}
	client, err := git.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository.getRepositoryDefaultBranch", "client_error", err)
		return nil, err
	}

	input := git.GetBranchArgs{
		RepositoryId: types.String(repo.Id.String()),
		Name:         types.String(*repo.DefaultBranch),
	}

	branch, err := client.GetBranch(ctx, input)
	if err != nil {
		// A repository with no branches (e.g. an empty/uninitialized repo)
		// returns VS403403 "Cannot find any branches". Leave the column null.
		if isNoBranchesError(err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_git_repository.getRepositoryDefaultBranch", "api_error", err)
		return nil, err
	}

	return branch, nil
}

// isNoBranchesError returns true when the given error indicates that a
// repository has no branches to resolve (e.g. an empty or uninitialized
// repository), in which case branch lookups should be treated as absent
// rather than as a hard failure.
func isNoBranchesError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "vs403403") ||
		strings.Contains(msg, "cannot find any branches") ||
		strings.Contains(msg, "no branches exist")
}
