package api

import (
	"encoding/json"
	"errors"
	"io"
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
	evidence, err := h.deps.Services.Evidence.List(r.Context(), id, evidenceType)
	if err != nil {
		writeError(w, http.StatusNotFound, APIError{Code: "EVIDENCE_NOT_FOUND", Message: "Unable to load ChangeRequest evidence", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusOK, evidence, map[string]any{"count": len(evidence), "requestId": requestIDFromContext(r.Context())})
}

type lifecycleTransitionRequest struct {
	Actor   string `json:"actor"`
	Message string `json:"message"`
}

func (h *Handler) submitChange(w http.ResponseWriter, r *http.Request, id string) {
	h.transitionLifecycle(w, r, id, "submit")
}

func (h *Handler) approveChange(w http.ResponseWriter, r *http.Request, id string) {
	h.transitionLifecycle(w, r, id, "approve")
}

func (h *Handler) rejectChange(w http.ResponseWriter, r *http.Request, id string) {
	h.transitionLifecycle(w, r, id, "reject")
}

func (h *Handler) startExecution(w http.ResponseWriter, r *http.Request, id string) {
	h.transitionLifecycle(w, r, id, "start-execution")
}

func (h *Handler) completeExecution(w http.ResponseWriter, r *http.Request, id string) {
	h.transitionLifecycle(w, r, id, "complete-execution")
}

func (h *Handler) failExecution(w http.ResponseWriter, r *http.Request, id string) {
	h.transitionLifecycle(w, r, id, "fail-execution")
}

func (h *Handler) closeChange(w http.ResponseWriter, r *http.Request, id string) {
	h.transitionLifecycle(w, r, id, "close")
}

func (h *Handler) cancelChange(w http.ResponseWriter, r *http.Request, id string) {
	h.transitionLifecycle(w, r, id, "cancel")
}

func (h *Handler) createBranch(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.deps.Services.Changes.CreateBranch(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, APIError{Code: "GITLAB_CREATE_BRANCH_FAILED", Message: "Unable to create GitLab branch for ChangeRequest", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) updateFiles(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.deps.Services.Changes.UpdateFiles(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, APIError{Code: "GITLAB_UPDATE_FILES_FAILED", Message: "Unable to update GitLab files for ChangeRequest", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) validateChange(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.deps.Services.Changes.Validate(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, APIError{Code: "TEKTON_VALIDATE_FAILED", Message: "Unable to start Tekton validation pipeline for ChangeRequest", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) checkValidation(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.deps.Services.Changes.CheckValidation(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, APIError{Code: "TEKTON_CHECK_VALIDATION_FAILED", Message: "Unable to check Tekton validation pipeline for ChangeRequest", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) checkDeployment(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.deps.Services.Changes.CheckDeployment(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, APIError{Code: "ARGOCD_CHECK_DEPLOYMENT_FAILED", Message: "Unable to check Argo CD deployment for ChangeRequest", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) openMergeRequest(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.deps.Services.Changes.OpenMergeRequest(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, APIError{Code: "GITLAB_OPEN_MERGE_REQUEST_FAILED", Message: "Unable to open GitLab merge request for ChangeRequest", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) mergeRequest(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.deps.Services.Changes.MergeRequest(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, APIError{Code: "GITLAB_MERGE_REQUEST_FAILED", Message: "Unable to merge GitLab merge request for ChangeRequest", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) syncChange(w http.ResponseWriter, r *http.Request, id string) {
	h.markWorkflowStep(w, r, id, "SyncRunning")
}

func (h *Handler) collectEvidence(w http.ResponseWriter, r *http.Request, id string) {
	result, err := h.deps.Services.Changes.CollectEvidence(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, APIError{Code: "EVIDENCE_COLLECTION_FAILED", Message: "Unable to collect post-deployment evidence for ChangeRequest", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) transitionLifecycle(w http.ResponseWriter, r *http.Request, id string, action string) {
	var req lifecycleTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, APIError{Code: "VALIDATION_INVALID_REQUEST", Message: "Invalid JSON request", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}

	result, err := h.deps.Services.Changes.TransitionLifecycle(r.Context(), id, action, req.Actor, req.Message)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, APIError{Code: "CHANGE_TRANSITION_INVALID", Message: "Unable to apply ChangeRequest lifecycle transition", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}

	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}

func (h *Handler) markWorkflowStep(w http.ResponseWriter, r *http.Request, id string, status string) {
	result, err := h.deps.Services.Changes.MarkStep(r.Context(), id, status)
	if err != nil {
		writeError(w, http.StatusNotFound, APIError{Code: "CHANGE_NOT_FOUND", Message: "Unable to update ChangeRequest status", TechnicalMessage: err.Error(), Recoverable: true}, nil)
		return
	}
	writeJSON(w, http.StatusAccepted, result, map[string]any{"requestId": requestIDFromContext(r.Context())})
}
