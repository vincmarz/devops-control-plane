# DevOps Control Plane — Phase 10 Operability Closure

## Document metadata

- **Project:** DevOps Control Plane
- **Phase:** 10 — Operability / production operations
- **Closure status:** Completed as an advanced operational baseline
- **Date:** 2026-07-03
- **Scope:** Formal closure of Phase 10 after completion of PostgreSQL operability, backup/restore, observability, smoke testing, disaster recovery, maintenance operations and production readiness documentation
- **Audience:** Project maintainers, platform engineers, DevOps engineers, service owner, operations team and final technical documentation readers

---

## 1. Executive summary

Phase 10 — Operability / production operations is formally closed as:

```text
COMPLETED AS AN ADVANCED OPERATIONAL BASELINE
```

The DevOps Control Plane now has a consolidated production-oriented operational layer including:

- PostgreSQL operability inventory;
- PostgreSQL backup/restore baseline;
- isolated PostgreSQL restore validation;
- observability baseline inventory;
- detailed operability health-check runbook;
- automated operability smoke-test script;
- disaster recovery operational baseline;
- maintenance operations runbook;
- production readiness checklist.

This phase does not claim that every possible enterprise production-hardening item is fully implemented. Instead, it establishes a strong and repeatable operational baseline suitable for an advanced MVP / production-oriented readiness review, with known limitations explicitly documented.

---

## 2. Phase 10 closure decision

Closure decision:

```text
Phase 10 — Operability / production operations: CLOSED
Closure level: Completed as advanced operational baseline
```

Rationale:

```text
All planned Phase 10 operational deliverables have been created, validated where applicable, committed and pushed to main/origin main.
Runtime smoke tests passed with PASS=35, WARN=0, FAIL=0.
Operational runbooks and scripts are now available in the repository.
Known production gaps are documented in the production readiness checklist.
```

---

## 3. Completed sub-phases

### 3.1 Phase 10.1 — PostgreSQL operability inventory

Status:

```text
COMPLETED
```

Summary:

- PostgreSQL runtime inventoried in read-only mode.
- Pod, Deployment, Service, endpoint and PVC validated.
- PostgreSQL pod was Running and Ready.
- PVC `postgresql-data` was Bound on StorageClass `ocp-rbd-rook`.
- PostgreSQL resources, probes and Secret usage were inspected without printing sensitive values.
- Initial production gaps were identified, including single replica, `latest` image and need for backup/restore procedures.

---

### 3.2 Phase 10.2 — PostgreSQL backup/restore baseline

Status:

```text
COMPLETED
```

Summary:

- PostgreSQL backup was created with `pg_dump`.
- SHA256 checksum was generated and validated.
- PostgreSQL tooling compatibility was handled using PostgreSQL 16.14 tooling inside the pod.
- TOC list was generated.
- Restore validation plan was created.
- Isolated restore test was executed in temporary namespace `devops-control-plane-restore-test`.
- Validation queries succeeded.
- Temporary namespace was cleaned up.
- Repository was aligned with PostgreSQL backup/restore runbook.

Main deliverable:

```text
docs/runbooks/postgresql-backup-restore.md
```

---

### 3.3 Phase 10.3 — Observability / health-check baseline

Status:

```text
COMPLETED
```

This phase was completed through four sub-steps.

#### 10.3.1 — Observability baseline inventory read-only

Status:

```text
COMPLETED
```

Validated runtime baseline:

```text
Namespace devops-control-plane: Active
DCP pod: Running 2/2, restart 0
PostgreSQL pod: Running 1/1, restart 0
Route /readyz: HTTP 200
Route /livez: HTTP 200
Route anonymous /api/v1/changes: HTTP 403
Route anonymous /ui/dashboard: HTTP 403
Backend direct /readyz: HTTP 200
Backend direct /livez: HTTP 200
Backend direct anonymous API: HTTP 401
Backend direct admin-header API: HTTP 200
PostgreSQL connectivity: OK
PostgreSQL counts: change_requests=1, change_events=34, evidences=18
NetworkPolicy: present
RBAC/group-reader: present
```

Evidence directory:

```text
/tmp/dcp-10.3-observability-baseline
```

#### 10.3.2 — Operability health-check runbook

Status:

```text
COMPLETED
```

Commit:

```text
18d3832 Add operability health check runbook
```

Deliverable:

```text
docs/runbooks/operability-health-check.md
```

#### 10.3.3 — Operability smoke-test script

Status:

```text
COMPLETED
```

Commit:

```text
daabedd Add operability smoke test script
```

Deliverables:

```text
scripts/operability/README.md
scripts/operability/health-check.sh
```

Runtime validation:

```text
SKIP_CAN_I=false ./scripts/operability/health-check.sh
PASS=35
WARN=0
FAIL=0
```

Evidence directory:

```text
/tmp/dcp-operability-smoke-test-20260703-102101
```

#### 10.3.4 — Update runbook with smoke-test script reference

Status:

```text
COMPLETED
```

Commit:

```text
1a0c24e Reference operability smoke test in runbook
```

Summary:

- The operability health-check runbook now references the automated smoke-test script.
- Default execution and `SKIP_CAN_I=false` execution are documented.
- Expected baseline `PASS=35 WARN=0 FAIL=0` is documented.
- Exit codes and evidence directory are documented.

---

### 3.4 Phase 10.4 — Probes/health/resource tuning oauth-proxy

Status:

```text
COMPLETED
```

Commit:

```text
c0f5ac3 Set oauth proxy resource requests and limits
```

Summary:

- `oauth-proxy` resource requests and limits were added.
- Runtime validation succeeded.
- Route health and anonymous access policy remained correct.
- Repository was aligned.

Current baseline:

```text
oauth-proxy requests: cpu=10m, memory=32Mi
oauth-proxy limits:   cpu=100m, memory=128Mi
```

---

### 3.5 Phase 10.5 — Disaster recovery operational baseline

Status:

```text
COMPLETED
```

Commit:

```text
3470abc Add disaster recovery runbook
```

Deliverable:

```text
docs/runbooks/disaster-recovery.md
```

Summary:

- Protected components documented.
- RPO/RTO baseline documented.
- Covered and non-covered DR scenarios documented.
- Backup artifacts and checksum validation documented.
- Isolated restore target documented.
- Restore validation queries documented.
- Incident recovery decision flow documented.
- Evidence package, roles, limitations and production recommendations documented.

---

### 3.6 Phase 10.6 — Maintenance operations runbook

Status:

```text
COMPLETED
```

Commit:

```text
f564e79 Add maintenance operations runbook
```

Deliverable:

```text
docs/runbooks/maintenance-operations.md
```

Summary:

- Pre-maintenance checklist documented.
- Application rollout and rollback documented.
- ConfigMap maintenance documented.
- Secret/token rotation coordination documented.
- Trust bundle maintenance documented.
- OAuth Proxy maintenance documented.
- AuthN/AuthZ, RBAC and NetworkPolicy validations documented.
- PostgreSQL maintenance documented.
- Argo CD, GitLab and Tekton maintenance considerations documented.
- Post-maintenance validation through `scripts/operability/health-check.sh` documented.
- Closure and escalation criteria documented.

---

### 3.7 Phase 10.7 — Production readiness checklist

Status:

```text
COMPLETED
```

Commit:

```text
6e70de9 Add production readiness checklist
```

Deliverable:

```text
docs/production-readiness-checklist.md
```

Summary:

- Readiness classification documented.
- Mandatory smoke-test baseline documented.
- Availability, health endpoints, OAuth Proxy, AuthN/AuthZ, TLS, Secret/token, RBAC, NetworkPolicy and PostgreSQL readiness documented.
- Backup/restore, DR, maintenance, observability and integration readiness documented.
- UI and repository/documentation readiness documented.
- Known limitations and accepted risks documented.
- Go/no-go criteria documented.
- Remaining production-hardening items documented.

---

## 4. Repository state at closure

Latest Phase 10 closure-related commit before this document:

```text
6e70de9 Add production readiness checklist
```

Expected repository state after adding this closure document:

```text
main and origin/main aligned
working tree clean
```

Recommended verification commands:

```bash
git status --short
git log --oneline -10
```

---

## 5. Current technical posture

The DevOps Control Plane can now be described as:

```text
Advanced MVP / production-oriented operational baseline
```

Strong points:

```text
Backend foundation complete
Change lifecycle and audit complete
GitLab MR workflow complete
Tekton technical integration complete
Tekton GitOps validation strengthened
Argo CD check and runtime/evidence strengthened
UI MVP advanced
OpenShift deployment working
OAuth Proxy enabled
Auth middleware enabled
OpenShift group lookup enabled
TLS strict baseline implemented
Static Kubernetes token removed
RBAC least-privilege baseline implemented
PostgreSQL NetworkPolicy baseline implemented
PostgreSQL backup/restore validated
Isolated restore tested
Operability runbook available
Automated smoke-test script available and validated
Disaster recovery runbook available
Maintenance operations runbook available
Production readiness checklist available
```

---

## 6. Known limitations after Phase 10

The following items remain known production-hardening topics and are intentionally not treated as blockers for closing Phase 10 as an advanced operational baseline.

```text
PostgreSQL currently single-replica
Some image references still use latest
Scheduled backup automation not yet implemented in repository baseline
Backup retention policy not yet formalized
Off-cluster backup storage not yet formalized
Full OpenShift cluster DR is out of scope for this runbook set
GitLab, Argo CD and Tekton DR are out of scope for this specific baseline
Strict deny-all egress NetworkPolicy design is deferred
Fase 0 ADR/documentation still requires refresh
Fase 11 CLI remains optional and deferred
Fase 12 final technical document remains in incremental production
```

These limitations are tracked in:

```text
docs/production-readiness-checklist.md
```

---

## 7. Formal roadmap update

Updated roadmap status:

```text
Fase 0  — Documentazione e ADR                         PARZIALE / DA AGGIORNARE
Fase 1  — Backend foundation Go/PostgreSQL              COMPLETATA
Fase 2  — Change lifecycle + audit                      COMPLETATA
Fase 3  — GitLab MR workflow                            COMPLETATA
Fase 4  — Tekton technical integration                  COMPLETATA
Fase 5  — Tekton GitOps validation pipeline             COMPLETATA / RAFFORZATA
Fase 6  — Argo CD check + runtime/evidence              COMPLETATA / RAFFORZATA
Fase 7  — UI MVP avanzata                               COMPLETATA
Fase 8  — Deploy DevOps Control Plane su OpenShift      COMPLETATA MVP FUNZIONANTE
Fase 9  — Security hardening / production readiness     COMPLETATA COME BASELINE AVANZATA
Fase 10 — Operability / production operations           COMPLETATA COME BASELINE OPERATIVA AVANZATA
Fase 11 — CLI devopsctl                                 OPZIONALE / DA RIPIANIFICARE
Fase 12 — Documento tecnico finale                      IN PRODUZIONE INCREMENTALE
```

---

## 8. Recommended next step

After closing Phase 10, the recommended next step is:

```text
Fase 12 — Documento tecnico finale
```

Before or during Phase 12, Phase 0 should also be refreshed:

```text
Fase 0 — Documentazione e ADR: update required
```

Recommended documentation consolidation topics:

```text
Architecture overview
ADR index refresh
Security model
Operational model
Backup/restore and DR model
Maintenance model
Production readiness summary
Known limitations and future roadmap
Newbie-friendly end-to-end explanation
```

---

## 9. Closure statement

Phase 10 is formally closed with the following statement:

```text
The DevOps Control Plane Phase 10 — Operability / production operations is completed as an advanced operational baseline. The project now includes validated health-check automation, operational runbooks, disaster recovery baseline, maintenance procedures and production readiness checklist. Remaining production-hardening items are documented and can be managed through future roadmap actions, service-owner acceptance or Phase 12 final documentation.
```

---

## 10. Revision history

| Date | Phase | Description |
|---|---:|---|
| 2026-07-03 | 10 closure | Formal closure document for Phase 10 Operability / production operations. |

## Post-Phase 15 operability consolidation addendum

Status: Completed  
Phase reference: 10.12.1  
Last updated: 2026-07-09

### Purpose

This addendum refreshes the Phase 10 operability closure after completion of Phase 15 and the subsequent post-closure multi-cluster readiness reinforcement.

The original Phase 10 closure remains valid as an advanced operational baseline. This addendum clarifies that the operational documentation has now been realigned with the current runtime state of the DevOps Control Plane.

### Current operational baseline

The current validated runtime baseline is namespace-isolated on the available `ocp-dev` OpenShift cluster:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

This namespace-isolated baseline has been validated through runtime smoke tests and published through the annotated tag:

`namespace-isolated-baseline-20260709`

### Phase 15 alignment

Phase 15 is completed as a multi-cluster code-ready baseline.

Physical cross-cluster runtime validation remains deferred because no additional OpenShift cluster is currently available.

The following capabilities are now reflected in the operational baseline:

- Environment Catalog and namespace-aware runtime model;
- Argo CD Applications for dev, staging and production;
- Tekton validation evidence for staging and production;
- UI visibility for `Environments / Namespaces`;
- runtime evidence and validation evidence on ChangeRequest detail pages;
- simulated staging and production external-cluster readiness tests;
- fail-closed behavior for missing or disabled runtime providers;
- Secret reference, RBAC and runtime factory operational guardrails.

### Updated runbook alignment

The operability health-check runbook has been refreshed after Phase 15.

The refreshed runbook now covers:

- DevOps Control Plane pod readiness;
- `/readyz` validation;
- dashboard UI validation;
- Argo CD Application matrix for dev, staging and production;
- deployment readiness matrix across the three namespaces;
- route `/healthz` matrix;
- Tekton validation PipelineRun checks;
- UI ChangeRequest detail checks;
- incident triage quick-reference;
- evidence checklist;
- Secret, RBAC and runtime factory guardrails.

Updated runbook:

`docs/runbooks/operability-health-check.md`

### Troubleshooting alignment

The troubleshooting matrix now covers the current post-Phase 15 failure modes, including:

- DevOps Control Plane pod not ready;
- `/readyz` non-200;
- dashboard unavailable;
- missing latest ChangeRequest;
- missing `Environments / Namespaces` UI section;
- Argo CD OutOfSync or Degraded;
- deployment not ready in one namespace;
- route `/healthz` failure;
- Tekton PipelineRun failure;
- missing Tekton validation evidence in UI;
- wrong environment or namespace mapping;
- missing or disabled runtime provider;
- Secret loading fail-closed behavior;
- runtime factory fail-closed behavior.

### Guardrail alignment

The operational guardrails now explicitly state that operators must not bypass safety mechanisms during troubleshooting.

The following behavior is expected and must be preserved:

- Secret loading disabled by default;
- runtime factories disabled by default;
- missing provider fails closed;
- disabled provider fails closed;
- non allow-listed Secret reference fails closed;
- unsupported kubeconfig input fails closed;
- unsupported raw CA input fails closed;
- missing token value fails closed;
- no raw Secret value in logs, evidence, documentation or tickets;
- no silent fallback to `ocp-dev` when a different target cluster is expected.

### Current Phase 10 status

Phase 10 can now be considered:

`COMPLETED AS AN ADVANCED OPERABILITY BASELINE, REALIGNED AFTER PHASE 15`

This does not mean that every enterprise production-hardening item is complete.

It means that the project now has an operationally usable and documented baseline aligned with the current runtime model, including namespace-isolated multi-environment operations and multi-cluster code-readiness guardrails.

### Remaining production-hardening topics

The following items remain future production-hardening topics:

- physical multi-cluster runtime validation when additional clusters are available;
- production-grade PostgreSQL high availability, if required by the target service profile;
- immutable image or digest pinning strategy;
- formal SLO and SLA definitions;
- dedicated DR procedures for GitLab, Argo CD and Tekton;
- production-grade alerting and metrics integration;
- periodic restore exercises with measured RPO and RTO;
- final service-owner acceptance.

### Closure statement

Phase 10 is consolidated as an advanced operability baseline aligned with the current DevOps Control Plane runtime.

The platform can be operated, checked and troubleshot using the refreshed runbooks while future physical multi-cluster onboarding remains deferred by infrastructure availability.
