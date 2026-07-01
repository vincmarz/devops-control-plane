# AuthN/AuthZ Runbook

## Purpose

This runbook documents the planned authentication and authorization model for the DevOps Control Plane.

The target strategy is header-based AuthN/AuthZ behind an OpenShift OAuth Proxy or equivalent trusted reverse proxy.

## Current state

Current Phase 9.3 baseline found:

- no application-level login;
- no application session cookie;
- no CSRF protection on UI forms;
- no role-based authorization middleware;
- UI displays static `admin` placeholder;
- Route exposes the application directly with edge TLS;
- Deployment has no OAuth proxy sidecar or front proxy.

Current exposure:

```text
Browser -> OpenShift Route -> DevOps Control Plane
```

Target exposure:

```text
Browser -> OpenShift Route -> OAuth Proxy -> DevOps Control Plane
```

## Target identity headers

The application will trust identity headers only when deployed behind the trusted proxy.

Candidate headers:

```text
X-Forwarded-User
X-Forwarded-Groups
X-Auth-Request-User
```

The exact header names must be validated during implementation based on the selected proxy configuration.

## Target roles

```text
viewer
operator
approver
admin
```

### viewer

Can view UI and API read-only resources.

### operator

Can execute technical DevOps actions.

### approver

Can approve/reject/close governed changes and execute approval-sensitive actions.

### admin

Can perform all actions and access future administrative configuration.

## Endpoint classification

### Viewer

Read-only UI:

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
```

Read-only API:

```text
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

### Operator

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

### Approver

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

### Admin

```text
GET /ui/settings
```

Admin can also execute every operator and approver action.

## Implementation checklist

### 1. Add application identity model

Define an internal identity object with:

```text
username
groups
roles
source
```

Possible source values:

```text
anonymous
header
system
```

### 2. Add authorization policy

Define a central policy mapping:

```text
method + route pattern + action -> required role
```

The middleware should reject requests when:

- identity is missing;
- no role is mapped;
- the route is not classified;
- the user does not have the required role.

### 3. Add role mapping from groups

Example configuration shape:

```text
AUTH_ENABLED=true
AUTH_TRUSTED_HEADER_USER=X-Forwarded-User
AUTH_TRUSTED_HEADER_GROUPS=X-Forwarded-Groups
AUTH_GROUP_VIEWER=devops-cp-viewers
AUTH_GROUP_OPERATOR=devops-cp-operators
AUTH_GROUP_APPROVER=devops-cp-approvers
AUTH_GROUP_ADMIN=devops-cp-admins
```

### 4. Replace static UI user

Replace the static UI display `admin` with the authenticated user from the request context.

### 5. Enrich audit actor

Today several actions use static or technical actors. The target is to populate actor fields with the authenticated user.

### 6. Plan CSRF protection

For UI forms using POST, add CSRF protection or verify that the OAuth proxy/session topology mitigates cross-site action risks.

## OpenShift deployment target

Target topology:

```text
Route -> OAuth Proxy Service/Port -> DevOps Control Plane upstream
```

Important boundary:

The backend application must not be reachable directly by end users once trusted headers are accepted.

## Validation tests

### Viewer

Expected allowed:

```text
GET /ui/dashboard
GET /ui/changes
GET /api/v1/changes
```

Expected denied:

```text
POST /api/v1/changes/{id}/validate
POST /api/v1/changes/{id}/approve
```

### Operator

Expected allowed:

```text
POST /api/v1/changes/{id}/validate
POST /api/v1/changes/{id}/check-deployment
POST /api/v1/changes/{id}/collect-evidence
```

Expected denied:

```text
POST /api/v1/changes/{id}/approve
POST /api/v1/changes/{id}/close
```

### Approver

Expected allowed:

```text
POST /api/v1/changes/{id}/approve
POST /api/v1/changes/{id}/close
POST /api/v1/changes/{id}/merge-request
```

### Admin

Expected allowed:

```text
GET /ui/settings
all operator actions
all approver actions
```

## Failure modes

### Missing identity headers

Expected response:

```text
401 Unauthorized
```

### Authenticated but insufficient role

Expected response:

```text
403 Forbidden
```

### Unknown endpoint/action

Expected response:

```text
403 Forbidden
```

The policy must fail closed.

## Phase 9.3 deliverables

- ADR documenting the header-based OAuth Proxy strategy.
- This runbook with endpoint-to-role classification.
- Follow-up implementation patch for middleware, role mapping and tests.
