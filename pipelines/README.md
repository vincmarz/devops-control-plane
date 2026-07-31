# Tekton pipelines

This directory contains Tekton resources used by the DevOps Control Plane lab.

## Source of truth

The source-of-truth pipeline definitions are:

```text
pipelines/validate-gitops.yaml
pipelines/go-build-and-push.yaml
```

`validate-gitops` validates the GitOps repository. `go-build-and-push` builds
an application source commit and publishes an image with immutable digest results.

Runtime namespace:

```text
devops-ci-demo
```

Pipeline name:

```text
validate-gitops
```

The DevOps Control Plane OpenShift deployment must be configured with:

```text
TEKTON_NAMESPACE=devops-ci-demo
TEKTON_PIPELINE_NAME=validate-gitops
TEKTON_BUILD_PIPELINE_NAME=go-build-and-push
TEKTON_GIT_URL=https://github.com/vincmarz/demo-app-gitops.git
TEKTON_GIT_REVISION_TEMPLATE=change/{changeNumber}
TEKTON_VALIDATION_PATH=apps/demo-go-color-app
TEKTON_IMAGE=unused
```

## What the pipeline validates

The pipeline:

1. clones the GitOps repository using the `git-clone` Tekton Task;
2. checks that `VALIDATION_PATH` exists;
3. lists YAML manifests under the GitOps path;
4. rejects obvious secret-like content and Kubernetes `Secret` manifests;
5. validates rendered Kustomize output with `oc apply --dry-run=client`;
6. validates plain manifests outside `kustomization.yaml` descriptors.

## Compatibility notes

The `IMAGE` parameter and `dockerconfig` workspace are currently retained for compatibility with the existing DevOps Control Plane Tekton adapter, but this validation pipeline does not use them directly.

## MVP policy and anti-secret guardrails

The current MVP validation pipeline includes lightweight policy checks directly in the Tekton validation script.

### Blocked

The pipeline fails when it finds:

- Kubernetes `Secret` manifests under the validated GitOps path;
- obvious inline secret-like values, such as `password: ...`, `token: ...`, `client_secret`, `authToken`, `secret_key`, private key headers, AWS access key markers, bearer/authorization inline values.

### Allowed as references

The MVP does not block normal references to externally managed secrets, for example:

```yaml
secretName: existing-platform-secret
imagePullSecrets:
  - name: existing-pull-secret
```

These references are expected to point to secrets managed outside the GitOps repository.

### Future hardening

Future phases may replace or complement the shell-based checks with a policy engine such as OPA/Conftest or cluster-side policy admission through Kyverno/OpenShift policy mechanisms.

## Application build pipeline

The build pipeline accepts `GIT_URL`, `GIT_REVISION`, and `IMAGE`. The Control
Plane must pass an immutable commit as `GIT_REVISION` and a deterministic image
tag instead of relying on `latest`.

The pipeline exposes these results:

```text
SOURCE_COMMIT
IMAGE_URL
IMAGE_DIGEST
```

`git-clone` and `buildah` are namespaced Tasks supplied and managed by OpenShift
Pipelines. Their full Task definitions are intentionally not duplicated here.
