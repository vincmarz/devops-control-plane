package api

import (
	"os"
	"strings"
	"testing"
)

func TestGitWorkflowErrorsUseProviderNeutralPublicContract(t *testing.T) {
	content, err := os.ReadFile("change_handlers.go")
	if err != nil {
		t.Fatalf("ReadFile change_handlers.go: %v", err)
	}
	text := string(content)
	required := []string{
		`Code: "GIT_CREATE_BRANCH_FAILED", Message: "Unable to create Git branch for ChangeRequest"`,
		`Code: "GIT_UPDATE_FILES_FAILED", Message: "Unable to update Git files for ChangeRequest"`,
		`Code: "GIT_OPEN_REVIEW_REQUEST_FAILED", Message: "Unable to open Git review request for ChangeRequest"`,
		`Code: "GIT_MERGE_REVIEW_REQUEST_FAILED", Message: "Unable to merge Git review request for ChangeRequest"`,
	}
	for _, value := range required {
		if !strings.Contains(text, value) {
			t.Fatalf("change_handlers.go does not contain %q", value)
		}
	}
	forbidden := []string{
		"GITLAB_CREATE_BRANCH_FAILED",
		"GITLAB_UPDATE_FILES_FAILED",
		"GITLAB_OPEN_MERGE_REQUEST_FAILED",
		"GITLAB_MERGE_REQUEST_FAILED",
		"Unable to create GitLab branch for ChangeRequest",
		"Unable to update GitLab files for ChangeRequest",
		"Unable to open GitLab merge request for ChangeRequest",
		"Unable to merge GitLab merge request for ChangeRequest",
	}
	for _, value := range forbidden {
		if strings.Contains(text, value) {
			t.Fatalf("change_handlers.go still contains provider-specific contract %q", value)
		}
	}
}
