package domain

import "time"

const (
	ChangeStatusDraft     = "draft"
	ChangeStatusSubmitted = "submitted"
	ChangeStatusApproved  = "approved"
	ChangeStatusRejected  = "rejected"
	ChangeStatusExecuting = "executing"
	ChangeStatusExecuted  = "executed"
	ChangeStatusFailed    = "failed"
	ChangeStatusClosed    = "closed"
	ChangeStatusCancelled = "cancelled"
)

const (
	RiskLow      = "low"
	RiskMedium   = "medium"
	RiskHigh     = "high"
	RiskCritical = "critical"
)

const (
	ChangeEventCreated                = "change_created"
	ChangeEventSubmitted              = "change_submitted"
	ChangeEventApproved               = "change_approved"
	ChangeEventRejected               = "change_rejected"
	ChangeEventExecutionStarted       = "change_execution_started"
	ChangeEventExecutionCompleted     = "change_execution_completed"
	ChangeEventExecutionFailed        = "change_execution_failed"
	ChangeEventClosed                 = "change_closed"
	ChangeEventCancelled              = "change_cancelled"
	ChangeEventTechnicalStepCompleted = "technical_step_completed"
)

type CreateChangeRequest struct {
	Title             string         `json:"title"`
	ApplicationName   string         `json:"applicationName"`
	TargetEnvironment string         `json:"targetEnvironment"`
	ChangeType        string         `json:"changeType"`
	RiskLevel         string         `json:"riskLevel"`
	RequestedBy       string         `json:"requestedBy"`
	Description       string         `json:"description"`
	Payload           map[string]any `json:"payload"`
}

type ChangeRequest struct {
	ID                string         `json:"id"`
	ChangeNumber      string         `json:"changeNumber"`
	Title             string         `json:"title"`
	ApplicationName   string         `json:"applicationName"`
	TargetEnvironment string         `json:"targetEnvironment"`
	ChangeType        string         `json:"changeType"`
	Status            string         `json:"status"`
	RiskLevel         string         `json:"riskLevel"`
	RequestedBy       string         `json:"requestedBy,omitempty"`
	Description       string         `json:"description,omitempty"`
	Payload           map[string]any `json:"payload,omitempty"`

	Git     map[string]any `json:"git,omitempty"`
	Tekton  map[string]any `json:"tekton,omitempty"`
	ArgoCD  map[string]any `json:"argocd,omitempty"`
	Runtime map[string]any `json:"runtime,omitempty"`

	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	SubmittedAt *time.Time `json:"submittedAt,omitempty"`
	ApprovedAt  *time.Time `json:"approvedAt,omitempty"`
	RejectedAt  *time.Time `json:"rejectedAt,omitempty"`
	ExecutedAt  *time.Time `json:"executedAt,omitempty"`
	ClosedAt    *time.Time `json:"closedAt,omitempty"`
	CancelledAt *time.Time `json:"cancelledAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
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
