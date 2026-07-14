# ChangeRequest Application-to-Environment Binding Runtime Validation

## Purpose

This document records the runtime validation of the fail-closed binding between a `ChangeRequest.applicationName` value and the logical application configured for the selected `targetEnvironment`.

The validation was performed after merge commit `898bed5`, which integrated pull request `#24`, titled `Validate ChangeRequest application environment binding`.

## Validated behavior

The implemented rule is:

```text
ChangeRequest.applicationName
must match
EnvironmentDefinition.applicationName
for the selected targetEnvironment
```

A matching request is accepted. A mismatching request is rejected before persistence and before execution of technical workflows.

## Environment Catalog baseline

The versioned manifest and the runtime ConfigMap were compared and found equivalent after regenerating both comparison files.

The validated namespace-isolated mappings are:

```text
dev        -> ocp-dev / devops-ci-demo       / demo-go-color-app
staging    -> ocp-dev / devops-ci-staging    / demo-go-color-app-staging
production -> ocp-dev / devops-ci-production / demo-go-color-app-production
```

All three environments use the logical application name:

```text
demo-go-color-app
```

The runtime ConfigMap is:

```text
namespace: devops-control-plane
name: devops-control-plane-environments
```

The ConfigMap is mounted in the application container at:

```text
/etc/dcp-environments
```

## Build and deployment evidence

The application image was built from merge commit:

```text
898bed5
```

The immutable image tag is:

```text
devops-control-plane:898bed5
```

The published image digest is:

```text
sha256:70c6ad6b7856b5bbe0ee09c2b3e39b20f5c81b22434c5cf2ecbd0165af34a27b
```

The ImageStream digest and the running pod image digest matched exactly.

The deployment and runtime pod were:

```text
deployment: devops-control-plane
pod: devops-control-plane-84568648cb-8bkmx
```

The rollout completed successfully. At the end of validation:

```text
phase: Running
container readiness: true,true
restart count: 0,0
```

The readiness endpoint returned:

```json
{
  "data": {
    "checks": {
      "configuration": "ok",
      "database": "ok"
    },
    "status": "ready"
  },
  "error": null
}
```

## Positive runtime validation

A request with a matching application and environment binding was submitted:

```text
applicationName: demo-go-color-app
targetEnvironment: dev
```

Result:

```text
HTTP status: 201
ChangeRequest: CHG-2026-0051
status: draft
riskLevel: medium
requestedBy: operator@example.test
```

This confirms that a valid application-to-environment binding is accepted and persisted.

## Fail-closed runtime validation

A request with a mismatching application was submitted:

```text
applicationName: payments
targetEnvironment: dev
```

Result:

```text
HTTP status: 422
error code: VALIDATION_INVALID_REQUEST
recoverable: true
```

The technical message was:

```text
applicationName "payments" is not configured for targetEnvironment "dev"; expected "demo-go-color-app"
```

This confirms that an inconsistent binding is rejected before persistence and before technical workflow execution.

## RBAC context

The runtime validation used the operator group configured in the deployed ConfigMap:

```text
AUTH_GROUP_OPERATOR=devops-control-plane-operators
```

The initial use of the unit-test-only group `cp-operators` returned HTTP `403`, confirming that runtime tests must use the deployed RBAC configuration rather than test fixture values.

No credentials, bearer tokens, cookie secrets, or raw Secret values are recorded in this document.

## Source and test evidence

The implementation added:

- `EnvironmentCatalog.ValidateApplicationBinding`;
- fail-closed wiring in `ChangeService.Create`;
- application bindings in the versioned Environment Catalog manifest;
- catalog validation tests;
- ChangeService wiring validation;
- API fixture alignment.

The following tests passed without using the existing Go build cache:

```text
go test -count=1 ./internal/app ./internal/api
go test -count=1 ./...
```

The pull request CI check also completed successfully.

## Conclusion

The application-to-environment binding is implemented, tested, deployed, and validated end-to-end on OpenShift.

The validated behavior is:

```text
matching application binding    -> accepted
mismatching application binding -> rejected fail-closed with HTTP 422
```

The runtime remains healthy after rollout, with matching image digests, successful readiness checks, and no container restarts.
