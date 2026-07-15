package config

import "testing"

func TestLoadGitHubConfiguration(t *testing.T) {
	t.Setenv("GITHUB_API_URL", "https://github.example/api/v3")
	t.Setenv("GITHUB_TOKEN", "token")
	t.Setenv("GITHUB_TIMEOUT_SECONDS", "45")
	t.Setenv("GITHUB_INSECURE_TLS", "true")
	t.Setenv("GITHUB_CA_FILE", "/tmp/github-ca.crt")
	cfg := Load()
	if cfg.GitHubAPIURL != "https://github.example/api/v3" || cfg.GitHubToken != "token" || cfg.GitHubTimeoutSeconds != 45 || !cfg.GitHubInsecureTLS || cfg.GitHubCAFile != "/tmp/github-ca.crt" {
		t.Fatalf("GitHub config = %#v", cfg)
	}
}

func TestLoadDefaultsGitHubConfiguration(t *testing.T) {
	t.Setenv("GITHUB_API_URL", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GITHUB_TIMEOUT_SECONDS", "")
	t.Setenv("GITHUB_INSECURE_TLS", "")
	t.Setenv("GITHUB_CA_FILE", "")
	cfg := Load()
	if cfg.GitHubAPIURL != "https://api.github.com" || cfg.GitHubTimeoutSeconds != 30 || cfg.GitHubInsecureTLS {
		t.Fatalf("GitHub defaults = %#v", cfg)
	}
}
