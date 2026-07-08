package app

import (
	"context"
	"errors"
	"testing"
)

func TestNewKubernetesRuntimeClientProviderFactoryAwareRegistryNormalizesNilPlaceholders(t *testing.T) {
	r := NewKubernetesRuntimeClientProviderFactoryAwareRegistry(
		EmptyKubernetesRuntimeClientProviderRegistry(),
		nil,
		nil,
	)

	selection := RuntimeClientProviderSelection{
		SecretRefsConfigured: true,
		SecretRefs: RuntimeClientSecretRefs{
			ClusterName: "ocp-staging",
			Kubernetes: &RuntimeSecretReference{
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

func TestKubernetesRuntimeClientProviderFactoryAwareRegistryReturnsRegistryErrorWhenNoSecretRefs(t *testing.T) {
	r := NewKubernetesRuntimeClientProviderFactoryAwareRegistry(
		EmptyKubernetesRuntimeClientProviderRegistry(),
		EmptyKubernetesRuntimeClientFactory{},
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
	if errors.Is(err, ErrKubernetesRuntimeClientFactoryNotConfigured) {
		t.Fatal("without SecretRefsConfigured the wrapper must not delegate to the factory")
	}
}

func TestKubernetesRuntimeClientProviderFactoryAwareRegistryFallsBackToLoaderWhenSecretRefsConfigured(t *testing.T) {
	r := NewKubernetesRuntimeClientProviderFactoryAwareRegistry(
		EmptyKubernetesRuntimeClientProviderRegistry(),
		EmptyKubernetesRuntimeClientFactory{},
		EmptyRuntimeSecretValueLoader{},
	)

	selection := RuntimeClientProviderSelection{
		SecretRefsConfigured: true,
		SecretRefs: RuntimeClientSecretRefs{
			ClusterName: "ocp-staging",
			Kubernetes: &RuntimeSecretReference{
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
