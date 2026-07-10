# DevOps Control Plane — Documentation Index

This is the entry point for all project documentation. Files are grouped by
purpose: living reference on top, point-in-time validation evidence at the
bottom.

> Language policy: all documentation is English (see
> [documentation-language-policy.md](documentation-language-policy.md)).
> Existing Italian documents are tracked as migration targets.

---

## Start here (suggested reading order)

New to the project? Read in this order:

1. [00-vision.md](00-vision.md) — what the project is and why it exists.
2. [01-scope-mvp.md](01-scope-mvp.md) — what is in and out of scope.
3. [05-architecture.md](05-architecture.md) — how the system is put together.
4. [13-api-design.md](13-api-design.md) — the API surface.
5. [adr/README.md](adr/README.md) — the decisions behind the design.

For an end-to-end narrative, see the
[Final technical guide](final-technical-guide/final-technical-guide-it.md) (Italian).

---

## 1. Core specifications

The numbered `00`–`13` set is the canonical product and technical specification.

| Doc | Topic |
|---|---|
| [00-vision.md](00-vision.md) | Product vision, guiding principles, scope direction |
| [01-scope-mvp.md](01-scope-mvp.md) | MVP scope: in scope vs out of scope |
| [02-personas-use-cases.md](02-personas-use-cases.md) | Personas and use cases |
| [03-functional-requirements.md](03-functional-requirements.md) | Functional requirements |
| [04-non-functional-requirements.md](04-non-functional-requirements.md) | Non-functional requirements |
| [05-architecture.md](05-architecture.md) | System architecture and component model |
| [06-argocd-integration.md](06-argocd-integration.md) | Argo CD integration design |
| [07-gitlab-integration.md](07-gitlab-integration.md) | GitLab integration design |
| [08-tekton-integration.md](08-tekton-integration.md) | Tekton validation integration design |
| [09-security-rbac.md](09-security-rbac.md) | Security model and RBAC |
| [10-data-model.md](10-data-model.md) | Data model |
| [11-change-workflows.md](11-change-workflows.md) | ChangeRequest workflows |
| [12-evidence-model.md](12-evidence-model.md) | Evidence model |
| [13-api-design.md](13-api-design.md) | HTTP API design |

## 2. Architecture Decision Records

Full index and rules: [adr/README.md](adr/README.md).

| ADR | Decision |
|---|---|
| [ADR-0001](adr/ADR-0001-git-source-of-truth.md) | Git as source of truth |
| [ADR-0002](adr/ADR-0002-argocd-as-gitops-engine.md) | Argo CD as GitOps engine |
| [ADR-0003](adr/ADR-0003-tekton-validation-engine.md) | Tekton as validation engine |
| [ADR-0004](adr/ADR-0004-postgresql-change-history.md) | PostgreSQL as change history and evidence store |
| [ADR-0005](adr/ADR-0005-api-first-before-web-ui.md) | API-first before full Web UI |
| [ADR-0006](adr/ADR-0006-adapter-based-architecture.md) | Adapter-based architecture |
| [ADR-0007](adr/ADR-0007-gitlab-api-as-git-provider.md) | GitLab API as Git provider integration |
| [ADR-0008](adr/ADR-0008-kubernetes-api-for-tekton.md) | Kubernetes API for Tekton integration |
| [ADR-0009](adr/ADR-0009-authn-authz-strategy.md) | AuthN/AuthZ strategy |
| [ADR-0010](adr/ADR-0010-oauth-proxy-deployment-design.md) | OAuth Proxy deployment design |
| [ADR-0011](adr/ADR-0011-multi-environment-model.md) | Multi-environment model |

## 3. Governance and documentation policy

| Doc | Purpose |
|---|---|
| [documentation-language-policy.md](documentation-language-policy.md) | Official repository language policy (English) |
| [numbered-documentation-migration-plan.md](numbered-documentation-migration-plan.md) | Plan to migrate the `00`–`13` set |
| [italian-documentation-inventory.md](italian-documentation-inventory.md) | Inventory of Italian content to migrate |
| [production-readiness-checklist.md](production-readiness-checklist.md) | Production readiness checklist |

## 4. Design and configuration models

Living design documents for runtime, environments and multi-cluster support.

| Doc | Topic |
|---|---|
| [multi-environment-model.md](multi-environment-model.md) | Multi-environment model |
| [environment-configuration-model.md](environment-configuration-model.md) | Environment configuration model |
| [runtime-client-secret-config-model.md](runtime-client-secret-config-model.md) | Runtime client secret configuration model |
| [runtime-client-secret-reference-loading.md](runtime-client-secret-reference-loading.md) | Runtime client secret reference loading |
| [controlled-runtime-secret-loader-and-factory-enablement-design.md](controlled-runtime-secret-loader-and-factory-enablement-design.md) | Secret loader and factory enablement design |
| [real-runtime-client-factory-implementation-design.md](real-runtime-client-factory-implementation-design.md) | Runtime client factory implementation design |
| [real-runtime-client-factory-implementation-readiness.md](real-runtime-client-factory-implementation-readiness.md) | Runtime client factory readiness |
| [single-non-production-multi-cluster-enablement-plan.md](single-non-production-multi-cluster-enablement-plan.md) | Single non-production multi-cluster enablement plan |
| [multi-cluster-environment-enablement-request.md](multi-cluster-environment-enablement-request.md) | Multi-cluster environment enablement request |
| [postgresql-integration-notes.md](postgresql-integration-notes.md) | PostgreSQL integration notes |
| [phase-1-postgresql-change-repository.md](phase-1-postgresql-change-repository.md) | Phase 1 — PostgreSQL change repository |
| [phase-10-operability-closure.md](phase-10-operability-closure.md) | Phase 10 — operability closure |
| [skeleton-notes.md](skeleton-notes.md) | Initial skeleton notes |

## 5. Runbooks (operations)

Operational procedures under [runbooks/](runbooks/).

| Runbook | Purpose |
|---|---|
| [operability-health-check.md](runbooks/operability-health-check.md) | Operability health-check procedure |
| [postgresql-backup-restore.md](runbooks/postgresql-backup-restore.md) | PostgreSQL backup and restore |
| [disaster-recovery.md](runbooks/disaster-recovery.md) | Disaster recovery |
| [maintenance-operations.md](runbooks/maintenance-operations.md) | Maintenance operations |
| [secrets-rotation.md](runbooks/secrets-rotation.md) | Secret and token rotation |
| [authn-authz.md](runbooks/authn-authz.md) | AuthN/AuthZ operations |
| [oauth-proxy.md](runbooks/oauth-proxy.md) | OAuth Proxy operations |
| [oauth-proxy-implementation.md](runbooks/oauth-proxy-implementation.md) | OAuth Proxy implementation |
| [oauth-proxy-live-patch-rollout.md](runbooks/oauth-proxy-live-patch-rollout.md) | OAuth Proxy live patch rollout |
| [oauth-proxy-runtime-only-retry.md](runbooks/oauth-proxy-runtime-only-retry.md) | OAuth Proxy runtime-only retry |

## 6. Final technical guide

The synthesized end-to-end guide under [final-technical-guide/](final-technical-guide/).

| Doc | Purpose |
|---|---|
| [final-technical-guide-it.md](final-technical-guide/final-technical-guide-it.md) | Full technical guide (Italian localized deliverable) |
| [README.md](final-technical-guide/README.md) | Guide overview |
| [outline.md](final-technical-guide/outline.md) | Guide outline |
| [source-map.md](final-technical-guide/source-map.md) | Source mapping |
| [writing-plan.md](final-technical-guide/writing-plan.md) | Writing plan |

> **Localized deliverable**: `final-technical-guide-it.md` is a declared Italian
> deliverable per documentation-language-policy.md §5.2. Commands, resource names,
> API names and technical identifiers remain in their original format.

---

## 7. Validation and results reports (point-in-time evidence)

These are **historical evidence logs** tied to specific runtime commits and
images. They are kept for audit and traceability and are **not** living
reference — expect them to reflect the state at the time they were written.

<!-- Consider moving these under docs/validation/ to separate point-in-time
     evidence from living reference documentation. -->

| Report |
|---|
| [multi-user-validation-results.md](multi-user-validation-results.md) |
| [multi-user-multi-environment-validation-results.md](multi-user-multi-environment-validation-results.md) |
| [multi-user-multi-environment-validation-matrix.md](multi-user-multi-environment-validation-matrix.md) |
| [role-aware-ui-action-visibility-results.md](role-aware-ui-action-visibility-results.md) |
| [environment-catalog-ui-action-results.md](environment-catalog-ui-action-results.md) |
| [environment-visibility-ui-results.md](environment-visibility-ui-results.md) |
| [environment-cluster-resolver-results.md](environment-cluster-resolver-results.md) |
| [cluster-registry-baseline-results.md](cluster-registry-baseline-results.md) |
| [lifecycle-runtime-status-clarity-results.md](lifecycle-runtime-status-clarity-results.md) |
| [target-environment-validation-results.md](target-environment-validation-results.md) |
| [technical-runtime-target-preparation-results.md](technical-runtime-target-preparation-results.md) |
| [argocd-runtime-provider-wiring-results.md](argocd-runtime-provider-wiring-results.md) |
| [tekton-runtime-provider-wiring-results.md](tekton-runtime-provider-wiring-results.md) |
| [kubernetes-runtime-provider-wiring-results.md](kubernetes-runtime-provider-wiring-results.md) |
| [kubernetes-secret-loader-wiring-results.md](kubernetes-secret-loader-wiring-results.md) |
| [factory-aware-runtime-provider-main-wiring-results.md](factory-aware-runtime-provider-main-wiring-results.md) |
| [controlled-enablement-plumbing-validation-results.md](controlled-enablement-plumbing-validation-results.md) |
| [runtime-non-regression-factories-disabled-results.md](runtime-non-regression-factories-disabled-results.md) |
| [simulated-staging-enablement-results.md](simulated-staging-enablement-results.md) |
| [simulated-production-enablement-results.md](simulated-production-enablement-results.md) |
| [simulated-multi-environment-validation-results.md](simulated-multi-environment-validation-results.md) |
| [runtime-evidence-dashboard-maintenance-alignment.md](runtime-evidence-dashboard-maintenance-alignment.md) |

---

## Maintenance

When adding a document, add it to the correct section above. When adding an
ADR, also update [adr/README.md](adr/README.md). Keep filenames in lowercase
kebab-case and all new content in English.
