package config

import "testing"

// TestLoadDefaultsInsecureTLSFlagsToFalse encodes the security invariant that
// TLS certificate verification is ON by default for every runtime integration:
// the *_INSECURE_TLS flags must only become true when explicitly configured.
func TestLoadDefaultsInsecureTLSFlagsToFalse(t *testing.T) {
	t.Setenv("ARGOCD_INSECURE_TLS", "")
	t.Setenv("GITLAB_INSECURE_TLS", "")
	t.Setenv("KUBERNETES_INSECURE_TLS", "")

	cfg := Load()

	if cfg.ArgoCDInsecureTLS {
		t.Error("ArgoCDInsecureTLS must default to false (TLS verification on)")
	}
	if cfg.GitLabInsecureTLS {
		t.Error("GitLabInsecureTLS must default to false (TLS verification on)")
	}
	if cfg.KubernetesInsecureTLS {
		t.Error("KubernetesInsecureTLS must default to false (TLS verification on)")
	}
}

// TestLoadParsesInsecureTLSFlagsWhenExplicitlyEnabled verifies the opt-in path:
// the flags become true only when the operator explicitly sets them.
func TestLoadParsesInsecureTLSFlagsWhenExplicitlyEnabled(t *testing.T) {
	t.Setenv("ARGOCD_INSECURE_TLS", "true")
	t.Setenv("GITLAB_INSECURE_TLS", "true")
	t.Setenv("KUBERNETES_INSECURE_TLS", "true")

	cfg := Load()

	if !cfg.ArgoCDInsecureTLS {
		t.Error("ArgoCDInsecureTLS should be true when ARGOCD_INSECURE_TLS=true")
	}
	if !cfg.GitLabInsecureTLS {
		t.Error("GitLabInsecureTLS should be true when GITLAB_INSECURE_TLS=true")
	}
	if !cfg.KubernetesInsecureTLS {
		t.Error("KubernetesInsecureTLS should be true when KUBERNETES_INSECURE_TLS=true")
	}
}
