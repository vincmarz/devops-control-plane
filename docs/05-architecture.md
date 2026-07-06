# DevOps Control Plane — Architecture

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 05 — Architecture
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
- **Status:** Rewritten in English and refreshed to align with the current advanced MVP, security, operability and multi-environment baseline
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document describes the current architecture of the DevOps Control Plane.

The original architecture document described the initial MVP direction. This refreshed version updates the architecture to reflect the current implementation baseline, including:

- Go backend;
- PostgreSQL persistence;
- versioned `/api/v1` API;
- server-side Web UI;
- GitLab integration;
- Tekton integration through Kubernetes API;
- Argo CD integration;
- Kubernetes/OpenShift runtime evidence collection;
- OpenShift OAuth Proxy;
- backend AuthN/AuthZ;
- OpenShift group lookup;
- TLS strict and trust bundle model;
- Secret/token handling;
- RBAC least privilege;
- PostgreSQL NetworkPolicy;
- operability runbooks and smoke-test automation;
- backup/restore, disaster recovery and maintenance baselines;
- multi-developer UI validation;
- multi-environment direction for `dev`, `staging` and `production`.

The goal is to provide a technical architecture that is implementable, auditable, understandable for onboarding readers and aligned with the current repository language policy.

---

## 2. Executive summary

The DevOps Control Plane is a Go application deployed on OpenShift. It orchestrates GitOps workflows by coordinating external systems rather than replacing them.

```text
GitLab       = repository, branch, commit and merge request workflow
Tekton       = GitOps validation and technical execution diagnostics
Argo CD      = GitOps deployment state and reconciliation status
OpenShift    = runtime platform and Kubernetes/OpenShift evidence source
PostgreSQL   = ChangeRequest, events and evidence store
Go backend   = orchestration, API, UI, AuthZ, adapters and evidence correlation
OAuth Proxy  = OpenShift-backed authentication gateway for Route traffic
```

The main architectural principle remains:

```text
Every permanent desired-state change must be represented in Git.
```

The Control Plane is not an alternative runtime editor for the OpenShift cluster. It is an orchestration, governance, evidence and audit layer over the GitOps toolchain.

---

## 3. High-level architecture

```text
User / Browser / API Client
        |
        v
OpenShift Route
        |
        v
OpenShift OAuth Proxy
        |
        v
DevOps Control Plane Go backend
        |
        +--> PostgreSQL repository layer
        |
        +--> GitLab adapter
        |
        +--> Tekton adapter through Kubernetes API
        |
        +--> Argo CD adapter
        |
        +--> Kubernetes/OpenShift runtime evidence adapter
        |
        +--> AuthN/AuthZ middleware and OpenShift group resolver
```

External systems:

```text
GitLab API
Argo CD API
Kubernetes/OpenShift API
Tekton CRDs
PostgreSQL
OpenShift OAuth
OpenShift groups
OpenShift image registry
```

---

## 4. Architectural principles

### 4.1 Git remains the source of desired state

Application workload changes must be represented as Git changes. Runtime imperative commands may be used only for troubleshooting or controlled operational procedures, not as the final workflow state.

### 4.2 Argo CD remains the reconciliation engine

The Control Plane checks and interprets Argo CD state. It does not implement its own deployment engine.

### 4.3 Tekton remains the validation engine

The Control Plane creates and observes Tekton `PipelineRun` and `TaskRun` resources through Kubernetes API.

### 4.4 PostgreSQL stores functional history and evidence

GitLab, Tekton and Argo CD provide technical histories. PostgreSQL stores the Control Plane model: ChangeRequests, events, runtime status and evidence.

### 4.5 Security and operability are part of the architecture

OAuth Proxy, backend AuthZ, RBAC, TLS, Secret handling, NetworkPolicy and runbooks are not add-ons. They are part of the architecture baseline.

---

## 5. Component model

## 5.1 OpenShift Route and OAuth Proxy

### Responsibilities

- expose the application through OpenShift Route;
- authenticate browser/API users through OpenShift OAuth;
- block anonymous access to UI and API routes;
- allow unauthenticated access to safe health endpoints;
- forward trusted identity headers to the backend.

### Current behavior

- anonymous `/api/v1/changes` through Route returns HTTP 403;
- anonymous `/ui/dashboard` through Route returns HTTP 403;
- `/readyz` and `/livez` return HTTP 200 when healthy.

---

## 5.2 HTTP API layer

### Responsibilities

- expose versioned REST endpoints under `/api/v1`;
- validate HTTP input;
- call application services;
- translate errors into structured responses;
- enforce backend AuthZ through middleware;
- avoid direct low-level integration logic in handlers.

### Core endpoint areas

```text
GET  /readyz
GET  /livez
GET  /api/v1/applications
GET  /api/v1/applications/{name}
GET  /api/v1/changes
GET  /api/v1/changes/{id}
POST /api/v1/changes
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
GET  /api/v1/changes/{id}/events
```

---

## 5.3 Web UI layer

### Responsibilities

- provide server-side rendered pages;
- expose dashboard, applications and ChangeRequest views;
- expose evidence and audit event pages;
- provide server-side technical actions;
- show requester and runtime status;
- keep UI labels in English.

### Current UI baseline

- dashboard with KPI counters;
- recent changes limited to five items;
- `View all` link to full ChangeRequest list;
- requester visible in recent changes and `/ui/changes`;
- ChangeRequest detail with actions, evidence and audit links;
- UI wrapper for raw Changes API.

---

## 5.4 Domain layer

### Responsibilities

The domain layer contains the conceptual model:

- Application;
- ChangeRequest;
- ChangeEvent;
- Evidence;
- lifecycle status;
- runtime status;
- Git references;
- Tekton validation results;
- Argo CD deployment results;
- Kubernetes/OpenShift evidence summaries.

### Rules

- domain types do not depend on GitLab HTTP details;
- domain types do not depend on Argo CD HTTP details;
- domain types do not depend on concrete Kubernetes clients;
- domain and service logic must be unit-testable.

---

## 5.5 Application services

### Responsibilities

Application services orchestrate business use cases and integration ports.

Current service responsibilities include:

- listing and retrieving applications;
- creating and reading ChangeRequests;
- lifecycle transitions;
- technical status updates;
- GitLab workflow orchestration;
- Tekton validation orchestration;
- Argo CD deployment checks;
- evidence collection;
- audit event persistence.

---

## 5.6 PostgreSQL repository layer

### Responsibilities

The PostgreSQL layer stores persistent functional state.

Main entities:

- `applications` where used by the baseline;
- `change_requests`;
- `change_events`;
- `evidences`.

### Rules

- timestamps must be preserved;
- JSON payloads must be sanitized;
- Secret values must not be persisted;
- critical state changes must be consistent;
- backup and restore procedures must protect this data.

---

## 5.7 Evidence collector

### Responsibilities

The evidence collector normalizes and sanitizes evidence from external systems.

Evidence sources:

- Tekton validation status and diagnostics;
- Argo CD deployment status;
- Kubernetes/OpenShift Deployment, Pod, Service and Route state;
- ChangeRequest metadata;
- GitLab references.

### Rules

- do not store tokens;
- do not store raw Secret values;
- include useful summaries;
- preserve enough payload for audit and troubleshooting;
- include target environment where available.

---

# 6. Adapter layer

The adapter layer isolates the application/domain logic from external APIs.

## 6.1 GitLab adapter

### Responsibilities

- create branches;
- update files;
- open merge requests;
- merge merge requests;
- normalize GitLab API errors;
- return external references to the ChangeService.

### Rules

- GitLab token must never be logged;
- branch conflicts must be handled clearly;
- file update failures must be readable;
- commit and merge references must be recorded;
- TLS strict mode must be supported.

---

## 6.2 Argo CD adapter

### Responsibilities

- list applications;
- get application details;
- read sync and health status;
- read revision and conditions;
- check deployment state;
- preserve warnings such as `OrphanedResourceWarning`.

### Rules

- Argo CD token must never be logged;
- TLS strict mode must be used;
- healthy applications with warnings must not be automatically classified as failed;
- deployment status must be mapped to Control Plane runtime status.

---

## 6.3 Kubernetes/OpenShift adapter

### Responsibilities

- read runtime Deployment state;
- read Pod status and restart counts;
- read Service state;
- read OpenShift Route state;
- use ServiceAccount token fallback;
- use ServiceAccount CA where applicable;
- support runtime evidence collection.

### Rules

- runtime evidence collection is read-only;
- ServiceAccount permissions must remain least-privilege;
- static Kubernetes tokens should not be required;
- Secret values must not be read unless explicitly required and approved.

---

## 6.4 Tekton adapter

### Responsibilities

- create `PipelineRun` resources;
- read `PipelineRun` status;
- list associated `TaskRun` resources;
- collect diagnostics;
- map Tekton status to validation runtime status.

### Rules

- `PipelineRun` names and labels should include ChangeRequest context;
- timeout behavior must be explicit;
- validation evidence must be stored;
- TaskRun failures must be visible.

---

## 6.5 AuthN/AuthZ components

### Responsibilities

- read trusted user and group headers;
- resolve OpenShift group membership when enabled;
- map groups to roles;
- enforce endpoint/action authorization;
- fail closed for unknown endpoints or unauthorized actions.

### Current roles

```text
viewer
operator
approver
admin
```

Future environment-aware authorization will include target environment and approval policy.

---

# 7. Go package architecture

Current package layout follows an adapter-based architecture.

Representative structure:

```text
cmd/
internal/
  api/
  app/
  domain/
  adapters/
    argocd/
    gitlab/
    kubernetes/
    tekton/
    tlsutil/
  config/
  database/
  logging/
  workflow/
migrations/
manifests/
pipelines/
scripts/
docs/
```

Package rules:

- `domain` contains core models;
- `app` orchestrates use cases through interfaces;
- `adapters` implement concrete integrations;
- `api` exposes HTTP handlers and UI handlers;
- `database` encapsulates PostgreSQL access;
- `config` loads runtime configuration;
- `scripts` contains operational automation.

---

# 8. Main architectural flows

## 8.1 Application discovery

```text
User or API client
  -> OAuth Proxy where accessed through Route
  -> HTTP API or UI handler
  -> Application service
  -> Argo CD adapter
  -> Argo CD API
  -> normalized Application view
  -> JSON response or UI page
```

---

## 8.2 ChangeRequest creation

```text
User or API client
  -> POST /api/v1/changes
  -> AuthZ middleware
  -> ChangeService validation
  -> PostgreSQL create ChangeRequest
  -> PostgreSQL create change event
  -> API response
  -> UI list and detail visibility
```

Required fields include:

```text
title
applicationName
changeType
requestedBy
```

`targetEnvironment` currently defaults to `dev` when omitted and will become fail-closed against the environment catalog in a future implementation.

---

## 8.3 GitLab technical workflow

```text
ChangeRequest
  -> create branch
  -> update GitOps files
  -> open merge request
  -> optionally merge merge request
  -> store external references
  -> update runtime status
  -> create audit events
```

The runtime cluster is not directly modified as desired state.

---

## 8.4 Tekton validation workflow

```text
ChangeRequest
  -> create Tekton PipelineRun
  -> Tekton clones branch/revision
  -> Tekton validates manifests and policy checks
  -> Control Plane checks PipelineRun
  -> Control Plane lists TaskRuns
  -> validation evidence is stored
  -> runtime status becomes ValidationSucceeded or ValidationFailed
```

---

## 8.5 Argo CD deployment check workflow

```text
ChangeRequest
  -> Argo CD adapter reads Application
  -> sync status and health status are mapped
  -> runtime status is updated
  -> warnings are preserved
```

Mapping examples:

```text
Synced + Healthy       -> DeploymentSyncedHealthy
OutOfSync              -> DeploymentOutOfSync
Progressing            -> DeploymentProgressing
Degraded               -> DeploymentDegraded
Unknown/other          -> DeploymentUnknown
```

---

## 8.6 Runtime evidence workflow

```text
ChangeRequest
  -> Argo CD state collection
  -> Kubernetes/OpenShift runtime collection
  -> Deployment, Pods, Service, Route checks
  -> diagnostics summary generation
  -> sanitized evidence persisted in PostgreSQL
  -> UI evidence rendering
```

---

## 8.7 Operability smoke-test workflow

```text
Operator
  -> scripts/operability/health-check.sh
  -> namespace/workload checks
  -> route checks
  -> backend checks through port-forward
  -> PostgreSQL counts
  -> NetworkPolicy and RBAC checks
  -> evidence directory
```

This workflow is read-only and must not print Secret values.

---

# 9. Data flow and persistence

## 9.1 Transient data

Transient data should not be persisted if it contains sensitive values.

Examples:

- API tokens;
- raw HTTP responses containing sensitive headers;
- temporary YAML files;
- kube client session data;
- decoded Secret values.

## 9.2 Persistent data

Persistent data includes:

- ChangeRequest metadata;
- lifecycle and technical events;
- GitLab external references;
- Tekton external references;
- Argo CD deployment state;
- sanitized evidence payloads;
- diagnostics summaries.

## 9.3 Forbidden persistent data

The following must not be persisted:

- GitLab token;
- Argo CD token;
- Kubernetes bearer token;
- kubeconfig;
- PostgreSQL password in clear text;
- Kubernetes Secret values;
- Docker auth JSON values;
- private keys.

---

# 10. Deployment architecture on OpenShift

## 10.1 Current runtime model

Current OpenShift deployment baseline:

```text
Namespace: devops-control-plane
Deployment: devops-control-plane
Application container: devops-control-plane
Sidecar container: oauth-proxy
Service: devops-control-plane
Route: devops-control-plane
ConfigMaps: application configuration and trust bundle
Secrets: application tokens and database URL
ServiceAccount: devops-control-plane-oauth-proxy
PostgreSQL: in-cluster deployment with PVC
NetworkPolicy: restrict PostgreSQL ingress
```

## 10.2 Container requirements

- non-privileged runtime;
- configurable HTTP port;
- `/readyz` readiness endpoint;
- `/livez` liveness endpoint;
- ConfigMap/Secret-driven configuration;
- no embedded secrets in the image;
- resource requests and limits for application and OAuth Proxy containers.

## 10.3 Image build and deployment

Images are built with Podman and pushed to the OpenShift registry.

Given bastion space constraints, builds should use temporary storage under `/tmp` rather than `/home`.

---

# 11. Security architecture

## 11.1 Trust boundaries

```text
Browser/API client
  -> OpenShift Route
  -> OAuth Proxy
  -> Go backend
  -> PostgreSQL
  -> external APIs
```

Main trust boundaries:

- public Route;
- OAuth Proxy to backend;
- trusted headers;
- database access;
- Secret access;
- external API access;
- evidence persistence.

## 11.2 Secrets

Main Secret keys include:

```text
GITLAB_TOKEN
ARGOCD_AUTH_TOKEN
DATABASE_URL
```

`KUBERNETES_TOKEN` is not required in the current baseline because Kubernetes/OpenShift access uses ServiceAccount token fallback.

Rules:

- no token in Git;
- no token in logs;
- no token in evidence;
- no decoded Secret value in operational outputs;
- token exposure requires rotation.

## 11.3 RBAC

The runtime ServiceAccount requires least-privilege permissions:

- create and read Tekton PipelineRuns in the target namespace;
- read TaskRuns;
- read application runtime resources for evidence;
- read OpenShift groups through a dedicated cluster binding;
- no default access to Secrets;
- no cluster-admin permissions.

## 11.4 TLS and trust bundle

TLS strict mode is required for supported integrations.

The architecture supports an app-dedicated trust bundle mounted into the application container.

---

# 12. Operability architecture

Operability is documented through runbooks and scripts.

Key assets:

```text
docs/runbooks/operability-health-check.md
scripts/operability/health-check.sh
docs/runbooks/postgresql-backup-restore.md
docs/runbooks/disaster-recovery.md
docs/runbooks/maintenance-operations.md
docs/production-readiness-checklist.md
docs/phase-10-operability-closure.md
```

Operability requirements include:

- read-only health checks;
- sanitized evidence collection;
- PostgreSQL backup and restore validation;
- isolated restore tests;
- DR runbook;
- maintenance runbook;
- production readiness checklist.

---

# 13. Error architecture

## 13.1 Error model

Errors should include:

```yaml
code: VALIDATION_INVALID_REQUEST
message: Unable to create ChangeRequest
technicalMessage: title is required
recoverable: true
```

Rules:

- expose technical detail only when safe;
- do not expose Secret values;
- include recoverability where useful;
- persist errors as events when associated with ChangeRequests.

## 13.2 Error families

```text
VALIDATION_*
AUTH_*
GITLAB_*
ARGOCD_*
KUBERNETES_*
TEKTON_*
DATABASE_*
WORKFLOW_*
```

---

# 14. Multi-developer architecture considerations

The UI and backend must support multiple users creating ChangeRequests.

Current behavior:

- `requestedBy` is stored and displayed;
- `/ui/changes` shows all ChangeRequests;
- dashboard recent changes shows only the latest five;
- ChangeRequest detail remains individually addressable;
- audit events and evidence remain per ChangeRequest.

This prevents the dashboard from becoming a long unbounded history while preserving the full list.

---

# 15. Multi-environment architecture direction

The selected future architecture is:

```text
Single DevOps Control Plane instance
Multiple target environments
Correlated ChangeRequests for promotion
Environment-aware RBAC/AuthZ
```

Canonical environments:

```text
dev
staging
production
```

Design documents:

```text
docs/adr/ADR-0011-multi-environment-model.md
docs/multi-environment-model.md
docs/environment-configuration-model.md
docs/change-promotion-model.md
```

Expected future architecture additions:

- environment catalog loaded from ConfigMap;
- environment-specific GitLab path mapping;
- environment-specific Tekton namespace/pipeline mapping;
- environment-specific Argo CD Application mapping;
- environment-specific Kubernetes/OpenShift namespace mapping;
- production-specific AuthZ policy;
- promotion metadata such as `promotionGroupID` and `promotedFromChangeNumber`.

---

# 16. Architectural decisions

Core ADRs include:

```text
ADR-0001 — Git as source of truth
ADR-0002 — Argo CD as GitOps engine
ADR-0003 — Tekton as validation engine
ADR-0004 — PostgreSQL as change history and evidence store
ADR-0005 — API-first before full Web UI
ADR-0006 — Adapter-based architecture
ADR-0007 — GitLab API as Git provider integration
ADR-0008 — Kubernetes API for Tekton integration
ADR-0009 — AuthN/AuthZ strategy
ADR-0010 — OAuth Proxy deployment design
ADR-0011 — Multi-environment DevOps Control Plane model
```

The ADR index must remain aligned with `docs/adr/README.md`.

---

# 17. Architecture boundaries and non-goals

## 17.1 In scope

- single Go backend service;
- PostgreSQL-backed ChangeRequest and evidence store;
- GitLab API adapter;
- Argo CD API adapter;
- Kubernetes/OpenShift adapter;
- Tekton adapter;
- server-side Web UI;
- OpenShift deployment with OAuth Proxy;
- AuthN/AuthZ and group lookup;
- current dev environment runtime baseline;
- future multi-environment design.

## 17.2 Out of scope for the current baseline

- microservices split;
- event bus architecture;
- full enterprise ITSM integration;
- full environment management UI;
- full production promotion workflow implementation;
- complete Git provider abstraction;
- automatic AppProject generation;
- full policy engine UI;
- full replacement of Argo CD UI or Tekton Dashboard.

---

# 18. Architecture validation checklist

The architecture is valid when:

- Go backend builds and runs;
- PostgreSQL connectivity works;
- `/readyz` and `/livez` work;
- APIs use `/api/v1`;
- UI pages render through authenticated access;
- OAuth Proxy protects UI and API routes;
- backend AuthZ is fail-closed;
- GitLab, Tekton, Argo CD and Kubernetes integrations are adapter-based;
- evidence is persisted and sanitized;
- RBAC is least-privilege;
- Secrets are not printed;
- TLS strict mode is enabled where supported;
- smoke-test script validates the runtime;
- documentation and ADRs are aligned.

---

# 19. Key message

The DevOps Control Plane architecture must remain simple, modular and consistent with GitOps.

The value is not building a new deployment engine. The value is providing a safe orchestration, governance, audit and evidence layer for the flow:

```text
GitLab change
  -> Tekton validation
  -> Argo CD deployment check
  -> OpenShift runtime evidence
  -> PostgreSQL history
  -> Web UI visibility
  -> Operability baseline
```

Each component has a clear responsibility. Each workflow must produce useful evidence for requesters, operators, approvers, platform engineers, security reviewers, operations engineers and auditors.

---

## 20. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial architecture document in Italian. |
| 2026-07-06 | 0.2 | Rewritten in English and refreshed to align with the current advanced MVP, security, operability, UI, multi-developer and multi-environment baseline. |
