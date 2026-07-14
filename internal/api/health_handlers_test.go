package api_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/api"
	"github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/config"
	"github.com/vincmarz/devops-control-plane/internal/database"
)

// stubPinger is a fake database.Pinger used to drive the readiness endpoint
// without a real database connection. A non-nil err simulates a DB outage.
type stubPinger struct{ err error }

func (s stubPinger) Ping(ctx context.Context) error { return s.err }

// responseEnvelope mirrors api.Response for decoding in the external test
// package (which cannot rely on unexported details).
type responseEnvelope struct {
	Data  map[string]any `json:"data"`
	Meta  map[string]any `json:"meta"`
	Error *struct {
		Code             string `json:"code"`
		Message          string `json:"message"`
		TechnicalMessage string `json:"technicalMessage"`
		Recoverable      bool   `json:"recoverable"`
	} `json:"error"`
}

// newTestServer builds the real HTTP router with in-memory dependencies and
// returns a running httptest server. Only DB is wired: the endpoints under test
// touch nothing else, so the remaining services intentionally stay nil.
func newTestServer(t *testing.T, pinger database.Pinger) *httptest.Server {
	t.Helper()
	deps := api.Dependencies{
		Config: config.Config{},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Services: app.Services{
			DB: pinger,
		},
	}
	srv := httptest.NewServer(api.NewRouter(deps))
	t.Cleanup(srv.Close)
	return srv
}

func decodeEnvelope(t *testing.T, resp *http.Response) responseEnvelope {
	t.Helper()
	defer resp.Body.Close()
	var env responseEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return env
}

// TestHealthzEndpoint exercises the full middleware chain (request-id, logging,
// auth bypass for public endpoints) end to end and asserts the JSON envelope.
func TestHealthzEndpoint(t *testing.T) {
	srv := newTestServer(t, stubPinger{})

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	env := decodeEnvelope(t, resp)
	if env.Error != nil {
		t.Errorf("expected no error object, got %+v", env.Error)
	}
	if env.Data["status"] != "ok" {
		t.Errorf("expected data.status \"ok\", got %v", env.Data["status"])
	}
	if env.Data["service"] != "devops-control-plane" {
		t.Errorf("expected data.service \"devops-control-plane\", got %v", env.Data["service"])
	}
	if id, _ := env.Meta["requestId"].(string); id == "" {
		t.Error("expected a non-empty meta.requestId")
	}
}

// TestReadyzWhenDatabaseHealthy asserts the readiness contract used by the
// Kubernetes readiness probe: a successful DB ping yields 200 / ready.
func TestReadyzWhenDatabaseHealthy(t *testing.T) {
	srv := newTestServer(t, stubPinger{err: nil})

	resp, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	env := decodeEnvelope(t, resp)
	if env.Data["status"] != "ready" {
		t.Errorf("expected data.status \"ready\", got %v", env.Data["status"])
	}
	checks, _ := env.Data["checks"].(map[string]any)
	if checks["database"] != "ok" {
		t.Errorf("expected checks.database \"ok\", got %v", checks["database"])
	}
}

// TestReadyzWhenDatabaseDown asserts the failure side of the readiness
// contract: a failing DB ping must yield 503 so the pod is taken out of
// rotation instead of receiving traffic it cannot serve.
func TestReadyzWhenDatabaseDown(t *testing.T) {
	srv := newTestServer(t, stubPinger{err: context.DeadlineExceeded})

	resp, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", resp.StatusCode)
	}

	env := decodeEnvelope(t, resp)
	if env.Data["status"] != "not-ready" {
		t.Errorf("expected data.status \"not-ready\", got %v", env.Data["status"])
	}
	checks, _ := env.Data["checks"].(map[string]any)
	if checks["database"] != "not-configured" {
		t.Errorf("expected checks.database \"not-configured\", got %v", checks["database"])
	}
}
