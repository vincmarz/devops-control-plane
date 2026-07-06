# Environment Catalog and UI Action Visibility Results

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** Environment Catalog and UI action visibility results
- **Phase:** 14.5.3 — Document Environment Catalog and UI action visibility results
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Baseline runtime image before Environment Catalog:** `devops-control-plane:92ca49f`
- **Environment Catalog runtime image:** `devops-control-plane:e3aecdf`
- **UI action visibility runtime image:** `devops-control-plane:041608d`
- **Code commits:**
  - `e3aecdf` — `Add environment catalog baseline`
  - `041608d` — `Use environment catalog for UI action visibility`
- **Related phases:**
  - 14.5.1 — UI action state consistency inventory
  - 14.5.2a — Environment Catalog runtime baseline
  - 14.5.2b — Use Environment Catalog semantics for UI action visibility
- **Status:** Completed validation report
- **Language:** English

---

## 1. Purpose

This document records the outcome of Phase 14.5, which refined UI action visibility using Environment Catalog semantics.

The purpose of this phase was to remove the previous mismatch where the UI displayed and exposed technical action paths for historical ChangeRequests whose `targetEnvironment` was no longer valid or enabled.

The phase introduced a first Environment Catalog runtime baseline and then used the same catalog semantics to control the ChangeRequest detail UI technical action section.

The core outcome is:

```text
Enabled environment:
  technical actions visible

Configured but disabled environment:
  warning visible
  technical actions hidden

Not configured environment:
  warning visible
  technical actions hidden
```

---

## 2. Scope

This phase covered:

- inventory of current UI action rendering;
- validation of action behavior by role;
- creation of an in-code Environment Catalog baseline;
- repository manifest for an Environment Catalog ConfigMap;
- runtime validation of target environment semantics;
- UI action visibility based on Environment Catalog semantics;
- runtime validation of UI behavior for enabled, disabled and not configured environments.

This phase did not implement dynamic runtime loading from the Environment Catalog ConfigMap.

This phase did not enable `staging` or `production`.

This phase did not change the technical execution logic for GitLab, Tekton, Argo CD or Kubernetes/OpenShift integrations.

---

## 3. Background

The DevOps Control Plane already had fail-closed `targetEnvironment` validation from Phase 14.2.

Before the Environment Catalog baseline, that validation was intentionally simple:

```text
allowed:
  dev

rejected:
  staging
  production
  unknown-env
```

However, the earlier implementation did not distinguish between:

```text
configured but disabled
not configured
```

This distinction became important after UI inventory found historical records such as:

```text
CHG-2026-0011 -> targetEnvironment=unknown-env
CHG-2026-0012 -> targetEnvironment=production
```

Those records were created before fail-closed validation was introduced. The UI still displayed technical actions for them, even though their target environments were not currently enabled.

---

## 4. Phase 14.5.1 — UI action state consistency inventory

### 4.1 Inventory objective

Phase 14.5.1 checked whether UI technical actions were consistent with:

- process lifecycle status;
- technical runtime status;
- user role;
- target environment validity;
- historical records created before environment validation was hardened.

### 4.2 Runtime baseline

The inventory used runtime image:

```text
devops-control-plane:92ca49f
```

Repository HEAD at that time was:

```text
ef7fbb9 Document environment visibility UI results
```

### 4.3 Sample ChangeRequests

The inventory used the following sample records:

```text
CHG-2026-0001 -> targetEnvironment=dev, runtimeStatus=EvidenceCollected
CHG-2026-0010 -> targetEnvironment=dev
CHG-2026-0011 -> targetEnvironment=unknown-env, historical pre-fix record
CHG-2026-0012 -> targetEnvironment=production, historical pre-fix record
```

### 4.4 API status and environment summary

Observed API state:

```text
changeNumber=CHG-2026-0001
status=draft
runtimeStatus=EvidenceCollected
targetEnvironment=dev
requestedBy=vmarzario
title=OpenShift runtime smoke test
END=CHG-2026-0001

changeNumber=CHG-2026-0010
status=draft
runtimeStatus=
targetEnvironment=dev
requestedBy=target-env-admin-a
title=Target environment validation test A
END=CHG-2026-0010

changeNumber=CHG-2026-0011
status=draft
runtimeStatus=
targetEnvironment=unknown-env
requestedBy=target-env-admin-b
title=Target environment validation test B
END=CHG-2026-0011

changeNumber=CHG-2026-0012
status=draft
runtimeStatus=
targetEnvironment=production
requestedBy=target-env-admin-c
title=Target environment validation test C
END=CHG-2026-0012
```

### 4.5 UI action rendering before refinement

The inventory showed that all four ChangeRequest detail pages displayed the same technical action sections and action paths.

For `CHG-2026-0001`:

```text
Recommended next actions=True
Advanced/manual actions=True
check-validation=True
collect-evidence=True
create-branch=True
update-files=True
open-merge-request=True
merge-request=True
check-deployment=True
```

For `CHG-2026-0010`:

```text
Recommended next actions=True
Advanced/manual actions=True
check-validation=True
collect-evidence=True
create-branch=True
update-files=True
open-merge-request=True
merge-request=True
check-deployment=True
```

For historical `CHG-2026-0011` with `unknown-env`:

```text
Recommended next actions=True
Advanced/manual actions=True
check-validation=True
collect-evidence=True
create-branch=True
update-files=True
open-merge-request=True
merge-request=True
check-deployment=True
```

For historical `CHG-2026-0012` with `production`:

```text
Recommended next actions=True
Advanced/manual actions=True
check-validation=True
collect-evidence=True
create-branch=True
update-files=True
open-merge-request=True
merge-request=True
check-deployment=True
```

### 4.6 UI action behavior by role on historical production record

The inventory also tested the `check-validation` UI action on `CHG-2026-0012`.

Observed result:

```text
admin_production_check_validation=303
operator_production_check_validation=303
viewer_production_check_validation=403
```

Interpretation:

```text
Viewer remained correctly blocked by AuthZ.
Admin and operator could reach the UI technical action handler.
```

This confirmed that role-based AuthZ was still working, but UI action visibility was not aware of whether the target environment was currently enabled.

### 4.7 Gap confirmed

The confirmed gap was:

```text
The UI displayed and exposed technical actions for historical ChangeRequests whose targetEnvironment was not currently enabled or not configured.
```

This gap was not primarily an AuthZ issue.

The issue was UI state consistency with environment readiness.

---

## 5. Phase 14.5.2a — Environment Catalog runtime baseline

### 5.1 Commit

Code commit:

```text
e3aecdf Add environment catalog baseline
```

### 5.2 Files changed

```text
internal/app/change_service.go
internal/app/environment_catalog.go
internal/app/change_service_target_environment_test.go
manifests/configmap-environments.yaml
```

### 5.3 Environment Catalog baseline

The first Environment Catalog baseline is implemented in code through:

```text
internal/app/environment_catalog.go
```

The baseline catalog contains:

```text
dev:
  configured: true
  enabled: true
  allowTechnicalActions: true

staging:
  configured: true
  enabled: false
  allowTechnicalActions: false

production:
  configured: true
  enabled: false
  allowTechnicalActions: false
```

Unknown values such as `unknown-env` are treated as not configured.

### 5.4 Repository ConfigMap manifest

A repository baseline manifest was added:

```text
manifests/configmap-environments.yaml
```

The ConfigMap name is:

```text
devops-control-plane-environments
```

The manifest contains an `environments.yaml` key with `dev`, `staging` and `production` definitions.

The manifest was validated using:

```text
oc apply --dry-run=client -f manifests/configmap-environments.yaml
```

Observed result:

```text
configmap/devops-control-plane-environments created (dry run)
```

### 5.5 Create ChangeRequest validation behavior

`ChangeService.Create` was updated so that:

```text
missing targetEnvironment -> EnvironmentCatalog.DefaultEnvironment()
targetEnvironment value -> EnvironmentCatalog.ValidateCreateTargetEnvironment(value)
```

The old hardcoded `dev`-only helper was replaced by catalog semantics.

### 5.6 Code validation

Validation before commit:

```text
gofmt OK
go test ./... OK
git diff --check OK
oc apply --dry-run=client -f manifests/configmap-environments.yaml OK
```

---

## 6. Runtime validation for Environment Catalog baseline

### 6.1 Runtime image

The Environment Catalog baseline was built and deployed as:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:e3aecdf
```

### 6.2 Readiness

Runtime readiness:

```text
/readyz -> HTTP 200
```

### 6.3 Missing `targetEnvironment`

Request without `targetEnvironment` returned:

```text
HTTP 201
changeNumber=CHG-2026-0016
targetEnvironment=dev
requestedBy=env-catalog-admin-a
status=draft
```

Conclusion:

```text
The default environment is resolved through the catalog and defaults to dev.
```

### 6.4 `targetEnvironment=dev`

Request with `targetEnvironment=dev` returned:

```text
HTTP 201
changeNumber=CHG-2026-0017
targetEnvironment=dev
requestedBy=env-catalog-admin-b
status=draft
```

Conclusion:

```text
dev is configured and enabled.
```

### 6.5 `targetEnvironment=staging`

Request with `targetEnvironment=staging` returned:

```text
HTTP 422
errorCode=VALIDATION_INVALID_REQUEST
technicalMessage=targetEnvironment "staging" is currently disabled
```

Conclusion:

```text
staging is configured but disabled.
```

### 6.6 `targetEnvironment=production`

Request with `targetEnvironment=production` returned:

```text
HTTP 422
errorCode=VALIDATION_INVALID_REQUEST
technicalMessage=targetEnvironment "production" is currently disabled
```

Conclusion:

```text
production is configured but disabled.
```

### 6.7 `targetEnvironment=unknown-env`

Request with `targetEnvironment=unknown-env` returned:

```text
HTTP 422
errorCode=VALIDATION_INVALID_REQUEST
technicalMessage=targetEnvironment "unknown-env" is not configured
```

Conclusion:

```text
unknown-env is not configured.
```

### 6.8 Runtime validation summary

```text
readyz=200
missing_target_env_http=201
dev_target_env_http=201
staging_target_env_http=422
production_target_env_http=422
unknown_target_env_http=422
```

The runtime distinction between enabled, configured but disabled, and not configured environments was confirmed.

---

## 7. Phase 14.5.2b — UI action visibility using Environment Catalog semantics

### 7.1 Commit

Code commit:

```text
041608d Use environment catalog for UI action visibility
```

### 7.2 File changed

```text
internal/api/ui_handlers.go
```

### 7.3 UI implementation summary

The UI handler now imports the application package as:

```text
appsvc
```

The UI template function map now registers:

```text
environmentAllowsTechnicalActions
environmentActionWarning
```

The ChangeRequest detail page `Technical actions` section is now rendered according to:

```text
environmentAllowsTechnicalActions .SelectedChange
```

If the current ChangeRequest environment does not allow technical actions, the UI shows a warning and a technical actions unavailable message.

### 7.4 Warning semantics

The UI now differentiates:

```text
Target environment is empty.
Target environment <name> is not configured.
Target environment <name> is currently disabled.
Target environment <name> does not allow technical actions.
```

For historical records in the current dataset, this results in:

```text
Target environment unknown-env is not configured.
Target environment production is currently disabled.
```

### 7.5 Code validation

Validation before commit:

```text
gofmt OK
go test ./... OK
git diff --check OK
class="sys" sidebar check OK
```

---

## 8. Runtime validation for UI action visibility

### 8.1 Runtime image

The UI action visibility refinement was built and deployed as:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:041608d
```

### 8.2 HTTP summary

Runtime summary:

```text
readyz=200
ui_0001_http=200
ui_0010_http=200
ui_0011_http=200
ui_0012_http=200
```

### 8.3 `CHG-2026-0001` — dev

Observed UI check:

```text
Technical actions=True
Recommended next actions=True
Advanced/manual actions=True
check-validation=True
collect-evidence=True
create-branch=True
update-files=True
open-merge-request=True
merge-request=True
check-deployment=True
Target environment unknown-env is not configured=False
Target environment production is currently disabled=False
Technical actions are not available because the target environment is not currently enabled for automation.=False
```

Conclusion:

```text
Technical actions remain available for dev.
```

### 8.4 `CHG-2026-0010` — dev

Observed UI check:

```text
Technical actions=True
Recommended next actions=True
Advanced/manual actions=True
check-validation=True
collect-evidence=True
create-branch=True
update-files=True
open-merge-request=True
merge-request=True
check-deployment=True
Target environment unknown-env is not configured=False
Target environment production is currently disabled=False
Technical actions are not available because the target environment is not currently enabled for automation.=False
```

Conclusion:

```text
Technical actions remain available for dev records.
```

### 8.5 `CHG-2026-0011` — unknown-env historical record

Observed UI check:

```text
Technical actions=True
Recommended next actions=False
Advanced/manual actions=False
check-validation=False
collect-evidence=False
create-branch=False
update-files=False
open-merge-request=False
merge-request=False
check-deployment=False
Target environment unknown-env is not configured=True
Target environment production is currently disabled=False
Technical actions are not available because the target environment is not currently enabled for automation.=True
```

Conclusion:

```text
The UI identifies unknown-env as not configured and hides technical actions.
```

### 8.6 `CHG-2026-0012` — production historical record

Observed UI check:

```text
Technical actions=True
Recommended next actions=False
Advanced/manual actions=False
check-validation=False
collect-evidence=False
create-branch=False
update-files=False
open-merge-request=False
merge-request=False
check-deployment=False
Target environment unknown-env is not configured=False
Target environment production is currently disabled=True
Technical actions are not available because the target environment is not currently enabled for automation.=True
```

Conclusion:

```text
The UI identifies production as configured but currently disabled and hides technical actions.
```

---

## 9. Final behavior after Phase 14.5

### 9.1 Create ChangeRequest behavior

```text
missing targetEnvironment -> HTTP 201, default dev
targetEnvironment=dev -> HTTP 201
targetEnvironment=staging -> HTTP 422, currently disabled
targetEnvironment=production -> HTTP 422, currently disabled
targetEnvironment=unknown-env -> HTTP 422, not configured
```

### 9.2 UI action visibility behavior

```text
dev:
  technical actions visible
  recommended actions visible
  advanced/manual actions visible

production:
  warning visible
  technical actions unavailable message visible
  recommended actions hidden
  advanced/manual actions hidden
  technical action paths hidden

unknown-env:
  warning visible
  technical actions unavailable message visible
  recommended actions hidden
  advanced/manual actions hidden
  technical action paths hidden
```

---

## 10. Security and governance rationale

This phase improves safety and governance by preventing the UI from recommending or exposing technical workflow entry points for environments that are not currently enabled.

This avoids accidental operator confusion around historical records.

The implementation also prepares the project for future controlled environment enablement.

Instead of checking only whether a value equals `dev`, the project now uses catalog semantics:

```text
configured
not configured
enabled
disabled
allowTechnicalActions
```

This gives a cleaner path for future staging enablement.

---

## 11. Residual notes

### 11.1 Runtime ConfigMap loading is not implemented yet

The repository now includes:

```text
manifests/configmap-environments.yaml
```

However, Phase 14.5.2a uses an in-code baseline catalog.

A future phase can load the catalog from this ConfigMap or from a mounted file.

### 11.2 Technical backend actions are still protected by AuthZ

The UI now hides action paths for disabled and not configured environments.

Backend AuthZ remains in place and continues to enforce role-based access.

Future backend service-layer checks may also use Environment Catalog semantics for defense in depth.

### 11.3 Historical records remain readable

Historical records remain readable in dashboard, list and detail views.

The UI does not hide the ChangeRequest itself.

The UI only hides technical action entry points when the environment is not currently suitable for automation.

---

## 12. Recommended next steps

Recommended next work items:

```text
14.5.4 — Document Phase 14.5 closure or include this document as closure
14.6 — Runtime Environment Catalog ConfigMap loading
14.7 — Controlled staging enablement design
14.8 — Environment-aware AuthZ and approval policies
14.9 — Production enablement gates
```

A practical next engineering step is:

```text
Load Environment Catalog from a mounted ConfigMap with fail-closed fallback.
```

The recommended production-safe behavior is:

```text
If the runtime catalog cannot be loaded, fall back to the conservative in-code catalog:
  dev enabled
  staging disabled
  production disabled
```

---

## 13. Acceptance summary

Phase 14.5 is considered complete because:

- UI action rendering was inventoried;
- the gap for historical disabled or not configured environments was confirmed;
- the Environment Catalog baseline was introduced;
- create-time environment validation now distinguishes enabled, disabled and not configured environments;
- UI action visibility now uses Environment Catalog semantics;
- runtime validation confirmed expected behavior for `dev`, `production` and `unknown-env`;
- no additional environment was enabled by accident;
- historical records remain readable but no longer expose technical actions when their target environment is not available for automation.
