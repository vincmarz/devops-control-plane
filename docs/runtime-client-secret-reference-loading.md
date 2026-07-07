# Runtime client Secret reference loading

This document describes the loading model for multi-cluster runtime client Secret references.

## Scope

Phase 15.8.2 loads only references to Kubernetes Secrets and keys.

It does not read Secret values and does not create runtime clients.

## Configuration file

Default path:

```text
/etc/dcp-runtime-client-secrets/secret-refs.yaml
```

Override environment variable:

```text
DCP_RUNTIME_CLIENT_SECRET_REFS_FILE
```

## Example

```yaml
clusters:
  - clusterName: ocp-staging
    kubernetes:
      namespace: devops-control-plane
      name: ocp-staging-kube
      tokenKey: token
      caKey: ca.crt
    tekton:
      namespace: devops-control-plane
      name: ocp-staging-kube
      tokenKey: token
      caKey: ca.crt
    argocd:
      namespace: devops-control-plane
      name: ocp-staging-argocd
      tokenKey: token
      baseURLKey: baseURL
      caKey: ca.crt
```

## Safety behavior

If the default file is missing, the application uses an empty registry.

This preserves the current safe behavior:

```text
ocp-dev uses the current runtime provider
staging remains disabled
production remains disabled
no Secret values are loaded
```
