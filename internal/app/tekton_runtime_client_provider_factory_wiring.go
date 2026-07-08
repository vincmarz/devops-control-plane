package app

import "context"

// TektonRuntimeClientProviderFactoryAwareRegistry combines the existing
// TektonRuntimeClientProviderRegistry with an optional
// TektonRuntimeClientFactory and RuntimeSecretValueLoader.
//
// The default configuration wires EmptyTektonRuntimeClientFactory and
// EmptyRuntimeSecretValueLoader so the factory fallback path always fails
// safely with an explicit "not configured" error, until real implementations
// are wired in a later phase.
type TektonRuntimeClientProviderFactoryAwareRegistry struct {
	registry TektonRuntimeClientProviderRegistry
	factory  TektonRuntimeClientFactory
	loader   RuntimeSecretValueLoader
}

// NewTektonRuntimeClientProviderFactoryAwareRegistry builds a factory-aware
// wrapper. Nil factory is normalized to EmptyTektonRuntimeClientFactory, nil
// loader is normalized to EmptyRuntimeSecretValueLoader.
func NewTektonRuntimeClientProviderFactoryAwareRegistry(
	registry TektonRuntimeClientProviderRegistry,
	factory TektonRuntimeClientFactory,
	loader RuntimeSecretValueLoader,
) TektonRuntimeClientProviderFactoryAwareRegistry {
	if factory == nil {
		factory = EmptyTektonRuntimeClientFactory{}
	}
	if loader == nil {
		loader = EmptyRuntimeSecretValueLoader{}
	}
	return TektonRuntimeClientProviderFactoryAwareRegistry{
		registry: registry,
		factory:  factory,
		loader:   loader,
	}
}

// Resolve returns a TektonRuntimeClient for the given selection. It first
// delegates to the existing provider registry, which handles the current
// cluster provider. If the registry cannot resolve a client and the selection
// carries configured Secret references, it attempts to build one through the
// runtime Secret value loader and the Tekton runtime client factory. With the
// conservative default placeholders, that fallback path always fails with an
// explicit "not configured" error.
func (r TektonRuntimeClientProviderFactoryAwareRegistry) Resolve(
	ctx context.Context,
	selection RuntimeClientProviderSelection,
) (TektonRuntimeClient, error) {
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
	request := TektonRuntimeClientFactoryRequest{
		Target:       selection.Target,
		SecretRefs:   selection.SecretRefs,
		SecretValues: values,
	}
	return r.factory.BuildTektonRuntimeClient(ctx, request)
}
