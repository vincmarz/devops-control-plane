package app

import "testing"

func TestRuntimeClientProviderRegistrySelectWithSecretRefs(t *testing.T) {
	target := TechnicalRuntimeTarget{TargetEnvironment: "dev", EnvironmentName: "dev", ClusterName: "ocp-dev", ClusterEnabled: true, KubernetesNamespace: "devops-ci-demo", TektonNamespace: "devops-ci-demo", TektonPipelineName: "validate-gitops", ArgoCDApplicationName: "demo-go-color-app", GitTargetBranch: "main"}
	refsRegistry, err := NewRuntimeClientSecretRefsRegistry([]RuntimeClientSecretRefs{{ClusterName: "ocp-dev", Kubernetes: &RuntimeSecretReference{Namespace: "devops-control-plane", Name: "ocp-dev-kube", TokenKey: "token"}}})
	if err != nil {
		t.Fatalf("NewRuntimeClientSecretRefsRegistry returned error %v", err)
	}
	selection, err := DefaultRuntimeClientProviderRegistry().SelectWithSecretRefs(target, refsRegistry)
	if err != nil {
		t.Fatalf("SelectWithSecretRefs returned error %v", err)
	}
	if !selection.SecretRefsConfigured {
		t.Fatal("SecretRefsConfigured = false, want true")
	}
	if selection.SecretRefs.Kubernetes == nil || selection.SecretRefs.Kubernetes.Name != "ocp-dev-kube" {
		t.Fatalf("unexpected kubernetes secret refs: %+v", selection.SecretRefs.Kubernetes)
	}
}

func TestRuntimeClientProviderRegistrySelectWithSecretRefsAllowsMissingRefs(t *testing.T) {
	target := TechnicalRuntimeTarget{TargetEnvironment: "dev", EnvironmentName: "dev", ClusterName: "ocp-dev", ClusterEnabled: true, KubernetesNamespace: "devops-ci-demo", TektonNamespace: "devops-ci-demo", TektonPipelineName: "validate-gitops", ArgoCDApplicationName: "demo-go-color-app", GitTargetBranch: "main"}
	selection, err := DefaultRuntimeClientProviderRegistry().SelectWithSecretRefs(target, EmptyRuntimeClientSecretRefsRegistry())
	if err != nil {
		t.Fatalf("SelectWithSecretRefs returned error %v", err)
	}
	if selection.SecretRefsConfigured {
		t.Fatal("SecretRefsConfigured = true, want false")
	}
}
