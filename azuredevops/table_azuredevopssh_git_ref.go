package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsGitRef(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_git_ref",
		Description: "Retrieve refs (branches and tags) for Git repositories.",
		List: &plugin.ListConfig{
			ParentHydrate: listGitRepositories,
			Hydrate:       listGitRefs,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "repository_id", Require: plugin.Optional},
				{Name: "name", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "repository_id",
				Description: "The ID of the repository the ref belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "The ID of the project this repository belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The full ref name, for example refs/heads/main.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "object_id",
				Description: "The object ID the ref currently points to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "peeled_object_id",
				Description: "The object ID the ref peeled to (e.g. the tagged commit for an annotated tag).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_locked",
				Description: "True if the ref is locked.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "creator",
				Description: "The identity of the user who created the ref.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "is_locked_by",
				Description: "The identity that locked the ref.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "statuses",
				Description: "A list of commit statuses for the ref.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "url",
				Description: "The URL of the ref.",
				Type:        proto.ColumnType_STRING,
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

// RepositoryRef enriches a GitRef with its owning repository context.
type RepositoryRef struct {
	git.GitRef
	RepositoryId string
	ProjectId    string
}

func listGitRefs(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
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
		plugin.Logger(ctx).Error("azuredevopssh_git_ref.listGitRefs", "connection_error", err)
		return nil, err
	}
	client, err := git.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_ref.listGitRefs", "client_error", err)
		return nil, err
	}

	projectId := ""
	if repo.Project != nil && repo.Project.Id != nil {
		projectId = repo.Project.Id.String()
	}

	input := git.GetRefsArgs{
		RepositoryId:    types.String(repo.Id.String()),
		Project:         types.String(projectId),
		Filter:          types.String("refs/"),
		IncludeStatuses: types.Bool(true),
		Top:             types.Int(1000),
	}

	// Paginate using the continuation token returned in the response headers.
	for {
		response, err := client.GetRefs(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_git_ref.listGitRefs", "api_error", err)
			return nil, err
		}
		if response == nil || response.Value == nil || len(response.Value) == 0 {
			return nil, nil
		}

		for i := range response.Value {
			d.StreamListItem(ctx, RepositoryRef{response.Value[i], repo.Id.String(), projectId})

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		// Respect a limit if present.
		if d.QueryContext.Limit != nil {
			return nil, nil
		}

		if response.ContinuationToken == "" {
			return nil, nil
		}

		input.ContinuationToken = types.String(response.ContinuationToken)
	}
}