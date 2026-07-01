# ADR-0002: AuthN/AuthZ Strategy for DevOps Control Plane

## Status

Proposed for Phase 9.3.

## Context

The DevOps Control Plane currently exposes a server-side web UI and REST API through an OpenShift Route.

Current exposure baseline:

- Route `devops-control-plane` is exposed with edge TLS termination.
- The application has no built-in user login, session handling, role model or authorization middleware.
- The UI currently shows a static `admin` placeholder.
- UI and API endpoints include read-only views and mutating technical/lifecycle actions.
- Outbound API clients use tokens for Argo CD, GitLab, Kubernetes and Tekton, but those are integration credentials, not end-user authentication.

This means that any user who can reach the Route can potentially access UI/API data and execute available actions.

The project runs on OpenShift, so the preferred production-oriented model is to delegate authentication to an OpenShift-aware front component and keep authorization decisions explicit in the application.

## Decision

Use header-based AuthN/AuthZ behind an OpenShift OAuth Proxy or equivalent trusted reverse proxy.

Target request flow:

```text
Browser
  -> OpenShift Route
  -> OAuth Proxy
  -> DevOps Control Plane
```

The proxy authenticates the user and forwards trusted identity headers to the application, for example:

```text
X-Forwarded-User
X-Forwarded-Groups
X-Auth-Request-User
```

The DevOps Control Plane will implement an application middleware that:

1. reads trusted identity headers;
2. maps users/groups to application roles;
3. enforces endpoint/action-level authorization;
4. makes the authenticated actor available to audit/event generation;
5. rejects mutating actions when the caller does not have the required role.

## Role model

Initial roles:

```text
viewer
operator
approver
admin
```

### viewer

Can access read-only UI/API data:

- dashboard;
- applications;
- change list/detail;
- evidence;
- audit events.

### operator

Includes `viewer` permissions and can execute technical actions:

- validate;
- check-validation;
- check-deployment;
- collect-evidence;
- create-branch;
- update-files;
- open-merge-request.

### approver

Includes read access and can execute governed lifecycle actions:

- approve;
- reject;
- start-execution;
- complete-execution;
- fail-execution;
- close;
- cancel;
- merge-request, if the process requires approver ownership of merge.

### admin

Can execute all actions and access administrative pages such as settings.

## Endpoint authorization matrix

### Public or probe-only endpoints

```text
GET /healthz    probe/public depending on route exposure
GET /readyz     probe/public depending on route exposure
```

### Viewer endpoints

```text
GET /
GET /ui
GET /ui/dashboard
GET /ui/changes
GET /ui/changes/{id}
GET /ui/changes/{id}/evidence
GET /ui/changes/{id}/events
GET /ui/applications
GET /ui/applications/{name}
GET /ui/changes-api

GET /api/v1/applications
GET /api/v1/applications/{name}
GET /api/v1/applications/{name}/resources
GET /api/v1/applications/{name}/history
GET /api/v1/applications/{name}/runtime
GET /api/v1/changes
GET /api/v1/changes/{id}
GET /api/v1/changes/{id}/events
GET /api/v1/changes/{id}/evidence
GET /api/v1/changes/{id}/evidence/{type}
```

### Admin endpoints

```text
GET /ui/settings
```

### Operator endpoints

```text
POST /api/v1/changes
POST /api/v1/changes/{id}/validate
POST /api/v1/changes/{id}/check-validation
POST /api/v1/changes/{id}/check-deployment
POST /api/v1/changes/{id}/collect-evidence
POST /api/v1/changes/{id}/create-branch
POST /api/v1/changes/{id}/update-files
POST /api/v1/changes/{id}/open-merge-request

POST /ui/changes/{id}/actions/validate
POST /ui/changes/{id}/actions/check-validation
POST /ui/changes/{id}/actions/check-deployment
POST /ui/changes/{id}/actions/collect-evidence
POST /ui/changes/{id}/actions/create-branch
POST /ui/changes/{id}/actions/update-files
POST /ui/changes/{id}/actions/open-merge-request
```

### Approver or admin endpoints

```text
POST /api/v1/changes/{id}/submit
POST /api/v1/changes/{id}/approve
POST /api/v1/changes/{id}/reject
POST /api/v1/changes/{id}/start-execution
POST /api/v1/changes/{id}/complete-execution
POST /api/v1/changes/{id}/fail-execution
POST /api/v1/changes/{id}/close
POST /api/v1/changes/{id}/cancel
POST /api/v1/changes/{id}/merge-request

POST /ui/changes/{id}/actions/merge-request
```

## Security constraints

Header-based authentication is safe only if the application receives requests exclusively from the trusted proxy.

Required controls:

- the backend Route must not be directly exposed to end users when the proxy is introduced;
- externally supplied identity headers must not be accepted from untrusted clients;
- the proxy must overwrite, not append, identity headers;
- application authorization must fail closed when identity headers are missing;
- mutating UI actions should also include CSRF protection or be constrained through proxy/session controls.

## Consequences

Positive:

- integrates naturally with OpenShift;
- avoids implementing password handling in the Go application;
- allows group-based authorization;
- supports audit actor enrichment;
- keeps the application role model explicit and testable.

Trade-offs:

- requires additional OpenShift deployment/proxy configuration;
- requires careful header trust boundaries;
- requires endpoint authorization middleware and tests;
- requires migration from static UI user placeholder to real identity.

## Implementation phases

### Phase 9.3.1

Current exposure baseline and endpoint inventory.

### Phase 9.3.2

Role model and endpoint/action classification.

### Phase 9.3.3

Application middleware design:

- identity object;
- role mapping;
- authorization policy;
- audit actor propagation.

### Phase 9.3.4

OpenShift OAuth Proxy deployment plan:

- proxy container or sidecar;
- service/route topology;
- required secrets;
- trusted header handling.

### Phase 9.3.5

First implementation:

- middleware for header identity;
- fail-closed mode behind feature flag;
- role mapping through configuration;
- tests for allowed/denied actions.

## Current decision

Proceed with header-based AuthN/AuthZ behind an OpenShift OAuth Proxy as the target strategy for DevOps Control Plane production readiness.
