package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainWiresKubernetesRuntimeClientProviderRegistry(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	text := string(content)

	check := "app.WithKubernetesRuntimeClientProviderRegistry(app.DefaultKubernetesRuntimeClientProviderRegistry(kubernetesRuntimeClient))"
	if !strings.Contains(text, check) {
		t.Fatalf("main.go should contain %q", check)
	}
}
