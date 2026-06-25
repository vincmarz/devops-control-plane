package domain

import "time"

type CreateChangeRequest struct {
	ApplicationName string         `json:"applicationName"`
	ChangeType      string         `json:"changeType"`
	RequestedBy     string         `json:"requestedBy"`
	Description     string         `json:"description"`
	Payload         map[string]any `json:"payload"`
}

type ChangeRequest struct {
	ID              string         `json:"id"`
	ChangeNumber    string         `json:"changeNumber"`
	ApplicationName string         `json:"applicationName"`
	ChangeType      string         `json:"changeType"`
	Status          string         `json:"status"`
	RequestedBy     string         `json:"requestedBy,omitempty"`
	Description     string         `json:"description,omitempty"`
	Payload         map[string]any `json:"payload,omitempty"`
	Git             map[string]any `json:"git,omitempty"`
	Tekton          map[string]any `json:"tekton,omitempty"`
	ArgoCD          map[string]any `json:"argocd,omitempty"`
	Runtime         map[string]any `json:"runtime,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	CompletedAt     *time.Time     `json:"completedAt,omitempty"`
}

type ChangeEvent struct {
	EventType      string         `json:"eventType"`
	PreviousStatus string         `json:"previousStatus,omitempty"`
	NewStatus      string         `json:"newStatus,omitempty"`
	Message        string         `json:"message,omitempty"`
	ErrorCode      string         `json:"errorCode,omitempty"`
	Source         string         `json:"source,omitempty"`
	Payload        map[string]any `json:"payload,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
}
