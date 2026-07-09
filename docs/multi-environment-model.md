# DevOps Control Plane — Multi-environment Model

## 1. Purpose

This document expands the architectural decision captured in `docs/adr/ADR-0011-multi-environment-model.md` and defines the initial target model for supporting multiple environments in the DevOps Control Plane.

The goal is to prepare the application for scenarios where the same Control Plane manages changes across:

```text
dev
staging
production
```

This document is a design baseline. It does not implement the full model yet.

---

## 2. Selected architectural direction

The selected direction is:

```text
Single DevOps Control Plane instance
Multi-environment target model
Correlated ChangeRequests for promotion
Environment-aware RBAC/AuthZ policy
```

This means there is one central DevOps Control Plane application and database, while each ChangeRequest explicitly targets one environment.

Promotion is represented by correlated ChangeRequests rather than by one monolithic multi-stage ChangeRequest.

---

## 3. Why a single Control Plane instance

A single multi-environment Control Plane provides:

- unified dashboard;
- unified audit trail;
- unified evidence store;
- one operational surface for platform and DevOps teams;
- easier cross-environment visibility;
- direct reuse of the current `targetEnvironment` field;
- centralized production-readiness governance.

The trade-off is that the Control Plane must enforce stronger guardrails:

- environment validation;
- environment-specific authorization;
- production-specific approval requirements;
- explicit integration mapping;
- clear UI environment context.

---

## 4. Environment catalog

The first implementation should introduce an explicit supported environment catalog.

Initial environments:

```text
dev
staging
production
```

Each environment should eventually include:

```text
name
displayName
description
enabled
riskProfile
approvalPolicy
kubernetesNamespace
tektonNamespace
tektonPipelineName
gitOpsPath
argoCdApplicationName
gitlabTargetBranch
allowedViewerGroups
allowedOperatorGroups
allowedApproverGroups
allowedAdminGroups
evidencePolicy
```

For the first increment, this can be represented in ConfigMap-driven configuration.

A future phase can evaluate whether this should move to a database table.

---

## 5. Recommended configuration approach

### 5.1 Initial approach — ConfigMap-driven

For the MVP multi-environment increment, use ConfigMap-driven configuration.

Benefits:

- no database migration required for the first design phase;
- simple repository alignment;
- visible operational configuration;
- easy OpenShift review;
- consistent with the current deployment model.

Example conceptual configuration:

```yaml
environments:
  dev:
    displayName: Development
    kubernetesNamespace: devops-ci-demo
    tektonNamespace: devops-ci-demo
    tektonPipelineName: validate-gitops
    gitOpsPath: apps/demo-go-color-app/overlays/dev
    argoCdApplicationName: demo-go-color-app-dev
    gitlabTargetBranch: main
    approvalPolicy: relaxed

  staging:
    displayName: Staging
    kubernetesNamespace: devops-ci-staging
    tektonNamespace: devops-ci-staging
    tektonPipelineName: validate-gitops
    gitOpsPath: apps/demo-go-color-app/overlays/staging
    argoCdApplicationName: demo-go-color-app-staging
    gitlabTargetBranch: main
    approvalPolicy: required

  production:
    displayName: Production
    kubernetesNamespace: devops-ci-production
    tektonNamespace: devops-ci-production
    tektonPipelineName: validate-gitops
    gitOpsPath: apps/demo-go-color-app/overlays/production
    argoCdApplicationName: demo-go-color-app-production
    gitlabTargetBranch: main
    approvalPolicy: strict
```

This is a conceptual example. The exact format should be defined in Phase 13.2.

---

## 6. ChangeRequest promotion model

The selected promotion model is correlated ChangeRequests.

### 6.1 Current model

A ChangeRequest already includes:

```text
changeNumber
applicationName
targetEnvironment
status
runtimeStatus
requestedBy
```

### 6.2 Future promotion metadata

To support promotion, introduce correlation metadata such as:

```text
promotionGroupID
promotedFromChangeNumber
promotedToChangeNumber
sourceEnvironment
targetEnvironment
```

Possible flow:

```text
CHG-2026-0101
  targetEnvironment: dev
  promotionGroupID: PROMO-2026-0001

CHG-2026-0102
  targetEnvironment: staging
  promotedFromChangeNumber: CHG-2026-0101
  promotionGroupID: PROMO-2026-0001

CHG-2026-0103
  targetEnvironment: production
  promotedFromChangeNumber: CHG-2026-0102
  promotionGroupID: PROMO-2026-0001
```

### 6.3 Why correlated changes

This approach allows:

- separate approval per environment;
- separate evidence per environment;
- separate audit events per environment;
- clear production authorization boundary;
- incremental evolution of the current domain model.

---

## 7. Environment-aware RBAC/AuthZ model

The current roles are:

```text
viewer
operator
approver
admin
```

The multi-environment model should preserve these roles but evaluate them with environment context.

### 7.1 Suggested policy

```text
dev:
  viewer: read
  operator: create, validate, collect evidence, execute standard technical actions
  approver: approve when needed
  admin: all controlled actions

staging:
  viewer: read
  operator: create and validate
  approver: approve and promote
  admin: all controlled actions

production:
  viewer: read
  operator: propose only
  approver: approve production change
  production approver: approve and authorize production execution
  admin: emergency/admin actions with audit
```

### 7.2 Possible OpenShift groups

Existing groups can be reused:

```text
devops-control-plane-viewers
devops-control-plane-operators
devops-control-plane-approvers
devops-control-plane-admins
```

Additional production-specific groups may be added:

```text
devops-control-plane-production-approvers
devops-control-plane-production-operators
```

The exact names should be decided in Phase 13.8.

---

## 8. UI model

The UI should evolve from the current static environment display to a real environment selector.

### 8.1 Environment selector

Possible values:

```text
all
dev
staging
production
```

Expected behavior:

```text
all:
  aggregate dashboard across all environments

dev:
  show only dev ChangeRequests and dev runtime context

staging:
  show only staging ChangeRequests and staging runtime context

production:
  show only production ChangeRequests and production runtime context
```

### 8.2 Dashboard cards

KPI cards should respect the selected environment filter.

Example:

```text
Applications [N]
Completed changes [N]
Running changes [N]
Failed changes [N]
Collected evidence [N]
```

### 8.3 Recent changes

Recent changes should display:

```text
changeNumber
applicationName
targetEnvironment
requestedBy
runtimeStatus
```

The already implemented limit of 5 remains valid.

### 8.4 Change list filters

The `/ui/changes` page should eventually support filters:

```text
environment
requester
application
lifecycle status
runtime status
```

---

## 9. GitOps layout

The recommended GitOps layout is `base + overlays`.

Example:

```text
apps/demo-go-color-app/base
apps/demo-go-color-app/overlays/dev
apps/demo-go-color-app/overlays/staging
apps/demo-go-color-app/overlays/production
```

Benefits:

- common manifests are shared;
- environment-specific differences are explicit;
- Kustomize validation is natural;
- promotion can be represented as controlled changes to overlays.

---

## 10. Argo CD integration model

Argo CD application mapping should become environment-aware.

Example:

```text
demo-go-color-app + dev        -> demo-go-color-app-dev
demo-go-color-app + staging   -> demo-go-color-app-staging
demo-go-color-app + production -> demo-go-color-app-production
```

The DevOps Control Plane should resolve the Argo CD Application based on:

```text
applicationName
targetEnvironment
```

This impacts:

- application list;
- application details;
- deployment checks;
- deployment evidence collection;
- UI links.

---

## 11. Tekton integration model

Tekton validation should become environment-aware.

Potential approaches:

### Option A — One Tekton namespace, parameterized by environment

Example:

```text
TEKTON_NAMESPACE=devops-ci-demo
VALIDATION_PATH=apps/demo-go-color-app/overlays/dev
```

### Option B — One Tekton namespace per environment

Example:

```text
devops-ci-demo
devops-ci-staging
devops-ci-production
```

Recommendation for the first implementation:

```text
Use parameterized validation first.
Evaluate namespace separation before production enforcement.
```

Production may require stricter isolation.

---

## 12. Evidence and audit requirements

Evidence and audit must include environment context.

Required future fields or payload attributes:

```text
changeNumber
applicationName
targetEnvironment
actor
requestedBy
evidenceType
externalRef
createdAt
```

For promotion chains, evidence should also include:

```text
promotionGroupID
promotedFromChangeNumber
```

The UI should make the target environment visible in:

- Change detail;
- Evidence pages;
- Audit log pages;
- Dashboard recent changes;
- Change list.

---

## 13. Security guardrails

Minimum guardrails:

```text
Unknown targetEnvironment is rejected.
Production actions require production-appropriate role.
Production approval cannot be bypassed by dev operator role.
All actions include actor and targetEnvironment in audit events.
Secret values are never displayed.
Environment configuration is reviewed through repository changes.
Production integrations use TLS strict mode.
RBAC remains least-privilege.
```

---

## 14. Suggested implementation roadmap

```text
13.1 — Multi-environment architecture decision
13.2 — Environment configuration model
13.3 — Domain model extension for promotion metadata
13.4 — UI environment selector and filters design
13.5 — GitOps layout alignment for overlays
13.6 — Argo CD environment mapping
13.7 — Tekton validation parameterization
13.8 — Environment-aware RBAC/AuthZ policy
13.9 — Evidence and audit enrichment
13.10 — MVP implementation for dev and staging
13.11 — Production environment hardening
```

---

## 15. Current decision summary

Accepted decisions:

```text
Use one single multi-environment DevOps Control Plane instance.
Use correlated ChangeRequests for promotion.
Use environment-aware RBAC/AuthZ policy.
Start with ConfigMap-driven environment configuration.
Preserve PostgreSQL as the central audit/evidence store.
Keep production guardrails explicit and fail-closed.
```

---

## 16. Next step

The next recommended phase is:

```text
Phase 13.2 — Environment configuration model
```

Expected deliverables:

```text
docs/environment-configuration-model.md
initial ConfigMap design
validation rules
migration impact assessment
implementation plan
```

## Namespace-isolated Tekton validation closure

Status: Completed
Related phase: 15.8.9.7
Commit: `79194cd` — `Add environment-specific Tekton validation paths`

The namespace-isolated multi-environment model has been validated end-to-end for Tekton GitOps validation on the current development OpenShift cluster.

The validated topology is:

```text
dev        -> ocp-dev / devops-ci-demo
staging    -> ocp-dev / devops-ci-staging
production -> ocp-dev / devops-ci-production
```

The final validation confirmed that staging and production no longer validate the root GitOps application path that renders development namespace resources. Instead, each environment validates its own overlay path.

Final successful PipelineRuns:

```text
devops-ci-staging    devops-cp-validate-chg-2026-0049-nd7rm  Succeeded
devops-ci-production devops-cp-validate-chg-2026-0050-8wqtv  Succeeded
```

This completes the namespace-isolated validation baseline before introducing a physically separate non-production cluster.

