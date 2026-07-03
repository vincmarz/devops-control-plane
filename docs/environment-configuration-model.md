# DevOps Control Plane — Environment Configuration Model

## Document metadata

- **Project:** DevOps Control Plane
- **Phase:** 13.2 — Environment configuration model
- **Status:** Draft design baseline for incremental implementation
- **Date:** 2026-07-03
- **Scope:** Technical configuration model for supporting multiple target environments from a single DevOps Control Plane instance
- **Related ADR:** `docs/adr/ADR-0011-multi-environment-model.md`
- **Related design document:** `docs/multi-environment-model.md`

---

## 1. Purpose

This document defines the initial technical model for configuring multiple target environments in the DevOps Control Plane.

The selected architecture, formalized in `ADR-0011`, is:

```text
Single DevOps Control Plane instance
Multi-environment target model
Correlated ChangeRequests for promotion
Environment-aware RBAC/AuthZ policy
```

The environment configuration model must allow the Control Plane to resolve, for each ChangeRequest target environment:

- Kubernetes/OpenShift runtime namespace;
- Tekton validation configuration;
- GitOps path;
- Argo CD Application mapping;
- GitLab target branch or target path policy;
- approval policy;
- RBAC/AuthZ groups and action rules;
- evidence collection policy;
- operational safety guardrails.

This document is intentionally design-first. It does not implement the runtime model yet.

---

## 2. Design goals

The environment configuration model must satisfy the following goals.

```text
Explicitly define supported environments.
Reject unknown target environments fail-closed.
Keep the first implementation simple and repository-aligned.
Avoid database migration in the first configuration increment if possible.
Support dev, staging and production as first target environments.
Allow environment-specific Tekton, Argo CD, Kubernetes and GitOps mappings.
Allow environment-specific RBAC/AuthZ policy.
Keep production guardrails explicit.
Remain compatible with the existing ChangeRequest targetEnvironment field.
```

---

## 3. Non-goals for the first increment

The first increment does not aim to implement:

```text
Full environment management UI
Dynamic runtime editing of environment definitions
Database-backed environment catalog
Full production dual-approval workflow
Full GitOps repository restructuring
Full staging/production OpenShift namespace provisioning
Full Argo CD multi-application rollout
Full Tekton namespace-per-environment migration
```

These items can be introduced incrementally after the configuration model is accepted.

---

## 4. Selected configuration approach

### 4.1 Initial approach

The initial implementation should use a ConfigMap-driven environment catalog.

Rationale:

- consistent with current OpenShift deployment model;
- easy to review through Git;
- no immediate PostgreSQL migration required;
- suitable for MVP and production-oriented baseline;
- supports operator-controlled configuration;
- can later evolve to a database table if needed.

### 4.2 Future option

A future phase may introduce a database-backed `environments` table if environment definitions need to be managed dynamically by the Control Plane UI/API.

The current recommendation is:

```text
Phase 13.2 and initial implementation: ConfigMap-driven
Future phase if required: database-backed environment catalog
```

---

## 5. Environment naming convention

The initial supported environments are:

```text
dev
staging
production
```

The environment name must be:

- lowercase;
- stable;
- used as the canonical key in configuration;
- stored in `ChangeRequest.targetEnvironment`;
- used for filtering, authorization and mapping;
- treated as an internal identifier, not necessarily as the display label.

Display labels can be configured separately:

```text
dev        -> Development
staging   -> Staging
production -> Production
```

---

## 6. Recommended ConfigMap model

### 6.1 ConfigMap name

Recommended ConfigMap:

```text
devops-control-plane-environments
```

Recommended key:

```text
environments.yaml
```

This keeps the main application configuration separate from the environment catalog.

### 6.2 Conceptual ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: devops-control-plane-environments
  namespace: devops-control-plane
data:
  environments.yaml: |
    defaultEnvironment: dev
    environments:
      dev:
        displayName: Development
        description: Development environment for integration and fast validation.
        enabled: true
        riskProfile: low-to-medium
        approvalPolicy: relaxed
        kubernetes:
          namespace: devops-ci-demo
        tekton:
          namespace: devops-ci-demo
          pipelineName: validate-gitops
          serviceAccount: pipeline
          workspacePVC: pipeline-workspace
          validationPath: apps/demo-go-color-app/overlays/dev
        argocd:
          applicationName: demo-go-color-app-dev
        gitlab:
          targetBranch: main
          gitOpsPath: apps/demo-go-color-app/overlays/dev
        authz:
          viewerGroups:
            - devops-control-plane-viewers
          operatorGroups:
            - devops-control-plane-operators
          approverGroups:
            - devops-control-plane-approvers
          adminGroups:
            - devops-control-plane-admins
        evidence:
          collectDeploymentEvidence: true
          requireEvidenceBeforeClose: false

      staging:
        displayName: Staging
        description: Controlled test/pre-production validation environment.
        enabled: true
        riskProfile: medium
        approvalPolicy: required
        kubernetes:
          namespace: devops-ci-staging
        tekton:
          namespace: devops-ci-staging
          pipelineName: validate-gitops
          serviceAccount: pipeline
          workspacePVC: pipeline-workspace
          validationPath: apps/demo-go-color-app/overlays/staging
        argocd:
          applicationName: demo-go-color-app-staging
        gitlab:
          targetBranch: main
          gitOpsPath: apps/demo-go-color-app/overlays/staging
        authz:
          viewerGroups:
            - devops-control-plane-viewers
          operatorGroups:
            - devops-control-plane-operators
          approverGroups:
            - devops-control-plane-approvers
          adminGroups:
            - devops-control-plane-admins
        evidence:
          collectDeploymentEvidence: true
          requireEvidenceBeforeClose: true

      production:
        displayName: Production
        description: Production environment. Strict approval and evidence requirements apply.
        enabled: false
        riskProfile: high
        approvalPolicy: strict
        kubernetes:
          namespace: devops-ci-production
        tekton:
          namespace: devops-ci-production
          pipelineName: validate-gitops
          serviceAccount: pipeline
          workspacePVC: pipeline-workspace
          validationPath: apps/demo-go-color-app/overlays/production
        argocd:
          applicationName: demo-go-color-app-production
        gitlab:
          targetBranch: main
          gitOpsPath: apps/demo-go-color-app/overlays/production
        authz:
          viewerGroups:
            - devops-control-plane-viewers
          operatorGroups:
            - devops-control-plane-production-operators
          approverGroups:
            - devops-control-plane-production-approvers
          adminGroups:
            - devops-control-plane-admins
        evidence:
          collectDeploymentEvidence: true
          requireEvidenceBeforeClose: true
```

### 6.3 Why `production.enabled=false` initially

Production should be represented in the model from the beginning, but it should remain disabled until:

- production namespace exists;
- production RBAC is validated;
- production Argo CD Application exists;
- production GitOps overlay exists;
- production approval policy is implemented;
- service owner accepts RPO/RTO and operational guardrails.

---

## 7. Configuration schema

### 7.1 Top-level fields

```text
defaultEnvironment
environments
```

### 7.2 Environment fields

Each environment should include:

```text
displayName
description
enabled
riskProfile
approvalPolicy
kubernetes
tekton
argocd
gitlab
authz
evidence
```

### 7.3 `riskProfile`

Allowed values:

```text
low
low-to-medium
medium
high
critical
```

### 7.4 `approvalPolicy`

Allowed values:

```text
relaxed
required
strict
```

Meaning:

```text
relaxed:
  dev-oriented policy; standard operator actions can be allowed within guardrails

required:
  approval required before controlled execution/promotion

strict:
  production-oriented policy; production-specific approval and evidence required
```

---

## 8. Fail-closed validation rules

The application must reject invalid environment configuration and invalid target environment usage.

### 8.1 Startup validation

At startup, the application should validate:

```text
environments.yaml exists or fallback legacy config is explicitly enabled
at least one environment is enabled
defaultEnvironment exists
defaultEnvironment is enabled
every environment has a unique key
every enabled environment has kubernetes.namespace
every enabled environment has tekton.namespace
every enabled environment has tekton.pipelineName
every enabled environment has gitlab.gitOpsPath
every enabled environment has argocd.applicationName
every enabled environment has approvalPolicy
every enabled environment has authz group mapping
```

If validation fails, the recommended behavior is:

```text
fail startup for invalid explicit multi-environment configuration
```

For the transition phase, a compatibility mode can be allowed only if explicitly configured.

### 8.2 ChangeRequest validation

When creating or acting on a ChangeRequest:

```text
if targetEnvironment is empty:
  default to configured defaultEnvironment only if allowed

if targetEnvironment is unknown:
  reject request

if targetEnvironment is disabled:
  reject request

if actor does not have required environment permission:
  reject request
```

### 8.3 Production guardrail

For `production`, the default should be:

```text
fail closed until production policy is explicitly enabled and validated
```

---

## 9. Compatibility with current configuration

The current runtime uses single-value settings such as:

```text
KUBERNETES_NAMESPACE=devops-ci-demo
TEKTON_NAMESPACE=devops-ci-demo
TEKTON_PIPELINE_NAME=validate-gitops
TEKTON_VALIDATION_PATH=apps/demo-go-color-app
```

The first implementation can support a transition strategy:

```text
If devops-control-plane-environments ConfigMap is absent:
  use legacy single-environment config as dev

If devops-control-plane-environments ConfigMap is present:
  use environment catalog and validate it fail-closed
```

This keeps the current runtime stable while enabling progressive migration.

---

## 10. Runtime resolution model

The application should resolve environment-specific settings from `targetEnvironment`.

Example flow:

```text
ChangeRequest targetEnvironment = staging

Resolve environment config:
  env = environments["staging"]

Tekton validation:
  namespace = env.tekton.namespace
  pipelineName = env.tekton.pipelineName
  validationPath = env.tekton.validationPath

Argo CD check:
  applicationName = env.argocd.applicationName

Kubernetes evidence:
  namespace = env.kubernetes.namespace

GitLab update:
  gitOpsPath = env.gitlab.gitOpsPath
  targetBranch = env.gitlab.targetBranch

AuthZ:
  approvalPolicy = env.approvalPolicy
  authz groups = env.authz
```

---

## 11. Impact on existing components

### 11.1 ChangeService

The ChangeService should receive an environment resolver dependency, or access environment configuration through application services.

Expected responsibilities:

```text
validate targetEnvironment
resolve environment policy for actions
use environment mappings for technical workflows
include environment context in events
```

### 11.2 GitLab adapter usage

GitLab calls should become environment-aware through resolved paths and branch policies.

Potential changes:

```text
update-files uses env.gitlab.gitOpsPath
open-merge-request uses env.gitlab.targetBranch
commit messages include targetEnvironment
MR title includes targetEnvironment
```

### 11.3 Tekton adapter usage

Tekton calls should use:

```text
env.tekton.namespace
env.tekton.pipelineName
env.tekton.serviceAccount
env.tekton.workspacePVC
env.tekton.validationPath
```

### 11.4 Argo CD adapter usage

Argo CD checks should use:

```text
env.argocd.applicationName
```

rather than assuming a direct mapping from `applicationName` to one single Argo CD Application.

### 11.5 Kubernetes evidence collector

Runtime evidence should use:

```text
env.kubernetes.namespace
```

and include `targetEnvironment` in payload summaries.

### 11.6 UI

The UI should add environment context to:

```text
Dashboard environment selector
Recent changes
Change list filters
Change detail
Evidence pages
Audit pages
```

---

## 12. AuthZ policy integration

The existing AuthN/AuthZ model should be extended with environment context.

Current dimensions:

```text
actor
role
action
endpoint
```

Future dimensions:

```text
actor
role
action
endpoint
targetEnvironment
approvalPolicy
```

Example future rule:

```text
operator can validate dev
operator can validate staging
operator cannot approve production
production-approver can approve production
admin can perform emergency actions with audit
```

---

## 13. Evidence and audit requirements

Every environment-aware workflow should record:

```text
targetEnvironment
actor
requestedBy
action
runtimeStatus
externalRef
```

Promotion-related workflows should also record:

```text
promotionGroupID
promotedFromChangeNumber
```

This is important for:

- audit trail;
- production readiness;
- incident review;
- DR and evidence correlation;
- final technical documentation.

---

## 14. OpenShift object impact

The first configuration phase may introduce:

```text
ConfigMap devops-control-plane-environments
optional mount or env reference for environments.yaml
application parser/loader support
tests for valid and invalid environment config
```

It should not immediately require:

```text
new production namespace
new production Argo CD Application
new production Tekton pipeline
production RBAC grants
```

Those can be added in later implementation phases.

---

## 15. Repository impact

Expected future files:

```text
manifests/devops-control-plane-environments-template.yaml
docs/environment-configuration-model.md
internal/config/environments.go
internal/config/environments_test.go
```

Potential later files:

```text
docs/runbooks/environment-configuration.md
docs/runbooks/environment-promotion.md
docs/adr/ADR-0012-environment-configuration-source.md
```

---

## 16. Testing strategy

### 16.1 Unit tests

Test cases:

```text
valid dev-only configuration
valid dev/staging/production configuration
missing defaultEnvironment
unknown defaultEnvironment
defaultEnvironment disabled
missing required namespace
missing Tekton pipeline
invalid approvalPolicy
unknown targetEnvironment rejected
disabled targetEnvironment rejected
```

### 16.2 Runtime smoke tests

Extend future smoke tests to validate:

```text
environment config loaded
supported environments listed
unknown environment rejected
existing dev workflow still works
```

### 16.3 UI tests

Validate:

```text
environment selector displays configured environments
dashboard filter works
/ui/changes filter by environment works
Change detail shows target environment clearly
```

---

## 17. Migration plan

Recommended incremental plan:

### Step 1 — Document model

Create this document and align with ADR-0011.

### Step 2 — Introduce ConfigMap template

Add a non-applied repository template for environment configuration.

### Step 3 — Add parser and tests

Add config package support to parse the environment catalog.

### Step 4 — Legacy compatibility

Map current single-environment settings to implicit `dev` if the environment catalog is absent.

### Step 5 — Validate create ChangeRequest

Reject unknown target environments.

### Step 6 — Start resolving Tekton/Kubernetes/Argo CD settings by environment

Start with dev behavior unchanged.

### Step 7 — Add UI environment selector

Initially read-only/filter-only.

### Step 8 — Introduce staging

Add `staging` as enabled only after required namespaces and mappings exist.

### Step 9 — Prepare production disabled

Keep `production` present but disabled until production guardrails are implemented.

---

## 18. Initial acceptance criteria

Phase 13.2 can be considered complete when:

```text
environment configuration model is documented
ConfigMap-driven approach is selected
initial schema is defined
validation rules are defined
legacy compatibility strategy is defined
implementation impact is documented
migration plan is documented
```

No runtime implementation is required to close this design phase.

---

## 19. Decision summary

The initial environment configuration model is:

```text
ConfigMap-driven
explicit environment catalog
canonical environment keys: dev, staging, production
default environment: dev
unknown environments rejected fail-closed
production present but disabled until production guardrails are ready
environment-specific mappings for Kubernetes, Tekton, Argo CD, GitLab, AuthZ and evidence policy
legacy single-environment compatibility retained during transition
```

---

## 20. Next recommended phase

```text
Phase 13.3 — Environment-aware ChangeRequest and promotion metadata design
```

Recommended deliverables:

```text
docs/change-promotion-model.md
promotion metadata design
migration impact assessment
API/UI impact notes
```
