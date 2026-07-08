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

func TestLoadDefaultsRuntimeEnablementFlagsToDisabled(t *testing.T) {
	t.Setenv("DCP_RUNTIME_SECRET_LOADER_ENABLED", "")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORIES_ENABLED", "")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_KUBERNETES_ENABLED", "")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_TEKTON_ENABLED", "")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_ARGOCD_ENABLED", "")

	cfg := Load()

	if cfg.RuntimeSecretLoaderEnabled {
		t.Fatal("RuntimeSecretLoaderEnabled default = true, want false")
	}
	if cfg.RuntimeClientFactoriesEnabled {
		t.Fatal("RuntimeClientFactoriesEnabled default = true, want false")
	}
	if cfg.RuntimeClientFactoryKubernetesEnabled {
		t.Fatal("RuntimeClientFactoryKubernetesEnabled default = true, want false")
	}
	if cfg.RuntimeClientFactoryTektonEnabled {
		t.Fatal("RuntimeClientFactoryTektonEnabled default = true, want false")
	}
	if cfg.RuntimeClientFactoryArgoCDEnabled {
		t.Fatal("RuntimeClientFactoryArgoCDEnabled default = true, want false")
	}
}

func TestLoadParsesRuntimeEnablementFlags(t *testing.T) {
	t.Setenv("DCP_RUNTIME_SECRET_LOADER_ENABLED", "true")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORIES_ENABLED", "true")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_KUBERNETES_ENABLED", "true")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_TEKTON_ENABLED", "true")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_ARGOCD_ENABLED", "true")

	cfg := Load()

	if !cfg.RuntimeSecretLoaderEnabled {
		t.Fatal("RuntimeSecretLoaderEnabled = false, want true")
	}
	if !cfg.RuntimeClientFactoriesEnabled {
		t.Fatal("RuntimeClientFactoriesEnabled = false, want true")
	}
	if !cfg.RuntimeClientFactoryKubernetesEnabled {
		t.Fatal("RuntimeClientFactoryKubernetesEnabled = false, want true")
	}
	if !cfg.RuntimeClientFactoryTektonEnabled {
		t.Fatal("RuntimeClientFactoryTektonEnabled = false, want true")
	}
	if !cfg.RuntimeClientFactoryArgoCDEnabled {
		t.Fatal("RuntimeClientFactoryArgoCDEnabled = false, want true")
	}
}

func TestLoadParsesRuntimeEnablementFlagsAsFalse(t *testing.T) {
	t.Setenv("DCP_RUNTIME_SECRET_LOADER_ENABLED", "false")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORIES_ENABLED", "false")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_KUBERNETES_ENABLED", "false")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_TEKTON_ENABLED", "false")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_ARGOCD_ENABLED", "false")

	cfg := Load()

	if cfg.RuntimeSecretLoaderEnabled {
		t.Fatal("RuntimeSecretLoaderEnabled = true, want false")
	}
	if cfg.RuntimeClientFactoriesEnabled {
		t.Fatal("RuntimeClientFactoriesEnabled = true, want false")
	}
	if cfg.RuntimeClientFactoryKubernetesEnabled {
		t.Fatal("RuntimeClientFactoryKubernetesEnabled = true, want false")
	}
	if cfg.RuntimeClientFactoryTektonEnabled {
		t.Fatal("RuntimeClientFactoryTektonEnabled = true, want false")
	}
	if cfg.RuntimeClientFactoryArgoCDEnabled {
		t.Fatal("RuntimeClientFactoryArgoCDEnabled = true, want false")
	}
}

func TestLoadFallsBackToDisabledForInvalidRuntimeEnablementFlags(t *testing.T) {
	t.Setenv("DCP_RUNTIME_SECRET_LOADER_ENABLED", "not-a-bool")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORIES_ENABLED", "not-a-bool")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_KUBERNETES_ENABLED", "not-a-bool")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_TEKTON_ENABLED", "not-a-bool")
	t.Setenv("DCP_RUNTIME_CLIENT_FACTORY_ARGOCD_ENABLED", "not-a-bool")

	cfg := Load()

	if cfg.RuntimeSecretLoaderEnabled {
		t.Fatal("RuntimeSecretLoaderEnabled invalid value fallback = true, want false")
	}
	if cfg.RuntimeClientFactoriesEnabled {
		t.Fatal("RuntimeClientFactoriesEnabled invalid value fallback = true, want false")
	}
	if cfg.RuntimeClientFactoryKubernetesEnabled {
		t.Fatal("RuntimeClientFactoryKubernetesEnabled invalid value fallback = true, want false")
	}
	if cfg.RuntimeClientFactoryTektonEnabled {
		t.Fatal("RuntimeClientFactoryTektonEnabled invalid value fallback = true, want false")
	}
	if cfg.RuntimeClientFactoryArgoCDEnabled {
		t.Fatal("RuntimeClientFactoryArgoCDEnabled invalid value fallback = true, want false")
	}
}
