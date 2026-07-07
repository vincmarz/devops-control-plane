package app

import (
	"context"
	"errors"
	"fmt"
)

// ArgoCDRuntimeClient is the minimal Argo CD runtime capability required by
// ChangeService deployment check workflows.
//
// The interface is intentionally independent from the concrete Argo CD adapter.
// Adapter-specific response models are mapped in the composition root or in
// future provider implementations.
type ArgoCDRuntimeClient interface {
	CheckDeployment(ctx context.Context, applicationName string) (ArgoCDDeploymentResult, error)
}

// ArgoCDRuntimeClientProvider resolves the Argo CD runtime client associated
// with a RuntimeClientProviderSelection.
//
// In phase 15.8.6.2 only the current-cluster provider is supported. Real
// multi-cluster Argo CD client construction is intentionally deferred.
type ArgoCDRuntimeClientProvider interface {
	ResolveArgoCDRuntimeClient(ctx context.Context, selection RuntimeClientProviderSelection) (ArgoCDRuntimeClient, error)
}

// ArgoCDRuntimeClientProviderRegistry selects an Argo CD runtime client provider
// by runtime provider cluster name.
type ArgoCDRuntimeClientProviderRegistry struct {
	providers map[string]ArgoCDRuntimeClientProvider
}

// NewArgoCDRuntimeClientProviderRegistry builds an Argo CD provider registry
// from explicit provider mappings.
func NewArgoCDRuntimeClientProviderRegistry(providers map[string]ArgoCDRuntimeClientProvider) ArgoCDRuntimeClientProviderRegistry {
	registry := ArgoCDRuntimeClientProviderRegistry{providers: map[string]ArgoCDRuntimeClientProvider{}}
	for clusterName, provider := range providers {
		name := normalizeRuntimeProviderClusterName(clusterName)
		if name == "" || provider == nil {
			continue
		}
		registry.providers[name] = provider
	}
	return registry
}

// EmptyArgoCDRuntimeClientProviderRegistry returns an empty registry.
func EmptyArgoCDRuntimeClientProviderRegistry() ArgoCDRuntimeClientProviderRegistry {
	return ArgoCDRuntimeClientProviderRegistry{providers: map[string]ArgoCDRuntimeClientProvider{}}
}

// DefaultArgoCDRuntimeClientProviderRegistry returns the conservative baseline
// registry used before real multi-cluster Argo CD clients are introduced.
//
// If currentClient is nil, the registry is empty. This preserves existing
// behavior for deployments where Argo CD integration is disabled.
func DefaultArgoCDRuntimeClientProviderRegistry(currentClient ArgoCDRuntimeClient) ArgoCDRuntimeClientProviderRegistry {
	if currentClient == nil {
		return EmptyArgoCDRuntimeClientProviderRegistry()
	}
	return NewArgoCDRuntimeClientProviderRegistry(map[string]ArgoCDRuntimeClientProvider{
		"ocp-dev": CurrentClusterArgoCDRuntimeClientProvider(currentClient),
	})
}

// CurrentClusterArgoCDRuntimeClientProvider creates a provider returning the
// currently configured Argo CD runtime client.
func CurrentClusterArgoCDRuntimeClientProvider(client ArgoCDRuntimeClient) ArgoCDRuntimeClientProvider {
	return currentClusterArgoCDRuntimeClientProvider{client: client}
}

type currentClusterArgoCDRuntimeClientProvider struct {
	client ArgoCDRuntimeClient
}

func (p currentClusterArgoCDRuntimeClientProvider) ResolveArgoCDRuntimeClient(ctx context.Context, selection RuntimeClientProviderSelection) (ArgoCDRuntimeClient, error) {
	if p.client == nil {
		return nil, errors.New("current Argo CD runtime client is not configured")
	}
	if selection.Provider.ClusterName == "" {
		return nil, errors.New("runtime provider cluster name is required for Argo CD client selection")
	}
	if !selection.Provider.CurrentCluster {
		return nil, fmt.Errorf("runtime provider for cluster %q is not a current-cluster Argo CD provider", selection.Provider.ClusterName)
	}
	if !selection.Provider.ArgoCDProvider {
		return nil, fmt.Errorf("runtime provider for cluster %q does not expose Argo CD capability", selection.Provider.ClusterName)
	}
	return p.client, nil
}

// Resolve selects the Argo CD runtime client for a provider selection.
func (r ArgoCDRuntimeClientProviderRegistry) Resolve(ctx context.Context, selection RuntimeClientProviderSelection) (ArgoCDRuntimeClient, error) {
	clusterName := normalizeRuntimeProviderClusterName(selection.Provider.ClusterName)
	if clusterName == "" {
		return nil, errors.New("runtime provider cluster name is required for Argo CD client selection")
	}
	if r.providers == nil {
		return nil, fmt.Errorf("Argo CD runtime client provider for cluster %q is not configured", clusterName)
	}
	provider, ok := r.providers[clusterName]
	if !ok {
		return nil, fmt.Errorf("Argo CD runtime client provider for cluster %q is not configured", clusterName)
	}
	return provider.ResolveArgoCDRuntimeClient(ctx, selection)
}
