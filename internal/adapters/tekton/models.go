package tekton

type PipelineRunRef struct {
	Name      string
	Namespace string
	UID       string
}
type PipelineRunStatus struct {
	Name              string
	Namespace         string
	UID               string
	Status            string
	Reason            string
	Message           string
	CreationTimestamp string
	CompletionTime    string
	Results           []PipelineRunResult
}
type TaskRunStatus struct {
	Name             string
	Namespace        string
	PipelineTaskName string
	TaskName         string
	Status           string
	Reason           string
	Message          string
	StartTime        string
	CompletionTime   string
}

type PipelineRunResult struct {
	Name  string
	Value string
}
