# Controlled Runtime Secret Loader and Factory Enablement Design

## Purpose

This document defines the controlled enablement design for phase 15.8.9.

The DevOps Control Plane now has:

- factory-aware runtime provider registries wired conservatively in `main.go`;
- concrete runtime client factory adapters for Kubernetes, Tekton and Argo CD;
- runtime Secret reference models;
- an allow-list based Kubernetes Secret value loader;
- runtime validation proving that current-cluster behavior has no regression while factories remain disabled.

Phase 15.8.9 is a design phase. It must not enable real Secret reads yet and must not wire concrete factories into runtime execution.

## Current Baseline

The current validated baseline is:

```text
7f49645 Document factory-disabled runtime non-regression validation
```

The current runtime state remains conservative:

```text
main.go uses factory-aware registries
main.go keeps EmptyRuntimeSecretValueLoader
concrete factories are implemented but not wired
ocp-dev/current remains the active runtime baseline
staging and production remain disabled
no Kubernetes Secret value is read at runtime
```

Validated implementation context:

```text
b87cd55 Add Kubernetes runtime client factory adapter
7d02979 Add Tekton runtime client factory adapter
65b99d2 Add Argo CD runtime client factory adapter
52b8ac2 Document real runtime client factory readiness
7f49645 Document factory-disabled runtime non-regression validation
```

## Design Goals

The controlled enablement design must satisfy these goals:

1. Enable real runtime client construction only through explicit configuration.
2. Keep the default runtime posture fail-closed.
3. Keep `ocp-dev/current` as the non-regression baseline.
4. Enable one capability at a time.
5. Ensure Secret references and Secret values are never confused.
6. Ensure Secret values are never logged, serialized or included in errors.
7. Preserve staging and production disabled until dedicated environment enablement phases.
8. Provide clean rollback to the current factory-disabled behavior.

## Non-Goals

This phase does not:

- wire concrete factories into `main.go`;
- replace `EmptyRuntimeSecretValueLoader` in runtime;
- create or mount real runtime client Secrets;
- read any real Kubernetes Secret value;
- enable staging;
- enable production;
- change the cluster registry default behavior;
- change the environment catalog default behavior;
- introduce kubeconfig-based runtime client construction;
- introduce raw CA handling from Secret values.

## Current Components

### Runtime Secret References

Runtime Secret references are modeled by:

```text
internal/app/runtime_client_secret_model.go
internal/app/runtime_client_secret_loading.go
```

The model stores only cluster name, Secret namespace, Secret name and Secret key names. It must never store actual Secret values.

### Runtime Secret Value Loader

The runtime Secret value loader contract is defined by:

```text
internal/app/runtime_secret_value_loader.go
```

The current default is:

```text
EmptyRuntimeSecretValueLoader
```

A guarded implementation already exists:

```text
AllowListKubernetesSecretValueLoader
```

It requires an allow-list configuration, a KubernetesSecretGetter implementation and validated RuntimeClientSecretRefs.

### Concrete Factories

Concrete factories exist in the composition layer:

```text
cmd/devops-control-plane/kubernetes_runtime_client_factory_adapter.go
cmd/devops-control-plane/tekton_runtime_client_factory_adapter.go
cmd/devops-control-plane/argocd_runtime_client_factory_adapter.go
```

They are currently unit-tested but not wired in `main.go`.

## Controlled Enablement Principle

Real runtime client enablement must require all of the following to be true:

```text
runtime secret reference registry configured
secret loader allow-list configured
KubernetesSecretGetter configured
concrete factory explicitly wired
target cluster provider enabled
target environment enabled
capability-specific validation completed
```

If any condition is missing, the runtime must fail closed.

## Proposed Enablement Flags

Future enablement should be controlled by explicit configuration flags.

Suggested configuration variables:

```text
DCP_RUNTIME_SECRET_LOADER_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORIES_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_KUBERNETES_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_TEKTON_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_ARGOCD_ENABLED=false
```

Default values must be equivalent to `false`.

This allows staged progression:

1. Secret reference registry configured but Secret reading disabled.
2. Secret loader enabled but factories disabled.
3. One factory enabled for a non-production target.
4. Runtime provider enabled for one target cluster.
5. Environment enabled only after dedicated validation.

## Proposed Secret Loader Configuration

The allow-list configuration should be explicit and static at startup.

Suggested configuration source:

```text
DCP_RUNTIME_SECRET_LOADER_ALLOWED_REFS_FILE=/etc/dcp-runtime-client-secrets/allowed-refs.yaml
```

The file should define allowed Secret references by cluster, namespace and name.

Example structure:

```yaml
allowedClusters:
  - ocp-staging
allowedRefs:
  - clusterName: ocp-staging
    namespace: devops-control-plane
    name: staging-runtime-client
```

This file still stores only references, never Secret values.

## Proposed Runtime Secret Reference Configuration

The existing reference registry path should remain the source of Secret references:

```text
DCP_RUNTIME_CLIENT_SECRET_REFS_FILE=/etc/dcp-runtime-client-secrets/secret-refs.yaml
```

Example structure:

```yaml
clusters:
  - clusterName: ocp-staging
    kubernetes:
      namespace: devops-control-plane
      name: staging-runtime-client
      tokenKey: kubernetes-token
    tekton:
      namespace: devops-control-plane
      name: staging-runtime-client
      tokenKey: kubernetes-token
    argocd:
      namespace: devops-control-plane
      name: staging-argocd-runtime-client
      tokenKey: argocd-token
      baseURLKey: argocd-base-url
```

The runtime Secret reference registry must continue to reject values that look like actual credentials.

## Proposed KubernetesSecretGetter Wiring

The concrete Secret getter should be built from a Kubernetes adapter that is already configured for the control-plane cluster.

The getter must satisfy:

```text
app.KubernetesSecretGetter
```

Guardrails:

- read only allowed Secret namespace/name pairs;
- never print Secret data;
- return raw Secret data only to the loader;
- rely on Kubernetes RBAC to restrict Secret access;
- be constructed only when `DCP_RUNTIME_SECRET_LOADER_ENABLED=true`.

## Proposed Factory Wiring Sequence

Factory wiring should be added in a later implementation phase and must remain conditional.

Recommended sequence:

1. Add configuration parsing for enablement flags.
2. Add allow-list config loading.
3. Add concrete KubernetesSecretGetter construction.
4. Wire `AllowListKubernetesSecretValueLoader` only when the loader flag is enabled.
5. Wire concrete factories only when global and capability-specific factory flags are enabled.
6. Keep nil factories otherwise so existing default behavior remains unchanged.

Conceptual logic:

```text
if runtimeSecretLoaderEnabled:
    loader = AllowListKubernetesSecretValueLoader
else:
    loader = EmptyRuntimeSecretValueLoader

if runtimeClientFactoriesEnabled and kubernetesFactoryEnabled:
    kubernetesFactory = concrete Kubernetes factory
else:
    kubernetesFactory = nil
```

## Staged Validation Plan

### Stage 1 — Compile-time and unit validation

Expected checks:

```text
go test ./...
git diff --check
```

Validation focus:

- config parsing defaults to disabled;
- invalid allow-list config fails closed;
- missing Secret getter fails closed;
- missing Secret values fail closed;
- concrete factories remain disabled by default.

### Stage 2 — Runtime non-regression with all flags disabled

Expected runtime matrix:

```text
/readyz -> 200
create dev -> 201
collect-evidence -> 202
validate -> 202
check-validation -> 202 or valid in-progress state
check-deployment -> 202 with Synced/Healthy
staging -> 422 disabled
production -> 422 disabled
```

### Stage 3 — Secret loader enabled but factories disabled

Goal:

- prove the loader can be configured and still not used for current-cluster runtime paths;
- prove no factory fallback is triggered for `ocp-dev/current`;
- prove disabled environments remain disabled.

Expected:

```text
ocp-dev/current still succeeds through current provider
staging/production still 422
no Secret value appears in logs or responses
```

### Stage 4 — Single non-production factory dry-run

This stage must be performed only after a dedicated runtime target is enabled in configuration and after rollback is documented.

Recommended first target:

```text
staging only
one runtime capability only
```

Do not enable production in the same phase.

## Rollback Strategy

Rollback must be simple and immediate.

Primary rollback is configuration-only:

```text
DCP_RUNTIME_SECRET_LOADER_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORIES_ENABLED=false
```

After rollback:

```text
EmptyRuntimeSecretValueLoader is used
concrete factories are not used
current-cluster providers continue to work
non-current targets fail closed
```

A deployment rollout restart may be required for environment-variable based configuration.

## Observability Requirements

Logs may include safe summaries only:

```text
clusterName
providerCurrent
secretRefsConfigured
keysPresent booleans
loader/factory enabled flags
```

Logs must not include:

```text
tokens
kubeconfig content
CA content
Authorization headers
raw Secret data
Secret value strings
```

Error messages must describe missing configuration without echoing sensitive values.

## Security Review Checklist

Before enabling any real Secret read:

- confirm Kubernetes RBAC restricts the control-plane ServiceAccount to only required Secrets;
- confirm allow-list file contains only expected cluster/namespace/name triples;
- confirm Secret reference file contains only references and key names;
- confirm no code path logs `RuntimeSecretValueSet` raw values;
- confirm `SafeSummary` is used for diagnostics;
- confirm kubeconfig and raw CA paths remain disabled or explicitly designed;
- confirm rollback flags are documented and tested;
- confirm staging/production enablement is separate from Secret loader enablement.

## Recommended Phase Breakdown

Recommended follow-up phases:

```text
15.8.9.1 Document controlled enablement design
15.8.9.2 Add disabled-by-default config flags and tests
15.8.9.3 Add allow-list loader config parsing and tests
15.8.9.4 Add concrete KubernetesSecretGetter wiring behind disabled flag
15.8.9.5 Wire loader conditionally with factories still disabled
15.8.9.6 Runtime non-regression with loader flag disabled
15.8.9.7 Runtime non-regression with loader enabled but factories disabled
15.8.9.8 Prepare single-cluster single-capability factory enablement plan
```

No later subphase should enable staging or production without its own validation and rollback checklist.

## Conclusion

The controlled enablement approach keeps the current validated runtime stable while allowing the project to progress toward real multi-cluster support.

The key rule is that Secret loading, factory wiring and environment enablement must stay separate.

Each step must be independently testable and reversible.
