package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainWiresFactoryAwareRuntimeProviderRegistries(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile main.go: %v", err)
	}
	checks := []string{
		"var runtimeSecretValueLoader app.RuntimeSecretValueLoader = app.EmptyRuntimeSecretValueLoader{}",
		"app.NewKubernetesRuntimeClientProviderFactoryAwareRegistry(",
		"app.NewTektonRuntimeClientProviderFactoryAwareRegistry(",
		"app.NewArgoCDRuntimeClientProviderFactoryAwareRegistry(",
		"tektonRuntimeClientProviderRegistry.Resolve(ctx, selection)",
		"argoCDRuntimeClientProviderRegistry.Resolve(ctx, selection)",
	}
	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Fatalf("main.go does not contain %q", check)
		}
	}
}

func TestMainDoesNotUseDirectTektonOrArgoCDDefaultRegistryResolution(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile main.go: %v", err)
	}
	forbidden := []string{
		"app.DefaultTektonRuntimeClientProviderRegistry(currentTektonRuntimeClient{client: tektonClient}).Resolve(ctx, selection)",
		"app.DefaultArgoCDRuntimeClientProviderRegistry(currentArgoCDRuntimeClient{client: argoCDClient}).Resolve(ctx, selection)",
	}
	for _, check := range forbidden {
		if strings.Contains(string(content), check) {
			t.Fatalf("main.go still contains direct default registry resolution %q", check)
		}
	}
}
