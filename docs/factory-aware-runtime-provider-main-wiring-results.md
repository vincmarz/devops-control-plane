# Factory-Aware Runtime Provider Main Wiring Results

## Purpose

This document records the results of the factory-aware runtime provider main wiring work completed in phase 15.8.7.3 and the runtime validation completed in phase 15.8.7.4.

The scope was intentionally conservative:

- prepare the main application wiring for real multi-cluster runtime clients;
- keep the existing `ocp-dev` / current-cluster runtime behavior unchanged;
- keep staging and production disabled;
- keep Secret value loading disabled by default;
- avoid reading or exposing any Kubernetes Secret values;
- validate the runtime paths for Kubernetes, Tekton, and Argo CD without regressions.

## Validated Commit

The runtime validation was performed against commit:

```text
f657f44 Wire factory-aware runtime provider registries in main
```

At validation time, `main` and `origin/main` were aligned and the source working tree was clean.

## Implementation Summary

### Factory-aware registries wired in `main.go`

The application main wiring now uses factory-aware runtime provider registries for the three runtime integrations:

- Kubernetes runtime evidence collection;
- Tekton validation and validation status checks;
- Argo CD deployment status checks.

The following factory-aware registries are wired from `main.go`:

- `NewKubernetesRuntimeClientProviderFactoryAwareRegistry`;
- `NewTektonRuntimeClientProviderFactoryAwareRegistry`;
- `NewArgoCDRuntimeClientProviderFactoryAwareRegistry`.

The current configuration remains conservative. The registries are wired with `EmptyRuntimeSecretValueLoader`, and no real Secret value loading is enabled.

### ChangeService resolver contract

`ChangeService` no longer depends directly on the concrete Kubernetes runtime client provider registry type. It now depends on a resolver contract that exposes the minimal runtime resolution operation required by the service.

This allows both of the following to satisfy the dependency:

- the existing default Kubernetes runtime client provider registry;
- the new factory-aware Kubernetes runtime provider registry wrapper.

This preserves existing behavior while preparing the service for future real multi-cluster runtime client factories.

### Runtime Secret loading remains disabled

The runtime Secret loading path remains fail-closed.

The current wiring uses:

```text
EmptyRuntimeSecretValueLoader
```

Consequences:

- no Kubernetes Secret is read;
- no Secret value is logged or serialized;
- no staging or production runtime client is enabled by the factory fallback;
- fallback paths for Secret-configured targets remain explicitly not configured until a later enablement phase introduces real loaders and factories.

## Runtime Validation Setup

Runtime validation was performed through a direct port-forward to the application container, bypassing the OAuth proxy for trusted-header validation.

The Service exposes only the OAuth proxy port:

```text
service: devops-control-plane
service port: 8443
targetPort: oauth-proxy
```

The application container exposes:

```text
container: devops-control-plane
container port: 8080
```

The successful validation path used:

```text
pod/devops-control-plane-699bf6bdb7-l9wlx 18107:8080
BASE_URL=http://127.0.0.1:18107
```

A readiness check through the direct application port returned:

```text
readyz=200
```

## Runtime RBAC Notes

The runtime ConfigMap defines the active RBAC groups as:

```text
AUTH_GROUP_ADMIN=devops-control-plane-admins
AUTH_GROUP_APPROVER=devops-control-plane-approvers
AUTH_GROUP_OPERATOR=devops-control-plane-operators
AUTH_GROUP_VIEWER=devops-control-plane-viewers
AUTH_HEADER_USER=X-Forwarded-User
AUTH_HEADER_GROUPS=X-Forwarded-Groups
AUTH_HEADER_ALT_USER=X-Auth-Request-User
AUTH_OPENSHIFT_GROUP_LOOKUP_ENABLED=true
```

During validation, requests using old or non-configured group names correctly failed authorization with HTTP 403.

Examples of non-configured groups that were rejected:

```text
devops-operators
devops-developers
```

The configured operator group succeeded:

```text
devops-control-plane-operators
```

## Create ChangeRequest Contract Notes

The create ChangeRequest endpoint requires a complete payload including at least:

- `title`;
- `changeType`;
- `applicationName`;
- `targetEnvironment`.

Runtime validation confirmed that incomplete payloads were rejected by schema validation after RBAC was satisfied:

```text
missing title      -> HTTP 422, title is required
missing changeType -> HTTP 422, changeType is required
```

The successful create request used:

```text
title=Runtime validation for factory-aware main wiring
changeType=standard
applicationName=demo-go-color-app
targetEnvironment=dev
```

The API created:

```text
CHANGE_NUMBER=CHG-2026-0032
HTTP 201
```

## Runtime Validation Results

### Readiness

```text
/readyz -> HTTP 200
```

The readiness response reported:

```text
configuration=ok
database=ok
status=ready
```

### Create ChangeRequest on dev

```text
POST /api/v1/changes -> HTTP 201
CHANGE_NUMBER=CHG-2026-0032
requestedBy=operator
status=draft
riskLevel=medium
```

### Kubernetes runtime provider path: collect evidence

```text
POST /api/v1/changes/CHG-2026-0032/collect-evidence -> HTTP 202
runtimeStatus=EvidenceCollected
```

The evidence response included a sanitized deployment evidence record for:

```text
applicationName=demo-go-color-app
evidenceType=deployment
name=deployment-evidence
sanitized=true
```

### Tekton runtime provider path: validate

```text
POST /api/v1/changes/CHG-2026-0032/validate -> HTTP 202
runtimeStatus=ValidationRunning
```

A real Tekton PipelineRun was created:

```text
namespace=devops-ci-demo
pipelineRunName=devops-cp-validate-chg-2026-0032-l4w7c
```

### Tekton runtime provider path: check validation

```text
POST /api/v1/changes/CHG-2026-0032/check-validation -> HTTP 202
runtimeStatus=ValidationRunning
```

The PipelineRun was observed as not yet completed at check time:

```text
status=Unknown
reason=Pending
message=PipelineRun has no Succeeded condition yet
```

This is a valid intermediate runtime state and confirms that the provider-aware Tekton check path resolved the runtime client and queried the PipelineRun successfully.

### Argo CD runtime provider path: check deployment

```text
POST /api/v1/changes/CHG-2026-0032/check-deployment -> HTTP 202
runtimeStatus=DeploymentSyncedHealthy
```

The Argo CD application state was resolved as:

```text
applicationName=demo-go-color-app
project=devops-ci-demo
syncStatus=Synced
healthStatus=Healthy
revision=6bfdd07a96b4a23569a82bfb1e252e48eed248bc
```

### Disabled environment guardrails

The staging and production environments remain disabled.

Validation results:

```text
POST /api/v1/changes targetEnvironment=staging    -> HTTP 422
POST /api/v1/changes targetEnvironment=production -> HTTP 422
```

The API returned explicit validation errors:

```text
targetEnvironment "staging" is currently disabled
targetEnvironment "production" is currently disabled
```

## Security and Secret Handling

The validation did not enable real Secret loading.

The following guardrails remained in place:

- no Kubernetes Secret value was read;
- no Kubernetes Secret value was printed;
- no token, CA bundle content, kubeconfig content, Argo CD token, or Tekton credential was exposed;
- staging and production stayed disabled;
- the factory fallback path remained fail-closed through `EmptyRuntimeSecretValueLoader`.

## Outcome

Phase 15.8.7.3 and phase 15.8.7.4 are completed as a conservative baseline for later real multi-cluster runtime client enablement.

Validated outcomes:

- factory-aware runtime provider registries are wired from `main.go`;
- `ChangeService` accepts a resolver contract instead of the concrete Kubernetes registry;
- Kubernetes collect-evidence works through the provider-aware path;
- Tekton validate and check-validation work through the provider-aware path;
- Argo CD check-deployment works through the provider-aware path;
- `ocp-dev` / current-cluster remains the active baseline;
- staging and production remain disabled;
- Secret value loading remains disabled and fail-closed;
- no runtime regression was observed.

## Next Steps

Recommended next steps:

1. Keep the current baseline unchanged until real runtime factories are introduced.
2. Introduce concrete runtime client factories only behind explicit configuration and allow-list guardrails.
3. Enable real Secret value loading only after the loader has a concrete Kubernetes Secret getter and strict allow-list configuration.
4. Validate staging and production in dedicated enablement phases, one environment at a time.
5. Keep runtime validation focused on one integration surface at a time to avoid mixing Secret loading, factory creation, and environment enablement in a single change.
