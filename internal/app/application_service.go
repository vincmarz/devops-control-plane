package app

import (
	"context"
	"errors"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type ApplicationService struct{}

func NewApplicationService() *ApplicationService { return &ApplicationService{} }

func (s *ApplicationService) List(ctx context.Context) ([]domain.Application, error) {
	return []domain.Application{
		{
			Name:            "demo-go-color-app",
			ArgoCDNamespace: "openshift-gitops",
			Project:         "devops-ci-demo",
			TargetNamespace: "devops-ci-demo",
			RepoURL:         "https://gitlab.example.local/group/demo-app-gitops.git",
			TargetRevision:  "main",
			Path:            "apps/demo-go-color-app",
			SyncStatus:      "Unknown",
			HealthStatus:    "Unknown",
			CurrentRevision: "",
		},
	}, nil
}

func (s *ApplicationService) Get(ctx context.Context, name string) (domain.Application, error) {
	apps, _ := s.List(ctx)
	for _, app := range apps {
		if app.Name == name {
			app.Source = map[string]string{"repoUrl": app.RepoURL, "targetRevision": app.TargetRevision, "path": app.Path}
			app.Destination = map[string]string{"server": "https://kubernetes.default.svc", "namespace": app.TargetNamespace}
			return app, nil
		}
	}
	return domain.Application{}, errors.New("application not found in placeholder service")
}

func (s *ApplicationService) Resources(ctx context.Context, name string) []domain.ApplicationResource {
	return []domain.ApplicationResource{
		{Group: "apps", Kind: "Deployment", Namespace: "devops-ci-demo", Name: name, Status: "Unknown", Health: "Unknown", Orphaned: false},
		{Group: "", Kind: "Service", Namespace: "devops-ci-demo", Name: name, Status: "Unknown", Health: "Unknown", Orphaned: false},
	}
}

func (s *ApplicationService) History(ctx context.Context, name string) []domain.ApplicationHistoryItem {
	return []domain.ApplicationHistoryItem{}
}

func (s *ApplicationService) Runtime(ctx context.Context, name string) map[string]any {
	return map[string]any{
		"namespace": "devops-ci-demo",
		"deployment": map[string]any{"name": name, "desiredReplicas": 0, "readyReplicas": 0, "availableReplicas": 0},
		"pods": []any{},
	}
}
