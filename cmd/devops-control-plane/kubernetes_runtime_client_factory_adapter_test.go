package main

import (
	"context"
	"errors"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/app"
)

var _ app.KubernetesRuntimeClientFactory = kubernetesRuntimeClientFactoryAdapter{}

func validKubernetesRuntimeClientFactoryAdapterTarget() app.TechnicalRuntimeTarget {
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

func validKubernetesRuntimeClientFactoryAdapterRefs() app.RuntimeClientSecretRefs {
	return app.RuntimeClientSecretRefs{
		ClusterName: "ocp-staging",
		Kubernetes: &app.RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "staging-runtime-client",
			TokenKey:  "token",
		},
	}
}

func validKubernetesRuntimeClientFactoryAdapterRequest() app.KubernetesRuntimeClientFactoryRequest {
	return app.KubernetesRuntimeClientFactoryRequest{
		Target:     validKubernetesRuntimeClientFactoryAdapterTarget(),
		SecretRefs: validKubernetesRuntimeClientFactoryAdapterRefs(),
		SecretValues: app.NewRuntimeSecretValueSet(map[app.RuntimeSecretValueKey]string{
			app.RuntimeSecretValueKubernetesToken: "redacted-token-for-unit-test",
		}),
	}
}

func TestKubernetesRuntimeClientFactoryAdapterBuildsTokenBasedClient(t *testing.T) {
	factory := newKubernetesRuntimeClientFactoryAdapter(kubernetesRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{"ocp-staging": "https://api.staging.example:6443"},
		TimeoutSeconds: 15,
		InsecureTLS:    true,
	})

	client, err := factory.BuildKubernetesRuntimeEvidenceClient(context.Background(), validKubernetesRuntimeClientFactoryAdapterRequest())
	if err != nil {
		t.Fatalf("BuildKubernetesRuntimeEvidenceClient returned error %v", err)
	}
	if client == nil {
		t.Fatal("BuildKubernetesRuntimeEvidenceClient returned nil client")
	}
}

func TestKubernetesRuntimeClientFactoryAdapterRejectsInvalidRequest(t *testing.T) {
	factory := newKubernetesRuntimeClientFactoryAdapter(kubernetesRuntimeClientFactoryAdapterConfig{})

	client, err := factory.BuildKubernetesRuntimeEvidenceClient(context.Background(), app.KubernetesRuntimeClientFactoryRequest{})
	if err == nil {
		t.Fatal("BuildKubernetesRuntimeEvidenceClient returned nil error for invalid request")
	}
	if client != nil {
		t.Fatal("BuildKubernetesRuntimeEvidenceClient returned client for invalid request")
	}
}

func TestKubernetesRuntimeClientFactoryAdapterRequiresTokenValue(t *testing.T) {
	factory := newKubernetesRuntimeClientFactoryAdapter(kubernetesRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{"ocp-staging": "https://api.staging.example:6443"},
	})
	request := validKubernetesRuntimeClientFactoryAdapterRequest()
	request.SecretValues = app.EmptyRuntimeSecretValueSet()

	client, err := factory.BuildKubernetesRuntimeEvidenceClient(context.Background(), request)
	if !errors.Is(err, app.ErrRuntimeSecretValueNotAvailable) {
		t.Fatalf("BuildKubernetesRuntimeEvidenceClient error = %v, want ErrRuntimeSecretValueNotAvailable", err)
	}
	if client != nil {
		t.Fatal("BuildKubernetesRuntimeEvidenceClient returned client without token")
	}
}

func TestKubernetesRuntimeClientFactoryAdapterRequiresAPIURLForCluster(t *testing.T) {
	factory := newKubernetesRuntimeClientFactoryAdapter(kubernetesRuntimeClientFactoryAdapterConfig{})

	client, err := factory.BuildKubernetesRuntimeEvidenceClient(context.Background(), validKubernetesRuntimeClientFactoryAdapterRequest())
	if !errors.Is(err, errKubernetesRuntimeClientFactoryAPIURLNotConfigured) {
		t.Fatalf("BuildKubernetesRuntimeEvidenceClient error = %v, want errKubernetesRuntimeClientFactoryAPIURLNotConfigured", err)
	}
	if client != nil {
		t.Fatal("BuildKubernetesRuntimeEvidenceClient returned client without API URL")
	}
}

func TestKubernetesRuntimeClientFactoryAdapterRejectsKubeconfigReferenceUntilSupported(t *testing.T) {
	factory := newKubernetesRuntimeClientFactoryAdapter(kubernetesRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{"ocp-staging": "https://api.staging.example:6443"},
	})
	request := validKubernetesRuntimeClientFactoryAdapterRequest()
	request.SecretRefs.Kubernetes.TokenKey = ""
	request.SecretRefs.Kubernetes.KubeconfigKey = "kubeconfig"
	request.SecretValues = app.NewRuntimeSecretValueSet(map[app.RuntimeSecretValueKey]string{
		app.RuntimeSecretValueKubernetesKubeconfig: "redacted-kubeconfig-for-unit-test",
	})

	client, err := factory.BuildKubernetesRuntimeEvidenceClient(context.Background(), request)
	if !errors.Is(err, errKubernetesRuntimeClientFactoryKubeconfigUnsupported) {
		t.Fatalf("BuildKubernetesRuntimeEvidenceClient error = %v, want errKubernetesRuntimeClientFactoryKubeconfigUnsupported", err)
	}
	if client != nil {
		t.Fatal("BuildKubernetesRuntimeEvidenceClient returned client for unsupported kubeconfig input")
	}
}

func TestKubernetesRuntimeClientFactoryAdapterRejectsRawCAReferenceUntilSupported(t *testing.T) {
	factory := newKubernetesRuntimeClientFactoryAdapter(kubernetesRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{"ocp-staging": "https://api.staging.example:6443"},
	})
	request := validKubernetesRuntimeClientFactoryAdapterRequest()
	request.SecretRefs.Kubernetes.CAKey = "ca.crt"
	request.SecretValues = app.NewRuntimeSecretValueSet(map[app.RuntimeSecretValueKey]string{
		app.RuntimeSecretValueKubernetesToken: "redacted-token-for-unit-test",
		app.RuntimeSecretValueKubernetesCA:    "redacted-ca-for-unit-test",
	})

	client, err := factory.BuildKubernetesRuntimeEvidenceClient(context.Background(), request)
	if !errors.Is(err, errKubernetesRuntimeClientFactoryRawCAUnsupported) {
		t.Fatalf("BuildKubernetesRuntimeEvidenceClient error = %v, want errKubernetesRuntimeClientFactoryRawCAUnsupported", err)
	}
	if client != nil {
		t.Fatal("BuildKubernetesRuntimeEvidenceClient returned client for unsupported raw CA input")
	}
}

func TestKubernetesRuntimeClientFactoryAdapterNormalizesClusterAPIURLKeys(t *testing.T) {
	factory := newKubernetesRuntimeClientFactoryAdapter(kubernetesRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{" OCP-STAGING ": " https://api.staging.example:6443 "},
		InsecureTLS:    true,
	})

	client, err := factory.BuildKubernetesRuntimeEvidenceClient(context.Background(), validKubernetesRuntimeClientFactoryAdapterRequest())
	if err != nil {
		t.Fatalf("BuildKubernetesRuntimeEvidenceClient returned error %v", err)
	}
	if client == nil {
		t.Fatal("BuildKubernetesRuntimeEvidenceClient returned nil client")
	}
}
