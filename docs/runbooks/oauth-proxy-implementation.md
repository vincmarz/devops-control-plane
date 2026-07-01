# Runbook: OAuth Proxy repository-only implementation

## Purpose

This runbook documents the repository-only implementation introduced in Phase 9.3.9.

The objective is to provide a reviewable and dry-run-validatable Kustomize overlay for introducing OpenShift OAuth Proxy in front of the DevOps Control Plane, without changing the current runtime during this phase.

## Runtime safety principle

Phase 9.3.9 is repository-only.

Do not run `oc apply` against the live namespace as part of this phase unless explicitly entering the later runtime rollout phase.

The current baseline remains:

```text
AUTH_ENABLED=false
Route edge -> Service http -> application port 8080
```

## New repository layout

```text
manifests/oauth-proxy/overlay/
├── README.md
├── kustomization.yaml
├── oauth-proxy-serviceaccount.yaml
├── oauth-proxy-cookie-secret-template.yaml
├── oauth-proxy-tekton-rolebinding.yaml
├── patch-deployment-oauth-proxy.yaml
├── patch-service-oauth-proxy.yaml
└── patch-route-oauth-proxy.yaml

manifests/oauth-proxy/cluster/
├── README.md
└── oauth-proxy-auth-delegator-clusterrolebinding-template.yaml
```

## Design choices

### Dedicated Pod ServiceAccount

The OAuth Proxy sidecar and the application container share the same Pod ServiceAccount because Kubernetes assigns a single ServiceAccount at Pod level.

The overlay therefore switches the workload Pod ServiceAccount to:

```text
devops-control-plane-oauth-proxy
```

To preserve existing Tekton runtime behavior, the overlay adds a namespaced RoleBinding that binds the existing Role:

```text
devops-control-plane-tekton-runtime
```

to the new ServiceAccount.

### OAuth redirect reference

The OAuth Proxy ServiceAccount includes the OpenShift OAuth redirect reference annotation pointing to the `devops-control-plane` Route.

This allows the ServiceAccount to act as an OpenShift OAuth client for the application Route.

### Service serving certificate

The overlay patches the Service with OpenShift service serving certificate annotation:

```text
service.beta.openshift.io/serving-cert-secret-name: devops-control-plane-oauth-proxy-tls
```

The generated Secret is mounted into the OAuth Proxy container and used through:

```text
--tls-cert=/etc/tls/private/tls.crt
--tls-key=/etc/tls/private/tls.key
```

### Cookie secret

The overlay includes a cookie Secret template with a placeholder value.

Before runtime use, generate a strong random secret and replace the placeholder without committing real secret values if the repository is public or shared.

Example runtime-safe creation command:

```bash
COOKIE_SECRET="$(openssl rand -base64 32)"

oc create secret generic devops-control-plane-oauth-proxy-cookie \
  -n devops-control-plane \
  --from-literal=cookie-secret="${COOKIE_SECRET}" \
  --dry-run=client -o yaml
```

Do not print or paste real secret values in tickets, chat, logs, or documentation.

### Middleware remains disabled for first rollout

The first runtime rollout must keep:

```text
AUTH_ENABLED=false
```

Only after OAuth login, proxy forwarding and header propagation are validated should the middleware be enabled:

```text
AUTH_ENABLED=true
```

## Repository-only validation

Render the overlay:

```bash
oc kustomize --load-restrictor=LoadRestrictionsNone manifests/oauth-proxy/overlay >/tmp/dcp-oauth-proxy-rendered.yaml
```

Client-side dry-run:

```bash
oc apply --dry-run=client -f /tmp/dcp-oauth-proxy-rendered.yaml
```

Inspect critical rendered values:

```bash
grep -n "serviceAccountName: devops-control-plane-oauth-proxy" /tmp/dcp-oauth-proxy-rendered.yaml

grep -n "name: oauth-proxy" /tmp/dcp-oauth-proxy-rendered.yaml

grep -n "targetPort: oauth-proxy" /tmp/dcp-oauth-proxy-rendered.yaml

grep -n "termination: reencrypt" /tmp/dcp-oauth-proxy-rendered.yaml
```

## Cluster-scoped auth delegation

The OAuth Proxy ServiceAccount requires auth delegation.

Preferred operational command:

```bash
oc adm policy add-role-to-user \
  system:auth-delegator \
  system:serviceaccount:devops-control-plane:devops-control-plane-oauth-proxy
```

A ClusterRoleBinding template is provided separately under:

```text
manifests/oauth-proxy/cluster/
```

It is intentionally not included in the namespaced overlay Kustomization.

## Runtime rollout preview

Later runtime rollout should be staged:

1. Apply ServiceAccount, cookie Secret, RoleBinding, Deployment/Service/Route changes.
2. Keep `AUTH_ENABLED=false`.
3. Validate OAuth login redirect and UI access.
4. Validate proxy to application connectivity.
5. Validate injected identity/group headers if a temporary diagnostic endpoint or logs are available.
6. Enable `AUTH_ENABLED=true` only after successful proxy validation.
7. Validate viewer/operator/approver/admin behavior.

## Rollback

Fast rollback for authorization issues:

```bash
oc set env deployment/devops-control-plane \
  -n devops-control-plane \
  AUTH_ENABLED=false
```

Rollback for proxy wiring issues:

```bash
oc apply -k manifests
```

or explicitly restore the pre-proxy Deployment, Service and Route from the base manifests.

## Security notes

- Do not commit real cookie secrets.
- Do not print OAuth tokens, application tokens or Kubernetes tokens.
- Do not expose application port `8080` directly once header-based auth is enabled.
- Treat `X-Forwarded-User` and `X-Forwarded-Groups` as trusted only when traffic is guaranteed to come from the OAuth Proxy path.
