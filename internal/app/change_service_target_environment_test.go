package app

import "testing"

func TestIsAllowedTargetEnvironment(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		allowed bool
	}{
		{name: "dev is allowed", value: "dev", allowed: true},
		{name: "staging is not enabled yet", value: "staging", allowed: false},
		{name: "production is not enabled yet", value: "production", allowed: false},
		{name: "unknown environment is rejected", value: "unknown-env", allowed: false},
		{name: "empty value is rejected by helper", value: "", allowed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAllowedTargetEnvironment(tt.value); got != tt.allowed {
				t.Fatalf("isAllowedTargetEnvironment(%q) = %v, want %v", tt.value, got, tt.allowed)
			}
		})
	}
}
