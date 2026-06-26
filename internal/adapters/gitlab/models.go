package gitlab

import "time"

type RepositoryFile struct {
	FilePath       string `json:"file_path"`
	Branch         string `json:"branch"`
	Ref            string `json:"ref,omitempty"`
	ContentDecoded string `json:"content_decoded,omitempty"`
	CommitID       string `json:"commit_id,omitempty"`
	LastCommitID   string `json:"last_commit_id,omitempty"`
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

// Branch rappresenta la risposta GitLab per un repository branch.
type Branch struct {
	Name               string `json:"name"`
	Commit             Commit `json:"commit"`
	Merged             bool   `json:"merged"`
	Protected          bool   `json:"protected"`
	DevelopersCanPush  bool   `json:"developers_can_push"`
	DevelopersCanMerge bool   `json:"developers_can_merge"`
	CanPush            bool   `json:"can_push"`
	Default            bool   `json:"default"`
	WebURL             string `json:"web_url"`
}

// Commit rappresenta i campi principali di un commit GitLab.
type Commit struct {
	ID             string    `json:"id"`
	ShortID        string    `json:"short_id"`
	CreatedAt      time.Time `json:"created_at"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	AuthoredDate   time.Time `json:"authored_date"`
	CommitterName  string    `json:"committer_name"`
	CommitterEmail string    `json:"committer_email"`
	CommittedDate  time.Time `json:"committed_date"`
	WebURL         string    `json:"web_url"`
}
