package api

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/config"
)

type Dependencies struct {
	Config   config.Config
	Logger   *slog.Logger
	Services app.Services
}

type Handler struct {
	deps Dependencies
}

func NewRouter(deps Dependencies) http.Handler {
	h := &Handler{deps: deps}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /readyz", h.readyz)

	mux.HandleFunc("GET /api/v1/applications", h.listApplications)
	mux.HandleFunc("GET /api/v1/applications/", h.applicationSubrouter)

	mux.HandleFunc("GET /api/v1/changes", h.listChanges)
	mux.HandleFunc("POST /api/v1/changes", h.createChange)
	mux.HandleFunc("GET /api/v1/changes/", h.changeSubrouter)
	mux.HandleFunc("POST /api/v1/changes/", h.changeSubrouter)

	return requestIDMiddleware(loggingMiddleware(deps.Logger, mux))
}

func (h *Handler) applicationSubrouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/applications/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, APIError{Code: "NOT_FOUND", Message: "Application path not found"}, nil)
		return
	}

	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		h.getApplication(w, r, parts[0])
	case len(parts) == 2 && parts[1] == "resources" && r.Method == http.MethodGet:
		h.getApplicationResources(w, r, parts[0])
	case len(parts) == 2 && parts[1] == "history" && r.Method == http.MethodGet:
		h.getApplicationHistory(w, r, parts[0])
	case len(parts) == 2 && parts[1] == "runtime" && r.Method == http.MethodGet:
		h.getApplicationRuntime(w, r, parts[0])
	case len(parts) == 2 && parts[1] == "sync" && r.Method == http.MethodPost:
		h.syncApplication(w, r, parts[0])
	default:
		writeError(w, http.StatusNotFound, APIError{Code: "NOT_FOUND", Message: "Application route not found"}, nil)
	}
}

func (h *Handler) changeSubrouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/changes/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, APIError{Code: "NOT_FOUND", Message: "Change path not found"}, nil)
		return
	}

	id := parts[0]
	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		h.getChange(w, r, id)
	case len(parts) == 2 && parts[1] == "events" && r.Method == http.MethodGet:
		h.getChangeEvents(w, r, id)
	case len(parts) == 2 && parts[1] == "evidence" && r.Method == http.MethodGet:
		h.getChangeEvidence(w, r, id, "")
	case len(parts) == 2 && parts[1] == "submit" && r.Method == http.MethodPost:
		h.submitChange(w, r, id)
	case len(parts) == 2 && parts[1] == "approve" && r.Method == http.MethodPost:
		h.approveChange(w, r, id)
	case len(parts) == 2 && parts[1] == "reject" && r.Method == http.MethodPost:
		h.rejectChange(w, r, id)
	case len(parts) == 2 && parts[1] == "start-execution" && r.Method == http.MethodPost:
		h.startExecution(w, r, id)
	case len(parts) == 2 && parts[1] == "complete-execution" && r.Method == http.MethodPost:
		h.completeExecution(w, r, id)
	case len(parts) == 2 && parts[1] == "fail-execution" && r.Method == http.MethodPost:
		h.failExecution(w, r, id)
	case len(parts) == 2 && parts[1] == "close" && r.Method == http.MethodPost:
		h.closeChange(w, r, id)
	case len(parts) == 2 && parts[1] == "cancel" && r.Method == http.MethodPost:
		h.cancelChange(w, r, id)
	case len(parts) == 3 && parts[1] == "evidence" && r.Method == http.MethodGet:
		h.getChangeEvidence(w, r, id, parts[2])
	case len(parts) == 2 && parts[1] == "validate" && r.Method == http.MethodPost:
		h.validateChange(w, r, id)
	case len(parts) == 2 && parts[1] == "check-validation" && r.Method == http.MethodPost:
		h.checkValidation(w, r, id)
	case len(parts) == 2 && parts[1] == "sync" && r.Method == http.MethodPost:
		h.syncChange(w, r, id)
	case len(parts) == 2 && parts[1] == "collect-evidence" && r.Method == http.MethodPost:
		h.collectEvidence(w, r, id)
	case len(parts) == 2 && parts[1] == "create-branch" && r.Method == http.MethodPost:
		h.createBranch(w, r, id)
	case len(parts) == 2 && parts[1] == "update-files" && r.Method == http.MethodPost:
		h.updateFiles(w, r, id)
	case len(parts) == 2 && parts[1] == "open-merge-request" && r.Method == http.MethodPost:
		h.openMergeRequest(w, r, id)
	case len(parts) == 2 && parts[1] == "merge-request" && r.Method == http.MethodPost:
		h.mergeRequest(w, r, id)
	default:
		writeError(w, http.StatusNotFound, APIError{Code: "NOT_FOUND", Message: "Change route not found"}, nil)
	}
}
