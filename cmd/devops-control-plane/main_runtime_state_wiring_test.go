package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainWiresChangeRuntimeStateStore(t *testing.T) {
	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile main.go: %v", err)
	}
	text := string(content)
	want := "app.WithChangeRuntimeStateStore(repositories.RuntimeStates)"
	if !strings.Contains(text, want) {
		t.Fatalf("main.go does not contain %q", want)
	}
}
