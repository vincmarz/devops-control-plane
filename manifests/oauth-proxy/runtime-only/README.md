# OAuth Proxy runtime-only overlay

This overlay is for the safe retry of Phase 9.3.10.

It intentionally renders only OAuth Proxy runtime wiring resources and patches:

- OAuth Proxy ServiceAccount
- OAuth Proxy Tekton RoleBinding
- Deployment patch
- Service patch
- Route patch

It does not include or render:

- application Secret templates
- PostgreSQL Secret template
- PostgreSQL Deployment/Service/PVC
- OAuth cookie Secret template
- ConfigMap changes

## Render

```bash
oc kustomize \
  --load-restrictor=LoadRestrictionsNone \
  manifests/oauth-proxy/runtime-only \
  >/tmp/dcp-oauth-proxy-runtime-only-rendered.yaml
```

## Negative check

```bash
grep -n "kind: Secret\|name: devops-control-plane-secrets\|name: postgresql-secret\|kind: PersistentVolumeClaim\|name: postgresql" \
  /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml || echo "OK: no Secret/PostgreSQL resources rendered"
```

## Dry-run

```bash
oc apply --dry-run=client -f /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml
```
