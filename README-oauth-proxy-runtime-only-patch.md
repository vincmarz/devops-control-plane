# DevOps Control Plane - OAuth Proxy runtime-only retry patch set

This archive contains a safer runtime-only patch set for retrying Phase 9.3.10.

It intentionally does not render or apply any application Secret template or PostgreSQL resource.

## Goal

Retry OAuth Proxy staged runtime rollout while keeping:

```text
AUTH_ENABLED=false
```

and without touching:

```text
devops-control-plane-secrets
postgresql-secret
PostgreSQL Deployment/Service/PVC
```

## Files included

```text
docs/runbooks/oauth-proxy-runtime-only-retry.md
manifests/oauth-proxy/runtime-only/README.md
manifests/oauth-proxy/runtime-only/kustomization.yaml
manifests/oauth-proxy/runtime-only/oauth-proxy-serviceaccount.yaml
manifests/oauth-proxy/runtime-only/oauth-proxy-tekton-rolebinding.yaml
manifests/oauth-proxy/runtime-only/patch-deployment-oauth-proxy.yaml
manifests/oauth-proxy/runtime-only/patch-service-oauth-proxy.yaml
manifests/oauth-proxy/runtime-only/patch-route-oauth-proxy.yaml
```

## Apply to repository

```bash
unzip -o devops-control-plane-oauth-proxy-runtime-only.zip
rm -f devops-control-plane-oauth-proxy-runtime-only.zip
```

## Repository-only validation

```bash
git diff --check

oc kustomize \
  --load-restrictor=LoadRestrictionsNone \
  manifests/oauth-proxy/runtime-only \
  >/tmp/dcp-oauth-proxy-runtime-only-rendered.yaml

# Must not render any Secret or PostgreSQL resource
grep -n "kind: Secret\|name: devops-control-plane-secrets\|name: postgresql-secret\|kind: PersistentVolumeClaim\|name: postgresql" \
  /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml || echo "OK: no Secret/PostgreSQL resources rendered"

oc apply --dry-run=client -f /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml
```

## Runtime rule

Before real apply, verify existing live Secrets and cookie secret. Do not print secret values.
