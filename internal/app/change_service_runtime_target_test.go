package app

import (
	"os"
	"strings"
	"testing"
)

func TestChangeServiceTechnicalWorkflowsResolveRuntimeTarget(t *testing.T) {
	content, err := os.ReadFile("change_service.go")
	if err != nil {
		t.Fatalf("failed to read change_service.go: %v", err)
	}
	text := string(content)

	for _, method := range []string{"Validate", "CheckValidation", "CheckDeployment", "CollectEvidence"} {
		marker := "func (s *ChangeService) " + method
		start := strings.Index(text, marker)
		if start == -1 {
			t.Fatalf("method %s not found", method)
		}
		end := strings.Index(text[start+1:], "\nfunc ")
		block := text[start:]
		if end != -1 {
			block = text[start : start+1+end]
		}
		if !strings.Contains(block, "resolveTechnicalRuntimeTarget(ctx, change)") {
			t.Fatalf("method %s should resolve TechnicalRuntimeTarget before technical execution", method)
		}
	}
}

func TestChangeServiceExposesTechnicalRuntimeTargetResolverOption(t *testing.T) {
	content, err := os.ReadFile("change_service.go")
	if err != nil {
		t.Fatalf("failed to read change_service.go: %v", err)
	}
	text := string(content)

	checks := []string{
		"type TechnicalRuntimeTargetResolverFunc",
		"func WithTechnicalRuntimeTargetResolver(",
		"func WithTechnicalRuntimeTargetResolverFunc(",
		"technicalRuntimeTargetResolver",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("change_service.go should contain %q", check)
		}
	}
}
