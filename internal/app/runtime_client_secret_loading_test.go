package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRuntimeClientSecretRefsRegistryFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret-refs.yaml")
	content := []byte(`clusters:
  - clusterName: ocp-staging
    kubernetes:
      namespace: devops-control-plane
      name: ocp-staging-kube
      tokenKey: token
      caKey: ca.crt
    tekton:
      namespace: devops-control-plane
      name: ocp-staging-kube
      tokenKey: token
      caKey: ca.crt
    argocd:
      namespace: devops-control-plane
      name: ocp-staging-argocd
      tokenKey: token
      baseURLKey: baseURL
      caKey: ca.crt
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	registry, err := LoadRuntimeClientSecretRefsRegistryFromFile(path)
	if err != nil {
		t.Fatalf("LoadRuntimeClientSecretRefsRegistryFromFile returned error %v", err)
	}

	refs, ok := registry.Resolve("ocp-staging")
	if !ok {
		t.Fatal("Resolve(ocp-staging) returned ok=false")
	}
	if refs.ClusterName != "ocp-staging" {
		t.Fatalf("ClusterName = %q, want ocp-staging", refs.ClusterName)
	}
	if refs.Kubernetes == nil || refs.Kubernetes.Name != "ocp-staging-kube" {
		t.Fatalf("unexpected kubernetes refs: %+v", refs.Kubernetes)
	}
	if refs.ArgoCD == nil || refs.ArgoCD.BaseURLKey != "baseURL" {
		t.Fatalf("unexpected argocd refs: %+v", refs.ArgoCD)
	}
}

func TestLoadRuntimeClientSecretRefsRegistryRejectsSecretValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret-refs.yaml")
	content := []byte(`clusters:
  - clusterName: ocp-staging
    kubernetes:
      namespace: devops-control-plane
      name: ocp-staging-kube
      tokenKey: "Bearer abc"
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	_, err := LoadRuntimeClientSecretRefsRegistryFromFile(path)
	if err == nil {
		t.Fatal("LoadRuntimeClientSecretRefsRegistryFromFile returned nil error")
	}
	if !strings.Contains(err.Error(), "secret value") {
		t.Fatalf("error = %q, want secret value", err.Error())
	}
}

func TestDefaultRuntimeClientSecretRefsRegistryFallsBackWhenFileMissing(t *testing.T) {
	t.Setenv("DCP_RUNTIME_CLIENT_SECRET_REFS_FILE", filepath.Join(t.TempDir(), "missing.yaml"))

	registry := DefaultRuntimeClientSecretRefsRegistry()
	if _, ok := registry.Resolve("ocp-staging"); ok {
		t.Fatal("Resolve(ocp-staging) returned ok=true for missing file fallback")
	}
}

func TestDefaultRuntimeClientSecretRefsRegistryLoadsConfiguredFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret-refs.yaml")
	content := []byte(`clusters:
  - clusterName: OCP-STAGING
    kubernetes:
      namespace: devops-control-plane
      name: ocp-staging-kube
      tokenKey: token
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	t.Setenv("DCP_RUNTIME_CLIENT_SECRET_REFS_FILE", path)

	registry := DefaultRuntimeClientSecretRefsRegistry()
	refs, ok := registry.Resolve("ocp-staging")
	if !ok {
		t.Fatal("Resolve(ocp-staging) returned ok=false")
	}
	if refs.ClusterName != "ocp-staging" {
		t.Fatalf("ClusterName = %q, want ocp-staging", refs.ClusterName)
	}
}

func TestRuntimeClientSecretRefsRegistrySafeSummary(t *testing.T) {
	registry, err := NewRuntimeClientSecretRefsRegistry([]RuntimeClientSecretRefs{
		{
			ClusterName: "ocp-staging",
			Kubernetes:  &RuntimeSecretReference{Namespace: "devops-control-plane", Name: "ocp-staging-kube", TokenKey: "token"},
		},
	})
	if err != nil {
		t.Fatalf("NewRuntimeClientSecretRefsRegistry returned error %v", err)
	}

	summary := registry.SafeSummary()
	if len(summary) != 1 {
		t.Fatalf("summary length = %d, want 1", len(summary))
	}
	if summary[0]["clusterName"] != "ocp-staging" {
		t.Fatalf("clusterName = %v, want ocp-staging", summary[0]["clusterName"])
	}
}
