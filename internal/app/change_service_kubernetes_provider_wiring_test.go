package app

import (
	"os"
	"strings"
	"testing"
)

func TestChangeServiceCollectEvidenceUsesKubernetesRuntimeProvider(t *testing.T) {
	content, err := os.ReadFile("change_service.go")
	if err != nil {
		t.Fatalf("failed to read change_service.go: %v", err)
	}
	text := string(content)

	checks := []string{
		"func WithKubernetesRuntimeClientProviderRegistry(",
		"kubernetesRuntimeClientProviderRegistry",
		"collectKubernetesRuntimeEvidence(ctx, change, providerSelection)",
		"selection.Target.KubernetesNamespace",
		"client.CollectRuntimeEvidence(ctx, namespace, change.ApplicationName)",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("change_service.go should contain %q", check)
		}
	}
}
