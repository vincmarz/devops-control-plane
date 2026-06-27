package app

import (
	"context"
	"time"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type EvidenceListStore interface {
	List(ctx context.Context, idOrNumber string, evidenceType string) ([]domain.Evidence, error)
}

type EvidenceService struct {
	store EvidenceListStore
}

func NewEvidenceService(stores ...EvidenceListStore) *EvidenceService {
	service := &EvidenceService{}
	if len(stores) > 0 {
		service.store = stores[0]
	}
	return service
}

func (s *EvidenceService) List(ctx context.Context, changeID, evidenceType string) ([]domain.Evidence, error) {
	if s.store != nil {
		return s.store.List(ctx, changeID, evidenceType)
	}

	placeholder := domain.Evidence{
		ID:           "ev-placeholder",
		EvidenceType: "workflow-summary",
		Name:         "workflow-summary-placeholder",
		Summary:      "Evidence service placeholder. Real evidence persistence will be implemented with PostgreSQL.",
		Sanitized:    true,
		CreatedAt:    time.Now().UTC(),
	}
	if evidenceType != "" && evidenceType != placeholder.EvidenceType {
		return []domain.Evidence{}, nil
	}
	return []domain.Evidence{placeholder}, nil
}
