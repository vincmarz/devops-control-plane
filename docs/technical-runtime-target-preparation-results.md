# 15.7.5 — TechnicalRuntimeTarget Preparation Results

## 1. Purpose

This document records the results of **Phase 15.7 — Multi-cluster runtime client selection preparation**, with specific focus on the introduction, wiring and runtime validation of the `TechnicalRuntimeTarget` model.

The goal of this phase was to prepare DevOps Control Plane technical workflows for future multi-cluster runtime client selection while preserving the current safe runtime behavior:

```text
dev        -> enabled and operational
staging    -> configured but disabled
production -> configured but disabled
```

This phase does **not** introduce real multi-cluster Kubernetes, Tekton or Argo CD clients yet. Instead, this phase establishes the application-layer target resolution model that later phases can use to select appropriate clients.

---

## 2. Starting point

At the beginning of Phase 15.7, the project had already completed the following multi-environment and cluster-readiness phases:

```text
15.3 — Cluster Registry baseline
15.4 — Environment-to-cluster resolver baseline
15.5 — Multi-user / multi-environment validation matrix
15.6 — Role-aware UI action visibility refinement
```

The latest documented baseline before Phase 15.7 was:

```text
babaf76 Document role-aware UI action visibility results
```

The runtime image previously validated before the new technical target preparation was:

```text
7d79087 Add role-aware UI action visibility
```

The repository was clean and aligned with `main` / `origin/main` before the Phase 15.7 implementation work started.

---

## 3. Phase 15.7 scope

### 3.1 In scope

Phase 15.7 covered the following preparation activities:

```text
15.7.1 — Inventory current runtime client instantiation
15.7.2 — Define TechnicalRuntimeTarget model
15.7.3 — Wire TechnicalRuntimeTarget into technical workflow preparation
15.7.4 — Runtime validation for TechnicalRuntimeTarget preparation
15.7.5 — Document TechnicalRuntimeTarget preparation results
```

### 3.2 Out of scope

The following items were intentionally kept out of the current phase:

```text
creating real multi-cluster Kubernetes clients
creating real multi-cluster Tekton clients
creating real multi-cluster Argo CD clients
introducing staging or production tokens
enabling staging
enabling production
executing technical workflows on staging or production
changing OpenShift RBAC
changing OAuth Proxy configuration
changing current dev runtime execution behavior
```

This was intentional to keep Phase 15.7 small, testable and safe.

---

## 4. 15.7.1 — Runtime client instantiation inventory

### 4.1 Inventory summary

The inventory produced the following summary:

```text
git_head=babaf76 Document role-aware UI action visibility results
kubernetes_client_new_refs=1
tekton_client_new_refs=1
argocd_client_new_refs=1
cfg_kubernetes_namespace_refs=1
cfg_tekton_namespace_refs=4
environment_cluster_resolver_refs=26
resolve_technical_action_target_refs=3
```

### 4.2 Runtime image at inventory time

The runtime deployment was using:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:7d79087
```

Runtime mounts were present:

```text
mount=dcp-trust-bundle path=/etc/dcp-trust
mount=dcp-environments path=/etc/dcp-environments
mount=dcp-clusters path=/etc/dcp-clusters
```

### 4.3 Findings from inventory

The inventory confirmed that the technical runtime clients were still instantiated as single runtime clients from global configuration.

Kubernetes client creation:

```text
cmd/devops-control-plane/main.go:98
kubernetesadapter.New(...)
```

Tekton client creation:

```text
cmd/devops-control-plane/main.go:107
tektonadapter.New(...)
```

Argo CD client creation:

```text
cmd/devops-control-plane/main.go:159
argocdadapter.New(...)
```

Kubernetes runtime evidence was still using the global namespace:

```text
cfg.KubernetesNamespace
```

Tekton validation and status checks were still using global Tekton configuration:

```text
cfg.TektonNamespace
cfg.TektonPipelineName
```

Argo CD was still using global configuration:

```text
cfg.ArgoCDBaseURL
cfg.ArgoCDAuthToken
```

### 4.4 Architecture assessment

The inventory confirmed a positive architectural aspect: `ChangeService` is already adapter-independent and uses injected application ports/functions such as:

```text
TektonRunPipelineFunc
TektonCheckValidationFunc
ArgoCDCheckDeploymentFunc
KubernetesRuntimeEvidenceCollectorFunc
GitCreateBranchFunc
GitCommitFilesFunc
GitOpenMergeRequestFunc
GitMergeRequestFunc
```

This means the future multi-cluster client selection can be introduced at the composition root and application-service boundary instead of forcing concrete adapters directly into the domain/application layer.

### 4.5 Resolver availability

The existing resolver was already present:

```text
EnvironmentClusterResolver
ResolveEnabledTarget
ResolveTechnicalActionTarget
DefaultClusterRegistry
ClusterDefinition
```

However, before Phase 15.7.3, technical runtime workflows were not yet resolving a runtime target before execution.

---

## 5. 15.7.2 — TechnicalRuntimeTarget model

### 5.1 Commit

The model was introduced with commit:

```text
f997b89 Add technical runtime target model
```

### 5.2 Files added

```text
internal/app/technical_runtime_target.go
internal/app/technical_runtime_target_test.go
```

Commit delta:

```text
2 files changed, 269 insertions(+)
```

### 5.3 Model purpose

The `TechnicalRuntimeTarget` model describes the target technical execution context resolved from a ChangeRequest target environment.

The model prepares the application for this future chain:

```text
ChangeRequest.targetEnvironment
  -> EnvironmentClusterResolver
  -> Environment Catalog
  -> Cluster Registry
  -> TechnicalRuntimeTarget
  -> future runtime client provider
```

### 5.4 Model fields

The model includes the following key fields:

```text
TargetEnvironment
EnvironmentName
EnvironmentDisplayName
ClusterName
ClusterDisplayName
ClusterEnabled
KubernetesNamespace
TektonNamespace
TektonPipelineName
ArgoCDApplicationName
GitTargetBranch
```

These fields represent the minimum technical metadata required by the current Kubernetes, Tekton, Argo CD and GitOps-oriented workflows.

### 5.5 Resolver

The following resolver was introduced:

```text
TechnicalRuntimeTargetResolver
```

The resolver uses:

```text
EnvironmentClusterResolver
```

and resolves enabled technical targets via:

```text
ResolveTechnicalActionTarget
```

### 5.6 Test coverage

The tests cover:

```text
dev resolves successfully
missing targetEnvironment falls back to default dev
staging is rejected because disabled
production is rejected because disabled
unknown-env is rejected as not configured
missing technical metadata is rejected
missing Tekton pipeline name is rejected
```

During implementation, one test was corrected to include:

```text
AllowTechnicalActions: true
```

This was necessary because the technical target resolver correctly rejects environments that do not allow technical actions before validating downstream runtime metadata.

### 5.7 Local validation

The model was validated with:

```text
go test ./... OK
git diff --check OK
```

---

## 6. 15.7.3 — Wiring TechnicalRuntimeTarget into technical workflow preparation

### 6.1 Commit

The wiring was introduced with commit:

```text
6a4f313 Wire technical runtime target preparation
```

### 6.2 Files modified or added

```text
cmd/devops-control-plane/main.go
internal/app/change_service.go
internal/app/technical_runtime_target.go
internal/app/change_service_runtime_target_test.go
```

### 6.3 Composition root wiring

The composition root now wires the default technical runtime target resolver into `ChangeService`:

```text
app.WithTechnicalRuntimeTargetResolver(app.DefaultTechnicalRuntimeTargetResolver(cfg.TektonPipelineName))
```

This keeps resolver construction close to runtime configuration, without introducing concrete runtime client selection yet.

### 6.4 ChangeService wiring

`ChangeService` now contains:

```text
technicalRuntimeTargetResolver TechnicalRuntimeTargetResolverFunc
```

The following options are available:

```text
WithTechnicalRuntimeTargetResolver
WithTechnicalRuntimeTargetResolverFunc
```

The internal helper is:

```text
resolveTechnicalRuntimeTarget(ctx, change)
```

### 6.5 Technical workflows covered

The technical runtime target is now resolved before the following workflows:

```text
Validate
CheckValidation
CheckDeployment
CollectEvidence
```

Relevant checks confirmed resolver calls in these methods:

```text
resolveTechnicalRuntimeTarget(ctx, change)
```

The observed line references after implementation were:

```text
Validate         -> resolveTechnicalRuntimeTarget
CheckValidation -> resolveTechnicalRuntimeTarget
CheckDeployment -> resolveTechnicalRuntimeTarget
CollectEvidence -> resolveTechnicalRuntimeTarget
```

### 6.6 ResolveChange helper

`TechnicalRuntimeTargetResolver` also exposes:

```text
ResolveChange(ctx, change)
```

This helper resolves the runtime target directly from a `domain.ChangeRequest`.

### 6.7 Source-level regression tests

A source-level test was introduced:

```text
TestChangeServiceTechnicalWorkflowsResolveRuntimeTarget
```

This test asserts that the identified technical workflows resolve the technical runtime target before execution.

A second test verifies the availability of the resolver options and field wiring.

### 6.8 Recovery during implementation

During patching, two implementation recovery actions were required:

1. A malformed comment insertion in `technical_runtime_target.go` was repaired.
2. The missing `technicalRuntimeTargetResolver` field was added to the `ChangeService` struct.

After these corrections:

```text
go test ./... OK
git diff --check OK
```

---

## 7. 15.7.4 — Runtime validation

### 7.1 Runtime image

The new runtime image was built, pushed and deployed:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:6a4f313
```

Build and push completed successfully.

Deployment rollout completed successfully:

```text
deployment "devops-control-plane" successfully rolled out
```

### 7.2 Runtime mounts

Runtime mounts remained correct:

```text
container=devops-control-plane
image=image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:6a4f313
mount=dcp-trust-bundle path=/etc/dcp-trust
mount=dcp-environments path=/etc/dcp-environments
mount=dcp-clusters path=/etc/dcp-clusters
```

### 7.3 Evidence directory

Evidence was collected under a directory matching:

```text
/tmp/dcp-15.7.4-runtime-target-validation-YYYYMMDD-HHMMSS
```

### 7.4 Repository state during validation

The repository head during validation was:

```text
6a4f313 Wire technical runtime target preparation
```

The latest commits were:

```text
6a4f313 Wire technical runtime target preparation
f997b89 Add technical runtime target model
babaf76 Document role-aware UI action visibility results
7d79087 Add role-aware UI action visibility
8a1b70d Document multi-user multi-environment validation results
```

### 7.5 Readiness

Readiness validation succeeded:

```text
readyz=200
```

---

## 8. Runtime create matrix results

### 8.1 dev

Request:

```text
POST /api/v1/changes
 targetEnvironment=dev
```

Result:

```text
create_dev_http=201
```

Created ChangeRequest:

```text
CHG-2026-0027
```

Response summary:

```text
CASE=create-dev
changeNumber=CHG-2026-0027
targetEnvironment=dev
requestedBy=runtime-target-admin-dev
status=draft
runtimeStatus=
errorCode=
technicalMessage=
END_CASE=create-dev
```

Interpretation:

```text
dev remains enabled and operational.
```

### 8.2 staging

Request:

```text
POST /api/v1/changes
 targetEnvironment=staging
```

Result:

```text
create_staging_http=422
```

Response summary:

```text
CASE=create-staging
errorCode=VALIDATION_INVALID_REQUEST
technicalMessage=targetEnvironment "staging" is currently disabled
END_CASE=create-staging
```

Interpretation:

```text
staging remains configured but disabled.
```

### 8.3 production

Request:

```text
POST /api/v1/changes
 targetEnvironment=production
```

Result:

```text
create_production_http=422
```

Response summary:

```text
CASE=create-production
errorCode=VALIDATION_INVALID_REQUEST
technicalMessage=targetEnvironment "production" is currently disabled
END_CASE=create-production
```

Interpretation:

```text
production remains configured but disabled.
```

---

## 9. Runtime technical action matrix results

The technical action validation used:

```text
POST /api/v1/changes/CHG-2026-0027/collect-evidence
```

This path is important because `CollectEvidence` is one of the workflows now resolving the `TechnicalRuntimeTarget` before execution.

### 9.1 viewer

Result:

```text
collect_evidence_viewer_http=403
```

Response body was plain text:

```text
forbidden: insufficient role
```

Interpretation:

```text
viewer remains denied.
```

### 9.2 operator

Result:

```text
collect_evidence_operator_http=202
```

Response summary:

```text
CASE=collect-operator
changeNumber=CHG-2026-0027
status=draft
runtimeStatus=EvidenceCollected
errorCode=
technicalMessage=
END_CASE=collect-operator
```

Interpretation:

```text
operator remains allowed.
```

### 9.3 approver

Result:

```text
collect_evidence_approver_http=403
```

Response body was plain text:

```text
forbidden: insufficient role
```

Interpretation:

```text
approver remains denied for technical action execution.
```

### 9.4 admin

Result:

```text
collect_evidence_admin_http=202
```

Response summary:

```text
CASE=collect-admin
changeNumber=CHG-2026-0027
status=draft
runtimeStatus=EvidenceCollected
errorCode=
technicalMessage=
END_CASE=collect-admin
```

Interpretation:

```text
admin remains allowed.
```

---

## 10. Final HTTP summary

The runtime HTTP summary was:

```text
readyz=200
create_dev_http=201
create_staging_http=422
create_production_http=422
runtime_target_change_number=CHG-2026-0027
collect_evidence_viewer_http=403
collect_evidence_operator_http=202
collect_evidence_approver_http=403
collect_evidence_admin_http=202
```

This matches the expected behavior.

---

## 11. Acceptance criteria

### AC-15.7.4-001 — Runtime image deployed

Status:

```text
PASSED
```

Evidence:

```text
image=.../devops-control-plane:6a4f313
```

### AC-15.7.4-002 — Runtime mounts preserved

Status:

```text
PASSED
```

Evidence:

```text
mount=dcp-environments path=/etc/dcp-environments
mount=dcp-clusters path=/etc/dcp-clusters
```

### AC-15.7.4-003 — Readiness successful

Status:

```text
PASSED
```

Evidence:

```text
readyz=200
```

### AC-15.7.4-004 — dev remains operational

Status:

```text
PASSED
```

Evidence:

```text
create_dev_http=201
runtime_target_change_number=CHG-2026-0027
```

### AC-15.7.4-005 — staging remains disabled

Status:

```text
PASSED
```

Evidence:

```text
create_staging_http=422
targetEnvironment "staging" is currently disabled
```

### AC-15.7.4-006 — production remains disabled

Status:

```text
PASSED
```

Evidence:

```text
create_production_http=422
targetEnvironment "production" is currently disabled
```

### AC-15.7.4-007 — collect-evidence policy preserved

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

### AC-15.7.4-008 — TechnicalRuntimeTarget preparation does not break existing runtime execution

Status:

```text
PASSED
```

Evidence:

```text
operator/admin collect-evidence returned 202
runtimeStatus=EvidenceCollected
```

---

## 12. Current state after Phase 15.7.4

The current repository state is:

```text
HEAD -> main
origin/main
6a4f313 Wire technical runtime target preparation
```

The runtime image validated is:

```text
6a4f313
```

The working tree after runtime validation was clean:

```text
git status --short
```

returned no output.

---

## 13. Phase 15.7 status

Current status:

```text
15.7 — Multi-cluster runtime client selection preparation
IN PROGRESS
```

Completed sub-phases:

```text
15.7.1 — Inventory current runtime client instantiation              COMPLETED
15.7.2 — Define TechnicalRuntimeTarget model                         COMPLETED
15.7.3 — Wire TechnicalRuntimeTarget into technical workflow prep     COMPLETED
15.7.4 — Runtime validation for TechnicalRuntimeTarget preparation    COMPLETED
15.7.5 — Document TechnicalRuntimeTarget preparation results          COMPLETED
```

---

## 14. Remaining limitation

The project is now ready for client-selection preparation, but actual runtime clients are still single-instance.

Current state:

```text
Kubernetes client: single runtime client from global config
Tekton client: single runtime client from global config
Argo CD client: single runtime client from global config
```

The new layer resolves runtime target metadata, but does not yet select a different client per cluster.

This is expected for Phase 15.7.

---

## 15. Recommended next phase

Recommended next step:

```text
15.7.6 — Prepare Kubernetes/Tekton/Argo CD client provider abstractions
```

Alternative naming:

```text
15.8 — Runtime client provider abstraction
```

Recommended scope:

```text
introduce provider interfaces for technical clients
keep dev/current cluster as the only available provider
map TechnicalRuntimeTarget.ClusterName to the current provider
return explicit errors for disabled or unknown runtime providers
avoid introducing real staging/production credentials
avoid enabling staging/production
```

Suggested acceptance criteria:

```text
dev target resolves to current runtime provider
staging remains disabled
production remains disabled
unknown target remains rejected
no runtime behavior regression
no new secret exposure
```

---

## 16. Final conclusion

Phase 15.7 successfully introduced the first runtime preparation layer required for future multi-cluster operation.

The project now has:

```text
Environment Catalog
Cluster Registry
EnvironmentClusterResolver
TechnicalRuntimeTarget
TechnicalRuntimeTargetResolver
technical workflow preparation wiring
runtime validation on image 6a4f313
```

The runtime behavior remains stable and conservative:

```text
dev operational
staging disabled
production disabled
technical RBAC preserved
existing runtime clients still used
```

This establishes a safe foundation for the next phase: client provider abstraction and, eventually, true multi-cluster runtime client selection.
