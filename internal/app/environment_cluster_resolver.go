package app

import (
	"fmt"
	"strings"
)

type EnvironmentClusterResolution struct {
	TargetEnvironment string
	Environment       EnvironmentDefinition
	Cluster           ClusterDefinition
}

type EnvironmentClusterResolver struct {
	catalog  EnvironmentCatalog
	registry ClusterRegistry
}

func DefaultEnvironmentClusterResolver() EnvironmentClusterResolver {
	return NewEnvironmentClusterResolver(DefaultEnvironmentCatalog(), DefaultClusterRegistry())
}

func NewEnvironmentClusterResolver(catalog EnvironmentCatalog, registry ClusterRegistry) EnvironmentClusterResolver {
	return EnvironmentClusterResolver{catalog: catalog, registry: registry}
}

func (r EnvironmentClusterResolver) Resolve(targetEnvironment string) (EnvironmentClusterResolution, error) {
	targetEnvironment = normalizeEnvironmentName(targetEnvironment)
	if targetEnvironment == "" {
		targetEnvironment = r.catalog.DefaultEnvironment()
	}

	environment, ok := r.catalog.Resolve(targetEnvironment)
	if !ok {
		return EnvironmentClusterResolution{}, fmt.Errorf("targetEnvironment %q is not configured", targetEnvironment)
	}

	clusterName := strings.TrimSpace(environment.ClusterName)
	if clusterName == "" {
		return EnvironmentClusterResolution{}, fmt.Errorf("targetEnvironment %q does not define clusterName", targetEnvironment)
	}

	cluster, ok := r.registry.Resolve(clusterName)
	if !ok {
		return EnvironmentClusterResolution{}, fmt.Errorf("targetEnvironment %q references clusterName %q which is not configured", targetEnvironment, clusterName)
	}

	return EnvironmentClusterResolution{
		TargetEnvironment: targetEnvironment,
		Environment:       environment,
		Cluster:           cluster,
	}, nil
}

func (r EnvironmentClusterResolver) ResolveEnabledTarget(targetEnvironment string) (EnvironmentClusterResolution, error) {
	resolution, err := r.Resolve(targetEnvironment)
	if err != nil {
		return EnvironmentClusterResolution{}, err
	}

	if !resolution.Environment.Enabled {
		return EnvironmentClusterResolution{}, fmt.Errorf("targetEnvironment %q is currently disabled", resolution.TargetEnvironment)
	}
	if !resolution.Cluster.Enabled {
		return EnvironmentClusterResolution{}, fmt.Errorf("clusterName %q for targetEnvironment %q is currently disabled", resolution.Cluster.Name, resolution.TargetEnvironment)
	}

	return resolution, nil
}

func (r EnvironmentClusterResolver) ResolveTechnicalActionTarget(targetEnvironment string) (EnvironmentClusterResolution, error) {
	resolution, err := r.ResolveEnabledTarget(targetEnvironment)
	if err != nil {
		return EnvironmentClusterResolution{}, err
	}

	if !resolution.Environment.AllowTechnicalActions {
		return EnvironmentClusterResolution{}, fmt.Errorf("targetEnvironment %q does not allow technical actions", resolution.TargetEnvironment)
	}

	return resolution, nil
}
