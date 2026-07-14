package app

import (
	"errors"
	"fmt"
	"strings"
)

const (
	GitOpsConsumerArgoCD = "argocd"
	GitOpsConsumerTekton = "tekton"
)

type GitOpsRepositoryTarget struct {
	Provider      string
	ProviderRef   string
	ProjectID     int
	ProjectPath   string
	RepositoryURL string
	DefaultBranch string
	ConsumedBy    []string
}

type GitOpsBindingResolverFunc func(applicationName string) (RepositoryBinding, error)

type GitOpsRepositoryTargetResolver struct{ resolveBinding GitOpsBindingResolverFunc }

func NewGitOpsRepositoryTargetResolver(resolveBinding GitOpsBindingResolverFunc) GitOpsRepositoryTargetResolver {
	return GitOpsRepositoryTargetResolver{resolveBinding: resolveBinding}
}

func (r GitOpsRepositoryTargetResolver) Resolve(applicationName, consumer string) (GitOpsRepositoryTarget, error) {
	applicationName = strings.TrimSpace(applicationName)
	consumer = strings.ToLower(strings.TrimSpace(consumer))
	if applicationName == "" {
		return GitOpsRepositoryTarget{}, errors.New("application name is required for GitOps repository resolution")
	}
	if consumer == "" {
		return GitOpsRepositoryTarget{}, errors.New("GitOps consumer is required")
	}
	if r.resolveBinding == nil {
		return GitOpsRepositoryTarget{}, errors.New("GitOps binding resolver is not configured")
	}
	binding, err := r.resolveBinding(applicationName)
	if err != nil {
		return GitOpsRepositoryTarget{}, err
	}
	return NewGitOpsRepositoryTarget(binding, consumer)
}

func NewGitOpsRepositoryTarget(binding RepositoryBinding, consumer string) (GitOpsRepositoryTarget, error) {
	consumer = strings.ToLower(strings.TrimSpace(consumer))
	if consumer == "" {
		return GitOpsRepositoryTarget{}, errors.New("GitOps consumer is required")
	}
	target := GitOpsRepositoryTarget{
		Provider:      strings.ToLower(strings.TrimSpace(binding.Provider)),
		ProviderRef:   strings.TrimSpace(binding.ProviderRef),
		ProjectID:     binding.ProjectID,
		ProjectPath:   strings.TrimSpace(binding.ProjectPath),
		RepositoryURL: strings.TrimSpace(binding.RepositoryURL),
		DefaultBranch: strings.TrimSpace(binding.DefaultBranch),
		ConsumedBy:    normalizeGitOpsConsumers(binding.ConsumedBy),
	}
	if strings.ToLower(strings.TrimSpace(binding.Role)) != RepositoryRoleGitOps {
		return GitOpsRepositoryTarget{}, fmt.Errorf("repository binding role %q is not GitOps", binding.Role)
	}
	if target.Provider == "" {
		return GitOpsRepositoryTarget{}, errors.New("GitOps repository target provider is required")
	}
	if target.ProviderRef == "" {
		return GitOpsRepositoryTarget{}, errors.New("GitOps repository target providerRef is required")
	}
	if target.ProjectPath == "" {
		return GitOpsRepositoryTarget{}, errors.New("GitOps repository target projectPath is required")
	}
	if target.RepositoryURL == "" {
		return GitOpsRepositoryTarget{}, errors.New("GitOps repository target repositoryURL is required")
	}
	if target.DefaultBranch == "" {
		return GitOpsRepositoryTarget{}, errors.New("GitOps repository target defaultBranch is required")
	}
	if !target.SupportsConsumer(consumer) {
		return GitOpsRepositoryTarget{}, fmt.Errorf("GitOps repository target is not configured for consumer %q", consumer)
	}
	return target, nil
}

func (t GitOpsRepositoryTarget) SupportsConsumer(consumer string) bool {
	consumer = strings.ToLower(strings.TrimSpace(consumer))
	for _, configured := range t.ConsumedBy {
		if configured == consumer {
			return true
		}
	}
	return false
}

func normalizeGitOpsConsumers(consumers []string) []string {
	normalized := make([]string, 0, len(consumers))
	seen := map[string]bool{}
	for _, consumer := range consumers {
		consumer = strings.ToLower(strings.TrimSpace(consumer))
		if consumer == "" || seen[consumer] {
			continue
		}
		seen[consumer] = true
		normalized = append(normalized, consumer)
	}
	return normalized
}

func NormalizeGitRepositoryURL(repositoryURL string) string {
	normalized := strings.ToLower(strings.TrimSpace(repositoryURL))
	normalized = strings.TrimRight(normalized, "/")
	normalized = strings.TrimSuffix(normalized, ".git")
	return normalized
}

func (t GitOpsRepositoryTarget) MatchesRepositoryURL(observedRepositoryURL string) bool {
	configured := NormalizeGitRepositoryURL(t.RepositoryURL)
	observed := NormalizeGitRepositoryURL(observedRepositoryURL)
	return configured != "" && configured == observed
}
