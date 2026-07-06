# DevOps Control Plane — Data Model

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 10 — Data Model
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
  - `docs/05-architecture.md`
  - `docs/06-argocd-integration.md`
  - `docs/07-gitlab-integration.md`
  - `docs/08-tekton-integration.md`
  - `docs/09-security-rbac.md`
- **Status:** Rewritten in English and refreshed while preserving the original PostgreSQL-as-functional-history intent
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document defines the data model of the DevOps Control Plane.

The original document established the core data-model spirit of the project:

```text
PostgreSQL is the functional change store.
ChangeRequest is the central entity.
ChangeEvent provides the auditable timeline.
Evidence provides technical proof.
JSONB stores extensible integration payloads.
No secrets are persisted in the database.
GitLab, Tekton, Argo CD and OpenShift data are correlated into one readable operational history.
```

This refreshed version preserves that intent and aligns the document with the implemented advanced MVP baseline, including:

- current PostgreSQL tables;
- ChangeRequest lifecycle status and runtime status separation;
- `requestedBy` and multi-developer visibility;
- `targetEnvironment` and future environment catalog alignment;
- validation and deployment evidence;
- Tekton `PipelineRun` and `TaskRun` diagnostics;
- Argo CD deployment status and warnings;
- Kubernetes/OpenShift runtime evidence and diagnostics;
- future promotion metadata such as `promotionGroupID` and `promotedFromChangeNumber`;
- backup, restore and disaster recovery implications for persisted data.

The data model must support the central objective of the product:

```text
trace GitOps changes end to end in a functional, auditable and evidence-driven way
```

---

## 2. Data model principles

## 2.1 PostgreSQL stores the functional history

PostgreSQL is the internal database of the DevOps Control Plane.

PostgreSQL stores:

- ChangeRequests;
- lifecycle events;
- technical workflow events;
- validation evidence;
- deployment evidence;
- runtime status;
- audit-oriented payloads;
- external references to GitLab, Tekton, Argo CD and OpenShift.

PostgreSQL does not replace:

- GitLab Git history;
- Argo CD history;
- Tekton objects or Tekton Results;
- Kubernetes events;
- OpenShift audit logs.

PostgreSQL provides a normalized functional view that makes those external systems understandable together.

## 2.2 No secrets in the database

The database must not contain:

- GitLab tokens;
- Argo CD tokens;
- Kubernetes bearer tokens;
- kubeconfigs;
- PostgreSQL passwords in clear text;
- Kubernetes Secret values;
- Docker auth JSON values;
- private keys;
- raw logs containing credentials.

All evidence payloads must be sanitized before persistence.

## 2.3 Relational columns and JSONB are both intentional

The design uses relational columns for frequently queried, indexed and stable fields.

The design uses `jsonb` for extensible technical payloads such as:

- ChangeRequest creation payload;
- event payloads;
- Tekton diagnostics;
- Argo CD status details;
- Kubernetes/OpenShift runtime evidence;
- future promotion metadata details.

Rule:

```text
Use relational columns for stable search and filtering.
Use JSONB for variable integration payloads and diagnostics.
```

## 2.4 Lightweight event sourcing

Every significant lifecycle or technical action creates a `change_events` row.

This is not full event sourcing, but it provides a reliable audit timeline.

Examples:

```text
change_created
change_submitted
change_approved
technical_step_completed BranchCreated
technical_step_completed CommitCreated
technical_step_completed ValidationSucceeded
technical_step_completed DeploymentSyncedHealthy
technical_step_completed EvidenceCollected
```

## 2.5 Evidence is a first-class model

Evidence is not an afterthought. Evidence is one of the reasons the Control Plane exists.

Evidence must answer operational and audit questions such as:

- which validation ran;
- which task failed;
- which Argo CD state was observed;
- which deployment state was observed;
- which warnings were present;
- whether the payload was sanitized.

---

## 3. Main entities

## 3.1 Application

An Application represents an Argo CD Application known to the Control Plane.

Primary source:

```text
Argo CD API
```

Additional metadata can include:

- Git repository information;
- target namespace;
- GitOps path;
- target revision;
- current revision;
- sync status;
- health status;
- environment mapping in future phases.

The current implementation can operate with live Argo CD data and does not require every Application to be pre-cached.

## 3.2 ChangeRequest

A ChangeRequest represents a requested GitOps change.

It is the central entity of the system.

A ChangeRequest captures:

- change number;
- title;
- application name;
- target environment;
- change type;
- risk level;
- requester;
- description;
- lifecycle status;
- runtime status;
- timestamps;
- request payload.

The current model separates lifecycle status from runtime status.

## 3.3 ChangeEvent

A ChangeEvent represents one point in the history of a ChangeRequest.

Events can represent:

- lifecycle transitions;
- technical workflow completions;
- validation results;
- deployment checks;
- evidence collection;
- failures;
- authorization-relevant activity where applicable.

## 3.4 Evidence

Evidence represents technical proof associated with a ChangeRequest.

Current evidence types include:

```text
validation
deployment
```

Evidence payloads can contain GitLab, Tekton, Argo CD and Kubernetes/OpenShift data as long as the payload is sanitized.

## 3.5 Future integration snapshot

A future `integration_snapshots` table may be introduced if evidence payloads become too broad or if raw normalized snapshots must be retained separately.

For the current baseline, `evidences.payload` is sufficient.

---

## 4. Logical relationship diagram

```text
applications
    |
    | 1:N optional
    v
change_requests
    |
    | 1:N
    v
change_events

change_requests
    |
    | 1:N
    v
evidences

change_requests
    |
    | logical references
    v
GitLab branch / commit / merge request

change_requests
    |
    | logical references
    v
Tekton PipelineRun / TaskRun diagnostics

change_requests
    |
    | logical references
    v
Argo CD deployment status

change_requests
    |
    | logical references
    v
Kubernetes/OpenShift runtime evidence
```

---

## 5. Current PostgreSQL tables

The implemented MVP baseline is centered on:

```text
applications
change_requests
change_events
evidences
```

Future tables may include:

```text
users
application_repositories
application_environments
workflow_locks
integration_snapshots
promotion_groups
change_promotions
audit_exports
```

---

## 6. Table: applications

## 6.1 Purpose

Stores known application metadata, normally discovered from Argo CD API or configured for workflow correlation.

## 6.2 Conceptual fields

```sql
CREATE TABLE applications (
    id                  uuid PRIMARY KEY,
    name                text NOT NULL,
    argocd_namespace    text NOT NULL,
    argocd_project      text,
    target_namespace    text,

    repo_provider       text,
    repo_url            text,
    repo_project_id     text,
    repo_default_branch text,
    repo_path           text,

    target_revision     text,
    current_revision    text,
    sync_status         text,
    health_status       text,

    last_seen_at        timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    UNIQUE (argocd_namespace, name)
);
```

## 6.3 Notes

`name` is the Argo CD Application name.

`repo_provider` is currently expected to be:

```text
gitlab
```

`repo_project_id` can be a GitLab numeric project ID or an encoded project path, depending on the adapter configuration.

`sync_status` and `health_status` represent the latest observed Argo CD values when cached.

Future environment-aware behavior may move some fields to an environment mapping model.

---

## 7. Table: change_requests

## 7.1 Purpose

Stores the current functional and technical state of a ChangeRequest.

This is the central table of the Control Plane.

## 7.2 Current conceptual fields

```sql
CREATE TABLE change_requests (
    id                  uuid PRIMARY KEY,
    change_number       text UNIQUE NOT NULL,

    application_id      uuid REFERENCES applications(id),
    application_name    text NOT NULL,

    title               text,
    change_type         text NOT NULL,
    risk_level          text,
    target_environment  text,
    requested_by        text,
    description         text,

    status              text NOT NULL,
    runtime_status      text,
    request_payload     jsonb,

    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    completed_at        timestamptz
);
```

Some early design fields for GitLab, Tekton and Argo CD can be represented either as columns or as event/evidence payloads. The current implementation favors keeping integration details in events and evidence payloads where appropriate.

## 7.3 Change number

Human-readable format:

```text
CHG-YYYY-NNNN
```

Example:

```text
CHG-2026-0006
```

The UUID remains the technical primary key. The change number is the operator-friendly identifier.

## 7.4 Lifecycle status

Current lifecycle states:

```text
draft
submitted
approved
executing
executed
closed
```

Lifecycle status represents process state, not technical execution state.

## 7.5 Runtime status

Runtime status represents technical execution state.

Examples:

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

## 7.6 Target environment

`target_environment` identifies the intended environment.

Current canonical values:

```text
dev
staging
production
```

Current baseline defaults to `dev` where needed. Future implementation must validate this field against the environment catalog fail-closed.

## 7.7 Request payload

Example standard payload:

```json
{
  "applicationName": "demo-go-color-app",
  "targetEnvironment": "dev",
  "changeType": "standard"
}
```

Example future ConfigMap payload:

```json
{
  "configMapName": "demo-go-color-app-config",
  "values": {
    "APP_VERSION": "v3-green",
    "PAGE_COLOR": "#28A745"
  }
}
```

---

## 8. Table: change_events

## 8.1 Purpose

Stores the timeline of each ChangeRequest.

## 8.2 Conceptual fields

```sql
CREATE TABLE change_events (
    id                  uuid PRIMARY KEY,
    change_request_id   uuid NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,

    event_type          text NOT NULL,
    previous_status     text,
    new_status          text,
    message             text,
    technical_message   text,
    error_code          text,
    source              text,
    payload             jsonb,

    created_at          timestamptz NOT NULL DEFAULT now()
);
```

## 8.3 Event types

Current and recommended event types include:

```text
change_created
change_submitted
change_approved
change_execution_started
change_execution_completed
change_closed
technical_step_completed
technical_step_failed
evidence_collected
```

Technical step details are usually carried in the payload, for example:

```json
{
  "step": "ValidationSucceeded",
  "lifecycleStatus": "draft",
  "previousRuntimeStatus": "ValidationRunning"
}
```

## 8.4 Event source

Recommended sources:

```text
system
user
gitlab
argocd
tekton
kubernetes
workflow
authz
```

## 8.5 Payload examples

Branch event:

```json
{
  "branch": "change/CHG-2026-0006",
  "projectId": "12345",
  "ref": "main"
}
```

Validation failure:

```json
{
  "step": "ValidationFailed",
  "failedTasks": ["clone-repository"],
  "summary": "Tekton validation failed in task clone-repository"
}
```

---

## 9. Table: evidences

## 9.1 Purpose

Stores sanitized technical evidence associated with a ChangeRequest.

## 9.2 Conceptual fields

```sql
CREATE TABLE evidences (
    id                  uuid PRIMARY KEY,
    change_request_id   uuid NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,

    evidence_type       text NOT NULL,
    name                text,
    summary             text,
    content             text,
    payload             jsonb,
    external_ref        text,

    sanitized           boolean NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now()
);
```

## 9.3 Current evidence types

Current baseline uses:

```text
validation
deployment
```

Future or logical evidence categories may include:

```text
gitlab
tekton
argocd
kubernetes-runtime
health-check
diff-summary
workflow-summary
security-validation
```

## 9.4 Validation evidence example

```json
{
  "provider": "tekton",
  "namespace": "devops-ci-demo",
  "pipelineRunName": "devops-cp-validate-chg-2026-0006-xxxxx",
  "pipelineName": "validate-gitops",
  "gitRevision": "change/CHG-2026-0006",
  "validationPath": "apps/demo-go-color-app",
  "status": "Succeeded",
  "reason": "Succeeded",
  "taskRuns": [
    {
      "name": "clone-repository",
      "status": "Succeeded"
    },
    {
      "name": "validate-gitops-manifests",
      "status": "Succeeded"
    }
  ]
}
```

## 9.5 Failed validation evidence example

```json
{
  "provider": "tekton",
  "pipelineRunName": "devops-cp-validate-chg-2026-0001-xxxxx",
  "status": "Failed",
  "failedTaskCount": 1,
  "failedTasks": ["clone-repository"],
  "summary": "Tekton validation failed in task clone-repository"
}
```

## 9.6 Deployment evidence example

```json
{
  "change": {
    "changeNumber": "CHG-2026-0006",
    "applicationName": "demo-go-color-app",
    "targetEnvironment": "dev"
  },
  "argocd": {
    "syncStatus": "Synced",
    "healthStatus": "Healthy",
    "revision": "<revision>",
    "conditions": [
      {
        "type": "OrphanedResourceWarning",
        "message": "Application has 5 orphaned resources"
      }
    ]
  },
  "kubernetes": {
    "deployment": {
      "name": "demo-go-color-app",
      "desiredReplicas": 2,
      "readyReplicas": 2,
      "availableReplicas": 2
    },
    "pods": [
      {
        "name": "demo-go-color-app-xxxxx",
        "phase": "Running",
        "ready": true,
        "restartCount": 0
      }
    ]
  },
  "diagnostics": {
    "argocdSynced": true,
    "argocdHealthy": true,
    "deploymentReady": true,
    "readyReplicas": "2/2",
    "podsReady": "2/2",
    "totalRestarts": 0,
    "warnings": ["OrphanedResourceWarning: Application has 5 orphaned resources"]
  }
}
```

---

## 10. Future tables

## 10.1 users

A future users table may be introduced when user identity needs to be normalized beyond trusted headers.

For the current baseline, `requested_by` remains a text field derived from request data or authenticated identity context.

## 10.2 workflow_locks

A future workflow lock table may prevent conflicting changes against the same application, path, environment or resource.

Example resource key:

```text
demo-go-color-app:dev:apps/demo-go-color-app/configmap.yaml
```

## 10.3 integration_snapshots

A future integration snapshot table may store normalized snapshots from external systems if evidence payloads become too large or too specialized.

Providers:

```text
gitlab
argocd
tekton
kubernetes
```

## 10.4 promotion metadata and promotion tables

Future multi-environment promotion can start with nullable fields on `change_requests`:

```text
promotion_group_id
promoted_from_change_number
```

A richer future design may introduce:

```text
change_promotions
promotion_groups
```

The initial recommendation is to start with nullable metadata fields and introduce separate promotion tables only if workflow state requires them.

---

## 11. Status and error model

## 11.1 Error codes

Errors associated with ChangeRequests should be stored in `change_events.error_code`.

Recommended families:

```text
GITLAB_*
ARGOCD_*
TEKTON_*
KUBERNETES_*
DATABASE_*
VALIDATION_*
AUTH_*
SECURITY_*
WORKFLOW_*
```

Examples:

```text
GITLAB_BRANCH_EXISTS
GITLAB_FILE_NOT_FOUND
ARGOCD_APPLICATION_OUT_OF_SYNC
TEKTON_PIPELINERUN_FAILED
VALIDATION_SECRET_DETECTED
AUTH_FORBIDDEN
WORKFLOW_CONFLICT_ACTIVE_CHANGE
```

## 11.2 Final and non-final states

Lifecycle final states:

```text
closed
```

Technical terminal runtime states can include:

```text
ValidationSucceeded
ValidationFailed
DeploymentSyncedHealthy
DeploymentOutOfSync
DeploymentDegraded
EvidenceCollected
```

The project intentionally separates lifecycle and runtime status so that a technical step can complete without incorrectly changing business/process lifecycle state.

---

## 12. Data security rules

## 12.1 Forbidden columns

Do not create columns such as:

```text
gitlab_token
argocd_token
kubernetes_token
kubeconfig
secret_value
raw_secret_yaml
password
```

unless there is a formally approved secure design. Current baseline does not require these columns.

## 12.2 Sanitized payloads

Every `payload` and `content` field must be sanitized if it comes from external logs, command output or API responses.

Rule:

```text
If the payload may contain secrets and cannot be sanitized confidently, do not persist it as raw content.
```

## 12.3 Sanitized flag

The `evidences.sanitized` flag must be `true` for current stored evidence.

If future evidence references external raw logs, that external storage model must be documented with retention and access controls.

---

## 13. Current migration baseline

The initial migration creates the four core tables:

```text
applications
change_requests
change_events
evidences
```

The exact SQL lives in the repository migrations and should be treated as source of truth for runtime schema.

This design document describes the intended model and evolution direction. When code and migrations differ, the migration files and Go repository code must be reviewed as the runtime truth and this document should be updated.

---

## 14. Useful queries

## 14.1 Recent changes

```sql
SELECT
    change_number,
    application_name,
    target_environment,
    status,
    runtime_status,
    requested_by,
    created_at,
    completed_at
FROM change_requests
ORDER BY created_at DESC
LIMIT 50;
```

## 14.2 ChangeRequest timeline

```sql
SELECT
    event_type,
    previous_status,
    new_status,
    message,
    error_code,
    source,
    payload,
    created_at
FROM change_events
WHERE change_request_id = $1
ORDER BY created_at ASC;
```

## 14.3 ChangeRequest evidence

```sql
SELECT
    evidence_type,
    name,
    summary,
    sanitized,
    external_ref,
    created_at
FROM evidences
WHERE change_request_id = $1
ORDER BY created_at ASC;
```

## 14.4 Active changes by application and environment

```sql
SELECT
    change_number,
    application_name,
    target_environment,
    status,
    runtime_status,
    requested_by,
    created_at
FROM change_requests
WHERE application_name = $1
  AND target_environment = $2
  AND status NOT IN ('closed')
ORDER BY created_at ASC;
```

---

## 15. Mapping to API

## 15.1 `GET /api/v1/changes`

Primary source:

```text
change_requests
```

Returned fields should include:

- change number;
- title;
- application;
- target environment;
- lifecycle status;
- runtime status;
- requester;
- creation timestamp.

## 15.2 `GET /api/v1/changes/{id}`

Primary sources:

```text
change_requests
change_events
evidences
```

Returned data should include:

- ChangeRequest detail;
- lifecycle and runtime status;
- events;
- GitLab references from event/evidence payloads;
- Tekton validation status;
- Argo CD deployment status;
- Kubernetes/OpenShift evidence.

## 15.3 `GET /api/v1/changes/{id}/evidence`

Primary source:

```text
evidences
```

Returned data should include evidence summaries and sanitized payloads.

## 15.4 `GET /api/v1/applications`

Primary source:

```text
Argo CD API live data
applications cache where available
```

---

## 16. Mapping to workflows

## 16.1 GitLab workflow

GitLab branch, file update, commit and merge request details are represented through:

- ChangeRequest runtime status;
- ChangeEvent payloads;
- evidence payloads where needed.

## 16.2 Tekton validation workflow

Tekton validation is represented through:

- runtime status `ValidationRunning`, `ValidationSucceeded` or `ValidationFailed`;
- ChangeEvent payloads;
- validation evidence;
- TaskRun diagnostics inside evidence payload.

## 16.3 Argo CD deployment check workflow

Argo CD deployment state is represented through:

- runtime status such as `DeploymentSyncedHealthy`;
- ChangeEvent payloads;
- deployment evidence payload.

## 16.4 Runtime evidence workflow

Kubernetes/OpenShift state is represented through deployment evidence and diagnostics.

---

## 17. Consistency rules

## 17.1 ChangeRequest without application ID

`application_id` may remain nullable because a ChangeRequest can be created before the Application cache exists or when the system uses live Argo CD data.

`application_name` remains mandatory.

## 17.2 Final lifecycle status and completion timestamp

When lifecycle status reaches a final business state, `completed_at` should be set where applicable.

The current implementation may keep lifecycle and runtime closure intentionally separate; consistency rules should not collapse the two models.

## 17.3 Tekton references

If validation evidence contains a PipelineRun name, the namespace and pipeline name should also be available in the payload.

## 17.4 Argo CD references

If deployment evidence contains an Argo CD Application, the evidence should include sync status, health status and revision where available.

## 17.5 Target environment

Every new ChangeRequest should include `target_environment` or be defaulted to `dev` by explicit compatibility logic.

Future validation must reject unknown target environments.

---

## 18. Backup, restore and retention

## 18.1 Backup

PostgreSQL stores functional history and evidence. It must be backed up.

Current baseline includes:

- PostgreSQL backup runbook;
- checksum validation;
- restore validation plan;
- isolated restore test;
- disaster recovery runbook.

## 18.2 Restore validation

Restore validation must prove that core tables and representative ChangeRequests can be restored.

Representative validation queries should check:

- `applications` count;
- `change_requests` count;
- `change_events` count;
- `evidences` count;
- known ChangeRequest records.

## 18.3 Evidence retention

Evidence can grow over time.

Current guidance:

- store summaries and sanitized payloads;
- avoid large raw logs;
- keep content excerpts bounded;
- consider external storage or Tekton Results integration only in future phases.

---

## 19. Multi-environment data model direction

The multi-environment model introduces environment-aware ChangeRequests and future promotion chains.

Canonical environments:

```text
dev
staging
production
```

Current and future data model requirements:

- `target_environment` must be visible and queryable;
- environment configuration remains externalized in the initial design;
- future promotion metadata can link related ChangeRequests;
- production workflows must remain guarded by RBAC/AuthZ and approval policy.

Future fields:

```text
promotion_group_id
promoted_from_change_number
```

These fields should be nullable for backward compatibility.

---

## 20. Data model validation checklist

The data model baseline is ready when:

- the core tables exist;
- each ChangeRequest has a readable change number;
- lifecycle and runtime status are distinguishable;
- requester is visible;
- target environment is represented;
- events are persisted;
- evidence records are persisted;
- evidence is sanitized;
- no Secret value is part of the schema;
- GitLab, Tekton, Argo CD and Kubernetes/OpenShift references can be correlated;
- backup and restore procedures protect the data;
- future promotion metadata has a clear extension path.

---

## 21. Relationship with other documents

This document informs and is informed by:

- `migrations/`;
- `internal/domain/`;
- `internal/database/`;
- `docs/11-change-workflows.md`;
- `docs/12-evidence-model.md`;
- `docs/13-api-design.md`;
- `docs/environment-configuration-model.md`;
- `docs/change-promotion-model.md`;
- `docs/runbooks/postgresql-backup-restore.md`;
- `docs/runbooks/disaster-recovery.md`;
- `docs/adr/ADR-0004-postgresql-change-history.md`.

---

## 22. Key message

The data model must make every GitOps change reconstructable.

The database must answer questions such as:

```text
Who requested the change?
Which application and environment were targeted?
What was the requested change?
Which GitLab branch, commit or merge request represented the change?
Which Tekton PipelineRun validated the change?
Which TaskRun failed, if validation failed?
Which Argo CD state was observed?
Which Kubernetes/OpenShift runtime state was collected?
Which evidence is available?
Was the evidence sanitized?
```

The database must not replace GitLab, Tekton, Argo CD or OpenShift. The database must correlate information from those systems into one readable, auditable and evidence-driven functional history.

This preserves the original spirit of the project: one small but solid data model that makes GitOps workflows understandable, traceable and safe to audit.

---

## 23. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial data model document in Italian. |
| 2026-07-06 | 0.2 | Rewritten in English and refreshed while preserving the original PostgreSQL change-store, event timeline and evidence model intent and aligning it with the current advanced MVP and multi-environment direction. |
