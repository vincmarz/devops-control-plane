package argocd

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
