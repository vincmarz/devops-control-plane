# PostgreSQL Backup and Restore Runbook

## Scope

This runbook documents the PostgreSQL backup and restore baseline for the DevOps Control Plane.

It covers:

- manual logical backup with `pg_dump` custom format;
- checksum generation and verification;
- dump content verification with `pg_restore -l`;
- restore validation plan;
- isolated restore test on a temporary PostgreSQL instance;
- cleanup and evidence collection;
- operational notes and known limitations.

This runbook intentionally does not define an automated production backup schedule yet. Automation, retention, external storage and alerting are covered by future operability phases.

## Safety rules

- Do not print database passwords.
- Do not print Secret values.
- Do not restore into the live DevOps Control Plane database during validation tests.
- Do not run destructive commands such as `drop database` or `drop schema` against the live database.
- Use a temporary isolated PostgreSQL instance for restore testing.
- Remove temporary dump files from pods after copying or validating them.
- Keep backup files and checksum files together.
- Treat dump files as sensitive operational artifacts.

## Current validated environment

```text
Namespace: devops-control-plane
PostgreSQL pod: postgresql-66f87b7545-rtxmx
PostgreSQL image: registry.redhat.io/rhel9/postgresql-16:latest
Database version: PostgreSQL 16.14
Backup file: dcp-postgresql-20260702-160033.dump
Backup directory on bastion: /tmp/dcp-10.2-postgresql-backup
Backup size: 21269 bytes, approximately 21K
Backup format: pg_dump custom format
Checksum SHA256: f35affbb14f235a1b3cc21613e1ed677db08d40f1116ed918fc9beb212e19eaa
```

## Important version note

The bastion currently provides:

```text
pg_restore (PostgreSQL) 10.23
```

The backup was created by PostgreSQL 16.14 tools:

```text
Dumped from database version: 16.14
Dumped by pg_dump version: 16.14
```

For this reason, the dump TOC list and restore validation were performed with `pg_restore` 16.14 inside the PostgreSQL pod instead of using the older bastion `pg_restore` 10.23.

## Phase 10.2.1 — Manual backup baseline

### Variables

```bash
DCP_NS="devops-control-plane"
PG_APP="postgresql"
BACKUP_DIR="/tmp/dcp-10.2-postgresql-backup"

rm -rf "$BACKUP_DIR"
mkdir -p "$BACKUP_DIR"

PG_POD="$(oc get pod -n "$DCP_NS" -l app=postgresql -o jsonpath='{.items[0].metadata.name}')"

echo "DCP_NS=${DCP_NS}"
echo "PG_APP=${PG_APP}"
echo "PG_POD=${PG_POD}"
echo "BACKUP_DIR=${BACKUP_DIR}"
```

### Verify PostgreSQL tools in the pod

```bash
oc exec -n "$DCP_NS" "$PG_POD" -- sh -c '
command -v pg_dump
command -v pg_restore
command -v psql
'
```

Expected result:

```text
/usr/bin/pg_dump
/usr/bin/pg_restore
/usr/bin/psql
```

### Verify required environment variables without printing values

```bash
oc exec -n "$DCP_NS" "$PG_POD" -- sh -c '
for k in POSTGRESQL_USER POSTGRESQL_DATABASE POSTGRESQL_PASSWORD; do
  if printenv "$k" >/dev/null 2>&1; then
    echo "$k present"
  else
    echo "$k missing"
  fi
done
'
```

Expected result:

```text
POSTGRESQL_USER present
POSTGRESQL_DATABASE present
POSTGRESQL_PASSWORD present
```

### Create a logical backup

```bash
BACKUP_NAME="dcp-postgresql-$(date +%Y%m%d-%H%M%S).dump"

echo "BACKUP_NAME=${BACKUP_NAME}"
```

```bash
oc exec -n "$DCP_NS" "$PG_POD" -- sh -c '
set -eu

BACKUP_FILE="/tmp/'"$BACKUP_NAME"'"

PGPASSWORD="${POSTGRESQL_PASSWORD}" pg_dump \
  -h 127.0.0.1 \
  -p 5432 \
  -U "${POSTGRESQL_USER}" \
  -d "${POSTGRESQL_DATABASE}" \
  -F c \
  -f "${BACKUP_FILE}"

wc -c "${BACKUP_FILE}"
'
```

Validated result from the baseline execution:

```text
21269 /tmp/dcp-postgresql-20260702-160033.dump
```

### Verify dump content without restoring

```bash
oc exec -n "$DCP_NS" "$PG_POD" -- sh -c '
set -eu

BACKUP_FILE="/tmp/'"$BACKUP_NAME"'"

pg_restore -l "${BACKUP_FILE}" | head -40
'
```

Validated content included:

```text
TABLE public applications
TABLE public change_events
TABLE public change_requests
TABLE public evidences
TABLE DATA public applications
TABLE DATA public change_events
TABLE DATA public change_requests
TABLE DATA public evidences
```

### Copy the backup to the bastion

```bash
oc cp "${DCP_NS}/${PG_POD}:/tmp/${BACKUP_NAME}" "${BACKUP_DIR}/${BACKUP_NAME}"
```

```bash
ls -lh "${BACKUP_DIR}/${BACKUP_NAME}"
wc -c "${BACKUP_DIR}/${BACKUP_NAME}"
```

### Generate checksum

```bash
sha256sum "${BACKUP_DIR}/${BACKUP_NAME}" > "${BACKUP_DIR}/${BACKUP_NAME}.sha256"
cat "${BACKUP_DIR}/${BACKUP_NAME}.sha256"
```

Validated checksum:

```text
f35affbb14f235a1b3cc21613e1ed677db08d40f1116ed918fc9beb212e19eaa
```

### Save backup metadata

```bash
cat > "${BACKUP_DIR}/backup-metadata.txt" <<EOF
Backup type: PostgreSQL logical backup
Format: pg_dump custom format
Namespace: ${DCP_NS}
Pod: ${PG_POD}
Backup file: ${BACKUP_NAME}
Created at: $(date --iso-8601=seconds)
Verification: pg_restore -l executed
Secrets printed: no
EOF
```

### Remove temporary dump from the PostgreSQL pod

```bash
oc exec -n "$DCP_NS" "$PG_POD" -- sh -c '
rm -f "/tmp/'"$BACKUP_NAME"'"
'
```

```bash
oc exec -n "$DCP_NS" "$PG_POD" -- sh -c '
if test -f "/tmp/'"$BACKUP_NAME"'"; then
  echo "backup still present in pod"
else
  echo "backup removed from pod"
fi
'
```

Expected result:

```text
backup removed from pod
```

## Phase 10.2.2 — Restore validation plan

### Verify local backup checksum

```bash
BACKUP_DIR="/tmp/dcp-10.2-postgresql-backup"
BACKUP_NAME="dcp-postgresql-20260702-160033.dump"

sha256sum -c "${BACKUP_DIR}/${BACKUP_NAME}.sha256"
```

Validated result:

```text
/tmp/dcp-10.2-postgresql-backup/dcp-postgresql-20260702-160033.dump: OK
```

### Generate TOC list using PostgreSQL 16.14 tools in the pod

Because the bastion `pg_restore` is version 10.23, the TOC list must be generated with `pg_restore` 16.14 from the PostgreSQL pod.

```bash
DCP_NS="devops-control-plane"
PG_POD="postgresql-66f87b7545-rtxmx"
BACKUP_DIR="/tmp/dcp-10.2-postgresql-backup"
BACKUP_NAME="dcp-postgresql-20260702-160033.dump"

oc cp "${BACKUP_DIR}/${BACKUP_NAME}" "${DCP_NS}/${PG_POD}:/tmp/${BACKUP_NAME}"
```

```bash
oc exec -n "$DCP_NS" "$PG_POD" -- sh -c '
set -eu

BACKUP_FILE="/tmp/'"$BACKUP_NAME"'"
TOC_FILE="/tmp/restore-toc-list.txt"

pg_restore -l "${BACKUP_FILE}" > "${TOC_FILE}"

wc -l "${TOC_FILE}"
head -80 "${TOC_FILE}"
'
```

Validated result:

```text
54 /tmp/restore-toc-list.txt
```

### Copy TOC list to bastion and cleanup pod

```bash
oc cp "${DCP_NS}/${PG_POD}:/tmp/restore-toc-list.txt" "${BACKUP_DIR}/restore-toc-list.txt"
```

```bash
oc exec -n "$DCP_NS" "$PG_POD" -- sh -c '
rm -f "/tmp/'"$BACKUP_NAME"'" "/tmp/restore-toc-list.txt"
'
```

```bash
oc exec -n "$DCP_NS" "$PG_POD" -- sh -c '
if test -f "/tmp/'"$BACKUP_NAME"'"; then
  echo "backup still present in pod"
else
  echo "backup removed from pod"
fi

if test -f "/tmp/restore-toc-list.txt"; then
  echo "toc still present in pod"
else
  echo "toc removed from pod"
fi
'
```

Expected result:

```text
backup removed from pod
toc removed from pod
```

### Restore validation plan target

Recommended isolated target:

```text
Temporary namespace: devops-control-plane-restore-test
Temporary PostgreSQL deployment: postgresql-restore-test
Temporary database: devops_control_plane_restore_test
```

Validation queries:

```sql
select count(*) as applications_count from applications;
select count(*) as change_requests_count from change_requests;
select count(*) as change_events_count from change_events;
select count(*) as evidences_count from evidences;
select change_number, application_name, target_environment, status, runtime_status
from change_requests
where change_number = 'CHG-2026-0001';
```

## Phase 10.2.3 — Isolated restore test

### Restore test target

Validated isolated restore environment:

```text
Namespace: devops-control-plane-restore-test
Deployment: postgresql-restore-test
Pod: postgresql-restore-test-6d567c8d99-q2cvm
Database: devops_control_plane_restore_test
Backup: dcp-postgresql-20260702-160033.dump
```

### Restore execution result

The dump was copied into the restore pod and restored with `pg_restore` into the temporary database.

Validated dump size inside restore pod:

```text
21269 /tmp/dcp-postgresql-20260702-160033.dump
```

The `pg_restore` command completed without visible errors.

### Restore validation output

Validated query output:

```text
applications_count: 0
change_requests_count: 1
change_events_count: 34
evidences_count: 18

CHG-2026-0001 | demo-go-color-app | dev | draft | EvidenceCollected
```

### Interpretation of applications_count equals zero

`applications_count=0` is not considered a restore failure.

The dump contains the `applications` table and its table data entry. The restored content reflects the source dataset at backup time. The application `demo-go-color-app` is recoverable from `change_requests.application_name` for `CHG-2026-0001`.

### Restore evidence directory

Evidence from the restore test was saved in:

```text
/tmp/dcp-10.2.3-postgresql-restore-test
```

Files:

```text
namespace.yaml
postgresql-restore-deployment.yaml
postgresql-restore-pod.yaml
postgresql-restore-service.yaml
restore-validation-queries.txt
```

### Temporary dump cleanup

The copied dump was removed from the restore pod.

Expected and validated result:

```text
backup removed from restore pod
```

## Cleanup temporary restore namespace

If no further inspection is required, remove the restore namespace:

```bash
oc delete project devops-control-plane-restore-test
```

Verify deletion:

```bash
oc get namespace devops-control-plane-restore-test
```

Expected result after termination completes:

```text
Error from server (NotFound): namespaces "devops-control-plane-restore-test" not found
```

## Success criteria

The backup and restore baseline is successful if:

- `pg_dump` completes successfully.
- The dump file is greater than zero bytes.
- `pg_restore -l` succeeds with PostgreSQL 16.14 tools.
- The dump is copied to the bastion.
- A SHA256 checksum is generated and verified.
- Restore is executed only in an isolated PostgreSQL instance.
- `pg_restore` completes without errors in the isolated instance.
- Main tables are present after restore.
- Key data is recoverable after restore.
- `CHG-2026-0001` is present after restore.
- `demo-go-color-app` is recoverable from restored `change_requests.application_name`.
- No production Secret values are printed.
- The live database is not modified.
- Temporary dump files are removed from pods.

## Failure criteria

The backup and restore baseline is not acceptable if:

- `pg_dump` fails.
- The dump file is empty.
- The checksum validation fails.
- `pg_restore -l` fails with a compatible PostgreSQL version.
- Restore is attempted against the live database.
- Required tables are missing after restore.
- `CHG-2026-0001` cannot be recovered.
- Secret or password values are printed.
- Temporary dump files remain in pods after the procedure.

## Current limitations and next improvements

This runbook documents a manually validated backup and isolated restore procedure. It does not yet provide:

- automated scheduled backups;
- retention policy;
- external durable backup storage;
- encrypted backup archive handling;
- backup monitoring and alerting;
- periodic restore drill automation;
- RPO/RTO targets;
- production-grade HA PostgreSQL topology.

These topics belong to future Operability phases.
