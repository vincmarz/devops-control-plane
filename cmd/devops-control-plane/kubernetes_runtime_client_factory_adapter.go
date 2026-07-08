package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	kubernetesadapter "github.com/vincmarz/devops-control-plane/internal/adapters/kubernetes"
	"github.com/vincmarz/devops-control-plane/internal/app"
)

var errKubernetesRuntimeClientFactoryAPIURLNotConfigured = errors.New("kubernetes runtime client factory adapter: api URL is not configured for the requested cluster")

var errKubernetesRuntimeClientFactoryKubeconfigUnsupported = errors.New("kubernetes runtime client factory adapter: kubeconfig secret value is not supported yet")

var errKubernetesRuntimeClientFactoryRawCAUnsupported = errors.New("kubernetes runtime client factory adapter: raw CA secret value is not supported yet")

type kubernetesRuntimeClientFactoryAdapterConfig struct {
	ClusterAPIURLs map[string]string
	TimeoutSeconds int
	InsecureTLS    bool
	CAFiles        map[string]string
}

// kubernetesRuntimeClientFactoryAdapter builds concrete Kubernetes runtime
// evidence clients from validated app-layer factory requests.
//
// The adapter is intentionally local to the composition package so internal/app
// remains a contract layer and does not import concrete adapters.
//
// This factory is not wired in main.go in phase 15.8.8.3. It is introduced with
// unit tests only, preserving the current EmptyRuntimeSecretValueLoader
// fail-closed runtime baseline.
type kubernetesRuntimeClientFactoryAdapter struct {
	clusterAPIURLs map[string]string
	timeoutSeconds int
	insecureTLS    bool
	caFiles        map[string]string
}

func newKubernetesRuntimeClientFactoryAdapter(config kubernetesRuntimeClientFactoryAdapterConfig) kubernetesRuntimeClientFactoryAdapter {
	return kubernetesRuntimeClientFactoryAdapter{
		clusterAPIURLs: normalizeStringMap(config.ClusterAPIURLs),
		timeoutSeconds: config.TimeoutSeconds,
		insecureTLS:    config.InsecureTLS,
		caFiles:        normalizeStringMap(config.CAFiles),
	}
}

func (f kubernetesRuntimeClientFactoryAdapter) BuildKubernetesRuntimeEvidenceClient(_ context.Context, request app.KubernetesRuntimeClientFactoryRequest) (app.KubernetesRuntimeEvidenceClient, error) {
	if err := app.ValidateKubernetesRuntimeClientFactoryRequest(request); err != nil {
		return nil, err
	}
	if request.SecretRefs.Kubernetes == nil {
		return nil, errors.New("kubernetes runtime client factory adapter: kubernetes secret reference is required")
	}
	if strings.TrimSpace(request.SecretRefs.Kubernetes.KubeconfigKey) != "" {
		return nil, errKubernetesRuntimeClientFactoryKubeconfigUnsupported
	}
	if strings.TrimSpace(request.SecretRefs.Kubernetes.CAKey) != "" {
		return nil, errKubernetesRuntimeClientFactoryRawCAUnsupported
	}

	token, err := request.SecretValues.Resolve(app.RuntimeSecretValueKubernetesToken)
	if err != nil {
		return nil, err
	}

	clusterName := normalizeFactoryClusterName(request.Target.ClusterName)
	apiURL := strings.TrimSpace(f.clusterAPIURLs[clusterName])
	if apiURL == "" {
		return nil, fmt.Errorf("%w: %s", errKubernetesRuntimeClientFactoryAPIURLNotConfigured, clusterName)
	}

	client, err := kubernetesadapter.New(kubernetesadapter.Config{
		APIURL:         apiURL,
		Token:          token,
		TimeoutSeconds: f.timeoutSeconds,
		InsecureTLS:    f.insecureTLS,
		CAFile:         strings.TrimSpace(f.caFiles[clusterName]),
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func normalizeStringMap(input map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range input {
		normalizedKey := normalizeFactoryClusterName(key)
		if normalizedKey == "" {
			continue
		}
		out[normalizedKey] = strings.TrimSpace(value)
	}
	return out
}

func normalizeFactoryClusterName(clusterName string) string {
	return strings.ToLower(strings.TrimSpace(clusterName))
}
