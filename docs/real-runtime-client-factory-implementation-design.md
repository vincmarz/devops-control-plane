# Real Runtime Client Factory Implementation Design

## Purpose

This document records the design for phase 15.8.8.2: preparation of real runtime client factory implementations for Kubernetes, Tekton and Argo CD.

The design is intentionally conservative. It prepares the codebase for real multi-cluster runtime clients without enabling real Secret reads, without changing the current `main.go` runtime behavior and without enabling staging or production.

## Current Baseline

The current validated baseline is:

```text
main/origin main: aligned
latest documentation commit: ab06f3e Document factory-aware runtime provider validation
factory-aware main wiring commit: f657f44 Wire factory-aware runtime provider registries in main
```

Phase 15.8.7.4 validated the factory-aware main wiring on `ocp-dev` / current-cluster through the already configured runtime clients.

Validated runtime paths:

- Kubernetes collect-evidence;
- Tekton validate;
- Tekton check-validation;
- Argo CD check-deployment.

Current safety assumptions remain unchanged:

- `ocp-dev` / current-cluster remains the only active runtime baseline;
- staging and production remain disabled;
- `EmptyRuntimeSecretValueLoader` remains wired from `main.go`;
- no Kubernetes Secret value is read;
- no Secret value is logged or serialized;
- factory fallback paths remain fail-closed.

## Existing Contracts

The application layer already defines the runtime client factory contracts:

```text
internal/app/kubernetes_runtime_client_factory.go
internal/app/tekton_runtime_client_factory.go
internal/app/argocd_runtime_client_factory.go
```

The contracts are:

- `KubernetesRuntimeClientFactory`;
- `TektonRuntimeClientFactory`;
- `ArgoCDRuntimeClientFactory`.

Each factory receives:

- a resolved `TechnicalRuntimeTarget`;
- validated `RuntimeClientSecretRefs`;
- an in-memory `RuntimeSecretValueSet`.

Each contract has a conservative empty implementation:

- `EmptyKubernetesRuntimeClientFactory`;
- `EmptyTektonRuntimeClientFactory`;
- `EmptyArgoCDRuntimeClientFactory`.

These empty implementations must remain the default until explicit enablement.

## Existing Factory-Aware Wiring

The factory-aware provider wrappers are already present:

```text
internal/app/kubernetes_runtime_client_provider_factory_wiring.go
internal/app/tekton_runtime_client_provider_factory_wiring.go
internal/app/argocd_runtime_client_provider_factory_wiring.go
```

Each wrapper follows the same resolution order:

1. Try the existing provider registry first.
2. If the existing provider resolves, return the current-cluster client.
3. If the existing provider cannot resolve and no Secret references are configured, return the original provider error.
4. If Secret references are configured, ask the runtime Secret value loader to load values.
5. Build a real runtime client through the concrete factory.

With the current wiring, step 4 always fails closed because `EmptyRuntimeSecretValueLoader` is used.

## Existing Adapter Constructors

The concrete runtime clients can be built through existing adapter constructors:

```text
internal/adapters/kubernetes.New(cfg, opts...)
internal/adapters/tekton.New(cfg, opts...)
internal/adapters/argocd.New(cfg)
```

The relevant adapter configuration structures are:

```text
kubernetes.Config {
  APIURL
  Token
  TimeoutSeconds
  InsecureTLS
  CAFile
}

tekton.Config {
  APIURL
  Token
  TimeoutSeconds
  InsecureTLS
  CAFile
}

argocd.Config {
  BaseURL
  AuthToken
  TimeoutSeconds
  InsecureTLS
  CAFile
}
```

## Layering Decision

Concrete real runtime client factories must not be implemented inside `internal/app` if they need to import adapter packages.

`internal/app` remains the application and contract layer. It should continue to define:

- factory interfaces;
- request contracts;
- validation helpers;
- provider registry wrappers;
- secret reference models;
- secret value loader contracts.

Concrete factories should live in the composition/runtime layer, currently the `cmd/devops-control-plane` package, because this package already wires the concrete adapters and owns the application composition root.

This keeps dependency direction clean:

```text
cmd/devops-control-plane -> internal/app
cmd/devops-control-plane -> internal/adapters/*
internal/app             -> no dependency on internal/adapters/*
```

## Proposed Concrete Factory Types

The first concrete factory implementations should be local to the composition package and should not be wired into `main.go` yet.

Proposed files:

```text
cmd/devops-control-plane/kubernetes_runtime_client_factory_adapter.go
cmd/devops-control-plane/tekton_runtime_client_factory_adapter.go
cmd/devops-control-plane/argocd_runtime_client_factory_adapter.go
```

Proposed tests:

```text
cmd/devops-control-plane/kubernetes_runtime_client_factory_adapter_test.go
cmd/devops-control-plane/tekton_runtime_client_factory_adapter_test.go
cmd/devops-control-plane/argocd_runtime_client_factory_adapter_test.go
```

The concrete factories should satisfy the application contracts:

```text
app.KubernetesRuntimeClientFactory
app.TektonRuntimeClientFactory
app.ArgoCDRuntimeClientFactory
```

## Kubernetes Factory Design

The Kubernetes runtime client factory will build a `KubernetesRuntimeEvidenceClient` from a validated factory request.

Validation sequence:

1. Call `app.ValidateKubernetesRuntimeClientFactoryRequest(request)`.
2. Reject disabled targets through the existing `TechnicalRuntimeTarget.Validate()` path.
3. Resolve required Secret value keys through `app.RequiredKubernetesRuntimeSecretValueKeys(request.SecretRefs)`.
4. Read required values from `request.SecretValues.Resolve(...)`.
5. Build a `kubernetesadapter.Config` without logging Secret values.
6. Call `kubernetesadapter.New(...)`.
7. Return the resulting client as `app.KubernetesRuntimeEvidenceClient`.

The expected Secret values are:

```text
RuntimeSecretValueKubernetesToken
RuntimeSecretValueKubernetesCA optional, only if referenced
RuntimeSecretValueKubernetesKubeconfig optional, only if referenced
```

Initial implementation should support token-based construction first.

Kubeconfig-based construction should remain a future extension unless an adapter contract exists to consume kubeconfig directly.

## Tekton Factory Design

Tekton is accessed through the target cluster Kubernetes API.

The Tekton runtime client factory will build a `TektonRuntimeClient` from a validated factory request.

Validation sequence:

1. Call `app.ValidateTektonRuntimeClientFactoryRequest(request)`.
2. Resolve required Secret value keys through `app.RequiredTektonRuntimeSecretValueKeys(request.SecretRefs)`.
3. Read required values from `request.SecretValues.Resolve(...)`.
4. Build a `tektonadapter.Config` without logging Secret values.
5. Call `tektonadapter.New(...)`.
6. Wrap the adapter using `currentTektonRuntimeClient` so it satisfies `app.TektonRuntimeClient`.

The factory should reuse the existing adapter wrapper instead of duplicating Tekton request mapping logic.

## Argo CD Factory Design

The Argo CD runtime client factory will build an `ArgoCDRuntimeClient` from a validated factory request.

Validation sequence:

1. Call `app.ValidateArgoCDRuntimeClientFactoryRequest(request)`.
2. Resolve required Secret value keys through `app.RequiredArgoCDRuntimeSecretValueKeys(request.SecretRefs)`.
3. Read required values from `request.SecretValues.Resolve(...)`.
4. Build an `argocdadapter.Config` without logging Secret values.
5. Call `argocdadapter.New(...)`.
6. Wrap the adapter using `currentArgoCDRuntimeClient` so it satisfies `app.ArgoCDRuntimeClient`.

Expected Secret values:

```text
RuntimeSecretValueArgoCDToken
RuntimeSecretValueArgoCDBaseURL
RuntimeSecretValueArgoCDCA optional, only if referenced
```

## API URL and CA Handling

Current adapter constructors expect CA material through `CAFile`, not raw CA content.

The current `RuntimeSecretValueSet` can hold raw CA values, but it does not represent a filesystem path. Therefore, the first concrete factory implementation must not silently write CA contents to disk.

Conservative initial rule:

- token and base URL values may be used directly;
- raw CA values must not be written to temporary files;
- if a CA value is required and the adapter only accepts `CAFile`, the factory must fail closed with a clear error until a safe CA handling strategy is implemented.

A later phase can introduce an explicit CA material handling strategy if needed.

## Secret Handling Rules

Concrete factories must obey these rules:

- never log Secret values;
- never include Secret values in errors;
- never include Secret values in SafeSummary output;
- never retain references to caller-owned mutable maps;
- use `RuntimeSecretValueSet.Resolve(...)` only for required values;
- treat missing values as hard failures;
- fail closed for unsupported combinations such as kubeconfig-only inputs if the adapter cannot consume kubeconfig directly.

## Enablement Rules

Phase 15.8.8 implementation work must not wire the concrete factories into `main.go`.

Allowed in this phase:

- add concrete factory types;
- add constructor functions;
- add unit tests;
- add compile-time interface assertions;
- add documentation.

Not allowed in this phase:

- replacing `nil` factories in `main.go` with concrete factories;
- replacing `EmptyRuntimeSecretValueLoader` in `main.go`;
- enabling Secret reads in runtime;
- enabling staging;
- enabling production;
- reading live Kubernetes Secret values during tests.

## Unit Test Strategy

Tests should focus on contract compliance and fail-closed behavior.

Recommended test cases for each factory:

- factory satisfies the relevant `app.*RuntimeClientFactory` interface;
- invalid request returns the app-layer validation error;
- missing required Secret value returns `app.ErrRuntimeSecretValueNotAvailable` or a wrapped equivalent;
- CA-required request fails closed if raw CA cannot be safely passed to the adapter;
- kubeconfig-only Kubernetes/Tekton request fails closed unless explicit support is implemented;
- valid token-based configuration builds a client object without performing network calls.

The final point can be tested because adapter constructors only construct HTTP clients and validate configuration; they do not call the network.

## Non-Regression Validation

After each factory implementation step, run:

```text
go test ./...
git diff --check
git status --short
```

Runtime validation should be deferred until concrete factories are intentionally wired into the composition root and Secret loading is explicitly enabled through allow-listed configuration.

Until then, runtime validation should continue proving that:

- current-cluster provider behavior still works;
- factory fallback remains disabled;
- staging and production remain disabled.

## Next Implementation Steps

Recommended sequence:

1. Implement Kubernetes runtime client factory adapter with unit tests.
2. Implement Tekton runtime client factory adapter with unit tests.
3. Implement Argo CD runtime client factory adapter with unit tests.
4. Keep `main.go` unchanged.
5. Document implementation readiness.
6. Perform runtime non-regression validation with factories still not wired.

This keeps the project moving toward real multi-cluster support while preserving the validated baseline.
