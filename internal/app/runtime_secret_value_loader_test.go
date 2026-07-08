package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestNewRuntimeSecretValueSetIgnoresEmptyKeys(t *testing.T) {
	set := NewRuntimeSecretValueSet(map[RuntimeSecretValueKey]string{
		RuntimeSecretValueKubernetesToken: "tok",
		"":                                "ignored",
	})

	if !set.Has(RuntimeSecretValueKubernetesToken) {
		t.Fatal("expected kubernetesToken to be present")
	}
	if got, err := set.Resolve(RuntimeSecretValueKubernetesToken); err != nil || got != "tok" {
		t.Fatalf("Resolve(kubernetesToken) = %q, %v; want %q, nil", got, err, "tok")
	}
}

func TestRuntimeSecretValueSetResolveReturnsNotAvailableWhenMissing(t *testing.T) {
	set := NewRuntimeSecretValueSet(nil)
	if _, err := set.Resolve(RuntimeSecretValueKubernetesToken); !errors.Is(err, ErrRuntimeSecretValueNotAvailable) {
		t.Fatalf("Resolve on empty set = %v; want ErrRuntimeSecretValueNotAvailable", err)
	}
}

func TestRuntimeSecretValueSetHas(t *testing.T) {
	set := NewRuntimeSecretValueSet(map[RuntimeSecretValueKey]string{
		RuntimeSecretValueArgoCDToken: "argo",
	})

	if !set.Has(RuntimeSecretValueArgoCDToken) {
		t.Fatal("Has(argocdToken) = false, want true")
	}
	if set.Has(RuntimeSecretValueKubernetesToken) {
		t.Fatal("Has(kubernetesToken) = true, want false")
	}
}

func TestRuntimeSecretValueSetKeysAndIsEmpty(t *testing.T) {
	empty := EmptyRuntimeSecretValueSet()
	if !empty.IsEmpty() {
		t.Fatal("empty set IsEmpty = false, want true")
	}
	if len(empty.Keys()) != 0 {
		t.Fatalf("empty set Keys len = %d, want 0", len(empty.Keys()))
	}

	set := NewRuntimeSecretValueSet(map[RuntimeSecretValueKey]string{
		RuntimeSecretValueKubernetesToken: "tok",
		RuntimeSecretValueArgoCDToken:     "argo",
	})

	if set.IsEmpty() {
		t.Fatal("populated set IsEmpty = true, want false")
	}
	if len(set.Keys()) != 2 {
		t.Fatalf("Keys len = %d, want 2", len(set.Keys()))
	}
}

func TestRuntimeSecretValueSetSafeSummaryHidesValues(t *testing.T) {
	set := NewRuntimeSecretValueSet(map[RuntimeSecretValueKey]string{
		RuntimeSecretValueKubernetesToken: "top-secret-token",
		RuntimeSecretValueArgoCDToken:     "another-secret",
	})

	summary := set.SafeSummary()

	raw, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("json marshal SafeSummary error = %v", err)
	}

	body := string(raw)

	if strings.Contains(body, "top-secret-token") {
		t.Fatal("SafeSummary must not expose Secret values")
	}
	if strings.Contains(body, "another-secret") {
		t.Fatal("SafeSummary must not expose Secret values")
	}
	if !strings.Contains(body, string(RuntimeSecretValueKubernetesToken)) {
		t.Fatal("SafeSummary should list kubernetesToken as present key")
	}
	if !strings.Contains(body, string(RuntimeSecretValueArgoCDToken)) {
		t.Fatal("SafeSummary should list argocdToken as present key")
	}
	if !strings.Contains(body, "keysPresent") {
		t.Fatal("SafeSummary should include keysPresent field")
	}
}

func TestEmptyRuntimeSecretValueLoaderAlwaysReturnsNotConfigured(t *testing.T) {
	var loader RuntimeSecretValueLoader = EmptyRuntimeSecretValueLoader{}

	refs := RuntimeClientSecretRefs{ClusterName: "ocp-dev"}

	set, err := loader.LoadRuntimeSecretValues(context.Background(), refs)
	if !errors.Is(err, ErrRuntimeSecretValueLoaderNotConfigured) {
		t.Fatalf("LoadRuntimeSecretValues error = %v, want ErrRuntimeSecretValueLoaderNotConfigured", err)
	}
	if !set.IsEmpty() {
		t.Fatal("EmptyRuntimeSecretValueLoader must return an empty RuntimeSecretValueSet")
	}
}
