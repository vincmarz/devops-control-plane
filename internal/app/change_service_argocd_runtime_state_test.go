package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func TestCheckDeploymentPersistsArgoCDStateBeforeMarkStep(t *testing.T) {
	changeStore := &argocdFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-30", ApplicationName: "demo-go-color-app", TargetEnvironment: "staging"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: domain.ChangeRuntimeState{
		GitOps: domain.GitOpsRuntimeState{
			Provider:      "github",
			ProviderRef:   "github-public",
			ProjectPath:   "vincmarz/demo-app-gitops",
			RepositoryURL: "https://github.com/vincmarz/demo-app-gitops.git",
			DefaultBranch: "main",
			Revision:      "main",
		},
	}}
	service := NewChangeService(
		changeStore,
		WithChangeRuntimeStateStore(runtimeStore),
		WithArgoCDCheckDeployment(func(context.Context, domain.ChangeRequest) (ArgoCDDeploymentResult, error) {
			return ArgoCDDeploymentResult{
				ApplicationName:       "demo-go-color-app-staging",
				Project:               "default",
				SyncStatus:            "Synced",
				HealthStatus:          "Healthy",
				Revision:              "observed-sha",
				RepositoryURL:         "https://github.com/vincmarz/demo-app-gitops.git",
				TargetRevision:        "main",
				GitOpsProvider:        "github",
				GitOpsProviderRef:     "github-public",
				GitOpsProjectPath:     "vincmarz/demo-app-gitops",
				DeclaredRepositoryURL: "https://github.com/vincmarz/demo-app-gitops.git",
				DeclaredDefaultBranch: "main",
			}, nil
		}),
	)

	if _, err := service.CheckDeployment(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	if !runtimeStore.gitOpsCalled {
		t.Fatal("GitOps runtime state was not updated with the observed commit")
	}
	if runtimeStore.gitOps.CommitSHA != "observed-sha" || runtimeStore.gitOps.Revision != "main" {
		t.Fatalf("GitOps observed commit state = %#v", runtimeStore.gitOps)
	}
	if runtimeStore.gitOps.Provider != "github" || runtimeStore.gitOps.ProviderRef != "github-public" || runtimeStore.gitOps.ProjectPath != "vincmarz/demo-app-gitops" {
		t.Fatalf("GitOps identity state = %#v", runtimeStore.gitOps)
	}
	if !runtimeStore.argoCDCalled {
		t.Fatal("Argo CD runtime state was not persisted")
	}
	got := runtimeStore.argoCD
	if got.ApplicationName != "demo-go-color-app-staging" || got.Provider != "github" || got.ProviderRef != "github-public" {
		t.Fatalf("Argo CD identity state = %#v", got)
	}
	if got.DeclaredRepositoryURL != got.ObservedRepositoryURL || got.CorrelationStatus != "Matched" {
		t.Fatalf("Argo CD repository correlation = %#v", got)
	}
	if got.ObservedRevision != "observed-sha" || got.SyncStatus != "Synced" || got.HealthStatus != "Healthy" {
		t.Fatalf("Argo CD observed state = %#v", got)
	}
	if changeStore.markedStatus != "DeploymentSyncedHealthy" {
		t.Fatalf("marked status = %q", changeStore.markedStatus)
	}
}

func TestCheckDeploymentDoesNotMarkStepWhenArgoCDPersistenceFails(t *testing.T) {
	changeStore := &argocdFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-31", ApplicationName: "demo-go-color-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{argoCDErr: errors.New("database unavailable")}
	service := NewChangeService(
		changeStore,
		WithChangeRuntimeStateStore(runtimeStore),
		WithArgoCDCheckDeployment(func(context.Context, domain.ChangeRequest) (ArgoCDDeploymentResult, error) {
			return ArgoCDDeploymentResult{ApplicationName: "demo-go-color-app", SyncStatus: "Synced", HealthStatus: "Healthy"}, nil
		}),
	)

	_, err := service.CheckDeployment(context.Background(), changeStore.change.ChangeNumber)
	if err == nil || !strings.Contains(err.Error(), "persist Argo CD runtime state after checking deployment") {
		t.Fatalf("error = %v", err)
	}
	if changeStore.markedStatus != "" {
		t.Fatalf("MarkStep was called with %q", changeStore.markedStatus)
	}
}

func TestCheckDeploymentDoesNotMarkStepWhenGitOpsCommitPersistenceFails(t *testing.T) {
	changeStore := &argocdFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-32", ApplicationName: "demo-go-color-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{
		current:   domain.ChangeRuntimeState{GitOps: domain.GitOpsRuntimeState{Revision: "main"}},
		gitOpsErr: errors.New("database unavailable"),
	}
	service := NewChangeService(
		changeStore,
		WithChangeRuntimeStateStore(runtimeStore),
		WithArgoCDCheckDeployment(func(context.Context, domain.ChangeRequest) (ArgoCDDeploymentResult, error) {
			return ArgoCDDeploymentResult{
				ApplicationName:       "demo-go-color-app",
				SyncStatus:            "Synced",
				HealthStatus:          "Healthy",
				Revision:              "observed-sha",
				RepositoryURL:         "https://github.com/vincmarz/demo-app-gitops.git",
				DeclaredRepositoryURL: "https://github.com/vincmarz/demo-app-gitops.git",
			}, nil
		}),
	)

	_, err := service.CheckDeployment(context.Background(), changeStore.change.ChangeNumber)
	if err == nil || !strings.Contains(err.Error(), "persist observed GitOps commit after checking deployment") {
		t.Fatalf("error = %v", err)
	}
	if runtimeStore.argoCDCalled {
		t.Fatal("Argo CD runtime state was persisted after GitOps commit persistence failed")
	}
	if changeStore.markedStatus != "" {
		t.Fatalf("MarkStep was called with %q", changeStore.markedStatus)
	}
}
