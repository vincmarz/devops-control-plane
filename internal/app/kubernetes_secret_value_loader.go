package app

import (
	"context"
	"errors"
	"strings"
)

// KubernetesSecretValueLoaderAllowedRef identifies a Kubernetes Secret that a
// KubernetesSecretValueLoader is allowed to read at runtime.
//
// Only Secrets whose cluster, namespace and name match an entry in the
// allow-list can be loaded. This is the primary guardrail that keeps runtime
// Secret reads under explicit configuration control.
type KubernetesSecretValueLoaderAllowedRef struct {
	ClusterName string
	Namespace   string
	Name        string
}

// KubernetesSecretValueLoaderConfig is the immutable configuration used to
// build a KubernetesSecretValueLoader.
//
// AllowedRefs is the mandatory allow-list of Secrets that the loader is
// allowed to read. AllowedClusters, when non-empty, further restricts the
// loader to a specific set of cluster names.
//
// The configuration itself does not load any Secret; it only defines the
// static contract used by the concrete implementations.
type KubernetesSecretValueLoaderConfig struct {
	AllowedClusters []string
	AllowedRefs     []KubernetesSecretValueLoaderAllowedRef
}

// ErrKubernetesSecretValueLoaderClusterNotAllowed is returned when a runtime
// Secret load is attempted for a cluster that is not in the allow-list.
var ErrKubernetesSecretValueLoaderClusterNotAllowed = errors.New("kubernetes secret value loader: cluster is not in the allow-list")

// ErrKubernetesSecretValueLoaderRefNotAllowed is returned when a runtime Secret
// load is attempted for a (cluster, namespace, name) triple that is not in the
// allow-list.
var ErrKubernetesSecretValueLoaderRefNotAllowed = errors.New("kubernetes secret value loader: secret reference is not in the allow-list")

// ErrKubernetesSecretValueLoaderMissingKubernetesRef is returned when a runtime
// Secret load is attempted for RuntimeClientSecretRefs that do not carry a
// Kubernetes secret reference.
var ErrKubernetesSecretValueLoaderMissingKubernetesRef = errors.New("kubernetes secret value loader: kubernetes secret reference is required")

// KubernetesSecretValueLoader is the contract that concrete implementations
// must satisfy to read Secret values for the Kubernetes-family of runtime
// clients.
//
// Implementations are expected to:
//   - honour the AllowedClusters/AllowedRefs allow-list without exception
//   - never log or serialize Secret values
//   - never build a client themselves; they only return RuntimeSecretValueSet
//   - fail closed on any misconfiguration, unknown cluster or unknown ref
type KubernetesSecretValueLoader interface {
	RuntimeSecretValueLoader
}

// DisabledKubernetesSecretValueLoader is the conservative default that
// satisfies the KubernetesSecretValueLoader contract without reading any
// Secret. It always returns ErrRuntimeSecretValueLoaderNotConfigured.
type DisabledKubernetesSecretValueLoader struct {
	Config KubernetesSecretValueLoaderConfig
}

// LoadRuntimeSecretValues always returns ErrRuntimeSecretValueLoaderNotConfigured.
func (l DisabledKubernetesSecretValueLoader) LoadRuntimeSecretValues(_ context.Context, _ RuntimeClientSecretRefs) (RuntimeSecretValueSet, error) {
	return RuntimeSecretValueSet{}, ErrRuntimeSecretValueLoaderNotConfigured
}

// ValidateKubernetesSecretValueLoaderRequest checks whether the given
// RuntimeClientSecretRefs are allowed by the provided configuration. It never
// reads any Secret; it only enforces the allow-list guardrails and the
// structural requirements of the Kubernetes secret reference.
func ValidateKubernetesSecretValueLoaderRequest(config KubernetesSecretValueLoaderConfig, refs RuntimeClientSecretRefs) error {
	clusterName := strings.TrimSpace(refs.ClusterName)
	if clusterName == "" {
		return errors.New("kubernetes secret value loader request requires runtime secret references clusterName")
	}
	if refs.Kubernetes == nil {
		return ErrKubernetesSecretValueLoaderMissingKubernetesRef
	}
	if err := refs.Validate(); err != nil {
		return err
	}
	if len(config.AllowedClusters) > 0 {
		if !kubernetesSecretValueLoaderClusterAllowed(clusterName, config.AllowedClusters) {
			return ErrKubernetesSecretValueLoaderClusterNotAllowed
		}
	}
	namespace := strings.TrimSpace(refs.Kubernetes.Namespace)
	name := strings.TrimSpace(refs.Kubernetes.Name)
	if namespace == "" || name == "" {
		return errors.New("kubernetes secret value loader request requires kubernetes secret namespace and name")
	}
	if !kubernetesSecretValueLoaderRefAllowed(config.AllowedRefs, clusterName, namespace, name) {
		return ErrKubernetesSecretValueLoaderRefNotAllowed
	}
	return nil
}

func kubernetesSecretValueLoaderClusterAllowed(clusterName string, allowed []string) bool {
	for _, allowedCluster := range allowed {
		if strings.TrimSpace(allowedCluster) == clusterName {
			return true
		}
	}
	return false
}

func kubernetesSecretValueLoaderRefAllowed(allowed []KubernetesSecretValueLoaderAllowedRef, clusterName, namespace, name string) bool {
	for _, allowedRef := range allowed {
		if strings.TrimSpace(allowedRef.ClusterName) != clusterName {
			continue
		}
		if strings.TrimSpace(allowedRef.Namespace) != namespace {
			continue
		}
		if strings.TrimSpace(allowedRef.Name) != name {
			continue
		}
		return true
	}
	return false
}
