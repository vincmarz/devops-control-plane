package app

import (
	"context"
	"errors"
	"fmt"
)

// TektonRuntimePipelineRunRequest is the application-layer request used to
// start a Tekton validation PipelineRun.
//
// The request is intentionally independent from the concrete Tekton adapter
// package. Adapter-specific request objects are mapped in the composition root
// or in future provider implementations.
type TektonRuntimePipelineRunRequest struct {
	Namespace          string
	PipelineName       string
	GenerateName       string
	ChangeNumber       string
	ApplicationName    string
	GitURL             string
	GitRevision        string
	ValidationPath     string
	ServiceAccountName string
	Image              string
	WorkspacePVC       string
	DockerConfigSecret string
}

// TektonRuntimePipelineRunRef is the application-layer reference returned after
// creating a Tekton PipelineRun.
type TektonRuntimePipelineRunRef struct {
	Name      string
	Namespace string
	UID       string
}

// TektonRuntimeClient is the minimal Tekton runtime capability required by the
// ChangeService validation workflows.
//
// The interface does not import the concrete Tekton adapter. This keeps the app
// layer provider-aware while preserving adapter isolation.
type TektonRuntimeClient interface {
	CreatePipelineRun(ctx context.Context, request TektonRuntimePipelineRunRequest) (TektonRuntimePipelineRunRef, error)
	FindLatestPipelineRunByChange(ctx context.Context, namespace string, changeNumber string) (TektonValidationResult, error)
	ListTaskRunsByPipelineRun(ctx context.Context, namespace string, pipelineRunName string) ([]TektonTaskRunResult, error)
}

// TektonRuntimeClientProvider resolves the Tekton runtime client associated
// with a RuntimeClientProviderSelection.
//
// In phase 15.8.5.2 only the current-cluster provider is supported. Real
// multi-cluster client construction is intentionally deferred.
type TektonRuntimeClientProvider interface {
	ResolveTektonRuntimeClient(ctx context.Context, selection RuntimeClientProviderSelection) (TektonRuntimeClient, error)
}

// TektonRuntimeClientProviderRegistry selects a Tekton runtime client provider
// by runtime provider cluster name.
type TektonRuntimeClientProviderRegistry struct {
	providers map[string]TektonRuntimeClientProvider
}

// NewTektonRuntimeClientProviderRegistry builds a Tekton provider registry from
// explicit provider mappings.
func NewTektonRuntimeClientProviderRegistry(providers map[string]TektonRuntimeClientProvider) TektonRuntimeClientProviderRegistry {
	registry := TektonRuntimeClientProviderRegistry{providers: map[string]TektonRuntimeClientProvider{}}
	for clusterName, provider := range providers {
		name := normalizeRuntimeProviderClusterName(clusterName)
		if name == "" || provider == nil {
			continue
		}
		registry.providers[name] = provider
	}
	return registry
}

// EmptyTektonRuntimeClientProviderRegistry returns an empty registry.
func EmptyTektonRuntimeClientProviderRegistry() TektonRuntimeClientProviderRegistry {
	return TektonRuntimeClientProviderRegistry{providers: map[string]TektonRuntimeClientProvider{}}
}

// DefaultTektonRuntimeClientProviderRegistry returns the conservative baseline
// registry used before real multi-cluster Tekton clients are introduced.
//
// If currentClient is nil, the registry is empty. This preserves existing
// behavior for deployments where Tekton integration is disabled.
func DefaultTektonRuntimeClientProviderRegistry(currentClient TektonRuntimeClient) TektonRuntimeClientProviderRegistry {
	if currentClient == nil {
		return EmptyTektonRuntimeClientProviderRegistry()
	}
	return NewTektonRuntimeClientProviderRegistry(map[string]TektonRuntimeClientProvider{
		"ocp-dev": CurrentClusterTektonRuntimeClientProvider(currentClient),
	})
}

// CurrentClusterTektonRuntimeClientProvider creates a provider returning the
// currently configured Tekton runtime client.
func CurrentClusterTektonRuntimeClientProvider(client TektonRuntimeClient) TektonRuntimeClientProvider {
	return currentClusterTektonRuntimeClientProvider{client: client}
}

type currentClusterTektonRuntimeClientProvider struct {
	client TektonRuntimeClient
}

func (p currentClusterTektonRuntimeClientProvider) ResolveTektonRuntimeClient(ctx context.Context, selection RuntimeClientProviderSelection) (TektonRuntimeClient, error) {
	if p.client == nil {
		return nil, errors.New("current Tekton runtime client is not configured")
	}
	if selection.Provider.ClusterName == "" {
		return nil, errors.New("runtime provider cluster name is required for Tekton client selection")
	}
	if !selection.Provider.CurrentCluster {
		return nil, fmt.Errorf("runtime provider for cluster %q is not a current-cluster Tekton provider", selection.Provider.ClusterName)
	}
	if !selection.Provider.TektonProvider {
		return nil, fmt.Errorf("runtime provider for cluster %q does not expose Tekton capability", selection.Provider.ClusterName)
	}
	return p.client, nil
}

// Resolve selects the Tekton runtime client for a provider selection.
func (r TektonRuntimeClientProviderRegistry) Resolve(ctx context.Context, selection RuntimeClientProviderSelection) (TektonRuntimeClient, error) {
	clusterName := normalizeRuntimeProviderClusterName(selection.Provider.ClusterName)
	if clusterName == "" {
		return nil, errors.New("runtime provider cluster name is required for Tekton client selection")
	}
	if r.providers == nil {
		return nil, fmt.Errorf("Tekton runtime client provider for cluster %q is not configured", clusterName)
	}
	provider, ok := r.providers[clusterName]
	if !ok {
		return nil, fmt.Errorf("Tekton runtime client provider for cluster %q is not configured", clusterName)
	}
	return provider.ResolveTektonRuntimeClient(ctx, selection)
}
