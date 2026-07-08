package main

import (
	"github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/config"
)

// buildRuntimeSecretValueLoader gates construction of the runtime Secret value
// loader behind the disabled-by-default RuntimeSecretLoaderEnabled flag.
//
// When the flag is disabled, the function always returns
// app.EmptyRuntimeSecretValueLoader and never exposes the supplied getter.
//
// When the flag is enabled, the supplied KubernetesSecretGetter must be present;
// otherwise construction fails closed through buildRuntimeKubernetesSecretGetter.
//
// This helper is not wired in main.go in phase 15.8.9.5. It only prepares the
// composition boundary for a later conditional wiring phase and does not read
// any Kubernetes Secret.
func buildRuntimeSecretValueLoader(
	cfg config.Config,
	loaderConfig app.KubernetesSecretValueLoaderConfig,
	getter app.KubernetesSecretGetter,
) (app.RuntimeSecretValueLoader, error) {
	if !cfg.RuntimeSecretLoaderEnabled {
		return app.EmptyRuntimeSecretValueLoader{}, nil
	}

	builtGetter, err := buildRuntimeKubernetesSecretGetter(cfg, getter)
	if err != nil {
		return nil, err
	}

	return app.NewAllowListKubernetesSecretValueLoader(loaderConfig, builtGetter), nil
}
