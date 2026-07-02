package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddlewareDisabledAllowsRequest(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "false")
	h := withAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/changes/CHG-2026-0001/approve", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestAuthMiddlewareRequiresIdentityWhenEnabled(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	h := withAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/changes", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestAuthMiddlewareAllowsViewerReadOnly(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_VIEWER", "devops-cp-viewers")
	h := withAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/changes", nil)
	req.Header.Set("X-Forwarded-User", "viewer@example.test")
	req.Header.Set("X-Forwarded-Groups", "devops-cp-viewers")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestAuthMiddlewareDeniesViewerMutatingAction(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_VIEWER", "devops-cp-viewers")
	h := withAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/changes/CHG-2026-0001/validate", nil)
	req.Header.Set("X-Forwarded-User", "viewer@example.test")
	req.Header.Set("X-Forwarded-Groups", "devops-cp-viewers")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestAuthMiddlewareAllowsOperatorTechnicalAction(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_OPERATOR", "devops-cp-operators")
	h := withAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/changes/CHG-2026-0001/validate", nil)
	req.Header.Set("X-Forwarded-User", "operator@example.test")
	req.Header.Set("X-Forwarded-Groups", "devops-cp-operators")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestAuthMiddlewareAllowsApproverApprovalAction(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_APPROVER", "devops-cp-approvers")
	h := withAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/changes/CHG-2026-0001/approve", nil)
	req.Header.Set("X-Forwarded-User", "approver@example.test")
	req.Header.Set("X-Forwarded-Groups", "devops-cp-approvers")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestAuthMiddlewareAllowsAdminSettings(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_ADMIN", "devops-cp-admins")
	h := withAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodGet, "/ui/settings", nil)
	req.Header.Set("X-Forwarded-User", "admin@example.test")
	req.Header.Set("X-Forwarded-Groups", "devops-cp-admins")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestAuthMiddlewareAllowsPublicHealthEndpointsWhenEnabled(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")

	h := withAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, path := range []string{"/readyz", "/livez"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("expected status %d for %s, got %d", http.StatusNoContent, path, rr.Code)
		}
	}
}

func TestAuthMiddlewareAllowsPublicHealthHeadEndpointsWhenEnabled(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")

	h := withAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, path := range []string{"/readyz", "/livez"} {
		req := httptest.NewRequest(http.MethodHead, path, nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("expected status %d for HEAD %s, got %d", http.StatusNoContent, path, rr.Code)
		}
	}
}
