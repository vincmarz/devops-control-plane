package api

import "net/http"

func (h *Handler) listApplications(w http.ResponseWriter, r *http.Request) {
	apps, err := h.deps.Services.Applications.List(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, APIError{Code: "ARGO_LIST_FAILED", Message: "Unable to list applications", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusOK, apps, map[string]any{"count": len(apps), "requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) getApplication(w http.ResponseWriter, r *http.Request, name string) {
	app, err := h.deps.Services.Applications.Get(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, APIError{Code: "ARGO_APPLICATION_NOT_FOUND", Message: "Application not found", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusOK, app, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) getApplicationResources(w http.ResponseWriter, r *http.Request, name string) {
	resources := h.deps.Services.Applications.Resources(r.Context(), name)
	writeJSON(w, http.StatusOK, resources, map[string]any{"count": len(resources), "requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) getApplicationHistory(w http.ResponseWriter, r *http.Request, name string) {
	history := h.deps.Services.Applications.History(r.Context(), name)
	writeJSON(w, http.StatusOK, history, map[string]any{"count": len(history), "requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) getApplicationRuntime(w http.ResponseWriter, r *http.Request, name string) {
	runtime := h.deps.Services.Applications.Runtime(r.Context(), name)
	writeJSON(w, http.StatusOK, runtime, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) syncApplication(w http.ResponseWriter, r *http.Request, name string) {
	result := map[string]any{
		"application":  name,
		"status":       "SyncRequested",
		"syncStatus":   "Unknown",
		"healthStatus": "Unknown",
		"message":      "Argo CD adapter not implemented yet",
	}
	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}
