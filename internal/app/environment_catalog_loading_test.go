package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvironmentCatalogYAML(t *testing.T) {
	content := []byte(`defaultEnvironment: staging
environments:
  - name: dev
    displayName: Development
    enabled: true
    allowTechnicalActions: true
  - name: staging
    applicationName: demo-go-color-app
    displayName: Staging
    argocdApplicationName: demo-go-color-app-staging
    enabled: true
    allowTechnicalActions: true
  - name: production
    displayName: Production
    enabled: false
    allowTechnicalActions: false
`)

	catalog, err := ParseEnvironmentCatalogYAML(content)
	if err != nil {
		t.Fatalf("ParseEnvironmentCatalogYAML returned error %v", err)
	}
	if got := catalog.DefaultEnvironment(); got != "staging" {
		t.Fatalf("DefaultEnvironment() = %q, want staging", got)
	}
	if err := catalog.ValidateCreateTargetEnvironment("staging"); err != nil {
		t.Fatalf("staging should be enabled from runtime catalog: %v", err)
	}
	staging, ok := catalog.Resolve("staging")
	if !ok {
		t.Fatal("staging environment was not resolved")
	}
	if staging.ApplicationName != "demo-go-color-app" {
		t.Fatalf("ApplicationName = %q, want demo-go-color-app", staging.ApplicationName)
	}
	if staging.ArgoCDApplicationName != "demo-go-color-app-staging" {
		t.Fatalf("ArgoCDApplicationName = %q, want demo-go-color-app-staging", staging.ArgoCDApplicationName)
	}
	if catalog.AllowsTechnicalActions("production") {
		t.Fatal("production should not allow technical actions from runtime catalog")
	}
}

func TestParseEnvironmentCatalogYAMLRejectsMissingDefault(t *testing.T) {
	content := []byte(`defaultEnvironment: staging
environments:
  - name: dev
    enabled: true
`)

	_, err := ParseEnvironmentCatalogYAML(content)
	if err == nil {
		t.Fatal("ParseEnvironmentCatalogYAML returned nil error, want missing default environment error")
	}
}

func TestDefaultEnvironmentCatalogFallsBackWhenFileMissing(t *testing.T) {
	t.Setenv(environmentCatalogFileEnv, filepath.Join(t.TempDir(), "missing-environments.yaml"))

	catalog := DefaultEnvironmentCatalog()
	if got := catalog.DefaultEnvironment(); got != "dev" {
		t.Fatalf("DefaultEnvironment() = %q, want dev fallback", got)
	}
	if err := catalog.ValidateCreateTargetEnvironment("dev"); err != nil {
		t.Fatalf("dev should be enabled by fallback catalog: %v", err)
	}
	if err := catalog.ValidateCreateTargetEnvironment("staging"); err == nil {
		t.Fatal("staging should remain disabled by fallback catalog")
	}
}

func TestDefaultEnvironmentCatalogLoadsFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "environments.yaml")
	content := []byte(`defaultEnvironment: staging
environments:
  - name: dev
    enabled: true
    allowTechnicalActions: true
  - name: staging
    enabled: true
    allowTechnicalActions: true
  - name: production
    enabled: false
    allowTechnicalActions: false
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error %v", err)
	}
	t.Setenv(environmentCatalogFileEnv, path)

	catalog := DefaultEnvironmentCatalog()
	if got := catalog.DefaultEnvironment(); got != "staging" {
		t.Fatalf("DefaultEnvironment() = %q, want staging", got)
	}
	if err := catalog.ValidateCreateTargetEnvironment("staging"); err != nil {
		t.Fatalf("staging should be enabled from mounted catalog file: %v", err)
	}
}
