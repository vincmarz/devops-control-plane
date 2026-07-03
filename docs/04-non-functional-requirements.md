# DevOps Control Plane — Non-Functional Requirements

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 04 — Non-Functional Requirements
- **Version:** 0.2
- **Date:** 2026-07-03
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Previous documents:**
  - `docs/00-vision.md`
  - `docs/01-scope-mvp.md`
  - `docs/02-personas-use-cases.md`
  - `docs/03-functional-requirements.md`
- **Status:** Rewritten in English and refreshed to align with the current advanced MVP, security, operability and production-readiness baseline
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document defines the non-functional requirements of the DevOps Control Plane.

Non-functional requirements describe the expected qualities of the system independently from individual functional capabilities. This refreshed version aligns the original MVP-oriented document with the current advanced MVP and operational baseline.

This document covers:

- security;
- authentication and authorization;
- credential and Secret handling;
- auditability;
- observability;
- operability;
- reliability;
- resilience;
- performance;
- scalability;
- maintainability;
- configurability;
- usability;
- GitOps compliance;
- error handling;
- backup, restore and disaster recovery;
- OpenShift deployment requirements;
- testing;
- documentation and repository language requirements;
- multi-developer and multi-environment readiness.

This document complements `docs/03-functional-requirements.md` and should be used as input for architecture, deployment, runbooks, production readiness and final technical documentation.

---

## 2. Requirement conventions

### 2.1 Priority levels

| Priority | Meaning |
|---|---|
| MUST | Required for the current baseline, security or safe operation |
| SHOULD | Important and expected, but may be implemented incrementally |
| COULD | Useful future enhancement |
| WON'T | Explicitly outside the current baseline |

### 2.2 General principle

The DevOps Control Plane orchestrates GitOps workflows and integrates with systems that can affect application and platform runtime state:

- GitLab;
- Tekton;
- Argo CD;
- OpenShift/Kubernetes;
- PostgreSQL.

Therefore, even the MVP baseline must be designed with strong attention to:

- credential protection;
- traceability;
- predictable failure behavior;
- safe logging;
- separation between Git desired state and cluster runtime state;
- avoidance of permanent runtime changes outside Git;
- production-oriented operability.

---

# 3. Security requirements

## NFR-SEC-001 — Secure token and credential handling

### Priority

MUST

### Description

The system must handle tokens and credentials safely for:

- GitLab API;
- Argo CD API;
- Kubernetes/OpenShift API;
- PostgreSQL.

### Rules

- Tokens must not be stored in Git.
- Tokens must not be printed in application logs.
- Tokens must not be returned by APIs.
- Tokens must not be included in evidence payloads.
- Tokens and credentials must be provided through OpenShift Secrets or equivalent secure runtime mechanisms.
- Local `.env` files are allowed only for local development and must remain ignored by Git.

### Acceptance criteria

- `.gitignore` excludes `.env`, `*.env`, private keys and local secret artifacts.
- Authentication errors do not print token values.
- Evidence payloads are sanitized.
- Secret inventory procedures list keys only, never values.

### Current status

Implemented through OpenShift Secrets, runbooks and operational guardrails.

---

## NFR-SEC-002 — Least privilege

### Priority

MUST

### Description

Credentials and ServiceAccounts used by the DevOps Control Plane must have only the permissions required for the intended workflows.

### Rules

For GitLab:

- read configured repositories;
- create branches;
- update files;
- create commits;
- open and merge merge requests where allowed.

For Argo CD:

- read Applications;
- read resources and history;
- perform deployment checks;
- use a dedicated account with minimum required permissions.

For Kubernetes/OpenShift:

- create and read Tekton `PipelineRun` resources in the configured namespace;
- read `TaskRun` status and diagnostics;
- read runtime resources required for evidence;
- avoid cluster-admin privileges by default;
- deny access to application Secrets unless explicitly required and approved.

### Acceptance criteria

- Runtime RBAC is least-privilege.
- `oc auth can-i` checks are documented in the operability smoke test.
- The runtime ServiceAccount cannot read DevOps Control Plane Secrets.
- Production-related permissions are not granted before production guardrails are implemented.

### Current status

Implemented as an advanced baseline for the current OpenShift lab.

---

## NFR-SEC-003 — Separation of sensitive and non-sensitive configuration

### Priority

MUST

### Description

The system must distinguish between non-sensitive configuration and sensitive values.

### Rules

- Non-sensitive application configuration can be stored in ConfigMaps.
- Passwords, tokens, private keys and credentials must be stored in Secrets.
- The Control Plane must not create workflows that store Secrets in Git or ConfigMaps.
- Environment configuration should not contain secret values.

### Acceptance criteria

- ConfigMap examples contain no real tokens.
- Secret templates use placeholders only.
- Documentation clearly distinguishes ConfigMaps and Secrets.

### Current status

Implemented through manifests, runbooks and repository conventions.

---

## NFR-SEC-004 — Anti-secret validation

### Priority

MUST

### Description

The system must include anti-secret guardrails in GitOps validation.

### Patterns and objects

Validation should detect or block:

- Kubernetes `Secret` manifests where not explicitly allowed;
- inline password-like values;
- token-like values;
- client secrets;
- private keys;
- AWS access key patterns;
- bearer tokens or authorization headers;
- Docker auth configuration artifacts.

### Acceptance criteria

- Tekton validation blocks suspicious Secret-like content.
- Policy failure messages are actionable.
- Safe external Secret references can be allowed where policy permits.

### Current status

Implemented in the Tekton GitOps validation pipeline.

---

## NFR-SEC-005 — TLS and trust management

### Priority

MUST

### Description

The system must use strict TLS verification for supported integrations and must avoid relying on insecure TLS flags in the production-oriented baseline.

### Rules

- Argo CD integration must use TLS strict mode.
- GitLab integration must use TLS strict mode.
- Kubernetes/OpenShift integration must use ServiceAccount token and CA where possible.
- App-dedicated trust bundles should be used when cluster-wide trust bundles are insufficient.

### Acceptance criteria

- `ARGOCD_INSECURE_TLS=false`.
- `GITLAB_INSECURE_TLS=false`.
- `KUBERNETES_INSECURE_TLS=false`.
- The application can validate the relevant route certificates through a configured trust bundle.

### Current status

Implemented for the current baseline.

---

# 4. Authentication and authorization requirements

## NFR-AUTH-001 — OpenShift OAuth Proxy protection

### Priority

MUST

### Description

The UI and API exposed through the OpenShift Route must be protected by OpenShift OAuth Proxy.

### Acceptance criteria

- Anonymous Route access to API endpoints returns HTTP 403.
- Anonymous Route access to UI pages returns HTTP 403.
- Public health endpoints remain accessible.
- Browser login through OpenShift OAuth works.

### Current status

Implemented.

---

## NFR-AUTH-002 — Backend authorization

### Priority

MUST

### Description

The backend must enforce authorization based on trusted headers and resolved OpenShift groups.

### Roles

```text
viewer
operator
approver
admin
```

### Rules

- Anonymous direct backend API access must be rejected.
- Unknown or unclassified endpoints must fail closed.
- Operators, approvers and admins must have differentiated permissions.
- OpenShift group lookup must be supported.

### Acceptance criteria

- Viewer can read allowed resources.
- Operator can execute allowed technical actions.
- Approver can execute approval actions where authorized.
- Admin can perform administrative actions.
- Users without mapped groups are rejected.

### Current status

Implemented.

---

## NFR-AUTH-003 — Environment-aware authorization readiness

### Priority

SHOULD

### Description

Authorization must evolve to include target environment context.

### Future dimensions

```text
actor
role
action
targetEnvironment
approvalPolicy
```

### Acceptance criteria

- `dev`, `staging` and `production` can have differentiated authorization policy.
- Production actions require production-appropriate groups before being enabled.
- Unknown or disabled environments are rejected fail-closed.

### Current status

Design documented. Runtime implementation is future scope.

---

# 5. Auditability requirements

## NFR-AUD-001 — End-to-end ChangeRequest traceability

### Priority

MUST

### Description

Every ChangeRequest must be traceable from creation to final state.

### Minimum data

- ChangeRequest number;
- application;
- target environment;
- change type;
- requester;
- lifecycle status;
- runtime status;
- GitLab branch;
- commit;
- merge request;
- Tekton PipelineRun;
- Argo CD deployment state;
- evidence records;
- main timestamps.

### Acceptance criteria

- Every significant state transition creates an event in `change_events`.
- Every ChangeRequest has a readable timeline.
- Events can be ordered by timestamp.
- UI and API expose audit history.

### Current status

Implemented.

---

## NFR-AUD-002 — Persistent evidence

### Priority

MUST

### Description

Key evidence must be persisted in PostgreSQL or in storage referenced by PostgreSQL.

### Evidence types

- validation evidence;
- deployment evidence;
- runtime evidence;
- diagnostics summaries.

### Acceptance criteria

- Evidence is associated with a ChangeRequest.
- Evidence is available after application restart.
- Evidence contains no tokens or Secret values.
- Evidence can be reviewed through API and UI.

### Current status

Implemented.

---

## NFR-AUD-003 — Clear distinction between external histories

### Priority

SHOULD

### Description

The system must make clear the difference between:

- Git history;
- Argo CD history;
- Tekton execution history;
- DevOps Control Plane ChangeRequest history.

### Educational rule

```text
Git describes what changed in files.
Argo CD describes what was reconciled to the cluster.
Tekton describes validation and automation execution.
DevOps Control Plane describes why the change happened, who requested it, which workflow was executed and which evidence was collected.
```

### Acceptance criteria

- ChangeRequest detail separates GitLab, Tekton, Argo CD and evidence references.
- The UI does not merge unrelated histories into a single ambiguous status.

### Current status

Implemented in the current UI/API baseline, with ongoing refinement expected.

---

# 6. Observability and operability requirements

## NFR-OBS-001 — Structured logging

### Priority

MUST

### Description

The backend must produce useful logs without exposing sensitive values.

### Recommended fields

- timestamp;
- level;
- component;
- request ID if available;
- change number where applicable;
- application name where applicable;
- operation;
- status;
- error code where applicable.

### Rules

- Do not log tokens.
- Do not log passwords.
- Do not log full payloads when they may contain sensitive values.

### Acceptance criteria

- Errors produce log entries with actionable context.
- Sensitive values are not printed.

### Current status

Implemented as a baseline.

---

## NFR-OBS-002 — Health and readiness endpoints

### Priority

MUST

### Description

The service must expose health endpoints.

### Endpoints

```text
GET /livez
GET /readyz
```

### Rules

- `/livez` verifies that the process is alive.
- `/readyz` verifies at least database readiness.
- Health endpoints must be reachable through the Route without authentication.
- UI and API business endpoints must remain protected.

### Acceptance criteria

- Route `/livez` returns HTTP 200 when healthy.
- Route `/readyz` returns HTTP 200 when ready.
- Backend direct endpoints return HTTP 200 when healthy.

### Current status

Implemented.

---

## NFR-OBS-003 — Automated operability smoke test

### Priority

MUST

### Description

The repository must include an automated smoke-test script for runtime validation.

### Requirements

The smoke test should validate:

- namespace state;
- deployments;
- pods;
- services;
- route;
- PVC;
- logs;
- events;
- sanitized configuration;
- route endpoints;
- backend endpoints through port-forward;
- PostgreSQL connectivity and counts;
- NetworkPolicy;
- RBAC;
- `can-i` checks.

### Acceptance criteria

- The script creates an evidence directory.
- The script avoids printing Secret values.
- A healthy baseline can reach `PASS=35 WARN=0 FAIL=0`.
- Historical transient events are distinguishable from current runtime impact.

### Current status

Implemented in `scripts/operability/health-check.sh`.

---

## NFR-OBS-004 — Metrics readiness

### Priority

COULD

### Description

The system may expose Prometheus-compatible metrics in a future phase.

### Candidate metrics

- number of ChangeRequests created;
- number of completed and failed ChangeRequests;
- workflow duration;
- GitLab API errors;
- Argo CD API errors;
- Tekton validation failures;
- evidence collection duration.

### Current status

Future enhancement.

---

# 7. Reliability and resilience requirements

## NFR-REL-001 — Non-ambiguous failure states

### Priority

MUST

### Description

The system must avoid ambiguous states after failures.

### Rules

- Failed Tekton validation maps to `ValidationFailed`.
- Failed deployment checks map to explicit deployment runtime status where possible.
- Failed authorization must not execute actions.
- Failed evidence collection must return a readable error.
- Failures must create events where associated with a ChangeRequest.

### Acceptance criteria

- No workflow remains indefinitely in an intermediate state without diagnostic information.
- Every failure produces a readable error.
- Every relevant failure is auditable.

### Current status

Implemented for the main workflows.

---

## NFR-REL-002 — Controlled retry and idempotency

### Priority

SHOULD

### Description

Critical operations should support controlled retry without creating duplicate or inconsistent external resources.

### Examples

- Argo CD application retrieval;
- Tekton validation status polling;
- Argo CD deployment status checks;
- evidence collection;
- GitLab branch creation;
- GitLab merge request opening.

### Acceptance criteria

- Retry does not create uncontrolled duplicate branches.
- Retry does not create uncontrolled duplicate PipelineRuns.
- Retry behavior is documented where manual operation is required.

### Current status

Partially implemented through explicit technical actions and status checks.

---

## NFR-REL-003 — Explicit timeouts

### Priority

MUST

### Description

External integrations must use explicit timeouts where applicable.

### Dependencies

- GitLab API;
- Argo CD API;
- Kubernetes/OpenShift API;
- PostgreSQL;
- HTTP health checks.

### Acceptance criteria

- External calls do not hang indefinitely.
- Timeouts are configurable where needed.
- Timeout failures produce readable errors.

### Current status

Baseline behavior implemented, with further tuning expected.

---

# 8. Performance and scalability requirements

## NFR-PERF-001 — Baseline API performance

### Priority

SHOULD

### Description

Main APIs should respond in acceptable time for operational use.

### Initial targets

- `GET /livez`: below 100 ms in normal conditions;
- `GET /readyz`: below 500 ms in normal conditions;
- `GET /api/v1/applications`: below 5 seconds in small lab environments;
- `GET /api/v1/changes`: below 2 seconds for hundreds of ChangeRequests;
- ChangeRequest detail: below 2 seconds for normal evidence size.

### Notes

These targets are indicative for MVP and lab usage. They must be revisited for larger environments.

---

## NFR-PERF-002 — Long-running workflows must be asynchronous

### Priority

MUST

### Description

Long-running workflows such as Tekton validation and deployment checks must not block HTTP requests for excessive time.

### Rules

- A request starts or checks a workflow step.
- Runtime status is persisted.
- The client can poll or refresh the UI to observe progress.

### Acceptance criteria

- ChangeRequest creation returns quickly.
- Technical workflow status can be checked later.
- UI actions redirect after execution attempts.

### Current status

Implemented through explicit technical endpoints and UI actions.

---

## NFR-SCL-001 — Incremental scalability

### Priority

SHOULD

### Description

The system must support gradual growth.

### Initial targets

- multiple applications;
- hundreds of ChangeRequests;
- multiple developers;
- a small number of concurrent workflows;
- future multiple target environments.

### Acceptance criteria

- The data model does not assume a single application.
- UI lists remain readable with multiple changes.
- Environment mappings are configurable in future phases.

### Current status

Implemented for multiple ChangeRequests and multi-developer visibility. Multi-environment runtime support is future work.

---

# 9. Maintainability requirements

## NFR-MNT-001 — Modular adapter-based architecture

### Priority

MUST

### Description

The code must separate domain logic from external integrations.

### Expected adapters

- GitLab adapter;
- Argo CD adapter;
- Kubernetes adapter;
- Tekton adapter;
- PostgreSQL repository;
- evidence collector;
- AuthN/AuthZ middleware and resolver.

### Acceptance criteria

- Workflow logic does not directly depend on low-level HTTP details.
- Adapters can be tested separately.
- The composition root wires concrete implementations.

### Current status

Implemented.

---

## NFR-MNT-002 — Documentation must stay aligned

### Priority

MUST

### Description

Documentation must evolve with the project.

### Rules

- New architectural decisions must be documented through ADRs.
- New runbooks must be added for operational procedures.
- Numbered core documentation must be migrated to English.
- The ADR index must remain updated.
- Documentation must follow the repository language policy.

### Acceptance criteria

- Repository language policy exists.
- Italian documentation inventory exists.
- Numbered documentation migration plan exists.
- Core numbered documents are migrated incrementally.

### Current status

In progress. Product foundation documents are being migrated.

---

## NFR-MNT-003 — Consistent naming

### Priority

SHOULD

### Description

The project must use consistent naming across documentation, code, API and database.

### Canonical terms

- ChangeRequest;
- ChangeEvent;
- Evidence;
- Application;
- lifecycle status;
- runtime status;
- target environment;
- requester;
- promotion.

### Acceptance criteria

- Documentation and UI use English terminology.
- Environment names use `dev`, `staging`, `production`.

### Current status

In progress through documentation migration and terminology normalization.

---

# 10. Configurability and portability requirements

## NFR-CFG-001 — External configuration

### Priority

MUST

### Description

The system must read configuration from environment variables, ConfigMaps or mounted files.

### Non-sensitive configuration examples

```text
HTTP_ADDR
LOG_LEVEL
ARGOCD_BASE_URL
GITLAB_BASE_URL
TEKTON_NAMESPACE
KUBERNETES_NAMESPACE
AUTH_ENABLED
AUTH_HEADER_USER
AUTH_HEADER_GROUPS
```

### Sensitive configuration examples

```text
DATABASE_URL
ARGOCD_AUTH_TOKEN
GITLAB_TOKEN
```

### Acceptance criteria

- Missing required configuration fails with readable errors.
- Sensitive values are not printed.
- Runtime ConfigMaps can be inspected safely.

### Current status

Implemented.

---

## NFR-CFG-002 — Environment catalog readiness

### Priority

SHOULD

### Description

The system must evolve toward explicit multi-environment configuration.

### Canonical environments

```text
dev
staging
production
```

### Acceptance criteria

- The environment configuration model is documented.
- Unknown environments will be rejected fail-closed when implemented.
- `production` remains disabled until guardrails are ready.

### Current status

Documented design baseline.

---

# 11. Usability requirements

## NFR-UX-001 — Guided but transparent user experience

### Priority

SHOULD

### Description

The system should provide clear operational messages without hiding technical details.

### Requirements

- Explain major status values.
- Show evidence summaries.
- Show requester information.
- Show runtime status clearly.
- Keep UI labels in English.
- Provide links to detailed evidence and audit events.

### Acceptance criteria

- New users can follow the workflow through UI and documentation.
- Operators can still access technical detail when troubleshooting.

### Current status

Implemented as a UI baseline and evolving.

---

## NFR-UX-002 — Stable workflows before advanced UI expansion

### Priority

SHOULD

### Description

Advanced UI expansion should follow stable workflows rather than drive premature architecture.

### Acceptance criteria

- New UI actions map to existing backend endpoints.
- Backend AuthZ remains authoritative.
- UI rendering does not bypass workflow validation.

### Current status

Implemented for current UI actions.

---

# 12. GitOps compliance requirements

## NFR-GITOPS-001 — No permanent GitOps bypass

### Priority

MUST

### Description

The system must not permanently modify application runtime resources outside Git.

### Rules

- Application changes must go through GitLab and Argo CD.
- Imperative OpenShift commands are troubleshooting tools only.
- Evidence collection must not mutate runtime resources.

### Acceptance criteria

- GitLab technical workflow is the desired-state path.
- Runtime checks and evidence collection are read-only.
- Documentation reinforces Git as source of desired state.

### Current status

Implemented in the current workflow design.

---

## NFR-GITOPS-002 — Drift visibility

### Priority

SHOULD

### Description

The system should make GitOps drift visible and understandable.

### Acceptance criteria

- `OutOfSync` is visible.
- `Degraded` is visible.
- Warnings are preserved.
- Drift is not silently ignored.

### Current status

Partially implemented through Argo CD status and evidence diagnostics.

---

# 13. Backup, restore and disaster recovery requirements

## NFR-BCK-001 — PostgreSQL backup and restore

### Priority

MUST

### Description

PostgreSQL stores ChangeRequests, events and evidence. Backup and restore must be documented and validated.

### Acceptance criteria

- PostgreSQL backup runbook exists.
- Restore validation plan exists.
- Isolated restore test has been executed.
- Backup artifacts are checksummed.

### Current status

Implemented and validated.

---

## NFR-DR-001 — Disaster recovery runbook

### Priority

MUST

### Description

The repository must provide a disaster recovery runbook for the DevOps Control Plane baseline.

### Acceptance criteria

- Protected components are identified.
- RPO/RTO assumptions are documented.
- Restore target isolation is documented.
- Evidence package expectations are documented.

### Current status

Implemented.

---

## NFR-MAINT-001 — Maintenance operations runbook

### Priority

MUST

### Description

The repository must provide a maintenance operations runbook.

### Acceptance criteria

- Application rollout and rollback procedures are documented.
- ConfigMap maintenance is documented.
- Secret/token rotation coordination is documented.
- Post-maintenance validation is documented.

### Current status

Implemented.

---

# 14. OpenShift deployment requirements

## NFR-OCP-001 — Containerized runtime

### Priority

MUST

### Description

The service must run as a container in OpenShift.

### Acceptance criteria

- Container image builds successfully.
- Container exposes the configured HTTP port.
- Container does not require elevated privileges.
- Image can be pushed to the OpenShift registry.

### Current status

Implemented.

---

## NFR-OCP-002 — OpenShift deployment resources

### Priority

MUST

### Description

The repository must provide OpenShift/Kubernetes deployment resources.

### Required resources

- namespace/project assumptions;
- Deployment;
- Service;
- Route;
- ConfigMap;
- Secret template;
- ServiceAccount;
- Role/RoleBinding;
- NetworkPolicy where required.

### Acceptance criteria

- Manifests contain no real secrets.
- Runtime deployment is validated.
- OAuth Proxy sidecar is documented.
- Resource requests and limits are defined for application and proxy containers.

### Current status

Implemented.

---

# 15. Testing requirements

## NFR-TST-001 — Unit tests for domain and service logic

### Priority

MUST

### Description

Domain and service logic must be testable without real GitLab, Argo CD or Kubernetes dependencies.

### Acceptance criteria

- Lifecycle transitions are tested.
- Input validation is tested.
- Service behavior is tested with fake dependencies where applicable.

### Current status

Implemented for core service areas.

---

## NFR-TST-002 — Adapter tests with fake clients

### Priority

SHOULD

### Description

External adapters should be testable with fake or HTTP test clients.

### Acceptance criteria

- GitLab adapter can be tested without a real GitLab server.
- Argo CD adapter can be tested without a real Argo CD server.
- TLS utility behavior is tested.

### Current status

Implemented for several adapters.

---

## NFR-TST-003 — Evidence-aligned manual and runtime tests

### Priority

MUST

### Description

Important milestones must have repeatable validation steps and evidence.

### Acceptance criteria

- Runtime validations produce evidence directories or captured outputs.
- Smoke test produces an evidence directory.
- Restore tests produce backup/restore evidence.

### Current status

Implemented as part of Phase 10 operability baseline.

---

# 16. Documentation requirements

## NFR-DOC-001 — Progressive documentation

### Priority

MUST

### Description

Documentation must grow with the project and remain aligned with the implemented baseline.

### Required documentation areas

- vision;
- scope;
- personas and use cases;
- functional requirements;
- non-functional requirements;
- architecture;
- data model;
- workflows;
- API design;
- ADRs;
- runbooks;
- production readiness;
- final technical documentation.

### Acceptance criteria

- Documents are versioned in Git.
- Documents have purpose and status.
- Documentation language policy is followed.

### Current status

In progress.

---

## NFR-DOC-002 — ADRs for important decisions

### Priority

MUST

### Description

Every significant architectural decision must be documented through an ADR.

### Examples

- Git as source of truth;
- Argo CD as GitOps engine;
- Tekton as validation engine;
- PostgreSQL as change and evidence store;
- AuthN/AuthZ strategy;
- OAuth Proxy deployment design;
- multi-environment model.

### Acceptance criteria

- ADR files are written in English.
- ADR index is updated.
- ADR numbering remains consistent.

### Current status

Implemented and actively maintained.

---

## NFR-DOC-003 — Repository language policy

### Priority

MUST

### Description

Repository documentation must follow the official language policy.

### Rules

- Official repository language is English.
- UI language is English.
- ADR language is English.
- Runbook language is English.
- Existing Italian documentation is migrated incrementally.

### Acceptance criteria

- `docs/documentation-language-policy.md` exists.
- `docs/italian-documentation-inventory.md` exists.
- `docs/numbered-documentation-migration-plan.md` exists.
- New documentation is written in English.

### Current status

Implemented and migration in progress.

---

# 17. Multi-developer and multi-environment requirements

## NFR-MDEV-001 — Multi-developer readability

### Priority

MUST

### Description

The UI must remain readable when multiple users create ChangeRequests.

### Acceptance criteria

- Requester is visible.
- Dashboard recent changes is limited.
- Full history remains available.
- Multi-developer test scenario is documented.

### Current status

Implemented and validated.

---

## NFR-ENV-001 — Multi-environment readiness

### Priority

SHOULD

### Description

The architecture must support future multi-environment workflows.

### Canonical environments

```text
dev
staging
production
```

### Acceptance criteria

- Multi-environment architecture decision is documented.
- Environment configuration model is documented.
- Production remains disabled until guardrails are validated.
- Environment-aware RBAC/AuthZ is part of the design.

### Current status

Documented design baseline.

---

# 18. Compliance matrix

| Area | Key requirement | Priority |
|---|---|---|
| Security | No tokens in Git, logs or evidence | MUST |
| AuthN/AuthZ | OAuth Proxy and backend authorization | MUST |
| Audit | ChangeRequest traceability | MUST |
| Evidence | Persistent sanitized evidence | MUST |
| Observability | Health/readiness endpoints and smoke test | MUST |
| Reliability | Non-ambiguous failure states | MUST |
| Performance | Long workflows are asynchronous | MUST |
| Maintainability | Modular adapter-based architecture | MUST |
| Configurability | External configuration and Secret separation | MUST |
| GitOps | No permanent bypass of GitOps | MUST |
| Backup/Restore | PostgreSQL backup and restore runbook | MUST |
| OpenShift | Containerized non-privileged runtime | MUST |
| Documentation | English documentation and ADR alignment | MUST |
| Multi-developer | Requester and full history visibility | MUST |
| Multi-environment | Configurable future environment model | SHOULD |

---

## 19. Document completion criteria

This document is considered stable when:

- all major non-functional areas have at least one requirement;
- MUST requirements are aligned with architecture and deployment;
- security, AuthN/AuthZ and operability baselines are represented;
- backup, restore, DR and maintenance requirements are represented;
- documentation language requirements are represented;
- multi-developer and multi-environment readiness are represented.

---

## 20. Key message

The DevOps Control Plane must not only be functional. It must be secure, traceable, observable, operable and consistent with GitOps.

The value of the project is not only automating commands. The value is making changes:

```text
guided
repeatable
validated
traceable
auditable
secure
operable
consistent with GitOps
```

---

## 21. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial non-functional requirements document in Italian. |
| 2026-07-03 | 0.2 | Rewritten in English and refreshed to align with the current security, operability, production-readiness, multi-developer and multi-environment baseline. |
