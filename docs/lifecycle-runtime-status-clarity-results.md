# Lifecycle and Runtime Status Clarity Results

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** Lifecycle and runtime status clarity results
- **Phase:** 14.3.3 — Document lifecycle/runtime status UI/API clarity refinement
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Inventory baseline runtime image:** `devops-control-plane:399f829`
- **Validated runtime image:** `devops-control-plane:8362628`
- **Code commit:** `8362628` — `Clarify lifecycle and runtime status wording in UI`
- **Previous related phases:**
  - 14.3.1 — Lifecycle vs runtime status UI/API clarity inventory
  - 14.3.2 — Refine lifecycle/runtime status wording in UI
- **Status:** Completed validation report
- **Language:** English

---

## 1. Purpose

This document records the outcome of Phase 14.3, which reviewed and refined how the DevOps Control Plane presents the difference between the ChangeRequest process lifecycle status and the technical runtime status.

The goal of this phase was to reduce ambiguity for users and operators when a ChangeRequest has different lifecycle and runtime states at the same time.

A representative and valid example is:

```text
status=draft
runtimeStatus=EvidenceCollected
```

This means:

```text
The process lifecycle is still draft.
The latest technical runtime observation is EvidenceCollected.
```

This document explains the inventory, the wording refinement, the runtime validation and the final behavior.

---

## 2. Scope

This validation covered:

- API response behavior for `status` and `runtimeStatus`;
- UI ChangeRequest detail rendering;
- UI ChangeRequest list rendering;
- wording clarity for lifecycle versus runtime state;
- runtime validation after deployment of the UI refinement.

This phase did not change the underlying lifecycle state machine.

This phase did not change technical workflow behavior.

This phase did not change database schema or API field names.

---

## 3. Background

The DevOps Control Plane intentionally separates two concepts:

```text
Process lifecycle status
Technical runtime status
```

The process lifecycle status tracks governance and workflow approval state, for example:

```text
draft
submitted
approved
executing
executed
closed
```

The technical runtime status tracks the latest technical automation, validation, deployment or evidence observation, for example:

```text
BranchCreated
CommitCreated
MergeRequestOpened
MergeRequestMerged
ValidationRunning
ValidationSucceeded
ValidationFailed
DeploymentSyncedHealthy
EvidenceCollected
```

This distinction is important because a ChangeRequest can be technically advanced while its governance lifecycle remains in draft.

---

## 4. Phase 14.3.1 — Inventory

### 4.1 Repository and runtime baseline

Repository HEAD at inventory time:

```text
1cb1295 Document target environment validation results
```

Runtime image at inventory time:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:399f829
```

Backend port-forward was opened on port `18090`.

Readiness result:

```text
/readyz -> HTTP 200
```

### 4.2 API inventory

Endpoint tested:

```text
GET /api/v1/changes/CHG-2026-0001
```

Observed HTTP result:

```text
HTTP 200
```

Observed fields:

```text
changeNumber=CHG-2026-0001
status=draft
runtimeStatus=EvidenceCollected
applicationName=demo-go-color-app
targetEnvironment=dev
requestedBy=vmarzario
```

Conclusion:

```text
The API already exposes both status and runtimeStatus.
```

The API field names were not changed in this phase because they are stable contract fields and are already documented.

### 4.3 UI detail inventory

Endpoint tested:

```text
GET /ui/changes/CHG-2026-0001
```

Observed HTTP result:

```text
HTTP 200
```

Observed labels and values before the refinement:

```text
Change Request:=True
Lifecycle status=True
Runtime Status=True
draft=True
EvidenceCollected=True
Application=True
Requester=True
Environment=True
```

Conclusion:

```text
The UI detail page already showed both lifecycle and runtime status, but the labels could be more explicit.
```

### 4.4 UI list inventory

Endpoint tested:

```text
GET /ui/changes
```

Observed HTTP result:

```text
HTTP 200
```

Observed labels and values before the refinement:

```text
Change Requests=True
Lifecycle=True
Runtime=True
Requested by=True
Environment=True
CHG-2026-0001=True
EvidenceCollected=True
```

Conclusion:

```text
The UI list page already showed both columns, but the column names were short and could be clearer.
```

---

## 5. Phase 14.3.2 — UI wording refinement

### 5.1 Commit

Code commit:

```text
8362628 Clarify lifecycle and runtime status wording in UI
```

### 5.2 File changed

```text
internal/api/ui_handlers.go
```

### 5.3 Wording changes

The UI detail page labels were updated as follows:

```text
Lifecycle status -> Process lifecycle status
Runtime Status   -> Technical runtime status
```

The UI list page column headers were updated as follows:

```text
Lifecycle -> Process lifecycle
Runtime   -> Technical runtime
```

### 5.4 Explanatory section added

The ChangeRequest detail page now includes a `Status meaning` section.

The explanatory text is:

```text
Process lifecycle status tracks the governance and approval state of the ChangeRequest. Technical runtime status tracks the latest automation, validation, deployment or evidence observation.
```

### 5.5 Code validation

Validation completed successfully before commit:

```text
gofmt OK
go test ./... OK
git diff --check OK
```

### 5.6 Repository state

After commit and push:

```text
HEAD -> main
origin/main
working tree clean
```

---

## 6. Build, push and deployment

### 6.1 Image build and deployment

Image tag:

```text
8362628
```

Runtime image after rollout:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:8362628
```

Deployment rollout completed successfully:

```text
deployment "devops-control-plane" successfully rolled out
```

### 6.2 Runtime check

Backend port-forward was opened on port `18091`.

Readiness result:

```text
/readyz -> HTTP 200
```

---

## 7. Runtime validation after refinement

### 7.1 UI detail page validation

Endpoint tested:

```text
GET /ui/changes/CHG-2026-0001
```

Observed HTTP result:

```text
HTTP 200
```

Observed labels and values:

```text
Change Request:=True
Process lifecycle status=True
Technical runtime status=True
Status meaning=True
Process lifecycle status tracks the governance and approval state=True
Technical runtime status tracks the latest automation=True
draft=True
EvidenceCollected=True
```

Conclusion:

```text
The ChangeRequest detail page now clearly explains and displays both process lifecycle and technical runtime state.
```

### 7.2 UI list page validation

Endpoint tested:

```text
GET /ui/changes
```

Observed HTTP result:

```text
HTTP 200
```

Observed labels and values:

```text
Change Requests=True
Process lifecycle=True
Technical runtime=True
Requested by=True
Environment=True
CHG-2026-0001=True
EvidenceCollected=True
```

Conclusion:

```text
The ChangeRequest list now uses clearer column names for lifecycle and runtime state.
```

### 7.3 Runtime validation summary

```text
readyz=200
ui_change_detail_http=200
ui_changes_list_http=200
```

---

## 8. Final behavior

The API continues to expose stable field names:

```text
status
runtimeStatus
```

The UI now presents these fields with clearer user-facing terminology:

```text
Process lifecycle status
Technical runtime status
```

For the representative ChangeRequest `CHG-2026-0001`, the UI now communicates the following without ambiguity:

```text
Process lifecycle status = draft
Technical runtime status = EvidenceCollected
```

This makes it clear that the ChangeRequest can remain in a governance draft state while also having completed technical evidence collection.

---

## 9. Design rationale

The refinement intentionally changes UI wording only.

The API field names were preserved because:

- `status` and `runtimeStatus` are already part of the API contract;
- documentation already explains the distinction;
- changing API field names would introduce unnecessary compatibility risk;
- UI wording is the most appropriate place to improve readability for users.

The selected wording is explicit:

```text
Process lifecycle
Technical runtime
```

This helps both experienced operators and new users understand the separation between governance process and technical automation state.

---

## 10. Residual notes

### 10.1 UI wording is now clearer but not a policy engine

The UI now explains the meaning of the two statuses, but status consistency rules remain governed by the backend service and domain model.

### 10.2 Possible future enhancement

A future UI enhancement could add contextual hints for each specific state, for example:

```text
draft: The ChangeRequest has not yet been submitted for approval.
EvidenceCollected: Runtime evidence has been collected after deployment checks.
```

This was intentionally not included in Phase 14.3 to keep the change small and low-risk.

---

## 11. Acceptance summary

Phase 14.3 is considered complete because:

- the current API and UI behavior was inventoried;
- the UI wording was refined;
- the distinction between process lifecycle and technical runtime state is now explicit;
- the code was tested, committed and pushed;
- the image was deployed;
- runtime validation confirmed the new UI wording;
- no API compatibility change was introduced.
