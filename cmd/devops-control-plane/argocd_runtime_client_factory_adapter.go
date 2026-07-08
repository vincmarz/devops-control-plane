package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	argocdadapter "github.com/vincmarz/devops-control-plane/internal/adapters/argocd"
	"github.com/vincmarz/devops-control-plane/internal/app"
)

var errArgoCDRuntimeClientFactoryBaseURLNotConfigured = errors.New("argocd runtime client factory adapter: base URL is not configured for the requested cluster")

var errArgoCDRuntimeClientFactoryRawCAUnsupported = errors.New("argocd runtime client factory adapter: raw CA secret value is not supported yet")

type argoCDRuntimeClientFactoryAdapterConfig struct {
	ClusterBaseURLs map[string]string
	TimeoutSeconds  int
	InsecureTLS     bool
	CAFiles         map[string]string
}

// argoCDRuntimeClientFactoryAdapter builds concrete Argo CD runtime clients
// from validated app-layer factory requests.
//
// This factory is intentionally local to the composition package so
// internal/app remains a contract layer and does not import concrete adapters.
//
// This factory is not wired in main.go in phase 15.8.8.5. It is introduced
// with unit tests only, preserving the current EmptyRuntimeSecretValueLoader
// fail-closed runtime baseline.
type argoCDRuntimeClientFactoryAdapter struct {
	clusterBaseURLs map[string]string
	timeoutSeconds  int
	insecureTLS     bool
	caFiles         map[string]string
}

func newArgoCDRuntimeClientFactoryAdapter(config argoCDRuntimeClientFactoryAdapterConfig) argoCDRuntimeClientFactoryAdapter {
	return argoCDRuntimeClientFactoryAdapter{
		clusterBaseURLs: normalizeStringMap(config.ClusterBaseURLs),
		timeoutSeconds:  config.TimeoutSeconds,
		insecureTLS:     config.InsecureTLS,
		caFiles:         normalizeStringMap(config.CAFiles),
	}
}

func (f argoCDRuntimeClientFactoryAdapter) BuildArgoCDRuntimeClient(_ context.Context, request app.ArgoCDRuntimeClientFactoryRequest) (app.ArgoCDRuntimeClient, error) {
	if err := app.ValidateArgoCDRuntimeClientFactoryRequest(request); err != nil {
		return nil, err
	}
	if request.SecretRefs.ArgoCD == nil {
		return nil, errors.New("argocd runtime client factory adapter: argocd secret reference is required")
	}
	if strings.TrimSpace(request.SecretRefs.ArgoCD.CAKey) != "" {
		return nil, errArgoCDRuntimeClientFactoryRawCAUnsupported
	}

	token, err := request.SecretValues.Resolve(app.RuntimeSecretValueArgoCDToken)
	if err != nil {
		return nil, err
	}

	clusterName := normalizeFactoryClusterName(request.Target.ClusterName)
	baseURL := strings.TrimSpace(f.clusterBaseURLs[clusterName])
	if strings.TrimSpace(request.SecretRefs.ArgoCD.BaseURLKey) != "" {
		baseURLFromSecret, err := request.SecretValues.Resolve(app.RuntimeSecretValueArgoCDBaseURL)
		if err != nil {
			return nil, err
		}
		baseURL = strings.TrimSpace(baseURLFromSecret)
	}
	if baseURL == "" {
		return nil, fmt.Errorf("%w: %s", errArgoCDRuntimeClientFactoryBaseURLNotConfigured, clusterName)
	}

	client, err := argocdadapter.New(argocdadapter.Config{
		BaseURL:        baseURL,
		AuthToken:      token,
		TimeoutSeconds: f.timeoutSeconds,
		InsecureTLS:    f.insecureTLS,
		CAFile:         strings.TrimSpace(f.caFiles[clusterName]),
	})
	if err != nil {
		return nil, err
	}
	return currentArgoCDRuntimeClient{client: client}, nil
}
