package app

import (
	"strings"
	"testing"
)

func TestRuntimeClientSecretRefsValidateAllowsReferencesOnly(t *testing.T) {
	refs := RuntimeClientSecretRefs{
		ClusterName: "ocp-staging",
		Kubernetes:  &RuntimeSecretReference{Namespace: "devops-control-plane", Name: "ocp-staging-kube", TokenKey: "token", CAKey: "ca.crt"},
		Tekton:      &RuntimeSecretReference{Namespace: "devops-control-plane", Name: "ocp-staging-kube", TokenKey: "token", CAKey: "ca.crt"},
		ArgoCD:      &RuntimeSecretReference{Namespace: "devops-control-plane", Name: "ocp-staging-argocd", TokenKey: "token", BaseURLKey: "baseURL", CAKey: "ca.crt"},
	}

	if err := refs.Validate(); err != nil {
		t.Fatalf("Validate returned error %v", err)
	}
}

func TestRuntimeClientSecretRefsRejectsMissingClusterName(t *testing.T) {
	refs := RuntimeClientSecretRefs{}

	err := refs.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	if !strings.Contains(err.Error(), "clusterName") {
		t.Fatalf("Validate error = %q, want clusterName", err.Error())
	}
}

func TestRuntimeSecretReferenceRequiresNameNamespaceAndCredentialKey(t *testing.T) {
	cases := []struct {
		name string
		ref  RuntimeSecretReference
		want string
	}{
		{name: "missing namespace", ref: RuntimeSecretReference{Name: "cluster-secret", TokenKey: "token"}, want: "namespace"},
		{name: "missing name", ref: RuntimeSecretReference{Namespace: "devops-control-plane", TokenKey: "token"}, want: "name"},
		{name: "missing credential key", ref: RuntimeSecretReference{Namespace: "devops-control-plane", Name: "cluster-secret"}, want: "tokenKey or kubeconfigKey"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ref.Validate("kubernetes")
			if err == nil {
				t.Fatal("Validate returned nil error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Validate error = %q, want %q", err.Error(), tc.want)
			}
		})
	}
}

func TestRuntimeSecretReferenceRejectsSecretLikeValues(t *testing.T) {
	refs := []RuntimeSecretReference{
		{Namespace: "devops-control-plane", Name: "cluster-secret", TokenKey: "Bearer abc"},
		{Namespace: "devops-control-plane", Name: "cluster-secret", TokenKey: "token", CAKey: "-----BEGIN CERTIFICATE-----"},
		{Namespace: "devops-control-plane", Name: "cluster-secret", KubeconfigKey: "client-certificate-data: abc"},
	}

	for _, ref := range refs {
		err := ref.Validate("kubernetes")
		if err == nil {
			t.Fatalf("Validate returned nil error for ref %+v", ref)
		}
		if !strings.Contains(err.Error(), "secret value") {
			t.Fatalf("Validate error = %q, want secret value", err.Error())
		}
	}
}

func TestRuntimeClientSecretRefsSafeSummaryDoesNotExposeValues(t *testing.T) {
	refs := RuntimeClientSecretRefs{
		ClusterName: " OCP-STAGING ",
		Kubernetes:  &RuntimeSecretReference{Namespace: "devops-control-plane", Name: "ocp-staging-kube", TokenKey: "token", CAKey: "ca.crt"},
	}

	summary := refs.SafeSummary()
	if summary["clusterName"] != "ocp-staging" {
		t.Fatalf("clusterName summary = %v, want ocp-staging", summary["clusterName"])
	}
	kubernetes, ok := summary["kubernetes"].(map[string]any)
	if !ok {
		t.Fatalf("kubernetes summary has type %T", summary["kubernetes"])
	}
	if kubernetes["configured"] != true {
		t.Fatalf("configured = %v, want true", kubernetes["configured"])
	}
	if kubernetes["tokenKey"] != "token" {
		t.Fatalf("tokenKey = %v, want token", kubernetes["tokenKey"])
	}
}
