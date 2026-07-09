package app

import (
	"strings"
	"testing"
)

func simulatedExternalClusterCatalogAndRegistry() (EnvironmentCatalog, ClusterRegistry) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{
			Name:                  "dev",
			DisplayName:           "Development",
			Enabled:               true,
			ClusterName:           "ocp-dev",
			KubernetesNamespace:   "devops-ci-demo",
			TektonNamespace:       "devops-ci-demo",
			ArgoCDApplicationName: "demo-go-color-app",
			GitTargetBranch:       "main",
			ValidationPath:        "apps/demo-go-color-app",
			AllowTechnicalActions: true,
		},
		{
			Name:                  "staging",
			DisplayName:           "Staging",
			Enabled:               true,
			ClusterName:           "ocp-nonprod-simulated",
			KubernetesNamespace:   "devops-ci-staging",
			TektonNamespace:       "devops-ci-staging",
			ArgoCDApplicationName: "demo-go-color-app-staging",
			GitTargetBranch:       "change/CHG-2026-0049",
			ValidationPath:        "apps/demo-go-color-app/overlays/staging",
			AllowTechnicalActions: true,
		},
	}, "dev")

	registry := NewClusterRegistry([]ClusterDefinition{
		{
			Name:              "ocp-dev",
			DisplayName:       "OpenShift Development",
			Enabled:           true,
			APIURL:            "https://api.ocp-dev.example:6443",
			DefaultNamespace:  "devops-ci-demo",
			AllowedNamespaces: []string{"devops-ci-demo"},
		},
		{
			Name:              "ocp-nonprod-simulated",
			DisplayName:       "OpenShift Non Production Simulated",
			Enabled:           true,
			APIURL:            "https://api.ocp-nonprod-simulated.example:6443",
			DefaultNamespace:  "devops-ci-staging",
			AllowedNamespaces: []string{"devops-ci-staging"},
		},
	})

	return catalog, registry
}

func simulatedExternalClusterTarget(t testing.TB) TechnicalRuntimeTarget {
	t.Helper()

	catalog, registry := simulatedExternalClusterCatalogAndRegistry()
	resolver := NewTechnicalRuntimeTargetResolver(
		NewEnvironmentClusterResolver(catalog, registry),
		"validate-gitops",
	)

	target, err := resolver.Resolve("staging")
	if err != nil {
		t.Fatalf("Resolve(staging) returned error %v", err)
	}
	return target
}

func TestMultiClusterReadinessResolvesSimulatedExternalClusterWithoutFallback(t *testing.T) {
	target := simulatedExternalClusterTarget(t)

	if target.TargetEnvironment != "staging" {
		t.Fatalf("TargetEnvironment = %q", target.TargetEnvironment)
	}
	if target.ClusterName != "ocp-nonprod-simulated" {
		t.Fatalf("ClusterName = %q", target.ClusterName)
	}
	if target.ClusterName == "ocp-dev" {
		t.Fatal("staging target silently fell back to ocp-dev")
	}
	if target.KubernetesNamespace != "devops-ci-staging" {
		t.Fatalf("KubernetesNamespace = %q", target.KubernetesNamespace)
	}
	if target.TektonNamespace != "devops-ci-staging" {
		t.Fatalf("TektonNamespace = %q", target.TektonNamespace)
	}
	if target.ArgoCDApplicationName != "demo-go-color-app-staging" {
		t.Fatalf("ArgoCDApplicationName = %q", target.ArgoCDApplicationName)
	}
	if target.ValidationPath != "apps/demo-go-color-app/overlays/staging" {
		t.Fatalf("ValidationPath = %q", target.ValidationPath)
	}
}

func TestMultiClusterReadinessRejectsMissingRuntimeProviderForExternalCluster(t *testing.T) {
	target := simulatedExternalClusterTarget(t)

	_, err := DefaultRuntimeClientProviderRegistry().Select(target)
	if err == nil {
		t.Fatal("Select returned nil error for missing external runtime provider")
	}
	if !strings.Contains(err.Error(), "runtime provider for cluster \"ocp-nonprod-simulated\" is not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMultiClusterReadinessRejectsDisabledRuntimeProviderForExternalCluster(t *testing.T) {
	target := simulatedExternalClusterTarget(t)
	registry := NewRuntimeClientProviderRegistry([]RuntimeClientProvider{
		{
			ClusterName:        "ocp-nonprod-simulated",
			DisplayName:        "Disabled simulated non-production provider",
			Enabled:            false,
			CurrentCluster:     false,
			KubernetesProvider: true,
			TektonProvider:     true,
			ArgoCDProvider:     true,
		},
	})

	_, err := registry.Select(target)
	if err == nil {
		t.Fatal("Select returned nil error for disabled external runtime provider")
	}
	if !strings.Contains(err.Error(), "runtime provider for cluster \"ocp-nonprod-simulated\" is disabled") {
		t.Fatalf("unexpected error: %v", err)
	}
}
