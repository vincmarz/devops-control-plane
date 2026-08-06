package api

import "testing"

func TestCheckBuildRequiredRoles(t *testing.T) {
	roles, ok := requiredRolesForAction("check-build")
	if !ok {
		t.Fatal("check-build action is not registered")
	}
	if len(roles) != 2 || roles[0] != "operator" || roles[1] != "admin" {
		t.Fatalf("roles=%#v", roles)
	}
}
