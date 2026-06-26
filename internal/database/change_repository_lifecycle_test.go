package database

import (
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func TestLifecycleTransitionValidTransitions(t *testing.T) {
	tests := []struct {
		name          string
		action        string
		currentStatus string
		wantStatus    string
		wantEvent     string
		wantColumn    string
		wantCompleted bool
	}{
		{"submit draft to submitted", "submit", domain.ChangeStatusDraft, domain.ChangeStatusSubmitted, domain.ChangeEventSubmitted, "submitted_at", false},
		{"approve submitted to approved", "approve", domain.ChangeStatusSubmitted, domain.ChangeStatusApproved, domain.ChangeEventApproved, "approved_at", false},
		{"reject submitted to rejected", "reject", domain.ChangeStatusSubmitted, domain.ChangeStatusRejected, domain.ChangeEventRejected, "rejected_at", true},
		{"start execution approved to executing", "start-execution", domain.ChangeStatusApproved, domain.ChangeStatusExecuting, domain.ChangeEventExecutionStarted, "", false},
		{"complete execution executing to executed", "complete-execution", domain.ChangeStatusExecuting, domain.ChangeStatusExecuted, domain.ChangeEventExecutionCompleted, "executed_at", false},
		{"fail execution executing to failed", "fail-execution", domain.ChangeStatusExecuting, domain.ChangeStatusFailed, domain.ChangeEventExecutionFailed, "executed_at", true},
		{"close executed to closed", "close", domain.ChangeStatusExecuted, domain.ChangeStatusClosed, domain.ChangeEventClosed, "closed_at", true},
		{"close failed to closed", "close", domain.ChangeStatusFailed, domain.ChangeStatusClosed, domain.ChangeEventClosed, "closed_at", true},
		{"cancel draft to cancelled", "cancel", domain.ChangeStatusDraft, domain.ChangeStatusCancelled, domain.ChangeEventCancelled, "cancelled_at", true},
		{"cancel submitted to cancelled", "cancel", domain.ChangeStatusSubmitted, domain.ChangeStatusCancelled, domain.ChangeEventCancelled, "cancelled_at", true},
		{"cancel approved to cancelled", "cancel", domain.ChangeStatusApproved, domain.ChangeStatusCancelled, domain.ChangeEventCancelled, "cancelled_at", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotEvent, gotColumn, gotCompleted, err := lifecycleTransition(tt.action, tt.currentStatus)
			if err != nil {
				t.Fatalf("lifecycleTransition returned unexpected error: %v", err)
			}
			if gotStatus != tt.wantStatus {
				t.Fatalf("status = %q, want %q", gotStatus, tt.wantStatus)
			}
			if gotEvent != tt.wantEvent {
				t.Fatalf("event = %q, want %q", gotEvent, tt.wantEvent)
			}
			if gotColumn != tt.wantColumn {
				t.Fatalf("timestamp column = %q, want %q", gotColumn, tt.wantColumn)
			}
			if gotCompleted != tt.wantCompleted {
				t.Fatalf("markCompleted = %v, want %v", gotCompleted, tt.wantCompleted)
			}
		})
	}
}

func TestLifecycleTransitionInvalidTransitions(t *testing.T) {
	tests := []struct {
		name          string
		action        string
		currentStatus string
	}{
		{"submit from submitted is invalid", "submit", domain.ChangeStatusSubmitted},
		{"approve from draft is invalid", "approve", domain.ChangeStatusDraft},
		{"reject from approved is invalid", "reject", domain.ChangeStatusApproved},
		{"start execution from submitted is invalid", "start-execution", domain.ChangeStatusSubmitted},
		{"complete execution from approved is invalid", "complete-execution", domain.ChangeStatusApproved},
		{"fail execution from approved is invalid", "fail-execution", domain.ChangeStatusApproved},
		{"close from approved is invalid", "close", domain.ChangeStatusApproved},
		{"cancel from executing is invalid", "cancel", domain.ChangeStatusExecuting},
		{"unknown action is invalid", "unknown", domain.ChangeStatusDraft},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, _, err := lifecycleTransition(tt.action, tt.currentStatus)
			if err == nil {
				t.Fatalf("lifecycleTransition(%q, %q) returned nil error, want error", tt.action, tt.currentStatus)
			}
		})
	}
}

func TestCompletedAtClause(t *testing.T) {
	if got := completedAtClause(false); got != "" {
		t.Fatalf("completedAtClause(false) = %q, want empty string", got)
	}
	if got := completedAtClause(true); got != ", completed_at = now()" {
		t.Fatalf("completedAtClause(true) = %q, want completed_at assignment", got)
	}
}
