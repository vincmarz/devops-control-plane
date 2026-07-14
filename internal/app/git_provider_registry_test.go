package app

import (
	"context"
	"strings"
	"testing"
)

type fakeGitProvider struct {
	provider    string
	providerRef string
}

func (f fakeGitProvider) Provider() string    { return f.provider }
func (f fakeGitProvider) ProviderRef() string { return f.providerRef }
func (f fakeGitProvider) CreateBranch(context.Context, GitRepositoryTarget, string, string) error {
	return nil
}
func (f fakeGitProvider) CreateOrUpdateFile(context.Context, GitRepositoryTarget, string, string, string, string) error {
	return nil
}
func (f fakeGitProvider) OpenMergeRequest(context.Context, GitRepositoryTarget, string, string, string, string) (int, string, error) {
	return 1, "https://example.test/proposal/1", nil
}
func (f fakeGitProvider) MergeRequest(context.Context, GitRepositoryTarget, string, string, string) (int, string, string, error) {
	return 1, "https://example.test/proposal/1", "abc123", nil
}

func TestNewGitRepositoryTargetFromProviderNeutralBindings(t *testing.T) {
	cases := []RepositoryBinding{
		{Provider: "gitlab", ProviderRef: "gitlab-lab", Role: "source", ProjectID: 1, ProjectPath: "group/app", RepositoryURL: "https://gitlab.example/group/app.git", DefaultBranch: "main", WorkflowEnabled: true},
		{Provider: "github", ProviderRef: "github-public", Role: "source", ProjectPath: "org/app", RepositoryURL: "https://github.com/org/app.git", DefaultBranch: "main", WorkflowEnabled: true},
	}
	for _, binding := range cases {
		target, err := NewGitRepositoryTarget(binding)
		if err != nil {
			t.Fatalf("NewGitRepositoryTarget(%s): %v", binding.Provider, err)
		}
		if target.Provider != binding.Provider || target.ProviderRef != binding.ProviderRef || target.ProjectPath != binding.ProjectPath {
			t.Fatalf("unexpected target: %#v", target)
		}
	}
}

func TestGitProviderRegistryResolvesByProviderRef(t *testing.T) {
	registry, err := NewGitProviderRegistry([]GitProvider{
		fakeGitProvider{provider: "gitlab", providerRef: "gitlab-lab"},
		fakeGitProvider{provider: "github", providerRef: "github-public"},
	})
	if err != nil {
		t.Fatal(err)
	}
	cases := []GitRepositoryTarget{
		{Provider: "gitlab", ProviderRef: "gitlab-lab"},
		{Provider: "github", ProviderRef: "github-public"},
	}
	for _, target := range cases {
		provider, err := registry.Resolve(target)
		if err != nil {
			t.Fatalf("Resolve(%s): %v", target.ProviderRef, err)
		}
		if provider.Provider() != target.Provider {
			t.Fatalf("provider=%s want=%s", provider.Provider(), target.Provider)
		}
	}
}

func TestGitProviderRegistryFailsClosed(t *testing.T) {
	registry, err := NewGitProviderRegistry([]GitProvider{fakeGitProvider{provider: "gitlab", providerRef: "gitlab-lab"}})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		target GitRepositoryTarget
		want   string
	}{
		{"missing provider", GitRepositoryTarget{ProviderRef: "gitlab-lab"}, "provider is required"},
		{"missing providerRef", GitRepositoryTarget{Provider: "gitlab"}, "providerRef is required"},
		{"unregistered providerRef", GitRepositoryTarget{Provider: "github", ProviderRef: "github-public"}, "is not registered"},
		{"provider mismatch", GitRepositoryTarget{Provider: "github", ProviderRef: "gitlab-lab"}, `registered as "gitlab", not "github"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := registry.Resolve(tc.target)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v want=%q", err, tc.want)
			}
		})
	}
}

func TestGitProviderRegistryRejectsInvalidRegistrations(t *testing.T) {
	cases := []struct {
		name      string
		providers []GitProvider
		want      string
	}{
		{"nil provider", []GitProvider{nil}, "is nil"},
		{"missing provider", []GitProvider{fakeGitProvider{providerRef: "x"}}, "does not define provider"},
		{"missing providerRef", []GitProvider{fakeGitProvider{provider: "gitlab"}}, "does not define providerRef"},
		{"duplicate providerRef", []GitProvider{fakeGitProvider{provider: "gitlab", providerRef: "shared"}, fakeGitProvider{provider: "github", providerRef: "shared"}}, "registered more than once"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewGitProviderRegistry(tc.providers)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v want=%q", err, tc.want)
			}
		})
	}
}
