package gitlab

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/vincmarz/devops-control-plane/internal/app"
)

// ProviderClient is the minimal GitLab client contract required by the
// provider-neutral Git workflow adapter.
type ProviderClient interface {
	CreateBranch(ctx context.Context, projectID int, branch, ref string) (Branch, error)
	CreateOrUpdateFile(ctx context.Context, projectID int, branch, filePath, commitMessage, content string) (RepositoryFile, error)
	OpenMergeRequest(ctx context.Context, projectID int, sourceBranch, targetBranch, title, description string) (MergeRequest, error)
	FindOpenMergeRequest(ctx context.Context, projectID int, sourceBranch, targetBranch string) (MergeRequest, error)
	MergeMergeRequest(ctx context.Context, projectID, mergeRequestIID int, sha, mergeCommitMessage string) (MergeRequest, error)
}

// Provider adapts the GitLab REST client to the provider-neutral app.GitProvider contract.
type Provider struct {
	providerRef string
	client      ProviderClient
}

func NewProvider(providerRef string, client ProviderClient) (*Provider, error) {
	providerRef = strings.TrimSpace(providerRef)
	if providerRef == "" {
		return nil, errors.New("gitlab providerRef is required")
	}
	if client == nil {
		return nil, errors.New("gitlab provider client is required")
	}
	return &Provider{providerRef: providerRef, client: client}, nil
}

func (p *Provider) Provider() string    { return app.RepositoryProviderGitLab }
func (p *Provider) ProviderRef() string { return p.providerRef }

func (p *Provider) CreateBranch(ctx context.Context, target app.GitRepositoryTarget, branch, ref string) error {
	if err := p.validateTarget(target); err != nil {
		return err
	}
	_, err := p.client.CreateBranch(ctx, target.ProjectID, branch, ref)
	return err
}

func (p *Provider) CreateOrUpdateFile(ctx context.Context, target app.GitRepositoryTarget, branch, filePath, commitMessage, content string) error {
	if err := p.validateTarget(target); err != nil {
		return err
	}
	_, err := p.client.CreateOrUpdateFile(ctx, target.ProjectID, branch, filePath, commitMessage, content)
	return err
}

func (p *Provider) OpenMergeRequest(ctx context.Context, target app.GitRepositoryTarget, sourceBranch, targetBranch, title, description string) (int, string, error) {
	if err := p.validateTarget(target); err != nil {
		return 0, "", err
	}
	mergeRequest, err := p.client.OpenMergeRequest(ctx, target.ProjectID, sourceBranch, targetBranch, title, description)
	if err != nil {
		return 0, "", err
	}
	return mergeRequest.IID, mergeRequest.WebURL, nil
}

func (p *Provider) MergeRequest(ctx context.Context, target app.GitRepositoryTarget, sourceBranch, targetBranch, mergeCommitMessage string) (int, string, string, error) {
	if err := p.validateTarget(target); err != nil {
		return 0, "", "", err
	}
	mergeRequest, err := p.client.FindOpenMergeRequest(ctx, target.ProjectID, sourceBranch, targetBranch)
	if err != nil {
		return 0, "", "", err
	}
	merged, err := p.client.MergeMergeRequest(ctx, target.ProjectID, mergeRequest.IID, mergeRequest.SHA, mergeCommitMessage)
	if err != nil {
		return 0, "", "", err
	}
	return merged.IID, merged.WebURL, merged.MergeCommitSHA, nil
}

func (p *Provider) validateTarget(target app.GitRepositoryTarget) error {
	if strings.ToLower(strings.TrimSpace(target.Provider)) != app.RepositoryProviderGitLab {
		return fmt.Errorf("gitlab provider cannot handle provider %q", target.Provider)
	}
	if strings.TrimSpace(target.ProviderRef) != p.providerRef {
		return fmt.Errorf("gitlab providerRef %q cannot handle target providerRef %q", p.providerRef, target.ProviderRef)
	}
	if target.ProjectID <= 0 {
		return errors.New("gitlab repository target projectID must be greater than zero")
	}
	return nil
}

var _ app.GitProvider = (*Provider)(nil)
