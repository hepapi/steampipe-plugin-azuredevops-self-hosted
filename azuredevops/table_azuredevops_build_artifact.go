package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsBuildArtifact(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_build_artifact",
		Description: "Retrieve artifacts published by a single Azure DevOps build.",
		List: &plugin.ListConfig{
			Hydrate: listBuildArtifacts,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Required},
				{Name: "build_id", Require: plugin.Required},
				{Name: "name", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The artifact ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "name",
				Description: "Name of the artifact.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "build_id",
				Description: "ID of the build that published this artifact.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "ID of the project the build belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "source",
				Description: "ID of the job that produced this artifact.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "resource",
				Description: "Download location and type of the artifact.",
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

// BuildArtifactRow is one published artifact plus the build it was read from.
type BuildArtifactRow struct {
	build.BuildArtifact
	BuildId   int
	ProjectId string
}

func listBuildArtifacts(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	buildID, ok := qualInt(d, "build_id")
	if projectID == "" || !ok {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_build_artifact.listBuildArtifacts", "connection_error", err)
		return nil, err
	}
	client, err := build.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_build_artifact.listBuildArtifacts", "client_error", err)
		return nil, err
	}

	id := int(buildID)
	artifacts, err := client.GetArtifacts(ctx, build.GetArtifactsArgs{
		Project: types.String(projectID),
		BuildId: &id,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_build_artifact.listBuildArtifacts", "project_id", projectID, "build_id", id, "api_error", err)
		return nil, err
	}
	if artifacts == nil {
		return nil, nil
	}

	nameQual := qualString(d, "name")
	for _, artifact := range *artifacts {
		if nameQual != "" && (artifact.Name == nil || *artifact.Name != nameQual) {
			continue
		}
		d.StreamListItem(ctx, BuildArtifactRow{
			BuildArtifact: artifact,
			BuildId:       id,
			ProjectId:     projectID,
		})
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}
