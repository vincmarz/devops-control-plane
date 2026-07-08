package app

import (
	"context"
	"errors"
)

// RuntimeSecretValueKey identifies a specific Secret value inside a
// RuntimeSecretValueSet. Keys are stable and safe to log, values are not.
type RuntimeSecretValueKey string

const (
	RuntimeSecretValueKubernetesToken      RuntimeSecretValueKey = "kubernetesToken"
	RuntimeSecretValueKubernetesCA         RuntimeSecretValueKey = "kubernetesCA"
	RuntimeSecretValueKubernetesKubeconfig RuntimeSecretValueKey = "kubernetesKubeconfig"
	RuntimeSecretValueArgoCDToken          RuntimeSecretValueKey = "argocdToken"
	RuntimeSecretValueArgoCDBaseURL        RuntimeSecretValueKey = "argocdBaseURL"
	RuntimeSecretValueArgoCDCA             RuntimeSecretValueKey = "argocdCA"
)

var ErrRuntimeSecretValueLoaderNotConfigured = errors.New("runtime secret value loader is not configured")

var ErrRuntimeSecretValueNotAvailable = errors.New("runtime secret value is not available")

// RuntimeSecretValueSet groups Secret values resolved for a specific cluster.
//
// Design invariants:
//   - Values are stored only in memory and are never logged or serialized.
//   - SafeSummary intentionally exposes only which keys are present.
//   - There is no Values() accessor that returns the raw map.
type RuntimeSecretValueSet struct {
	values map[RuntimeSecretValueKey]string
}

func NewRuntimeSecretValueSet(values map[RuntimeSecretValueKey]string) RuntimeSecretValueSet {
	copied := map[RuntimeSecretValueKey]string{}
	for key, value := range values {
		if key == "" {
			continue
		}
		copied[key] = value
	}
	return RuntimeSecretValueSet{values: copied}
}

func EmptyRuntimeSecretValueSet() RuntimeSecretValueSet {
	return RuntimeSecretValueSet{values: map[RuntimeSecretValueKey]string{}}
}

func (s RuntimeSecretValueSet) Has(key RuntimeSecretValueKey) bool {
	if s.values == nil {
		return false
	}
	_, ok := s.values[key]
	return ok
}

func (s RuntimeSecretValueSet) Resolve(key RuntimeSecretValueKey) (string, error) {
	if s.values == nil {
		return "", ErrRuntimeSecretValueNotAvailable
	}
	value, ok := s.values[key]
	if !ok {
		return "", ErrRuntimeSecretValueNotAvailable
	}
	return value, nil
}

func (s RuntimeSecretValueSet) Keys() []RuntimeSecretValueKey {
	if len(s.values) == 0 {
		return nil
	}
	out := make([]RuntimeSecretValueKey, 0, len(s.values))
	for key := range s.values {
		out = append(out, key)
	}
	return out
}

func (s RuntimeSecretValueSet) IsEmpty() bool {
	return len(s.values) == 0
}

func (s RuntimeSecretValueSet) SafeSummary() map[string]any {
	present := map[string]bool{}
	for key := range s.values {
		present[string(key)] = true
	}
	return map[string]any{
		"keysPresent": present,
		"empty":       len(s.values) == 0,
	}
}

// RuntimeSecretValueLoader loads Secret values for a specific set of
// RuntimeClientSecretRefs. Implementations must never log or serialize
// Secret values.
type RuntimeSecretValueLoader interface {
	LoadRuntimeSecretValues(ctx context.Context, refs RuntimeClientSecretRefs) (RuntimeSecretValueSet, error)
}

// EmptyRuntimeSecretValueLoader is the conservative default loader.
// It never reads any Secret and always returns ErrRuntimeSecretValueLoaderNotConfigured.
type EmptyRuntimeSecretValueLoader struct{}

func (EmptyRuntimeSecretValueLoader) LoadRuntimeSecretValues(_ context.Context, _ RuntimeClientSecretRefs) (RuntimeSecretValueSet, error) {
	return RuntimeSecretValueSet{}, ErrRuntimeSecretValueLoaderNotConfigured
}
