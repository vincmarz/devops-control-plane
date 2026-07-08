package app

import (
	"context"
	"errors"
	"strings"
)

// ErrTektonRuntimeClientFactoryNotConfigured is returned when no Tekton runtime
// client factory is configured. Without an explicit factory, no Tekton runtime
// client can be built from Secret references.
var ErrTektonRuntimeClientFactoryNotConfigured = errors.New("tekton runtime client factory is not configured")

// ErrTektonRuntimeClientFactoryDisabled is returned when a Tekton runtime
// client factory refuses to build a client because the runtime provider target
// is disabled.
var ErrTektonRuntimeClientFactoryDisabled = errors.New("tekton runtime client factory is disabled for the requested target")

// TektonRuntimeClientFactoryRequest is the input contract used to build a
// per-cluster Tekton runtime client from Secret references and Secret values
// loaded at runtime.
type TektonRuntimeClientFactoryRequest struct {
	Target       TechnicalRuntimeTarget
	SecretRefs   RuntimeClientSecretRefs
	SecretValues RuntimeSecretValueSet
}

// TektonRuntimeClientFactory builds a TektonRuntimeClient for a specific
// runtime target. Implementations must never log or serialize Secret values,
// never build a client for a disabled or unknown cluster and never share
// mutable state across clusters.
type TektonRuntimeClientFactory interface {
	BuildTektonRuntimeClient(ctx context.Context, request TektonRuntimeClientFactoryRequest) (TektonRuntimeClient, error)
}

// EmptyTektonRuntimeClientFactory is the conservative default factory. It never
// builds a client and always fails with ErrTektonRuntimeClientFactoryNotConfigured.
type EmptyTektonRuntimeClientFactory struct{}

// BuildTektonRuntimeClient always returns ErrTektonRuntimeClientFactoryNotConfigured.
func (EmptyTektonRuntimeClientFactory) BuildTektonRuntimeClient(_ context.Context, _ TektonRuntimeClientFactoryRequest) (TektonRuntimeClient, error) {
	return nil, ErrTektonRuntimeClientFactoryNotConfigured
}

// ValidateTektonRuntimeClientFactoryRequest checks the request against the
// baseline guardrails before any concrete factory attempts to build a client.
//
// This function never inspects Secret values; it only validates references and
// metadata.
func ValidateTektonRuntimeClientFactoryRequest(request TektonRuntimeClientFactoryRequest) error {
	if err := request.Target.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(request.Target.TektonNamespace) == "" {
		return errors.New("tekton runtime client factory request requires target TektonNamespace")
	}
	if strings.TrimSpace(request.Target.TektonPipelineName) == "" {
		return errors.New("tekton runtime client factory request requires target TektonPipelineName")
	}
	if strings.TrimSpace(request.SecretRefs.ClusterName) == "" {
		return errors.New("tekton runtime client factory request requires runtime secret references clusterName")
	}
	if request.SecretRefs.Tekton == nil {
		return errors.New("tekton runtime client factory request requires tekton secret reference")
	}
	if err := request.SecretRefs.Validate(); err != nil {
		return err
	}
	return nil
}

// RequiredTektonRuntimeSecretValueKeys returns the set of Secret value keys
// that a concrete Tekton runtime client factory must find in the Secret value
// set before building a real client.
//
// Tekton is accessed through the Kubernetes API of the target cluster,
// therefore it reuses the Kubernetes-family Secret value keys.
//
// This helper does not read any Secret. It only enumerates the keys that the
// implementation contract expects.
func RequiredTektonRuntimeSecretValueKeys(refs RuntimeClientSecretRefs) []RuntimeSecretValueKey {
	if refs.Tekton == nil {
		return nil
	}
	keys := make([]RuntimeSecretValueKey, 0, 3)
	if strings.TrimSpace(refs.Tekton.TokenKey) != "" {
		keys = append(keys, RuntimeSecretValueKubernetesToken)
	}
	if strings.TrimSpace(refs.Tekton.KubeconfigKey) != "" {
		keys = append(keys, RuntimeSecretValueKubernetesKubeconfig)
	}
	if strings.TrimSpace(refs.Tekton.CAKey) != "" {
		keys = append(keys, RuntimeSecretValueKubernetesCA)
	}
	return keys
}
