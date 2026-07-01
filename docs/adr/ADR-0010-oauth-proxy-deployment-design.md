# ADR-0010: OpenShift OAuth Proxy Deployment Design

## Status

Accepted for design.

Runtime rollout is deferred to a later phase.

## Context

The DevOps Control Plane already includes configurable application-side AuthN/AuthZ middleware.
The middleware can be enabled through configuration and reads trusted identity headers provided by an upstream authentication component.

The current runtime baseline remains intentionally non-disruptive:

- `AUTH_ENABLED=false`
- the existing Route points to the DevOps Control Plane Service
- the current Route uses edge TLS termination
- the application container listens on port `8080`
- the existing Service targets port `http`

The target architecture is to place OpenShift OAuth Proxy in front of the application and let the backend authorize requests based on trusted headers.

## Decision

Use an OpenShift OAuth Proxy sidecar in the same Pod as the DevOps Control Plane application.

Target flow:

```text
User browser
  -> OpenShift Route
  -> Service port exposed for OAuth Proxy
  -> oauth-proxy sidecar
  -> http://localhost:8080
  -> DevOps Control Plane application
```

The application will not implement an interactive login flow.

Authentication will be delegated to OpenShift OAuth through OAuth Proxy.
Authorization will remain in the Go application middleware.

## Trusted headers

The application middleware will trust headers only when requests are received from the OAuth Proxy boundary.

Preferred headers:

```text
X-Forwarded-User
X-Forwarded-Groups
X-Auth-Request-User
```

The middleware remains fail-closed when enabled:

- missing identity headers: unauthorized
- missing or unmapped groups: forbidden
- insufficient role for an action: forbidden

## Role mapping

OpenShift groups map to application roles as follows:

```text
devops-cp-viewers    -> viewer
devops-cp-operators  -> operator
devops-cp-approvers  -> approver
devops-cp-admins     -> admin
```

The Go backend remains the source of authorization decisions for application endpoints and actions.

## Deployment model

The target Pod will contain two containers:

```text
Pod devops-control-plane
├── devops-control-plane
│   └── listens on 8080
└── oauth-proxy
    └── listens on 8443
    └── proxies to http://localhost:8080
```

The Service should expose the proxy port rather than direct application port when the OAuth Proxy runtime rollout is performed.

The Route should target the proxy Service port.

## Route TLS design

The preferred target Route TLS mode is:

```text
reencrypt
```

Target path:

```text
Browser --TLS--> OpenShift router --TLS--> oauth-proxy --HTTP localhost--> app
```

This is preferred over edge termination because it keeps TLS between the router and the OAuth Proxy container.

## ServiceAccount and OAuth delegation

A dedicated ServiceAccount will be created for OAuth Proxy:

```text
devops-control-plane-oauth-proxy
```

The ServiceAccount requires the `system:auth-delegator` role to validate OpenShift OAuth tokens.

Recommended operational command:

```bash
oc adm policy add-role-to-user \
  system:auth-delegator \
  system:serviceaccount:devops-control-plane:devops-control-plane-oauth-proxy
```

Cluster-scoped permissions should be applied explicitly by an operator/admin and not hidden inside the application Kustomize flow.

## Endpoint policy

Authentication should protect:

```text
/ui
/ui/*
/api/v1/*
```

Health endpoints should remain externally or operationally reachable according to the platform monitoring model:

```text
/healthz
/readyz
```

If health endpoints are routed through OAuth Proxy, they should be explicitly evaluated during implementation to avoid breaking OpenShift probes or external monitoring.

## Rollout strategy

Rollout must be staged.

### Stage 1: repository-only design

Add design documentation and template manifests only.

No runtime change.

### Stage 2: proxy introduced, app middleware still disabled

Deploy OAuth Proxy while keeping:

```text
AUTH_ENABLED=false
```

Validate:

- OpenShift login redirect
- authenticated UI access
- proxy-to-application connectivity
- trusted header injection
- no application behavior regression

### Stage 3: application middleware enabled

Enable:

```text
AUTH_ENABLED=true
```

Validate the role matrix:

- viewer
- operator
- approver
- admin

## Rollback

First rollback lever:

```bash
oc set env deployment/devops-control-plane \
  -n devops-control-plane \
  AUTH_ENABLED=false
```

Broader rollback:

- restore Service target back to application port
- restore Route to the previous Service target port
- remove OAuth Proxy sidecar

## Consequences

Positive:

- login delegated to OpenShift OAuth
- application code stays focused on authorization
- trusted user/group headers become the integration contract
- rollout can be staged and reversible

Trade-offs:

- Service/Route wiring changes are required during runtime rollout
- OAuth Proxy ServiceAccount requires auth delegation
- Route TLS mode and proxy certificate material must be handled carefully
- direct access to application port must be avoided to prevent header spoofing

## Current phase outcome

Phase 9.3.8 defines the target architecture and repository design artifacts.

Runtime implementation is intentionally deferred.
