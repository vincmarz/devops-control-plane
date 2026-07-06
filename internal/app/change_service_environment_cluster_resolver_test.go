package app

import (
	"os"
	"strings"
	"testing"
)

func TestChangeServiceCreatePathUsesEnvironmentClusterResolver(t *testing.T) {
	content, err := os.ReadFile("change_service.go")
	if err != nil {
		t.Fatalf("ReadFile(change_service.go) returned error %v", err)
	}

	if !strings.Contains(string(content), "DefaultEnvironmentClusterResolver().ResolveEnabledTarget(req.TargetEnvironment)") {
		t.Fatal("ChangeService create path should resolve targetEnvironment through EnvironmentClusterResolver")
	}
}
