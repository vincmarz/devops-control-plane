package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/vincmarz/devops-control-plane/internal/app"
)

// ProviderClient is the minimal GitHub client contract required by the
// provider-neutral Git workflow adapter.
type ProviderClient interface {
	CreateBranch(ctx context.Context, projectPath, branch, ref string) (Reference, error)
	CreateOrUpdateFile(ctx context.Context, projectPath, branch, filePath, commitMessage, content string) (RepositoryContent, error)
	OpenPullRequest(ctx context.Context, projectPath, sourceBranch, targetBranch, title, description string) (PullRequest, error)
	FindOpenPullRequest(ctx context.Context, projectPath, sourceBranch, targetBranch string) (PullRequest, error)
	MergePullRequest(ctx context.Context, projectPath string, number int, commitMessage string) (PullRequestMerge, error)
}

// Provider adapts the GitHub REST client to the provider-neutral app.GitProvider contract.
type Provider struct {
	providerRef string
	client      ProviderClient
}

func NewProvider(providerRef string, client ProviderClient) (*Provider, error) {
	providerRef = strings.TrimSpace(providerRef)
	if providerRef == "" {
		return nil, errors.New("github providerRef is required")
	}
	if client == nil {
		return nil, errors.New("github provider client is required")
	}
	return &Provider{providerRef: providerRef, client: client}, nil
}

func (p *Provider) Provider() string    { return app.RepositoryProviderGitHub }
func (p *Provider) ProviderRef() string { return p.providerRef }

func (p *Provider) CreateBranch(ctx context.Context, target app.GitRepositoryTarget, branch, ref string) error {
	if err := p.validateTarget(target); err != nil {
		return err
	}
	_, err := p.client.CreateBranch(ctx, target.ProjectPath, branch, ref)
	return err
}

func (p *Provider) CreateOrUpdateFile(ctx context.Context, target app.GitRepositoryTarget, branch, filePath, commitMessage, content string) error {
	if err := p.validateTarget(target); err != nil {
		return err
	}
	_, err := p.client.CreateOrUpdateFile(ctx, target.ProjectPath, branch, filePath, commitMessage, content)
	return err
}

func (p *Provider) OpenMergeRequest(ctx context.Context, target app.GitRepositoryTarget, sourceBranch, targetBranch, title, description string) (int, string, error) {
	if err := p.validateTarget(target); err != nil {
		return 0, "", err
	}
	pullRequest, err := p.client.OpenPullRequest(ctx, target.ProjectPath, sourceBranch, targetBranch, title, description)
	if err != nil {
		return 0, "", err
	}
	return pullRequest.Number, pullRequest.HTMLURL, nil
}

func (p *Provider) MergeRequest(ctx context.Context, target app.GitRepositoryTarget, sourceBranch, targetBranch, mergeCommitMessage string) (int, string, string, error) {
	if err := p.validateTarget(target); err != nil {
		return 0, "", "", err
	}
	pullRequest, err := p.client.FindOpenPullRequest(ctx, target.ProjectPath, sourceBranch, targetBranch)
	if err != nil {
		return 0, "", "", err
	}
	merged, err := p.client.MergePullRequest(ctx, target.ProjectPath, pullRequest.Number, mergeCommitMessage)
	if err != nil {
		return 0, "", "", err
	}
	return pullRequest.Number, pullRequest.HTMLURL, merged.SHA, nil
}

func (p *Provider) validateTarget(target app.GitRepositoryTarget) error {
	if strings.ToLower(strings.TrimSpace(target.Provider)) != app.RepositoryProviderGitHub {
		return fmt.Errorf("github provider cannot handle provider %q", target.Provider)
	}
	if strings.TrimSpace(target.ProviderRef) != p.providerRef {
		return fmt.Errorf("github providerRef %q cannot handle target providerRef %q", p.providerRef, target.ProviderRef)
	}
	if _, _, err := ParseProjectPath(target.ProjectPath); err != nil {
		return err
	}
	return nil
}

var _ app.GitProvider = (*Provider)(nil)
