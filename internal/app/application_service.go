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

type ApplicationGrouping struct {
	LogicalApplications    []map[string]any
	StandaloneApplications []map[string]any
}

// GroupByEnvironment correlates explicit Environment Catalog bindings with
// observed Argo CD Applications. It does not infer relationships from names or
// suffixes. Applications not referenced by the catalog remain standalone.
func (s *ApplicationService) GroupByEnvironment(applications []map[string]any, catalog EnvironmentCatalog) ApplicationGrouping {
	observed := make(map[string]map[string]any, len(applications))
	for _, application := range applications {
		name := applicationMapString(application, "name")
		if name != "" {
			observed[name] = application
		}
	}

	logicalByName := map[string]map[string]any{}
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

		group, ok := logicalByName[logicalName]
		if !ok {
			group = map[string]any{"name": logicalName, "environments": []map[string]any{}}
			logicalByName[logicalName] = group
			logicalOrder = append(logicalOrder, logicalName)
		}

		instance := map[string]any{
			"environment":           definition.Name,
			"displayName":           definition.DisplayName,
			"clusterName":           definition.ClusterName,
			"kubernetesNamespace":   definition.KubernetesNamespace,
			"argocdApplicationName": argoCDName,
			"observed":              false,
			"syncStatus":            "Not observed",
			"healthStatus":          "Not observed",
		}
		if application, found := observed[argoCDName]; found {
			instance["observed"] = true
			instance["syncStatus"] = application["syncStatus"]
			instance["healthStatus"] = application["healthStatus"]
			instance["revision"] = application["currentRevision"]
			instance["repoURL"] = application["repoURL"]
			instance["path"] = application["path"]
			used[argoCDName] = true
		}
		group["environments"] = append(group["environments"].([]map[string]any), instance)
	}

	logical := make([]map[string]any, 0, len(logicalOrder))
	for _, name := range logicalOrder {
		logical = append(logical, logicalByName[name])
	}
	standalone := make([]map[string]any, 0)
	for _, application := range applications {
		name := applicationMapString(application, "name")
		if name != "" && !used[name] {
			standalone = append(standalone, application)
		}
	}

	return ApplicationGrouping{
		LogicalApplications:    logical,
		StandaloneApplications: standalone,
	}
}

func applicationMapString(application map[string]any, key string) string {
	value, ok := application[key]
	if !ok || value == nil {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
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
