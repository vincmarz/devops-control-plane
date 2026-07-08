package app

import (
	"context"
	"errors"
	"testing"
)

func TestEmptyKubernetesRuntimeClientFactoryAlwaysReturnsNotConfigured(t *testing.T) {
	var factory KubernetesRuntimeClientFactory = EmptyKubernetesRuntimeClientFactory{}

	client, err := factory.BuildKubernetesRuntimeEvidenceClient(context.Background(), KubernetesRuntimeClientFactoryRequest{})
	if !errors.Is(err, ErrKubernetesRuntimeClientFactoryNotConfigured) {
		t.Fatalf("BuildKubernetesRuntimeEvidenceClient error = %v, want ErrKubernetesRuntimeClientFactoryNotConfigured", err)
	}
	if client != nil {
		t.Fatal("EmptyKubernetesRuntimeClientFactory must never return a client")
	}
}

func validKubernetesFactoryRequest() KubernetesRuntimeClientFactoryRequest {
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
		Kubernetes: &RuntimeSecretReference{
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
	return KubernetesRuntimeClientFactoryRequest{Target: target, SecretRefs: refs, SecretValues: values}
}

func TestValidateKubernetesRuntimeClientFactoryRequestAcceptsValidRequest(t *testing.T) {
	if err := ValidateKubernetesRuntimeClientFactoryRequest(validKubernetesFactoryRequest()); err != nil {
		t.Fatalf("ValidateKubernetesRuntimeClientFactoryRequest returned error %v", err)
	}
}

func TestValidateKubernetesRuntimeClientFactoryRequestRejectsMissingTargetMetadata(t *testing.T) {
	request := validKubernetesFactoryRequest()
	request.Target.KubernetesNamespace = ""

	if err := ValidateKubernetesRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for missing KubernetesNamespace, got nil")
	}
}

func TestValidateKubernetesRuntimeClientFactoryRequestRejectsMissingKubernetesSecretRef(t *testing.T) {
	request := validKubernetesFactoryRequest()
	request.SecretRefs.Kubernetes = nil

	if err := ValidateKubernetesRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for missing Kubernetes secret reference, got nil")
	}
}

func TestValidateKubernetesRuntimeClientFactoryRequestRejectsMissingClusterName(t *testing.T) {
	request := validKubernetesFactoryRequest()
	request.SecretRefs.ClusterName = ""

	if err := ValidateKubernetesRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for missing SecretRefs.ClusterName, got nil")
	}
}

func TestValidateKubernetesRuntimeClientFactoryRequestRejectsDisabledCluster(t *testing.T) {
	request := validKubernetesFactoryRequest()
	request.Target.ClusterEnabled = false

	if err := ValidateKubernetesRuntimeClientFactoryRequest(request); err == nil {
		t.Fatal("expected error for disabled cluster, got nil")
	}
}

func TestRequiredKubernetesRuntimeSecretValueKeysWithTokenAndCA(t *testing.T) {
	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-dev",
		Kubernetes: &RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "ocp-dev-kube",
			TokenKey:  "token",
			CAKey:     "ca.crt",
		},
	}

	keys := RequiredKubernetesRuntimeSecretValueKeys(refs)
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

func TestRequiredKubernetesRuntimeSecretValueKeysWithKubeconfigOnly(t *testing.T) {
	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-dev",
		Kubernetes: &RuntimeSecretReference{
			Namespace:     "devops-control-plane",
			Name:          "ocp-dev-kube",
			KubeconfigKey: "kubeconfig",
		},
	}

	keys := RequiredKubernetesRuntimeSecretValueKeys(refs)
	if len(keys) != 1 || keys[0] != RuntimeSecretValueKubernetesKubeconfig {
		t.Fatalf("keys = %v, want [kubernetesKubeconfig]", keys)
	}
}

func TestRequiredKubernetesRuntimeSecretValueKeysReturnsNilWhenKubernetesRefsAreMissing(t *testing.T) {
	if keys := RequiredKubernetesRuntimeSecretValueKeys(RuntimeClientSecretRefs{}); keys != nil {
		t.Fatalf("expected nil keys, got %v", keys)
	}
}

func TestEmptyKubernetesRuntimeClientFactoryIgnoresValidRequest(t *testing.T) {
	var factory KubernetesRuntimeClientFactory = EmptyKubernetesRuntimeClientFactory{}

	_, err := factory.BuildKubernetesRuntimeEvidenceClient(context.Background(), validKubernetesFactoryRequest())
	if !errors.Is(err, ErrKubernetesRuntimeClientFactoryNotConfigured) {
		t.Fatalf("BuildKubernetesRuntimeEvidenceClient error = %v, want ErrKubernetesRuntimeClientFactoryNotConfigured", err)
	}
}
