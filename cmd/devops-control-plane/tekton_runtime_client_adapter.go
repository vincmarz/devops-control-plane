package main

import (
	"context"
	"errors"

	"github.com/vincmarz/devops-control-plane/internal/adapters/tekton"
	"github.com/vincmarz/devops-control-plane/internal/app"
)

type currentTektonRuntimeClient struct {
	client *tekton.Client
}

func (c currentTektonRuntimeClient) CreatePipelineRun(ctx context.Context, request app.TektonRuntimePipelineRunRequest) (app.TektonRuntimePipelineRunRef, error) {
	if c.client == nil {
		return app.TektonRuntimePipelineRunRef{}, errors.New("current Tekton runtime client is not configured")
	}
	ref, err := c.client.CreatePipelineRun(ctx, tekton.CreatePipelineRunRequest{
		Namespace:          request.Namespace,
		PipelineName:       request.PipelineName,
		GenerateName:       request.GenerateName,
		ChangeNumber:       request.ChangeNumber,
		GitURL:             request.GitURL,
		GitRevision:        request.GitRevision,
		ValidationPath:     request.ValidationPath,
		ServiceAccountName: request.ServiceAccountName,
		Image:              request.Image,
		WorkspacePVC:       request.WorkspacePVC,
		DockerConfigSecret: request.DockerConfigSecret,
	})
	if err != nil {
		return app.TektonRuntimePipelineRunRef{}, err
	}
	return app.TektonRuntimePipelineRunRef{Name: ref.Name, Namespace: ref.Namespace, UID: ref.UID}, nil
}

func (c currentTektonRuntimeClient) FindLatestPipelineRunByChange(ctx context.Context, namespace string, changeNumber string) (app.TektonValidationResult, error) {
	return app.TektonValidationResult{}, errors.New("current Tekton runtime client check-validation provider is not wired yet")
}

func (c currentTektonRuntimeClient) ListTaskRunsByPipelineRun(ctx context.Context, namespace string, pipelineRunName string) ([]app.TektonTaskRunResult, error) {
	return nil, errors.New("current Tekton runtime client taskrun provider is not wired yet")
}
