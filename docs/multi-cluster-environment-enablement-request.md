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
