package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainWiresRuntimeSecretValueLoaderBuilder(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	body := string(content)

	required := []string{
		"var runtimeSecretValueLoader app.RuntimeSecretValueLoader = app.EmptyRuntimeSecretValueLoader{}",
		"buildRuntimeSecretValueLoader(",
		"runtime secret value loader initialization failed",
		"app.KubernetesSecretValueLoaderConfig{}",
	}
	for _, item := range required {
		if !strings.Contains(body, item) {
			t.Fatalf("main.go missing %q", item)
		}
	}
}

func TestMainDoesNotWireAllowListLoaderDirectly(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	body := string(content)

	forbidden := []string{
		"app.NewAllowListKubernetesSecretValueLoader(",
		"app.AllowListKubernetesSecretValueLoader{",
	}
	for _, item := range forbidden {
		if strings.Contains(body, item) {
			t.Fatalf("main.go must not wire allow-list loader directly; found %q", item)
		}
	}
}

func TestMainKeepsConcreteRuntimeFactoriesDisabled(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	body := string(content)

	forbidden := []string{
		"newKubernetesRuntimeClientFactoryAdapter(",
		"newTektonRuntimeClientFactoryAdapter(",
		"newArgoCDRuntimeClientFactoryAdapter(",
	}
	for _, item := range forbidden {
		if strings.Contains(body, item) {
			t.Fatalf("main.go must keep concrete factories disabled; found %q", item)
		}
	}
}
