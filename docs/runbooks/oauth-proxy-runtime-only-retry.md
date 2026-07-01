# Runbook: OAuth Proxy runtime-only retry

## Purpose

This runbook documents the safer retry for Phase 9.3.10 after the previous runtime attempt confirmed that OAuth Proxy started correctly but the rendered overlay also applied Secret templates and overwrote live Secret values.

This retry uses a minimal runtime-only Kustomize overlay that renders only:

- `ServiceAccount` for OAuth Proxy
- namespaced RoleBinding for preserving existing Tekton runtime permissions
- patched `Deployment`
- patched `Service`
- patched `Route`

It intentionally excludes:

- `secret-template.yaml`
- `postgresql-secret-template.yaml`
- `oauth-proxy-cookie-secret-template.yaml`
- PostgreSQL Deployment/Service/PVC
- ConfigMap changes

## Safety baseline

Before retry, confirm baseline:

```bash
oc get pods -n devops-control-plane -l app=devops-control-plane -o wide

oc get configmap devops-control-plane-config \
  -n devops-control-plane \
  -o jsonpath='AUTH_ENABLED={.data.AUTH_ENABLED}{"\n"}'

oc get service devops-control-plane \
  -n devops-control-plane \
  -o jsonpath='{range .spec.ports[*]}{.name}{" "}{.port}{" -> "}{.targetPort}{"\n"}{end}'

oc get route devops-control-plane \
  -n devops-control-plane \
  -o jsonpath='targetPort={.spec.port.targetPort}{" termination="}{.spec.tls.termination}{"\n"}'
```

Expected:

```text
AUTH_ENABLED=false
http 8080 -> http
targetPort=http termination=edge
```

## Verify live Secrets without printing values

```bash
oc get secret devops-control-plane-secrets \
  -n devops-control-plane \
  -o go-template='{{range $k,$v := .data}}{{printf "%s base64_len=%d\n" $k (len $v)}}{{end}}'

oc get secret postgresql-secret \
  -n devops-control-plane \
  -o go-template='{{range $k,$v := .data}}{{printf "%s base64_len=%d\n" $k (len $v)}}{{end}}'
```

Do not proceed if application token/password lengths look like placeholders.

## Ensure OAuth Proxy prerequisites

### Auth delegator grant

```bash
oc adm policy add-role-to-user \
  system:auth-delegator \
  system:serviceaccount:devops-control-plane:devops-control-plane-oauth-proxy
```

Verify:

```bash
oc auth can-i create tokenreviews.authentication.k8s.io \
  --as=system:serviceaccount:devops-control-plane:devops-control-plane-oauth-proxy

oc auth can-i create subjectaccessreviews.authorization.k8s.io \
  --as=system:serviceaccount:devops-control-plane:devops-control-plane-oauth-proxy
```

Expected:

```text
yes
yes
```

### Cookie secret

Verify existing cookie secret key:

```bash
oc get secret devops-control-plane-oauth-proxy-cookie \
  -n devops-control-plane \
  -o go-template='{{range $k,$v := .data}}{{printf "%s\n" $k}}{{end}}'
```

Expected:

```text
cookie-secret
```

If the cookie secret does not exist, create it without printing the value:

```bash
set +x
COOKIE_SECRET="$(openssl rand -base64 48 | tr -dc 'A-Za-z0-9' | head -c 32)"

oc create secret generic devops-control-plane-oauth-proxy-cookie \
  -n devops-control-plane \
  --from-literal=cookie-secret="${COOKIE_SECRET}" \
  --dry-run=client -o yaml \
  | oc apply -f -

unset COOKIE_SECRET
```

## Render runtime-only overlay

```bash
oc kustomize \
  --load-restrictor=LoadRestrictionsNone \
  manifests/oauth-proxy/runtime-only \
  >/tmp/dcp-oauth-proxy-runtime-only-rendered.yaml
```

## Critical negative checks

The render must not include any Secret template or PostgreSQL resource:

```bash
grep -n "kind: Secret\|name: devops-control-plane-secrets\|name: postgresql-secret\|kind: PersistentVolumeClaim\|name: postgresql" \
  /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml || echo "OK: no Secret/PostgreSQL resources rendered"
```

Expected:

```text
OK: no Secret/PostgreSQL resources rendered
```

## Critical positive checks

```bash
grep -n "serviceAccountName: devops-control-plane-oauth-proxy" /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml

grep -n "name: oauth-proxy" /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml

grep -n "targetPort: oauth-proxy" /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml

grep -n "termination: reencrypt" /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml

grep -n "serviceaccounts.openshift.io/oauth-redirectreference.primary" /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml
```

## Dry-run

```bash
oc apply --dry-run=client -f /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml
```

Expected resources only:

```text
serviceaccount/devops-control-plane-oauth-proxy
rolebinding/devops-control-plane-oauth-proxy-tekton-runtime
service/devops-control-plane
deployment.apps/devops-control-plane
route.route.openshift.io/devops-control-plane
```

No `secret/*`, no `deployment.apps/postgresql`, no `service/postgresql`, no PVC.

## Real apply

Only after checks are clean:

```bash
oc apply -f /tmp/dcp-oauth-proxy-runtime-only-rendered.yaml
```

Monitor:

```bash
oc rollout status deployment/devops-control-plane \
  -n devops-control-plane \
  --timeout=180s
```

## Runtime verification

```bash
oc get pods -n devops-control-plane -l app=devops-control-plane -o wide

POD="$(oc get pod -n devops-control-plane -l app=devops-control-plane -o jsonpath='{.items[0].metadata.name}')"

oc get pod "$POD" -n devops-control-plane \
  -o jsonpath='{range .spec.containers[*]}{.name}{"\n"}{end}'

oc get configmap devops-control-plane-config \
  -n devops-control-plane \
  -o jsonpath='AUTH_ENABLED={.data.AUTH_ENABLED}{"\n"}'

oc get service devops-control-plane \
  -n devops-control-plane \
  -o jsonpath='{range .spec.ports[*]}{.name}{" "}{.port}{" -> "}{.targetPort}{"\n"}{end}'

oc get route devops-control-plane \
  -n devops-control-plane \
  -o jsonpath='targetPort={.spec.port.targetPort}{" termination="}{.spec.tls.termination}{" insecure="}{.spec.tls.insecureEdgeTerminationPolicy}{"\n"}'
```

Expected:

```text
oauth-proxy
devops-control-plane
AUTH_ENABLED=false
https 8443 -> oauth-proxy
targetPort=https termination=reencrypt insecure=Redirect
```

## Rollback

If proxy wiring fails, restore baseline using direct patches:

```bash
oc patch service devops-control-plane \
  -n devops-control-plane \
  --type=json \
  -p='[
    {"op":"remove","path":"/metadata/annotations/service.beta.openshift.io~1serving-cert-secret-name"},
    {"op":"replace","path":"/spec/ports","value":[{"name":"http","port":8080,"protocol":"TCP","targetPort":"http"}]}
  ]' || \
oc patch service devops-control-plane \
  -n devops-control-plane \
  --type=json \
  -p='[
    {"op":"replace","path":"/spec/ports","value":[{"name":"http","port":8080,"protocol":"TCP","targetPort":"http"}]}
  ]'

oc patch route devops-control-plane \
  -n devops-control-plane \
  --type=merge \
  -p='{"spec":{"port":{"targetPort":"http"},"tls":{"termination":"edge","insecureEdgeTerminationPolicy":null}}}'

oc rollout undo deployment/devops-control-plane \
  -n devops-control-plane
```

Then validate `/readyz` and `/api/v1/changes`.
