package app

import (
	"strings"
	"testing"
)

func matrixBinding(provider, role, path, url string, projectID int, workflow bool) RepositoryBinding {
	return RepositoryBinding{Provider: provider, ProviderRef: provider + "-test", Role: role, ProjectID: projectID, ProjectPath: path, RepositoryURL: url, DefaultBranch: "main", WorkflowEnabled: workflow}
}

func TestApplicationCatalogProviderRoleMatrix(t *testing.T) {
	cases := []struct {
		name                   string
		source, gitops         RepositoryBinding
		wantSource, wantGitOps string
	}{
		{"A", matrixBinding("gitlab", "source", "group/a", "https://gitlab.example/group/a.git", 1, true), matrixBinding("github", "gitops", "org/a-gitops", "https://github.com/org/a-gitops.git", 0, false), "gitlab", "github"},
		{"B", matrixBinding("github", "source", "org/b", "https://github.com/org/b.git", 0, true), matrixBinding("gitlab", "gitops", "group/b-gitops", "https://gitlab.example/group/b-gitops.git", 2, false), "github", "gitlab"},
		{"C", matrixBinding("gitlab", "source", "group/c", "https://gitlab.example/group/c.git", 3, true), matrixBinding("gitlab", "gitops", "group/c-gitops", "https://gitlab.example/group/c-gitops.git", 4, false), "gitlab", "gitlab"},
		{"D", matrixBinding("github", "source", "org/d", "https://github.com/org/d.git", 0, true), matrixBinding("github", "gitops", "org/d-gitops", "https://github.com/org/d-gitops.git", 0, false), "github", "github"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			catalog, err := NewApplicationCatalog([]ApplicationDefinition{{Name: "app", Repositories: []RepositoryBinding{tc.source, tc.gitops}}})
			if err != nil {
				t.Fatal(err)
			}
			source, err := catalog.ResolveSourceBinding("app")
			if err != nil {
				t.Fatal(err)
			}
			gitops, err := catalog.ResolveGitOpsBinding("app")
			if err != nil {
				t.Fatal(err)
			}
			if source.Provider != tc.wantSource || gitops.Provider != tc.wantGitOps {
				t.Fatalf("source=%s gitops=%s", source.Provider, gitops.Provider)
			}
		})
	}
}

func TestParseApplicationCatalogYAMLProviderRefs(t *testing.T) {
	content := []byte(`applications:
  - name: demo
    repositories:
      - provider: gitlab
        providerRef: gitlab-lab
        role: source
        projectID: 1
        projectPath: group/demo
        repositoryURL: https://gitlab.example/group/demo.git
        defaultBranch: main
        workflowEnabled: true
      - provider: github
        providerRef: github-public
        role: gitops
        projectPath: org/demo-gitops
        repositoryURL: https://github.com/org/demo-gitops.git
        defaultBranch: main
`)
	catalog, err := ParseApplicationCatalogYAML(content)
	if err != nil {
		t.Fatal(err)
	}
	source, _ := catalog.ResolveSourceBinding("demo")
	gitops, _ := catalog.ResolveGitOpsBinding("demo")
	if source.ProviderRef != "gitlab-lab" || gitops.ProviderRef != "github-public" {
		t.Fatal("provider refs not preserved")
	}
}

func TestApplicationCatalogRejectsInvalidNeutralBindings(t *testing.T) {
	cases := []struct {
		name    string
		binding RepositoryBinding
		want    string
	}{
		{"unknown provider", RepositoryBinding{Provider: "bitbucket", ProviderRef: "bb", Role: "source", ProjectPath: "x/y", RepositoryURL: "https://x/y.git", DefaultBranch: "main"}, "is invalid"},
		{"missing provider ref", RepositoryBinding{Provider: "github", Role: "source", ProjectPath: "x/y", RepositoryURL: "https://github.com/x/y.git", DefaultBranch: "main"}, "requires providerRef"},
		{"missing path", RepositoryBinding{Provider: "github", ProviderRef: "gh", Role: "source", RepositoryURL: "https://github.com/x/y.git", DefaultBranch: "main"}, "requires projectPath"},
		{"missing URL", RepositoryBinding{Provider: "github", ProviderRef: "gh", Role: "source", ProjectPath: "x/y", DefaultBranch: "main"}, "requires repositoryURL"},
		{"GitLab missing ID", RepositoryBinding{Provider: "gitlab", ProviderRef: "gl", Role: "gitops", ProjectPath: "x/y", RepositoryURL: "https://gitlab.example/x/y.git", DefaultBranch: "main"}, "requires projectID"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewApplicationCatalog([]ApplicationDefinition{{Name: "app", Repositories: []RepositoryBinding{tc.binding}}})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v want=%s", err, tc.want)
			}
		})
	}
}

func TestApplicationCatalogRequiresEnabledSourceAndGitOpsBinding(t *testing.T) {
	catalog, err := NewApplicationCatalog([]ApplicationDefinition{{Name: "app", Repositories: []RepositoryBinding{matrixBinding("github", "source", "org/app", "https://github.com/org/app.git", 0, false)}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.ResolveSourceBinding("app"); err == nil {
		t.Fatal("expected disabled source error")
	}
	if _, err := catalog.ResolveGitOpsBinding("app"); err == nil {
		t.Fatal("expected missing GitOps error")
	}
}

func TestApplicationCatalogRejectsCredentialFields(t *testing.T) {
	content := []byte(`applications:
  - name: demo
    repositories:
      - provider: github
        providerRef: github-public
        role: source
        projectPath: org/demo
        repositoryURL: https://github.com/org/demo.git
        defaultBranch: main
        token: forbidden
`)
	_, err := ParseApplicationCatalogYAML(content)
	if err == nil || !strings.Contains(err.Error(), "field token not found") {
		t.Fatalf("error=%v", err)
	}
}
