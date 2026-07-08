package main

import (
	"errors"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/config"
)

func TestBuildRuntimeKubernetesClientFactoryDisabledByDefault(t *testing.T) {
	factory, err := buildRuntimeKubernetesClientFactory(config.Config{})
	if err != nil {
		t.Fatalf("buildRuntimeKubernetesClientFactory returned error %v", err)
	}
	if factory != nil {
		t.Fatalf("factory type = %T, want nil", factory)
	}
}

func TestBuildRuntimeKubernetesClientFactoryRequiresGlobalFlag(t *testing.T) {
	factory, err := buildRuntimeKubernetesClientFactory(config.Config{
		RuntimeClientFactoryKubernetesEnabled: true,
		KubernetesAPIURL:                      "https://api.dev.example:6443",
	})
	if err != nil {
		t.Fatalf("buildRuntimeKubernetesClientFactory returned error %v", err)
	}
	if factory != nil {
		t.Fatalf("factory type = %T, want nil", factory)
	}
}

func TestBuildRuntimeKubernetesClientFactoryRequiresAPIURLWhenEnabled(t *testing.T) {
	factory, err := buildRuntimeKubernetesClientFactory(config.Config{
		RuntimeClientFactoriesEnabled:         true,
		RuntimeClientFactoryKubernetesEnabled: true,
	})
	if !errors.Is(err, errRuntimeKubernetesClientFactoryAPIURLNotConfigured) {
		t.Fatalf("buildRuntimeKubernetesClientFactory error = %v, want errRuntimeKubernetesClientFactoryAPIURLNotConfigured", err)
	}
	if factory != nil {
		t.Fatalf("factory type = %T, want nil", factory)
	}
}

func TestBuildRuntimeKubernetesClientFactoryReturnsFactoryWhenEnabled(t *testing.T) {
	factory, err := buildRuntimeKubernetesClientFactory(config.Config{
		RuntimeClientFactoriesEnabled:         true,
		RuntimeClientFactoryKubernetesEnabled: true,
		KubernetesAPIURL:                      "https://api.dev.example:6443",
		TektonTimeoutSeconds:                  30,
	})
	if err != nil {
		t.Fatalf("buildRuntimeKubernetesClientFactory returned error %v", err)
	}
	if factory == nil {
		t.Fatal("buildRuntimeKubernetesClientFactory returned nil factory")
	}
}

func TestBuildRuntimeTektonClientFactoryDisabledByDefault(t *testing.T) {
	factory, err := buildRuntimeTektonClientFactory(config.Config{})
	if err != nil {
		t.Fatalf("buildRuntimeTektonClientFactory returned error %v", err)
	}
	if factory != nil {
		t.Fatalf("factory type = %T, want nil", factory)
	}
}

func TestBuildRuntimeTektonClientFactoryRequiresGlobalFlag(t *testing.T) {
	factory, err := buildRuntimeTektonClientFactory(config.Config{
		RuntimeClientFactoryTektonEnabled: true,
		KubernetesAPIURL:                  "https://api.dev.example:6443",
	})
	if err != nil {
		t.Fatalf("buildRuntimeTektonClientFactory returned error %v", err)
	}
	if factory != nil {
		t.Fatalf("factory type = %T, want nil", factory)
	}
}

func TestBuildRuntimeTektonClientFactoryRequiresAPIURLWhenEnabled(t *testing.T) {
	factory, err := buildRuntimeTektonClientFactory(config.Config{
		RuntimeClientFactoriesEnabled:     true,
		RuntimeClientFactoryTektonEnabled: true,
	})
	if !errors.Is(err, errRuntimeTektonClientFactoryAPIURLNotConfigured) {
		t.Fatalf("buildRuntimeTektonClientFactory error = %v, want errRuntimeTektonClientFactoryAPIURLNotConfigured", err)
	}
	if factory != nil {
		t.Fatalf("factory type = %T, want nil", factory)
	}
}

func TestBuildRuntimeTektonClientFactoryReturnsFactoryWhenEnabled(t *testing.T) {
	factory, err := buildRuntimeTektonClientFactory(config.Config{
		RuntimeClientFactoriesEnabled:     true,
		RuntimeClientFactoryTektonEnabled: true,
		KubernetesAPIURL:                  "https://api.dev.example:6443",
		TektonTimeoutSeconds:              30,
	})
	if err != nil {
		t.Fatalf("buildRuntimeTektonClientFactory returned error %v", err)
	}
	if factory == nil {
		t.Fatal("buildRuntimeTektonClientFactory returned nil factory")
	}
}

func TestBuildRuntimeArgoCDClientFactoryDisabledByDefault(t *testing.T) {
	factory, err := buildRuntimeArgoCDClientFactory(config.Config{})
	if err != nil {
		t.Fatalf("buildRuntimeArgoCDClientFactory returned error %v", err)
	}
	if factory != nil {
		t.Fatalf("factory type = %T, want nil", factory)
	}
}

func TestBuildRuntimeArgoCDClientFactoryRequiresGlobalFlag(t *testing.T) {
	factory, err := buildRuntimeArgoCDClientFactory(config.Config{
		RuntimeClientFactoryArgoCDEnabled: true,
		ArgoCDBaseURL:                     "https://argocd.dev.example",
	})
	if err != nil {
		t.Fatalf("buildRuntimeArgoCDClientFactory returned error %v", err)
	}
	if factory != nil {
		t.Fatalf("factory type = %T, want nil", factory)
	}
}

func TestBuildRuntimeArgoCDClientFactoryRequiresBaseURLWhenEnabled(t *testing.T) {
	factory, err := buildRuntimeArgoCDClientFactory(config.Config{
		RuntimeClientFactoriesEnabled:     true,
		RuntimeClientFactoryArgoCDEnabled: true,
	})
	if !errors.Is(err, errRuntimeArgoCDClientFactoryBaseURLNotConfigured) {
		t.Fatalf("buildRuntimeArgoCDClientFactory error = %v, want errRuntimeArgoCDClientFactoryBaseURLNotConfigured", err)
	}
	if factory != nil {
		t.Fatalf("factory type = %T, want nil", factory)
	}
}

func TestBuildRuntimeArgoCDClientFactoryReturnsFactoryWhenEnabled(t *testing.T) {
	factory, err := buildRuntimeArgoCDClientFactory(config.Config{
		RuntimeClientFactoriesEnabled:     true,
		RuntimeClientFactoryArgoCDEnabled: true,
		ArgoCDBaseURL:                     "https://argocd.dev.example",
		ArgoCDTimeoutSeconds:              30,
	})
	if err != nil {
		t.Fatalf("buildRuntimeArgoCDClientFactory returned error %v", err)
	}
	if factory == nil {
		t.Fatal("buildRuntimeArgoCDClientFactory returned nil factory")
	}
}
