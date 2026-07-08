package app

import (
	"context"
	"errors"
	"strings"
)

// ErrArgoCDRuntimeClientFactoryNotConfigured is returned when no Argo CD
// runtime client factory is configured. Without an explicit factory, no
// Argo CD runtime client can be built from Secret references.
var ErrArgoCDRuntimeClientFactoryNotConfigured = errors.New("argocd runtime client factory is not configured")

// ErrArgoCDRuntimeClientFactoryDisabled is returned when an Argo CD runtime
// client factory refuses to build a client because the runtime provider target
// is disabled.
var ErrArgoCDRuntimeClientFactoryDisabled = errors.New("argocd runtime client factory is disabled for the requested target")

// ArgoCDRuntimeClientFactoryRequest is the input contract used to build a
// per-cluster Argo CD runtime client from Secret references and Secret values
// loaded at runtime.
type ArgoCDRuntimeClientFactoryRequest struct {
	Target       TechnicalRuntimeTarget
	SecretRefs   RuntimeClientSecretRefs
	SecretValues RuntimeSecretValueSet
}

// ArgoCDRuntimeClientFactory builds an ArgoCDRuntimeClient for a specific
// runtime target. Implementations must never log or serialize Secret values,
// never build a client for a disabled or unknown cluster and never share
// mutable state across clusters.
type ArgoCDRuntimeClientFactory interface {
	BuildArgoCDRuntimeClient(ctx context.Context, request ArgoCDRuntimeClientFactoryRequest) (ArgoCDRuntimeClient, error)
}

// EmptyArgoCDRuntimeClientFactory is the conservative default factory. It never
// builds a client and always fails with ErrArgoCDRuntimeClientFactoryNotConfigured.
type EmptyArgoCDRuntimeClientFactory struct{}

// BuildArgoCDRuntimeClient always returns ErrArgoCDRuntimeClientFactoryNotConfigured.
func (EmptyArgoCDRuntimeClientFactory) BuildArgoCDRuntimeClient(_ context.Context, _ ArgoCDRuntimeClientFactoryRequest) (ArgoCDRuntimeClient, error) {
	return nil, ErrArgoCDRuntimeClientFactoryNotConfigured
}

// ValidateArgoCDRuntimeClientFactoryRequest checks the request against the
// baseline guardrails before any concrete factory attempts to build a client.
//
// This function never inspects Secret values; it only validates references and
// metadata.
func ValidateArgoCDRuntimeClientFactoryRequest(request ArgoCDRuntimeClientFactoryRequest) error {
	if err := request.Target.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(request.Target.ArgoCDApplicationName) == "" {
		return errors.New("argocd runtime client factory request requires target ArgoCDApplicationName")
	}
	if strings.TrimSpace(request.SecretRefs.ClusterName) == "" {
		return errors.New("argocd runtime client factory request requires runtime secret references clusterName")
	}
	if request.SecretRefs.ArgoCD == nil {
		return errors.New("argocd runtime client factory request requires argocd secret reference")
	}
	if err := request.SecretRefs.Validate(); err != nil {
		return err
	}
	return nil
}

// RequiredArgoCDRuntimeSecretValueKeys returns the set of Secret value keys
// that a concrete Argo CD runtime client factory must find in the Secret value
// set before building a real client.
//
// This helper does not read any Secret. It only enumerates the keys that the
// implementation contract expects.
func RequiredArgoCDRuntimeSecretValueKeys(refs RuntimeClientSecretRefs) []RuntimeSecretValueKey {
	if refs.ArgoCD == nil {
		return nil
	}
	keys := make([]RuntimeSecretValueKey, 0, 3)
	if strings.TrimSpace(refs.ArgoCD.TokenKey) != "" {
		keys = append(keys, RuntimeSecretValueArgoCDToken)
	}
	if strings.TrimSpace(refs.ArgoCD.BaseURLKey) != "" {
		keys = append(keys, RuntimeSecretValueArgoCDBaseURL)
	}
	if strings.TrimSpace(refs.ArgoCD.CAKey) != "" {
		keys = append(keys, RuntimeSecretValueArgoCDCA)
	}
	return keys
}
