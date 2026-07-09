package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const environmentCatalogFileEnv = "ENVIRONMENT_CATALOG_FILE"

type EnvironmentDefinition struct {
	Name                  string `yaml:"name"`
	DisplayName           string `yaml:"displayName"`
	Enabled               bool   `yaml:"enabled"`
	Category              string `yaml:"category"`
	Description           string `yaml:"description"`
	ClusterName           string `yaml:"clusterName"`
	KubernetesNamespace   string `yaml:"kubernetesNamespace"`
	TektonNamespace       string `yaml:"tektonNamespace"`
	ArgoCDApplicationName string `yaml:"argocdApplicationName"`
	GitTargetBranch       string `yaml:"gitTargetBranch"`
	ValidationPath        string `yaml:"validationPath"`
	AllowTechnicalActions bool   `yaml:"allowTechnicalActions"`
	AllowPromotionSource  bool   `yaml:"allowPromotionSource"`
	AllowPromotionTarget  bool   `yaml:"allowPromotionTarget"`
}

type EnvironmentCatalog struct {
	defaultEnvironment string
	environments       map[string]EnvironmentDefinition
}

type environmentCatalogFile struct {
	DefaultEnvironment string                  `yaml:"defaultEnvironment"`
	Environments       []EnvironmentDefinition `yaml:"environments"`
}

func DefaultEnvironmentCatalog() EnvironmentCatalog {
	catalog, err := LoadEnvironmentCatalogFromFile(os.Getenv(environmentCatalogFileEnv))
	if err == nil {
		return catalog
	}

	return DefaultEnvironmentCatalogFallback()
}

func DefaultEnvironmentCatalogFallback() EnvironmentCatalog {
	return NewEnvironmentCatalog([]EnvironmentDefinition{
		{
			Name:                  "dev",
			DisplayName:           "Development",
			Enabled:               true,
			Category:              "development",
			Description:           "Current active development environment.",
			ClusterName:           "ocp-dev",
			KubernetesNamespace:   "devops-ci-demo",
			TektonNamespace:       "devops-ci-demo",
			ArgoCDApplicationName: "demo-go-color-app",
			GitTargetBranch:       "main",
			ValidationPath:        "apps/demo-go-color-app",
			AllowTechnicalActions: true,
			AllowPromotionSource:  true,
			AllowPromotionTarget:  false,
		},
		{
			Name:                 "staging",
			DisplayName:          "Staging",
			Enabled:              false,
			Category:             "preproduction",
			Description:          "Future controlled staging environment. Not enabled yet.",
			ClusterName:          "ocp-staging",
			AllowPromotionTarget: true,
		},
		{
			Name:                 "production",
			DisplayName:          "Production",
			Enabled:              false,
			Category:             "production",
			Description:          "Future production environment. Not enabled yet.",
			ClusterName:          "ocp-production",
			AllowPromotionTarget: true,
		},
	}, "dev")
}

func LoadEnvironmentCatalogFromFile(path string) (EnvironmentCatalog, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return EnvironmentCatalog{}, errors.New("environment catalog file path is empty")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return EnvironmentCatalog{}, fmt.Errorf("read environment catalog file: %w", err)
	}

	return ParseEnvironmentCatalogYAML(content)
}

func ParseEnvironmentCatalogYAML(content []byte) (EnvironmentCatalog, error) {
	var parsed environmentCatalogFile
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		return EnvironmentCatalog{}, fmt.Errorf("parse environment catalog yaml: %w", err)
	}
	if len(parsed.Environments) == 0 {
		return EnvironmentCatalog{}, errors.New("environment catalog does not define any environments")
	}

	catalog := NewEnvironmentCatalog(parsed.Environments, parsed.DefaultEnvironment)
	if _, ok := catalog.Resolve(catalog.DefaultEnvironment()); !ok {
		return EnvironmentCatalog{}, fmt.Errorf("default environment %q is not configured", catalog.DefaultEnvironment())
	}

	return catalog, nil
}

func NewEnvironmentCatalog(definitions []EnvironmentDefinition, defaultEnvironment string) EnvironmentCatalog {
	catalog := EnvironmentCatalog{
		defaultEnvironment: normalizeEnvironmentName(defaultEnvironment),
		environments:       map[string]EnvironmentDefinition{},
	}

	for _, definition := range definitions {
		name := normalizeEnvironmentName(definition.Name)
		if name == "" {
			continue
		}
		definition.Name = name
		definition.ClusterName = normalizeClusterName(definition.ClusterName)
		definition.ValidationPath = strings.TrimSpace(definition.ValidationPath)
		catalog.environments[name] = definition
	}

	if catalog.defaultEnvironment == "" {
		catalog.defaultEnvironment = "dev"
	}

	return catalog
}

func (c EnvironmentCatalog) DefaultEnvironment() string {
	if strings.TrimSpace(c.defaultEnvironment) == "" {
		return "dev"
	}
	return c.defaultEnvironment
}

func (c EnvironmentCatalog) Resolve(name string) (EnvironmentDefinition, bool) {
	definition, ok := c.environments[normalizeEnvironmentName(name)]
	return definition, ok
}

func (c EnvironmentCatalog) ValidateCreateTargetEnvironment(name string) error {
	name = normalizeEnvironmentName(name)
	if name == "" {
		return errors.New("targetEnvironment is required")
	}

	definition, ok := c.Resolve(name)
	if !ok {
		return fmt.Errorf("targetEnvironment %q is not configured", name)
	}
	if !definition.Enabled {
		return fmt.Errorf("targetEnvironment %q is currently disabled", name)
	}

	return nil
}

func (c EnvironmentCatalog) IsEnabled(name string) bool {
	definition, ok := c.Resolve(name)
	return ok && definition.Enabled
}

func (c EnvironmentCatalog) AllowsTechnicalActions(name string) bool {
	definition, ok := c.Resolve(name)
	return ok && definition.Enabled && definition.AllowTechnicalActions
}

func isAllowedTargetEnvironment(value string) bool {
	return DefaultEnvironmentCatalog().IsEnabled(value)
}

func normalizeEnvironmentName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
