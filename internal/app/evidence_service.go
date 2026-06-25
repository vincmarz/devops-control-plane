package app

import (
	"context"
	"time"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type EvidenceService struct{}

func NewEvidenceService() *EvidenceService { return &EvidenceService{} }

func (s *EvidenceService) List(ctx context.Context, changeID, evidenceType string) []domain.Evidence {
	placeholder := domain.Evidence{
		ID:           "ev-placeholder",
		EvidenceType: "workflow-summary",
		Name:         "workflow-summary-placeholder",
		Summary:      "Evidence service placeholder. Real evidence persistence will be implemented with PostgreSQL.",
		Sanitized:    true,
		CreatedAt:    time.Now().UTC(),
	}
	if evidenceType != "" && evidenceType != placeholder.EvidenceType {
		return []domain.Evidence{}
	}
	return []domain.Evidence{placeholder}
}
