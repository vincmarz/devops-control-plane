package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainWiresRuntimeClientFactoryBuilders(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	body := string(content)

	required := []string{
		"buildRuntimeKubernetesClientFactory(cfg)",
		"buildRuntimeTektonClientFactory(cfg)",
		"buildRuntimeArgoCDClientFactory(cfg)",
		"runtime kubernetes client factory initialization failed",
		"runtime tekton client factory initialization failed",
		"runtime argocd client factory initialization failed",
		"kubernetesRuntimeClientFactory,",
		"tektonRuntimeClientFactory,",
		"argoCDRuntimeClientFactory,",
	}
	for _, item := range required {
		if !strings.Contains(body, item) {
			t.Fatalf("main.go missing %q", item)
		}
	}
}

func TestMainDoesNotInstantiateRuntimeFactoriesDirectlyInRegistries(t *testing.T) {
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
			t.Fatalf("main.go must use gated factory builders, found direct constructor %q", item)
		}
	}
}
