package app

import "github.com/vincmarz/devops-control-plane/internal/domain"

// ApplicationTopology is the application-level view that correlates logical
// applications, environment instances, and standalone Argo CD Applications.
type ApplicationTopology struct {
	LogicalApplications    []LogicalApplication       `json:"logicalApplications"`
	StandaloneApplications []ArgoCDApplicationSummary `json:"standaloneApplications"`
}

// LogicalApplication is the functional application identity shared by its
// environment-specific runtime instances.
type LogicalApplication struct {
	Name         string                `json:"name"`
	Environments []EnvironmentInstance `json:"environments"`
}

// EnvironmentInstance describes one logical application in one configured
// environment and carries the observed Argo CD state when available.
type EnvironmentInstance struct {
	Environment           string `json:"environment"`
	DisplayName           string `json:"displayName"`
	ClusterName           string `json:"clusterName"`
	KubernetesNamespace   string `json:"kubernetesNamespace"`
	ArgoCDApplicationName string `json:"argocdApplicationName"`
	Observed              bool   `json:"observed"`
	SyncStatus            string `json:"syncStatus"`
	HealthStatus          string `json:"healthStatus"`
	Revision              string `json:"revision,omitempty"`
	RepositoryURL         string `json:"repoURL,omitempty"`
	Path                  string `json:"path,omitempty"`
}

// ArgoCDApplicationSummary represents an observed Argo CD Application that is
// not bound to an Environment Catalog entry.
type ArgoCDApplicationSummary struct {
	Name            string `json:"name"`
	TargetNamespace string `json:"targetNamespace"`
	SyncStatus      string `json:"syncStatus"`
	HealthStatus    string `json:"healthStatus"`
	CurrentRevision string `json:"currentRevision,omitempty"`
	RepositoryURL   string `json:"repoURL,omitempty"`
	Path            string `json:"path,omitempty"`
}

func NewArgoCDApplicationSummary(application domain.Application) ArgoCDApplicationSummary {
	return ArgoCDApplicationSummary{
		Name:            application.Name,
		TargetNamespace: application.TargetNamespace,
		SyncStatus:      application.SyncStatus,
		HealthStatus:    application.HealthStatus,
		CurrentRevision: application.CurrentRevision,
		RepositoryURL:   application.RepoURL,
		Path:            application.Path,
	}
}
