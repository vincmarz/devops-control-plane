package app

import (
	"os"
	"strings"
	"testing"
)

func TestChangeServiceCreateValidatesApplicationBinding(t *testing.T) {
	content, err := os.ReadFile("change_service.go")
	if err != nil {
		t.Fatalf("read change_service.go: %v", err)
	}

	source := string(content)
	environmentValidation := "environmentCatalog.ValidateCreateTargetEnvironment(req.TargetEnvironment)"
	applicationValidation := "environmentCatalog.ValidateApplicationBinding(req.TargetEnvironment, req.ApplicationName)"
	clusterResolution := "DefaultEnvironmentClusterResolver().ResolveEnabledTarget(req.TargetEnvironment)"

	environmentIndex := strings.Index(source, environmentValidation)
	applicationIndex := strings.Index(source, applicationValidation)
	clusterIndex := strings.Index(source, clusterResolution)

	if environmentIndex < 0 {
		t.Fatal("ChangeService Create does not validate targetEnvironment")
	}
	if applicationIndex < 0 {
		t.Fatal("ChangeService Create does not validate the application binding")
	}
	if clusterIndex < 0 {
		t.Fatal("ChangeService Create does not resolve the enabled cluster target")
	}
	if !(environmentIndex < applicationIndex && applicationIndex < clusterIndex) {
		t.Fatal("application binding validation must run after environment validation and before cluster resolution")
	}
}
