package api

import "net/http"

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "devops-control-plane",
	}, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) readyz(w http.ResponseWriter, r *http.Request) {
	checks := map[string]string{
		"configuration": "ok",
		"database":      "ok",
	}

	if err := h.deps.Services.DB.Ping(r.Context()); err != nil {
		checks["database"] = "not-configured"
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not-ready", "checks": checks}, map[string]any{"requestId": requestIDFromContext(r.Context())})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "checks": checks}, map[string]any{"requestId": requestIDFromContext(r.Context())})
}
