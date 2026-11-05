package azuredevops

import (
	"context"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsBuildTimeline(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_build_timeline",
		Description: "Retrieve timeline records for a single Azure DevOps build.",
		List: &plugin.ListConfig{
			Hydrate: listBuildTimeline,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Required},
				{Name: "build_id", Require: plugin.Required},
				{Name: "id", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The timeline record ID.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "build_id",
				Description: "ID of the build this record belongs to.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "ID of the project the build belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "timeline_id",
				Description: "ID of the timeline that contains this record.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "Name of the timeline record.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "Type of the record, such as Stage, Phase, Job, or Task.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "state",
				Description: "State of the record (pending, inProgress, or completed).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "result",
				Description: "Result of the record.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "parent_id",
				Description: "ID of the parent timeline record.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "order",
				Description: "Ordinal of this record relative to its siblings.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "identifier",
				Description: "Identifier that stays the same across attempts.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "attempt",
				Description: "Attempt number of the record.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "error_count",
				Description: "Number of errors produced by this operation.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "warning_count",
				Description: "Number of warnings produced by this operation.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "percent_complete",
				Description: "Completion percentage of the record.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "worker_name",
				Description: "Name of the agent that ran the operation.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "start_time",
				Description: "Time the record started.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("StartTime.Time"),
			},
			{
				Name:        "finish_time",
				Description: "Time the record finished.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("FinishTime.Time"),
			},
			{
				Name:        "log",
				Description: "Reference to the log produced by this operation.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "task",
				Description: "Task represented by this timeline record.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "issues",
				Description: "Issues reported by this record.",
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

// BuildTimelineRecord is one timeline step plus the build it was read from.
type BuildTimelineRecord struct {
	build.TimelineRecord
	BuildId    int
	ProjectId  string
	TimelineId string
}

func listBuildTimeline(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	buildID, ok := qualInt(d, "build_id")
	if projectID == "" || !ok {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_build_timeline.listBuildTimeline", "connection_error", err)
		return nil, err
	}
	client, err := build.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_build_timeline.listBuildTimeline", "client_error", err)
		return nil, err
	}

	id := int(buildID)
	timeline, err := client.GetBuildTimeline(ctx, build.GetBuildTimelineArgs{
		Project: types.String(projectID),
		BuildId: &id,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_build_timeline.listBuildTimeline", "project_id", projectID, "build_id", id, "api_error", err)
		return nil, err
	}
	if timeline == nil || timeline.Records == nil {
		return nil, nil
	}

	timelineID := ""
	if timeline.Id != nil {
		timelineID = timeline.Id.String()
	}
	recordID := qualString(d, "id")
	for _, record := range *timeline.Records {
		if recordID != "" && (record.Id == nil || !strings.EqualFold(record.Id.String(), recordID)) {
			continue
		}
		d.StreamListItem(ctx, BuildTimelineRecord{
			TimelineRecord: record,
			BuildId:        id,
			ProjectId:      projectID,
			TimelineId:     timelineID,
		})
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}
