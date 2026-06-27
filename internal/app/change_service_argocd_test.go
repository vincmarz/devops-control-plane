package app

import (
	"context"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func TestMapArgoCDDeploymentRuntimeStatus(t *testing.T) {
	tests := []struct {
		name   string
		sync   string
		health string
		want   string
	}{
		{name: "synced healthy", sync: "Synced", health: "Healthy", want: "DeploymentSyncedHealthy"},
		{name: "synced progressing", sync: "Synced", health: "Progressing", want: "DeploymentProgressing"},
		{name: "out of sync", sync: "OutOfSync", health: "Healthy", want: "DeploymentOutOfSync"},
		{name: "degraded wins", sync: "Synced", health: "Degraded", want: "DeploymentDegraded"},
		{name: "unknown", sync: "Unknown", health: "Missing", want: "DeploymentUnknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapArgoCDDeploymentRuntimeStatus(tt.sync, tt.health)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckDeployment(t *testing.T) {
	store := &argocdFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-2026-0005", ApplicationName: "demo-go-color-app"}}
	service := NewChangeService(store, WithArgoCDCheckDeployment(func(ctx context.Context, change domain.ChangeRequest) (ArgoCDDeploymentResult, error) {
		return ArgoCDDeploymentResult{ApplicationName: change.ApplicationName, Project: "default", SyncStatus: "Synced", HealthStatus: "Healthy", Revision: "abc123"}, nil
	}))

	result, err := service.CheckDeployment(context.Background(), "CHG-2026-0005")
	if err != nil {
		t.Fatalf("CheckDeployment() error = %v", err)
	}
	if store.markedStatus != "DeploymentSyncedHealthy" {
		t.Fatalf("marked status = %q", store.markedStatus)
	}
	argo, ok := result["argocd"].(map[string]any)
	if !ok {
		t.Fatalf("argocd payload missing: %+v", result)
	}
	if argo["applicationName"] != "demo-go-color-app" || argo["revision"] != "abc123" {
		t.Fatalf("unexpected argocd payload: %+v", argo)
	}
}

type argocdFakeStore struct {
	change       domain.ChangeRequest
	markedStatus string
}

func (s *argocdFakeStore) List(ctx context.Context) ([]domain.ChangeRequest, error) { return nil, nil }
func (s *argocdFakeStore) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, nil
}
func (s *argocdFakeStore) Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error) {
	return s.change, nil
}
func (s *argocdFakeStore) Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error) {
	return nil, nil
}
func (s *argocdFakeStore) TransitionLifecycle(ctx context.Context, idOrNumber string, action string, actor string, message string) (map[string]any, error) {
	return nil, nil
}
func (s *argocdFakeStore) MarkStep(ctx context.Context, idOrNumber string, status string) (map[string]any, error) {
	s.markedStatus = status
	return map[string]any{"changeNumber": idOrNumber, "runtimeStatus": status}, nil
}
