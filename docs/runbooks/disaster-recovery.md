# DevOps Control Plane — Disaster Recovery Operational Baseline

## Document metadata

- **Project:** DevOps Control Plane
- **Phase:** 10.5 — Disaster recovery operational baseline
- **Scope:** Operational disaster recovery baseline for the DevOps Control Plane runtime and PostgreSQL data layer on OpenShift
- **Audience:** Platform engineers, DevOps engineers, application operators, technical leads and onboarding readers
- **Execution mode:** Mostly read-only. Restore actions must be executed only in an isolated namespace unless an approved production incident procedure explicitly requires otherwise.
- **Security posture:** Do not print Secret values, database passwords, GitLab tokens or Argo CD tokens.
- **Related runbooks:**
  - `docs/runbooks/postgresql-backup-restore.md`
  - `docs/runbooks/operability-health-check.md`
  - `scripts/operability/health-check.sh`

---

## 1. Purpose

This runbook defines the disaster recovery operational baseline for the DevOps Control Plane.

The objective is to provide a clear, repeatable and safe procedure to:

1. understand what must be protected;
2. define baseline RPO/RTO expectations;
3. identify the required backup artifacts;
4. validate backup integrity;
5. execute restore tests in an isolated namespace;
6. collect evidence for audit and operational review;
7. define success/failure criteria;
8. document current limitations and production recommendations.

This runbook is intentionally conservative. It assumes that destructive actions against the active runtime are not allowed during regular validation activities.

---

## 2. Protected components

The current DevOps Control Plane baseline contains the following components.

### 2.1 Application runtime

- OpenShift namespace: `devops-control-plane`
- Deployment: `devops-control-plane`
- Containers:
  - `oauth-proxy`
  - `devops-control-plane`
- Service: `devops-control-plane`
- Route: `devops-control-plane`
- ConfigMap: `devops-control-plane-config`
- Secret: `devops-control-plane-secrets`
- Trust bundle ConfigMap: `dcp-app-trust-bundle`
- OAuth Proxy TLS and cookie Secret resources

### 2.2 Database runtime

- Deployment: `postgresql`
- Service: `postgresql`
- PVC: `postgresql-data`
- Secret: `postgresql-secret`
- Database: `devops_control_plane`

### 2.3 External dependencies

The DevOps Control Plane integrates with external or adjacent systems:

- GitLab API;
- Argo CD API;
- Kubernetes/OpenShift API;
- Tekton Pipelines API;
- OpenShift OAuth Proxy and OpenShift group API.

The primary state that must be protected for DevOps Control Plane recovery is PostgreSQL data. External systems are not backed up by this runbook and must have their own backup/restore policies.

---

## 3. Recovery objectives

The following values are the current operational baseline and must be reviewed before production promotion.

### 3.1 RPO — Recovery Point Objective

Current baseline:

```text
RPO = time since the latest valid PostgreSQL backup
```

For the current MVP/lab baseline, the RPO depends on how frequently manual or scheduled PostgreSQL backups are produced.

Production recommendation:

```text
Target RPO should be explicitly defined by the service owner.
Recommended initial target: <= 24 hours for low-volume MVP usage.
```

If the DevOps Control Plane becomes a critical production approval/audit system, a lower RPO may be required.

### 3.2 RTO — Recovery Time Objective

Current baseline:

```text
RTO = time required to provision an isolated PostgreSQL runtime, restore the dump, validate data and reconnect/redeploy the application if needed
```

Production recommendation:

```text
Target RTO should be explicitly defined by the service owner.
Recommended initial target: <= 4 hours for low-volume MVP usage.
```

The current baseline has already validated an isolated PostgreSQL restore test, but production RTO requires a timed exercise.

---

## 4. Disaster scenarios covered

This baseline covers the following scenarios:

1. PostgreSQL data corruption detected after a change or incident.
2. PostgreSQL PVC loss or unrecoverable storage failure.
3. Accidental deletion of PostgreSQL data.
4. Failed application deployment requiring rollback plus data validation.
5. Need to prove recoverability during an operational review.
6. Need to validate a backup in an isolated namespace without touching production runtime.

---

## 5. Disaster scenarios not fully covered

The following scenarios require additional platform-level procedures:

1. Full OpenShift cluster disaster recovery.
2. Loss of the internal OpenShift image registry.
3. Loss of GitLab data.
4. Loss of Argo CD state.
5. Loss of Tekton namespace resources.
6. Loss of cluster-wide OAuth/OpenShift identity provider configuration.
7. Regional or site-level disaster.
8. Compromise of tokens or credentials.

For token or credential compromise, use the relevant secrets rotation runbook and rotate exposed credentials immediately.

---

## 6. Safety rules

During DR validation:

- Do not restore into the active production namespace unless explicitly approved as part of an incident procedure.
- Prefer an isolated namespace for restore validation.
- Do not print Secret values.
- Do not decode Secret values in shared logs.
- Do not overwrite active PVCs.
- Do not delete active PostgreSQL resources during a test.
- Do not change external systems during a restore validation unless explicitly required.
- Preserve all evidence in a timestamped directory.
- If a token/password is accidentally printed, treat it as exposed and rotate it.

Commands that are safe for inventory:

```bash
oc get
oc describe
oc logs
oc exec with read-only SQL queries
pg_restore --list
sha256sum
```

Commands that require explicit approval in production:

```bash
oc delete
oc apply
oc patch
oc replace
oc scale
oc rollout restart
pg_restore into an active production database
DROP DATABASE
DROP TABLE
TRUNCATE
```

---

## 7. Required backup artifacts

A valid PostgreSQL backup package should contain at least:

```text
backup-dump-file.dump
backup-dump-file.dump.sha256
backup-metadata.txt
restore-toc-list.txt
restore-validation-plan.txt
```

Recommended naming convention:

```text
dcp-postgresql-YYYYMMDD-HHMMSS.dump
dcp-postgresql-YYYYMMDD-HHMMSS.dump.sha256
backup-metadata-YYYYMMDD-HHMMSS.txt
restore-toc-list-YYYYMMDD-HHMMSS.txt
restore-validation-plan-YYYYMMDD-HHMMSS.txt
```

Current validated example from the baseline:

```text
dcp-postgresql-20260702-160033.dump
```

with checksum:

```text
f35affbb14f235a1b3cc21613e1ed677db08d40f1116ed918fc9beb212e19eaa
```

The dump was validated using PostgreSQL 16.14 tooling from inside the PostgreSQL pod because the bastion `pg_restore` version was older and not aligned.

---

## 8. Evidence directory

Use a timestamped directory for every DR validation:

```bash
export DCP_NS="devops-control-plane"
export RESTORE_NS="devops-control-plane-restore-test"
export EVIDENCE_DIR="/tmp/dcp-dr-baseline-$(date +%Y%m%d-%H%M%S)"

mkdir -p "$EVIDENCE_DIR"

date -Is | tee "$EVIDENCE_DIR/00-timestamp.txt"

echo "DCP_NS=$DCP_NS" | tee "$EVIDENCE_DIR/00-context.txt"
echo "RESTORE_NS=$RESTORE_NS" | tee -a "$EVIDENCE_DIR/00-context.txt"
echo "EVIDENCE_DIR=$EVIDENCE_DIR" | tee -a "$EVIDENCE_DIR/00-context.txt"
```

---

## 9. Pre-DR health-check

Before any DR test, collect the current runtime health.

Recommended automated check:

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

Save or reference the generated evidence directory in the DR evidence package.

Manual minimum check:

```bash
oc get pods -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/01-source-pods.txt"
oc get deploy -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/02-source-deployments.txt"
oc get pvc -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/03-source-pvc.txt"
oc get route -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/04-source-routes.txt"
```

Expected source baseline:

```text
DevOps Control Plane pod Running 2/2
PostgreSQL pod Running 1/1
postgresql-data PVC Bound
Route present and healthy
```

---

## 10. Backup integrity validation

Set the backup directory and dump name:

```bash
export BACKUP_DIR="/tmp/dcp-10.2-postgresql-backup"
export BACKUP_DUMP="$BACKUP_DIR/dcp-postgresql-20260702-160033.dump"
export BACKUP_SHA256="$BACKUP_DUMP.sha256"
```

Verify files exist:

```bash
ls -lh "$BACKUP_DUMP" "$BACKUP_SHA256" | tee "$EVIDENCE_DIR/10-backup-files.txt"
```

Verify checksum:

```bash
cd "$BACKUP_DIR"
sha256sum -c "$(basename "$BACKUP_SHA256")" | tee "$EVIDENCE_DIR/11-backup-checksum-validation.txt"
```

Expected result:

```text
 dcp-postgresql-YYYYMMDD-HHMMSS.dump: OK
```

If checksum validation fails:

- do not use the backup;
- locate another valid backup;
- escalate to the service owner/platform owner;
- preserve the failed checksum evidence.

---

## 11. Restore validation target

Restore tests must use an isolated namespace by default:

```text
devops-control-plane-restore-test
```

Recommended isolated restore components:

```text
Namespace:  devops-control-plane-restore-test
Deployment: postgresql-restore-test
Service:    postgresql-restore-test
Database:   devops_control_plane_restore_test
PVC:        postgresql-restore-test-data
```

Do not restore into:

```text
devops-control-plane/postgresql
```

unless the situation is a formally approved production recovery.

---

## 12. Isolated restore procedure reference

The detailed PostgreSQL restore procedure is maintained in:

```text
docs/runbooks/postgresql-backup-restore.md
```

At a high level, the isolated restore test must:

1. create a temporary namespace;
2. create temporary PostgreSQL credentials without printing values;
3. create a PVC and PostgreSQL deployment;
4. wait for the restore PostgreSQL pod to become Ready;
5. copy the backup dump into the restore pod;
6. create the target database;
7. run `pg_restore` using PostgreSQL 16-compatible tooling;
8. execute validation queries;
9. remove the dump from the pod;
10. collect evidence;
11. delete the temporary namespace after review.

---

## 13. Restore validation queries

After restore, validate minimum dataset consistency:

```sql
select count(*) as applications_count from applications;
select count(*) as change_requests_count from change_requests;
select count(*) as change_events_count from change_events;
select count(*) as evidences_count from evidences;
```

For the current baseline, expected values are:

```text
applications_count = 0
change_requests_count = 1
change_events_count = 34
evidences_count = 18
```

The `applications` table being empty is expected for the current dataset. The application name is still recoverable from `change_requests.application_name`.

Validate the known change request:

```sql
select
  change_number,
  application_name,
  target_environment,
  status,
  runtime_status
from change_requests
where change_number = 'CHG-2026-0001';
```

Expected result:

```text
change_number = CHG-2026-0001
application_name = demo-go-color-app
target_environment = dev
status = draft
runtime_status = EvidenceCollected
```

---

## 14. Post-restore application validation

After a successful isolated database restore, the production runtime should remain untouched.

Validate source runtime again:

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

If the source runtime changed during a restore test, stop and investigate immediately. Restore validation should not impact the active namespace.

---

## 15. Cleanup procedure

After restore validation and evidence review, remove the temporary namespace:

```bash
oc delete namespace "$RESTORE_NS"
```

Verify deletion:

```bash
oc get namespace "$RESTORE_NS" 2>&1 | tee "$EVIDENCE_DIR/90-restore-namespace-delete-check.txt"
```

Expected result:

```text
Error from server (NotFound): namespaces "devops-control-plane-restore-test" not found
```

Do not delete the source namespace or source PVC.

---

## 16. DR success criteria

A DR validation is successful when all of the following are true:

- backup dump exists;
- checksum validation succeeds;
- TOC list can be generated or has already been generated with compatible PostgreSQL tooling;
- isolated restore PostgreSQL pod becomes Ready;
- `pg_restore` completes without errors;
- validation queries return expected counts;
- known change request `CHG-2026-0001` is present and consistent;
- restore dump is removed from the temporary pod;
- temporary namespace is deleted after validation;
- source runtime smoke test remains healthy;
- evidence package is complete.

---

## 17. DR failure criteria

A DR validation must be considered failed if any of the following occur:

- backup file is missing;
- checksum validation fails;
- `pg_restore --list` fails with compatible tooling;
- isolated PostgreSQL pod does not become Ready;
- `pg_restore` fails;
- expected tables are missing;
- validation counts are inconsistent without explanation;
- known change request is missing;
- source runtime is impacted by the restore test;
- temporary namespace cannot be cleaned up;
- Secret values are exposed in logs or evidence.

If Secret values are exposed, rotate the affected credentials immediately.

---

## 18. Incident recovery decision flow

During an actual incident, use this decision flow.

### 18.1 Confirm impact

Check whether the issue affects:

- only the UI/API runtime;
- only PostgreSQL;
- both application and PostgreSQL;
- external integrations;
- OpenShift cluster/platform.

Run quick health-check if possible:

```bash
unset EVIDENCE_DIR
./scripts/operability/health-check.sh
```

### 18.2 Identify last known good backup

Locate latest backup with:

- valid checksum;
- matching PostgreSQL major version compatibility;
- usable TOC list;
- evidence from previous restore validation if available.

### 18.3 Choose recovery target

Preferred sequence:

1. restore into isolated namespace;
2. validate data;
3. decide whether production data replacement is required;
4. execute production restore only with explicit approval.

### 18.4 Production restore approval

Before restoring into production, obtain explicit approval from the responsible owner and confirm:

- selected backup timestamp;
- expected data loss based on RPO;
- expected downtime based on RTO;
- rollback plan;
- communication plan;
- credential handling plan;
- evidence owner.

---

## 19. Evidence package

A complete DR evidence package should include:

```text
00-timestamp.txt
00-context.txt
01-source-pods.txt
02-source-deployments.txt
03-source-pvc.txt
04-source-routes.txt
10-backup-files.txt
11-backup-checksum-validation.txt
restore-toc-list.txt
restore-validation-plan.txt
restore-pod-status.txt
pg-restore-output.txt
validation-query-counts.txt
validation-query-known-change.txt
post-restore-smoke-test-summary.txt
90-restore-namespace-delete-check.txt
```

Before sharing externally, review all files and remove any accidental sensitive data.

---

## 20. Roles and responsibilities

### Service owner

- Approves RPO/RTO targets.
- Approves production restore decisions.
- Accepts or rejects data loss risk.

### Platform/OpenShift operator

- Provides namespace, storage and runtime support.
- Validates PVC and cluster health.
- Supports restore namespace creation and cleanup.

### DevOps Control Plane maintainer

- Executes application-level validation.
- Runs smoke tests.
- Validates data consistency after restore.
- Updates runbooks and evidence.

### Security owner

- Handles credential exposure events.
- Confirms token rotation requirements.
- Reviews evidence for sensitive content before external sharing.

---

## 21. Current limitations

The current DR baseline has the following limitations:

1. PostgreSQL currently runs as a single replica.
2. PostgreSQL image uses `latest`; production should pin immutable tags or digests.
3. Backup execution is currently manual unless scheduled externally.
4. Backup retention policy is not yet formalized.
5. Off-cluster backup storage is not yet formalized.
6. Production RPO/RTO have not yet been formally approved by a service owner.
7. Restore has been validated in an isolated namespace, not as a production replacement exercise.
8. GitLab, Argo CD and Tekton backup/restore are out of scope for this runbook.
9. Full OpenShift cluster DR is out of scope for this runbook.

---

## 22. Production recommendations

Before production promotion, consider implementing:

1. Scheduled PostgreSQL backups.
2. Off-cluster backup storage.
3. Backup encryption at rest.
4. Backup retention policy.
5. Periodic restore drills.
6. Timed RTO measurement.
7. Explicit RPO/RTO sign-off.
8. Immutable image references.
9. PostgreSQL HA evaluation.
10. Alerting for missed backups.
11. Alerting for failed restore validation jobs.
12. Evidence retention policy.
13. Separate DR procedure for GitLab, Argo CD and cluster-level components.

---

## 23. Recommended validation frequency

Suggested baseline:

```text
Backup creation:             at least daily for production-like usage
Checksum validation:         every backup
Restore validation:          monthly or after major schema/application changes
Smoke test:                  after every deployment/change
RPO/RTO review:              quarterly or after service criticality changes
Runbook review:              after every incident or DR exercise
```

For MVP/lab usage, a manual restore validation after each major milestone is acceptable.

---

## 24. Revision history

| Date | Phase | Description |
|---|---:|---|
| 2026-07-03 | 10.5 | Initial disaster recovery operational baseline for DevOps Control Plane. |
