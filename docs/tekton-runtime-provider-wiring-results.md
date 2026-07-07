# 15.8.5.6 — Tekton Runtime Provider Wiring Results

## 1. Purpose

This document records the implementation and runtime validation results for **Phase 15.8.5 — Tekton runtime client provider per cluster**.

The purpose of this phase was to make the Tekton validation workflows provider-aware while preserving the current safe runtime behavior:

```text
dev        -> enabled and operational
staging    -> configured but disabled
production -> configured but disabled
```

The phase follows the same staged approach already used for the Kubernetes provider baseline. The implementation introduces Tekton provider abstractions and wires the `Validate` and `CheckValidation` workflows through the runtime target and provider-selection chain.

No real staging or production Tekton clients are created in this phase. No Kubernetes Secret values are read.

---

## 2. Starting point

Before Phase 15.8.5, the project had already completed:

```text
15.8.1 — Multi-cluster secret/config model
15.8.2 — Runtime client Secret reference loading
15.8.3 — Wire Secret reference registry into provider preparation
15.8.4 — Kubernetes client provider per cluster
```

The Kubernetes runtime evidence path had already become provider-aware through:

```text
TechnicalRuntimeTarget
  -> RuntimeClientProviderSelection
  -> KubernetesRuntimeClientProviderRegistry
  -> KubernetesRuntimeEvidenceClient
```

The Tekton workflows were still using the globally configured Tekton client and runtime configuration values in the composition root.

---

## 3. Phase scope

### 3.1 Completed sub-phases

```text
15.8.5.1 — Inventory Tekton adapter constructor and runtime usage     COMPLETED
15.8.5.2 — Define Tekton runtime client provider abstraction          COMPLETED
15.8.5.3 — Wire Tekton provider into Validate                         COMPLETED
15.8.5.4 — Wire Tekton provider into CheckValidation                  COMPLETED
15.8.5.5 — Runtime validation for Tekton runtime provider wiring      COMPLETED
15.8.5.6 — Document Tekton runtime provider wiring results            COMPLETED
```

### 3.2 Out of scope

The following items were intentionally out of scope for this phase:

```text
creating real Tekton clients for staging
creating real Tekton clients for production
reading Kubernetes Secret values
reading service account token values
printing token values or kubeconfig values
enabling staging
enabling production
changing OAuth Proxy configuration
changing OpenShift RBAC
changing GitLab workflows
changing Argo CD provider behavior
```

This phase focused only on making the existing dev/current Tekton client reachable through a provider-aware chain.

---

## 4. 15.8.5.1 — Inventory results

### 4.1 Summary

The Tekton inventory produced the following summary:

```text
git_head=315acfb Document Kubernetes runtime provider wiring results
tekton_new_refs=1
run_pipeline_refs=10
find_latest_pipeline_run_refs=9
list_taskruns_refs=6
tekton_adapter_new_refs=1
tekton_client_refs=4
cfg_tekton_namespace_refs=4
provider_selection_refs=36
```

### 4.2 Tekton adapter constructor

The Tekton adapter constructor was identified as:

```text
internal/adapters/tekton/client.go:42
func New(cfg Config, opts ...Option) (*Client, error)
```

The adapter configuration was identified as:

```text
internal/adapters/tekton/client.go:21
type Config struct
```

### 4.3 Existing single-client wiring

The Tekton client was created once in the composition root:

```text
cmd/devops-control-plane/main.go:109
tektonadapter.New(tektonadapter.Config{
  APIURL: cfg.KubernetesAPIURL,
  Token: cfg.KubernetesToken,
  TimeoutSeconds: cfg.TektonTimeoutSeconds,
  InsecureTLS: cfg.KubernetesInsecureTLS,
  CAFile: cfg.KubernetesCAFile,
})
```

This confirmed that Tekton was still using:

```text
single runtime client
single Kubernetes API URL/token/CA configuration
single runtime namespace
```

### 4.4 Validate workflow before provider wiring

The `Validate` workflow used:

```text
cfg.TektonNamespace
cfg.TektonPipelineName
cfg.TektonServiceAccount
cfg.TektonGitURL
cfg.TektonValidationPath
```

The relevant inventory references were:

```text
cmd/devops-control-plane/main.go:114
WithTektonRunPipeline(...)

cmd/devops-control-plane/main.go:121
Namespace: cfg.TektonNamespace

cmd/devops-control-plane/main.go:122
PipelineName: cfg.TektonPipelineName
```

### 4.5 CheckValidation workflow before provider wiring

`CheckValidation` used the global Tekton namespace:

```text
cmd/devops-control-plane/main.go:137
tektonClient.FindLatestPipelineRunByChange(ctx, cfg.TektonNamespace, change.ChangeNumber)

cmd/devops-control-plane/main.go:146
tektonClient.ListTaskRunsByPipelineRun(ctx, cfg.TektonNamespace, status.Name)
```

### 4.6 Application layer readiness

The ChangeService already had Tekton application ports:

```text
type TektonRunPipelineFunc func(ctx context.Context, change domain.ChangeRequest) (string, string, string, error)
type TektonCheckValidationFunc func(ctx context.Context, change domain.ChangeRequest) (TektonValidationResult, error)
```

The technical workflows were already resolving provider selection through:

```text
resolveRuntimeClientProviderSelection(ctx, change)
```

This made the Tekton workflow suitable for provider-aware wiring.

---

## 5. 15.8.5.2 — Tekton runtime client provider abstraction

### 5.1 Commit

The Tekton provider abstraction was introduced with:

```text
53cec68 Add Tekton runtime client provider abstraction
```

### 5.2 Files added

```text
internal/app/tekton_runtime_client_provider.go
internal/app/tekton_runtime_client_provider_test.go
```

### 5.3 Abstractions introduced

The following application-layer abstractions were added:

```text
TektonRuntimePipelineRunRequest
TektonRuntimePipelineRunRef
TektonRuntimeClient
TektonRuntimeClientProvider
TektonRuntimeClientProviderRegistry
CurrentClusterTektonRuntimeClientProvider
```

### 5.4 Default behavior

The default registry is conservative:

```text
ocp-dev -> current Tekton runtime client, if configured
ocp-staging -> provider not configured
ocp-production -> provider not configured
```

If no current Tekton runtime client is configured, the default Tekton provider registry remains empty.

### 5.5 Safety posture

The abstraction does not read Secret values and does not build real multi-cluster clients. It only defines the app-layer boundary used by later wiring.

---

## 6. 15.8.5.3 — Tekton provider wiring into Validate

### 6.1 Commit

The Validate workflow was wired with:

```text
b8a430a Wire Tekton runtime provider into validation
```

### 6.2 Files modified or added

```text
cmd/devops-control-plane/main.go
internal/app/tekton_runtime_client_provider.go
cmd/devops-control-plane/main_tekton_provider_validate_wiring_test.go
cmd/devops-control-plane/tekton_runtime_client_adapter.go
cmd/devops-control-plane/tekton_runtime_client_adapter_test.go
```

### 6.3 Validate provider-aware chain

The `Validate` workflow now follows:

```text
Validate
  -> TechnicalRuntimeTarget
  -> RuntimeClientProviderSelection
  -> TektonRuntimeClientProviderRegistry
  -> current Tekton runtime client
  -> CreatePipelineRun
```

### 6.4 Target-aware Tekton configuration

The `Validate` flow now uses:

```text
target.TektonNamespace
target.TektonPipelineName
```

instead of direct global values:

```text
cfg.TektonNamespace
cfg.TektonPipelineName
```

For the current dev baseline, the effective namespace and pipeline remain equivalent to the existing configuration:

```text
TektonNamespace=devops-ci-demo
TektonPipelineName=<configured validation pipeline>
```

### 6.5 Adapter bridge

A bridge named:

```text
currentTektonRuntimeClient
```

was introduced to adapt the concrete Tekton adapter to the app-layer interface:

```text
TektonRuntimeClient
```

During implementation, the app-level request was extended with the fields required by the concrete Tekton adapter:

```text
GenerateName
Image
WorkspacePVC
DockerConfigSecret
```

The concrete adapter request does not receive unsupported app-only fields such as `ApplicationName`.

### 6.6 Validation

The code path was validated with:

```text
go test ./... OK
git diff --check OK
```

---

## 7. 15.8.5.4 — Tekton provider wiring into CheckValidation

### 7.1 Commit

The CheckValidation workflow was wired with:

```text
8cf9ec5 Wire Tekton runtime provider into validation check
```

### 7.2 Files modified or added

```text
cmd/devops-control-plane/main.go
cmd/devops-control-plane/tekton_runtime_client_adapter.go
cmd/devops-control-plane/main_tekton_provider_check_validation_wiring_test.go
```

### 7.3 CheckValidation provider-aware chain

The `CheckValidation` workflow now follows:

```text
CheckValidation
  -> TechnicalRuntimeTarget
  -> RuntimeClientProviderSelection
  -> TektonRuntimeClientProviderRegistry
  -> current Tekton runtime client
  -> FindLatestPipelineRunByChange(target.TektonNamespace, change.ChangeNumber)
  -> ListTaskRunsByPipelineRun(target.TektonNamespace, status.PipelineRunName)
```

### 7.4 Target-aware runtime lookup

The runtime lookup now uses:

```text
target.TektonNamespace
```

for both:

```text
FindLatestPipelineRunByChange
ListTaskRunsByPipelineRun
```

The validation result now records:

```text
PipelineName: target.TektonPipelineName
```

instead of using direct global config for the pipeline name.

### 7.5 Validation

The code path was validated with:

```text
go test ./... OK
git diff --check OK
```

---

## 8. 15.8.5.5 — Runtime validation

### 8.1 Runtime image

Runtime validation was executed against image:

```text
8cf9ec5 Wire Tekton runtime provider into validation check
```

Expected deployed image shape:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:8cf9ec5
```

### 8.2 Readiness

Readiness passed:

```text
readyz=200
```

### 8.3 Create matrix

The create matrix produced:

```text
create_dev_http=201
create_staging_http=422
create_production_http=422
tekton_provider_change_number=CHG-2026-0030
```

Interpretation:

```text
dev remains operational
staging remains disabled
production remains disabled
```

### 8.4 ChangeRequest created for runtime validation

The runtime validation created:

```text
CHG-2026-0030
```

The `create-dev` summary was:

```text
CASE=create-dev
changeNumber=CHG-2026-0030
targetEnvironment=dev
requestedBy=tekton-provider-admin-dev
status=draft
runtimeStatus=
pipelineRunName=
namespace=
errorCode=
technicalMessage=
END_CASE=create-dev
```

### 8.5 Disabled environment validation

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

## 9. Validate runtime matrix

Endpoint tested:

```text
POST /api/v1/changes/CHG-2026-0030/validate
```

### 9.1 viewer

Observed:

```text
validate_viewer_http=403
body_prefix=forbidden: insufficient role
```

Interpretation:

```text
viewer remains denied.
```

### 9.2 operator

Observed:

```text
validate_operator_http=202
runtimeStatus=ValidationRunning
```

Response summary:

```text
CASE=validate-operator
changeNumber=CHG-2026-0030
status=draft
runtimeStatus=ValidationRunning
errorCode=
technicalMessage=
END_CASE=validate-operator
```

Interpretation:

```text
operator remains allowed and a Tekton validation PipelineRun was triggered through the provider-aware path.
```

### 9.3 approver

Observed:

```text
validate_approver_http=403
body_prefix=forbidden: insufficient role
```

Interpretation:

```text
approver remains denied for technical validation execution.
```

### 9.4 admin

Admin validate was intentionally skipped:

```text
validate_admin_http=skipped
```

Reason:

```text
operator already created a PipelineRun for CHG-2026-0030; admin validate was skipped to avoid creating a second PipelineRun on the same ChangeRequest.
```

---

## 10. CheckValidation runtime matrix

Endpoint tested:

```text
POST /api/v1/changes/CHG-2026-0030/check-validation
```

### 10.1 First operator check

Observed:

```text
check_validation_operator_http=202
runtimeStatus=ValidationRunning
```

Response summary:

```text
CASE=check-validation-operator
changeNumber=CHG-2026-0030
status=draft
runtimeStatus=ValidationRunning
errorCode=
technicalMessage=
END_CASE=check-validation-operator
```

### 10.2 Second operator check

Observed:

```text
check_validation_operator_second_http=202
runtimeStatus=ValidationFailed
```

Response summary:

```text
CASE=check-validation-operator-second
changeNumber=CHG-2026-0030
status=draft
runtimeStatus=ValidationFailed
errorCode=
technicalMessage=
END_CASE=check-validation-operator-second
```

Interpretation:

```text
CheckValidation successfully located and evaluated the Tekton validation state through the provider-aware path.
The final validation result was ValidationFailed, which is an application/workload validation outcome, not a provider wiring regression.
```

---

## 11. Final HTTP summary

The final runtime validation summary was:

```text
readyz=200
create_dev_http=201
create_staging_http=422
create_production_http=422
tekton_provider_change_number=CHG-2026-0030
validate_viewer_http=403
validate_operator_http=202
validate_approver_http=403
validate_admin_http=skipped
check_validation_operator_http=202
check_validation_operator_second_http=202
```

This matches the expected no-regression behavior.

---

## 12. Acceptance criteria

### AC-15.8.5-001 — Tekton provider abstraction exists

Status:

```text
PASSED
```

Evidence:

```text
TektonRuntimePipelineRunRequest
TektonRuntimePipelineRunRef
TektonRuntimeClient
TektonRuntimeClientProvider
TektonRuntimeClientProviderRegistry
CurrentClusterTektonRuntimeClientProvider
```

### AC-15.8.5-002 — Validate uses provider-aware path

Status:

```text
PASSED
```

Evidence:

```text
DefaultTektonRuntimeClientProviderRegistry(...).Resolve(ctx, selection)
TektonRuntimePipelineRunRequest
Namespace: target.TektonNamespace
PipelineName: target.TektonPipelineName
```

### AC-15.8.5-003 — CheckValidation uses provider-aware path

Status:

```text
PASSED
```

Evidence:

```text
FindLatestPipelineRunByChange(ctx, target.TektonNamespace, change.ChangeNumber)
ListTaskRunsByPipelineRun(ctx, target.TektonNamespace, status.PipelineRunName)
PipelineName: target.TektonPipelineName
```

### AC-15.8.5-004 — dev create remains operational

Status:

```text
PASSED
```

Evidence:

```text
create_dev_http=201
tekton_provider_change_number=CHG-2026-0030
```

### AC-15.8.5-005 — staging remains disabled

Status:

```text
PASSED
```

Evidence:

```text
create_staging_http=422
targetEnvironment "staging" is currently disabled
```

### AC-15.8.5-006 — production remains disabled

Status:

```text
PASSED
```

Evidence:

```text
create_production_http=422
targetEnvironment "production" is currently disabled
```

### AC-15.8.5-007 — Validate RBAC preserved

Status:

```text
PASSED
```

Evidence:

```text
viewer=403
operator=202
approver=403
admin=skipped intentionally
```

### AC-15.8.5-008 — CheckValidation remains operational

Status:

```text
PASSED
```

Evidence:

```text
check_validation_operator_http=202
check_validation_operator_second_http=202
runtimeStatus moved from ValidationRunning to ValidationFailed
```

---

## 13. Security assessment

The phase preserved the expected security guardrails:

```text
no Secret values read
no token values printed
no kubeconfig values printed
no staging credentials introduced
no production credentials introduced
no staging enablement
no production enablement
```

The Tekton provider currently maps only:

```text
ocp-dev -> current Tekton runtime client
```

Real staging and production Tekton clients remain future work.

---

## 14. Current state after 15.8.5

The Tekton validation path is now provider-aware for the current dev cluster.

Current supported behavior:

```text
ocp-dev:
  Tekton provider configured through current runtime client
  Validate operational
  CheckValidation operational

ocp-staging:
  environment disabled
  Tekton provider not configured for real use

ocp-production:
  environment disabled
  Tekton provider not configured for real use
```

---

## 15. Recommended next phase

Recommended next step:

```text
15.8.6 — Argo CD runtime client provider per cluster
```

Expected scope:

```text
inventory Argo CD adapter constructor and runtime usage
define Argo CD runtime client provider abstraction
wire Argo CD provider into CheckDeployment
preserve dev behavior
keep staging/production disabled
avoid reading Secret values until explicit real-client factory phase
```

Alternative follow-up before Argo CD:

```text
15.8.5.7 — Document/update full multi-provider architecture diagram
```

However, Argo CD is the natural next runtime integration because deployment checks are the remaining major provider-aware workflow.

---

## 16. Final conclusion

Phase 15.8.5 successfully introduced and validated Tekton runtime provider wiring.

The DevOps Control Plane now has provider-aware Tekton validation workflows:

```text
Validate
  -> TechnicalRuntimeTarget
  -> RuntimeClientProviderSelection
  -> TektonRuntimeClientProviderRegistry
  -> TektonRuntimeClient
  -> CreatePipelineRun

CheckValidation
  -> TechnicalRuntimeTarget
  -> RuntimeClientProviderSelection
  -> TektonRuntimeClientProviderRegistry
  -> TektonRuntimeClient
  -> FindLatestPipelineRunByChange
  -> ListTaskRunsByPipelineRun
```

Runtime validation confirmed no regression:

```text
readyz=200
dev create=201
staging=422
production=422
validate viewer=403
validate operator=202
validate approver=403
check-validation operator=202
second check-validation operator=202
```

This establishes a validated Tekton provider-aware baseline for `ocp-dev/current provider` and prepares the project for the Argo CD provider phase.
