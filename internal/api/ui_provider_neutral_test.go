package api

import (
	"os"
	"strings"
	"testing"
)

func TestUIUsesProviderNeutralGitReviewTerminology(t *testing.T) {
	content, err := os.ReadFile("ui_handlers.go")
	if err != nil {
		t.Fatalf("ReadFile ui_handlers.go: %v", err)
	}
	text := string(content)
	required := []string{
		`"Open Review Request"`,
		`"Merge Review Request"`,
		`"Open the Git review request."`,
		`"Merge the approved Git review request."`,
		`"Merge the review request when governance allows it."`,
		`Git Providers`,
	}
	for _, value := range required {
		if !strings.Contains(text, value) {
			t.Fatalf("ui_handlers.go does not contain %q", value)
		}
	}
	forbidden := []string{
		`GitLab`,
		`Open Merge Request`,
		`Merge the GitLab MR`,
	}
	for _, value := range forbidden {
		if strings.Contains(text, value) {
			t.Fatalf("ui_handlers.go still contains provider-specific text %q", value)
		}
	}
}

func TestUIKeepsCompatibleGitWorkflowActionIdentifiers(t *testing.T) {
	content, err := os.ReadFile("ui_handlers.go")
	if err != nil {
		t.Fatalf("ReadFile ui_handlers.go: %v", err)
	}
	text := string(content)
	required := []string{
		`uiAction("open-merge-request",`,
		`uiAction("merge-request",`,
	}
	for _, value := range required {
		if !strings.Contains(text, value) {
			t.Fatalf("ui_handlers.go does not preserve action identifier %q", value)
		}
	}
}
