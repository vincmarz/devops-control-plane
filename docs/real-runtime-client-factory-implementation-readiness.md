# Real Runtime Client Factory Implementation Readiness

## Purpose

This document records the implementation readiness status for phase 15.8.8 after adding concrete runtime client factory adapters for Kubernetes, Tekton and Argo CD.

The scope of this readiness milestone is intentionally conservative:

- concrete factories exist and are covered by unit tests;
- the factories live in the composition layer, not in `internal/app`;
- `internal/app` remains the contract and validation layer;
- `main.go` still does not wire the concrete factories;
- `EmptyRuntimeSecretValueLoader` remains the runtime default;
- no Kubernetes Secret values are read at runtime;
- staging and production remain disabled.

This phase prepares the project for future real multi-cluster client enablement without changing the validated `ocp-dev` / current-cluster runtime behavior.

## Validated Implementation Commits

The following implementation commits are part of this readiness baseline:

```text
50eab45 Document real runtime client factory design
b87cd55 Add Kubernetes runtime client factory adapter
7d02979 Add Tekton runtime client factory adapter
65b99d2 Add Argo CD runtime client factory adapter
```

At the end of the implementation steps, `go test ./...` was green and the working tree was clean.

## Layering Baseline

The implementation follows the design decision documented in:

```text
docs/real-runtime-client-factory-implementation-design.md
```

The dependency direction remains:

```text
cmd/devops-control-plane -> internal/app
cmd/devops-control-plane -> internal/adapters/*
internal/app             -> no dependency on internal/adapters/*
```

This keeps the application layer free from concrete adapter dependencies while allowing the composition root to construct real adapter-backed runtime clients in the future.

## Implemented Factory Adapters

### Kubernetes runtime client factory adapter

Implemented files:

```text
cmd/devops-control-plane/kubernetes_runtime_client_factory_adapter.go
cmd/devops-control-plane/kubernetes_runtime_client_factory_adapter_test.go
```

The Kubernetes factory adapter satisfies:

```text
app.KubernetesRuntimeClientFactory
```

Readiness behavior:

- validates `app.KubernetesRuntimeClientFactoryRequest`;
- requires a configured API URL for the target cluster;
- resolves `RuntimeSecretValueKubernetesToken` from `RuntimeSecretValueSet`;
- builds a concrete Kubernetes adapter client through `kubernetesadapter.New(...)`;
- returns an `app.KubernetesRuntimeEvidenceClient`;
- rejects kubeconfig-based inputs until explicit support is implemented;
- rejects raw CA values loaded from Secret until a safe CA handling strategy is implemented.

The factory is token-based only in this readiness phase.

### Tekton runtime client factory adapter

Implemented files:

```text
cmd/devops-control-plane/tekton_runtime_client_factory_adapter.go
cmd/devops-control-plane/tekton_runtime_client_factory_adapter_test.go
```

The Tekton factory adapter satisfies:

```text
app.TektonRuntimeClientFactory
```

Readiness behavior:

- validates `app.TektonRuntimeClientFactoryRequest`;
- requires a configured API URL for the target cluster;
- resolves `RuntimeSecretValueKubernetesToken` from `RuntimeSecretValueSet`;
- builds a concrete Tekton adapter client through `tektonadapter.New(...)`;
- wraps the adapter with `currentTektonRuntimeClient`;
- returns an `app.TektonRuntimeClient`;
- rejects kubeconfig-based inputs until explicit support is implemented;
- rejects raw CA values loaded from Secret until a safe CA handling strategy is implemented.

The factory reuses the existing Tekton runtime adapter wrapper so PipelineRun request and response mapping remain centralized.

### Argo CD runtime client factory adapter

Implemented files:

```text
cmd/devops-control-plane/argocd_runtime_client_factory_adapter.go
cmd/devops-control-plane/argocd_runtime_client_factory_adapter_test.go
```

The Argo CD factory adapter satisfies:

```text
app.ArgoCDRuntimeClientFactory
```

Readiness behavior:

- validates `app.ArgoCDRuntimeClientFactoryRequest`;
- resolves `RuntimeSecretValueArgoCDToken` from `RuntimeSecretValueSet`;
- uses an Argo CD BaseURL from factory configuration or, when `BaseURLKey` is configured, from `RuntimeSecretValueSet`;
- builds a concrete Argo CD adapter client through `argocdadapter.New(...)`;
- wraps the adapter with `currentArgoCDRuntimeClient`;
- returns an `app.ArgoCDRuntimeClient`;
- rejects raw CA values loaded from Secret until a safe CA handling strategy is implemented.

The factory reuses the existing Argo CD runtime adapter wrapper so deployment status response mapping remains centralized.

## Unit Test Coverage

Each factory adapter includes unit tests for readiness and fail-closed behavior.

Common coverage includes:

- compile-time interface assertion against the relevant `app.*RuntimeClientFactory` contract;
- valid token-based client construction without network calls;
- invalid request rejection through app-layer validation;
- missing Secret value failure;
- missing cluster endpoint configuration failure;
- unsupported raw CA rejection;
- unsupported kubeconfig rejection for Kubernetes-family clients;
- normalized cluster key handling.

Argo CD-specific coverage also includes:

- BaseURL from factory configuration;
- BaseURL from `RuntimeSecretValueSet` when `BaseURLKey` is configured;
- missing BaseURL Secret value failure when `BaseURLKey` is configured.

The tests do not read Kubernetes Secrets and do not perform live network calls.

## Security Guardrails

The readiness implementation preserves the established security guardrails:

- no Secret value is logged;
- no Secret value is serialized;
- no Secret value is included in errors;
- no real Kubernetes Secret is read;
- no runtime Secret loader is enabled;
- no factory is wired into `main.go`;
- staging and production remain disabled;
- kubeconfig values are rejected until explicit support is designed;
- raw CA values from Secret are rejected until safe CA handling is designed.

## Runtime Enablement Status

The concrete factories are present but not runtime-enabled.

Current runtime wiring remains conservative:

```text
main.go uses factory-aware registries
factory arguments remain nil
RuntimeSecretValueLoader remains EmptyRuntimeSecretValueLoader
```

Therefore:

- `ocp-dev` / current-cluster continues to use the already configured runtime clients;
- factory fallback remains fail-closed;
- Secret-configured non-current clusters are not activated;
- staging and production remain disabled.

## Explicit Non-Goals for This Phase

This phase does not:

- wire concrete factories into `main.go`;
- replace `EmptyRuntimeSecretValueLoader`;
- enable any real Secret reader;
- enable staging;
- enable production;
- support kubeconfig-based client construction;
- support raw CA material loaded from Secret;
- create temporary CA files;
- introduce new runtime configuration variables for real factory enablement.

## Readiness Conclusion

The repository is ready for a future controlled enablement phase where concrete factories may be wired behind explicit configuration and allow-listed Secret loading.

Readiness achieved:

- factory contracts already exist in `internal/app`;
- factory-aware registries are already wired conservatively in `main.go`;
- concrete Kubernetes, Tekton and Argo CD factory adapters exist in the composition layer;
- unit tests validate construction and fail-closed paths;
- the current runtime baseline remains unchanged and safe.

## Recommended Next Step

The recommended next step is runtime non-regression validation with factories still disabled:

```text
15.8.8.7 — Runtime non-regression validation with factories still disabled
```

The purpose of that validation is to prove that adding the concrete factory adapter code did not change runtime behavior because the factories are not yet wired and the runtime Secret value loader remains empty.

Expected validation outcomes:

```text
/readyz -> 200
create dev ChangeRequest -> 201 with configured RBAC group and complete payload
collect-evidence -> 202
validate -> 202
check-validation -> 202 or a valid in-progress state
check-deployment -> 202 with Synced/Healthy state
staging -> 422 disabled
production -> 422 disabled
```
