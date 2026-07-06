package app

import "testing"

func TestDefaultEnvironmentCatalogValidateCreateTargetEnvironment(t *testing.T) {
	catalog := DefaultEnvironmentCatalog()

	tests := []struct {
		name    string
		value   string
		wantErr string
	}{
		{name: "dev is enabled", value: "dev", wantErr: ""},
		{name: "staging is configured but disabled", value: "staging", wantErr: "targetEnvironment \"staging\" is currently disabled"},
		{name: "production is configured but disabled", value: "production", wantErr: "targetEnvironment \"production\" is currently disabled"},
		{name: "unknown environment is not configured", value: "unknown-env", wantErr: "targetEnvironment \"unknown-env\" is not configured"},
		{name: "empty value is rejected", value: "", wantErr: "targetEnvironment is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := catalog.ValidateCreateTargetEnvironment(tt.value)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateCreateTargetEnvironment(%q) returned error %v", tt.value, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateCreateTargetEnvironment(%q) returned nil error, want %q", tt.value, tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("ValidateCreateTargetEnvironment(%q) error = %q, want %q", tt.value, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestDefaultEnvironmentCatalogMetadata(t *testing.T) {
	catalog := DefaultEnvironmentCatalog()

	if got := catalog.DefaultEnvironment(); got != "dev" {
		t.Fatalf("DefaultEnvironment() = %q, want dev", got)
	}

	dev, ok := catalog.Resolve("dev")
	if !ok {
		t.Fatal("dev environment not found")
	}
	if !dev.Enabled {
		t.Fatal("dev environment should be enabled")
	}
	if !dev.AllowTechnicalActions {
		t.Fatal("dev environment should allow technical actions")
	}

	production, ok := catalog.Resolve("production")
	if !ok {
		t.Fatal("production environment should be configured")
	}
	if production.Enabled {
		t.Fatal("production environment should be disabled initially")
	}
	if production.AllowTechnicalActions {
		t.Fatal("production environment should not allow technical actions initially")
	}
}

func TestIsAllowedTargetEnvironmentUsesCatalog(t *testing.T) {
	if !isAllowedTargetEnvironment("dev") {
		t.Fatal("dev should be allowed by the default catalog")
	}
	if isAllowedTargetEnvironment("staging") {
		t.Fatal("staging should not be allowed while disabled")
	}
	if isAllowedTargetEnvironment("production") {
		t.Fatal("production should not be allowed while disabled")
	}
	if isAllowedTargetEnvironment("unknown-env") {
		t.Fatal("unknown-env should not be allowed when not configured")
	}
}
