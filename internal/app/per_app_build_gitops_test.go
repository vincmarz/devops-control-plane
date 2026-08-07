package app

import "testing"

// TestResolveBuildImage verifies that StartBuild resolves the build image from
// the per-application source binding, falling back to the global build image
// when the binding does not set one. This enables truly independent
// applications (e.g. demo-go-quote-app) to publish to their own image stream
// instead of the global demo-go-color-app default.
func TestResolveBuildImage(t *testing.T) {
	cases := []struct {
		name     string
		image    string
		fallback string
		want     string
	}{
		{"per-app override", "reg/ns/demo-go-quote-app", "reg/ns/demo-go-color-app", "reg/ns/demo-go-quote-app"},
		{"empty falls back to global", "", "reg/ns/demo-go-color-app", "reg/ns/demo-go-color-app"},
		{"whitespace falls back to global", "   ", "reg/ns/demo-go-color-app", "reg/ns/demo-go-color-app"},
		{"override is trimmed", "  reg/ns/demo-go-quote-app  ", "reg/ns/demo-go-color-app", "reg/ns/demo-go-quote-app"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveBuildImage(RepositoryBinding{BuildImage: tc.image}, tc.fallback)
			if got != tc.want {
				t.Fatalf("resolveBuildImage() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestResolveKustomizationPath verifies that UpdateGitOps resolves the GitOps
// kustomization path from the per-application GitOps binding, falling back to
// the global configured path when the binding does not set one. This lets
// multiple applications share a single GitOps repository while each targets its
// own apps/<app>/kustomization.yaml.
func TestResolveKustomizationPath(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		fallback string
		want     string
	}{
		{"per-app override", "apps/demo-go-quote-app/kustomization.yaml", "apps/demo-go-color-app/kustomization.yaml", "apps/demo-go-quote-app/kustomization.yaml"},
		{"empty falls back to global", "", "apps/demo-go-color-app/kustomization.yaml", "apps/demo-go-color-app/kustomization.yaml"},
		{"whitespace falls back to global", "   ", "apps/demo-go-color-app/kustomization.yaml", "apps/demo-go-color-app/kustomization.yaml"},
		{"override is trimmed", "  apps/demo-go-quote-app/kustomization.yaml  ", "apps/demo-go-color-app/kustomization.yaml", "apps/demo-go-quote-app/kustomization.yaml"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveKustomizationPath(RepositoryBinding{KustomizationPath: tc.path}, tc.fallback)
			if got != tc.want {
				t.Fatalf("resolveKustomizationPath() = %q, want %q", got, tc.want)
			}
		})
	}
}
