package api

import "testing"

func TestLatestValidationEvidenceFindsTektonEvidence(t *testing.T) {
	items := []map[string]any{
		{"name": "deployment-evidence", "evidenceType": "deployment"},
		{
			"name":         "tekton-validation-evidence",
			"evidenceType": "validation",
			"sanitized":    true,
			"summary":      "Tekton validation completed without failed TaskRuns",
			"payload": map[string]any{
				"tekton": map[string]any{
					"pipelineRunName": "devops-cp-validate-chg-2026-0049-nd7rm",
					"tektonNamespace": "devops-ci-staging",
					"pipelineName":    "validate-gitops",
					"revision":        "change/CHG-2026-0049",
					"validationPath":  "apps/demo-go-color-app/overlays/staging",
					"status":          "True",
					"reason":          "Succeeded",
				},
				"diagnostics": map[string]any{
					"failedTaskCount": 0,
				},
			},
		},
	}

	ev := latestValidationEvidence(items)
	if ev == nil {
		t.Fatal("latestValidationEvidence returned nil")
	}
	if got := validationField(ev, "pipelineRunName"); got != "devops-cp-validate-chg-2026-0049-nd7rm" {
		t.Fatalf("pipelineRunName = %v", got)
	}
	if got := validationField(ev, "tektonNamespace"); got != "devops-ci-staging" {
		t.Fatalf("tektonNamespace = %v", got)
	}
	if got := validationField(ev, "validationPath"); got != "apps/demo-go-color-app/overlays/staging" {
		t.Fatalf("validationPath = %v", got)
	}
	if got := validationField(ev, "failedTaskCount"); got != 0 {
		t.Fatalf("failedTaskCount = %v", got)
	}
	if got := validationField(ev, "summary"); got != "Tekton validation completed without failed TaskRuns" {
		t.Fatalf("summary = %v", got)
	}
}

func TestLatestEvidenceStillPrefersDeploymentEvidence(t *testing.T) {
	items := []map[string]any{
		{
			"name":         "tekton-validation-evidence",
			"evidenceType": "validation",
			"payload": map[string]any{
				"gitops": map[string]any{"validationPath": "apps/demo-go-color-app/overlays/staging"},
			},
		},
		{
			"name":         "deployment-evidence",
			"evidenceType": "deployment",
			"payload": map[string]any{
				"kubernetes": map[string]any{"deployment": map[string]any{"namespace": "devops-ci-staging"}},
			},
		},
	}

	ev := latestEvidence(items)
	if ev == nil {
		t.Fatal("latestEvidence returned nil")
	}
	if got := get(ev, "name"); got != "deployment-evidence" {
		t.Fatalf("latestEvidence selected %v", got)
	}
}
