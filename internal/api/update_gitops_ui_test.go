package api

import "testing"

func updateGitOpsActionNames(actions []map[string]any) map[string]bool {
	result := map[string]bool{}
	for _, action := range actions {
		name, _ := action["name"].(string)
		result[name] = true
	}
	return result
}

func TestUpdateGitOpsRecommendedAfterBuildSucceeded(t *testing.T) {
	actions := recommendedActions(map[string]any{"runtimeStatus": "BuildSucceeded"})
	if !updateGitOpsActionNames(actions)["update-gitops"] {
		t.Fatalf("actions=%#v", actions)
	}
	primary, _ := actions[0]["primary"].(bool)
	if !primary {
		t.Fatalf("update-gitops not primary: %#v", actions[0])
	}
}

func TestCheckDeploymentRecommendedAfterGitOpsUpdated(t *testing.T) {
	actions := recommendedActions(map[string]any{"runtimeStatus": "GitOpsUpdated"})
	if !updateGitOpsActionNames(actions)["check-deployment"] {
		t.Fatalf("actions=%#v", actions)
	}
}

func TestUpdateGitOpsAvailableAsAdvancedAction(t *testing.T) {
	if !updateGitOpsActionNames(allUIActions())["update-gitops"] {
		t.Fatal("update-gitops missing from all UI actions")
	}
}
