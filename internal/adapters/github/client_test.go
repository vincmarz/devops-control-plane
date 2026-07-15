package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := New(Config{APIURL: server.URL, Token: "token", TimeoutSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestParseProjectPath(t *testing.T) {
	owner, repository, err := ParseProjectPath(" owner/repository ")
	if err != nil || owner != "owner" || repository != "repository" {
		t.Fatalf("owner=%q repository=%q err=%v", owner, repository, err)
	}
	for _, value := range []string{"", "owner", "owner/repository/extra", "/repository"} {
		if _, _, err := ParseProjectPath(value); err == nil {
			t.Fatalf("ParseProjectPath(%q) returned nil error", value)
		}
	}
}

func TestCreateBranchUsesBaseReferenceSHA(t *testing.T) {
	calls := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatal("missing bearer token")
		}
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			if r.Method != http.MethodGet || r.URL.Path != "/repos/org/repo/git/ref/heads/main" {
				t.Fatalf("first request: %s %s", r.Method, r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"ref":"refs/heads/main","object":{"sha":"abc123"}}`))
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != "/repos/org/repo/git/refs" {
			t.Fatalf("second request: %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]string
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload["ref"] != "refs/heads/change/CHG-1" || payload["sha"] != "abc123" {
			t.Fatalf("payload=%#v", payload)
		}
		_, _ = w.Write([]byte(`{"ref":"refs/heads/change/CHG-1","object":{"sha":"abc123"}}`))
	})
	if _, err := client.CreateBranch(context.Background(), "org/repo", "change/CHG-1", "main"); err != nil {
		t.Fatal(err)
	}
}

func TestCreateOrUpdateFileCreatesWhenMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var payload map[string]string
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload["sha"] != "" || payload["branch"] != "change/CHG-1" || payload["content"] == "" {
			t.Fatalf("payload=%#v", payload)
		}
		_, _ = w.Write([]byte(`{"content":{"sha":"newsha"}}`))
	})
	result, err := client.CreateOrUpdateFile(context.Background(), "org/repo", "change/CHG-1", "manifests/change.yaml", "update", "content")
	if err != nil || result.SHA != "newsha" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestPullRequestWorkflow(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"number":7,"html_url":"https://github.example/org/repo/pull/7"}`))
		case r.Method == http.MethodGet:
			if r.URL.Query().Get("state") != "open" || r.URL.Query().Get("head") != "org:change/CHG-1" || r.URL.Query().Get("base") != "main" {
				t.Fatalf("query=%s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`[{"number":7,"html_url":"https://github.example/org/repo/pull/7"}]`))
		case r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"sha":"mergedsha","merged":true,"message":"merged"}`))
		}
	})
	opened, err := client.OpenPullRequest(context.Background(), "org/repo", "change/CHG-1", "main", "title", "body")
	if err != nil || opened.Number != 7 {
		t.Fatalf("opened=%#v err=%v", opened, err)
	}
	found, err := client.FindOpenPullRequest(context.Background(), "org/repo", "change/CHG-1", "main")
	if err != nil || found.Number != 7 {
		t.Fatalf("found=%#v err=%v", found, err)
	}
	merged, err := client.MergePullRequest(context.Background(), "org/repo", found.Number, "merge")
	if err != nil || merged.SHA != "mergedsha" {
		t.Fatalf("merged=%#v err=%v", merged, err)
	}
}

func TestNewValidation(t *testing.T) {
	tests := []Config{{Token: "token"}, {APIURL: "https://api.github.com"}, {APIURL: "not-a-url", Token: "token"}}
	for _, cfg := range tests {
		if _, err := New(cfg); err == nil {
			t.Fatalf("New(%#v) returned nil error", cfg)
		}
	}
	client, err := New(Config{APIURL: "https://api.github.com/", Token: " token "})
	if err != nil || client.apiURL != "https://api.github.com" || strings.TrimSpace(client.token) != "token" {
		t.Fatalf("client=%#v err=%v", client, err)
	}
}
