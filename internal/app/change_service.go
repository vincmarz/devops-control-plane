package app

import (
	"context"
	"errors"

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
	if req.ApplicationName == "" || req.ChangeType == "" {
		return domain.ChangeRequest{}, errors.New("applicationName and changeType are required")
	}
	return s.store.Create(ctx, req)
}

func (s *ChangeService) Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error) {
	return s.store.Get(ctx, idOrNumber)
}

func (s *ChangeService) Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error) {
	return s.store.Events(ctx, idOrNumber)
}

func (s *ChangeService) MarkStep(ctx context.Context, idOrNumber string, status string) (map[string]any, error) {
	return s.store.MarkStep(ctx, idOrNumber, status)
}
