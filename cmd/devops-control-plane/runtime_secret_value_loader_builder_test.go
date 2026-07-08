package main

import (
	"errors"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/config"
)

func TestBuildRuntimeSecretValueLoaderReturnsEmptyLoaderWhenDisabled(t *testing.T) {
	getter := &fakeRuntimeKubernetesSecretGetter{}

	loader, err := buildRuntimeSecretValueLoader(
		config.Config{RuntimeSecretLoaderEnabled: false},
		app.KubernetesSecretValueLoaderConfig{},
		getter,
	)
	if err != nil {
		t.Fatalf("buildRuntimeSecretValueLoader returned error %v", err)
	}
	if _, ok := loader.(app.EmptyRuntimeSecretValueLoader); !ok {
		t.Fatalf("loader type = %T, want app.EmptyRuntimeSecretValueLoader", loader)
	}
	if getter.calls != 0 {
		t.Fatalf("GetSecret was called %d times, want 0", getter.calls)
	}
}

func TestBuildRuntimeSecretValueLoaderReturnsEmptyLoaderWhenDisabledAndGetterNil(t *testing.T) {
	loader, err := buildRuntimeSecretValueLoader(
		config.Config{RuntimeSecretLoaderEnabled: false},
		app.KubernetesSecretValueLoaderConfig{},
		nil,
	)
	if err != nil {
		t.Fatalf("buildRuntimeSecretValueLoader returned error %v", err)
	}
	if _, ok := loader.(app.EmptyRuntimeSecretValueLoader); !ok {
		t.Fatalf("loader type = %T, want app.EmptyRuntimeSecretValueLoader", loader)
	}
}

func TestBuildRuntimeSecretValueLoaderRejectsNilGetterWhenEnabled(t *testing.T) {
	loader, err := buildRuntimeSecretValueLoader(
		config.Config{RuntimeSecretLoaderEnabled: true},
		app.KubernetesSecretValueLoaderConfig{},
		nil,
	)
	if !errors.Is(err, errRuntimeKubernetesSecretGetterNotConfigured) {
		t.Fatalf("buildRuntimeSecretValueLoader error = %v, want errRuntimeKubernetesSecretGetterNotConfigured", err)
	}
	if loader != nil {
		t.Fatalf("loader type = %T, want nil", loader)
	}
}

func TestBuildRuntimeSecretValueLoaderReturnsAllowListLoaderWhenEnabled(t *testing.T) {
	getter := &fakeRuntimeKubernetesSecretGetter{}
	loaderConfig := app.KubernetesSecretValueLoaderConfig{
		AllowedClusters: []string{"ocp-staging"},
		AllowedRefs: []app.KubernetesSecretValueLoaderAllowedRef{
			{ClusterName: "ocp-staging", Namespace: "devops-control-plane", Name: "staging-runtime-client"},
		},
	}

	loader, err := buildRuntimeSecretValueLoader(
		config.Config{RuntimeSecretLoaderEnabled: true},
		loaderConfig,
		getter,
	)
	if err != nil {
		t.Fatalf("buildRuntimeSecretValueLoader returned error %v", err)
	}
	if _, ok := loader.(app.AllowListKubernetesSecretValueLoader); !ok {
		t.Fatalf("loader type = %T, want app.AllowListKubernetesSecretValueLoader", loader)
	}
	if getter.calls != 0 {
		t.Fatalf("GetSecret was called %d times, want 0", getter.calls)
	}
}

func TestBuildRuntimeSecretValueLoaderDoesNotReadSecrets(t *testing.T) {
	getter := &fakeRuntimeKubernetesSecretGetter{}

	loader, err := buildRuntimeSecretValueLoader(
		config.Config{RuntimeSecretLoaderEnabled: true},
		app.KubernetesSecretValueLoaderConfig{},
		getter,
	)
	if err != nil {
		t.Fatalf("buildRuntimeSecretValueLoader returned error %v", err)
	}
	if loader == nil {
		t.Fatal("buildRuntimeSecretValueLoader returned nil loader")
	}
	if getter.calls != 0 {
		t.Fatalf("GetSecret was called %d times, want 0", getter.calls)
	}
}
