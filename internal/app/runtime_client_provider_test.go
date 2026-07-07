package app

import (
	"strings"
	"testing"
)

func TestDefaultRuntimeClientProviderRegistryResolvesOCPDev(t *testing.T) {
	registry := DefaultRuntimeClientProviderRegistry()

	provider, err := registry.Resolve("ocp-dev")
	if err != nil {
		t.Fatalf("Resolve(ocp-dev) returned error %v", err)
	}

	if provider.ClusterName != "ocp-dev" {
		t.Fatalf("ClusterName = %q, want ocp-dev", provider.ClusterName)
	}
	if !provider.Enabled {
		t.Fatal("Enabled = false, want true")
	}
	if !provider.CurrentCluster {
		t.Fatal("CurrentCluster = false, want true")
	}
	if !provider.KubernetesProvider || !provider.TektonProvider || !provider.ArgoCDProvider {
		t.Fatalf("provider capabilities = kubernetes:%v tekton:%v argocd:%v, want all true", provider.KubernetesProvider, provider.TektonProvider, provider.ArgoCDProvider)
	}
}

func TestDefaultRuntimeClientProviderRegistryRejectsProviderNotConfigured(t *testing.T) {
	registry := DefaultRuntimeClientProviderRegistry()

	_, err := registry.Resolve("ocp-staging")
	if err == nil {
		t.Fatal("Resolve(ocp-staging) returned nil error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("Resolve(ocp-staging) error = %q, want not configured", err.Error())
	}
}

func TestRuntimeClientProviderRegistryRejectsDisabledProvider(t *testing.T) {
	registry := NewRuntimeClientProviderRegistry([]RuntimeClientProvider{
		{ClusterName: "ocp-staging", Enabled: false},
	})

	_, err := registry.Resolve("ocp-staging")
	if err == nil {
		t.Fatal("Resolve(ocp-staging) returned nil error")
	}
	if !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("Resolve(ocp-staging) error = %q, want disabled", err.Error())
	}
}

func TestRuntimeClientProviderRegistrySelectsProviderForTechnicalRuntimeTarget(t *testing.T) {
	target := TechnicalRuntimeTarget{
		TargetEnvironment:     "dev",
		EnvironmentName:       "dev",
		ClusterName:           "ocp-dev",
		ClusterEnabled:        true,
		KubernetesNamespace:   "devops-ci-demo",
		TektonNamespace:       "devops-ci-demo",
		TektonPipelineName:    "validate-gitops",
		ArgoCDApplicationName: "demo-go-color-app",
		GitTargetBranch:       "main",
	}

	selection, err := DefaultRuntimeClientProviderRegistry().Select(target)
	if err != nil {
		t.Fatalf("Select(target) returned error %v", err)
	}

	if selection.Target.ClusterName != "ocp-dev" {
		t.Fatalf("selection target cluster = %q, want ocp-dev", selection.Target.ClusterName)
	}
	if selection.Provider.ClusterName != "ocp-dev" {
		t.Fatalf("selection provider cluster = %q, want ocp-dev", selection.Provider.ClusterName)
	}
	if !selection.Provider.CurrentCluster {
		t.Fatal("selection provider CurrentCluster = false, want true")
	}
}

func TestRuntimeClientProviderRegistrySelectRejectsInvalidTarget(t *testing.T) {
	target := TechnicalRuntimeTarget{ClusterName: "ocp-dev"}

	_, err := DefaultRuntimeClientProviderRegistry().Select(target)
	if err == nil {
		t.Fatal("Select(invalid target) returned nil error")
	}
	if !strings.Contains(err.Error(), "target environment") {
		t.Fatalf("Select(invalid target) error = %q, want target environment", err.Error())
	}
}

func TestRuntimeClientProviderClusterNameIsNormalized(t *testing.T) {
	registry := NewRuntimeClientProviderRegistry([]RuntimeClientProvider{
		CurrentClusterRuntimeProvider(" OCP-DEV ", "Current provider"),
	})

	provider, err := registry.Resolve("ocp-dev")
	if err != nil {
		t.Fatalf("Resolve(ocp-dev) returned error %v", err)
	}
	if provider.ClusterName != "ocp-dev" {
		t.Fatalf("ClusterName = %q, want ocp-dev", provider.ClusterName)
	}
}
