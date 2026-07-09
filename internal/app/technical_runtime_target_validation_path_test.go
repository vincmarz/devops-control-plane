package app

import "testing"

func TestTechnicalRuntimeTargetResolverCarriesValidationPath(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{
			Name:                  "staging",
			Enabled:               true,
			ClusterName:           "ocp-dev",
			KubernetesNamespace:   "devops-ci-staging",
			TektonNamespace:       "devops-ci-staging",
			ArgoCDApplicationName: "demo-go-color-app-staging",
			GitTargetBranch:       "main",
			ValidationPath:        "apps/demo-go-color-app/overlays/staging",
			AllowTechnicalActions: true,
		},
	}, "staging")
	registry := DefaultClusterRegistryFallback()
	resolver := NewTechnicalRuntimeTargetResolver(NewEnvironmentClusterResolver(catalog, registry), "validate-gitops")

	target, err := resolver.Resolve("staging")
	if err != nil {
		t.Fatalf("Resolve returned error %v", err)
	}
	if target.ValidationPath != "apps/demo-go-color-app/overlays/staging" {
		t.Fatalf("ValidationPath = %q", target.ValidationPath)
	}
}
