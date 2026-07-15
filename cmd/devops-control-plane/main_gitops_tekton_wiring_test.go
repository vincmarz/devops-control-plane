package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainWiresApplicationGitOpsBindingIntoTekton(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile main.go: %v", err)
	}
	text := string(content)
	required := []string{
		"app.NewGitOpsRepositoryTargetResolver(applicationCatalog.ResolveGitOpsBinding)",
		"gitOpsRepositoryTargetResolver.Resolve(change.ApplicationName, app.GitOpsConsumerTekton)",
		"revision := gitOpsTarget.DefaultBranch",
		"GitURL:             gitOpsTarget.RepositoryURL",
		"GitURL: gitOpsTarget.RepositoryURL, GitRevision: revision",
	}
	for _, value := range required {
		if !strings.Contains(text, value) {
			t.Fatalf("main.go does not contain %q", value)
		}
	}
}

func TestMainDoesNotUseGlobalTektonGitRepositoryAtRuntime(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile main.go: %v", err)
	}
	text := string(content)
	forbidden := []string{
		"GitURL:             cfg.TektonGitURL",
		"GitURL: cfg.TektonGitURL",
		"revision := cfg.TektonGitRevision",
	}
	for _, value := range forbidden {
		if strings.Contains(text, value) {
			t.Fatalf("main.go still contains global Tekton Git runtime wiring %q", value)
		}
	}
}
