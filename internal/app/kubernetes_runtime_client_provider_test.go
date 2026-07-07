package app

import (
	"context"
	"strings"
	"testing"
)

type fakeKubernetesRuntimeEvidenceClient struct {
	called          bool
	lastNamespace   string
	lastApplication string
}

func (f *fakeKubernetesRuntimeEvidenceClient) CollectRuntimeEvidence(ctx context.Context, namespace string, applicationName string) (map[string]any, error) {
	f.called = true
	f.lastNamespace = namespace
	f.lastApplication = applicationName
	return map[string]any{"summary": "ok"}, nil
}

func TestDefaultKubernetesRuntimeClientProviderRegistryResolvesCurrentProvider(t *testing.T) {
	client := &fakeKubernetesRuntimeEvidenceClient{}
	registry := DefaultKubernetesRuntimeClientProviderRegistry(client)
	selection := RuntimeClientProviderSelection{
		Target: TechnicalRuntimeTarget{ClusterName: "ocp-dev", KubernetesNamespace: "devops-ci-demo"},
		Provider: RuntimeClientProvider{
			ClusterName:        "ocp-dev",
			Enabled:            true,
			CurrentCluster:     true,
			KubernetesProvider: true,
		},
	}

	resolved, err := registry.Resolve(context.Background(), selection)
	if err != nil {
		t.Fatalf("Resolve returned error %v", err)
	}
	if resolved != client {
		t.Fatal("resolved client does not match current client")
	}
}

func TestDefaultKubernetesRuntimeClientProviderRegistryIsEmptyWhenCurrentClientMissing(t *testing.T) {
	registry := DefaultKubernetesRuntimeClientProviderRegistry(nil)
	selection := RuntimeClientProviderSelection{Provider: RuntimeClientProvider{ClusterName: "ocp-dev"}}

	_, err := registry.Resolve(context.Background(), selection)
	if err == nil {
		t.Fatal("Resolve returned nil error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("Resolve error = %q, want not configured", err.Error())
	}
}

func TestKubernetesRuntimeClientProviderRegistryRejectsUnknownCluster(t *testing.T) {
	registry := DefaultKubernetesRuntimeClientProviderRegistry(&fakeKubernetesRuntimeEvidenceClient{})
	selection := RuntimeClientProviderSelection{Provider: RuntimeClientProvider{ClusterName: "ocp-staging"}}

	_, err := registry.Resolve(context.Background(), selection)
	if err == nil {
		t.Fatal("Resolve returned nil error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("Resolve error = %q, want not configured", err.Error())
	}
}

func TestCurrentClusterKubernetesRuntimeClientProviderRequiresCurrentClusterProvider(t *testing.T) {
	provider := CurrentClusterKubernetesRuntimeClientProvider(&fakeKubernetesRuntimeEvidenceClient{})
	selection := RuntimeClientProviderSelection{
		Provider: RuntimeClientProvider{
			ClusterName:        "ocp-dev",
			CurrentCluster:     false,
			KubernetesProvider: true,
		},
	}

	_, err := provider.ResolveKubernetesRuntimeEvidenceClient(context.Background(), selection)
	if err == nil {
		t.Fatal("ResolveKubernetesRuntimeEvidenceClient returned nil error")
	}
	if !strings.Contains(err.Error(), "current-cluster") {
		t.Fatalf("error = %q, want current-cluster", err.Error())
	}
}

func TestCurrentClusterKubernetesRuntimeClientProviderRequiresKubernetesCapability(t *testing.T) {
	provider := CurrentClusterKubernetesRuntimeClientProvider(&fakeKubernetesRuntimeEvidenceClient{})
	selection := RuntimeClientProviderSelection{
		Provider: RuntimeClientProvider{
			ClusterName:        "ocp-dev",
			CurrentCluster:     true,
			KubernetesProvider: false,
		},
	}

	_, err := provider.ResolveKubernetesRuntimeEvidenceClient(context.Background(), selection)
	if err == nil {
		t.Fatal("ResolveKubernetesRuntimeEvidenceClient returned nil error")
	}
	if !strings.Contains(err.Error(), "Kubernetes capability") {
		t.Fatalf("error = %q, want Kubernetes capability", err.Error())
	}
}

func TestKubernetesRuntimeEvidenceClientInterfaceCanCollectEvidence(t *testing.T) {
	client := &fakeKubernetesRuntimeEvidenceClient{}
	var evidenceClient KubernetesRuntimeEvidenceClient = client

	result, err := evidenceClient.CollectRuntimeEvidence(context.Background(), "devops-ci-demo", "demo-go-color-app")
	if err != nil {
		t.Fatalf("CollectRuntimeEvidence returned error %v", err)
	}
	if result["summary"] != "ok" {
		t.Fatalf("summary = %v, want ok", result["summary"])
	}
	if !client.called || client.lastNamespace != "devops-ci-demo" || client.lastApplication != "demo-go-color-app" {
		t.Fatalf("client call state = called:%v namespace:%q application:%q", client.called, client.lastNamespace, client.lastApplication)
	}
}
