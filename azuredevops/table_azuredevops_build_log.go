package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsBuildLog(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_build_log",
		Description: "Retrieve log metadata for a single Azure DevOps build. Log line content is not downloaded.",
		List: &plugin.ListConfig{
			Hydrate: listBuildLogs,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Required},
				{Name: "build_id", Require: plugin.Required},
				{Name: "id", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The log ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "build_id",
				Description: "ID of the build this log belongs to.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "ID of the project the build belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "Type of the log location.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "URL of the log resource.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "line_count",
				Description: "Number of lines in the log.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("LineCount").Transform(uint64ToInt),
			},
			{
				Name:        "created_on",
				Description: "Time the log was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedOn.Time"),
			},
			{
				Name:        "last_changed_on",
				Description: "Time the log was last changed.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("LastChangedOn.Time"),
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Url"),
			},
		}),
	}
}

// BuildLogRow is log metadata plus the build it was read from.
type BuildLogRow struct {
	build.BuildLog
	BuildId   int
	ProjectId string
}

func listBuildLogs(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	buildID, ok := qualInt(d, "build_id")
	if projectID == "" || !ok {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_build_log.listBuildLogs", "connection_error", err)
		return nil, err
	}
	client, err := build.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_build_log.listBuildLogs", "client_error", err)
		return nil, err
	}

	id := int(buildID)
	logs, err := client.GetBuildLogs(ctx, build.GetBuildLogsArgs{
		Project: types.String(projectID),
		BuildId: &id,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_build_log.listBuildLogs", "project_id", projectID, "build_id", id, "api_error", err)
		return nil, err
	}
	if logs == nil {
		return nil, nil
	}

	logID, hasLogID := qualInt(d, "id")
	for _, log := range *logs {
		if hasLogID && (log.Id == nil || int64(*log.Id) != logID) {
			continue
		}
		d.StreamListItem(ctx, BuildLogRow{
			BuildLog:  log,
			BuildId:   id,
			ProjectId: projectID,
		})
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func uint64ToInt(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return nil, nil
	}
	switch v := d.Value.(type) {
	case uint64:
		return int64(v), nil
	case *uint64:
		if v == nil {
			return nil, nil
		}
		return int64(*v), nil
	default:
		return d.Value, nil
	}
}
