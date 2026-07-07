package app

import (
	"context"
	"errors"
	"fmt"
)

// KubernetesRuntimeEvidenceClient is the minimal runtime client capability used
// by ChangeService to collect Kubernetes/OpenShift runtime evidence.
//
// The interface intentionally mirrors the existing adapter method shape without
// importing the concrete Kubernetes adapter into the application layer.
type KubernetesRuntimeEvidenceClient interface {
	CollectRuntimeEvidence(ctx context.Context, namespace string, applicationName string) (map[string]any, error)
}

// KubernetesRuntimeClientProvider resolves the Kubernetes runtime evidence
// client associated with a RuntimeClientProviderSelection.
//
// Implementations must not log or expose credentials. In phase 15.8.4.2 only
// the current-cluster provider is supported.
type KubernetesRuntimeClientProvider interface {
	ResolveKubernetesRuntimeEvidenceClient(ctx context.Context, selection RuntimeClientProviderSelection) (KubernetesRuntimeEvidenceClient, error)
}

// KubernetesRuntimeClientProviderRegistry selects a Kubernetes runtime client
// provider by runtime provider cluster name.
type KubernetesRuntimeClientProviderRegistry struct {
	providers map[string]KubernetesRuntimeClientProvider
}

// NewKubernetesRuntimeClientProviderRegistry builds a Kubernetes provider
// registry from explicit provider mappings.
func NewKubernetesRuntimeClientProviderRegistry(providers map[string]KubernetesRuntimeClientProvider) KubernetesRuntimeClientProviderRegistry {
	registry := KubernetesRuntimeClientProviderRegistry{providers: map[string]KubernetesRuntimeClientProvider{}}
	for clusterName, provider := range providers {
		name := normalizeRuntimeProviderClusterName(clusterName)
		if name == "" || provider == nil {
			continue
		}
		registry.providers[name] = provider
	}
	return registry
}

// EmptyKubernetesRuntimeClientProviderRegistry returns an empty registry.
func EmptyKubernetesRuntimeClientProviderRegistry() KubernetesRuntimeClientProviderRegistry {
	return KubernetesRuntimeClientProviderRegistry{providers: map[string]KubernetesRuntimeClientProvider{}}
}

// DefaultKubernetesRuntimeClientProviderRegistry returns the conservative
// baseline registry used before real multi-cluster Kubernetes clients are
// introduced.
//
// If currentClient is nil, the registry is empty. This preserves existing
// behavior for deployments where Kubernetes runtime evidence is disabled.
func DefaultKubernetesRuntimeClientProviderRegistry(currentClient KubernetesRuntimeEvidenceClient) KubernetesRuntimeClientProviderRegistry {
	if currentClient == nil {
		return EmptyKubernetesRuntimeClientProviderRegistry()
	}
	return NewKubernetesRuntimeClientProviderRegistry(map[string]KubernetesRuntimeClientProvider{
		"ocp-dev": CurrentClusterKubernetesRuntimeClientProvider(currentClient),
	})
}

// CurrentClusterKubernetesRuntimeClientProvider creates a provider that returns
// the currently configured Kubernetes runtime evidence client.
func CurrentClusterKubernetesRuntimeClientProvider(client KubernetesRuntimeEvidenceClient) KubernetesRuntimeClientProvider {
	return currentClusterKubernetesRuntimeClientProvider{client: client}
}

type currentClusterKubernetesRuntimeClientProvider struct {
	client KubernetesRuntimeEvidenceClient
}

func (p currentClusterKubernetesRuntimeClientProvider) ResolveKubernetesRuntimeEvidenceClient(ctx context.Context, selection RuntimeClientProviderSelection) (KubernetesRuntimeEvidenceClient, error) {
	if p.client == nil {
		return nil, errors.New("current Kubernetes runtime evidence client is not configured")
	}
	if selection.Provider.ClusterName == "" {
		return nil, errors.New("runtime provider cluster name is required for Kubernetes client selection")
	}
	if !selection.Provider.CurrentCluster {
		return nil, fmt.Errorf("runtime provider for cluster %q is not a current-cluster Kubernetes provider", selection.Provider.ClusterName)
	}
	if !selection.Provider.KubernetesProvider {
		return nil, fmt.Errorf("runtime provider for cluster %q does not expose Kubernetes capability", selection.Provider.ClusterName)
	}
	return p.client, nil
}

// Resolve selects the Kubernetes runtime evidence client for a provider
// selection.
func (r KubernetesRuntimeClientProviderRegistry) Resolve(ctx context.Context, selection RuntimeClientProviderSelection) (KubernetesRuntimeEvidenceClient, error) {
	clusterName := normalizeRuntimeProviderClusterName(selection.Provider.ClusterName)
	if clusterName == "" {
		return nil, errors.New("runtime provider cluster name is required for Kubernetes client selection")
	}
	if r.providers == nil {
		return nil, fmt.Errorf("Kubernetes runtime client provider for cluster %q is not configured", clusterName)
	}
	provider, ok := r.providers[clusterName]
	if !ok {
		return nil, fmt.Errorf("Kubernetes runtime client provider for cluster %q is not configured", clusterName)
	}
	return provider.ResolveKubernetesRuntimeEvidenceClient(ctx, selection)
}
