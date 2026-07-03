# ADR-0011 — Multi-environment DevOps Control Plane model

## Status

Accepted for design baseline.

## Date

2026-07-03

## Context

The DevOps Control Plane currently operates mainly against the `dev` runtime baseline. The current implementation already includes a `targetEnvironment` field in the ChangeRequest domain model, and the UI currently displays an environment selector placeholder with `dev`.

The project has reached an advanced operational baseline:

- backend foundation with Go and PostgreSQL is complete;
- ChangeRequest lifecycle and audit are implemented;
- GitLab merge request workflow is implemented;
- Tekton validation integration and GitOps policy validation are implemented;
- Argo CD deployment checks and Kubernetes/OpenShift runtime evidence are implemented;
- UI dashboard and change pages are implemented;
- OpenShift deployment is operational;
- OAuth Proxy and AuthN/AuthZ are enabled;
- RBAC least-privilege and NetworkPolicy baseline are implemented;
- backup/restore, disaster recovery, maintenance and production readiness runbooks are available.

The next architectural question is how the DevOps Control Plane should support multiple target environments such as:

```text
dev
staging
production
```

Multi-environment support impacts:

- environment configuration;
- GitOps repository layout;
- Argo CD application mapping;
- Tekton validation parameters;
- ChangeRequest lifecycle and promotion model;
- AuthN/AuthZ and RBAC policy;
- evidence and audit semantics;
- UI dashboard and filtering;
- production readiness and operational procedures.

## Decision

The DevOps Control Plane will follow these architectural decisions.

### Decision 1 — Single multi-environment Control Plane instance

Use a single DevOps Control Plane instance capable of managing multiple target environments.

Chosen option:

```text
Variant B — One single multi-environment DevOps Control Plane instance
```

Rejected baseline alternative:

```text
One DevOps Control Plane instance per environment
```

Rationale:

- the project is conceptually a central Control Plane;
- the database already provides a central audit and evidence store;
- the UI already provides a unified dashboard and ChangeRequest list;
- `targetEnvironment` already exists in the domain model;
- a centralized model enables cross-environment visibility;
- operational and security complexity can be controlled through explicit environment configuration and RBAC/AuthZ policies.

### Decision 2 — Correlated ChangeRequests for promotion

Use correlated ChangeRequests to model promotion between environments.

Chosen option:

```text
Model 3 — Correlated ChangeRequests
```

Expected pattern:

```text
CHG-2026-0101 targets dev
CHG-2026-0102 targets staging and is promotedFrom CHG-2026-0101
CHG-2026-0103 targets production and is promotedFrom CHG-2026-0102
```

Rationale:

- each environment keeps its own explicit approval and evidence record;
- each ChangeRequest remains simple and targeted to one environment;
- audit remains clear;
- promotion traceability is preserved through correlation fields;
- the current model can evolve incrementally without redesigning the entire lifecycle around a complex multi-stage ChangeRequest object.

### Decision 3 — Environment-aware RBAC/AuthZ policy

Use differentiated RBAC/AuthZ policy per target environment.

Expected policy direction:

```text
dev:
  operators can create and execute standard technical workflows within guardrails

staging:
  operators can propose and validate
  approvers are required for controlled promotion and execution

production:
  operators can propose
  production approvers or admins are required for approval and execution
  stricter evidence and audit requirements apply
```

Rationale:

- production has higher risk than dev;
- environment policy must be explicit and auditable;
- the current roles `viewer`, `operator`, `approver`, `admin` can be extended with production-specific groups;
- OpenShift group lookup and trusted-header AuthN/AuthZ already provide a foundation for environment-aware authorization.

## Consequences

### Positive consequences

- Central dashboard across environments.
- Unified audit and evidence store.
- Clear promotion chain using correlated ChangeRequests.
- Stronger production governance through environment-aware authorization.
- Incremental implementation path from the current single-environment baseline.
- Easier final documentation because there is one Control Plane architecture.

### Negative consequences and trade-offs

- Configuration model becomes more complex.
- RBAC/AuthZ logic must include environment context.
- UI filtering and dashboard aggregation must become environment-aware.
- GitOps, Argo CD and Tekton integrations require environment mapping.
- A single Control Plane instance becomes a more critical operational component.
- Misconfiguration could affect multiple target environments if guardrails are weak.

### Required guardrails

- Explicit supported environment list.
- Environment-specific configuration mapping.
- Default deny for unknown environment values.
- Environment-aware AuthZ checks.
- Production-specific approval policy.
- Clear evidence attribution by environment.
- Audit events must include target environment and actor.
- UI must make the selected environment visible.
- Production actions must not be available from dev-only roles.

## Implementation direction

Phase 13 should proceed incrementally.

Recommended sub-phases:

```text
13.1 — Multi-environment architecture decision
13.2 — Environment configuration model
13.3 — Environment-aware ChangeRequest and promotion metadata design
13.4 — UI environment selector and filtering design
13.5 — GitOps layout for dev/staging/production
13.6 — Argo CD application mapping
13.7 — Tekton validation parameterization
13.8 — Environment-aware RBAC/AuthZ policy
13.9 — Evidence and audit environment enrichment
13.10 — MVP implementation for dev and staging
13.11 — Production-readiness extension for production
```

## Initial target environment model

Initial environment names:

```text
dev
staging
production
```

Initial high-level mapping:

```text
dev:
  displayName: Development
  riskProfile: low-to-medium
  approvalPolicy: relaxed

staging:
  displayName: Test / pre-production validation
  riskProfile: medium
  approvalPolicy: required

production:
  displayName: Production
  riskProfile: high
  approvalPolicy: strict
```

## Open questions

The following topics are intentionally deferred to the next design phases:

1. Exact ConfigMap format for environment definitions.
2. Whether environment definitions should eventually move to a database table.
3. Exact GitOps repository layout for overlays.
4. Required Argo CD Application naming convention.
5. Whether Tekton runs in one namespace or per-environment namespaces.
6. Exact production approval group names.
7. Whether production execution requires dual approval.
8. How environment filters should be represented in URLs.
9. Whether the UI default view should be `all` or `dev`.
10. Whether promotion links should be stored as `promotedFromChangeID`, `promotionGroupID`, or both.

## Decision summary

```text
Single Control Plane instance: accepted
Correlated ChangeRequests promotion model: accepted
Environment-aware RBAC/AuthZ policy: accepted
Implementation approach: incremental
```

## References

- `docs/phase-10-operability-closure.md`
- `docs/production-readiness-checklist.md`
- `docs/runbooks/authn-authz.md`
- `docs/runbooks/operability-health-check.md`
- `docs/runbooks/maintenance-operations.md`
