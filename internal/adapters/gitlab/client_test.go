package gitlab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
