package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func TestOpenMergeRequestPersistsIncrementalGitHubProposalState(t *testing.T) {
	changeStore := &createBranchFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-10", ApplicationName: "github-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: domain.ChangeRuntimeState{
		ChangeRequestID: "change-id",
		Source: domain.SourceRuntimeState{
			Provider: "github", ProviderRef: "github-public", ProjectPath: "org/app",
			RepositoryURL: "https://github.com/org/app.git", DefaultBranch: "main",
			Branch: "change/CHG-10", CommitSHA: "existing-source-sha",
		},
	}}
	binding := RepositoryBinding{
		Provider: "github", ProviderRef: "github-public", Role: RepositoryRoleSource,
		ProjectPath: "org/app", RepositoryURL: "https://github.com/org/app.git",
		DefaultBranch: "main", WorkflowEnabled: true,
	}
	service := providerAwareServiceForRuntimeStateTest(t, changeStore, runtimeStore, binding)

	if _, err := service.OpenMergeRequest(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	got := runtimeStore.state
	if got.CommitSHA != "existing-source-sha" {
		t.Fatalf("commit SHA was not preserved: %#v", got)
	}
	if got.TargetBranch != "main" || got.ProposalNumber != 7 || got.ProposalURL != "https://example.test/mr/7" || got.ProposalState != "open" {
		t.Fatalf("proposal state = %#v", got)
	}
	if got.MergeCommitSHA != "" {
		t.Fatalf("unexpected merge commit SHA: %#v", got)
	}
	if !changeStore.markStepCalled || changeStore.markedStatus != "MergeRequestOpened" {
		t.Fatalf("MarkStep state = called:%v status:%q", changeStore.markStepCalled, changeStore.markedStatus)
	}
}

func TestMergeRequestPersistsIncrementalGitLabProposalState(t *testing.T) {
	changeStore := &createBranchFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-11", ApplicationName: "gitlab-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: domain.ChangeRuntimeState{
		ChangeRequestID: "change-id",
		Source: domain.SourceRuntimeState{
			Provider: "gitlab", ProviderRef: "gitlab-lab", ProjectID: 42, ProjectPath: "group/app",
			RepositoryURL: "https://gitlab.example/group/app.git", DefaultBranch: "trunk",
			Branch: "change/CHG-11", CommitSHA: "existing-source-sha",
			TargetBranch: "trunk", ProposalNumber: 7, ProposalURL: "https://example.test/mr/7", ProposalState: "open",
		},
	}}
	binding := RepositoryBinding{
		Provider: "gitlab", ProviderRef: "gitlab-lab", Role: RepositoryRoleSource,
		ProjectID: 42, ProjectPath: "group/app", RepositoryURL: "https://gitlab.example/group/app.git",
		DefaultBranch: "trunk", WorkflowEnabled: true,
	}
	service := providerAwareServiceForRuntimeStateTest(t, changeStore, runtimeStore, binding)

	if _, err := service.MergeRequest(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	got := runtimeStore.state
	if got.CommitSHA != "existing-source-sha" {
		t.Fatalf("commit SHA was not preserved: %#v", got)
	}
	if got.TargetBranch != "trunk" || got.ProposalNumber != 7 || got.ProposalURL != "https://example.test/mr/7" || got.ProposalState != "merged" {
		t.Fatalf("proposal state = %#v", got)
	}
	if got.MergeCommitSHA != "abc123" {
		t.Fatalf("merge commit SHA = %q", got.MergeCommitSHA)
	}
	if !changeStore.markStepCalled || changeStore.markedStatus != "MergeRequestMerged" {
		t.Fatalf("MarkStep state = called:%v status:%q", changeStore.markStepCalled, changeStore.markedStatus)
	}
}

func TestOpenMergeRequestReconstructsMissingSourceState(t *testing.T) {
	changeStore := &createBranchFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-12", ApplicationName: "github-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: domain.ChangeRuntimeState{ChangeRequestID: "change-id"}}
	binding := RepositoryBinding{
		Provider: "github", ProviderRef: "github-public", Role: RepositoryRoleSource,
		ProjectPath: "org/app", RepositoryURL: "https://github.com/org/app.git",
		DefaultBranch: "main", WorkflowEnabled: true,
	}
	service := providerAwareServiceForRuntimeStateTest(t, changeStore, runtimeStore, binding)

	if _, err := service.OpenMergeRequest(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	got := runtimeStore.state
	if got.Provider != "github" || got.ProviderRef != "github-public" || got.ProjectPath != "org/app" || got.Branch != "change/CHG-12" {
		t.Fatalf("reconstructed source state = %#v", got)
	}
}

func TestOpenMergeRequestDoesNotMarkStepWhenRuntimeStateReadFails(t *testing.T) {
	changeStore := &createBranchFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-13", ApplicationName: "github-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{getErr: errors.New("database unavailable")}
	binding := RepositoryBinding{
		Provider: "github", ProviderRef: "github-public", Role: RepositoryRoleSource,
		ProjectPath: "org/app", RepositoryURL: "https://github.com/org/app.git",
		DefaultBranch: "main", WorkflowEnabled: true,
	}
	service := providerAwareServiceForRuntimeStateTest(t, changeStore, runtimeStore, binding)

	_, err := service.OpenMergeRequest(context.Background(), changeStore.change.ChangeNumber)
	if err == nil || !strings.Contains(err.Error(), "persist source runtime state after opening merge request") {
		t.Fatalf("error = %v", err)
	}
	if changeStore.markStepCalled {
		t.Fatal("MarkStep was called after runtime-state read failure")
	}
}

func TestMergeRequestDoesNotMarkStepWhenRuntimeStateWriteFails(t *testing.T) {
	changeStore := &createBranchFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-14", ApplicationName: "gitlab-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{err: errors.New("database unavailable")}
	binding := RepositoryBinding{
		Provider: "gitlab", ProviderRef: "gitlab-lab", Role: RepositoryRoleSource,
		ProjectID: 42, ProjectPath: "group/app", RepositoryURL: "https://gitlab.example/group/app.git",
		DefaultBranch: "main", WorkflowEnabled: true,
	}
	service := providerAwareServiceForRuntimeStateTest(t, changeStore, runtimeStore, binding)

	_, err := service.MergeRequest(context.Background(), changeStore.change.ChangeNumber)
	if err == nil || !strings.Contains(err.Error(), "persist source runtime state after merging merge request") {
		t.Fatalf("error = %v", err)
	}
	if changeStore.markStepCalled {
		t.Fatal("MarkStep was called after runtime-state write failure")
	}
}
