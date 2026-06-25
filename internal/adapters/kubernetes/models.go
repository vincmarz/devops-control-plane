package kubernetes

type DeploymentSummary struct {
	Namespace         string
	Name              string
	DesiredReplicas   int32
	ReadyReplicas     int32
	AvailableReplicas int32
	UpdatedReplicas   int32
}

type PodSummary struct {
	Namespace    string
	Name         string
	Phase        string
	Ready        bool
	RestartCount int32
}
