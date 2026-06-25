package api

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Data  any            `json:"data"`
	Meta  map[string]any `json:"meta"`
	Error *APIError      `json:"error"`
}

type APIError struct {
	Code             string `json:"code"`
	Message          string `json:"message"`
	TechnicalMessage string `json:"technicalMessage,omitempty"`
	SuggestedAction  string `json:"suggestedAction,omitempty"`
	Recoverable      bool   `json:"recoverable"`
}

func writeJSON(w http.ResponseWriter, status int, data any, meta map[string]any) {
	if meta == nil {
		meta = map[string]any{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Response{Data: data, Meta: meta, Error: nil})
}

func writeError(w http.ResponseWriter, status int, err APIError, meta map[string]any) {
	if meta == nil {
		meta = map[string]any{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Response{Data: nil, Meta: meta, Error: &err})
}
