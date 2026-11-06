package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/audit"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

const auditLogBatchSize = 100

func tableAzureDevOpsAuditLog(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_audit_log",
		Description: "Retrieve Azure DevOps audit log entries.",
		List: &plugin.ListConfig{
			Hydrate: listAuditLogs,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "Audit event identifier.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "action_id",
				Description: "Action that was audited, such as Git.CreateRepo.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "area",
				Description: "Area of Azure DevOps where the action occurred.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "category",
				Description: "Action category, such as create, modify, or access.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "category_display_name",
				Description: "Display name of the action category.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "actor_display_name",
				Description: "Display name of the user who initiated the action.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "actor_upn",
				Description: "UPN of the user who initiated the action.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ActorUPN"),
			},
			{
				Name:        "project_id",
				Description: "Project associated with the event, when the event has one.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_name",
				Description: "Name of the project associated with the event.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "ip_address",
				Description: "IP address where the event originated.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "authentication_mechanism",
				Description: "How the actor authenticated.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "details",
				Description: "Decorated details for the event.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "timestamp",
				Description: "When the event occurred.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("Timestamp.Time"),
			},
			{
				Name:        "data",
				Description: "Extra data attached to the event.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ActionId"),
			},
		}),
	}
}

func listAuditLogs(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_audit_log.listAuditLogs", "connection_error", err)
		return nil, err
	}
	client, err := audit.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_audit_log.listAuditLogs", "client_error", err)
		return nil, err
	}

	var token *string
	for {
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
		batch := auditLogBatchSize
		result, err := client.QueryLog(ctx, audit.QueryLogArgs{
			BatchSize:         &batch,
			ContinuationToken: token,
			SkipAggregation:   boolPtr(true),
		})
		if err != nil {
			plugin.Logger(ctx).Error("azuredevops_audit_log.listAuditLogs", "api_error", err)
			return nil, err
		}
		if result == nil || result.DecoratedAuditLogEntries == nil || len(*result.DecoratedAuditLogEntries) == 0 {
			return nil, nil
		}
		for _, item := range *result.DecoratedAuditLogEntries {
			d.StreamListItem(ctx, item)
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		if result.HasMore == nil || !*result.HasMore || result.ContinuationToken == nil || *result.ContinuationToken == "" {
			return nil, nil
		}
		token = result.ContinuationToken
	}
}
