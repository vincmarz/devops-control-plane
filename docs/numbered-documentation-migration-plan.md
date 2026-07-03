# DevOps Control Plane — Numbered Core Documentation Migration Plan

## Document metadata

- **Project:** DevOps Control Plane
- **Phase:** 0.3 — Numbered core documentation migration plan
- **Status:** Migration plan baseline
- **Date:** 2026-07-03
- **Scope:** Numbered core documentation files under `docs/`, currently identified as containing Italian or mixed-language content
- **Policy reference:** `docs/documentation-language-policy.md`
- **Inventory reference:** `docs/italian-documentation-inventory.md`

---

## 1. Purpose

This document defines the migration plan for the numbered core documentation set of the DevOps Control Plane repository.

The repository language policy defines English as the official repository language. The documentation inventory identified that the numbered documentation files from `00` to `13` are still largely written in Italian or contain mixed-language content.

The purpose of this plan is to migrate the numbered documentation set in a controlled way, avoiding a blind literal translation of outdated documents.

The migration approach is:

```text
translate + refresh + normalize terminology
```

not simply:

```text
literal translation
```

---

## 2. Background

The project has evolved significantly since the first numbered documents were created.

The current repository now includes:

- Go backend foundation;
- PostgreSQL persistence;
- ChangeRequest lifecycle and audit events;
- GitLab API integration;
- Tekton validation workflow;
- Argo CD deployment checks;
- Kubernetes/OpenShift runtime evidence;
- server-side UI dashboard and change pages;
- OpenShift deployment;
- OAuth Proxy;
- header-based AuthN/AuthZ;
- OpenShift group lookup;
- TLS strict baseline;
- Secret/token rotation runbook;
- RBAC least-privilege baseline;
- NetworkPolicy baseline;
- PostgreSQL backup/restore validation;
- operability smoke-test script;
- disaster recovery runbook;
- maintenance operations runbook;
- production readiness checklist;
- multi-developer UI validation;
- multi-environment design for `dev`, `staging` and `production`.

For this reason, older documents must be refreshed while being translated.

---

## 3. Official language decision

The official repository language is:

```text
English
```

The numbered documentation migration must use the terminology and rules defined in:

```text
docs/documentation-language-policy.md
```

Canonical environment names:

```text
dev
staging
production
```

Display names:

```text
Development
Staging
Production
```

Do not use Italian canonical environment names such as:

```text
collaudo
produzione
```

---

## 4. Migration principles

The numbered documentation migration must follow these principles:

1. Preserve useful technical meaning.
2. Remove or update obsolete statements.
3. Align terminology with current ADRs and runbooks.
4. Keep filenames stable during the first migration pass to avoid breaking links.
5. Prefer English headings and English body text.
6. Keep commands, resource names and identifiers unchanged unless they are intentionally renamed.
7. Do not translate Kubernetes, OpenShift, GitLab, Tekton, Argo CD or API identifiers.
8. Mark superseded content explicitly instead of silently deleting important history.
9. Keep each commit focused by document or small batch.
10. Run `git diff --check` before every commit.

---

## 5. Numbered documentation scope

The migration initially covers the numbered files under `docs/` from `00` to `13`.

Expected scope:

```text
docs/00-vision.md
docs/01-scope-mvp.md
docs/02-personas-use-cases.md
docs/03-functional-requirements.md
docs/04-non-functional-requirements.md
docs/05-architecture.md
docs/06-argocd-integration.md
docs/07-gitlab-integration.md
docs/08-tekton-integration.md
docs/09-kubernetes-openshift-integration.md
docs/10-*.md
docs/11-*.md
docs/12-*.md
docs/13-api-design.md
```

The exact list should be verified against the repository tree before each migration batch.

---

## 6. File classification model

Each numbered document should be classified using one of the following migration actions.

```text
TRANSLATE_AND_REFRESH
  Translate to English and update content to current project state.

REWRITE
  Replace old content with a new English document because the original is too outdated.

SUPERSEDE
  Keep the file but mark it as superseded by newer documentation.

MERGE_INTO_FINAL_DOCUMENT
  Use the content as input for Phase 12 final technical documentation instead of maintaining it as a standalone detailed document.

REVIEW_ONLY
  Review residual language or terminology issues only.
```

---

## 7. Migration batches

### 7.1 Batch A — Product foundation

Target files:

```text
docs/00-vision.md
docs/01-scope-mvp.md
docs/02-personas-use-cases.md
docs/03-functional-requirements.md
docs/04-non-functional-requirements.md
```

Recommended action:

```text
TRANSLATE_AND_REFRESH
```

Purpose:

- define the product vision in English;
- align MVP scope with the current advanced baseline;
- update personas and use cases;
- update functional requirements;
- update non-functional requirements based on Phase 9 and Phase 10 outcomes.

Priority:

```text
P1
```

---

### 7.2 Batch B — Architecture and integrations

Target files:

```text
docs/05-architecture.md
docs/06-argocd-integration.md
docs/07-gitlab-integration.md
docs/08-tekton-integration.md
docs/09-kubernetes-openshift-integration.md
```

Recommended action:

```text
TRANSLATE_AND_REFRESH
```

Purpose:

- align architecture with Go/PostgreSQL/OpenShift runtime;
- include OAuth Proxy and AuthN/AuthZ baseline;
- include TLS strict, RBAC and NetworkPolicy baseline;
- include evidence and audit architecture;
- include multi-environment direction;
- align integration descriptions with actual adapters and runtime behavior.

Priority:

```text
P1
```

---

### 7.3 Batch C — API, operations and roadmap documents

Target files:

```text
docs/10-*.md
docs/11-*.md
docs/12-*.md
docs/13-api-design.md
```

Recommended action:

```text
TRANSLATE_AND_REFRESH or SUPERSEDE, depending on current relevance
```

Purpose:

- align API design with implemented endpoints;
- align operational documents with Phase 10 runbooks;
- mark obsolete planning documents as superseded if needed;
- prepare inputs for Phase 12 final technical documentation.

Priority:

```text
P2
```

---

## 8. Initial document migration matrix

| File | Priority | Initial action | Notes |
|---|---:|---|---|
| `docs/00-vision.md` | P1 | TRANSLATE_AND_REFRESH | Product vision must reflect the current Control Plane scope. |
| `docs/01-scope-mvp.md` | P1 | TRANSLATE_AND_REFRESH | Scope must be aligned with security, operability and multi-environment direction. |
| `docs/02-personas-use-cases.md` | P1 | TRANSLATE_AND_REFRESH | Personas should include requester, operator, approver, admin and platform operator. |
| `docs/03-functional-requirements.md` | P1 | TRANSLATE_AND_REFRESH | Requirements should include implemented workflows and upcoming multi-environment requirements. |
| `docs/04-non-functional-requirements.md` | P1 | TRANSLATE_AND_REFRESH | Must include security, operability, backup/restore, DR and production readiness. |
| `docs/05-architecture.md` | P1 | TRANSLATE_AND_REFRESH | Must align with current Go/PostgreSQL/OpenShift/OAuth Proxy architecture. |
| `docs/06-argocd-integration.md` | P1 | TRANSLATE_AND_REFRESH | Must align with TLS strict, evidence diagnostics and future environment mapping. |
| `docs/07-gitlab-integration.md` | P1 | TRANSLATE_AND_REFRESH | Must align with branch, file update, MR open and merge flows. |
| `docs/08-tekton-integration.md` | P1 | TRANSLATE_AND_REFRESH | Must align with validation evidence, diagnostics and anti-secret policy checks. |
| `docs/09-kubernetes-openshift-integration.md` | P1 | TRANSLATE_AND_REFRESH | Must align with ServiceAccount token fallback and runtime evidence collection. |
| `docs/10-*.md` | P2 | REVIEW | Decide whether to translate, refresh or supersede based on current content. |
| `docs/11-*.md` | P2 | REVIEW | CLI-related content may remain deferred if Phase 11 is still optional. |
| `docs/12-*.md` | P2 | MERGE_INTO_FINAL_DOCUMENT | Use as input for Phase 12 if applicable. |
| `docs/13-api-design.md` | P1 | TRANSLATE_AND_REFRESH | Must align with current API and AuthZ model. |

---

## 9. Translation and refresh rules

When migrating a document:

1. Translate headings to English.
2. Translate body text to English.
3. Replace Italian environment terms with `dev`, `staging`, `production`.
4. Replace Italian process terms with the official English terms.
5. Preserve commands and API paths.
6. Preserve Kubernetes/OpenShift/GitLab/Tekton/Argo CD identifiers.
7. Update outdated statements to reflect current implemented state.
8. Add references to newer ADRs or runbooks where useful.
9. Do not remove historical decisions without noting supersession.
10. Run quality checks before commit.

Recommended quality checks:

```bash
git diff --check

grep -n "collaudo\|produzione\|Questo\|Questa\|Obiettivo\|Regola" docs/<file>.md
```

The grep check is a heuristic and may return false positives.

---

## 10. Superseded document handling

If a numbered document is no longer accurate and should not be maintained as-is, use a supersession block at the top:

```text
> Status: Superseded
> Superseded by: <new document path>
> Reason: <short explanation>
```

Avoid deleting documents during the first migration pass unless there is a dedicated cleanup decision.

---

## 11. Link preservation strategy

During the first migration pass, keep existing filenames stable.

Do not rename:

```text
00-vision.md
01-scope-mvp.md
02-personas-use-cases.md
...
13-api-design.md
```

Rationale:

- avoids breaking existing links;
- keeps history easy to follow;
- separates language migration from file taxonomy cleanup.

A future consistency pass may propose renaming if needed.

---

## 12. Recommended execution order

Recommended execution phases:

```text
Phase 0.4 — Translate and refresh product foundation docs 00-04
Phase 0.5 — Translate and refresh architecture and integration docs 05-09
Phase 0.6 — Translate and refresh API / operational / roadmap docs 10-13
Phase 0.7 — Documentation consistency pass
```

Recommended commit granularity:

```text
one document per commit for large files
small coherent batches for short files
```

---

## 13. Acceptance criteria for Phase 0.3

Phase 0.3 is complete when:

```text
numbered documentation migration scope is documented
migration batches are defined
initial migration matrix is documented
translation and refresh rules are documented
superseded document handling is documented
execution order is documented
repository language policy is referenced
```

No translation of numbered documents is required to close Phase 0.3.

---

## 14. Next recommended phase

```text
Phase 0.4 — Translate and refresh product foundation docs 00-04
```

Recommended first document:

```text
docs/00-vision.md
```

---

## 15. Revision history

| Date | Phase | Description |
|---|---:|---|
| 2026-07-03 | 0.3 | Initial numbered core documentation migration plan. |
