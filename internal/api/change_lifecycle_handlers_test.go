package api_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

// transitionStore is an app.ChangeStore focused on lifecycle transitions: it
// records the arguments it receives and returns a configurable result or error.
// The remaining methods fail loudly if unexpectedly called.
type transitionStore struct {
	result map[string]any
	err    error

	gotID      string
	gotAction  string
	gotActor   string
	gotMessage string
}

func (s *transitionStore) TransitionLifecycle(ctx context.Context, idOrNumber, action, actor, message string) (map[string]any, error) {
	s.gotID, s.gotAction, s.gotActor, s.gotMessage = idOrNumber, action, actor, message
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

func (s *transitionStore) List(ctx context.Context) ([]domain.ChangeRequest, error) {
	return nil, errors.New("not implemented")
}

func (s *transitionStore) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, errors.New("not implemented")
}

func (s *transitionStore) Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, errors.New("not implemented")
}

func (s *transitionStore) Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error) {
	return nil, errors.New("not implemented")
}

func (s *transitionStore) MarkStep(ctx context.Context, idOrNumber, status string) (map[string]any, error) {
	return nil, errors.New("not implemented")
}

func approverHeaders() map[string]string {
	return map[string]string{
		"X-Forwarded-User":   "approver@example.test",
		"X-Forwarded-Groups": "cp-approvers",
	}
}

// TestApproveChangeAsApprover verifies the happy path of a lifecycle transition:
// an approver approves a change, the handler forwards action/actor/id to the
// store and returns 202 with the store's result.
func TestApproveChangeAsApprover(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_APPROVER", "cp-approvers")

	store := &transitionStore{result: map[string]any{"id": "1", "status": "approved"}}
	srv := newChangesServer(t, store)

	body := `{"actor":"jane","message":"looks good"}`
	resp := doRequest(t, http.MethodPost, srv.URL+"/api/v1/changes/1/approve", body, approverHeaders())

	if resp.StatusCode != http.StatusAccepted {
		resp.Body.Close()
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}

	env := decodeEnvelope(t, resp)
	if env.Data["status"] != "approved" {
		t.Errorf("expected data.status \"approved\", got %v", env.Data["status"])
	}
	if store.gotAction != "approve" {
		t.Errorf("expected store action \"approve\", got %q", store.gotAction)
	}
	if store.gotActor != "jane" {
		t.Errorf("expected actor \"jane\", got %q", store.gotActor)
	}
	if store.gotID != "1" {
		t.Errorf("expected id \"1\", got %q", store.gotID)
	}
}

// TestApproveChangeForbiddenForViewer verifies RBAC: approving requires
// approver/admin, so a viewer must receive 403 and never reach the store.
func TestApproveChangeForbiddenForViewer(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_VIEWER", "cp-viewers")

	store := &transitionStore{}
	srv := newChangesServer(t, store)

	resp := doRequest(t, http.MethodPost, srv.URL+"/api/v1/changes/1/approve", `{"actor":"bob"}`, viewerHeaders())
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for a viewer approving a change, got %d", resp.StatusCode)
	}
	if store.gotAction != "" {
		t.Errorf("store must not be reached on a forbidden request, got action %q", store.gotAction)
	}
}

// TestApproveChangeInvalidTransition verifies that a rejected transition from
// the store surfaces as 422 with an error envelope.
func TestApproveChangeInvalidTransition(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_APPROVER", "cp-approvers")

	store := &transitionStore{err: errors.New(`invalid lifecycle transition "approve" from status "draft"`)}
	srv := newChangesServer(t, store)

	resp := doRequest(t, http.MethodPost, srv.URL+"/api/v1/changes/1/approve", `{"actor":"jane"}`, approverHeaders())

	if resp.StatusCode != http.StatusUnprocessableEntity {
		resp.Body.Close()
		t.Fatalf("expected 422 for an invalid transition, got %d", resp.StatusCode)
	}

	env := decodeEnvelope(t, resp)
	if env.Error == nil {
		t.Fatal("expected an error object in the envelope")
	}
}

// TestSubmitChangeMissingActor verifies service-level validation at the HTTP
// boundary: an empty body means no actor, which must be rejected with 422
// before the store is ever called.
func TestSubmitChangeMissingActor(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_APPROVER", "cp-approvers")

	store := &transitionStore{result: map[string]any{"status": "submitted"}}
	srv := newChangesServer(t, store)

	resp := doRequest(t, http.MethodPost, srv.URL+"/api/v1/changes/1/submit", "", approverHeaders())
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for a missing actor, got %d", resp.StatusCode)
	}
	if store.gotAction != "" {
		t.Errorf("store must not be called when actor is missing, got action %q", store.gotAction)
	}
}
