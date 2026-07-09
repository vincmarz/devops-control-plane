# DevOps Control Plane — Evidence Model

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 12 — Evidence Model
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
  - `docs/10-data-model.md`
  - `docs/11-change-workflows.md`
- **Status:** Rewritten in English and refreshed while preserving the original evidence-first and sanitized-evidence intent
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document defines the evidence model of the DevOps Control Plane.

The original document established one of the most important ideas of the project:

```text
Executing a change is not enough.
The system must be able to prove what happened, when it happened, which tools were involved and what the outcome was.
```

This refreshed version preserves that original evidence-first intent and updates the document to match the current advanced MVP baseline.

The document describes:

- what evidence means in the DevOps Control Plane;
- which evidence is currently collected;
- when evidence is collected;
- how evidence is normalized;
- how evidence is stored in PostgreSQL;
- which data must never be stored;
- how GitLab, Tekton, Argo CD and Kubernetes/OpenShift evidence is correlated;
- how evidence supports audit, troubleshooting, UI visibility and onboarding;
- how evidence will evolve for `dev`, `staging` and `production` workflows.

Evidence is central to the value of the DevOps Control Plane. Without evidence, the system would only orchestrate technical actions. With evidence, the system becomes auditable, explainable and useful for operations and governance.

---

## 2. Definition of evidence

In the DevOps Control Plane, evidence is a technical, normalized and sanitized record associated with a ChangeRequest.

Evidence must help answer questions such as:

```text
Which Git branch was created?
Which commit or merge request represented the change?
Which Tekton PipelineRun validated the change?
Which TaskRun failed, if validation failed?
Which Argo CD state was observed?
Was the application Synced and Healthy?
Which runtime state was collected from OpenShift?
Were Pods Running and Ready?
Were warnings or errors observed?
Was the evidence sanitized?
```

Evidence is not raw logging. Evidence is curated technical proof that supports reconstruction of a workflow.

---

## 3. Evidence model principles

## 3.1 Evidence first

Every meaningful workflow step should produce evidence or an auditable event.

Core rule:

```text
If a workflow changes Git, validates with Tekton, checks Argo CD or reads OpenShift runtime state, it should produce evidence or an event that can be reviewed later.
```

## 3.2 Evidence must be sanitized

Evidence must be useful, but safe.

Evidence must not contain:

- tokens;
- passwords;
- Secret values;
- kubeconfigs;
- Authorization headers;
- private keys;
- Docker auth JSON;
- raw Kubernetes Secret YAML;
- unfiltered full logs.

## 3.3 Evidence must be readable

Each evidence record should include:

- evidence type;
- name;
- human-readable summary;
- normalized technical payload;
- timestamp;
- external reference where available;
- sanitized flag.

## 3.4 Evidence must be correlated

Every evidence record must be associated with:

```text
change_request_id
```

When possible, evidence payloads should also include:

- `changeNumber`;
- `applicationName`;
- `targetEnvironment`;
- `provider`;
- `namespace`;
- `resourceName`;
- `revision`;
- `status`.

## 3.5 Evidence is append-oriented

Evidence should be treated as a historical record.

The default model is:

```text
Do not overwrite old evidence.
Append new evidence when new observations are collected.
```

---

## 4. Current evidence types

The current implementation baseline uses the following evidence types:

```text
validation
deployment
```

These types intentionally aggregate multiple external systems:

- `validation` evidence focuses on Tekton validation and GitOps policy checks;
- `deployment` evidence focuses on Argo CD state and Kubernetes/OpenShift runtime state.

The original conceptual categories remain useful for future refinement:

```text
gitlab
tekton
argocd
kubernetes-runtime
health-check
diff-summary
workflow-summary
security-validation
error-summary
```

The current data model can support both compact evidence types and richer payload categorization.

---

## 5. Evidence table

The primary data model is defined in `docs/10-data-model.md`.

Current conceptual table:

```sql
CREATE TABLE evidences (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
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

Rules:

- `payload` must be sanitized before storage;
- `sanitized` should be `true` for current evidence;
- `content` should contain only bounded excerpts or safe text;
- large raw logs should not be stored directly in PostgreSQL.

---

## 6. Validation evidence

## 6.1 Purpose

Validation evidence proves that a GitOps change was validated by Tekton before the workflow progressed.

Validation evidence is generated from:

- Tekton `PipelineRun` status;
- related `TaskRun` diagnostics;
- Git URL and revision;
- validation path;
- GitOps policy guardrail results;
- anti-secret checks.

## 6.2 When to collect it

Validation evidence is collected when:

- validation is checked and reaches a terminal state;
- validation succeeds;
- validation fails;
- TaskRun diagnostics are available;
- policy guardrails detect a violation.

## 6.3 Successful validation example

```json
{
  "provider": "tekton",
  "namespace": "devops-ci-demo",
  "pipelineName": "validate-gitops",
  "pipelineRunName": "devops-cp-validate-chg-2026-0006-xxxxx",
  "status": "Succeeded",
  "reason": "Succeeded",
  "gitRevision": "change/CHG-2026-0006",
  "validationPath": "apps/demo-go-color-app",
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

## 6.4 Failed validation example

```json
{
  "provider": "tekton",
  "namespace": "devops-ci-demo",
  "pipelineRunName": "devops-cp-validate-chg-2026-0001-xxxxx",
  "status": "Failed",
  "reason": "Failed",
  "failedTaskCount": 1,
  "failedTasks": [
    "clone-repository"
  ],
  "summary": "Tekton validation failed in task clone-repository",
  "errorCode": "TEKTON_PIPELINERUN_FAILED"
}
```

## 6.5 TaskRun diagnostics

TaskRun diagnostics should include:

- TaskRun name;
- PipelineRun name;
- pipeline task name;
- status;
- reason;
- message;
- failed step where available;
- safe summary.

The purpose is to avoid forcing operators to manually inspect every TaskRun before understanding the failure point.

---

## 7. Deployment evidence

## 7.1 Purpose

Deployment evidence proves which deployment state was observed after a change or during evidence collection.

Deployment evidence combines:

- ChangeRequest metadata;
- Argo CD status;
- Kubernetes/OpenShift runtime state;
- diagnostics summary;
- warnings.

## 7.2 When to collect it

Deployment evidence is collected when:

- a deployment state check has been performed;
- runtime evidence collection is requested;
- a post-change audit snapshot is needed;
- troubleshooting requires a point-in-time runtime view.

## 7.3 Deployment evidence example

```json
{
  "change": {
    "changeNumber": "CHG-2026-0006",
    "applicationName": "demo-go-color-app",
    "targetEnvironment": "dev",
    "lifecycleStatus": "draft",
    "runtimeStatus": "EvidenceCollected"
  },
  "argocd": {
    "application": "demo-go-color-app",
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
      "namespace": "devops-ci-demo",
      "desiredReplicas": 2,
      "readyReplicas": 2,
      "availableReplicas": 2,
      "updatedReplicas": 2,
      "generationObserved": true
    },
    "pods": [
      {
        "name": "demo-go-color-app-xxxxx",
        "phase": "Running",
        "ready": true,
        "restartCount": 0
      }
    ],
    "service": {
      "name": "demo-go-color-app",
      "available": true
    },
    "route": {
      "name": "demo-go-color-app",
      "available": true
    }
  },
  "diagnostics": {
    "argocdSynced": true,
    "argocdHealthy": true,
    "deploymentReady": true,
    "generationObserved": true,
    "readyReplicas": "2/2",
    "podsReady": "2/2",
    "totalRestarts": 0,
    "serviceAvailable": true,
    "routeAvailable": true,
    "warnings": [
      "OrphanedResourceWarning: Application has 5 orphaned resources"
    ],
    "summary": "Application demo-go-color-app is Synced/Healthy with 2/2 replicas ready; warnings: 1"
  }
}
```

---

## 8. GitLab evidence

## 8.1 Purpose

GitLab evidence proves which Git change represented the desired-state change.

Although the current implementation often stores GitLab references through events, GitLab evidence remains a useful logical evidence category.

## 8.2 When to collect it

GitLab evidence can be collected:

- after branch creation;
- after file update or commit;
- after merge request opening;
- after merge request merge;
- when GitLab operations fail.

## 8.3 Minimum payload

```json
{
  "provider": "gitlab",
  "projectId": "12345",
  "repoUrl": "https://gitlab.example.local/group/demo-app-gitops.git",
  "sourceBranch": "change/CHG-2026-0006",
  "targetBranch": "main",
  "commitSha": "<sha>",
  "commitShortSha": "<short-sha>",
  "mergeRequestIid": 42,
  "mergeRequestUrl": "https://gitlab.example.local/group/repo/-/merge_requests/42",
  "mergeRequestState": "opened",
  "filesChanged": [
    "apps/demo-go-color-app/deployment.yaml"
  ]
}
```

## 8.4 Suggested summary

```text
GitLab change created on branch change/CHG-2026-0006 and represented by merge request !42.
```

---

## 9. Diff summary evidence

## 9.1 Purpose

Diff summary evidence explains what changed in GitOps files without forcing a user to read a full raw diff.

## 9.2 Example: replica change

```json
{
  "files": [
    {
      "path": "apps/demo-go-color-app/deployment.yaml",
      "changeType": "update",
      "summary": "spec.replicas changed from 2 to 3",
      "fields": [
        {
          "field": "spec.replicas",
          "oldValue": 2,
          "newValue": 3
        }
      ]
    }
  ]
}
```

## 9.3 Redaction rule

If a modified field appears sensitive, values must be redacted.

Example:

```json
{
  "field": "data.PASSWORD",
  "oldValue": "***REDACTED***",
  "newValue": "***REDACTED***",
  "redacted": true
}
```

---

## 10. Argo CD evidence

## 10.1 Purpose

Argo CD evidence proves which GitOps deployment state was observed.

## 10.2 Successful status example

```json
{
  "provider": "argocd",
  "application": "demo-go-color-app",
  "targetEnvironment": "dev",
  "syncStatus": "Synced",
  "healthStatus": "Healthy",
  "revision": "<revision>",
  "conditions": []
}
```

## 10.3 Warning example

```json
{
  "provider": "argocd",
  "application": "demo-go-color-app",
  "syncStatus": "Synced",
  "healthStatus": "Healthy",
  "conditions": [
    {
      "type": "OrphanedResourceWarning",
      "message": "Application has 5 orphaned resources"
    }
  ]
}
```

Rule:

```text
OrphanedResourceWarning must be preserved, but it must not automatically convert a Synced/Healthy Application into a failed deployment.
```

---

## 11. Kubernetes/OpenShift runtime evidence

## 11.1 Purpose

Kubernetes/OpenShift runtime evidence proves the actual runtime state observed from the cluster.

## 11.2 Target resources

Current runtime evidence can include:

- Deployment;
- Pods;
- Service;
- Route;
- selected runtime metadata;
- diagnostics derived from runtime state.

## 11.3 Deployment payload example

```json
{
  "provider": "kubernetes",
  "namespace": "devops-ci-demo",
  "kind": "Deployment",
  "name": "demo-go-color-app",
  "desiredReplicas": 2,
  "readyReplicas": 2,
  "availableReplicas": 2,
  "updatedReplicas": 2,
  "observedGenerationMatches": true
}
```

## 11.4 Pod summary example

```json
{
  "provider": "kubernetes",
  "namespace": "devops-ci-demo",
  "kind": "PodSummary",
  "selector": "app=demo-go-color-app",
  "pods": [
    {
      "name": "demo-go-color-app-abcde",
      "phase": "Running",
      "ready": true,
      "restartCount": 0
    }
  ]
}
```

## 11.5 ConfigMap evidence caution

ConfigMaps are not Secrets, but ConfigMaps may still contain sensitive data in poorly designed applications.

If a ConfigMap key appears sensitive, the value must be redacted.

---

## 12. Health-check evidence

## 12.1 Purpose

Health-check evidence proves that an application or Control Plane endpoint responded at a given time.

## 12.2 Preferred endpoints

```text
/healthz
/readyz
/livez
/
```

The exact endpoint depends on the target application or component.

## 12.3 Payload example

```json
{
  "provider": "http",
  "url": "https://demo-go-color-app-devops-ci-demo.apps.example.local/healthz",
  "statusCode": 200,
  "responseExcerpt": "ok",
  "durationMs": 42,
  "timestamp": "2026-06-25T15:10:00+02:00"
}
```

Rules:

- do not store pages containing sensitive data;
- limit response excerpts;
- use short timeouts;
- record errors if endpoint is unreachable.

---

## 13. Security-validation evidence

## 13.1 Purpose

Security-validation evidence proves that minimum safety guardrails were executed.

Current security validation is mostly represented through Tekton validation evidence and policy guardrail output.

## 13.2 Successful check example

```json
{
  "provider": "devops-control-plane",
  "check": "anti-secret-check",
  "status": "Passed",
  "filesScanned": [
    "apps/demo-go-color-app/deployment.yaml",
    "apps/demo-go-color-app/configmap.yaml"
  ],
  "matches": []
}
```

## 13.3 Failed check example

```json
{
  "provider": "devops-control-plane",
  "check": "anti-secret-check",
  "status": "Failed",
  "errorCode": "VALIDATION_SECRET_DETECTED",
  "matches": [
    {
      "file": "apps/demo-go-color-app/configmap.yaml",
      "pattern": "password",
      "value": "***REDACTED***"
    }
  ]
}
```

---

## 14. Workflow summary evidence

## 14.1 Purpose

Workflow summary evidence provides a human-readable final or intermediate summary of the ChangeRequest.

Current implementation may use event history plus validation/deployment evidence instead of a distinct `workflow-summary` evidence type, but the concept remains useful.

## 14.2 Completed example

```json
{
  "changeNumber": "CHG-2026-0006",
  "application": "demo-go-color-app",
  "targetEnvironment": "dev",
  "finalRuntimeStatus": "EvidenceCollected",
  "summary": "Application reached Synced/Healthy state and deployment evidence was collected.",
  "tekton": {
    "status": "Succeeded"
  },
  "argocd": {
    "syncStatus": "Synced",
    "healthStatus": "Healthy"
  },
  "runtime": {
    "readyReplicas": "2/2"
  }
}
```

## 14.3 Failed example

```json
{
  "changeNumber": "CHG-2026-0001",
  "application": "demo-go-color-app",
  "targetEnvironment": "dev",
  "failedPhase": "Tekton validation",
  "errorCode": "TEKTON_PIPELINERUN_FAILED",
  "message": "Tekton validation failed in task clone-repository.",
  "suggestedAction": "Review TaskRun diagnostics associated with the ChangeRequest."
}
```

---

## 15. Error-summary evidence

## 15.1 Purpose

Error-summary evidence normalizes complex failures into a readable and auditable format.

## 15.2 Standard payload

```json
{
  "code": "TEKTON_PIPELINERUN_FAILED",
  "phase": "Validation",
  "technicalMessage": "PipelineRun devops-cp-validate-chg-2026-0001 failed",
  "message": "Tekton validation failed for the ChangeRequest.",
  "suggestedAction": "Review TaskRun diagnostics associated with the ChangeRequest.",
  "recoverable": true
}
```

---

## 16. Evidence naming

Recommended evidence naming pattern:

```text
<purpose>-evidence
```

Current examples:

```text
tekton-validation-evidence
deployment-evidence
```

The original provider-based naming remains useful for future logical categories:

```text
gitlab-commit-CHG-2026-0006
tekton-validation-CHG-2026-0006
argocd-deployment-CHG-2026-0006
runtime-deployment-CHG-2026-0006
workflow-summary-CHG-2026-0006
```

---

## 17. Sanitization

## 17.1 Sanitization responsibility

Every adapter, service or evidence builder must sanitize payloads before persistence.

Minimum suspicious patterns:

```text
token
password
secret
client_secret
authToken
secret_key
private_key
PRIVATE KEY
BEGIN RSA
AWS access key patterns
.dockerconfigjson
Authorization:
Bearer
```

## 17.2 Redaction format

Use:

```text
***REDACTED***
```

Example:

```json
{
  "key": "DATABASE_PASSWORD",
  "value": "***REDACTED***"
}
```

## 17.3 Sanitized flag

Current evidence should have:

```text
sanitized = true
```

If future evidence references unsanitized external storage, that design must explicitly define access controls and retention rules.

---

## 18. Size and retention

## 18.1 Recommended limits

```text
summary: short and human-readable
content: bounded excerpt only
payload: normalized JSONB
logExcerpt: bounded by configuration, for example 10 KiB
```

## 18.2 Large data

For large data:

- store a summary;
- store an external reference if needed;
- avoid full dumps in PostgreSQL;
- consider Tekton Results or external object storage in future phases.

## 18.3 Retention

Current baseline keeps evidence in PostgreSQL.

Future work may introduce:

- retention policy;
- audit export;
- compression;
- external storage;
- pruning of non-critical evidence.

---

## 19. Evidence lifecycle

## 19.1 Creation

Evidence can be created by:

- ChangeService;
- Evidence Service;
- GitLab workflow;
- Tekton validation workflow;
- Argo CD deployment check workflow;
- Kubernetes/OpenShift runtime evidence workflow;
- operability scripts where appropriate.

## 19.2 Update model

Evidence should be append-oriented.

Rule:

```text
Do not overwrite historical evidence. Add a new evidence record for a new observation.
```

## 19.3 Deletion model

Current baseline does not include ordinary application-level evidence deletion.

Future policy may define:

- retention-based deletion;
- audit export before pruning;
- preservation of final evidence per ChangeRequest.

---

## 20. Useful evidence queries

## 20.1 Evidence by ChangeRequest

```sql
SELECT
    evidence_type,
    name,
    summary,
    sanitized,
    created_at
FROM evidences
WHERE change_request_id = $1
ORDER BY created_at ASC;
```

## 20.2 Evidence by type

```sql
SELECT
    name,
    summary,
    created_at
FROM evidences
WHERE change_request_id = $1
  AND evidence_type = $2
ORDER BY created_at ASC;
```

## 20.3 Recent deployment evidence

```sql
SELECT
    name,
    summary,
    external_ref,
    created_at
FROM evidences
WHERE evidence_type = 'deployment'
ORDER BY created_at DESC
LIMIT 20;
```

---

## 21. API mapping

Current evidence endpoints use `/api/v1`.

```text
GET /api/v1/changes/{id}/evidence
GET /api/v1/changes/{id}/evidence/validation
GET /api/v1/changes/{id}/evidence/deployment
```

The UI also exposes evidence through ChangeRequest detail pages.

Legacy endpoint references without `/api/v1` should be considered historical and updated during documentation migration.

---

## 22. UI rendering

Evidence should be understandable in the Web UI.

Current UI direction:

- show compact runtime evidence cards;
- show deployment diagnostics;
- show Pod, Service and Route details;
- expose detailed payloads when useful;
- avoid hiding warnings;
- avoid exposing sensitive values.

Raw payloads may be useful for technical users, but they must remain sanitized.

---

## 23. Multi-environment evidence direction

Future evidence must include environment context.

Canonical environments:

```text
dev
staging
production
```

Evidence should include:

```text
targetEnvironment
environment-specific namespace
environment-specific Argo CD Application
environment-specific Tekton PipelineRun
environment-specific GitOps path
```

Promotion evidence should also include:

```text
promotionGroupID
promotedFromChangeNumber
```

Production evidence may require stricter retention, approval and audit rules.

---

## 24. Evidence readiness checklist

A completed technical workflow should have:

- ChangeRequest metadata;
- GitLab branch, commit or merge request references where applicable;
- validation evidence when validation was executed;
- TaskRun diagnostics when Tekton was involved;
- Argo CD status;
- Kubernetes/OpenShift runtime evidence;
- diagnostics summary;
- warning and error details where present;
- sanitized payloads;
- no tokens, passwords or Secret values.

A failed workflow should have:

- failed phase;
- error code;
- readable message;
- suggested action where possible;
- last observed state;
- sanitized diagnostics;
- partial evidence if available.

---

## 25. Relationship with other documents

This document informs and is informed by:

- `internal/app/evidence_service.go`;
- `internal/domain/evidence.go`;
- `internal/database/repositories.go`;
- `docs/10-data-model.md`;
- `docs/11-change-workflows.md`;
- `docs/13-api-design.md`;
- `docs/07-gitlab-integration.md`;
- `docs/08-tekton-integration.md`;
- `docs/06-argocd-integration.md`;
- `docs/environment-configuration-model.md`;
- `docs/change-promotion-model.md`;
- future retention and audit export policies.

---

## 26. Key message

Evidence turns technical automation into a governable process.

The DevOps Control Plane must be able to say:

```text
Git was changed.
Tekton validated the change.
Argo CD deployment state was checked.
OpenShift runtime state was observed.
A safe and readable proof was saved for each significant step.
```

Without evidence, the system would only be a wrapper around commands and APIs.

With evidence, the system becomes an auditable DevOps Control Plane that supports operations, troubleshooting, onboarding and governance.

This preserves the original evidence-first spirit of the project while aligning the model with the current validation and deployment evidence baseline.

---

## 27. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial evidence model document in Italian. |
| 2026-07-06 | 0.2 | Rewritten in English and refreshed while preserving the original evidence-first and sanitized-evidence model and aligning it with validation evidence, deployment evidence, diagnostics and multi-environment direction. |

## Post-Phase 15 runtime and validation evidence alignment

Status: Active evidence baseline  
Phase reference: 13.2  
Last updated: 2026-07-09

### Purpose

This section refreshes the Evidence Model after completion of Phase 15 and the post-closure simulated staging and production cluster readiness validation.

The current DevOps Control Plane baseline is namespace-isolated on the available `ocp-dev` OpenShift cluster:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

Physical cross-cluster runtime validation remains deferred because no additional OpenShift cluster is currently available.

The evidence model is now aligned with the namespace-isolated runtime baseline, the UI runtime evidence rendering, Tekton validation evidence and the multi-cluster code-ready guardrails.

### Evidence categories in the current baseline

The current baseline uses the following evidence categories:

- ChangeRequest lifecycle evidence;
- audit event evidence;
- GitLab workflow evidence;
- Argo CD deployment evidence;
- Kubernetes/OpenShift runtime evidence;
- Tekton validation evidence;
- deployment diagnostics;
- validation diagnostics;
- sanitized raw evidence for operator review.

The most important post-Phase 15 addition is that validation evidence and runtime evidence are now visible and useful in the UI, instead of being only backend records.

### Runtime evidence

Runtime evidence represents the observed state of the target runtime environment.

For the current namespace-isolated baseline, runtime evidence may refer to:

- target environment;
- cluster name;
- Kubernetes namespace;
- deployment name;
- deployment readiness;
- available replicas;
- updated replicas;
- pods;
- services;
- routes;
- health endpoint result;
- Argo CD Application status.

Runtime evidence must always preserve the target environment and namespace, because dev, staging and production currently share the same physical OpenShift cluster.

This avoids ambiguity between:

- `devops-ci-demo`;
- `devops-ci-staging`;
- `devops-ci-production`.

### Tekton validation evidence

Tekton validation evidence represents the result of a validation PipelineRun.

The current model records, renders or derives the following fields:

- target environment;
- Tekton namespace;
- Pipeline name;
- PipelineRun name;
- Git revision or branch;
- validation path;
- status;
- reason;
- failed task count;
- failed tasks, when available;
- evidence sanitization status.

The validation path is environment-specific.

Validated examples:

- staging validation path: `apps/demo-go-color-app/overlays/staging`;
- production validation path: `apps/demo-go-color-app/overlays/production`.

### Latest validation evidence

The UI now uses the latest validation evidence associated with the selected ChangeRequest.

This is important because a ChangeRequest can accumulate multiple events and multiple evidence records over time.

The UI must show the most relevant validation result for the operator, especially after `check-validation`.

The latest validation evidence card is expected to show:

- PipelineRun;
- Tekton namespace;
- Pipeline;
- Git revision;
- validation path;
- status;
- reason;
- failed task count;
- sanitized evidence state.

### Validated staging evidence

The final namespace-isolated staging validation record is:

- ChangeRequest: `CHG-2026-0049`;
- environment: `staging`;
- Tekton namespace: `devops-ci-staging`;
- PipelineRun: `devops-cp-validate-chg-2026-0049-nd7rm`;
- validation path: `apps/demo-go-color-app/overlays/staging`;
- failed task count: `0`;
- evidence sanitized: `true`;
- result: `Succeeded`.

### Validated production evidence

The final namespace-isolated production validation record is:

- ChangeRequest: `CHG-2026-0050`;
- environment: `production`;
- Tekton namespace: `devops-ci-production`;
- PipelineRun: `devops-cp-validate-chg-2026-0050-8wqtv`;
- validation path: `apps/demo-go-color-app/overlays/production`;
- failed task count: `0`;
- evidence sanitized: `true`;
- result: `Succeeded`.

### UI rendering expectations

The ChangeRequest detail UI must render evidence in an operator-friendly way.

Expected UI behavior:

- runtime evidence is shown in compact cards;
- latest Tekton validation evidence is shown when available;
- failed task count is visible;
- validation path is visible;
- Tekton namespace is visible;
- PipelineRun name is visible;
- sanitized state is visible;
- raw evidence remains available only as sanitized diagnostic material.

The dashboard also reflects the evidence model by selecting the latest ChangeRequest and showing the current `Environments / Namespaces` mapping.

### Evidence sanitization

Evidence must remain sanitized.

Evidence may contain operational metadata such as:

- namespace names;
- resource names;
- PipelineRun names;
- status values;
- reason values;
- Git revision or branch names;
- validation paths.

Evidence must not contain:

- raw Secret values;
- bearer tokens;
- kubeconfig payloads;
- private keys;
- decoded Secret content;
- sensitive credential material.

The expected sanitized state for successful validation evidence is:

`evidence sanitized=true`

### Relationship with multi-cluster readiness

The evidence model is now compatible with future real multi-cluster onboarding.

The current physical runtime baseline is still namespace-isolated on `ocp-dev`.

However, the codebase has also validated simulated external cluster targets:

- `staging` -> `ocp-staging-simulated`;
- `production` -> `ocp-production-simulated`.

This means evidence must continue to preserve:

- target environment;
- resolved cluster name;
- Kubernetes namespace;
- Tekton namespace;
- Argo CD application;
- validation path.

When real clusters become available, evidence must prove that the target did not silently fall back to `ocp-dev`.

### Operational interpretation

Operators should use evidence to answer the following questions:

1. Which ChangeRequest was validated?
2. Which environment was targeted?
3. Which namespace was targeted?
4. Which PipelineRun executed?
5. Which GitOps path was validated?
6. Did the validation succeed?
7. Were there failed tasks?
8. Was the evidence sanitized?
9. Does the UI show the same result as the backend evidence?
10. Is the target consistent with the expected environment mapping?

### Closure statement

The evidence model is aligned with the current post-Phase 15 runtime baseline.

The DevOps Control Plane now provides evidence that is useful for operational validation, UI inspection, troubleshooting and future multi-cluster onboarding.

Physical cross-cluster evidence remains deferred until additional OpenShift clusters are available.

Namespace-isolated runtime evidence and simulated external-cluster readiness evidence are both documented and validated.
