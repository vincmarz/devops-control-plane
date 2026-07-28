package api

import (
	"os"
	"strings"
	"testing"
)

func TestUIChangeDetailExposesPersistedRuntimeSections(t *testing.T) {
	content, err := os.ReadFile("ui_handlers.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	required := []string{
		"ChangeRuntimeState",
		"GetRuntimeState(r.Context(), id)",
		"Technical runtime state",
		"Source repository",
		"GitOps repository",
		"Tekton validation state",
		"Argo CD deployment state",
		"Kubernetes runtime state",
		"No runtime state recorded",
	}
	for _, value := range required {
		if !strings.Contains(text, value) {
			t.Fatalf("ui_handlers.go does not contain %q", value)
		}
	}
}

func TestUIRuntimeStateDoesNotExposeSensitiveConfiguration(t *testing.T) {
	content, err := os.ReadFile("ui_handlers.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	forbidden := []string{"secretValue", "authorizationHeader", "runtimeSecretValue"}
	for _, value := range forbidden {
		if strings.Contains(text, value) {
			t.Fatalf("ui_handlers.go contains sensitive runtime field %q", value)
		}
	}
}
