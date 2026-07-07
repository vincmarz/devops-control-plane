# 15.8.6.5 — Argo CD Runtime Provider Wiring Results

## 1. Purpose

This document records the implementation and runtime validation results for **Phase 15.8.6 — Argo CD runtime client provider per cluster**.

The goal of this phase was to make the Argo CD deployment-check workflow provider-aware while preserving the existing safe runtime baseline:

```text
dev        -> enabled and operational
staging    -> configured but disabled
production -> configured but disabled
```

The phase follows the same incremental provider-aware approach already completed for Kubernetes and Tekton.

No real staging or production Argo CD clients were created in this phase. No Kubernetes Secret values were read. No staging or production enablement was performed.

---

## 2. Starting point

Before Phase 15.8.6, the project had already completed:

```text
15.8.1 — Multi-cluster secret/config model
15.8.2 — Runtime client Secret reference loading
15.8.3 — Wire Secret reference registry into provider preparation
15.8.4 — Kubernetes client provider per cluster
15.8.5 — Tekton runtime client provider per cluster
```

At that point, the following runtime workflows were already provider-aware:

```text
Kubernetes collect-evidence
Tekton validate
Tekton check-validation
```

The remaining major technical workflow still using a direct current/global runtime client was:

```text
Argo CD check-deployment
```

---

## 3. Phase scope

### 3.1 Completed sub-phases

```text
15.8.6.1 — Inventory Argo CD adapter constructor and runtime usage      COMPLETED
15.8.6.2 — Define Argo CD runtime client provider abstraction           COMPLETED
15.8.6.3 — Wire Argo CD provider into CheckDeployment                   COMPLETED
15.8.6.4 — Runtime validation for Argo CD runtime provider wiring       COMPLETED
15.8.6.5 — Document Argo CD runtime provider wiring results             COMPLETED
```

### 3.2 Out of scope

The following items were intentionally kept out of scope:

```text
creating real Argo CD clients for staging
creating real Argo CD clients for production
reading Kubernetes Secret values
reading Argo CD token values from Kubernetes Secrets
printing token values
introducing staging credentials
introducing production credentials
enabling staging
enabling production
changing OAuth Proxy configuration
changing OpenShift RBAC
changing Tekton provider behavior
changing Kubernetes provider behavior
```

This phase focused only on introducing and validating the provider-aware Argo CD baseline for the current dev runtime client.

---

## 4. 15.8.6.1 — Inventory results

### 4.1 Summary

The Argo CD inventory produced the following summary:

```text
git_head=3040366 Document Tekton runtime provider wiring results
argocd_new_refs=1
check_deployment_refs=16
get_application_refs=8
argocd_adapter_new_refs=1
argocd_client_refs=4
with_argocd_check_deployment_refs=3
argocd_application_name_refs=11
provider_selection_refs=45
```

### 4.2 Argo CD adapter constructor

The Argo CD adapter configuration was identified as:

```text
internal/adapters/argocd/models.go:5
type Config struct
```

The constructor was identified as:

```text
internal/adapters/argocd/client.go:24
func New(cfg Config) (*Client, error)
```

### 4.3 Existing single-client wiring

The current Argo CD client was created once in the composition root:

```text
cmd/devops-control-plane/main.go:185
argocdadapter.New(argocdadapter.Config{
  BaseURL: cfg.ArgoCDBaseURL,
  AuthToken: cfg.ArgoCDAuthToken,
  TimeoutSeconds: cfg.ArgoCDTimeoutSeconds,
  InsecureTLS: cfg.ArgoCDInsecureTLS,
  CAFile: cfg.ArgoCDCAFile,
})
```

This confirmed that Argo CD was still using:

```text
single runtime client
global Argo CD BaseURL
global Argo CD AuthToken
global timeout / TLS / CA settings
```

### 4.4 CheckDeployment before provider wiring

The ChangeService used the application-level port:

```text
type ArgoCDCheckDeploymentFunc func(ctx context.Context, change domain.ChangeRequest) (ArgoCDDeploymentResult, error)
```

The `main.go` wiring used the concrete current Argo CD client directly:

```text
argoCDClient.GetApplication(ctx, change.ApplicationName)
```

Observed references:

```text
cmd/devops-control-plane/main.go:192
cmd/devops-control-plane/main.go:196
```

Therefore, before provider wiring, `CheckDeployment` still used:

```text
change.ApplicationName
```

instead of:

```text
TechnicalRuntimeTarget.ArgoCDApplicationName
```

### 4.5 TechnicalRuntimeTarget readiness

The target model already contained:

```text
ArgoCDApplicationName
```

Observed references included:

```text
internal/app/environment_catalog.go
internal/app/technical_runtime_target.go
internal/app/technical_runtime_target_test.go
```

The dev baseline configured:

```text
ArgoCDApplicationName=demo-go-color-app
```

### 4.6 Provider selection readiness

`CheckDeployment` already passed through:

```text
resolveRuntimeClientProviderSelection(ctx, change)
```

This made the workflow ready to receive an Argo CD runtime provider registry.

---

## 5. 15.8.6.2 — Argo CD runtime client provider abstraction

### 5.1 Commit

The Argo CD provider abstraction was introduced with:

```text
b631be6 Add Argo CD runtime client provider abstraction
```

### 5.2 Files added

```text
internal/app/argocd_runtime_client_provider.go
internal/app/argocd_runtime_client_provider_test.go
```

### 5.3 Abstractions introduced

The following app-layer abstractions were added:

```text
ArgoCDRuntimeClient
ArgoCDRuntimeClientProvider
ArgoCDRuntimeClientProviderRegistry
CurrentClusterArgoCDRuntimeClientProvider
```

### 5.4 Default behavior

The default registry is conservative:

```text
ocp-dev -> current Argo CD runtime client, if configured
ocp-staging -> provider not configured
ocp-production -> provider not configured
```

If no current Argo CD runtime client is configured, the default Argo CD provider registry remains empty.

### 5.5 Validation

The code was validated with:

```text
go test ./... OK
git diff --check OK
```

---

## 6. 15.8.6.3 — Argo CD provider wiring into CheckDeployment

### 6.1 Commit

The `CheckDeployment` wiring was introduced with:

```text
cf6463e Wire Argo CD runtime provider into deployment check
```

### 6.2 Files modified or added

```text
cmd/devops-control-plane/main.go
cmd/devops-control-plane/argocd_runtime_client_adapter.go
cmd/devops-control-plane/argocd_runtime_client_adapter_test.go
cmd/devops-control-plane/main_argocd_provider_check_deployment_wiring_test.go
```

### 6.3 Adapter bridge

A bridge named:

```text
currentArgoCDRuntimeClient
```

was introduced to adapt the concrete Argo CD adapter to the app-layer interface:

```text
ArgoCDRuntimeClient
```

### 6.4 Provider-aware CheckDeployment chain

The `CheckDeployment` workflow now follows:

```text
CheckDeployment
  -> TechnicalRuntimeTarget
  -> RuntimeClientProviderSelection
  -> ArgoCDRuntimeClientProviderRegistry
  -> current Argo CD runtime client
  -> CheckDeployment(ctx, target.ArgoCDApplicationName)
```

### 6.5 Target-aware application name

The workflow now uses:

```text
target.ArgoCDApplicationName
```

instead of:

```text
change.ApplicationName
```

For the current dev baseline, both still resolve to:

```text
demo-go-color-app
```

### 6.6 Validation

The code path was validated with:

```text
go test ./... OK
git diff --check OK
```

---

## 7. 15.8.6.4 — Runtime validation

### 7.1 Runtime image

Runtime validation was executed against image:

```text
cf6463e Wire Argo CD runtime provider into deployment check
```

Expected deployed image shape:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:cf6463e
```

### 7.2 Readiness

Readiness passed:

```text
readyz=200
```

### 7.3 Create matrix

The create matrix produced:

```text
create_dev_http=201
create_staging_http=422
create_production_http=422
argocd_provider_change_number=CHG-2026-0031
```

Interpretation:

```text
dev remains operational
staging remains disabled
production remains disabled
```

### 7.4 ChangeRequest created for runtime validation

The runtime validation created:

```text
CHG-2026-0031
```

The create-dev response summary was:

```text
CASE=create-dev
changeNumber=CHG-2026-0031
targetEnvironment=dev
requestedBy=argocd-provider-admin-dev
status=draft
runtimeStatus=
applicationName=demo-go-color-app
syncStatus=
healthStatus=
revision=
errorCode=
technicalMessage=
END_CASE=create-dev
```

### 7.5 Disabled environment validation

Staging remained disabled:

```text
CASE=create-staging
errorCode=VALIDATION_INVALID_REQUEST
technicalMessage=targetEnvironment "staging" is currently disabled
END_CASE=create-staging
```

Production remained disabled:

```text
CASE=create-production
errorCode=VALIDATION_INVALID_REQUEST
technicalMessage=targetEnvironment "production" is currently disabled
END_CASE=create-production
```

---

## 8. CheckDeployment runtime matrix

Endpoint tested:

```text
POST /api/v1/changes/CHG-2026-0031/check-deployment
```

### 8.1 viewer

Observed:

```text
check_deployment_viewer_http=403
body_prefix=forbidden: insufficient role
```

Interpretation:

```text
viewer remains denied.
```

### 8.2 operator

Observed:

```text
check_deployment_operator_http=202
runtimeStatus=DeploymentSyncedHealthy
```

Response summary:

```text
CASE=check-deployment-operator
changeNumber=CHG-2026-0031
status=draft
runtimeStatus=DeploymentSyncedHealthy
errorCode=
technicalMessage=
END_CASE=check-deployment-operator
```

Interpretation:

```text
operator remains allowed and Argo CD deployment status check succeeds through the provider-aware path.
```

### 8.3 approver

Observed:

```text
check_deployment_approver_http=403
body_prefix=forbidden: insufficient role
```

Interpretation:

```text
approver remains denied for technical deployment check execution.
```

### 8.4 admin

Observed:

```text
check_deployment_admin_http=202
runtimeStatus=DeploymentSyncedHealthy
```

Response summary:

```text
CASE=check-deployment-admin
changeNumber=CHG-2026-0031
status=draft
runtimeStatus=DeploymentSyncedHealthy
errorCode=
technicalMessage=
END_CASE=check-deployment-admin
```

Interpretation:

```text
admin remains allowed and Argo CD deployment status check succeeds through the provider-aware path.
```

---

## 9. Final HTTP summary

The final runtime validation summary was:

```text
readyz=200
create_dev_http=201
create_staging_http=422
create_production_http=422
argocd_provider_change_number=CHG-2026-0031
check_deployment_viewer_http=403
check_deployment_operator_http=202
check_deployment_approver_http=403
check_deployment_admin_http=202
```

This matches the expected no-regression behavior.

---

## 10. Acceptance criteria

### AC-15.8.6-001 — Argo CD provider abstraction exists

Status:

```text
PASSED
```

Evidence:

```text
ArgoCDRuntimeClient
ArgoCDRuntimeClientProvider
ArgoCDRuntimeClientProviderRegistry
CurrentClusterArgoCDRuntimeClientProvider
```

### AC-15.8.6-002 — CheckDeployment uses provider-aware path

Status:

```text
PASSED
```

Evidence:

```text
DefaultArgoCDRuntimeClientProviderRegistry(...).Resolve(ctx, selection)
argoCDRuntimeClient.CheckDeployment(ctx, target.ArgoCDApplicationName)
```

### AC-15.8.6-003 — dev create remains operational

Status:

```text
PASSED
```

Evidence:

```text
create_dev_http=201
argocd_provider_change_number=CHG-2026-0031
```

### AC-15.8.6-004 — staging remains disabled

Status:

```text
PASSED
```

Evidence:

```text
create_staging_http=422
targetEnvironment "staging" is currently disabled
```

### AC-15.8.6-005 — production remains disabled

Status:

```text
PASSED
```

Evidence:

```text
create_production_http=422
targetEnvironment "production" is currently disabled
```

### AC-15.8.6-006 — CheckDeployment RBAC preserved

Status:

```text
PASSED
```

Evidence:

```text
viewer=403
operator=202
approver=403
admin=202
```

### AC-15.8.6-007 — Deployment status remains operational

Status:

```text
PASSED
```

Evidence:

```text
operator runtimeStatus=DeploymentSyncedHealthy
admin runtimeStatus=DeploymentSyncedHealthy
```

---

## 11. Security assessment

The phase preserved the expected security guardrails:

```text
no Secret values read
no Argo CD token values printed
no kubeconfig values printed
no staging credentials introduced
no production credentials introduced
no staging enablement
no production enablement
```

The Argo CD provider currently maps only:

```text
ocp-dev -> current Argo CD runtime client
```

Real staging and production Argo CD clients remain future work.

---

## 12. Current state after 15.8.6

The following runtime workflows are now provider-aware for the current dev cluster:

```text
Kubernetes collect-evidence
Tekton validate
Tekton check-validation
Argo CD check-deployment
```

Current supported behavior:

```text
ocp-dev:
  Kubernetes provider configured through current runtime client
  Tekton provider configured through current runtime client
  Argo CD provider configured through current runtime client

ocp-staging:
  environment disabled
  runtime providers not configured for real use

ocp-production:
  environment disabled
  runtime providers not configured for real use
```

---

## 13. Recommended next phase

Recommended next step:

```text
15.8.7 — Real multi-cluster client factory preparation
```

Expected scope:

```text
introduce safe runtime client factory contracts
load Secret references without printing values
define explicit enablement gates for staging and production
keep staging/production disabled until real credentials and validation are available
```

Alternative next step if documentation consolidation is preferred:

```text
15.8.6.6 — Document full provider-aware runtime architecture summary
```

---

## 14. Final conclusion

Phase 15.8.6 successfully introduced and validated Argo CD runtime provider wiring.

The DevOps Control Plane now has provider-aware deployment checks:

```text
CheckDeployment
  -> TechnicalRuntimeTarget
  -> RuntimeClientProviderSelection
  -> ArgoCDRuntimeClientProviderRegistry
  -> ArgoCDRuntimeClient
  -> CheckDeployment(target.ArgoCDApplicationName)
```

Runtime validation confirmed no regression:

```text
readyz=200
dev create=201
staging=422
production=422
check-deployment viewer=403
check-deployment operator=202
check-deployment approver=403
check-deployment admin=202
runtimeStatus=DeploymentSyncedHealthy
```

This completes the provider-aware baseline for Kubernetes, Tekton and Argo CD on `ocp-dev/current provider` and prepares the project for the next step: real multi-cluster client factory and controlled staging/production enablement.
