package main

import (
	"context"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/app"
)

func TestCurrentTektonRuntimeClientRequiresClient(t *testing.T) {
	client := currentTektonRuntimeClient{}
	_, err := client.CreatePipelineRun(context.Background(), app.TektonRuntimePipelineRunRequest{})
	if err == nil {
		t.Fatal("CreatePipelineRun returned nil error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("error = %q, want not configured", err.Error())
	}
}
