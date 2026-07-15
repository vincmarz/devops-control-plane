package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainWiresApplicationGitOpsBindingIntoArgoCD(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	required := []string{
		"gitOpsRepositoryTargetResolver.Resolve(change.ApplicationName, app.GitOpsConsumerArgoCD)",
		"deployment.DeclaredRepositoryURL = gitOpsTarget.RepositoryURL",
		"deployment.DeclaredDefaultBranch = gitOpsTarget.DefaultBranch",
	}
	for _, value := range required {
		if !strings.Contains(text, value) {
			t.Fatalf("main.go does not contain %q", value)
		}
	}
}
