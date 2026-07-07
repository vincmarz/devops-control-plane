package app

import (
	"errors"
	"fmt"
	"strings"
)

// RuntimeClientProvider describes an application-level runtime provider entry.
//
// This abstraction is intentionally metadata-only in phase 15.7.6.1. It does
// not expose concrete Kubernetes, Tekton or Argo CD clients yet. The purpose is
// to introduce a safe selection boundary that can later be extended with real
// per-cluster clients.
type RuntimeClientProvider struct {
	ClusterName        string
	DisplayName        string
	Enabled            bool
	CurrentCluster     bool
	KubernetesProvider bool
	TektonProvider     bool
	ArgoCDProvider     bool
}

// RuntimeClientProviderSelection is the result of selecting a runtime provider
// for a resolved TechnicalRuntimeTarget.
type RuntimeClientProviderSelection struct {
	Target               TechnicalRuntimeTarget
	Provider             RuntimeClientProvider
	SecretRefsConfigured bool
	SecretRefs           RuntimeClientSecretRefs
}

// RuntimeClientProviderRegistry contains the configured runtime client provider
// entries keyed by cluster name.
type RuntimeClientProviderRegistry struct {
	providers map[string]RuntimeClientProvider
}

// NewRuntimeClientProviderRegistry returns a provider registry from explicit
// provider definitions.
func NewRuntimeClientProviderRegistry(providers []RuntimeClientProvider) RuntimeClientProviderRegistry {
	registry := RuntimeClientProviderRegistry{providers: map[string]RuntimeClientProvider{}}
	for _, provider := range providers {
		name := normalizeRuntimeProviderClusterName(provider.ClusterName)
		if name == "" {
			continue
		}
		provider.ClusterName = name
		if strings.TrimSpace(provider.DisplayName) == "" {
			provider.DisplayName = name
		}
		registry.providers[name] = provider
	}
	return registry
}

// DefaultRuntimeClientProviderRegistry returns the conservative baseline used
// before real multi-cluster client secrets are introduced.
//
// Only ocp-dev maps to the current runtime provider. Staging and production do
// not receive client providers in this phase.
func DefaultRuntimeClientProviderRegistry() RuntimeClientProviderRegistry {
	return NewRuntimeClientProviderRegistry([]RuntimeClientProvider{
		CurrentClusterRuntimeProvider("ocp-dev", "Current OpenShift runtime provider"),
	})
}

// CurrentClusterRuntimeProvider returns a provider entry that points to the
// currently configured runtime clients.
func CurrentClusterRuntimeProvider(clusterName string, displayName string) RuntimeClientProvider {
	name := normalizeRuntimeProviderClusterName(clusterName)
	return RuntimeClientProvider{
		ClusterName:        name,
		DisplayName:        strings.TrimSpace(displayName),
		Enabled:            true,
		CurrentCluster:     true,
		KubernetesProvider: true,
		TektonProvider:     true,
		ArgoCDProvider:     true,
	}
}

// Resolve selects a provider by cluster name.
func (r RuntimeClientProviderRegistry) Resolve(clusterName string) (RuntimeClientProvider, error) {
	name := normalizeRuntimeProviderClusterName(clusterName)
	if name == "" {
		return RuntimeClientProvider{}, errors.New("runtime provider cluster name is required")
	}
	provider, ok := r.providers[name]
	if !ok {
		return RuntimeClientProvider{}, fmt.Errorf("runtime provider for cluster %q is not configured", name)
	}
	if !provider.Enabled {
		return RuntimeClientProvider{}, fmt.Errorf("runtime provider for cluster %q is disabled", name)
	}
	return provider, nil
}

// Select resolves the runtime provider for a TechnicalRuntimeTarget.
func (r RuntimeClientProviderRegistry) Select(target TechnicalRuntimeTarget) (RuntimeClientProviderSelection, error) {
	if err := target.Validate(); err != nil {
		return RuntimeClientProviderSelection{}, err
	}
	provider, err := r.Resolve(target.ClusterName)
	if err != nil {
		return RuntimeClientProviderSelection{}, err
	}
	return RuntimeClientProviderSelection{Target: target, Provider: provider}, nil
}

// SelectWithSecretRefs resolves the runtime provider and enriches the selection
// with optional Secret references for the selected provider cluster.
//
// This method does not read Kubernetes Secret values. It only attaches validated
// references loaded by RuntimeClientSecretRefsRegistry.
func (r RuntimeClientProviderRegistry) SelectWithSecretRefs(target TechnicalRuntimeTarget, refsRegistry RuntimeClientSecretRefsRegistry) (RuntimeClientProviderSelection, error) {
	selection, err := r.Select(target)
	if err != nil {
		return RuntimeClientProviderSelection{}, err
	}
	refs, ok := refsRegistry.Resolve(selection.Provider.ClusterName)
	if ok {
		selection.SecretRefsConfigured = true
		selection.SecretRefs = refs
	}
	return selection, nil
}

// SafeSummary returns a non-sensitive summary of the provider selection.
func (s RuntimeClientProviderSelection) SafeSummary() map[string]any {
	summary := map[string]any{
		"targetEnvironment":    s.Target.TargetEnvironment,
		"clusterName":          s.Target.ClusterName,
		"providerClusterName":  s.Provider.ClusterName,
		"providerDisplayName":  s.Provider.DisplayName,
		"providerCurrent":      s.Provider.CurrentCluster,
		"secretRefsConfigured": s.SecretRefsConfigured,
	}
	if s.SecretRefsConfigured {
		summary["secretRefs"] = s.SecretRefs.SafeSummary()
	}
	return summary
}

func normalizeRuntimeProviderClusterName(clusterName string) string {
	return strings.ToLower(strings.TrimSpace(clusterName))
}
