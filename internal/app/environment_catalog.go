package app

import (
	"errors"
	"fmt"
	"strings"
)

type EnvironmentDefinition struct {
	Name                  string
	DisplayName           string
	Enabled               bool
	Category              string
	Description           string
	KubernetesNamespace   string
	TektonNamespace       string
	ArgoCDApplicationName string
	GitTargetBranch       string
	AllowTechnicalActions bool
	AllowPromotionSource  bool
	AllowPromotionTarget  bool
}

type EnvironmentCatalog struct {
	defaultEnvironment string
	environments       map[string]EnvironmentDefinition
}

func DefaultEnvironmentCatalog() EnvironmentCatalog {
	return NewEnvironmentCatalog([]EnvironmentDefinition{
		{
			Name:                  "dev",
			DisplayName:           "Development",
			Enabled:               true,
			Category:              "development",
			Description:           "Current active development environment.",
			KubernetesNamespace:   "devops-ci-demo",
			TektonNamespace:       "devops-ci-demo",
			ArgoCDApplicationName: "demo-go-color-app",
			GitTargetBranch:       "main",
			AllowTechnicalActions: true,
			AllowPromotionSource:  true,
			AllowPromotionTarget:  false,
		},
		{
			Name:                  "staging",
			DisplayName:           "Staging",
			Enabled:               false,
			Category:              "preproduction",
			Description:           "Future controlled staging environment. Not enabled yet.",
			KubernetesNamespace:   "",
			TektonNamespace:       "",
			ArgoCDApplicationName: "",
			GitTargetBranch:       "",
			AllowTechnicalActions: false,
			AllowPromotionSource:  false,
			AllowPromotionTarget:  true,
		},
		{
			Name:                  "production",
			DisplayName:           "Production",
			Enabled:               false,
			Category:              "production",
			Description:           "Future production environment. Not enabled yet.",
			KubernetesNamespace:   "",
			TektonNamespace:       "",
			ArgoCDApplicationName: "",
			GitTargetBranch:       "",
			AllowTechnicalActions: false,
			AllowPromotionSource:  false,
			AllowPromotionTarget:  true,
		},
	}, "dev")
}

func NewEnvironmentCatalog(definitions []EnvironmentDefinition, defaultEnvironment string) EnvironmentCatalog {
	catalog := EnvironmentCatalog{
		defaultEnvironment: strings.TrimSpace(defaultEnvironment),
		environments:       map[string]EnvironmentDefinition{},
	}

	for _, definition := range definitions {
		name := normalizeEnvironmentName(definition.Name)
		if name == "" {
			continue
		}
		definition.Name = name
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
