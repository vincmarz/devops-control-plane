package app

import (
	"context"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func TestCheckValidationCreatesValidationEvidenceOnTerminalStatus(t *testing.T) {
	changeStore := &validationEvidenceFakeChangeStore{change: domain.ChangeRequest{ID: "35beb7e0-f8b4-4fd9-b859-0c4e47daecab", ChangeNumber: "CHG-2026-0001", ApplicationName: "demo-go-color-app", TargetEnvironment: "dev", Status: domain.ChangeStatusDraft}}
	evidenceStore := &validationEvidenceFakeEvidenceStore{}

	svc := NewChangeService(
		changeStore,
		WithEvidenceStore(evidenceStore),
		WithTektonCheckValidation(func(ctx context.Context, ch domain.ChangeRequest) (TektonValidationResult, error) {
			return TektonValidationResult{
				PipelineRunName: "devops-cp-validate-chg-2026-0001-abcde",
				Namespace:       "devops-ci-demo",
				UID:             "pipeline-run-uid",
				Status:          "True",
				Reason:          "Succeeded",
				Message:         "Tasks Completed: 2 (Failed: 0, Cancelled 0), Skipped: 0",
				PipelineName:    "validate-gitops",
				GitURL:          "https://github.com/vincmarz/demo-app-gitops.git",
				GitRevision:     "change/CHG-2026-0001",
				ValidationPath:  "apps/demo-go-color-app",
				TaskRuns: []TektonTaskRunResult{
					{Name: "pr-1-clone-repository", Namespace: "devops-ci-demo", PipelineTaskName: "clone-repository", TaskName: "git-clone", Status: "True", Reason: "Succeeded", Message: "clone completed"},
				},
			}, nil
		}),
	)

	result, err := svc.CheckValidation(context.Background(), "CHG-2026-0001")
	if err != nil {
		t.Fatalf("CheckValidation() error = %v", err)
	}
	if changeStore.markedStatus != "ValidationSucceeded" {
		t.Fatalf("marked status=%q, want ValidationSucceeded", changeStore.markedStatus)
	}
	if evidenceStore.created.EvidenceType != "validation" {
		t.Fatalf("evidence type=%q, want validation", evidenceStore.created.EvidenceType)
	}
	if evidenceStore.created.Name != "tekton-validation-evidence" {
		t.Fatalf("evidence name=%q, want tekton-validation-evidence", evidenceStore.created.Name)
	}
	if !evidenceStore.created.Sanitized {
		t.Fatal("validation evidence must be sanitized")
	}
	if evidenceStore.created.ExternalRef != "devops-cp-validate-chg-2026-0001-abcde" {
		t.Fatalf("externalRef=%q, want PipelineRun name", evidenceStore.created.ExternalRef)
	}
	tekton, ok := evidenceStore.created.Payload["tekton"].(map[string]any)
	if !ok {
		t.Fatalf("tekton payload missing or invalid: %#v", evidenceStore.created.Payload)
	}
	if tekton["status"] != "True" || tekton["reason"] != "Succeeded" {
		t.Fatalf("tekton payload=%#v", tekton)
	}
	gitops, ok := evidenceStore.created.Payload["gitops"].(map[string]any)
	if !ok {
		t.Fatalf("gitops payload missing or invalid: %#v", evidenceStore.created.Payload)
	}
	if gitops["pipelineName"] != "validate-gitops" || gitops["revision"] != "change/CHG-2026-0001" {
		t.Fatalf("gitops payload=%#v", gitops)
	}
	taskRuns, ok := evidenceStore.created.Payload["taskRuns"].([]map[string]any)
	if !ok || len(taskRuns) != 1 {
		t.Fatalf("taskRuns payload missing or invalid: %#v", evidenceStore.created.Payload["taskRuns"])
	}
	diagnostics, ok := evidenceStore.created.Payload["diagnostics"].(map[string]any)
	if !ok {
		t.Fatalf("diagnostics payload missing or invalid: %#v", evidenceStore.created.Payload["diagnostics"])
	}
	if diagnostics["failedTaskCount"] != 0 {
		t.Fatalf("diagnostics payload=%#v", diagnostics)
	}
	if result["evidence"] == nil {
		t.Fatalf("result evidence summary missing: %#v", result)
	}
}

func TestCheckValidationDoesNotCreateValidationEvidenceWhileRunning(t *testing.T) {
	changeStore := &validationEvidenceFakeChangeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0001", ApplicationName: "demo-go-color-app", TargetEnvironment: "dev", Status: domain.ChangeStatusDraft}}
	evidenceStore := &validationEvidenceFakeEvidenceStore{}

	svc := NewChangeService(
		changeStore,
		WithEvidenceStore(evidenceStore),
		WithTektonCheckValidation(func(ctx context.Context, ch domain.ChangeRequest) (TektonValidationResult, error) {
			return TektonValidationResult{PipelineRunName: "pr-running", Namespace: "devops-ci-demo", Status: "Unknown", Reason: "Running", Message: "PipelineRun is still running"}, nil
		}),
	)

	_, err := svc.CheckValidation(context.Background(), "CHG-2026-0001")
	if err != nil {
		t.Fatalf("CheckValidation() error = %v", err)
	}
	if changeStore.markedStatus != "ValidationRunning" {
		t.Fatalf("marked status=%q, want ValidationRunning", changeStore.markedStatus)
	}
	if evidenceStore.createCalled {
		t.Fatal("validation evidence must not be created while PipelineRun is still running")
	}
}

type validationEvidenceFakeChangeStore struct {
	change       domain.ChangeRequest
	markedStatus string
}

func (s *validationEvidenceFakeChangeStore) List(ctx context.Context) ([]domain.ChangeRequest, error) {
	return []domain.ChangeRequest{s.change}, nil
}
func (s *validationEvidenceFakeChangeStore) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	return s.change, nil
}
func (s *validationEvidenceFakeChangeStore) Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error) {
	return s.change, nil
}
func (s *validationEvidenceFakeChangeStore) Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error) {
	return nil, nil
}
func (s *validationEvidenceFakeChangeStore) TransitionLifecycle(ctx context.Context, idOrNumber string, action string, actor string, message string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *validationEvidenceFakeChangeStore) MarkStep(ctx context.Context, idOrNumber string, status string) (map[string]any, error) {
	s.markedStatus = status
	return map[string]any{"runtimeStatus": status}, nil
}

type validationEvidenceFakeEvidenceStore struct {
	createCalled bool
	created      domain.Evidence
}

func (s *validationEvidenceFakeEvidenceStore) Create(ctx context.Context, changeID string, evidence domain.Evidence) (domain.Evidence, error) {
	s.createCalled = true
	s.created = evidence
	s.created.ID = "validation-evidence-id"
	return s.created, nil
}
