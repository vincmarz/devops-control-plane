package app

import (
	"context"
	"errors"
	"strings"
)

// KubernetesSecretGetter is the minimal contract that a concrete Kubernetes
// client must satisfy in order for AllowListKubernetesSecretValueLoader to
// read Secrets at runtime.
//
// Implementations must return the Secret data as a map of key to raw bytes,
// preserving the on-cluster representation. They must never log or serialize
// Secret values.
type KubernetesSecretGetter interface {
	GetSecret(ctx context.Context, namespace string, name string) (map[string][]byte, error)
}

// ErrKubernetesSecretValueLoaderGetterNotConfigured is returned by
// AllowListKubernetesSecretValueLoader when it is constructed without a
// KubernetesSecretGetter.
var ErrKubernetesSecretValueLoaderGetterNotConfigured = errors.New("kubernetes secret value loader: getter is not configured")

// ErrKubernetesSecretValueLoaderKeyMissing is returned when the Secret does
// not contain a key that the Secret reference declared as required.
var ErrKubernetesSecretValueLoaderKeyMissing = errors.New("kubernetes secret value loader: required key is missing from the secret")

// AllowListKubernetesSecretValueLoader implements RuntimeSecretValueLoader by
// combining an allow-list configuration with a KubernetesSecretGetter that
// reads Secrets from a Kubernetes API.
//
// This loader is the application-level orchestration point. It never talks to
// a Kubernetes API directly; it delegates to the Getter after enforcing the
// allow-list. All guardrails from ValidateKubernetesSecretValueLoaderRequest
// are applied before the Getter is invoked, and every step fails closed on
// misconfiguration.
type AllowListKubernetesSecretValueLoader struct {
	config KubernetesSecretValueLoaderConfig
	getter KubernetesSecretGetter
}

// NewAllowListKubernetesSecretValueLoader builds the loader from a config and a
// Getter. A nil Getter is preserved so the loader can be constructed as a
// placeholder; in that case LoadRuntimeSecretValues always returns
// ErrKubernetesSecretValueLoaderGetterNotConfigured.
func NewAllowListKubernetesSecretValueLoader(config KubernetesSecretValueLoaderConfig, getter KubernetesSecretGetter) AllowListKubernetesSecretValueLoader {
	return AllowListKubernetesSecretValueLoader{config: config, getter: getter}
}

// LoadRuntimeSecretValues enforces the allow-list, delegates the Secret read to
// the Getter, and maps the Kubernetes-family keys into a RuntimeSecretValueSet.
//
// The set never contains empty entries and never exposes Secret values via
// SafeSummary. Missing required keys always fail with
// ErrKubernetesSecretValueLoaderKeyMissing so the caller does not silently
// build a client with partial credentials.
func (l AllowListKubernetesSecretValueLoader) LoadRuntimeSecretValues(ctx context.Context, refs RuntimeClientSecretRefs) (RuntimeSecretValueSet, error) {
	if l.getter == nil {
		return RuntimeSecretValueSet{}, ErrKubernetesSecretValueLoaderGetterNotConfigured
	}
	if err := ValidateKubernetesSecretValueLoaderRequest(l.config, refs); err != nil {
		return RuntimeSecretValueSet{}, err
	}
	namespace := strings.TrimSpace(refs.Kubernetes.Namespace)
	name := strings.TrimSpace(refs.Kubernetes.Name)
	data, err := l.getter.GetSecret(ctx, namespace, name)
	if err != nil {
		return RuntimeSecretValueSet{}, err
	}
	values := map[RuntimeSecretValueKey]string{}
	tokenKey := strings.TrimSpace(refs.Kubernetes.TokenKey)
	if tokenKey != "" {
		value, ok := data[tokenKey]
		if !ok {
			return RuntimeSecretValueSet{}, ErrKubernetesSecretValueLoaderKeyMissing
		}
		values[RuntimeSecretValueKubernetesToken] = string(value)
	}
	caKey := strings.TrimSpace(refs.Kubernetes.CAKey)
	if caKey != "" {
		value, ok := data[caKey]
		if !ok {
			return RuntimeSecretValueSet{}, ErrKubernetesSecretValueLoaderKeyMissing
		}
		values[RuntimeSecretValueKubernetesCA] = string(value)
	}
	kubeconfigKey := strings.TrimSpace(refs.Kubernetes.KubeconfigKey)
	if kubeconfigKey != "" {
		value, ok := data[kubeconfigKey]
		if !ok {
			return RuntimeSecretValueSet{}, ErrKubernetesSecretValueLoaderKeyMissing
		}
		values[RuntimeSecretValueKubernetesKubeconfig] = string(value)
	}
	if len(values) == 0 {
		return EmptyRuntimeSecretValueSet(), nil
	}
	return NewRuntimeSecretValueSet(values), nil
}
