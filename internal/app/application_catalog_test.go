package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseApplicationCatalogYAML(t *testing.T) {
	content := []byte(`applications:
  - name: demo-go-color-app
    repositories:
      - provider: gitlab
        role: source
        projectID: 1
        projectPath: devops-lab/demo-go-color-app-gitops
        defaultBranch: main
        workflowEnabled: true
      - provider: github
        role: gitops
        repositoryURL: https://github.com/vincmarz/demo-app-gitops.git
        defaultBranch: main
        consumedBy:
          - argocd
          - tekton
`)
	catalog, err := ParseApplicationCatalogYAML(content)
	if err != nil {
		t.Fatalf("ParseApplicationCatalogYAML returned error %v", err)
	}
	source, err := catalog.ResolveRepositoryBinding("demo-go-color-app", "gitlab", "source")
	if err != nil {
		t.Fatalf("resolve GitLab source binding: %v", err)
	}
	if source.ProjectID != 1 || source.ProjectPath != "devops-lab/demo-go-color-app-gitops" || !source.WorkflowEnabled {
		t.Fatalf("unexpected source binding: %#v", source)
	}
	gitops, err := catalog.ResolveRepositoryBinding("demo-go-color-app", "github", "gitops")
	if err != nil {
		t.Fatalf("resolve GitHub GitOps binding: %v", err)
	}
	if gitops.RepositoryURL != "https://github.com/vincmarz/demo-app-gitops.git" || strings.Join(gitops.ConsumedBy, ",") != "argocd,tekton" {
		t.Fatalf("unexpected GitOps binding: %#v", gitops)
	}
}

func TestApplicationCatalogRejectsDuplicateApplication(t *testing.T) {
	content := []byte(`applications:
  - name: demo
    repositories:
      - provider: github
        role: gitops
        repositoryURL: https://example.test/one.git
        defaultBranch: main
  - name: demo
    repositories:
      - provider: github
        role: gitops
        repositoryURL: https://example.test/two.git
        defaultBranch: main
`)
	_, err := ParseApplicationCatalogYAML(content)
	if err == nil || !strings.Contains(err.Error(), `application "demo" is configured more than once`) {
		t.Fatalf("duplicate application error = %v", err)
	}
}

func TestApplicationCatalogRejectsAmbiguousSourceBindings(t *testing.T) {
	_, err := NewApplicationCatalog([]ApplicationDefinition{{
		Name: "demo",
		Repositories: []RepositoryBinding{
			{Provider: "gitlab", Role: "source", ProjectID: 1, ProjectPath: "group/one", DefaultBranch: "main", WorkflowEnabled: true},
			{Provider: "github", Role: "source", RepositoryURL: "https://example.test/two.git", DefaultBranch: "main", WorkflowEnabled: true},
		},
	}})
	if err == nil || !strings.Contains(err.Error(), "more than one workflow-enabled source binding") {
		t.Fatalf("ambiguous source error = %v", err)
	}
}

func TestApplicationCatalogRejectsIncompleteBindings(t *testing.T) {
	tests := []struct {
		name    string
		binding RepositoryBinding
		want    string
	}{
		{name: "missing provider", binding: RepositoryBinding{Role: "source", DefaultBranch: "main"}, want: "provider is required"},
		{name: "invalid role", binding: RepositoryBinding{Provider: "gitlab", Role: "mirror", DefaultBranch: "main"}, want: `role "mirror" is invalid`},
		{name: "missing GitLab project ID", binding: RepositoryBinding{Provider: "gitlab", Role: "source", ProjectPath: "group/project", DefaultBranch: "main"}, want: "requires projectID"},
		{name: "missing GitLab project path", binding: RepositoryBinding{Provider: "gitlab", Role: "source", ProjectID: 1, DefaultBranch: "main"}, want: "requires projectPath"},
		{name: "missing GitOps URL", binding: RepositoryBinding{Provider: "github", Role: "gitops", DefaultBranch: "main"}, want: "requires repositoryURL"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewApplicationCatalog([]ApplicationDefinition{{Name: "demo", Repositories: []RepositoryBinding{test.binding}}})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestApplicationCatalogRejectsCredentialFields(t *testing.T) {
	content := []byte(`applications:
  - name: demo
    repositories:
      - provider: gitlab
        role: source
        projectID: 1
        projectPath: group/project
        defaultBranch: main
        token: forbidden
`)
	_, err := ParseApplicationCatalogYAML(content)
	if err == nil || !strings.Contains(err.Error(), "field token not found") {
		t.Fatalf("credential field error = %v", err)
	}
}

func TestDefaultApplicationCatalogLoadsFileAndFallsBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "applications.yaml")
	content := []byte(`applications:
  - name: payments
    repositories:
      - provider: github
        role: gitops
        repositoryURL: https://example.test/payments.git
        defaultBranch: main
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write application catalog: %v", err)
	}
	t.Setenv(applicationCatalogFileEnv, path)
	if _, ok := DefaultApplicationCatalog().Resolve("payments"); !ok {
		t.Fatal("file-backed application catalog did not load payments")
	}
	t.Setenv(applicationCatalogFileEnv, filepath.Join(t.TempDir(), "missing.yaml"))
	if _, ok := DefaultApplicationCatalog().Resolve("demo-go-color-app"); !ok {
		t.Fatal("fallback application catalog did not load demo-go-color-app")
	}
}
