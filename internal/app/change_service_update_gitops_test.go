package app

import (
	"context"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

const gitOpsUpdateDigest = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

type recordingGitOpsProvider struct {
	provider        string
	providerRef     string
	createdBranch   string
	createdFilePath string
	createdContent  string
	openedBranch    string
	mergedBranch    string
	mergeCommitSHA  string
}

func (p *recordingGitOpsProvider) Provider() string    { return p.provider }
func (p *recordingGitOpsProvider) ProviderRef() string { return p.providerRef }
func (p *recordingGitOpsProvider) CreateBranch(_ context.Context, _ GitRepositoryTarget, branch, _ string) error {
	p.createdBranch = branch
	return nil
}
func (p *recordingGitOpsProvider) CreateOrUpdateFile(_ context.Context, _ GitRepositoryTarget, branch, filePath, _, content string) (GitFileUpdateResult, error) {
	p.createdFilePath = filePath
	p.createdContent = content
	return GitFileUpdateResult{FilePath: filePath, Branch: branch, CommitSHA: "file-commit"}, nil
}
func (p *recordingGitOpsProvider) OpenMergeRequest(_ context.Context, _ GitRepositoryTarget, sourceBranch, _, _, _ string) (int, string, error) {
	p.openedBranch = sourceBranch
	return 7, "https://example/pr/7", nil
}
func (p *recordingGitOpsProvider) MergeRequest(_ context.Context, _ GitRepositoryTarget, sourceBranch, _, _ string) (int, string, string, error) {
	p.mergedBranch = sourceBranch
	return 7, "https://example/pr/7", p.mergeCommitSHA, nil
}

func gitOpsBinding() RepositoryBinding {
	return RepositoryBinding{
		Provider: "github", ProviderRef: "github-public", Role: RepositoryRoleGitOps,
		ProjectPath: "vincmarz/demo-app-gitops", RepositoryURL: "https://github.com/vincmarz/demo-app-gitops.git",
		DefaultBranch: "main",
	}
}

func newUpdateGitOpsService(store *validateFakeStore, runtimeStore *sourceRuntimeStateStoreFake, provider *recordingGitOpsProvider) *ChangeService {
	registry, err := NewGitProviderRegistry([]GitProvider{provider})
	if err != nil {
		panic(err)
	}
	return NewChangeService(
		store,
		WithChangeRuntimeStateStore(runtimeStore),
		WithGitOpsBindingResolverFunc(func(string) (RepositoryBinding, error) { return gitOpsBinding(), nil }),
		WithGitProviderResolver(registry),
		WithGitOpsImageKustomizationPath("apps/demo-go-color-app/kustomization.yaml"),
	)
}

func updateGitOpsRuntimeState() domain.ChangeRuntimeState {
	return domain.ChangeRuntimeState{
		Artifact: domain.ArtifactRuntimeState{
			ImageRepository: "image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app",
			ImageDigest:     gitOpsUpdateDigest,
			Status:          "True",
			Reason:          "Succeeded",
		},
	}
}

func TestUpdateGitOpsWritesDigestAndPersistsBeforeMarkStep(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0072", ApplicationName: "demo-go-color-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: updateGitOpsRuntimeState()}
	provider := &recordingGitOpsProvider{provider: "github", providerRef: "github-public", mergeCommitSHA: "merge-sha"}
	service := newUpdateGitOpsService(store, runtimeStore, provider)

	if _, err := service.UpdateGitOps(context.Background(), store.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	if provider.createdBranch != "gitops/CHG-2026-0072" || provider.mergedBranch != "gitops/CHG-2026-0072" {
		t.Fatalf("branches created=%q merged=%q", provider.createdBranch, provider.mergedBranch)
	}
	if provider.createdFilePath != "apps/demo-go-color-app/kustomization.yaml" {
		t.Fatalf("filePath=%q", provider.createdFilePath)
	}
	if !strings.Contains(provider.createdContent, "digest: "+gitOpsUpdateDigest) {
		t.Fatalf("content missing digest: %q", provider.createdContent)
	}
	if !strings.Contains(provider.createdContent, "newName: image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app") {
		t.Fatalf("content missing newName: %q", provider.createdContent)
	}
	if !runtimeStore.gitOpsCalled {
		t.Fatal("gitops state not persisted")
	}
	if runtimeStore.gitOps.CommitSHA != "merge-sha" || runtimeStore.gitOps.Provider != "github" {
		t.Fatalf("gitops state=%#v", runtimeStore.gitOps)
	}
	if !store.markStepCalled || store.markedStatus != "GitOpsUpdated" {
		t.Fatalf("MarkStep=called:%v status:%q", store.markStepCalled, store.markedStatus)
	}
}

func TestUpdateGitOpsFailsClosed(t *testing.T) {
	cases := []struct {
		name  string
		state domain.ChangeRuntimeState
		want  string
	}{
		{name: "missing repository", state: domain.ChangeRuntimeState{Artifact: domain.ArtifactRuntimeState{ImageDigest: gitOpsUpdateDigest}}, want: "image repository"},
		{name: "invalid digest", state: domain.ChangeRuntimeState{Artifact: domain.ArtifactRuntimeState{ImageRepository: "repo", ImageDigest: "not-a-digest"}}, want: "valid image digest"},
		{name: "empty digest", state: domain.ChangeRuntimeState{Artifact: domain.ArtifactRuntimeState{ImageRepository: "repo"}}, want: "valid image digest"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0073", ApplicationName: "demo-go-color-app"}}
			runtimeStore := &sourceRuntimeStateStoreFake{current: tc.state}
			provider := &recordingGitOpsProvider{provider: "github", providerRef: "github-public"}
			service := newUpdateGitOpsService(store, runtimeStore, provider)
			_, err := service.UpdateGitOps(context.Background(), store.change.ChangeNumber)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v want=%q", err, tc.want)
			}
			if provider.createdBranch != "" || provider.openedBranch != "" || store.markStepCalled {
				t.Fatalf("side effects: branch=%q opened=%q mark=%v", provider.createdBranch, provider.openedBranch, store.markStepCalled)
			}
		})
	}
}

func TestUpdateGitOpsRequiresRuntimeStore(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0074", ApplicationName: "demo-go-color-app"}}
	service := NewChangeService(store, WithGitOpsImageKustomizationPath("apps/demo-go-color-app/kustomization.yaml"))
	if _, err := service.UpdateGitOps(context.Background(), store.change.ChangeNumber); err == nil || !strings.Contains(err.Error(), "runtime state store") {
		t.Fatalf("err=%v", err)
	}
}

func TestGeneratedGitOpsKustomizationIsDeterministic(t *testing.T) {
	out := generatedGitOpsKustomization("repo/app", gitOpsUpdateDigest)
	for _, want := range []string{"kind: Kustomization", "resources:", "- deployment.yaml", "images:", "name: repo/app", "newName: repo/app", "digest: " + gitOpsUpdateDigest} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}
