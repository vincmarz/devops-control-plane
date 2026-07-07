package app

import (
	"os"
	"strings"
	"testing"
)

func TestChangeServiceWiresRuntimeClientSecretRefsRegistry(t *testing.T) {
	content, err := os.ReadFile("change_service.go")
	if err != nil {
		t.Fatalf("failed to read change_service.go: %v", err)
	}
	text := string(content)
	checks := []string{"func WithRuntimeClientSecretRefsRegistry(", "runtimeClientSecretRefsRegistry", "SecretRefsConfigured", "selection.SecretRefs"}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("change_service.go should contain %q", check)
		}
	}
}
