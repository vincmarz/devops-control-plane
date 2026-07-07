# Runtime client secret/config model

This document defines the initial secret/config model for future multi-cluster runtime clients.

## Security rules

- Secret values must never be stored in Git.
- Secret values must never be printed in logs, documentation or validation output.
- Repository-controlled configuration may only contain Kubernetes Secret references and key names.
- Staging and production remain disabled until explicit enablement phases.

## Initial secret reference shape

```yaml
clusterName: ocp-staging
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

The model is intentionally not wired into runtime clients in phase 15.8.1.
