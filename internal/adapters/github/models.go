package github

// Config configures a GitHub REST API client. APIURL supports both GitHub.com
// and GitHub Enterprise API roots.
type Config struct {
	APIURL         string
	Token          string
	TimeoutSeconds int
	InsecureTLS    bool
	CAFile         string
}

type Reference struct {
	Ref    string `json:"ref"`
	Object struct {
		SHA string `json:"sha"`
	} `json:"object"`
}

type RepositoryContent struct {
	SHA string `json:"sha"`
}

type PullRequest struct {
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
	State   string `json:"state"`
	Head    struct {
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

type PullRequestMerge struct {
	SHA     string `json:"sha"`
	Merged  bool   `json:"merged"`
	Message string `json:"message"`
}
