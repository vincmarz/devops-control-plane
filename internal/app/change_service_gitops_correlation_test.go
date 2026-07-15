package app

import (
	"context"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func TestCheckDeploymentReportsMatchedGitOpsRepository(t *testing.T) {
	store := &argocdFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-1", ApplicationName: "demo-go-color-app"}}
	service := NewChangeService(store, WithArgoCDCheckDeployment(func(context.Context, domain.ChangeRequest) (ArgoCDDeploymentResult, error) {
		return ArgoCDDeploymentResult{
			ApplicationName: "demo-go-color-app", SyncStatus: "Synced", HealthStatus: "Healthy",
			RepositoryURL: "https://github.com/vincmarz/demo-app-gitops/", TargetRevision: "main",
			GitOpsProvider: "github", GitOpsProviderRef: "github-public", GitOpsProjectPath: "vincmarz/demo-app-gitops",
			DeclaredRepositoryURL: "https://github.com/vincmarz/demo-app-gitops.git", DeclaredDefaultBranch: "main",
		}, nil
	}))
	result, err := service.CheckDeployment(context.Background(), "CHG-1")
	if err != nil {
		t.Fatal(err)
	}
	gitops := result["argocd"].(map[string]any)["gitops"].(map[string]any)
	if gitops["correlationStatus"] != "Matched" || gitops["providerRef"] != "github-public" {
		t.Fatalf("gitops = %#v", gitops)
	}
}

func TestCheckDeploymentReportsGitOpsRepositoryMismatch(t *testing.T) {
	store := &argocdFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-2", ApplicationName: "demo-go-color-app"}}
	service := NewChangeService(store, WithArgoCDCheckDeployment(func(context.Context, domain.ChangeRequest) (ArgoCDDeploymentResult, error) {
		return ArgoCDDeploymentResult{
			ApplicationName: "demo-go-color-app", SyncStatus: "Synced", HealthStatus: "Healthy",
			RepositoryURL:         "https://gitlab.example/group/demo-app-gitops.git",
			DeclaredRepositoryURL: "https://github.com/vincmarz/demo-app-gitops.git",
		}, nil
	}))
	result, err := service.CheckDeployment(context.Background(), "CHG-2")
	if err != nil {
		t.Fatal(err)
	}
	gitops := result["argocd"].(map[string]any)["gitops"].(map[string]any)
	if gitops["correlationStatus"] != "Mismatch" {
		t.Fatalf("gitops = %#v", gitops)
	}
}
