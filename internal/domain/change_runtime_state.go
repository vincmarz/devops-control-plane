package domain

import "time"

type ChangeRuntimeState struct {
	ChangeRequestID string                  `json:"changeRequestID"`
	Source          SourceRuntimeState      `json:"source"`
	GitOps          GitOpsRuntimeState      `json:"gitops"`
	Tekton          TektonRuntimeState      `json:"tekton"`
	ArgoCD          ArgoCDRuntimeState      `json:"argocd"`
	Runtime         RuntimeObservationState `json:"runtime"`
	CreatedAt       time.Time               `json:"createdAt,omitempty"`
	UpdatedAt       time.Time               `json:"updatedAt,omitempty"`
}

type SourceRuntimeState struct {
	Provider      string `json:"provider,omitempty"`
	ProviderRef   string `json:"providerRef,omitempty"`
	ProjectID     int    `json:"projectID,omitempty"`
	ProjectPath   string `json:"projectPath,omitempty"`
	RepositoryURL string `json:"repositoryURL,omitempty"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
	Branch        string `json:"branch,omitempty"`
	CommitSHA     string `json:"commitSHA,omitempty"`
}

type GitOpsRuntimeState struct {
	Provider      string `json:"provider,omitempty"`
	ProviderRef   string `json:"providerRef,omitempty"`
	ProjectID     int    `json:"projectID,omitempty"`
	ProjectPath   string `json:"projectPath,omitempty"`
	RepositoryURL string `json:"repositoryURL,omitempty"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
	Revision      string `json:"revision,omitempty"`
	CommitSHA     string `json:"commitSHA,omitempty"`
}

type TektonRuntimeState struct {
	Namespace       string `json:"namespace,omitempty"`
	PipelineName    string `json:"pipelineName,omitempty"`
	PipelineRunName string `json:"pipelineRunName,omitempty"`
	UID             string `json:"uid,omitempty"`
	GitURL          string `json:"gitURL,omitempty"`
	GitRevision     string `json:"gitRevision,omitempty"`
	ValidationPath  string `json:"validationPath,omitempty"`
	Status          string `json:"status,omitempty"`
	Reason          string `json:"reason,omitempty"`
	Message         string `json:"message,omitempty"`
}

type ArgoCDRuntimeState struct {
	ApplicationName        string `json:"applicationName,omitempty"`
	Provider               string `json:"provider,omitempty"`
	ProviderRef            string `json:"providerRef,omitempty"`
	ProjectPath            string `json:"projectPath,omitempty"`
	DeclaredRepositoryURL  string `json:"declaredRepositoryURL,omitempty"`
	ObservedRepositoryURL  string `json:"observedRepositoryURL,omitempty"`
	DeclaredDefaultBranch  string `json:"declaredDefaultBranch,omitempty"`
	ObservedTargetRevision string `json:"observedTargetRevision,omitempty"`
	ObservedRevision       string `json:"observedRevision,omitempty"`
	SyncStatus             string `json:"syncStatus,omitempty"`
	HealthStatus           string `json:"healthStatus,omitempty"`
	CorrelationStatus      string `json:"correlationStatus,omitempty"`
}

type RuntimeObservationState struct {
	ClusterName  string `json:"clusterName,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
	ResourceKind string `json:"resourceKind,omitempty"`
	ResourceName string `json:"resourceName,omitempty"`
	Status       string `json:"status,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
}
