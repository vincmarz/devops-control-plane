package app

import (
	"context"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func TestCollectEvidence(t *testing.T) {
	changeStore := &collectEvidenceFakeChangeStore{change: domain.ChangeRequest{ID: "change-id-1", ChangeNumber: "CHG-2026-0005", ApplicationName: "demo-go-color-app", TargetEnvironment: "dev", Status: "draft"}}
	evidenceStore := &collectEvidenceFakeEvidenceStore{}

	service := NewChangeService(
		changeStore,
		WithEvidenceStore(evidenceStore),
		WithDeploymentEvidenceCollector(func(ctx context.Context, change domain.ChangeRequest) (domain.Evidence, error) {
			return domain.Evidence{EvidenceType: "deployment", Name: "deployment-evidence", Summary: "summary", Sanitized: true, Payload: map[string]any{"applicationName": change.ApplicationName}}, nil
		}),
	)

	result, err := service.CollectEvidence(context.Background(), "CHG-2026-0005")
	if err != nil {
		t.Fatalf("CollectEvidence() error = %v", err)
	}
	if evidenceStore.changeID != "change-id-1" {
		t.Fatalf("evidence changeID = %q", evidenceStore.changeID)
	}
	if changeStore.markedStatus != "EvidenceCollected" {
		t.Fatalf("marked status = %q", changeStore.markedStatus)
	}
	evidence, ok := result["evidence"].(map[string]any)
	if !ok {
		t.Fatalf("missing evidence result: %+v", result)
	}
	if evidence["evidenceType"] != "deployment" || evidence["name"] != "deployment-evidence" {
		t.Fatalf("unexpected evidence result: %+v", evidence)
	}
}

type collectEvidenceFakeChangeStore struct {
	change       domain.ChangeRequest
	markedStatus string
}

func (s *collectEvidenceFakeChangeStore) List(ctx context.Context) ([]domain.ChangeRequest, error) {
	return nil, nil
}
func (s *collectEvidenceFakeChangeStore) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, nil
}
func (s *collectEvidenceFakeChangeStore) Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error) {
	return s.change, nil
}
func (s *collectEvidenceFakeChangeStore) Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error) {
	return nil, nil
}
func (s *collectEvidenceFakeChangeStore) TransitionLifecycle(ctx context.Context, idOrNumber string, action string, actor string, message string) (map[string]any, error) {
	return nil, nil
}
func (s *collectEvidenceFakeChangeStore) MarkStep(ctx context.Context, idOrNumber string, status string) (map[string]any, error) {
	s.markedStatus = status
	return map[string]any{"changeNumber": idOrNumber, "runtimeStatus": status, "previousRuntimeStatus": "DeploymentSyncedHealthy"}, nil
}

type collectEvidenceFakeEvidenceStore struct {
	changeID string
}

func (s *collectEvidenceFakeEvidenceStore) Create(ctx context.Context, changeID string, evidence domain.Evidence) (domain.Evidence, error) {
	s.changeID = changeID
	evidence.ID = "evidence-id-1"
	evidence.ChangeNumber = "CHG-2026-0005"
	return evidence, nil
}
