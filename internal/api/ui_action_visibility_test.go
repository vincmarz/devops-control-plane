package api

import (
	"context"
	"testing"
)

func TestUserCanExecuteTechnicalActions(t *testing.T) {
	tests := []struct {
		name  string
		roles map[string]bool
		want  bool
	}{
		{name: "viewer", roles: map[string]bool{"viewer": true}, want: false},
		{name: "approver", roles: map[string]bool{"approver": true}, want: false},
		{name: "operator", roles: map[string]bool{"operator": true}, want: true},
		{name: "admin", roles: map[string]bool{"admin": true}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), identityContextKey, authIdentity{Roles: tt.roles})
			got := userCanExecuteTechnicalActions(ctx)
			if got != tt.want {
				t.Fatalf("userCanExecuteTechnicalActions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithUIActionVisibility(t *testing.T) {
	ctx := context.WithValue(context.Background(), identityContextKey, authIdentity{Roles: map[string]bool{"operator": true}})
	change := map[string]any{"changeNumber": "CHG-2026-TEST"}

	visible := withUIActionVisibility(ctx, change)
	if !userCanSeeTechnicalActions(visible) {
		t.Fatal("expected operator to see technical actions")
	}
	if _, ok := change["uiCanSeeTechnicalActions"]; ok {
		t.Fatal("withUIActionVisibility should not mutate the original change map")
	}
}

func TestUserCanSeeTechnicalActionsDefaultsToFalse(t *testing.T) {
	if userCanSeeTechnicalActions(map[string]any{}) {
		t.Fatal("expected missing visibility flag to default to false")
	}
}
