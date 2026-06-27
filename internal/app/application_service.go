package app

import (
	"context"
	"errors"
	"strings"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type ArgoCDApplicationClient interface {
	ListApplications(ctx context.Context) ([]domain.Application, error)
	GetApplication(ctx context.Context, name string) (domain.Application, error)
}

type ApplicationService struct {
	argocd ArgoCDApplicationClient
}

func NewApplicationService(clients ...ArgoCDApplicationClient) *ApplicationService {
	service := &ApplicationService{}
	if len(clients) > 0 {
		service.argocd = clients[0]
	}
	return service
}

func (s *ApplicationService) List(ctx context.Context) ([]domain.Application, error) {
	if s.argocd == nil {
		return nil, errors.New("argocd application client is not configured")
	}
	return s.argocd.ListApplications(ctx)
}

func (s *ApplicationService) Get(ctx context.Context, name string) (domain.Application, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Application{}, errors.New("application name is required")
	}
	if s.argocd == nil {
		return domain.Application{}, errors.New("argocd application client is not configured")
	}
	return s.argocd.GetApplication(ctx, name)
}

func (s *ApplicationService) Resources(ctx context.Context, name string) []domain.ApplicationResource {
	app, err := s.Get(ctx, name)
	if err != nil {
		return []domain.ApplicationResource{}
	}
	return []domain.ApplicationResource{
		{Group: "apps", Kind: "Deployment", Namespace: app.TargetNamespace, Name: app.Name, Status: app.SyncStatus, Health: app.HealthStatus, Orphaned: false},
		{Group: "", Kind: "Service", Namespace: app.TargetNamespace, Name: app.Name, Status: app.SyncStatus, Health: app.HealthStatus, Orphaned: false},
	}
}

func (s *ApplicationService) History(ctx context.Context, name string) []domain.ApplicationHistoryItem {
	return []domain.ApplicationHistoryItem{}
}

func (s *ApplicationService) Runtime(ctx context.Context, name string) map[string]any {
	app, err := s.Get(ctx, name)
	if err != nil {
		return map[string]any{"application": name, "error": err.Error()}
	}
	return map[string]any{
		"application":     app.Name,
		"namespace":       app.TargetNamespace,
		"syncStatus":      app.SyncStatus,
		"healthStatus":    app.HealthStatus,
		"currentRevision": app.CurrentRevision,
		"source":          app.Source,
		"destination":     app.Destination,
	}
}
