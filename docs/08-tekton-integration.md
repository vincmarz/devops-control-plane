# DevOps Control Plane — Tekton Integration

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 08 — Tekton Integration
- **Version:** 0.2
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Previous documents:**
  - `docs/00-vision.md`
  - `docs/01-scope-mvp.md`
  - `docs/02-personas-use-cases.md`
  - `docs/03-functional-requirements.md`
  - `docs/04-non-functional-requirements.md`
  - `docs/05-architecture.md`
  - `docs/06-argocd-integration.md`
  - `docs/07-gitlab-integration.md`
- **Status:** Rewritten in English and refreshed while preserving the original validation-engine intent
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document describes how the DevOps Control Plane integrates with Tekton / Red Hat OpenShift Pipelines.

The original document defined Tekton as the validation and technical automation engine for GitOps changes. This refreshed version preserves that original architectural intent while aligning the document with the current implementation baseline.

The document defines:

- why Tekton is used in the project;
- how the Control Plane interacts with Tekton through Kubernetes API;
- the role of `PipelineRun` and `TaskRun` resources;
- the current GitOps validation pipeline;
- validation parameters and execution model;
- validation runtime status mapping;
- TaskRun diagnostics and failure analysis;
- validation evidence persisted in PostgreSQL;
- anti-secret and manifest guardrails;
- RBAC and security requirements;
- current `/api/v1` validation endpoints;
- future environment-aware Tekton mapping for `dev`, `staging` and `production`.

Tekton does not replace GitLab or Argo CD. In the DevOps Control Plane architecture, Tekton validates that a GitOps change is technically acceptable before the change is considered safe to move forward in the workflow.

---

## 2. Role of Tekton in the architecture

Tekton has the role of **Validation Engine**.

Tekton is responsible for:

- executing validation pipelines;
- cloning the GitOps repository and selected branch or revision;
- validating YAML and rendered manifests;
- executing Kustomize validation where applicable;
- running dry-run checks;
- enforcing anti-secret and GitOps manifest guardrails;
- producing `PipelineRun` and `TaskRun` status;
- producing technical logs and diagnostics associated with validation.

The DevOps Control Plane is responsible for:

- creating `PipelineRun` resources through Kubernetes API;
- passing ChangeRequest, Git and path parameters to Tekton;
- checking `PipelineRun` status;
- collecting associated `TaskRun` diagnostics;
- storing validation evidence in PostgreSQL;
- updating ChangeRequest runtime status;
- exposing validation status and diagnostics through API and UI;
- preserving a clear audit trail.

---

## 3. Integration principles

### 3.1 Tekton validates GitOps change proposals

Tekton validates the GitOps branch or revision produced through GitLab workflow.

The intended flow remains:

```text
GitLab branch
  -> Tekton PipelineRun
  -> TaskRun validation
  -> validation evidence
  -> Argo CD deployment state only after validation is acceptable
```

This preserves the original project spirit:

```text
Validation must be repeatable, observable and auditable.
```

### 3.2 The Control Plane uses Kubernetes API directly

The architectural integration path is:

```text
DevOps Control Plane -> Kubernetes API -> Tekton CRDs
```

The backend Go application must not depend on the `tkn` CLI or shell wrappers for product runtime behavior.

The `tkn` CLI remains useful for troubleshooting, manual validation and lab diagnostics, but it is not the Control Plane runtime integration mechanism.

### 3.3 `PipelineRun` and `TaskRun` are first-class evidence sources

A `PipelineRun` represents a validation execution. `TaskRun` resources represent the detailed validation steps.

The Control Plane must preserve enough information to answer:

- which validation was started;
- which Git revision was validated;
- which tasks ran;
- which task failed, if any;
- which reason and message Tekton reported;
- which evidence was stored for audit.

### 3.4 Tekton does not decide what is deployed

Tekton validates the proposed GitOps change.

GitLab remains the source of declarative change. Argo CD remains the GitOps reconciliation engine. OpenShift remains the runtime platform.

---

## 4. Current implementation baseline

Current Tekton-related capabilities include:

- Tekton integration through Kubernetes API;
- `POST /api/v1/changes/{id}/validate` to start validation;
- `POST /api/v1/changes/{id}/check-validation` to read validation result;
- `PipelineRun` creation for ChangeRequests;
- validation runtime statuses:
  - `ValidationRunning`;
  - `ValidationSucceeded`;
  - `ValidationFailed`;
- `TaskRun` diagnostics collection;
- validation evidence persisted in PostgreSQL;
- failure diagnostics exposed through API and UI;
- hardened GitOps validation pipeline;
- anti-secret and manifest guardrails;
- pipeline source of truth in `pipelines/validate-gitops.yaml`;
- runtime RBAC least-privilege baseline for Tekton resources;
- relationship with GitLab branch workflow and Argo CD deployment checks.

---

## 5. Relevant Tekton concepts

## 5.1 Task

A `Task` defines a reusable unit of work.

Examples in the validation workflow:

- clone repository;
- validate GitOps manifests;
- run Kustomize rendering;
- run dry-run checks;
- apply anti-secret checks;
- produce validation output.

## 5.2 TaskRun

A `TaskRun` is the execution of a `Task`.

The Control Plane uses `TaskRun` information to identify:

- executed validation steps;
- task status;
- failure reason;
- failure message;
- start and completion times;
- diagnostics useful for operators and reviewers.

## 5.3 Pipeline

A `Pipeline` defines the validation workflow as a sequence or graph of tasks.

The current source-of-truth pipeline is:

```text
pipelines/validate-gitops.yaml
```

The logical pipeline name in the current runtime baseline is:

```text
validate-gitops
```

## 5.4 PipelineRun

A `PipelineRun` instantiates and executes the validation `Pipeline`.

Each ChangeRequest validation attempt creates or refers to a dedicated `PipelineRun` associated with the ChangeRequest context.

Example generated name pattern:

```text
devops-cp-validate-chg-2026-0006-xxxxx
```

---

## 6. Validation pipeline baseline

## 6.1 Purpose

The validation pipeline validates a GitOps branch before deployment checks and evidence collection continue.

The pipeline should prove that:

- the Git branch or revision exists;
- the expected GitOps path exists;
- manifests can be rendered or validated;
- unsafe Secret-like content is not introduced;
- blocked Kubernetes Secret manifests are detected;
- the change can be reviewed through repeatable technical evidence.

## 6.2 Current validation tasks

The current pipeline baseline includes the logical validation path:

```text
clone-repository
validate-gitops-manifests
```

The validation task performs policy checks such as:

- anti-secret and manifest guardrails;
- Kubernetes `Secret` manifest detection;
- inline secret-like value detection;
- dry-run and manifest checks where configured;
- readable policy violation messages.

## 6.3 Original candidate task model

The original design described a broader task model that remains useful as future direction:

```text
clone-repository
show-context
yaml-validate
kustomize-build
server-side-dry-run
anti-secret-check
appproject-policy-check
report
```

This intent remains valid. The current implementation can evolve toward more explicit tasks over time while preserving the same validation purpose.

---

## 7. PipelineRun creation

## 7.1 Validation trigger

Validation is triggered by:

```text
POST /api/v1/changes/{id}/validate
```

Main flow:

```text
ChangeRequest
  -> ChangeService.Validate
  -> Tekton adapter creates PipelineRun
  -> runtime status becomes ValidationRunning
  -> technical event is recorded
```

## 7.2 Required parameters

The Control Plane passes validation context such as:

```yaml
changeNumber: CHG-2026-0006
applicationName: demo-go-color-app
gitUrl: https://github.com/vincmarz/demo-app-gitops.git
gitRevision: change/CHG-2026-0006
validationPath: apps/demo-go-color-app
pipelineName: validate-gitops
namespace: devops-ci-demo
```

The exact values are resolved from the ChangeRequest, application configuration and future environment catalog.

## 7.3 PipelineRun metadata

A `PipelineRun` should include labels or generated names that make it traceable to the ChangeRequest.

Recommended metadata:

```text
changeNumber
applicationName
targetEnvironment
managedBy=devops-control-plane
```

---

## 8. Tekton adapter

## 8.1 Responsibilities

The Tekton adapter encapsulates Tekton resource handling through Kubernetes API.

The rest of the application must not depend on raw CRD payloads, `tkn` command output or shell wrappers.

Current adapter responsibilities include:

- create `PipelineRun`;
- read `PipelineRun` status;
- list related `TaskRun` resources;
- map Tekton conditions to Control Plane runtime status;
- return diagnostics to the ChangeService.

## 8.2 Conceptual interface

```go
type TektonAdapter interface {
    CreatePipelineRun(ctx context.Context, change domain.ChangeRequest) (name string, namespace string, uid string, err error)
    CheckValidation(ctx context.Context, change domain.ChangeRequest) (TektonValidationResult, error)
    ListTaskRunsByPipelineRun(ctx context.Context, namespace string, pipelineRunName string) ([]TektonTaskRunResult, error)
}
```

The exact implementation can evolve, but the architectural boundary remains:

```text
application services depend on Tekton ports, not raw Kubernetes client details
```

## 8.3 Package location

Current package area:

```text
internal/adapters/tekton/
```

Expected responsibilities:

```text
client.go        -> Kubernetes client and Tekton request handling
models.go        -> normalized Tekton models
pipelineruns.go  -> PipelineRun creation and retrieval
taskruns.go      -> TaskRun listing and mapping
errors.go        -> Tekton/Kubernetes error normalization
```

---

## 9. Normalized data model

## 9.1 PipelineRun reference

```yaml
name: devops-cp-validate-chg-2026-0006-xxxxx
namespace: devops-ci-demo
uid: <uid>
changeNumber: CHG-2026-0006
applicationName: demo-go-color-app
```

## 9.2 PipelineRun status

```yaml
name: devops-cp-validate-chg-2026-0006-xxxxx
namespace: devops-ci-demo
status: Succeeded
reason: Succeeded
message: "Tasks Completed: 2 completed, 0 failed"
pipelineName: validate-gitops
gitURL: https://github.com/vincmarz/demo-app-gitops.git
gitRevision: change/CHG-2026-0006
validationPath: apps/demo-go-color-app
```

## 9.3 TaskRun diagnostics

```yaml
name: devops-cp-validate-chg-2026-0006-xxxxx-validate-gitops-manifests
namespace: devops-ci-demo
pipelineTaskName: validate-gitops-manifests
taskName: validate-gitops-manifests
status: Succeeded
reason: Succeeded
message: policy checks passed
```

Failure example:

```yaml
name: devops-cp-validate-chg-2026-0001-xxxxx-clone-repository
namespace: devops-ci-demo
pipelineTaskName: clone-repository
status: Failed
reason: Failed
message: step-prepare-and-run exited with code 1
```

---

## 10. Runtime status mapping

Tekton conditions map to Control Plane validation runtime statuses.

| Tekton condition | Control Plane runtime status |
|---|---|
| `Succeeded=True` | `ValidationSucceeded` |
| `Succeeded=False` | `ValidationFailed` |
| `Succeeded=Unknown` | `ValidationRunning` |
| timeout or missing terminal state | `ValidationFailed` or timeout-specific error, depending on implementation |

Rules:

- successful validation requires terminal Tekton success;
- failed Tekton validation must not be hidden;
- TaskRun diagnostics must identify failed tasks when possible;
- lifecycle status must not be overwritten by runtime status.

---

## 11. Validation evidence

## 11.1 Evidence creation

Validation evidence is created when Tekton validation reaches a terminal state.

Evidence type:

```text
validation
```

Evidence includes:

- ChangeRequest metadata;
- Tekton namespace;
- PipelineRun name;
- PipelineRun UID;
- pipeline name;
- Git URL;
- Git revision;
- validation path;
- Tekton status;
- reason;
- message;
- TaskRun diagnostics;
- sanitized payload.

## 11.2 Successful validation evidence

```yaml
provider: tekton
namespace: devops-ci-demo
pipelineRunName: devops-cp-validate-chg-2026-0006-xxxxx
status: Succeeded
reason: Succeeded
gitRevision: change/CHG-2026-0006
validationPath: apps/demo-go-color-app
taskRuns:
  - name: clone-repository
    status: Succeeded
  - name: validate-gitops-manifests
    status: Succeeded
```

## 11.3 Failed validation evidence

```yaml
provider: tekton
pipelineRunName: devops-cp-validate-chg-2026-0001-xxxxx
status: Failed
failedTaskCount: 1
failedTasks:
  - clone-repository
summary: Tekton validation failed in task clone-repository
```

Evidence must not include tokens or Secret values.

---

## 12. Failure diagnostics

The Control Plane must make Tekton failures understandable.

Current diagnostics include:

- failed task count;
- failed task names;
- TaskRun status;
- TaskRun reason;
- TaskRun message;
- validation summary.

Example summary:

```text
Tekton validation failed in task clone-repository
```

The UI and API should expose this information so the operator does not need to manually inspect all TaskRuns before understanding the failure point.

---

## 13. Anti-secret and GitOps guardrails

## 13.1 Purpose

The validation pipeline must prevent unsafe GitOps content from being accepted silently.

## 13.2 Guardrails

Current and expected guardrails include:

- block Kubernetes `Secret` manifests where policy forbids them;
- detect inline secret-like values;
- detect common token, password, private key and authorization patterns;
- allow safe external Secret references when policy permits;
- produce clear `policy violation` messages.

## 13.3 Examples of suspicious patterns

```text
password
token
client_secret
authToken
secret_key
private_key
PRIVATE KEY
AWS access key patterns
bearer
authorization
```

## 13.4 Design rule

Anti-secret checks must be treated as GitOps safety guardrails, not as a replacement for an enterprise Secret management platform.

---

## 14. Configuration

Tekton configuration baseline:

```text
TEKTON_NAMESPACE=devops-ci-demo
TEKTON_PIPELINE_NAME=validate-gitops
TEKTON_TIMEOUT_SECONDS=900
TEKTON_POLL_INTERVAL_SECONDS=5
TEKTON_VALIDATION_PATH=apps/demo-go-color-app
```

Future environment-aware configuration will resolve these values by target environment.

Rules:

- namespace must be configurable;
- pipeline name must be configurable;
- validation path must be configurable;
- timeouts must be explicit;
- credentials used by pipelines must come from Secrets;
- no token values may be printed in PipelineRun logs.

---

## 15. RBAC and security

## 15.1 Runtime ServiceAccount

The DevOps Control Plane runtime ServiceAccount must be able to:

- create `PipelineRun` resources in the configured Tekton namespace;
- get/list `PipelineRun` resources;
- get/list `TaskRun` resources;
- read runtime resources required for evidence where configured.

It must not have broad cluster-admin permissions.

## 15.2 Pipeline ServiceAccount

The PipelineRun ServiceAccount must have only the permissions required by the validation pipeline.

Depending on the pipeline design, it may require:

- repository clone credentials;
- read access to resources needed for dry-run or validation;
- no unnecessary access to Secrets.

## 15.3 Secret handling

- Git credentials used by the pipeline must be stored in Secrets.
- Git tokens must not be printed in Tekton logs.
- Secret values must not be stored in evidence.
- Validation output must be sanitized.

---

## 16. API endpoints related to Tekton validation

Current validation endpoints:

```text
POST /api/v1/changes/{id}/validate
POST /api/v1/changes/{id}/check-validation
GET  /api/v1/changes/{id}/evidence
GET  /api/v1/changes/{id}/evidence/validation
```

Legacy references without `/api/v1` should be considered historical and updated during documentation migration.

---

## 17. Relationship with GitLab and Argo CD

Tekton validation sits between GitLab change creation and Argo CD deployment state.

```text
GitLab branch / file update / merge request
  -> Tekton validation
  -> Argo CD deployment check
  -> Kubernetes/OpenShift runtime evidence
```

A validation failure should prevent the change from being treated as technically validated.

A validation success does not itself deploy anything. Deployment state is still observed through Argo CD and runtime evidence collection.

---

## 18. Multi-environment direction

Future environment-aware behavior will resolve Tekton settings from the environment catalog.

Example mapping:

```text
dev:
  namespace: devops-ci-demo
  pipelineName: validate-gitops
  validationPath: apps/demo-go-color-app/overlays/dev

staging:
  namespace: devops-ci-staging
  pipelineName: validate-gitops
  validationPath: apps/demo-go-color-app/overlays/staging

production:
  namespace: devops-ci-production
  pipelineName: validate-gitops
  validationPath: apps/demo-go-color-app/overlays/production
```

Production validation must remain guarded by environment-aware RBAC/AuthZ before production workflows are enabled.

---

## 19. Testing strategy

## 19.1 Unit tests

Test:

- PipelineRun request construction;
- PipelineRun status mapping;
- TaskRun status mapping;
- failed task diagnostics;
- validation evidence payload construction;
- error mapping;
- timeout behavior where implemented.

## 19.2 Fake Kubernetes client tests

Workflow services should be testable without a real cluster.

Example fake flow:

```text
CreatePipelineRun -> returns PipelineRunRef
CheckValidation -> Running
CheckValidation -> Succeeded
ListTaskRuns -> returns task statuses
```

## 19.3 Runtime validation

Runtime validation should cover:

- PipelineRun creation through the Control Plane;
- `check-validation` reading terminal status;
- TaskRun diagnostics collection;
- validation evidence persistence;
- policy guardrail success case;
- policy guardrail failure case;
- RBAC permissions for create/get/list PipelineRun and TaskRun resources.

## 19.4 Failure scenarios

Validate or simulate:

- namespace not found;
- pipeline not found;
- ServiceAccount permission failure;
- repository clone failure;
- manifest validation failure;
- anti-secret policy violation;
- PipelineRun failure;
- TaskRun diagnostics unavailable;
- timeout.

---

## 20. Completion checklist

The Tekton integration baseline is considered ready when:

- Tekton configuration is loaded safely;
- Kubernetes client can create PipelineRuns;
- PipelineRun contains ChangeRequest context;
- PipelineRun status can be checked;
- TaskRun diagnostics are collected;
- validation evidence is persisted;
- failed tasks are visible;
- anti-secret and manifest guardrails are active;
- runtime status is updated correctly;
- RBAC is least-privilege;
- runtime validation is documented.

---

## 21. Relationship with other documents

This document informs and is informed by:

- `docs/05-architecture.md`;
- `docs/07-gitlab-integration.md`;
- `docs/06-argocd-integration.md`;
- `docs/09-kubernetes-openshift-integration.md`;
- `docs/13-api-design.md`;
- `docs/04-non-functional-requirements.md`;
- `docs/environment-configuration-model.md`;
- `docs/change-promotion-model.md`;
- `docs/adr/ADR-0003-tekton-validation-engine.md`;
- `docs/adr/ADR-0008-kubernetes-api-for-tekton.md`.

---

## 22. Key message

Tekton transforms GitOps validation from a manual activity into a repeatable, observable and auditable process.

The Control Plane uses Tekton to answer a simple operational question:

```text
Is this GitOps change technically safe enough to proceed?
```

Tekton does not decide what is deployed. Tekton validates the GitOps change before Argo CD deployment state and OpenShift runtime evidence are evaluated.

This preserves the original spirit of the project: validation is explicit, traceable and associated with the ChangeRequest.

---

## 23. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial Tekton integration document in Italian. |
| 2026-07-06 | 0.2 | Rewritten in English and refreshed while preserving the original Tekton validation-engine intent and aligning it with the implemented validation, diagnostics and evidence baseline. |
