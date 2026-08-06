package app

import (
	"context"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

// recordingGitLabGitOpsProvider is a fail-closed GitLab-style provider fake that
// captures the resolved GitRepositoryTarget it receives on every operation. It
// asserts, at contract level, that UpdateGitOps stays provider-neutral and that
// a GitLab gitops binding propagates its numeric ProjectID all the way to the
// provider, exactly as the real internal/adapters/gitlab/provider.go expects.
type recordingGitLabGitOpsProvider struct {
	provider        string
	providerRef     string
	mergeCommitSHA  string
	createBranchTgt GitRepositoryTarget
	fileTgt         GitRepositoryTarget
	openTgt         GitRepositoryTarget
	mergeTgt        GitRepositoryTarget
	createdFilePath string
	createdContent  string
}

func (p *recordingGitLabGitOpsProvider) Provider() string    { return p.provider }
func (p *recordingGitLabGitOpsProvider) ProviderRef() string { return p.providerRef }
func (p *recordingGitLabGitOpsProvider) CreateBranch(_ context.Context, target GitRepositoryTarget, _ string, _ string) error {
	p.createBranchTgt = target
	return nil
}
func (p *recordingGitLabGitOpsProvider) CreateOrUpdateFile(_ context.Context, target GitRepositoryTarget, branch, filePath, _, content string) (GitFileUpdateResult, error) {
	p.fileTgt = target
	p.createdFilePath = filePath
	p.createdContent = content
	return GitFileUpdateResult{FilePath: filePath, Branch: branch, CommitSHA: "gitlab-file-commit"}, nil
}
func (p *recordingGitLabGitOpsProvider) OpenMergeRequest(_ context.Context, target GitRepositoryTarget, _, _, _, _ string) (int, string, error) {
	p.openTgt = target
	return 9, "https://gitlab.example/mr/9", nil
}
func (p *recordingGitLabGitOpsProvider) MergeRequest(_ context.Context, target GitRepositoryTarget, _, _, _ string) (int, string, string, error) {
	p.mergeTgt = target
	return 9, "https://gitlab.example/mr/9", p.mergeCommitSHA, nil
}

func gitlabGitOpsBinding() RepositoryBinding {
	return RepositoryBinding{
		Provider: "gitlab", ProviderRef: "gitlab-lab", Role: RepositoryRoleGitOps,
		ProjectID: 42, ProjectPath: "devops-lab/demo-app-gitops",
		RepositoryURL: "https://gitlab.example/devops-lab/demo-app-gitops.git",
		DefaultBranch: "main",
	}
}

func gitlabUpdateGitOpsRuntimeState() domain.ChangeRuntimeState {
	return domain.ChangeRuntimeState{
		Artifact: domain.ArtifactRuntimeState{
			ImageRepository: "registry.example/team/demo-go-color-app",
			ImageDigest:     gitOpsUpdateDigest,
			Status:          "True",
			Reason:          "Succeeded",
		},
	}
}

func newGitLabUpdateGitOpsService(store *validateFakeStore, runtimeStore *sourceRuntimeStateStoreFake, provider *recordingGitLabGitOpsProvider) *ChangeService {
	registry, err := NewGitProviderRegistry([]GitProvider{provider})
	if err != nil {
		panic(err)
	}
	return NewChangeService(
		store,
		WithChangeRuntimeStateStore(runtimeStore),
		WithGitOpsBindingResolverFunc(func(string) (RepositoryBinding, error) { return gitlabGitOpsBinding(), nil }),
		WithGitProviderResolver(registry),
		WithGitOpsImageKustomizationPath("apps/demo-go-color-app/kustomization.yaml"),
	)
}

// TestUpdateGitOpsIsProviderNeutralForGitLab proves scenarios B and C at the
// contract level: with a GitLab gitops binding, UpdateGitOps selects the GitLab
// provider, propagates a target carrying provider=gitlab, providerRef=gitlab-lab
// and the numeric ProjectID on every operation, writes the deterministic
// Kustomize images override with the immutable digest, and persists gitops_state
// before MarkStep GitOpsUpdated.
func TestUpdateGitOpsIsProviderNeutralForGitLab(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0090", ApplicationName: "demo-go-color-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: gitlabUpdateGitOpsRuntimeState()}
	provider := &recordingGitLabGitOpsProvider{provider: "gitlab", providerRef: "gitlab-lab", mergeCommitSHA: "gitlab-merge-sha"}
	service := newGitLabUpdateGitOpsService(store, runtimeStore, provider)

	if _, err := service.UpdateGitOps(context.Background(), store.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}

	// ProjectID must reach the provider on every operation (GitLab uses it for API v4).
	for name, tgt := range map[string]GitRepositoryTarget{
		"createBranch": provider.createBranchTgt,
		"file":         provider.fileTgt,
		"openMR":       provider.openTgt,
		"mergeMR":      provider.mergeTgt,
	} {
		if tgt.Provider != "gitlab" || tgt.ProviderRef != "gitlab-lab" {
			t.Fatalf("%s target provider=%q ref=%q", name, tgt.Provider, tgt.ProviderRef)
		}
		if tgt.ProjectID != 42 {
			t.Fatalf("%s target ProjectID=%d, want 42", name, tgt.ProjectID)
		}
	}

	if provider.createdFilePath != "apps/demo-go-color-app/kustomization.yaml" {
		t.Fatalf("filePath=%q", provider.createdFilePath)
	}
	if !strings.Contains(provider.createdContent, "digest: "+gitOpsUpdateDigest) {
		t.Fatalf("content missing digest: %q", provider.createdContent)
	}
	if !strings.Contains(provider.createdContent, "newName: registry.example/team/demo-go-color-app") {
		t.Fatalf("content missing newName: %q", provider.createdContent)
	}

	if !runtimeStore.gitOpsCalled {
		t.Fatal("gitops state not persisted")
	}
	if runtimeStore.gitOps.Provider != "gitlab" || runtimeStore.gitOps.ProviderRef != "gitlab-lab" {
		t.Fatalf("gitops state provider=%#v", runtimeStore.gitOps)
	}
	if runtimeStore.gitOps.ProjectID != 42 {
		t.Fatalf("gitops state ProjectID=%d, want 42", runtimeStore.gitOps.ProjectID)
	}
	if runtimeStore.gitOps.CommitSHA != "gitlab-merge-sha" {
		t.Fatalf("gitops state commit=%q", runtimeStore.gitOps.CommitSHA)
	}
	if !store.markStepCalled || store.markedStatus != "GitOpsUpdated" {
		t.Fatalf("MarkStep=called:%v status:%q", store.markStepCalled, store.markedStatus)
	}
}
