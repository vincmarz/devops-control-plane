# Git Provider and GitOps Runtime Correlation

## 1. Purpose

This document clarifies the relationship between source repositories, the SCM provider, the GitOps repository, Argo CD Applications, environments, namespaces, and ChangeRequests in the DevOps Control Plane validated technical baseline.

GitHub, GitLab CE, Argo CD, and Tekton have distinct roles. They must not be described as one automated chain unless that correlation has been explicitly demonstrated.

## 2. Baseline summary

The current baseline includes:

- the `devops-control-plane` product repository hosted on GitHub;
- internal GitLab CE used as an operational SCM provider through REST APIs;
- GitHub repository `demo-app-gitops` used as the current source of truth by Argo CD and Tekton;
- one primary logical application, `demo-go-color-app`;
- three environment instances: dev, staging, and production;
- one historical standalone Argo CD Application, `demo-app`.

No automatic synchronization between the internal GitLab project and the GitHub repository consumed by Argo CD has been demonstrated.

## 3. Distinct Git roles

### 3.1 Product source repository

```text
Provider: GitHub
Repository: vincmarz/devops-control-plane
```

Role:

- Go backend source code;
- documentation;
- pull requests;
- GitHub Actions CI quality gate;
- Containerfile and deployment assets.

### 3.2 GitLab development and validation project

```text
Provider: internal GitLab CE
Group: devops-lab
Project: devops-lab/demo-go-color-app-gitops
```

Demonstrated role:

- GitLab REST adapter validation;
- branch creation;
- file updates;
- Merge Request creation and lookup;
- merge execution;
- ChangeRequest event and evidence production.

### 3.3 GitOps runtime repository

```text
Provider: GitHub
Repository: https://github.com/vincmarz/demo-app-gitops.git
```

Demonstrated role:

- source repository configured in current Argo CD Applications;
- repository used by Tekton GitOps validation;
- desired state reconciled to OpenShift namespaces.

## 4. GitLab integration in the backend

`internal/adapters/gitlab` contains a concrete REST client implementing:

```text
Ping
CreateBranch
CreateOrUpdateFile
OpenMergeRequest
FindOpenMergeRequest
MergeMergeRequest
```

`cmd/devops-control-plane/main.go` creates and injects the client when the GitLab base URL, token, and project ID are configured.

```text
ChangeService
-> Git operation ports based on function types
-> GitLab adapter
-> GitLab REST API v4
```

GitLab is used as an SCM provider through APIs, not as the CI engine for the Control Plane repository.

## 5. SCM multiprovider boundary

Git operations are abstracted through application ports, but GitLab is the observed concrete implementation.

The current baseline does not include:

- an equivalent GitHub SCM adapter;
- a generic SCM provider registry;
- an SCM provider factory;
- dynamic SCM provider selection per logical application.

The SCM component is therefore not yet a complete multiprovider framework comparable to the provider-aware Kubernetes, Tekton, and Argo CD runtime model.

## 6. Source of the Applications UI

The Applications UI does not list GitLab projects. The implemented call chain is:

```text
uiApplicationsData
-> ApplicationService.List
-> ArgoCDApplicationClient.ListApplications
-> Argo CD GET /api/v1/applications
```

The original count of four Applications represented four observed Argo CD resources, not four logical applications and not four GitLab projects.

PR `#18`, merge commit `086e6e6`, corrected the UI wording to explicitly describe Argo CD Applications observed through Argo CD.

## 7. Observed Argo CD Applications

### 7.1 Historical standalone baseline

```text
Name: demo-app
Repository: https://github.com/vincmarz/demo-app-gitops.git
Path: manifests
Namespace: devops-ci-demo
Status: Synced / Healthy
Condition: OrphanedResourceWarning
```

An end-to-end GitLab workflow for `demo-app` has not been demonstrated.

### 7.2 Development instance

```text
Name: demo-go-color-app
Path: apps/demo-go-color-app
Namespace: devops-ci-demo
Status: Synced / Healthy
```

### 7.3 Staging instance

```text
Name: demo-go-color-app-staging
Path: apps/demo-go-color-app/overlays/staging
Namespace: devops-ci-staging
Status: Synced / Healthy
```

### 7.4 Production instance

```text
Name: demo-go-color-app-production
Path: apps/demo-go-color-app/overlays/production
Namespace: devops-ci-production
Status: Synced / Healthy
```

All four Argo CD Applications use `https://github.com/vincmarz/demo-app-gitops.git`.

## 8. Logical application and environment instances

The tested baseline is single-application and multi-environment:

```text
demo-go-color-app
|-- dev
|   |-- Argo CD Application: demo-go-color-app
|   `-- namespace: devops-ci-demo
|-- staging
|   |-- Argo CD Application: demo-go-color-app-staging
|   `-- namespace: devops-ci-staging
`-- production
    |-- Argo CD Application: demo-go-color-app-production
    `-- namespace: devops-ci-production
```

All three environments currently run on `ocp-dev`. The demonstrated isolation boundary is the namespace. Multi-cluster support is code-ready, but physical cross-cluster validation remains deferred.

## 9. Explicit Environment Catalog bindings

PR `#19`, merge commit `61783d2`, introduced the explicit field:

```yaml
applicationName: demo-go-color-app
```

Each environment is bound to the logical application and its concrete Argo CD Application.

```yaml
- name: dev
  applicationName: demo-go-color-app
  clusterName: ocp-dev
  kubernetesNamespace: devops-ci-demo
  tektonNamespace: devops-ci-demo
  argocdApplicationName: demo-go-color-app
  validationPath: apps/demo-go-color-app
```

```yaml
- name: staging
  applicationName: demo-go-color-app
  clusterName: ocp-dev
  kubernetesNamespace: devops-ci-staging
  tektonNamespace: devops-ci-staging
  argocdApplicationName: demo-go-color-app-staging
  validationPath: apps/demo-go-color-app/overlays/staging
```

```yaml
- name: production
  applicationName: demo-go-color-app
  clusterName: ocp-dev
  kubernetesNamespace: devops-ci-production
  tektonNamespace: devops-ci-production
  argocdApplicationName: demo-go-color-app-production
  validationPath: apps/demo-go-color-app/overlays/production
```

The correlation is not inferred from environment suffixes or GitOps paths.

## 10. Updated UI representation

The UI now presents:

```text
Logical Applications: 1
Environment instances: 3
Standalone Argo CD Applications: 1
```

`demo-go-color-app` is grouped across dev, staging, and production. `demo-app` is displayed separately as a standalone Argo CD Application.

If a catalog binding exists but the corresponding Argo CD Application is not observed, the UI displays `Not observed`.

## 11. Deployment evidence

The grouping change was deployed with:

- Pull request: `#19`
- Merge commit: `61783d2`
- Image tag: `devops-control-plane:61783d2`
- Image digest: `sha256:7d9297431d370fe8c23a32cc68a52f300fa0332fbe34ef28fd99ede5bfecdd7e`
- Observed pod: `devops-control-plane-77f6d6d758-jnvlj`
- Pod phase: `Running`
- Backend ready: `true`
- Backend restarts: `0`
- OAuth proxy ready: `true`
- OAuth proxy restarts: `0`

The ConfigMap contains three `applicationName: demo-go-color-app` bindings. Readiness reported `configuration: ok`, `database: ok`, and `status: ready`. An authenticated browser check confirmed the grouping.

## 12. ChangeRequest and environment resolution

A ChangeRequest stores:

```text
applicationName
targetEnvironment
```

The Technical Runtime Target resolves cluster, Kubernetes namespace, Tekton namespace, Argo CD Application, target Git branch, and validation path from the catalog.

This prevents `demo-go-color-app-production` from being treated as an independent logical application.

## 13. Correlation not demonstrated

The following chain has not been demonstrated:

```text
GitLab CE
-> automatic mirror or promotion
-> GitHub demo-app-gitops
-> Argo CD
```

The following paths have been demonstrated separately:

```text
DevOps Control Plane
-> GitLab REST API
-> branch / file / Merge Request / merge
```

```text
GitHub demo-app-gitops
-> Argo CD and Tekton
-> OpenShift namespaces
```

Documentation must not claim an automatic mirror or promotion workflow without dedicated operational evidence.

## 14. Designed capability, tested baseline, and historical resources

### 14.1 Designed capability

- Multiple logical applications
- Multiple environments
- Provider-aware runtime clients
- Multi-cluster readiness
- Abstract SCM operations

### 14.2 Tested baseline

- Logical application: `demo-go-color-app`
- SCM provider: GitLab CE
- GitOps runtime repository: GitHub `demo-app-gitops`
- Environments: dev, staging, production
- Physical cluster: `ocp-dev`
- Isolation: namespaces
- Environment-specific Argo CD Applications: three
- GitLab branch/Merge Request/merge workflow: validated
- GitLab-to-GitHub synchronization: not demonstrated

### 14.3 Historical resource

- `demo-app`
- Standalone Argo CD Application
- Path: `manifests`
- `Synced/Healthy` with `OrphanedResourceWarning`
- End-to-end GitLab workflow not demonstrated

## 15. Normative terminology

- **Logical application**: functional identity used by a ChangeRequest.
- **Argo CD Application**: runtime resource binding repository, path, and Kubernetes destination.
- **Environment instance**: deployment of a logical application into a specific environment.
- **SCM repository**: repository on which the Change Service performs Git and Merge Request operations.
- **GitOps runtime repository**: repository consumed by Argo CD and Tekton.
- **Standalone Argo CD Application**: Argo CD resource not associated with a logical application through the Environment Catalog.

## 16. Remaining work

- Formalize the SCM binding per logical application
- Model SCM repositories separately from GitOps runtime repositories
- Expose repository correlation status in the UI or evidence
- Decide whether Argo CD should consume GitLab directly, a mirror, or GitHub
- Avoid adding a placeholder GitHub SCM adapter without a real requirement
- Define explicit repository correlation modes
- Review the orphaned resources associated with `demo-app`
