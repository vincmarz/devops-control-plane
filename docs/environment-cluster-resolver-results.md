# Environment-to-cluster Resolver Results

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** Environment-to-cluster resolver results
- **Phase:** 15.4.4 — Document Environment-to-cluster resolver results
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Baseline before Phase 15.4:** `7793186` — `Document cluster registry baseline results`
- **Resolver baseline commit:** `383f1f0` — `Add environment cluster resolver baseline`
- **ChangeService validation wiring commit/runtime image:** `0132f06` — `Use environment cluster resolver in change validation`
- **Related phases:**
  - 15.1.1 — Add multi-cluster environment enablement request
  - 15.2.2 — Load Environment Catalog from mounted file
  - 15.3.2 — Add Cluster Registry baseline
  - 15.3.3 — Document Cluster Registry baseline results
  - 15.4.1 — Inventory current client selection points
  - 15.4.2 — Add Environment Cluster Resolver baseline
  - 15.4.3 — Use Environment Cluster Resolver in ChangeService validation path
- **Status:** Completed validation report
- **Language:** English

---

## 1. Purpose

This document records the result of Phase 15.4, which introduced the first environment-to-cluster resolution layer in the DevOps Control Plane.

The purpose of this phase was to start using the metadata introduced by the Environment Catalog and Cluster Registry in the application layer, while preserving the existing runtime behavior and without enabling `staging` or `production`.

The resolver provides the following logical chain:

```text
ChangeRequest.targetEnvironment
  -> Environment Catalog
  -> environment.clusterName
  -> Cluster Registry
  -> cluster metadata
```

This phase is intentionally conservative. It does not yet change Kubernetes, Tekton or Argo CD adapter wiring. It only introduces and then uses the resolver in the ChangeService create validation path.

---

## 2. Scope

Phase 15.4 covers:

- inventory of current single-cluster client selection points;
- identification of current adapter wiring limitations;
- introduction of the `EnvironmentClusterResolver` baseline;
- unit validation of environment-to-cluster resolution cases;
- wiring of the resolver into the ChangeService create validation path;
- runtime validation that the previous `targetEnvironment` behavior remains unchanged.

Phase 15.4 does not cover:

- dynamic Kubernetes client selection by cluster;
- dynamic Tekton client selection by cluster;
- dynamic Argo CD client selection by cluster or Argo CD instance;
- usage of real staging or production cluster credentials;
- enabling `staging` or `production`;
- changing UI action visibility behavior.

---

## 3. Starting point

Before Phase 15.4, the DevOps Control Plane already had:

```text
Environment Catalog runtime loading
Cluster Registry runtime loading
Environment Catalog clusterName metadata
Cluster Registry baseline entries
```

The repository baseline before Phase 15.4 was:

```text
7793186 Document cluster registry baseline results
52de0c1 Add cluster registry baseline
4d5f524 Align Go module dependencies with build image
4c1706e Load environment catalog from mounted file
a203112 Add multi-cluster environment enablement request
```

The runtime image before the resolver wiring was:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:52de0c1
```

The runtime had the following mounts:

```text
/etc/dcp-trust
/etc/dcp-environments
/etc/dcp-clusters
```

The target environment behavior before Phase 15.4 was:

```text
missing targetEnvironment -> HTTP 201, default dev
targetEnvironment=dev -> HTTP 201
targetEnvironment=staging -> HTTP 422, currently disabled
targetEnvironment=production -> HTTP 422, currently disabled
targetEnvironment=unknown-env -> HTTP 422, not configured
```

---

## 4. Phase 15.4.1 — Current client selection inventory

### 4.1 Objective

The inventory identified where the current runtime still uses single-cluster global configuration.

The goal was to understand where future phases must introduce environment-to-cluster client selection.

### 4.2 Inventory summary

Observed summary:

```text
git_head=7793186 Document cluster registry baseline results
kubernetes_api_url_refs=5
tekton_namespace_refs=2
argocd_base_url_refs=2
cluster_name_refs=14
cluster_registry_file_refs=2
```

### 4.3 Current single-cluster configuration references

The inventory confirmed active references to:

```text
KUBERNETES_API_URL
KUBERNETES_TOKEN
KUBERNETES_CA_FILE
KUBERNETES_NAMESPACE
TEKTON_NAMESPACE
TEKTON_PIPELINE_NAME
ARGOCD_BASE_URL
ARGOCD_AUTH_TOKEN
```

The main files involved are:

```text
internal/config/config.go
manifests/configmap.yaml
cmd/devops-control-plane/main.go
```

### 4.4 Adapter constructor inventory

Adapter constructors identified:

```text
internal/adapters/argocd/client.go: New
internal/adapters/gitlab/client.go: New
internal/adapters/kubernetes/client.go: New
internal/adapters/tekton/client.go: New
```

Application service constructor identified:

```text
internal/app/application_service.go: NewApplicationService
```

Runtime wiring for Argo CD currently happens in:

```text
cmd/devops-control-plane/main.go
```

### 4.5 Tekton namespace usage

Tekton namespace usage is currently based on global configuration:

```text
cmd/devops-control-plane/main.go: Namespace: cfg.TektonNamespace
cmd/devops-control-plane/main.go: FindLatestPipelineRunByChange(ctx, cfg.TektonNamespace, ...)
cmd/devops-control-plane/main.go: ListTaskRunsByPipelineRun(ctx, cfg.TektonNamespace, ...)
```

This confirms that Tekton execution and status lookup remain single-cluster and single-namespace for now.

### 4.6 Kubernetes namespace usage

Kubernetes runtime evidence collection currently uses:

```text
cfg.KubernetesNamespace
```

Main call site:

```text
cmd/devops-control-plane/main.go
```

This confirms that runtime evidence collection is still single-cluster and single-namespace.

### 4.7 Argo CD usage

Argo CD is still configured from global settings:

```text
ARGOCD_BASE_URL
ARGOCD_AUTH_TOKEN
ARGOCD_CA_FILE
ARGOCD_INSECURE_TLS
```

Main wiring:

```text
cmd/devops-control-plane/main.go
```

This means the current implementation still assumes one configured Argo CD endpoint.

### 4.8 Existing Cluster Registry and Environment Catalog metadata

Cluster Registry and Environment Catalog metadata are already present:

```text
clusterName in EnvironmentDefinition
DefaultClusterRegistry
CLUSTER_REGISTRY_FILE
```

However, before Phase 15.4.2 and 15.4.3 this metadata was not used by the ChangeService create path.

---

## 5. Phase 15.4.2 — Environment Cluster Resolver baseline

### 5.1 Commit

The resolver baseline was implemented and pushed with:

```text
383f1f0 Add environment cluster resolver baseline
```

### 5.2 Files introduced

New files:

```text
internal/app/environment_cluster_resolver.go
internal/app/environment_cluster_resolver_test.go
```

### 5.3 Resolver model

The resolver introduces:

```text
EnvironmentClusterResolution
EnvironmentClusterResolver
```

The resolution object contains:

```text
TargetEnvironment
Environment
Cluster
```

### 5.4 Resolver chain

The resolver implements the logical chain:

```text
targetEnvironment
  -> Environment Catalog
  -> clusterName
  -> Cluster Registry
```

### 5.5 Resolver functions

The resolver baseline provides:

```text
DefaultEnvironmentClusterResolver
NewEnvironmentClusterResolver
Resolve
ResolveEnabledTarget
ResolveTechnicalActionTarget
```

### 5.6 Baseline behavior

The baseline resolver supports the following cases:

```text
missing targetEnvironment -> dev -> ocp-dev
dev -> ocp-dev
staging -> ocp-staging, while environment and cluster remain disabled
production -> ocp-production, while environment and cluster remain disabled
unknown-env -> targetEnvironment not configured
environment without clusterName -> error
environment with unknown clusterName -> error
```

### 5.7 Unit validation

The tests validate:

```text
default environment resolution
configured disabled environment resolution
unknown environment rejection
missing clusterName rejection
unknown clusterName rejection
enabled target behavior
technical action target behavior for dev
```

Local validation before commit:

```text
go test ./... OK
git diff --check OK
```

---

## 6. Phase 15.4.3 — ChangeService validation path wiring

### 6.1 Commit

The ChangeService validation path wiring was implemented and pushed with:

```text
0132f06 Use environment cluster resolver in change validation
```

### 6.2 Files changed

Changed or added files:

```text
internal/app/change_service.go
internal/app/change_service_environment_cluster_resolver_test.go
```

### 6.3 ChangeService behavior

The ChangeService create path now invokes:

```text
DefaultEnvironmentClusterResolver().ResolveEnabledTarget(req.TargetEnvironment)
```

This introduces environment-to-cluster resolution into create validation.

### 6.4 Preserved behavior

The wiring intentionally preserves the existing validation behavior:

```text
missing targetEnvironment -> accepted as dev
dev -> accepted
staging -> rejected because targetEnvironment is disabled
production -> rejected because targetEnvironment is disabled
unknown-env -> rejected because targetEnvironment is not configured
```

### 6.5 Regression test

A regression test was added to confirm that the ChangeService create path uses the Environment Cluster Resolver.

The test checks that `change_service.go` contains:

```text
DefaultEnvironmentClusterResolver().ResolveEnabledTarget(req.TargetEnvironment)
```

This is a lightweight source-level guard for the current incremental phase. A future refactor may replace this with dependency-injected resolver tests.

### 6.6 Local validation

Local validation after commit:

```text
git status --short -> clean
go test ./... -> OK
git diff --check -> OK
```

---

## 7. Runtime deployment for Phase 15.4.3

### 7.1 Runtime image

The runtime was deployed with image:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:0132f06
```

### 7.2 Runtime mounts

Runtime deployment confirmed:

```text
container=oauth-proxy
image=registry.redhat.io/openshift4/ose-oauth-proxy:latest
mount=oauth-proxy-tls path=/etc/tls/private

container=devops-control-plane
image=image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:0132f06
mount=dcp-trust-bundle path=/etc/dcp-trust
mount=dcp-environments path=/etc/dcp-environments
mount=dcp-clusters path=/etc/dcp-clusters
```

This confirms that the application runtime had both the Environment Catalog and Cluster Registry file mounts available.

---

## 8. Runtime validation for Phase 15.4.3

### 8.1 Readiness

Readiness result:

```text
readyz=200
```

### 8.2 Create ChangeRequest without targetEnvironment

Observed HTTP result:

```text
missing_target_env_http=201
```

Expected behavior:

```text
The request defaults to dev and resolves dev -> ocp-dev.
```

### 8.3 Create ChangeRequest with targetEnvironment=dev

Observed HTTP result:

```text
dev_target_env_http=201
```

Expected behavior:

```text
dev is enabled and resolves to ocp-dev.
```

### 8.4 Create ChangeRequest with targetEnvironment=staging

Observed HTTP result:

```text
staging_target_env_http=422
```

Expected behavior:

```text
staging is configured and resolves to ocp-staging, but staging remains disabled.
```

### 8.5 Create ChangeRequest with targetEnvironment=production

Observed HTTP result:

```text
production_target_env_http=422
```

Expected behavior:

```text
production is configured and resolves to ocp-production, but production remains disabled.
```

### 8.6 Create ChangeRequest with targetEnvironment=unknown-env

Observed HTTP result:

```text
unknown_target_env_http=422
```

Expected behavior:

```text
unknown-env is not configured and is rejected.
```

### 8.7 Runtime validation summary

Final runtime validation summary:

```text
readyz=200
missing_target_env_http=201
dev_target_env_http=201
staging_target_env_http=422
production_target_env_http=422
unknown_target_env_http=422
```

The runtime behavior remained unchanged after the resolver was wired into the ChangeService create validation path.

---

## 9. Current architecture after Phase 15.4

The application now contains the following layers:

```text
Environment Catalog
Cluster Registry
Environment Cluster Resolver
ChangeService create validation using resolver
```

The logical model is now:

```text
ChangeRequest.targetEnvironment
  -> Environment Catalog
  -> environment.clusterName
  -> Cluster Registry
  -> cluster metadata
```

The execution model is still intentionally conservative:

```text
technical adapters still use global current-cluster configuration
staging remains disabled
production remains disabled
no external cluster credentials are used
```

---

## 10. Current limitations

The following limitations remain by design:

```text
Kubernetes client selection is not yet environment-aware.
Tekton client selection is not yet environment-aware.
Argo CD client selection is not yet environment-aware.
GitOps path selection is not yet fully environment-aware in runtime operations.
No staging OpenShift cluster is connected.
No production OpenShift cluster is connected.
No multi-cluster credentials are mounted or consumed.
```

The resolver is the application-level bridge needed before implementing those later runtime selections.

---

## 11. Security considerations

This phase does not introduce new secrets.

The resolver uses metadata already available in:

```text
Environment Catalog
Cluster Registry
```

The Cluster Registry continues to store only non-secret metadata and references such as:

```text
clusterName
apiURL
caConfigMapRef
tokenSecretRef
defaultNamespace
allowedNamespaces
```

No bearer token, password, private key or raw kubeconfig is stored or printed by this phase.

---

## 12. Acceptance summary

Phase 15.4 is accepted because:

- current single-cluster client selection points were inventoried;
- an Environment Cluster Resolver was added;
- resolver unit tests validate expected environment-to-cluster mappings;
- ChangeService create validation now uses the resolver;
- existing `targetEnvironment` behavior is preserved;
- runtime image `0132f06` was deployed;
- runtime mounts for Environment Catalog and Cluster Registry were present;
- `/readyz` returned HTTP 200;
- create validation behavior matched the expected baseline;
- staging and production remained disabled;
- unknown environments remained rejected.

---

## 13. Recommended next step

Proceed with a controlled next phase:

```text
Phase 15.5 — Multi-user / multi-environment / multi-cluster validation matrix
```

or, if continuing the client-selection implementation first:

```text
Phase 15.4.5 — Document or expose resolved cluster metadata for diagnostics
Phase 15.5 — Environment-aware technical workflow routing baseline
```

Recommended next engineering target:

```text
Pass resolved environment and cluster metadata deeper into technical workflow services without yet switching clients.
```

This would allow logs, evidence and tests to confirm the intended target cluster while keeping actual execution on the current dev cluster until real staging and production clusters are available.
