# 15.6.4 — Role-aware UI Action Visibility Results

## 1. Purpose

This document records the implementation and runtime validation results for **Phase 15.6 — Role-aware UI action visibility refinement** in the DevOps Control Plane project.

Phase 15.6 was introduced after the multi-user and multi-environment validation matrix identified a user-experience inconsistency in the ChangeRequest detail page.

The server-side authorization model was already correct: roles that were not allowed to execute technical actions were rejected by the backend. However, the UI still displayed the technical action area to roles that could not execute those actions.

The objective of this phase was to align UI visibility with the existing backend authorization policy while keeping backend enforcement unchanged.

---

## 2. Related finding

### FINDING-15.5.2d-001 — UI action visibility is not role-filtered

During **15.5.2d — UI action visibility matrix**, the ChangeRequest detail page showed the **Technical actions** section and related action groups to roles that were not authorized to execute technical actions.

Observed behavior before the fix:

```text
viewer:
  detail page: 200
  Technical actions visible: yes
  POST collect-evidence: 403

approver:
  detail page: 200
  Technical actions visible: yes
  POST collect-evidence: 403

operator:
  detail page: 200
  Technical actions visible: yes
  POST collect-evidence: 303

admin:
  detail page: 200
  Technical actions visible: yes
  POST collect-evidence: 303
```

Security impact:

```text
No server-side authorization bypass was found.
```

User experience impact:

```text
Users without technical execution permissions could see technical actions that would later fail at execution time.
```

Resolution target:

```text
viewer and approver should be able to read ChangeRequest details, but should not see executable technical action controls.
operator and admin should continue to see and execute technical actions where the target environment allows automation.
```

---

## 3. Phase scope

### In scope

Phase 15.6 covered:

```text
15.6.1 — Inventory current UI action visibility and AuthZ helpers
15.6.2 — Add role-aware UI action visibility
15.6.3 — Runtime validation for role-aware UI action visibility
15.6.4 — Document role-aware UI action visibility results
```

### Out of scope

The following items were intentionally left out of this phase:

```text
changing server-side POST authorization rules
changing OAuth Proxy configuration
changing OpenShift RBAC
changing Environment Catalog semantics
changing Cluster Registry semantics
enabling staging or production
routing technical execution to non-dev clusters
redesigning the entire UI template structure
```

---

## 4. Baseline before remediation

At the beginning of the phase, the repository state was:

```text
8a1b70d Document multi-user multi-environment validation results
```

The runtime image before the UI visibility fix was:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:0132f06
```

The runtime deployment already included the required multi-environment and cluster-aware mounts:

```text
mount=dcp-trust-bundle path=/etc/dcp-trust
mount=dcp-environments path=/etc/dcp-environments
mount=dcp-clusters path=/etc/dcp-clusters
```

The previously validated backend behavior was:

```text
POST /ui/changes/{changeNumber}/actions/collect-evidence

viewer   -> 403
operator -> 303
approver -> 403
admin    -> 303
```

The issue was limited to UI visibility.

---

## 5. 15.6.1 — Inventory current UI action visibility and AuthZ helpers

### 5.1 Inventory summary

The inventory produced the following summary:

```text
git_head=8a1b70d Document multi-user multi-environment validation results
ui_change_action_refs=2
recommended_actions_refs=4
advanced_actions_refs=3
environment_allows_technical_actions_refs=3
insufficient_role_refs=1
template_action_posts_refs=2
```

### 5.2 Key files identified

The primary UI implementation file was:

```text
internal/api/ui_handlers.go
```

The primary authorization middleware file was:

```text
internal/api/auth_middleware.go
```

The main UI functions and helpers identified were:

```text
uiChangeAction
recommendedActions
advancedActions
environmentAllowsTechnicalActions
environmentActionWarning
```

The existing server-side authorization path was enforced in:

```text
internal/api/auth_middleware.go
```

The authorization failure message was confirmed as:

```text
forbidden: insufficient role
```

### 5.3 Root cause

The ChangeRequest detail template displayed the technical action section using an environment-only condition:

```text
environmentAllowsTechnicalActions .SelectedChange
```

This condition checked whether the target environment allowed technical actions, but did not check whether the authenticated user role was allowed to execute technical actions.

The server-side POST authorization policy was already role-aware, but the UI visibility logic was not.

---

## 6. 15.6.2 — Implementation summary

### 6.1 Commit

The fix was committed as:

```text
7d79087 Add role-aware UI action visibility
```

### 6.2 Files changed

The implementation changed or added:

```text
internal/api/ui_handlers.go
internal/api/ui_action_visibility_test.go
```

Commit delta:

```text
2 files changed, 80 insertions(+), 5 deletions(-)
create mode 100644 internal/api/ui_action_visibility_test.go
```

### 6.3 Implementation model

The implementation introduced role-aware UI action visibility helpers in the API/UI layer.

The core concept is:

```text
ChangeRequest detail page remains readable for authorized read roles.
Technical action controls are visible only when:
  target environment allows technical actions
  and the authenticated user role can execute technical actions
```

The UI map enrichment is performed when converting the domain ChangeRequest into the map used by templates.

The relevant transformation became:

```go
selected := withUIActionVisibility(r.Context(), toMap(change))
```

This enrichment is applied in the ChangeRequest-related UI views where a selected ChangeRequest is rendered.

Confirmed occurrences:

```text
selected := withUIActionVisibility(r.Context(), toMap(change))
```

present in the relevant UI handlers.

### 6.4 Role-aware flag

The helper enriches the UI map with a role-derived property equivalent to:

```text
uiCanSeeTechnicalActions
```

The template uses this property through a helper equivalent to:

```text
userCanSeeTechnicalActions .SelectedChange
```

The template condition became conceptually:

```text
environmentAllowsTechnicalActions(.SelectedChange)
AND
userCanSeeTechnicalActions(.SelectedChange)
```

This means:

```text
viewer:
  can read details
  cannot see executable technical action groups

approver:
  can read details
  cannot see executable technical action groups

operator:
  can read details
  can see technical action groups on enabled environments

admin:
  can read details
  can see technical action groups on enabled environments
```

### 6.5 Server-side enforcement preserved

The implementation did not weaken or replace backend enforcement.

The POST endpoint is still protected by the existing authorization middleware. UI visibility is a usability improvement, not a security boundary.

The security boundary remains:

```text
server-side AuthZ middleware
```

---

## 7. Local validation

After implementation and recovery from an initial type mismatch, local validation passed.

### 7.1 Issue encountered during implementation

The first patch attempted to apply the UI map helper directly to a `domain.ChangeRequest` object:

```go
change = withUIActionVisibility(r.Context(), change)
```

This failed because:

```text
withUIActionVisibility expects map[string]any
change was domain.ChangeRequest
```

Compilation errors were observed in `internal/api/ui_handlers.go`.

### 7.2 Recovery

The invalid enrichment lines were removed:

```text
removed_invalid_lines=3
```

The correct enrichment was then applied to the mapped representation:

```text
selected_enriched=3
```

The final valid form was:

```go
selected := withUIActionVisibility(r.Context(), toMap(change))
```

### 7.3 Successful local validation

The final local validation passed:

```text
go test ./... OK
git diff --check OK
```

The working tree before commit contained only:

```text
M  internal/api/ui_handlers.go
?? internal/api/ui_action_visibility_test.go
```

The final commit was pushed successfully:

```text
7d79087 Add role-aware UI action visibility
```

---

## 8. Runtime deployment validation

### 8.1 Runtime image

The runtime validation was executed against image:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:7d79087
```

### 8.2 Runtime mounts

The expected mounts remained present:

```text
mount=dcp-trust-bundle path=/etc/dcp-trust
mount=dcp-environments path=/etc/dcp-environments
mount=dcp-clusters path=/etc/dcp-clusters
```

This confirms that the role-aware UI change did not disturb the multi-environment and cluster-aware runtime baseline.

### 8.3 Test ChangeRequest

The runtime validation reused the existing dev ChangeRequest:

```text
CHG-2026-0026
```

This ChangeRequest had previously been used for the technical action matrix on dev.

---

## 9. 15.6.3 — Runtime validation results

### 9.1 Readiness

Runtime readiness was successful:

```text
readyz=200
```

### 9.2 UI detail HTTP access matrix

Final HTTP summary:

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
no-role remains denied
viewer can read detail
operator can read detail
approver can read detail
admin can read detail
```

This confirms that read access behavior remained aligned with the previously validated read matrix.

---

## 10. UI role-aware visibility results

### 10.1 no-role

Observed:

```text
CASE=no-role
body_bytes=29
technical_actions_section=False
recommended_next_actions=False
advanced_manual_actions=False
validate_label=False
collect_evidence_label=False
technical_unavailable_message=False
forbidden_text=True
END_CASE=no-role
```

Interpretation:

```text
no-role users cannot read the ChangeRequest detail page.
```

This is expected.

### 10.2 viewer

Observed:

```text
CASE=viewer
body_bytes=17797
technical_actions_section=True
recommended_next_actions=False
advanced_manual_actions=False
validate_label=False
collect_evidence_label=False
technical_unavailable_message=True
forbidden_text=False
END_CASE=viewer
```

Interpretation:

```text
viewer users can read the ChangeRequest detail page.
viewer users no longer see executable technical action groups.
viewer users see an informational message indicating that technical actions are not available.
```

This is the intended behavior.

### 10.3 operator

Observed:

```text
CASE=operator
body_bytes=19816
technical_actions_section=True
recommended_next_actions=True
advanced_manual_actions=True
validate_label=True
collect_evidence_label=False
technical_unavailable_message=False
forbidden_text=False
END_CASE=operator
```

Interpretation:

```text
operator users can read the ChangeRequest detail page.
operator users can see technical action groups.
operator users can see recommended and advanced/manual technical actions.
```

This is the intended behavior.

### 10.4 approver

Observed:

```text
CASE=approver
body_bytes=17797
technical_actions_section=True
recommended_next_actions=False
advanced_manual_actions=False
validate_label=False
collect_evidence_label=False
technical_unavailable_message=True
forbidden_text=False
END_CASE=approver
```

Interpretation:

```text
approver users can read the ChangeRequest detail page.
approver users no longer see executable technical action groups.
approver users see an informational message indicating that technical actions are not available.
```

This is the intended behavior.

### 10.5 admin

Observed:

```text
CASE=admin
body_bytes=19816
technical_actions_section=True
recommended_next_actions=True
advanced_manual_actions=True
validate_label=True
collect_evidence_label=False
technical_unavailable_message=False
forbidden_text=False
END_CASE=admin
```

Interpretation:

```text
admin users can read the ChangeRequest detail page.
admin users can see technical action groups.
admin users can see recommended and advanced/manual technical actions.
```

This is the intended behavior.

---

## 11. POST enforcement retest

The UI POST action enforcement was retested for:

```text
POST /ui/changes/CHG-2026-0026/actions/collect-evidence
```

Final POST summary:

```text
ui_action_collect_evidence_viewer=403
ui_action_collect_evidence_operator=303
ui_action_collect_evidence_approver=403
ui_action_collect_evidence_admin=303
```

Interpretation:

```text
viewer remains denied
operator remains allowed and redirected by the UI flow
approver remains denied
admin remains allowed and redirected by the UI flow
```

This confirms that backend and UI POST enforcement remained unchanged and correct.

---

## 12. Finding closure

### FINDING-15.5.2d-001 status

```text
CLOSED
```

Reason:

```text
The UI now hides executable technical action groups from viewer and approver roles.
The UI continues to show executable technical action groups to operator and admin roles.
Server-side POST enforcement remains unchanged and correct.
```

Residual note:

```text
The Technical actions section title may remain visible for viewer and approver users, but executable action groups are hidden and an informational message is shown.
```

This is acceptable and improves clarity because users can understand why no technical actions are available to their role.

---

## 13. Security assessment

### 13.1 Security boundary

The security boundary remains server-side authorization:

```text
internal/api/auth_middleware.go
```

The UI visibility layer is not relied on as a security control.

### 13.2 Enforcement remains correct

The POST validation confirms that unauthorized roles remain blocked:

```text
viewer -> 403
approver -> 403
```

Authorized technical roles remain allowed:

```text
operator -> 303
admin -> 303
```

### 13.3 No sensitive data exposure introduced

The implementation does not introduce new tokens, secrets, cluster credentials or sensitive runtime configuration.

---

## 14. Operational notes

### 14.1 Runtime image alignment

The repository head for this remediation is:

```text
7d79087 Add role-aware UI action visibility
```

The runtime image was updated to the same commit tag:

```text
7d79087
```

### 14.2 Existing runtime configuration preserved

The following runtime capabilities remained available:

```text
Environment Catalog mounted from /etc/dcp-environments
Cluster Registry mounted from /etc/dcp-clusters
OAuth Proxy sidecar present
Readiness successful
```

### 14.3 Working tree status

At the end of validation:

```text
git status --short
```

returned no output.

---

## 15. Acceptance criteria

### AC-15.6-001 — viewer detail access remains available

Status:

```text
PASSED
```

Evidence:

```text
ui_detail_viewer=200
```

### AC-15.6-002 — viewer executable technical actions hidden

Status:

```text
PASSED
```

Evidence:

```text
viewer:
  recommended_next_actions=False
  advanced_manual_actions=False
  validate_label=False
  technical_unavailable_message=True
```

### AC-15.6-003 — approver detail access remains available

Status:

```text
PASSED
```

Evidence:

```text
ui_detail_approver=200
```

### AC-15.6-004 — approver executable technical actions hidden

Status:

```text
PASSED
```

Evidence:

```text
approver:
  recommended_next_actions=False
  advanced_manual_actions=False
  validate_label=False
  technical_unavailable_message=True
```

### AC-15.6-005 — operator technical actions remain visible

Status:

```text
PASSED
```

Evidence:

```text
operator:
  recommended_next_actions=True
  advanced_manual_actions=True
  validate_label=True
```

### AC-15.6-006 — admin technical actions remain visible

Status:

```text
PASSED
```

Evidence:

```text
admin:
  recommended_next_actions=True
  advanced_manual_actions=True
  validate_label=True
```

### AC-15.6-007 — POST enforcement remains unchanged

Status:

```text
PASSED
```

Evidence:

```text
ui_action_collect_evidence_viewer=403
ui_action_collect_evidence_operator=303
ui_action_collect_evidence_approver=403
ui_action_collect_evidence_admin=303
```

---

## 16. Final status

```text
15.6 — Role-aware UI action visibility refinement
COMPLETED
```

Completed sub-phases:

```text
15.6.1 — Inventory current UI action visibility and AuthZ helpers    COMPLETED
15.6.2 — Add role-aware UI action visibility                         COMPLETED
15.6.3 — Runtime validation for role-aware UI action visibility       COMPLETED
15.6.4 — Document role-aware UI action visibility results             COMPLETED
```

---

## 17. Recommended next steps

Recommended next phase:

```text
15.7 — Role-aware UI regression and documentation consolidation
```

Potential follow-up items:

```text
add automated UI rendering tests for role-aware action visibility
add explicit documentation of UI routes and canonical dashboard paths
add regression tests for /ui/ versus /ui and /ui/dashboard behavior
review whether the Technical actions section title should remain visible for non-technical roles
extend role-aware visibility to future promotion and multi-cluster execution actions
```

Suggested acceptance baseline for future regression:

```text
viewer:
  read detail page
  no executable technical action buttons
  POST technical action denied

approver:
  read detail page
  no executable technical action buttons
  POST technical action denied

operator:
  read detail page
  technical action buttons visible where environment allows
  POST technical action allowed

admin:
  read detail page
  technical action buttons visible where environment allows
  POST technical action allowed
```

---

## 18. Appendix — Evidence references

Runtime validation evidence was collected under a directory matching:

```text
/tmp/dcp-15.6.3-role-aware-ui-runtime-validation-YYYYMMDD-HHMMSS
```

Important evidence files:

```text
06-readyz-code.txt
10-ui-detail-no-role-code.txt
11-ui-detail-viewer-code.txt
12-ui-detail-operator-code.txt
13-ui-detail-approver-code.txt
14-ui-detail-admin-code.txt
20-ui-action-visibility-summary.txt
21-ui-detail-http-summary.txt
31-ui-action-collect-evidence-viewer-code.txt
32-ui-action-collect-evidence-operator-code.txt
33-ui-action-collect-evidence-approver-code.txt
34-ui-action-collect-evidence-admin-code.txt
35-ui-action-post-summary.txt
```
