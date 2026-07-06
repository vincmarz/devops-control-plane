package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultClusterRegistryFallback(t *testing.T) {
	registry := DefaultClusterRegistryFallback()

	if !registry.IsEnabled("ocp-dev") {
		t.Fatal("ocp-dev should be enabled in fallback registry")
	}
	if registry.IsEnabled("ocp-staging") {
		t.Fatal("ocp-staging should be disabled in fallback registry")
	}
	if registry.IsEnabled("ocp-production") {
		t.Fatal("ocp-production should be disabled in fallback registry")
	}
}

func TestParseClusterRegistryYAML(t *testing.T) {
	content := []byte(`clusters:
  - name: ocp-dev
    displayName: OpenShift Development
    enabled: true
    apiURL: https://api.dev.example:6443
    caConfigMapRef: dcp-cluster-ocp-dev-ca
    tokenSecretRef: dcp-cluster-ocp-dev-token
    defaultNamespace: devops-ci-demo
    allowedNamespaces:
      - devops-ci-demo
  - name: ocp-staging
    displayName: OpenShift Staging
    enabled: false
    apiURL: ""
    caConfigMapRef: dcp-cluster-ocp-staging-ca
    tokenSecretRef: dcp-cluster-ocp-staging-token
    defaultNamespace: devops-ci-staging
    allowedNamespaces:
      - devops-ci-staging
`)

	registry, err := ParseClusterRegistryYAML(content)
	if err != nil {
		t.Fatalf("ParseClusterRegistryYAML returned error %v", err)
	}
	dev, ok := registry.Resolve("ocp-dev")
	if !ok {
		t.Fatal("ocp-dev should be configured")
	}
	if dev.APIURL != "https://api.dev.example:6443" {
		t.Fatalf("ocp-dev apiURL = %q", dev.APIURL)
	}
	if registry.IsEnabled("ocp-staging") {
		t.Fatal("ocp-staging should be disabled")
	}
}

func TestDefaultClusterRegistryLoadsFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clusters.yaml")
	content := []byte(`clusters:
  - name: ocp-dev
    enabled: true
    apiURL: https://api.dev.example:6443
    defaultNamespace: devops-ci-demo
    allowedNamespaces:
      - devops-ci-demo
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error %v", err)
	}
	t.Setenv(clusterRegistryFileEnv, path)

	registry := DefaultClusterRegistry()
	if !registry.IsEnabled("ocp-dev") {
		t.Fatal("ocp-dev should be enabled from mounted registry file")
	}
}

func TestDefaultClusterRegistryFallsBackWhenFileMissing(t *testing.T) {
	t.Setenv(clusterRegistryFileEnv, filepath.Join(t.TempDir(), "missing-clusters.yaml"))

	registry := DefaultClusterRegistry()
	if !registry.IsEnabled("ocp-dev") {
		t.Fatal("ocp-dev should be enabled by fallback registry")
	}
}

func TestClusterRegistryValidatesEnvironmentCatalogClusterReferences(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{Name: "dev", Enabled: true, ClusterName: "ocp-dev"},
		{Name: "staging", Enabled: false, ClusterName: "ocp-staging"},
		{Name: "production", Enabled: false, ClusterName: "ocp-production"},
	}, "dev")
	registry := DefaultClusterRegistryFallback()

	if err := registry.ValidateEnvironmentCatalog(catalog); err != nil {
		t.Fatalf("ValidateEnvironmentCatalog returned error %v", err)
	}
}

func TestClusterRegistryRejectsUnknownEnvironmentClusterReference(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{Name: "dev", Enabled: true, ClusterName: "missing-cluster"},
	}, "dev")
	registry := DefaultClusterRegistryFallback()

	if err := registry.ValidateEnvironmentCatalog(catalog); err == nil {
		t.Fatal("ValidateEnvironmentCatalog returned nil error for an unknown cluster reference")
	}
}
