# DevOps Control Plane — Functional Requirements

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 03 — Functional Requirements
- **Version:** 0.2
- **Date:** 2026-07-03
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Previous documents:**
  - `docs/00-vision.md`
  - `docs/01-scope-mvp.md`
  - `docs/02-personas-use-cases.md`
- **Status:** Rewritten in English and refreshed to align with the current advanced MVP, security, operability and multi-environment baseline
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document defines the functional requirements of the DevOps Control Plane.

The original version described the first MVP requirements. The current version updates those requirements to reflect the implemented advanced MVP baseline and the current roadmap direction.

The document defines:

- application discovery requirements;
- ChangeRequest requirements;
- GitLab workflow requirements;
- Tekton validation requirements;
- Argo CD deployment check requirements;
- Kubernetes/OpenShift evidence requirements;
- Web UI requirements;
- AuthN/AuthZ requirements;
- audit and evidence requirements;
- operability requirements;
- multi-developer requirements;
- multi-environment requirements;
- error handling requirements;
- acceptance criteria.

This document is an input for architecture, API design, data model, runbooks and final technical documentation.

---

## 2. Requirement conventions

### 2.1 Priority levels

| Priority | Meaning |
|---|---|
| MUST | Required for the current advanced MVP or required for safe operation |
| SHOULD | Important and expected, but can be implemented incrementally |
| COULD | Useful for a later roadmap phase |
| WON'T | Explicitly outside the current baseline |

### 2.2 Requirement structure

Each requirement should include:

- identifier;
- priority;
- description;
- expected input or trigger;
- expected output or state;
- rules;
- acceptance criteria;
- current implementation status where relevant.

### 2.3 Fundamental GitOps rule

Every permanent application workload change must be represented as a Git change.

The system must not use imperative runtime mutations as the final desired-state workflow.

Commands such as the following may be useful for troubleshooting, but they must not represent the final DevOps Control Plane workflow:

```bash
oc edit deployment
oc patch deployment
oc set env deployment
oc scale deployment
```

---

## 3. Functional areas

The functional requirements are grouped into the following areas:

1. Application discovery and detail;
2. ChangeRequest management;
3. lifecycle and runtime status;
4. GitLab operations;
5. Tekton validation;
6. Argo CD deployment checks;
7. Kubernetes/OpenShift runtime evidence;
8. evidence and audit;
9. Web UI;
10. AuthN/AuthZ and security;
11. operability;
12. multi-developer behavior;
13. multi-environment direction;
14. error handling.

---

# 4. Application discovery and detail

## FR-APP-001 — List Argo CD applications

### Priority

MUST

### Description

The system must retrieve from Argo CD API the list of Applications visible to the DevOps Control Plane.

### Output

For each Application, the system should expose at least:

```yaml
name: demo-go-color-app
argocdNamespace: openshift-gitops
project: default
targetNamespace: devops-ci-demo
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
targetRevision: main
path: apps/demo-go-color-app
syncStatus: Synced
healthStatus: Healthy
currentRevision: <revision>
```

### Rules

- The system must normalize the main fields.
- The system must handle unreachable Argo CD API errors.
- The system must not expose Argo CD tokens in logs or responses.
- The system must preserve warnings such as `OrphanedResourceWarning` without automatically treating a healthy application as failed.

### Acceptance criteria

- `GET /api/v1/applications` returns the application list.
- Each record contains at least name, project, namespace, sync, health and revision.
- Argo CD API errors are returned in a readable form.

### Current status

Implemented.

---

## FR-APP-002 — Show application detail

### Priority

MUST

### Description

The system must show operational details for a single Argo CD Application.

### Input

```text
applicationName
```

### Output

- name;
- project;
- target namespace;
- repository;
- path;
- target revision;
- current revision;
- sync status;
- health status;
- conditions;
- destination server;
- resource summary.

### Acceptance criteria

- `GET /api/v1/applications/{name}` returns detail consistent with Argo CD.
- The UI shows application detail to authorized users.
- Known warnings are visible but not overclassified as failures.

### Current status

Implemented.

---

## FR-APP-003 — Show managed resources and runtime views

### Priority

SHOULD

### Description

The system should expose managed resources, runtime information and Argo CD history where available.

### Expected resources

- Deployments;
- Pods;
- Services;
- Routes;
- ConfigMaps where relevant;
- Argo CD resource conditions.

### Acceptance criteria

- `GET /api/v1/applications/{name}/resources` returns resources where available.
- `GET /api/v1/applications/{name}/history` returns Argo CD history where available.
- `GET /api/v1/applications/{name}/runtime` returns runtime information where available.

### Current status

Partially implemented and available through the current API/UI baseline.

---

# 5. ChangeRequest management

## FR-CHG-001 — Create ChangeRequest

### Priority

MUST

### Description

The system must create a ChangeRequest for each tracked GitOps change.

### Minimum input

```yaml
title: Update demo application
applicationName: demo-go-color-app
targetEnvironment: dev
changeType: standard
riskLevel: medium
requestedBy: developer-a
description: Multi-developer dashboard validation test
```

### Rules

- `title` is required.
- `applicationName` is required.
- `changeType` is required.
- `requestedBy` is required.
- `targetEnvironment` defaults to `dev` until environment validation is enabled.
- `riskLevel` defaults to `medium` when omitted.
- Future environment validation must reject unknown or disabled environments fail-closed.

### Acceptance criteria

- The ChangeRequest is stored in PostgreSQL.
- A unique `CHG-YYYY-NNNN` change number is generated.
- A creation event is recorded.
- The ChangeRequest appears in API and UI.

### Current status

Implemented.

---

## FR-CHG-002 — List and get ChangeRequests

### Priority

MUST

### Description

The system must allow authorized users to list ChangeRequests and view ChangeRequest details.

### Endpoints

```text
GET /api/v1/changes
GET /api/v1/changes/{id}
```

### Acceptance criteria

- ChangeRequests are returned with application, requester, target environment, lifecycle status and runtime status.
- The UI `/ui/changes` shows all ChangeRequests.
- The UI shows the requester through the `REQUESTED BY` column.
- Change detail includes events and evidence links.

### Current status

Implemented and validated in a multi-developer scenario.

---

## FR-CHG-003 — Manage lifecycle transitions

### Priority

MUST

### Description

The system must support functional lifecycle transitions independently from technical runtime status.

### Lifecycle states

```text
draft
submitted
approved
executing
executed
closed
```

### Lifecycle endpoints

```text
POST /api/v1/changes/{id}/submit
POST /api/v1/changes/{id}/approve
POST /api/v1/changes/{id}/start-execution
POST /api/v1/changes/{id}/complete-execution
POST /api/v1/changes/{id}/close
```

### Acceptance criteria

- Valid transitions succeed.
- Invalid transitions fail with readable errors.
- Every lifecycle transition creates a `change_events` record.
- Lifecycle status does not overwrite technical runtime status.

### Current status

Implemented and tested.

---

## FR-CHG-004 — Track runtime status

### Priority

MUST

### Description

The system must track technical runtime state separately from lifecycle state.

### Example runtime statuses

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
EvidenceCollected
```

### Acceptance criteria

- Technical actions update runtime status.
- Runtime status changes create events.
- UI and API show runtime status clearly.

### Current status

Implemented.

---

# 6. GitLab operations

## FR-GIT-001 — Configure GitLab target repository

### Priority

MUST

### Description

The system must know or resolve the GitLab metadata needed for target GitOps repositories.

### Required configuration

```yaml
provider: gitlab
baseUrl: https://gitlab.example.local
projectId: "12345"
defaultBranch: main
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
path: apps/demo-go-color-app
```

### Acceptance criteria

- Each managed Application can be associated with a GitLab repository and path.
- Tokens are read from Secrets and never printed.
- GitLab TLS strict mode is supported.

### Current status

Implemented for the lab GitLab target.

---

## FR-GIT-002 — Create branch

### Priority

MUST

### Description

The system must create a dedicated branch for a ChangeRequest.

### Naming convention

```text
change/<change-number>
```

Example:

```text
change/CHG-2026-0006
```

### Acceptance criteria

- The branch is created through GitLab API.
- The branch starts from the configured target branch.
- The branch reference is recorded in a technical event.

### Current status

Implemented.

---

## FR-GIT-003 — Update GitOps files

### Priority

MUST

### Description

The system must update GitOps files in a controlled way through GitLab API.

### Rules

- Commits must include ChangeRequest context.
- Generated content must not contain tokens or secrets.
- The target path will become environment-aware in a future phase.

### Acceptance criteria

- The file update creates a GitLab commit.
- The commit reference is recorded.
- The runtime cluster is not directly modified as desired state.

### Current status

Implemented.

---

## FR-GIT-004 — Open merge request

### Priority

MUST

### Description

The system must open a GitLab merge request for a ChangeRequest where the workflow requires review.

### Acceptance criteria

- Merge request IID and URL are recorded.
- Source and target branches are visible.
- Runtime status is updated.

### Current status

Implemented.

---

## FR-GIT-005 — Merge merge request

### Priority

SHOULD

### Description

The system should support merging a GitLab merge request when the actor is authorized and the workflow allows it.

### Acceptance criteria

- Merge is executed through GitLab API.
- Merge commit SHA is recorded.
- Runtime status is updated.
- Authorization checks are applied.

### Current status

Implemented for the current lab workflow.

---

# 7. Tekton validation

## FR-TKN-001 — Create validation PipelineRun

### Priority

MUST

### Description

The system must create a Tekton `PipelineRun` through Kubernetes API.

### Input

- Git repository URL;
- branch or revision;
- validation path;
- ChangeRequest number;
- pipeline name;
- target namespace.

### Acceptance criteria

- The `PipelineRun` is created in the configured namespace.
- The `PipelineRun` name, namespace and UID are recorded.
- Runtime status becomes `ValidationRunning` where appropriate.

### Current status

Implemented.

---

## FR-TKN-002 — Check validation result

### Priority

MUST

### Description

The system must check the latest Tekton validation result for a ChangeRequest.

### Acceptance criteria

- Successful validation maps to `ValidationSucceeded`.
- Failed validation maps to `ValidationFailed`.
- Tekton status, reason and message are returned.
- The system handles missing or unknown PipelineRuns.

### Current status

Implemented.

---

## FR-TKN-003 — Collect TaskRun diagnostics

### Priority

MUST

### Description

The system must collect TaskRun diagnostics for failed or completed validation runs.

### Acceptance criteria

- TaskRuns associated with the PipelineRun are listed.
- Failed tasks are identified.
- Diagnostics are included in the API response and validation evidence.

### Current status

Implemented.

---

## FR-TKN-004 — Enforce GitOps policy guardrails

### Priority

MUST

### Description

Tekton validation must enforce GitOps guardrails including anti-secret and unsupported manifest checks.

### Acceptance criteria

- Kubernetes Secret manifests are blocked by policy.
- Inline secret-like values are detected.
- Safe config references can be allowed.
- Policy failure messages are actionable.

### Current status

Implemented in the validation pipeline baseline.

---

# 8. Argo CD deployment checks

## FR-ARGO-001 — Check deployment state

### Priority

MUST

### Description

The system must check Argo CD deployment state for a ChangeRequest application.

### Runtime status mapping

```text
Synced + Healthy       -> DeploymentSyncedHealthy
OutOfSync              -> DeploymentOutOfSync
Progressing            -> DeploymentProgressing
Degraded               -> DeploymentDegraded
Unknown/other          -> DeploymentUnknown
```

### Acceptance criteria

- The application status is read through Argo CD API.
- Runtime status is updated.
- Argo CD errors are readable.
- TLS strict mode is supported.

### Current status

Implemented.

---

## FR-ARGO-002 — Interpret Argo CD warnings

### Priority

SHOULD

### Description

The system should preserve Argo CD warnings and conditions without overclassifying them.

### Acceptance criteria

- `OrphanedResourceWarning` is captured.
- A healthy application with warnings can still be reported as healthy.
- Warnings are included in evidence diagnostics.

### Current status

Implemented in deployment diagnostics.

---

# 9. Kubernetes/OpenShift runtime evidence

## FR-K8S-001 — Collect deployment evidence

### Priority

MUST

### Description

The system must collect runtime evidence from Kubernetes/OpenShift.

### Target resources

- Deployment;
- Pods;
- Service;
- Route;
- runtime diagnostics.

### Acceptance criteria

- Evidence is associated with the ChangeRequest.
- Evidence is sanitized.
- Evidence includes runtime readiness information.
- Static Kubernetes token is not required.

### Current status

Implemented with ServiceAccount token fallback.

---

## FR-K8S-002 — Summarize deployment diagnostics

### Priority

SHOULD

### Description

The system should summarize runtime evidence for UI and audit readability.

### Diagnostics data

- Argo CD synced;
- Argo CD healthy;
- deployment ready;
- generation observed;
- ready replicas;
- pods ready;
- total restarts;
- Service available;
- Route available;
- warnings.

### Acceptance criteria

- Diagnostics are stored in the evidence payload.
- UI renders diagnostics in readable cards.

### Current status

Implemented.

---

# 10. Evidence and audit

## FR-EVD-001 — Store evidence records

### Priority

MUST

### Description

The system must store evidence records in PostgreSQL.

### Evidence types

```text
validation
deployment
```

### Acceptance criteria

- Evidence is linked to a ChangeRequest.
- Evidence has a name, type, summary, external reference, payload and timestamp.
- Evidence can be retrieved through API and UI.
- Evidence is sanitized.

### Current status

Implemented.

---

## FR-EVD-002 — Store audit events

### Priority

MUST

### Description

The system must store lifecycle and technical events for ChangeRequests.

### Acceptance criteria

- Each significant action creates an event.
- Events include event type, actor where available, payload and timestamp.
- Events are visible through API and UI.

### Current status

Implemented.

---

# 11. Web UI

## FR-UI-001 — Provide dashboard

### Priority

MUST

### Description

The system must provide a dashboard for operational visibility.

### Requirements

- show KPI counters;
- show application summary;
- show recent changes;
- limit recent changes to five records;
- provide `View all` link to the full ChangeRequest list;
- show requester information;
- keep UI labels in English.

### Current status

Implemented and validated.

---

## FR-UI-002 — Provide ChangeRequest pages

### Priority

MUST

### Description

The system must provide UI pages for listing and inspecting ChangeRequests.

### Requirements

- list all ChangeRequests;
- show change number;
- show application;
- show requester;
- show target environment;
- show lifecycle status;
- show runtime status;
- open detail page;
- show evidence and audit links.

### Current status

Implemented and validated in a multi-developer scenario.

---

## FR-UI-003 — Provide server-side technical actions

### Priority

SHOULD

### Description

The UI should expose server-side action buttons for allowed technical actions.

### Acceptance criteria

- Actions post to backend endpoints.
- Flash or error messages are displayed.
- Backend AuthZ remains authoritative.

### Current status

Implemented.

---

# 12. AuthN/AuthZ and security

## FR-AUTH-001 — Protect UI and API with OpenShift OAuth Proxy

### Priority

MUST

### Description

The system must protect UI and API routes through OpenShift OAuth Proxy.

### Acceptance criteria

- Anonymous UI/API requests through Route are blocked.
- Health endpoints remain accessible.
- Authenticated browser access works through OpenShift login.

### Current status

Implemented.

---

## FR-AUTH-002 — Enforce backend authorization

### Priority

MUST

### Description

The backend must enforce authorization based on trusted headers and resolved groups.

### Roles

```text
viewer
operator
approver
admin
```

### Acceptance criteria

- Anonymous backend API access is rejected.
- Unclassified endpoints are denied fail-closed.
- Operators, approvers and admins have differentiated permissions.
- OpenShift group lookup is supported.

### Current status

Implemented.

---

## FR-SEC-001 — Handle secrets safely

### Priority

MUST

### Description

The system must handle tokens and credentials safely.

### Rules

- No token in Git.
- No token in logs.
- No token in evidence.
- Secret inventories must avoid printing values.
- Exposed tokens must be rotated.

### Current status

Implemented through runbooks and operational practices.

---

# 13. Operability

## FR-OPS-001 — Provide health endpoints

### Priority

MUST

### Description

The system must expose health endpoints.

### Endpoints

```text
GET /readyz
GET /livez
```

### Acceptance criteria

- Route health endpoints return HTTP 200 when healthy.
- Backend direct health endpoints return HTTP 200 when healthy.
- Health endpoints bypass auth safely.

### Current status

Implemented.

---

## FR-OPS-002 — Provide operability smoke test

### Priority

MUST

### Description

The repository must provide an automated smoke-test script.

### Acceptance criteria

- Script collects evidence directory.
- Script validates workloads, routes, backend, PostgreSQL, NetworkPolicy and RBAC.
- Script avoids printing Secret values.

### Current status

Implemented in `scripts/operability/health-check.sh`.

---

## FR-OPS-003 — Provide backup, restore, DR and maintenance runbooks

### Priority

MUST

### Description

The repository must document operational procedures.

### Required runbooks

- PostgreSQL backup/restore;
- disaster recovery;
- maintenance operations;
- operability health-check;
- secrets rotation;
- AuthN/AuthZ.

### Current status

Implemented.

---

# 14. Multi-developer behavior

## FR-MDEV-001 — Show requester

### Priority

MUST

### Description

The system must show which requester created a ChangeRequest.

### Acceptance criteria

- `/ui/changes` includes `REQUESTED BY` information.
- Dashboard recent changes includes requester context.
- Multi-developer test data remains understandable.

### Current status

Implemented and validated.

---

## FR-MDEV-002 — Limit dashboard recent changes

### Priority

MUST

### Description

The dashboard must remain readable when many ChangeRequests exist.

### Acceptance criteria

- Dashboard recent changes shows only the latest five changes.
- `/ui/changes` shows the full list.
- The `View all` link is available.

### Current status

Implemented and validated with six ChangeRequests.

---

# 15. Multi-environment direction

## FR-ENV-001 — Store target environment

### Priority

MUST

### Description

Every ChangeRequest must include a target environment.

### Canonical environments

```text
dev
staging
production
```

### Acceptance criteria

- The field is persisted and visible.
- Current default behavior remains compatible with `dev`.
- Future validation rejects unknown or disabled environments fail-closed.

### Current status

`targetEnvironment` exists. Full environment catalog validation is future implementation.

---

## FR-ENV-002 — Resolve environment configuration

### Priority

SHOULD

### Description

The system should resolve Kubernetes, Tekton, Argo CD, GitLab, AuthZ and evidence policy mappings from an environment catalog.

### Acceptance criteria

- Environment configuration model is documented.
- ConfigMap-driven catalog is selected for the first increment.
- `production` remains disabled until production guardrails are implemented.

### Current status

Documented design baseline.

---

## FR-ENV-003 — Support promotion metadata

### Priority

COULD

### Description

The system should support correlated ChangeRequests for promotion.

### Future metadata

```text
promotionGroupID
promotedFromChangeNumber
```

### Acceptance criteria

- Promotion metadata design is documented.
- Runtime implementation is planned for a future phase.

### Current status

Design in progress.

---

# 16. Error handling

## FR-ERR-001 — Return readable errors

### Priority

MUST

### Description

The system must transform technical failures into readable operational errors.

### Examples

- invalid request payload;
- missing required field;
- GitLab branch conflict;
- Tekton validation failure;
- Argo CD API error;
- Kubernetes evidence collection error;
- authorization failure.

### Acceptance criteria

- Errors include code, message, technical message where safe and recoverability flag.
- Errors linked to ChangeRequests are recorded as events where applicable.
- No Secret values are returned.

### Current status

Implemented for current APIs, with further refinement expected.

---

## FR-ERR-002 — Avoid ambiguous failure states

### Priority

MUST

### Description

A failed action must not leave the ChangeRequest in an ambiguous state.

### Rules

- Failed validation maps to `ValidationFailed`.
- Failed deployment check maps to an explicit deployment runtime status where possible.
- Failed evidence collection must return a readable error.
- Failed authorization must not execute the action.

### Current status

Implemented for main workflows.

---

# 17. Traceability matrix

| Use Case | Main requirements |
|---|---|
| UC-001 View application list | FR-APP-001, FR-UI-001 |
| UC-002 View application detail | FR-APP-002, FR-APP-003 |
| UC-003 Create ChangeRequest | FR-CHG-001, FR-CHG-002 |
| UC-004 Execute GitLab workflow | FR-GIT-002, FR-GIT-003, FR-GIT-004 |
| UC-005 Open and merge MR | FR-GIT-004, FR-GIT-005, FR-AUTH-002 |
| UC-006 Validate with Tekton | FR-TKN-001, FR-TKN-002, FR-TKN-003, FR-TKN-004 |
| UC-007 Check Argo CD deployment | FR-ARGO-001, FR-ARGO-002 |
| UC-008 Collect evidence | FR-K8S-001, FR-K8S-002, FR-EVD-001 |
| UC-009 Review history | FR-CHG-002, FR-EVD-002, FR-UI-002 |
| UC-010 Multi-developer dashboard | FR-MDEV-001, FR-MDEV-002 |
| UC-011 Operability health-check | FR-OPS-001, FR-OPS-002, FR-OPS-003 |
| UC-012 Environment promotion | FR-ENV-001, FR-ENV-002, FR-ENV-003 |

---

## 18. Out-of-scope requirements

The following are outside the current baseline:

- full enterprise ITSM integration;
- full environment administration UI;
- full production dual-approval implementation;
- complete multi-cluster orchestration;
- Secret management platform capabilities;
- visual YAML editor;
- AI assistant integration;
- complete Git provider abstraction;
- CLI `devopsctl`, currently optional and deferred.

---

## 19. Document completion criteria

This document is considered stable when:

- requirements are written in English;
- priorities are assigned;
- current implementation status is clear;
- API paths are aligned to `/api/v1`;
- security and operability requirements are included;
- multi-developer and multi-environment requirements are represented;
- requirements trace to current personas and use cases.

---

## 20. Key message

The functional requirements keep the DevOps Control Plane focused on a useful and safe operational baseline.

The system must support a reliable end-to-end flow:

```text
Application discovery
  -> ChangeRequest
  -> GitLab branch / commit / merge request
  -> Tekton validation
  -> Argo CD deployment check
  -> OpenShift runtime evidence
  -> PostgreSQL history
  -> Web UI visibility
  -> Operability and audit
```

The current baseline already implements this core flow. The next scope evolution is environment-aware promotion across:

```text
dev -> staging -> production
```

---

## 21. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial functional requirements document in Italian. |
| 2026-07-03 | 0.2 | Rewritten in English and refreshed to align with the current advanced MVP, security, operability, multi-developer and multi-environment baseline. |
