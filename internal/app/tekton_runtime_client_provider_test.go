package app

import (
	"context"
	"strings"
	"testing"
)

type fakeTektonRuntimeClient struct {
	createCalled         bool
	lastCreateRequest    TektonRuntimePipelineRunRequest
	findLatestCalled     bool
	lastFindNamespace    string
	lastFindChangeNumber string
	listTaskRunsCalled   bool
	lastTaskRunNamespace string
	lastPipelineRunName  string
}

func (f *fakeTektonRuntimeClient) CreatePipelineRun(ctx context.Context, request TektonRuntimePipelineRunRequest) (TektonRuntimePipelineRunRef, error) {
	f.createCalled = true
	f.lastCreateRequest = request
	return TektonRuntimePipelineRunRef{Name: "pr-1", Namespace: request.Namespace, UID: "uid-1"}, nil
}

func (f *fakeTektonRuntimeClient) FindLatestPipelineRunByChange(ctx context.Context, namespace string, changeNumber string) (TektonValidationResult, error) {
	f.findLatestCalled = true
	f.lastFindNamespace = namespace
	f.lastFindChangeNumber = changeNumber
	return TektonValidationResult{PipelineRunName: "pr-1", Namespace: namespace, Status: "True", Reason: "Succeeded"}, nil
}

func (f *fakeTektonRuntimeClient) ListTaskRunsByPipelineRun(ctx context.Context, namespace string, pipelineRunName string) ([]TektonTaskRunResult, error) {
	f.listTaskRunsCalled = true
	f.lastTaskRunNamespace = namespace
	f.lastPipelineRunName = pipelineRunName
	return []TektonTaskRunResult{{Name: "task-1", Status: "True", Reason: "Succeeded"}}, nil
}

func TestDefaultTektonRuntimeClientProviderRegistryResolvesCurrentProvider(t *testing.T) {
	client := &fakeTektonRuntimeClient{}
	registry := DefaultTektonRuntimeClientProviderRegistry(client)
	selection := RuntimeClientProviderSelection{
		Target: TechnicalRuntimeTarget{ClusterName: "ocp-dev", TektonNamespace: "devops-ci-demo", TektonPipelineName: "validate-gitops"},
		Provider: RuntimeClientProvider{
			ClusterName:    "ocp-dev",
			Enabled:        true,
			CurrentCluster: true,
			TektonProvider: true,
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

func TestDefaultTektonRuntimeClientProviderRegistryIsEmptyWhenCurrentClientMissing(t *testing.T) {
	registry := DefaultTektonRuntimeClientProviderRegistry(nil)
	selection := RuntimeClientProviderSelection{Provider: RuntimeClientProvider{ClusterName: "ocp-dev"}}

	_, err := registry.Resolve(context.Background(), selection)
	if err == nil {
		t.Fatal("Resolve returned nil error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("Resolve error = %q, want not configured", err.Error())
	}
}

func TestTektonRuntimeClientProviderRegistryRejectsUnknownCluster(t *testing.T) {
	registry := DefaultTektonRuntimeClientProviderRegistry(&fakeTektonRuntimeClient{})
	selection := RuntimeClientProviderSelection{Provider: RuntimeClientProvider{ClusterName: "ocp-staging"}}

	_, err := registry.Resolve(context.Background(), selection)
	if err == nil {
		t.Fatal("Resolve returned nil error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("Resolve error = %q, want not configured", err.Error())
	}
}

func TestCurrentClusterTektonRuntimeClientProviderRequiresCurrentClusterProvider(t *testing.T) {
	provider := CurrentClusterTektonRuntimeClientProvider(&fakeTektonRuntimeClient{})
	selection := RuntimeClientProviderSelection{
		Provider: RuntimeClientProvider{
			ClusterName:    "ocp-dev",
			CurrentCluster: false,
			TektonProvider: true,
		},
	}

	_, err := provider.ResolveTektonRuntimeClient(context.Background(), selection)
	if err == nil {
		t.Fatal("ResolveTektonRuntimeClient returned nil error")
	}
	if !strings.Contains(err.Error(), "current-cluster") {
		t.Fatalf("error = %q, want current-cluster", err.Error())
	}
}

func TestCurrentClusterTektonRuntimeClientProviderRequiresTektonCapability(t *testing.T) {
	provider := CurrentClusterTektonRuntimeClientProvider(&fakeTektonRuntimeClient{})
	selection := RuntimeClientProviderSelection{
		Provider: RuntimeClientProvider{
			ClusterName:    "ocp-dev",
			CurrentCluster: true,
			TektonProvider: false,
		},
	}

	_, err := provider.ResolveTektonRuntimeClient(context.Background(), selection)
	if err == nil {
		t.Fatal("ResolveTektonRuntimeClient returned nil error")
	}
	if !strings.Contains(err.Error(), "Tekton capability") {
		t.Fatalf("error = %q, want Tekton capability", err.Error())
	}
}

func TestTektonRuntimeClientInterfaceCanRunAndInspectPipelineRuns(t *testing.T) {
	client := &fakeTektonRuntimeClient{}
	var tektonClient TektonRuntimeClient = client

	ref, err := tektonClient.CreatePipelineRun(context.Background(), TektonRuntimePipelineRunRequest{
		Namespace:          "devops-ci-demo",
		PipelineName:       "validate-gitops",
		ChangeNumber:       "CHG-2026-0001",
		ApplicationName:    "demo-go-color-app",
		GitURL:             "https://example.invalid/repo.git",
		GitRevision:        "main",
		ValidationPath:     "manifests",
		ServiceAccountName: "pipeline",
	})
	if err != nil {
		t.Fatalf("CreatePipelineRun returned error %v", err)
	}
	if ref.Name != "pr-1" || ref.Namespace != "devops-ci-demo" || ref.UID != "uid-1" {
		t.Fatalf("unexpected ref: %+v", ref)
	}
	if !client.createCalled || client.lastCreateRequest.PipelineName != "validate-gitops" {
		t.Fatalf("create call state unexpected: called=%v request=%+v", client.createCalled, client.lastCreateRequest)
	}

	status, err := tektonClient.FindLatestPipelineRunByChange(context.Background(), "devops-ci-demo", "CHG-2026-0001")
	if err != nil {
		t.Fatalf("FindLatestPipelineRunByChange returned error %v", err)
	}
	if status.PipelineRunName != "pr-1" || status.Status != "True" {
		t.Fatalf("unexpected status: %+v", status)
	}

	taskRuns, err := tektonClient.ListTaskRunsByPipelineRun(context.Background(), "devops-ci-demo", "pr-1")
	if err != nil {
		t.Fatalf("ListTaskRunsByPipelineRun returned error %v", err)
	}
	if len(taskRuns) != 1 || taskRuns[0].Name != "task-1" {
		t.Fatalf("unexpected taskRuns: %+v", taskRuns)
	}
}
