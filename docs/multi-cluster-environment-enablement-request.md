# Multi-cluster Environment Enablement Request

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** Multi-cluster environment enablement request
- **Phase:** 15.1.1 — Add multi-cluster environment enablement request
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Current baseline commit:** `f96d751` — `Document environment catalog UI action results`
- **Related baseline commits:**
  - `a4e9b28` — `Document multi-environment architecture decision`
  - `9664329` — `Document environment configuration model`
  - `e3aecdf` — `Add environment catalog baseline`
  - `041608d` — `Use environment catalog for UI action visibility`
  - `f96d751` — `Document environment catalog UI action results`
- **Language:** English
- **Status:** Request and planning document

---

## 1. Purpose

This document defines the technical request and target architecture required to evolve the DevOps Control Plane from the current single-cluster operational baseline into a multi-cluster, multi-environment control plane.

The current lab has only one OpenShift cluster available. That cluster is sufficient for the current development baseline and for limited simulation. However, the target architecture must not be limited to multiple namespaces inside one OpenShift cluster.

The target architecture is:

```text
One DevOps Control Plane instance
managing multiple OpenShift clusters
mapped to logical environments:
  dev
  staging
  production
```

The functional mapping is:

```text
dev        -> development environment
staging    -> collaudo
production -> produzione
```

This document is the formal request to prepare the platform, configuration model, security model and validation plan needed to support that architecture.

---

## 2. Current state

### 2.1 Current OpenShift availability

At the time of writing, only one OpenShift cluster is available for the DevOps Control Plane lab and runtime validation.

The current operational environment maps to:

```text
targetEnvironment=dev
```

The current runtime namespace used by the demo application and technical workflows is:

```text
devops-ci-demo
```

### 2.2 Current DevOps Control Plane runtime

The DevOps Control Plane currently has the following baseline capabilities:

```text
Backend Go service
PostgreSQL persistence
ChangeRequest lifecycle and audit events
GitLab technical workflow integration
Tekton validation integration
Argo CD deployment check integration
Kubernetes/OpenShift runtime evidence collection
Server-side Web UI
OpenShift OAuth Proxy integration
Role-based AuthZ
Environment Catalog baseline
```

### 2.3 Current Environment Catalog baseline

The current Environment Catalog baseline is implemented in code and represented in the repository with a ConfigMap manifest.

Current environment states:

```text
dev         configured, enabled
staging     configured, disabled
production  configured, disabled
unknown-env not configured
```

Current runtime create behavior:

```text
missing targetEnvironment -> HTTP 201, default dev
targetEnvironment=dev -> HTTP 201
targetEnvironment=staging -> HTTP 422, currently disabled
targetEnvironment=production -> HTTP 422, currently disabled
targetEnvironment=unknown-env -> HTTP 422, not configured
```

### 2.4 Current UI action visibility baseline

Current UI behavior after Phase 14.5:

```text
dev:
  technical actions visible

production disabled:
  warning visible
  technical actions hidden

unknown-env not configured:
  warning visible
  technical actions hidden
```

---

## 3. Target architecture

### 3.1 Primary architecture decision

The target model is:

```text
Single DevOps Control Plane instance
multi-environment aware
multi-cluster ready
```

The DevOps Control Plane must manage environments that may point to separate OpenShift clusters.

The target architecture is not simply:

```text
one OpenShift cluster
multiple namespaces
```

That single-cluster, multi-namespace pattern can be used only as a temporary lab simulation when additional clusters are not yet available.

### 3.2 Target environment-to-cluster mapping

Target mapping:

```text
dev:
  cluster: OpenShift development cluster
  status: enabled first

staging:
  cluster: OpenShift staging/collaudo cluster
  status: disabled initially, then enabled after platform readiness and validation

production:
  cluster: OpenShift production cluster
  status: disabled until production gates are approved
```

### 3.3 Important terminology

Repository and runtime terminology must use English environment names:

```text
dev
staging
production
```

Italian functional terms may be used in explanatory prose only:

```text
collaudo = staging
produzione = production
```

---

## 4. Required model evolution

### 4.1 Current limitation

The current configuration model still contains several single-cluster settings, such as:

```text
KUBERNETES_API_URL
KUBERNETES_CA_FILE
KUBERNETES_NAMESPACE
TEKTON_NAMESPACE
```

These settings are acceptable as the current `dev` default, but they are not sufficient as a long-term multi-cluster model.

### 4.2 Required target model

The target runtime model must resolve technical integration settings from the selected environment.

Target resolution flow:

```text
ChangeRequest.targetEnvironment
  -> Environment Catalog
  -> Cluster reference
  -> Cluster Registry
  -> Kubernetes/OpenShift client configuration
  -> Tekton namespace and pipeline configuration
  -> Argo CD application mapping
  -> GitOps path and branch mapping
  -> Evidence collection target
```

---

## 5. Environment Catalog target

The Environment Catalog must evolve from basic enabled/disabled semantics into cluster-aware environment metadata.

### 5.1 Target fields

Each environment definition should include at least:

```text
name
displayName
enabled
category
description
clusterName
clusterAPIURL reference or clusterRef
kubernetesNamespace
tektonNamespace
argocdApplicationName
gitTargetBranch
validationPath
allowTechnicalActions
allowPromotionSource
allowPromotionTarget
requiredApproverGroup
operatorGroup
viewerGroup
adminGroup
```

### 5.2 Target catalog example

```yaml
environments:
  - name: dev
    displayName: Development
    enabled: true
    category: development
    description: Active development environment.
    clusterName: ocp-dev
    kubernetesNamespace: devops-ci-demo
    tektonNamespace: devops-ci-demo
    argocdApplicationName: demo-go-color-app-dev
    gitTargetBranch: main
    validationPath: apps/demo-go-color-app/overlays/dev
    allowTechnicalActions: true
    allowPromotionSource: true
    allowPromotionTarget: false
    viewerGroup: devops-control-plane-dev-viewers
    operatorGroup: devops-control-plane-dev-operators
    approverGroup: devops-control-plane-dev-approvers
    adminGroup: devops-control-plane-dev-admins

  - name: staging
    displayName: Staging
    enabled: false
    category: preproduction
    description: Controlled staging environment for pre-production validation.
    clusterName: ocp-staging
    kubernetesNamespace: devops-ci-staging
    tektonNamespace: devops-ci-staging
    argocdApplicationName: demo-go-color-app-staging
    gitTargetBranch: main
    validationPath: apps/demo-go-color-app/overlays/staging
    allowTechnicalActions: false
    allowPromotionSource: false
    allowPromotionTarget: true
    viewerGroup: devops-control-plane-staging-viewers
    operatorGroup: devops-control-plane-staging-operators
    approverGroup: devops-control-plane-staging-approvers
    adminGroup: devops-control-plane-staging-admins

  - name: production
    displayName: Production
    enabled: false
    category: production
    description: Production environment. Disabled until production gates are approved.
    clusterName: ocp-production
    kubernetesNamespace: devops-ci-production
    tektonNamespace: ""
    argocdApplicationName: demo-go-color-app-production
    gitTargetBranch: main
    validationPath: apps/demo-go-color-app/overlays/production
    allowTechnicalActions: false
    allowPromotionSource: false
    allowPromotionTarget: true
    viewerGroup: devops-control-plane-production-viewers
    operatorGroup: devops-control-plane-production-operators
    approverGroup: devops-control-plane-production-approvers
    adminGroup: devops-control-plane-production-admins
```

### 5.3 Security note

The Environment Catalog must not store tokens, passwords, private keys or bearer credentials.

The Environment Catalog may contain non-secret references, such as:

```text
clusterName
namespace
Argo CD application name
GitOps path
RBAC group names
Secret reference names
ConfigMap reference names
```

Secrets must remain in Kubernetes/OpenShift Secret resources or external secret management systems.

---

## 6. Cluster Registry target

A multi-cluster DevOps Control Plane requires a Cluster Registry or equivalent configuration model.

### 6.1 Required Cluster Registry fields

Each cluster entry should include:

```text
clusterName
apiURL
caBundleConfigMapRef
tokenSecretRef
connectionMode
defaultNamespace
allowedNamespaces
description
enabled
```

### 6.2 Target Cluster Registry example

```yaml
clusters:
  - name: ocp-dev
    apiURL: https://api.dev.example:6443
    caBundleConfigMapRef: dcp-cluster-ocp-dev-ca
    tokenSecretRef: dcp-cluster-ocp-dev-token
    defaultNamespace: devops-ci-demo
    allowedNamespaces:
      - devops-ci-demo
    enabled: true

  - name: ocp-staging
    apiURL: https://api.staging.example:6443
    caBundleConfigMapRef: dcp-cluster-ocp-staging-ca
    tokenSecretRef: dcp-cluster-ocp-staging-token
    defaultNamespace: devops-ci-staging
    allowedNamespaces:
      - devops-ci-staging
    enabled: false

  - name: ocp-production
    apiURL: https://api.production.example:6443
    caBundleConfigMapRef: dcp-cluster-ocp-production-ca
    tokenSecretRef: dcp-cluster-ocp-production-token
    defaultNamespace: devops-ci-production
    allowedNamespaces:
      - devops-ci-production
    enabled: false
```

### 6.3 Secret handling

For each managed OpenShift cluster, credentials must be stored in dedicated Secret resources.

Example logical Secret names:

```text
dcp-cluster-ocp-dev-token
dcp-cluster-ocp-staging-token
dcp-cluster-ocp-production-token
```

No token value must be printed in operational output, evidence, logs or documentation.

### 6.4 CA handling

For each managed OpenShift cluster, the DevOps Control Plane must trust the target cluster API server certificate.

Possible reference objects:

```text
dcp-cluster-ocp-dev-ca
dcp-cluster-ocp-staging-ca
dcp-cluster-ocp-production-ca
```

The trust model must stay application-scoped and must not require cluster-wide OpenShift CA/proxy changes.

---

## 7. OpenShift platform requirements

### 7.1 Required clusters

Target architecture requires three logical cluster targets:

```text
ocp-dev
ocp-staging
ocp-production
```

For the current lab, only the current cluster may exist. That is acceptable for simulation, but the design and configuration must remain multi-cluster ready.

### 7.2 Required namespaces per cluster

Recommended namespaces:

```text
ocp-dev:
  devops-ci-demo

ocp-staging:
  devops-ci-staging

ocp-production:
  devops-ci-production
```

If only one cluster is available temporarily, equivalent namespaces may be created in the current cluster for simulation. Such simulation must be clearly marked as temporary and not considered the final production topology.

### 7.3 RBAC per target cluster

For dev and staging technical automation, the DevOps Control Plane service identity requires least-privilege permissions.

Minimum permissions:

```text
create pipelineruns.tekton.dev
get pipelineruns.tekton.dev
list pipelineruns.tekton.dev
get taskruns.tekton.dev
list taskruns.tekton.dev
get deployments.apps
list pods
get services
get routes.route.openshift.io
```

Explicitly denied permissions:

```text
get secrets
list secrets
create secrets
update secrets
delete pods
update deployments.apps
delete deployments.apps
```

For production, the initial recommended permission model is read-only evidence collection only.

Production technical mutation permissions should not be granted until production gates are formally approved.

---

## 8. GitOps repository requirements

### 8.1 Target repository structure

Recommended GitOps layout:

```text
apps/demo-go-color-app/base
apps/demo-go-color-app/overlays/dev
apps/demo-go-color-app/overlays/staging
apps/demo-go-color-app/overlays/production
```

The DevOps Control Plane should resolve `validationPath` through the Environment Catalog.

### 8.2 Branching model

Recommended branch model remains:

```text
change/CHG-YYYY-NNNN
```

The target branch can remain:

```text
main
```

Promotion must not be modeled as uncontrolled direct runtime mutation.

Promotion must be represented through correlated ChangeRequests and GitOps changes.

### 8.3 Promotion model

Target promotion flow:

```text
dev ChangeRequest
  -> related staging ChangeRequest
    -> related production ChangeRequest
```

The relationship between ChangeRequests must preserve traceability.

---

## 9. Argo CD requirements

### 9.1 Application model

Recommended Argo CD Applications:

```text
demo-go-color-app-dev
demo-go-color-app-staging
demo-go-color-app-production
```

### 9.2 Cluster targeting

Each Argo CD Application must target the corresponding OpenShift cluster and namespace.

```text
demo-go-color-app-dev:
  cluster: ocp-dev
  namespace: devops-ci-demo

 demo-go-color-app-staging:
  cluster: ocp-staging
  namespace: devops-ci-staging

 demo-go-color-app-production:
  cluster: ocp-production
  namespace: devops-ci-production
```

### 9.3 Production state

Production Argo CD Application may be prepared but must not imply that the DevOps Control Plane enables production technical actions.

The Environment Catalog remains the source of whether the DevOps Control Plane allows actions for production.

---

## 10. Tekton requirements

### 10.1 Dev

Current Tekton baseline:

```text
cluster: current dev cluster
namespace: devops-ci-demo
pipeline: validate-gitops
```

### 10.2 Staging

Staging should have an equivalent validation pipeline:

```text
cluster: ocp-staging
namespace: devops-ci-staging
pipeline: validate-gitops
validationPath: apps/demo-go-color-app/overlays/staging
```

### 10.3 Production

Production should initially be disabled for technical automation.

Recommended initial production policy:

```text
allowTechnicalActions: false
create PipelineRun: not allowed
read-only evidence: allowed only after explicit RBAC setup
```

---

## 11. Argo CD and Tekton multi-cluster strategy options

### 11.1 Option A — Central Argo CD managing multiple clusters

One Argo CD instance manages Applications targeting multiple OpenShift clusters.

Pros:

```text
centralized deployment visibility
single Argo CD API integration for DevOps Control Plane
consistent GitOps management
```

Cons:

```text
requires Argo CD cluster credentials for each target cluster
production security controls must be strict
```

### 11.2 Option B — Per-cluster Argo CD instances

Each OpenShift cluster has its own Argo CD instance.

Pros:

```text
stronger cluster isolation
production Argo CD separated from development/collaudo
```

Cons:

```text
DevOps Control Plane must support multiple Argo CD base URLs and tokens
more complex catalog and secret mapping
```

### 11.3 Recommendation

For the current roadmap, the DevOps Control Plane should be designed to support both models through catalog references.

The first operational implementation can use the existing Argo CD model, but the catalog should not hardcode the assumption that a single Argo CD instance always manages all clusters.

---

## 12. AuthZ and group model requirements

### 12.1 Current global groups

Current groups:

```text
devops-control-plane-viewers
devops-control-plane-operators
devops-control-plane-approvers
devops-control-plane-admins
```

### 12.2 Required environment-aware groups

Recommended groups:

```text
devops-control-plane-dev-viewers
devops-control-plane-dev-operators
devops-control-plane-dev-approvers
devops-control-plane-dev-admins

devops-control-plane-staging-viewers
devops-control-plane-staging-operators
devops-control-plane-staging-approvers
devops-control-plane-staging-admins

devops-control-plane-production-viewers
devops-control-plane-production-operators
devops-control-plane-production-approvers
devops-control-plane-production-admins
```

### 12.3 Authorization target

Authorization decisions should eventually consider:

```text
user identity
user groups
action type
target environment
target cluster
ChangeRequest lifecycle status
technical runtime status
Environment Catalog policy
```

### 12.4 Initial compatibility

Global groups may remain as a compatibility layer, but environment-specific groups should become the preferred model for multi-environment operations.

---

## 13. Multi-user and multi-environment test matrix

### 13.1 Roles

The test matrix must cover:

```text
viewer
operator
approver
admin
no-role
```

### 13.2 Environments

Before staging enablement:

```text
dev enabled
staging disabled
production disabled
unknown-env not configured
```

After controlled staging enablement:

```text
dev enabled
staging enabled
production disabled
unknown-env not configured
```

### 13.3 Cluster modes

The test matrix must distinguish:

```text
single-cluster simulation mode
multi-cluster real mode
```

Single-cluster simulation validates logical routing and policy.

Multi-cluster real mode validates actual API connectivity, credentials, RBAC, Argo CD, Tekton and evidence across clusters.

### 13.4 Initial matrix before staging enablement

Viewer:

```text
read dev -> allowed
read staging historical/disabled records -> allowed if authorized to read
technical action dev -> denied
technical action staging -> denied or not visible
technical action production -> denied or not visible
```

Operator:

```text
read dev -> allowed
technical action dev -> allowed
technical action staging disabled -> not visible / not allowed
technical action production disabled -> not visible / not allowed
approval action -> denied
```

Approver:

```text
read dev -> allowed
approval action dev -> allowed where lifecycle permits
technical action dev -> denied
technical action staging disabled -> denied or not visible
technical action production disabled -> denied or not visible
```

Admin:

```text
read all configured environments -> allowed
action dev -> allowed
action staging disabled -> blocked by Environment Catalog semantics
action production disabled -> blocked by Environment Catalog semantics
```

No-role:

```text
read -> denied
action -> denied
UI dashboard -> denied
```

---

## 14. Staging enablement criteria

Staging may be enabled only when all of the following conditions are true:

```text
staging OpenShift cluster identified
network reachability from DCP to staging API server verified
staging CA bundle available to DCP
staging token Secret available to DCP without printing token values
staging namespace created
staging RBAC least privilege applied
staging GitOps overlay exists
staging Argo CD Application exists
staging Tekton validation pipeline exists
staging Environment Catalog entry completed
staging UI visibility validated
staging multi-user AuthZ validated
staging evidence collection validated
```

Only after these checks should the Environment Catalog be changed to:

```text
staging.enabled=true
staging.allowTechnicalActions=true
```

---

## 15. Production enablement gates

Production must remain disabled until strict gates are approved.

Minimum gates:

```text
production OpenShift cluster identified
production namespace created
production RBAC reviewed and approved
production Argo CD Application created or planned
production GitOps overlay reviewed
production credentials stored securely
production read-only evidence validated
production approval group defined
production operator group restricted
mandatory promotion from staging defined
mandatory staging evidence required
rollback runbook available
maintenance window policy available
audit and retention policy available
production go/no-go checklist approved
```

Only after explicit approval should production change from:

```text
enabled=false
allowTechnicalActions=false
```

to a more permissive state.

---

## 16. Required implementation roadmap

### 16.1 Phase 15.2 — Runtime Environment Catalog ConfigMap loading

Implement loading of the Environment Catalog from a mounted ConfigMap or file.

Expected behavior:

```text
valid catalog loaded -> use runtime catalog
catalog missing -> fallback to conservative in-code catalog
catalog invalid -> fail closed or fallback to conservative catalog with clear warning
```

### 16.2 Phase 15.3 — Cluster Registry baseline

Introduce Cluster Registry model and repository manifest.

Expected behavior:

```text
environment.clusterName -> cluster registry entry
cluster registry entry -> API URL / CA ref / token ref / allowed namespaces
```

### 16.3 Phase 15.4 — Multi-cluster client selection

Refactor Kubernetes/Tekton adapters to select clients based on target environment and cluster.

### 16.4 Phase 15.5 — Controlled staging enablement

Enable staging after prerequisites are ready.

### 16.5 Phase 15.6 — Multi-user / multi-environment / multi-cluster validation

Run full role and environment matrix.

### 16.6 Phase 15.7 — Production readiness gates

Keep production disabled while documenting and validating required production controls.

---

## 17. Deliverables requested from platform owners

The DevOps Control Plane engineering work requires the following platform inputs.

### 17.1 For staging

```text
OpenShift staging cluster API URL
staging CA bundle
staging ServiceAccount/token or approved authentication method
staging namespace
staging RBAC Role/RoleBinding
staging Argo CD Application or Argo CD cluster registration
staging Tekton pipeline namespace and pipeline
staging GitOps overlay path
staging users/groups for viewer/operator/approver/admin roles
network reachability confirmation from DCP to staging API server
```

### 17.2 For production

```text
OpenShift production cluster API URL
production CA bundle
production ServiceAccount/token or approved authentication method
production namespace
production read-only RBAC baseline
production Argo CD Application or Argo CD cluster registration plan
production GitOps overlay path
production users/groups for viewer/operator/approver/admin roles
production approval policy
production rollback and maintenance runbook references
network reachability confirmation from DCP to production API server
```

---

## 18. Acceptance criteria for Phase 15.1.1

This phase is complete when this document is committed and pushed to the repository.

The document must clearly state that:

```text
only one OpenShift cluster is currently available
single-cluster multi-namespace is only a temporary simulation model
the target architecture is multi-cluster
dev, staging and production may map to distinct OpenShift clusters
Environment Catalog must become cluster-aware
a Cluster Registry or equivalent model is required
staging must be enabled before production
production must remain disabled until production gates are approved
multi-user and multi-environment tests must include multi-cluster readiness
```

---

## 19. Final recommendation

Proceed in this order:

```text
1. Commit this multi-cluster enablement request.
2. Implement runtime Environment Catalog ConfigMap loading.
3. Introduce Cluster Registry baseline.
4. Refactor technical adapters for environment-to-cluster selection.
5. Prepare staging platform prerequisites.
6. Enable staging with controlled tests.
7. Execute multi-user and multi-environment validation.
8. Keep production disabled until production gates are approved.
```

This order keeps the DevOps Control Plane secure, auditable and production-ready while enabling progressive multi-cluster expansion.

## 15.8.9.7 — Document namespace-isolated Tekton validation path fix and final runtime validation

Status: Completed
Date: 2026-07-09
Commit: `79194cd` — `Add environment-specific Tekton validation paths`

### Objective

Document the final namespace-isolated Tekton validation performed for the simulated staging and simulated production environments running on the current development OpenShift cluster.

This phase closes the runtime validation loop for the namespace-isolated topology:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

The validation confirmed that Tekton PipelineRuns are created in the environment-specific Tekton namespace and that the GitOps validation path is resolved from the Environment Catalog instead of using a single legacy global fallback path.

### Root cause addressed

During the first staging and production validation attempts, Tekton received the legacy global validation path:

```text
apps/demo-go-color-app
```

That path contained manifests and kustomization output targeting the development namespace:

```text
devops-ci-demo
```

As a result, the staging and production Tekton ServiceAccounts attempted to validate resources in the development namespace and correctly received RBAC authorization failures.

The issue was not resolved by granting cross-namespace permissions. The correct fix was to make the validation path environment-specific.

### Runtime fix

The Environment Catalog now supports an environment-specific `validationPath` field.

The value is propagated through:

1. `EnvironmentDefinition`
2. `TechnicalRuntimeTarget`
3. Tekton PipelineRun creation
4. Tekton validation evidence reporting

The legacy `TEKTON_VALIDATION_PATH` remains available only as a fallback when an environment does not define `validationPath`.

### Effective validation paths

```text
dev        -> apps/demo-go-color-app
staging    -> apps/demo-go-color-app/overlays/staging
production -> apps/demo-go-color-app/overlays/production
```

### Runtime validation evidence

The final namespace-isolated Tekton validation completed successfully.

#### Staging

```text
ChangeRequest: CHG-2026-0049
Namespace: devops-ci-staging
PipelineRun: devops-cp-validate-chg-2026-0049-nd7rm
Result: Succeeded
check-validation HTTP status: 202
failedTaskCount: 0
Evidence sanitized: true
```

#### Production

```text
ChangeRequest: CHG-2026-0050
Namespace: devops-ci-production
PipelineRun: devops-cp-validate-chg-2026-0050-8wqtv
Result: Succeeded
check-validation HTTP status: 202
failedTaskCount: 0
Evidence sanitized: true
```

### Validation commands executed

The relevant test and runtime checks completed successfully:

```text
go test ./internal/app ./cmd/devops-control-plane ./internal/adapters/tekton
```

Final repository state after commit and push:

```text
git status --short
```

Expected result:

```text
<empty>
```

### Outcome

The namespace-isolated staging and production Tekton validation is now operational on the current development OpenShift cluster. The DevOps Control Plane can create and check Tekton validation PipelineRuns for all three simulated environments without requiring cross-namespace RBAC shortcuts.

## Phase 15.8.10 — UI runtime evidence alignment and multi-environment visibility closure

**Status:** Completed  
**Completion date:** 2026-07-09  
**Code commit:** `58805ef` — `Align UI with namespace-isolated runtime evidence`

### Summary

Phase 15.8.10 closed the Web UI alignment gap after the namespace-isolated `dev`, `staging` and `production` runtime validation flows were completed on the shared `ocp-dev` OpenShift cluster.

The phase did not change the backend execution model. It made the UI accurately reflect the runtime evidence already produced by the DevOps Control Plane.

### Completed outcomes

- The dashboard now selects the most recent `ChangeRequest` instead of the historical hardcoded `CHG-2026-0005`.
- The static `Environment dev` topbar placeholder was removed.
- The topbar now shows `Environments / Namespaces` for:
  - `dev` -> `devops-ci-demo`;
  - `staging` -> `devops-ci-staging`;
  - `production` -> `devops-ci-production`.
- The user indicator is displayed in an adjacent topbar segment aligned with the environment summary.
- The `ChangeRequest` detail view now loads all evidence types for the selected change.
- The UI renders Tekton validation evidence when available.
- The Tekton validation card shows PipelineRun, Tekton namespace, Pipeline, Git revision, validation path, status, reason, failed task count and sanitization state.

### Runtime validation

The UI was validated with namespace-isolated staging and production records.

Validated staging record:

- `CHG-2026-0049`;
- Tekton namespace: `devops-ci-staging`;
- PipelineRun: `devops-cp-validate-chg-2026-0049-nd7rm`;
- validation path: `apps/demo-go-color-app/overlays/staging`;
- failed task count: `0`;
- evidence sanitized: `true`.

Validated production record:

- `CHG-2026-0050`;
- Tekton namespace: `devops-ci-production`;
- PipelineRun: `devops-cp-validate-chg-2026-0050-8wqtv`;
- validation path: `apps/demo-go-color-app/overlays/production`;
- failed task count: `0`;
- evidence sanitized: `true`.

Automated validation completed successfully with:

`go test ./internal/api ./internal/app ./cmd/devops-control-plane`

### Multi-cluster objective alignment

This phase preserves the broader multi-cluster objective.

The current topology remains intentionally namespace-isolated on the available `ocp-dev` cluster:

- `dev` -> `ocp-dev` / `devops-ci-demo`;
- `staging` -> `ocp-dev` / `devops-ci-staging`;
- `production` -> `ocp-dev` / `devops-ci-production`.

This is still a simulated multi-environment model, not yet a physical multi-cluster deployment.

The UI now makes the environment-to-namespace mapping explicit while the backend remains aligned with the provider-aware runtime design, environment catalog, cluster registry, runtime target resolution and controlled Secret-loader/factory enablement model.

### Next direction

The next work should continue toward controlled real multi-cluster enablement when an additional non-production cluster becomes available.

Until then, the namespace-isolated topology on `ocp-dev` remains the validated proving ground for environment-aware UI behavior, runtime target selection, evidence sanitization, provider-aware clients and future physical multi-cluster onboarding.

## Phase 15.8.11.1 — Final namespace-isolated runtime smoke matrix

**Status:** Completed  
**Validation date:** 2026-07-09  
**Evidence directory:** `/tmp/dcp-15-8-11-1-20260709-145735`

### Purpose

This validation step captured the final namespace-isolated runtime smoke matrix for the current multi-environment baseline running on the available `ocp-dev` OpenShift cluster.

The goal was to confirm that the simulated `dev`, `staging` and `production` environments remain operational after the latest UI runtime evidence alignment and documentation updates.

### Validated topology

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

### Runtime smoke results

The following runtime checks completed successfully:

- DevOps Control Plane readiness endpoint returned HTTP `200`.
- Dashboard UI returned HTTP `200`.
- Argo CD Applications were `Synced` and `Healthy` for all three environments.
- `demo-go-color-app` deployment was ready in all three namespaces.
- Route `/healthz` returned HTTP `200` for all three environments.
- Final staging Tekton PipelineRun was `Succeeded`.
- Final production Tekton PipelineRun was `Succeeded`.
- UI detail pages for `CHG-2026-0049` and `CHG-2026-0050` returned HTTP `200`.

### Detailed validation evidence

Argo CD status:

- `dev`: `sync=Synced`, `health=Healthy`, revision `2d66c51831c282856b27397bcd3e0aeba51e435c`
- `staging`: `sync=Synced`, `health=Healthy`, revision `2d66c51831c282856b27397bcd3e0aeba51e435c`
- `production`: `sync=Synced`, `health=Healthy`, revision `2d66c51831c282856b27397bcd3e0aeba51e435c`

Deployment readiness:

- `dev`: `ready=2/2`, `available=2`, `updated=2`
- `staging`: `ready=2/2`, `available=2`, `updated=2`
- `production`: `ready=2/2`, `available=2`, `updated=2`

Route health checks:

- `dev_healthz_http=200`
- `staging_healthz_http=200`
- `production_healthz_http=200`

Tekton validation:

- `staging`: `devops-cp-validate-chg-2026-0049-nd7rm`, `status=True`, `reason=Succeeded`
- `production`: `devops-cp-validate-chg-2026-0050-8wqtv`, `status=True`, `reason=Succeeded`

UI validation:

- `CHG-2026-0049` detail page returned HTTP `200`
- `CHG-2026-0050` detail page returned HTTP `200`

### Conclusion

The namespace-isolated multi-environment baseline is stable across `dev`, `staging` and `production` on `ocp-dev`.

This confirms that the current simulated multi-environment topology is ready to be used as the controlled baseline for the next planning steps toward real multi-cluster enablement.

This does not replace the future physical multi-cluster objective. It confirms that the environment catalog, GitOps overlays, Argo CD Applications, Tekton validation paths, runtime evidence model and UI visibility are aligned before onboarding a real non-production cluster.

## Phase 15.9.1 — Physical cluster availability constraint and multi-cluster readiness closure strategy

**Status:** Completed as planning and closure strategy  
**Date:** 2026-07-09  
**Baseline tag:** `namespace-isolated-baseline-20260709`  
**Baseline commit:** `af6ddb3` — `Document phase 15.8.11.1 runtime smoke matrix`

### Purpose

Phase 15.9.1 documents the infrastructure constraint that currently prevents physical multi-cluster runtime validation.

At this point, only the `ocp-dev` OpenShift cluster is available. There is no separate real non-production cluster available and there is no separate real production cluster available for onboarding.

The project must therefore avoid remaining open indefinitely while waiting for unavailable infrastructure. The correct closure strategy is to declare the DevOps Control Plane as multi-cluster-ready from an architecture, configuration, runtime targeting, Secret reference, client factory and operational readiness perspective, while explicitly deferring physical cross-cluster runtime validation until an additional OpenShift cluster becomes available.

### Current validated topology

The currently validated topology remains namespace-isolated on `ocp-dev`:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

This topology is no longer treated as an accidental workaround. It is the official validated fallback topology until real clusters become available.

### Infrastructure constraint

The following physical multi-cluster targets are currently unavailable:

- a real non-production OpenShift cluster for `staging`;
- a real production OpenShift cluster for `production`.

Because these clusters are unavailable, the project cannot complete a physical cross-cluster runtime smoke test at this time.

This is an infrastructure availability constraint, not a blocker in the DevOps Control Plane architecture.

### Multi-cluster readiness closure strategy

The DevOps Control Plane is considered multi-cluster-ready when the following capabilities are present and validated or safely gated:

- Environment Catalog supports logical environment modeling.
- Cluster Registry abstraction supports future physical cluster targets.
- Runtime target resolution is environment-aware.
- Kubernetes, Tekton and Argo CD runtime clients are provider-aware by design.
- Secret references are modeled without exposing raw Secret values.
- Runtime Secret loader and concrete factories remain disabled by default unless explicitly enabled.
- Secret reference allow-list validation is available.
- Readiness gates fail closed when required configuration is incomplete or unsafe.
- Validation paths are environment-specific.
- Runtime evidence is environment-aware and sanitized.
- The UI shows the validated environment-to-namespace mapping.
- The namespace-isolated dev, staging and production baseline is fully validated.

### Closure statement

Physical multi-cluster runtime validation is explicitly deferred because no additional OpenShift cluster is currently available.

The multi-cluster readiness work is closed as a validated readiness baseline, not as a physical multi-cluster deployment.

The validated baseline is:

- namespace-isolated;
- multi-environment;
- environment-aware;
- evidence-aware;
- provider-aware in the runtime design;
- ready for future physical cluster onboarding.

### Future real-cluster onboarding condition

When a real additional OpenShift cluster becomes available, onboarding must resume from this baseline.

The first real-cluster onboarding should not redesign the current model. It should only provide the missing infrastructure inputs:

- cluster identity;
- API server URL;
- certificate authority reference;
- token Secret reference;
- target namespace;
- Tekton namespace;
- Argo CD access model;
- minimum RBAC;
- readiness validation;
- rollback plan.

### Required future validation

When a real cluster becomes available, the first physical multi-cluster validation must prove:

- the target environment no longer falls back silently to `ocp-dev`;
- readiness fails closed if Secret references, RBAC or runtime factories are incomplete;
- `create`, `check-deployment`, `validate`, `check-validation` and `collect-evidence` work for the real target environment;
- no raw Secret values appear in logs, evidence or UI;
- the existing `dev` baseline remains unaffected;
- rollback to the namespace-isolated fallback topology is possible and documented.

### Impact on Phase 15

This phase allows Phase 15 to proceed toward closure without waiting for unavailable infrastructure.

Recommended Phase 15 closure wording:

`Phase 15 — Multi-environment / multi-cluster readiness baseline: completed as multi-cluster-ready baseline. Physical cross-cluster runtime validation is deferred by infrastructure availability.`

### Next direction

The next planning work should focus on finalizing the multi-cluster readiness checklist and the deferred real-cluster onboarding contract.

The project should not assume near-term availability of a non-production or production cluster. Future onboarding must be treated as conditional on infrastructure availability.

## Phase 15.9.2.3 — Multi-cluster code readiness test coverage

Status: Completed  
Date: 2026-07-09  
Code commit: `68a8b2e` — `Add simulated external cluster fail-closed tests`

### Purpose

Phase 15.9.2.3 documents the test coverage added to prove that the DevOps Control Plane is being prepared for real multi-cluster execution even though no additional physical OpenShift cluster is currently available.

The goal is to keep progressing on code readiness without depending on unavailable infrastructure.

### Context

The validated runtime baseline remains namespace-isolated on `ocp-dev`:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

Physical cross-cluster runtime validation is deferred by infrastructure availability.

However, code readiness continues. The system must be ready to support a future topology where an environment resolves to a cluster different from `ocp-dev`.

### Test coverage added

The following test file was added:

`internal/app/multicluster_readiness_failclosed_test.go`

The test suite validates a simulated external non-production cluster named:

`ocp-nonprod-simulated`

The tests prove that:

- `staging` can resolve to a cluster different from `ocp-dev`;
- the resolved `TechnicalRuntimeTarget` preserves the external cluster name;
- the runtime target does not silently fall back to `ocp-dev`;
- Kubernetes and Tekton namespaces are resolved from the environment metadata;
- Argo CD application metadata is resolved from the environment metadata;
- environment-specific validation path is preserved;
- runtime provider selection fails closed when the external cluster provider is missing;
- runtime provider selection fails closed when the external cluster provider is configured but disabled.

### Validated fail-closed behavior

The following fail-closed cases are now covered:

- missing runtime provider for `ocp-nonprod-simulated`;
- disabled runtime provider for `ocp-nonprod-simulated`;
- no silent fallback from `staging` to `ocp-dev`.

This behavior is required for safe future multi-cluster enablement.

If a future environment points to a real cluster and the runtime provider is missing or disabled, the DevOps Control Plane must fail explicitly instead of executing against the wrong cluster.

### Automated validation

The following tests were executed successfully:

`go test ./internal/app -count=1`

The broader validation was also executed successfully:

`go test ./internal/api ./internal/app ./cmd/devops-control-plane`

### Multi-cluster readiness impact

This phase strengthens the multi-cluster code-readiness baseline.

The project now has test evidence that the runtime target model can represent a future external cluster and that runtime provider selection does not degrade into unsafe fallback behavior.

This is important because the lack of a second physical cluster should not block code preparation.

### Closure statement

The DevOps Control Plane remains physically validated on a namespace-isolated topology only.

At the same time, the codebase is now more strongly validated for future real multi-cluster onboarding through simulated external-cluster fail-closed tests.

Physical runtime validation remains deferred.

Code readiness continues.

## Phase 15.9.3.2 — Secret reference and runtime factory fail-closed coverage

Status: Completed  
Date: 2026-07-09  
Scope: Multi-cluster code-readiness guardrails

### Purpose

Phase 15.9.3.2 documents the review of Secret reference handling, runtime Secret loading and runtime client factory fail-closed behavior.

The goal is to confirm that the DevOps Control Plane can continue progressing toward code-level multi-cluster readiness even when physical cross-cluster runtime validation is deferred by infrastructure availability.

### Review summary

The review confirmed that the codebase already includes the required safety boundaries for future real multi-cluster onboarding.

The following areas were reviewed:

- runtime Secret reference model;
- Secret reference registry;
- Secret reference validation;
- Secret value loader;
- Kubernetes Secret getter readiness gate;
- allow-list based Secret loading;
- Kubernetes runtime client factory;
- Tekton runtime client factory;
- Argo CD runtime client factory;
- provider-aware runtime factory wiring;
- disabled-by-default runtime factory flags;
- fail-closed fallback behavior.

### Secret reference model

The codebase includes a dedicated Secret reference model based on `RuntimeClientSecretRefs`.

Secret references are modeled as metadata and are attached to runtime provider selections without exposing raw Secret values.

The review confirmed the following elements:

- `RuntimeClientSecretRefs`;
- `RuntimeClientSecretRefsRegistry`;
- `RuntimeClientSecretRefs.Validate`;
- `RuntimeClientProviderRegistry.SelectWithSecretRefs`;
- sanitized `SafeSummary` output.

This supports future real-cluster onboarding because cluster credentials can be referenced without being embedded directly in configuration, logs or evidence.

### Runtime Secret loader

The runtime Secret value loading path is guarded by conservative defaults.

The following protections are present:

- default empty loader;
- disabled Kubernetes Secret value loader;
- allow-list based Kubernetes Secret value loader;
- Kubernetes Secret getter readiness boundary;
- validation of allowed clusters and allowed Secret references;
- fail-closed behavior for missing getter, missing references and missing required keys.

The runtime Secret loader remains disabled by default unless explicitly enabled through configuration.

### Runtime client factories

The following factory abstractions are present:

- Kubernetes runtime client factory;
- Tekton runtime client factory;
- Argo CD runtime client factory.

Each factory has a conservative empty implementation that does not build a client and returns a clear not-configured error.

The following factory enablement flags default to disabled:

- `RuntimeClientFactoryKubernetesEnabled`;
- `RuntimeClientFactoryTektonEnabled`;
- `RuntimeClientFactoryArgoCDEnabled`.

This prevents accidental construction of real external-cluster clients before the operator explicitly enables the required capability.

### Adapter guardrails

The concrete runtime factory adapters include additional safety checks.

The review confirmed guardrails for:

- missing Kubernetes API URL;
- missing Tekton API URL;
- missing Argo CD base URL;
- missing token value;
- unsupported kubeconfig references;
- unsupported raw CA references;
- invalid factory requests.

These guardrails are important because future real-cluster onboarding must not silently accept incomplete or unsafe credential material.

### Fail-closed behavior

The review confirmed the following fail-closed behavior:

- missing runtime provider fails explicitly;
- disabled runtime provider fails explicitly;
- missing Secret references prevent factory-aware client construction;
- missing allow-list entries prevent Secret loading;
- disabled Secret loader does not read Secrets;
- empty factories do not return clients;
- factory builders remain disabled unless global and capability-specific flags are enabled.

This is aligned with the security expectation that a future external cluster target must never silently fall back to `ocp-dev`.

### Automated validation

The following validation completed successfully during the review:

`go test ./internal/app -count=1`

The broader validation also completed successfully:

`go test ./internal/api ./internal/app ./cmd/devops-control-plane`

### Multi-cluster readiness impact

This phase confirms that the DevOps Control Plane has the required fail-closed technical foundations for future real multi-cluster onboarding.

Physical cross-cluster runtime validation remains deferred because no additional OpenShift cluster is currently available.

However, the codebase continues to be prepared for multi-cluster operation through explicit runtime target resolution, provider selection, Secret references, allow-list validation and disabled-by-default client factory enablement.

### Closure statement

Phase 15.9.3.2 confirms that Secret reference and runtime factory guardrails are in place and validated.

The project remains physically validated on the namespace-isolated `ocp-dev` topology, while code-level multi-cluster readiness continues to be actively consolidated.
