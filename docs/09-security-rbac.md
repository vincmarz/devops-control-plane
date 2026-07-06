# DevOps Control Plane — Security and RBAC

## Document metadata

- **Project:** DevOps Control Plane
- **Document:** 09 — Security and RBAC
- **Version:** 0.2
- **Date:** 2026-07-06
- **Owner:** Vincenzo Marzario
- **Repository:** `https://github.com/vincmarz/devops-control-plane`
- **Previous documents:**
  - `docs/00-vision.md`
  - `docs/01-scope-mvp.md`
  - `docs/02-personas-use-cases.md`
  - `docs/03-functional-requirements.md`
  - `docs/04-non-functional-requirements.md`
  - `docs/05-architecture.md`
  - `docs/06-argocd-integration.md`
  - `docs/07-gitlab-integration.md`
  - `docs/08-tekton-integration.md`
- **Status:** Rewritten in English and refreshed while preserving the original least-privilege and sanitized-evidence security intent
- **Language:** English
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This document defines the security, credential-handling and RBAC model of the DevOps Control Plane.

The original document established the right security spirit from the beginning:

```text
least privilege
configuration separated from secrets
no secrets in Git
no secrets in logs
no secrets in evidence
sanitized evidence
controlled blast radius
lab versus stable production-oriented usage clearly separated
```

This refreshed version preserves that original intent and updates the document to match the current implementation baseline, including:

- OpenShift OAuth Proxy;
- trusted-header AuthN/AuthZ;
- OpenShift group lookup;
- roles `viewer`, `operator`, `approver` and `admin`;
- RBAC least-privilege baseline;
- ClusterRoleBinding for OpenShift group lookup;
- ServiceAccount token fallback for Kubernetes/OpenShift API access;
- removal of static `KUBERNETES_TOKEN` from the application Secret;
- dedicated Argo CD account and token;
- TLS strict mode;
- application-dedicated trust bundle;
- GitLab TLS strict mode;
- PostgreSQL NetworkPolicy;
- secrets rotation runbook;
- operability smoke-test security checks;
- future environment-aware security model for `dev`, `staging` and `production`.

The goal is to ensure that the Control Plane makes GitOps changes safer and more auditable without becoming a new uncontrolled privileged entry point.

---

## 2. Security principles

## 2.1 Least privilege

Every identity used by the DevOps Control Plane must have only the permissions required for its specific function.

Core rule:

```text
No token and no ServiceAccount should be cluster-admin by default.
```

Elevated permissions may be acceptable only during controlled laboratory work and must not become the stable runtime baseline.

## 2.2 Configuration and secrets must be separated

Non-sensitive configuration belongs in ConfigMaps or plain environment configuration.

Examples:

- Argo CD base URL;
- GitLab base URL;
- Tekton namespace;
- Tekton pipeline name;
- Kubernetes target namespace;
- timeout values;
- log level;
- AuthN/AuthZ feature flags;
- supported environment names.

Sensitive values belong in Secrets.

Examples:

- GitLab token;
- Argo CD token;
- PostgreSQL connection string if it contains credentials;
- PostgreSQL password;
- private keys;
- Git credentials used by pipelines.

## 2.3 Git is untrusted for real secrets

The repository must never contain real credentials.

Forbidden values include:

- tokens;
- passwords;
- real kubeconfigs;
- private certificates;
- private SSH keys;
- real Docker auth configuration;
- real Kubernetes Secrets;
- `.env` files containing sensitive values.

Only placeholders and templates are allowed.

Example allowed template:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: devops-control-plane-secrets
stringData:
  GITLAB_TOKEN: "<replace-me>"
  ARGOCD_AUTH_TOKEN: "<replace-me>"
```

## 2.4 Evidence must be useful but sanitized

Evidence is one of the main values of the DevOps Control Plane, but evidence must not become a data leakage channel.

Core rule:

```text
Evidence first, but sanitized evidence.
```

## 2.5 Runtime changes must not bypass GitOps

Security also includes preventing uncontrolled runtime drift.

The Control Plane must not use direct runtime mutations as the permanent desired-state mechanism.

---

## 3. Trust boundaries

## 3.1 Current trust boundaries

Current request and integration path:

```text
Browser / API Client
  -> OpenShift Route
  -> OpenShift OAuth Proxy
  -> DevOps Control Plane Go backend
  -> PostgreSQL
  -> GitLab API
  -> Argo CD API
  -> Kubernetes/OpenShift API
  -> Tekton CRDs
```

Each trust boundary must enforce:

- authentication where required;
- authorization where required;
- timeout handling;
- safe logging;
- error normalization;
- masking or exclusion of sensitive values.

## 3.2 HTTP ingress boundary

The current stable baseline uses OpenShift OAuth Proxy in front of the Go backend.

Rules:

- anonymous API and UI access through the Route must be denied;
- health endpoints must remain reachable without authentication;
- trusted identity headers are accepted only from the OAuth Proxy boundary;
- backend AuthZ remains authoritative for actions and endpoints.

Current behavior:

```text
GET /readyz through Route -> HTTP 200 when ready
GET /livez through Route  -> HTTP 200 when live
GET /api/v1/changes anonymous through Route -> HTTP 403
GET /ui/dashboard anonymous through Route -> HTTP 403
```

## 3.3 External API boundary

Every external adapter must:

- read tokens from Secrets;
- avoid printing tokens;
- use explicit timeouts;
- normalize errors;
- avoid returning raw sensitive payloads to clients;
- avoid persisting credentials in PostgreSQL.

---

## 4. Authentication and authorization model

## 4.1 OpenShift OAuth Proxy

OpenShift OAuth Proxy authenticates users at the Route boundary and forwards trusted identity headers to the backend.

Expected trusted headers include:

```text
X-Forwarded-User
X-Forwarded-Groups
```

The exact header names are configurable.

## 4.2 Backend AuthN/AuthZ

The backend reads trusted headers and enforces authorization.

Current roles:

```text
viewer
operator
approver
admin
```

Rules:

- unknown users must not receive implicit privileges;
- users without mapped groups must be rejected;
- unclassified endpoints must fail closed;
- health endpoints are public by explicit policy;
- API and UI business endpoints require authorization.

## 4.3 OpenShift group lookup

The backend can resolve group membership through the OpenShift API when group headers are not sufficient or not present.

This requires a dedicated cluster-scoped permission to read OpenShift groups.

Current baseline includes a dedicated ClusterRoleBinding for group lookup.

## 4.4 Environment-aware AuthZ direction

Future multi-environment support must evaluate authorization with environment context.

Future dimensions:

```text
actor
role
action
targetEnvironment
approvalPolicy
```

Expected direction:

```text
dev:
  operators can execute standard guarded workflows

staging:
  operators can propose and validate
  approvers are required for controlled execution or promotion

production:
  operators can propose only
  production approvers or admins are required for approval and execution
```

Production workflows must remain disabled until environment-aware AuthZ and production guardrails are implemented and validated.

---

## 5. Required credentials

## 5.1 GitLab

Sensitive variable:

```text
GITLAB_TOKEN
```

Usage:

- read repositories;
- read files;
- create branches;
- update files;
- create commits;
- open merge requests;
- merge merge requests where policy allows;
- read merge request state.

Recommendation:

- use a Project Access Token when a single project is sufficient;
- use a Group Access Token when multiple repositories in a group are required;
- use a controlled service account or bot user where appropriate;
- avoid personal access tokens for stable usage;
- use the minimum privileges compatible with the workflow.

## 5.2 Argo CD

Sensitive variable:

```text
ARGOCD_AUTH_TOKEN
```

Usage:

- list allowed Applications;
- get Application details;
- read resources and history where needed;
- check deployment status;
- collect Argo CD evidence.

Current stable pattern:

```text
dedicated Argo CD account for DevOps Control Plane
minimum required permissions
API token stored in OpenShift Secret
```

Avoid global admin tokens as stable runtime configuration.

## 5.3 PostgreSQL

Sensitive configuration:

```text
DATABASE_URL
DATABASE_PASSWORD
```

If `DATABASE_URL` contains user/password, it must be treated as sensitive and stored in Secret.

## 5.4 Kubernetes/OpenShift

When the DevOps Control Plane runs in OpenShift, Kubernetes/OpenShift API access uses the mounted ServiceAccount token and CA.

Current baseline:

```text
KUBERNETES_TOKEN is not required in the application Secret
ServiceAccount token fallback is used
ServiceAccount CA is used where applicable
```

For local development only, a `KUBECONFIG` may be used, but it must never be committed or included in evidence.

---

## 6. ConfigMaps and Secrets

## 6.1 Application ConfigMap

Non-sensitive configuration may be stored in a ConfigMap.

Example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: devops-control-plane-config
  namespace: devops-control-plane
data:
  HTTP_ADDR: ":8080"
  LOG_LEVEL: "info"
  ARGOCD_BASE_URL: "https://openshift-gitops-server-openshift-gitops.apps.example.local"
  GITLAB_BASE_URL: "https://gitlab.example.local"
  TEKTON_NAMESPACE: "devops-ci-demo"
  TEKTON_PIPELINE_NAME: "validate-gitops"
  KUBERNETES_NAMESPACE: "devops-ci-demo"
  AUTH_ENABLED: "true"
  AUTH_OPENSHIFT_GROUP_LOOKUP_ENABLED: "true"
```

## 6.2 Application Secret

Sensitive values must be stored in Secret.

Example template with placeholders only:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: devops-control-plane-secrets
  namespace: devops-control-plane
type: Opaque
stringData:
  GITLAB_TOKEN: "<replace-me>"
  ARGOCD_AUTH_TOKEN: "<replace-me>"
  DATABASE_URL: "postgres://devops_cp:<replace-me>@postgresql:5432/devops_control_plane?sslmode=disable"
```

Real values must be applied through a secure operational procedure outside Git.

## 6.3 Trust bundle ConfigMap

TLS strict mode can require an application-dedicated trust bundle.

Current baseline uses:

```text
ConfigMap: dcp-app-trust-bundle
Mount: /etc/dcp-trust/ca-bundle.crt
```

This supports strict TLS for integrations such as Argo CD and GitLab without modifying OpenShift cluster-wide trust configuration.

---

## 7. ServiceAccounts

## 7.1 Runtime ServiceAccount

Current runtime uses a ServiceAccount associated with the OAuth Proxy deployment baseline.

The runtime ServiceAccount is responsible for:

- running the backend and proxy deployment;
- creating Tekton `PipelineRun` resources where authorized;
- reading Tekton `PipelineRun` and `TaskRun` resources;
- reading runtime resources required for evidence;
- reading OpenShift group membership through the dedicated group reader permission.

## 7.2 PipelineRun ServiceAccount

The Tekton PipelineRun ServiceAccount is responsible for executing validation tasks.

It should have only the permissions required for validation:

- clone repository using configured credentials;
- validate manifests;
- perform dry-run checks if required;
- avoid broad access to Secrets or cluster-wide resources.

---

## 8. RBAC baseline

## 8.1 Tekton runtime permissions

The runtime ServiceAccount must have least-privilege access to the configured Tekton namespace.

Required permissions include:

```text
create/get/list PipelineRun
get/list TaskRun
```

The exact namespace is currently `devops-ci-demo` for the lab baseline and should become environment-aware in future phases.

## 8.2 Runtime evidence permissions

For runtime evidence, the ServiceAccount requires read-only access to authorized application namespaces.

Typical resources:

```text
Deployments
Pods
Services
Routes
```

Rules:

- read-only access where possible;
- no default access to Secrets;
- no cluster-admin;
- namespace-specific RoleBindings preferred.

## 8.3 OpenShift group lookup permissions

The backend group resolver requires permission to read OpenShift groups.

This is cluster-scoped by nature and must remain limited to the minimum resource and verbs required.

## 8.4 PostgreSQL NetworkPolicy

PostgreSQL ingress must be restricted.

Current baseline includes a NetworkPolicy allowing PostgreSQL ingress only from DevOps Control Plane pods on TCP 5432.

This reduces namespace blast radius without introducing a broad deny-all policy before a dedicated network policy design phase.

---

## 9. GitLab authorization model

The GitLab token must allow only the operations required by the workflow.

Minimum capabilities:

- read repository;
- read files;
- create branch;
- update files or create commits;
- open merge request;
- merge merge request where policy allows;
- read merge request state.

Preferred token order:

1. Project Access Token;
2. Group Access Token;
3. controlled bot or service account token;
4. Personal Access Token only for development or lab.

---

## 10. Argo CD authorization model

The Argo CD token must allow only required Application operations.

Expected capabilities:

- list allowed Applications;
- get allowed Application details;
- read Application status and conditions;
- read resources and history where required;
- perform deployment checks.

If future sync operations are enabled through the Control Plane, sync permission must be granted only to the required Applications and environments.

---

## 11. PostgreSQL security

## 11.1 Dedicated database user

Use a dedicated database user, for example:

```text
devops_cp
```

Avoid running the application as PostgreSQL superuser.

## 11.2 Minimum database permissions

The application user should have permissions only on the DevOps Control Plane database/schema.

Example conceptual setup:

```sql
CREATE USER devops_cp WITH PASSWORD '<replace-me>';
CREATE DATABASE devops_control_plane OWNER devops_cp;
```

More advanced separation between schema owner and application user can be introduced later.

## 11.3 Backup and restore protection

PostgreSQL backups contain ChangeRequests, events and evidence.

Backup artifacts must be protected and checksummed. Restore tests must be performed in isolated namespaces or safe targets.

---

## 12. Secure logging

## 12.1 Forbidden log content

Do not log:

- GitLab token;
- Argo CD token;
- PostgreSQL password;
- Kubernetes bearer token;
- kubeconfig;
- Kubernetes Secret values;
- Authorization headers;
- Docker auth JSON;
- private keys.

## 12.2 Masking examples

If configuration needs to be logged, sensitive values must be masked.

```text
GITLAB_TOKEN=****
ARGOCD_AUTH_TOKEN=****
DATABASE_URL=postgres://devops_cp:****@postgresql:5432/devops_control_plane
```

---

## 13. Evidence sanitization

## 13.1 Allowed evidence content

Evidence can include:

- Application name;
- namespace;
- target environment;
- Argo CD sync and health status;
- Git commit SHA;
- merge request URL;
- Tekton PipelineRun name;
- TaskRun status;
- Deployment status;
- Pod status;
- Service and Route status;
- sanitized technical errors.

## 13.2 Forbidden evidence content

Evidence must not include:

- tokens;
- passwords;
- Secret values;
- kubeconfig;
- private keys;
- raw Secret YAML;
- Docker auth JSON;
- Authorization headers.

---

## 14. Anti-secret validation

## 14.1 Minimum detection patterns

The validation pipeline should detect common suspicious patterns such as:

```text
token
password
secret
client_secret
authToken
secret_key
private_key
PRIVATE KEY
BEGIN RSA
AWS access key patterns
bearer
authorization
.dockerconfigjson
config.json
```

## 14.2 Action on detection

If suspicious content is detected, validation must fail or require explicit review according to policy.

Recommended runtime status:

```text
ValidationFailed
```

Recommended error family:

```text
VALIDATION_SECRET_DETECTED
```

---

## 15. Namespace and blast radius

## 15.1 Dedicated namespace

Use a dedicated namespace:

```text
devops-control-plane
```

Benefits:

- clearer separation;
- isolated Secrets;
- explicit RBAC;
- dedicated deployment and service;
- simpler NetworkPolicy model.

## 15.2 Authorized application namespaces

The Control Plane must operate only on explicitly authorized target namespaces.

Current lab namespace:

```text
devops-ci-demo
```

Future environment-aware configuration will map environments to namespaces.

---

## 16. Versioned and non-versioned files

## 16.1 Versioned files

Allowed in Git:

- manifest templates;
- non-sensitive ConfigMaps;
- Secret templates with placeholders;
- ServiceAccounts;
- Roles and RoleBindings;
- NetworkPolicies;
- Tekton pipelines;
- documentation;
- Go code;
- SQL migrations.

## 16.2 Non-versioned files

Not allowed in Git:

- real `.env` files;
- tokens;
- passwords;
- private keys;
- kubeconfigs;
- real Secrets;
- private certificates;
- evidence containing sensitive values.

---

## 17. Threat model

| Threat | Impact | Mitigation |
|---|---|---|
| GitLab token committed to Git | Repository compromise | Secret storage, `.gitignore`, anti-secret validation |
| Argo CD token printed in logs | Unauthorized application access | masking, no token logging, token rotation |
| Overprivileged ServiceAccount | Broad cluster impact | least-privilege RBAC, namespace RoleBindings |
| Evidence contains Secret values | Data leakage | evidence sanitization |
| PipelineRun has excessive permissions | Runtime mutation risk | dedicated PipelineRun ServiceAccount |
| GitOps drift is hidden | Non-auditable runtime state | Argo CD status and evidence |
| Change branch not correlated | Poor audit trail | ChangeRequest number in branch/events |
| Anonymous UI/API access | Unauthorized access | OAuth Proxy and backend AuthZ |
| Stale token after exposure | Unauthorized access | documented rotation procedure |
| Production action by dev-only role | Production impact | future environment-aware AuthZ |

---

## 18. Security validation checklist

Before a milestone is considered stable:

- `.env` is ignored by Git;
- no token is present in Git;
- real Secrets are not committed;
- tokens do not appear in logs;
- evidence does not contain Secret values;
- ServiceAccount is not cluster-admin;
- RoleBindings are namespace-scoped where possible;
- GitLab token has minimum required privileges;
- Argo CD token is dedicated and not global admin where possible;
- Kubernetes static token is not required in the current runtime baseline;
- PipelineRun uses a controlled ServiceAccount;
- anti-secret validation is active;
- timeouts are configured;
- HTTP 401/403 errors are handled without leaks;
- OAuth Proxy protects UI and API routes;
- backend AuthZ is fail-closed;
- OpenShift group lookup can be validated;
- NetworkPolicy for PostgreSQL is present;
- TLS strict mode is enabled where supported.

---

## 19. Security roadmap

Possible future improvements:

- richer environment-aware AuthZ;
- production-specific approver groups;
- dual approval for production changes;
- integration with Vault or other Secret managers;
- automated token rotation;
- policy engine integration with OPA, Kyverno or similar tools;
- audit export;
- commit signing;
- signature verification;
- advanced secret scanning;
- SIEM integration;
- more explicit deny-all NetworkPolicy design.

---

## 20. Relationship with other documents

This document informs and is informed by:

- `manifests/serviceaccount.yaml`;
- `manifests/role.yaml`;
- `manifests/rolebinding.yaml`;
- `manifests/configmap.yaml`;
- `manifests/secret-template.yaml`;
- `manifests/networkpolicies/postgresql-ingress-from-devops-control-plane.yaml`;
- `pipelines/validate-gitops.yaml`;
- `docs/04-non-functional-requirements.md`;
- `docs/05-architecture.md`;
- `docs/07-gitlab-integration.md`;
- `docs/08-tekton-integration.md`;
- `docs/runbooks/secrets-rotation.md`;
- `docs/runbooks/authn-authz.md`;
- `docs/adr/ADR-0009-authn-authz-strategy.md`;
- `docs/adr/ADR-0010-oauth-proxy-deployment-design.md`;
- `docs/adr/ADR-0011-multi-environment-model.md`.

---

## 21. Key message

Security must be designed into the DevOps Control Plane from the first MVP onward.

The Control Plane orchestrates GitLab, Tekton, Argo CD and OpenShift. A mistake in credentials, RBAC or evidence handling can have significant operational impact.

The core rule remains:

```text
Least privilege, no secrets in Git, no secrets in logs, no secrets in evidence.
```

The DevOps Control Plane must make GitOps changes safer and more traceable, not introduce a new ungoverned risk surface.

---

## 22. Revision history

| Date | Version | Description |
|---|---:|---|
| 2026-06-25 | 0.1 | Initial security and RBAC document in Italian. |
| 2026-07-06 | 0.2 | Rewritten in English and refreshed while preserving the original least-privilege and sanitized-evidence security intent and aligning it with OAuth Proxy, AuthN/AuthZ, RBAC, TLS, NetworkPolicy and environment-aware security direction. |
