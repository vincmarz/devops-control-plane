package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	argocdadapter "github.com/vincmarz/devops-control-plane/internal/adapters/argocd"
)

func TestCurrentArgoCDRuntimeClientMapsObservedRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"metadata":{"name":"demo-go-color-app"},"spec":{"project":"default","source":{"repoURL":"https://github.com/vincmarz/demo-app-gitops.git","targetRevision":"main","path":"manifests"}},"status":{"sync":{"status":"Synced","revision":"abc123"},"health":{"status":"Healthy"}}}`))
	}))
	defer server.Close()

	client, err := argocdadapter.New(argocdadapter.Config{BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	result, err := (currentArgoCDRuntimeClient{client: client}).CheckDeployment(context.Background(), "demo-go-color-app")
	if err != nil {
		t.Fatal(err)
	}
	if result.RepositoryURL != "https://github.com/vincmarz/demo-app-gitops.git" || result.TargetRevision != "main" || result.Revision != "abc123" {
		t.Fatalf("result = %#v", result)
	}
}
