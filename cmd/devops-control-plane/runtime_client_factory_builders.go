package main

import (
	"errors"
	"strings"

	"github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/config"
)

var errRuntimeKubernetesClientFactoryAPIURLNotConfigured = errors.New("runtime kubernetes client factory api URL is not configured")

var errRuntimeTektonClientFactoryAPIURLNotConfigured = errors.New("runtime tekton client factory api URL is not configured")

var errRuntimeArgoCDClientFactoryBaseURLNotConfigured = errors.New("runtime argocd client factory base URL is not configured")

// buildRuntimeKubernetesClientFactory gates construction of the concrete
// Kubernetes runtime client factory behind the disabled-by-default global and
// capability-specific runtime factory flags.
//
// The helper does not read Secrets and is not wired into main.go in phase
// 15.8.9.8. It only prepares the controlled construction boundary for a later
// main wiring phase.
func buildRuntimeKubernetesClientFactory(cfg config.Config) (app.KubernetesRuntimeClientFactory, error) {
	if !cfg.RuntimeClientFactoriesEnabled || !cfg.RuntimeClientFactoryKubernetesEnabled {
		return nil, nil
	}
	apiURL := strings.TrimSpace(cfg.KubernetesAPIURL)
	if apiURL == "" {
		return nil, errRuntimeKubernetesClientFactoryAPIURLNotConfigured
	}
	return newKubernetesRuntimeClientFactoryAdapter(kubernetesRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{"ocp-dev": apiURL},
		TimeoutSeconds: cfg.TektonTimeoutSeconds,
		InsecureTLS:    cfg.KubernetesInsecureTLS,
		CAFiles:        map[string]string{"ocp-dev": strings.TrimSpace(cfg.KubernetesCAFile)},
	}), nil
}

// buildRuntimeTektonClientFactory gates construction of the concrete Tekton
// runtime client factory behind the disabled-by-default global and
// capability-specific runtime factory flags.
//
// Tekton uses the Kubernetes API of the runtime cluster, so the current safe
// readiness builder uses the Kubernetes API URL and TLS settings from config.
func buildRuntimeTektonClientFactory(cfg config.Config) (app.TektonRuntimeClientFactory, error) {
	if !cfg.RuntimeClientFactoriesEnabled || !cfg.RuntimeClientFactoryTektonEnabled {
		return nil, nil
	}
	apiURL := strings.TrimSpace(cfg.KubernetesAPIURL)
	if apiURL == "" {
		return nil, errRuntimeTektonClientFactoryAPIURLNotConfigured
	}
	return newTektonRuntimeClientFactoryAdapter(tektonRuntimeClientFactoryAdapterConfig{
		ClusterAPIURLs: map[string]string{"ocp-dev": apiURL},
		TimeoutSeconds: cfg.TektonTimeoutSeconds,
		InsecureTLS:    cfg.KubernetesInsecureTLS,
		CAFiles:        map[string]string{"ocp-dev": strings.TrimSpace(cfg.KubernetesCAFile)},
	}), nil
}

// buildRuntimeArgoCDClientFactory gates construction of the concrete Argo CD
// runtime client factory behind the disabled-by-default global and
// capability-specific runtime factory flags.
func buildRuntimeArgoCDClientFactory(cfg config.Config) (app.ArgoCDRuntimeClientFactory, error) {
	if !cfg.RuntimeClientFactoriesEnabled || !cfg.RuntimeClientFactoryArgoCDEnabled {
		return nil, nil
	}
	baseURL := strings.TrimSpace(cfg.ArgoCDBaseURL)
	if baseURL == "" {
		return nil, errRuntimeArgoCDClientFactoryBaseURLNotConfigured
	}
	return newArgoCDRuntimeClientFactoryAdapter(argoCDRuntimeClientFactoryAdapterConfig{
		ClusterBaseURLs: map[string]string{"ocp-dev": baseURL},
		TimeoutSeconds:  cfg.ArgoCDTimeoutSeconds,
		InsecureTLS:     cfg.ArgoCDInsecureTLS,
		CAFiles:         map[string]string{"ocp-dev": strings.TrimSpace(cfg.ArgoCDCAFile)},
	}), nil
}
