package app

import (
	"errors"
	"fmt"
	"strings"
)

// TechnicalRuntimeTarget describes the resolved runtime target for technical
// operations associated with a ChangeRequest target environment.
//
// The model is intentionally read-only in this phase. It prepares the
// application layer for future multi-cluster client selection without changing
// the current runtime adapters or enabling disabled environments.
type TechnicalRuntimeTarget struct {
	TargetEnvironment      string
	EnvironmentName        string
	EnvironmentDisplayName string
	ClusterName            string
	ClusterDisplayName     string
	ClusterEnabled         bool
	KubernetesNamespace    string
	TektonNamespace        string
	TektonPipelineName     string
	ArgoCDApplicationName  string
	GitTargetBranch        string
}

// TechnicalRuntimeTargetResolver resolves the technical runtime target for a
// ChangeRequest target environment by reusing the EnvironmentClusterResolver.
type TechnicalRuntimeTargetResolver struct {
	environmentClusterResolver EnvironmentClusterResolver
	defaultTektonPipelineName  string
}

// DefaultTechnicalRuntimeTargetResolver builds a resolver from the runtime
// Environment Catalog and Cluster Registry.
func DefaultTechnicalRuntimeTargetResolver(defaultTektonPipelineName string) TechnicalRuntimeTargetResolver {
	return NewTechnicalRuntimeTargetResolver(DefaultEnvironmentClusterResolver(), defaultTektonPipelineName)
}

// NewTechnicalRuntimeTargetResolver returns a resolver using explicit
// dependencies. Tests should prefer this constructor to avoid relying on runtime
// files or environment variables.
func NewTechnicalRuntimeTargetResolver(environmentClusterResolver EnvironmentClusterResolver, defaultTektonPipelineName string) TechnicalRuntimeTargetResolver {
	return TechnicalRuntimeTargetResolver{
		environmentClusterResolver: environmentClusterResolver,
		defaultTektonPipelineName:  strings.TrimSpace(defaultTektonPipelineName),
	}
}

// Resolve resolves and validates the technical runtime target for an enabled
// target environment.
//
// Disabled environments continue to be rejected. This preserves the current
// safety baseline: dev is operational, while staging and production remain
// configured but disabled.
func (r TechnicalRuntimeTargetResolver) Resolve(targetEnvironment string) (TechnicalRuntimeTarget, error) {
	resolution, err := r.environmentClusterResolver.ResolveTechnicalActionTarget(targetEnvironment)
	if err != nil {
		return TechnicalRuntimeTarget{}, err
	}

	environment := resolution.Environment
	cluster := resolution.Cluster

	if !cluster.Enabled {
		return TechnicalRuntimeTarget{}, fmt.Errorf("cluster %q is currently disabled", cluster.Name)
	}

	target := TechnicalRuntimeTarget{
		TargetEnvironment:      environment.Name,
		EnvironmentName:        environment.Name,
		EnvironmentDisplayName: strings.TrimSpace(environment.DisplayName),
		ClusterName:            cluster.Name,
		ClusterDisplayName:     strings.TrimSpace(cluster.DisplayName),
		ClusterEnabled:         cluster.Enabled,
		KubernetesNamespace:    strings.TrimSpace(environment.KubernetesNamespace),
		TektonNamespace:        strings.TrimSpace(environment.TektonNamespace),
		TektonPipelineName:     r.defaultTektonPipelineName,
		ArgoCDApplicationName:  strings.TrimSpace(environment.ArgoCDApplicationName),
		GitTargetBranch:        strings.TrimSpace(environment.GitTargetBranch),
	}

	if target.EnvironmentDisplayName == "" {
		target.EnvironmentDisplayName = target.EnvironmentName
	}
	if target.ClusterDisplayName == "" {
		target.ClusterDisplayName = target.ClusterName
	}
	if target.TektonNamespace == "" {
		target.TektonNamespace = target.KubernetesNamespace
	}

	if err := target.Validate(); err != nil {
		return TechnicalRuntimeTarget{}, err
	}

	return target, nil
}

// Validate checks that the resolved technical target contains the minimum
// metadata required by the existing technical integrations.
func (t TechnicalRuntimeTarget) Validate() error {
	if strings.TrimSpace(t.TargetEnvironment) == "" {
		return errors.New("target environment is required")
	}
	if strings.TrimSpace(t.EnvironmentName) == "" {
		return errors.New("environment name is required")
	}
	if strings.TrimSpace(t.ClusterName) == "" {
		return errors.New("cluster name is required")
	}
	if !t.ClusterEnabled {
		return fmt.Errorf("cluster %q is currently disabled", t.ClusterName)
	}
	if strings.TrimSpace(t.KubernetesNamespace) == "" {
		return fmt.Errorf("environment %q kubernetesNamespace is required for technical runtime target", t.EnvironmentName)
	}
	if strings.TrimSpace(t.TektonNamespace) == "" {
		return fmt.Errorf("environment %q tektonNamespace is required for technical runtime target", t.EnvironmentName)
	}
	if strings.TrimSpace(t.TektonPipelineName) == "" {
		return errors.New("tekton pipeline name is required for technical runtime target")
	}
	if strings.TrimSpace(t.ArgoCDApplicationName) == "" {
		return fmt.Errorf("environment %q argocdApplicationName is required for technical runtime target", t.EnvironmentName)
	}
	if strings.TrimSpace(t.GitTargetBranch) == "" {
		return fmt.Errorf("environment %q gitTargetBranch is required for technical runtime target", t.EnvironmentName)
	}

	return nil
}
