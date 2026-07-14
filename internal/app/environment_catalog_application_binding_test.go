package app

import (
	"strings"
	"testing"
)

func TestEnvironmentCatalogValidateApplicationBinding(t *testing.T) {
	catalog := NewEnvironmentCatalog([]EnvironmentDefinition{
		{
			Name:            "dev",
			ApplicationName: "demo-go-color-app",
			Enabled:         true,
		},
		{
			Name:    "staging",
			Enabled: true,
		},
	}, "dev")

	tests := []struct {
		name              string
		targetEnvironment string
		applicationName   string
		wantErr           string
	}{
		{
			name:              "matching binding",
			targetEnvironment: "dev",
			applicationName:   "demo-go-color-app",
		},
		{
			name:              "surrounding spaces are trimmed",
			targetEnvironment: " dev ",
			applicationName:   " demo-go-color-app ",
		},
		{
			name:              "different application is rejected",
			targetEnvironment: "dev",
			applicationName:   "payments",
			wantErr:           `applicationName "payments" is not configured for targetEnvironment "dev"; expected "demo-go-color-app"`,
		},
		{
			name:              "application comparison is case sensitive",
			targetEnvironment: "dev",
			applicationName:   "Demo-Go-Color-App",
			wantErr:           `applicationName "Demo-Go-Color-App" is not configured for targetEnvironment "dev"; expected "demo-go-color-app"`,
		},
		{
			name:              "missing catalog binding is rejected",
			targetEnvironment: "staging",
			applicationName:   "demo-go-color-app",
			wantErr:           `targetEnvironment "staging" does not define applicationName`,
		},
		{
			name:              "unknown environment is rejected",
			targetEnvironment: "unknown",
			applicationName:   "demo-go-color-app",
			wantErr:           `targetEnvironment "unknown" is not configured`,
		},
		{
			name:            "missing environment is rejected",
			applicationName: "demo-go-color-app",
			wantErr:         "targetEnvironment is required",
		},
		{
			name:              "missing application is rejected",
			targetEnvironment: "dev",
			wantErr:           "applicationName is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := catalog.ValidateApplicationBinding(test.targetEnvironment, test.applicationName)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateApplicationBinding returned error %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateApplicationBinding returned nil error, want %q", test.wantErr)
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("ValidateApplicationBinding error = %q, want %q", err.Error(), test.wantErr)
			}
		})
	}
}
