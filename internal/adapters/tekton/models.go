package tekton

type PipelineRunRef struct {
	Name      string
	Namespace string
	UID       string
}
type PipelineRunStatus struct {
	Name      string
	Namespace string
	Status    string
	Reason    string
	Message   string
}
type TaskRunStatus struct {
	Name     string
	TaskName string
	Status   string
	Reason   string
}
