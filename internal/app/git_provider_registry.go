package app

import (
	"errors"
	"fmt"
	"strings"
)

// GitProviderRegistry resolves a concrete Git provider instance by providerRef
// and verifies that the registered adapter type matches the repository target.
type GitProviderRegistry struct {
	providers map[string]GitProvider
}

func NewGitProviderRegistry(providers []GitProvider) (GitProviderRegistry, error) {
	registry := GitProviderRegistry{providers: map[string]GitProvider{}}
	for index, provider := range providers {
		if provider == nil {
			return GitProviderRegistry{}, fmt.Errorf("git provider at index %d is nil", index)
		}
		providerType := strings.ToLower(strings.TrimSpace(provider.Provider()))
		providerRef := strings.TrimSpace(provider.ProviderRef())
		if providerType == "" {
			return GitProviderRegistry{}, fmt.Errorf("git provider at index %d does not define provider", index)
		}
		if providerRef == "" {
			return GitProviderRegistry{}, fmt.Errorf("git provider at index %d does not define providerRef", index)
		}
		if _, exists := registry.providers[providerRef]; exists {
			return GitProviderRegistry{}, fmt.Errorf("git providerRef %q is registered more than once", providerRef)
		}
		registry.providers[providerRef] = provider
	}
	return registry, nil
}

func (r GitProviderRegistry) Resolve(target GitRepositoryTarget) (GitProvider, error) {
	providerType := strings.ToLower(strings.TrimSpace(target.Provider))
	providerRef := strings.TrimSpace(target.ProviderRef)
	if providerType == "" {
		return nil, errors.New("git repository target provider is required")
	}
	if providerRef == "" {
		return nil, errors.New("git repository target providerRef is required")
	}
	provider, ok := r.providers[providerRef]
	if !ok {
		return nil, fmt.Errorf("git providerRef %q is not registered", providerRef)
	}
	registeredType := strings.ToLower(strings.TrimSpace(provider.Provider()))
	if registeredType != providerType {
		return nil, fmt.Errorf("git providerRef %q is registered as %q, not %q", providerRef, registeredType, providerType)
	}
	return provider, nil
}
