package api

import "testing"

func TestUpdateGitOpsRequiredRoles(t *testing.T) {
	roles, ok := requiredRolesForAction("update-gitops")
	if !ok {
		t.Fatal("update-gitops action is not registered")
	}
	if len(roles) != 2 || roles[0] != "operator" || roles[1] != "admin" {
		t.Fatalf("roles=%#v", roles)
	}
}
