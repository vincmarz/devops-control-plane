# OAuth Proxy implementation overlay

This directory contains the Phase 9.3.9 repository-only OAuth Proxy implementation overlay.

It is intended to be rendered and validated before any runtime rollout.

## Render

```bash
oc kustomize --load-restrictor=LoadRestrictionsNone manifests/oauth-proxy/overlay >/tmp/dcp-oauth-proxy-rendered.yaml
```

## Client-side dry-run

```bash
oc apply --dry-run=client -f /tmp/dcp-oauth-proxy-rendered.yaml
```

## Runtime rollout rule

First rollout must keep application-side middleware disabled:

```text
AUTH_ENABLED=false
```

Enable `AUTH_ENABLED=true` only after OAuth Proxy login and header propagation are validated.

## Cluster-scoped auth delegation

The `system:auth-delegator` grant is provided separately under `manifests/oauth-proxy/cluster/` and is not included in this overlay.
