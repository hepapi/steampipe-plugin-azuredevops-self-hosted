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

func tableAzureDevOpsReleaseDefinition(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_release_definition",
		Description: "Retrieve Azure DevOps classic release definitions.",
		List: &plugin.ListConfig{
			ParentHydrate: listProjects,
			Hydrate:       listReleaseDefinitions,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "project_id", Require: plugin.Optional},
				{Name: "id", Require: plugin.Optional},
				{Name: "name", Require: plugin.Optional},
				{Name: "path", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "project_id"}),
			Hydrate:    getReleaseDefinition,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The release definition ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "Name of the release definition.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "path",
				Description: "Folder path of the release definition.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this definition belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "Description of the release definition.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "revision",
				Description: "Revision number of the definition.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "is_deleted",
				Description: "True when the definition has been deleted.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "url",
				Description: "REST URL of the release definition.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "created_on",
				Description: "Time the definition was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedOn.Time"),
			},
			{
				Name:        "modified_on",
				Description: "Time the definition was last modified.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("ModifiedOn.Time"),
			},
			{
				Name:        "created_by",
				Description: "Identity that created the definition.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "modified_by",
				Description: "Identity that last modified the definition.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "artifacts",
				Description: "Artifacts linked to the definition.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "environments",
				Description: "Environments (stages) configured on the definition.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "tags",
				Description: "Tags on the definition.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "variables",
				Description: "Variables defined on the release definition.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "triggers",
				Description: "Triggers configured on the definition.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
		}),
	}
}

// ReleaseDefinitionRow is a release definition plus the project it was listed from.
// The list API often omits projectReference, so the project id is carried here.
type ReleaseDefinitionRow struct {
	release.ReleaseDefinition
	ProjectId string
}

func listReleaseDefinitions(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
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
		plugin.Logger(ctx).Error("azuredevopssh_release_definition.listReleaseDefinitions", "connection_error", err)
		return nil, err
	}
	client, err := release.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_release_definition.listReleaseDefinitions", "client_error", err)
		return nil, err
	}

	expand := release.ReleaseDefinitionExpandsValues.Environments
	top := 50
	if d.QueryContext.Limit != nil && *d.QueryContext.Limit > 0 && int(*d.QueryContext.Limit) < top {
		top = int(*d.QueryContext.Limit)
	}
	input := release.GetReleaseDefinitionsArgs{
		Project: types.String(projectID),
		Expand:  &expand,
		Top:     &top,
	}
	if id, ok := qualInt(d, "id"); ok {
		input.DefinitionIdFilter = &[]string{strconv.FormatInt(id, 10)}
	}
	if name := qualString(d, "name"); name != "" {
		input.SearchText = &name
		exact := true
		input.IsExactNameMatch = &exact
	}
	if path := qualString(d, "path"); path != "" {
		input.Path = &path
	}

	for {
		definitions, err := client.GetReleaseDefinitions(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_release_definition.listReleaseDefinitions", "project_id", projectID, "api_error", err)
			return nil, err
		}
		for _, definition := range definitions.Value {
			d.StreamListItem(ctx, ReleaseDefinitionRow{
				ReleaseDefinition: definition,
				ProjectId:         projectID,
			})
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		if definitions.ContinuationToken == "" {
			break
		}
		token := definitions.ContinuationToken
		input.ContinuationToken = &token
	}
	return nil, nil
}

func getReleaseDefinition(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	id, ok := qualInt(d, "id")
	if projectID == "" || !ok {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_release_definition.getReleaseDefinition", "connection_error", err)
		return nil, err
	}
	client, err := release.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_release_definition.getReleaseDefinition", "client_error", err)
		return nil, err
	}

	definitionID := int(id)
	definition, err := client.GetReleaseDefinition(ctx, release.GetReleaseDefinitionArgs{
		Project:      types.String(projectID),
		DefinitionId: &definitionID,
	})
	if err != nil {
		if absentFromCollection(d, err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_release_definition.getReleaseDefinition", "project_id", projectID, "id", definitionID, "api_error", err)
		return nil, err
	}
	if definition == nil {
		return nil, nil
	}
	return &ReleaseDefinitionRow{
		ReleaseDefinition: *definition,
		ProjectId:         projectID,
	}, nil
}
