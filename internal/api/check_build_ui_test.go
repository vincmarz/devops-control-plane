package api

import "testing"

func checkBuildActionNames(actions []map[string]any) map[string]bool {
	result := map[string]bool{}
	for _, action := range actions {
		name, _ := action["name"].(string)
		result[name] = true
	}
	return result
}

func TestCheckBuildRecommendedWhileBuildRunning(t *testing.T) {
	actions := recommendedActions(map[string]any{"runtimeStatus": "BuildRunning"})
	if !checkBuildActionNames(actions)["check-build"] {
		t.Fatalf("actions=%#v", actions)
	}
	primary, _ := actions[0]["primary"].(bool)
	if !primary {
		t.Fatalf("check-build not primary: %#v", actions[0])
	}
}

func TestStartBuildRecommendedAfterBuildFailed(t *testing.T) {
	actions := recommendedActions(map[string]any{"runtimeStatus": "BuildFailed"})
	if !checkBuildActionNames(actions)["start-build"] {
		t.Fatalf("actions=%#v", actions)
	}
}

func TestCheckBuildAvailableAsAdvancedAction(t *testing.T) {
	if !checkBuildActionNames(allUIActions())["check-build"] {
		t.Fatal("check-build missing from all UI actions")
	}
}
