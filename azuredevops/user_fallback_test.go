package azuredevops

import (
	"net/url"
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/turbot/go-kit/types"
)

func TestGraphUserFromIdentity(t *testing.T) {
	active := true
	user := false
	memberOf := []string{"vssgp.admins", " "}
	item := identity.Identity{
		IsContainer:         &user,
		IsActive:            &active,
		Descriptor:          types.String("win.legacy"),
		SubjectDescriptor:   types.String("win.ada"),
		ProviderDisplayName: types.String("Ada Lovelace"),
		MemberOf:            &memberOf,
		Properties: map[string]interface{}{
			"Account": map[string]interface{}{"$value": `CONTOSO\ada`},
			"Mail":    map[string]interface{}{"$value": "ada@contoso.com"},
			"Domain":  map[string]interface{}{"$value": "CONTOSO"},
			"Origin":  "AAD",
		},
	}

	row, ok := graphUserFromIdentity(item)
	if !ok {
		t.Fatal("expected a user row")
	}
	if stringValue(row.Descriptor) != "win.ada" || stringValue(row.LegacyDescriptor) != "win.legacy" {
		t.Fatalf("descriptors = %s %s", stringValue(row.Descriptor), stringValue(row.LegacyDescriptor))
	}
	if stringValue(row.DisplayName) != "Ada Lovelace" || stringValue(row.PrincipalName) != `CONTOSO\ada` {
		t.Fatalf("names = %s %s", stringValue(row.DisplayName), stringValue(row.PrincipalName))
	}
	if stringValue(row.MailAddress) != "ada@contoso.com" || stringValue(row.Domain) != "CONTOSO" {
		t.Fatalf("mail/domain = %s %s", stringValue(row.MailAddress), stringValue(row.Domain))
	}
	if stringValue(row.Origin) != "aad" || stringValue(row.DirectoryAlias) != "ada" {
		t.Fatalf("origin/alias = %s %s", stringValue(row.Origin), stringValue(row.DirectoryAlias))
	}
	if row.MembershipActive == nil || !*row.MembershipActive {
		t.Fatalf("membership = %v", row.MembershipActive)
	}
	if row.Memberships == nil || len(*row.Memberships) != 1 || stringValue((*row.Memberships)[0].ContainerDescriptor) != "vssgp.admins" {
		t.Fatalf("memberships = %+v", row.Memberships)
	}
}

func TestGraphUserFromIdentitySkipsGroups(t *testing.T) {
	container := true
	_, ok := graphUserFromIdentity(identity.Identity{
		IsContainer: &container,
		Descriptor:  types.String("vssgp.group"),
	})
	if ok {
		t.Fatal("expected a group to be skipped")
	}
}

func TestClassifyCollectionUsers(t *testing.T) {
	container := true
	user := false
	group := identity.Identity{
		IsContainer: &container,
		Descriptor:  types.String("valid-users"),
		Members:     &[]string{"team-a", "win.ada", "team-a"},
	}
	ada := identity.Identity{
		IsContainer:         &user,
		SubjectDescriptor:   types.String("win.ada"),
		Descriptor:          types.String("legacy-ada"),
		ProviderDisplayName: types.String("Ada"),
	}
	queued := map[string]struct{}{"valid-users": {}}
	emitted := map[string]struct{}{}

	users, next := classifyCollectionUsers([]identity.Identity{group, ada}, queued, emitted)
	if len(users) != 1 || stringValue(users[0].Descriptor) != "win.ada" || stringValue(users[0].DisplayName) != "Ada" {
		t.Fatalf("users = %+v", users)
	}
	if len(next) != 1 || next[0] != "team-a" {
		t.Fatalf("next = %+v", next)
	}

	users, next = classifyCollectionUsers([]identity.Identity{group, ada}, queued, emitted)
	if len(users) != 0 || len(next) != 0 {
		t.Fatalf("second pass users=%+v next=%+v", users, next)
	}

	// A descriptor is marked queued when it is discovered as a member. The
	// later read of that identity still has to produce a row.
	queued["win.bea"] = struct{}{}
	bea := identity.Identity{
		Descriptor:          types.String("win.bea"),
		ProviderDisplayName: types.String("Bea"),
	}
	users, next = classifyCollectionUsers([]identity.Identity{bea}, queued, emitted)
	if len(users) != 1 || stringValue(users[0].Descriptor) != "win.bea" || len(next) != 0 {
		t.Fatalf("queued user = %+v next=%+v", users, next)
	}
}

func TestOriginFromDescriptor(t *testing.T) {
	cases := map[string]string{
		"aad.ada": "aad",
		"msa.ada": "msa",
		"win.ada": "ad",
		"svc.ada": "vsts",
		"imp.ada": "vsts",
		"other":   "",
	}
	for descriptor, origin := range cases {
		if got := originFromDescriptor(descriptor); got != origin {
			t.Fatalf("%s origin = %s", descriptor, got)
		}
	}
}

func TestGraphUsersEndpoint(t *testing.T) {
	endpoint := graphUsersEndpoint("https://code.example.net/hecollection/", "next page", "6.0-preview.1")
	parsed, err := url.Parse(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Path != "/hecollection/_apis/graph/users" {
		t.Fatalf("path = %s", parsed.Path)
	}
	if parsed.Query().Get("api-version") != "6.0-preview.1" || parsed.Query().Get("continuationToken") != "next page" {
		t.Fatalf("query = %s", parsed.RawQuery)
	}

	userEndpoint := graphUserEndpoint("https://code.example.net/hecollection", "win.Uy0;", "6.0-preview.1")
	parsed, err = url.Parse(userEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.EscapedPath() != "/hecollection/_apis/graph/users/win.Uy0%3B" {
		t.Fatalf("user path = %s", parsed.EscapedPath())
	}
}

func TestUnsupportedAPIVersion(t *testing.T) {
	status := 400
	message := "The requested api-version is not supported."
	if !unsupportedAPIVersion(azuredevops.WrappedError{StatusCode: &status, Message: &message}) {
		t.Fatal("expected an api-version error")
	}
	if unsupportedAPIVersion(azuredevops.WrappedError{StatusCode: &status, Message: types.String("denied")}) {
		t.Fatal("a generic 400 is not an api-version error")
	}
}

func TestFirstContinuation(t *testing.T) {
	if firstContinuation(nil) != "" {
		t.Fatal("nil continuation should be empty")
	}
	tokens := []string{"  next  "}
	if firstContinuation(&tokens) != "next" {
		t.Fatalf("continuation = %q", firstContinuation(&tokens))
	}
}
