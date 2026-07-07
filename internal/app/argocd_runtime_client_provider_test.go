package app

import (
	"context"
	"strings"
	"testing"
)

type fakeArgoCDRuntimeClient struct {
	called              bool
	lastApplicationName string
}

func (f *fakeArgoCDRuntimeClient) CheckDeployment(ctx context.Context, applicationName string) (ArgoCDDeploymentResult, error) {
	f.called = true
	f.lastApplicationName = applicationName
	return ArgoCDDeploymentResult{ApplicationName: applicationName, Project: "default", SyncStatus: "Synced", HealthStatus: "Healthy", Revision: "abc123"}, nil
}

func TestDefaultArgoCDRuntimeClientProviderRegistryResolvesCurrentProvider(t *testing.T) {
	client := &fakeArgoCDRuntimeClient{}
	registry := DefaultArgoCDRuntimeClientProviderRegistry(client)
	selection := RuntimeClientProviderSelection{
		Target: TechnicalRuntimeTarget{ClusterName: "ocp-dev", ArgoCDApplicationName: "demo-go-color-app"},
		Provider: RuntimeClientProvider{
			ClusterName:    "ocp-dev",
			Enabled:        true,
			CurrentCluster: true,
			ArgoCDProvider: true,
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

func TestDefaultArgoCDRuntimeClientProviderRegistryIsEmptyWhenCurrentClientMissing(t *testing.T) {
	registry := DefaultArgoCDRuntimeClientProviderRegistry(nil)
	selection := RuntimeClientProviderSelection{Provider: RuntimeClientProvider{ClusterName: "ocp-dev"}}

	_, err := registry.Resolve(context.Background(), selection)
	if err == nil {
		t.Fatal("Resolve returned nil error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("Resolve error = %q, want not configured", err.Error())
	}
}

func TestArgoCDRuntimeClientProviderRegistryRejectsUnknownCluster(t *testing.T) {
	registry := DefaultArgoCDRuntimeClientProviderRegistry(&fakeArgoCDRuntimeClient{})
	selection := RuntimeClientProviderSelection{Provider: RuntimeClientProvider{ClusterName: "ocp-staging"}}

	_, err := registry.Resolve(context.Background(), selection)
	if err == nil {
		t.Fatal("Resolve returned nil error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("Resolve error = %q, want not configured", err.Error())
	}
}

func TestCurrentClusterArgoCDRuntimeClientProviderRequiresCurrentClusterProvider(t *testing.T) {
	provider := CurrentClusterArgoCDRuntimeClientProvider(&fakeArgoCDRuntimeClient{})
	selection := RuntimeClientProviderSelection{
		Provider: RuntimeClientProvider{
			ClusterName:    "ocp-dev",
			CurrentCluster: false,
			ArgoCDProvider: true,
		},
	}

	_, err := provider.ResolveArgoCDRuntimeClient(context.Background(), selection)
	if err == nil {
		t.Fatal("ResolveArgoCDRuntimeClient returned nil error")
	}
	if !strings.Contains(err.Error(), "current-cluster") {
		t.Fatalf("error = %q, want current-cluster", err.Error())
	}
}

func TestCurrentClusterArgoCDRuntimeClientProviderRequiresArgoCDCapability(t *testing.T) {
	provider := CurrentClusterArgoCDRuntimeClientProvider(&fakeArgoCDRuntimeClient{})
	selection := RuntimeClientProviderSelection{
		Provider: RuntimeClientProvider{
			ClusterName:    "ocp-dev",
			CurrentCluster: true,
			ArgoCDProvider: false,
		},
	}

	_, err := provider.ResolveArgoCDRuntimeClient(context.Background(), selection)
	if err == nil {
		t.Fatal("ResolveArgoCDRuntimeClient returned nil error")
	}
	if !strings.Contains(err.Error(), "Argo CD capability") {
		t.Fatalf("error = %q, want Argo CD capability", err.Error())
	}
}

func TestArgoCDRuntimeClientInterfaceCanCheckDeployment(t *testing.T) {
	client := &fakeArgoCDRuntimeClient{}
	var runtimeClient ArgoCDRuntimeClient = client

	result, err := runtimeClient.CheckDeployment(context.Background(), "demo-go-color-app")
	if err != nil {
		t.Fatalf("CheckDeployment returned error %v", err)
	}
	if result.ApplicationName != "demo-go-color-app" || result.SyncStatus != "Synced" || result.HealthStatus != "Healthy" {
		t.Fatalf("unexpected deployment result: %+v", result)
	}
	if !client.called || client.lastApplicationName != "demo-go-color-app" {
		t.Fatalf("client call state = called:%v application:%q", client.called, client.lastApplicationName)
	}
}
