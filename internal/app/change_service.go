package app

import (
	"context"
	"errors"
	"strings"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type ChangeStore interface {
	List(ctx context.Context) ([]domain.ChangeRequest, error)
	Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error)
	Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error)
	Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error)
	MarkStep(ctx context.Context, idOrNumber string, status string) (map[string]any, error)
}

type ChangeService struct {
	store ChangeStore
}

func NewChangeService(store ChangeStore) *ChangeService {
	return &ChangeService{store: store}
}

func (s *ChangeService) List(ctx context.Context) ([]domain.ChangeRequest, error) {
	return s.store.List(ctx)
}

func (s *ChangeService) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.ApplicationName = strings.TrimSpace(req.ApplicationName)
	req.TargetEnvironment = strings.TrimSpace(req.TargetEnvironment)
	req.ChangeType = strings.TrimSpace(req.ChangeType)
	req.RiskLevel = strings.TrimSpace(req.RiskLevel)
	req.RequestedBy = strings.TrimSpace(req.RequestedBy)

	if req.Title == "" {
		return domain.ChangeRequest{}, errors.New("title is required")
	}
	if req.ApplicationName == "" {
		return domain.ChangeRequest{}, errors.New("applicationName is required")
	}
	if req.ChangeType == "" {
		return domain.ChangeRequest{}, errors.New("changeType is required")
	}
	if req.RequestedBy == "" {
		return domain.ChangeRequest{}, errors.New("requestedBy is required")
	}

	if req.TargetEnvironment == "" {
		req.TargetEnvironment = "dev"
	}
	if req.RiskLevel == "" {
		req.RiskLevel = domain.RiskMedium
	}

	if !isAllowedRiskLevel(req.RiskLevel) {
		return domain.ChangeRequest{}, errors.New("riskLevel must be one of: low, medium, high, critical")
	}

	return s.store.Create(ctx, req)
}

func (s *ChangeService) Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error) {
	return s.store.Get(ctx, idOrNumber)
}

func (s *ChangeService) Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error) {
	return s.store.Events(ctx, idOrNumber)
}

// MarkStep registra uno step tecnico del workflow.
// Nota importante:
// Non deve modificare lo stato governato della ChangeRequest.
// Gli step tecnici come BranchCreated, CommitCreated o SyncRunning devono
// essere tracciati come eventi/runtime status, non come lifecycle status.
func (s *ChangeService) MarkStep(ctx context.Context, idOrNumber string, status string) (map[string]any, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return nil, errors.New("workflow step status is required")
	}
	return s.store.MarkStep(ctx, idOrNumber, status)
}

func isAllowedRiskLevel(riskLevel string) bool {
	switch riskLevel {
	case domain.RiskLow, domain.RiskMedium, domain.RiskHigh, domain.RiskCritical:
		return true
	default:
		return false
	}
}
