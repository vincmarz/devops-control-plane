# DevOps Control Plane - OAuth Proxy live-patch-only rollout procedure

This archive contains the deliverables for Phase 9.3.10a.

The procedure is intentionally live-patch-only and avoids applying repository base templates that can overwrite live resources.

## Scope

Prepare a safer runtime retry procedure for OpenShift OAuth Proxy while keeping:

```text
AUTH_ENABLED=false
```

and without applying or rendering:

```text
kind: Secret
ConfigMap dcp-app-trust-bundle
PostgreSQL resources
base Kustomize overlays
repository template secrets
```

## Files included

```text
docs/runbooks/oauth-proxy-live-patch-rollout.md
scripts/oauth-proxy-live-patch/00-precheck.sh
scripts/oauth-proxy-live-patch/01-prepare-prereqs.sh
scripts/oauth-proxy-live-patch/02-apply-live-patches.sh
scripts/oauth-proxy-live-patch/03-verify-rollout.sh
scripts/oauth-proxy-live-patch/99-rollback.sh
```

## Repository apply

```bash
unzip -o devops-control-plane-oauth-proxy-live-patch-only.zip
rm -f devops-control-plane-oauth-proxy-live-patch-only.zip
chmod +x scripts/oauth-proxy-live-patch/*.sh
```

## Repository validation

```bash
git diff --check
find scripts/oauth-proxy-live-patch -maxdepth 1 -type f -name '*.sh' -print -exec bash -n {} \;
```

## Runtime safety

Read the runbook before running the scripts.
Do not run the apply script until precheck and prerequisites are clean.
