package app

import "context"

// ArgoCDRuntimeClientProviderFactoryAwareRegistry combines the existing
// ArgoCDRuntimeClientProviderRegistry with an optional
// ArgoCDRuntimeClientFactory and RuntimeSecretValueLoader.
//
// The default configuration wires EmptyArgoCDRuntimeClientFactory and
// EmptyRuntimeSecretValueLoader so the factory fallback path always fails
// safely with an explicit "not configured" error, until real implementations
// are wired in a later phase.
type ArgoCDRuntimeClientProviderFactoryAwareRegistry struct {
	registry ArgoCDRuntimeClientProviderRegistry
	factory  ArgoCDRuntimeClientFactory
	loader   RuntimeSecretValueLoader
}

// NewArgoCDRuntimeClientProviderFactoryAwareRegistry builds a factory-aware
// wrapper. Nil factory is normalized to EmptyArgoCDRuntimeClientFactory, nil
// loader is normalized to EmptyRuntimeSecretValueLoader.
func NewArgoCDRuntimeClientProviderFactoryAwareRegistry(
	registry ArgoCDRuntimeClientProviderRegistry,
	factory ArgoCDRuntimeClientFactory,
	loader RuntimeSecretValueLoader,
) ArgoCDRuntimeClientProviderFactoryAwareRegistry {
	if factory == nil {
		factory = EmptyArgoCDRuntimeClientFactory{}
	}
	if loader == nil {
		loader = EmptyRuntimeSecretValueLoader{}
	}
	return ArgoCDRuntimeClientProviderFactoryAwareRegistry{
		registry: registry,
		factory:  factory,
		loader:   loader,
	}
}

// Resolve returns an ArgoCDRuntimeClient for the given selection. It first
// delegates to the existing provider registry, which handles the current
// cluster provider. If the registry cannot resolve a client and the selection
// carries configured Secret references, it attempts to build one through the
// runtime Secret value loader and the Argo CD runtime client factory. With
// the conservative default placeholders, that fallback path always fails with
// an explicit "not configured" error.
func (r ArgoCDRuntimeClientProviderFactoryAwareRegistry) Resolve(
	ctx context.Context,
	selection RuntimeClientProviderSelection,
) (ArgoCDRuntimeClient, error) {
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
	request := ArgoCDRuntimeClientFactoryRequest{
		Target:       selection.Target,
		SecretRefs:   selection.SecretRefs,
		SecretValues: values,
	}
	return r.factory.BuildArgoCDRuntimeClient(ctx, request)
}
