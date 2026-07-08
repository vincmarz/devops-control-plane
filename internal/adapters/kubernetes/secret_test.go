package kubernetes

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestSecretClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	client, err := New(Config{APIURL: server.URL, Token: "test-token"}, WithHTTPClient(server.Client()))
	if err != nil {
		server.Close()
		t.Fatalf("New returned error: %v", err)
	}
	return client, server
}

func TestGetSecretDecodesBase64Values(t *testing.T) {
	tokenPlain := "top-secret-token"
	caPlain := "top-secret-ca"
	tokenEncoded := base64.StdEncoding.EncodeToString([]byte(tokenPlain))
	caEncoded := base64.StdEncoding.EncodeToString([]byte(caPlain))

	var capturedPath string
	var capturedAuth string
	client, server := newTestSecretClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]any{
			"kind": "Secret",
			"data": map[string]any{
				"token":  tokenEncoded,
				"ca.crt": caEncoded,
			},
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	data, err := client.GetSecret(context.Background(), "devops-control-plane", "ocp-dev-kube")
	if err != nil {
		t.Fatalf("GetSecret returned error: %v", err)
	}
	if capturedPath != "/api/v1/namespaces/devops-control-plane/secrets/ocp-dev-kube" {
		t.Fatalf("unexpected request path: %q", capturedPath)
	}
	if !strings.HasPrefix(capturedAuth, "Bearer ") {
		t.Fatalf("expected Bearer token in Authorization header, got %q", capturedAuth)
	}
	if got := string(data["token"]); got != tokenPlain {
		t.Fatalf("token = %q, want %q", got, tokenPlain)
	}
	if got := string(data["ca.crt"]); got != caPlain {
		t.Fatalf("ca.crt = %q, want %q", got, caPlain)
	}
}

func TestGetSecretRejectsEmptyNamespace(t *testing.T) {
	client, server := newTestSecretClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called when namespace is empty")
	}))
	defer server.Close()

	if _, err := client.GetSecret(context.Background(), "", "ocp-dev-kube"); err == nil {
		t.Fatal("expected error for empty namespace, got nil")
	}
}

func TestGetSecretRejectsEmptyName(t *testing.T) {
	client, server := newTestSecretClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called when name is empty")
	}))
	defer server.Close()

	if _, err := client.GetSecret(context.Background(), "devops-control-plane", ""); err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestGetSecretReturnsEmptyMapWhenDataFieldIsMissing(t *testing.T) {
	client, server := newTestSecretClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]any{
			"kind": "Secret",
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	data, err := client.GetSecret(context.Background(), "devops-control-plane", "ocp-dev-kube")
	if err != nil {
		t.Fatalf("GetSecret returned error: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected empty data, got %d entries", len(data))
	}
}

func TestGetSecretReturnsErrorWhenValueIsNotBase64(t *testing.T) {
	client, server := newTestSecretClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]any{
			"kind": "Secret",
			"data": map[string]any{
				"token": "not-base64!!!",
			},
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	if _, err := client.GetSecret(context.Background(), "devops-control-plane", "ocp-dev-kube"); err == nil {
		t.Fatal("expected base64 decode error, got nil")
	}
}

func TestGetSecretPropagatesServerError(t *testing.T) {
	client, server := newTestSecretClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"kind":"Status","code":404,"reason":"NotFound"}`))
	}))
	defer server.Close()

	if _, err := client.GetSecret(context.Background(), "devops-control-plane", "missing"); err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}
