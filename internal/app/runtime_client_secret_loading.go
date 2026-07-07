package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultRuntimeClientSecretRefsPath = "/etc/dcp-runtime-client-secrets/secret-refs.yaml"

// RuntimeClientSecretRefsDocument is the file-level representation of runtime
// client Secret references.
//
// The document stores only Kubernetes Secret references and key names. It must
// never store actual token, kubeconfig, CA, password or API credential values.
type RuntimeClientSecretRefsDocument struct {
	Clusters []RuntimeClientSecretRefs `yaml:"clusters" json:"clusters"`
}

// RuntimeClientSecretRefsRegistry stores validated Secret references by cluster
// name.
type RuntimeClientSecretRefsRegistry struct {
	refs map[string]RuntimeClientSecretRefs
}

// NewRuntimeClientSecretRefsRegistry validates and indexes runtime client
// Secret reference definitions.
func NewRuntimeClientSecretRefsRegistry(items []RuntimeClientSecretRefs) (RuntimeClientSecretRefsRegistry, error) {
	registry := RuntimeClientSecretRefsRegistry{refs: map[string]RuntimeClientSecretRefs{}}
	for _, item := range items {
		clusterName := normalizeRuntimeProviderClusterName(item.ClusterName)
		if clusterName == "" {
			return RuntimeClientSecretRefsRegistry{}, errors.New("clusterName is required for runtime client secret references")
		}
		item.ClusterName = clusterName
		if err := item.Validate(); err != nil {
			return RuntimeClientSecretRefsRegistry{}, err
		}
		registry.refs[clusterName] = item
	}
	return registry, nil
}

// EmptyRuntimeClientSecretRefsRegistry returns an empty registry. It is used as
// the safe fallback before real multi-cluster Secret references are configured.
func EmptyRuntimeClientSecretRefsRegistry() RuntimeClientSecretRefsRegistry {
	return RuntimeClientSecretRefsRegistry{refs: map[string]RuntimeClientSecretRefs{}}
}

// LoadRuntimeClientSecretRefsRegistryFromFile loads Secret references from a
// YAML file.
//
// Only references are loaded. No Kubernetes Secret values are read here.
func LoadRuntimeClientSecretRefsRegistryFromFile(path string) (RuntimeClientSecretRefsRegistry, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return RuntimeClientSecretRefsRegistry{}, errors.New("runtime client secret references file path is required")
	}

	content, err := os.ReadFile(trimmedPath)
	if err != nil {
		return RuntimeClientSecretRefsRegistry{}, err
	}

	var document RuntimeClientSecretRefsDocument
	if err := yaml.Unmarshal(content, &document); err != nil {
		return RuntimeClientSecretRefsRegistry{}, fmt.Errorf("failed to parse runtime client secret references file %q: %w", trimmedPath, err)
	}

	return NewRuntimeClientSecretRefsRegistry(document.Clusters)
}

// DefaultRuntimeClientSecretRefsRegistry loads the optional runtime Secret
// reference registry from the configured file path.
//
// If the file is not present, an empty registry is returned. This preserves the
// current conservative runtime behavior and avoids making staging or production
// active accidentally.
func DefaultRuntimeClientSecretRefsRegistry() RuntimeClientSecretRefsRegistry {
	path := strings.TrimSpace(os.Getenv("DCP_RUNTIME_CLIENT_SECRET_REFS_FILE"))
	if path == "" {
		path = defaultRuntimeClientSecretRefsPath
	}

	registry, err := LoadRuntimeClientSecretRefsRegistryFromFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return EmptyRuntimeClientSecretRefsRegistry()
		}
		return EmptyRuntimeClientSecretRefsRegistry()
	}
	return registry
}

// Resolve returns the Secret references configured for a cluster.
func (r RuntimeClientSecretRefsRegistry) Resolve(clusterName string) (RuntimeClientSecretRefs, bool) {
	if r.refs == nil {
		return RuntimeClientSecretRefs{}, false
	}
	item, ok := r.refs[normalizeRuntimeProviderClusterName(clusterName)]
	return item, ok
}

// SafeSummary returns a non-sensitive representation of the registry.
func (r RuntimeClientSecretRefsRegistry) SafeSummary() []map[string]any {
	if r.refs == nil {
		return nil
	}
	summary := make([]map[string]any, 0, len(r.refs))
	for _, item := range r.refs {
		summary = append(summary, item.SafeSummary())
	}
	return summary
}
