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

func tableAzureDevOpsReleaseEnvironment(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_release_environment",
		Description: "Retrieve environments (stages) from Azure DevOps classic releases.",
		List: &plugin.ListConfig{
			ParentHydrate: listProjects,
			Hydrate:       listReleaseEnvironments,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "project_id", Require: plugin.Optional},
				{Name: "release_id", Require: plugin.Optional},
				{Name: "id", Require: plugin.Optional},
				{Name: "status", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The release environment ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "Name of the environment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "Deployment status of the environment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "rank",
				Description: "Order of the environment in the release.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this environment belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "release_id",
				Description: "ID of the release this environment belongs to.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "definition_environment_id",
				Description: "ID of the environment on the release definition.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "trigger_reason",
				Description: "Reason the environment deployment was triggered.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "time_to_deploy",
				Description: "Seconds spent deploying this environment.",
				Type:        proto.ColumnType_DOUBLE,
			},
			{
				Name:        "created_on",
				Description: "Time the environment was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedOn.Time"),
			},
			{
				Name:        "modified_on",
				Description: "Time the environment was last modified.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("ModifiedOn.Time"),
			},
			{
				Name:        "owner",
				Description: "Identity that owns the environment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "variables",
				Description: "Variables scoped to this environment.",
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

// ReleaseEnvironmentRow is one release stage plus the project it was listed from.
type ReleaseEnvironmentRow struct {
	release.ReleaseEnvironment
	ProjectId string
}

func listReleaseEnvironments(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
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
		plugin.Logger(ctx).Error("azuredevops_release_environment.listReleaseEnvironments", "connection_error", err)
		return nil, err
	}
	client, err := release.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_release_environment.listReleaseEnvironments", "client_error", err)
		return nil, err
	}

	expand := release.ReleaseExpandsValues.Environments
	top := 50
	input := release.GetReleasesArgs{
		Project: types.String(projectID),
		Expand:  &expand,
		Top:     &top,
	}
	if releaseID, ok := qualInt(d, "release_id"); ok {
		id := int(releaseID)
		input.ReleaseIdFilter = &[]int{id}
	}

	environmentID, hasEnvironmentID := qualInt(d, "id")
	statusQual := qualString(d, "status")

	for {
		releases, err := client.GetReleases(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevops_release_environment.listReleaseEnvironments", "project_id", projectID, "api_error", err)
			return nil, err
		}
		for _, rel := range releases.Value {
			if rel.Environments == nil {
				continue
			}
			for _, env := range *rel.Environments {
				if env.ReleaseId == nil && rel.Id != nil {
					env.ReleaseId = rel.Id
				}
				if hasEnvironmentID && (env.Id == nil || int64(*env.Id) != environmentID) {
					continue
				}
				if statusQual != "" && (env.Status == nil || !strings.EqualFold(string(*env.Status), statusQual)) {
					continue
				}
				d.StreamListItem(ctx, ReleaseEnvironmentRow{
					ReleaseEnvironment: env,
					ProjectId:          projectID,
				})
				if d.RowsRemaining(ctx) == 0 {
					return nil, nil
				}
			}
		}
		if releases.ContinuationToken == "" {
			break
		}
		token, err := strconv.Atoi(releases.ContinuationToken)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevops_release_environment.listReleaseEnvironments", "continuation_token_error", err)
			return nil, err
		}
		input.ContinuationToken = &token
	}
	return nil, nil
}
