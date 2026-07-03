# DevOps Control Plane — Vision

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 00 — Vision
- **Version:** 0.2
- **Date:** 2026-07-03
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Status:** Refreshed vision aligned with the current advanced MVP and operational baseline
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Context

Modern OpenShift and GitOps platforms rely on multiple specialized tools working together:

- **Git** stores the desired application and platform state.
- **Argo CD / OpenShift GitOps** reconciles the desired state from Git with the runtime state in the cluster.
- **Tekton / OpenShift Pipelines** validates, automates and records technical execution steps.
- **OpenShift / Kubernetes** provides the runtime platform.
- **GitOps repositories** make infrastructure and application changes reviewable and auditable.
- **PostgreSQL** can provide an application-level audit and evidence store that complements Git, Argo CD and Kubernetes histories.

These tools are powerful, but a production-oriented GitOps workflow still requires a significant amount of operational knowledge.

A simple change, such as updating an application version, changing a `ConfigMap`, adjusting replicas or promoting a deployment between environments, can involve many steps:

1. identify the correct GitOps manifest;
2. create a branch;
3. update YAML or Kustomize overlays;
4. validate syntax and policy rules;
5. open or update a merge request;
6. run Tekton validation;
7. merge the approved change;
8. verify Argo CD sync and health;
9. verify Kubernetes/OpenShift runtime state;
10. collect evidence;
11. update the ChangeRequest runtime status;
12. provide a clear audit trail;
13. rollback or recover if something goes wrong.

The **DevOps Control Plane** exists to standardize these workflows and make them safer, more repeatable, more auditable and easier to understand for experienced and less experienced operators.

---

## 2. Problem statement

GitOps changes can be performed manually by combining:

- `git` commands;
- GitLab API or UI actions;
- Argo CD commands or UI actions;
- `oc` / `kubectl` commands;
- Tekton / `tkn` operations;
- manual YAML editing;
- manual evidence collection.

This approach works, but it introduces risks:

- long command sequences that are hard to remember;
- copy and paste mistakes;
- changes applied directly to the cluster without Git traceability;
- invalid YAML committed without validation;
- missing policy checks;
- incomplete evidence collection;
- inconsistent audit trails;
- unclear ownership of who requested, approved or executed a change;
- difficult rollback analysis;
- difficulty onboarding new team members;
- inconsistent handling of development, staging and production environments.

The DevOps Control Plane reduces these risks by providing an orchestration and governance layer over GitLab, Tekton, Argo CD and OpenShift.

---

## 3. Product vision

The DevOps Control Plane is an internal platform application that helps teams manage GitOps changes through a guided, auditable and production-oriented workflow.

The Control Plane should allow authorized users to:

- view applications managed by Argo CD;
- inspect sync status, health status, current revision and runtime diagnostics;
- create and track ChangeRequests;
- execute standard GitLab branch, file update, merge request and merge workflows;
- run Tekton validation pipelines;
- collect validation and deployment evidence;
- check Argo CD deployment state;
- collect Kubernetes/OpenShift runtime evidence;
- view audit events and evidence in a Web UI;
- enforce AuthN/AuthZ policies;
- provide a clear operational baseline for backup, restore, disaster recovery and maintenance;
- evolve toward multi-environment workflows across `dev`, `staging` and `production`.

The Control Plane does not replace Git, Argo CD, Tekton or OpenShift. It coordinates and governs their usage.

---

## 4. Guiding principles

### 4.1 Git remains the source of desired state

Permanent application and platform changes must be represented in Git.

A GitOps change should be traceable through:

- branch;
- commit;
- merge request;
- merge into target branch;
- Argo CD reconciliation;
- runtime verification;
- evidence collection.

The Control Plane must not use imperative runtime commands as the final desired-state mechanism.

Examples of commands that may be useful for troubleshooting but must not represent the final GitOps workflow are:

```bash
oc edit deployment
oc patch deployment
oc set env deployment
oc scale deployment
```

### 4.2 Argo CD remains the GitOps reconciliation engine

The DevOps Control Plane does not implement its own deployment engine.

Argo CD remains responsible for:

- reading desired state from Git;
- comparing desired and runtime state;
- reporting `Synced`, `OutOfSync`, `Healthy` and `Degraded` states;
- performing sync operations where appropriate;
- exposing application resources, history and conditions.

The Control Plane integrates with Argo CD through the Argo CD API.

### 4.3 Tekton provides validation and technical execution evidence

Tekton is used as the validation and automation engine for GitOps changes.

Validation can include:

- cloning the selected Git revision;
- validating manifests;
- building Kustomize output;
- running dry-run checks;
- enforcing anti-secret and manifest guardrail policies;
- collecting TaskRun and PipelineRun diagnostics.

The Control Plane integrates with Tekton through the Kubernetes API by creating and observing `PipelineRun` and `TaskRun` resources.

### 4.4 PostgreSQL stores the functional audit and evidence model

Git and Argo CD provide technical history, but they do not fully answer functional audit questions such as:

- who requested the change?
- why was the change requested?
- which workflow steps were executed?
- which Tekton validation run was associated with the change?
- which evidence was collected?
- which runtime state was observed after deployment?
- which user performed each action?

The DevOps Control Plane uses PostgreSQL to store:

- ChangeRequests;
- lifecycle events;
- technical workflow events;
- evidence records;
- runtime status;
- audit context.

### 4.5 API-first foundation with an operational Web UI

The project started with an API-first approach to stabilize the backend workflows before investing in the UI.

The current baseline now includes a server-side Web UI that exposes:

- dashboard;
- application views;
- ChangeRequest list and detail pages;
- technical actions;
- audit event views;
- evidence views;
- multi-developer visibility through requester information;
- dashboard KPI counters and recent changes.

The UI must remain aligned with backend workflows and must not hide operational risk.

### 4.6 Security and operability are first-class concerns

The Control Plane must be safe to operate in a production-oriented environment.

This requires:

- OAuth Proxy integration;
- trusted-header AuthN/AuthZ;
- OpenShift group lookup;
- least-privilege RBAC;
- TLS strict integration where possible;
- safe Secret and token handling;
- NetworkPolicy baseline;
- PostgreSQL backup and restore runbooks;
- disaster recovery runbook;
- maintenance operations runbook;
- production readiness checklist;
- automated smoke-test script.

---

## 5. Core capabilities

### 5.1 Application discovery and runtime visibility

The Control Plane provides visibility into Argo CD applications and their runtime state.

Typical application attributes include:

- application name;
- Argo CD project;
- target namespace;
- Git repository;
- GitOps path;
- target revision;
- sync status;
- health status;
- current revision;
- known warnings or conditions.

### 5.2 ChangeRequest management

The Control Plane provides a ChangeRequest model to track functional and technical change execution.

A ChangeRequest captures:

- title;
- application name;
- target environment;
- change type;
- risk level;
- requester;
- description;
- lifecycle status;
- runtime status;
- audit events;
- evidence.

### 5.3 GitLab workflow orchestration

The Control Plane integrates with GitLab API to support:

- branch creation;
- GitOps file updates;
- merge request creation;
- merge request merge;
- commit and merge references;
- traceability from ChangeRequest to Git activity.

The project source code can be hosted on GitHub while the product integrates with GitLab for target application and GitOps repositories.

### 5.4 Tekton validation

The Control Plane can trigger Tekton validation and check the result.

The validation workflow includes:

- creating a `PipelineRun`;
- checking `PipelineRun` status;
- collecting `TaskRun` diagnostics;
- recording validation evidence;
- detecting validation failures;
- exposing diagnostics in API and UI.

### 5.5 Argo CD deployment checks

The Control Plane can check Argo CD deployment state and map it to runtime status values such as:

- deployment synced and healthy;
- deployment progressing;
- deployment out of sync;
- deployment degraded;
- deployment unknown.

### 5.6 Kubernetes/OpenShift runtime evidence

The Control Plane collects runtime evidence from Kubernetes/OpenShift resources, including:

- Deployment status;
- Pod readiness;
- restart counts;
- Service availability;
- Route availability;
- runtime diagnostics.

Evidence is sanitized before being stored or displayed.

### 5.7 Web UI

The Web UI provides a human-friendly operational layer.

Current UI capabilities include:

- dashboard with KPI counters;
- recent changes limited to the latest items;
- full ChangeRequest list;
- requester visibility for multi-developer scenarios;
- ChangeRequest detail pages;
- recommended and advanced technical actions;
- evidence views;
- audit event views;
- settings page;
- UI wrapper for raw Changes API.

---

## 6. Current implementation baseline

The current implementation is an advanced MVP and production-oriented baseline.

Completed capabilities include:

- Go backend skeleton and service structure;
- PostgreSQL integration with migrations;
- ChangeRequest persistence;
- lifecycle transitions and audit events;
- GitLab branch, update files, open merge request and merge request workflows;
- Tekton validation integration;
- Tekton validation evidence and failure diagnostics;
- GitOps validation policy hardening;
- Argo CD application client and deployment check workflow;
- post-deployment evidence collection;
- Kubernetes/OpenShift runtime evidence collection;
- UI dashboard and server-side actions;
- UI evidence rendering and deployment diagnostics;
- OpenShift deployment;
- OAuth Proxy integration;
- AuthN/AuthZ middleware;
- OpenShift group lookup;
- RBAC least-privilege baseline;
- TLS strict baseline for integrations;
- removal of static Kubernetes token through ServiceAccount token fallback;
- NetworkPolicy baseline for PostgreSQL;
- PostgreSQL backup/restore validation;
- operability runbooks;
- automated smoke-test script;
- disaster recovery runbook;
- maintenance runbook;
- production readiness checklist;
- multi-developer dashboard validation;
- initial multi-environment architecture decision and configuration model.

---

## 7. Security and governance baseline

The current security baseline includes:

- OpenShift OAuth Proxy in front of the application;
- backend authorization based on trusted headers;
- OpenShift group lookup;
- roles such as `viewer`, `operator`, `approver` and `admin`;
- anonymous API and UI access blocked through the Route;
- health endpoints exposed safely;
- TLS strict configuration for Argo CD, GitLab and Kubernetes/OpenShift integrations;
- app-dedicated trust bundle for route CA validation;
- Secret inventory and rotation runbook;
- static Kubernetes token removal;
- least-privilege RBAC for Tekton and runtime evidence collection;
- PostgreSQL ingress NetworkPolicy.

Security is expected to evolve with environment-aware policies for `dev`, `staging` and `production`.

---

## 8. Operability and production-readiness baseline

The Control Plane has a documented operational baseline.

Key operational deliverables include:

- PostgreSQL operability inventory;
- PostgreSQL backup and restore runbook;
- isolated restore validation;
- operability health-check runbook;
- automated health-check script;
- OAuth Proxy resource tuning;
- disaster recovery runbook;
- maintenance operations runbook;
- production readiness checklist;
- formal Phase 10 closure document.

The latest healthy operational baseline expects a smoke-test result similar to:

```text
PASS=35
WARN=0
FAIL=0
```

A temporary failure caused only by historical rollout events can be considered non-blocking when runtime health, Route checks, backend checks, PostgreSQL checks, NetworkPolicy and RBAC checks are all healthy.

---

## 9. Multi-developer collaboration

The Control Plane must help multiple developers and operators work on changes without losing visibility.

The current UI supports:

- requester visibility in the ChangeRequest list;
- requester visibility in dashboard recent changes;
- dashboard recent changes limited to the latest five items;
- full history available through `/ui/changes`;
- audit events and evidence per ChangeRequest.

A controlled multi-developer validation demonstrated that multiple ChangeRequests from different requesters remain visible and traceable.

---

## 10. Multi-environment direction

The next architectural direction is multi-environment support.

The selected model is:

```text
Single DevOps Control Plane instance
Multiple target environments
Correlated ChangeRequests for promotion
Environment-aware RBAC/AuthZ
```

Canonical environment keys are:

```text
dev
staging
production
```

The target promotion flow is:

```text
dev -> staging -> production
```

The initial design introduces:

- environment configuration model;
- ConfigMap-driven environment catalog;
- future promotion metadata such as `promotionGroupID` and `promotedFromChangeNumber`;
- environment-aware GitLab, Tekton, Argo CD and Kubernetes/OpenShift mappings;
- production-specific guardrails.

---

## 11. What the DevOps Control Plane is not

The DevOps Control Plane is not:

- a replacement for Git;
- a replacement for Argo CD;
- a replacement for Tekton;
- a replacement for OpenShift or Kubernetes;
- a full ITSM system;
- a generic multi-cloud orchestration product;
- a Secret management platform;
- a complete clone of Argo CD UI or Tekton Dashboard.

The Control Plane is a governance, orchestration and audit layer that makes these tools easier and safer to use together.

---

## 12. Initial and future scope

### 12.1 Current advanced MVP scope

Current scope includes:

- application visibility;
- ChangeRequest lifecycle;
- GitLab technical workflow;
- Tekton validation;
- Argo CD deployment check;
- runtime evidence collection;
- server-side UI;
- audit and evidence storage;
- AuthN/AuthZ;
- OpenShift deployment;
- operational runbooks.

### 12.2 Future scope

Future scope includes:

- full environment-aware workflows;
- promotion from `dev` to `staging` to `production`;
- environment selector and filters in the UI;
- richer production approval policies;
- improved reporting and final technical documentation;
- optional CLI support if Phase 11 is re-prioritized.

---

## 13. Success criteria

The DevOps Control Plane is successful if it helps the team:

- reduce manual GitOps execution errors;
- make change execution repeatable;
- make evidence collection consistent;
- improve auditability;
- clarify who requested and executed actions;
- simplify onboarding for new team members;
- preserve Git as the source of truth;
- integrate safely with Argo CD, Tekton, GitLab and OpenShift;
- operate with clear backup, restore, DR and maintenance procedures;
- support future `dev`, `staging` and `production` workflows with explicit guardrails.

---

## 14. Key message

The DevOps Control Plane helps DevOps and platform teams execute GitOps changes in a guided, secure and auditable way.

It does not replace Git, Argo CD, Tekton or OpenShift.

It provides the orchestration, governance, evidence and audit layer that makes these tools safer and more reliable for daily operational workflows.

---

## 15. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial project vision in Italian. |
| 2026-07-03 | 0.2 | Rewritten in English and refreshed to align with the current advanced MVP, security, operability and multi-environment baseline. |
