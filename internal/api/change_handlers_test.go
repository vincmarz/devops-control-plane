package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/api"
	"github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/config"
	"github.com/vincmarz/devops-control-plane/internal/domain"
)

// fakeChangeStore is an in-memory app.ChangeStore for exercising the change
// HTTP handlers without a database. Only List and Create are meaningful here;
// the rest satisfy the interface and fail loudly if unexpectedly called.
type fakeChangeStore struct {
	listResult   []domain.ChangeRequest
	listErr      error
	createResult domain.ChangeRequest
	createErr    error
	createdWith  domain.CreateChangeRequest // captured for assertions
}

func (f *fakeChangeStore) List(ctx context.Context) ([]domain.ChangeRequest, error) {
	return f.listResult, f.listErr
}

func (f *fakeChangeStore) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	f.createdWith = req
	if f.createErr != nil {
		return domain.ChangeRequest{}, f.createErr
	}
	res := f.createResult
	res.RequestedBy = req.RequestedBy
	return res, nil
}

func (f *fakeChangeStore) Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, errors.New("not implemented")
}

func (f *fakeChangeStore) Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeChangeStore) TransitionLifecycle(ctx context.Context, idOrNumber, action, actor, message string) (map[string]any, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeChangeStore) MarkStep(ctx context.Context, idOrNumber, status string) (map[string]any, error) {
	return nil, errors.New("not implemented")
}

// newChangesServer builds the real router wired with a ChangeService backed by
// the given fake store.
func newChangesServer(t *testing.T, store app.ChangeStore) *httptest.Server {
	t.Helper()
	deps := api.Dependencies{
		Config: config.Config{},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Services: app.Services{
			Changes: app.NewChangeService(store),
		},
	}
	srv := httptest.NewServer(api.NewRouter(deps))
	t.Cleanup(srv.Close)
	return srv
}

// doRequest issues an HTTP request with optional headers and returns the response.
func doRequest(t *testing.T, method, url, body string, headers map[string]string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	return resp
}

func operatorHeaders() map[string]string {
	return map[string]string{
		"X-Forwarded-User":   "operator@example.test",
		"X-Forwarded-Groups": "cp-operators",
	}
}

func viewerHeaders() map[string]string {
	return map[string]string{
		"X-Forwarded-User":   "viewer@example.test",
		"X-Forwarded-Groups": "cp-viewers",
	}
}

// TestListChangesAuthenticatedViewer verifies routing, viewer authorization and
// that the handler returns the store's data as a JSON array with a count.
func TestListChangesAuthenticatedViewer(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_VIEWER", "cp-viewers")

	store := &fakeChangeStore{listResult: []domain.ChangeRequest{
		{ID: "1", ChangeNumber: "CHG-2026-0001", Title: "first", Status: "draft"},
		{ID: "2", ChangeNumber: "CHG-2026-0002", Title: "second", Status: "submitted"},
	}}
	srv := newChangesServer(t, store)

	resp := doRequest(t, http.MethodGet, srv.URL+"/api/v1/changes", "", viewerHeaders())
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var env struct {
		Data []map[string]any `json:"data"`
		Meta map[string]any   `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if len(env.Data) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(env.Data))
	}
	if count, _ := env.Meta["count"].(float64); int(count) != 2 {
		t.Errorf("expected meta.count 2, got %v", env.Meta["count"])
	}
}

// TestListChangesRequiresAuthentication verifies that, with auth enabled, a
// request without trusted identity headers is rejected with 401.
func TestListChangesRequiresAuthentication(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")

	srv := newChangesServer(t, &fakeChangeStore{})

	resp := doRequest(t, http.MethodGet, srv.URL+"/api/v1/changes", "", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without identity headers, got %d", resp.StatusCode)
	}
}

// TestCreateChangeForbiddenForViewer verifies RBAC at the HTTP boundary: creating
// a change requires operator/admin, so a viewer must receive 403.
func TestCreateChangeForbiddenForViewer(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_VIEWER", "cp-viewers")

	srv := newChangesServer(t, &fakeChangeStore{})

	body := `{"title":"x","applicationName":"app","changeType":"config","targetEnvironment":"dev"}`
	resp := doRequest(t, http.MethodPost, srv.URL+"/api/v1/changes", body, viewerHeaders())
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for a viewer creating a change, got %d", resp.StatusCode)
	}
}

// TestCreateChangeAsOperator verifies the happy path: an operator creates a
// change, the handler stamps the authenticated username onto the request and
// returns 201 with the created resource.
func TestCreateChangeAsOperator(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_OPERATOR", "cp-operators")

	store := &fakeChangeStore{createResult: domain.ChangeRequest{
		ID:           "generated-id",
		ChangeNumber: "CHG-2026-0007",
		Status:       "draft",
	}}
	srv := newChangesServer(t, store)

	body := `{"title":"Bump image","applicationName":"demo-go-color-app","changeType":"image-update","targetEnvironment":"dev"}`
	resp := doRequest(t, http.MethodPost, srv.URL+"/api/v1/changes", body, operatorHeaders())

	if resp.StatusCode != http.StatusCreated {
		resp.Body.Close()
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	env := decodeEnvelope(t, resp)
	if env.Data["status"] != "draft" {
		t.Errorf("expected data.status \"draft\", got %v", env.Data["status"])
	}
	// The handler must stamp the authenticated identity onto the request.
	if store.createdWith.RequestedBy != "operator@example.test" {
		t.Errorf("expected RequestedBy from identity, got %q", store.createdWith.RequestedBy)
	}
}

// TestCreateChangeRejectsApplicationEnvironmentMismatch verifies that a
// ChangeRequest cannot target an environment configured for a different
// logical application. The request must fail before store persistence.
func TestCreateChangeRejectsApplicationEnvironmentMismatch(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_OPERATOR", "cp-operators")

	store := &fakeChangeStore{}
	srv := newChangesServer(t, store)

	body := `{"title":"Reject mismatching application binding","applicationName":"payments","changeType":"config","targetEnvironment":"dev"}`
	resp := doRequest(t, http.MethodPost, srv.URL+"/api/v1/changes", body, operatorHeaders())

	if resp.StatusCode != http.StatusUnprocessableEntity {
		resp.Body.Close()
		t.Fatalf("expected 422 for an application environment mismatch, got %d", resp.StatusCode)
	}

	env := decodeEnvelope(t, resp)
	if env.Error == nil {
		t.Fatal("expected an error object in the envelope")
	}
	if env.Error.Code != "VALIDATION_INVALID_REQUEST" {
		t.Errorf("error code = %q, want VALIDATION_INVALID_REQUEST", env.Error.Code)
	}
	if env.Error.Message != "Unable to create ChangeRequest" {
		t.Errorf("error message = %q, want Unable to create ChangeRequest", env.Error.Message)
	}

	expectedTechnicalMessage := `applicationName "payments" is not configured for targetEnvironment "dev"; expected "demo-go-color-app"`
	if env.Error.TechnicalMessage != expectedTechnicalMessage {
		t.Errorf("technical message = %q, want %q", env.Error.TechnicalMessage, expectedTechnicalMessage)
	}
	if !env.Error.Recoverable {
		t.Error("expected recoverable validation error")
	}

	if store.createdWith.Title != "" || store.createdWith.ApplicationName != "" || store.createdWith.TargetEnvironment != "" {
		t.Fatalf("store.Create was unexpectedly called with %#v", store.createdWith)
	}
}

// TestCreateChangeRejectsMalformedJSON verifies malformed input yields 400.
func TestCreateChangeRejectsMalformedJSON(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_OPERATOR", "cp-operators")

	srv := newChangesServer(t, &fakeChangeStore{})

	resp := doRequest(t, http.MethodPost, srv.URL+"/api/v1/changes", "{not-json", operatorHeaders())
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d", resp.StatusCode)
	}
}

// TestCreateChangeValidationError verifies that domain validation failures
// surface as 422 with an error envelope (here: a missing title).
func TestCreateChangeValidationError(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_OPERATOR", "cp-operators")

	srv := newChangesServer(t, &fakeChangeStore{})

	body := `{"applicationName":"app","changeType":"config","targetEnvironment":"dev"}`
	resp := doRequest(t, http.MethodPost, srv.URL+"/api/v1/changes", body, operatorHeaders())

	if resp.StatusCode != http.StatusUnprocessableEntity {
		resp.Body.Close()
		t.Fatalf("expected 422 for a missing title, got %d", resp.StatusCode)
	}

	env := decodeEnvelope(t, resp)
	if env.Error == nil {
		t.Fatal("expected an error object in the envelope")
	}
}
