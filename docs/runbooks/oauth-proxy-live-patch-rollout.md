# Runbook: OAuth Proxy live-patch-only rollout

## Purpose

This runbook documents Phase 9.3.10a: a safer live-patch-only rollout procedure for retrying OpenShift OAuth Proxy in front of the DevOps Control Plane.

The previous runtime attempts showed that OAuth Proxy itself can start correctly, but broad Kustomize renders can unintentionally apply templates and overwrite live resources.

This procedure avoids that class of incident by using only targeted live operations.

## Non-negotiable safety rules

Do not apply repository base manifests or overlays during this procedure.

Do not apply or render resources of these types/names:

```text
kind: Secret
ConfigMap/dcp-app-trust-bundle
Deployment/postgresql
Service/postgresql
PersistentVolumeClaim/postgresql-data
```

Do not touch:

```text
devops-control-plane-secrets
postgresql-secret
dcp-app-trust-bundle
```

The rollout must keep:

```text
AUTH_ENABLED=false
```

Application-side middleware must not be enabled until OAuth Proxy login and forwarding are validated.

## Target staged state

```text
Route targetPort=https termination=reencrypt
  -> Service https 8443 -> oauth-proxy
  -> oauth-proxy sidecar on 8443
  -> http://localhost:8080
  -> devops-control-plane application
```

The application container continues to read:

```text
envFrom ConfigMap/devops-control-plane-config
envFrom Secret/devops-control-plane-secrets
volume ConfigMap/dcp-app-trust-bundle mounted at /etc/dcp-trust
```

## Procedure overview

1. Precheck baseline and live resources.
2. Verify `dcp-app-trust-bundle` contains a valid PEM certificate.
3. Verify application and PostgreSQL Secrets are not placeholders.
4. Ensure OAuth Proxy ServiceAccount and auth delegation.
5. Ensure OAuth Proxy cookie Secret exists without printing its value.
6. Add serving-cert annotation to the existing Service without changing ports yet.
7. Wait for the serving certificate Secret.
8. Patch Deployment with sidecar and new ServiceAccount.
9. Wait for the new Pod to become ready with both containers.
10. Switch Service and Route to OAuth Proxy.
11. Validate `AUTH_ENABLED=false`, Service/Route, and OAuth redirect behavior.
12. Roll back immediately if the new Pod fails or app logs show startup errors.

## Manual rollback summary

If anything fails:

```bash
scripts/oauth-proxy-live-patch/99-rollback.sh
```

The rollback script restores:

```text
Service http 8080 -> http
Route targetPort=http termination=edge
Deployment serviceAccountName=devops-control-plane
Deployment without oauth-proxy sidecar
Deployment without oauth-proxy-tls volume
```

It does not modify application Secrets, PostgreSQL, or `dcp-app-trust-bundle`.

## Important lesson learned

Do not use Kustomize overlays that include base template resources for runtime experiments.
Use only narrow live patches or standalone manifests that include only the exact objects that must change.
