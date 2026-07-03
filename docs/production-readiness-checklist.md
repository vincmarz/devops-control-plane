# DevOps Control Plane — Production Readiness Checklist

## Document metadata

- **Project:** DevOps Control Plane
- **Phase:** 10.7 — Production readiness checklist
- **Scope:** Final operational checklist to assess production readiness of the DevOps Control Plane MVP baseline on OpenShift
- **Audience:** Service owner, platform engineers, DevOps engineers, security reviewers, operations team and onboarding readers
- **Execution mode:** Primarily read-only validation; remediation actions must follow the referenced runbooks
- **Security posture:** Do not print Secret values, tokens, database passwords or kubeconfig credentials
- **Related runbooks and scripts:**
  - `docs/runbooks/operability-health-check.md`
  - `scripts/operability/health-check.sh`
  - `docs/runbooks/postgresql-backup-restore.md`
  - `docs/runbooks/disaster-recovery.md`
  - `docs/runbooks/maintenance-operations.md`
  - `docs/runbooks/secrets-rotation.md`
  - `docs/runbooks/authn-authz.md`

---

## 1. Purpose

This checklist provides a consolidated go/no-go assessment for the DevOps Control Plane production readiness baseline.

It does not replace the detailed runbooks. Instead, it acts as an executive and operational checklist that references the runbooks and highlights:

1. what is already ready;
2. what must be validated before production promotion;
3. what is acceptable for MVP/lab baseline;
4. what remains a known limitation;
5. what requires explicit service-owner acceptance.

The checklist can be used for:

- production readiness review;
- pre-release validation;
- operational handover;
- audit preparation;
- onboarding of new maintainers;
- post-incident baseline confirmation.

---

## 2. Readiness classification

Use the following status values.

```text
READY       = validated and acceptable for the target baseline
ACCEPTED    = known limitation accepted by service owner for this phase
CONDITIONAL = acceptable only with documented constraints or compensating controls
NOT READY   = blocker for production promotion
N/A         = not applicable to the current deployment profile
```

A production promotion decision should not be based only on technical status. The service owner must explicitly accept RPO/RTO, known limitations and operational responsibilities.

---

## 3. Current baseline summary

Current validated baseline:

```text
Namespace: devops-control-plane
Application pod: Running 2/2, restart 0
PostgreSQL pod: Running 1/1, restart 0
OAuth Proxy: enabled as sidecar
Auth middleware: enabled
OpenShift group lookup: enabled
TLS strict mode: enabled for Argo CD, GitLab and Kubernetes/OpenShift
KUBERNETES_TOKEN: removed from application Secret
PostgreSQL backup/restore: validated with isolated restore test
Operability smoke test: available and validated
Disaster recovery runbook: available
Maintenance operations runbook: available
```

Latest known healthy smoke-test baseline:

```text
PASS=35
WARN=0
FAIL=0
```

Expected route behavior:

```text
/readyz                         HTTP 200
/livez                          HTTP 200
/api/v1/changes anonymous       HTTP 403
/ui/dashboard anonymous         HTTP 403
```

Expected backend direct behavior through controlled port-forward:

```text
/readyz                         HTTP 200
/livez                          HTTP 200
/api/v1/changes anonymous       HTTP 401
/api/v1/changes admin headers   HTTP 200
```

---

## 4. Mandatory pre-review smoke test

Before filling this checklist, run:

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

Expected result:

```text
PASS=35
WARN=0
FAIL=0
```

Record the generated evidence directory:

```text
Evidence directory: /tmp/dcp-operability-smoke-test-YYYYMMDD-HHMMSS
```

If the smoke test returns `FAIL > 0`, stop the readiness review and follow:

```text
docs/runbooks/operability-health-check.md
```

---

## 5. Application availability checklist

### 5.1 Workload readiness

Required checks:

```bash
oc get deploy -n devops-control-plane -o wide
oc get pods -n devops-control-plane -o wide
```

Acceptance criteria:

```text
Deployment devops-control-plane available replicas >= 1
Deployment postgresql available replicas >= 1
DCP pod Running 2/2
PostgreSQL pod Running 1/1
Restart counters stable and ideally 0
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

### 5.2 Service and Route readiness

Required checks:

```bash
oc get svc -n devops-control-plane -o wide
oc get route -n devops-control-plane -o wide
```

Acceptance criteria:

```text
Service devops-control-plane exposes HTTPS/OAuth Proxy port 8443
Service postgresql exposes 5432
Route devops-control-plane exists
Route termination is reencrypt/Redirect
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 6. Health endpoints checklist

Required checks:

```bash
unset EVIDENCE_DIR
./scripts/operability/health-check.sh
```

Acceptance criteria:

```text
Route /readyz returns HTTP 200
Route /livez returns HTTP 200
Backend /readyz returns HTTP 200
Backend /livez returns HTTP 200
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 7. AuthN/AuthZ checklist

### 7.1 OAuth Proxy readiness

Required checks:

```bash
oc get deployment devops-control-plane -n devops-control-plane -o json \
  | jq -r '.spec.template.spec.containers[] | select(.name=="oauth-proxy") | .args[]'
```

Acceptance criteria:

```text
OAuth Proxy container is present
Provider is openshift
Upstream points to localhost:8080
Health skip-auth regex includes ^/readyz$ and ^/livez$
User headers are passed
Bearer/access/basic auth forwarding is disabled
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

### 7.2 Anonymous access policy

Acceptance criteria:

```text
Route anonymous /api/v1/changes returns HTTP 403
Route anonymous /ui/dashboard returns HTTP 403
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

### 7.3 Backend authorization policy

Acceptance criteria:

```text
Backend anonymous /api/v1/changes returns HTTP 401 or 403
Backend admin-header /api/v1/changes returns HTTP 200
Endpoint not classified by AuthZ policy is denied
Role matrix has been validated according to docs/runbooks/authn-authz.md
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

### 7.4 OpenShift group lookup

Required checks:

```bash
oc get clusterrolebinding devops-control-plane-group-reader -o yaml
oc get group devops-control-plane-admins -o yaml 2>/dev/null || true
oc get group devops-control-plane-operators -o yaml 2>/dev/null || true
oc get group devops-control-plane-approvers -o yaml 2>/dev/null || true
oc get group devops-control-plane-viewers -o yaml 2>/dev/null || true
```

Acceptance criteria:

```text
AUTH_OPENSHIFT_GROUP_LOOKUP_ENABLED=true
ClusterRoleBinding devops-control-plane-group-reader exists
Expected OpenShift groups exist or are intentionally deferred
Real user access has been validated by browser/UI or controlled request
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 8. TLS and trust checklist

Required sanitized configuration check:

```bash
oc get configmap devops-control-plane-config -n devops-control-plane -o json \
  | jq -r '.data | to_entries[] | select(.key | test("TLS|CA_FILE|INSECURE|ARGOCD|GITLAB|KUBERNETES")) | .key + "=" + .value' \
  | sort
```

Acceptance criteria:

```text
ARGOCD_INSECURE_TLS=false
GITLAB_INSECURE_TLS=false
KUBERNETES_INSECURE_TLS=false
ARGOCD_CA_FILE=/etc/dcp-trust/ca-bundle.crt
GITLAB_CA_FILE=/etc/dcp-trust/ca-bundle.crt
Kubernetes/OpenShift uses ServiceAccount CA fallback or valid CA configuration
App-dedicated trust bundle exists and is mounted
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 9. Secrets and token management checklist

Required checks without printing values:

```bash
oc get secret devops-control-plane-secrets -n devops-control-plane -o json \
  | jq -r '.data | keys[]' \
  | sort

oc get secret devops-control-plane-secrets -n devops-control-plane -o json \
  | jq -r '.data | to_entries[] | .key + " base64_length=" + (.value | length | tostring)' \
  | sort
```

Acceptance criteria:

```text
Required keys are present: ARGOCD_AUTH_TOKEN, DATABASE_URL, GITLAB_TOKEN
KUBERNETES_TOKEN is not required and should not be present
No Secret values are committed to the repository
No Secret values are printed in operational evidence
Secrets rotation runbook exists
Latest known exposed Argo CD token has been rotated
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 10. RBAC checklist

Required checks:

```bash
DCP_SA="system:serviceaccount:devops-control-plane:devops-control-plane-oauth-proxy"

oc auth can-i create pipelineruns.tekton.dev -n devops-ci-demo --as "$DCP_SA"
oc auth can-i get deployments -n devops-ci-demo --as "$DCP_SA"
oc auth can-i list pods -n devops-ci-demo --as "$DCP_SA"
oc auth can-i get secrets -n devops-control-plane --as "$DCP_SA"
```

Acceptance criteria:

```text
create pipelineruns.tekton.dev in devops-ci-demo: yes
get deployments in devops-ci-demo: yes
list pods in devops-ci-demo: yes
get secrets in devops-control-plane: no
system:auth-delegator binding exists for OAuth Proxy
OpenShift group-reader binding exists for group lookup
No broad admin binding is used by the application runtime ServiceAccount
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 11. NetworkPolicy checklist

Required checks:

```bash
oc get networkpolicy -n devops-control-plane -o wide
oc get networkpolicy postgresql-ingress-from-devops-control-plane -n devops-control-plane -o yaml
```

Acceptance criteria:

```text
NetworkPolicy postgresql-ingress-from-devops-control-plane exists
Policy selects app=postgresql
Policy allows ingress from app=devops-control-plane on TCP 5432
Application can still connect to PostgreSQL
No untested deny-all egress policy is introduced
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 12. PostgreSQL readiness checklist

Required checks:

```bash
oc get deploy postgresql -n devops-control-plane -o wide
oc get pod -n devops-control-plane -l app=postgresql -o wide
oc get pvc postgresql-data -n devops-control-plane -o wide
```

Read-only SQL validation:

```bash
PG_POD="$(oc get pod -n devops-control-plane -l app=postgresql -o jsonpath='{.items[0].metadata.name}')"

oc exec -n devops-control-plane "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select now() as db_time, current_database() as database_name, current_user as user_name;"'

oc exec -n devops-control-plane "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as change_requests_count from change_requests;"'

oc exec -n devops-control-plane "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as change_events_count from change_events;"'

oc exec -n devops-control-plane "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as evidences_count from evidences;"'
```

Acceptance criteria:

```text
PostgreSQL deployment available
PostgreSQL pod Running 1/1
PVC postgresql-data Bound
Database connection succeeds
Expected tables are queryable
Counts are consistent with current baseline
```

Current known baseline counts:

```text
change_requests = 1
change_events = 34
evidences = 18
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 13. Backup and restore checklist

Reference:

```text
docs/runbooks/postgresql-backup-restore.md
```

Required evidence:

```text
latest backup dump file exists
checksum file exists
checksum validation succeeds
restore TOC list is available
restore validation plan is available
isolated restore test has been executed successfully
```

Current validated baseline:

```text
dcp-postgresql-20260702-160033.dump
SHA256 f35affbb14f235a1b3cc21613e1ed677db08d40f1116ed918fc9beb212e19eaa
isolated restore namespace devops-control-plane-restore-test validated and cleaned up
```

Acceptance criteria:

```text
Backup can be validated by checksum
Restore can be completed in isolated namespace
Validation queries return expected results
Temporary restore namespace is cleaned up
Backup/restore runbook is committed in repository
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 14. Disaster recovery checklist

Reference:

```text
docs/runbooks/disaster-recovery.md
```

Acceptance criteria:

```text
Protected components are documented
RPO/RTO are declared, even if still requiring formal service owner approval
DR scenarios covered and not covered are documented
Backup artifacts are defined
Restore target isolation is documented
Success/failure criteria are defined
Incident recovery decision flow is documented
Roles and responsibilities are defined
Current limitations are documented
```

Service owner decision required:

```text
[ ] RPO accepted
[ ] RTO accepted
[ ] Current single-replica PostgreSQL limitation accepted or remediation planned
[ ] Backup retention policy accepted or remediation planned
[ ] Off-cluster backup storage accepted or remediation planned
```

Status:

```text
[ ] READY
[ ] ACCEPTED
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 15. Maintenance operations checklist

Reference:

```text
docs/runbooks/maintenance-operations.md
```

Acceptance criteria:

```text
Pre-maintenance checks are documented
Application rollout procedure is documented
Rollback procedure is documented
ConfigMap maintenance procedure is documented
Secret/token rotation coordination is documented
Trust bundle maintenance is documented
OAuth Proxy maintenance is documented
AuthN/AuthZ validation is documented
RBAC validation is documented
NetworkPolicy validation is documented
PostgreSQL maintenance is documented
Post-maintenance smoke test is documented
Closure and escalation criteria are documented
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 16. Observability and operability checklist

References:

```text
docs/runbooks/operability-health-check.md
scripts/operability/health-check.sh
```

Acceptance criteria:

```text
Manual health-check runbook exists
Automated smoke-test script exists
Smoke-test script has been validated runtime
Evidence directory is generated automatically
Script supports optional can-i checks
Exit codes are documented
Latest known runtime test result is PASS=35 WARN=0 FAIL=0
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 17. External integrations checklist

### 17.1 GitLab

Acceptance criteria:

```text
GITLAB_BASE_URL configured
GITLAB_PROJECT_ID configured
GITLAB_INSECURE_TLS=false
GITLAB_CA_FILE configured
GITLAB_TOKEN present as Secret key only
Branch/update/MR workflow previously validated
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

### 17.2 Tekton

Acceptance criteria:

```text
TEKTON_NAMESPACE configured
TEKTON_PIPELINE_NAME=validate-gitops
TEKTON_WORKSPACE_PVC configured
Runtime ServiceAccount can create PipelineRuns
Validation evidence and diagnostics are implemented
Policy/anti-secret hardening is implemented
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

### 17.3 Argo CD

Acceptance criteria:

```text
ARGOCD_BASE_URL configured
ARGOCD_INSECURE_TLS=false
ARGOCD_CA_FILE configured
Dedicated Argo CD account/token baseline documented
Application read/check-deployment workflow validated
Deployment evidence collection validated
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

### 17.4 Kubernetes/OpenShift runtime evidence

Acceptance criteria:

```text
KUBERNETES_INSECURE_TLS=false
Static KUBERNETES_TOKEN removed
ServiceAccount token fallback implemented and validated
Runtime evidence collection works
Least-privilege RBAC is applied
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 18. UI readiness checklist

Acceptance criteria:

```text
UI dashboard is available to authorized users
Anonymous UI access is blocked through the Route
Change list and change details are rendered
Runtime evidence is rendered in user-friendly form
Raw JSON links are wrapped by UI pages where applicable
UI language is English
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 19. Repository and documentation checklist

Acceptance criteria:

```text
main and origin/main are aligned
working tree is clean
runbooks are committed
operability scripts are committed
manifest changes are committed
ADR documentation is reviewed or explicitly marked as pending
technical final document is in progress
```

Required checks:

```bash
git status --short
git log --oneline -5
```

Known status:

```text
Fase 0 documentation/ADR: partial and still to update
Fase 12 final technical document: in incremental production
```

Status:

```text
[ ] READY
[ ] CONDITIONAL
[ ] NOT READY
```

Notes:

```text

```

---

## 20. Known limitations and accepted risks

The following limitations must be explicitly accepted or converted into remediation items.

```text
[ ] PostgreSQL currently single-replica
[ ] Some images still use latest tags
[ ] Scheduled backup automation not yet implemented in repository baseline
[ ] Backup retention policy not yet formalized
[ ] Off-cluster backup storage not yet formalized
[ ] Full OpenShift cluster DR is out of scope
[ ] GitLab, Argo CD and Tekton DR are out of scope for this runbook set
[ ] Strict deny-all egress NetworkPolicy design is deferred
[ ] Fase 0 ADR/documentation still partial
[ ] Fase 11 CLI is optional and deferred
```

Service owner acceptance:

```text
Owner: ___________________________
Date:  ___________________________
Decision:
[ ] Accepted for MVP production-like baseline
[ ] Accepted only for lab/non-production
[ ] Not accepted; remediation required before production
```

---

## 21. Go/no-go decision

### 21.1 Go criteria

A `GO` decision requires:

```text
[ ] Smoke test PASS=35 WARN=0 FAIL=0 or equivalent current baseline with no FAIL
[ ] Application pods Ready
[ ] PostgreSQL Ready
[ ] Health endpoints OK
[ ] Anonymous API/UI blocked
[ ] Admin/authorized access works
[ ] TLS strict mode enabled
[ ] Secret inventory clean and no token leakage known
[ ] RBAC least privilege validated
[ ] NetworkPolicy baseline present
[ ] Backup/restore validated
[ ] DR runbook available
[ ] Maintenance runbook available
[ ] Known limitations accepted by service owner
[ ] Evidence package archived
```

### 21.2 No-go criteria

A `NO-GO` decision is required if any of the following are true:

```text
[ ] Smoke test has FAIL > 0
[ ] Application pod not Ready
[ ] PostgreSQL pod not Ready
[ ] Anonymous API/UI returns HTTP 200
[ ] TLS strict mode disabled unexpectedly
[ ] Secret/token value exposed and not rotated
[ ] PostgreSQL backup integrity cannot be validated
[ ] Restore validation has never succeeded
[ ] RBAC grants broad access to runtime ServiceAccount without approval
[ ] Required runbooks missing
[ ] Service owner does not accept RPO/RTO or known limitations
```

### 21.3 Final decision

```text
Decision:
[ ] GO
[ ] CONDITIONAL GO
[ ] NO-GO

Decision owner: ___________________________
Date:           ___________________________
Evidence dir:   ___________________________
Notes:          ___________________________
```

---

## 22. Production readiness summary for current baseline

Based on the latest validated phases, the DevOps Control Plane can be described as:

```text
Advanced MVP / production-oriented baseline
```

Strong points:

```text
Backend foundation complete
Change lifecycle and audit complete
GitLab MR workflow complete
Tekton validation strengthened
Argo CD and Kubernetes runtime evidence strengthened
UI MVP advanced
OpenShift deployment working
OAuth Proxy and AuthN/AuthZ enabled
TLS strict baseline implemented
Static Kubernetes token removed
RBAC least privilege baseline implemented
PostgreSQL NetworkPolicy baseline implemented
Backup/restore validated
Operability health-check runbook available
Automated smoke-test script available and validated
Disaster recovery baseline documented
Maintenance operations documented
```

Remaining production-hardening items:

```text
Formal service owner RPO/RTO acceptance
Scheduled/off-cluster backup strategy
Image pinning instead of latest
PostgreSQL HA assessment
Full monitoring/alerting dashboards
Log retention/forwarding policy
Cluster/external systems DR procedures
Fase 0 ADR/documentation refresh
Fase 12 final technical document completion
```

---

## 23. Revision history

| Date | Phase | Description |
|---|---:|---|
| 2026-07-03 | 10.7 | Initial production readiness checklist for the DevOps Control Plane advanced MVP baseline. |
