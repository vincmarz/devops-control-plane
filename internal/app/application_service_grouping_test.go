package app

import (
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func TestApplicationServiceGroupByEnvironment(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{Name: "dev", ApplicationName: "demo-go-color-app", ClusterName: "ocp-dev", KubernetesNamespace: "devops-ci-demo", ArgoCDApplicationName: "demo-go-color-app"},
		{Name: "staging", ApplicationName: "demo-go-color-app", ClusterName: "ocp-dev", KubernetesNamespace: "devops-ci-staging", ArgoCDApplicationName: "demo-go-color-app-staging"},
		{Name: "production", ApplicationName: "demo-go-color-app", ClusterName: "ocp-dev", KubernetesNamespace: "devops-ci-production", ArgoCDApplicationName: "demo-go-color-app-production"},
	}, "dev")
	applications := []domain.Application{
		{Name: "demo-app", TargetNamespace: "devops-ci-demo", SyncStatus: "Synced", HealthStatus: "Healthy"},
		{Name: "demo-go-color-app", TargetNamespace: "devops-ci-demo", SyncStatus: "Synced", HealthStatus: "Healthy"},
		{Name: "demo-go-color-app-production", TargetNamespace: "devops-ci-production", SyncStatus: "Synced", HealthStatus: "Healthy"},
		{Name: "demo-go-color-app-staging", TargetNamespace: "devops-ci-staging", SyncStatus: "Synced", HealthStatus: "Healthy"},
	}

	grouping := NewApplicationService().GroupByEnvironment(applications, catalog)
	if len(grouping.LogicalApplications) != 1 {
		t.Fatalf("logical applications = %d, want 1", len(grouping.LogicalApplications))
	}
	if grouping.LogicalApplications[0].Name != "demo-go-color-app" {
		t.Fatalf("logical application name = %v", grouping.LogicalApplications[0].Name)
	}
	environments := grouping.LogicalApplications[0].Environments
	if len(environments) != 3 {
		t.Fatalf("environment instances = %d, want 3", len(environments))
	}
	for index, want := range []string{"dev", "staging", "production"} {
		if environments[index].Environment != want {
			t.Fatalf("environment %d = %v, want %s", index, environments[index].Environment, want)
		}
	}
	if len(grouping.StandaloneApplications) != 1 || grouping.StandaloneApplications[0].Name != "demo-app" {
		t.Fatalf("standalone applications = %#v", grouping.StandaloneApplications)
	}
}

func TestApplicationServiceGroupByEnvironmentDoesNotInferFromSuffixes(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{Name: "staging", ApplicationName: "payments", ArgoCDApplicationName: "release-candidate"},
	}, "staging")
	applications := []domain.Application{
		{Name: "payments-staging", SyncStatus: "Synced", HealthStatus: "Healthy"},
		{Name: "release-candidate", SyncStatus: "Synced", HealthStatus: "Healthy"},
	}

	grouping := NewApplicationService().GroupByEnvironment(applications, catalog)
	environments := grouping.LogicalApplications[0].Environments
	if environments[0].ArgoCDApplicationName != "release-candidate" {
		t.Fatal("binding did not use explicit Argo CD application name")
	}
	if len(grouping.StandaloneApplications) != 1 || grouping.StandaloneApplications[0].Name != "payments-staging" {
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
	environments := grouping.LogicalApplications[0].Environments
	if environments[0].Observed != false {
		t.Fatal("missing application should not be observed")
	}
	if environments[0].HealthStatus != "Not observed" {
		t.Fatalf("health status = %v, want Not observed", environments[0].HealthStatus)
	}
}
