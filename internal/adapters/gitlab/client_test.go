package gitlab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRequiresBaseURL(t *testing.T) {
	_, err := New(Config{Token: "token"})
	if err == nil {
		t.Fatal("New returned nil error, want base URL validation error")
	}
}

func TestNewRequiresToken(t *testing.T) {
	_, err := New(Config{BaseURL: "https://gitlab.example.local"})
	if err == nil {
		t.Fatal("New returned nil error, want token validation error")
	}
}

func TestCreateBranch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v4/projects/1/repository/branches" {
			t.Fatalf("path = %s, want /api/v4/projects/1/repository/branches", r.URL.Path)
		}
		if got := r.Header.Get("PRIVATE-TOKEN"); got != "test-token" {
			t.Fatalf("PRIVATE-TOKEN = %q, want test-token", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm returned error: %v", err)
		}
		if got := r.Form.Get("branch"); got != "change/CHG-2026-0002" {
			t.Fatalf("branch = %q, want change/CHG-2026-0002", got)
		}
		if got := r.Form.Get("ref"); got != "main" {
			t.Fatalf("ref = %q, want main", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":      "change/CHG-2026-0002",
			"protected": false,
			"default":   false,
			"can_push":  true,
			"commit": map[string]any{
				"id":         "78d93eb6df08afc62a7dd7e3aa2235b5b276c98c",
				"short_id":   "78d93eb6",
				"created_at": "2026-06-26T15:59:10.000+02:00",
				"title":      "Initial commit",
				"message":    "Initial commit",
				"web_url":    "https://gitlab.example.local/commit/78d93eb6",
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL:        server.URL,
		Token:          "test-token",
		TimeoutSeconds: 5,
	}, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	branch, err := client.CreateBranch(context.Background(), 1, " change/CHG-2026-0002 ", " main ")
	if err != nil {
		t.Fatalf("CreateBranch returned error: %v", err)
	}
	if branch.Name != "change/CHG-2026-0002" {
		t.Fatalf("branch.Name = %q, want change/CHG-2026-0002", branch.Name)
	}
	if branch.Commit.ID != "78d93eb6df08afc62a7dd7e3aa2235b5b276c98c" {
		t.Fatalf("branch.Commit.ID = %q", branch.Commit.ID)
	}
	if !branch.CanPush {
		t.Fatal("branch.CanPush = false, want true")
	}
}

func TestCreateBranchValidation(t *testing.T) {
	client, err := New(Config{BaseURL: "https://gitlab.example.local", Token: "token"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	tests := []struct {
		name      string
		projectID int
		branch    string
		ref       string
	}{
		{name: "missing project", projectID: 0, branch: "change/test", ref: "main"},
		{name: "missing branch", projectID: 1, branch: "", ref: "main"},
		{name: "missing ref", projectID: 1, branch: "change/test", ref: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateBranch(context.Background(), tt.projectID, tt.branch, tt.ref)
			if err == nil {
				t.Fatal("CreateBranch returned nil error, want validation error")
			}
		})
	}
}

func TestCreateBranchReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Branch already exists"}`))
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, Token: "test-token"}, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	_, err = client.CreateBranch(context.Background(), 1, "change/CHG-2026-0002", "main")
	if err == nil {
		t.Fatal("CreateBranch returned nil error, want API error")
	}
}

func TestCreateOrUpdateFileCreatesFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.EscapedPath() != "/api/v4/projects/1/repository/files/manifests%2Fchg-2026-0003-control-plane.yaml" {
			t.Fatalf("escaped path = %s path = %s", r.URL.EscapedPath(), r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm returned error: %v", err)
		}
		if got := r.Form.Get("branch"); got != "change/CHG-2026-0003" {
			t.Fatalf("branch = %q", got)
		}
		if got := r.Form.Get("commit_message"); got != "Add generated manifest for CHG-2026-0003" {
			t.Fatalf("commit message = %q", got)
		}
		if got := r.Form.Get("content"); !strings.Contains(got, "changeNumber: CHG-2026-0003") {
			t.Fatalf("content does not contain change number: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"file_path": "manifests/chg-2026-0003-control-plane.yaml", "branch": "change/CHG-2026-0003"})
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, Token: "test-token"}, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	file, err := client.CreateOrUpdateFile(context.Background(), 1, "change/CHG-2026-0003", "manifests/chg-2026-0003-control-plane.yaml", "Add generated manifest for CHG-2026-0003", "changeNumber: CHG-2026-0003")
	if err != nil {
		t.Fatalf("CreateOrUpdateFile returned error: %v", err)
	}
	if file.FilePath != "manifests/chg-2026-0003-control-plane.yaml" {
		t.Fatalf("file path = %q", file.FilePath)
	}
}

func TestCreateOrUpdateFileFallsBackToUpdate(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"A file with this name already exists"}`))
		case http.MethodPut:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"file_path": "manifests/chg-2026-0003-control-plane.yaml", "branch": "change/CHG-2026-0003"})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, Token: "test-token"}, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	_, err = client.CreateOrUpdateFile(context.Background(), 1, "change/CHG-2026-0003", "manifests/chg-2026-0003-control-plane.yaml", "Update manifest", "content")
	if err != nil {
		t.Fatalf("CreateOrUpdateFile returned error: %v", err)
	}
	if len(methods) != 2 || methods[0] != http.MethodPost || methods[1] != http.MethodPut {
		t.Fatalf("methods = %#v, want POST then PUT", methods)
	}
}

func TestOpenMergeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v4/projects/1/merge_requests" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm returned error: %v", err)
		}
		if got := r.Form.Get("source_branch"); got != "change/CHG-2026-0003" {
			t.Fatalf("source_branch = %q", got)
		}
		if got := r.Form.Get("target_branch"); got != "main" {
			t.Fatalf("target_branch = %q", got)
		}
		if got := r.Form.Get("title"); got != "CHG-2026-0003 - GitOps change for demo-go-color-app" {
			t.Fatalf("title = %q", got)
		}
		if got := r.Form.Get("remove_source_branch"); got != "false" {
			t.Fatalf("remove_source_branch = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            10,
			"iid":           1,
			"project_id":    1,
			"title":         "CHG-2026-0003 - GitOps change for demo-go-color-app",
			"state":         "opened",
			"source_branch": "change/CHG-2026-0003",
			"target_branch": "main",
			"web_url":       "https://gitlab.example.local/group/project/-/merge_requests/1",
		})
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, Token: "test-token"}, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	mr, err := client.OpenMergeRequest(context.Background(), 1, "change/CHG-2026-0003", "main", "CHG-2026-0003 - GitOps change for demo-go-color-app", "description")
	if err != nil {
		t.Fatalf("OpenMergeRequest returned error: %v", err)
	}
	if mr.IID != 1 {
		t.Fatalf("mr.IID = %d, want 1", mr.IID)
	}
	if mr.State != "opened" {
		t.Fatalf("mr.State = %q, want opened", mr.State)
	}
	if mr.WebURL == "" {
		t.Fatal("mr.WebURL is empty")
	}
}

func TestOpenMergeRequestValidation(t *testing.T) {
	client, err := New(Config{BaseURL: "https://gitlab.example.local", Token: "token"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	tests := []struct {
		name         string
		projectID    int
		sourceBranch string
		targetBranch string
		title        string
	}{
		{name: "missing project", projectID: 0, sourceBranch: "change/test", targetBranch: "main", title: "title"},
		{name: "missing source", projectID: 1, sourceBranch: "", targetBranch: "main", title: "title"},
		{name: "missing target", projectID: 1, sourceBranch: "change/test", targetBranch: "", title: "title"},
		{name: "missing title", projectID: 1, sourceBranch: "change/test", targetBranch: "main", title: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.OpenMergeRequest(context.Background(), tt.projectID, tt.sourceBranch, tt.targetBranch, tt.title, "description")
			if err == nil {
				t.Fatal("OpenMergeRequest returned nil error, want validation error")
			}
		})
	}
}
