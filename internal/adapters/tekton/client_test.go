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
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"metadata": map[string]any{"name": "devops-cp-validate-chg-2026-0005-abcde", "namespace": "devops-ci-demo", "uid": "uid-123"}})
	}))
	defer srv.Close()
	c, err := New(Config{APIURL: srv.URL, Token: "test-token"}, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}
	ref, err := c.CreatePipelineRun(context.Background(), CreatePipelineRunRequest{Namespace: "devops-ci-demo", PipelineName: "go-build-and-push", ServiceAccountName: "pipeline", GenerateName: "devops-cp-validate-chg-2026-0005-", ChangeNumber: "CHG-2026-0005", GitURL: "https://github.com/vincmarz/demo-go-color-app.git", GitRevision: "main", Image: "image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app:latest", WorkspacePVC: "pipeline-workspace", DockerConfigSecret: "pipeline-registry-dockerconfig"})
	if err != nil {
		t.Fatal(err)
	}
	if ref.Name != "devops-cp-validate-chg-2026-0005-abcde" || ref.Namespace != "devops-ci-demo" || ref.UID != "uid-123" {
		t.Fatalf("bad ref %#v", ref)
	}
}
