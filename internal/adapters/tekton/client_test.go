package tekton

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRequiresAPIURL(t *testing.T) {
	_, err := New(Config{Token: "token"})
	if err == nil {
		t.Fatal("want API URL validation error")
	}
}
func TestNewRequiresToken(t *testing.T) {
	_, err := New(Config{APIURL: "https://api.example.local:6443"})
	if err == nil {
		t.Fatal("want token validation error")
	}
}
func TestCreatePipelineRun(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s", r.Method)
		}
		if r.URL.Path != "/apis/tekton.dev/v1/namespaces/devops-ci-demo/pipelineruns" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("bad auth")
		}
		var p map[string]any
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Fatal(err)
		}
		spec := p["spec"].(map[string]any)
		if spec["pipelineRef"].(map[string]any)["name"] != "go-build-and-push" {
			t.Fatalf("bad pipelineRef")
		}
		params := spec["params"].([]any)
		var gotValidationPath string
		for _, raw := range params {
			param := raw.(map[string]any)
			if param["name"] == "VALIDATION_PATH" {
				gotValidationPath, _ = param["value"].(string)
			}
		}
		if gotValidationPath != "apps/demo-go-color-app" {
			t.Fatalf("VALIDATION_PATH = %q", gotValidationPath)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"metadata": map[string]any{"name": "devops-cp-validate-chg-2026-0005-abcde", "namespace": "devops-ci-demo", "uid": "uid-123"}})
	}))
	defer srv.Close()
	c, err := New(Config{APIURL: srv.URL, Token: "test-token"}, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}
	ref, err := c.CreatePipelineRun(context.Background(), CreatePipelineRunRequest{Namespace: "devops-ci-demo", PipelineName: "go-build-and-push", ServiceAccountName: "pipeline", GenerateName: "devops-cp-validate-chg-2026-0005-", ChangeNumber: "CHG-2026-0005", GitURL: "https://github.com/vincmarz/demo-go-color-app.git", GitRevision: "main", Image: "image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app:latest", WorkspacePVC: "pipeline-workspace", DockerConfigSecret: "pipeline-registry-dockerconfig", ValidationPath: "apps/demo-go-color-app"})
	if err != nil {
		t.Fatal(err)
	}
	if ref.Name != "devops-cp-validate-chg-2026-0005-abcde" || ref.Namespace != "devops-ci-demo" || ref.UID != "uid-123" {
		t.Fatalf("bad ref %#v", ref)
	}
}

func TestFindLatestPipelineRunByChange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method=%s, want GET", r.Method)
		}
		if r.URL.Path != "/apis/tekton.dev/v1/namespaces/devops-ci-demo/pipelineruns" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if got := r.URL.Query().Get("labelSelector"); got != "devops-control-plane/change-number=CHG-2026-0005" {
			t.Fatalf("labelSelector=%q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{{"metadata": map[string]any{"name": "older", "namespace": "devops-ci-demo", "uid": "uid-old", "creationTimestamp": "2026-06-26T16:09:08Z"}, "status": map[string]any{"conditions": []map[string]any{{"type": "Succeeded", "status": "False", "reason": "Failed", "message": "failed"}}}}, {"metadata": map[string]any{"name": "newer", "namespace": "devops-ci-demo", "uid": "uid-new", "creationTimestamp": "2026-06-26T16:16:33Z"}, "status": map[string]any{"completionTime": "2026-06-26T16:18:32Z", "conditions": []map[string]any{{"type": "Succeeded", "status": "True", "reason": "Succeeded", "message": "Tasks Completed: 2"}}}}}})
	}))
	defer server.Close()
	client, err := New(Config{APIURL: server.URL, Token: "test-token"}, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatal(err)
	}
	status, err := client.FindLatestPipelineRunByChange(context.Background(), "devops-ci-demo", "CHG-2026-0005")
	if err != nil {
		t.Fatalf("FindLatestPipelineRunByChange returned error: %v", err)
	}
	if status.Name != "newer" || status.Status != "True" || status.Reason != "Succeeded" {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestFindLatestPipelineRunByChangeNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()
	client, err := New(Config{APIURL: server.URL, Token: "test-token"}, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.FindLatestPipelineRunByChange(context.Background(), "devops-ci-demo", "CHG-2026-0005")
	if err == nil {
		t.Fatal("FindLatestPipelineRunByChange returned nil error, want not found error")
	}
}
