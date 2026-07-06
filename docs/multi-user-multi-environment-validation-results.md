# Multi-user / Multi-environment Validation Results

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** Multi-user / multi-environment validation results
- **Phase:** 15.5.2e — Document multi-user multi-environment validation results
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Validation matrix definition commit:** `be4858d` — `Define multi-user multi-environment validation matrix`
- **Runtime resolver image:** `0132f06` — `Use environment cluster resolver in change validation`
- **Related phases:**
  - 15.5.1 — Define multi-user multi-environment validation matrix
  - 15.5.2a — Runtime baseline and admin create matrix
  - 15.5.2b — no-role / viewer / operator / approver / admin read matrix
  - 15.5.2c — Technical action matrix on dev
  - 15.5.2d — UI action visibility matrix
- **Status:** Completed validation report
- **Language:** English

---

## 1. Purpose

This document records the runtime validation results for the multi-user / multi-environment validation matrix executed during Phase 15.5.2.

The validation confirms the current behavior of the DevOps Control Plane after introducing:

```text
Environment Catalog runtime loading
Cluster Registry runtime loading
Environment-to-cluster resolver
ChangeService resolver-based environment validation
```

The validation was intentionally performed without enabling `staging` or `production`.

The current validated runtime behavior remains conservative:

```text
dev is enabled
staging is configured but disabled
production is configured but disabled
unknown-env is not configured
```

---

## 2. Runtime and repository baseline

### 2.1 Repository baseline

The repository head during the validation matrix definition was:

```text
be4858d Define multi-user multi-environment validation matrix
```

Recent repository history included:

```text
be4858d Define multi-user multi-environment validation matrix
e765c31 Document environment cluster resolver results
0132f06 Use environment cluster resolver in change validation
383f1f0 Add environment cluster resolver baseline
7793186 Document cluster registry baseline results
```

### 2.2 Runtime image baseline

The runtime image validated during the execution was:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:0132f06
```

This image contains the ChangeService environment-cluster resolver wiring.

### 2.3 Runtime mounts

Runtime mounts confirmed during the validation runs:

```text
container=oauth-proxy
image=registry.redhat.io/openshift4/ose-oauth-proxy:latest
mount=oauth-proxy-tls path=/etc/tls/private

container=devops-control-plane
image=image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:0132f06
mount=dcp-trust-bundle path=/etc/dcp-trust
mount=dcp-environments path=/etc/dcp-environments
mount=dcp-clusters path=/etc/dcp-clusters
```

The runtime therefore had access to:

```text
/etc/dcp-environments/environments.yaml
/etc/dcp-clusters/clusters.yaml
```

---

## 3. Environment and cluster baseline

### 3.1 Environment Catalog baseline

The active environment model was:

```text
dev:
  configured: true
  enabled: true
  clusterName: ocp-dev
  allowTechnicalActions: true

staging:
  configured: true
  enabled: false
  clusterName: ocp-staging
  allowTechnicalActions: false

production:
  configured: true
  enabled: false
  clusterName: ocp-production
  allowTechnicalActions: false

unknown-env:
  configured: false
```

### 3.2 Cluster Registry baseline

The active cluster model was:

```text
ocp-dev:
  configured: true
  enabled: true

ocp-staging:
  configured: true
  enabled: false

ocp-production:
  configured: true
  enabled: false
```

### 3.3 Resolver chain under validation

The validated application-level chain was:

```text
ChangeRequest.targetEnvironment
  -> Environment Catalog
  -> environment.clusterName
  -> Cluster Registry
  -> cluster metadata
```

The technical adapters still execute using the current single-cluster runtime configuration. The validation matrix did not test real execution against separate staging or production clusters.

---

## 4. Phase 15.5.2a — Runtime baseline and admin create matrix

### 4.1 Objective

The goal of Phase 15.5.2a was to validate the runtime baseline and the admin create matrix across target environments.

The cases tested were:

```text
missing targetEnvironment
dev
staging
production
unknown-env
```

### 4.2 Runtime readiness

Readiness result:

```text
readyz=200
```

### 4.3 Admin create matrix summary

Observed result:

```text
readyz=200
admin_missing_target_env_http=201
admin_dev_target_env_http=201
admin_staging_target_env_http=422
admin_production_target_env_http=422
admin_unknown_target_env_http=422
```

### 4.4 Interpretation

The admin create matrix confirms:

```text
missing targetEnvironment -> default dev -> accepted
dev -> accepted
staging -> rejected because disabled
production -> rejected because disabled
unknown-env -> rejected because not configured
```

This confirms that the ChangeService resolver wiring preserved the expected environment validation behavior.

### 4.5 Phase 15.5.2a status

```text
15.5.2a — Runtime baseline and admin create matrix
COMPLETED
```

---

## 5. Phase 15.5.2b — Read matrix

### 5.1 Objective

The goal of Phase 15.5.2b was to validate read access across roles for API and UI read endpoints.

Roles tested:

```text
no-role
viewer
operator
approver
admin
```

Endpoints tested:

```text
GET /api/v1/changes
GET /ui/changes
GET /ui/dashboard
```

### 5.2 API read matrix

Observed result:

```text
api_read_no_role=403
api_read_viewer=200
api_read_operator=200
api_read_approver=200
api_read_admin=200
```

Interpretation:

```text
no-role is correctly denied.
viewer, operator, approver and admin can read the API change list.
```

### 5.3 UI changes list matrix

Observed result:

```text
ui_changes_no_role=403
ui_changes_viewer=200
ui_changes_operator=200
ui_changes_approver=200
ui_changes_admin=200
```

Interpretation:

```text
no-role is correctly denied.
viewer, operator, approver and admin can read the UI change list.
```

### 5.4 Dashboard path correction

Initial validation used:

```text
/ui/
```

The initial `/ui/` request returned `403` for all roles. Code inspection confirmed that `/ui/` is not the canonical route.

Canonical dashboard routes are:

```text
/
/ui
/ui/dashboard
```

The dashboard matrix was retested using:

```text
/ui/dashboard
```

### 5.5 Canonical dashboard matrix

Observed result:

```text
readyz_dashboard_retest=200
ui_dashboard_canonical_no_role=403
ui_dashboard_canonical_viewer=200
ui_dashboard_canonical_operator=200
ui_dashboard_canonical_approver=200
ui_dashboard_canonical_admin=200
```

Interpretation:

```text
no-role is correctly denied.
viewer, operator, approver and admin can read the canonical dashboard route.
```

### 5.6 Finding: non-canonical /ui/ path

```text
FINDING-15.5.2b-001:
  Initial dashboard matrix used non-canonical path /ui/.
  Canonical dashboard path /ui/dashboard was retested successfully.
  No role authorization defect was confirmed.
```

### 5.7 Phase 15.5.2b status

```text
15.5.2b — no-role / viewer / operator / approver / admin read matrix
COMPLETED
```

---

## 6. Phase 15.5.2c — Technical action matrix on dev

### 6.1 Objective

The goal of Phase 15.5.2c was to validate technical action execution on an enabled `dev` ChangeRequest using different roles.

The technical action selected was:

```text
POST /api/v1/changes/{changeNumber}/collect-evidence
```

### 6.2 Baseline ChangeRequest

A dedicated dev ChangeRequest was created:

```text
create_dev_baseline_http=201
change_number=CHG-2026-0026
```

### 6.3 API technical action matrix

Observed result:

```text
readyz=200
create_dev_baseline_http=201
change_number=CHG-2026-0026
collect_evidence_no_role=403
collect_evidence_viewer=403
collect_evidence_operator=202
collect_evidence_approver=403
collect_evidence_admin=202
```

### 6.4 Response interpretation

For denied users, response body was plain text:

```text
forbidden: insufficient role
```

For `operator` and `admin`, the action succeeded and returned application JSON with:

```text
runtimeStatus=EvidenceCollected
status=draft
```

### 6.5 Authorization conclusion

The technical action policy is enforced correctly:

```text
no-role   -> denied
viewer    -> denied
operator  -> allowed
approver  -> denied
admin     -> allowed
```

### 6.6 Phase 15.5.2c status

```text
15.5.2c — Technical action matrix on dev
COMPLETED
```

---

## 7. Phase 15.5.2d — UI action visibility matrix

### 7.1 Objective

The goal of Phase 15.5.2d was to validate UI visibility and UI POST behavior for technical actions on the same dev ChangeRequest:

```text
CHG-2026-0026
```

UI detail page:

```text
GET /ui/changes/CHG-2026-0026
```

UI action endpoint:

```text
POST /ui/changes/CHG-2026-0026/actions/collect-evidence
```

### 7.2 UI detail access matrix

Observed result:

```text
readyz=200
ui_detail_no_role=403
ui_detail_viewer=200
ui_detail_operator=200
ui_detail_approver=200
ui_detail_admin=200
```

Interpretation:

```text
no-role is correctly denied.
viewer, operator, approver and admin can read the ChangeRequest detail page.
```

### 7.3 UI visibility result

Observed HTML visibility for `viewer`, `operator`, `approver` and `admin`:

```text
technical_actions_section=True
recommended_next_actions=True
advanced_manual_actions=True
validate_label=True
```

Observed HTML visibility for `no-role`:

```text
body_bytes=29
forbidden_text=True
```

### 7.4 UI POST action matrix

Observed result:

```text
ui_action_collect_evidence_no_role=403
ui_action_collect_evidence_viewer=403
ui_action_collect_evidence_operator=303
ui_action_collect_evidence_approver=403
ui_action_collect_evidence_admin=303
```

Interpretation:

```text
no-role   -> denied
viewer    -> denied
operator  -> allowed, UI redirect 303
approver  -> denied
admin     -> allowed, UI redirect 303
```

The `303` response is expected UI redirect behavior after a successful POST action.

### 7.5 Finding: UI action visibility not role-filtered

```text
FINDING-15.5.2d-001:
  UI action visibility is not role-filtered.
```

Details:

```text
viewer and approver can see the Technical actions section,
Recommended next actions,
Advanced/manual actions,
and at least the Validate action label.

However, POST execution is correctly denied by RBAC/AuthZ.
```

Impact:

```text
Security enforcement: OK
User experience and action visibility: requires improvement
```

### 7.6 Security conclusion

There is no confirmed security bypass. The backend and UI POST authorization correctly deny unauthorized execution.

The finding is limited to UI visibility and should be handled as a UX/AuthZ consistency refinement.

### 7.7 Phase 15.5.2d status

```text
15.5.2d — UI action visibility matrix
COMPLETED WITH FINDING
```

---

## 8. Overall validation summary

### 8.1 Completed validation phases

```text
15.5.2a — Runtime baseline and admin create matrix      COMPLETED
15.5.2b — Read matrix                                   COMPLETED
15.5.2c — Technical action matrix on dev                COMPLETED
15.5.2d — UI action visibility matrix                   COMPLETED WITH FINDING
```

### 8.2 Main HTTP results

```text
Admin create matrix:
  missing targetEnvironment -> 201
  dev -> 201
  staging -> 422
  production -> 422
  unknown-env -> 422

Read matrix:
  no-role API/UI read -> 403
  viewer/operator/approver/admin API/UI read -> 200

Technical action API matrix:
  no-role -> 403
  viewer -> 403
  operator -> 202
  approver -> 403
  admin -> 202

UI action POST matrix:
  no-role -> 403
  viewer -> 403
  operator -> 303
  approver -> 403
  admin -> 303
```

### 8.3 Environment behavior

Environment behavior remained correct:

```text
dev remains enabled and operational.
staging remains disabled.
production remains disabled.
unknown-env remains not configured.
```

### 8.4 Role behavior

Role behavior remained correct at enforcement level:

```text
no-role is denied protected access.
viewer can read but cannot execute collect-evidence.
operator can read and execute collect-evidence on dev.
approver can read but cannot execute collect-evidence.
admin can read and execute collect-evidence on dev.
```

---

## 9. Findings

### 9.1 FINDING-15.5.2b-001 — Non-canonical dashboard path

Initial dashboard testing used:

```text
/ui/
```

This returned `403` for all roles. The canonical path:

```text
/ui/dashboard
```

was retested and behaved correctly.

Status:

```text
Closed as test path correction.
```

Recommendation:

```text
Future UI tests should use /ui/dashboard or /ui, not /ui/.
```

Optional future improvement:

```text
Add redirect or explicit route handling for /ui/ to improve browser usability.
```

### 9.2 FINDING-15.5.2d-001 — UI actions visible to roles that cannot execute them

The UI detail page shows technical action sections to `viewer` and `approver` even though those roles are denied execution.

Status:

```text
Open UX/AuthZ consistency finding.
```

Security impact:

```text
No security bypass confirmed.
```

Recommendation:

```text
Introduce role-aware UI action visibility so only authorized roles see executable technical actions.
```

Suggested follow-up phase:

```text
15.6 — Role-aware UI action visibility refinement
```

---

## 10. Final acceptance status

Phase 15.5.2 runtime validation is accepted because:

```text
runtime readiness is healthy;
repository working tree remained clean;
admin create behavior matches the environment policy;
read access behaves correctly across roles;
technical action enforcement is correct across roles;
UI POST enforcement is correct across roles;
dev remains operational;
staging and production remain disabled;
unknown-env remains rejected;
no security bypass was observed.
```

The only open item is UI visibility refinement, not backend authorization.

---

## 11. Recommended next phase

Proceed with:

```text
15.6 — Role-aware UI action visibility refinement
```

Recommended goal:

```text
Use the same authorization semantics that protect POST actions to also hide or disable technical action buttons in the UI for unauthorized roles.
```

Initial acceptance criteria for the follow-up:

```text
viewer can read the detail page but does not see executable technical action buttons;
approver can read the detail page but does not see collect-evidence technical action buttons;
operator sees collect-evidence where allowed;
admin sees collect-evidence where allowed;
POST enforcement remains unchanged;
unit tests and runtime UI evidence confirm consistency.
```
