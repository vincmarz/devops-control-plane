package argocd

import "github.com/vincmarz/devops-control-plane/internal/domain"

type Config struct {
	BaseURL        string
	AuthToken      string
	TimeoutSeconds int
	InsecureTLS    bool
}

type ApplicationStatus struct {
	Name            string
	Project         string
	SyncStatus      string
	HealthStatus    string
	CurrentRevision string
}

type SyncResult struct {
	Application string
	Revision    string
	Phase       string
	Message     string
}

type applicationListResponse struct {
	Items []applicationResponse `json:"items"`
}

type applicationResponse struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Project string `json:"project"`
		Source  struct {
			RepoURL        string `json:"repoURL"`
			TargetRevision string `json:"targetRevision"`
			Path           string `json:"path"`
		} `json:"source"`
		Destination struct {
			Server    string `json:"server"`
			Namespace string `json:"namespace"`
			Name      string `json:"name"`
		} `json:"destination"`
	} `json:"spec"`
	Status struct {
		Sync struct {
			Status   string `json:"status"`
			Revision string `json:"revision"`
		} `json:"sync"`
		Health struct {
			Status string `json:"status"`
		} `json:"health"`
		Conditions []map[string]any `json:"conditions"`
	} `json:"status"`
}

func (a applicationResponse) toDomain() domain.Application {
	app := domain.Application{
		Name:            a.Metadata.Name,
		ArgoCDNamespace: a.Metadata.Namespace,
		Project:         a.Spec.Project,
		TargetNamespace: a.Spec.Destination.Namespace,
		RepoURL:         a.Spec.Source.RepoURL,
		TargetRevision:  a.Spec.Source.TargetRevision,
		Path:            a.Spec.Source.Path,
		SyncStatus:      a.Status.Sync.Status,
		HealthStatus:    a.Status.Health.Status,
		CurrentRevision: a.Status.Sync.Revision,
		Conditions:      a.Status.Conditions,
	}
	app.Source = map[string]string{"repoUrl": app.RepoURL, "targetRevision": app.TargetRevision, "path": app.Path}
	app.Destination = map[string]string{"server": a.Spec.Destination.Server, "namespace": app.TargetNamespace}
	if a.Spec.Destination.Name != "" {
		app.Destination["name"] = a.Spec.Destination.Name
	}
	return app
}
