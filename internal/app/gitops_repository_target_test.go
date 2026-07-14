package app

import (
	"errors"
	"strings"
	"testing"
)

func validGitOpsBinding(provider string) RepositoryBinding {
	projectID := 0
	projectPath := "org/demo-gitops"
	repositoryURL := "https://github.com/org/demo-gitops.git"
	providerRef := "github-public"
	if provider == RepositoryProviderGitLab {
		projectID = 11
		projectPath = "group/demo-gitops"
		repositoryURL = "https://gitlab.example/group/demo-gitops.git"
		providerRef = "gitlab-lab"
	}
	return RepositoryBinding{Provider: provider, ProviderRef: providerRef, Role: RepositoryRoleGitOps, ProjectID: projectID, ProjectPath: projectPath, RepositoryURL: repositoryURL, DefaultBranch: "main", ConsumedBy: []string{"argocd", "tekton"}}
}

func TestGitOpsRepositoryTargetProviderMatrix(t *testing.T) {
	for _, provider := range []string{RepositoryProviderGitLab, RepositoryProviderGitHub} {
		t.Run(provider, func(t *testing.T) {
			target, err := NewGitOpsRepositoryTarget(validGitOpsBinding(provider), GitOpsConsumerTekton)
			if err != nil {
				t.Fatal(err)
			}
			if target.Provider != provider {
				t.Fatalf("provider = %q", target.Provider)
			}
			if !target.SupportsConsumer(GitOpsConsumerArgoCD) || !target.SupportsConsumer(GitOpsConsumerTekton) {
				t.Fatalf("consumers = %#v", target.ConsumedBy)
			}
		})
	}
}

func TestGitOpsRepositoryTargetResolverUsesApplicationBinding(t *testing.T) {
	resolver := NewGitOpsRepositoryTargetResolver(func(applicationName string) (RepositoryBinding, error) {
		if applicationName != "demo-go-color-app" {
			return RepositoryBinding{}, errors.New("application is not configured")
		}
		return validGitOpsBinding(RepositoryProviderGitHub), nil
	})
	target, err := resolver.Resolve(" demo-go-color-app ", " TEKTON ")
	if err != nil {
		t.Fatal(err)
	}
	if target.ProviderRef != "github-public" || target.DefaultBranch != "main" {
		t.Fatalf("target = %#v", target)
	}
}

func TestGitOpsRepositoryTargetRejectsInvalidBindings(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*RepositoryBinding)
		consumer string
		want     string
	}{
		{"wrong role", func(b *RepositoryBinding) { b.Role = RepositoryRoleSource }, GitOpsConsumerTekton, "is not GitOps"},
		{"missing provider", func(b *RepositoryBinding) { b.Provider = "" }, GitOpsConsumerTekton, "provider is required"},
		{"missing providerRef", func(b *RepositoryBinding) { b.ProviderRef = "" }, GitOpsConsumerTekton, "providerRef is required"},
		{"missing projectPath", func(b *RepositoryBinding) { b.ProjectPath = "" }, GitOpsConsumerTekton, "projectPath is required"},
		{"missing repositoryURL", func(b *RepositoryBinding) { b.RepositoryURL = "" }, GitOpsConsumerTekton, "repositoryURL is required"},
		{"missing defaultBranch", func(b *RepositoryBinding) { b.DefaultBranch = "" }, GitOpsConsumerTekton, "defaultBranch is required"},
		{"consumer not configured", func(b *RepositoryBinding) { b.ConsumedBy = []string{"argocd"} }, GitOpsConsumerTekton, "not configured for consumer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binding := validGitOpsBinding(RepositoryProviderGitHub)
			tt.mutate(&binding)
			_, err := NewGitOpsRepositoryTarget(binding, tt.consumer)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestGitOpsRepositoryTargetResolverFailsClosed(t *testing.T) {
	resolver := NewGitOpsRepositoryTargetResolver(nil)
	_, err := resolver.Resolve("app", GitOpsConsumerTekton)
	if err == nil || !strings.Contains(err.Error(), "resolver is not configured") {
		t.Fatalf("error = %v", err)
	}
}

func TestGitOpsRepositoryURLCorrelation(t *testing.T) {
	target, err := NewGitOpsRepositoryTarget(validGitOpsBinding(RepositoryProviderGitHub), GitOpsConsumerArgoCD)
	if err != nil {
		t.Fatal(err)
	}
	for _, repositoryURL := range []string{"https://github.com/org/demo-gitops.git", " HTTPS://GITHUB.COM/ORG/DEMO-GITOPS.GIT ", "https://github.com/org/demo-gitops/"} {
		if !target.MatchesRepositoryURL(repositoryURL) {
			t.Fatalf("repository URL %q should match", repositoryURL)
		}
	}
	if target.MatchesRepositoryURL("https://gitlab.example/org/demo-gitops.git") {
		t.Fatal("different provider URL must not match")
	}
}
