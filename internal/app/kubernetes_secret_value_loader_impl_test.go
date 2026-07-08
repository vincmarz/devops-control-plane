package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type fakeKubernetesSecretGetter struct {
	calls []fakeKubernetesSecretGetterCall
	data  map[string]map[string][]byte
	err   error
}

type fakeKubernetesSecretGetterCall struct {
	Namespace string
	Name      string
}

func (f *fakeKubernetesSecretGetter) GetSecret(_ context.Context, namespace string, name string) (map[string][]byte, error) {
	f.calls = append(f.calls, fakeKubernetesSecretGetterCall{Namespace: namespace, Name: name})
	if f.err != nil {
		return nil, f.err
	}
	if data, ok := f.data[namespace+"/"+name]; ok {
		return data, nil
	}
	return nil, errors.New("fake getter: secret not found")
}

func newAllowedStagingConfig() KubernetesSecretValueLoaderConfig {
	return KubernetesSecretValueLoaderConfig{
		AllowedClusters: []string{"ocp-staging"},
		AllowedRefs: []KubernetesSecretValueLoaderAllowedRef{
			{ClusterName: "ocp-staging", Namespace: "devops-control-plane", Name: "ocp-staging-kube"},
		},
	}
}

func newValidStagingRefs() RuntimeClientSecretRefs {
	return RuntimeClientSecretRefs{
		ClusterName: "ocp-staging",
		Kubernetes: &RuntimeSecretReference{
			Namespace: "devops-control-plane",
			Name:      "ocp-staging-kube",
			TokenKey:  "token",
			CAKey:     "ca.crt",
		},
	}
}

func TestAllowListKubernetesSecretValueLoaderRejectsMissingGetter(t *testing.T) {
	loader := NewAllowListKubernetesSecretValueLoader(newAllowedStagingConfig(), nil)

	_, err := loader.LoadRuntimeSecretValues(context.Background(), newValidStagingRefs())
	if !errors.Is(err, ErrKubernetesSecretValueLoaderGetterNotConfigured) {
		t.Fatalf("LoadRuntimeSecretValues error = %v, want ErrKubernetesSecretValueLoaderGetterNotConfigured", err)
	}
}

func TestAllowListKubernetesSecretValueLoaderRejectsClusterOutsideAllowList(t *testing.T) {
	getter := &fakeKubernetesSecretGetter{}
	loader := NewAllowListKubernetesSecretValueLoader(newAllowedStagingConfig(), getter)

	refs := newValidStagingRefs()
	refs.ClusterName = "ocp-production"

	_, err := loader.LoadRuntimeSecretValues(context.Background(), refs)
	if !errors.Is(err, ErrKubernetesSecretValueLoaderClusterNotAllowed) {
		t.Fatalf("LoadRuntimeSecretValues error = %v, want ErrKubernetesSecretValueLoaderClusterNotAllowed", err)
	}
	if len(getter.calls) != 0 {
		t.Fatalf("getter must not be called when cluster is not allowed, calls = %d", len(getter.calls))
	}
}

func TestAllowListKubernetesSecretValueLoaderRejectsRefOutsideAllowList(t *testing.T) {
	getter := &fakeKubernetesSecretGetter{}
	loader := NewAllowListKubernetesSecretValueLoader(newAllowedStagingConfig(), getter)

	refs := newValidStagingRefs()
	refs.Kubernetes.Namespace = "unexpected-namespace"

	_, err := loader.LoadRuntimeSecretValues(context.Background(), refs)
	if !errors.Is(err, ErrKubernetesSecretValueLoaderRefNotAllowed) {
		t.Fatalf("LoadRuntimeSecretValues error = %v, want ErrKubernetesSecretValueLoaderRefNotAllowed", err)
	}
	if len(getter.calls) != 0 {
		t.Fatalf("getter must not be called when ref is not allowed, calls = %d", len(getter.calls))
	}
}

func TestAllowListKubernetesSecretValueLoaderReturnsTokenAndCA(t *testing.T) {
	getter := &fakeKubernetesSecretGetter{
		data: map[string]map[string][]byte{
			"devops-control-plane/ocp-staging-kube": {
				"token":  []byte("secret-token-value"),
				"ca.crt": []byte("secret-ca-value"),
			},
		},
	}
	loader := NewAllowListKubernetesSecretValueLoader(newAllowedStagingConfig(), getter)

	set, err := loader.LoadRuntimeSecretValues(context.Background(), newValidStagingRefs())
	if err != nil {
		t.Fatalf("LoadRuntimeSecretValues error = %v", err)
	}
	if set.IsEmpty() {
		t.Fatal("expected non-empty RuntimeSecretValueSet")
	}
	if got, err := set.Resolve(RuntimeSecretValueKubernetesToken); err != nil || got != "secret-token-value" {
		t.Fatalf("Resolve(kubernetesToken) = %q, %v; want %q, nil", got, err, "secret-token-value")
	}
	if got, err := set.Resolve(RuntimeSecretValueKubernetesCA); err != nil || got != "secret-ca-value" {
		t.Fatalf("Resolve(kubernetesCA) = %q, %v; want %q, nil", got, err, "secret-ca-value")
	}
	if len(getter.calls) != 1 {
		t.Fatalf("getter must be called exactly once, calls = %d", len(getter.calls))
	}
	if getter.calls[0].Namespace != "devops-control-plane" || getter.calls[0].Name != "ocp-staging-kube" {
		t.Fatalf("unexpected getter call = %+v", getter.calls[0])
	}
}

func TestAllowListKubernetesSecretValueLoaderRejectsMissingRequiredKey(t *testing.T) {
	getter := &fakeKubernetesSecretGetter{
		data: map[string]map[string][]byte{
			"devops-control-plane/ocp-staging-kube": {
				"token": []byte("secret-token-value"),
			},
		},
	}
	loader := NewAllowListKubernetesSecretValueLoader(newAllowedStagingConfig(), getter)

	refs := newValidStagingRefs()
	refs.Kubernetes.CAKey = "missing-ca-key"

	_, err := loader.LoadRuntimeSecretValues(context.Background(), refs)
	if !errors.Is(err, ErrKubernetesSecretValueLoaderKeyMissing) {
		t.Fatalf("LoadRuntimeSecretValues error = %v, want ErrKubernetesSecretValueLoaderKeyMissing", err)
	}
}

func TestAllowListKubernetesSecretValueLoaderPropagatesGetterError(t *testing.T) {
	getterErr := errors.New("boom")
	getter := &fakeKubernetesSecretGetter{err: getterErr}
	loader := NewAllowListKubernetesSecretValueLoader(newAllowedStagingConfig(), getter)

	_, err := loader.LoadRuntimeSecretValues(context.Background(), newValidStagingRefs())
	if !errors.Is(err, getterErr) {
		t.Fatalf("LoadRuntimeSecretValues error = %v, want %v", err, getterErr)
	}
}

func TestAllowListKubernetesSecretValueLoaderSafeSummaryDoesNotExposeValues(t *testing.T) {
	getter := &fakeKubernetesSecretGetter{
		data: map[string]map[string][]byte{
			"devops-control-plane/ocp-staging-kube": {
				"token":  []byte("top-secret-token"),
				"ca.crt": []byte("top-secret-ca"),
			},
		},
	}
	loader := NewAllowListKubernetesSecretValueLoader(newAllowedStagingConfig(), getter)

	set, err := loader.LoadRuntimeSecretValues(context.Background(), newValidStagingRefs())
	if err != nil {
		t.Fatalf("LoadRuntimeSecretValues error = %v", err)
	}

	summary := set.SafeSummary()
	raw, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("json marshal SafeSummary error = %v", err)
	}
	body := string(raw)

	if strings.Contains(body, "top-secret-token") {
		t.Fatal("SafeSummary must not expose Secret values")
	}
	if strings.Contains(body, "top-secret-ca") {
		t.Fatal("SafeSummary must not expose Secret values")
	}
}
func TestAllowListKubernetesSecretValueLoaderReturnsOnlyTokenWhenCANotConfigured(t *testing.T) {
	getter := &fakeKubernetesSecretGetter{
		data: map[string]map[string][]byte{
			"devops-control-plane/ocp-staging-kube": {
				"token": []byte("secret-token-value"),
			},
		},
	}
	loader := NewAllowListKubernetesSecretValueLoader(newAllowedStagingConfig(), getter)

	refs := newValidStagingRefs()
	refs.Kubernetes.CAKey = ""

	set, err := loader.LoadRuntimeSecretValues(context.Background(), refs)
	if err != nil {
		t.Fatalf("LoadRuntimeSecretValues error = %v", err)
	}
	if set.IsEmpty() {
		t.Fatal("expected non-empty RuntimeSecretValueSet when token key is configured")
	}
	if got, err := set.Resolve(RuntimeSecretValueKubernetesToken); err != nil || got != "secret-token-value" {
		t.Fatalf("Resolve(kubernetesToken) = %q, %v; want %q, nil", got, err, "secret-token-value")
	}
	if set.Has(RuntimeSecretValueKubernetesCA) {
		t.Fatal("RuntimeSecretValueSet must not contain kubernetesCA when CAKey is not configured")
	}
}
