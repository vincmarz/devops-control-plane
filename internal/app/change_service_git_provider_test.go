package app

import (
	"context"
	"errors"
	"github.com/vincmarz/devops-control-plane/internal/domain"
	"strings"
	"testing"
)

type providerAwareFake struct {
	provider    string
	providerRef string
	target      GitRepositoryTarget
	operation   string
}

func (p *providerAwareFake) Provider() string    { return p.provider }
func (p *providerAwareFake) ProviderRef() string { return p.providerRef }
func (p *providerAwareFake) CreateBranch(context.Context, GitRepositoryTarget, string, string) error {
	p.operation = "create"
	return nil
}
func (p *providerAwareFake) CreateOrUpdateFile(_ context.Context, target GitRepositoryTarget, branch, filePath, message, content string) error {
	p.target = target
	p.operation = "update"
	return nil
}
func (p *providerAwareFake) OpenMergeRequest(context.Context, GitRepositoryTarget, string, string, string, string) (int, string, error) {
	return 7, "https://example.test/mr/7", nil
}
func (p *providerAwareFake) MergeRequest(context.Context, GitRepositoryTarget, string, string, string) (int, string, string, error) {
	return 7, "https://example.test/mr/7", "abc123", nil
}

func TestChangeServiceUsesProviderAwareSourceBinding(t *testing.T) {
	store := &createBranchFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-1", ApplicationName: "demo-go-color-app", TargetEnvironment: "dev"}}
	provider := &providerAwareFake{provider: "gitlab", providerRef: "gitlab-lab"}
	registry, err := NewGitProviderRegistry([]GitProvider{provider})
	if err != nil {
		t.Fatal(err)
	}
	binding := RepositoryBinding{Provider: "gitlab", ProviderRef: "gitlab-lab", Role: "source", ProjectID: 42, ProjectPath: "group/app", RepositoryURL: "https://gitlab.example/group/app.git", DefaultBranch: "trunk", WorkflowEnabled: true}
	service := NewChangeService(store, WithGitSourceBindingResolverFunc(func(string) (RepositoryBinding, error) { return binding, nil }), WithGitProviderResolver(registry))
	result, err := service.UpdateFiles(context.Background(), "CHG-1")
	if err != nil {
		t.Fatal(err)
	}
	metadata := result["git"].(map[string]any)
	if metadata["provider"] != "gitlab" || metadata["providerRef"] != "gitlab-lab" || metadata["projectID"] != 42 || metadata["defaultBranch"] != "trunk" {
		t.Fatalf("metadata=%#v", metadata)
	}
}

func TestChangeServiceFailsClosedForUnregisteredGitHubSource(t *testing.T) {
	store := &createBranchFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-1", ApplicationName: "demo-go-color-app"}}
	registry, err := NewGitProviderRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	binding := RepositoryBinding{Provider: "github", ProviderRef: "github-public", Role: "source", ProjectPath: "org/app", RepositoryURL: "https://github.com/org/app.git", DefaultBranch: "main", WorkflowEnabled: true}
	service := NewChangeService(store, WithGitSourceBindingResolverFunc(func(string) (RepositoryBinding, error) { return binding, nil }), WithGitProviderResolver(registry))
	_, err = service.CreateBranch(context.Background(), "CHG-1")
	if err == nil || !strings.Contains(err.Error(), `git providerRef "github-public" is not registered`) {
		t.Fatalf("error=%v", err)
	}
}

func TestChangeServiceFailsClosedForMissingApplicationBinding(t *testing.T) {
	store := &createBranchFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-1", ApplicationName: "unknown"}}
	registry, _ := NewGitProviderRegistry(nil)
	service := NewChangeService(store, WithGitSourceBindingResolverFunc(func(string) (RepositoryBinding, error) {
		return RepositoryBinding{}, errors.New("application is not configured")
	}), WithGitProviderResolver(registry))
	_, err := service.CreateBranch(context.Background(), "CHG-1")
	if err == nil || !strings.Contains(err.Error(), "application is not configured") {
		t.Fatalf("error=%v", err)
	}
}
