package gitlab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestWaitUntilMergeablePollsUntilMergeable reproduces the real GitLab race:
// the merge request is "checking" right after creation and only later becomes
// "mergeable". WaitUntilMergeable must keep polling until it is mergeable.
func TestWaitUntilMergeablePollsUntilMergeable(t *testing.T) {
	restore := shrinkMergePolling(t, time.Millisecond, 2*time.Second)
	defer restore()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v4/projects/1/merge_requests/7" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		status := "checking"
		if atomic.AddInt32(&calls, 1) >= 3 {
			status = "mergeable"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"iid":                   7,
			"merge_status":          "checking",
			"detailed_merge_status": status,
		})
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, Token: "token"})
	if err != nil {
		t.Fatal(err)
	}

	mergeRequest, err := client.WaitUntilMergeable(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("WaitUntilMergeable returned error: %v", err)
	}
	if mergeRequest.DetailedMergeStatus != "mergeable" {
		t.Fatalf("detailed_merge_status = %q, want mergeable", mergeRequest.DetailedMergeStatus)
	}
	if got := atomic.LoadInt32(&calls); got < 3 {
		t.Fatalf("expected at least 3 polls, got %d", got)
	}
}

// TestWaitUntilMergeableReturnsImmediatelyWhenMergeable verifies the fast path:
// a single GET is enough when the MR is already mergeable.
func TestWaitUntilMergeableReturnsImmediatelyWhenMergeable(t *testing.T) {
	restore := shrinkMergePolling(t, time.Millisecond, 2*time.Second)
	defer restore()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"iid":                   7,
			"detailed_merge_status": "mergeable",
		})
	}))
	defer server.Close()

	client, _ := New(Config{BaseURL: server.URL, Token: "token"})
	if _, err := client.WaitUntilMergeable(context.Background(), 1, 7); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 poll, got %d", got)
	}
}

// TestWaitUntilMergeableTimesOut ensures the function is fail-closed when the
// merge request never leaves the checking state.
func TestWaitUntilMergeableTimesOut(t *testing.T) {
	restore := shrinkMergePolling(t, time.Millisecond, 20*time.Millisecond)
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"iid":                   7,
			"detailed_merge_status": "checking",
		})
	}))
	defer server.Close()

	client, _ := New(Config{BaseURL: server.URL, Token: "token"})
	_, err := client.WaitUntilMergeable(context.Background(), 1, 7)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

// TestWaitUntilMergeableFailsOnNonMergeable ensures a clearly blocking status is
// reported immediately instead of proceeding to a merge GitLab would reject.
func TestWaitUntilMergeableFailsOnNonMergeable(t *testing.T) {
	restore := shrinkMergePolling(t, time.Millisecond, 2*time.Second)
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"iid":                   7,
			"detailed_merge_status": "conflict",
		})
	}))
	defer server.Close()

	client, _ := New(Config{BaseURL: server.URL, Token: "token"})
	_, err := client.WaitUntilMergeable(context.Background(), 1, 7)
	if err == nil || !strings.Contains(err.Error(), "not mergeable") {
		t.Fatalf("expected not mergeable error, got %v", err)
	}
}

// TestWaitUntilMergeableFallsBackToLegacyMergeStatus covers older GitLab
// versions that only populate merge_status.
func TestWaitUntilMergeableFallsBackToLegacyMergeStatus(t *testing.T) {
	restore := shrinkMergePolling(t, time.Millisecond, 2*time.Second)
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"iid":          7,
			"merge_status": "can_be_merged",
		})
	}))
	defer server.Close()

	client, _ := New(Config{BaseURL: server.URL, Token: "token"})
	if _, err := client.WaitUntilMergeable(context.Background(), 1, 7); err != nil {
		t.Fatalf("expected legacy merge_status to be accepted, got %v", err)
	}
}

// shrinkMergePolling temporarily reduces the polling knobs for fast tests and
// returns a restore function.
func shrinkMergePolling(t *testing.T, interval, timeout time.Duration) func() {
	t.Helper()
	oldInterval, oldTimeout := mergePollInterval, mergePollTimeout
	mergePollInterval, mergePollTimeout = interval, timeout
	return func() { mergePollInterval, mergePollTimeout = oldInterval, oldTimeout }
}
