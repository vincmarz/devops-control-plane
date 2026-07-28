package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainDeploymentEvidenceCollectorUsesResolvedArgoCDApplicationName(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile main.go: %v", err)
	}
	text := string(content)
	required := []string{
		"WithDeploymentEvidenceCollector(func(ctx context.Context, change domain.ChangeRequest, target app.TechnicalRuntimeTarget)",
		"argoCDClient.GetApplication(ctx, target.ArgoCDApplicationName)",
	}
	for _, value := range required {
		if !strings.Contains(text, value) {
			t.Fatalf("main.go does not contain %q", value)
		}
	}
	if strings.Contains(text, "argoCDClient.GetApplication(ctx, change.ApplicationName)") {
		t.Fatal("deployment evidence collector still uses the logical application name")
	}
}
