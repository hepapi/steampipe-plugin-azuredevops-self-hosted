package azuredevops

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// workItemPageSize is the maximum number of ids accepted by GetWorkItemsBatch.
const workItemPageSize = 200

func tableAzureDevOpsWorkItem(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_work_item",
		Description: "Retrieve work items from Azure DevOps projects.",
		List: &plugin.ListConfig{
			Hydrate: listWorkItems,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Optional},
				{Name: "id", Require: plugin.Optional},
				{Name: "work_item_type", Require: plugin.Optional},
				{Name: "state", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "project_id"}),
			Hydrate:    getWorkItem,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The work item ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "rev",
				Description: "Revision number of the work item.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this work item belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_name",
				Description: "Name of the project this work item belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(workItemStringField("System.TeamProject")),
			},
			{
				Name:        "work_item_type",
				Description: "Type of the work item, such as Bug, Task, or User Story.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(workItemStringField("System.WorkItemType")),
			},
			{
				Name:        "state",
				Description: "Current state of the work item.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(workItemStringField("System.State")),
			},
			{
				Name:        "reason",
				Description: "Reason the work item is in its current state.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(workItemStringField("System.Reason")),
			},
			{
				Name:        "area_path",
				Description: "Area path of the work item.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(workItemStringField("System.AreaPath")),
			},
			{
				Name:        "iteration_path",
				Description: "Iteration path of the work item.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(workItemStringField("System.IterationPath")),
			},
			{
				Name:        "tags",
				Description: "Semicolon-separated tags on the work item.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(workItemStringField("System.Tags")),
			},
			{
				Name:        "priority",
				Description: "Priority of the work item, when the process defines Microsoft.VSTS.Common.Priority.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.From(workItemIntField("Microsoft.VSTS.Common.Priority")),
			},
			{
				Name:        "severity",
				Description: "Severity of the work item, when the process defines Microsoft.VSTS.Common.Severity.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(workItemStringField("Microsoft.VSTS.Common.Severity")),
			},
			{
				Name:        "parent_id",
				Description: "ID of the parent work item, when System.Parent is set.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.From(workItemIntField("System.Parent")),
			},
			{
				Name:        "assigned_to_name",
				Description: "Display name of the identity assigned to the work item.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(workItemIdentityName("System.AssignedTo")),
			},
			{
				Name:        "assigned_to",
				Description: "Identity assigned to the work item.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.From(workItemRawField("System.AssignedTo")),
			},
			{
				Name:        "created_by",
				Description: "Identity that created the work item.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.From(workItemRawField("System.CreatedBy")),
			},
			{
				Name:        "changed_by",
				Description: "Identity that last changed the work item.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.From(workItemRawField("System.ChangedBy")),
			},
			{
				Name:        "created_date",
				Description: "Time the work item was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.From(workItemTimeField("System.CreatedDate")),
			},
			{
				Name:        "changed_date",
				Description: "Time the work item was last changed.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.From(workItemTimeField("System.ChangedDate")),
			},
			{
				Name:        "description",
				Description: "Description of the work item, stored as HTML.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(workItemStringField("System.Description")),
			},
			{
				Name:        "url",
				Description: "REST URL of the work item.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "fields",
				Description: "Full field map for the work item, including process-specific fields.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.From(workItemFieldsColumn),
			},
			{
				Name:        "links",
				Description: "REST reference links for the work item.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Links"),
			},
			{
				Name:        "title",
				Description: "Title of the work item.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.From(workItemStringField("System.Title")),
			},
		}),
	}
}

// ProjectWorkItem is one work item plus the project it was read from.
type ProjectWorkItem struct {
	workitemtracking.WorkItem
	ProjectId string
}

func listWorkItems(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_work_item.listWorkItems", "connection_error", err)
		return nil, err
	}
	client, err := workitemtracking.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_work_item.listWorkItems", "client_error", err)
		return nil, err
	}

	// A project_id qual reads that project even when it is outside the first page.
	// The SQL limit applies to work item rows, not to how many projects are visited.
	if projectID := qualString(d, "project_id"); projectID != "" {
		return nil, listProjectWorkItems(ctx, d, client, projectID)
	}

	coreClient, err := core.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_work_item.listWorkItems", "client_error", err)
		return nil, err
	}
	input := core.GetProjectsArgs{Top: types.Int(100)}
	for {
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
		response, err := coreClient.GetProjects(ctx, input)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_work_item.listWorkItems", "api_error", err)
			return nil, err
		}
		for _, project := range response.Value {
			if project.Id == nil {
				continue
			}
			if err := listProjectWorkItems(ctx, d, client, project.Id.String()); err != nil {
				return nil, err
			}
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		if response.ContinuationToken == "" {
			return nil, nil
		}
		continuationToken, err := strconv.Atoi(response.ContinuationToken)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_work_item.listWorkItems", "continuation_token_error", err)
			return nil, err
		}
		input.ContinuationToken = types.Int(continuationToken)
	}
}

func listProjectWorkItems(ctx context.Context, d *plugin.QueryData, client workitemtracking.Client, projectID string) error {
	var idFilter *int
	if id, ok := qualInt(d, "id"); ok {
		value := int(id)
		idFilter = &value
	}

	lastID := 0
	for {
		top := workItemPageSize
		if remaining := int(d.RowsRemaining(ctx)); remaining < top {
			top = remaining
		}
		if top <= 0 {
			return nil
		}

		query := workItemListWiql(lastID, idFilter, qualString(d, "work_item_type"), qualString(d, "state"))
		result, err := client.QueryByWiql(ctx, workitemtracking.QueryByWiqlArgs{
			Project: types.String(projectID),
			Wiql:    &workitemtracking.Wiql{Query: types.String(query)},
			Top:     types.Int(top),
		})
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_work_item.listWorkItems", "project_id", projectID, "api_error", err)
			return err
		}
		if result == nil || result.WorkItems == nil || len(*result.WorkItems) == 0 {
			return nil
		}

		ids := make([]int, 0, len(*result.WorkItems))
		maxID := lastID
		for _, ref := range *result.WorkItems {
			if ref.Id == nil {
				continue
			}
			ids = append(ids, *ref.Id)
			if *ref.Id > maxID {
				maxID = *ref.Id
			}
		}
		if len(ids) == 0 || maxID == lastID {
			return nil
		}

		if err := streamWorkItems(ctx, d, client, projectID, ids); err != nil {
			return err
		}
		if d.RowsRemaining(ctx) == 0 || idFilter != nil || len(ids) < top {
			return nil
		}
		lastID = maxID
	}
}

func streamWorkItems(ctx context.Context, d *plugin.QueryData, client workitemtracking.Client, projectID string, ids []int) error {
	expand := workitemtracking.WorkItemExpandValues.Links
	policy := workitemtracking.WorkItemErrorPolicyValues.Omit
	for start := 0; start < len(ids); start += workItemPageSize {
		end := start + workItemPageSize
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[start:end]
		items, err := client.GetWorkItemsBatch(ctx, workitemtracking.GetWorkItemsBatchArgs{
			Project: types.String(projectID),
			WorkItemGetRequest: &workitemtracking.WorkItemBatchGetRequest{
				Ids:         &chunk,
				Expand:      &expand,
				ErrorPolicy: &policy,
			},
		})
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_work_item.listWorkItems", "project_id", projectID, "api_error", err)
			return err
		}
		if items == nil {
			continue
		}
		for _, item := range *items {
			d.StreamListItem(ctx, ProjectWorkItem{WorkItem: item, ProjectId: projectID})
			if d.RowsRemaining(ctx) == 0 {
				return nil
			}
		}
	}
	return nil
}

func getWorkItem(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	id, ok := qualInt(d, "id")
	if projectID == "" || !ok {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_work_item.getWorkItem", "connection_error", err)
		return nil, err
	}
	client, err := workitemtracking.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_work_item.getWorkItem", "client_error", err)
		return nil, err
	}

	workItemID := int(id)
	expand := workitemtracking.WorkItemExpandValues.Links
	item, err := client.GetWorkItem(ctx, workitemtracking.GetWorkItemArgs{
		Id:      &workItemID,
		Project: types.String(projectID),
		Expand:  &expand,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_work_item.getWorkItem", "project_id", projectID, "id", workItemID, "api_error", err)
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return &ProjectWorkItem{WorkItem: *item, ProjectId: projectID}, nil
}

func workItemListWiql(afterID int, id *int, workItemType, state string) string {
	var query strings.Builder
	query.WriteString("SELECT [System.Id] FROM WorkItems WHERE [System.TeamProject] = @project")
	if id != nil {
		fmt.Fprintf(&query, " AND [System.Id] = %d", *id)
	} else {
		fmt.Fprintf(&query, " AND [System.Id] > %d", afterID)
	}
	if workItemType != "" {
		query.WriteString(" AND [System.WorkItemType] = ")
		query.WriteString(wiqlString(workItemType))
	}
	if state != "" {
		query.WriteString(" AND [System.State] = ")
		query.WriteString(wiqlString(state))
	}
	query.WriteString(" ORDER BY [System.Id]")
	return query.String()
}

func wiqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func projectWorkItem(item interface{}) (ProjectWorkItem, bool) {
	switch value := item.(type) {
	case ProjectWorkItem:
		return value, true
	case *ProjectWorkItem:
		if value == nil {
			return ProjectWorkItem{}, false
		}
		return *value, true
	default:
		return ProjectWorkItem{}, false
	}
}

func workItemField(item interface{}, referenceName string) interface{} {
	workItem, ok := projectWorkItem(item)
	if !ok || workItem.Fields == nil {
		return nil
	}
	return (*workItem.Fields)[referenceName]
}

func workItemFieldsColumn(_ context.Context, d *transform.TransformData) (interface{}, error) {
	workItem, ok := projectWorkItem(d.HydrateItem)
	if !ok || workItem.Fields == nil {
		return nil, nil
	}
	return *workItem.Fields, nil
}

func workItemRawField(referenceName string) transform.TransformFunc {
	return func(_ context.Context, d *transform.TransformData) (interface{}, error) {
		return workItemField(d.HydrateItem, referenceName), nil
	}
}

func workItemStringField(referenceName string) transform.TransformFunc {
	return func(_ context.Context, d *transform.TransformData) (interface{}, error) {
		value, ok := workItemField(d.HydrateItem, referenceName).(string)
		if !ok || value == "" {
			return nil, nil
		}
		return value, nil
	}
}

func workItemIntField(referenceName string) transform.TransformFunc {
	return func(_ context.Context, d *transform.TransformData) (interface{}, error) {
		switch value := workItemField(d.HydrateItem, referenceName).(type) {
		case int:
			return value, nil
		case int32:
			return int(value), nil
		case int64:
			return int(value), nil
		case float32:
			return int(value), nil
		case float64:
			return int(value), nil
		default:
			return nil, nil
		}
	}
}

func workItemTimeField(referenceName string) transform.TransformFunc {
	return func(_ context.Context, d *transform.TransformData) (interface{}, error) {
		value, ok := workItemField(d.HydrateItem, referenceName).(string)
		if !ok || value == "" {
			return nil, nil
		}
		parsed, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return nil, nil
		}
		return parsed, nil
	}
}

func workItemIdentityName(referenceName string) transform.TransformFunc {
	return func(_ context.Context, d *transform.TransformData) (interface{}, error) {
		switch value := workItemField(d.HydrateItem, referenceName).(type) {
		case string:
			if value == "" {
				return nil, nil
			}
			return value, nil
		case map[string]interface{}:
			if name, ok := value["displayName"].(string); ok && name != "" {
				return name, nil
			}
			if name, ok := value["uniqueName"].(string); ok && name != "" {
				return name, nil
			}
		}
		return nil, nil
	}
}
