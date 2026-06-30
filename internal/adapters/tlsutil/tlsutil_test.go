package tlsutil

import "testing"

func TestTLSConfigFromCAFileRequiresPath(t *testing.T) {
	if _, err := TLSConfigFromCAFile(""); err == nil {
		t.Fatal("expected error for empty CA file path")
	}
}
