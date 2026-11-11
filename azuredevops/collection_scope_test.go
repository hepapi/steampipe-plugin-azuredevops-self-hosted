package azuredevops

import (
	"net/http"
	"strings"
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func strPtr(value string) *string {
	return &value
}

func TestResolveCollectionPlan(t *testing.T) {
	tests := []struct {
		name   string
		config azureDevOpsConfig
		multi  bool
		allow  []string
		deny   []string
		server string
	}{
		{
			name:   "bare host discovers every collection",
			config: azureDevOpsConfig{OrganizationURL: strPtr("https://code.example.net")},
			multi:  true,
			server: "https://code.example.net",
		},
		{
			name:   "collection path stays single",
			config: azureDevOpsConfig{OrganizationURL: strPtr("https://code.example.net/DefaultCollection")},
			multi:  false,
			server: "https://code.example.net",
		},
		{
			name:   "services organization stays single",
			config: azureDevOpsConfig{OrganizationURL: strPtr("https://dev.azure.com/example")},
			multi:  false,
			server: "https://dev.azure.com",
		},
		{
			name: "allow list",
			config: azureDevOpsConfig{
				OrganizationURL: strPtr("https://code.example.net"),
				Collections:     []string{" One ", "Two"},
			},
			multi:  true,
			allow:  []string{"One", "Two"},
			server: "https://code.example.net",
		},
		{
			name: "deny list discovers then excludes",
			config: azureDevOpsConfig{
				OrganizationURL:     strPtr("https://code.example.net/DefaultCollection"),
				ExcludedCollections: []string{"LegacyTfs"},
			},
			multi:  true,
			deny:   []string{"LegacyTfs"},
			server: "https://code.example.net",
		},
		{
			name: "wildcard forces discovery and keeps the deny list",
			config: azureDevOpsConfig{
				OrganizationURL:     strPtr("https://host.example/tfs/DefaultCollection"),
				ServerURL:           strPtr("https://host.example/tfs"),
				Collections:         []string{"*"},
				ExcludedCollections: []string{"LegacyTfs"},
			},
			multi:  true,
			deny:   []string{"LegacyTfs"},
			server: "https://host.example/tfs",
		},
		{
			name: "wildcard ignores other allow names",
			config: azureDevOpsConfig{
				OrganizationURL: strPtr("https://code.example.net/DefaultCollection"),
				Collections:     []string{"*", "Foo"},
			},
			multi:  true,
			server: "https://code.example.net",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := resolveCollectionPlan(test.config)
			if plan.Multi != test.multi {
				t.Fatalf("multi = %v, want %v", plan.Multi, test.multi)
			}
			if strings.Join(plan.Allow, ",") != strings.Join(test.allow, ",") {
				t.Fatalf("allow = %#v, want %#v", plan.Allow, test.allow)
			}
			if strings.Join(plan.Deny, ",") != strings.Join(test.deny, ",") {
				t.Fatalf("deny = %#v, want %#v", plan.Deny, test.deny)
			}
			if plan.ServerURL != test.server {
				t.Fatalf("server = %s, want %s", plan.ServerURL, test.server)
			}
		})
	}
}

func TestSelectCollections(t *testing.T) {
	server := "https://code.example.net"
	discovered := []projectCollection{
		{Name: "Git", URL: collectionBaseURL(server, "Git")},
		{Name: "LegacyTfs", URL: collectionBaseURL(server, "LegacyTfs")},
		{Name: "Archive", URL: collectionBaseURL(server, "Archive")},
	}

	selected, err := selectCollections(server, discovered, []string{"git", "Archive"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := collectionNames(selected); got != "Git,Archive" {
		t.Fatalf("allow list = %s", got)
	}

	selected, err = selectCollections(server, discovered, nil, []string{"legacytfs"})
	if err != nil {
		t.Fatal(err)
	}
	if got := collectionNames(selected); got != "Git,Archive" {
		t.Fatalf("deny list = %s", got)
	}

	selected, err = selectCollections(server, discovered, []string{"Git", "LegacyTfs"}, []string{"LegacyTfs"})
	if err != nil {
		t.Fatal(err)
	}
	if got := collectionNames(selected); got != "Git" {
		t.Fatalf("allow minus deny = %s", got)
	}

	_, err = selectCollections(server, discovered, []string{"Missing"}, nil)
	if err == nil || !strings.Contains(err.Error(), "Missing") || !strings.Contains(err.Error(), "Git, LegacyTfs, Archive") {
		t.Fatalf("allow typo error = %v", err)
	}

	_, err = selectCollections(server, discovered, nil, []string{"Missing"})
	if err == nil || !strings.Contains(err.Error(), "excluded collection \"Missing\"") || !strings.Contains(err.Error(), "discovered collections:") {
		t.Fatalf("deny typo error = %v", err)
	}

	_, err = selectCollections(server, discovered, nil, []string{"Git", "LegacyTfs", "Archive"})
	if err == nil || !strings.Contains(err.Error(), "no project collections left") {
		t.Fatalf("empty filter error = %v", err)
	}

	_, err = selectCollections(server, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "no project collections were discovered") {
		t.Fatalf("empty discovery error = %v", err)
	}
}

func TestCollectionBaseURL(t *testing.T) {
	if got := collectionBaseURL("https://code.example.net/", "DefaultCollection"); got != "https://code.example.net/DefaultCollection" {
		t.Fatalf("base url = %s", got)
	}
	if got := collectionBaseURL("https://host.example/tfs", "Git"); got != "https://host.example/tfs/Git" {
		t.Fatalf("virtual directory url = %s", got)
	}
}

func TestIsNotFound(t *testing.T) {
	status := http.StatusNotFound
	message := "missing"
	value := azuredevops.WrappedError{StatusCode: &status, Message: &message}
	if !isNotFound(value) {
		t.Fatal("value 404 was not recognized")
	}
	if !isNotFound(&value) {
		t.Fatal("pointer 404 was not recognized")
	}
	other := http.StatusBadRequest
	if isNotFound(azuredevops.WrappedError{StatusCode: &other, Message: &message}) {
		t.Fatal("400 was treated as not found")
	}
	if isNotFound(azuredevops.WrappedError{Message: &message}) {
		t.Fatal("error without a status was treated as not found")
	}
}

func TestAbsentFromCollection(t *testing.T) {
	status := http.StatusNotFound
	err := azuredevops.WrappedError{StatusCode: &status}
	multi := &plugin.QueryData{EqualsQuals: plugin.KeyColumnEqualsQualMap{
		"collection_url": proto.NewQualValue("https://code.example.net/Git"),
	}}
	if !absentFromCollection(multi, err) {
		t.Fatal("expected a multi-collection 404 to be absent")
	}
	single := &plugin.QueryData{}
	if absentFromCollection(single, err) {
		t.Fatal("single-collection 404 must stay an error")
	}
}

func collectionNames(collections []projectCollection) string {
	names := make([]string, len(collections))
	for i, collection := range collections {
		names[i] = collection.Name
	}
	return strings.Join(names, ",")
}
