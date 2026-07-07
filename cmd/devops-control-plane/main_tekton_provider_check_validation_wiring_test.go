package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainWiresTektonRuntimeProviderIntoCheckValidation(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	text := string(content)

	checks := []string{
		"DefaultTechnicalRuntimeTargetResolver(cfg.TektonPipelineName).Resolve(change.TargetEnvironment)",
		"DefaultRuntimeClientProviderRegistry().Select(target)",
		"DefaultTektonRuntimeClientProviderRegistry(currentTektonRuntimeClient{client: tektonClient}).Resolve(ctx, selection)",
		"FindLatestPipelineRunByChange(ctx, target.TektonNamespace, change.ChangeNumber)",
		"ListTaskRunsByPipelineRun(ctx, target.TektonNamespace, status.PipelineRunName)",
		"PipelineName: target.TektonPipelineName",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("main.go should contain %q", check)
		}
	}
}
