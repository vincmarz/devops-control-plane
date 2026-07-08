package app

import "context"

// KubernetesRuntimeClientProviderFactoryAwareRegistry combines the existing
// KubernetesRuntimeClientProviderRegistry with an optional
// KubernetesRuntimeClientFactory and RuntimeSecretValueLoader.
//
// This factory-aware wrapper prepares the path from current-cluster providers
// to real multi-cluster clients built from Secret references.
//
// The default configuration wires EmptyKubernetesRuntimeClientFactory and
// EmptyRuntimeSecretValueLoader so the factory fallback path always fails
// safely with an explicit "not configured" error, until real implementations
// are wired in a later phase.
type KubernetesRuntimeClientProviderFactoryAwareRegistry struct {
	registry KubernetesRuntimeClientProviderRegistry
	factory  KubernetesRuntimeClientFactory
	loader   RuntimeSecretValueLoader
}

// NewKubernetesRuntimeClientProviderFactoryAwareRegistry builds a factory-aware
// wrapper. Nil factory is normalized to EmptyKubernetesRuntimeClientFactory,
// nil loader is normalized to EmptyRuntimeSecretValueLoader.
func NewKubernetesRuntimeClientProviderFactoryAwareRegistry(
	registry KubernetesRuntimeClientProviderRegistry,
	factory KubernetesRuntimeClientFactory,
	loader RuntimeSecretValueLoader,
) KubernetesRuntimeClientProviderFactoryAwareRegistry {
	if factory == nil {
		factory = EmptyKubernetesRuntimeClientFactory{}
	}
	if loader == nil {
		loader = EmptyRuntimeSecretValueLoader{}
	}
	return KubernetesRuntimeClientProviderFactoryAwareRegistry{
		registry: registry,
		factory:  factory,
		loader:   loader,
	}
}

// Resolve returns a KubernetesRuntimeEvidenceClient for the given selection.
//
// It first delegates to the existing provider registry, which handles the
// current-cluster provider for ocp-dev. If the registry cannot resolve a
// client and the selection carries configured Secret references, it attempts
// to build a client through the runtime Secret value loader and the Kubernetes
// runtime client factory. With the conservative default placeholders, that
// fallback path always fails with an explicit "not configured" error.
func (r KubernetesRuntimeClientProviderFactoryAwareRegistry) Resolve(
	ctx context.Context,
	selection RuntimeClientProviderSelection,
) (KubernetesRuntimeEvidenceClient, error) {
	client, err := r.registry.Resolve(ctx, selection)
	if err == nil {
		return client, nil
	}
	if !selection.SecretRefsConfigured {
		return nil, err
	}
	values, loaderErr := r.loader.LoadRuntimeSecretValues(ctx, selection.SecretRefs)
	if loaderErr != nil {
		return nil, loaderErr
	}
	request := KubernetesRuntimeClientFactoryRequest{
		Target:       selection.Target,
		SecretRefs:   selection.SecretRefs,
		SecretValues: values,
	}
	return r.factory.BuildKubernetesRuntimeEvidenceClient(ctx, request)
}
