# 15.8.4.5 — Kubernetes Runtime Provider Wiring Results

## 1. Purpose

This document records the implementation and runtime validation results for **Phase 15.8.4 — Kubernetes client provider per cluster**, with specific focus on the Kubernetes runtime provider abstraction and its integration into the `collect-evidence` workflow.

The purpose of this phase was to move the Kubernetes/OpenShift evidence collection path from a purely global, single-client model toward a provider-aware selection model based on the already introduced runtime target chain:

```text
ChangeRequest.targetEnvironment
  -> EnvironmentClusterResolver
  -> TechnicalRuntimeTarget
  -> RuntimeClientProviderSelection
  -> KubernetesRuntimeClientProviderRegistry
  -> KubernetesRuntimeEvidenceClient
```

The phase intentionally preserved the conservative runtime posture:

```text
dev        -> enabled and operational
staging    -> configured but disabled
production -> configured but disabled
```

No real staging or production Kubernetes client was created in this phase.

---

## 2. Starting point

Before this phase, the DevOps Control Plane already had the following building blocks:

```text
Environment Catalog
Cluster Registry
EnvironmentClusterResolver
TechnicalRuntimeTarget
RuntimeClientProvider
RuntimeClientProviderSelection
RuntimeClientSecretRefs
RuntimeClientSecretRefsRegistry
```

The latest relevant commits before the Kubernetes provider wiring were:

```text
1e02659 Add multi-cluster runtime secret reference model
7866133 Add runtime client secret reference loading
f5d83bf Wire runtime secret references into provider preparation
```

At this point, the application could resolve the logical runtime target and provider metadata, but Kubernetes runtime evidence collection was still using the existing globally configured Kubernetes client and namespace.

---

## 3. Phase scope

### 3.1 Completed sub-phases

```text
15.8.4.1 — Inventory Kubernetes adapter constructor and runtime usage       COMPLETED
15.8.4.2 — Define Kubernetes runtime client provider abstraction            COMPLETED
15.8.4.3 — Wire Kubernetes runtime provider into collect-evidence prep      COMPLETED
15.8.4.4 — Runtime validation for Kubernetes runtime provider wiring        COMPLETED
15.8.4.5 — Document Kubernetes runtime provider wiring results              COMPLETED
```

### 3.2 Out of scope

The following items were intentionally out of scope for this phase:

```text
creating real Kubernetes clients for staging
creating real Kubernetes clients for production
reading Kubernetes Secret values
reading service account tokens from Kubernetes Secrets
printing token values or kubeconfig values
creating Tekton multi-cluster providers
creating Argo CD multi-cluster providers
enabling staging
enabling production
changing OAuth Proxy configuration
changing OpenShift RBAC
```

This phase was focused only on the Kubernetes evidence provider abstraction and no-regression runtime validation.

---

## 4. 15.8.4.1 — Inventory results

The inventory confirmed the existing Kubernetes adapter and runtime usage.

### 4.1 Summary

```text
git_head=f5d83bf Wire runtime secret references into provider preparation
kubernetes_new_refs=1
collect_runtime_evidence_refs=2
kubernetes_adapter_new_refs=1
kubernetes_runtime_client_refs=2
kubernetes_secret_refs=65
provider_selection_refs=26
```

### 4.2 Kubernetes adapter constructor

The Kubernetes adapter constructor was identified as:

```text
internal/adapters/kubernetes/client.go:45
func New(cfg Config, opts ...Option) (*Client, error)
```

The adapter config was identified as:

```text
internal/adapters/kubernetes/client.go:21
type Config struct
```

### 4.3 Evidence collection method

The Kubernetes runtime evidence method was identified as:

```text
internal/adapters/kubernetes/client.go:80
func (c *Client) CollectRuntimeEvidence(ctx context.Context, namespace string, applicationName string) (map[string]any, error)
```

### 4.4 Previous single-client wiring

Before the provider wiring, the main runtime path still used the globally configured namespace:

```text
cmd/devops-control-plane/main.go:105
return kubernetesRuntimeClient.CollectRuntimeEvidence(ctx, cfg.KubernetesNamespace, change.ApplicationName)
```

This meant that even though a `TechnicalRuntimeTarget` existed, Kubernetes evidence collection did not yet use:

```text
TechnicalRuntimeTarget.KubernetesNamespace
```

### 4.5 ChangeService readiness

The ChangeService already exposed the application-level collector function:

```text
type KubernetesRuntimeEvidenceCollectorFunc func(ctx context.Context, change domain.ChangeRequest) (map[string]any, error)
```

The `CollectEvidence` workflow was already resolving provider selection through:

```text
resolveRuntimeClientProviderSelection(ctx, change)
```

This made the ChangeService ready for Kubernetes provider wiring.

---

## 5. 15.8.4.2 — Kubernetes runtime client provider abstraction

### 5.1 Commit

The Kubernetes provider abstraction was introduced with:

```text
3545042 Add Kubernetes runtime client provider abstraction
```

### 5.2 Files added

```text
internal/app/kubernetes_runtime_client_provider.go
internal/app/kubernetes_runtime_client_provider_test.go
```

### 5.3 Abstractions introduced

The following application-layer abstractions were added:

```text
KubernetesRuntimeEvidenceClient
KubernetesRuntimeClientProvider
KubernetesRuntimeClientProviderRegistry
CurrentClusterKubernetesRuntimeClientProvider
```

### 5.4 Conservative default behavior

The default provider registry is conservative:

```text
ocp-dev -> current Kubernetes runtime evidence client, if configured
ocp-staging -> provider not configured
ocp-production -> provider not configured
```

If no current Kubernetes runtime evidence client is configured, the default registry remains empty.

### 5.5 Safety posture

The abstraction does not read Secret values and does not create real multi-cluster clients.

It only provides a selection boundary that later phases can extend.

---

## 6. 15.8.4.3 — Kubernetes provider wiring into collect-evidence

### 6.1 Commit

The collect-evidence wiring was introduced with:

```text
29f3d6d Wire Kubernetes runtime provider into evidence collection
```

### 6.2 Files modified or added

```text
cmd/devops-control-plane/main.go
cmd/devops-control-plane/main_kubernetes_provider_wiring_test.go
internal/app/change_service.go
internal/app/change_service_kubernetes_provider_wiring_test.go
```

### 6.3 Main wiring

The composition root now wires the Kubernetes runtime client provider registry using the current Kubernetes runtime evidence client:

```text
app.WithKubernetesRuntimeClientProviderRegistry(
  app.DefaultKubernetesRuntimeClientProviderRegistry(kubernetesRuntimeClient),
)
```

This means:

```text
ocp-dev -> current Kubernetes runtime evidence client
```

### 6.4 ChangeService wiring

The ChangeService now exposes:

```text
WithKubernetesRuntimeClientProviderRegistry
```

and stores:

```text
kubernetesRuntimeClientProviderRegistry KubernetesRuntimeClientProviderRegistry
```

A helper was introduced:

```text
collectKubernetesRuntimeEvidence(ctx, change, providerSelection)
```

### 6.5 New collect-evidence chain

The new collection chain is:

```text
CollectEvidence
  -> resolveRuntimeClientProviderSelection(ctx, change)
  -> collectKubernetesRuntimeEvidence(ctx, change, providerSelection)
  -> KubernetesRuntimeClientProviderRegistry.Resolve(ctx, providerSelection)
  -> KubernetesRuntimeEvidenceClient
  -> CollectRuntimeEvidence(ctx, providerSelection.Target.KubernetesNamespace, change.ApplicationName)
```

### 6.6 Important behavior change

The Kubernetes namespace now comes from:

```text
providerSelection.Target.KubernetesNamespace
```

instead of the global configuration field:

```text
cfg.KubernetesNamespace
```

For the current dev baseline the value remains equivalent:

```text
devops-ci-demo
```

This is the required intermediate step before real multi-cluster Kubernetes clients are enabled.

---

## 7. 15.8.4.4 — Runtime validation

### 7.1 Runtime image

Runtime validation was executed against image:

```text
29f3d6d Wire Kubernetes runtime provider into evidence collection
```

Expected container image shape:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:29f3d6d
```

### 7.2 Runtime readiness

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
k8s_provider_change_number=CHG-2026-0029
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
CHG-2026-0029
```

The create-dev response summary was:

```text
CASE=create-dev
changeNumber=CHG-2026-0029
targetEnvironment=dev
requestedBy=k8s-provider-admin-dev
status=draft
runtimeStatus=
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

## 8. Collect-evidence runtime validation

The technical action tested was:

```text
POST /api/v1/changes/CHG-2026-0029/collect-evidence
```

This directly validates the new provider-aware Kubernetes evidence collection path.

### 8.1 viewer

Observed:

```text
collect_evidence_viewer_http=403
body_prefix=forbidden: insufficient role
```

Interpretation:

```text
viewer remains denied.
```

### 8.2 operator

Observed:

```text
collect_evidence_operator_http=202
runtimeStatus=EvidenceCollected
```

Response summary:

```text
CASE=collect-operator
changeNumber=CHG-2026-0029
status=draft
runtimeStatus=EvidenceCollected
errorCode=
technicalMessage=
END_CASE=collect-operator
```

Interpretation:

```text
operator remains allowed and Kubernetes runtime evidence collection succeeds.
```

### 8.3 approver

Observed:

```text
collect_evidence_approver_http=403
body_prefix=forbidden: insufficient role
```

Interpretation:

```text
approver remains denied for technical action execution.
```

### 8.4 admin

Observed:

```text
collect_evidence_admin_http=202
runtimeStatus=EvidenceCollected
```

Response summary:

```text
CASE=collect-admin
changeNumber=CHG-2026-0029
status=draft
runtimeStatus=EvidenceCollected
errorCode=
technicalMessage=
END_CASE=collect-admin
```

Interpretation:

```text
admin remains allowed and Kubernetes runtime evidence collection succeeds.
```

---

## 9. Final HTTP summary

The final HTTP validation summary was:

```text
readyz=200
create_dev_http=201
create_staging_http=422
create_production_http=422
k8s_provider_change_number=CHG-2026-0029
collect_evidence_viewer_http=403
collect_evidence_operator_http=202
collect_evidence_approver_http=403
collect_evidence_admin_http=202
```

This matches the expected no-regression behavior.

---

## 10. Acceptance criteria

### AC-15.8.4-001 — Kubernetes provider abstraction exists

Status:

```text
PASSED
```

Evidence:

```text
KubernetesRuntimeEvidenceClient
KubernetesRuntimeClientProvider
KubernetesRuntimeClientProviderRegistry
CurrentClusterKubernetesRuntimeClientProvider
```

### AC-15.8.4-002 — current-cluster provider baseline exists

Status:

```text
PASSED
```

Evidence:

```text
ocp-dev -> current Kubernetes runtime evidence client
```

### AC-15.8.4-003 — collect-evidence uses Kubernetes provider selection

Status:

```text
PASSED
```

Evidence:

```text
collectKubernetesRuntimeEvidence(ctx, change, providerSelection)
KubernetesRuntimeClientProviderRegistry.Resolve(ctx, selection)
selection.Target.KubernetesNamespace
```

### AC-15.8.4-004 — dev remains operational

Status:

```text
PASSED
```

Evidence:

```text
create_dev_http=201
k8s_provider_change_number=CHG-2026-0029
```

### AC-15.8.4-005 — staging remains disabled

Status:

```text
PASSED
```

Evidence:

```text
create_staging_http=422
targetEnvironment "staging" is currently disabled
```

### AC-15.8.4-006 — production remains disabled

Status:

```text
PASSED
```

Evidence:

```text
create_production_http=422
targetEnvironment "production" is currently disabled
```

### AC-15.8.4-007 — technical RBAC preserved

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

### AC-15.8.4-008 — runtime evidence still collected for operator/admin

Status:

```text
PASSED
```

Evidence:

```text
operator runtimeStatus=EvidenceCollected
admin runtimeStatus=EvidenceCollected
```

---

## 11. Security assessment

The phase preserved the security model:

```text
no Secret values read
no token values printed
no kubeconfig values printed
no staging credentials introduced
no production credentials introduced
no staging enablement
no production enablement
```

The application still uses the existing current-cluster Kubernetes runtime evidence client for dev.

The Secret reference model remains metadata-only at this point.

---

## 12. Current state after 15.8.4

The Kubernetes runtime evidence path is now provider-aware for the current dev cluster.

Current supported runtime behavior:

```text
ocp-dev:
  Kubernetes provider configured through current runtime client
  collect-evidence operational

ocp-staging:
  environment disabled
  Kubernetes provider not configured for real use

ocp-production:
  environment disabled
  Kubernetes provider not configured for real use
```

The project is now ready to repeat the same provider pattern for other runtime integrations.

---

## 13. Recommended next phase

Recommended next step:

```text
15.8.5 — Tekton runtime client provider per cluster
```

Expected scope:

```text
inventory Tekton adapter constructor and runtime usage
define Tekton runtime client provider abstraction
wire Tekton provider into validate/check-validation preparation
preserve dev behavior
keep staging/production disabled
avoid reading Secret values until explicit client factory phase
```

Alternative path, if prioritizing deployment status over validation pipeline:

```text
15.8.6 — Argo CD runtime client provider per cluster
```

However, Tekton is the natural next step because the validation workflow is already strongly tied to environment-specific namespaces and pipeline names.

---

## 14. Final conclusion

Phase 15.8.4 successfully introduced and validated Kubernetes runtime provider wiring.

The DevOps Control Plane now has a provider-aware Kubernetes evidence path:

```text
ChangeRequest.targetEnvironment
  -> TechnicalRuntimeTarget
  -> RuntimeClientProviderSelection
  -> KubernetesRuntimeClientProviderRegistry
  -> KubernetesRuntimeEvidenceClient
  -> runtime evidence collection
```

The runtime validation confirmed no regression:

```text
readyz=200
dev create=201
staging=422
production=422
viewer collect-evidence=403
operator collect-evidence=202
approver collect-evidence=403
admin collect-evidence=202
```

This establishes a safe and validated baseline for real multi-cluster Kubernetes client expansion in later phases.
