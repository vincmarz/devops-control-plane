package domain

import "time"

type Evidence struct {
	ID           string         `json:"id"`
	ChangeNumber string         `json:"changeNumber,omitempty"`
	EvidenceType string         `json:"evidenceType"`
	Name         string         `json:"name,omitempty"`
	Summary      string         `json:"summary,omitempty"`
	Content      string         `json:"content,omitempty"`
	Payload      map[string]any `json:"payload,omitempty"`
	ExternalRef  string         `json:"externalRef,omitempty"`
	Sanitized    bool           `json:"sanitized"`
	CreatedAt    time.Time      `json:"createdAt"`
}
