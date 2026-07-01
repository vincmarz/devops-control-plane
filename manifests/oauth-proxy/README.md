# OAuth Proxy design manifests

This directory contains design/template manifests for introducing OpenShift OAuth Proxy in front of DevOps Control Plane.

These files are part of Phase 9.3.8 and are not intended to be applied directly without review.

## Files

- `serviceaccount.yaml`: dedicated ServiceAccount for OAuth Proxy
- `deployment-design.yaml`: target Deployment shape with an OAuth Proxy sidecar
- `service-design.yaml`: target Service shape exposing the proxy port
- `route-design.yaml`: target Route shape using reencrypt TLS and targeting the proxy port

## Current repository baseline

At the time of this design, the current repository contains:

- `manifests/deployment.yaml`
- `manifests/service.yaml`
- `manifests/route.yaml`

The current Route targets Service port `http` and uses edge TLS termination.
The current application container exposes port `8080`.

## Required runtime grant

The OAuth Proxy ServiceAccount requires auth delegation:

```bash
oc adm policy add-role-to-user \
  system:auth-delegator \
  system:serviceaccount:devops-control-plane:devops-control-plane-oauth-proxy
```

## Rollout principle

First introduce OAuth Proxy with application middleware disabled:

```text
AUTH_ENABLED=false
```

Only after login and header propagation are validated, enable:

```text
AUTH_ENABLED=true
```
