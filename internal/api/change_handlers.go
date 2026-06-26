package api

import (
	"encoding/json"
	"net/http"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func (h *Handler) listChanges(w http.ResponseWriter, r *http.Request) {
	changes, err := h.deps.Services.Changes.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, APIError{Code: "DATABASE_ERROR", Message: "Unable to list ChangeRequests", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
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
	events, err := h.deps.Services.Changes.Events(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, APIError{Code: "CHANGE_EVENTS_NOT_FOUND", Message: "Unable to load ChangeRequest events", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusOK, events, map[string]any{"count": len(events), "requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) getChangeEvidence(w http.ResponseWriter, r *http.Request, id, evidenceType string) {
	evidence := h.deps.Services.Evidence.List(r.Context(), id, evidenceType)
	writeJSON(w, http.StatusOK, evidence, map[string]any{"count": len(evidence), "requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) createBranch(w http.ResponseWriter, r *http.Request, id string) {
	h.markWorkflowStep(w, r, id, "BranchCreated")
}

func (h *Handler) updateFiles(w http.ResponseWriter, r *http.Request, id string) {
	h.markWorkflowStep(w, r, id, "CommitCreated")
}

func (h *Handler) validateChange(w http.ResponseWriter, r *http.Request, id string) {
	h.markWorkflowStep(w, r, id, "ValidationRunning")
}

func (h *Handler) openMergeRequest(w http.ResponseWriter, r *http.Request, id string) {
	h.markWorkflowStep(w, r, id, "MergeRequestOpened")
}

func (h *Handler) syncChange(w http.ResponseWriter, r *http.Request, id string) {
	h.markWorkflowStep(w, r, id, "SyncRunning")
}

func (h *Handler) collectEvidence(w http.ResponseWriter, r *http.Request, id string) {
	h.markWorkflowStep(w, r, id, "EvidenceCollected")
}

func (h *Handler) markWorkflowStep(w http.ResponseWriter, r *http.Request, id string, status string) {
	result, err := h.deps.Services.Changes.MarkStep(r.Context(), id, status)
	if err != nil {
		writeError(w, http.StatusNotFound, APIError{Code: "CHANGE_NOT_FOUND", Message: "Unable to update ChangeRequest status", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}
