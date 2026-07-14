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

// GroupByEnvironment correlates explicit Environment Catalog bindings with
// observed Argo CD Applications. It does not infer relationships from names or
// suffixes. Applications not referenced by the catalog remain standalone.
func (s *ApplicationService) GroupByEnvironment(applications []domain.Application, catalog EnvironmentCatalog) ApplicationTopology {
	observed := make(map[string]domain.Application, len(applications))
	for _, application := range applications {
		name := strings.TrimSpace(application.Name)
		if name != "" {
			observed[name] = application
		}
	}

	logicalByName := map[string]*LogicalApplication{}
	logicalOrder := []string{}
	used := map[string]bool{}
	for _, environmentName := range []string{"dev", "staging", "production"} {
		definition, ok := catalog.Resolve(environmentName)
		if !ok {
			continue
		}
		logicalName := strings.TrimSpace(definition.ApplicationName)
		argoCDName := strings.TrimSpace(definition.ArgoCDApplicationName)
		if logicalName == "" || argoCDName == "" {
			continue
		}

		logical, ok := logicalByName[logicalName]
		if !ok {
			logical = &LogicalApplication{Name: logicalName, Environments: []EnvironmentInstance{}}
			logicalByName[logicalName] = logical
			logicalOrder = append(logicalOrder, logicalName)
		}

		instance := EnvironmentInstance{
			Environment:           definition.Name,
			DisplayName:           definition.DisplayName,
			ClusterName:           definition.ClusterName,
			KubernetesNamespace:   definition.KubernetesNamespace,
			ArgoCDApplicationName: argoCDName,
			Observed:              false,
			SyncStatus:            "Not observed",
			HealthStatus:          "Not observed",
		}
		if application, found := observed[argoCDName]; found {
			instance.Observed = true
			instance.SyncStatus = application.SyncStatus
			instance.HealthStatus = application.HealthStatus
			instance.Revision = application.CurrentRevision
			instance.RepositoryURL = application.RepoURL
			instance.Path = application.Path
			used[argoCDName] = true
		}
		logical.Environments = append(logical.Environments, instance)
	}

	logicalApplications := make([]LogicalApplication, 0, len(logicalOrder))
	for _, name := range logicalOrder {
		logicalApplications = append(logicalApplications, *logicalByName[name])
	}

	standalone := make([]ArgoCDApplicationSummary, 0)
	for _, application := range applications {
		name := strings.TrimSpace(application.Name)
		if name != "" && !used[name] {
			standalone = append(standalone, NewArgoCDApplicationSummary(application))
		}
	}

	return ApplicationTopology{
		LogicalApplications:    logicalApplications,
		StandaloneApplications: standalone,
	}
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
