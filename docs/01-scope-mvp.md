# DevOps Control Plane — MVP and Advanced Baseline Scope

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 01 — Scope
- **Version:** 0.2
- **Date:** 2026-07-03
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Previous document:** `docs/00-vision.md`
- **Status:** Rewritten in English and refreshed to align with the current advanced MVP, security and operational baseline
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document defines the functional and technical scope of the DevOps Control Plane.

The original version described the first MVP scope. The project has since evolved into an advanced MVP with a production-oriented operational baseline. This refreshed version preserves the original intent but updates the scope to reflect the current implementation and near-term roadmap.

The document clarifies:

- what the current baseline includes;
- which capabilities are considered part of the advanced MVP;
- which capabilities are outside the current scope;
- which integrations are required;
- which workflows are supported;
- which operational and security expectations are in scope;
- which items belong to future phases.

This document should be used as a scope reference for backend, API, adapter, UI, security, operability and documentation work.

---

## 2. Scope summary

The DevOps Control Plane provides a guided orchestration and governance layer above GitLab, Tekton, Argo CD and OpenShift.

The current baseline allows authorized users and operators to:

1. view GitOps applications managed by Argo CD;
2. inspect essential GitOps and runtime state;
3. create and track ChangeRequests;
4. execute GitLab branch, file update, merge request and merge workflows;
5. validate changes through Tekton;
6. check Argo CD deployment state;
7. collect Kubernetes/OpenShift runtime evidence;
8. store audit events and evidence in PostgreSQL;
9. use a server-side Web UI for dashboard, changes, evidence and actions;
10. authenticate through OpenShift OAuth Proxy;
11. authorize actions through trusted headers, OpenShift group lookup and backend AuthZ;
12. operate with backup/restore, disaster recovery, maintenance and readiness runbooks;
13. support multi-developer visibility;
14. evolve toward `dev`, `staging` and `production` environment support.

The DevOps Control Plane is not a replacement for GitLab, Argo CD, Tekton or OpenShift. It coordinates and governs their usage.

---

## 3. Technology baseline

### 3.1 Backend

```text
Go
```

The backend is responsible for:

- HTTP API exposure;
- server-side Web UI rendering;
- ChangeRequest orchestration;
- lifecycle transitions;
- audit event creation;
- GitLab API integration;
- Argo CD API integration;
- Kubernetes API integration for Tekton;
- OpenShift/Kubernetes runtime evidence collection;
- PostgreSQL persistence;
- AuthN/AuthZ enforcement;
- configuration loading;
- health and readiness endpoints.

### 3.2 Frontend / UI

```text
Go server-side HTML templates
```

The current UI is no longer just a placeholder. It includes:

- dashboard;
- KPI counters;
- recent changes limited to the latest five items;
- applications view;
- full ChangeRequest list;
- requester visibility;
- ChangeRequest detail;
- server-side technical actions;
- evidence pages;
- audit event pages;
- settings page;
- UI wrapper for raw Changes API.

The UI must remain English according to the repository language policy.

### 3.3 Database

```text
PostgreSQL
```

PostgreSQL stores:

- ChangeRequests;
- lifecycle and technical events;
- evidence records;
- runtime status;
- functional audit history;
- references to external systems.

PostgreSQL is part of the operational baseline and is covered by backup/restore and disaster recovery runbooks.

### 3.4 Git integration

```text
GitLab API
```

The Control Plane integrates with GitLab API to:

- create branches;
- update GitOps files;
- create commits;
- open merge requests;
- merge merge requests;
- store external references in ChangeRequest events;
- support GitOps traceability.

The DevOps Control Plane source code may be hosted on GitHub, while target application and GitOps repositories are integrated through GitLab API.

### 3.5 Argo CD integration

```text
Argo CD API
```

The Control Plane uses Argo CD API to:

- list applications;
- get application details;
- read sync and health status;
- read revision and conditions;
- check deployment state;
- collect deployment evidence;
- expose deployment diagnostics.

TLS strict mode is part of the current security baseline.

### 3.6 Tekton integration

```text
Kubernetes API
```

The Control Plane interacts with Tekton through Kubernetes API resources:

- `PipelineRun` creation;
- `PipelineRun` status polling;
- `TaskRun` status collection;
- validation evidence creation;
- validation failure diagnostics.

Tekton validation includes GitOps policy checks and anti-secret guardrails.

### 3.7 OpenShift / Kubernetes integration

The Control Plane uses OpenShift/Kubernetes APIs to:

- collect runtime deployment evidence;
- inspect Deployments, Pods, Services and Routes;
- use ServiceAccount token fallback for Kubernetes authentication;
- enforce runtime RBAC least privilege;
- operate behind OpenShift OAuth Proxy.

---

## 4. In-scope capabilities

## 4.1 Application discovery

### Description

The system must discover and display applications managed by Argo CD.

### Minimum data

- Application name;
- Argo CD namespace;
- Argo CD project;
- target namespace;
- repository URL;
- GitOps path;
- target revision;
- sync status;
- health status;
- current revision;
- conditions and warnings where available.

### Current status

This capability is implemented as part of the Argo CD integration and UI baseline.

---

## 4.2 Application details

### Description

The system must expose operational detail for an Argo CD Application.

### Minimum data

- sync status;
- health status;
- current revision;
- repository URL;
- target revision;
- Git path;
- target namespace;
- managed resources;
- orphaned resources or conditions;
- history where available.

### Current status

Application detail is available through API and UI. The implementation is expected to evolve further with environment-aware Argo CD mapping.

---

## 4.3 ChangeRequest lifecycle

### Description

The system must provide a ChangeRequest model with functional lifecycle and technical runtime status.

### In-scope lifecycle capabilities

- create ChangeRequest;
- submit ChangeRequest;
- approve ChangeRequest;
- start execution;
- complete execution;
- close ChangeRequest;
- record lifecycle events;
- record technical events;
- keep runtime status separate from lifecycle status.

### Current status

Implemented with PostgreSQL persistence and tests.

---

## 4.4 GitLab branch and file update workflow

### Description

The system must support controlled GitOps file updates through GitLab API.

### In-scope actions

- create branch;
- update GitOps files;
- create commit;
- open merge request;
- merge merge request where allowed;
- record technical events and external references.

### Current status

Implemented and runtime validated for the lab environment.

---

## 4.5 Tekton validation

### Description

The system must delegate validation to Tekton before considering a change technically validated.

### In-scope validation

- create `PipelineRun`;
- check latest PipelineRun by ChangeRequest;
- collect TaskRun diagnostics;
- store validation evidence;
- report validation failures clearly;
- enforce GitOps anti-secret and manifest guardrails.

### Current status

Implemented with validation evidence, diagnostics and hardened policy checks.

---

## 4.6 Argo CD deployment check

### Description

The system must check the deployment state through Argo CD and map Argo CD status to Control Plane runtime status.

### In-scope runtime statuses

- `DeploymentSyncedHealthy`;
- `DeploymentProgressing`;
- `DeploymentOutOfSync`;
- `DeploymentDegraded`;
- `DeploymentUnknown`.

### Current status

Implemented and validated against the demo application.

---

## 4.7 Runtime evidence collection

### Description

The system must collect technical evidence after deployment.

### Evidence scope

- ChangeRequest metadata;
- Argo CD sync and health state;
- Git revision;
- Kubernetes/OpenShift Deployment state;
- Pods and readiness;
- restart counts;
- Service availability;
- Route availability;
- diagnostics summary;
- warnings and conditions.

### Current status

Implemented with sanitized deployment evidence and UI rendering.

---

## 4.8 Functional audit history

### Description

The system must provide a functional history that allows operators to understand what happened without manually correlating GitLab, Tekton, Argo CD and Kubernetes data.

### Minimum data

- ChangeRequest number;
- application;
- target environment;
- requester;
- lifecycle status;
- runtime status;
- technical events;
- evidence;
- timestamps;
- external references.

### Current status

Implemented through PostgreSQL-backed ChangeRequests, change events and evidence records.

---

## 4.9 Web UI

### Description

The Web UI must provide an operational interface for common workflows and status review.

### Current UI scope

- dashboard;
- KPI counters;
- recent changes;
- applications list;
- ChangeRequest list;
- ChangeRequest detail;
- requester visibility;
- technical actions;
- audit event view;
- evidence view;
- raw API wrapper page.

### Current status

Implemented and validated, including multi-developer display behavior.

---

## 4.10 AuthN/AuthZ and security baseline

### Description

The system must protect APIs and UI through authenticated and authorized access.

### Current security scope

- OpenShift OAuth Proxy;
- trusted-header authentication;
- backend AuthZ middleware;
- OpenShift group lookup;
- roles: `viewer`, `operator`, `approver`, `admin`;
- anonymous API/UI blocked through Route;
- public health endpoints;
- TLS strict integration baseline;
- app-dedicated trust bundle;
- Secret rotation runbook;
- Kubernetes ServiceAccount token fallback;
- RBAC least privilege;
- PostgreSQL NetworkPolicy.

### Current status

Implemented as an advanced security baseline.

---

## 4.11 Operability baseline

### Description

The system must be operable and support production-oriented maintenance procedures.

### Current operability scope

- PostgreSQL operability inventory;
- backup and restore runbook;
- isolated restore validation;
- operability health-check runbook;
- automated smoke-test script;
- disaster recovery runbook;
- maintenance operations runbook;
- production readiness checklist.

### Current status

Implemented and formally closed as Phase 10 advanced operational baseline.

---

## 4.12 Multi-developer visibility

### Description

The system must remain understandable when multiple developers or operators create and execute ChangeRequests.

### Current scope

- `requestedBy` visible in ChangeRequest list;
- requester visible in dashboard recent changes;
- dashboard recent changes limited to the latest five entries;
- full history available through `/ui/changes`;
- audit events available per ChangeRequest.

### Current status

Validated through controlled multi-developer test data.

---

## 4.13 Multi-environment direction

### Description

The system must evolve toward environment-aware workflows.

### Target environments

```text
dev
staging
production
```

### In-scope design baseline

- single multi-environment Control Plane instance;
- correlated ChangeRequests for promotion;
- environment-aware RBAC/AuthZ;
- ConfigMap-driven environment catalog;
- future promotion metadata such as `promotionGroupID` and `promotedFromChangeNumber`.

### Current status

Architecture decision and environment configuration model are documented. Runtime implementation is future scope.

---

## 5. Out-of-scope capabilities for the current baseline

The following capabilities are outside the current advanced MVP baseline:

- full enterprise multi-tenant model;
- full ITSM integration;
- ServiceNow/Jira integration;
- dynamic Secret management platform;
- full production dual-approval workflow;
- full environment management UI;
- database-backed environment catalog;
- advanced visual YAML editor;
- AI assistant integration;
- complete replacement of Argo CD UI;
- complete replacement of Tekton Dashboard;
- complete GitHub, Gitea or Bitbucket target provider support;
- full OpenShift cluster disaster recovery;
- DR of GitLab, Argo CD and Tekton external systems;
- CLI `devopsctl`, currently optional and deferred.

These items may be considered in later roadmap phases.

---

## 6. Non-functional requirements

### 6.1 Security

The system must:

- keep GitLab tokens in OpenShift Secrets;
- keep Argo CD tokens in OpenShift Secrets;
- keep PostgreSQL credentials in OpenShift Secrets;
- avoid printing tokens or passwords in logs;
- avoid committing secrets to Git;
- enforce AuthN/AuthZ on APIs and UI;
- use TLS strict mode where possible;
- use least-privilege RBAC;
- restrict PostgreSQL ingress through NetworkPolicy;
- provide documented rotation procedures.

### 6.2 Auditability

Every ChangeRequest must provide enough audit data to answer:

- who requested the change;
- who executed actions;
- when actions happened;
- which workflow steps were executed;
- which external systems were involved;
- which evidence was collected;
- what the final state was.

### 6.3 Repeatability

A workflow must be repeatable and evidence-driven.

When a validation or deployment check fails, the system should record:

- failed phase;
- reason;
- diagnostic payload;
- related external references;
- recommended troubleshooting direction where possible.

### 6.4 Separation of responsibilities

Responsibilities remain separated:

- GitLab manages repositories, branches, commits and merge requests;
- Argo CD manages GitOps reconciliation;
- Tekton manages validation and automation;
- OpenShift/Kubernetes provides runtime state;
- PostgreSQL stores functional audit and evidence;
- DevOps Control Plane orchestrates and correlates.

### 6.5 Operability

The system must be supported by:

- health and readiness endpoints;
- smoke-test automation;
- backup/restore procedures;
- disaster recovery procedures;
- maintenance procedures;
- production readiness checklist.

---

## 7. ChangeRequest state model

The current model distinguishes between lifecycle status and runtime status.

### 7.1 Lifecycle status

Lifecycle status represents business/process state, such as:

```text
draft
submitted
approved
executing
executed
closed
```

### 7.2 Runtime status

Runtime status represents technical execution state, such as:

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

### 7.3 Event rule

Every significant transition must create a `change_events` record.

---

## 8. API scope

The API currently follows a versioned `/api/v1` model and includes endpoints for:

### Applications

```text
GET /api/v1/applications
GET /api/v1/applications/{name}
GET /api/v1/applications/{name}/resources
GET /api/v1/applications/{name}/history
GET /api/v1/applications/{name}/runtime
```

### ChangeRequests

```text
POST /api/v1/changes
GET  /api/v1/changes
GET  /api/v1/changes/{id}
POST /api/v1/changes/{id}/submit
POST /api/v1/changes/{id}/approve
POST /api/v1/changes/{id}/start-execution
POST /api/v1/changes/{id}/complete-execution
POST /api/v1/changes/{id}/close
POST /api/v1/changes/{id}/create-branch
POST /api/v1/changes/{id}/update-files
POST /api/v1/changes/{id}/open-merge-request
POST /api/v1/changes/{id}/merge-request
POST /api/v1/changes/{id}/validate
POST /api/v1/changes/{id}/check-validation
POST /api/v1/changes/{id}/check-deployment
POST /api/v1/changes/{id}/collect-evidence
GET  /api/v1/changes/{id}/evidence
GET  /api/v1/changes/{id}/evidence/{type}
GET  /api/v1/changes/{id}/events
```

### Health

```text
GET /readyz
GET /livez
```

Legacy or early endpoint references should be updated to the current `/api/v1` model during documentation migration.

---

## 9. Implementation milestones — current status

The initial MVP milestone plan has been superseded by actual implementation progress.

### Completed foundation

- Go project bootstrap;
- PostgreSQL integration;
- migrations;
- ChangeRequest model;
- lifecycle transitions;
- audit events;
- health endpoints.

### Completed integrations

- GitLab create branch;
- GitLab update files;
- GitLab open merge request;
- GitLab merge request merge;
- Tekton validation;
- Tekton diagnostics;
- Argo CD application client;
- Argo CD deployment check;
- Kubernetes/OpenShift runtime evidence.

### Completed UI baseline

- dashboard;
- applications pages;
- ChangeRequest pages;
- evidence pages;
- audit pages;
- technical action forms;
- requester visibility;
- dashboard recent changes limit.

### Completed security and operations baseline

- OAuth Proxy;
- AuthN/AuthZ;
- OpenShift group lookup;
- TLS strict baseline;
- RBAC least privilege;
- NetworkPolicy;
- PostgreSQL backup/restore;
- DR;
- maintenance;
- production readiness.

### Design in progress

- multi-environment support;
- promotion metadata;
- final technical documentation;
- documentation language migration.

---

## 10. Global acceptance criteria for the advanced MVP baseline

The advanced MVP baseline is considered achieved when the following are true:

1. the service runs on OpenShift;
2. PostgreSQL persistence is operational;
3. health and readiness endpoints work;
4. Argo CD applications can be queried;
5. ChangeRequests can be created and tracked;
6. lifecycle and technical events are persisted;
7. GitLab technical workflow is available;
8. Tekton validation can be triggered and checked;
9. validation evidence is stored;
10. Argo CD deployment state can be checked;
11. Kubernetes/OpenShift runtime evidence is collected;
12. the UI exposes dashboard, changes, evidence and audit views;
13. OAuth Proxy protects UI and API routes;
14. backend AuthZ enforces role-based access;
15. Secrets are not printed or committed;
16. TLS strict baseline is implemented;
17. RBAC least privilege is implemented;
18. PostgreSQL NetworkPolicy is implemented;
19. backup/restore is documented and validated;
20. disaster recovery and maintenance runbooks exist;
21. production readiness checklist exists;
22. multi-developer visibility is validated.

---

## 11. Main risks and mitigations

### 11.1 Workflow complexity

Risk:

```text
Building too many workflows before stabilizing the core path.
```

Mitigation:

```text
Keep workflows incremental and evidence-driven. Validate each technical step before expanding the scope.
```

### 11.2 Integration drift

Risk:

```text
GitLab, Tekton, Argo CD and OpenShift behavior may drift from repository assumptions.
```

Mitigation:

```text
Use adapters, runtime validation, smoke tests and runbooks.
```

### 11.3 GitOps drift

Risk:

```text
Runtime changes may bypass Git.
```

Mitigation:

```text
Keep Git as the source of desired state. Treat imperative runtime commands as troubleshooting tools, not final workflow state.
```

### 11.4 Credential exposure

Risk:

```text
GitLab, Argo CD, Kubernetes or database credentials may be exposed in logs or output.
```

Mitigation:

```text
Use OpenShift Secrets, avoid decoding values in shared logs, enforce anti-secret checks and rotate exposed credentials.
```

### 11.5 Production complexity

Risk:

```text
Production environments require stricter approval, evidence and operational controls.
```

Mitigation:

```text
Introduce environment-aware RBAC/AuthZ, promotion chains and production-specific guardrails before enabling production workflows.
```

---

## 12. Recommended execution direction

The recommended direction is:

```text
1. Keep the current advanced MVP stable.
2. Continue documentation language migration.
3. Complete multi-environment design.
4. Introduce promotion metadata incrementally.
5. Add environment-aware UI filtering.
6. Extend environment-aware AuthZ.
7. Enable staging before production.
8. Keep production disabled until guardrails are validated.
```

---

## 13. Key message

The DevOps Control Plane scope is to provide a small but solid orchestration, governance and audit layer for GitOps workflows.

The project started with a first MVP scope, but it has evolved into an advanced operational baseline with security, evidence, UI, backup/restore and production-readiness capabilities.

The next scope evolution is controlled multi-environment support across:

```text
dev -> staging -> production
```

The priority remains the same:

```text
safe GitOps workflows
clear audit trail
repeatable validation
consistent evidence
operational readiness
```

---

## 14. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial MVP scope in Italian. |
| 2026-07-03 | 0.2 | Rewritten in English and refreshed to align with the current advanced MVP, security and operational baseline. |
