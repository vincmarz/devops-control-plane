package gitlab

type RepositoryFile struct {
	FilePath       string
	Ref            string
	ContentDecoded string
	CommitID       string
	LastCommitID   string
}

type CommitAction struct {
	Action   string
	FilePath string
	Content  string
}

type MergeRequest struct {
	IID          int
	Title        string
	State        string
	SourceBranch string
	TargetBranch string
	WebURL       string
}
