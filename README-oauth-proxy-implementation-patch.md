# DevOps Control Plane - OAuth Proxy repository-only implementation patch

This archive contains the repository-only implementation deliverables for Phase 9.3.9.

It adds a deployable Kustomize overlay under:

```text
manifests/oauth-proxy/overlay/
```

The overlay is intended for review and dry-run validation first. Do not apply it to the runtime cluster until the staged rollout phase.

## Files included

- docs/runbooks/oauth-proxy-implementation.md
- manifests/oauth-proxy/overlay/README.md
- manifests/oauth-proxy/overlay/kustomization.yaml
- manifests/oauth-proxy/overlay/oauth-proxy-serviceaccount.yaml
- manifests/oauth-proxy/overlay/oauth-proxy-cookie-secret-template.yaml
- manifests/oauth-proxy/overlay/oauth-proxy-tekton-rolebinding.yaml
- manifests/oauth-proxy/overlay/patch-deployment-oauth-proxy.yaml
- manifests/oauth-proxy/overlay/patch-service-oauth-proxy.yaml
- manifests/oauth-proxy/overlay/patch-route-oauth-proxy.yaml
- manifests/oauth-proxy/cluster/README.md
- manifests/oauth-proxy/cluster/oauth-proxy-auth-delegator-clusterrolebinding-template.yaml

## Apply to repository

```bash
unzip -o devops-control-plane-oauth-proxy-implementation.zip
```

## Validate repository-only

```bash
git diff --check

# inspect generated overlay without applying it
oc kustomize --load-restrictor=LoadRestrictionsNone manifests/oauth-proxy/overlay >/tmp/dcp-oauth-proxy-rendered.yaml

# client-side validation only
oc apply --dry-run=client -f /tmp/dcp-oauth-proxy-rendered.yaml
```

## Important

This phase does not change the current runtime by itself. Runtime changes happen only if the overlay is applied later.

The overlay keeps `AUTH_ENABLED=false` during the first proxy rollout stage. Application-side authorization must be enabled only after OAuth Proxy login and header propagation are validated.
