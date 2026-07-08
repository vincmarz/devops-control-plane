package main

import (
	"context"
	"errors"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/config"
)

type fakeRuntimeKubernetesSecretGetter struct {
	calls int
}

func (f *fakeRuntimeKubernetesSecretGetter) GetSecret(_ context.Context, _ string, _ string) (map[string][]byte, error) {
	f.calls++
	return map[string][]byte{"token": []byte("redacted-token-for-unit-test")}, nil
}

func TestBuildRuntimeKubernetesSecretGetterReturnsNilWhenLoaderDisabled(t *testing.T) {
	getter := &fakeRuntimeKubernetesSecretGetter{}

	built, err := buildRuntimeKubernetesSecretGetter(config.Config{RuntimeSecretLoaderEnabled: false}, getter)
	if err != nil {
		t.Fatalf("buildRuntimeKubernetesSecretGetter returned error %v", err)
	}
	if built != nil {
		t.Fatal("buildRuntimeKubernetesSecretGetter returned getter while loader is disabled")
	}
	if getter.calls != 0 {
		t.Fatalf("GetSecret was called %d times, want 0", getter.calls)
	}
}

func TestBuildRuntimeKubernetesSecretGetterReturnsNilWhenLoaderDisabledAndGetterNil(t *testing.T) {
	built, err := buildRuntimeKubernetesSecretGetter(config.Config{RuntimeSecretLoaderEnabled: false}, nil)
	if err != nil {
		t.Fatalf("buildRuntimeKubernetesSecretGetter returned error %v", err)
	}
	if built != nil {
		t.Fatal("buildRuntimeKubernetesSecretGetter returned getter while loader is disabled")
	}
}

func TestBuildRuntimeKubernetesSecretGetterRejectsNilGetterWhenLoaderEnabled(t *testing.T) {
	built, err := buildRuntimeKubernetesSecretGetter(config.Config{RuntimeSecretLoaderEnabled: true}, nil)
	if !errors.Is(err, errRuntimeKubernetesSecretGetterNotConfigured) {
		t.Fatalf("buildRuntimeKubernetesSecretGetter error = %v, want errRuntimeKubernetesSecretGetterNotConfigured", err)
	}
	if built != nil {
		t.Fatal("buildRuntimeKubernetesSecretGetter returned getter when input getter is nil")
	}
}

func TestBuildRuntimeKubernetesSecretGetterReturnsGetterWhenLoaderEnabled(t *testing.T) {
	getter := &fakeRuntimeKubernetesSecretGetter{}

	built, err := buildRuntimeKubernetesSecretGetter(config.Config{RuntimeSecretLoaderEnabled: true}, getter)
	if err != nil {
		t.Fatalf("buildRuntimeKubernetesSecretGetter returned error %v", err)
	}
	if built == nil {
		t.Fatal("buildRuntimeKubernetesSecretGetter returned nil getter")
	}
	if built != app.KubernetesSecretGetter(getter) {
		t.Fatal("buildRuntimeKubernetesSecretGetter did not return the supplied getter")
	}
	if getter.calls != 0 {
		t.Fatalf("GetSecret was called %d times, want 0", getter.calls)
	}
}

func TestBuildRuntimeKubernetesSecretGetterDoesNotReadSecrets(t *testing.T) {
	getter := &fakeRuntimeKubernetesSecretGetter{}

	built, err := buildRuntimeKubernetesSecretGetter(config.Config{RuntimeSecretLoaderEnabled: true}, getter)
	if err != nil {
		t.Fatalf("buildRuntimeKubernetesSecretGetter returned error %v", err)
	}
	if built == nil {
		t.Fatal("buildRuntimeKubernetesSecretGetter returned nil getter")
	}
	if getter.calls != 0 {
		t.Fatalf("GetSecret was called %d times, want 0", getter.calls)
	}
}
