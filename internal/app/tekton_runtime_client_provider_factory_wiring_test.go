package app

import (
	"context"
	"errors"
	"testing"
)

func TestNewTektonRuntimeClientProviderFactoryAwareRegistryNormalizesNilPlaceholders(t *testing.T) {
	r := NewTektonRuntimeClientProviderFactoryAwareRegistry(
		EmptyTektonRuntimeClientProviderRegistry(),
		nil,
		nil,
	)

	selection := RuntimeClientProviderSelection{
		SecretRefsConfigured: true,
		SecretRefs: RuntimeClientSecretRefs{
			ClusterName: "ocp-staging",
			Tekton: &RuntimeSecretReference{
				Namespace: "devops-control-plane",
				Name:      "ocp-staging-kube",
				TokenKey:  "token",
			},
		},
	}

	_, err := r.Resolve(context.Background(), selection)
	if !errors.Is(err, ErrRuntimeSecretValueLoaderNotConfigured) {
		t.Fatalf("Resolve error = %v, want ErrRuntimeSecretValueLoaderNotConfigured", err)
	}
}

func TestTektonRuntimeClientProviderFactoryAwareRegistryReturnsRegistryErrorWhenNoSecretRefs(t *testing.T) {
	r := NewTektonRuntimeClientProviderFactoryAwareRegistry(
		EmptyTektonRuntimeClientProviderRegistry(),
		EmptyTektonRuntimeClientFactory{},
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
	if errors.Is(err, ErrTektonRuntimeClientFactoryNotConfigured) {
		t.Fatal("without SecretRefsConfigured the wrapper must not delegate to the factory")
	}
}

func TestTektonRuntimeClientProviderFactoryAwareRegistryFallsBackToLoaderWhenSecretRefsConfigured(t *testing.T) {
	r := NewTektonRuntimeClientProviderFactoryAwareRegistry(
		EmptyTektonRuntimeClientProviderRegistry(),
		EmptyTektonRuntimeClientFactory{},
		EmptyRuntimeSecretValueLoader{},
	)

	selection := RuntimeClientProviderSelection{
		SecretRefsConfigured: true,
		SecretRefs: RuntimeClientSecretRefs{
			ClusterName: "ocp-staging",
			Tekton: &RuntimeSecretReference{
				Namespace: "devops-control-plane",
				Name:      "ocp-staging-kube",
				TokenKey:  "token",
			},
		},
	}

	_, err := r.Resolve(context.Background(), selection)
	if !errors.Is(err, ErrRuntimeSecretValueLoaderNotConfigured) {
		t.Fatalf("Resolve error = %v, want ErrRuntimeSecretValueLoaderNotConfigured", err)
	}
}
