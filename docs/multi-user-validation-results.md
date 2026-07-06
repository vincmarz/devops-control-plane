# Multi-user Validation Results

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** Multi-user validation results
- **Phase:** 14.1.6 — Document multi-user validation outcome
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Runtime commit validated:** `133a496` — `Derive requester from authenticated identity`
- **Runtime image validated:** `image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:133a496`
- **Status:** Completed validation report
- **Language:** English

---

## 1. Purpose

This document records the outcome of the multi-user and role-based validation performed for the DevOps Control Plane during Phase 14.1.

The goal of this validation was to confirm that the current authenticated runtime behaves correctly with multiple requesters, role-based API access, server-side Web UI access and UI action authorization.

This document is intentionally evidence-oriented. It captures what was validated, which runtime was used, what was fixed, which results were observed and which gaps or follow-up items remain.

---

## 2. Validation scope

The validation covered the following areas:

- current AuthN/AuthZ runtime inventory;
- requester source behavior;
- remediation to derive `requestedBy` from authenticated identity;
- API authorization matrix using trusted headers;
- Web UI dashboard and full ChangeRequest list behavior;
- Web UI technical action permissions by role;
- multi-developer visibility through `requestedBy`;
- dashboard recent changes behavior;
- no-role and anonymous access behavior;
- least-privilege runtime RBAC sanity checks.

The validation was performed against the `dev` runtime baseline.

---

## 3. Runtime baseline

### 3.1 Authentication and authorization configuration

Runtime configuration confirmed:

```text
AUTH_ENABLED=true
AUTH_OPENSHIFT_GROUP_LOOKUP_ENABLED=true
```

Trusted identity headers:

```text
AUTH_HEADER_USER=X-Forwarded-User
AUTH_HEADER_ALT_USER=X-Auth-Request-User
AUTH_HEADER_GROUPS=X-Forwarded-Groups
```

Role-to-group mapping:

```text
viewer   -> devops-control-plane-viewers
operator -> devops-control-plane-operators
approver -> devops-control-plane-approvers
admin    -> devops-control-plane-admins
```

### 3.2 Runtime Secret inventory

The application Secret contains only the following keys:

```text
ARGOCD_AUTH_TOKEN
DATABASE_URL
GITLAB_TOKEN
```

The following key is intentionally absent:

```text
KUBERNETES_TOKEN
```

This confirms that Kubernetes/OpenShift API access continues to rely on the ServiceAccount token fallback rather than on a static Kubernetes token stored in the application Secret.

### 3.3 Runtime ServiceAccount and RBAC

Runtime ServiceAccount:

```text
devops-control-plane-oauth-proxy
```

Validated RBAC checks:

```text
create pipelineruns.tekton.dev in devops-ci-demo -> yes
get taskruns.tekton.dev in devops-ci-demo       -> yes
get deployments.apps in devops-ci-demo          -> yes
list pods in devops-ci-demo                     -> yes
get secrets in devops-control-plane             -> no
```

This confirms the expected least-privilege baseline for the runtime ServiceAccount.

---

## 4. Phase 14.1.1 — AuthN/AuthZ runtime inventory

### 4.1 Result

Phase 14.1.1 completed successfully.

Observed results:

```text
/readyz through Route              -> HTTP 200
/livez through Route               -> HTTP 200
/api/v1/changes anonymous          -> HTTP 403
/ui/dashboard anonymous            -> HTTP 403
```

### 4.2 Conclusion

The runtime matches the expected OpenShift OAuth Proxy and backend AuthZ baseline:

- health endpoints are public;
- business API routes are protected;
- UI routes are protected;
- anonymous access to API and UI is denied;
- the runtime ServiceAccount has least-privilege access;
- static `KUBERNETES_TOKEN` is not present in the application Secret.

---

## 5. Phase 14.1.2 — Requester source behavior before the fix

### 5.1 Test A — payload requester versus authenticated identity

Input:

```text
X-Forwarded-User=requester-header-test
payload requestedBy=requester-payload-test
```

Observed result:

```text
HTTP 201
ChangeRequest=CHG-2026-0007
saved requestedBy=requester-payload-test
```

### 5.2 Test B — missing requestedBy in payload

Input:

```text
X-Forwarded-User=requester-header-test-b
payload without requestedBy
```

Observed result:

```text
HTTP 422
error.code=VALIDATION_INVALID_REQUEST
technicalMessage="requestedBy is required"
```

### 5.3 Conclusion before remediation

Before the fix, `requestedBy` was payload-controlled and mandatory in the JSON payload, even when `AUTH_ENABLED=true`.

This behavior was acceptable for controlled tests, but it allowed logical impersonation in authenticated scenarios because a client could submit a requester different from the authenticated user.

---

## 6. Phase 14.1.2a — Remediation: derive requester from authenticated identity

### 6.1 Code change

Commit:

```text
133a496 Derive requester from authenticated identity
```

Files changed:

```text
internal/api/auth_middleware.go
internal/api/change_handlers.go
internal/api/auth_middleware_test.go
```

Implemented behavior:

```text
AUTH_ENABLED=true:
  requestedBy is derived from the authenticated identity in the request context.
  requestedBy from the JSON payload is overwritten.

AUTH_ENABLED=false:
  existing development/test behavior is preserved.
  requestedBy is still read from the JSON payload and validated by the service layer.
```

A helper was added:

```text
authenticatedUsernameFromContext
```

The `createChange` handler now overwrites `req.RequestedBy` with the authenticated username when AuthN/AuthZ is enabled.

### 6.2 Code validation

Validation completed before commit:

```text
gofmt OK
go test ./... OK
git diff --check OK
```

### 6.3 Build and deployment

Image build completed successfully using temporary Podman storage under `/tmp` to avoid pressure on `/home`.

Image tag:

```text
devops-control-plane:133a496
```

The image was pushed to the OpenShift registry using:

```text
--authfile ${HOME}/.docker/config.json
--tls-verify=false
```

Deployment image updated to:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:133a496
```

Rollout completed successfully.

### 6.4 Runtime retest after remediation

#### Test A — payload requester different from authenticated identity

Input:

```text
X-Forwarded-User=requester-header-after-fix
payload requestedBy=requester-payload-after-fix
```

Observed result:

```text
HTTP 201
ChangeRequest=CHG-2026-0008
saved requestedBy=requester-header-after-fix
```

#### Test B — missing requestedBy in payload

Input:

```text
X-Forwarded-User=requester-header-after-fix-b
payload without requestedBy
```

Observed result:

```text
HTTP 201
ChangeRequest=CHG-2026-0009
saved requestedBy=requester-header-after-fix-b
```

### 6.5 Conclusion after remediation

The remediation was successful.

With `AUTH_ENABLED=true`, `requestedBy` now derives from the authenticated/trusted identity and no longer depends on the client-provided payload.

---

## 7. Phase 14.1.3 — API authorization matrix with trusted headers

### 7.1 Test setup

Validated using backend port-forward and trusted headers.

Correct group mapping used for the final clean matrix:

```text
VIEWER_GROUP=devops-control-plane-viewers
OPERATOR_GROUP=devops-control-plane-operators
APPROVER_GROUP=devops-control-plane-approvers
ADMIN_GROUP=devops-control-plane-admins
```

Baseline check:

```text
/readyz -> HTTP 200
```

### 7.2 Final API matrix results

```text
readyz=200
admin_get_changes=200
viewer_get_changes=200
missing_identity_get_changes=401
viewer_check_validation=403
operator_check_validation=422
operator_approve=403
approver_approve=422
approver_check_validation=403
admin_check_validation=422
no_role_get_changes=403
```

### 7.3 Interpretation

The results confirm the expected API AuthZ model:

- missing identity is rejected with HTTP 401;
- viewer can read but cannot execute technical actions;
- operator can reach technical actions but cannot approve;
- approver can reach approval action but cannot execute technical validation actions;
- admin can read and reach technical actions;
- users without mapped roles are denied with HTTP 403.

The HTTP 422 responses are expected service-layer responses, not AuthZ failures. They confirm that the request passed authorization and reached application logic, but the target ChangeRequest state or technical context did not allow successful execution of the operation.

---

## 8. Phase 14.1.4 — UI dashboard and full list multi-user validation

### 8.1 Test setup

Validated using backend port-forward and admin trusted headers.

Observed HTTP results:

```text
/readyz       -> HTTP 200
/ui/dashboard -> HTTP 200
/ui/changes   -> HTTP 200
```

### 8.2 Dashboard recent changes behavior

The dashboard contains:

```text
Recent changes=True
View all=True
```

The dashboard shows the latest five ChangeRequests:

```text
CHG-2026-0009=True
CHG-2026-0008=True
CHG-2026-0007=True
CHG-2026-0006=True
CHG-2026-0005=True
```

Older ChangeRequests are excluded from the dashboard recent changes section:

```text
CHG-2026-0004=False
CHG-2026-0003=False
CHG-2026-0002=False
CHG-2026-0001=False
```

### 8.3 Requester visibility on dashboard

Confirmed requester values shown on dashboard recent changes:

```text
requester-header-after-fix-b=True
requester-header-after-fix=True
requester-payload-test=True
developer-e=True
developer-d=True
```

Note: `requester-payload-test` belongs to `CHG-2026-0007`, which was created before the remediation that derives requester from authenticated identity.

### 8.4 Full list behavior

The full list `/ui/changes` shows all ChangeRequests:

```text
CHG-2026-0001=True
CHG-2026-0002=True
CHG-2026-0003=True
CHG-2026-0004=True
CHG-2026-0005=True
CHG-2026-0006=True
CHG-2026-0007=True
CHG-2026-0008=True
CHG-2026-0009=True
```

All expected requesters are visible:

```text
vmarzario=True
developer-a=True
developer-b=True
developer-c=True
developer-d=True
developer-e=True
requester-payload-test=True
requester-header-after-fix=True
requester-header-after-fix-b=True
```

### 8.5 UI labels

The ChangeRequest list includes:

```text
Change Requests=True
Requested by=True
Environment=True
Lifecycle=True
Runtime=True
Action=True
```

The dashboard includes:

```text
Applications=True
Completed changes=True
Running changes=True
Failed changes=True
Collected evidence=True
Recent changes=True
View all=True
```

### 8.6 Conclusion

The UI correctly supports multi-developer visibility:

- dashboard remains concise with the latest five ChangeRequests;
- `/ui/changes` shows the complete list;
- `requestedBy` is visible in both dashboard and full list context;
- lifecycle and runtime information is visible in the full list.

---

## 9. Phase 14.1.5 — UI action permissions by role

### 9.1 Test setup

Validated using backend port-forward and trusted role headers.

Observed summary:

```text
readyz=200
viewer_dashboard=200
no_role_dashboard=403
viewer_ui_check_validation=403
operator_ui_check_validation=303
approver_ui_check_validation=403
admin_ui_check_validation=303
viewer_ui_collect_evidence=403
operator_ui_collect_evidence=303
approver_ui_collect_evidence=403
admin_ui_collect_evidence=303
```

### 9.2 Interpretation

The UI server-side action authorization is correct:

- viewer can read the dashboard;
- no-role user cannot access the dashboard;
- viewer cannot execute technical UI actions;
- operator can reach technical UI handlers;
- approver cannot execute technical UI actions;
- admin can reach technical UI handlers.

HTTP 303 is expected for UI action handlers because the server-side UI redirects back to the ChangeRequest detail page with a flash or error message after the action is handled.

### 9.3 Conclusion

The server-side UI respects backend AuthZ expectations:

```text
viewer  -> read-only
operator -> technical actions
approver -> approval-oriented role, not technical actions
admin -> technical actions allowed where workflow state permits
no-role -> denied
```

---

## 10. Final conclusion

Phase 14.1 confirms that the DevOps Control Plane has a solid multi-user and AuthZ baseline.

Validated final state:

```text
OAuth Proxy and backend AuthZ enabled.
OpenShift group lookup enabled.
Health endpoints remain public.
API and UI are protected for anonymous users.
requestedBy is derived from authenticated identity when AUTH_ENABLED=true.
Viewer role is read-only.
Operator role can reach technical actions.
Approver role can reach approval actions but not technical actions.
Admin role can reach read and technical actions.
Users without mapped roles are denied.
Dashboard shows latest five ChangeRequests.
Full ChangeRequest list shows all ChangeRequests.
requestedBy is visible for multi-developer traceability.
UI action handlers respect AuthZ.
```

---

## 11. Residual notes and known considerations

### 11.1 Historical ChangeRequest before requester remediation

`CHG-2026-0007` keeps:

```text
requestedBy=requester-payload-test
```

This is expected because that record was created before the remediation in commit `133a496`.

Records created after the remediation correctly derive `requestedBy` from authenticated identity.

### 11.2 HTTP 422 in authorization tests

Several tests returned HTTP 422 for authorized roles.

These are not AuthZ failures. They mean the request reached the service layer and failed due to workflow state or missing technical context, such as no Tekton PipelineRun associated with the selected ChangeRequest.

### 11.3 Production remains disabled as a runtime workflow target

These validations were performed on the current `dev` runtime baseline.

Production-oriented controls are documented, but production workflow enablement remains a future guarded step requiring environment-aware AuthZ, approval rules and runtime configuration validation.

---

## 12. Recommended next steps

Recommended follow-up phases:

```text
14.2 — Runtime targetEnvironment validation baseline
14.3 — Lifecycle vs runtime status UI/API clarity refinement
14.4 — AuthZ runtime validation matrix documentation or runbook alignment
14.5 — UI action state consistency refinement
14.6 — Retry and idempotency policy
14.7 — Evidence UI grouping refinement
```

The immediate recommended next phase is:

```text
14.2 — Runtime targetEnvironment validation baseline
```

---

## 13. Acceptance summary

Phase 14.1 is considered complete because:

- AuthN/AuthZ runtime inventory is valid;
- requester source behavior was identified and remediated;
- remediation was committed, pushed, built, deployed and runtime-validated;
- API AuthZ matrix is coherent;
- UI dashboard/list multi-user behavior is correct;
- UI action permissions are role-aware;
- no critical AuthZ gap remains in the validated `dev` baseline.
