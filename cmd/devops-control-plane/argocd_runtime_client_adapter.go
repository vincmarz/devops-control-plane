package main

import (
	"context"
	"errors"

	"github.com/vincmarz/devops-control-plane/internal/adapters/argocd"
	"github.com/vincmarz/devops-control-plane/internal/app"
)

type currentArgoCDRuntimeClient struct {
	client *argocd.Client
}

func (c currentArgoCDRuntimeClient) CheckDeployment(ctx context.Context, applicationName string) (app.ArgoCDDeploymentResult, error) {
	if c.client == nil {
		return app.ArgoCDDeploymentResult{}, errors.New("current Argo CD runtime client is not configured")
	}
	argoApp, err := c.client.GetApplication(ctx, applicationName)
	if err != nil {
		return app.ArgoCDDeploymentResult{}, err
	}
	return app.ArgoCDDeploymentResult{
		ApplicationName: argoApp.Name,
		Project:         argoApp.Project,
		SyncStatus:      argoApp.SyncStatus,
		HealthStatus:    argoApp.HealthStatus,
		Revision:        argoApp.CurrentRevision,
		RepositoryURL:   argoApp.RepoURL,
		TargetRevision:  argoApp.TargetRevision,
	}, nil
}
