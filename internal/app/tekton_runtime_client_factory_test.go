package app

import (
	"context"
	"errors"
	"testing"
)

func TestEmptyTektonRuntimeClientFactoryAlwaysReturnsNotConfigured(t *testing.T) {
	var factory TektonRuntimeClientFactory = EmptyTektonRuntimeClientFactory{}

	client, err := factory.BuildTektonRuntimeClient(context.Background(), TektonRuntimeClientFactoryRequest{})
	if !errors.Is(err, ErrTektonRuntimeClientFactoryNotConfigured) {
		t.Fatalf("BuildTektonRuntimeClient error = %v, want ErrTektonRuntimeClientFactoryNotConfigured", err)
	}
	if client != nil {
		t.Fatal("EmptyTektonRuntimeClientFactory must never return a client")
	}
}

func validTektonFactoryRequest() TektonRuntimeClientFactoryRequest {
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
		Tekton: &RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "ocp-dev-kube",
			TokenKey:  "token",
			CAKey:     "ca.crt",
		},
	}
	values := NewRuntimeSecretValueSet(map[RuntimeSecretValueKey]string{
		RuntimeSecretValueKubernetesToken: "dummy-token-for-tests",
		RuntimeSecretValueKubernetesCA:    "dummy-ca-for-tests",
	})
	return TektonRuntimeClientFactoryRequest{Target: target, SecretRefs: refs, SecretValues: values}
}

func TestValidateTektonRuntimeClientFactoryRequestAcceptsValidRequest(t *testing.T) {
	if err := ValidateTektonRuntimeClientFactoryRequest(validTektonFactoryRequest()); err != nil {
		t.Fatalf("ValidateTektonRuntimeClientFactoryRequest returned error %v", err)
	}
}

func TestValidateTektonRuntimeClientFactoryRequestRejectsMissingTektonNamespace(t *testing.T) {
	request := validTektonFactoryRequest()
	request.Target.TektonNamespace = ""

	if err := ValidateTektonRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for missing TektonNamespace, got nil")
	}
}

func TestValidateTektonRuntimeClientFactoryRequestRejectsMissingTektonPipelineName(t *testing.T) {
	request := validTektonFactoryRequest()
	request.Target.TektonPipelineName = ""

	if err := ValidateTektonRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for missing TektonPipelineName, got nil")
	}
}

func TestValidateTektonRuntimeClientFactoryRequestRejectsMissingTektonSecretRef(t *testing.T) {
	request := validTektonFactoryRequest()
	request.SecretRefs.Tekton = nil

	if err := ValidateTektonRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for missing Tekton secret reference, got nil")
	}
}

func TestValidateTektonRuntimeClientFactoryRequestRejectsMissingClusterName(t *testing.T) {
	request := validTektonFactoryRequest()
	request.SecretRefs.ClusterName = ""

	if err := ValidateTektonRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for missing SecretRefs.ClusterName, got nil")
	}
}

func TestValidateTektonRuntimeClientFactoryRequestRejectsDisabledCluster(t *testing.T) {
	request := validTektonFactoryRequest()
	request.Target.ClusterEnabled = false

	if err := ValidateTektonRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for disabled cluster, got nil")
	}
}

func TestRequiredTektonRuntimeSecretValueKeysWithTokenAndCA(t *testing.T) {
	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-dev",
		Tekton: &RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "ocp-dev-kube",
			TokenKey:  "token",
			CAKey:     "ca.crt",
		},
	}

	keys := RequiredTektonRuntimeSecretValueKeys(refs)
	if len(keys) != 2 {
		t.Fatalf("keys len = %d, want 2", len(keys))
	}

	has := map[RuntimeSecretValueKey]bool{}
	for _, key := range keys {
		has[key] = true
	}
	if !has[RuntimeSecretValueKubernetesToken] {
		t.Fatal("expected kubernetesToken key")
	}
	if !has[RuntimeSecretValueKubernetesCA] {
		t.Fatal("expected kubernetesCA key")
	}
}

func TestRequiredTektonRuntimeSecretValueKeysWithKubeconfigOnly(t *testing.T) {
	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-dev",
		Tekton: &RuntimeSecretReference{
			Namespace:     "devops-control-plane",
			Name:          "ocp-dev-kube",
			KubeconfigKey: "kubeconfig",
		},
	}

	keys := RequiredTektonRuntimeSecretValueKeys(refs)
	if len(keys) != 1 || keys[0] != RuntimeSecretValueKubernetesKubeconfig {
		t.Fatalf("keys = %v, want [kubernetesKubeconfig]", keys)
	}
}

func TestRequiredTektonRuntimeSecretValueKeysReturnsNilWhenTektonRefsAreMissing(t *testing.T) {
	if keys := RequiredTektonRuntimeSecretValueKeys(RuntimeClientSecretRefs{}); keys != nil {
		t.Fatalf("expected nil keys, got %v", keys)
	}
}

func TestEmptyTektonRuntimeClientFactoryIgnoresValidRequest(t *testing.T) {
	var factory TektonRuntimeClientFactory = EmptyTektonRuntimeClientFactory{}

	_, err := factory.BuildTektonRuntimeClient(context.Background(), validTektonFactoryRequest())
	if !errors.Is(err, ErrTektonRuntimeClientFactoryNotConfigured) {
		t.Fatalf("BuildTektonRuntimeClient error = %v, want ErrTektonRuntimeClientFactoryNotConfigured", err)
	}
}
