package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const clusterRegistryFileEnv = "CLUSTER_REGISTRY_FILE"

type ClusterDefinition struct {
	Name              string   `yaml:"name"`
	DisplayName       string   `yaml:"displayName"`
	Enabled           bool     `yaml:"enabled"`
	APIURL            string   `yaml:"apiURL"`
	CAConfigMapRef    string   `yaml:"caConfigMapRef"`
	TokenSecretRef    string   `yaml:"tokenSecretRef"`
	DefaultNamespace  string   `yaml:"defaultNamespace"`
	AllowedNamespaces []string `yaml:"allowedNamespaces"`
	Description       string   `yaml:"description"`
}

type ClusterRegistry struct {
	clusters map[string]ClusterDefinition
}

type clusterRegistryFile struct {
	Clusters []ClusterDefinition `yaml:"clusters"`
}

func DefaultClusterRegistry() ClusterRegistry {
	registry, err := LoadClusterRegistryFromFile(os.Getenv(clusterRegistryFileEnv))
	if err == nil {
		return registry
	}

	return DefaultClusterRegistryFallback()
}

func DefaultClusterRegistryFallback() ClusterRegistry {
	return NewClusterRegistry([]ClusterDefinition{
		{
			Name:              "ocp-dev",
			DisplayName:       "OpenShift Development",
			Enabled:           true,
			APIURL:            "https://api.ocp4.mim.lan:6443",
			CAConfigMapRef:    "dcp-cluster-ocp-dev-ca",
			TokenSecretRef:    "dcp-cluster-ocp-dev-token",
			DefaultNamespace:  "devops-ci-demo",
			AllowedNamespaces: []string{"devops-ci-demo"},
			Description:       "Current OpenShift cluster used as the active dev environment.",
		},
		{
			Name:              "ocp-staging",
			DisplayName:       "OpenShift Staging",
			Enabled:           false,
			APIURL:            "",
			CAConfigMapRef:    "dcp-cluster-ocp-staging-ca",
			TokenSecretRef:    "dcp-cluster-ocp-staging-token",
			DefaultNamespace:  "devops-ci-staging",
			AllowedNamespaces: []string{"devops-ci-staging"},
			Description:       "Future staging/collaudo OpenShift cluster. Not connected yet.",
		},
		{
			Name:              "ocp-production",
			DisplayName:       "OpenShift Production",
			Enabled:           false,
			APIURL:            "",
			CAConfigMapRef:    "dcp-cluster-ocp-production-ca",
			TokenSecretRef:    "dcp-cluster-ocp-production-token",
			DefaultNamespace:  "devops-ci-production",
			AllowedNamespaces: []string{"devops-ci-production"},
			Description:       "Future production OpenShift cluster. Disabled until production gates are approved.",
		},
	})
}

func LoadClusterRegistryFromFile(path string) (ClusterRegistry, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return ClusterRegistry{}, errors.New("cluster registry file path is empty")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return ClusterRegistry{}, fmt.Errorf("read cluster registry file: %w", err)
	}

	return ParseClusterRegistryYAML(content)
}

func ParseClusterRegistryYAML(content []byte) (ClusterRegistry, error) {
	var parsed clusterRegistryFile
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		return ClusterRegistry{}, fmt.Errorf("parse cluster registry yaml: %w", err)
	}
	if len(parsed.Clusters) == 0 {
		return ClusterRegistry{}, errors.New("cluster registry does not define any clusters")
	}

	registry := NewClusterRegistry(parsed.Clusters)
	if len(registry.clusters) == 0 {
		return ClusterRegistry{}, errors.New("cluster registry does not contain any named clusters")
	}

	return registry, nil
}

func NewClusterRegistry(definitions []ClusterDefinition) ClusterRegistry {
	registry := ClusterRegistry{clusters: map[string]ClusterDefinition{}}

	for _, definition := range definitions {
		name := normalizeClusterName(definition.Name)
		if name == "" {
			continue
		}
		definition.Name = name
		registry.clusters[name] = definition
	}

	return registry
}

func (r ClusterRegistry) Resolve(name string) (ClusterDefinition, bool) {
	definition, ok := r.clusters[normalizeClusterName(name)]
	return definition, ok
}

func (r ClusterRegistry) IsEnabled(name string) bool {
	definition, ok := r.Resolve(name)
	return ok && definition.Enabled
}

func (r ClusterRegistry) ValidateConfiguredCluster(name string) error {
	name = normalizeClusterName(name)
	if name == "" {
		return errors.New("clusterName is required")
	}
	if _, ok := r.Resolve(name); !ok {
		return fmt.Errorf("clusterName %q is not configured", name)
	}
	return nil
}

func (r ClusterRegistry) ValidateEnvironmentCatalog(catalog EnvironmentCatalog) error {
	for environmentName, environment := range catalog.environments {
		clusterName := strings.TrimSpace(environment.ClusterName)
		if clusterName == "" {
			return fmt.Errorf("environment %q does not define clusterName", environmentName)
		}
		if err := r.ValidateConfiguredCluster(clusterName); err != nil {
			return fmt.Errorf("environment %q references invalid cluster: %w", environmentName, err)
		}
	}
	return nil
}

func normalizeClusterName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
