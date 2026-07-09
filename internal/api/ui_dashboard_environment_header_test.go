package api

import "testing"

func TestPreferredChangeFallsBackToMostRecentWhenPreferredIsEmpty(t *testing.T) {
	changes := []map[string]any{
		{"changeNumber": "CHG-2026-0050"},
		{"changeNumber": "CHG-2026-0049"},
	}

	selected := preferredChange(changes, "")
	if got := changeNumberOrID(selected); got != "CHG-2026-0050" {
		t.Fatalf("selected change = %q", got)
	}
}

func TestEnvironmentSummariesIncludesFallbackDev(t *testing.T) {
	summaries := environmentSummaries()
	if len(summaries) == 0 {
		t.Fatal("environmentSummaries returned no environments")
	}
	if got := summaries[0]["name"]; got != "dev" {
		t.Fatalf("first environment name = %v", got)
	}
	if got := summaries[0]["kubernetesNamespace"]; got == "" {
		t.Fatal("first environment kubernetesNamespace is empty")
	}
}
