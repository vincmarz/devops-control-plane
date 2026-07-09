package app

import "testing"

func TestParseEnvironmentCatalogYAMLLoadsValidationPath(t *testing.T) {
	content := []byte(`defaultEnvironment: staging
environments:
  - name: staging
    enabled: true
    allowTechnicalActions: true
    validationPath: apps/demo-go-color-app/overlays/staging
`)

	catalog, err := ParseEnvironmentCatalogYAML(content)
	if err != nil {
		t.Fatalf("ParseEnvironmentCatalogYAML returned error %v", err)
	}

	environment, ok := catalog.Resolve("staging")
	if !ok {
		t.Fatal("staging environment was not resolved")
	}
	if environment.ValidationPath != "apps/demo-go-color-app/overlays/staging" {
		t.Fatalf("ValidationPath = %q", environment.ValidationPath)
	}
}
