package app

import (
	"context"
	"errors"
	"testing"
)

func TestNewArgoCDRuntimeClientProviderFactoryAwareRegistryNormalizesNilPlaceholders(t *testing.T) {
	r := NewArgoCDRuntimeClientProviderFactoryAwareRegistry(
		EmptyArgoCDRuntimeClientProviderRegistry(),
		nil,
		nil,
	)

	selection := RuntimeClientProviderSelection{
		SecretRefsConfigured: true,
		SecretRefs: RuntimeClientSecretRefs{
			ClusterName: "ocp-staging",
			ArgoCD: &RuntimeSecretReference{
				Namespace: "devops-control-plane",
				Name:      "ocp-staging-argocd",
				TokenKey:  "token",
			},
		},
	}

	_, err := r.Resolve(context.Background(), selection)
	if !errors.Is(err, ErrRuntimeSecretValueLoaderNotConfigured) {
		t.Fatalf("Resolve error = %v, want ErrRuntimeSecretValueLoaderNotConfigured", err)
	}
}

func TestArgoCDRuntimeClientProviderFactoryAwareRegistryReturnsRegistryErrorWhenNoSecretRefs(t *testing.T) {
	r := NewArgoCDRuntimeClientProviderFactoryAwareRegistry(
		EmptyArgoCDRuntimeClientProviderRegistry(),
		EmptyArgoCDRuntimeClientFactory{},
		EmptyRuntimeSecretValueLoader{},
	)

	selection := RuntimeClientProviderSelection{
		SecretRefsConfigured: false,
	}

	_, err := r.Resolve(context.Background(), selection)
	if err == nil {
		t.Fatal("expected error from empty registry, got nil")
	}
	if errors.Is(err, ErrRuntimeSecretValueLoaderNotConfigured) {
		t.Fatal("without SecretRefsConfigured the wrapper must not delegate to the loader")
	}
	if errors.Is(err, ErrArgoCDRuntimeClientFactoryNotConfigured) {
		t.Fatal("without SecretRefsConfigured the wrapper must not delegate to the factory")
	}
}

func TestArgoCDRuntimeClientProviderFactoryAwareRegistryFallsBackToLoaderWhenSecretRefsConfigured(t *testing.T) {
	r := NewArgoCDRuntimeClientProviderFactoryAwareRegistry(
		EmptyArgoCDRuntimeClientProviderRegistry(),
		EmptyArgoCDRuntimeClientFactory{},
		EmptyRuntimeSecretValueLoader{},
	)

	selection := RuntimeClientProviderSelection{
		SecretRefsConfigured: true,
		SecretRefs: RuntimeClientSecretRefs{
			ClusterName: "ocp-staging",
			ArgoCD: &RuntimeSecretReference{
				Namespace: "devops-control-plane",
				Name:      "ocp-staging-argocd",
				TokenKey:  "token",
			},
		},
	}

	_, err := r.Resolve(context.Background(), selection)
	if !errors.Is(err, ErrRuntimeSecretValueLoaderNotConfigured) {
		t.Fatalf("Resolve error = %v, want ErrRuntimeSecretValueLoaderNotConfigured", err)
	}
}
