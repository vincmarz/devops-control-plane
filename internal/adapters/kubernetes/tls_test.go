package kubernetes

import (
	"crypto/tls"
	"net/http"
	"testing"
)

// transportTLSConfig returns the *tls.Config the client's transport would use,
// or nil when it relies on Go's secure defaults.
func transportTLSConfig(t *testing.T, c *Client) *tls.Config {
	t.Helper()
	tr, ok := c.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", c.httpClient.Transport)
	}
	return tr.TLSClientConfig
}

// TestNewVerifiesTLSByDefault encodes the security invariant: without an
// explicit InsecureTLS opt-in the client must not disable certificate
// verification (a nil TLSClientConfig means Go's secure defaults apply).
func TestNewVerifiesTLSByDefault(t *testing.T) {
	c, err := New(Config{APIURL: "https://k8s.example.test", Token: "test-token"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if cfg := transportTLSConfig(t, c); cfg != nil && cfg.InsecureSkipVerify {
		t.Error("default client must verify TLS certificates (InsecureSkipVerify must be false)")
	}
}

// TestNewDisablesTLSVerificationOnlyWhenRequested verifies the opt-in path:
// InsecureTLS=true is the only way to turn verification off.
func TestNewDisablesTLSVerificationOnlyWhenRequested(t *testing.T) {
	c, err := New(Config{APIURL: "https://k8s.example.test", Token: "test-token", InsecureTLS: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	cfg := transportTLSConfig(t, c)
	if cfg == nil || !cfg.InsecureSkipVerify {
		t.Error("InsecureTLS=true must set InsecureSkipVerify=true on the transport")
	}
}