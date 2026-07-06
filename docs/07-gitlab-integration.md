# DevOps Control Plane — GitLab Integration

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 07 — GitLab Integration
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
- **Status:** Rewritten in English and refreshed while preserving the original GitOps-oriented integration intent
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document describes how the DevOps Control Plane integrates with GitLab.

The original document defined GitLab as the system that hosts the target GitOps repositories and as the source of declarative changes later reconciled by Argo CD. This refreshed version preserves that original architectural intent while updating the document to match the current implementation baseline.

The document defines:

- GitLab responsibilities in the DevOps Control Plane architecture;
- responsibilities of the GitLab adapter;
- branch, file update, commit and merge request workflows;
- naming conventions;
- authentication and token handling;
- normalized data model;
- error handling;
- evidence expectations;
- security requirements;
- current `/api/v1` endpoints;
- relationship with ChangeRequests, Tekton validation, Argo CD deployment checks and future environment-aware GitOps paths.

GitLab is the system that keeps GitOps changes explicit, reviewable, auditable and reversible.

---

## 2. Role of GitLab in the architecture

GitLab has the role of Git provider and change collaboration platform.

GitLab is responsible for:

- hosting target GitOps repositories;
- exposing repository files;
- managing branches;
- storing commits;
- managing merge requests;
- keeping Git history;
- providing auditable links to branches, commits and merge requests.

The DevOps Control Plane is responsible for:

- using GitLab API through a dedicated adapter;
- resolving the target project, branch and path;
- creating a ChangeRequest branch;
- updating GitOps files in a controlled way;
- creating commits;
- opening merge requests where review is required;
- merging merge requests where policy and authorization allow it;
- correlating GitLab references with ChangeRequests, Tekton validation, Argo CD deployment checks and runtime evidence;
- preventing tokens or secrets from being committed to Git.

---

## 3. Integration principles

### 3.1 GitLab is the source of declarative change

The DevOps Control Plane must produce Git changes, not permanent runtime mutations.

The intended flow is:

```text
ChangeRequest
  -> GitLab branch
  -> GitOps file update
  -> commit or merge request
  -> Tekton validation
  -> merge into target branch where required
  -> Argo CD deployment state check
  -> OpenShift runtime evidence
  -> PostgreSQL audit and evidence history
```

This preserves the original project rule:

```text
Every permanent desired-state change must pass through Git.
```

### 3.2 DevOps Control Plane uses GitLab API

The product integration must use GitLab REST API through a Go adapter.

The `git` CLI remains useful for development, troubleshooting and manual validation, but the application runtime workflow must not depend on local shell Git commands.

### 3.3 Every change must be traceable

Each ChangeRequest should retain GitLab-related references such as:

- project ID;
- repository URL;
- source branch;
- target branch;
- commit SHA;
- merge request IID and URL, when present;
- requester;
- logical actor;
- commit message;
- merge request title;
- merge result;
- timestamps;
- related technical events.

### 3.4 Branch plus merge request remains the preferred model

The original document allowed two modes:

```text
Mode A — Direct commit to target branch
Mode B — Branch plus merge request
```

The preferred model remains:

```text
Branch plus merge request
```

Rationale:

- clearer review;
- better auditability;
- safer production-oriented workflow;
- explicit separation between proposal and application;
- better alignment with Tekton validation before merge;
- clearer rollback through revert commit or revert merge request.

Direct commit can remain useful only for controlled lab or bootstrap scenarios, if explicitly allowed by policy.

---

## 4. Current implementation baseline

The GitLab integration currently includes:

- GitLab REST client;
- branch creation workflow;
- GitOps file update workflow;
- open merge request workflow;
- merge request merge workflow;
- external reference persistence through ChangeRequest events;
- runtime status updates such as `BranchCreated`, `CommitCreated`, `MergeRequestOpened` and `MergeRequestMerged`;
- integration with Tekton validation;
- integration with downstream Argo CD deployment checks and evidence collection;
- token handling through OpenShift Secret;
- TLS strict baseline for GitLab access.

The implemented workflow preserves the original intent of making GitLab the auditable source of GitOps change.

---

## 5. GitLab repository configuration

## 5.1 Required repository metadata

A managed application must be associated with GitLab repository metadata.

Example baseline:

```yaml
applicationName: demo-go-color-app
repoProvider: gitlab
gitlabBaseUrl: https://gitlab.example.local
gitlabProjectId: "12345"
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
defaultBranch: main
path: apps/demo-go-color-app
targetEnvironment: dev
```

### Rules

- `gitlabProjectId` may be a numeric ID or URL-encoded project path depending on GitLab API usage.
- `defaultBranch` must be configurable.
- `path` must align with the GitOps path used by Argo CD.
- Future environment-aware behavior must resolve GitOps path from the environment catalog.

---

## 6. GitLab adapter

## 6.1 Responsibilities

The GitLab adapter encapsulates all interactions with GitLab API.

The rest of the application must not know raw GitLab endpoints, token headers, file encoding details or raw GitLab error payloads.

Current responsibilities:

- create branch;
- create or update repository files;
- open merge request;
- merge merge request;
- map GitLab responses to internal models;
- normalize GitLab errors;
- support TLS strict mode.

Future responsibilities may include richer read-only repository metadata, diff retrieval and environment-aware path resolution.

## 6.2 Conceptual interface

Conceptual adapter interface:

```go
type GitLabAdapter interface {
    CreateBranch(ctx context.Context, projectID int, branch string, ref string) error
    CreateOrUpdateFile(ctx context.Context, projectID int, branch string, filePath string, commitMessage string, content string) error
    OpenMergeRequest(ctx context.Context, projectID int, sourceBranch string, targetBranch string, title string, description string) (int, string, error)
    MergeRequest(ctx context.Context, projectID int, sourceBranch string, targetBranch string, mergeCommitMessage string) (int, string, string, error)
}
```

The exact implementation can evolve, but services should depend on ports rather than raw HTTP implementation details.

## 6.3 Package location

Current package area:

```text
internal/adapters/gitlab/
```

Expected responsibilities:

```text
client.go           -> HTTP client, auth header, base URL, timeout
models.go           -> GitLab DTOs and normalized models
branches.go         -> branch operations
files.go            -> file create/update operations
merge_requests.go   -> merge request operations
errors.go           -> GitLab error mapping
```

---

## 7. Branch workflow

## 7.1 Branch naming

The branch should be derived from the ChangeRequest number.

Current practical convention:

```text
change/<change-number>
```

Example:

```text
change/CHG-2026-0006
```

The original extended convention remains valid as a future option:

```text
change/<change-number>-<change-type>
```

The current shorter form is easier to correlate and avoids overly long branch names.

## 7.2 Branch creation flow

```text
ChangeRequest
  -> build branch name
  -> GitLab adapter CreateBranch(projectId, branch, ref=main)
  -> record technical event BranchCreated
  -> update runtime status
```

### Rules

- The branch starts from the configured target branch or reference.
- The branch name must be normalized.
- Existing branch conflicts must produce a readable error.
- The branch reference must be recorded in the ChangeRequest event history.

---

## 8. File update and commit workflow

## 8.1 File update purpose

The Control Plane updates GitOps files as part of a ChangeRequest technical workflow.

Examples:

- generated control-plane marker file;
- deployment manifest update;
- ConfigMap update;
- Kustomize overlay update in the future environment-aware model.

## 8.2 Commit message convention

Commit messages should include the ChangeRequest number.

Example:

```text
CHG-2026-0006 Update demo-go-color-app GitOps files
```

For generated workflow commits, concise messages are acceptable if the ChangeRequest event carries additional context.

## 8.3 Rules

- The commit must include ChangeRequest context.
- Generated content must not contain tokens or secrets.
- Anti-secret validation must be performed before commit or through Tekton validation before merge.
- The commit SHA or resulting GitLab reference must be recorded where available.

---

## 9. Merge request workflow

## 9.1 Merge request creation

The Control Plane can open a merge request from the ChangeRequest branch to the target branch.

Example title:

```text
[CHG-2026-0006] Update demo-go-color-app via DevOps Control Plane
```

Recommended description sections:

```text
Change Summary
Application
Target Environment
Requested By
Files Changed
Validation Status
GitOps Flow
Rollback
```

## 9.2 Merge request merge

The Control Plane can merge a merge request when policy and authorization allow it.

The merge result should record:

- merge request IID;
- merge request URL;
- merge commit SHA;
- target branch;
- runtime status `MergeRequestMerged`;
- technical event payload.

## 9.3 Rules

- Merge request source branch is the ChangeRequest branch.
- Merge request target branch is the configured target branch, usually `main`.
- Merge action must be authorized.
- Merge result must be auditable.
- Production merge behavior must remain guarded by future environment-aware AuthZ.

---

## 10. Normalized GitLab data model

## 10.1 Git repository

```yaml
provider: gitlab
baseUrl: https://gitlab.example.local
projectId: "12345"
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
defaultBranch: main
```

## 10.2 Git branch

```yaml
name: change/CHG-2026-0006
ref: main
webUrl: https://gitlab.example.local/group/repo/-/tree/change/CHG-2026-0006
```

## 10.3 Git commit

```yaml
id: <full-sha>
shortId: <short-sha>
title: CHG-2026-0006 Update demo-go-color-app GitOps files
webUrl: https://gitlab.example.local/group/repo/-/commit/<sha>
```

## 10.4 Merge request

```yaml
iid: 42
id: 123456
title: "[CHG-2026-0006] Update demo-go-color-app via DevOps Control Plane"
state: opened
sourceBranch: change/CHG-2026-0006
targetBranch: main
webUrl: https://gitlab.example.local/group/repo/-/merge_requests/42
mergeStatus: can_be_merged
```

---

## 11. Authentication and token handling

## 11.1 Token-based authentication

Configuration baseline:

```text
GITLAB_BASE_URL=https://gitlab.example.local
GITLAB_TOKEN=<from Secret>
GITLAB_TIMEOUT_SECONDS=30
GITLAB_DEFAULT_PROJECT_ID=
GITLAB_DEFAULT_BRANCH=main
GITLAB_INSECURE_TLS=false
GITLAB_CA_FILE=/etc/dcp-trust/ca-bundle.crt
```

### Rules

- `GITLAB_TOKEN` must come from OpenShift Secret.
- The token must not be logged.
- The token must not be returned by API responses.
- The token must not be persisted in PostgreSQL.
- The token must not be included in evidence payloads.
- HTTP 401 and 403 errors must be normalized.

## 11.2 Token type

For lab usage, a technical token may be used. For production-oriented use, prefer a governed token with the minimum permissions compatible with the workflow.

Possible token types:

- Project Access Token;
- Group Access Token;
- service account or bot user token;
- Personal Access Token only for controlled lab or development use.

### Security rule

Use the most restrictive token scope that supports the required workflow.

---

## 12. TLS model

GitLab integration must support TLS strict mode.

Baseline:

```text
GITLAB_INSECURE_TLS=false
```

If the GitLab Route or endpoint uses a CA not available in the default trust store, the application trust bundle must include the required CA.

Rules:

- insecure TLS must not be used as the production-oriented default;
- CA bundle configuration must be explicit;
- TLS errors must be readable and must not expose token values.

---

## 13. ChangeRequest to GitLab mapping

A ChangeRequest should be correlated with GitLab references through events and payloads.

Example logical mapping:

```yaml
changeNumber: CHG-2026-0006
applicationName: demo-go-color-app
targetEnvironment: dev
changeType: standard
git:
  provider: gitlab
  projectId: "12345"
  repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
  sourceBranch: change/CHG-2026-0006
  targetBranch: main
  commitSha: <sha>
  mergeRequestIid: 42
  mergeRequestUrl: https://gitlab.example.local/group/repo/-/merge_requests/42
```

Current implementation stores most of this information through technical events and responses rather than a separate rich Git table.

---

## 14. GitLab-related events

Recommended event names:

```text
BranchCreated
CommitCreated
MergeRequestOpened
MergeRequestMerged
GitOperationFailed
```

Events should include safe payload fields such as:

- project ID;
- branch;
- target branch;
- commit SHA;
- merge request IID;
- merge request URL;
- actor;
- target environment.

Events must not include token values.

---

## 15. File modification strategy

## 15.1 Controlled file modifications

The Control Plane should modify GitOps files through controlled logic rather than arbitrary text editing.

Preferred approaches:

1. structured YAML parsing and update;
2. targeted patch logic for known objects;
3. controlled generated files for lab and validation workflows.

### Rule

Avoid fragile text manipulation when the Workflow is intended for production-oriented usage.

## 15.2 Diff and review

The workflow should make changes reviewable through GitLab commits and merge requests.

Diffs must not include secrets.

## 15.3 Concurrency

The system should eventually detect conflicting active ChangeRequests affecting the same application, target environment or GitOps path.

This is especially important before enabling production workflows.

---

## 16. GitLab error model

## 16.1 Recommended error codes

```text
GITLAB_AUTH_FAILED
GITLAB_FORBIDDEN
GITLAB_PROJECT_NOT_FOUND
GITLAB_FILE_NOT_FOUND
GITLAB_REF_NOT_FOUND
GITLAB_BRANCH_EXISTS
GITLAB_BRANCH_CREATE_FAILED
GITLAB_COMMIT_FAILED
GITLAB_MR_CREATE_FAILED
GITLAB_MR_MERGE_FAILED
GITLAB_MR_NOT_FOUND
GITLAB_RATE_LIMITED
GITLAB_TIMEOUT
GITLAB_UNKNOWN_ERROR
```

## 16.2 Normalized error example

```yaml
code: GITLAB_FILE_NOT_FOUND
message: GitLab file not found
technicalMessage: 404 File Not Found
recoverable: true
suggestedAction: Verify project ID, branch and file path.
```

## 16.3 Branch already exists

If the branch already exists, the system should:

- reuse it only if clearly associated with the same ChangeRequest and policy allows it;
- otherwise fail with a readable conflict error;
- avoid silently writing to an unrelated branch.

---

## 17. GitLab evidence

## 17.1 Minimal evidence

```yaml
provider: gitlab
projectId: "12345"
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
sourceBranch: change/CHG-2026-0006
targetBranch: main
commitSha: <sha>
mergeRequestIid: 42
mergeRequestUrl: https://gitlab.example.local/group/repo/-/merge_requests/42
filesChanged:
  - apps/demo-go-color-app/deployment.yaml
```

## 17.2 Error evidence

```yaml
provider: gitlab
operation: CreateBranch
errorCode: GITLAB_BRANCH_EXISTS
message: Branch change/CHG-2026-0006 already exists
```

Evidence must be sanitized and must not include token values.

---

## 18. API endpoints related to GitLab workflows

The current Control Plane APIs use `/api/v1`.

Change workflow actions:

```text
POST /api/v1/changes/{id}/create-branch
POST /api/v1/changes/{id}/update-files
POST /api/v1/changes/{id}/open-merge-request
POST /api/v1/changes/{id}/merge-request
GET  /api/v1/changes/{id}
GET  /api/v1/changes/{id}/events
GET  /api/v1/changes/{id}/evidence
```

Legacy endpoint references without `/api/v1` should be considered historical and updated during documentation migration.

---

## 19. Integration with Tekton and Argo CD

GitLab workflow is not the end of the change lifecycle.

The GitLab branch, commit and merge request are followed by:

```text
Tekton validation
  -> Argo CD deployment check
  -> Kubernetes/OpenShift runtime evidence
  -> PostgreSQL evidence and audit history
```

Tekton validates the GitOps content. Argo CD reflects deployment state. OpenShift evidence verifies runtime behavior.

---

## 20. Multi-environment direction

Future environment-aware behavior will resolve GitOps path and target branch from the environment catalog.

Example mapping:

```text
demo-go-color-app + dev        -> apps/demo-go-color-app/overlays/dev
demo-go-color-app + staging    -> apps/demo-go-color-app/overlays/staging
demo-go-color-app + production -> apps/demo-go-color-app/overlays/production
```

Production workflows must remain disabled until production guardrails, approval policy and environment-aware AuthZ are implemented and validated.

---

## 21. Security requirements specific to GitLab

- Do not log `GITLAB_TOKEN`.
- Do not return token values through API.
- Do not persist token values in PostgreSQL.
- Do not include token values in evidence.
- Use token with least required privileges.
- Prefer project, group or service account tokens over personal tokens for production-oriented usage.
- Define token expiration and rotation procedures.
- Keep `.env` out of Git.
- Run anti-secret validation before merge.
- Use TLS strict mode.

---

## 22. Testing strategy

## 22.1 Unit tests

Test:

- branch naming;
- path normalization;
- GitLab response mapping;
- merge request mapping;
- error mapping;
- token redaction behavior where applicable.

## 22.2 Fake GitLab adapter tests

Workflow services must be testable without real GitLab.

Example fake flow:

```text
CreateBranch -> success
CreateOrUpdateFile -> success
OpenMergeRequest -> returns IID and URL
MergeRequest -> returns merge commit SHA
```

## 22.3 Runtime validation

Runtime validation should cover:

- create branch from a ChangeRequest;
- update a file in the ChangeRequest branch;
- open a merge request;
- merge a merge request;
- verify ChangeRequest events;
- verify GitLab references in responses and evidence;
- verify token values are not printed.

## 22.4 Failure scenarios

Scenarios to validate or simulate:

- invalid token;
- wrong project ID;
- branch already exists;
- file update failure;
- merge request creation failure;
- merge request merge failure;
- timeout;
- TLS trust error.

---

## 23. Completion checklist

The GitLab integration baseline is considered ready when:

- GitLab configuration is loaded safely;
- token is read from Secret and not logged;
- TLS strict mode works;
- a ChangeRequest branch can be created;
- GitOps files can be updated;
- a merge request can be opened;
- a merge request can be merged where authorized;
- GitLab external references are stored in events or evidence;
- errors are normalized;
- GitLab workflow integrates with Tekton validation and Argo CD deployment checks;
- runtime validation is documented.

---

## 24. Relationship with other documents

This document informs and is informed by:

- `docs/05-architecture.md`;
- `docs/08-tekton-integration.md`;
- `docs/06-argocd-integration.md`;
- `docs/13-api-design.md`;
- `docs/04-non-functional-requirements.md`;
- `docs/environment-configuration-model.md`;
- `docs/change-promotion-model.md`;
- `docs/adr/ADR-0007-gitlab-api-as-git-provider.md`;
- `docs/runbooks/secrets-rotation.md`.

---

## 25. Key message

The GitLab integration is what keeps the DevOps Control Plane aligned with GitOps.

The system must not change the cluster directly as the desired state. The system must produce tracked Git changes:

```text
branch
  -> file update
  -> commit
  -> merge request
  -> validation
  -> merge
  -> Argo CD deployment state
```

This preserves the original spirit of the project: every change should be readable, reviewable, auditable and reversible.

---

## 26. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial GitLab integration document in Italian. |
| 2026-07-06 | 0.2 | Rewritten in English and refreshed while preserving the original GitOps-oriented GitLab integration model and aligning it with the implemented advanced MVP baseline. |
