# Runtime Evidence, Dashboard and Maintenance Alignment

Status: Completed  
Phase: 13.6 — Final Phase 13 documentation alignment summary  
Owner: Vincenzo Marzario  
Last updated: 2026-07-09

## 1. Purpose

This document formally closes Phase 13 — Runtime evidence, dashboard and maintenance refinement — as completed and aligned with the current post-Phase 15 DevOps Control Plane baseline.

The purpose of Phase 13 alignment was not to add new runtime features. Its purpose was to refresh the documentation so that evidence, dashboard behavior, ChangeRequest workflows and maintenance operations accurately describe the current implemented system.

## 2. Final status

Phase 13 is now considered:

`COMPLETED AND DOCUMENTATION-ALIGNED AFTER PHASE 15`

The DevOps Control Plane now has documented alignment for:

- runtime evidence;
- Tekton validation evidence;
- Argo CD deployment evidence;
- dashboard latest ChangeRequest behavior;
- UI environment and namespace visibility;
- ChangeRequest detail evidence rendering;
- maintenance operations after Phase 15;
- namespace-isolated dev, staging and production runtime;
- multi-cluster code-ready baseline;
- fail-closed guardrails for providers, Secret references and factories.

## 3. Current runtime baseline

The current validated runtime baseline remains namespace-isolated on the available `ocp-dev` OpenShift cluster:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

Physical cross-cluster runtime validation remains deferred because no additional OpenShift cluster is currently available.

The codebase remains multi-cluster code-ready and has been validated with simulated external cluster targets for staging and production.

## 4. Documents aligned during Phase 13

The following documents were refreshed or aligned.

### Evidence model

Document:

`docs/12-evidence-model.md`

Commit:

`f43a6ab` — `Refresh evidence model after Phase 15 runtime validation`

Alignment completed:

- runtime evidence;
- Tekton validation evidence;
- latest validation evidence;
- failed task count;
- evidence sanitization;
- environment-specific validation paths;
- UI evidence rendering;
- staging evidence for `CHG-2026-0049`;
- production evidence for `CHG-2026-0050`;
- multi-cluster readiness implications.

### Dashboard and UI architecture

Document:

`docs/05-architecture.md`

Commit:

`dd9273b` — `Refresh dashboard and UI architecture after Phase 15`

Alignment completed:

- dashboard latest ChangeRequest behavior;
- `Environments / Namespaces` topbar;
- user context segment;
- Tekton validation evidence card;
- runtime evidence cards;
- namespace-isolated UI representation;
- multi-cluster code-ready posture;
- simulated staging and production external-cluster readiness.

### Maintenance operations

Document:

`docs/runbooks/maintenance-operations.md`

Commit:

`912483d` — `Refresh maintenance operations after Phase 15`

Alignment completed:

- post-Phase 15 maintenance validation;
- pre-maintenance and post-maintenance smoke matrix;
- Argo CD validation for dev, staging and production;
- Tekton validation for staging and production;
- UI evidence validation after maintenance;
- Environment Catalog maintenance;
- Cluster Registry and provider maintenance;
- Secret, RBAC and factory maintenance;
- future real multi-cluster maintenance notes.

### Change workflows

Document:

`docs/11-change-workflows.md`

Commit:

`32c9c5e` — `Refresh change workflows after Phase 15`

Alignment completed:

- create workflow;
- runtime target resolution;
- collect-evidence workflow;
- check-deployment workflow;
- validate workflow;
- check-validation workflow;
- staging validation evidence;
- production validation evidence;
- UI workflow visibility;
- evidence rendering behavior;
- multi-cluster code-ready workflow behavior.

## 5. Evidence model alignment summary

The evidence model now describes the current behavior of the platform.

Evidence can now be understood as operational proof for:

- which ChangeRequest was processed;
- which environment was targeted;
- which namespace was used;
- which deployment was checked;
- which Argo CD Application was observed;
- which Tekton PipelineRun executed;
- which validation path was used;
- whether validation succeeded;
- whether any task failed;
- whether the evidence was sanitized.

The evidence model explicitly distinguishes between:

- runtime evidence;
- deployment evidence;
- Tekton validation evidence;
- sanitized diagnostic evidence.

## 6. Dashboard alignment summary

The dashboard is now documented as an operational surface, not only as a basic UI MVP.

The current dashboard behavior includes:

- latest ChangeRequest selection;
- environment and namespace visibility;
- runtime evidence visibility;
- Tekton validation evidence visibility;
- ChangeRequest detail evidence rendering;
- operator-oriented summary of the runtime state.

The UI no longer hides staging and production behind a static dev-only representation.

## 7. Maintenance alignment summary

Maintenance documentation now reflects the fact that operators must validate all three logical environments:

- dev;
- staging;
- production.

Maintenance validation now includes:

- Argo CD Application health;
- application deployment readiness;
- route health checks;
- Tekton validation state;
- runtime evidence checks;
- UI evidence checks;
- Secret and RBAC guardrail checks;
- provider and factory fail-closed checks.

## 8. Workflow alignment summary

The Change Workflow documentation now reflects the current end-to-end flow:

1. create ChangeRequest;
2. resolve target environment;
3. resolve runtime target;
4. collect runtime evidence;
5. check deployment;
6. start Tekton validation;
7. check Tekton validation;
8. persist sanitized evidence;
9. render evidence in UI.

The workflow now explicitly covers dev, staging and production as namespace-isolated environments on `ocp-dev`.

## 9. Multi-cluster readiness alignment

Phase 13 documentation now aligns with the Phase 15 outcome.

Current position:

- physical cross-cluster validation is deferred;
- namespace-isolated runtime baseline is validated;
- multi-cluster code readiness is completed;
- simulated staging external-cluster readiness is validated;
- simulated production external-cluster readiness is validated;
- missing runtime provider fails closed;
- disabled runtime provider fails closed;
- Secret and factory guardrails are documented.

## 10. Remaining documentation work

The following broader documentation activities remain outside Phase 13 closure:

- final organic technical document for the whole project;
- broader architecture narrative for new readers;
- final beginner-friendly explanation of namespace isolation, GitOps, Argo CD, Tekton and control-plane workflows;
- future real-cluster onboarding documentation when additional infrastructure becomes available.

These items belong primarily to Phase 12 and ongoing Phase 0 documentation work.

## 11. Closure statement

Phase 13 is formally closed as documentation-aligned after Phase 15.

The DevOps Control Plane documentation now reflects the implemented runtime evidence model, dashboard behavior, maintenance procedures and ChangeRequest workflow for the current namespace-isolated multi-environment baseline.

Physical multi-cluster documentation will be extended when a real additional OpenShift cluster becomes available.
