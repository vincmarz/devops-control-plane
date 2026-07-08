package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tektonadapter "github.com/vincmarz/devops-control-plane/internal/adapters/tekton"
	"github.com/vincmarz/devops-control-plane/internal/app"
)

var errTektonRuntimeClientFactoryAPIURLNotConfigured = errors.New("tekton runtime client factory adapter: api URL is not configured for the requested cluster")

var errTektonRuntimeClientFactoryKubeconfigUnsupported = errors.New("tekton runtime client factory adapter: kubeconfig secret value is not supported yet")

var errTektonRuntimeClientFactoryRawCAUnsupported = errors.New("tekton runtime client factory adapter: raw CA secret value is not supported yet")

type tektonRuntimeClientFactoryAdapterConfig struct {
	ClusterAPIURLs map[string]string
	TimeoutSeconds int
	InsecureTLS    bool
	CAFiles        map[string]string
}

// tektonRuntimeClientFactoryAdapter builds concrete Tekton runtime clients from
// validated app-layer factory requests.
//
// Tekton is accessed through the Kubernetes API of the target cluster, so this
// factory uses the same cluster API URL and Kubernetes-family credential model
// used by the Kubernetes runtime client factory adapter.
//
// This factory is not wired in main.go in phase 15.8.8.4. It is introduced with
// unit tests only, preserving the current EmptyRuntimeSecretValueLoader
// fail-closed runtime baseline.
type tektonRuntimeClientFactoryAdapter struct {
	clusterAPIURLs map[string]string
	timeoutSeconds int
	insecureTLS    bool
	caFiles        map[string]string
}

func newTektonRuntimeClientFactoryAdapter(config tektonRuntimeClientFactoryAdapterConfig) tektonRuntimeClientFactoryAdapter {
	return tektonRuntimeClientFactoryAdapter{
		clusterAPIURLs: normalizeStringMap(config.ClusterAPIURLs),
		timeoutSeconds: config.TimeoutSeconds,
		insecureTLS:    config.InsecureTLS,
		caFiles:        normalizeStringMap(config.CAFiles),
	}
}

func (f tektonRuntimeClientFactoryAdapter) BuildTektonRuntimeClient(_ context.Context, request app.TektonRuntimeClientFactoryRequest) (app.TektonRuntimeClient, error) {
	if err := app.ValidateTektonRuntimeClientFactoryRequest(request); err != nil {
		return nil, err
	}
	if request.SecretRefs.Tekton == nil {
		return nil, errors.New("tekton runtime client factory adapter: tekton secret reference is required")
	}
	if strings.TrimSpace(request.SecretRefs.Tekton.KubeconfigKey) != "" {
		return nil, errTektonRuntimeClientFactoryKubeconfigUnsupported
	}
	if strings.TrimSpace(request.SecretRefs.Tekton.CAKey) != "" {
		return nil, errTektonRuntimeClientFactoryRawCAUnsupported
	}

	token, err := request.SecretValues.Resolve(app.RuntimeSecretValueKubernetesToken)
	if err != nil {
		return nil, err
	}

	clusterName := normalizeFactoryClusterName(request.Target.ClusterName)
	apiURL := strings.TrimSpace(f.clusterAPIURLs[clusterName])
	if apiURL == "" {
		return nil, fmt.Errorf("%w: %s", errTektonRuntimeClientFactoryAPIURLNotConfigured, clusterName)
	}

	client, err := tektonadapter.New(tektonadapter.Config{
		APIURL:         apiURL,
		Token:          token,
		TimeoutSeconds: f.timeoutSeconds,
		InsecureTLS:    f.insecureTLS,
		CAFile:         strings.TrimSpace(f.caFiles[clusterName]),
	})
	if err != nil {
		return nil, err
	}
	return currentTektonRuntimeClient{client: client}, nil
}
