# Runbook: OpenShift OAuth Proxy for DevOps Control Plane

## Purpose

This runbook documents the target deployment model for protecting DevOps Control Plane UI and API endpoints with OpenShift OAuth Proxy.

This runbook is design-oriented for Phase 9.3.8. Runtime application of the manifests must be performed in later phases only.

## Current baseline

Expected current runtime baseline:

```text
AUTH_ENABLED=false
Route -> Service -> application container port 8080
middleware available but disabled
```

Check:

```bash
oc get deployment devops-control-plane -n devops-control-plane
oc get route devops-control-plane -n devops-control-plane -o yaml
oc get service devops-control-plane -n devops-control-plane -o yaml
oc get configmap devops-control-plane-config -n devops-control-plane -o yaml | grep AUTH_ENABLED -n
```

## Target architecture

```text
User Browser
  -> OpenShift Route
  -> Service
  -> OAuth Proxy sidecar on 8443
  -> DevOps Control Plane app on localhost:8080
```

## Target ServiceAccount

```bash
oc create serviceaccount devops-control-plane-oauth-proxy \
  -n devops-control-plane
```

Grant auth delegation:

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

## Target OAuth Proxy arguments

Target logical arguments:

```text
--provider=openshift
--https-address=:8443
--upstream=http://localhost:8080
--set-xauthrequest=true
--pass-user-bearer-token=false
--pass-access-token=false
--pass-basic-auth=false
```

Implementation may require additional arguments for certificates, cookie secret, OpenShift service account integration, redirect URL and skip-auth routes.
Those details should be validated in Phase 9.3.9/9.3.10 against the exact image version available in the cluster.

## Application configuration

Initial proxy rollout should keep:

```text
AUTH_ENABLED=false
```

Only after proxy login/header behavior is validated:

```text
AUTH_ENABLED=true
```

Expected auth header configuration:

```text
AUTH_USER_HEADERS=X-Forwarded-User,X-Auth-Request-User
AUTH_GROUP_HEADERS=X-Forwarded-Groups
AUTH_VIEWER_GROUPS=devops-cp-viewers
AUTH_OPERATOR_GROUPS=devops-cp-operators
AUTH_APPROVER_GROUPS=devops-cp-approvers
AUTH_ADMIN_GROUPS=devops-cp-admins
```

## OpenShift groups

Create groups if not already present:

```bash
oc adm groups new devops-cp-viewers
oc adm groups new devops-cp-operators
oc adm groups new devops-cp-approvers
oc adm groups new devops-cp-admins
```

Add users:

```bash
oc adm groups add-users devops-cp-viewers USERNAME
oc adm groups add-users devops-cp-operators USERNAME
oc adm groups add-users devops-cp-approvers USERNAME
oc adm groups add-users devops-cp-admins USERNAME
```

Check user groups:

```bash
oc get groups | grep devops-cp
oc describe group devops-cp-viewers
oc describe group devops-cp-operators
oc describe group devops-cp-approvers
oc describe group devops-cp-admins
```

## Runtime validation plan

### Stage A: proxy present, middleware disabled

```text
AUTH_ENABLED=false
```

Validate:

```bash
APP_ROUTE="$(oc get route devops-control-plane -n devops-control-plane -o jsonpath='{.spec.host}')"

curl -k -I "https://${APP_ROUTE}/ui/dashboard"
curl -k -I "https://${APP_ROUTE}/readyz"
```

Browser validation:

- access the Route URL
- confirm redirect to OpenShift OAuth
- login with a test user
- confirm redirect back to the DevOps Control Plane UI

### Stage B: middleware enabled

```bash
oc set env deployment/devops-control-plane \
  -n devops-control-plane \
  AUTH_ENABLED=true
```

Validate role behavior with real users assigned to OpenShift groups.

## Expected role behavior

Viewer:

```text
read UI/API: allowed
technical actions: forbidden
approval actions: forbidden
```

Operator:

```text
read UI/API: allowed
technical runtime actions: allowed
approval actions: forbidden
```

Approver:

```text
read UI/API: allowed
approval actions: allowed when domain state permits
operator-only actions: forbidden unless also member of operator/admin group
```

Admin:

```text
full access according to application middleware policy
```

## Rollback

Fast rollback:

```bash
oc set env deployment/devops-control-plane \
  -n devops-control-plane \
  AUTH_ENABLED=false
```

If the proxy wiring itself causes issues:

- restore the previous Service target port
- restore previous Route target port/TLS mode
- remove OAuth Proxy sidecar
- keep the application ConfigMap with `AUTH_ENABLED=false`

## Safety notes

Do not print OAuth tokens, session secrets, application tokens or Kubernetes tokens in logs or chat.

Do not trust user-provided `X-Forwarded-User` or `X-Forwarded-Groups` headers unless traffic is guaranteed to arrive from the OAuth Proxy trusted boundary.

Do not expose the application container directly once `AUTH_ENABLED=true` is used with header-based authentication.
