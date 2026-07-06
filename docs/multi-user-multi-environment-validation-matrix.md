# Multi-user / Multi-environment / Multi-cluster Validation Matrix

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** Multi-user / multi-environment / multi-cluster validation matrix
- **Phase:** 15.5.1 — Define multi-user multi-environment validation matrix
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Current baseline commit:** `e765c31` — `Document environment cluster resolver results`
- **Runtime resolver commit:** `0132f06` — `Use environment cluster resolver in change validation`
- **Related phases:**
  - 14.1 — Multi-user validation baseline
  - 14.2 — Target environment validation baseline
  - 14.5 — Environment Catalog and UI action visibility
  - 15.1 — Multi-cluster enablement request
  - 15.2 — Environment Catalog runtime loading
  - 15.3 — Cluster Registry baseline
  - 15.4 — Environment-to-cluster resolver baseline
- **Status:** Validation plan
- **Language:** English

---

## 1. Purpose

This document defines the validation matrix for testing the DevOps Control Plane with multiple user roles, multiple target environments and the current multi-cluster metadata baseline.

The purpose of Phase 15.5.1 is to define what must be tested before executing the validation runs.

The current runtime is already environment-aware and cluster-aware at the application metadata layer:

```text
ChangeRequest.targetEnvironment
  -> Environment Catalog
  -> environment.clusterName
  -> Cluster Registry
  -> cluster metadata
```

However, technical adapters still execute against the current dev OpenShift cluster configuration. Therefore this matrix validates authorization, environment validation and cluster-aware metadata behavior without enabling staging or production.

---

## 2. Current baseline

The current expected repository baseline is:

```text
e765c31 Document environment cluster resolver results
0132f06 Use environment cluster resolver in change validation
383f1f0 Add environment cluster resolver baseline
7793186 Document cluster registry baseline results
52de0c1 Add cluster registry baseline
```

The current runtime image expected for the test baseline is:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:0132f06
```

The runtime should have the following mounts:

```text
/etc/dcp-trust
/etc/dcp-environments
/etc/dcp-clusters
```

The current Environment Catalog state is:

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

The current Cluster Registry state is:

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

---

## 3. Validation principles

The validation must confirm the following principles:

```text
1. no-role users cannot access protected API/UI paths;
2. viewers can read where allowed but cannot execute technical or approval actions;
3. operators can execute technical actions only on enabled environments;
4. approvers can execute approval lifecycle actions only where lifecycle and role allow it;
5. admins can execute allowed operations on enabled environments;
6. staging remains disabled;
7. production remains disabled;
8. unknown-env remains not configured;
9. disabled environments must not become operational because of role privileges;
10. cluster-aware metadata must not imply multi-cluster execution yet.
```

---

## 4. User roles under test

The test matrix uses the current global groups:

```text
devops-control-plane-viewers
devops-control-plane-operators
devops-control-plane-approvers
devops-control-plane-admins
```

The logical users for validation are:

```text
no-role-user
viewer-user
operator-user
approver-user
admin-user
```

The corresponding test headers are:

```text
no-role-user:
  X-Forwarded-User: matrix-no-role
  X-Forwarded-Groups: none or unrelated group

viewer-user:
  X-Forwarded-User: matrix-viewer
  X-Forwarded-Groups: devops-control-plane-viewers

operator-user:
  X-Forwarded-User: matrix-operator
  X-Forwarded-Groups: devops-control-plane-operators

approver-user:
  X-Forwarded-User: matrix-approver
  X-Forwarded-Groups: devops-control-plane-approvers

admin-user:
  X-Forwarded-User: matrix-admin
  X-Forwarded-Groups: devops-control-plane-admins
```

For admin-only setup operations, use:

```text
X-Forwarded-User: matrix-admin
X-Forwarded-Groups: devops-control-plane-admins
```

---

## 5. Environments under test

The environment dimension is:

```text
dev
staging
production
unknown-env
missing targetEnvironment
```

Expected environment behavior:

```text
missing targetEnvironment -> default dev -> HTTP 201 for authorized create
dev -> HTTP 201 for authorized create
staging -> HTTP 422 targetEnvironment "staging" is currently disabled
production -> HTTP 422 targetEnvironment "production" is currently disabled
unknown-env -> HTTP 422 targetEnvironment "unknown-env" is not configured
```

---

## 6. Cluster dimension under test

The cluster metadata dimension is:

```text
ocp-dev
ocp-staging
ocp-production
```

Expected cluster behavior in this phase:

```text
dev -> ocp-dev -> configured and enabled
staging -> ocp-staging -> configured but disabled
production -> ocp-production -> configured but disabled
unknown-env -> no environment -> no cluster resolution
```

Important: Phase 15.5 validates cluster metadata resolution indirectly through the environment validation baseline and resolver wiring. It does not validate real API communication with multiple OpenShift clusters.

---

## 7. API endpoints under test

### 7.1 Read endpoints

Read endpoints to validate:

```text
GET /readyz
GET /api/v1/changes
GET /api/v1/changes/{changeNumber}
GET /ui/
GET /ui/changes
GET /ui/changes/{changeNumber}
```

Expected behavior depends on AuthZ policy.

Baseline expectations:

```text
admin -> allowed
operator -> allowed for normal read paths
approver -> allowed for normal read paths
viewer -> allowed for normal read paths
no-role -> denied for protected read paths
```

### 7.2 Create endpoint

Create endpoint:

```text
POST /api/v1/changes
```

Expected behavior:

```text
admin + missing/dev -> 201
admin + staging -> 422
admin + production -> 422
admin + unknown-env -> 422
unknown role + any create -> 403
```

For viewer/operator/approver create permissions, the matrix must follow the current AuthZ implementation. If create is admin/operator-only, viewer and approver must be denied.

### 7.3 Lifecycle endpoints

Lifecycle endpoints to validate where applicable:

```text
POST /api/v1/changes/{changeNumber}/submit
POST /api/v1/changes/{changeNumber}/approve
POST /api/v1/changes/{changeNumber}/start-execution
POST /api/v1/changes/{changeNumber}/complete-execution
POST /api/v1/changes/{changeNumber}/close
```

Expected role behavior:

```text
viewer -> denied for lifecycle mutation
operator -> allowed only for execution-oriented lifecycle actions, if current policy allows
approver -> allowed for approval actions, if lifecycle state allows
admin -> allowed where lifecycle state allows
no-role -> denied
```

### 7.4 Technical workflow endpoints

Technical actions to validate:

```text
POST /api/v1/changes/{changeNumber}/validate
POST /api/v1/changes/{changeNumber}/check-validation
POST /api/v1/changes/{changeNumber}/collect-evidence
POST /api/v1/changes/{changeNumber}/create-branch
POST /api/v1/changes/{changeNumber}/update-files
POST /api/v1/changes/{changeNumber}/open-merge-request
POST /api/v1/changes/{changeNumber}/merge-request
POST /api/v1/changes/{changeNumber}/check-deployment
```

Expected environment behavior:

```text
dev -> technical actions visible/available for authorized operator/admin
staging -> technical actions unavailable because environment disabled
production -> technical actions unavailable because environment disabled
unknown-env -> technical actions unavailable because environment not configured
```

Expected role behavior:

```text
viewer -> technical actions denied
operator -> technical actions allowed on dev where current AuthZ allows
approver -> technical actions denied unless current policy explicitly allows
admin -> technical actions allowed on dev
no-role -> denied
```

---

## 8. UI behavior under test

UI pages to validate:

```text
/ui/
/ui/changes
/ui/changes/{devChange}
/ui/changes/{stagingHistoricalOrSyntheticChange}
/ui/changes/{productionHistoricalOrSyntheticChange}
/ui/changes/{unknownHistoricalOrSyntheticChange}
```

Expected UI behavior:

```text
dev:
  Technical actions visible for authorized operator/admin
  Recommended next actions visible where applicable
  Advanced/manual actions visible where applicable

staging disabled:
  warning visible
  technical actions hidden or unavailable

production disabled:
  warning visible
  technical actions hidden or unavailable

unknown-env:
  warning visible
  technical actions hidden or unavailable
```

Expected UI role behavior:

```text
viewer:
  read-only views allowed
  technical actions not executable

operator:
  technical actions visible/executable on dev
  approval actions not executable unless policy allows

approver:
  approval actions visible/executable where lifecycle allows
  technical actions not executable unless policy allows

admin:
  full actions on dev where lifecycle and runtime state allow

no-role:
  protected UI denied
```

---

## 9. Core matrix — Create ChangeRequest

### 9.1 Admin create matrix

Admin create test cases:

```text
admin + missing targetEnvironment -> 201, targetEnvironment=dev
admin + dev -> 201, targetEnvironment=dev
admin + staging -> 422, staging currently disabled
admin + production -> 422, production currently disabled
admin + unknown-env -> 422, unknown-env not configured
```

### 9.2 Non-admin create matrix

Non-admin create behavior depends on current AuthZ policy. The expected baseline should be validated, not guessed.

Minimum required checks:

```text
no-role + dev create -> 403
viewer + dev create -> expected 403 unless current policy allows create
operator + dev create -> expected according to current policy
approver + dev create -> expected 403 unless current policy allows create
```

If operator create is allowed by current policy, additional checks must confirm:

```text
operator + staging -> 422 or 403 depending on ordering of AuthZ vs validation
operator + production -> 422 or 403 depending on ordering of AuthZ vs validation
operator + unknown-env -> 422 or 403 depending on ordering of AuthZ vs validation
```

The actual behavior must be documented from runtime evidence.

---

## 10. Core matrix — Read access

Read access matrix:

```text
no-role:
  GET /api/v1/changes -> 403
  GET /ui/changes -> 403

viewer:
  GET /api/v1/changes -> 200
  GET /ui/changes -> 200

operator:
  GET /api/v1/changes -> 200
  GET /ui/changes -> 200

approver:
  GET /api/v1/changes -> 200
  GET /ui/changes -> 200

admin:
  GET /api/v1/changes -> 200
  GET /ui/changes -> 200
```

Any deviation must be recorded as actual policy behavior.

---

## 11. Core matrix — Technical actions on dev

Technical action validation must use a dev ChangeRequest created for the test run.

Expected minimum behavior:

```text
no-role + technical action dev -> 403
viewer + technical action dev -> 403
operator + technical action dev -> allowed where current policy allows
approver + technical action dev -> 403 unless current policy allows
admin + technical action dev -> allowed where lifecycle/runtime permits
```

High-value technical endpoints to start with:

```text
check-validation
collect-evidence
check-deployment
```

GitLab and merge actions may have state preconditions, so they should be tested separately if a workflow-ready ChangeRequest is available.

---

## 12. Core matrix — Disabled environments

The disabled environment matrix validates that role privileges do not override environment state.

Expected behavior:

```text
admin + staging create -> 422 disabled
admin + production create -> 422 disabled
operator + staging technical action -> not available or rejected
operator + production technical action -> not available or rejected
admin + staging technical action -> not available or rejected
admin + production technical action -> not available or rejected
viewer + staging technical action -> denied
viewer + production technical action -> denied
no-role + staging/production anything -> denied
```

Important: if a historical staging or production ChangeRequest exists from before validation hardening, read access may remain allowed. Technical actions must remain hidden or unavailable.

---

## 13. Core matrix — Unknown environment

Unknown environment behavior must remain fail-closed.

Expected behavior:

```text
admin + unknown-env create -> 422 not configured
operator + unknown-env create/action -> rejected or denied
viewer + unknown-env action -> denied
no-role + unknown-env access/action -> denied
```

Historical unknown-env records, if present, may remain readable for authorized users. Technical actions must remain hidden or unavailable.

---

## 14. Evidence requirements

Each validation run must capture:

```text
git HEAD
runtime image
runtime mounts
Environment Catalog live YAML
Cluster Registry live YAML
readyz result
HTTP code for each test case
response body for each test case
summary file
```

Suggested evidence directory naming:

```text
/tmp/dcp-15.5.2-multi-user-multi-environment-runtime-test-YYYYMMDD-HHMMSS
```

Required summary files:

```text
00-date.txt
01-git-log.txt
02-runtime-image-mounts.txt
03-environments-yaml.txt
04-clusters-yaml.txt
05-readyz-code.txt
10-create-matrix-summary.txt
20-read-matrix-summary.txt
30-ui-matrix-summary.txt
40-technical-actions-summary.txt
99-final-summary.txt
```

---

## 15. Recommended test order

Recommended execution order:

```text
1. Confirm repository clean and runtime image/mount baseline.
2. Confirm /readyz is 200.
3. Run admin create matrix across missing/dev/staging/production/unknown-env.
4. Run no-role read and create checks.
5. Run viewer read-only checks.
6. Run operator checks on dev.
7. Run approver lifecycle checks on a prepared dev ChangeRequest.
8. Run admin technical checks on dev.
9. Validate UI action visibility for dev, production disabled and unknown-env historical records.
10. Produce final summary.
```

---

## 16. Acceptance criteria for Phase 15.5 validation execution

The future execution phase is accepted when runtime evidence confirms:

```text
no-role is denied protected access;
viewer cannot execute technical actions;
operator can execute only allowed dev technical actions;
approver can execute only allowed approval actions;
admin can execute allowed dev actions;
staging remains disabled for all roles;
production remains disabled for all roles;
unknown-env remains not configured;
UI does not expose technical actions for disabled or unknown environments;
resolved environment-to-cluster metadata does not change current execution behavior;
working tree remains clean after tests.
```

---

## 17. Out of scope for this validation matrix

The following are out of scope for Phase 15.5.1 and should be handled later:

```text
real staging cluster connection
real production cluster connection
per-cluster Kubernetes token loading
per-cluster CA loading
per-cluster Tekton client construction
per-cluster Argo CD endpoint selection
production enablement
staging enablement
promotion flow execution dev -> staging -> production
```

---

## 18. Next phase

Proceed with:

```text
Phase 15.5.2 — Execute multi-user / multi-environment validation matrix
```

The execution phase should follow this document and produce runtime evidence under `/tmp`.
