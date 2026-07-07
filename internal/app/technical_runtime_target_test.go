package app

import (
	"strings"
	"testing"
)

func TestTechnicalRuntimeTargetResolverResolvesDevFallback(t *testing.T) {
	resolver := NewTechnicalRuntimeTargetResolver(
		NewEnvironmentClusterResolver(DefaultEnvironmentCatalogFallback(), DefaultClusterRegistryFallback()),
		"validate-gitops",
	)

	target, err := resolver.Resolve("dev")
	if err != nil {
		t.Fatalf("Resolve(dev) returned error %v", err)
	}

	if target.TargetEnvironment != "dev" {
		t.Fatalf("TargetEnvironment = %q, want dev", target.TargetEnvironment)
	}
	if target.EnvironmentName != "dev" {
		t.Fatalf("EnvironmentName = %q, want dev", target.EnvironmentName)
	}
	if target.ClusterName != "ocp-dev" {
		t.Fatalf("ClusterName = %q, want ocp-dev", target.ClusterName)
	}
	if !target.ClusterEnabled {
		t.Fatal("ClusterEnabled = false, want true")
	}
	if target.KubernetesNamespace != "devops-ci-demo" {
		t.Fatalf("KubernetesNamespace = %q, want devops-ci-demo", target.KubernetesNamespace)
	}
	if target.TektonNamespace != "devops-ci-demo" {
		t.Fatalf("TektonNamespace = %q, want devops-ci-demo", target.TektonNamespace)
	}
	if target.TektonPipelineName != "validate-gitops" {
		t.Fatalf("TektonPipelineName = %q, want validate-gitops", target.TektonPipelineName)
	}
	if target.ArgoCDApplicationName != "demo-go-color-app" {
		t.Fatalf("ArgoCDApplicationName = %q, want demo-go-color-app", target.ArgoCDApplicationName)
	}
	if target.GitTargetBranch != "main" {
		t.Fatalf("GitTargetBranch = %q, want main", target.GitTargetBranch)
	}
}

func TestTechnicalRuntimeTargetResolverUsesDefaultEnvironment(t *testing.T) {
	resolver := NewTechnicalRuntimeTargetResolver(
		NewEnvironmentClusterResolver(DefaultEnvironmentCatalogFallback(), DefaultClusterRegistryFallback()),
		"validate-gitops",
	)

	target, err := resolver.Resolve("")
	if err != nil {
		t.Fatalf("Resolve(empty) returned error %v", err)
	}
	if target.EnvironmentName != "dev" {
		t.Fatalf("EnvironmentName = %q, want dev", target.EnvironmentName)
	}
	if target.ClusterName != "ocp-dev" {
		t.Fatalf("ClusterName = %q, want ocp-dev", target.ClusterName)
	}
}

func TestTechnicalRuntimeTargetResolverRejectsDisabledEnvironments(t *testing.T) {
	resolver := NewTechnicalRuntimeTargetResolver(
		NewEnvironmentClusterResolver(DefaultEnvironmentCatalogFallback(), DefaultClusterRegistryFallback()),
		"validate-gitops",
	)

	for _, environment := range []string{"staging", "production"} {
		_, err := resolver.Resolve(environment)
		if err == nil {
			t.Fatalf("Resolve(%s) returned nil error", environment)
		}
		if !strings.Contains(err.Error(), "disabled") {
			t.Fatalf("Resolve(%s) error = %q, want disabled", environment, err.Error())
		}
	}
}

func TestTechnicalRuntimeTargetResolverRejectsUnknownEnvironment(t *testing.T) {
	resolver := NewTechnicalRuntimeTargetResolver(
		NewEnvironmentClusterResolver(DefaultEnvironmentCatalogFallback(), DefaultClusterRegistryFallback()),
		"validate-gitops",
	)

	_, err := resolver.Resolve("unknown-env")
	if err == nil {
		t.Fatal("Resolve(unknown-env) returned nil error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("Resolve(unknown-env) error = %q, want not configured", err.Error())
	}
}

func TestTechnicalRuntimeTargetResolverRejectsMissingTechnicalMetadata(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{Name: "dev", Enabled: true, ClusterName: "ocp-dev", AllowTechnicalActions: true},
	}, "dev")
	resolver := NewTechnicalRuntimeTargetResolver(
		NewEnvironmentClusterResolver(catalog, DefaultClusterRegistryFallback()),
		"validate-gitops",
	)

	_, err := resolver.Resolve("dev")
	if err == nil {
		t.Fatal("Resolve(dev) returned nil error for missing technical metadata")
	}
	if !strings.Contains(err.Error(), "kubernetesNamespace") {
		t.Fatalf("Resolve(dev) error = %q, want kubernetesNamespace", err.Error())
	}
}

func TestTechnicalRuntimeTargetValidateRequiresPipelineName(t *testing.T) {
	target := TechnicalRuntimeTarget{
		TargetEnvironment:     "dev",
		EnvironmentName:       "dev",
		ClusterName:           "ocp-dev",
		ClusterEnabled:        true,
		KubernetesNamespace:   "devops-ci-demo",
		TektonNamespace:       "devops-ci-demo",
		ArgoCDApplicationName: "demo-go-color-app",
		GitTargetBranch:       "main",
	}

	err := target.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for missing TektonPipelineName")
	}
	if !strings.Contains(err.Error(), "tekton pipeline name") {
		t.Fatalf("Validate error = %q, want tekton pipeline name", err.Error())
	}
}
