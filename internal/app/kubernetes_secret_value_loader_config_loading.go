package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// KubernetesSecretValueLoaderConfigDocument is the file-level representation of
// the allow-list used by AllowListKubernetesSecretValueLoader.
//
// The document stores only cluster names and Kubernetes Secret references. It
// must never contain Secret values.
type KubernetesSecretValueLoaderConfigDocument struct {
	AllowedClusters []string                                        `yaml:"allowedClusters" json:"allowedClusters"`
	AllowedRefs     []KubernetesSecretValueLoaderAllowedRefDocument `yaml:"allowedRefs" json:"allowedRefs"`
}

// KubernetesSecretValueLoaderAllowedRefDocument is the YAML/JSON representation
// of one allow-listed Secret reference.
type KubernetesSecretValueLoaderAllowedRefDocument struct {
	ClusterName string `yaml:"clusterName" json:"clusterName"`
	Namespace   string `yaml:"namespace" json:"namespace"`
	Name        string `yaml:"name" json:"name"`
}

// NewKubernetesSecretValueLoaderConfig validates and normalizes the static
// allow-list configuration used by the Kubernetes Secret value loader.
//
// This function does not read any Kubernetes Secret. It only validates and
// normalizes cluster names and Secret reference metadata.
func NewKubernetesSecretValueLoaderConfig(
	allowedClusters []string,
	allowedRefs []KubernetesSecretValueLoaderAllowedRef,
) (KubernetesSecretValueLoaderConfig, error) {
	config := KubernetesSecretValueLoaderConfig{
		AllowedClusters: make([]string, 0, len(allowedClusters)),
		AllowedRefs:     make([]KubernetesSecretValueLoaderAllowedRef, 0, len(allowedRefs)),
	}

	for _, clusterName := range allowedClusters {
		normalized := normalizeRuntimeProviderClusterName(clusterName)
		if normalized == "" {
			return KubernetesSecretValueLoaderConfig{}, errors.New("kubernetes secret value loader config allowedClusters contains an empty cluster name")
		}
		if err := rejectSecretLikeValue("allowedClusters", normalized); err != nil {
			return KubernetesSecretValueLoaderConfig{}, err
		}
		config.AllowedClusters = append(config.AllowedClusters, normalized)
	}

	for _, ref := range allowedRefs {
		normalizedRef, err := normalizeKubernetesSecretValueLoaderAllowedRef(ref)
		if err != nil {
			return KubernetesSecretValueLoaderConfig{}, err
		}
		config.AllowedRefs = append(config.AllowedRefs, normalizedRef)
	}

	return config, nil
}

// LoadKubernetesSecretValueLoaderConfigFromFile loads and validates the
// allow-list configuration from a YAML file.
//
// Only allow-list metadata is loaded. No Kubernetes Secret value is read here.
func LoadKubernetesSecretValueLoaderConfigFromFile(path string) (KubernetesSecretValueLoaderConfig, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return KubernetesSecretValueLoaderConfig{}, errors.New("kubernetes secret value loader config file path is required")
	}

	content, err := os.ReadFile(trimmedPath)
	if err != nil {
		return KubernetesSecretValueLoaderConfig{}, err
	}

	var document KubernetesSecretValueLoaderConfigDocument
	if err := yaml.Unmarshal(content, &document); err != nil {
		return KubernetesSecretValueLoaderConfig{}, fmt.Errorf("failed to parse kubernetes secret value loader config file %q: %w", trimmedPath, err)
	}

	refs := make([]KubernetesSecretValueLoaderAllowedRef, 0, len(document.AllowedRefs))
	for _, ref := range document.AllowedRefs {
		refs = append(refs, KubernetesSecretValueLoaderAllowedRef{
			ClusterName: ref.ClusterName,
			Namespace:   ref.Namespace,
			Name:        ref.Name,
		})
	}

	return NewKubernetesSecretValueLoaderConfig(document.AllowedClusters, refs)
}

func normalizeKubernetesSecretValueLoaderAllowedRef(ref KubernetesSecretValueLoaderAllowedRef) (KubernetesSecretValueLoaderAllowedRef, error) {
	clusterName := normalizeRuntimeProviderClusterName(ref.ClusterName)
	namespace := strings.TrimSpace(ref.Namespace)
	name := strings.TrimSpace(ref.Name)

	if clusterName == "" {
		return KubernetesSecretValueLoaderAllowedRef{}, errors.New("kubernetes secret value loader config allowedRefs requires clusterName")
	}
	if namespace == "" {
		return KubernetesSecretValueLoaderAllowedRef{}, errors.New("kubernetes secret value loader config allowedRefs requires namespace")
	}
	if name == "" {
		return KubernetesSecretValueLoaderAllowedRef{}, errors.New("kubernetes secret value loader config allowedRefs requires name")
	}
	if err := rejectSecretLikeValue("clusterName", clusterName); err != nil {
		return KubernetesSecretValueLoaderAllowedRef{}, err
	}
	if err := rejectSecretLikeValue("namespace", namespace); err != nil {
		return KubernetesSecretValueLoaderAllowedRef{}, err
	}
	if err := rejectSecretLikeValue("name", name); err != nil {
		return KubernetesSecretValueLoaderAllowedRef{}, err
	}

	return KubernetesSecretValueLoaderAllowedRef{
		ClusterName: clusterName,
		Namespace:   namespace,
		Name:        name,
	}, nil
}
