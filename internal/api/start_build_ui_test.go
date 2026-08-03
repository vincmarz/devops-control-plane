package api

import "testing"

func actionNames(actions []map[string]any) map[string]bool {
	result := map[string]bool{}
	for _, action := range actions {
		name, _ := action["name"].(string)
		result[name] = true
	}
	return result
}

func TestStartBuildIsRecommendedAfterMerge(t *testing.T) {
	actions := recommendedActions(map[string]any{"runtimeStatus": "MergeRequestMerged"})
	if !actionNames(actions)["start-build"] {
		t.Fatalf("actions = %#v", actions)
	}
	primary, _ := actions[0]["primary"].(bool)
	if !primary {
		t.Fatalf("Start Build is not primary: %#v", actions[0])
	}
}

func TestStartBuildIsNotRecommendedWhileRunning(t *testing.T) {
	actions := recommendedActions(map[string]any{"runtimeStatus": "BuildRunning"})
	if actionNames(actions)["start-build"] {
		t.Fatalf("actions = %#v", actions)
	}
}

func TestStartBuildIsAvailableAsAdvancedAction(t *testing.T) {
	if !actionNames(allUIActions())["start-build"] {
		t.Fatal("Start Build is missing from all UI actions")
	}
}
