package app

import (
	"errors"
	"fmt"
	"strings"
)

// RuntimeSecretReference describes where a runtime client credential is stored.
//
// This model only stores Kubernetes Secret references and Secret key names. It
// must never store the secret value itself. It is intentionally independent from
// concrete Kubernetes, Tekton and Argo CD client construction so it can be used
// safely as a configuration contract before real multi-cluster clients are
// introduced.
type RuntimeSecretReference struct {
	Namespace     string `yaml:"namespace" json:"namespace"`
	Name          string `yaml:"name" json:"name"`
	TokenKey      string `yaml:"tokenKey,omitempty" json:"tokenKey,omitempty"`
	CAKey         string `yaml:"caKey,omitempty" json:"caKey,omitempty"`
	KubeconfigKey string `yaml:"kubeconfigKey,omitempty" json:"kubeconfigKey,omitempty"`
	BaseURLKey    string `yaml:"baseURLKey,omitempty" json:"baseURLKey,omitempty"`
}

// RuntimeClientSecretRefs groups the secret references required to build real
// runtime clients for a cluster.
//
// In phase 15.8.1 this structure is a safe model only. No runtime code reads
// Secret values and no staging or production provider is enabled by this model.
type RuntimeClientSecretRefs struct {
	ClusterName string                  `yaml:"clusterName" json:"clusterName"`
	Kubernetes  *RuntimeSecretReference `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty"`
	Tekton      *RuntimeSecretReference `yaml:"tekton,omitempty" json:"tekton,omitempty"`
	ArgoCD      *RuntimeSecretReference `yaml:"argocd,omitempty" json:"argocd,omitempty"`
}

// Validate checks that the model contains only safe references and that each
// configured reference contains enough metadata to locate a Kubernetes Secret.
func (r RuntimeClientSecretRefs) Validate() error {
	clusterName := normalizeRuntimeProviderClusterName(r.ClusterName)
	if clusterName == "" {
		return errors.New("clusterName is required for runtime client secret references")
	}

	items := []struct {
		label string
		ref   *RuntimeSecretReference
	}{
		{label: "kubernetes", ref: r.Kubernetes},
		{label: "tekton", ref: r.Tekton},
		{label: "argocd", ref: r.ArgoCD},
	}

	for _, item := range items {
		if item.ref == nil {
			continue
		}
		if err := item.ref.Validate(item.label); err != nil {
			return fmt.Errorf("cluster %q %s secret reference invalid: %w", clusterName, item.label, err)
		}
	}

	return nil
}

// Validate checks a single secret reference.
func (r RuntimeSecretReference) Validate(label string) error {
	if strings.TrimSpace(r.Namespace) == "" {
		return errors.New("namespace is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(r.TokenKey) == "" && strings.TrimSpace(r.KubeconfigKey) == "" {
		return errors.New("tokenKey or kubeconfigKey is required")
	}
	if err := rejectSecretLikeValue("namespace", r.Namespace); err != nil {
		return err
	}
	if err := rejectSecretLikeValue("name", r.Name); err != nil {
		return err
	}
	if err := rejectSecretLikeValue("tokenKey", r.TokenKey); err != nil {
		return err
	}
	if err := rejectSecretLikeValue("caKey", r.CAKey); err != nil {
		return err
	}
	if err := rejectSecretLikeValue("kubeconfigKey", r.KubeconfigKey); err != nil {
		return err
	}
	if err := rejectSecretLikeValue("baseURLKey", r.BaseURLKey); err != nil {
		return err
	}
	return nil
}

// SafeSummary returns a non-sensitive representation suitable for logs,
// diagnostics and documentation.
func (r RuntimeClientSecretRefs) SafeSummary() map[string]any {
	return map[string]any{
		"clusterName": normalizeRuntimeProviderClusterName(r.ClusterName),
		"kubernetes":  safeSecretReferenceSummary(r.Kubernetes),
		"tekton":      safeSecretReferenceSummary(r.Tekton),
		"argocd":      safeSecretReferenceSummary(r.ArgoCD),
	}
}

func safeSecretReferenceSummary(ref *RuntimeSecretReference) map[string]any {
	if ref == nil {
		return map[string]any{"configured": false}
	}
	return map[string]any{
		"configured":    true,
		"namespace":     strings.TrimSpace(ref.Namespace),
		"name":          strings.TrimSpace(ref.Name),
		"tokenKey":      strings.TrimSpace(ref.TokenKey),
		"caKey":         strings.TrimSpace(ref.CAKey),
		"kubeconfigKey": strings.TrimSpace(ref.KubeconfigKey),
		"baseURLKey":    strings.TrimSpace(ref.BaseURLKey),
	}
}

func rejectSecretLikeValue(field string, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	lower := strings.ToLower(trimmed)
	suspicious := []string{
		"bearer ",
		"-----begin",
		"-----end",
		"eyj",
		"kubeconfig:",
		"client-certificate-data",
		"client-key-data",
		"password=",
		"token=",
	}
	for _, marker := range suspicious {
		if strings.Contains(lower, marker) {
			return fmt.Errorf("%s appears to contain a secret value instead of a secret reference", field)
		}
	}
	return nil
}
