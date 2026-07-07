package main

import (
	"context"
	"encoding/json"
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
	var result app.ArgoCDDeploymentResult
	content, err := json.Marshal(argoApp)
	if err != nil {
		return app.ArgoCDDeploymentResult{}, err
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return app.ArgoCDDeploymentResult{}, err
	}
	if result.ApplicationName == "" {
		result.ApplicationName = applicationName
	}
	return result, nil
}
