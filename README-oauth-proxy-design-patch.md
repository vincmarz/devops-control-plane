# DevOps Control Plane - OAuth Proxy deployment design patch

This archive contains the design-only deliverables for Phase 9.3.8.

It does not modify the existing runtime manifests directly.

## Files included

- docs/adr/ADR-0010-oauth-proxy-deployment-design.md
- docs/runbooks/oauth-proxy.md
- manifests/oauth-proxy/README.md
- manifests/oauth-proxy/serviceaccount.yaml
- manifests/oauth-proxy/deployment-design.yaml
- manifests/oauth-proxy/service-design.yaml
- manifests/oauth-proxy/route-design.yaml

## Apply

From the repository root:

```bash
unzip -o devops-control-plane-oauth-proxy-design.zip

git status --short
find manifests/oauth-proxy -maxdepth 1 -type f | sort
```

## Validate

```bash
git diff --check

# YAML parse validation if yq is available
yq e '.' manifests/oauth-proxy/*.yaml >/dev/null
```

## Important

The files under `manifests/oauth-proxy/` are design/template manifests. Do not apply them directly to the runtime cluster until Phase 9.3.9/9.3.10.
