package main

import (
	"errors"

	"github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/config"
)

var errRuntimeKubernetesSecretGetterNotConfigured = errors.New("runtime kubernetes secret getter is not configured")

// buildRuntimeKubernetesSecretGetter gates construction of the runtime
// KubernetesSecretGetter behind the disabled-by-default runtime Secret loader
// flag.
//
// This helper intentionally does not read any Kubernetes Secret. It only
// decides whether a pre-built getter may be exposed to the future allow-list
// runtime Secret value loader.
//
// The helper is not wired in main.go in phase 15.8.9.4. It is introduced as a
// construction-readiness boundary with unit tests only.
func buildRuntimeKubernetesSecretGetter(cfg config.Config, getter app.KubernetesSecretGetter) (app.KubernetesSecretGetter, error) {
	if !cfg.RuntimeSecretLoaderEnabled {
		return nil, nil
	}
	if getter == nil {
		return nil, errRuntimeKubernetesSecretGetterNotConfigured
	}
	return getter, nil
}
