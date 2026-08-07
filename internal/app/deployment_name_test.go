package app

import "testing"

// TestResolveDeploymentName verifies that evidence collection resolves the
// Kubernetes Deployment name from the environment-provided
// KubernetesDeploymentName, falling back to the application name when it is not
// set. This guards the scenario C fix where the logical application name
// (e.g. "demo-go-color-app-gitlab-full") differs from the real Deployment name
// (e.g. "demo-go-color-app").
func TestResolveDeploymentName(t *testing.T) {
	cases := []struct {
		name            string
		target          TechnicalRuntimeTarget
		applicationName string
		want            string
	}{
		{
			name:            "explicit deployment name overrides application name",
			target:          TechnicalRuntimeTarget{KubernetesDeploymentName: "demo-go-color-app"},
			applicationName: "demo-go-color-app-gitlab-full",
			want:            "demo-go-color-app",
		},
		{
			name:            "empty deployment name falls back to application name",
			target:          TechnicalRuntimeTarget{KubernetesDeploymentName: ""},
			applicationName: "demo-go-color-app",
			want:            "demo-go-color-app",
		},
		{
			name:            "whitespace deployment name falls back to application name",
			target:          TechnicalRuntimeTarget{KubernetesDeploymentName: "   "},
			applicationName: "demo-go-color-app",
			want:            "demo-go-color-app",
		},
		{
			name:            "explicit deployment name is trimmed",
			target:          TechnicalRuntimeTarget{KubernetesDeploymentName: "  demo-go-color-app  "},
			applicationName: "demo-go-color-app-gitlab-full",
			want:            "demo-go-color-app",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveDeploymentName(tc.target, tc.applicationName)
			if got != tc.want {
				t.Fatalf("resolveDeploymentName() = %q, want %q", got, tc.want)
			}
		})
	}
}
