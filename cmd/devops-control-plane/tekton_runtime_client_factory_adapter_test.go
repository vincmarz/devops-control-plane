package main

import (
	"context"
	"errors"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/app"
)

var _ app.TektonRuntimeClientFactory = tektonRuntimeClientFactoryAdapter{}

func validTektonRuntimeClientFactoryAdapterTarget() app.TechnicalRuntimeTarget {
	return app.TechnicalRuntimeTarget{
		TargetEnvironment:      "staging",
		EnvironmentName:        "staging",
		EnvironmentDisplayName: "Staging",
		ClusterName:            "ocp-staging",
		ClusterDisplayName:     "OpenShift Staging",
		ClusterEnabled:         true,
		KubernetesNamespace:    "devops-ci-staging",
		TektonNamespace:        "devops-ci-staging",
		TektonPipelineName:     "validate-gitops",
		ArgoCDApplicationName:  "demo-go-color-app",
		GitTargetBranch:        "main",
	}
}

func validTektonRuntimeClientFactoryAdapterRefs() app.RuntimeClientSecretRefs {
	return app.RuntimeClientSecretRefs{
		ClusterName: "ocp-staging",
		Tekton: &app.RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "staging-tekton-runtime-client",
			TokenKey:  "token",
		},
	}
}

func validTektonRuntimeClientFactoryAdapterRequest() app.TektonRuntimeClientFactoryRequest {
	return app.TektonRuntimeClientFactoryRequest{
		Target:     validTektonRuntimeClientFactoryAdapterTarget(),
		SecretRefs: validTektonRuntimeClientFactoryAdapterRefs(),
		SecretValues: app.NewRuntimeSecretValueSet(map[app.RuntimeSecretValueKey]string{
			app.RuntimeSecretValueKubernetesToken: "redacted-token-for-unit-test",
		}),
	}
}

func TestTektonRuntimeClientFactoryAdapterBuildsTokenBasedClient(t *testing.T) {
	factory := newTektonRuntimeClientFactoryAdapter(tektonRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{"ocp-staging": "https://api.staging.example:6443"},
		TimeoutSeconds: 15,
		InsecureTLS:    true,
	})

	client, err := factory.BuildTektonRuntimeClient(context.Background(), validTektonRuntimeClientFactoryAdapterRequest())
	if err != nil {
		t.Fatalf("BuildTektonRuntimeClient returned error %v", err)
	}
	if client == nil {
		t.Fatal("BuildTektonRuntimeClient returned nil client")
	}
}

func TestTektonRuntimeClientFactoryAdapterRejectsInvalidRequest(t *testing.T) {
	factory := newTektonRuntimeClientFactoryAdapter(tektonRuntimeClientFactoryAdapterConfig{})

	client, err := factory.BuildTektonRuntimeClient(context.Background(), app.TektonRuntimeClientFactoryRequest{})
	if err == nil {
		t.Fatal("BuildTektonRuntimeClient returned nil error for invalid request")
	}
	if client != nil {
		t.Fatal("BuildTektonRuntimeClient returned client for invalid request")
	}
}

func TestTektonRuntimeClientFactoryAdapterRequiresTokenValue(t *testing.T) {
	factory := newTektonRuntimeClientFactoryAdapter(tektonRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{"ocp-staging": "https://api.staging.example:6443"},
	})
	request := validTektonRuntimeClientFactoryAdapterRequest()
	request.SecretValues = app.EmptyRuntimeSecretValueSet()

	client, err := factory.BuildTektonRuntimeClient(context.Background(), request)
	if !errors.Is(err, app.ErrRuntimeSecretValueNotAvailable) {
		t.Fatalf("BuildTektonRuntimeClient error = %v, want ErrRuntimeSecretValueNotAvailable", err)
	}
	if client != nil {
		t.Fatal("BuildTektonRuntimeClient returned client without token")
	}
}

func TestTektonRuntimeClientFactoryAdapterRequiresAPIURLForCluster(t *testing.T) {
	factory := newTektonRuntimeClientFactoryAdapter(tektonRuntimeClientFactoryAdapterConfig{})

	client, err := factory.BuildTektonRuntimeClient(context.Background(), validTektonRuntimeClientFactoryAdapterRequest())
	if !errors.Is(err, errTektonRuntimeClientFactoryAPIURLNotConfigured) {
		t.Fatalf("BuildTektonRuntimeClient error = %v, want errTektonRuntimeClientFactoryAPIURLNotConfigured", err)
	}
	if client != nil {
		t.Fatal("BuildTektonRuntimeClient returned client without API URL")
	}
}

func TestTektonRuntimeClientFactoryAdapterRejectsKubeconfigReferenceUntilSupported(t *testing.T) {
	factory := newTektonRuntimeClientFactoryAdapter(tektonRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{"ocp-staging": "https://api.staging.example:6443"},
	})
	request := validTektonRuntimeClientFactoryAdapterRequest()
	request.SecretRefs.Tekton.TokenKey = ""
	request.SecretRefs.Tekton.KubeconfigKey = "kubeconfig"
	request.SecretValues = app.NewRuntimeSecretValueSet(map[app.RuntimeSecretValueKey]string{
		app.RuntimeSecretValueKubernetesKubeconfig: "redacted-kubeconfig-for-unit-test",
	})

	client, err := factory.BuildTektonRuntimeClient(context.Background(), request)
	if !errors.Is(err, errTektonRuntimeClientFactoryKubeconfigUnsupported) {
		t.Fatalf("BuildTektonRuntimeClient error = %v, want errTektonRuntimeClientFactoryKubeconfigUnsupported", err)
	}
	if client != nil {
		t.Fatal("BuildTektonRuntimeClient returned client for unsupported kubeconfig input")
	}
}

func TestTektonRuntimeClientFactoryAdapterRejectsRawCAReferenceUntilSupported(t *testing.T) {
	factory := newTektonRuntimeClientFactoryAdapter(tektonRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{"ocp-staging": "https://api.staging.example:6443"},
	})
	request := validTektonRuntimeClientFactoryAdapterRequest()
	request.SecretRefs.Tekton.CAKey = "ca.crt"
	request.SecretValues = app.NewRuntimeSecretValueSet(map[app.RuntimeSecretValueKey]string{
		app.RuntimeSecretValueKubernetesToken: "redacted-token-for-unit-test",
		app.RuntimeSecretValueKubernetesCA:    "redacted-ca-for-unit-test",
	})

	client, err := factory.BuildTektonRuntimeClient(context.Background(), request)
	if !errors.Is(err, errTektonRuntimeClientFactoryRawCAUnsupported) {
		t.Fatalf("BuildTektonRuntimeClient error = %v, want errTektonRuntimeClientFactoryRawCAUnsupported", err)
	}
	if client != nil {
		t.Fatal("BuildTektonRuntimeClient returned client for unsupported raw CA input")
	}
}

func TestTektonRuntimeClientFactoryAdapterNormalizesClusterAPIURLKeys(t *testing.T) {
	factory := newTektonRuntimeClientFactoryAdapter(tektonRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{" OCP-STAGING ": " https://api.staging.example:6443 "},
		InsecureTLS:    true,
	})

	client, err := factory.BuildTektonRuntimeClient(context.Background(), validTektonRuntimeClientFactoryAdapterRequest())
	if err != nil {
		t.Fatalf("BuildTektonRuntimeClient returned error %v", err)
	}
	if client == nil {
		t.Fatal("BuildTektonRuntimeClient returned nil client")
	}
}
