package main

import (
	"context"
	"errors"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/app"
)

var _ app.ArgoCDRuntimeClientFactory = argoCDRuntimeClientFactoryAdapter{}

func validArgoCDRuntimeClientFactoryAdapterTarget() app.TechnicalRuntimeTarget {
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

func validArgoCDRuntimeClientFactoryAdapterRefs() app.RuntimeClientSecretRefs {
	return app.RuntimeClientSecretRefs{
		ClusterName: "ocp-staging",
		ArgoCD: &app.RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "staging-argocd-runtime-client",
			TokenKey:  "token",
		},
	}
}

func validArgoCDRuntimeClientFactoryAdapterRequest() app.ArgoCDRuntimeClientFactoryRequest {
	return app.ArgoCDRuntimeClientFactoryRequest{
		Target:     validArgoCDRuntimeClientFactoryAdapterTarget(),
		SecretRefs: validArgoCDRuntimeClientFactoryAdapterRefs(),
		SecretValues: app.NewRuntimeSecretValueSet(map[app.RuntimeSecretValueKey]string{
			app.RuntimeSecretValueArgoCDToken: "redacted-token-for-unit-test",
		}),
	}
}

func TestArgoCDRuntimeClientFactoryAdapterBuildsTokenBasedClientWithConfiguredBaseURL(t *testing.T) {
	factory := newArgoCDRuntimeClientFactoryAdapter(argoCDRuntimeClientFactoryAdapterConfig{
		ClusterBaseURLs: map[string]string{"ocp-staging": "https://argocd.staging.example"},
		TimeoutSeconds:  15,
		InsecureTLS:     true,
	})

	client, err := factory.BuildArgoCDRuntimeClient(context.Background(), validArgoCDRuntimeClientFactoryAdapterRequest())
	if err != nil {
		t.Fatalf("BuildArgoCDRuntimeClient returned error %v", err)
	}
	if client == nil {
		t.Fatal("BuildArgoCDRuntimeClient returned nil client")
	}
}

func TestArgoCDRuntimeClientFactoryAdapterBuildsClientWithBaseURLFromSecretValues(t *testing.T) {
	factory := newArgoCDRuntimeClientFactoryAdapter(argoCDRuntimeClientFactoryAdapterConfig{
		TimeoutSeconds: 15,
		InsecureTLS:    true,
	})
	request := validArgoCDRuntimeClientFactoryAdapterRequest()
	request.SecretRefs.ArgoCD.BaseURLKey = "baseURL"
	request.SecretValues = app.NewRuntimeSecretValueSet(map[app.RuntimeSecretValueKey]string{
		app.RuntimeSecretValueArgoCDToken:   "redacted-token-for-unit-test",
		app.RuntimeSecretValueArgoCDBaseURL: "https://argocd-from-secret.staging.example",
	})

	client, err := factory.BuildArgoCDRuntimeClient(context.Background(), request)
	if err != nil {
		t.Fatalf("BuildArgoCDRuntimeClient returned error %v", err)
	}
	if client == nil {
		t.Fatal("BuildArgoCDRuntimeClient returned nil client")
	}
}

func TestArgoCDRuntimeClientFactoryAdapterRejectsInvalidRequest(t *testing.T) {
	factory := newArgoCDRuntimeClientFactoryAdapter(argoCDRuntimeClientFactoryAdapterConfig{})

	client, err := factory.BuildArgoCDRuntimeClient(context.Background(), app.ArgoCDRuntimeClientFactoryRequest{})
	if err == nil {
		t.Fatal("BuildArgoCDRuntimeClient returned nil error for invalid request")
	}
	if client != nil {
		t.Fatal("BuildArgoCDRuntimeClient returned client for invalid request")
	}
}

func TestArgoCDRuntimeClientFactoryAdapterRequiresTokenValue(t *testing.T) {
	factory := newArgoCDRuntimeClientFactoryAdapter(argoCDRuntimeClientFactoryAdapterConfig{
		ClusterBaseURLs: map[string]string{"ocp-staging": "https://argocd.staging.example"},
	})
	request := validArgoCDRuntimeClientFactoryAdapterRequest()
	request.SecretValues = app.EmptyRuntimeSecretValueSet()

	client, err := factory.BuildArgoCDRuntimeClient(context.Background(), request)
	if !errors.Is(err, app.ErrRuntimeSecretValueNotAvailable) {
		t.Fatalf("BuildArgoCDRuntimeClient error = %v, want ErrRuntimeSecretValueNotAvailable", err)
	}
	if client != nil {
		t.Fatal("BuildArgoCDRuntimeClient returned client without token")
	}
}

func TestArgoCDRuntimeClientFactoryAdapterRequiresConfiguredBaseURL(t *testing.T) {
	factory := newArgoCDRuntimeClientFactoryAdapter(argoCDRuntimeClientFactoryAdapterConfig{})

	client, err := factory.BuildArgoCDRuntimeClient(context.Background(), validArgoCDRuntimeClientFactoryAdapterRequest())
	if !errors.Is(err, errArgoCDRuntimeClientFactoryBaseURLNotConfigured) {
		t.Fatalf("BuildArgoCDRuntimeClient error = %v, want errArgoCDRuntimeClientFactoryBaseURLNotConfigured", err)
	}
	if client != nil {
		t.Fatal("BuildArgoCDRuntimeClient returned client without base URL")
	}
}

func TestArgoCDRuntimeClientFactoryAdapterRequiresBaseURLSecretValueWhenReferenced(t *testing.T) {
	factory := newArgoCDRuntimeClientFactoryAdapter(argoCDRuntimeClientFactoryAdapterConfig{
		ClusterBaseURLs: map[string]string{"ocp-staging": "https://argocd.staging.example"},
	})
	request := validArgoCDRuntimeClientFactoryAdapterRequest()
	request.SecretRefs.ArgoCD.BaseURLKey = "baseURL"

	client, err := factory.BuildArgoCDRuntimeClient(context.Background(), request)
	if !errors.Is(err, app.ErrRuntimeSecretValueNotAvailable) {
		t.Fatalf("BuildArgoCDRuntimeClient error = %v, want ErrRuntimeSecretValueNotAvailable", err)
	}
	if client != nil {
		t.Fatal("BuildArgoCDRuntimeClient returned client without base URL secret value")
	}
}

func TestArgoCDRuntimeClientFactoryAdapterRejectsRawCAReferenceUntilSupported(t *testing.T) {
	factory := newArgoCDRuntimeClientFactoryAdapter(argoCDRuntimeClientFactoryAdapterConfig{
		ClusterBaseURLs: map[string]string{"ocp-staging": "https://argocd.staging.example"},
	})
	request := validArgoCDRuntimeClientFactoryAdapterRequest()
	request.SecretRefs.ArgoCD.CAKey = "ca.crt"
	request.SecretValues = app.NewRuntimeSecretValueSet(map[app.RuntimeSecretValueKey]string{
		app.RuntimeSecretValueArgoCDToken: "redacted-token-for-unit-test",
		app.RuntimeSecretValueArgoCDCA:    "redacted-ca-for-unit-test",
	})

	client, err := factory.BuildArgoCDRuntimeClient(context.Background(), request)
	if !errors.Is(err, errArgoCDRuntimeClientFactoryRawCAUnsupported) {
		t.Fatalf("BuildArgoCDRuntimeClient error = %v, want errArgoCDRuntimeClientFactoryRawCAUnsupported", err)
	}
	if client != nil {
		t.Fatal("BuildArgoCDRuntimeClient returned client for unsupported raw CA input")
	}
}

func TestArgoCDRuntimeClientFactoryAdapterNormalizesClusterBaseURLKeys(t *testing.T) {
	factory := newArgoCDRuntimeClientFactoryAdapter(argoCDRuntimeClientFactoryAdapterConfig{
		ClusterBaseURLs: map[string]string{" OCP-STAGING ": " https://argocd.staging.example "},
		InsecureTLS:     true,
	})

	client, err := factory.BuildArgoCDRuntimeClient(context.Background(), validArgoCDRuntimeClientFactoryAdapterRequest())
	if err != nil {
		t.Fatalf("BuildArgoCDRuntimeClient returned error %v", err)
	}
	if client == nil {
		t.Fatal("BuildArgoCDRuntimeClient returned nil client")
	}
}
