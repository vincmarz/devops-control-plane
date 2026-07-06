# Environment Visibility UI Results

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** Environment visibility UI results
- **Phase:** 14.4.3 — Document environment visibility UI results
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Inventory runtime image:** `devops-control-plane:8362628`
- **Validated runtime image:** `devops-control-plane:92ca49f`
- **Code commit:** `92ca49f` — `Show target environment in dashboard recent changes`
- **Previous related phases:**
  - 14.4.1 — Environment visibility in dashboard and UI inventory
  - 14.4.2 — Show target environment in dashboard recent changes
- **Status:** Completed validation report
- **Language:** English

---

## 1. Purpose

This document records the outcome of Phase 14.4, which validated and refined how the DevOps Control Plane Web UI displays `targetEnvironment` in the dashboard and ChangeRequest views.

The purpose of the phase was to ensure that users can clearly understand which environment a ChangeRequest targets, especially when the UI is used in preparation for the future multi-environment model.

The phase focused on visibility only.

This phase did not enable additional environments.

This phase did not change the fail-closed runtime validation introduced in Phase 14.2.

---

## 2. Scope

The validation covered the following UI areas:

- `/ui/dashboard`;
- `/ui/changes`;
- `/ui/changes/CHG-2026-0010`;
- `/ui/changes/CHG-2026-0011`;
- `/ui/changes/CHG-2026-0012`.

The target records used during validation were:

```text
CHG-2026-0010 -> targetEnvironment=dev
CHG-2026-0011 -> targetEnvironment=unknown-env
CHG-2026-0012 -> targetEnvironment=production
```

Important note:

```text
CHG-2026-0011 and CHG-2026-0012 are historical records created before fail-closed targetEnvironment validation was introduced.
```

They are useful for UI visibility testing, but they do not mean that `unknown-env` or `production` are currently enabled runtime environments.

---

## 3. Background

The DevOps Control Plane now validates `targetEnvironment` fail-closed at runtime.

Current runtime behavior after Phase 14.2 is:

```text
missing targetEnvironment -> default dev
targetEnvironment=dev -> allowed
targetEnvironment=unknown-env -> rejected
targetEnvironment=production -> rejected
targetEnvironment=staging -> rejected until explicitly enabled
```

The UI must therefore distinguish two separate concerns:

```text
1. Runtime validation of newly created ChangeRequests.
2. Read-only display of historical ChangeRequest records already stored in PostgreSQL.
```

The goal of Phase 14.4 was to make sure that historical and current ChangeRequests display their `targetEnvironment` clearly in the UI, without implying that non-enabled environments are currently available for new changes.

---

## 4. Phase 14.4.1 — Inventory current environment rendering

### 4.1 Repository and runtime baseline

Repository HEAD at inventory time:

```text
612947f Document lifecycle and runtime status clarity results
```

Runtime image at inventory time:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:8362628
```

Backend port-forward was opened on port `18092`.

Readiness result:

```text
/readyz -> HTTP 200
```

### 4.2 HTTP results

The following UI routes were retrieved successfully:

```text
/ui/dashboard -> HTTP 200
/ui/changes -> HTTP 200
/ui/changes/CHG-2026-0010 -> HTTP 200
/ui/changes/CHG-2026-0011 -> HTTP 200
/ui/changes/CHG-2026-0012 -> HTTP 200
```

Inventory summary:

```text
readyz=200
dashboard_http=200
changes_http=200
detail_0010_http=200
detail_0011_http=200
detail_0012_http=200
```

---

## 5. Dashboard inventory before refinement

### 5.1 Dashboard check

The dashboard was checked for recent changes and environment values.

Observed result before refinement:

```text
Recent changes=True
View all=True
CHG-2026-0010=False
CHG-2026-0011=True
CHG-2026-0012=True
dev=True
unknown-env=False
production=False
Environment=True
```

### 5.2 Interpretation

The dashboard was available and displayed recent ChangeRequests.

However, the environment value was not rendered explicitly in the Recent changes row.

The evidence was:

```text
CHG-2026-0011=True
CHG-2026-0012=True
unknown-env=False
production=False
```

This means the historical records were present in the dashboard recent changes section, but their `targetEnvironment` values were not visible in the dashboard row.

The value `Environment=True` was present, but it could come from other UI areas, such as the topbar environment placeholder. It was not sufficient proof that the Recent changes row rendered `targetEnvironment`.

### 5.3 Dashboard gap

The inventory confirmed a UI gap:

```text
Dashboard Recent changes did not show targetEnvironment explicitly.
```

This mattered because a multi-environment-ready UI must not hide which environment a ChangeRequest targets.

---

## 6. Full ChangeRequest list inventory

### 6.1 `/ui/changes` check

Observed result:

```text
Change Requests=True
Environment=True
CHG-2026-0010=True
CHG-2026-0011=True
CHG-2026-0012=True
dev=True
unknown-env=True
production=True
Process lifecycle=True
Technical runtime=True
```

### 6.2 Interpretation

The full ChangeRequest list already displayed the environment column correctly.

The list correctly showed:

```text
CHG-2026-0010 -> dev
CHG-2026-0011 -> unknown-env
CHG-2026-0012 -> production
```

This confirmed that `/ui/changes` was already environment-visible.

---

## 7. ChangeRequest detail inventory

### 7.1 `CHG-2026-0010` detail

Observed result:

```text
Change Request:=True
CHG-2026-0010=True
Environment=True
dev=True
Process lifecycle status=True
Technical runtime status=True
```

Conclusion:

```text
The detail page correctly displays Environment=dev.
```

### 7.2 `CHG-2026-0011` detail

Observed result:

```text
Change Request:=True
CHG-2026-0011=True
Environment=True
unknown-env=True
Process lifecycle status=True
Technical runtime status=True
```

Conclusion:

```text
The detail page correctly displays Environment=unknown-env for the historical pre-fix record.
```

### 7.3 `CHG-2026-0012` detail

Observed result:

```text
Change Request:=True
CHG-2026-0012=True
Environment=True
production=True
Process lifecycle status=True
Technical runtime status=True
```

Conclusion:

```text
The detail page correctly displays Environment=production for the historical pre-fix record.
```

---

## 8. Phase 14.4.1 conclusion

Inventory conclusion:

```text
/ui/changes -> OK
/ui/changes/{id} -> OK
/ui/dashboard -> needs refinement
```

The gap was limited to the dashboard Recent changes section.

The list and detail pages already displayed `targetEnvironment` correctly.

---

## 9. Phase 14.4.2 — Dashboard refinement

### 9.1 Commit

Code commit:

```text
92ca49f Show target environment in dashboard recent changes
```

### 9.2 File changed

```text
internal/api/ui_handlers.go
```

### 9.3 UI rendering before refinement

Before the refinement, the dashboard Recent changes row rendered conceptually as:

```text
applicationName · requestedBy
```

### 9.4 UI rendering after refinement

After the refinement, the dashboard Recent changes row renders as:

```text
applicationName · Environment: targetEnvironment · Requested by: requestedBy
```

Example intended renderings:

```text
demo-go-color-app · Environment: dev · Requested by: target-env-admin-a
demo-go-color-app · Environment: unknown-env · Requested by: target-env-admin-b
demo-go-color-app · Environment: production · Requested by: target-env-admin-c
```

### 9.5 Safety statement

This change is a UI visibility improvement only.

It does not:

- enable `production`;
- enable `unknown-env`;
- change environment validation;
- change API behavior;
- change database schema;
- change workflow routing.

Runtime creation remains fail-closed according to Phase 14.2.

---

## 10. Code validation

Validation completed successfully before commit:

```text
gofmt OK
go test ./... OK
git diff --check OK
```

After commit and push:

```text
HEAD -> main
origin/main
working tree clean
```

---

## 11. Build, push and deployment

### 11.1 Runtime image

Validated runtime image after rollout:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:92ca49f
```

### 11.2 Rollout

Deployment rollout completed successfully:

```text
deployment "devops-control-plane" successfully rolled out
```

---

## 12. Runtime retest after refinement

### 12.1 Test setup

The dashboard was retrieved through backend port-forward using trusted admin headers.

Readiness result:

```text
/readyz -> HTTP 200
```

Dashboard result:

```text
/ui/dashboard -> HTTP 200
```

### 12.2 Dashboard validation result

Observed result after refinement:

```text
Recent changes=True
View all=True
Environment:=True
Requested by:=True
CHG-2026-0011=True
CHG-2026-0012=True
Environment: unknown-env=True
Environment: production=True
Requested by: target-env-admin-b=True
Requested by: target-env-admin-c=True
```

### 12.3 Runtime retest summary

```text
readyz=200
dashboard_http=200
```

### 12.4 Conclusion

The runtime retest confirmed that the dashboard Recent changes section now displays `targetEnvironment` explicitly.

The specific historical records `CHG-2026-0011` and `CHG-2026-0012` now show their stored environment values directly in the dashboard Recent changes section:

```text
Environment: unknown-env
Environment: production
```

The requester is also shown explicitly:

```text
Requested by: target-env-admin-b
Requested by: target-env-admin-c
```

---

## 13. Final behavior

The UI now behaves as follows:

```text
Dashboard Recent changes:
  shows ChangeRequest number
  shows application name
  shows Environment: targetEnvironment
  shows Requested by: requestedBy
  shows technical runtime badge

Full ChangeRequest list:
  shows Environment column
  shows Process lifecycle column
  shows Technical runtime column

ChangeRequest detail:
  shows Environment
  shows Process lifecycle status
  shows Technical runtime status
```

This completes the environment visibility baseline across the main UI views.

---

## 14. Historical records note

The records below were created before fail-closed `targetEnvironment` validation was introduced:

```text
CHG-2026-0011 -> unknown-env
CHG-2026-0012 -> production
```

They remain visible because the UI must accurately display stored historical data.

They must not be interpreted as current runtime enablement of `unknown-env` or `production`.

Current runtime creation behavior remains:

```text
dev -> allowed
unknown-env -> rejected
production -> rejected
staging -> rejected until explicitly enabled
```

---

## 15. Recommended next steps

Recommended follow-up items:

```text
14.5 — UI action state consistency refinement
14.6 — Retry and idempotency policy
14.7 — Evidence UI grouping refinement
14.8 — Environment catalog runtime model
14.9 — Controlled staging enablement
```

A future environment catalog should make it possible to distinguish in UI between:

```text
stored targetEnvironment value
currently enabled environment
currently disabled environment
historical invalid environment
```

That is intentionally outside the scope of Phase 14.4.

---

## 16. Acceptance summary

Phase 14.4 is considered complete because:

- the dashboard, list and detail views were inventoried;
- the list and detail views already displayed environment correctly;
- the dashboard Recent changes environment visibility gap was identified;
- the dashboard was refined to show `Environment:` explicitly;
- the code was tested, committed, pushed, built and deployed;
- runtime validation confirmed the new dashboard rendering;
- no environment enablement behavior was changed.
