package domain

type Application struct {
	Name              string            `json:"name"`
	ArgoCDNamespace   string            `json:"argocdNamespace"`
	Project           string            `json:"project"`
	TargetNamespace   string            `json:"targetNamespace"`
	RepoURL           string            `json:"repoUrl"`
	TargetRevision    string            `json:"targetRevision"`
	Path              string            `json:"path"`
	SyncStatus        string            `json:"syncStatus"`
	HealthStatus      string            `json:"healthStatus"`
	CurrentRevision   string            `json:"currentRevision"`
	Source            map[string]string `json:"source,omitempty"`
	Destination       map[string]string `json:"destination,omitempty"`
	Conditions        []map[string]any  `json:"conditions,omitempty"`
}

type ApplicationResource struct {
	Group     string `json:"group"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Status    string `json:"status,omitempty"`
	Health    string `json:"health,omitempty"`
	Orphaned  bool   `json:"orphaned"`
}

type ApplicationHistoryItem struct {
	ID         int    `json:"id"`
	Revision   string `json:"revision"`
	DeployedAt string `json:"deployedAt"`
	SourcePath string `json:"sourcePath"`
}
