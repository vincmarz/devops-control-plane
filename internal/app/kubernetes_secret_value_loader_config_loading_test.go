package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadKubernetesSecretValueLoaderConfigFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "allowed-refs.yaml")
	content := []byte(`allowedClusters:
  - OCP-STAGING
allowedRefs:
  - clusterName: OCP-STAGING
    namespace: devops-control-plane
    name: staging-runtime-client
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	config, err := LoadKubernetesSecretValueLoaderConfigFromFile(path)
	if err != nil {
		t.Fatalf("LoadKubernetesSecretValueLoaderConfigFromFile returned error %v", err)
	}
	if len(config.AllowedClusters) != 1 || config.AllowedClusters[0] != "ocp-staging" {
		t.Fatalf("AllowedClusters = %#v, want [ocp-staging]", config.AllowedClusters)
	}
	if len(config.AllowedRefs) != 1 {
		t.Fatalf("AllowedRefs length = %d, want 1", len(config.AllowedRefs))
	}
	ref := config.AllowedRefs[0]
	if ref.ClusterName != "ocp-staging" || ref.Namespace != "devops-control-plane" || ref.Name != "staging-runtime-client" {
		t.Fatalf("unexpected allowed ref: %#v", ref)
	}
}

func TestLoadKubernetesSecretValueLoaderConfigFromFileRejectsEmptyPath(t *testing.T) {
	_, err := LoadKubernetesSecretValueLoaderConfigFromFile("")
	if err == nil {
		t.Fatal("LoadKubernetesSecretValueLoaderConfigFromFile returned nil error for empty path")
	}
}

func TestLoadKubernetesSecretValueLoaderConfigFromFileReturnsReadError(t *testing.T) {
	_, err := LoadKubernetesSecretValueLoaderConfigFromFile(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("LoadKubernetesSecretValueLoaderConfigFromFile returned nil error for missing file")
	}
}

func TestLoadKubernetesSecretValueLoaderConfigFromFileRejectsInvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "allowed-refs.yaml")
	if err := os.WriteFile(path, []byte("allowedRefs: ["), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	_, err := LoadKubernetesSecretValueLoaderConfigFromFile(path)
	if err == nil {
		t.Fatal("LoadKubernetesSecretValueLoaderConfigFromFile returned nil error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Fatalf("error = %q, want failed to parse", err.Error())
	}
}

func TestLoadKubernetesSecretValueLoaderConfigFromFileRejectsSecretLikeValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "allowed-refs.yaml")
	content := []byte(`allowedRefs:
  - clusterName: ocp-staging
    namespace: devops-control-plane
    name: "Bearer abc"
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	_, err := LoadKubernetesSecretValueLoaderConfigFromFile(path)
	if err == nil {
		t.Fatal("LoadKubernetesSecretValueLoaderConfigFromFile returned nil error for secret-like value")
	}
	if !strings.Contains(err.Error(), "secret value") {
		t.Fatalf("error = %q, want secret value", err.Error())
	}
}

func TestLoadKubernetesSecretValueLoaderConfigFromFileAllowsEmptyDocumentFailClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "allowed-refs.yaml")
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	config, err := LoadKubernetesSecretValueLoaderConfigFromFile(path)
	if err != nil {
		t.Fatalf("LoadKubernetesSecretValueLoaderConfigFromFile returned error %v", err)
	}
	if len(config.AllowedClusters) != 0 {
		t.Fatalf("AllowedClusters length = %d, want 0", len(config.AllowedClusters))
	}
	if len(config.AllowedRefs) != 0 {
		t.Fatalf("AllowedRefs length = %d, want 0", len(config.AllowedRefs))
	}

	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-staging",
		Kubernetes: &RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "staging-runtime-client",
			TokenKey:  "token",
		},
	}
	if err := ValidateKubernetesSecretValueLoaderRequest(config, refs); err == nil {
		t.Fatal("ValidateKubernetesSecretValueLoaderRequest returned nil error for empty allow-list")
	}
}

func TestNewKubernetesSecretValueLoaderConfigRejectsIncompleteAllowedRef(t *testing.T) {
	_, err := NewKubernetesSecretValueLoaderConfig(nil, []KubernetesSecretValueLoaderAllowedRef{
		{ClusterName: "ocp-staging", Namespace: "devops-control-plane"},
	})
	if err == nil {
		t.Fatal("NewKubernetesSecretValueLoaderConfig returned nil error for incomplete ref")
	}
}
