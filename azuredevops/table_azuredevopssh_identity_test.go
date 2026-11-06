package azuredevops

import (
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/turbot/go-kit/types"
)

func TestIdentityQueryValues(t *testing.T) {
	direct := identity.QueryMembershipValues.Direct
	query := identityQueryValues(identity.ReadIdentitiesArgs{
		SearchFilter:    types.String("General"),
		FilterValue:     types.String("Project Collection"),
		QueryMembership: &direct,
	})

	if query.Get("api-version") != "6.0" {
		t.Fatalf("api-version = %s", query.Get("api-version"))
	}
	if query.Get("searchFilter") != "General" || query.Get("filterValue") != "Project Collection" {
		t.Fatalf("search query = %s", query.Encode())
	}
	if query.Get("queryMembership") != "direct" {
		t.Fatalf("queryMembership = %s", query.Get("queryMembership"))
	}
}

func TestMembershipDirection(t *testing.T) {
	got, err := membershipDirection("")
	if err != nil || got != "down" {
		t.Fatalf("default direction = %s, %v", got, err)
	}
	if _, err := membershipDirection("sideways"); err == nil {
		t.Fatal("expected an error for an unknown direction")
	}
}

func TestMembershipEdges(t *testing.T) {
	down := membershipEdges("down", "group", []string{"user-a", "user-b"}, []string{"parent"})
	if len(down) != 2 || down[0].Container != "group" || down[0].Member != "user-a" {
		t.Fatalf("down edges = %+v", down)
	}

	up := membershipEdges("up", "user", []string{"ignored"}, []string{"parent"})
	if len(up) != 1 || up[0].Container != "parent" || up[0].Member != "user" {
		t.Fatalf("up edges = %+v", up)
	}
}
