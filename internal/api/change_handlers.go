package api

import (
	"encoding/json"
	"net/http"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func (h *Handler) listChanges(w http.ResponseWriter, r *http.Request) {
	changes := h.deps.Services.Changes.List(r.Context())
	writeJSON(w, http.StatusOK, changes, map[string]any{"count": len(changes), "requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) createChange(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "VALIDATION_INVALID_REQUEST", Message: "Invalid JSON request", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}

	change, err := h.deps.Services.Changes.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, APIError{Code: "VALIDATION_INVALID_REQUEST", Message: "Unable to create ChangeRequest", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusCreated, change, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) getChange(w http.ResponseWriter, r *http.Request, id string) {
	change, err := h.deps.Services.Changes.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, APIError{Code: "CHANGE_NOT_FOUND", Message: "ChangeRequest not found", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusOK, change, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) getChangeEvents(w http.ResponseWriter, r *http.Request, id string) {
	events := h.deps.Services.Changes.Events(r.Context(), id)
	writeJSON(w, http.StatusOK, events, map[string]any{"count": len(events), "requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) getChangeEvidence(w http.ResponseWriter, r *http.Request, id, evidenceType string) {
	evidence := h.deps.Services.Evidence.List(r.Context(), id, evidenceType)
	writeJSON(w, http.StatusOK, evidence, map[string]any{"count": len(evidence), "requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) createBranch(w http.ResponseWriter, r *http.Request, id string) {
	writeJSON(w, http.StatusAccepted, h.deps.Services.Changes.MarkStep(id, "BranchCreated"), map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) updateFiles(w http.ResponseWriter, r *http.Request, id string) {
	writeJSON(w, http.StatusAccepted, h.deps.Services.Changes.MarkStep(id, "CommitCreated"), map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) validateChange(w http.ResponseWriter, r *http.Request, id string) {
	writeJSON(w, http.StatusAccepted, h.deps.Services.Changes.MarkStep(id, "ValidationRunning"), map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) openMergeRequest(w http.ResponseWriter, r *http.Request, id string) {
	writeJSON(w, http.StatusAccepted, h.deps.Services.Changes.MarkStep(id, "MergeRequestOpened"), map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) syncChange(w http.ResponseWriter, r *http.Request, id string) {
	writeJSON(w, http.StatusAccepted, h.deps.Services.Changes.MarkStep(id, "SyncRunning"), map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) collectEvidence(w http.ResponseWriter, r *http.Request, id string) {
	writeJSON(w, http.StatusAccepted, h.deps.Services.Changes.MarkStep(id, "EvidenceCollected"), map[string]any{"requestId": requestIDFromContext(r.Context())})
}
