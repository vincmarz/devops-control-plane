package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesExplicitKubernetesTokenWhenProvided(t *testing.T) {
	t.Setenv("KUBERNETES_TOKEN", "explicit-token")
	t.Setenv("KUBERNETES_TOKEN_FILE", filepath.Join(t.TempDir(), "missing-token"))

	cfg := Load()

	if cfg.KubernetesToken != "explicit-token" {
		t.Fatalf("expected explicit Kubernetes token to win")
	}
}

func TestLoadFallsBackToServiceAccountTokenFile(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "token")
	if err := os.WriteFile(tokenFile, []byte("service-account-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("KUBERNETES_TOKEN", "")
	t.Setenv("KUBERNETES_TOKEN_FILE", tokenFile)

	cfg := Load()

	if cfg.KubernetesToken != "service-account-token" {
		t.Fatalf("expected token from service account file, got %q", cfg.KubernetesToken)
	}
}

func TestLoadBuildsKubernetesAPIURLFromServiceEnvironment(t *testing.T) {
	t.Setenv("KUBERNETES_API_URL", "")
	t.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
	t.Setenv("KUBERNETES_SERVICE_PORT", "6443")

	cfg := Load()

	if cfg.KubernetesAPIURL != "https://10.0.0.1:6443" {
		t.Fatalf("unexpected Kubernetes API URL: %q", cfg.KubernetesAPIURL)
	}
}

func TestLoadDefaultsKubernetesAPIURLPortTo443(t *testing.T) {
	t.Setenv("KUBERNETES_API_URL", "")
	t.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
	t.Setenv("KUBERNETES_SERVICE_PORT", "")

	cfg := Load()

	if cfg.KubernetesAPIURL != "https://10.0.0.1:443" {
		t.Fatalf("unexpected Kubernetes API URL: %q", cfg.KubernetesAPIURL)
	}
}

func TestLoadUsesServiceAccountCAFileWhenPresent(t *testing.T) {
	dir := t.TempDir()
	caFile := filepath.Join(dir, "ca.crt")
	if err := os.WriteFile(caFile, []byte("ca"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("KUBERNETES_CA_FILE", "")
	t.Setenv("KUBERNETES_SERVICEACCOUNT_CA_FILE", caFile)

	cfg := Load()

	if cfg.KubernetesCAFile != caFile {
		t.Fatalf("unexpected Kubernetes CA file: %q", cfg.KubernetesCAFile)
	}
}
