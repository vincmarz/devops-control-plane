# Runtime Non-Regression Validation With Factories Still Disabled

## Purpose

This document records the runtime non-regression validation completed in phase 15.8.8.7 after adding concrete runtime client factory adapters for Kubernetes, Tekton and Argo CD.

The goal was to prove that adding the concrete factory adapter code did not change the current runtime behavior because the factories are not yet wired into `main.go` and the runtime Secret value loader remains disabled.

## Validated Commit

Validation was performed against:

```text
52b8ac2 Document real runtime client factory readiness
```

The source branch was aligned with `origin/main` and the working tree was clean before evidence collection.

Recent implementation context:

```text
65b99d2 Add Argo CD runtime client factory adapter
7d02979 Add Tekton runtime client factory adapter
b87cd55 Add Kubernetes runtime client factory adapter
50eab45 Document real runtime client factory design
```

## Pre-Validation Source Checks

The following source checks confirmed that the runtime still uses the conservative baseline:

```text
runtimeSecretValueLoader := app.EmptyRuntimeSecretValueLoader{}
NewKubernetesRuntimeClientProviderFactoryAwareRegistry present
NewTektonRuntimeClientProviderFactoryAwareRegistry present
NewArgoCDRuntimeClientProviderFactoryAwareRegistry present
```

The concrete factory adapter constructors were not referenced in `main.go`:

```text
newKubernetesRuntimeClientFactoryAdapter absent
newTektonRuntimeClientFactoryAdapter absent
newArgoCDRuntimeClientFactoryAdapter absent
```

Automated tests were green:

```text
go test ./... -> OK
```

## Runtime Validation Setup

Runtime validation used a direct application port-forward to bypass the OAuth proxy and validate trusted-header RBAC behavior consistently with previous runtime validations.

```text
BASE_URL=http://127.0.0.1:18107
```

The validation used the configured runtime RBAC group:

```text
X-Forwarded-User: operator
X-Forwarded-Groups: devops-control-plane-operators
```

## Readiness Result

The readiness endpoint returned:

```text
readyz=200
```

This confirmed that application configuration and database connectivity were healthy before executing the runtime workflows.

## ChangeRequest Creation

A dev ChangeRequest was created using a complete API payload including `title` and `changeType`.

Result:

```text
create_dev=201
CHANGE_NUMBER=CHG-2026-0033
```

The created ChangeRequest had:

```text
applicationName=demo-go-color-app
targetEnvironment=dev
changeType=standard
status=draft
riskLevel=medium
requestedBy=operator
```

## Kubernetes Runtime Path: collect-evidence

The Kubernetes runtime evidence path returned:

```text
collect_evidence=202
```

This validates that the existing current-cluster Kubernetes runtime provider path remains operational after adding the concrete Kubernetes factory adapter code.

The concrete factory adapter was not used because it is not wired in `main.go`.

## Tekton Runtime Path: validate

The Tekton validate path returned:

```text
validate=202
```

A real Tekton PipelineRun was created:

```text
pipelineRunName=devops-cp-validate-chg-2026-0033-5bv8n
namespace=devops-ci-demo
```

This validates that the existing current-cluster Tekton runtime provider path remains operational after adding the concrete Tekton factory adapter code.

The concrete factory adapter was not used because it is not wired in `main.go`.

## Tekton Runtime Path: check-validation

The Tekton check-validation path returned:

```text
check_validation=202
```

The PipelineRun was observed in a valid in-progress state:

```text
status=Unknown
reason=Running
message=Tasks Completed: 0 (Failed: 0, Cancelled 0), Incomplete: 2, Skipped: 0
```

A TaskRun was observed as pending:

```text
taskRun=devops-cp-validate-chg-2026-0033-5bv8n-clone-repository
reason=Pending
message=pod status "Ready":"False"; message: "containers with unready status: [step-prepare-and-run]"
```

This is a valid transient Tekton state and confirms that the provider-aware Tekton check path resolved the current runtime client correctly.

## Argo CD Runtime Path: check-deployment

The Argo CD check-deployment path returned:

```text
check_deployment=202
runtimeStatus=DeploymentSyncedHealthy
```

The Argo CD application state was:

```text
applicationName=demo-go-color-app
project=devops-ci-demo
syncStatus=Synced
healthStatus=Healthy
revision=6bfdd07a96b4a23569a82bfb1e252e48eed248bc
```

This validates that the existing current-cluster Argo CD runtime provider path remains operational after adding the concrete Argo CD factory adapter code.

The concrete factory adapter was not used because it is not wired in `main.go`.

## Disabled Environment Guardrails

Staging and production remained disabled.

Results:

```text
create_staging=422
create_production=422
```

The API returned explicit validation errors:

```text
targetEnvironment "staging" is currently disabled
targetEnvironment "production" is currently disabled
```

## Secret Handling Confirmation

This validation did not enable real Secret loading.

The following guardrails remained in place:

- `EmptyRuntimeSecretValueLoader` remained configured in `main.go`;
- no concrete runtime client factory adapter was wired into `main.go`;
- no Kubernetes Secret value was read;
- no token, kubeconfig, CA bundle or Argo CD credential was printed;
- staging and production were not activated.

## Validation Summary

Final runtime validation matrix:

```text
readyz=200
create_dev=201
CHANGE_NUMBER=CHG-2026-0033
collect_evidence=202
validate=202
check_validation=202
check_deployment=202
create_staging=422
create_production=422
```

## Conclusion

Phase 15.8.8.7 completed successfully.

The addition of concrete Kubernetes, Tekton and Argo CD runtime client factory adapters did not change the current runtime behavior because the factories remain disabled in the composition root.

The validated baseline remains:

- `ocp-dev` / current-cluster runtime remains operational;
- factory-aware registries remain wired conservatively;
- concrete factories remain not wired;
- runtime Secret loading remains disabled;
- staging and production remain disabled;
- no runtime regression was observed.

## Next Step

The project can now proceed to the next controlled enablement phase.

Recommended next phase:

```text
15.8.9 — Controlled runtime Secret loader and factory enablement design
```

That phase should remain design-first and should not enable real Secret reads until the allow-list configuration, Secret getter wiring and rollback strategy are explicitly defined and validated.
