package azuredevops

import (
	"context"
	"testing"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func TestWorkItemListWiql(t *testing.T) {
	got := workItemListWiql(0, nil, "", "")
	want := "SELECT [System.Id] FROM WorkItems WHERE [System.TeamProject] = @project AND [System.Id] > 0 ORDER BY [System.Id]"
	if got != want {
		t.Fatalf("first page:\n got %s\nwant %s", got, want)
	}

	id := 42
	got = workItemListWiql(10, &id, "User's Story", "Active")
	want = "SELECT [System.Id] FROM WorkItems WHERE [System.TeamProject] = @project AND [System.Id] = 42 AND [System.WorkItemType] = 'User''s Story' AND [System.State] = 'Active' ORDER BY [System.Id]"
	if got != want {
		t.Fatalf("filtered:\n got %s\nwant %s", got, want)
	}
}

func TestWorkItemFieldTransforms(t *testing.T) {
	fields := map[string]interface{}{
		"System.Title":                   "Fix login",
		"System.WorkItemType":            "Bug",
		"Microsoft.VSTS.Common.Priority": float64(2),
		"System.CreatedDate":             "2024-03-01T10:11:12.123Z",
		"System.AssignedTo": map[string]interface{}{
			"displayName": "Ada Lovelace",
			"uniqueName":  "ada@example.com",
		},
	}
	item := ProjectWorkItem{
		WorkItem:  workitemtracking.WorkItem{Fields: &fields},
		ProjectId: "project",
	}
	data := &transform.TransformData{HydrateItem: item}

	title, err := workItemStringField("System.Title")(context.Background(), data)
	if err != nil || title != "Fix login" {
		t.Fatalf("title = %#v, %v", title, err)
	}
	priority, err := workItemIntField("Microsoft.VSTS.Common.Priority")(context.Background(), data)
	if err != nil || priority != 2 {
		t.Fatalf("priority = %#v, %v", priority, err)
	}
	created, err := workItemTimeField("System.CreatedDate")(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	createdAt, ok := created.(time.Time)
	if !ok || createdAt.UTC().Format(time.RFC3339) != "2024-03-01T10:11:12Z" {
		t.Fatalf("created = %#v", created)
	}
	name, err := workItemIdentityName("System.AssignedTo")(context.Background(), data)
	if err != nil || name != "Ada Lovelace" {
		t.Fatalf("assigned_to_name = %#v, %v", name, err)
	}

	missing, err := workItemIntField("System.Parent")(context.Background(), data)
	if err != nil || missing != nil {
		t.Fatalf("missing parent = %#v, %v", missing, err)
	}
}
