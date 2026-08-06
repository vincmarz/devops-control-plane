package gitlab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestMergeMergeRequestOmitsShaWhenEmpty proves that when the provider passes an
// empty sha, the adapter does NOT send the sha form field. This is the fix for
// the GitLab 405 Method Not Allowed observed when a stale sha is sent: GitLab
// treats sha as an optimistic lock and rejects the merge if HEAD has moved.
func TestMergeMergeRequestOmitsShaWhenEmpty(t *testing.T) {
	var gotForm url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method=%s", r.Method)
		}
		if r.URL.Path != "/api/v4/projects/1/merge_requests/7/merge" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		gotForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"iid": 7, "web_url": "https://gitlab.example/mr/7", "merge_commit_sha": "merge-sha"})
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, Token: "token"}, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatal(err)
	}
	merged, err := client.MergeMergeRequest(context.Background(), 1, 7, "", "merge message")
	if err != nil {
		t.Fatal(err)
	}
	if merged.MergeCommitSHA != "merge-sha" {
		t.Fatalf("mergeCommitSHA=%q", merged.MergeCommitSHA)
	}
	if _, present := gotForm["sha"]; present {
		t.Fatalf("sha field must be omitted when empty, got form: %v", gotForm)
	}
	if gotForm.Get("merge_commit_message") != "merge message" {
		t.Fatalf("merge_commit_message=%q", gotForm.Get("merge_commit_message"))
	}
}

// TestMergeMergeRequestSendsShaWhenProvided keeps the optimistic-lock behavior
// available for callers that still choose to pass a sha explicitly.
func TestMergeMergeRequestSendsShaWhenProvided(t *testing.T) {
	var gotForm url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"iid": 7, "merge_commit_sha": "merge-sha"})
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, Token: "token"}, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.MergeMergeRequest(context.Background(), 1, 7, "abc123", "merge message"); err != nil {
		t.Fatal(err)
	}
	if gotForm.Get("sha") != "abc123" {
		t.Fatalf("sha=%q, want abc123", gotForm.Get("sha"))
	}
}
