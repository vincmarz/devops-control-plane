# DevOps Control Plane — Personas and Use Cases

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 02 — Personas and Use Cases
- **Version:** 0.2
- **Date:** 2026-07-03
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Previous document:** `docs/01-scope-mvp.md`
- **Status:** Rewritten in English and refreshed to align with the current advanced MVP, security, operability and multi-environment baseline
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document describes the primary personas and use cases for the DevOps Control Plane.

The original version focused on the first MVP. The current version updates the personas and use cases to reflect the current advanced MVP baseline, including:

- ChangeRequest lifecycle and audit;
- GitLab technical workflow;
- Tekton validation and diagnostics;
- Argo CD deployment checks;
- Kubernetes/OpenShift runtime evidence;
- Web UI;
- OAuth Proxy and AuthN/AuthZ;
- requester visibility and multi-developer scenarios;
- operational runbooks and production-readiness baseline;
- future multi-environment direction across `dev`, `staging` and `production`.

This document provides input for:

- `docs/03-functional-requirements.md`;
- `docs/05-architecture.md`;
- `docs/13-api-design.md`;
- the final technical documentation.

---

## 2. User-centered design principles

The DevOps Control Plane must be designed around real operational workflows.

The system should:

1. guide users through standard GitOps changes;
2. reduce manual mistakes across GitLab, YAML, Tekton, Argo CD and OpenShift;
3. provide clear audit and evidence records;
4. keep Git as the source of desired state;
5. expose enough detail for experts without overwhelming onboarding users;
6. make requester, actor, status and evidence visible;
7. enforce AuthN/AuthZ and environment-specific guardrails;
8. support operational troubleshooting and production readiness.

The UI and API should not hide risk. Instead, the Control Plane should make the workflow understandable, repeatable and auditable.

---

## 3. Primary personas

## 3.1 P1 — Requester / Developer

### Description

The Requester or Developer initiates a change for an application managed through GitOps.

The requester may not execute every technical action directly, but the requester needs visibility into the status of the requested change.

### Goals

The Requester wants to:

- request a change through a clear form or API;
- specify application, environment, title, type, risk and description;
- understand the current status of the ChangeRequest;
- see whether validation, deployment and evidence collection succeeded;
- know who executed or approved follow-up actions;
- avoid losing visibility when multiple developers create changes at the same time.

### Example needs

- “I want to request an update to the demo application.”
- “I want to see whether my ChangeRequest is still in draft or has been executed.”
- “I want to know whether the change reached `Synced` and `Healthy` state.”
- “I want to find my ChangeRequest in the full change list.”

### Required information

- ChangeRequest number;
- application name;
- target environment;
- requester;
- lifecycle status;
- runtime status;
- validation result;
- deployment state;
- collected evidence.

---

## 3.2 P2 — DevOps Operator

### Description

The DevOps Operator is the primary operational user of the system.

The operator executes standard GitOps workflows on applications already managed through Argo CD and OpenShift.

### Goals

The DevOps Operator wants to:

- quickly understand application status;
- execute standard technical actions without manually running long command chains;
- reduce YAML and Git workflow mistakes;
- create branches and merge requests through guided actions;
- run Tekton validation;
- check Argo CD deployment state;
- collect runtime evidence;
- troubleshoot failed validations or deployments.

### Example needs

- “I want to validate a ChangeRequest through Tekton.”
- “I want to open a merge request for a generated GitOps change.”
- “I want to check whether Argo CD reports the application as `Synced` and `Healthy`.”
- “I want to collect evidence after deployment.”

### Required information

- Argo CD application status;
- GitLab branch, commit and merge request references;
- Tekton PipelineRun and TaskRun diagnostics;
- Kubernetes/OpenShift runtime state;
- runtime status;
- evidence records;
- recommended next actions.

---

## 3.3 P3 — Approver / Reviewer

### Description

The Approver or Reviewer validates whether a change is acceptable before approval, merge, execution or promotion.

This persona may be a senior DevOps engineer, platform owner, technical lead or environment-specific approver.

### Goals

The Approver wants to:

- understand what is changing;
- inspect validation results;
- assess risk level;
- review GitLab merge request links and evidence;
- approve or reject lifecycle transitions;
- enforce stricter rules for staging and production environments;
- understand rollback or recovery options.

### Example needs

- “Which files were changed?”
- “Did Tekton validation succeed?”
- “Is the application healthy after deployment?”
- “Is this change allowed for production?”
- “Which evidence supports this approval?”

### Required information

- ChangeRequest summary;
- title, type and risk level;
- requester;
- target environment;
- validation evidence;
- deployment evidence;
- GitLab merge request;
- audit events;
- approval policy.

---

## 3.4 P4 — Platform Engineer

### Description

The Platform Engineer manages OpenShift, Argo CD, Tekton integrations, RBAC and platform guardrails.

The Platform Engineer cares about application changes and also about whether the workflow respects platform rules.

### Goals

The Platform Engineer wants to:

- verify that application resources are consistent with GitOps governance;
- detect policy issues before runtime impact;
- validate RBAC and NetworkPolicy behavior;
- ensure the Control Plane uses least-privilege access;
- verify that runtime evidence collection is safe;
- keep production guardrails explicit and fail-closed.

### Example needs

- “Does the runtime ServiceAccount have only the expected permissions?”
- “Does this workflow require access to resources outside the expected namespace?”
- “Is the PostgreSQL NetworkPolicy still present?”
- “Are health endpoints publicly accessible while API and UI are protected?”
- “Is the target environment configured correctly?”

### Required information

- environment configuration;
- Kubernetes namespace mapping;
- Argo CD Application mapping;
- Tekton namespace and pipeline mapping;
- RBAC checks;
- NetworkPolicy status;
- OpenShift group bindings;
- runtime evidence payloads;
- operational smoke-test output.

---

## 3.5 P5 — Auditor / Change Manager

### Description

The Auditor or Change Manager needs to reconstruct what happened across the lifecycle of a change.

This persona may not modify manifests or operate the platform directly.

### Goals

The Auditor wants to:

- know who requested a change;
- know who executed actions;
- know when each action happened;
- review GitLab, Tekton, Argo CD and Kubernetes references;
- confirm that evidence was collected;
- inspect final state;
- understand whether production-readiness requirements were respected.

### Example needs

- “Who requested `CHG-2026-0006`?”
- “Which PipelineRun validated the change?”
- “Which evidence was collected after deployment?”
- “Was the anonymous UI/API access blocked?”
- “Which user or group was allowed to perform the action?”

### Required information

- ChangeRequest number;
- requester;
- actor information in audit events;
- timestamps;
- lifecycle status;
- runtime status;
- GitLab external references;
- Tekton external references;
- Argo CD state;
- Kubernetes/OpenShift evidence;
- production readiness notes where applicable.

---

## 3.6 P6 — Security Reviewer

### Description

The Security Reviewer verifies that the Control Plane enforces authentication, authorization, secret handling and secure integration practices.

### Goals

The Security Reviewer wants to:

- confirm that OAuth Proxy protects the UI and APIs;
- confirm that backend AuthZ is fail-closed;
- verify that OpenShift group lookup works;
- confirm that tokens are not printed or committed;
- review TLS strict settings;
- validate least-privilege RBAC;
- verify that production actions require stricter permissions.

### Example needs

- “Can an anonymous user access `/ui/dashboard`?”
- “Can an operator approve a production change?”
- “Are GitLab and Argo CD tokens stored only in Secrets?”
- “Is `KUBERNETES_TOKEN` removed from the application Secret?”
- “Does the runtime ServiceAccount have access to Secrets?”

### Required information

- AuthN/AuthZ runbook;
- OAuth Proxy configuration;
- backend role mapping;
- OpenShift group mapping;
- Secret inventory without values;
- RBAC `can-i` results;
- TLS and trust configuration;
- security audit events.

---

## 3.7 P7 — Operations Engineer

### Description

The Operations Engineer runs health-checks, maintenance, backup/restore validation and disaster recovery procedures.

### Goals

The Operations Engineer wants to:

- validate runtime health;
- run the automated smoke-test script;
- collect an evidence package;
- validate PostgreSQL backup and restore;
- execute maintenance or rollback procedures;
- understand whether a failure is runtime-impacting or historical/transient.

### Example needs

- “Is the current runtime healthy?”
- “Did the smoke test fail because of a current problem or an old rollout event?”
- “Can PostgreSQL be restored in an isolated namespace?”
- “Which runbook should be used for maintenance?”

### Required information

- health-check runbook;
- smoke-test output;
- event and log evidence;
- PostgreSQL counts;
- backup artifact checksums;
- DR runbook;
- maintenance runbook;
- production readiness checklist.

---

## 3.8 P8 — Newbie / Onboarding Reader

### Description

The Newbie or onboarding reader has basic or early knowledge of GitOps, OpenShift, Argo CD, Tekton and Kubernetes.

This persona needs guided documentation and UI visibility, not a black-box automation system.

### Goals

The onboarding reader wants to:

- understand why Git is the source of desired state;
- understand the difference between GitLab, Tekton, Argo CD and OpenShift;
- follow the lifecycle of a ChangeRequest;
- understand validation and evidence;
- learn from standardized workflows;
- map UI concepts to backend/API/runtime concepts.

### Example needs

- “Why should the change go through Git instead of `oc edit`?”
- “What does `OutOfSync` mean?”
- “What does `EvidenceCollected` mean?”
- “Why is Tekton validation required?”
- “What does the Control Plane add on top of Argo CD?”

### Required information

- clear UI labels;
- clear workflow steps;
- evidence summaries;
- readable operational messages;
- final technical documentation;
- links to relevant runbooks and ADRs.

---

## 4. Main use cases

## 4.1 UC-001 — View application list

### Primary persona

DevOps Operator

### Secondary personas

Platform Engineer, Newbie / Onboarding Reader

### Goal

Display applications managed by Argo CD and their main status.

### Preconditions

- Argo CD API is configured.
- A valid Argo CD token is available through Secret.
- At least one visible Application exists.

### Main flow

1. The user opens the Applications page or calls the application API.
2. The backend queries Argo CD API.
3. The backend normalizes application data.
4. The UI/API displays application name, project, namespace, sync status, health status and revision.

### Expected output

```text
Application          Project          Sync      Health    Revision
demo-go-color-app    default          Synced    Healthy   <revision>
```

### Acceptance criteria

- Applications are listed through API and UI.
- Main sync and health state are visible.
- Argo CD errors are returned clearly.
- Tokens are not exposed in logs or output.

---

## 4.2 UC-002 — View application detail

### Primary persona

DevOps Operator

### Secondary personas

Platform Engineer, Approver

### Goal

Display detailed operational state for an Argo CD Application.

### Main flow

1. The user selects an application.
2. The backend reads application detail from Argo CD.
3. The backend reads resources, health, sync, revision and available conditions.
4. The UI displays the result.

### Acceptance criteria

- The UI clearly shows `Synced`, `OutOfSync`, `Healthy` and `Degraded` states.
- Known warnings such as orphaned resources are visible.
- Healthy applications with warnings are not automatically treated as failed.

---

## 4.3 UC-003 — Create a ChangeRequest

### Primary persona

Requester / Developer

### Secondary personas

DevOps Operator, Newbie / Onboarding Reader

### Goal

Create a ChangeRequest with enough metadata to support workflow execution and audit.

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

### Main flow

1. The user creates a ChangeRequest through API or UI.
2. The backend validates required fields.
3. The backend stores the ChangeRequest in PostgreSQL.
4. The backend creates a change event.
5. The UI/API returns the ChangeRequest number.

### Acceptance criteria

- Required fields are validated.
- Unknown or disabled target environments must be rejected when environment validation is implemented.
- The ChangeRequest is visible in `/ui/changes`.
- The requester is visible in the UI.

---

## 4.4 UC-004 — Execute GitLab branch and file update workflow

### Primary persona

DevOps Operator

### Goal

Represent a GitOps change through GitLab branch and file updates.

### Main flow

1. The operator triggers branch creation.
2. The Control Plane creates a GitLab branch.
3. The operator triggers file update.
4. The Control Plane updates a GitOps manifest.
5. The Control Plane records technical events.
6. The ChangeRequest runtime status is updated.

### Acceptance criteria

- The runtime cluster is not modified directly as desired state.
- GitLab external references are recorded.
- A technical event is created for each step.

---

## 4.5 UC-005 — Open and merge a GitLab merge request

### Primary persona

DevOps Operator

### Secondary persona

Approver / Reviewer

### Goal

Open and optionally merge a GitLab merge request as part of a controlled workflow.

### Main flow

1. The Control Plane opens a merge request from the change branch to the target branch.
2. Reviewers inspect the merge request.
3. The authorized user merges the merge request where policy allows.
4. The Control Plane records merge request references and status.

### Acceptance criteria

- Merge request URL and ID are stored.
- Merge state is visible.
- Merge action is authorized.
- Merge result is recorded as a technical event.

---

## 4.6 UC-006 — Validate a change with Tekton

### Primary persona

DevOps Operator

### Secondary personas

Approver / Reviewer, Platform Engineer

### Goal

Validate a GitOps branch before considering the change technically safe.

### Main flow

1. The Control Plane creates a Tekton `PipelineRun`.
2. Tekton clones the selected branch/revision.
3. Tekton validates manifests and GitOps policy checks.
4. Tekton runs anti-secret checks.
5. The Control Plane checks `PipelineRun` and `TaskRun` status.
6. The Control Plane records validation evidence.

### Acceptance criteria

- `ValidationSucceeded` is assigned only after successful Tekton completion.
- Failed TaskRuns are visible through diagnostics.
- Validation evidence is stored and available through API/UI.

---

## 4.7 UC-007 — Check Argo CD deployment status

### Primary persona

DevOps Operator

### Secondary persona

Approver / Reviewer

### Goal

Check whether the application is deployed, synced and healthy through Argo CD.

### Main flow

1. The operator triggers deployment check.
2. The Control Plane calls Argo CD API.
3. The Control Plane maps Argo CD status to runtime status.
4. The result is stored and visible in UI/API.

### Acceptance criteria

- `Synced` and `Healthy` map to `DeploymentSyncedHealthy`.
- `OutOfSync`, `Progressing`, `Degraded` and unknown states are handled explicitly.
- Argo CD API errors are readable.

---

## 4.8 UC-008 — Collect deployment evidence

### Primary persona

Auditor / Change Manager

### Secondary personas

DevOps Operator, Operations Engineer

### Goal

Collect and store post-deployment evidence for a ChangeRequest.

### Evidence scope

- ChangeRequest summary;
- Argo CD status;
- Git revision;
- Kubernetes/OpenShift Deployment state;
- Pod readiness and restart counts;
- Service availability;
- Route availability;
- diagnostics summary;
- warnings and conditions.

### Acceptance criteria

- Evidence is associated with the ChangeRequest.
- Evidence is sanitized.
- Evidence is visible through API and UI.
- Evidence does not contain tokens or Secret values.

---

## 4.9 UC-009 — Review ChangeRequest history

### Primary persona

Auditor / Change Manager

### Secondary personas

Requester / Developer, Approver / Reviewer

### Goal

Review functional and technical history for a ChangeRequest.

### Main flow

1. The user opens the ChangeRequest list.
2. The system displays ChangeRequest number, application, requester, environment, lifecycle and runtime status.
3. The user opens a ChangeRequest detail page.
4. The system displays events, evidence and technical actions.

### Acceptance criteria

- It is possible to reconstruct the lifecycle of a change.
- Events are ordered and readable.
- Evidence is linked to the ChangeRequest.
- Requester visibility is available.

---

## 4.10 UC-010 — Use dashboard in a multi-developer scenario

### Primary persona

DevOps Operator

### Secondary personas

Requester / Developer, Auditor / Change Manager

### Goal

Keep the dashboard readable while multiple users create ChangeRequests.

### Main flow

1. Multiple developers create ChangeRequests.
2. The dashboard KPI counters aggregate over all loaded changes.
3. The dashboard recent changes section shows only the latest five changes.
4. The `/ui/changes` page shows the full list.
5. The requester is visible in both the dashboard and the full ChangeRequest list.

### Acceptance criteria

- Recent changes is limited to five items.
- The full list remains available through `View all`.
- Requester is visible.
- Change ordering is newest first.

---

## 4.11 UC-011 — Operability health-check

### Primary persona

Operations Engineer

### Goal

Validate current runtime health and collect an evidence directory.

### Main flow

1. The operator runs the smoke-test script.
2. The script checks namespace, deployments, pods, services, routes and PVC.
3. The script checks Route and backend health endpoints.
4. The script checks PostgreSQL connectivity and counts.
5. The script checks NetworkPolicy and RBAC.
6. The script writes an evidence package.

### Acceptance criteria

- Runtime readiness is validated.
- Evidence directory is generated.
- Port-forward conflicts are identifiable.
- Historical rollout events are distinguishable from current runtime impact.

---

## 4.12 UC-012 — Prepare environment-aware promotion

### Primary persona

Approver / Reviewer

### Secondary personas

Platform Engineer, Security Reviewer

### Goal

Prepare a future promotion workflow across `dev`, `staging` and `production`.

### Main flow

1. A change is validated in `dev`.
2. A related ChangeRequest is created for `staging`.
3. A later related ChangeRequest is created for `production`.
4. Promotion metadata links the related ChangeRequests.
5. Environment-aware AuthZ enforces stricter policies for production.

### Acceptance criteria

- Promotion model is documented.
- Environment configuration model is documented.
- Unknown or disabled environments are rejected fail-closed when implemented.
- Production remains disabled until guardrails are ready.

---

## 5. Out-of-scope use cases for the current baseline

The following use cases are outside the current advanced MVP baseline:

- full enterprise user/group management UI;
- full ITSM integration;
- ServiceNow or Jira integration;
- full application Secret management;
- full production dual approval;
- automatic multi-step rollback;
- complete policy engine UI;
- automatic AppProject generation;
- advanced visual YAML editor;
- full Backstage-style developer portal;
- full multi-cluster orchestration;
- full environment management UI;
- complete Git provider abstraction beyond GitLab;
- CLI `devopsctl`, currently optional and deferred.

---

## 6. Persona to use case matrix

| Persona | Primary use cases |
|---|---|
| Requester / Developer | UC-003, UC-009, UC-010 |
| DevOps Operator | UC-001, UC-002, UC-004, UC-005, UC-006, UC-007, UC-008, UC-010 |
| Approver / Reviewer | UC-005, UC-006, UC-007, UC-009, UC-012 |
| Platform Engineer | UC-001, UC-002, UC-006, UC-011, UC-012 |
| Auditor / Change Manager | UC-008, UC-009, UC-010 |
| Security Reviewer | UC-003, UC-005, UC-011, UC-012 |
| Operations Engineer | UC-008, UC-011 |
| Newbie / Onboarding Reader | UC-001, UC-002, UC-003, UC-009 |

---

## 7. Information visibility matrix

| Information | Requester | Operator | Approver | Platform | Auditor | Security | Operations | Newbie |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Application status | yes | yes | yes | yes | read | read | read | yes |
| ChangeRequest requester | yes | yes | yes | yes | yes | yes | read | yes |
| Lifecycle status | yes | yes | yes | yes | yes | yes | read | yes |
| Runtime status | yes | yes | yes | yes | yes | yes | yes | yes |
| GitLab references | read | yes | yes | read | yes | read | read | read |
| Tekton diagnostics | read | yes | yes | yes | yes | read | read | guided |
| Deployment evidence | read | yes | yes | yes | yes | read | yes | guided |
| RBAC/AuthZ details | no | limited | limited | yes | read | yes | read | no |
| Operational runbooks | no | read | read | yes | read | read | yes | guided |
| Environment promotion metadata | read | yes | yes | yes | yes | yes | read | guided |

---

## 8. Educational information requirements

The Control Plane should help users understand the workflow.

For each major workflow, the system should make clear:

- what is being done;
- why the action is needed;
- which external tool is involved;
- which file or resource is affected;
- which branch, commit or merge request was created;
- what Tekton validated;
- what Argo CD checked;
- what Kubernetes/OpenShift evidence was collected.

Example educational message:

```text
The requested change is represented in Git. Tekton validates the GitOps manifests before the change is considered technically safe. Argo CD then reconciles the target application from Git to the OpenShift runtime.
```

---

## 9. Error explanation requirements

The Control Plane should translate complex technical failures into actionable operational messages.

### Example 1 — Resource not permitted by Argo CD AppProject

Technical error:

```text
resource :ConfigMap is not permitted in project devops-ci-demo
```

Suggested message:

```text
The ConfigMap is present in the GitOps repository, but the Argo CD AppProject does not allow ConfigMap resources for this application. Review the namespaceResourceWhitelist before syncing.
```

### Example 2 — Invalid Kubernetes manifest

Technical error:

```text
valueFrom: Invalid value: must specify configMapKeyRef, secretKeyRef, fieldRef or resourceFieldRef
```

Suggested message:

```text
The Deployment contains an incomplete valueFrom block. Verify that configMapKeyRef, secretKeyRef, fieldRef or resourceFieldRef is correctly specified under valueFrom.
```

### Example 3 — Argo CD operation already in progress

Technical error:

```text
another operation is already in progress
```

Suggested message:

```text
Argo CD already has an operation in progress for this Application. Wait for the current operation to finish or investigate whether the operation is stuck.
```

### Example 4 — Local Git context error

Technical error:

```text
fatal: not a git repository
```

Suggested message:

```text
The Git command was executed outside the repository directory. Enter the correct repository directory before running Git operations.
```

---

## 10. Example journey — ChangeRequest validation and evidence

```text
1. A requester creates a ChangeRequest for demo-go-color-app.
2. The Control Plane stores the ChangeRequest in PostgreSQL.
3. A DevOps Operator creates a GitLab branch.
4. The Control Plane updates the GitOps file.
5. The Control Plane opens a GitLab merge request.
6. Tekton validation is triggered.
7. The Control Plane checks PipelineRun and TaskRun status.
8. Validation evidence is stored.
9. An authorized user merges the merge request.
10. Argo CD deployment status is checked.
11. Kubernetes/OpenShift runtime evidence is collected.
12. The ChangeRequest runtime status becomes EvidenceCollected.
13. The full history remains visible in the UI.
```

---

## 11. Example journey — Multi-developer dashboard

```text
1. Developer A creates CHG-2026-0002.
2. Developer B creates CHG-2026-0003.
3. Developer C creates CHG-2026-0004.
4. Developer D creates CHG-2026-0005.
5. Developer E creates CHG-2026-0006.
6. The dashboard shows only the latest five recent changes.
7. The full /ui/changes page shows all changes.
8. The REQUESTED BY column identifies each requester.
```

---

## 12. Minimum application configuration data

To manage an application, the system must know or resolve:

```yaml
applicationName: demo-go-color-app
argocdNamespace: openshift-gitops
argocdProject: default
targetNamespace: devops-ci-demo
repoProvider: gitlab
repoProjectId: "<gitlab-project-id>"
repoUrl: "https://gitlab.example.local/group/demo-app-gitops.git"
targetRevision: main
path: apps/demo-go-color-app
defaultBranch: main
targetEnvironment: dev
```

In the future environment-aware model, this mapping must be resolved through the environment catalog.

---

## 13. Relationship with other documents

This document informs:

- `docs/03-functional-requirements.md`;
- `docs/04-non-functional-requirements.md`;
- `docs/05-architecture.md`;
- `docs/13-api-design.md`;
- `docs/multi-environment-model.md`;
- `docs/environment-configuration-model.md`;
- `docs/change-promotion-model.md`;
- Phase 12 final technical documentation.

---

## 14. Key message

The DevOps Control Plane must be designed around real operational users and their workflows.

The system should be guided but transparent:

```text
Show what the system is doing.
Show why the system is doing it.
Show which external tool is involved.
Show which evidence is produced.
Show who requested and executed the change.
```

This balance helps onboarding users without hiding the technical details required by DevOps, platform, security and operations teams.

---

## 15. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial personas and use cases document in Italian. |
| 2026-07-03 | 0.2 | Rewritten in English and refreshed to align with the current advanced MVP, security, operability, multi-developer and multi-environment baseline. |
