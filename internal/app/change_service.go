package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type ChangeService struct {
	mu      sync.Mutex
	counter int
	items   map[string]domain.ChangeRequest
	events  map[string][]domain.ChangeEvent
}

func NewChangeService() *ChangeService {
	return &ChangeService{items: map[string]domain.ChangeRequest{}, events: map[string][]domain.ChangeEvent{}}
}

func (s *ChangeService) List(ctx context.Context) []domain.ChangeRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.ChangeRequest, 0, len(s.items))
	for _, item := range s.items {
		out = append(out, item)
	}
	return out
}

func (s *ChangeService) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	if req.ApplicationName == "" || req.ChangeType == "" {
		return domain.ChangeRequest{}, errors.New("applicationName and changeType are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	now := time.Now().UTC()
	id := fmt.Sprintf("chg-%04d", s.counter)
	changeNumber := fmt.Sprintf("CHG-%d-%04d", now.Year(), s.counter)
	change := domain.ChangeRequest{
		ID:              id,
		ChangeNumber:    changeNumber,
		ApplicationName: req.ApplicationName,
		ChangeType:      req.ChangeType,
		Status:          "Created",
		RequestedBy:     req.RequestedBy,
		Description:     req.Description,
		Payload:         req.Payload,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	s.items[id] = change
	s.events[id] = append(s.events[id], domain.ChangeEvent{EventType: "Created", NewStatus: "Created", Message: "ChangeRequest created", Source: "workflow", CreatedAt: now})
	return change, nil
}

func (s *ChangeService) Get(ctx context.Context, id string) (domain.ChangeRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	change, ok := s.items[id]
	if !ok {
		return domain.ChangeRequest{}, errors.New("change not found")
	}
	return change, nil
}

func (s *ChangeService) Events(ctx context.Context, id string) []domain.ChangeEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.ChangeEvent{}, s.events[id]...)
}

func (s *ChangeService) MarkStep(id, status string) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	change, ok := s.items[id]
	if ok {
		previous := change.Status
		change.Status = status
		change.UpdatedAt = now
		s.items[id] = change
		s.events[id] = append(s.events[id], domain.ChangeEvent{EventType: status, PreviousStatus: previous, NewStatus: status, Message: "Workflow placeholder step executed", Source: "workflow", CreatedAt: now})
	}
	return map[string]any{"id": id, "status": status, "message": "placeholder workflow step, adapter not implemented yet"}
}
