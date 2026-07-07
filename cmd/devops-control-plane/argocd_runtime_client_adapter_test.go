package main

import (
	"context"
	"strings"
	"testing"
)

func TestCurrentArgoCDRuntimeClientRequiresClient(t *testing.T) {
	client := currentArgoCDRuntimeClient{}
	_, err := client.CheckDeployment(context.Background(), "demo-go-color-app")
	if err == nil {
		t.Fatal("CheckDeployment returned nil error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("error = %q, want not configured", err.Error())
	}
}
