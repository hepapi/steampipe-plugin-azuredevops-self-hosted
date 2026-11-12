package azuredevops

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/turbot/go-kit/types"
)

func TestGraphGroupFromIdentity(t *testing.T) {
	active := true
	container := true
	memberOf := []string{"vssgp.admins", " "}
	id := uuid.MustParse("d47d025a-ce2f-4a79-8618-e8862ade30dd")
	item := identity.Identity{
		IsContainer:         &container,
		IsActive:            &active,
		Id:                  &id,
		Descriptor:          types.String("Microsoft.TeamFoundation.Identity;S-1-9-1"),
		SubjectDescriptor:   types.String("vssgp.admins-group"),
		ProviderDisplayName: types.String("Project Collection Administrators"),
		MemberOf:            &memberOf,
		Properties: map[string]interface{}{
			"Account":     map[string]interface{}{"$value": `[TEAM FOUNDATION]\Project Collection Administrators`},
			"Description": map[string]interface{}{"$value": "Collection administrators"},
			"Domain":      map[string]interface{}{"$value": "vstfs:///Framework/IdentityDomain/x"},
			"ScopeId":     map[string]interface{}{"$value": "scope-1"},
			"Mail":        map[string]interface{}{"$value": "admins@contoso.com"},
		},
	}

	row, ok := graphGroupFromIdentity(item)
	if !ok {
		t.Fatal("expected a group row")
	}
	if stringValue(row.Descriptor) != "vssgp.admins-group" || stringValue(row.LegacyDescriptor) != "Microsoft.TeamFoundation.Identity;S-1-9-1" {
		t.Fatalf("descriptors = %s %s", stringValue(row.Descriptor), stringValue(row.LegacyDescriptor))
	}
	if stringValue(row.DisplayName) != "Project Collection Administrators" {
		t.Fatalf("display name = %s", stringValue(row.DisplayName))
	}
	if stringValue(row.PrincipalName) != `[TEAM FOUNDATION]\Project Collection Administrators` {
		t.Fatalf("principal = %s", stringValue(row.PrincipalName))
	}
	if stringValue(row.Description) != "Collection administrators" || stringValue(row.Domain) != "scope-1" {
		t.Fatalf("description/domain = %s %s", stringValue(row.Description), stringValue(row.Domain))
	}
	if stringValue(row.MailAddress) != "admins@contoso.com" || stringValue(row.Origin) != "vsts" {
		t.Fatalf("mail/origin = %s %s", stringValue(row.MailAddress), stringValue(row.Origin))
	}
	if stringValue(row.SubjectKind) != "group" || stringValue(row.OriginId) != id.String() {
		t.Fatalf("kind/origin id = %s %s", stringValue(row.SubjectKind), stringValue(row.OriginId))
	}
	if row.MembershipActive == nil || !*row.MembershipActive {
		t.Fatalf("membership = %v", row.MembershipActive)
	}
	if row.Memberships == nil || len(*row.Memberships) != 1 || stringValue((*row.Memberships)[0].ContainerDescriptor) != "vssgp.admins" {
		t.Fatalf("memberships = %+v", row.Memberships)
	}
}

func TestGraphGroupFromIdentitySkipsUsers(t *testing.T) {
	user := false
	_, ok := graphGroupFromIdentity(identity.Identity{
		IsContainer: &user,
		Descriptor:  types.String("win.ada"),
	})
	if ok {
		t.Fatal("expected a user to be skipped")
	}
}

func TestClassifyCollectionGroups(t *testing.T) {
	container := true
	user := false
	admins := identity.Identity{
		IsContainer:         &container,
		Descriptor:          types.String("valid-users"),
		SubjectDescriptor:   types.String("vssgp.valid-users"),
		ProviderDisplayName: types.String("Valid Users"),
		Members:             &[]string{"team-a", "win.ada", "team-a"},
	}
	ada := identity.Identity{
		IsContainer:         &user,
		SubjectDescriptor:   types.String("win.ada"),
		Descriptor:          types.String("legacy-ada"),
		ProviderDisplayName: types.String("Ada"),
	}
	queued := map[string]struct{}{"valid-users": {}}
	emitted := map[string]struct{}{}

	groups, next := classifyCollectionGroups([]identity.Identity{admins, ada}, queued, emitted)
	if len(groups) != 1 || stringValue(groups[0].Descriptor) != "vssgp.valid-users" || stringValue(groups[0].DisplayName) != "Valid Users" {
		t.Fatalf("groups = %+v", groups)
	}
	if len(next) != 2 || next[0] != "team-a" || next[1] != "win.ada" {
		t.Fatalf("next = %+v", next)
	}

	groups, next = classifyCollectionGroups([]identity.Identity{admins, ada}, queued, emitted)
	if len(groups) != 0 || len(next) != 0 {
		t.Fatalf("second pass groups=%+v next=%+v", groups, next)
	}

	team := identity.Identity{
		IsContainer:         &container,
		Descriptor:          types.String("team-a"),
		ProviderDisplayName: types.String("Team A"),
		Members:             &[]string{"win.ada"},
	}
	groups, next = classifyCollectionGroups([]identity.Identity{team}, queued, emitted)
	if len(groups) != 1 || stringValue(groups[0].Descriptor) != "team-a" || len(next) != 0 {
		t.Fatalf("nested group = %+v next=%+v", groups, next)
	}
}

func TestGraphGroupsEndpoint(t *testing.T) {
	endpoint := graphGroupsEndpoint("https://code.example.net/itcollection/", "next page", "6.0-preview.1")
	parsed, err := url.Parse(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Path != "/itcollection/_apis/graph/groups" {
		t.Fatalf("path = %s", parsed.Path)
	}
	if parsed.Query().Get("api-version") != "6.0-preview.1" || parsed.Query().Get("continuationToken") != "next page" {
		t.Fatalf("query = %s", parsed.RawQuery)
	}

	groupEndpoint := graphGroupEndpoint("https://code.example.net/itcollection", "vssgp.Uy0;", "6.0-preview.1")
	parsed, err = url.Parse(groupEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.EscapedPath() != "/itcollection/_apis/graph/groups/vssgp.Uy0%3B" {
		t.Fatalf("group path = %s", parsed.EscapedPath())
	}
}
