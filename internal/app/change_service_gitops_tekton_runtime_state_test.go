package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func gitOpsTektonTestResult() TektonPipelineRunResult {
	return TektonPipelineRunResult{
		PipelineRunName: "devops-cp-validate-chg-20",
		Namespace:       "devops-ci-demo",
		UID:             "uid-20",
		PipelineName:    "validate-gitops",
		GitOpsTarget: GitOpsRepositoryTarget{
			Provider: "github", ProviderRef: "github-public", ProjectPath: "vincmarz/demo-app-gitops",
			RepositoryURL: "https://github.com/vincmarz/demo-app-gitops.git", DefaultBranch: "main",
		},
		GitRevision:    "main",
		ValidationPath: "apps/demo-go-color-app",
	}
}

func TestValidatePersistsGitOpsAndTektonBeforeMarkStep(t *testing.T) {
	changeStore := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-20", ApplicationName: "demo-go-color-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{}
	result := gitOpsTektonTestResult()
	svc := NewChangeService(
		changeStore,
		WithChangeRuntimeStateStore(runtimeStore),
		WithTektonRunPipeline(func(context.Context, domain.ChangeRequest) (TektonPipelineRunResult, error) { return result, nil }),
	)

	if _, err := svc.Validate(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	if !runtimeStore.gitOpsCalled || !runtimeStore.tektonCalled {
		t.Fatalf("runtime persistence calls: GitOps=%v Tekton=%v", runtimeStore.gitOpsCalled, runtimeStore.tektonCalled)
	}
	if runtimeStore.gitOps.RepositoryURL != result.GitOpsTarget.RepositoryURL || runtimeStore.gitOps.Revision != "main" {
		t.Fatalf("GitOps state = %#v", runtimeStore.gitOps)
	}
	if runtimeStore.tekton.PipelineRunName != result.PipelineRunName || runtimeStore.tekton.Status != "Unknown" || runtimeStore.tekton.Reason != "Running" {
		t.Fatalf("Tekton state = %#v", runtimeStore.tekton)
	}
	if !changeStore.markStepCalled || changeStore.markedStatus != "ValidationRunning" {
		t.Fatalf("MarkStep state = called:%v status:%q", changeStore.markStepCalled, changeStore.markedStatus)
	}
}

func TestValidateDoesNotMarkStepWhenGitOpsPersistenceFails(t *testing.T) {
	changeStore := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-21"}}
	runtimeStore := &sourceRuntimeStateStoreFake{gitOpsErr: errors.New("database unavailable")}
	svc := NewChangeService(
		changeStore,
		WithChangeRuntimeStateStore(runtimeStore),
		WithTektonRunPipeline(func(context.Context, domain.ChangeRequest) (TektonPipelineRunResult, error) {
			return gitOpsTektonTestResult(), nil
		}),
	)

	_, err := svc.Validate(context.Background(), changeStore.change.ChangeNumber)
	if err == nil || !strings.Contains(err.Error(), "persist GitOps runtime state after starting Tekton validation") {
		t.Fatalf("error = %v", err)
	}
	if changeStore.markStepCalled {
		t.Fatal("MarkStep was called after GitOps persistence failure")
	}
	if runtimeStore.tektonCalled {
		t.Fatal("Tekton state was persisted after GitOps persistence failure")
	}
}

func TestValidateDoesNotMarkStepWhenTektonPersistenceFails(t *testing.T) {
	changeStore := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-22"}}
	runtimeStore := &sourceRuntimeStateStoreFake{tektonErr: errors.New("database unavailable")}
	svc := NewChangeService(
		changeStore,
		WithChangeRuntimeStateStore(runtimeStore),
		WithTektonRunPipeline(func(context.Context, domain.ChangeRequest) (TektonPipelineRunResult, error) {
			return gitOpsTektonTestResult(), nil
		}),
	)

	_, err := svc.Validate(context.Background(), changeStore.change.ChangeNumber)
	if err == nil || !strings.Contains(err.Error(), "persist Tekton runtime state after starting validation") {
		t.Fatalf("error = %v", err)
	}
	if changeStore.markStepCalled {
		t.Fatal("MarkStep was called after Tekton persistence failure")
	}
}

func TestCheckValidationPreservesExistingTektonMetadata(t *testing.T) {
	changeStore := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-23"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: domain.ChangeRuntimeState{Tekton: domain.TektonRuntimeState{
		PipelineName: "validate-gitops", GitURL: "https://github.com/vincmarz/demo-app-gitops.git",
		GitRevision: "main", ValidationPath: "apps/demo-go-color-app",
	}}}
	svc := NewChangeService(
		changeStore,
		WithChangeRuntimeStateStore(runtimeStore),
		WithTektonCheckValidation(func(context.Context, domain.ChangeRequest) (TektonValidationResult, error) {
			return TektonValidationResult{PipelineRunName: "pr-23", Namespace: "devops-ci-demo", UID: "uid-23", Status: "True", Reason: "Succeeded", Message: "Tasks Completed: 2"}, nil
		}),
	)

	if _, err := svc.CheckValidation(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	got := runtimeStore.tekton
	if got.PipelineName != "validate-gitops" || got.GitURL == "" || got.GitRevision != "main" || got.ValidationPath == "" {
		t.Fatalf("existing Tekton metadata was not preserved: %#v", got)
	}
	if got.PipelineRunName != "pr-23" || got.Status != "True" || got.Reason != "Succeeded" {
		t.Fatalf("observed Tekton metadata = %#v", got)
	}
	if changeStore.markedStatus != "ValidationSucceeded" {
		t.Fatalf("marked status = %q", changeStore.markedStatus)
	}
}

func TestCheckValidationDoesNotMarkStepWhenTektonPersistenceFails(t *testing.T) {
	changeStore := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-24"}}
	runtimeStore := &sourceRuntimeStateStoreFake{tektonErr: errors.New("database unavailable")}
	svc := NewChangeService(
		changeStore,
		WithChangeRuntimeStateStore(runtimeStore),
		WithTektonCheckValidation(func(context.Context, domain.ChangeRequest) (TektonValidationResult, error) {
			return TektonValidationResult{PipelineRunName: "pr-24", Namespace: "devops-ci-demo", Status: "False", Reason: "Failed"}, nil
		}),
	)

	_, err := svc.CheckValidation(context.Background(), changeStore.change.ChangeNumber)
	if err == nil || !strings.Contains(err.Error(), "persist Tekton runtime state after checking validation") {
		t.Fatalf("error = %v", err)
	}
	if changeStore.markStepCalled {
		t.Fatal("MarkStep was called after Tekton persistence failure")
	}
}
