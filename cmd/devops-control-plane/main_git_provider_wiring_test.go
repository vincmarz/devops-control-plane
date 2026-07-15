package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainWiresIndependentGitLabAndGitHubProviders(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	required := []string{
		`githubadapter "github.com/vincmarz/devops-control-plane/internal/adapters/github"`,
		`gitProviders := make([]app.GitProvider, 0, 2)`,
		`if cfg.GitHubToken != ""`,
		`githubadapter.NewProvider("github-public", gitHubClient)`,
		`gitlabadapter.NewProvider("gitlab-lab", gitLabClient)`,
		`app.NewGitProviderRegistry(gitProviders)`,
		`app.WithGitProviderResolver(gitProviderRegistry)`,
	}
	for _, value := range required {
		if !strings.Contains(text, value) {
			t.Fatalf("main.go does not contain %q", value)
		}
	}
}

func TestMainBuildsSingleGitProviderRegistry(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(content), "app.NewGitProviderRegistry(") != 1 {
		t.Fatal("main.go must build exactly one Git provider registry")
	}
}
