package argocd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/applications/demo-go-color-app" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("unexpected authorization header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"metadata":{"name":"demo-go-color-app","namespace":"openshift-gitops"},"spec":{"project":"default","source":{"repoURL":"https://gitlab.example.local/devops/demo.git","targetRevision":"main","path":"manifests"},"destination":{"server":"https://kubernetes.default.svc","namespace":"devops-ci-demo"}},"status":{"sync":{"status":"Synced","revision":"abc123"},"health":{"status":"Healthy"}}}`))
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, AuthToken: "token-123", TimeoutSeconds: 5})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	app, err := client.GetApplication(context.Background(), "demo-go-color-app")
	if err != nil {
		t.Fatalf("GetApplication() error = %v", err)
	}
	if app.Name != "demo-go-color-app" || app.SyncStatus != "Synced" || app.HealthStatus != "Healthy" || app.CurrentRevision != "abc123" {
		t.Fatalf("unexpected app: %+v", app)
	}
}

func TestListApplications(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/applications" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"metadata":{"name":"demo-go-color-app","namespace":"openshift-gitops"},"spec":{"project":"default"},"status":{"sync":{"status":"Synced","revision":"abc123"},"health":{"status":"Healthy"}}}]}`))
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	apps, err := client.ListApplications(context.Background())
	if err != nil {
		t.Fatalf("ListApplications() error = %v", err)
	}
	if len(apps) != 1 || apps[0].Name != "demo-go-color-app" {
		t.Fatalf("unexpected apps: %+v", apps)
	}
}
