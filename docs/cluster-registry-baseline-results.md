# Cluster Registry Baseline Results

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** Cluster Registry baseline results
- **Phase:** 15.3.3 — Document Cluster Registry baseline results
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Baseline commit before this phase:** `4d5f524` — `Align Go module dependencies with build image`
- **Cluster Registry baseline commit/runtime image:** `52de0c1` — `Add cluster registry baseline`
- **Related phases:**
  - 15.1.1 — Add multi-cluster environment enablement request
  - 15.2.2 — Load Environment Catalog from mounted file
  - 15.3.1 — Cluster Registry inventory
  - 15.3.2 — Add Cluster Registry baseline
- **Status:** Completed validation report
- **Language:** English

---

## 1. Purpose

This document records the result of Phase 15.3, which introduced the first Cluster Registry baseline for the DevOps Control Plane.

The purpose of this baseline is to prepare the DevOps Control Plane for a future multi-cluster OpenShift architecture while keeping the current runtime behavior conservative and safe.

The target architecture remains:

```text
One DevOps Control Plane instance
managing multiple logical environments
each environment mapped to a cluster identity
```

Target logical mapping:

```text
dev        -> ocp-dev
staging    -> ocp-staging
production -> ocp-production
```

At the time of this phase, only one OpenShift cluster is available for runtime validation. Therefore this phase introduces the registry and environment-to-cluster metadata but does not yet implement multi-cluster client selection.

---

## 2. Scope

Phase 15.3 covers:

- inventory of the current runtime and repository state before Cluster Registry introduction;
- introduction of a Cluster Registry application model;
- introduction of a Cluster Registry ConfigMap manifest;
- introduction of `clusterName` in the Environment Catalog model and manifest;
- runtime mounting of the Cluster Registry ConfigMap;
- runtime validation that the existing `targetEnvironment` behavior remains unchanged;
- confirmation that the DevOps Control Plane remains fail-closed for disabled or unknown environments.

Phase 15.3 does not cover:

- selecting Kubernetes clients dynamically per cluster;
- selecting Tekton clients dynamically per cluster;
- selecting Argo CD instances dynamically per cluster;
- connecting to real staging or production OpenShift clusters;
- enabling `staging` or `production`.

Those capabilities are planned for later phases.

---

## 3. Starting point

Before this phase, the DevOps Control Plane already had runtime Environment Catalog loading from a mounted ConfigMap file.

The Environment Catalog was loaded from:

```text
/etc/dcp-environments/environments.yaml
```

The runtime variable was:

```text
ENVIRONMENT_CATALOG_FILE=/etc/dcp-environments/environments.yaml
```

The existing catalog states were:

```text
dev         configured, enabled
staging     configured, disabled
production  configured, disabled
unknown-env not configured
```

The validated behavior before this phase was:

```text
missing targetEnvironment -> HTTP 201, default dev
targetEnvironment=dev -> HTTP 201
targetEnvironment=staging -> HTTP 422, currently disabled
targetEnvironment=production -> HTTP 422, currently disabled
targetEnvironment=unknown-env -> HTTP 422, not configured
```

---

## 4. Phase 15.3.1 — Cluster Registry inventory

### 4.1 Repository baseline

The repository baseline at the start of the inventory was:

```text
4d5f524 Align Go module dependencies with build image
4c1706e Load environment catalog from mounted file
a203112 Add multi-cluster environment enablement request
```

The working tree was clean before the Cluster Registry patch was applied.

### 4.2 Runtime baseline

The DevOps Control Plane runtime image at inventory time was:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:4d5f524
```

Runtime mounts on the application container were:

```text
mount=dcp-trust-bundle path=/etc/dcp-trust
mount=dcp-environments path=/etc/dcp-environments
```

There was no Cluster Registry mount yet.

### 4.3 Environment Catalog runtime loading confirmed

The live configuration exposed:

```text
ENVIRONMENT_CATALOG_FILE=/etc/dcp-environments/environments.yaml
```

The repository and live Environment Catalog contained:

```text
dev:
  enabled: true
  kubernetesNamespace: devops-ci-demo
  tektonNamespace: devops-ci-demo
  argocdApplicationName: demo-go-color-app

staging:
  enabled: false

production:
  enabled: false
```

### 4.4 Cluster Registry absence confirmed

The inventory confirmed that no Cluster Registry implementation existed yet.

Observed counts:

```text
cluster_registry_go_files_count=0
cluster_configmap_files_count=0
```

No `CLUSTER_REGISTRY_FILE` setting was present before this phase.

### 4.5 Single-cluster technical configuration still present

The inventory confirmed that the technical runtime still contains single-cluster defaults:

```text
KUBERNETES_API_URL
KUBERNETES_NAMESPACE
TEKTON_NAMESPACE
TEKTON_PIPELINE_NAME
```

This is expected and remains acceptable for the current baseline.

The single-cluster technical configuration will be addressed in a later environment-to-cluster client selection phase.

---

## 5. Phase 15.3.2 — Cluster Registry baseline implementation

### 5.1 Commit

The baseline was implemented and pushed with:

```text
52de0c1 Add cluster registry baseline
```

### 5.2 Files introduced

New files:

```text
internal/app/cluster_registry.go
internal/app/cluster_registry_test.go
manifests/configmap-clusters.yaml
```

### 5.3 Files updated

Updated files:

```text
internal/app/environment_catalog.go
manifests/configmap-environments.yaml
manifests/configmap.yaml
manifests/deployment.yaml
```

### 5.4 Cluster Registry model

The new application model introduces:

```text
ClusterDefinition
ClusterRegistry
```

The baseline `ClusterDefinition` contains non-secret cluster metadata:

```text
name
displayName
enabled
apiURL
caConfigMapRef
tokenSecretRef
defaultNamespace
allowedNamespaces
description
```

The model deliberately stores only references to sensitive objects, not sensitive values.

No token, password or private key is stored in the Cluster Registry.

### 5.5 Cluster Registry functions

The implementation introduces:

```text
DefaultClusterRegistry
DefaultClusterRegistryFallback
LoadClusterRegistryFromFile
ParseClusterRegistryYAML
NewClusterRegistry
Resolve
IsEnabled
ValidateConfiguredCluster
ValidateEnvironmentCatalog
```

The fallback registry remains conservative and prepares the future target topology without enabling non-dev clusters.

### 5.6 Cluster Registry baseline entries

The baseline Cluster Registry contains:

```text
ocp-dev:
  enabled: true
  apiURL: https://api.ocp4.mim.lan:6443
  defaultNamespace: devops-ci-demo

ocp-staging:
  enabled: false
  apiURL: ""
  defaultNamespace: devops-ci-staging

ocp-production:
  enabled: false
  apiURL: ""
  defaultNamespace: devops-ci-production
```

The `apiURL` for `ocp-staging` and `ocp-production` is intentionally empty because those clusters are not connected yet.

### 5.7 Repository ConfigMap

The new repository manifest is:

```text
manifests/configmap-clusters.yaml
```

It defines:

```text
ConfigMap: devops-control-plane-clusters
Key: clusters.yaml
```

The ConfigMap contains the baseline logical clusters:

```text
ocp-dev
ocp-staging
ocp-production
```

### 5.8 Runtime configuration

The application ConfigMap now defines:

```text
CLUSTER_REGISTRY_FILE=/etc/dcp-clusters/clusters.yaml
```

### 5.9 Runtime mount target

The repository Deployment manifest now includes:

```text
volumeMount:
  dcp-clusters -> /etc/dcp-clusters

volume:
  dcp-clusters -> ConfigMap devops-control-plane-clusters
```

The mounted file path is:

```text
/etc/dcp-clusters/clusters.yaml
```

### 5.10 Environment Catalog cluster mapping

The Environment Catalog model now includes:

```text
clusterName
```

The baseline environment-to-cluster mapping is:

```text
dev        -> ocp-dev
staging    -> ocp-staging
production -> ocp-production
```

This mapping is present both in the Go fallback catalog and in the repository ConfigMap manifest.

---

## 6. Validation before commit

The implementation was validated before commit with:

```text
go test ./...
git diff --check
oc apply --dry-run=client -f manifests/configmap-clusters.yaml
oc apply --dry-run=client -f manifests/configmap-environments.yaml
oc apply --dry-run=client -f manifests/configmap.yaml
oc apply --dry-run=client -f manifests/deployment.yaml
```

Observed results:

```text
go test ./... OK
git diff --check OK
configmap/devops-control-plane-clusters created (dry run)
configmap/devops-control-plane-environments configured (dry run)
configmap/devops-control-plane-config configured (dry run)
deployment.apps/devops-control-plane configured (dry run)
```

Targeted checks confirmed:

```text
CLUSTER_REGISTRY_FILE present in manifests/configmap.yaml
dcp-clusters mount present in manifests/deployment.yaml
clusterName present in manifests/configmap-environments.yaml
DefaultClusterRegistry present
LoadClusterRegistryFromFile present
ValidateEnvironmentCatalog present
```

---

## 7. Runtime deployment

### 7.1 Build and push

The image was built and pushed with tag:

```text
52de0c1
```

The build completed successfully using the existing Go builder image:

```text
docker.io/library/golang:1.22
```

The pushed runtime image is:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:52de0c1
```

### 7.2 ConfigMaps applied live

The following ConfigMaps were applied live:

```text
devops-control-plane-clusters
devops-control-plane-environments
devops-control-plane-config
```

`devops-control-plane-clusters` contained:

```text
ocp-dev
ocp-staging
ocp-production
```

`devops-control-plane-config` contained:

```text
CLUSTER_REGISTRY_FILE=/etc/dcp-clusters/clusters.yaml
```

### 7.3 Deployment rollout

The Deployment was updated to runtime image:

```text
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:52de0c1
```

The rollout completed successfully.

### 7.4 Runtime mount patch

Because the live Deployment includes OAuth Proxy runtime customizations, the full repository Deployment manifest was not applied directly.

Instead, a targeted live JSON patch was used to add:

```text
volume: dcp-clusters
volumeMount: /etc/dcp-clusters
```

This preserved the existing OAuth Proxy sidecar and existing live mounts.

### 7.5 Runtime mount validation

Runtime container state after rollout and patch:

```text
container=oauth-proxy
image=registry.redhat.io/openshift4/ose-oauth-proxy:latest
mount=oauth-proxy-tls path=/etc/tls/private

container=devops-control-plane
image=image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:52de0c1
mount=dcp-trust-bundle path=/etc/dcp-trust
mount=dcp-environments path=/etc/dcp-environments
mount=dcp-clusters path=/etc/dcp-clusters
```

The mounted Cluster Registry file was visible in the pod as:

```text
/etc/dcp-clusters/clusters.yaml -> ..data/clusters.yaml
```

---

## 8. Runtime validation

### 8.1 Readiness

Readiness check:

```text
/readyz -> HTTP 200
```

### 8.2 Missing target environment

Request without `targetEnvironment` returned:

```text
HTTP 201
changeNumber=CHG-2026-0020
targetEnvironment=dev
requestedBy=cluster-registry-admin-a
status=draft
```

Conclusion:

```text
Default environment behavior remains unchanged.
```

### 8.3 Dev target environment

Request with `targetEnvironment=dev` returned:

```text
HTTP 201
changeNumber=CHG-2026-0021
targetEnvironment=dev
requestedBy=cluster-registry-admin-b
status=draft
```

Conclusion:

```text
dev remains enabled after Cluster Registry introduction.
```

### 8.4 Staging target environment

Request with `targetEnvironment=staging` returned:

```text
HTTP 422
errorCode=VALIDATION_INVALID_REQUEST
technicalMessage=targetEnvironment "staging" is currently disabled
```

Conclusion:

```text
staging remains configured but disabled.
```

### 8.5 Production target environment

Request with `targetEnvironment=production` returned:

```text
HTTP 422
errorCode=VALIDATION_INVALID_REQUEST
technicalMessage=targetEnvironment "production" is currently disabled
```

Conclusion:

```text
production remains configured but disabled.
```

### 8.6 Unknown target environment

Request with `targetEnvironment=unknown-env` returned:

```text
HTTP 422
errorCode=VALIDATION_INVALID_REQUEST
technicalMessage=targetEnvironment "unknown-env" is not configured
```

Conclusion:

```text
unknown-env remains not configured.
```

### 8.7 Runtime validation summary

```text
readyz=200
missing_target_env_http=201
dev_target_env_http=201
staging_target_env_http=422
production_target_env_http=422
unknown_target_env_http=422
```

---

## 9. Operational notes

### 9.1 Port-forward note

During one retest attempt, local port `18098` was already occupied by an existing listener.

Observed message:

```text
bind: address already in use
```

The final API validation nevertheless completed successfully because a listener on `18098` was available and answered the expected requests.

Before future retests, use a fresh port or verify the listener state.

### 9.2 Accidental local files cleanup

Some accidental local files were created by malformed shell input during repeated build/push commands.

The files were removed explicitly:

```text
build
--authfile
--root
--runroot
--tls-verify=false
--tmpdir
push
devops-control-plane:52de0c1
Podman step/hash-like filenames
```

Final repository status was clean.

---

## 10. Current limitations

The Cluster Registry baseline is now present, but the technical adapters still use the current single-cluster configuration.

Current execution still relies on settings such as:

```text
KUBERNETES_API_URL
KUBERNETES_NAMESPACE
TEKTON_NAMESPACE
TEKTON_PIPELINE_NAME
```

The Cluster Registry is not yet used to select runtime Kubernetes, Tekton or Argo CD clients.

This is intentional for this phase.

The next technical phase must introduce environment-to-cluster client selection in a controlled and testable way.

---

## 11. Security considerations

The Cluster Registry stores only non-secret metadata and Secret/ConfigMap reference names.

It does not store:

```text
bearer tokens
passwords
private keys
client certificates
raw kubeconfigs
```

Future cluster credentials must remain in:

```text
Kubernetes/OpenShift Secrets
ServiceAccount tokens
external secret management systems
```

Evidence and logs must continue to avoid printing sensitive values.

---

## 12. Acceptance summary

Phase 15.3 is accepted because:

- the absence of Cluster Registry was inventoried;
- the Cluster Registry model was introduced;
- the Cluster Registry repository manifest was added;
- the Environment Catalog now includes `clusterName`;
- `dev`, `staging` and `production` now map to `ocp-dev`, `ocp-staging` and `ocp-production`;
- the Cluster Registry ConfigMap is applied live;
- the Cluster Registry file is mounted into the pod;
- the runtime image `52de0c1` is deployed;
- `/readyz` returns HTTP 200;
- existing target environment validation behavior remains unchanged;
- staging and production remain disabled;
- unknown environments remain rejected;
- the working tree was clean after cleanup.

---

## 13. Recommended next step

Proceed with:

```text
Phase 15.4 — Environment-to-cluster client selection design/runtime baseline
```

Recommended initial goal for Phase 15.4:

```text
Resolve targetEnvironment -> Environment Catalog -> clusterName -> Cluster Registry
```

without yet enabling staging or production.

The next runtime increment should prove that technical workflows can determine the intended cluster metadata for a ChangeRequest, while still executing only against the current dev cluster until multi-cluster credentials and target clusters are ready.
