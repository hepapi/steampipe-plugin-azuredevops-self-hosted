package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsGitTag(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_git_tag",
		Description: "Retrieve tags for Git repositories.",
		List: &plugin.ListConfig{
			ParentHydrate: listGitRepositories,
			Hydrate:       listGitTags,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "repository_id", Require: plugin.Optional},
				{Name: "name", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "repository_id",
				Description: "The ID of the repository the tag belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "The ID of the project this repository belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The full ref name of the tag, for example refs/tags/v1.0.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "object_id",
				Description: "The object ID (SHA-1) the tag ref currently points to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "peeled_object_id",
				Description: "The object ID the annotated tag peels to (the tagged commit).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "creator",
				Description: "The identity of the user who created the tag.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "message",
				Description: "The tag message for annotated tags.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getTagMessage,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "is_locked",
				Description: "True if the tag is locked.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_locked_by",
				Description: "The identity that locked the tag.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "statuses",
				Description: "A list of commit statuses for the tag.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "url",
				Description: "The URL of the tag.",
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

// RepositoryTag enriches a GitRef (of type tag) with owning repository context.
type RepositoryTag struct {
	git.GitRef
	RepositoryId string
	ProjectId    string
}

func listGitTags(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
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
		plugin.Logger(ctx).Error("azuredevops_git_tag.listGitTags", "connection_error", err)
		return nil, err
	}
	client, err := git.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_git_tag.listGitTags", "client_error", err)
		return nil, err
	}

	projectId := ""
	if repo.Project != nil && repo.Project.Id != nil {
		projectId = repo.Project.Id.String()
	}

	input := git.GetRefsArgs{
		RepositoryId:    types.String(repo.Id.String()),
		Project:         types.String(projectId),
		Filter:          types.String("refs/tags/"),
		IncludeStatuses: types.Bool(true),
		PeelTags:        types.Bool(true),
		Top:             types.Int(1000),
	}

	// Paginate using the continuation token returned in the response headers.
	for {
		response, err := client.GetRefs(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevops_git_tag.listGitTags", "api_error", err)
			return nil, err
		}
		if response == nil || response.Value == nil || len(response.Value) == 0 {
			return nil, nil
		}

		for i := range response.Value {
			d.StreamListItem(ctx, RepositoryTag{response.Value[i], repo.Id.String(), projectId})

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

// getTagMessage resolves the message for an annotated tag by fetching the
// annotated tag object. Lightweight tags have no annotated message, so the
// lookup error (including the "not found" response for lightweight tags) is
// swallowed and the column is left null.
func getTagMessage(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	tag, ok := h.Item.(RepositoryTag)
	if !ok || tag.ObjectId == nil {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		return nil, err
	}
	client, err := git.NewClient(ctx, connection)
	if err != nil {
		return nil, err
	}

	annotatedTag, err := client.GetAnnotatedTag(ctx, git.GetAnnotatedTagArgs{
		Project:      types.String(tag.ProjectId),
		RepositoryId: types.String(tag.RepositoryId),
		ObjectId:     types.String(*tag.ObjectId),
	})
	if err != nil {
		// Lightweight tags don't have an associated annotated tag object.
		plugin.Logger(ctx).Debug("azuredevops_git_tag.getTagMessage", "get_annotated_tag_error", err.Error())
		return nil, nil
	}
	if annotatedTag == nil {
		return nil, nil
	}
	return annotatedTag.Message, nil
}