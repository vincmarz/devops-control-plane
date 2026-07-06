# DevOps Control Plane — Argo CD Integration

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 06 — Argo CD Integration
- **Version:** 0.2
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Previous documents:**
  - `docs/00-vision.md`
  - `docs/01-scope-mvp.md`
  - `docs/02-personas-use-cases.md`
  - `docs/03-functional-requirements.md`
  - `docs/04-non-functional-requirements.md`
  - `docs/05-architecture.md`
- **Status:** Rewritten in English and refreshed to align with the current advanced MVP, TLS, evidence and multi-environment baseline
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document describes how the DevOps Control Plane integrates with Argo CD / OpenShift GitOps.

The original document described the initial MVP integration design. This refreshed version updates the design to match the current implementation baseline and the near-term architecture direction.

The document covers:

- Argo CD responsibilities in the Control Plane architecture;
- the responsibilities of the DevOps Control Plane Argo CD adapter;
- current application discovery capabilities;
- deployment check workflow;
- runtime status mapping;
- treatment of warnings such as `OrphanedResourceWarning`;
- deployment evidence and diagnostics;
- TLS strict and application trust bundle requirements;
- token handling and security requirements;
- current `/api/v1` endpoint model;
- future environment-aware Argo CD mapping for `dev`, `staging` and `production`.

The DevOps Control Plane does not replace Argo CD. Argo CD remains the GitOps reconciliation engine.

---

## 2. Role of Argo CD in the architecture

Argo CD is the GitOps reconciliation engine.

Argo CD is responsible for:

- reading desired state from Git;
- comparing desired state with runtime cluster state;
- reporting `Synced`, `OutOfSync` and `Unknown` sync states;
- reporting `Healthy`, `Progressing`, `Degraded`, `Missing`, `Suspended` and `Unknown` health states;
- exposing Application metadata, revision, conditions and resources;
- exposing warnings such as `OrphanedResourceWarning`;
- enforcing AppProject constraints;
- maintaining Argo CD deployment history and operational state.

The DevOps Control Plane is responsible for:

- calling Argo CD API through a dedicated adapter;
- normalizing Argo CD data into internal models;
- correlating Argo CD state with ChangeRequests;
- mapping Argo CD state to Control Plane runtime status;
- preserving warnings and conditions;
- storing deployment evidence in PostgreSQL;
- exposing status and evidence through API and UI;
- keeping Argo CD tokens and TLS configuration safe;
- preparing environment-aware Argo CD Application mapping.

---

## 3. Integration principles

### 3.1 Argo CD remains the deployment state authority

The DevOps Control Plane does not implement its own deployment engine.

The desired flow remains:

```text
GitLab change
  -> Git branch / commit / merge request
  -> Argo CD reconciliation status
  -> OpenShift runtime state
  -> evidence and audit in PostgreSQL
```

### 3.2 Integration uses Argo CD API

The Go application integrates with Argo CD through Argo CD API calls encapsulated in the Argo CD adapter.

The `argocd` CLI can be useful for manual troubleshooting, but application runtime logic must not depend on parsing CLI output.

### 3.3 Argo CD history is distinct from Git and ChangeRequest history

The Control Plane must keep a clear distinction between:

```text
GitLab history        = commits, branches and merge requests
Argo CD history       = revisions and deployment/reconciliation history
Tekton history        = validation PipelineRuns and TaskRuns
Control Plane history = ChangeRequests, events and evidence
```

### 3.4 `OutOfSync` is context-dependent

`OutOfSync` means that runtime state does not match the desired state observed by Argo CD.

It may indicate:

- drift;
- a change not yet reconciled;
- a sync failure;
- temporary operational rollback;
- a manual runtime mutation;
- a new Git revision not yet applied.

The Control Plane must show `OutOfSync` clearly, but it must not blindly treat every `OutOfSync` state as the same class of application failure.

### 3.5 Warnings must be preserved

Warnings such as `OrphanedResourceWarning` are important evidence.

The Control Plane must preserve and expose those warnings without automatically overriding the health assessment when Argo CD reports the Application as `Healthy`.

---

## 4. Current implementation baseline

Current implemented Argo CD-related capabilities include:

- Argo CD application client;
- `GET /api/v1/applications` where applicable;
- `GET /api/v1/applications/{name}`;
- deployment check workflow through `POST /api/v1/changes/{id}/check-deployment`;
- runtime status mapping for deployment state;
- deployment evidence collection through `POST /api/v1/changes/{id}/collect-evidence`;
- deployment diagnostics enrichment;
- UI rendering for deployment diagnostics and evidence;
- TLS strict mode using an application-dedicated trust bundle;
- dedicated Argo CD API token handling through OpenShift Secret;
- token rotation procedure after exposure risk;
- preservation of `OrphanedResourceWarning` as warning evidence.

The original direct sync endpoint model is no longer the main implemented path. The current operational workflow focuses on checking deployment state and collecting evidence after GitLab/Tekton/GitOps workflow execution.

---

## 5. Argo CD adapter

## 5.1 Responsibilities

The Argo CD adapter encapsulates all Argo CD API interactions.

The rest of the application must not depend on raw Argo CD HTTP payload details, token handling or transport configuration.

### Current responsibilities

- create an authenticated Argo CD API client;
- retrieve Application data;
- map sync and health status;
- extract revision and conditions;
- preserve warnings;
- return normalized application results to application services;
- support deployment check workflows;
- use TLS strict settings and configured CA bundle.

### Future responsibilities

- resolve Argo CD Application by `applicationName` and `targetEnvironment`;
- support environment-specific Application mapping;
- optionally expose more history/resource detail where needed.

---

## 5.2 Conceptual interface

The current and future adapter can be represented conceptually as:

```go
type ArgoCDAdapter interface {
    ListApplications(ctx context.Context) ([]ApplicationSummary, error)
    GetApplication(ctx context.Context, name string) (*ApplicationDetail, error)
    CheckDeployment(ctx context.Context, applicationName string) (*DeploymentStatus, error)
}
```

The exact Go interface can evolve, but the design principle remains stable:

```text
application services depend on adapter ports, not raw HTTP implementation details
```

---

## 5.3 Package location

Current package area:

```text
internal/adapters/argocd/
```

Expected responsibilities by file or component:

```text
client.go    -> HTTP/API client and request handling
models.go    -> DTOs and normalized models
mapper.go    -> mapping from Argo CD payload to internal model
errors.go    -> error normalization
```

The package should remain independently testable through fake HTTP servers or fake clients.

---

## 6. Normalized data model

## 6.1 Application summary

Example normalized Application summary:

```yaml
name: demo-go-color-app
namespace: openshift-gitops
project: default
targetNamespace: devops-ci-demo
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
targetRevision: main
path: apps/demo-go-color-app
syncStatus: Synced
healthStatus: Healthy
currentRevision: <revision>
```

## 6.2 Application detail

Example normalized Application detail:

```yaml
name: demo-go-color-app
namespace: openshift-gitops
project: default
syncStatus: Synced
healthStatus: Healthy
currentRevision: <revision>
source:
  repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
  targetRevision: main
  path: apps/demo-go-color-app
destination:
  server: https://kubernetes.default.svc
  namespace: devops-ci-demo
conditions:
  - type: OrphanedResourceWarning
    message: Application has 5 orphaned resources
```

## 6.3 Deployment status

The Control Plane deployment check maps Argo CD data to runtime status.

Example:

```yaml
applicationName: demo-go-color-app
syncStatus: Synced
healthStatus: Healthy
revision: <revision>
runtimeStatus: DeploymentSyncedHealthy
conditions:
  - type: OrphanedResourceWarning
    message: Application has 5 orphaned resources
```

---

## 7. Authentication and token handling

## 7.1 Token-based authentication

The Control Plane authenticates to Argo CD with an API token stored in OpenShift Secret.

Configuration baseline:

```text
ARGOCD_BASE_URL=https://openshift-gitops-server-openshift-gitops.apps.example.local
ARGOCD_AUTH_TOKEN=<from Secret>
ARGOCD_INSECURE_TLS=false
ARGOCD_CA_FILE=/etc/dcp-trust/ca-bundle.crt
ARGOCD_TIMEOUT_SECONDS=30
```

### Rules

- `ARGOCD_AUTH_TOKEN` must come from Secret.
- The token must not be logged.
- The token must not be returned through API responses.
- The token must not be stored in PostgreSQL.
- The token must not appear in evidence payloads.
- Token exposure requires rotation.

## 7.2 Dedicated Argo CD account

The current operational baseline uses a dedicated Argo CD account for the DevOps Control Plane with API token capability and minimal required permissions.

This avoids using personal or admin credentials as a stable runtime pattern.

---

## 8. TLS and trust model

## 8.1 TLS strict baseline

The Argo CD integration must use TLS strict mode.

Current baseline:

```text
ARGOCD_INSECURE_TLS=false
ARGOCD_CA_FILE=/etc/dcp-trust/ca-bundle.crt
```

## 8.2 Application-dedicated trust bundle

The cluster-wide trusted bundle did not include the ingress route CA required for the Argo CD Route in the current lab baseline.

The validated approach is an application-dedicated trust bundle:

```text
ConfigMap: dcp-app-trust-bundle
Mount: /etc/dcp-trust/ca-bundle.crt
```

The trust bundle contains the route CA required to validate the Argo CD route certificate.

## 8.3 Rules

- Do not modify cluster-wide OpenShift trust settings for this application-only requirement.
- Keep TLS strict for Argo CD.
- Keep the trust bundle managed as application configuration.
- Validate the mounted CA file before disabling insecure mode.

---

## 9. Sync and health status mapping

## 9.1 Sync status

| Argo CD status | Meaning |
|---|---|
| `Synced` | Runtime state matches desired Git state according to Argo CD. |
| `OutOfSync` | Runtime state differs from desired Git state according to Argo CD. |
| `Unknown` | Argo CD cannot determine the sync state. |

## 9.2 Health status

| Argo CD status | Meaning |
|---|---|
| `Healthy` | Resources are considered healthy. |
| `Progressing` | Resources are still reconciling or rolling out. |
| `Degraded` | Resources are not healthy. |
| `Missing` | Expected resource is missing. |
| `Suspended` | Resource is intentionally suspended. |
| `Unknown` | Health state cannot be determined. |

## 9.3 Control Plane runtime status mapping

| Sync | Health | Runtime status |
|---|---|---|
| `Synced` | `Healthy` | `DeploymentSyncedHealthy` |
| `OutOfSync` | any | `DeploymentOutOfSync` |
| any | `Progressing` | `DeploymentProgressing` |
| any | `Degraded` | `DeploymentDegraded` |
| any | unknown or unsupported | `DeploymentUnknown` |

The mapping must remain explicit and auditable.

---

## 10. Conditions and warnings

## 10.1 `OrphanedResourceWarning`

Argo CD can report:

```text
OrphanedResourceWarning
```

This condition indicates that resources exist in the target namespace but are not managed by the Application.

### Control Plane rule

- Show the warning.
- Store the warning in evidence.
- Do not automatically mark a `Synced` and `Healthy` application as failed only because this warning exists.
- Include the warning count in deployment diagnostics.

## 10.2 AppProject resource restrictions

Argo CD may fail sync or report errors when an AppProject does not allow a resource.

Example:

```text
resource :ConfigMap is not permitted in project devops-ci-demo
```

Operational interpretation:

```text
The resource exists in the GitOps repository, but the AppProject does not allow the Application to manage that resource kind.
```

Suggested message:

```text
Review the AppProject namespaceResourceWhitelist and allow the required resource kind if the change is expected and approved.
```

---

## 11. Deployment check workflow

The current baseline uses deployment checks rather than relying only on an imperative sync flow.

### Endpoint

```text
POST /api/v1/changes/{id}/check-deployment
```

### Main flow

```text
ChangeRequest
  -> ChangeService.CheckDeployment
  -> Argo CD adapter GetApplication
  -> map sync and health status
  -> update runtime status
  -> create technical event
  -> return normalized deployment result
```

### Acceptance criteria

- `Synced` and `Healthy` produce `DeploymentSyncedHealthy`.
- `OutOfSync`, `Progressing`, `Degraded` and unknown states are handled explicitly.
- Previous runtime status is preserved in event payload where useful.
- Lifecycle status is not overwritten by runtime status.

---

## 12. Deployment evidence workflow

### Endpoint

```text
POST /api/v1/changes/{id}/collect-evidence
```

### Main flow

```text
ChangeRequest
  -> collect Argo CD state
  -> collect Kubernetes/OpenShift runtime state
  -> build sanitized deployment evidence
  -> build diagnostics summary
  -> persist evidence in PostgreSQL
  -> update runtime status to EvidenceCollected
```

### Evidence includes

- ChangeRequest metadata;
- Application name;
- target environment;
- Argo CD sync status;
- Argo CD health status;
- Git revision;
- Argo CD conditions;
- Kubernetes/OpenShift runtime evidence;
- deployment diagnostics.

---

## 13. Deployment diagnostics

Deployment evidence includes diagnostics derived from Argo CD and Kubernetes/OpenShift payloads.

Diagnostics include:

```text
argocdSynced
argocdHealthy
deploymentReady
generationObserved
readyReplicas
podsReady
totalRestarts
serviceAvailable
routeAvailable
warnings
summary
```

Example summary:

```text
Application demo-go-color-app is Synced/Healthy with 2/2 replicas ready; warnings: 1
```

Diagnostics are displayed in the UI and stored in the evidence payload.

---

## 14. Error model

## 14.1 Argo CD error codes

Recommended internal error family:

```text
ARGOCD_AUTH_FAILED
ARGOCD_FORBIDDEN
ARGOCD_APPLICATION_NOT_FOUND
ARGOCD_LIST_FAILED
ARGOCD_GET_FAILED
ARGOCD_SYNC_FAILED
ARGOCD_OPERATION_IN_PROGRESS
ARGOCD_RESOURCE_NOT_PERMITTED
ARGOCD_APPLICATION_DEGRADED
ARGOCD_APPLICATION_OUT_OF_SYNC
ARGOCD_UNKNOWN_ERROR
```

Current APIs may still expose broader validation or workflow error codes. Error refinements should preserve backward compatibility where possible.

## 14.2 Normalized error structure

Example:

```yaml
code: ARGOCD_RESOURCE_NOT_PERMITTED
message: Argo CD resource is not permitted
technicalMessage: resource :ConfigMap is not permitted in project devops-ci-demo
recoverable: true
suggestedAction: Review AppProject namespaceResourceWhitelist.
```

Rules:

- do not expose tokens;
- include technical details only when safe;
- make recoverability explicit where useful;
- record relevant failures as ChangeRequest events.

---

## 15. Argo CD evidence examples

## 15.1 Healthy evidence

```yaml
application: demo-go-color-app
targetEnvironment: dev
syncStatus: Synced
healthStatus: Healthy
revision: <revision>
conditions:
  - type: OrphanedResourceWarning
    message: Application has 5 orphaned resources
diagnostics:
  argocdSynced: true
  argocdHealthy: true
  warnings:
    - OrphanedResourceWarning: Application has 5 orphaned resources
```

## 15.2 Non-healthy evidence

```yaml
application: demo-go-color-app
targetEnvironment: dev
syncStatus: OutOfSync
healthStatus: Degraded
runtimeStatus: DeploymentDegraded
conditions:
  - type: Error
    message: Example condition message
```

---

## 16. API endpoints related to Argo CD

### Applications

```text
GET /api/v1/applications
GET /api/v1/applications/{name}
GET /api/v1/applications/{name}/resources
GET /api/v1/applications/{name}/history
GET /api/v1/applications/{name}/runtime
```

### ChangeRequest deployment and evidence

```text
POST /api/v1/changes/{id}/check-deployment
POST /api/v1/changes/{id}/collect-evidence
GET  /api/v1/changes/{id}/evidence
GET  /api/v1/changes/{id}/evidence/deployment
```

Legacy endpoint references without `/api/v1` should be considered historical and updated during documentation migration.

---

## 17. Security requirements specific to Argo CD

- Do not log `ARGOCD_AUTH_TOKEN`.
- Do not return Argo CD token values through API.
- Do not persist Argo CD token values in PostgreSQL.
- Do not include Argo CD token values in evidence.
- Use explicit timeouts.
- Use TLS strict mode.
- Use a dedicated service account or Argo CD account token.
- Rotate token immediately if exposed.
- Keep the token permission scope minimal.

---

## 18. Testing strategy

## 18.1 Unit tests

Test:

- raw Application payload mapping;
- sync and health status mapping;
- condition preservation;
- warning handling;
- error mapping;
- TLS utility behavior where applicable.

## 18.2 Runtime validation

Runtime validation should cover:

- `GET /api/v1/applications/{name}` returns the expected application;
- deployment check maps `Synced/Healthy` correctly;
- `OrphanedResourceWarning` is preserved;
- deployment evidence contains Argo CD state;
- TLS strict mode works with the mounted trust bundle;
- token rotation does not break runtime access after rollout.

## 18.3 Failure scenarios

Scenarios to validate or simulate:

- invalid token;
- expired token;
- missing Application;
- TLS trust failure;
- `OutOfSync` application;
- `Degraded` application;
- AppProject restriction;
- Argo CD API unavailable.

---

## 19. Multi-environment direction

The future environment-aware architecture will resolve Argo CD Application names through the environment catalog.

Example mapping:

```text
demo-go-color-app + dev        -> demo-go-color-app-dev
demo-go-color-app + staging    -> demo-go-color-app-staging
demo-go-color-app + production -> demo-go-color-app-production
```

The Control Plane should resolve Argo CD integration context from:

```text
applicationName
targetEnvironment
environment configuration
```

Production-related Argo CD workflows must remain disabled until production guardrails, RBAC/AuthZ and evidence policies are validated.

---

## 20. Relationship with other documents

This document informs and is informed by:

- `docs/05-architecture.md`;
- `docs/13-api-design.md`;
- `docs/04-non-functional-requirements.md`;
- `docs/multi-environment-model.md`;
- `docs/environment-configuration-model.md`;
- `docs/change-promotion-model.md`;
- `docs/adr/ADR-0002-argocd-as-gitops-engine.md`;
- `docs/runbooks/secrets-rotation.md`;
- `docs/runbooks/operability-health-check.md`.

---

## 21. Completion checklist

The Argo CD integration baseline is considered ready when:

- Argo CD configuration is loaded safely;
- token is read from Secret and not logged;
- TLS strict mode works;
- application detail can be retrieved;
- deployment check maps state correctly;
- warnings are preserved;
- deployment evidence includes Argo CD state;
- errors are normalized;
- evidence is persisted in PostgreSQL;
- UI shows the relevant deployment and evidence information;
- runtime validation is documented.

---

## 22. Key message

The Argo CD integration must be robust, secure and simple.

Argo CD provides GitOps reconciliation and application health insight. The DevOps Control Plane adds value by correlating Argo CD status with:

```text
GitLab workflow
Tekton validation
OpenShift runtime evidence
ChangeRequest history
PostgreSQL evidence
Web UI visibility
```

The integration must preserve Argo CD semantics while making application state understandable, auditable and operationally useful.

---

## 23. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial Argo CD integration document in Italian. |
| 2026-07-06 | 0.2 | Rewritten in English and refreshed to align with the current Argo CD adapter, TLS strict, deployment check, evidence and multi-environment baseline. |
