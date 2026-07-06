package app

import "testing"

func TestEnvironmentClusterResolverResolvesDefaultEnvironment(t *testing.T) {
	resolver := NewEnvironmentClusterResolver(DefaultEnvironmentCatalogFallback(), DefaultClusterRegistryFallback())

	resolution, err := resolver.Resolve("")
	if err != nil {
		t.Fatalf("Resolve returned error %v", err)
	}
	if resolution.TargetEnvironment != "dev" {
		t.Fatalf("TargetEnvironment = %q, want dev", resolution.TargetEnvironment)
	}
	if resolution.Environment.ClusterName != "ocp-dev" {
		t.Fatalf("Environment.ClusterName = %q, want ocp-dev", resolution.Environment.ClusterName)
	}
	if resolution.Cluster.Name != "ocp-dev" {
		t.Fatalf("Cluster.Name = %q, want ocp-dev", resolution.Cluster.Name)
	}
	if !resolution.Cluster.Enabled {
		t.Fatal("ocp-dev should be enabled")
	}
}

func TestEnvironmentClusterResolverResolvesConfiguredDisabledEnvironments(t *testing.T) {
	resolver := NewEnvironmentClusterResolver(DefaultEnvironmentCatalogFallback(), DefaultClusterRegistryFallback())

	staging, err := resolver.Resolve("staging")
	if err != nil {
		t.Fatalf("Resolve(staging) returned error %v", err)
	}
	if staging.Environment.Enabled {
		t.Fatal("staging environment should remain disabled")
	}
	if staging.Cluster.Name != "ocp-staging" {
		t.Fatalf("staging Cluster.Name = %q, want ocp-staging", staging.Cluster.Name)
	}
	if staging.Cluster.Enabled {
		t.Fatal("ocp-staging cluster should remain disabled")
	}

	production, err := resolver.Resolve("production")
	if err != nil {
		t.Fatalf("Resolve(production) returned error %v", err)
	}
	if production.Environment.Enabled {
		t.Fatal("production environment should remain disabled")
	}
	if production.Cluster.Name != "ocp-production" {
		t.Fatalf("production Cluster.Name = %q, want ocp-production", production.Cluster.Name)
	}
	if production.Cluster.Enabled {
		t.Fatal("ocp-production cluster should remain disabled")
	}
}

func TestEnvironmentClusterResolverRejectsUnknownEnvironment(t *testing.T) {
	resolver := NewEnvironmentClusterResolver(DefaultEnvironmentCatalogFallback(), DefaultClusterRegistryFallback())

	_, err := resolver.Resolve("unknown-env")
	if err == nil {
		t.Fatal("Resolve returned nil error for unknown-env")
	}
	if got, want := err.Error(), `targetEnvironment "unknown-env" is not configured`; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestEnvironmentClusterResolverRejectsMissingClusterName(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{Name: "dev", Enabled: true, ClusterName: ""},
	}, "dev")
	resolver := NewEnvironmentClusterResolver(catalog, DefaultClusterRegistryFallback())

	_, err := resolver.Resolve("dev")
	if err == nil {
		t.Fatal("Resolve returned nil error for missing clusterName")
	}
	if got, want := err.Error(), `targetEnvironment "dev" does not define clusterName`; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestEnvironmentClusterResolverRejectsUnknownClusterName(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{Name: "dev", Enabled: true, ClusterName: "missing-cluster"},
	}, "dev")
	resolver := NewEnvironmentClusterResolver(catalog, DefaultClusterRegistryFallback())

	_, err := resolver.Resolve("dev")
	if err == nil {
		t.Fatal("Resolve returned nil error for unknown clusterName")
	}
	if got, want := err.Error(), `targetEnvironment "dev" references clusterName "missing-cluster" which is not configured`; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestEnvironmentClusterResolverEnabledTargetPreservesDisabledEnvironmentBehavior(t *testing.T) {
	resolver := NewEnvironmentClusterResolver(DefaultEnvironmentCatalogFallback(), DefaultClusterRegistryFallback())

	resolution, err := resolver.ResolveEnabledTarget("dev")
	if err != nil {
		t.Fatalf("ResolveEnabledTarget(dev) returned error %v", err)
	}
	if resolution.Cluster.Name != "ocp-dev" {
		t.Fatalf("Cluster.Name = %q, want ocp-dev", resolution.Cluster.Name)
	}

	_, err = resolver.ResolveEnabledTarget("staging")
	if err == nil {
		t.Fatal("ResolveEnabledTarget(staging) returned nil error")
	}
	if got, want := err.Error(), `targetEnvironment "staging" is currently disabled`; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestEnvironmentClusterResolverTechnicalActionTarget(t *testing.T) {
	resolver := NewEnvironmentClusterResolver(DefaultEnvironmentCatalogFallback(), DefaultClusterRegistryFallback())

	resolution, err := resolver.ResolveTechnicalActionTarget("dev")
	if err != nil {
		t.Fatalf("ResolveTechnicalActionTarget(dev) returned error %v", err)
	}
	if resolution.Cluster.Name != "ocp-dev" {
		t.Fatalf("Cluster.Name = %q, want ocp-dev", resolution.Cluster.Name)
	}
}
