# DevOps Control Plane — Maintenance Operations Runbook

## Document metadata

- **Project:** DevOps Control Plane
- **Phase:** 10.6 — Maintenance operations runbook
- **Scope:** Routine and controlled maintenance operations for the DevOps Control Plane runtime on OpenShift
- **Audience:** Platform engineers, DevOps engineers, application operators, service owners and onboarding readers
- **Execution mode:** Controlled operational changes with explicit pre-checks, post-checks and rollback notes
- **Security posture:** Do not print Secret values, tokens, database passwords or kubeconfig credentials
- **Related runbooks and scripts:**
  - `docs/runbooks/operability-health-check.md`
  - `docs/runbooks/postgresql-backup-restore.md`
  - `docs/runbooks/disaster-recovery.md`
  - `docs/runbooks/secrets-rotation.md`
  - `docs/runbooks/authn-authz.md`
  - `scripts/operability/health-check.sh`

---

## 1. Purpose

This runbook defines the standard maintenance operations for the DevOps Control Plane deployed on OpenShift.

It covers safe operational procedures for:

1. pre-maintenance checks;
2. application rollout and image update validation;
3. rollback strategy;
4. post-maintenance smoke testing;
5. Secret and token rotation coordination;
6. trust bundle maintenance;
7. AuthN/AuthZ validation;
8. RBAC validation;
9. NetworkPolicy validation;
10. PostgreSQL maintenance considerations;
11. cleanup of temporary resources;
12. evidence collection and maintenance closure.

The runbook is designed to be used after the project has reached a working MVP and advanced security/operability baseline.

---

## 2. Operating principles

Every maintenance activity must follow these principles:

- make the smallest safe change possible;
- run pre-checks before touching runtime resources;
- collect evidence before and after the change;
- do not print Secret values;
- prefer declarative repository-aligned changes;
- avoid live patches unless they are explicitly part of a controlled remediation procedure;
- validate with the automated smoke test after every change;
- keep rollback instructions ready before applying a change;
- document deviations from the standard procedure.

---

## 3. Maintenance types

This runbook covers the following maintenance categories.

### 3.1 Routine validation

Examples:

- daily or weekly runtime health-check;
- post-deployment smoke test;
- validation after OpenShift/node maintenance;
- validation after external dependency maintenance.

### 3.2 Application maintenance

Examples:

- deploy a new DevOps Control Plane image;
- change ConfigMap values;
- update health probes;
- update resource requests/limits;
- update OAuth Proxy sidecar settings.

### 3.3 Security maintenance

Examples:

- rotate Argo CD token;
- rotate GitLab token;
- rotate database credentials;
- validate Secret keys without printing values;
- update trust bundle;
- validate AuthN/AuthZ role mapping.

### 3.4 Platform-facing maintenance

Examples:

- validate RBAC after namespace changes;
- validate NetworkPolicy behavior;
- validate Route/Service wiring;
- validate PostgreSQL PVC and backup posture.

---

## 4. Safety rules

Do not run destructive commands unless the maintenance change explicitly requires them and rollback is defined.

Avoid these commands during routine health-checks:

```bash
oc delete
oc replace
oc edit
oc patch
oc apply
oc rollout restart
oc rollout undo
```

These commands may be used during approved maintenance windows, but only with appropriate evidence and rollback notes:

```bash
oc apply
oc patch
oc rollout status
oc rollout undo
oc set image
oc scale
```

Never print sensitive values:

```bash
oc get secret ... -o yaml
oc get secret ... -o jsonpath='{.data...}' | base64 -d
env | grep TOKEN
env | grep PASSWORD
printenv
```

Allowed Secret inventory pattern:

```bash
oc get secret devops-control-plane-secrets -n devops-control-plane -o json \
  | jq -r '.data | keys[]' \
  | sort
```

Allowed Secret key length inventory:

```bash
oc get secret devops-control-plane-secrets -n devops-control-plane -o json \
  | jq -r '.data | to_entries[] | .key + " base64_length=" + (.value | length | tostring)' \
  | sort
```

---

## 5. Standard variables

Set these variables at the beginning of every maintenance session:

```bash
export DCP_NS="devops-control-plane"
export APP_DEPLOY="devops-control-plane"
export PG_DEPLOY="postgresql"
export APP_LABEL="app=devops-control-plane"
export PG_LABEL="app=postgresql"
export EVIDENCE_DIR="/tmp/dcp-maintenance-$(date +%Y%m%d-%H%M%S)"

mkdir -p "$EVIDENCE_DIR"

date -Is | tee "$EVIDENCE_DIR/00-timestamp.txt"

echo "DCP_NS=$DCP_NS" | tee "$EVIDENCE_DIR/00-context.txt"
echo "APP_DEPLOY=$APP_DEPLOY" | tee -a "$EVIDENCE_DIR/00-context.txt"
echo "PG_DEPLOY=$PG_DEPLOY" | tee -a "$EVIDENCE_DIR/00-context.txt"
echo "APP_LABEL=$APP_LABEL" | tee -a "$EVIDENCE_DIR/00-context.txt"
echo "PG_LABEL=$PG_LABEL" | tee -a "$EVIDENCE_DIR/00-context.txt"
echo "EVIDENCE_DIR=$EVIDENCE_DIR" | tee -a "$EVIDENCE_DIR/00-context.txt"
```

---

## 6. Pre-maintenance checklist

Before making any change, run the automated smoke test:

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

Expected baseline:

```text
PASS=35
WARN=0
FAIL=0
```

Record the generated evidence directory in the maintenance notes.

Then collect static inventory:

```bash
oc get deploy -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/01-pre-deployments.txt"
oc get pods -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/02-pre-pods.txt"
oc get svc -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/03-pre-services.txt"
oc get route -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/04-pre-routes.txt"
oc get pvc -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/05-pre-pvc.txt"
oc get events -n "$DCP_NS" --sort-by=.lastTimestamp | tail -n 80 | tee "$EVIDENCE_DIR/06-pre-events-tail.txt"
```

If the pre-maintenance smoke test fails, stop and troubleshoot before applying changes.

---

## 7. Application image update

### 7.1 Recommended approach

The preferred approach is repository-aligned deployment through the normal build/push and manifest update path.

The runtime deployment should be updated only after:

- code changes are committed;
- tests pass;
- image is built and pushed;
- manifest or deploy reference is aligned;
- rollback image reference is known.

### 7.2 Capture current image

```bash
oc get deployment "$APP_DEPLOY" -n "$DCP_NS" \
  -o jsonpath='{range .spec.template.spec.containers[*]}{.name}{"="}{.image}{"\n"}{end}' \
  | tee "$EVIDENCE_DIR/10-current-images.txt"
```

Expected containers:

```text
oauth-proxy=registry.redhat.io/openshift4/ose-oauth-proxy:latest
devops-control-plane=image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:<tag>
```

Production recommendation:

- avoid `latest` for production;
- prefer immutable tags or digest-pinned references.

### 7.3 Update image using `oc set image`

Use this only when a live runtime image update is approved.

```bash
export NEW_IMAGE="image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:<new-tag>"

oc set image deployment/"$APP_DEPLOY" \
  -n "$DCP_NS" \
  devops-control-plane="$NEW_IMAGE" \
  | tee "$EVIDENCE_DIR/11-set-image-output.txt"
```

Track rollout:

```bash
oc rollout status deployment/"$APP_DEPLOY" -n "$DCP_NS" \
  | tee "$EVIDENCE_DIR/12-rollout-status.txt"
```

Capture post-rollout pod:

```bash
oc get pods -n "$DCP_NS" -l "$APP_LABEL" -o wide \
  | tee "$EVIDENCE_DIR/13-post-rollout-pods.txt"
```

### 7.4 Post-image-update validation

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

Expected:

```text
PASS=35
WARN=0
FAIL=0
```

If this fails, proceed to rollback.

---

## 8. Application rollback

### 8.1 Rollback using previous ReplicaSet revision

Inspect rollout history:

```bash
oc rollout history deployment/"$APP_DEPLOY" -n "$DCP_NS" \
  | tee "$EVIDENCE_DIR/20-rollout-history.txt"
```

Rollback to previous revision:

```bash
oc rollout undo deployment/"$APP_DEPLOY" -n "$DCP_NS" \
  | tee "$EVIDENCE_DIR/21-rollout-undo.txt"
```

Track rollback:

```bash
oc rollout status deployment/"$APP_DEPLOY" -n "$DCP_NS" \
  | tee "$EVIDENCE_DIR/22-rollback-status.txt"
```

### 8.2 Rollback using known image

If a known-good image is documented, use:

```bash
export ROLLBACK_IMAGE="image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:<known-good-tag>"

oc set image deployment/"$APP_DEPLOY" \
  -n "$DCP_NS" \
  devops-control-plane="$ROLLBACK_IMAGE" \
  | tee "$EVIDENCE_DIR/23-rollback-set-image.txt"

oc rollout status deployment/"$APP_DEPLOY" -n "$DCP_NS" \
  | tee "$EVIDENCE_DIR/24-rollback-set-image-status.txt"
```

### 8.3 Post-rollback validation

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

Rollback is successful only if the smoke test returns:

```text
FAIL=0
```

---

## 9. ConfigMap maintenance

### 9.1 Capture current sanitized configuration

```bash
oc get configmap devops-control-plane-config -n "$DCP_NS" -o json \
  | jq -r '.data | to_entries[] | select(.key | test("TOKEN|PASSWORD|SECRET|DATABASE_URL") | not) | .key + "=" + .value' \
  | sort \
  | tee "$EVIDENCE_DIR/30-configmap-sanitized-before.txt"
```

### 9.2 Preferred update method

Preferred path:

1. update repository manifest;
2. validate diff;
3. commit and push;
4. apply using the approved deployment process.

Avoid ad-hoc runtime changes unless a controlled remediation requires it.

### 9.3 Runtime patch only when approved

Example for a non-sensitive flag:

```bash
oc patch configmap devops-control-plane-config \
  -n "$DCP_NS" \
  --type merge \
  -p '{"data":{"LOG_LEVEL":"info"}}' \
  | tee "$EVIDENCE_DIR/31-configmap-patch-output.txt"
```

If the application loads config only at startup, a rollout may be required:

```bash
oc rollout restart deployment/"$APP_DEPLOY" -n "$DCP_NS" \
  | tee "$EVIDENCE_DIR/32-rollout-restart-after-configmap.txt"

oc rollout status deployment/"$APP_DEPLOY" -n "$DCP_NS" \
  | tee "$EVIDENCE_DIR/33-rollout-status-after-configmap.txt"
```

### 9.4 Post-ConfigMap validation

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

---

## 10. Secret and token maintenance

Secret and token rotation must follow:

```text
docs/runbooks/secrets-rotation.md
```

This maintenance runbook only defines the coordination steps.

### 10.1 Secret inventory without values

```bash
oc get secret devops-control-plane-secrets -n "$DCP_NS" -o json \
  | jq -r '.data | keys[]' \
  | sort \
  | tee "$EVIDENCE_DIR/40-secret-keys-before.txt"

oc get secret devops-control-plane-secrets -n "$DCP_NS" -o json \
  | jq -r '.data | to_entries[] | .key + " base64_length=" + (.value | length | tostring)' \
  | sort \
  | tee "$EVIDENCE_DIR/41-secret-key-lengths-before.txt"
```

Expected keys:

```text
ARGOCD_AUTH_TOKEN
DATABASE_URL
GITLAB_TOKEN
```

`KUBERNETES_TOKEN` should not be required because the application uses the ServiceAccount token fallback.

### 10.2 Rotation principles

- Generate replacement tokens outside shared shell history where possible.
- Do not paste token values into tickets or chat.
- Load new values into the Secret without printing them.
- Restart/rollout the application only if required.
- Validate with smoke test.
- Destroy temporary token files securely.

### 10.3 Post-rotation validation

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

Additional functional checks may be required:

- Argo CD application read;
- GitLab branch/MR workflow;
- Tekton validation workflow;
- Kubernetes/OpenShift evidence collection.

---

## 11. Trust bundle maintenance

The current baseline uses an app-dedicated trust bundle:

```text
dcp-app-trust-bundle
```

mounted at:

```text
/etc/dcp-trust/ca-bundle.crt
```

Configuration references:

```text
ARGOCD_CA_FILE=/etc/dcp-trust/ca-bundle.crt
GITLAB_CA_FILE=/etc/dcp-trust/ca-bundle.crt
ARGOCD_INSECURE_TLS=false
GITLAB_INSECURE_TLS=false
KUBERNETES_INSECURE_TLS=false
```

### 11.1 Inspect trust bundle metadata without printing unrelated Secret values

```bash
oc get configmap dcp-app-trust-bundle -n "$DCP_NS" -o yaml \
  | tee "$EVIDENCE_DIR/50-dcp-app-trust-bundle.yaml"
```

### 11.2 Validate mounted file in the application pod

```bash
DCP_POD="$(oc get pod -n "$DCP_NS" -l "$APP_LABEL" -o jsonpath='{.items[0].metadata.name}')"

oc exec -n "$DCP_NS" "$DCP_POD" -c devops-control-plane -- \
  sh -c 'ls -l /etc/dcp-trust/ca-bundle.crt && wc -c /etc/dcp-trust/ca-bundle.crt' \
  | tee "$EVIDENCE_DIR/51-mounted-trust-bundle-file.txt"
```

### 11.3 Post-trust-bundle validation

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

If Argo CD or GitLab fails with `x509`, review the trust bundle and the current Route certificate chain.

---

## 12. OAuth Proxy maintenance

The current runtime uses OAuth Proxy as sidecar.

Important baseline arguments:

```text
--provider=openshift
--openshift-service-account=devops-control-plane-oauth-proxy
--https-address=:8443
--upstream=http://localhost:8080
--skip-auth-regex=^/readyz$
--skip-auth-regex=^/livez$
--pass-user-headers=true
--pass-user-bearer-token=false
--pass-access-token=false
--pass-basic-auth=false
```

### 12.1 Inspect current sidecar configuration

```bash
oc get deployment "$APP_DEPLOY" -n "$DCP_NS" -o json \
  | jq -r '.spec.template.spec.containers[] | select(.name=="oauth-proxy") | .args[]' \
  | tee "$EVIDENCE_DIR/60-oauth-proxy-args.txt"
```

### 12.2 Validate health policy

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

Expected:

```text
route readyz HTTP 200
route livez HTTP 200
route api changes anonymous HTTP 403
route ui dashboard anonymous HTTP 403
```

If health endpoints return `403`, verify both OAuth Proxy skip-auth regex and backend health middleware bypass.

---

## 13. AuthN/AuthZ maintenance

AuthN/AuthZ is based on trusted headers behind OpenShift OAuth Proxy and optional OpenShift group lookup.

Important configuration:

```text
AUTH_ENABLED=true
AUTH_HEADER_USER=X-Forwarded-User
AUTH_HEADER_ALT_USER=X-Auth-Request-User
AUTH_HEADER_GROUPS=X-Forwarded-Groups
AUTH_OPENSHIFT_GROUP_LOOKUP_ENABLED=true
AUTH_GROUP_VIEWER=devops-control-plane-viewers
AUTH_GROUP_OPERATOR=devops-control-plane-operators
AUTH_GROUP_APPROVER=devops-control-plane-approvers
AUTH_GROUP_ADMIN=devops-control-plane-admins
```

### 13.1 Inspect Auth configuration

```bash
oc get configmap devops-control-plane-config -n "$DCP_NS" -o json \
  | jq -r '.data | to_entries[] | select(.key | startswith("AUTH_")) | .key + "=" + .value' \
  | sort \
  | tee "$EVIDENCE_DIR/70-auth-config.txt"
```

### 13.2 Validate OpenShift groups

```bash
oc get group devops-control-plane-admins -o yaml 2>/dev/null \
  | tee "$EVIDENCE_DIR/71-group-admins.yaml" || true

oc get group devops-control-plane-operators -o yaml 2>/dev/null \
  | tee "$EVIDENCE_DIR/72-group-operators.yaml" || true

oc get group devops-control-plane-approvers -o yaml 2>/dev/null \
  | tee "$EVIDENCE_DIR/73-group-approvers.yaml" || true

oc get group devops-control-plane-viewers -o yaml 2>/dev/null \
  | tee "$EVIDENCE_DIR/74-group-viewers.yaml" || true
```

### 13.3 Post-Auth validation

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

For detailed role matrix validation, use the AuthN/AuthZ runbook.

---

## 14. RBAC maintenance

The runtime ServiceAccount is:

```text
system:serviceaccount:devops-control-plane:devops-control-plane-oauth-proxy
```

It requires least-privilege access for:

- Tekton PipelineRuns/TaskRuns in `devops-ci-demo`;
- Kubernetes/OpenShift runtime evidence collection;
- OpenShift group lookup through a cluster-scoped group-reader role.

### 14.1 Validate RBAC inventory

```bash
oc get rolebinding -n "$DCP_NS" -o wide \
  | tee "$EVIDENCE_DIR/80-rolebindings-dcp.txt"

oc get rolebinding -n devops-ci-demo -o wide \
  | tee "$EVIDENCE_DIR/81-rolebindings-devops-ci-demo.txt"

oc get clusterrolebinding \
  | grep -E 'devops-control-plane|oauth-proxy|group-reader|auth-delegator' \
  | tee "$EVIDENCE_DIR/82-clusterrolebindings-filtered.txt" || true
```

### 14.2 Validate `can-i`

```bash
DCP_SA="system:serviceaccount:devops-control-plane:devops-control-plane-oauth-proxy"

oc auth can-i create pipelineruns.tekton.dev -n devops-ci-demo --as "$DCP_SA" \
  | tee "$EVIDENCE_DIR/83-can-i-create-pipelineruns.txt"

oc auth can-i get deployments -n devops-ci-demo --as "$DCP_SA" \
  | tee "$EVIDENCE_DIR/84-can-i-get-deployments.txt"

oc auth can-i list pods -n devops-ci-demo --as "$DCP_SA" \
  | tee "$EVIDENCE_DIR/85-can-i-list-pods.txt"

oc auth can-i get secrets -n "$DCP_NS" --as "$DCP_SA" \
  | tee "$EVIDENCE_DIR/86-can-i-get-secrets-dcp.txt"
```

Expected:

```text
create pipelineruns.tekton.dev: yes
get deployments: yes
list pods: yes
get secrets in devops-control-plane: no
```

---

## 15. NetworkPolicy maintenance

The current safe baseline includes:

```text
postgresql-ingress-from-devops-control-plane
```

This policy allows ingress to PostgreSQL only from DevOps Control Plane pods on TCP 5432.

### 15.1 Inspect policies

```bash
oc get networkpolicy -n "$DCP_NS" -o wide \
  | tee "$EVIDENCE_DIR/90-networkpolicies.txt"

oc get networkpolicy postgresql-ingress-from-devops-control-plane -n "$DCP_NS" -o yaml \
  | tee "$EVIDENCE_DIR/91-postgresql-networkpolicy.yaml"
```

### 15.2 Validate application connectivity

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

If PostgreSQL connectivity fails after a NetworkPolicy change, rollback the policy or restore the known-good manifest.

Production note:

- do not introduce deny-all ingress/egress policies without a dedicated design and test plan;
- validate DNS, API server, GitLab, Argo CD and Tekton connectivity before enforcing strict egress.

---

## 16. PostgreSQL maintenance

PostgreSQL operations are covered primarily by:

```text
docs/runbooks/postgresql-backup-restore.md
docs/runbooks/disaster-recovery.md
```

### 16.1 Pre-maintenance backup

Before risky database maintenance, create and validate a backup.

Minimum requirements:

- dump file exists;
- checksum generated;
- checksum verified;
- TOC list generated with PostgreSQL-compatible tooling;
- restore plan documented.

### 16.2 Read-only database validation

```bash
PG_POD="$(oc get pod -n "$DCP_NS" -l "$PG_LABEL" -o jsonpath='{.items[0].metadata.name}')"

oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select now() as db_time, current_database() as database_name, current_user as user_name;"' \
  | tee "$EVIDENCE_DIR/100-postgresql-connectivity.txt"

oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as change_requests_count from change_requests;"' \
  | tee "$EVIDENCE_DIR/101-postgresql-change-requests-count.txt"

oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as change_events_count from change_events;"' \
  | tee "$EVIDENCE_DIR/102-postgresql-change-events-count.txt"

oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as evidences_count from evidences;"' \
  | tee "$EVIDENCE_DIR/103-postgresql-evidences-count.txt"
```

---

## 17. External integration maintenance

### 17.1 Argo CD

After Argo CD token, RBAC or TLS maintenance, validate:

- token is present as Secret key only;
- token value is not printed;
- TLS strict mode remains enabled;
- application read works through the DevOps Control Plane.

Relevant configuration:

```text
ARGOCD_BASE_URL
ARGOCD_CA_FILE
ARGOCD_INSECURE_TLS=false
ARGOCD_AUTH_TOKEN
```

### 17.2 GitLab

After GitLab token, TLS or project permission maintenance, validate:

- token is present as Secret key only;
- TLS strict mode remains enabled;
- GitLab project ID is correct;
- branch/MR workflow still works in a controlled test.

Relevant configuration:

```text
GITLAB_BASE_URL
GITLAB_CA_FILE
GITLAB_INSECURE_TLS=false
GITLAB_PROJECT_ID
GITLAB_TOKEN
```

### 17.3 Tekton

After Tekton pipeline, namespace or RBAC maintenance, validate:

- pipeline name is correct;
- workspace PVC exists;
- runtime ServiceAccount can create PipelineRuns;
- validation pipeline still succeeds for a safe revision.

Relevant configuration:

```text
TEKTON_NAMESPACE
TEKTON_PIPELINE_NAME
TEKTON_WORKSPACE_PVC
TEKTON_SERVICE_ACCOUNT
TEKTON_VALIDATION_PATH
```

---

## 18. Temporary resource cleanup

After maintenance, remove temporary local files and namespaces created for testing.

### 18.1 Local temporary files

Do not remove evidence directories until they have been reviewed or archived.

Remove temporary token files only after rotation validation:

```bash
shred -u /tmp/<temporary-token-file> 2>/dev/null || rm -f /tmp/<temporary-token-file>
```

### 18.2 Temporary namespaces

For restore tests:

```bash
oc delete namespace devops-control-plane-restore-test
```

Verify deletion:

```bash
oc get namespace devops-control-plane-restore-test 2>&1 \
  | tee "$EVIDENCE_DIR/110-restore-test-namespace-delete-check.txt"
```

---

## 19. Post-maintenance validation

Always run:

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

Expected:

```text
PASS=35
WARN=0
FAIL=0
```

Then collect final inventory:

```bash
oc get deploy -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/120-post-deployments.txt"
oc get pods -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/121-post-pods.txt"
oc get svc -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/122-post-services.txt"
oc get route -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/123-post-routes.txt"
oc get pvc -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/124-post-pvc.txt"
oc get events -n "$DCP_NS" --sort-by=.lastTimestamp | tail -n 80 | tee "$EVIDENCE_DIR/125-post-events-tail.txt"
```

---

## 20. Maintenance closure criteria

A maintenance activity can be closed when:

- planned changes were applied or explicitly skipped;
- rollback was not required, or rollback was completed successfully;
- post-maintenance smoke test completed with `FAIL=0`;
- no unexpected warning/error events are present;
- runtime pods are Ready;
- PostgreSQL connectivity is OK;
- anonymous Route access is still blocked for API/UI;
- evidence directory is documented;
- repository is aligned if manifests/scripts/runbooks changed;
- working tree is clean after commit/push where applicable.

---

## 21. Incident escalation criteria

Escalate if any of the following occur:

- application cannot become Ready after rollback;
- PostgreSQL is unavailable;
- PVC is not Bound;
- Route health endpoints fail after rollback;
- anonymous API/UI access returns HTTP 200;
- Secret value was exposed;
- token rotation fails and previous token is no longer usable;
- NetworkPolicy blocks required traffic and rollback fails;
- restore validation fails during a DR-related maintenance window.

---

## 22. Evidence package

A standard maintenance evidence package should include:

```text
00-timestamp.txt
00-context.txt
01-pre-deployments.txt
02-pre-pods.txt
03-pre-services.txt
04-pre-routes.txt
05-pre-pvc.txt
06-pre-events-tail.txt
10-current-images.txt
12-rollout-status.txt
13-post-rollout-pods.txt
30-configmap-sanitized-before.txt
40-secret-keys-before.txt
41-secret-key-lengths-before.txt
70-auth-config.txt
80-rolebindings-dcp.txt
81-rolebindings-devops-ci-demo.txt
82-clusterrolebindings-filtered.txt
83-can-i-create-pipelineruns.txt
84-can-i-get-deployments.txt
85-can-i-list-pods.txt
86-can-i-get-secrets-dcp.txt
90-networkpolicies.txt
100-postgresql-connectivity.txt
101-postgresql-change-requests-count.txt
102-postgresql-change-events-count.txt
103-postgresql-evidences-count.txt
120-post-deployments.txt
121-post-pods.txt
122-post-services.txt
123-post-routes.txt
124-post-pvc.txt
125-post-events-tail.txt
```

Also attach the generated smoke-test evidence directory when applicable.

Before sharing externally, review all files for accidental sensitive data.

---

## 23. Known limitations

Current baseline limitations:

1. Some image references still use `latest`.
2. PostgreSQL is single-replica.
3. Backup scheduling and retention are not yet automated in this repository baseline.
4. Strict deny-all NetworkPolicy design is intentionally deferred.
5. Production RPO/RTO require formal service owner approval.
6. External systems such as GitLab and Argo CD require their own maintenance and DR procedures.

---

## 24. Recommended maintenance frequency

Suggested baseline:

```text
Smoke test:                    after every change and at least weekly
Secret inventory:              monthly or before major releases
Token rotation:                according to organization security policy
RBAC validation:               after role/group/namespace changes
NetworkPolicy validation:      after any policy change
Backup validation:             every backup
Restore drill:                 monthly or after major schema changes
Trust bundle validation:       after certificate or Route CA changes
Runbook review:                after every incident or major maintenance
```

---

## 25. Revision history

| Date | Phase | Description |
|---|---:|---|
| 2026-07-03 | 10.6 | Initial maintenance operations runbook for routine and controlled DevOps Control Plane maintenance. |
