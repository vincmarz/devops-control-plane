package app

import (
	"context"
	"errors"
	"strings"
)

// GitRepositoryTarget is the provider-neutral repository identity resolved
// from an Application Catalog binding.
type GitRepositoryTarget struct {
	Provider      string
	ProviderRef   string
	ProjectID     int
	ProjectPath   string
	RepositoryURL string
	DefaultBranch string
}

// NewGitRepositoryTarget converts a validated repository binding into the
// minimal target required by Git workflow providers.
func NewGitRepositoryTarget(binding RepositoryBinding) (GitRepositoryTarget, error) {
	target := GitRepositoryTarget{
		Provider:      strings.ToLower(strings.TrimSpace(binding.Provider)),
		ProviderRef:   strings.TrimSpace(binding.ProviderRef),
		ProjectID:     binding.ProjectID,
		ProjectPath:   strings.TrimSpace(binding.ProjectPath),
		RepositoryURL: strings.TrimSpace(binding.RepositoryURL),
		DefaultBranch: strings.TrimSpace(binding.DefaultBranch),
	}
	if target.Provider == "" {
		return GitRepositoryTarget{}, errors.New("git repository target provider is required")
	}
	if target.ProviderRef == "" {
		return GitRepositoryTarget{}, errors.New("git repository target providerRef is required")
	}
	if target.ProjectPath == "" {
		return GitRepositoryTarget{}, errors.New("git repository target projectPath is required")
	}
	if target.RepositoryURL == "" {
		return GitRepositoryTarget{}, errors.New("git repository target repositoryURL is required")
	}
	if target.DefaultBranch == "" {
		return GitRepositoryTarget{}, errors.New("git repository target defaultBranch is required")
	}
	return target, nil
}

// GitProvider is the provider-neutral Git workflow contract. A GitLab adapter
// maps merge request operations directly; a GitHub adapter maps them to pull
// request operations while preserving the Control Plane contract.
type GitProvider interface {
	Provider() string
	ProviderRef() string
	CreateBranch(ctx context.Context, target GitRepositoryTarget, branch, ref string) error
	CreateOrUpdateFile(ctx context.Context, target GitRepositoryTarget, branch, filePath, commitMessage, content string) error
	OpenMergeRequest(ctx context.Context, target GitRepositoryTarget, sourceBranch, targetBranch, title, description string) (int, string, error)
	MergeRequest(ctx context.Context, target GitRepositoryTarget, sourceBranch, targetBranch, mergeCommitMessage string) (int, string, string, error)
}
