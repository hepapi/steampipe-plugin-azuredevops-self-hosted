package azuredevops

import (
	"errors"
	"testing"
)

func TestIsNoBranchesError(t *testing.T) {
	missing := errors.New(`Branch "refs/heads/master" does not exist in the c21376ef-3157-49eb-84fd-c1d9ab703b84 repository.`)
	if !isNoBranchesError(missing) {
		t.Fatal("deleted default branch should be absent")
	}
	if !isNoBranchesError(errors.New("VS403403: Cannot find any branches")) {
		t.Fatal("empty repository should be absent")
	}
	if isNoBranchesError(errors.New("connection reset")) {
		t.Fatal("unrelated error should fail the hydrate")
	}
}

func TestReleaseManagementUnavailable(t *testing.T) {
	missing := errors.New("Could not find dataspace with category ReleaseManagement and identifier 70b12a44-300e-4404-94b8-c3ff1dfe14f0")
	if !releaseManagementUnavailable(missing) {
		t.Fatal("project without release management should be skipped")
	}
	if releaseManagementUnavailable(errors.New("Could not find dataspace with category Git")) {
		t.Fatal("another dataspace should still fail")
	}
}
