package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type sourceRuntimeStateStoreFake struct {
	called     bool
	idOrNumber string
	state      domain.SourceRuntimeState
	err        error
}

func (f *sourceRuntimeStateStoreFake) UpsertSource(_ context.Context, idOrNumber string, state domain.SourceRuntimeState) error {
	f.called = true
	f.idOrNumber = idOrNumber
	f.state = state
	return f.err
}

func providerAwareServiceForRuntimeStateTest(t *testing.T, store ChangeStore, runtimeStore ChangeRuntimeStateStore, binding RepositoryBinding) *ChangeService {
	t.Helper()
	provider := &providerAwareFake{provider: binding.Provider, providerRef: binding.ProviderRef}
	registry, err := NewGitProviderRegistry([]GitProvider{provider})
	if err != nil {
		t.Fatal(err)
	}
	return NewChangeService(
		store,
		WithGitSourceBindingResolverFunc(func(string) (RepositoryBinding, error) { return binding, nil }),
		WithGitProviderResolver(registry),
		WithChangeRuntimeStateStore(runtimeStore),
	)
}

func TestCreateBranchPersistsProviderAwareGitHubSourceState(t *testing.T) {
	changeStore := &createBranchFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-1", ApplicationName: "github-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{}
	binding := RepositoryBinding{
		Provider: "github", ProviderRef: "github-public", Role: RepositoryRoleSource,
		ProjectPath: "org/app", RepositoryURL: "https://github.com/org/app.git",
		DefaultBranch: "main", WorkflowEnabled: true,
	}
	service := providerAwareServiceForRuntimeStateTest(t, changeStore, runtimeStore, binding)

	if _, err := service.CreateBranch(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	if !runtimeStore.called {
		t.Fatal("runtime state store was not called")
	}
	if runtimeStore.idOrNumber != changeStore.change.ID {
		t.Fatalf("runtime state change ID = %q, want %q", runtimeStore.idOrNumber, changeStore.change.ID)
	}
	got := runtimeStore.state
	if got.Provider != "github" || got.ProviderRef != "github-public" || got.ProjectID != 0 {
		t.Fatalf("provider metadata = %#v", got)
	}
	if got.ProjectPath != "org/app" || got.RepositoryURL != "https://github.com/org/app.git" {
		t.Fatalf("repository metadata = %#v", got)
	}
	if got.DefaultBranch != "main" || got.Branch != "change/CHG-1" || got.CommitSHA != "" {
		t.Fatalf("branch metadata = %#v", got)
	}
	if !changeStore.markStepCalled || changeStore.markedStatus != "BranchCreated" {
		t.Fatalf("MarkStep state = called:%v status:%q", changeStore.markStepCalled, changeStore.markedStatus)
	}
}

func TestUpdateFilesPersistsProviderAwareGitLabSourceState(t *testing.T) {
	changeStore := &createBranchFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2", ApplicationName: "gitlab-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{}
	binding := RepositoryBinding{
		Provider: "gitlab", ProviderRef: "gitlab-lab", Role: RepositoryRoleSource,
		ProjectID: 42, ProjectPath: "group/app", RepositoryURL: "https://gitlab.example/group/app.git",
		DefaultBranch: "trunk", WorkflowEnabled: true,
	}
	service := providerAwareServiceForRuntimeStateTest(t, changeStore, runtimeStore, binding)

	if _, err := service.UpdateFiles(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	got := runtimeStore.state
	if got.Provider != "gitlab" || got.ProviderRef != "gitlab-lab" || got.ProjectID != 42 {
		t.Fatalf("provider metadata = %#v", got)
	}
	if got.DefaultBranch != "trunk" || got.Branch != "change/CHG-2" {
		t.Fatalf("branch metadata = %#v", got)
	}
	if !changeStore.markStepCalled || changeStore.markedStatus != "CommitCreated" {
		t.Fatalf("MarkStep state = called:%v status:%q", changeStore.markStepCalled, changeStore.markedStatus)
	}
}

func TestCreateBranchDoesNotMarkStepWhenRuntimeStatePersistenceFails(t *testing.T) {
	changeStore := &createBranchFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-3", ApplicationName: "github-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{err: errors.New("database unavailable")}
	binding := RepositoryBinding{
		Provider: "github", ProviderRef: "github-public", Role: RepositoryRoleSource,
		ProjectPath: "org/app", RepositoryURL: "https://github.com/org/app.git",
		DefaultBranch: "main", WorkflowEnabled: true,
	}
	service := providerAwareServiceForRuntimeStateTest(t, changeStore, runtimeStore, binding)

	_, err := service.CreateBranch(context.Background(), changeStore.change.ChangeNumber)
	if err == nil || !strings.Contains(err.Error(), "persist source runtime state after creating branch") {
		t.Fatalf("error = %v", err)
	}
	if changeStore.markStepCalled {
		t.Fatal("MarkStep was called after runtime state persistence failure")
	}
}

func TestUpdateFilesDoesNotMarkStepWhenRuntimeStatePersistenceFails(t *testing.T) {
	changeStore := &createBranchFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-4", ApplicationName: "gitlab-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{err: errors.New("database unavailable")}
	binding := RepositoryBinding{
		Provider: "gitlab", ProviderRef: "gitlab-lab", Role: RepositoryRoleSource,
		ProjectID: 42, ProjectPath: "group/app", RepositoryURL: "https://gitlab.example/group/app.git",
		DefaultBranch: "main", WorkflowEnabled: true,
	}
	service := providerAwareServiceForRuntimeStateTest(t, changeStore, runtimeStore, binding)

	_, err := service.UpdateFiles(context.Background(), changeStore.change.ChangeNumber)
	if err == nil || !strings.Contains(err.Error(), "persist source runtime state after updating files") {
		t.Fatalf("error = %v", err)
	}
	if changeStore.markStepCalled {
		t.Fatal("MarkStep was called after runtime state persistence failure")
	}
}

func TestProviderAwareSourceActionsRemainCompatibleWithoutRuntimeStateStore(t *testing.T) {
	changeStore := &createBranchFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-5", ApplicationName: "github-app"}}
	binding := RepositoryBinding{
		Provider: "github", ProviderRef: "github-public", Role: RepositoryRoleSource,
		ProjectPath: "org/app", RepositoryURL: "https://github.com/org/app.git",
		DefaultBranch: "main", WorkflowEnabled: true,
	}
	service := providerAwareServiceForRuntimeStateTest(t, changeStore, nil, binding)

	if _, err := service.CreateBranch(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	if !changeStore.markStepCalled {
		t.Fatal("MarkStep was not called when runtime state store was omitted")
	}
}
