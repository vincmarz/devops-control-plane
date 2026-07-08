package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainWiresTektonRuntimeProviderIntoValidate(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	text := string(content)

	checks := []string{
		"DefaultTechnicalRuntimeTargetResolver(cfg.TektonPipelineName).Resolve(change.TargetEnvironment)",
		"DefaultRuntimeClientProviderRegistry().Select(target)",
		"tektonRuntimeClientProviderRegistry.Resolve(ctx, selection)",
		"TektonRuntimePipelineRunRequest",
		"Namespace:          target.TektonNamespace",
		"PipelineName:       target.TektonPipelineName",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("main.go should contain %q", check)
		}
	}
}
