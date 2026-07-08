package app

import (
	"context"
	"errors"
	"testing"
)

func TestDisabledKubernetesSecretValueLoaderAlwaysReturnsNotConfigured(t *testing.T) {
	var loader RuntimeSecretValueLoader = DisabledKubernetesSecretValueLoader{}

	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-dev",
		Kubernetes: &RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "ocp-dev-kube",
			TokenKey:  "token",
		},
	}

	set, err := loader.LoadRuntimeSecretValues(context.Background(), refs)
	if !errors.Is(err, ErrRuntimeSecretValueLoaderNotConfigured) {
		t.Fatalf("LoadRuntimeSecretValues error = %v, want ErrRuntimeSecretValueLoaderNotConfigured", err)
	}
	if !set.IsEmpty() {
		t.Fatal("DisabledKubernetesSecretValueLoader must return an empty RuntimeSecretValueSet")
	}
}

func TestValidateKubernetesSecretValueLoaderRequestAcceptsAllowedRef(t *testing.T) {
	config := KubernetesSecretValueLoaderConfig{
		AllowedClusters: []string{"ocp-staging"},
		AllowedRefs: []KubernetesSecretValueLoaderAllowedRef{
			{ClusterName: "ocp-staging", Namespace: "devops-control-plane", Name: "ocp-staging-kube"},
		},
	}

	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-staging",
		Kubernetes: &RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "ocp-staging-kube",
			TokenKey:  "token",
			CAKey:     "ca.crt",
		},
	}

	if err := ValidateKubernetesSecretValueLoaderRequest(config, refs); err != nil {
		t.Fatalf("ValidateKubernetesSecretValueLoaderRequest returned error %v", err)
	}
}

func TestValidateKubernetesSecretValueLoaderRequestRejectsClusterOutsideAllowList(t *testing.T) {
	config := KubernetesSecretValueLoaderConfig{
		AllowedClusters: []string{"ocp-staging"},
		AllowedRefs: []KubernetesSecretValueLoaderAllowedRef{
			{ClusterName: "ocp-staging", Namespace: "devops-control-plane", Name: "ocp-staging-kube"},
		},
	}

	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-production",
		Kubernetes: &RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "ocp-production-kube",
			TokenKey:  "token",
		},
	}

	err := ValidateKubernetesSecretValueLoaderRequest(config, refs)
	if !errors.Is(err, ErrKubernetesSecretValueLoaderClusterNotAllowed) {
		t.Fatalf("ValidateKubernetesSecretValueLoaderRequest error = %v, want ErrKubernetesSecretValueLoaderClusterNotAllowed", err)
	}
}

func TestValidateKubernetesSecretValueLoaderRequestRejectsRefOutsideAllowList(t *testing.T) {
	config := KubernetesSecretValueLoaderConfig{
		AllowedClusters: []string{"ocp-staging"},
		AllowedRefs: []KubernetesSecretValueLoaderAllowedRef{
			{ClusterName: "ocp-staging", Namespace: "devops-control-plane", Name: "ocp-staging-kube"},
		},
	}

	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-staging",
		Kubernetes: &RuntimeSecretReference{
			Namespace: "unexpected-namespace",
			Name:      "ocp-staging-kube",
			TokenKey:  "token",
		},
	}

	err := ValidateKubernetesSecretValueLoaderRequest(config, refs)
	if !errors.Is(err, ErrKubernetesSecretValueLoaderRefNotAllowed) {
		t.Fatalf("ValidateKubernetesSecretValueLoaderRequest error = %v, want ErrKubernetesSecretValueLoaderRefNotAllowed", err)
	}
}

func TestValidateKubernetesSecretValueLoaderRequestRejectsMissingKubernetesRef(t *testing.T) {
	config := KubernetesSecretValueLoaderConfig{
		AllowedClusters: []string{"ocp-staging"},
		AllowedRefs: []KubernetesSecretValueLoaderAllowedRef{
			{ClusterName: "ocp-staging", Namespace: "devops-control-plane", Name: "ocp-staging-kube"},
		},
	}

	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-staging",
	}

	err := ValidateKubernetesSecretValueLoaderRequest(config, refs)
	if !errors.Is(err, ErrKubernetesSecretValueLoaderMissingKubernetesRef) {
		t.Fatalf("ValidateKubernetesSecretValueLoaderRequest error = %v, want ErrKubernetesSecretValueLoaderMissingKubernetesRef", err)
	}
}

func TestValidateKubernetesSecretValueLoaderRequestRejectsMissingClusterName(t *testing.T) {
	config := KubernetesSecretValueLoaderConfig{
		AllowedClusters: []string{"ocp-staging"},
		AllowedRefs: []KubernetesSecretValueLoaderAllowedRef{
			{ClusterName: "ocp-staging", Namespace: "devops-control-plane", Name: "ocp-staging-kube"},
		},
	}

	refs := RuntimeClientSecretRefs{
		Kubernetes: &RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "ocp-staging-kube",
			TokenKey:  "token",
		},
	}

	if err := ValidateKubernetesSecretValueLoaderRequest(config, refs); err == nil {
		t.Fatal("expected error for missing SecretRefs.ClusterName, got nil")
	}
}
