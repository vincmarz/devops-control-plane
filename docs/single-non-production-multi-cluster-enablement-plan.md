# Single Non-Production Multi-Cluster Enablement Plan

## Purpose

This document defines the plan for phase 15.8.10: enabling a single non-production runtime target for real multi-cluster support in the DevOps Control Plane.

The phase is intentionally design-first. It does not enable staging/collaudo, does not create runtime Secrets, does not read Kubernetes Secrets, and does not change runtime flags.

The objective is to define a complete, reversible and testable plan before the first controlled real multi-cluster enablement.

## Current Baseline

The current repository baseline is:

```text
ab4c05c Document controlled enablement plumbing validation
```

The validated runtime baseline before this plan is:

```text
1ec5941 Wire disabled runtime client factory builders
```

The controlled enablement plumbing is complete and validated with all flags disabled:

```text
DCP_RUNTIME_SECRET_LOADER_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORIES_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_KUBERNETES_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_TEKTON_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_ARGOCD_ENABLED=false
```

Effective current runtime behavior remains:

```text
runtime Secret loader = EmptyRuntimeSecretValueLoader
runtime client factories = nil
current-cluster provider = active for ocp-dev/current
staging = disabled
production = disabled
no Kubernetes Secret values are read
```

## Target of This Plan

The first real multi-cluster enablement target should be a single non-production environment:

```text
targetEnvironment=staging
functional meaning=collaudo
clusterName=ocp-staging
```

Production is explicitly out of scope for this phase.

## Current Model Inventory

The current fallback environment catalog already defines:

```text
dev        enabled=true   clusterName=ocp-dev
staging    enabled=false  clusterName=ocp-staging
production enabled=false  clusterName=ocp-production
```

The current fallback cluster registry already defines:

```text
ocp-dev        enabled=true
ocp-staging    enabled=false
ocp-production enabled=false
```

The current runtime provider registry default still exposes only:

```text
ocp-dev current-cluster provider
```

Therefore, for staging/collaudo to work as a real non-current cluster, the enablement must eventually configure all of the following together:

```text
Environment Catalog entry for staging enabled
Cluster Registry entry for ocp-staging enabled
Runtime provider entry for ocp-staging enabled
Runtime Secret references for ocp-staging configured
Runtime Secret loader allow-list for ocp-staging configured
Runtime Secret loader flag enabled
Runtime client factory global flag enabled
Required capability-specific factory flags enabled
```

The exact implementation of runtime provider registry externalization may be a later phase if not already available at the time of actual enablement.

## Non-Goals

This plan does not:

- enable staging/collaudo;
- enable production;
- create Kubernetes Secrets;
- read Kubernetes Secret values;
- change `main.go`;
- set runtime enablement flags;
- apply OpenShift manifests;
- introduce kubeconfig-based client construction;
- introduce raw CA material handling from Secrets;
- bypass existing RBAC or allow-list controls.

## Required Configuration Artifacts

The first non-production enablement should be based on explicit mounted configuration files.

Required files:

```text
/etc/dcp-environments/environment-catalog.yaml
/etc/dcp-clusters/cluster-registry.yaml
/etc/dcp-runtime-client-secrets/secret-refs.yaml
/etc/dcp-runtime-client-secrets/allowed-refs.yaml
```

The files must contain metadata and references only. Secret values must be stored only in Kubernetes Secrets and must never be committed to Git.

## Environment Catalog Template

The staging/collaudo environment must remain disabled until the final enablement gate.

Initial planning template:

```yaml
defaultEnvironment: dev
environments:
  - name: dev
    displayName: Development
    enabled: true
    category: development
    description: Current active development environment.
    clusterName: ocp-dev
    kubernetesNamespace: devops-ci-demo
    tektonNamespace: devops-ci-demo
    argocdApplicationName: demo-go-color-app
    gitTargetBranch: main
    allowTechnicalActions: true
    allowPromotionSource: true
    allowPromotionTarget: false

  - name: staging
    displayName: Staging / Collaudo
    enabled: false
    category: preproduction
    description: Controlled staging/collaudo runtime target. Enable only after runtime validation gates pass.
    clusterName: ocp-staging
    kubernetesNamespace: devops-ci-staging
    tektonNamespace: devops-ci-staging
    argocdApplicationName: demo-go-color-app
    gitTargetBranch: main
    allowTechnicalActions: false
    allowPromotionSource: false
    allowPromotionTarget: true

  - name: production
    displayName: Production
    enabled: false
    category: production
    description: Production remains disabled until dedicated production gates are approved.
    clusterName: ocp-production
    allowTechnicalActions: false
    allowPromotionSource: false
    allowPromotionTarget: true
```

Final staging enablement must change only after all prerequisites are validated:

```yaml
staging.enabled: true
staging.allowTechnicalActions: true
```

Those changes must be performed in a dedicated enablement phase, not in this planning phase.

## Cluster Registry Template

Initial planning template:

```yaml
clusters:
  - name: ocp-dev
    displayName: OpenShift Development
    enabled: true
    apiURL: https://api.ocp4.mim.lan:6443
    caConfigMapRef: dcp-cluster-ocp-dev-ca
    tokenSecretRef: dcp-cluster-ocp-dev-token
    defaultNamespace: devops-ci-demo
    allowedNamespaces:
      - devops-ci-demo
    description: Current OpenShift cluster used as active dev baseline.

  - name: ocp-staging
    displayName: OpenShift Staging / Collaudo
    enabled: false
    apiURL: https://api.ocp-staging.example:6443
    caConfigMapRef: dcp-cluster-ocp-staging-ca
    tokenSecretRef: dcp-cluster-ocp-staging-token
    defaultNamespace: devops-ci-staging
    allowedNamespaces:
      - devops-ci-staging
    description: First controlled non-production multi-cluster runtime target.

  - name: ocp-production
    displayName: OpenShift Production
    enabled: false
    apiURL: ""
    caConfigMapRef: dcp-cluster-ocp-production-ca
    tokenSecretRef: dcp-cluster-ocp-production-token
    defaultNamespace: devops-ci-production
    allowedNamespaces:
      - devops-ci-production
    description: Production remains disabled.
```

The `apiURL` value for `ocp-staging` must be replaced with the real staging cluster API URL during the controlled enablement phase.

## Runtime Secret Reference Template

Runtime Secret references must be defined in:

```text
/etc/dcp-runtime-client-secrets/secret-refs.yaml
```

Template:

```yaml
clusters:
  - clusterName: ocp-staging
    kubernetes:
      namespace: devops-control-plane
      name: staging-runtime-client
      tokenKey: kubernetes-token
    tekton:
      namespace: devops-control-plane
      name: staging-runtime-client
      tokenKey: kubernetes-token
    argocd:
      namespace: devops-control-plane
      name: staging-argocd-runtime-client
      tokenKey: argocd-token
      baseURLKey: argocd-base-url
```

Rules:

- the file stores only references and key names;
- it must not contain actual tokens, kubeconfigs, CA bundles or passwords;
- kubeconfig keys must remain unset until kubeconfig support is explicitly designed;
- CA keys should remain unset until safe CA handling is explicitly designed.

## Runtime Secret Loader Allow-List Template

Runtime Secret loader allow-list must be defined in:

```text
/etc/dcp-runtime-client-secrets/allowed-refs.yaml
```

Template:

```yaml
allowedClusters:
  - ocp-staging
allowedRefs:
  - clusterName: ocp-staging
    namespace: devops-control-plane
    name: staging-runtime-client
  - clusterName: ocp-staging
    namespace: devops-control-plane
    name: staging-argocd-runtime-client
```

Rules:

- only expected staging/collaudo Secret references may be present;
- production Secrets must not be allow-listed in this phase;
- the allow-list must contain references only;
- the allow-list must be reviewed before enabling the Secret loader flag.

## Kubernetes Secret Requirements

The actual Kubernetes Secrets must be created manually or through approved infrastructure automation in the DevOps Control Plane namespace.

Expected Secret names:

```text
staging-runtime-client
staging-argocd-runtime-client
```

Expected keys:

```text
staging-runtime-client:
  kubernetes-token

staging-argocd-runtime-client:
  argocd-token
  argocd-base-url
```

Secret value rules:

- values must not be printed in terminals;
- values must not be committed to Git;
- `oc describe secret` should not be used for evidence;
- use safe metadata checks only, such as key names and Secret existence;
- Secret values must be rotated after test phases if temporary tokens are used.

## RBAC Requirements

The DevOps Control Plane ServiceAccount must have the minimum permissions required to read only the allow-listed runtime client Secrets.

Minimum desired permission in the control-plane namespace:

```text
apiGroups: [""]
resources: ["secrets"]
verbs: ["get"]
resourceNames:
  - staging-runtime-client
  - staging-argocd-runtime-client
```

No broad list/watch permission should be granted for Secrets.

The staging cluster runtime token must have only the permissions required for the selected capability.

For a phased approach, the first capability should be Kubernetes evidence collection only, then Tekton, then Argo CD.

## Recommended Capability Enablement Order

Do not enable all runtime capabilities at once.

Recommended order:

```text
1. Configuration-only validation with all flags disabled
2. Secret loader enabled, factories disabled
3. Kubernetes factory enabled for staging only
4. Tekton factory enabled for staging only
5. Argo CD factory enabled for staging only
6. staging environment enabled for technical actions
```

Production must not be introduced in this sequence.

## Flag Enablement Order

Initial state:

```text
DCP_RUNTIME_SECRET_LOADER_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORIES_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_KUBERNETES_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_TEKTON_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_ARGOCD_ENABLED=false
```

Stage A — configuration mounted, all runtime enablement disabled:

```text
DCP_RUNTIME_SECRET_LOADER_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORIES_ENABLED=false
```

Stage B — Secret loader enabled but factories disabled:

```text
DCP_RUNTIME_SECRET_LOADER_ENABLED=true
DCP_RUNTIME_CLIENT_FACTORIES_ENABLED=false
```

Stage C — Kubernetes factory only:

```text
DCP_RUNTIME_SECRET_LOADER_ENABLED=true
DCP_RUNTIME_CLIENT_FACTORIES_ENABLED=true
DCP_RUNTIME_CLIENT_FACTORY_KUBERNETES_ENABLED=true
DCP_RUNTIME_CLIENT_FACTORY_TEKTON_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_ARGOCD_ENABLED=false
```

Stage D — Tekton factory only after Kubernetes validation:

```text
DCP_RUNTIME_CLIENT_FACTORY_TEKTON_ENABLED=true
```

Stage E — Argo CD factory only after Tekton validation:

```text
DCP_RUNTIME_CLIENT_FACTORY_ARGOCD_ENABLED=true
```

## Validation Matrix

### Baseline validation with all flags disabled

Expected:

```text
readyz=200
create dev=201
collect-evidence dev=202
validate dev=202
check-validation dev=202 or coherent in-progress state
check-deployment dev=202 with Synced/Healthy
create staging=422 disabled
create production=422 disabled
```

### Secret loader enabled but factories disabled

Expected:

```text
readyz=200
dev runtime paths unchanged
staging still disabled or not actionable
production disabled
no Secret value in logs or responses
no factory fallback for staging
```

### Kubernetes factory enabled for staging only

Prerequisite:

```text
staging environment must still be disabled for general create until the technical action path is deliberately tested
```

Expected targeted validation:

```text
factory constructed only when global and Kubernetes flags are true
Secret references resolved only for ocp-staging
Secret loader allow-list permits only staging allowedRefs
Secret read limited to allow-listed Secret names
Kubernetes evidence collection succeeds only for allowed namespace
```

### Tekton factory enabled for staging only

Expected targeted validation:

```text
Tekton PipelineRun created in devops-ci-staging
PipelineRun status observable
TaskRun status observable
no access to devops-ci-production
no production target enabled
```

### Argo CD factory enabled for staging only

Expected targeted validation:

```text
Argo CD BaseURL resolved from allowed Secret reference
Argo CD application status returned for staging application
Synced/Healthy check succeeds or returns a coherent controlled error
no production Argo CD reference configured
```

## Rollback Plan

Rollback must be configuration-first.

Immediate rollback:

```text
DCP_RUNTIME_SECRET_LOADER_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORIES_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_KUBERNETES_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_TEKTON_ENABLED=false
DCP_RUNTIME_CLIENT_FACTORY_ARGOCD_ENABLED=false
```

Then restart deployment if flags are environment-variable based:

```text
oc rollout restart deployment/devops-control-plane -n devops-control-plane
oc rollout status deployment/devops-control-plane -n devops-control-plane
```

Expected rollback runtime state:

```text
readyz=200
dev current-cluster paths work
staging disabled
production disabled
no Secret read
```

## Stop Conditions

Stop the enablement immediately if any of the following occurs:

- `/readyz` fails;
- dev current-cluster runtime paths regress;
- staging unexpectedly becomes creatable before explicit enablement;
- production becomes creatable or technically actionable;
- any Secret value appears in logs, responses or evidence;
- Secret loader reads a Secret that is not allow-listed;
- factory construction succeeds without required global and capability flags;
- rollback does not restore the disabled posture.

## Required Evidence Without Secret Disclosure

Allowed evidence:

```text
HTTP status codes
ChangeRequest numbers
PipelineRun names
Argo CD sync/health status
ConfigMap and Secret metadata names
RBAC resourceNames
safe summaries
```

Forbidden evidence:

```text
token values
Authorization headers
kubeconfig content
CA content
Secret data payloads
base64 Secret values
raw environment dumps containing credentials
```

## Recommended Next Implementation Phases

Recommended follow-up sequence:

```text
15.8.10.1 Inventory environment/cluster/runtime enablement baseline
15.8.10.2 Document single non-production enablement plan
15.8.10.3 Add disabled example templates for staging Secret refs and allow-list
15.8.10.4 Add RBAC template for read-only allow-listed staging runtime Secrets
15.8.10.5 Runtime validation with templates mounted and all flags disabled
15.8.10.6 Controlled Secret loader enablement with factories disabled
15.8.10.7 Controlled Kubernetes-only staging factory enablement plan
```

The plan should be revisited before any real flag is set to true.

## Conclusion

The DevOps Control Plane is ready to plan the first single non-production multi-cluster enablement, but not yet to activate it.

The next safe step is to add reviewed, disabled templates for staging/collaudo Secret references, allow-list entries and RBAC, followed by another default-disabled runtime validation.
