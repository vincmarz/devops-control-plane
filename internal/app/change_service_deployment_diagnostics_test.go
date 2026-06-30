package app

import (
	"testing"
)

func TestDeploymentEvidenceDiagnosticsHealthyWithWarning(t *testing.T) {
	payload := map[string]any{
		"change": map[string]any{"applicationName": "demo-go-color-app"},
		"argocd": map[string]any{
			"applicationName": "demo-go-color-app",
			"syncStatus":      "Synced",
			"healthStatus":    "Healthy",
			"conditions": []map[string]any{
				{"type": "OrphanedResourceWarning", "message": "Application has 5 orphaned resources"},
			},
		},
		"kubernetes": map[string]any{
			"deployment": map[string]any{
				"desiredReplicas":    2,
				"readyReplicas":      2,
				"availableReplicas":  2,
				"updatedReplicas":    2,
				"generation":         6,
				"observedGeneration": 6,
			},
			"pods": []map[string]any{
				{"ready": true, "restartCount": 0},
				{"ready": true, "restartCount": 0},
			},
			"service": map[string]any{"clusterIP": "172.30.61.187"},
			"route":   map[string]any{"host": "demo-go-color-app-devops-ci-demo.apps.ocp4.mim.lan"},
		},
	}

	diagnostics := deploymentEvidenceDiagnostics(payload)
	if diagnostics["argocdSynced"] != true || diagnostics["argocdHealthy"] != true || diagnostics["deploymentReady"] != true {
		t.Fatalf("unexpected status diagnostics: %#v", diagnostics)
	}
	if diagnostics["readyReplicas"] != "2/2" || diagnostics["podsReady"] != "2/2" || diagnostics["totalRestarts"] != 0 {
		t.Fatalf("unexpected runtime diagnostics: %#v", diagnostics)
	}
	warnings, ok := diagnostics["warnings"].([]string)
	if !ok || len(warnings) != 1 || warnings[0] != "OrphanedResourceWarning: Application has 5 orphaned resources" {
		t.Fatalf("unexpected warnings: %#v", diagnostics["warnings"])
	}
}

func TestDeploymentEvidenceDiagnosticsDetectsRuntimeIssues(t *testing.T) {
	payload := map[string]any{
		"argocd": map[string]any{"applicationName": "demo-go-color-app", "syncStatus": "OutOfSync", "healthStatus": "Degraded"},
		"kubernetes": map[string]any{
			"deployment": map[string]any{"desiredReplicas": 2, "readyReplicas": 1, "availableReplicas": 1, "updatedReplicas": 1, "generation": 6, "observedGeneration": 5},
			"pods": []map[string]any{
				{"ready": true, "restartCount": 1},
				{"ready": false, "restartCount": 2},
			},
			"service": map[string]any{"error": "not found"},
			"route":   map[string]any{"error": "not found"},
		},
	}

	diagnostics := deploymentEvidenceDiagnostics(payload)
	if diagnostics["argocdSynced"] != false || diagnostics["argocdHealthy"] != false || diagnostics["deploymentReady"] != false {
		t.Fatalf("unexpected degraded diagnostics: %#v", diagnostics)
	}
	if diagnostics["readyReplicas"] != "1/2" || diagnostics["podsReady"] != "1/2" || diagnostics["totalRestarts"] != 3 {
		t.Fatalf("unexpected runtime diagnostics: %#v", diagnostics)
	}
	warnings, ok := diagnostics["warnings"].([]string)
	if !ok || len(warnings) != 4 {
		t.Fatalf("unexpected warnings: %#v", diagnostics["warnings"])
	}
}
