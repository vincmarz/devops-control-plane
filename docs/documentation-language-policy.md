# DevOps Control Plane — Repository Language Policy

## Document metadata

- **Project:** DevOps Control Plane
- **Phase:** 0.1 — Repository language policy
- **Status:** Accepted policy baseline for future documentation alignment
- **Date:** 2026-07-03
- **Scope:** Repository-wide language policy for documentation, ADRs, runbooks, UI labels, operational procedures and terminology
- **Audience:** Project maintainers, contributors, platform engineers, DevOps engineers, technical writers and onboarding readers

---

## 1. Purpose

This document defines the official language policy for the DevOps Control Plane repository.

The policy is needed because the repository currently contains a mix of English and Italian documentation. Recent product UI, ADRs, runbooks and operational documents are already mostly in English, while some early documents, such as `docs/00-vision.md`, still contain Italian content.

The objective is to make the repository consistent, maintainable and easier to read for a broader technical audience.

---

## 2. Decision

The official language of the DevOps Control Plane repository is:

```text
English
```

All new repository content must be written in English unless explicitly documented as an exception.

This applies to:

```text
Architecture documents
Architecture Decision Records
Runbooks
Operational procedures
README files
UI labels and user-facing text
API documentation
Configuration documentation
Manifest comments intended as project documentation
Script README files
Final technical documentation
```

---

## 3. Rationale

English is selected as the official repository language because:

- the UI is already in English;
- recent ADRs and operational runbooks are already in English;
- most cloud-native, Kubernetes, OpenShift, GitOps, Tekton, Argo CD and GitLab terminology is naturally expressed in English;
- repository content becomes easier to share with external teams, vendors and reviewers;
- onboarding material becomes more reusable outside a local Italian-only context;
- the final technical document can use one consistent terminology baseline;
- environment names such as `dev`, `staging` and `production` are standard and widely understood.

---

## 4. Scope

This policy applies to content committed to the repository.

In scope:

```text
docs
README files
ADR files
runbooks
scripts documentation
manifest documentation
pipeline documentation
UI labels
error messages intended for users or operators
configuration model documentation
final technical document
```

Out of scope:

```text
Local temporary notes not committed to the repository
Interactive working chat
Private scratch files outside the repository
External customer-specific documents if explicitly required in another language
```

---

## 5. Allowed exceptions

The following exceptions are allowed:

### 5.1 Historical content during migration

Existing Italian documents may remain temporarily in the repository while the documentation migration is in progress.

They must be tracked as migration targets and should not be expanded with new Italian content.

### 5.2 Customer-facing deliverables

If a specific stakeholder requires an Italian document, that document should be created as a separate deliverable and clearly marked as localized content.

Recommended naming pattern:

```text
*-it.md
```

Example:

```text
final-technical-summary-it.md
```

### 5.3 Legal, contractual or organizational names

Names of organizations, internal services, namespaces, groups or systems should not be translated if the original name is the canonical identifier.

### 5.4 Commit messages

Commit messages should preferably be in English. Historical commit messages in other languages are not rewritten.

---

## 6. Required terminology standard

The following terminology must be used consistently.

### 6.1 Environment names

Use canonical environment keys:

```text
dev
staging
production
```

Use display names:

```text
Development
Staging
Production
```

Do not use the following Italian terms in repository documentation as canonical environment names:

```text
collaudo
produzione
```

If historical references exist, migrate them to:

```text
collaudo   -> staging
produzione -> production
```

### 6.2 Core project terms

Use:

```text
ChangeRequest
change
promotion
environment
evidence
audit event
maintenance
recovery
restore
readiness
rollback
approval
runtime status
lifecycle status
```

Avoid mixed-language expressions such as:

```text
promozione ChangeRequest
ambiente production
evidenza deployment
manutenzione operations
```

### 6.3 UI and user-facing labels

UI labels must be English.

Examples:

```text
Dashboard
Applications
Recent changes
Change Requests
Requested by
Environment
Lifecycle status
Runtime status
Collected evidence
View all
Open
Settings
```

---

## 7. Documentation style rules

Repository documentation should follow these rules:

1. Use English for all headings and body text.
2. Use consistent technical terms across all documents.
3. Prefer simple and explicit wording suitable for onboarding readers.
4. Define acronyms at first use where useful.
5. Use code blocks for commands and configuration examples.
6. Avoid mixing Italian and English in the same paragraph.
7. Keep ADR titles and filenames in English.
8. Keep runbook names in English.
9. Keep file names in lowercase kebab-case when possible.
10. Update indexes when adding documents, especially ADR indexes.

---

## 8. ADR language rule

All ADRs must be written in English.

ADR filenames must follow this pattern:

```text
ADR-NNNN-short-english-slug.md
```

Examples:

```text
ADR-0009-authn-authz-strategy.md
ADR-0010-oauth-proxy-deployment-design.md
ADR-0011-multi-environment-model.md
```

The ADR index must be updated when an ADR is added, renamed or removed:

```text
docs/adr/README.md
```

---

## 9. Runbook language rule

All runbooks must be written in English.

Runbook filenames must use English names, for example:

```text
operability-health-check.md
postgresql-backup-restore.md
disaster-recovery.md
maintenance-operations.md
secrets-rotation.md
```

Runbooks must keep commands and resource names exactly as they are used in the runtime.

---

## 10. UI language rule

The DevOps Control Plane UI must remain English.

This includes:

```text
navigation labels
page titles
button labels
table headers
KPI labels
status descriptions
error messages intended for users
```

If localization is needed in the future, it must be implemented as an explicit product feature and not through mixed-language templates.

---

## 11. Migration strategy for existing Italian documents

Existing Italian documents should be migrated incrementally.

Recommended priority:

```text
1. docs/adr/README.md
2. docs/00-vision.md
3. docs/01-scope-mvp.md
4. docs/05-architecture.md
5. Core roadmap and API design documents
6. Remaining early design documents
```

Migration should preserve technical meaning. Do not remove useful historical context.

When migrating a document:

```text
translate headings and body text to English
normalize terminology
preserve commands and code blocks
preserve architecture decisions
update links if filenames change
run git diff --check
commit with a focused message
```

---

## 12. Inventory approach

To identify Italian content, maintainers can use targeted searches.

Example:

```bash
grep -R -n "Questo\|Questa\|Obiettivo\|Regola\|ambiente\|produzione\|collaudo\|manutenzione\|ripristino\|evidenza" docs
```

Recommended evidence file for migration planning:

```bash
grep -R -n "Questo\|Questa\|Obiettivo\|Regola\|ambiente\|produzione\|collaudo\|manutenzione\|ripristino\|evidenza" docs \
  > /tmp/dcp-docs-italian-terms-inventory.txt
```

The inventory should be used as a guide, not as an automatic translation source.

---

## 13. Repository language compliance checklist

Before adding or modifying documentation, verify:

```text
[ ] New content is written in English
[ ] UI labels are English
[ ] ADR title and filename are English
[ ] Runbook title and filename are English
[ ] Environment terms use dev/staging/production
[ ] No new Italian terminology has been introduced
[ ] ADR index is updated if needed
[ ] Documentation links still work
[ ] git diff --check passes
```

---

## 14. Future documentation alignment plan

Recommended next steps:

```text
Phase 0.2 — Italian documentation inventory
Phase 0.3 — Translate and align docs/00-vision.md
Phase 0.4 — Translate and align docs/01-scope-mvp.md
Phase 0.5 — Translate and align architecture/API core documents
Phase 12 — Use English-only final technical documentation
```

The migration can proceed incrementally and does not block ongoing implementation work, as long as new content follows this policy.

---

## 15. Decision summary

```text
Official repository language: English
UI language: English
ADR language: English
Runbook language: English
Canonical environments: dev, staging, production
Existing Italian documentation: migrate incrementally
Localized Italian documents: allowed only as explicitly marked separate deliverables
```

---

## 16. Revision history

| Date | Phase | Description |
|---|---:|---|
| 2026-07-03 | 0.1 | Initial repository language policy. |
