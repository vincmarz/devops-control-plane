package app

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const applicationCatalogFileEnv = "APPLICATION_CATALOG_FILE"

const (
	RepositoryProviderGitLab = "gitlab"
	RepositoryProviderGitHub = "github"
	RepositoryRoleSource     = "source"
	RepositoryRoleGitOps     = "gitops"
)

type RepositoryBinding struct {
	Provider        string   `yaml:"provider" json:"provider"`
	Role            string   `yaml:"role" json:"role"`
	ProjectID       int      `yaml:"projectID,omitempty" json:"projectID,omitempty"`
	ProjectPath     string   `yaml:"projectPath,omitempty" json:"projectPath,omitempty"`
	RepositoryURL   string   `yaml:"repositoryURL,omitempty" json:"repositoryURL,omitempty"`
	DefaultBranch   string   `yaml:"defaultBranch" json:"defaultBranch"`
	WorkflowEnabled bool     `yaml:"workflowEnabled,omitempty" json:"workflowEnabled,omitempty"`
	ConsumedBy      []string `yaml:"consumedBy,omitempty" json:"consumedBy,omitempty"`
}

type ApplicationDefinition struct {
	Name         string              `yaml:"name" json:"name"`
	Repositories []RepositoryBinding `yaml:"repositories" json:"repositories"`
}

type ApplicationCatalog struct {
	applications map[string]ApplicationDefinition
}

type applicationCatalogFile struct {
	Applications []ApplicationDefinition `yaml:"applications"`
}

func DefaultApplicationCatalog() ApplicationCatalog {
	catalog, err := LoadApplicationCatalogFromFile(os.Getenv(applicationCatalogFileEnv))
	if err == nil {
		return catalog
	}
	return DefaultApplicationCatalogFallback()
}

func DefaultApplicationCatalogFallback() ApplicationCatalog {
	catalog, err := NewApplicationCatalog([]ApplicationDefinition{
		{
			Name: "demo-go-color-app",
			Repositories: []RepositoryBinding{
				{Provider: RepositoryProviderGitLab, Role: RepositoryRoleSource, ProjectID: 1, ProjectPath: "devops-lab/demo-go-color-app-gitops", DefaultBranch: "main", WorkflowEnabled: true},
				{Provider: RepositoryProviderGitHub, Role: RepositoryRoleGitOps, RepositoryURL: "https://github.com/vincmarz/demo-app-gitops.git", DefaultBranch: "main", ConsumedBy: []string{"argocd", "tekton"}},
			},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("invalid default application catalog: %v", err))
	}
	return catalog
}

func LoadApplicationCatalogFromFile(path string) (ApplicationCatalog, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return ApplicationCatalog{}, errors.New("application catalog file path is empty")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return ApplicationCatalog{}, fmt.Errorf("read application catalog file: %w", err)
	}
	return ParseApplicationCatalogYAML(content)
}

func ParseApplicationCatalogYAML(content []byte) (ApplicationCatalog, error) {
	var parsed applicationCatalogFile
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(&parsed); err != nil {
		return ApplicationCatalog{}, fmt.Errorf("parse application catalog yaml: %w", err)
	}
	if len(parsed.Applications) == 0 {
		return ApplicationCatalog{}, errors.New("application catalog does not define any applications")
	}
	return NewApplicationCatalog(parsed.Applications)
}

func NewApplicationCatalog(definitions []ApplicationDefinition) (ApplicationCatalog, error) {
	catalog := ApplicationCatalog{applications: map[string]ApplicationDefinition{}}
	for index, definition := range definitions {
		definition.Name = normalizeApplicationName(definition.Name)
		if definition.Name == "" {
			return ApplicationCatalog{}, fmt.Errorf("application at index %d does not define name", index)
		}
		if _, exists := catalog.applications[definition.Name]; exists {
			return ApplicationCatalog{}, fmt.Errorf("application %q is configured more than once", definition.Name)
		}
		if len(definition.Repositories) == 0 {
			return ApplicationCatalog{}, fmt.Errorf("application %q does not define repository bindings", definition.Name)
		}

		seenBindings := map[string]bool{}
		activeSourceBindings := 0
		for repositoryIndex := range definition.Repositories {
			binding := &definition.Repositories[repositoryIndex]
			normalizeRepositoryBinding(binding)
			if err := validateRepositoryBinding(definition.Name, *binding); err != nil {
				return ApplicationCatalog{}, err
			}
			key := binding.Provider + ":" + binding.Role
			if seenBindings[key] {
				return ApplicationCatalog{}, fmt.Errorf("application %q defines duplicate repository binding %q", definition.Name, key)
			}
			seenBindings[key] = true
			if binding.Role == RepositoryRoleSource && binding.WorkflowEnabled {
				activeSourceBindings++
			}
		}
		if activeSourceBindings > 1 {
			return ApplicationCatalog{}, fmt.Errorf("application %q defines more than one workflow-enabled source binding", definition.Name)
		}
		catalog.applications[definition.Name] = definition
	}
	return catalog, nil
}

func (c ApplicationCatalog) Resolve(name string) (ApplicationDefinition, bool) {
	definition, ok := c.applications[normalizeApplicationName(name)]
	return definition, ok
}

func (c ApplicationCatalog) ResolveRepositoryBinding(applicationName, provider, role string) (RepositoryBinding, error) {
	definition, ok := c.Resolve(applicationName)
	if !ok {
		return RepositoryBinding{}, fmt.Errorf("application %q is not configured", normalizeApplicationName(applicationName))
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	role = strings.ToLower(strings.TrimSpace(role))
	for _, binding := range definition.Repositories {
		if binding.Provider == provider && binding.Role == role {
			return binding, nil
		}
	}
	return RepositoryBinding{}, fmt.Errorf("application %q does not define repository binding %s:%s", definition.Name, provider, role)
}

func normalizeApplicationName(name string) string {
	return strings.TrimSpace(name)
}

func normalizeRepositoryBinding(binding *RepositoryBinding) {
	binding.Provider = strings.ToLower(strings.TrimSpace(binding.Provider))
	binding.Role = strings.ToLower(strings.TrimSpace(binding.Role))
	binding.ProjectPath = strings.TrimSpace(binding.ProjectPath)
	binding.RepositoryURL = strings.TrimSpace(binding.RepositoryURL)
	binding.DefaultBranch = strings.TrimSpace(binding.DefaultBranch)
	for index := range binding.ConsumedBy {
		binding.ConsumedBy[index] = strings.ToLower(strings.TrimSpace(binding.ConsumedBy[index]))
	}
}

func validateRepositoryBinding(applicationName string, binding RepositoryBinding) error {
	if binding.Provider == "" {
		return fmt.Errorf("application %q repository binding provider is required", applicationName)
	}
	if binding.Role != RepositoryRoleSource && binding.Role != RepositoryRoleGitOps {
		return fmt.Errorf("application %q repository binding role %q is invalid", applicationName, binding.Role)
	}
	if binding.DefaultBranch == "" {
		return fmt.Errorf("application %q repository binding %s:%s does not define defaultBranch", applicationName, binding.Provider, binding.Role)
	}
	if binding.Provider == RepositoryProviderGitLab && binding.Role == RepositoryRoleSource {
		if binding.ProjectID <= 0 {
			return fmt.Errorf("application %q GitLab source binding requires projectID", applicationName)
		}
		if binding.ProjectPath == "" {
			return fmt.Errorf("application %q GitLab source binding requires projectPath", applicationName)
		}
	}
	if binding.Role == RepositoryRoleGitOps && binding.RepositoryURL == "" {
		return fmt.Errorf("application %q GitOps binding requires repositoryURL", applicationName)
	}
	return nil
}
