package app

import (
	"context"
	"errors"
	"testing"
)

func TestEmptyArgoCDRuntimeClientFactoryAlwaysReturnsNotConfigured(t *testing.T) {
	var factory ArgoCDRuntimeClientFactory = EmptyArgoCDRuntimeClientFactory{}

	client, err := factory.BuildArgoCDRuntimeClient(context.Background(), ArgoCDRuntimeClientFactoryRequest{})
	if !errors.Is(err, ErrArgoCDRuntimeClientFactoryNotConfigured) {
		t.Fatalf("BuildArgoCDRuntimeClient error = %v, want ErrArgoCDRuntimeClientFactoryNotConfigured", err)
	}
	if client != nil {
		t.Fatal("EmptyArgoCDRuntimeClientFactory must never return a client")
	}
}

func validArgoCDFactoryRequest() ArgoCDRuntimeClientFactoryRequest {
	target := TechnicalRuntimeTarget{
		TargetEnvironment:     "dev",
		EnvironmentName:       "dev",
		ClusterName:           "ocp-dev",
		ClusterEnabled:        true,
		KubernetesNamespace:   "devops-ci-demo",
		TektonNamespace:       "devops-ci-demo",
		TektonPipelineName:    "validate-gitops",
		ArgoCDApplicationName: "demo-go-color-app",
		GitTargetBranch:       "main",
	}
	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-dev",
		ArgoCD: &RuntimeSecretReference{
			Namespace:  "devops-control-plane",
			Name:       "ocp-dev-argocd",
			TokenKey:   "token",
			BaseURLKey: "baseURL",
			CAKey:      "ca.crt",
		},
	}
	values := NewRuntimeSecretValueSet(map[RuntimeSecretValueKey]string{
		RuntimeSecretValueArgoCDToken:   "dummy-token-for-tests",
		RuntimeSecretValueArgoCDBaseURL: "https://argocd.invalid",
		RuntimeSecretValueArgoCDCA:      "dummy-ca-for-tests",
	})
	return ArgoCDRuntimeClientFactoryRequest{Target: target, SecretRefs: refs, SecretValues: values}
}

func TestValidateArgoCDRuntimeClientFactoryRequestAcceptsValidRequest(t *testing.T) {
	if err := ValidateArgoCDRuntimeClientFactoryRequest(validArgoCDFactoryRequest()); err != nil {
		t.Fatalf("ValidateArgoCDRuntimeClientFactoryRequest returned error %v", err)
	}
}

func TestValidateArgoCDRuntimeClientFactoryRequestRejectsMissingArgoCDApplicationName(t *testing.T) {
	request := validArgoCDFactoryRequest()
	request.Target.ArgoCDApplicationName = ""

	if err := ValidateArgoCDRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for missing ArgoCDApplicationName, got nil")
	}
}

func TestValidateArgoCDRuntimeClientFactoryRequestRejectsMissingArgoCDSecretRef(t *testing.T) {
	request := validArgoCDFactoryRequest()
	request.SecretRefs.ArgoCD = nil

	if err := ValidateArgoCDRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for missing Argo CD secret reference, got nil")
	}
}

func TestValidateArgoCDRuntimeClientFactoryRequestRejectsMissingClusterName(t *testing.T) {
	request := validArgoCDFactoryRequest()
	request.SecretRefs.ClusterName = ""

	if err := ValidateArgoCDRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for missing SecretRefs.ClusterName, got nil")
	}
}

func TestValidateArgoCDRuntimeClientFactoryRequestRejectsDisabledCluster(t *testing.T) {
	request := validArgoCDFactoryRequest()
	request.Target.ClusterEnabled = false

	if err := ValidateArgoCDRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for disabled cluster, got nil")
	}
}

func TestRequiredArgoCDRuntimeSecretValueKeysWithTokenBaseURLAndCA(t *testing.T) {
	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-dev",
		ArgoCD: &RuntimeSecretReference{
			Namespace:  "devops-control-plane",
			Name:       "ocp-dev-argocd",
			TokenKey:   "token",
			BaseURLKey: "baseURL",
			CAKey:      "ca.crt",
		},
	}

	keys := RequiredArgoCDRuntimeSecretValueKeys(refs)
	if len(keys) != 3 {
		t.Fatalf("keys len = %d, want 3", len(keys))
	}

	has := map[RuntimeSecretValueKey]bool{}
	for _, key := range keys {
		has[key] = true
	}
	if !has[RuntimeSecretValueArgoCDToken] {
		t.Fatal("expected argocdToken key")
	}
	if !has[RuntimeSecretValueArgoCDBaseURL] {
		t.Fatal("expected argocdBaseURL key")
	}
	if !has[RuntimeSecretValueArgoCDCA] {
		t.Fatal("expected argocdCA key")
	}
}

func TestRequiredArgoCDRuntimeSecretValueKeysReturnsNilWhenArgoCDRefsAreMissing(t *testing.T) {
	if keys := RequiredArgoCDRuntimeSecretValueKeys(RuntimeClientSecretRefs{}); keys != nil {
		t.Fatalf("expected nil keys, got %v", keys)
	}
}

func TestEmptyArgoCDRuntimeClientFactoryIgnoresValidRequest(t *testing.T) {
	var factory ArgoCDRuntimeClientFactory = EmptyArgoCDRuntimeClientFactory{}

	_, err := factory.BuildArgoCDRuntimeClient(context.Background(), validArgoCDFactoryRequest())
	if !errors.Is(err, ErrArgoCDRuntimeClientFactoryNotConfigured) {
		t.Fatalf("BuildArgoCDRuntimeClient error = %v, want ErrArgoCDRuntimeClientFactoryNotConfigured", err)
	}
}
