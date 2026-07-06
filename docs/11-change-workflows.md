# DevOps Control Plane — Change Workflows

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 11 — Change Workflows
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
  - `docs/06-argocd-integration.md`
  - `docs/07-gitlab-integration.md`
  - `docs/08-tekton-integration.md`
  - `docs/09-security-rbac.md`
  - `docs/10-data-model.md`
- **Status:** Rewritten in English and refreshed while preserving the original GitOps workflow intent
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document describes the operational workflows of the DevOps Control Plane.

The original document was designed both as an implementation guide and as an educational guide for operators and onboarding readers. This refreshed version preserves that spirit while aligning the workflow model with the current advanced MVP baseline.

The document defines:

- the ChangeRequest lifecycle;
- runtime workflow states;
- allowed transitions;
- current implemented workflows;
- originally planned MVP workflows and their evolution;
- error handling rules;
- evidence collection expectations;
- integration points with GitLab, Tekton, Argo CD and OpenShift;
- security and governance checkpoints;
- completion and failure criteria;
- future multi-environment promotion direction.

The key objective remains unchanged:

```text
Make GitOps changes guided, repeatable, validated, auditable and understandable.
```

---

## 2. Fundamental principle

The DevOps Control Plane must not modify runtime resources directly as the permanent desired state.

Every application change must pass through Git.

The original workflow spirit remains:

```text
Change request
  -> GitLab branch / commit / merge request
  -> Tekton validation
  -> Argo CD deployment state
  -> OpenShift runtime validation
  -> Evidence
  -> Change history
```

Commands such as the following are not the permanent product workflow:

```bash
oc edit deployment
oc patch deployment
oc set env deployment
oc scale deployment
```

These commands may be used only for manual troubleshooting or explicitly documented emergency operations.

---

## 3. Current workflow model

The current implementation separates two concepts that were initially mixed in the first MVP state model:

```text
lifecycle status = business/process lifecycle
runtime status   = technical execution state
```

This separation is important because a technical step can complete without changing the business lifecycle state.

Example:

```text
lifecycle status = draft
runtime status   = EvidenceCollected
```

This means that technical evidence has been collected, but the business lifecycle has not necessarily been closed.

---

## 4. ChangeRequest lifecycle

## 4.1 Lifecycle states

Current lifecycle states:

```text
draft
submitted
approved
executing
executed
closed
```

These states describe the process view of a ChangeRequest.

## 4.2 Lifecycle transitions

Typical lifecycle path:

```text
draft
  -> submitted
  -> approved
  -> executing
  -> executed
  -> closed
```

Each transition must create a `change_events` record.

## 4.3 Lifecycle endpoint model

Current lifecycle endpoints include:

```text
POST /api/v1/changes/{id}/submit
POST /api/v1/changes/{id}/approve
POST /api/v1/changes/{id}/start-execution
POST /api/v1/changes/{id}/complete-execution
POST /api/v1/changes/{id}/close
```

Lifecycle transitions are subject to backend AuthZ.

---

## 5. Runtime status model

Runtime status describes technical execution progress.

Current and expected runtime statuses include:

```text
BranchCreated
CommitCreated
MergeRequestOpened
MergeRequestMerged
ValidationRunning
ValidationSucceeded
ValidationFailed
DeploymentSyncedHealthy
DeploymentProgressing
DeploymentOutOfSync
DeploymentDegraded
DeploymentUnknown
EvidenceCollected
```

Runtime status is updated by technical workflow actions such as:

- GitLab branch creation;
- GitOps file update;
- merge request opening;
- merge request merge;
- Tekton validation;
- Argo CD deployment check;
- runtime evidence collection.

Runtime status changes must create auditable events.

---

## 6. Workflow types

## 6.1 Currently implemented technical workflows

Current implemented technical workflows include:

```text
create ChangeRequest
create GitLab branch
update GitOps files
open GitLab merge request
merge GitLab merge request
run Tekton validation
check Tekton validation
check Argo CD deployment
collect deployment evidence
view ChangeRequest history and evidence
```

## 6.2 Original MVP workflow types

The original MVP workflow types remain useful as conceptual workflow families:

```text
WF-001 update-replicas
WF-002 update-app-version
WF-003 update-page-color
WF-004 update-configmap-values
WF-005 validate-only
WF-006 sync-only
WF-007 collect-evidence-only
```

Current implementation does not need every business workflow to be fully automated before the technical building blocks are stable.

The current architecture intentionally exposes explicit technical steps first, because explicit steps are easier to validate, debug and explain.

---

## 7. GitLab technical workflow

## 7.1 Purpose

The GitLab workflow represents the GitOps change as a branch, file update, commit and merge request.

This preserves the core GitOps rule:

```text
Git is the source of desired state.
```

## 7.2 Flow

```text
ChangeRequest
  -> create branch
  -> update GitOps files
  -> open merge request
  -> merge request where authorized
  -> record GitLab references
  -> update runtime status
  -> create audit events
```

## 7.3 Current endpoints

```text
POST /api/v1/changes/{id}/create-branch
POST /api/v1/changes/{id}/update-files
POST /api/v1/changes/{id}/open-merge-request
POST /api/v1/changes/{id}/merge-request
```

## 7.4 Expected events and runtime statuses

Typical runtime statuses:

```text
BranchCreated
CommitCreated
MergeRequestOpened
MergeRequestMerged
```

Typical event payload should include safe references such as:

- branch name;
- GitLab project ID;
- commit SHA where available;
- merge request IID;
- merge request URL;
- target branch;
- target environment.

## 7.5 Security rules

- Do not log GitLab token values.
- Do not store GitLab token values in PostgreSQL.
- Do not include token values in evidence.
- Git changes must be reviewed through merge request where policy requires it.
- Direct commit remains a lab/bootstrap exception, not the preferred governance model.

---

## 8. Tekton validation workflow

## 8.1 Purpose

Tekton validation verifies that the GitOps change is technically acceptable before the workflow proceeds.

Tekton answers the question:

```text
Is this GitOps change technically safe enough to continue?
```

## 8.2 Flow

```text
ChangeRequest
  -> POST /api/v1/changes/{id}/validate
  -> create Tekton PipelineRun
  -> runtime status becomes ValidationRunning
  -> POST /api/v1/changes/{id}/check-validation
  -> read PipelineRun status
  -> collect TaskRun diagnostics
  -> persist validation evidence
  -> runtime status becomes ValidationSucceeded or ValidationFailed
```

## 8.3 Current endpoints

```text
POST /api/v1/changes/{id}/validate
POST /api/v1/changes/{id}/check-validation
GET  /api/v1/changes/{id}/evidence/validation
```

## 8.4 Validation outcomes

```text
Tekton Succeeded=True    -> ValidationSucceeded
Tekton Succeeded=False   -> ValidationFailed
Tekton Succeeded=Unknown -> ValidationRunning
```

## 8.5 Failure handling

If validation fails:

- runtime status becomes `ValidationFailed`;
- `check-validation` returns diagnostics;
- failed TaskRuns are identified;
- validation evidence is stored;
- the workflow must not automatically continue to deployment state as if validation succeeded.

---

## 9. Argo CD deployment check workflow

## 9.1 Purpose

The Control Plane checks Argo CD deployment state after Git workflow and validation steps.

The current baseline focuses on deployment state checks rather than a broad automatic sync orchestration model.

## 9.2 Flow

```text
ChangeRequest
  -> POST /api/v1/changes/{id}/check-deployment
  -> read Argo CD Application
  -> map sync and health status
  -> update runtime status
  -> create technical event
```

## 9.3 Runtime status mapping

```text
Synced + Healthy       -> DeploymentSyncedHealthy
OutOfSync              -> DeploymentOutOfSync
Progressing            -> DeploymentProgressing
Degraded               -> DeploymentDegraded
Unknown/other          -> DeploymentUnknown
```

## 9.4 Warning handling

Warnings such as `OrphanedResourceWarning` must be preserved and shown as warnings.

A `Synced` and `Healthy` application with an orphaned resource warning should not automatically be marked as failed.

---

## 10. Runtime evidence workflow

## 10.1 Purpose

Runtime evidence verifies the actual OpenShift state after the GitOps workflow and deployment state checks.

## 10.2 Flow

```text
ChangeRequest
  -> POST /api/v1/changes/{id}/collect-evidence
  -> collect Argo CD state
  -> collect Kubernetes/OpenShift runtime state
  -> build diagnostics summary
  -> persist deployment evidence
  -> runtime status becomes EvidenceCollected
```

## 10.3 Current endpoint

```text
POST /api/v1/changes/{id}/collect-evidence
GET  /api/v1/changes/{id}/evidence/deployment
```

## 10.4 Evidence scope

Deployment evidence may include:

- ChangeRequest metadata;
- application name;
- target environment;
- Argo CD sync and health status;
- Argo CD revision and warnings;
- Deployment readiness;
- Pod readiness;
- restart counts;
- Service availability;
- Route availability;
- diagnostics summary.

Evidence must be sanitized.

---

## 11. End-to-end technical workflow baseline

A current end-to-end technical workflow can be represented as:

```text
Create ChangeRequest
  -> create GitLab branch
  -> update GitOps files
  -> open merge request
  -> run Tekton validation
  -> check Tekton validation
  -> merge merge request where authorized
  -> check Argo CD deployment
  -> collect deployment evidence
  -> review events and evidence in UI
```

This flow is intentionally explicit. It keeps each step observable and easier to validate.

---

## 12. WF-001 — update-replicas

## 12.1 Purpose

Modify the number of replicas through GitOps.

The workflow must update `spec.replicas` in Git, not scale the runtime Deployment directly.

## 12.2 Input

```yaml
applicationName: demo-go-color-app
targetEnvironment: dev
changeType: update-replicas
requestedBy: vmarzario
description: Scale demo-go-color-app from 2 to 3 replicas
payload:
  deploymentName: demo-go-color-app
  replicas: 3
```

## 12.3 Detailed flow

```text
1. Create ChangeRequest.
2. Resolve Application metadata.
3. Resolve GitLab repository and GitOps path.
4. Check active conflicting changes.
5. Create GitLab branch.
6. Read deployment manifest.
7. Update spec.replicas.
8. Generate reviewable diff.
9. Run anti-secret guardrails.
10. Commit files to branch.
11. Run Tekton validation.
12. Open merge request.
13. Merge where authorized.
14. Check Argo CD deployment state.
15. Collect runtime evidence.
16. Review final history and evidence.
```

## 12.4 Success criteria

The workflow is successful when:

- GitOps file is updated;
- commit or merge request represents the change;
- Tekton validation succeeds;
- merge or target branch update is approved according to policy;
- Argo CD reports an acceptable deployment state;
- runtime evidence is collected;
- evidence is persisted.

## 12.5 Failure examples

| Failure | Runtime status | Error family |
|---|---|---|
| Application not found | failure event | ARGOCD_* |
| GitOps file not found | failure event | GITLAB_* |
| Branch already exists | failure event | GITLAB_* |
| YAML invalid | ValidationFailed | VALIDATION_* |
| Tekton fails | ValidationFailed | TEKTON_* |
| Deployment degraded | DeploymentDegraded | ARGOCD_* |
| Runtime not ready | deployment/evidence failure | KUBERNETES_* |

---

## 13. WF-002 — update-app-version

## 13.1 Purpose

Update the application value `APP_VERSION` through GitOps.

The value can be defined:

1. inline in the Deployment;
2. in a ConfigMap referenced by the Deployment.

Preferred model:

```text
ConfigMap-managed APP_VERSION
```

## 13.2 Important rollout note

If `APP_VERSION` is consumed as an environment variable from a ConfigMap, existing Pods do not automatically receive the new value until Pods are recreated.

Future strategies:

- ConfigMap checksum annotation in the Pod template;
- ChangeRequest annotation in the Deployment;
- controlled rollout restart;
- separate Deployment change.

This decision should be formalized before enabling production-grade workflows.

---

## 14. WF-003 — update-page-color

## 14.1 Purpose

Update `PAGE_COLOR` through GitOps.

## 14.2 Input validation

`PAGE_COLOR` must match:

```text
#[0-9A-Fa-f]{6}
```

Valid examples:

```text
#1E90FF
#28A745
#FF0000
```

Invalid examples:

```text
blue
28A745
#XYZ123
#123
```

## 14.3 Evidence

If the application exposes a safe Route endpoint, the system may collect:

- HTTP status code;
- health endpoint status;
- safe page output if non-sensitive;
- observed value if safely exposed.

---

## 15. WF-004 — update-configmap-values

## 15.1 Purpose

Modify one or more keys in a ConfigMap managed through GitOps.

## 15.2 Validation rules

- ConfigMap must exist in the repository.
- ConfigMap must be included in `kustomization.yaml` where applicable.
- AppProject must allow `ConfigMap` where required.
- Values must not contain secrets.
- `PAGE_COLOR`, if present, must match the expected hex format.
- Key creation policy must be explicit.

## 15.3 Specific failure examples

| Failure | Runtime status | Error family |
|---|---|---|
| ConfigMap file missing | failure event | GITLAB_* |
| ConfigMap not included in Kustomize | ValidationFailed | VALIDATION_* |
| ConfigMap not allowed by AppProject | DeploymentOutOfSync or failure event | ARGOCD_* |
| Possible secret in values | ValidationFailed | VALIDATION_SECRET_DETECTED |
| Pods not refreshed | warning or failure based on policy | KUBERNETES_* |

---

## 16. WF-005 — validate-only

## 16.1 Purpose

Run Tekton validation on a branch or revision without proceeding to deployment state checks.

Useful for:

- preventive validation;
- merge request review;
- workflow testing;
- manifest troubleshooting.

## 16.2 Flow

```text
1. Create ChangeRequest or validation-oriented record.
2. Resolve Git metadata.
3. Create Tekton PipelineRun.
4. Check validation result.
5. Collect Tekton evidence.
6. End with ValidationSucceeded or ValidationFailed.
```

---

## 17. WF-006 — deployment-check-only

## 17.1 Purpose

Check Argo CD deployment state without creating a new Git change.

This replaces the older generic `sync-only` idea as the safer current baseline.

Useful when:

- Git is already updated;
- a merge request was merged manually;
- the operator needs to verify Argo CD status;
- a runtime state check is needed before evidence collection.

## 17.2 Flow

```text
1. Create or select ChangeRequest.
2. Read Argo CD Application.
3. Map sync and health status.
4. Update runtime status.
5. Record event.
```

---

## 18. WF-007 — collect-evidence-only

## 18.1 Purpose

Collect evidence for an Application or ChangeRequest without modifying Git and without triggering deployment actions.

Useful for:

- audit;
- troubleshooting;
- post-change verification;
- runtime baseline capture.

## 18.2 Flow

```text
1. Select ChangeRequest.
2. Read Argo CD status.
3. Read Kubernetes/OpenShift runtime resources.
4. Build diagnostics.
5. Save sanitized evidence.
6. Runtime status becomes EvidenceCollected where applicable.
```

---

## 19. Conflict handling

## 19.1 Active conflicting changes

The system should avoid unsafe concurrent workflows on the same application, target environment or GitOps path.

Initial rule:

```text
Block or warn when an active ChangeRequest targets the same application, environment and file path.
```

## 19.2 Example

Change A:

```text
update-configmap-values on apps/demo-go-color-app/configmap.yaml in dev
```

Change B:

```text
update-page-color on apps/demo-go-color-app/configmap.yaml in dev
```

Recommended action:

```text
Block Change B until Change A is completed, cancelled or explicitly superseded.
```

Recommended error family:

```text
WORKFLOW_CONFLICT_ACTIVE_CHANGE
```

---

## 20. Rollback handling

## 20.1 Principle

Definitive rollback should preferably pass through Git.

```text
git revert or revert merge request
  -> merge
  -> Argo CD deployment state
  -> OpenShift runtime evidence
```

## 20.2 Operational Argo CD rollback

An operational Argo CD rollback can move the cluster to a previous revision, but it does not modify Git.

The system must explain that this can result in:

```text
OutOfSync from main
Healthy
```

Meaning:

```text
runtime is healthy but not aligned with the current Git target branch
```

## 20.3 Future rollback workflow

Future workflow:

```text
rollback-git-revert
  -> create revert branch
  -> open revert merge request
  -> validate with Tekton
  -> merge
  -> check Argo CD deployment
  -> collect rollback evidence
```

---

## 21. Evidence collection standard

## 21.1 Evidence for a completed technical workflow

A completed technical workflow should have:

- ChangeRequest summary;
- GitLab branch, commit or merge request;
- validation PipelineRun;
- TaskRun diagnostics;
- Argo CD sync and health state;
- Kubernetes/OpenShift runtime status;
- warnings;
- timestamps;
- sanitized payloads.

## 21.2 Evidence for failed workflows

A failed workflow should preserve:

- failed phase;
- error code;
- technical message where safe;
- operational message;
- suggested action;
- last observed state;
- sanitized logs or diagnostics;
- partial evidence when useful.

---

## 22. Error handling standard

## 22.1 Normalized workflow error

Example:

```json
{
  "code": "ARGOCD_OPERATION_IN_PROGRESS",
  "technicalMessage": "another operation is already in progress",
  "message": "Argo CD already has an operation in progress for the Application.",
  "suggestedAction": "Wait for the operation to complete or inspect it manually.",
  "recoverable": true
}
```

## 22.2 Rules

- every workflow error must have an error code;
- every ChangeRequest-related error must create an event where applicable;
- tokens must not be logged;
- secrets must not be stored;
- previous status must not be lost;
- workflows must not remain indefinitely stuck.

---

## 23. Timeout behavior

## 23.1 Recommended timeouts

```text
GitLab API call: 30s
Argo CD API call: 30s
Tekton PipelineRun: 900s
Deployment check: 600s
Runtime readiness check: 300s
HTTP health check: 10s
```

## 23.2 Timeout rules

When timeout occurs:

- update runtime status or record an error event;
- save last observed state;
- save partial evidence where useful;
- propose manual action;
- avoid automatic unsafe continuation.

---

## 24. UI and API orchestration mapping

## 24.1 Create change

```text
POST /api/v1/changes
```

Creates a ChangeRequest.

## 24.2 Lifecycle actions

```text
POST /api/v1/changes/{id}/submit
POST /api/v1/changes/{id}/approve
POST /api/v1/changes/{id}/start-execution
POST /api/v1/changes/{id}/complete-execution
POST /api/v1/changes/{id}/close
```

## 24.3 Technical actions

```text
POST /api/v1/changes/{id}/create-branch
POST /api/v1/changes/{id}/update-files
POST /api/v1/changes/{id}/open-merge-request
POST /api/v1/changes/{id}/merge-request
POST /api/v1/changes/{id}/validate
POST /api/v1/changes/{id}/check-validation
POST /api/v1/changes/{id}/check-deployment
POST /api/v1/changes/{id}/collect-evidence
```

## 24.4 Full automation endpoint

A future endpoint may orchestrate multiple technical steps:

```text
POST /api/v1/changes/{id}/run
```

The current recommended model remains explicit step-by-step execution, because step-by-step execution is easier to validate, explain and audit.

---

## 25. Educational workflow view

Every workflow should show:

```text
current step
tool being used
why the step exists
files or resources involved
evidence produced
final or current status
```

Example:

```text
Step: Tekton validation
Tool: OpenShift Pipelines / Tekton
Purpose: validate GitOps manifests before deployment checks
Files: apps/demo-go-color-app/deployment.yaml
Evidence: PipelineRun devops-cp-validate-chg-2026-0006 succeeded
```

This preserves the original goal of making the system useful for both experienced operators and less experienced readers.

---

## 26. Multi-developer behavior

When multiple developers create ChangeRequests:

- `requestedBy` must be visible;
- dashboard recent changes shows the latest five changes;
- `/ui/changes` shows the full list;
- each ChangeRequest keeps its own event and evidence history;
- audit remains per ChangeRequest.

This avoids hiding older changes while keeping the dashboard readable.

---

## 27. Multi-environment promotion direction

Future workflows will support environment-aware promotion.

Canonical environment flow:

```text
dev -> staging -> production
```

Recommended future promotion workflow:

```text
Change in dev
  -> validation and evidence
  -> related ChangeRequest for staging
  -> staging validation and evidence
  -> related ChangeRequest for production
  -> production approval
  -> production validation and evidence
```

Future metadata:

```text
promotionGroupID
promotedFromChangeNumber
```

Production workflows must remain disabled until production guardrails, approval policy and environment-aware AuthZ are implemented.

---

## 28. Workflow readiness checklist

A workflow is ready when:

- it creates or uses a ChangeRequest;
- it creates events for each significant step;
- it modifies Git, not runtime desired state;
- it records GitLab branch, commit or merge request references;
- it runs Tekton validation where required;
- it blocks deployment progression on validation failure;
- it checks Argo CD deployment state;
- it collects sanitized evidence;
- it saves runtime status;
- it handles timeout;
- it never stores secrets;
- it produces understandable messages;
- it is authorized by backend AuthZ.

---

## 29. Relationship with other documents

This document informs and is informed by:

- `docs/13-api-design.md`;
- `docs/12-evidence-model.md`;
- `docs/10-data-model.md`;
- `docs/07-gitlab-integration.md`;
- `docs/08-tekton-integration.md`;
- `docs/06-argocd-integration.md`;
- `docs/09-security-rbac.md`;
- `docs/environment-configuration-model.md`;
- `docs/change-promotion-model.md`;
- `internal/app/change_service.go`;
- `internal/adapters/gitlab/`;
- `internal/adapters/argocd/`;
- `internal/adapters/tekton/`.

---

## 30. Key message

Workflows are the heart of the DevOps Control Plane.

The product must help the operator do the right thing:

```text
modify Git
validate with Tekton
check deployment with Argo CD
verify OpenShift runtime
save evidence
preserve history
```

The Control Plane must not hide GitOps. The Control Plane must make GitOps clearer, more guided, more repeatable and more auditable.

This preserves the original spirit of the project while aligning the workflow model with the current advanced MVP baseline.

---

## 31. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial change workflows document in Italian. |
| 2026-07-06 | 0.2 | Rewritten in English and refreshed while preserving the original GitOps-first workflow and educational intent and aligning it with the current lifecycle/runtime status, GitLab, Tekton, Argo CD, evidence and multi-environment baseline. |
