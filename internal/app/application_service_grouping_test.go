package app

import "testing"

func TestApplicationServiceGroupByEnvironment(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{Name: "dev", ApplicationName: "demo-go-color-app", ClusterName: "ocp-dev", KubernetesNamespace: "devops-ci-demo", ArgoCDApplicationName: "demo-go-color-app"},
		{Name: "staging", ApplicationName: "demo-go-color-app", ClusterName: "ocp-dev", KubernetesNamespace: "devops-ci-staging", ArgoCDApplicationName: "demo-go-color-app-staging"},
		{Name: "production", ApplicationName: "demo-go-color-app", ClusterName: "ocp-dev", KubernetesNamespace: "devops-ci-production", ArgoCDApplicationName: "demo-go-color-app-production"},
	}, "dev")
	applications := []map[string]any{
		{"name": "demo-app", "targetNamespace": "devops-ci-demo", "syncStatus": "Synced", "healthStatus": "Healthy"},
		{"name": "demo-go-color-app", "targetNamespace": "devops-ci-demo", "syncStatus": "Synced", "healthStatus": "Healthy"},
		{"name": "demo-go-color-app-production", "targetNamespace": "devops-ci-production", "syncStatus": "Synced", "healthStatus": "Healthy"},
		{"name": "demo-go-color-app-staging", "targetNamespace": "devops-ci-staging", "syncStatus": "Synced", "healthStatus": "Healthy"},
	}

	grouping := NewApplicationService().GroupByEnvironment(applications, catalog)
	if len(grouping.LogicalApplications) != 1 {
		t.Fatalf("logical applications = %d, want 1", len(grouping.LogicalApplications))
	}
	if grouping.LogicalApplications[0]["name"] != "demo-go-color-app" {
		t.Fatalf("logical application name = %v", grouping.LogicalApplications[0]["name"])
	}
	environments := grouping.LogicalApplications[0]["environments"].([]map[string]any)
	if len(environments) != 3 {
		t.Fatalf("environment instances = %d, want 3", len(environments))
	}
	for index, want := range []string{"dev", "staging", "production"} {
		if environments[index]["environment"] != want {
			t.Fatalf("environment %d = %v, want %s", index, environments[index]["environment"], want)
		}
	}
	if len(grouping.StandaloneApplications) != 1 || grouping.StandaloneApplications[0]["name"] != "demo-app" {
		t.Fatalf("standalone applications = %#v", grouping.StandaloneApplications)
	}
}

func TestApplicationServiceGroupByEnvironmentDoesNotInferFromSuffixes(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{Name: "staging", ApplicationName: "payments", ArgoCDApplicationName: "release-candidate"},
	}, "staging")
	applications := []map[string]any{
		{"name": "payments-staging", "syncStatus": "Synced", "healthStatus": "Healthy"},
		{"name": "release-candidate", "syncStatus": "Synced", "healthStatus": "Healthy"},
	}

	grouping := NewApplicationService().GroupByEnvironment(applications, catalog)
	environments := grouping.LogicalApplications[0]["environments"].([]map[string]any)
	if environments[0]["argocdApplicationName"] != "release-candidate" {
		t.Fatal("binding did not use explicit Argo CD application name")
	}
	if len(grouping.StandaloneApplications) != 1 || grouping.StandaloneApplications[0]["name"] != "payments-staging" {
		t.Fatalf("suffix-based application should remain standalone: %#v", grouping.StandaloneApplications)
	}
}

func TestApplicationServiceGroupByEnvironmentMarksMissingApplicationNotObserved(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{Name: "production", ApplicationName: "demo-go-color-app", KubernetesNamespace: "devops-ci-production", ArgoCDApplicationName: "demo-go-color-app-production"},
	}, "production")

	grouping := NewApplicationService().GroupByEnvironment(nil, catalog)
	if len(grouping.StandaloneApplications) != 0 {
		t.Fatalf("standalone applications = %d, want 0", len(grouping.StandaloneApplications))
	}
	environments := grouping.LogicalApplications[0]["environments"].([]map[string]any)
	if environments[0]["observed"] != false {
		t.Fatal("missing application should not be observed")
	}
	if environments[0]["healthStatus"] != "Not observed" {
		t.Fatalf("health status = %v, want Not observed", environments[0]["healthStatus"])
	}
}
