package main

import (
	"context"
	"encoding/json"
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
	if c.client == nil {
		return app.TektonValidationResult{}, errors.New("current Tekton runtime client is not configured")
	}
	status, err := c.client.FindLatestPipelineRunByChange(ctx, namespace, changeNumber)
	if err != nil {
		return app.TektonValidationResult{}, err
	}
	var result app.TektonValidationResult
	content, err := json.Marshal(status)
	if err != nil {
		return app.TektonValidationResult{}, err
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return app.TektonValidationResult{}, err
	}
	if result.PipelineRunName == "" {
		result.PipelineRunName = status.Name
	}
	return result, nil
}

func (c currentTektonRuntimeClient) ListTaskRunsByPipelineRun(ctx context.Context, namespace string, pipelineRunName string) ([]app.TektonTaskRunResult, error) {
	if c.client == nil {
		return nil, errors.New("current Tekton runtime client is not configured")
	}
	taskRuns, err := c.client.ListTaskRunsByPipelineRun(ctx, namespace, pipelineRunName)
	if err != nil {
		return nil, err
	}
	var result []app.TektonTaskRunResult
	content, err := json.Marshal(taskRuns)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}
	return result, nil
}
