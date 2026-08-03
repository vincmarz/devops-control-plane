package api

import "testing"

func TestStartBuildRequiredRoles(t *testing.T) {
	roles, ok := requiredRolesForAction("start-build")
	if !ok {
		t.Fatal("start-build action is not registered")
	}
	if len(roles) != 2 || roles[0] != "operator" || roles[1] != "admin" {
		t.Fatalf("roles = %#v", roles)
	}
}
