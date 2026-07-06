# Target Environment Validation Results

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** Target environment validation results
- **Phase:** 14.2.3 — Document targetEnvironment validation baseline
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Inventory baseline runtime image:** `devops-control-plane:133a496`
- **Validated runtime image:** `devops-control-plane:399f829`
- **Code commit:** `399f829` — `Validate target environment fail closed`
- **Previous related phase:** 14.2.1 — Inventory current `targetEnvironment` behavior
- **Remediation phase:** 14.2.2 — Add fail-closed `targetEnvironment` validation
- **Status:** Completed validation report
- **Language:** English

---

## 1. Purpose

This document records the outcome of Phase 14.2, which introduced and validated the first runtime baseline for fail-closed `targetEnvironment` validation in the DevOps Control Plane.

The goal of this phase was to align the current runtime behavior with the documented multi-environment direction while keeping the implementation deliberately conservative.

The current runtime remains a `dev` baseline. `staging` and `production` are documented future environments, but they are not enabled in runtime workflows until explicit environment catalog, environment-aware authorization, approval policy and promotion guardrails are implemented.

---

## 2. Scope

This validation covered:

- current code behavior for `targetEnvironment`;
- current runtime behavior before remediation;
- the gap between documented fail-closed environment validation and actual behavior;
- implementation of an initial fail-closed runtime validation baseline;
- unit test coverage for allowed and rejected environments;
- build, push and OpenShift rollout of the remediated image;
- runtime retest after deployment.

This phase did not introduce a full environment catalog yet.

This phase did not enable `staging` or `production` workflows.

---

## 3. Background

The DevOps Control Plane documentation now describes a future multi-environment model based on these canonical environments:

```text
dev
staging
production
```

The current runtime, however, operates on the `dev` baseline.

Before this phase, the ChangeRequest domain model already included:

```text
targetEnvironment
```

The UI also displayed the environment associated with a ChangeRequest.

However, runtime validation did not yet enforce a fail-closed list of enabled environments.

---

## 4. Phase 14.2.1 — Inventory current behavior

### 4.1 Repository baseline

Repository HEAD at inventory time:

```text
aa0c4ae Document multi-user validation results
```

Runtime image at inventory time:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:133a496
```

### 4.2 Code behavior before remediation

The service layer trimmed `targetEnvironment` and defaulted an empty value to `dev`:

```go
req.TargetEnvironment = strings.TrimSpace(req.TargetEnvironment)

if req.TargetEnvironment == "" {
    req.TargetEnvironment = "dev"
}
```

The repository layer required `targetEnvironment` to be non-empty, but it did not validate against an allowed environment list.

Therefore, before remediation:

```text
empty targetEnvironment -> defaulted to dev by service layer
non-empty arbitrary value -> accepted by service and repository
```

### 4.3 Runtime inventory tests before remediation

The runtime inventory used backend port-forward and trusted admin headers.

Readiness result:

```text
/readyz -> HTTP 200
```

#### Test A — missing `targetEnvironment`

Input:

```text
targetEnvironment omitted
```

Observed result:

```text
HTTP 201
ChangeRequest=CHG-2026-0010
targetEnvironment=dev
requestedBy=target-env-admin-a
status=draft
```

Conclusion:

```text
Missing targetEnvironment was correctly defaulted to dev.
```

#### Test B — unknown environment

Input:

```text
targetEnvironment=unknown-env
```

Observed result:

```text
HTTP 201
ChangeRequest=CHG-2026-0011
targetEnvironment=unknown-env
requestedBy=target-env-admin-b
status=draft
```

Conclusion:

```text
Unknown targetEnvironment was accepted before remediation.
```

#### Test C — production

Input:

```text
targetEnvironment=production
```

Observed result:

```text
HTTP 201
ChangeRequest=CHG-2026-0012
targetEnvironment=production
requestedBy=target-env-admin-c
status=draft
```

Conclusion:

```text
production was accepted before remediation, even though production workflows were not enabled.
```

### 4.4 Inventory summary before remediation

```text
readyz=200
missing_target_env_http=201
unknown_target_env_http=201
production_target_env_http=201
```

### 4.5 Gap confirmed

The inventory confirmed the expected runtime gap:

```text
missing targetEnvironment -> OK, defaulted to dev
unknown targetEnvironment -> NOT OK, accepted
production targetEnvironment -> NOT OK, accepted while production workflows are disabled
```

---

## 5. Phase 14.2.2 — Fail-closed validation implementation

### 5.1 Commit

Code commit:

```text
399f829 Validate target environment fail closed
```

### 5.2 Files changed

```text
internal/app/change_service.go
internal/app/change_service_target_environment_test.go
```

### 5.3 Implemented behavior

The service layer now keeps backward-compatible defaulting for omitted `targetEnvironment`:

```text
missing targetEnvironment -> dev
```

After defaulting, the service validates the value against an explicit allow-list.

Current allow-list:

```text
dev
```

Rejected until explicitly enabled:

```text
staging
production
unknown-env
any other non-enabled value
```

### 5.4 Implementation detail

The service now performs the following logic:

```go
if req.TargetEnvironment == "" {
    req.TargetEnvironment = "dev"
}
if !isAllowedTargetEnvironment(req.TargetEnvironment) {
    return domain.ChangeRequest{}, errors.New("targetEnvironment must be one of: dev")
}
```

A helper was added:

```go
func isAllowedTargetEnvironment(value string) bool {
    switch value {
    case "dev":
        return true
    default:
        return false
    }
}
```

### 5.5 Unit test coverage

A dedicated test file was added:

```text
internal/app/change_service_target_environment_test.go
```

The test covers:

```text
dev -> allowed
staging -> rejected
production -> rejected
unknown-env -> rejected
empty string -> rejected by helper
```

### 5.6 Code validation

Validation executed successfully:

```text
gofmt OK
go test ./... OK
git diff --check OK
```

### 5.7 Repository state

After commit and push:

```text
HEAD -> main
origin/main
working tree clean
```

---

## 6. Build, push and deployment

### 6.1 Image build

Image tag:

```text
399f829
```

Image build used temporary Podman storage under `/tmp`, consistent with the known bastion `/home` space constraints.

### 6.2 Registry push

The image was pushed to the OpenShift internal registry through the public registry route using:

```text
--authfile ${HOME}/.docker/config.json
--tls-verify=false
```

This is the same controlled lab approach used in previous phases for pushing from the bastion to the OpenShift image registry Route.

### 6.3 Deployment rollout

Deployment image was updated to:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:399f829
```

Rollout completed successfully:

```text
deployment "devops-control-plane" successfully rolled out
```

Runtime image after rollout:

```text
oauth-proxy image=registry.redhat.io/openshift4/ose-oauth-proxy:latest
devops-control-plane image=image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:399f829
```

---

## 7. Runtime retest after remediation

### 7.1 Test setup

Runtime retest used backend port-forward and trusted admin headers.

Readiness result:

```text
/readyz -> HTTP 200
```

### 7.2 Test A — missing `targetEnvironment`

Expected:

```text
HTTP 201
targetEnvironment=dev
```

Observed:

```text
missing_target_env_http=201
```

Conclusion:

```text
Backward-compatible defaulting to dev is preserved.
```

### 7.3 Test B — `targetEnvironment=dev`

Expected:

```text
HTTP 201
targetEnvironment=dev
```

Observed:

```text
dev_target_env_http=201
```

Conclusion:

```text
dev is explicitly allowed.
```

### 7.4 Test C — `targetEnvironment=unknown-env`

Expected:

```text
HTTP 422
```

Observed:

```text
unknown_target_env_http=422
```

Conclusion:

```text
Unknown targetEnvironment values are now rejected fail-closed.
```

### 7.5 Test D — `targetEnvironment=production`

Expected:

```text
HTTP 422
```

Observed:

```text
production_target_env_http=422
```

Conclusion:

```text
production is now rejected until explicitly enabled.
```

### 7.6 Runtime retest summary

```text
readyz=200
missing_target_env_http=201
dev_target_env_http=201
unknown_target_env_http=422
production_target_env_http=422
```

---

## 8. Final runtime behavior

The current runtime behavior is:

```text
allowed:
  dev

default:
  missing targetEnvironment -> dev

rejected:
  staging
  production
  unknown-env
  any non-enabled environment
```

This behavior intentionally favors safety over broad environment availability.

---

## 9. Security and governance rationale

The fail-closed model prevents users or clients from creating ChangeRequests against environments that are not yet explicitly configured and governed.

This is especially important for `production`, because production enablement requires at least:

- explicit environment catalog configuration;
- environment-aware AuthZ;
- approval policy;
- promotion workflow rules;
- evidence and audit expectations;
- operational readiness for production workflows;
- clear separation between `dev`, `staging` and `production` runtime targets.

Until those controls exist, only `dev` is enabled.

---

## 10. Relationship with multi-environment documentation

This implementation aligns runtime behavior with the documented future model in a conservative way.

The documentation describes:

```text
dev -> staging -> production
```

The runtime now enforces:

```text
dev only
```

This is intentional.

The future environment catalog will decide when and how `staging` and `production` become enabled.

---

## 11. Residual notes

### 11.1 Historical records created before remediation

The following records were created during the inventory before fail-closed validation was introduced:

```text
CHG-2026-0011 targetEnvironment=unknown-env
CHG-2026-0012 targetEnvironment=production
```

These records are historical evidence of the pre-remediation behavior.

They should not be interpreted as proof that these environments are enabled after commit `399f829`.

### 11.2 Error message

Current error message:

```text
targetEnvironment must be one of: dev
```

This is accurate for the current runtime allow-list.

When the environment catalog is introduced, the message should be generated from the configured enabled environments.

### 11.3 Staging

`staging` is documented as a future canonical environment, but it is intentionally not enabled yet.

A future phase should introduce `staging` only after environment-specific GitLab, Tekton, Argo CD, Kubernetes/OpenShift and AuthZ mappings are defined and tested.

---

## 12. Recommended next steps

Recommended next phases:

```text
14.3 — Lifecycle vs runtime status UI/API clarity refinement
14.4 — AuthZ runtime validation matrix documentation or runbook alignment
14.5 — UI action state consistency refinement
14.6 — Retry and idempotency policy
14.7 — Evidence UI grouping refinement
14.8 — Environment catalog runtime model
14.9 — Controlled staging enablement
```

The immediate next recommended implementation step is:

```text
Introduce a ConfigMap-driven environment catalog, while keeping only dev enabled by default.
```

---

## 13. Acceptance summary

Phase 14.2 is considered complete because:

- the previous runtime behavior was inventoried;
- the gap was confirmed with runtime evidence;
- fail-closed validation was implemented;
- unit tests were added;
- the code was committed and pushed;
- the image was built and deployed;
- runtime retest confirmed the target behavior;
- only `dev` is currently enabled;
- unknown and production target environments are rejected.
