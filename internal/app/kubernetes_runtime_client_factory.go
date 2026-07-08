package app

import (
	"context"
	"errors"
	"strings"
)

// ErrKubernetesRuntimeClientFactoryNotConfigured is returned when no Kubernetes
// runtime client factory is configured. Without an explicit factory, no
// Kubernetes runtime client can be built from Secret references.
var ErrKubernetesRuntimeClientFactoryNotConfigured = errors.New("kubernetes runtime client factory is not configured")

// ErrKubernetesRuntimeClientFactoryDisabled is returned when a Kubernetes
// runtime client factory refuses to build a client because the runtime provider
// target is disabled.
var ErrKubernetesRuntimeClientFactoryDisabled = errors.New("kubernetes runtime client factory is disabled for the requested target")

// KubernetesRuntimeClientFactoryRequest is the input contract used to build a
// per-cluster Kubernetes runtime evidence client from Secret references and
// Secret values loaded at runtime.
type KubernetesRuntimeClientFactoryRequest struct {
	Target       TechnicalRuntimeTarget
	SecretRefs   RuntimeClientSecretRefs
	SecretValues RuntimeSecretValueSet
}

// KubernetesRuntimeClientFactory builds a KubernetesRuntimeEvidenceClient for a
// specific runtime target. Implementations must never log or serialize Secret
// values, never build a client for a disabled or unknown cluster and never
// share mutable state across clusters.
type KubernetesRuntimeClientFactory interface {
	BuildKubernetesRuntimeEvidenceClient(ctx context.Context, request KubernetesRuntimeClientFactoryRequest) (KubernetesRuntimeEvidenceClient, error)
}

// EmptyKubernetesRuntimeClientFactory is the conservative default factory. It
// never builds a client and always fails with
// ErrKubernetesRuntimeClientFactoryNotConfigured.
type EmptyKubernetesRuntimeClientFactory struct{}

// BuildKubernetesRuntimeEvidenceClient always returns
// ErrKubernetesRuntimeClientFactoryNotConfigured.
func (EmptyKubernetesRuntimeClientFactory) BuildKubernetesRuntimeEvidenceClient(_ context.Context, _ KubernetesRuntimeClientFactoryRequest) (KubernetesRuntimeEvidenceClient, error) {
	return nil, ErrKubernetesRuntimeClientFactoryNotConfigured
}

// ValidateKubernetesRuntimeClientFactoryRequest checks the request against the
// baseline guardrails before any concrete factory attempts to build a client.
//
// This function never inspects Secret values; it only validates references and
// metadata.
func ValidateKubernetesRuntimeClientFactoryRequest(request KubernetesRuntimeClientFactoryRequest) error {
	if err := request.Target.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(request.SecretRefs.ClusterName) == "" {
		return errors.New("kubernetes runtime client factory request requires runtime secret references clusterName")
	}
	if request.SecretRefs.Kubernetes == nil {
		return errors.New("kubernetes runtime client factory request requires kubernetes secret reference")
	}
	if err := request.SecretRefs.Validate(); err != nil {
		return err
	}
	return nil
}

// RequiredKubernetesRuntimeSecretValueKeys returns the set of Secret value keys
// that a concrete Kubernetes runtime client factory must find in the Secret
// value set before building a real client.
//
// This helper does not read any Secret. It only enumerates the keys that the
// implementation contract expects.
func RequiredKubernetesRuntimeSecretValueKeys(refs RuntimeClientSecretRefs) []RuntimeSecretValueKey {
	if refs.Kubernetes == nil {
		return nil
	}
	keys := make([]RuntimeSecretValueKey, 0, 3)
	if strings.TrimSpace(refs.Kubernetes.TokenKey) != "" {
		keys = append(keys, RuntimeSecretValueKubernetesToken)
	}
	if strings.TrimSpace(refs.Kubernetes.KubeconfigKey) != "" {
		keys = append(keys, RuntimeSecretValueKubernetesKubeconfig)
	}
	if strings.TrimSpace(refs.Kubernetes.CAKey) != "" {
		keys = append(keys, RuntimeSecretValueKubernetesCA)
	}
	return keys
}
