package api

import (
	"testing"

	appsvc "github.com/vincmarz/devops-control-plane/internal/app"
)

func TestGroupApplicationsByEnvironment(t *testing.T) {
	catalog := appsvc.NewEnvironmentCatalog([]appsvc.EnvironmentDefinition{
		{Name: "dev", ApplicationName: "demo-go-color-app", ClusterName: "ocp-dev", KubernetesNamespace: "devops-ci-demo", ArgoCDApplicationName: "demo-go-color-app"},
		{Name: "staging", ApplicationName: "demo-go-color-app", ClusterName: "ocp-dev", KubernetesNamespace: "devops-ci-staging", ArgoCDApplicationName: "demo-go-color-app-staging"},
		{Name: "production", ApplicationName: "demo-go-color-app", ClusterName: "ocp-dev", KubernetesNamespace: "devops-ci-production", ArgoCDApplicationName: "demo-go-color-app-production"},
	}, "dev")
	apps := []map[string]any{
		{"name": "demo-app", "targetNamespace": "devops-ci-demo", "syncStatus": "Synced", "healthStatus": "Healthy"},
		{"name": "demo-go-color-app", "targetNamespace": "devops-ci-demo", "syncStatus": "Synced", "healthStatus": "Healthy"},
		{"name": "demo-go-color-app-production", "targetNamespace": "devops-ci-production", "syncStatus": "Synced", "healthStatus": "Healthy"},
		{"name": "demo-go-color-app-staging", "targetNamespace": "devops-ci-staging", "syncStatus": "Synced", "healthStatus": "Healthy"},
	}

	logical, standalone := groupApplicationsByEnvironment(apps, catalog)
	if len(logical) != 1 {
		t.Fatalf("logical applications = %d, want 1", len(logical))
	}
	if get(logical[0], "name") != "demo-go-color-app" {
		t.Fatalf("logical application name = %v", get(logical[0], "name"))
	}
	environments := get(logical[0], "environments").([]map[string]any)
	if len(environments) != 3 {
		t.Fatalf("environment instances = %d, want 3", len(environments))
	}
	for index, want := range []string{"dev", "staging", "production"} {
		if get(environments[index], "environment") != want {
			t.Fatalf("environment %d = %v, want %s", index, get(environments[index], "environment"), want)
		}
	}
	if len(standalone) != 1 || get(standalone[0], "name") != "demo-app" {
		t.Fatalf("standalone applications = %#v", standalone)
	}
}

func TestGroupApplicationsByEnvironmentDoesNotInferFromSuffixes(t *testing.T) {
	catalog := appsvc.NewEnvironmentCatalog([]appsvc.EnvironmentDefinition{
		{Name: "staging", ApplicationName: "payments", ArgoCDApplicationName: "release-candidate"},
	}, "staging")
	apps := []map[string]any{
		{"name": "payments-staging", "syncStatus": "Synced", "healthStatus": "Healthy"},
		{"name": "release-candidate", "syncStatus": "Synced", "healthStatus": "Healthy"},
	}

	logical, standalone := groupApplicationsByEnvironment(apps, catalog)
	environments := get(logical[0], "environments").([]map[string]any)
	if get(environments[0], "argocdApplicationName") != "release-candidate" {
		t.Fatalf("binding did not use explicit Argo CD application name")
	}
	if len(standalone) != 1 || get(standalone[0], "name") != "payments-staging" {
		t.Fatalf("suffix-based application should remain standalone: %#v", standalone)
	}
}

func TestGroupApplicationsByEnvironmentMarksMissingApplicationNotObserved(t *testing.T) {
	catalog := appsvc.NewEnvironmentCatalog([]appsvc.EnvironmentDefinition{
		{Name: "production", ApplicationName: "demo-go-color-app", KubernetesNamespace: "devops-ci-production", ArgoCDApplicationName: "demo-go-color-app-production"},
	}, "production")

	logical, standalone := groupApplicationsByEnvironment(nil, catalog)
	if len(standalone) != 0 {
		t.Fatalf("standalone applications = %d, want 0", len(standalone))
	}
	environments := get(logical[0], "environments").([]map[string]any)
	if get(environments[0], "observed") != false {
		t.Fatalf("missing application should not be observed")
	}
	if get(environments[0], "healthStatus") != "Not observed" {
		t.Fatalf("health status = %v, want Not observed", get(environments[0], "healthStatus"))
	}
}
