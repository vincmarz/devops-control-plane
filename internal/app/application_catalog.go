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
	ProviderRef     string   `yaml:"providerRef" json:"providerRef"`
	Role            string   `yaml:"role" json:"role"`
	ProjectID       int      `yaml:"projectID,omitempty" json:"projectID,omitempty"`
	ProjectPath     string   `yaml:"projectPath" json:"projectPath"`
	RepositoryURL   string   `yaml:"repositoryURL" json:"repositoryURL"`
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
	catalog, err := NewApplicationCatalog([]ApplicationDefinition{{
		Name: "demo-go-color-app",
		Repositories: []RepositoryBinding{
			{Provider: RepositoryProviderGitLab, ProviderRef: "gitlab-lab", Role: RepositoryRoleSource, ProjectID: 1, ProjectPath: "devops-lab/demo-go-color-app-gitops", RepositoryURL: "https://gitlab-devops-gitlab.apps.ocp4.mim.lan/devops-lab/demo-go-color-app-gitops.git", DefaultBranch: "main", WorkflowEnabled: true},
			{Provider: RepositoryProviderGitHub, ProviderRef: "github-public", Role: RepositoryRoleGitOps, ProjectPath: "vincmarz/demo-app-gitops", RepositoryURL: "https://github.com/vincmarz/demo-app-gitops.git", DefaultBranch: "main", ConsumedBy: []string{"argocd", "tekton"}},
		},
	}})
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
		definition.Name = strings.TrimSpace(definition.Name)
		if definition.Name == "" {
			return ApplicationCatalog{}, fmt.Errorf("application at index %d does not define name", index)
		}
		if _, exists := catalog.applications[definition.Name]; exists {
			return ApplicationCatalog{}, fmt.Errorf("application %q is configured more than once", definition.Name)
		}
		if len(definition.Repositories) == 0 {
			return ApplicationCatalog{}, fmt.Errorf("application %q does not define repository bindings", definition.Name)
		}
		seen := map[string]bool{}
		for i := range definition.Repositories {
			b := &definition.Repositories[i]
			normalizeRepositoryBinding(b)
			if err := validateRepositoryBinding(definition.Name, *b); err != nil {
				return ApplicationCatalog{}, err
			}
			key := b.Provider + ":" + b.Role
			if seen[key] {
				return ApplicationCatalog{}, fmt.Errorf("application %q defines duplicate repository binding %q", definition.Name, key)
			}
			seen[key] = true
		}
		catalog.applications[definition.Name] = definition
	}
	return catalog, nil
}

func (c ApplicationCatalog) Resolve(name string) (ApplicationDefinition, bool) {
	definition, ok := c.applications[strings.TrimSpace(name)]
	return definition, ok
}

func (c ApplicationCatalog) ResolveSourceBinding(applicationName string) (RepositoryBinding, error) {
	return c.resolveSingleBindingByRole(applicationName, RepositoryRoleSource, true)
}
func (c ApplicationCatalog) ResolveGitOpsBinding(applicationName string) (RepositoryBinding, error) {
	return c.resolveSingleBindingByRole(applicationName, RepositoryRoleGitOps, false)
}
func (c ApplicationCatalog) resolveSingleBindingByRole(applicationName, role string, requireEnabled bool) (RepositoryBinding, error) {
	definition, ok := c.Resolve(applicationName)
	if !ok {
		return RepositoryBinding{}, fmt.Errorf("application %q is not configured", strings.TrimSpace(applicationName))
	}
	matches := []RepositoryBinding{}
	for _, b := range definition.Repositories {
		if b.Role == role && (!requireEnabled || b.WorkflowEnabled) {
			matches = append(matches, b)
		}
	}
	if len(matches) == 0 {
		if requireEnabled {
			return RepositoryBinding{}, fmt.Errorf("application %q does not define a workflow-enabled %s binding", definition.Name, role)
		}
		return RepositoryBinding{}, fmt.Errorf("application %q does not define a %s binding", definition.Name, role)
	}
	if len(matches) > 1 {
		return RepositoryBinding{}, fmt.Errorf("application %q defines more than one %s binding", definition.Name, role)
	}
	return matches[0], nil
}

func (c ApplicationCatalog) ResolveRepositoryBinding(applicationName, provider, role string) (RepositoryBinding, error) {
	definition, ok := c.Resolve(applicationName)
	if !ok {
		return RepositoryBinding{}, fmt.Errorf("application %q is not configured", strings.TrimSpace(applicationName))
	}
	provider, role = strings.ToLower(strings.TrimSpace(provider)), strings.ToLower(strings.TrimSpace(role))
	for _, b := range definition.Repositories {
		if b.Provider == provider && b.Role == role {
			return b, nil
		}
	}
	return RepositoryBinding{}, fmt.Errorf("application %q does not define repository binding %s:%s", definition.Name, provider, role)
}

func normalizeRepositoryBinding(b *RepositoryBinding) {
	b.Provider = strings.ToLower(strings.TrimSpace(b.Provider))
	b.ProviderRef = strings.TrimSpace(b.ProviderRef)
	b.Role = strings.ToLower(strings.TrimSpace(b.Role))
	b.ProjectPath = strings.TrimSpace(b.ProjectPath)
	b.RepositoryURL = strings.TrimSpace(b.RepositoryURL)
	b.DefaultBranch = strings.TrimSpace(b.DefaultBranch)
	for i := range b.ConsumedBy {
		b.ConsumedBy[i] = strings.ToLower(strings.TrimSpace(b.ConsumedBy[i]))
	}
}

func validateRepositoryBinding(applicationName string, b RepositoryBinding) error {
	if b.Provider != RepositoryProviderGitLab && b.Provider != RepositoryProviderGitHub {
		return fmt.Errorf("application %q repository binding provider %q is invalid", applicationName, b.Provider)
	}
	if b.ProviderRef == "" {
		return fmt.Errorf("application %q repository binding %s requires providerRef", applicationName, b.Provider)
	}
	if b.Role != RepositoryRoleSource && b.Role != RepositoryRoleGitOps {
		return fmt.Errorf("application %q repository binding role %q is invalid", applicationName, b.Role)
	}
	if b.ProjectPath == "" {
		return fmt.Errorf("application %q repository binding %s:%s requires projectPath", applicationName, b.Provider, b.Role)
	}
	if b.RepositoryURL == "" {
		return fmt.Errorf("application %q repository binding %s:%s requires repositoryURL", applicationName, b.Provider, b.Role)
	}
	if b.DefaultBranch == "" {
		return fmt.Errorf("application %q repository binding %s:%s does not define defaultBranch", applicationName, b.Provider, b.Role)
	}
	if b.Provider == RepositoryProviderGitLab && b.ProjectID <= 0 {
		return fmt.Errorf("application %q GitLab binding requires projectID", applicationName)
	}
	return nil
}
