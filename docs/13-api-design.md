# DevOps Control Plane — API Design

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 13 — API Design
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
  - `docs/12-evidence-model.md`
- **Status:** Rewritten in English and refreshed while preserving the original API-first and workflow-driven intent
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document defines the HTTP/REST API design of the DevOps Control Plane.

The original document established the initial API spirit of the project:

```text
API-first
JSON as primary format
consistent responses
versioned API paths
no secrets in API responses
workflow-driven endpoints
alignment with data model, workflows and evidence
```

This refreshed version preserves that original intent and aligns the API documentation with the current advanced MVP baseline, including:

- canonical `/api/v1` endpoints;
- public `/readyz` and `/livez` health endpoints;
- ChangeRequest lifecycle actions;
- explicit technical workflow actions;
- GitLab branch, file update, merge request and merge workflow;
- Tekton validation and validation check workflow;
- Argo CD deployment check workflow;
- deployment evidence collection;
- evidence retrieval endpoints;
- OpenShift OAuth Proxy and backend AuthZ behavior;
- server-side Web UI relationship;
- multi-developer visibility;
- future multi-environment direction for `dev`, `staging` and `production`.

The API design guides the Go backend implementation and must remain aligned with:

- `docs/10-data-model.md`;
- `docs/11-change-workflows.md`;
- `docs/12-evidence-model.md`.

---

## 2. API principles

## 2.1 API-first, UI-aligned

The original principle was API-first before a complete UI. That principle remains valid.

The current baseline also includes a server-side Web UI. The UI must remain aligned with backend APIs and service workflows; it must not bypass backend validation or authorization.

Principle:

```text
Stable workflows and APIs first. UI actions are thin, server-side entry points into the same backend workflows.
```

## 2.2 JSON as primary application format

Application APIs use JSON.

Expected headers:

```http
Content-Type: application/json
Accept: application/json
```

HTML endpoints under `/ui` are separate server-side rendered UI routes.

## 2.3 Consistent response envelope

API responses should be predictable.

Single object response:

```json
{
  "data": {},
  "meta": {},
  "error": null
}
```

List response:

```json
{
  "data": [],
  "meta": {
    "count": 0,
    "limit": 50,
    "offset": 0
  },
  "error": null
}
```

Error response:

```json
{
  "data": null,
  "meta": {},
  "error": {
    "code": "VALIDATION_INVALID_REQUEST",
    "message": "Unable to create ChangeRequest",
    "technicalMessage": "title is required",
    "suggestedAction": "Provide the required title field.",
    "recoverable": true
  }
}
```

The current implementation may return a subset of optional fields. The target design remains the consistent envelope above.

## 2.4 No secrets in API responses

APIs must never return:

- GitLab tokens;
- Argo CD tokens;
- PostgreSQL passwords;
- Kubernetes bearer tokens;
- kubeconfigs;
- raw Kubernetes Secrets;
- Authorization headers;
- private keys;
- unredacted sensitive values.

## 2.5 Explicit workflow steps before full automation

The original design included a possible future automatic workflow endpoint. The current design prefers explicit step-by-step technical actions because they are easier to validate, explain and audit.

Future full automation can be added only after the step-by-step workflows are stable and governed.

---

## 3. API versioning

The canonical API prefix is:

```text
/api/v1
```

Examples:

```text
GET /api/v1/applications
GET /api/v1/changes
POST /api/v1/changes/{id}/validate
```

Legacy references without `/api/v1` should be treated as historical and updated during documentation migration.

---

## 4. Health endpoints

Health endpoints are intentionally outside `/api/v1`.

They must remain reachable through the Route without authentication, while business API and UI routes remain protected.

## 4.1 `GET /livez`

### Purpose

Checks that the process is alive.

### Request

```http
GET /livez
```

### Response 200

```json
{
  "status": "ok",
  "service": "devops-control-plane"
}
```

### Rules

- Must be safe for unauthenticated access.
- Must not expose sensitive configuration.
- Should not depend on deep external integrations.

## 4.2 `GET /readyz`

### Purpose

Checks that the service is ready to receive traffic.

Minimum checks:

- configuration loaded;
- PostgreSQL connectivity available.

### Request

```http
GET /readyz
```

### Response 200

```json
{
  "status": "ready",
  "checks": {
    "database": "ok",
    "configuration": "ok"
  }
}
```

### Response 503

```json
{
  "status": "not-ready",
  "checks": {
    "database": "failed",
    "configuration": "ok"
  }
}
```

---

## 5. Authentication and authorization behavior

## 5.1 Route-level authentication

The OpenShift Route is protected by OpenShift OAuth Proxy.

Expected behavior:

```text
GET /readyz anonymous -> 200 when healthy
GET /livez anonymous  -> 200 when healthy
GET /api/v1/changes anonymous through Route -> 403
GET /ui/dashboard anonymous through Route -> 403
```

## 5.2 Backend authorization

Backend AuthZ is authoritative.

The backend reads trusted identity headers from the OAuth Proxy boundary and can resolve OpenShift group membership.

Current roles:

```text
viewer
operator
approver
admin
```

Rules:

- unknown users are denied;
- unknown endpoints fail closed;
- actions require role-specific permission;
- health endpoints are explicitly public.

## 5.3 Authorization-aware APIs

Technical actions such as validation, approval, merge request merge and evidence collection must be checked by backend AuthZ.

Future environment-aware AuthZ will also evaluate:

```text
targetEnvironment
action
role
approvalPolicy
```

---

## 6. Applications API

## 6.1 `GET /api/v1/applications`

### Purpose

Returns the list of Argo CD Applications visible to the DevOps Control Plane.

### Query parameters

```text
project
namespace
syncStatus
healthStatus
limit
offset
```

### Request

```http
GET /api/v1/applications?limit=50&offset=0
```

### Response 200

```json
{
  "data": [
    {
      "name": "demo-go-color-app",
      "argocdNamespace": "openshift-gitops",
      "project": "default",
      "targetNamespace": "devops-ci-demo",
      "repoUrl": "https://gitlab.example.local/group/demo-app-gitops.git",
      "targetRevision": "main",
      "path": "apps/demo-go-color-app",
      "syncStatus": "Synced",
      "healthStatus": "Healthy",
      "currentRevision": "<revision>"
    }
  ],
  "meta": {
    "count": 1,
    "limit": 50,
    "offset": 0
  },
  "error": null
}
```

### Mapping

- Argo CD adapter;
- optional `applications` cache.

## 6.2 `GET /api/v1/applications/{name}`

### Purpose

Returns operational details for an Application.

### Request

```http
GET /api/v1/applications/demo-go-color-app
```

### Response 200

```json
{
  "data": {
    "name": "demo-go-color-app",
    "argocdNamespace": "openshift-gitops",
    "project": "default",
    "syncStatus": "Synced",
    "healthStatus": "Healthy",
    "currentRevision": "<revision>",
    "source": {
      "repoUrl": "https://gitlab.example.local/group/demo-app-gitops.git",
      "targetRevision": "main",
      "path": "apps/demo-go-color-app"
    },
    "destination": {
      "server": "https://kubernetes.default.svc",
      "namespace": "devops-ci-demo"
    },
    "conditions": [
      {
        "type": "OrphanedResourceWarning",
        "message": "Application has 5 orphaned resources"
      }
    ]
  },
  "meta": {},
  "error": null
}
```

### Error examples

```text
ARGOCD_APPLICATION_NOT_FOUND
ARGOCD_GET_FAILED
ARGOCD_AUTH_FAILED
```

## 6.3 Optional application subresources

The following endpoints are design targets and may be implemented incrementally:

```text
GET /api/v1/applications/{name}/resources
GET /api/v1/applications/{name}/history
GET /api/v1/applications/{name}/runtime
```

---

## 7. Changes API

## 7.1 `GET /api/v1/changes`

### Purpose

Returns the list of ChangeRequests.

### Query parameters

```text
applicationName
status
runtimeStatus
changeType
targetEnvironment
requestedBy
limit
offset
sort
```

### Response 200

```json
{
  "data": [
    {
      "id": "1d871b5c-8a7a-4c11-9420-d2bc070b4a12",
      "changeNumber": "CHG-2026-0006",
      "title": "Multi-developer dashboard validation",
      "applicationName": "demo-go-color-app",
      "targetEnvironment": "dev",
      "changeType": "standard",
      "status": "draft",
      "runtimeStatus": "EvidenceCollected",
      "requestedBy": "developer-e",
      "createdAt": "2026-07-03T12:00:00+02:00",
      "completedAt": null
    }
  ],
  "meta": {
    "count": 1,
    "limit": 50,
    "offset": 0
  },
  "error": null
}
```

### Notes

- `/ui/changes` uses similar data to show the full list.
- Dashboard recent changes shows only the latest five items.
- `requestedBy` must be visible for multi-developer readability.

## 7.2 `POST /api/v1/changes`

### Purpose

Creates a new ChangeRequest.

### Request

```json
{
  "title": "Multi-developer dashboard validation",
  "applicationName": "demo-go-color-app",
  "targetEnvironment": "dev",
  "changeType": "standard",
  "riskLevel": "medium",
  "requestedBy": "developer-a",
  "description": "Create a controlled ChangeRequest for dashboard validation",
  "payload": {}
}
```

### Response 201

```json
{
  "data": {
    "id": "1d871b5c-8a7a-4c11-9420-d2bc070b4a12",
    "changeNumber": "CHG-2026-0006",
    "applicationName": "demo-go-color-app",
    "targetEnvironment": "dev",
    "changeType": "standard",
    "status": "draft",
    "runtimeStatus": "Created"
  },
  "meta": {},
  "error": null
}
```

### Validation rules

Required or expected fields:

- `title`;
- `applicationName`;
- `changeType`;
- `requestedBy`;
- `targetEnvironment`, defaulted to `dev` where needed for compatibility.

Future environment validation must reject unknown or disabled environments fail-closed.

### Error examples

```text
VALIDATION_INVALID_REQUEST
WORKFLOW_CONFLICT_ACTIVE_CHANGE
AUTH_FORBIDDEN
```

## 7.3 `GET /api/v1/changes/{id}`

### Purpose

Returns ChangeRequest details, including status and useful references.

### Response 200

```json
{
  "data": {
    "id": "1d871b5c-8a7a-4c11-9420-d2bc070b4a12",
    "changeNumber": "CHG-2026-0006",
    "title": "Multi-developer dashboard validation",
    "applicationName": "demo-go-color-app",
    "targetEnvironment": "dev",
    "changeType": "standard",
    "riskLevel": "medium",
    "status": "draft",
    "runtimeStatus": "EvidenceCollected",
    "requestedBy": "developer-e",
    "description": "Create a controlled ChangeRequest for dashboard validation",
    "createdAt": "2026-07-03T12:00:00+02:00",
    "completedAt": null
  },
  "meta": {},
  "error": null
}
```

## 7.4 `GET /api/v1/changes/{id}/events`

### Purpose

Returns the ChangeRequest timeline.

### Response 200

```json
{
  "data": [
    {
      "eventType": "change_created",
      "previousStatus": null,
      "newStatus": "draft",
      "message": "ChangeRequest created",
      "source": "workflow",
      "createdAt": "2026-07-03T12:00:00+02:00"
    },
    {
      "eventType": "technical_step_completed",
      "previousStatus": "ValidationRunning",
      "newStatus": "ValidationSucceeded",
      "message": "Tekton validation succeeded",
      "source": "tekton",
      "payload": {
        "step": "ValidationSucceeded"
      },
      "createdAt": "2026-07-03T12:05:00+02:00"
    }
  ],
  "meta": {
    "count": 2
  },
  "error": null
}
```

---

## 8. Lifecycle action API

Lifecycle endpoints move the ChangeRequest through its business/process lifecycle.

```text
POST /api/v1/changes/{id}/submit
POST /api/v1/changes/{id}/approve
POST /api/v1/changes/{id}/start-execution
POST /api/v1/changes/{id}/complete-execution
POST /api/v1/changes/{id}/close
```

### Response example

```json
{
  "data": {
    "changeNumber": "CHG-2026-0006",
    "status": "submitted",
    "runtimeStatus": "EvidenceCollected"
  },
  "meta": {},
  "error": null
}
```

Rules:

- lifecycle status and runtime status must remain separate;
- invalid transitions must fail with clear errors;
- authorization is required for protected transitions such as approve.

---

## 9. Technical workflow action API

Technical workflow endpoints execute or check individual technical steps.

## 9.1 `POST /api/v1/changes/{id}/create-branch`

### Purpose

Creates a GitLab branch for the ChangeRequest.

### Response 200

```json
{
  "data": {
    "sourceBranch": "change/CHG-2026-0006",
    "runtimeStatus": "BranchCreated"
  },
  "meta": {},
  "error": null
}
```

## 9.2 `POST /api/v1/changes/{id}/update-files`

### Purpose

Updates GitOps files and creates a commit.

### Response 200

```json
{
  "data": {
    "runtimeStatus": "CommitCreated",
    "commitSha": "<sha>",
    "filesChanged": [
      "manifests/chg-2026-0006-control-plane.yaml"
    ]
  },
  "meta": {},
  "error": null
}
```

## 9.3 `POST /api/v1/changes/{id}/open-merge-request`

### Purpose

Opens a GitLab merge request.

### Response 200

```json
{
  "data": {
    "runtimeStatus": "MergeRequestOpened",
    "mergeRequestIid": 42,
    "mergeRequestUrl": "https://gitlab.example.local/group/repo/-/merge_requests/42"
  },
  "meta": {},
  "error": null
}
```

## 9.4 `POST /api/v1/changes/{id}/merge-request`

### Purpose

Merges the GitLab merge request where workflow policy and AuthZ allow it.

### Response 200

```json
{
  "data": {
    "runtimeStatus": "MergeRequestMerged",
    "mergeRequestIid": 42,
    "mergeCommitSha": "<sha>"
  },
  "meta": {},
  "error": null
}
```

## 9.5 `POST /api/v1/changes/{id}/validate`

### Purpose

Creates a Tekton `PipelineRun` for GitOps validation.

### Response 202

```json
{
  "data": {
    "runtimeStatus": "ValidationRunning",
    "pipelineRunName": "devops-cp-validate-chg-2026-0006-xxxxx"
  },
  "meta": {},
  "error": null
}
```

## 9.6 `POST /api/v1/changes/{id}/check-validation`

### Purpose

Checks Tekton validation result and collects diagnostics when terminal.

### Response 200 success

```json
{
  "data": {
    "runtimeStatus": "ValidationSucceeded",
    "pipelineRunName": "devops-cp-validate-chg-2026-0006-xxxxx",
    "status": "Succeeded",
    "diagnostics": {
      "failedTaskCount": 0,
      "failedTasks": [],
      "summary": "Tekton validation succeeded"
    }
  },
  "meta": {},
  "error": null
}
```

### Response 200 failure

```json
{
  "data": {
    "runtimeStatus": "ValidationFailed",
    "pipelineRunName": "devops-cp-validate-chg-2026-0001-xxxxx",
    "status": "Failed",
    "diagnostics": {
      "failedTaskCount": 1,
      "failedTasks": ["clone-repository"],
      "summary": "Tekton validation failed in task clone-repository"
    }
  },
  "meta": {},
  "error": null
}
```

## 9.7 `POST /api/v1/changes/{id}/check-deployment`

### Purpose

Checks Argo CD deployment state and maps sync/health to runtime status.

### Response 200

```json
{
  "data": {
    "runtimeStatus": "DeploymentSyncedHealthy",
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
  "meta": {},
  "error": null
}
```

## 9.8 `POST /api/v1/changes/{id}/collect-evidence`

### Purpose

Collects deployment evidence from Argo CD and Kubernetes/OpenShift.

### Response 202

```json
{
  "data": {
    "runtimeStatus": "EvidenceCollected",
    "evidenceType": "deployment",
    "summary": "Post-deployment evidence for demo-go-color-app"
  },
  "meta": {},
  "error": null
}
```

---

## 10. Evidence API

## 10.1 `GET /api/v1/changes/{id}/evidence`

### Purpose

Returns all evidence associated with a ChangeRequest.

### Response 200

```json
{
  "data": [
    {
      "id": "c77ce0b7-e24f-4095-9c9f-7b4cf3eb42b0",
      "evidenceType": "validation",
      "name": "tekton-validation-evidence",
      "summary": "Tekton validation succeeded",
      "sanitized": true,
      "createdAt": "2026-07-03T12:05:00+02:00"
    },
    {
      "id": "ed7b8be4-f3d5-46c6-b4bb-dc7445d6a9c9",
      "evidenceType": "deployment",
      "name": "deployment-evidence",
      "summary": "Post-deployment evidence for demo-go-color-app",
      "sanitized": true,
      "createdAt": "2026-07-03T12:10:00+02:00"
    }
  ],
  "meta": {
    "count": 2
  },
  "error": null
}
```

## 10.2 `GET /api/v1/changes/{id}/evidence/validation`

### Purpose

Returns validation evidence for a ChangeRequest.

## 10.3 `GET /api/v1/changes/{id}/evidence/deployment`

### Purpose

Returns deployment evidence for a ChangeRequest.

Evidence payloads must be sanitized.

---

## 11. Optional Git, validation and runtime query APIs

The following endpoints are useful design targets but can be implemented incrementally:

```text
GET /api/v1/changes/{id}/git
GET /api/v1/changes/{id}/validation
GET /api/v1/changes/{id}/validation/taskruns
GET /api/v1/changes/{id}/validation/logs
GET /api/v1/applications/{name}/runtime
GET /api/v1/applications/{name}/git/commits
```

When implemented, these endpoints must follow the same security and response rules.

---

## 12. Error model

## 12.1 HTTP status mapping

| HTTP | Usage |
|---|---|
| 200 | Operation succeeded or current state returned |
| 201 | Resource created |
| 202 | Asynchronous operation accepted or technical action started |
| 400 | Invalid request |
| 401 | Not authenticated |
| 403 | Not authorized |
| 404 | Resource not found |
| 409 | Workflow or external-resource conflict |
| 422 | Validation failed or invalid lifecycle transition |
| 500 | Internal error |
| 502 | External API error |
| 503 | Service not ready |
| 504 | External timeout |

## 12.2 Standard error response

```json
{
  "data": null,
  "meta": {
    "requestId": "req-123"
  },
  "error": {
    "code": "GITLAB_FILE_NOT_FOUND",
    "message": "The requested GitOps file does not exist in the repository or branch.",
    "technicalMessage": "404 File Not Found",
    "suggestedAction": "Verify project ID, branch and file path.",
    "recoverable": true
  }
}
```

## 12.3 Main error codes

```text
VALIDATION_INVALID_REQUEST
VALIDATION_SECRET_DETECTED
AUTH_UNAUTHENTICATED
AUTH_FORBIDDEN
WORKFLOW_CONFLICT_ACTIVE_CHANGE
GITLAB_FILE_NOT_FOUND
GITLAB_BRANCH_EXISTS
GITLAB_COMMIT_FAILED
GITLAB_MR_CREATE_FAILED
GITLAB_MR_MERGE_FAILED
ARGOCD_APPLICATION_NOT_FOUND
ARGOCD_OPERATION_IN_PROGRESS
ARGOCD_RESOURCE_NOT_PERMITTED
ARGOCD_SYNC_FAILED
TEKTON_PIPELINERUN_FAILED
TEKTON_PIPELINERUN_TIMEOUT
KUBERNETES_RUNTIME_NOT_READY
DATABASE_ERROR
```

---

## 13. Pagination, sorting and filtering

## 13.1 Pagination

Standard parameters:

```text
limit
offset
```

Recommended defaults:

```text
limit=50
offset=0
```

Recommended maximum:

```text
limit=200
```

## 13.2 Sorting

Parameter:

```text
sort
```

Examples:

```text
sort=createdAt:desc
sort=name:asc
```

## 13.3 Filtering

Supported filters should be added incrementally based on actual UI and operational needs.

Common filters:

```text
applicationName
targetEnvironment
status
runtimeStatus
requestedBy
changeType
```

---

## 14. Request ID and correlation

Every request should have a request ID.

Optional client header:

```http
X-Request-ID: req-123
```

If the client does not provide one, the backend may generate one.

The request ID should appear in:

- logs;
- error responses;
- optional diagnostic events.

---

## 15. API security requirements

- Do not expose tokens.
- Do not expose Secret values.
- Do not expose raw kubeconfig.
- Sanitize evidence payloads.
- Enforce backend AuthZ.
- Keep health endpoints public only by explicit policy.
- Deny unknown or unclassified endpoints fail-closed.
- Use timeouts for external calls.
- Return safe technical messages only.
- Keep UI actions mapped to backend-authorized actions.

---

## 16. Web UI relationship

The Web UI is server-side rendered and complements the API.

Current UI routes include:

```text
/
/ui
/ui/dashboard
/ui/changes
/ui/changes/{id}
/ui/changes-api
```

UI action handlers call backend services and should remain aligned with the API technical actions.

The UI must not bypass:

- validation;
- workflow rules;
- AuthZ;
- evidence sanitization.

---

## 17. Multi-developer API behavior

The API must support multi-developer visibility.

Requirements:

- `requestedBy` is included in ChangeRequest responses;
- `/api/v1/changes` returns the full list subject to filtering/pagination;
- dashboard can use recent changes limited to five items;
- full history remains available;
- evidence and events remain associated with the specific ChangeRequest.

---

## 18. Multi-environment API direction

Future environment-aware API behavior must include `targetEnvironment` consistently.

Canonical environments:

```text
dev
staging
production
```

Future request and response objects should include:

```text
targetEnvironment
promotionGroupID
promotedFromChangeNumber
```

Future environment-aware APIs should:

- reject unknown environments fail-closed;
- disable production until guardrails are implemented;
- enforce environment-specific AuthZ;
- resolve GitLab, Tekton, Argo CD and Kubernetes context from environment configuration.

---

## 19. Implementation order

Recommended implementation order, updated to current baseline:

1. `GET /livez`;
2. `GET /readyz`;
3. `GET /api/v1/changes`;
4. `POST /api/v1/changes`;
5. `GET /api/v1/changes/{id}`;
6. lifecycle action endpoints;
7. GitLab technical action endpoints;
8. Tekton validation endpoints;
9. Argo CD deployment check endpoint;
10. evidence collection endpoint;
11. evidence read endpoints;
12. Application read endpoints;
13. optional runtime and diagnostic endpoints;
14. OpenAPI specification.

---

## 20. Go package mapping

Representative package mapping:

```text
internal/api/router.go
internal/api/health_handlers.go
internal/api/application_handlers.go
internal/api/change_handlers.go
internal/api/ui_handlers.go
internal/api/errors.go
internal/api/response.go
```

Service and integration logic lives outside handlers:

```text
internal/app/
internal/domain/
internal/adapters/
internal/database/
internal/config/
```

Handlers should orchestrate request parsing, AuthZ context and service calls, not low-level external API operations.

---

## 21. API readiness checklist

The API baseline is ready when:

- `/livez` and `/readyz` work;
- `/api/v1` is the canonical API prefix;
- response envelope is consistent;
- error response is consistent;
- ChangeRequest creation works;
- ChangeRequest list and detail work;
- lifecycle actions are available;
- technical workflow steps are invokable;
- validation and deployment checks are observable;
- evidence is retrievable;
- input validation returns clear errors;
- authorization failures return 403;
- tokens and secrets are never returned;
- request logging is safe;
- UI actions map to backend services;
- documentation and implementation remain aligned.

---

## 22. Relationship with other documents

This document informs and is informed by:

- `internal/api/`;
- `internal/app/`;
- `internal/domain/`;
- `docs/10-data-model.md`;
- `docs/11-change-workflows.md`;
- `docs/12-evidence-model.md`;
- `docs/09-security-rbac.md`;
- `docs/environment-configuration-model.md`;
- `docs/change-promotion-model.md`;
- future OpenAPI specification.

---

## 23. Key message

The API must make the GitOps workflow clear, controllable and safe.

Each API should help users or the UI see or advance one step:

```text
Application discovery
Change creation
Lifecycle transition
Git change
Tekton validation
Argo CD deployment check
Runtime evidence collection
Evidence review
```

The design must remain simple, consistent, secure and explainable to people learning GitOps, Tekton, Argo CD and OpenShift.

This preserves the original API-first spirit of the project while aligning the endpoints with the current advanced MVP baseline.

---

## 24. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial API design document in Italian. |
| 2026-07-06 | 0.2 | Rewritten in English and refreshed while preserving the original API-first, JSON-first and workflow-driven API intent and aligning it with the current `/api/v1`, AuthZ, UI, validation, deployment and evidence baseline. |
