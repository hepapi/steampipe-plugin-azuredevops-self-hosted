package azuredevops

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/delegatedauthorization"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
)

func TestPersonalAccessTokenRowOmitsSecrets(t *testing.T) {
	secret := "secret-value"
	alternate := "alternate-secret"
	authorizationID := uuid.MustParse("4ab5764f-4193-4f1d-b995-64144880b7d7")
	accountID := uuid.MustParse("38aaa865-2c70-4bf7-a308-0c6539c38c1a")
	valid := true
	expires := azuredevops.Time{Time: time.Date(2026, 12, 1, 23, 46, 23, 0, time.UTC)}
	token := delegatedauthorization.SessionToken{
		AuthorizationId: &authorizationID,
		DisplayName:     patString("build"),
		Scope:           patString("vso.code vso.build"),
		Token:           &secret,
		AlternateToken:  &alternate,
		IsValid:         &valid,
		ValidTo:         &expires,
		TargetAccounts:  &[]uuid.UUID{accountID},
	}

	row := personalAccessTokenRowFrom(&token, patUser{
		SubjectDescriptor: "aad.ada",
		DisplayName:       "Ada",
	})

	if token.Token != nil || token.AlternateToken != nil {
		t.Fatal("token secrets were not cleared on the source object")
	}
	if row.AuthorizationID != authorizationID.String() || row.DisplayName != "build" || row.Scope != "vso.code vso.build" {
		t.Fatalf("metadata = %+v", row)
	}
	if row.SubjectDescriptor != "aad.ada" || row.UserDisplayName != "Ada" || row.IsValid == nil || !*row.IsValid {
		t.Fatalf("owner = %+v", row)
	}
	if row.ValidTo == nil || !row.ValidTo.Equal(expires.Time) {
		t.Fatalf("valid_to = %v", row.ValidTo)
	}
	if len(row.TargetAccounts) != 1 || row.TargetAccounts[0] != accountID.String() {
		t.Fatalf("target_accounts = %v", row.TargetAccounts)
	}

	rowType := reflect.TypeOf(row)
	for i := 0; i < rowType.NumField(); i++ {
		name := rowType.Field(i).Name
		if name == "Token" || name == "AlternateToken" {
			t.Fatalf("row exposes %s", name)
		}
	}
	encoded, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), secret) || strings.Contains(string(encoded), alternate) {
		t.Fatalf("marshaled row contains a secret: %s", encoded)
	}
}

func TestPatUsersFromIdentities(t *testing.T) {
	container := true
	user := false
	groups := identity.Identity{
		IsContainer:         &container,
		Descriptor:          patString("group-desc"),
		ProviderDisplayName: patString("Valid Users"),
	}
	ada := identity.Identity{
		IsContainer:         &user,
		SubjectDescriptor:   patString("aad.ada"),
		Descriptor:          patString("legacy-ada"),
		ProviderDisplayName: patString("Ada"),
	}
	bea := identity.Identity{
		Descriptor:          patString("legacy-bea"),
		CustomDisplayName:   patString("Bea"),
		ProviderDisplayName: patString("  "),
	}

	users := patUsersFromIdentities([]identity.Identity{groups, ada, ada, bea})
	if len(users) != 2 {
		t.Fatalf("users = %+v", users)
	}
	if users[0].SubjectDescriptor != "aad.ada" || users[0].DisplayName != "Ada" {
		t.Fatalf("first user = %+v", users[0])
	}
	if users[1].SubjectDescriptor != "legacy-bea" || users[1].DisplayName != "Bea" {
		t.Fatalf("second user = %+v", users[1])
	}
}

func TestClassifyPatMembers(t *testing.T) {
	container := true
	group := identity.Identity{
		IsContainer: &container,
		Descriptor:  patString("valid-users"),
		Members:     &[]string{"team-a", "aad.ada", "team-a"},
	}
	user := false
	ada := identity.Identity{
		IsContainer:         &user,
		SubjectDescriptor:   patString("aad.ada"),
		Descriptor:          patString("legacy-ada"),
		ProviderDisplayName: patString("Ada"),
	}
	seen := map[string]struct{}{"valid-users": {}}

	users, next := classifyPatMembers([]identity.Identity{group, ada}, seen)
	if len(users) != 1 || users[0].SubjectDescriptor != "aad.ada" || users[0].DisplayName != "Ada" {
		t.Fatalf("users = %+v", users)
	}
	if len(next) != 1 || next[0] != "team-a" {
		t.Fatalf("next = %+v", next)
	}

	users, next = classifyPatMembers([]identity.Identity{group, ada}, seen)
	if len(users) != 0 || len(next) != 0 {
		t.Fatalf("second pass users=%+v next=%+v", users, next)
	}
}

func TestPersonalAccessTokenEndpoint(t *testing.T) {
	endpoint := personalAccessTokenEndpoint("https://azuredevops.example.com/DefaultCollection/", "win.Uy0;", "next-page", 100)
	parsed, err := url.Parse(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.EscapedPath() != "/DefaultCollection/_apis/tokenadmin/personalaccesstokens/win.Uy0%3B" {
		t.Fatalf("path = %s", parsed.EscapedPath())
	}
	query := parsed.Query()
	if query.Get("api-version") != patAPIVersion || query.Get("isPublic") != "false" || query.Get("pageSize") != "100" {
		t.Fatalf("query = %s", query.Encode())
	}
	if query.Get("continuationToken") != "next-page" {
		t.Fatalf("continuationToken = %s", query.Get("continuationToken"))
	}
}

func TestContinuationTokenString(t *testing.T) {
	if continuationTokenString(nil) != "" {
		t.Fatal("nil continuation should be empty")
	}
	if continuationTokenString(json.RawMessage("null")) != "" {
		t.Fatal("null continuation should be empty")
	}
	if continuationTokenString(json.RawMessage(`"4ab5764f-4193-4f1d-b995-64144880b7d7"`)) != "4ab5764f-4193-4f1d-b995-64144880b7d7" {
		t.Fatal("string continuation was not decoded")
	}
}

func patString(value string) *string {
	return &value
}
