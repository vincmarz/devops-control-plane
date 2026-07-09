package app

import (
	"strings"
	"testing"
)

type simulatedClusterTargetCase struct {
	environmentName     string
	clusterName         string
	kubernetesNamespace string
	tektonNamespace     string
	argocdApplication   string
	gitTargetBranch     string
	validationPath      string
}

func simulatedStagingProductionCatalogAndRegistry() (EnvironmentCatalog, ClusterRegistry) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{
			Name:                  "dev",
			DisplayName:           "Development",
			Enabled:               true,
			ClusterName:           "ocp-dev",
			KubernetesNamespace:   "devops-ci-demo",
			TektonNamespace:       "devops-ci-demo",
			ArgoCDApplicationName: "demo-go-color-app",
			GitTargetBranch:       "main",
			ValidationPath:        "apps/demo-go-color-app",
			AllowTechnicalActions: true,
		},
		{
			Name:                  "staging",
			DisplayName:           "Staging",
			Enabled:               true,
			ClusterName:           "ocp-staging-simulated",
			KubernetesNamespace:   "devops-ci-staging",
			TektonNamespace:       "devops-ci-staging",
			ArgoCDApplicationName: "demo-go-color-app-staging",
			GitTargetBranch:       "change/CHG-2026-0049",
			ValidationPath:        "apps/demo-go-color-app/overlays/staging",
			AllowTechnicalActions: true,
		},
		{
			Name:                  "production",
			DisplayName:           "Production",
			Enabled:               true,
			ClusterName:           "ocp-production-simulated",
			KubernetesNamespace:   "devops-ci-production",
			TektonNamespace:       "devops-ci-production",
			ArgoCDApplicationName: "demo-go-color-app-production",
			GitTargetBranch:       "change/CHG-2026-0050",
			ValidationPath:        "apps/demo-go-color-app/overlays/production",
			AllowTechnicalActions: true,
		},
	}, "dev")

	registry := NewClusterRegistry([]ClusterDefinition{
		{
			Name:              "ocp-dev",
			DisplayName:       "OpenShift Development",
			Enabled:           true,
			APIURL:            "https://api.ocp-dev.example:6443",
			DefaultNamespace:  "devops-ci-demo",
			AllowedNamespaces: []string{"devops-ci-demo"},
		},
		{
			Name:              "ocp-staging-simulated",
			DisplayName:       "OpenShift Staging Simulated",
			Enabled:           true,
			APIURL:            "https://api.ocp-staging-simulated.example:6443",
			DefaultNamespace:  "devops-ci-staging",
			AllowedNamespaces: []string{"devops-ci-staging"},
		},
		{
			Name:              "ocp-production-simulated",
			DisplayName:       "OpenShift Production Simulated",
			Enabled:           true,
			APIURL:            "https://api.ocp-production-simulated.example:6443",
			DefaultNamespace:  "devops-ci-production",
			AllowedNamespaces: []string{"devops-ci-production"},
		},
	})

	return catalog, registry
}

func resolveSimulatedStagingProductionTarget(t testing.TB, environmentName string) TechnicalRuntimeTarget {
	t.Helper()

	catalog, registry := simulatedStagingProductionCatalogAndRegistry()
	resolver := NewTechnicalRuntimeTargetResolver(
		NewEnvironmentClusterResolver(catalog, registry),
		"validate-gitops",
	)

	target, err := resolver.Resolve(environmentName)
	if err != nil {
		t.Fatalf("Resolve(%s) returned error %v", environmentName, err)
	}
	return target
}

func simulatedStagingProductionCases() []simulatedClusterTargetCase {
	return []simulatedClusterTargetCase{
		{
			environmentName:     "staging",
			clusterName:         "ocp-staging-simulated",
			kubernetesNamespace: "devops-ci-staging",
			tektonNamespace:     "devops-ci-staging",
			argocdApplication:   "demo-go-color-app-staging",
			gitTargetBranch:     "change/CHG-2026-0049",
			validationPath:      "apps/demo-go-color-app/overlays/staging",
		},
		{
			environmentName:     "production",
			clusterName:         "ocp-production-simulated",
			kubernetesNamespace: "devops-ci-production",
			tektonNamespace:     "devops-ci-production",
			argocdApplication:   "demo-go-color-app-production",
			gitTargetBranch:     "change/CHG-2026-0050",
			validationPath:      "apps/demo-go-color-app/overlays/production",
		},
	}
}

func TestMultiClusterReadinessResolvesSimulatedStagingAndProductionClusters(t *testing.T) {
	for _, tc := range simulatedStagingProductionCases() {
		t.Run(tc.environmentName, func(t *testing.T) {
			target := resolveSimulatedStagingProductionTarget(t, tc.environmentName)

			if target.TargetEnvironment != tc.environmentName {
				t.Fatalf("TargetEnvironment = %q", target.TargetEnvironment)
			}
			if target.ClusterName != tc.clusterName {
				t.Fatalf("ClusterName = %q", target.ClusterName)
			}
			if target.ClusterName == "ocp-dev" {
				t.Fatalf("%s target silently fell back to ocp-dev", tc.environmentName)
			}
			if target.KubernetesNamespace != tc.kubernetesNamespace {
				t.Fatalf("KubernetesNamespace = %q", target.KubernetesNamespace)
			}
			if target.TektonNamespace != tc.tektonNamespace {
				t.Fatalf("TektonNamespace = %q", target.TektonNamespace)
			}
			if target.ArgoCDApplicationName != tc.argocdApplication {
				t.Fatalf("ArgoCDApplicationName = %q", target.ArgoCDApplicationName)
			}
			if target.GitTargetBranch != tc.gitTargetBranch {
				t.Fatalf("GitTargetBranch = %q", target.GitTargetBranch)
			}
			if target.ValidationPath != tc.validationPath {
				t.Fatalf("ValidationPath = %q", target.ValidationPath)
			}
		})
	}
}

func TestMultiClusterReadinessRejectsMissingRuntimeProvidersForSimulatedStagingAndProduction(t *testing.T) {
	for _, tc := range simulatedStagingProductionCases() {
		t.Run(tc.environmentName, func(t *testing.T) {
			target := resolveSimulatedStagingProductionTarget(t, tc.environmentName)

			_, err := DefaultRuntimeClientProviderRegistry().Select(target)
			if err == nil {
				t.Fatal("Select returned nil error for missing runtime provider")
			}
			expected := "runtime provider for cluster \"" + tc.clusterName + "\" is not configured"
			if !strings.Contains(err.Error(), expected) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMultiClusterReadinessRejectsDisabledRuntimeProvidersForSimulatedStagingAndProduction(t *testing.T) {
	for _, tc := range simulatedStagingProductionCases() {
		t.Run(tc.environmentName, func(t *testing.T) {
			target := resolveSimulatedStagingProductionTarget(t, tc.environmentName)
			registry := NewRuntimeClientProviderRegistry([]RuntimeClientProvider{
				{
					ClusterName:        tc.clusterName,
					DisplayName:        "Disabled simulated provider",
					Enabled:            false,
					CurrentCluster:     false,
					KubernetesProvider: true,
					TektonProvider:     true,
					ArgoCDProvider:     true,
				},
			})

			_, err := registry.Select(target)
			if err == nil {
				t.Fatal("Select returned nil error for disabled runtime provider")
			}
			expected := "runtime provider for cluster \"" + tc.clusterName + "\" is disabled"
			if !strings.Contains(err.Error(), expected) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
