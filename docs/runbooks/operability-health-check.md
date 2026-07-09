# DevOps Control Plane — Operability Health-Check Runbook

## Document metadata

- **Project:** DevOps Control Plane
- **Phase:** 10.3.2 — Operability health-check runbook
- **Scope:** Runtime health-check, troubleshooting and evidence collection for the DevOps Control Plane deployed on OpenShift
- **Audience:** Platform engineers, DevOps engineers, application operators and onboarding readers
- **Execution mode:** Read-only unless explicitly stated otherwise
- **Security posture:** Do not print Secret values, tokens or passwords
- **Last baseline reference:** Phase 10.3.1 — Observability baseline inventory read-only

---

## 1. Purpose

This runbook defines a repeatable operational health-check procedure for the DevOps Control Plane runtime on OpenShift.

It is intended to answer the following questions:

1. Is the DevOps Control Plane namespace available?
2. Are the application and PostgreSQL pods running and ready?
3. Are there recent warning/error events?
4. Are CPU and memory usage within the expected baseline?
5. Are health endpoints working?
6. Is anonymous access correctly blocked?
7. Is authenticated/authorized access working when trusted headers are provided directly to the backend for controlled testing?
8. Is PostgreSQL reachable and consistent with the expected dataset?
9. Are NetworkPolicy and RBAC objects present?
10. Which evidences should be collected before escalation?

This document is not a deployment procedure and does not replace backup/restore or disaster recovery runbooks.

---

## 2. Safety rules

Follow these rules during every execution:

- Do not print Secret values.
- Do not decode Secret values unless explicitly required by a controlled rotation procedure.
- Do not run destructive commands.
- Do not patch, edit, delete or apply resources during a health-check.
- Save command outputs into an evidence directory.
- Prefer short, targeted outputs when sharing results.
- If a token/password is accidentally printed, treat it as exposed and rotate it.

Commands intentionally avoided in this runbook:

```bash
oc apply
oc patch
oc edit
oc delete
oc replace
oc rollout restart
oc rollout undo
```

---

## 3. Runtime assumptions

The current baseline assumes:

- Namespace: `devops-control-plane`
- Application label: `app=devops-control-plane`
- PostgreSQL label: `app=postgresql`
- DevOps Control Plane Service exposed on port `8443`
- PostgreSQL Service exposed on port `5432`
- Route termination: `reencrypt/Redirect`
- OAuth Proxy enabled as sidecar
- Backend Auth middleware enabled
- OpenShift group lookup enabled
- Health endpoints `/readyz` and `/livez` publicly reachable through the OAuth Proxy skip-auth rules
- API/UI anonymous access blocked through the Route
- PostgreSQL data persisted on PVC `postgresql-data`

---

## 4. Expected healthy baseline

A healthy runtime should look similar to the following baseline:

```text
Namespace devops-control-plane: Active
DevOps Control Plane pod: Running, 2/2, restarts 0
PostgreSQL pod: Running, 1/1, restarts 0
Route /readyz: HTTP 200
Route /livez: HTTP 200
Route /api/v1/changes anonymous: HTTP 403
Route /ui/dashboard anonymous: HTTP 403
Backend direct /readyz via port-forward: HTTP 200
Backend direct /livez via port-forward: HTTP 200
Backend direct anonymous API: HTTP 401 or 403 depending on middleware/proxy path
Backend direct admin-header API test: HTTP 200
Namespace events: no relevant warning/error events
Application logs: no panic/fatal/error loops
OAuth Proxy logs: no authentication/TLS startup failure
PostgreSQL logs: no fatal/error/recovery loop
PostgreSQL connectivity: OK
NetworkPolicy for PostgreSQL ingress: present
Runtime RBAC for Tekton/Kubernetes evidence: present
OpenShift group-reader ClusterRoleBinding: present
```

Note about backend direct anonymous status:

- Through the Route, anonymous API/UI access should be blocked by OAuth Proxy and is expected to return `403`.
- Via direct port-forward to the backend container, anonymous API calls are handled by the application middleware and may return `401` if user identity headers are missing.
- Both are acceptable if the Route path remains protected and authorized backend testing succeeds.

---

## 5. Prepare environment and evidence directory

```bash
export DCP_NS="devops-control-plane"
export APP_LABEL="app=devops-control-plane"
export PG_LABEL="app=postgresql"
export EVIDENCE_DIR="/tmp/dcp-operability-health-check-$(date +%Y%m%d-%H%M%S)"

mkdir -p "$EVIDENCE_DIR"

date -Is | tee "$EVIDENCE_DIR/00-timestamp.txt"

echo "DCP_NS=$DCP_NS" | tee "$EVIDENCE_DIR/00-context.txt"
echo "APP_LABEL=$APP_LABEL" | tee -a "$EVIDENCE_DIR/00-context.txt"
echo "PG_LABEL=$PG_LABEL" | tee -a "$EVIDENCE_DIR/00-context.txt"
echo "EVIDENCE_DIR=$EVIDENCE_DIR" | tee -a "$EVIDENCE_DIR/00-context.txt"
```

---

## 6. Namespace and workload inventory

```bash
oc get ns "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/01-namespace.txt"

oc get deploy -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/02-deployments.txt"

oc get pods -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/03-pods.txt"

oc get rs -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/04-replicasets.txt"

oc get svc -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/05-services.txt"

oc get route -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/06-routes.txt"

oc get pvc -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/07-pvc.txt"
```

Expected result:

```text
Namespace is Active.
Deployment devops-control-plane is 1/1.
Deployment postgresql is 1/1.
DCP pod is Running 2/2 with 0 restarts.
PostgreSQL pod is Running 1/1 with 0 restarts.
Service devops-control-plane exposes 8443.
Service postgresql exposes 5432.
Route devops-control-plane is present.
PVC postgresql-data is Bound.
```

---

## 7. Identify active runtime pods

```bash
DCP_POD="$(oc get pod -n "$DCP_NS" -l "$APP_LABEL" -o jsonpath='{.items[0].metadata.name}')"
PG_POD="$(oc get pod -n "$DCP_NS" -l "$PG_LABEL" -o jsonpath='{.items[0].metadata.name}')"

echo "DCP_POD=$DCP_POD" | tee "$EVIDENCE_DIR/08-runtime-pods.txt"
echo "PG_POD=$PG_POD" | tee -a "$EVIDENCE_DIR/08-runtime-pods.txt"

oc get pod "$DCP_POD" -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/09-dcp-pod-wide.txt"
oc get pod "$PG_POD" -n "$DCP_NS" -o wide | tee "$EVIDENCE_DIR/10-postgresql-pod-wide.txt"
```

---

## 8. Container status and images

```bash
oc get pod "$DCP_POD" -n "$DCP_NS" -o jsonpath='{range .status.containerStatuses[*]}{.name}{" ready="}{.ready}{" restarts="}{.restartCount}{" image="}{.image}{"\n"}{end}' \
  | tee "$EVIDENCE_DIR/11-dcp-container-statuses.txt"

oc get pod "$PG_POD" -n "$DCP_NS" -o jsonpath='{range .status.containerStatuses[*]}{.name}{" ready="}{.ready}{" restarts="}{.restartCount}{" image="}{.image}{"\n"}{end}' \
  | tee "$EVIDENCE_DIR/12-postgresql-container-statuses.txt"
```

Expected result:

```text
devops-control-plane ready=true restarts=0
oauth-proxy ready=true restarts=0
postgresql ready=true restarts=0
```

Operational note:

- Image tags such as `latest` are acceptable for the current lab baseline but should be tracked as a production-hardening improvement.
- For production, prefer digest-pinned or immutable image references.

---

## 9. Resource requests, limits and current usage

```bash
oc get pod "$DCP_POD" -n "$DCP_NS" -o jsonpath='{range .spec.containers[*]}{.name}{" requests.cpu="}{.resources.requests.cpu}{" requests.memory="}{.resources.requests.memory}{" limits.cpu="}{.resources.limits.cpu}{" limits.memory="}{.resources.limits.memory}{"\n"}{end}' \
  | tee "$EVIDENCE_DIR/13-dcp-resources.txt"

oc get pod "$PG_POD" -n "$DCP_NS" -o jsonpath='{range .spec.containers[*]}{.name}{" requests.cpu="}{.resources.requests.cpu}{" requests.memory="}{.resources.requests.memory}{" limits.cpu="}{.resources.limits.cpu}{" limits.memory="}{.resources.limits.memory}{"\n"}{end}' \
  | tee "$EVIDENCE_DIR/14-postgresql-resources.txt"

oc adm top pod -n "$DCP_NS" | tee "$EVIDENCE_DIR/15-top-pods.txt"
```

Expected baseline example:

```text
DevOps Control Plane: low CPU and memory usage, no sustained pressure
PostgreSQL: low CPU and memory usage under idle/MVP workload
```

If `oc adm top` fails:

- Save the error output.
- Mark metrics as unavailable.
- Do not treat it as an application failure by itself.

---

## 10. Namespace events

```bash
oc get events -n "$DCP_NS" --sort-by=.lastTimestamp | tail -n 80 \
  | tee "$EVIDENCE_DIR/16-events-tail.txt"
```

Healthy result:

```text
No resources found in devops-control-plane namespace.
```

or only old/non-recurring normal events.

Investigate immediately if the events include:

```text
BackOff
Failed
FailedMount
FailedScheduling
Unhealthy
Killing
OOMKilled
CrashLoopBackOff
ImagePullBackOff
ErrImagePull
```

---

## 11. Application and OAuth Proxy logs

```bash
oc logs "$DCP_POD" -n "$DCP_NS" -c devops-control-plane --tail=120 \
  | tee "$EVIDENCE_DIR/17-dcp-app-logs-tail.txt"

oc logs "$DCP_POD" -n "$DCP_NS" -c oauth-proxy --tail=120 \
  | tee "$EVIDENCE_DIR/18-oauth-proxy-logs-tail.txt"

oc logs "$DCP_POD" -n "$DCP_NS" -c devops-control-plane --tail=300 \
  | grep -Ei 'error|panic|fatal|timeout|refused|denied|unauthorized|forbidden|x509|tls' \
  | tee "$EVIDENCE_DIR/19-dcp-app-log-warnings.txt" || true

oc logs "$DCP_POD" -n "$DCP_NS" -c oauth-proxy --tail=300 \
  | grep -Ei 'error|panic|fatal|timeout|refused|denied|unauthorized|forbidden|x509|tls' \
  | tee "$EVIDENCE_DIR/20-oauth-proxy-log-warnings.txt" || true
```

Healthy interpretation:

- Application logs should mainly show normal HTTP requests and health probes.
- OAuth Proxy logs should show startup, provider initialization, skip-auth regex compilation and HTTPS listener startup.
- A line containing `/etc/tls/private/tls.crt` from the dynamic serving controller is informational if not accompanied by errors.

Investigate if logs include recurring:

```text
panic
fatal
x509 certificate signed by unknown authority
connection refused
permission denied
unauthorized loops
forbidden loops for expected authorized users
upstream unavailable
database connection errors
```

---

## 12. PostgreSQL logs and connectivity

```bash
oc logs "$PG_POD" -n "$DCP_NS" --tail=120 \
  | tee "$EVIDENCE_DIR/21-postgresql-logs-tail.txt"

oc logs "$PG_POD" -n "$DCP_NS" --tail=300 \
  | grep -Ei 'error|fatal|panic|could not|timeout|refused|denied|checkpoint|recovery' \
  | tee "$EVIDENCE_DIR/22-postgresql-log-warnings.txt" || true

oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select now() as db_time, current_database() as database_name, current_user as user_name;"' \
  | tee "$EVIDENCE_DIR/23-postgresql-connectivity.txt"

oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as change_requests_count from change_requests;"' \
  | tee "$EVIDENCE_DIR/24-postgresql-change-requests-count.txt"

oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as change_events_count from change_events;"' \
  | tee "$EVIDENCE_DIR/25-postgresql-change-events-count.txt"

oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as evidences_count from evidences;"' \
  | tee "$EVIDENCE_DIR/26-postgresql-evidences-count.txt"
```

Healthy result:

```text
PostgreSQL accepts connections.
Database name is devops_control_plane.
Tables are queryable.
Counts are returned without errors.
```

Investigate if:

- `psql` cannot connect;
- database name is unexpected;
- tables are missing;
- queries fail with permission or relation errors;
- logs show recovery loops or fatal errors.

---

## 13. Sanitized configuration inventory

Do not print Secret values.

```bash
oc get configmap devops-control-plane-config -n "$DCP_NS" -o json \
  | jq -r '.data | to_entries[] | select(.key | test("TOKEN|PASSWORD|SECRET|DATABASE_URL") | not) | .key + "=" + .value' \
  | sort \
  | tee "$EVIDENCE_DIR/27-configmap-sanitized.txt"

oc get secret devops-control-plane-secrets -n "$DCP_NS" -o json \
  | jq -r '.data | keys[]' \
  | sort \
  | tee "$EVIDENCE_DIR/28-secret-keys-only.txt"

oc get secret devops-control-plane-secrets -n "$DCP_NS" -o json \
  | jq -r '.data | to_entries[] | .key + " base64_length=" + (.value | length | tostring)' \
  | sort \
  | tee "$EVIDENCE_DIR/29-secret-key-lengths-only.txt"
```

Healthy configuration indicators:

```text
AUTH_ENABLED=true
AUTH_OPENSHIFT_GROUP_LOOKUP_ENABLED=true
ARGOCD_INSECURE_TLS=false
GITLAB_INSECURE_TLS=false
KUBERNETES_INSECURE_TLS=false
ARGOCD_CA_FILE=/etc/dcp-trust/ca-bundle.crt
GITLAB_CA_FILE=/etc/dcp-trust/ca-bundle.crt
```

Expected Secret keys:

```text
ARGOCD_AUTH_TOKEN
DATABASE_URL
GITLAB_TOKEN
```

`KUBERNETES_TOKEN` should not be required when the ServiceAccount token fallback is active.

---

## 14. Backend health-check via port-forward

This test bypasses the Route and OAuth Proxy and talks directly to the backend container port.

```bash
oc port-forward -n "$DCP_NS" pod/"$DCP_POD" 18080:8080 > "$EVIDENCE_DIR/30-port-forward.log" 2>&1 &
PF_PID=$!

echo "PF_PID=$PF_PID" | tee "$EVIDENCE_DIR/31-port-forward-pid.txt"

sleep 3

curl -sS -o "$EVIDENCE_DIR/32-readyz-body.txt" -w "readyz HTTP %{http_code}\n" \
  http://127.0.0.1:18080/readyz \
  | tee "$EVIDENCE_DIR/32-readyz-status.txt"

curl -sS -o "$EVIDENCE_DIR/33-livez-body.txt" -w "livez HTTP %{http_code}\n" \
  http://127.0.0.1:18080/livez \
  | tee "$EVIDENCE_DIR/33-livez-status.txt"

curl -sS -o "$EVIDENCE_DIR/34-api-changes-anonymous-body.txt" -w "api changes anonymous HTTP %{http_code}\n" \
  http://127.0.0.1:18080/api/v1/changes \
  | tee "$EVIDENCE_DIR/34-api-changes-anonymous-status.txt"

curl -sS -o "$EVIDENCE_DIR/35-api-changes-admin-body.json" -w "api changes admin HTTP %{http_code}\n" \
  -H "X-Forwarded-User: operability-admin" \
  -H "X-Forwarded-Groups: devops-control-plane-admins" \
  http://127.0.0.1:18080/api/v1/changes \
  | tee "$EVIDENCE_DIR/35-api-changes-admin-status.txt"

kill "$PF_PID" 2>/dev/null || true
sleep 1

ps -p "$PF_PID" >/dev/null 2>&1 && echo "port-forward still running" || echo "port-forward stopped" \
  | tee "$EVIDENCE_DIR/36-port-forward-stop.txt"
```

Expected result:

```text
readyz HTTP 200
livez HTTP 200
api changes anonymous HTTP 401 or 403
api changes admin HTTP 200
port-forward stopped
```

If the admin-header test fails:

1. Verify `AUTH_ENABLED=true`.
2. Verify role mapping configuration.
3. Verify that `AUTH_HEADER_USER` and `AUTH_HEADER_GROUPS` match the headers used in the test.
4. Verify OpenShift group lookup only if testing with real users and no group header.
5. Check application logs for authorization failures.

---

## 15. Route health-check and anonymous access policy

```bash
APP_ROUTE="$(oc get route devops-control-plane -n "$DCP_NS" -o jsonpath='{.spec.host}')"

echo "APP_ROUTE=$APP_ROUTE" | tee "$EVIDENCE_DIR/37-route-host.txt"

curl --noproxy "$APP_ROUTE" -k -sS -o "$EVIDENCE_DIR/38-route-readyz-body.txt" -w "route readyz HTTP %{http_code}\n" \
  "https://$APP_ROUTE/readyz" \
  | tee "$EVIDENCE_DIR/38-route-readyz-status.txt"

curl --noproxy "$APP_ROUTE" -k -sS -o "$EVIDENCE_DIR/39-route-livez-body.txt" -w "route livez HTTP %{http_code}\n" \
  "https://$APP_ROUTE/livez" \
  | tee "$EVIDENCE_DIR/39-route-livez-status.txt"

curl --noproxy "$APP_ROUTE" -k -sS -o "$EVIDENCE_DIR/40-route-api-changes-anonymous-body.txt" -w "route api changes anonymous HTTP %{http_code}\n" \
  "https://$APP_ROUTE/api/v1/changes" \
  | tee "$EVIDENCE_DIR/40-route-api-changes-anonymous-status.txt"

curl --noproxy "$APP_ROUTE" -k -sS -o "$EVIDENCE_DIR/41-route-ui-dashboard-anonymous-body.txt" -w "route ui dashboard anonymous HTTP %{http_code}\n" \
  "https://$APP_ROUTE/ui/dashboard" \
  | tee "$EVIDENCE_DIR/41-route-ui-dashboard-anonymous-status.txt"
```

Expected result:

```text
route readyz HTTP 200
route livez HTTP 200
route api changes anonymous HTTP 403
route ui dashboard anonymous HTTP 403
```

If health endpoints return `403` through the Route:

1. Check OAuth Proxy arguments.
2. Confirm the skip-auth regex includes both `^/readyz$` and `^/livez$`.
3. Confirm the backend middleware allows public health endpoints.
4. Check whether the Route still targets the HTTPS service port.

If anonymous API/UI returns `200` through the Route:

1. Treat as a security issue.
2. Verify OAuth Proxy is still active.
3. Verify Route still points to the OAuth Proxy service port.
4. Verify `AUTH_ENABLED=true`.
5. Do not proceed with production promotion until remediated.

---

## 16. NetworkPolicy and RBAC checks

```bash
oc get networkpolicy -n "$DCP_NS" -o wide \
  | tee "$EVIDENCE_DIR/42-networkpolicies.txt"

oc get rolebinding -n "$DCP_NS" -o wide \
  | tee "$EVIDENCE_DIR/43-rolebindings-dcp-namespace.txt"

oc get rolebinding -n devops-ci-demo -o wide \
  | tee "$EVIDENCE_DIR/44-rolebindings-devops-ci-demo.txt"

oc get clusterrolebinding \
  | grep -E 'devops-control-plane|oauth-proxy|group-reader|auth-delegator' \
  | tee "$EVIDENCE_DIR/45-clusterrolebindings-filtered.txt" || true
```

Expected result:

```text
NetworkPolicy postgresql-ingress-from-devops-control-plane is present.
OAuth Proxy has system:auth-delegator binding.
DCP runtime ServiceAccount has required RoleBinding in devops-ci-demo.
ClusterRoleBinding devops-control-plane-group-reader is present.
```

---

## 17. Optional targeted `can-i` checks

Use these checks when authorization or integration runtime behavior is suspicious.

```bash
DCP_SA="system:serviceaccount:devops-control-plane:devops-control-plane-oauth-proxy"

oc auth can-i create pipelineruns.tekton.dev -n devops-ci-demo --as "$DCP_SA" \
  | tee "$EVIDENCE_DIR/46-can-i-create-pipelineruns.txt"

oc auth can-i get deployments -n devops-ci-demo --as "$DCP_SA" \
  | tee "$EVIDENCE_DIR/47-can-i-get-deployments.txt"

oc auth can-i list pods -n devops-ci-demo --as "$DCP_SA" \
  | tee "$EVIDENCE_DIR/48-can-i-list-pods.txt"

oc auth can-i get secrets -n devops-control-plane --as "$DCP_SA" \
  | tee "$EVIDENCE_DIR/49-can-i-get-secrets-dcp.txt"
```

Expected result:

```text
create pipelineruns.tekton.dev: yes
get deployments: yes
list pods: yes
get secrets in devops-control-plane: no
```

---

## 18. Summary generation

```bash
{
  echo "=== DevOps Control Plane operability health-check summary ==="
  echo "Timestamp: $(date -Is)"
  echo
  echo "Namespace:"
  oc get ns "$DCP_NS" --no-headers
  echo
  echo "Pods:"
  oc get pods -n "$DCP_NS" --no-headers
  echo
  echo "DCP container statuses:"
  oc get pod "$DCP_POD" -n "$DCP_NS" -o jsonpath='{range .status.containerStatuses[*]}{.name}{" ready="}{.ready}{" restarts="}{.restartCount}{"\n"}{end}'
  echo
  echo "PostgreSQL container statuses:"
  oc get pod "$PG_POD" -n "$DCP_NS" -o jsonpath='{range .status.containerStatuses[*]}{.name}{" ready="}{.ready}{" restarts="}{.restartCount}{"\n"}{end}'
  echo
  echo "Route smoke:"
  cat "$EVIDENCE_DIR/38-route-readyz-status.txt" 2>/dev/null || true
  cat "$EVIDENCE_DIR/39-route-livez-status.txt" 2>/dev/null || true
  cat "$EVIDENCE_DIR/40-route-api-changes-anonymous-status.txt" 2>/dev/null || true
  cat "$EVIDENCE_DIR/41-route-ui-dashboard-anonymous-status.txt" 2>/dev/null || true
  echo
  echo "PostgreSQL counts:"
  cat "$EVIDENCE_DIR/24-postgresql-change-requests-count.txt" 2>/dev/null || true
  cat "$EVIDENCE_DIR/25-postgresql-change-events-count.txt" 2>/dev/null || true
  cat "$EVIDENCE_DIR/26-postgresql-evidences-count.txt" 2>/dev/null || true
  echo
  echo "Evidence directory:"
  echo "$EVIDENCE_DIR"
} | tee "$EVIDENCE_DIR/50-summary.txt"
```

---

## 19. Automated smoke-test script

Starting from Phase 10.3.3, the manual checks described in this runbook are also available as an automated read-only smoke-test script:

```text
scripts/operability/health-check.sh
```

The script is intended to provide a fast and repeatable operational validation while preserving the same safety principles documented in this runbook:

- no Secret values are printed;
- Secret values are not decoded;
- no resources are applied, patched, edited, deleted or restarted;
- evidence files are written to an evidence directory;
- the final result is summarized with `PASS`, `WARN` and `FAIL` counters;
- the script returns a meaningful exit code.

### 19.1. Default execution

Use the default execution for a quick smoke test:

```bash
./scripts/operability/health-check.sh
```

By default, the script creates a timestamped evidence directory similar to:

```text
/tmp/dcp-operability-smoke-test-YYYYMMDD-HHMMSS
```

If `EVIDENCE_DIR` is already exported in the shell, the script reuses that value. To force a new timestamped evidence directory, unset it first:

```bash
unset EVIDENCE_DIR
./scripts/operability/health-check.sh
```

### 19.2. Execution with optional RBAC `can-i` checks

For a more complete operational validation, include the optional authorization checks:

```bash
unset EVIDENCE_DIR
SKIP_CAN_I=false ./scripts/operability/health-check.sh
```

Expected high-level result for the current baseline:

```text
PASS=35
WARN=0
FAIL=0
```

Expected smoke-test highlights:

```text
route readyz HTTP 200
route livez HTTP 200
route api changes anonymous HTTP 403
route ui dashboard anonymous HTTP 403
backend readyz HTTP 200
backend livez HTTP 200
backend api changes anonymous HTTP 401
backend api changes admin HTTP 200
```

### 19.3. Exit codes

The script uses the following exit codes:

```text
0 = all mandatory checks passed
1 = one or more mandatory checks failed
2 = local prerequisite or OpenShift context problem
```

### 19.4. Evidence package

The script stores evidence files under `EVIDENCE_DIR`. The main summary file is:

```text
$EVIDENCE_DIR/99-summary.txt
```

To print the final summary after execution:

```bash
cat "$EVIDENCE_DIR/99-summary.txt"
```

If `EVIDENCE_DIR` was not exported before execution, use the evidence directory printed by the script at the end of the run.

### 19.5. Relationship with this runbook

The script does not replace this runbook. Instead:

- use the script for fast routine checks and post-change smoke tests;
- use the runbook for detailed manual analysis, incident handling, onboarding and troubleshooting;
- if the script reports `FAIL > 0`, use the troubleshooting sections of this runbook to inspect the related evidence files and collect additional context.

Recommended operational usage:

```text
Routine health-check:              run scripts/operability/health-check.sh
Post-change validation:            run SKIP_CAN_I=false scripts/operability/health-check.sh
Incident initial evidence package: run the script, then attach the generated evidence directory
Detailed troubleshooting:          follow the manual sections of this runbook
```

## 20. OK/KO criteria

### Health-check OK

The health-check can be considered successful when all of the following are true:

- Namespace is Active.
- DevOps Control Plane deployment is available.
- PostgreSQL deployment is available.
- DCP pod is Running and Ready 2/2.
- PostgreSQL pod is Running and Ready 1/1.
- Restart counters are stable and ideally 0.
- `/readyz` returns HTTP 200.
- `/livez` returns HTTP 200.
- Anonymous API/UI access through the Route returns HTTP 403.
- Backend direct admin-header test returns HTTP 200.
- PostgreSQL connectivity test succeeds.
- No relevant recent warning/error events are present.
- No recurring fatal/panic/error patterns are present in logs.
- NetworkPolicy and RBAC baseline objects are present.

### Health-check KO

The health-check must be considered failed if any of the following occur:

- DCP pod is not Running or not Ready.
- PostgreSQL pod is not Running or not Ready.
- Repeated restarts are observed.
- Route health endpoints fail.
- Anonymous API/UI access through the Route returns HTTP 200.
- PostgreSQL connectivity fails.
- Logs show recurring fatal errors, TLS failures, database failures or authorization failures for expected authorized users.
- Required RBAC or NetworkPolicy objects are missing.

---

## 21. Troubleshooting quick guide

### DCP pod not Ready

Collect:

```bash
oc describe pod "$DCP_POD" -n "$DCP_NS"
oc logs "$DCP_POD" -n "$DCP_NS" -c devops-control-plane --tail=200
oc logs "$DCP_POD" -n "$DCP_NS" -c oauth-proxy --tail=200
oc get events -n "$DCP_NS" --sort-by=.lastTimestamp | tail -n 80
```

Common causes:

- application startup failure;
- invalid ConfigMap value;
- missing Secret key;
- invalid trust bundle;
- OAuth Proxy TLS/cookie secret issue;
- readiness/liveness probe failure.

### PostgreSQL pod not Ready

Collect:

```bash
oc describe pod "$PG_POD" -n "$DCP_NS"
oc logs "$PG_POD" -n "$DCP_NS" --tail=200
oc get pvc -n "$DCP_NS" -o wide
oc get events -n "$DCP_NS" --sort-by=.lastTimestamp | tail -n 80
```

Common causes:

- PVC not mounted;
- storage issue;
- corrupted or unavailable data directory;
- resource pressure;
- invalid PostgreSQL Secret configuration.

### Route health fails

Collect:

```bash
oc get route devops-control-plane -n "$DCP_NS" -o yaml
oc get svc devops-control-plane -n "$DCP_NS" -o yaml
oc describe pod "$DCP_POD" -n "$DCP_NS"
oc logs "$DCP_POD" -n "$DCP_NS" -c oauth-proxy --tail=200
```

Common causes:

- Route points to the wrong service port;
- OAuth Proxy not listening on 8443;
- TLS secret missing or invalid;
- skip-auth regex missing for health endpoints;
- backend health endpoint blocked by middleware.

### Authorized user cannot access UI/API

Collect:

```bash
oc get configmap devops-control-plane-config -n "$DCP_NS" -o json \
  | jq -r '.data | to_entries[] | select(.key | startswith("AUTH_")) | .key + "=" + .value' \
  | sort

oc get clusterrolebinding devops-control-plane-group-reader -o yaml
oc get group devops-control-plane-admins -o yaml 2>/dev/null || true
oc get group devops-control-plane-operators -o yaml 2>/dev/null || true
oc get group devops-control-plane-approvers -o yaml 2>/dev/null || true
oc get group devops-control-plane-viewers -o yaml 2>/dev/null || true
```

Common causes:

- user not member of expected OpenShift group;
- group lookup RBAC missing;
- wrong trusted header name;
- OAuth Proxy not passing user headers;
- backend role mapping misconfigured.

### Argo CD or GitLab integration fails

Collect only sanitized configuration and error logs:

```bash
oc get configmap devops-control-plane-config -n "$DCP_NS" -o json \
  | jq -r '.data | to_entries[] | select(.key | test("ARGOCD|GITLAB")) | select(.key | test("TOKEN|PASSWORD|SECRET") | not) | .key + "=" + .value' \
  | sort

oc logs "$DCP_POD" -n "$DCP_NS" -c devops-control-plane --tail=300 \
  | grep -Ei 'argocd|gitlab|x509|tls|unauthorized|forbidden|timeout|refused|error' || true
```

Common causes:

- expired token;
- missing or invalid CA bundle;
- TLS strict mode failure;
- remote API unavailable;
- insufficient RBAC on the external system.

If a token value is exposed while troubleshooting, rotate it immediately.

---

## 22. Evidence package to attach to an incident or review

At minimum, include:

```text
00-context.txt
01-namespace.txt
02-deployments.txt
03-pods.txt
05-services.txt
06-routes.txt
07-pvc.txt
11-dcp-container-statuses.txt
12-postgresql-container-statuses.txt
13-dcp-resources.txt
14-postgresql-resources.txt
15-top-pods.txt
16-events-tail.txt
19-dcp-app-log-warnings.txt
20-oauth-proxy-log-warnings.txt
22-postgresql-log-warnings.txt
23-postgresql-connectivity.txt
24-postgresql-change-requests-count.txt
25-postgresql-change-events-count.txt
26-postgresql-evidences-count.txt
27-configmap-sanitized.txt
28-secret-keys-only.txt
29-secret-key-lengths-only.txt
38-route-readyz-status.txt
39-route-livez-status.txt
40-route-api-changes-anonymous-status.txt
41-route-ui-dashboard-anonymous-status.txt
42-networkpolicies.txt
43-rolebindings-dcp-namespace.txt
44-rolebindings-devops-ci-demo.txt
45-clusterrolebindings-filtered.txt
50-summary.txt
```

Before sharing externally, review all files and remove any accidental sensitive value.

---

## 23. Current known gaps and production recommendations

The following items are not blockers for the current lab/MVP baseline, but should be tracked for production maturity:

1. Pin external images currently using `latest` to immutable tags or digests.
2. Define long-term monitoring dashboards and alerts.
3. Define log retention and forwarding policy.
4. Define backup retention and storage policy for PostgreSQL backups.
5. Consider PostgreSQL high availability if the target production profile requires it.
6. Consider PodDisruptionBudget and controlled rollout strategy for production.
7. Define SLO/SLA expectations for `/readyz`, `/livez`, API availability and evidence collection workflows.
8. Periodically test restore procedures in an isolated namespace.
9. Periodically validate AuthN/AuthZ role matrix after group or RBAC changes.
10. Periodically validate NetworkPolicy behavior before introducing stricter deny-all policies.

---

## 24. Revision history

| Date | Phase | Description |
|---|---:|---|
| 2026-07-03 | 10.3.2 | Initial operability health-check runbook based on Phase 10.3.1 observability baseline inventory. |
| 2026-07-03 | 10.3.4 | Added reference to the automated operability smoke-test script introduced in Phase 10.3.3. |

## Post-Phase 15 runtime health-check matrix

Status: Active operational baseline  
Phase reference: 10.9.1  
Last validated: 2026-07-09  
Related baseline tag: `namespace-isolated-baseline-20260709`

### Purpose

This section refreshes the operational health-check runbook after completion of Phase 15.

The current DevOps Control Plane runtime baseline is namespace-isolated on the available `ocp-dev` OpenShift cluster.

Current validated topology:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

Physical multi-cluster validation remains deferred because no additional OpenShift cluster is currently available.

The codebase is multi-cluster code-ready, but operational runtime checks must use the namespace-isolated baseline until real clusters become available.

### Current operational smoke matrix

A complete health-check must validate:

- DevOps Control Plane pod readiness;
- `/readyz` endpoint;
- dashboard UI HTTP response;
- Argo CD Applications for dev, staging and production;
- application deployment readiness in all three namespaces;
- route `/healthz` for all three environments;
- Tekton validation PipelineRuns for staging and production;
- UI detail pages for latest staging and production validation ChangeRequests;
- repository working tree cleanliness when the check is executed from the project repository.

### Evidence directory convention

Use a timestamped evidence directory for every operational health check.

Example command sequence:

- set `EVIDENCE_DIR` to a path under `/tmp`;
- create the directory;
- write command outputs into files under that directory;
- keep the evidence directory if any check fails.

Recommended naming pattern:

`/tmp/dcp-operability-health-YYYYMMDD-HHMMSS`

### Repository state check

When the runbook is executed from the repository root, verify that no accidental local change exists.

Command:

`git status --short`

Expected result: no output.

### DevOps Control Plane pod check

Validate the DevOps Control Plane pod in namespace `devops-control-plane`.

Expected result:

- DevOps Control Plane pod is `Running`;
- both containers are ready;
- the application image points to the DevOps Control Plane internal registry image.

Recommended evidence files:

- `01-dcp-pod.txt`
- `02-dcp-pods.txt`
- `03-dcp-images.txt`

### Direct readiness check

Use port-forward for direct application readiness validation.

Expected result:

`readyz_http=200`

Recommended evidence files:

- `04-port-forward.log`
- `05-readyz.json`
- `06-readyz-http.txt`

The port-forward process must be stopped at the end of the runbook.

### Dashboard UI check

Validate the dashboard through the direct application port.

Expected result:

`dashboard_http=200`

The dashboard should show:

- latest ChangeRequest;
- `Environments / Namespaces`;
- namespace mapping for dev, staging and production;
- user box in the topbar;
- runtime evidence card when available.

Recommended evidence files:

- `10-dashboard.html`
- `11-dashboard-http.txt`

### Argo CD application matrix

Validate the following Argo CD Applications:

- `demo-go-color-app`
- `demo-go-color-app-staging`
- `demo-go-color-app-production`

Expected result for all environments:

- sync status: `Synced`
- health status: `Healthy`

Recommended evidence files:

- `20-argocd-dev.txt`
- `21-argocd-staging.txt`
- `22-argocd-production.txt`

### Deployment readiness matrix

Validate `demo-go-color-app` deployment readiness in:

- `devops-ci-demo`
- `devops-ci-staging`
- `devops-ci-production`

Expected result:

- ready replicas match desired replicas;
- available replicas are present;
- updated replicas are present.

Recommended evidence files:

- `30-deploy-dev.txt`
- `31-deploy-staging.txt`
- `32-deploy-production.txt`

### Route health matrix

Validate route `/healthz` for all three environments.

Expected result:

- `dev_healthz_http=200`
- `staging_healthz_http=200`
- `production_healthz_http=200`

Recommended evidence files:

- `40-route-hosts.txt`
- `42-dev-healthz-http.txt`
- `44-staging-healthz-http.txt`
- `46-production-healthz-http.txt`

### Tekton validation matrix

Final known validated PipelineRuns:

- staging: `devops-cp-validate-chg-2026-0049-nd7rm`
- production: `devops-cp-validate-chg-2026-0050-8wqtv`

Expected result:

- status: `True`
- reason: `Succeeded`

Recommended evidence files:

- `50-tekton-staging.txt`
- `51-tekton-production.txt`

### UI ChangeRequest detail checks

Validate the ChangeRequest detail pages:

- `CHG-2026-0049`
- `CHG-2026-0050`

Expected result:

- `chg0049_ui_http=200`
- `chg0050_ui_http=200`

Recommended evidence files:

- `60-ui-chg-0049.html`
- `61-ui-chg-0049-http.txt`
- `70-ui-chg-0050.html`
- `71-ui-chg-0050-http.txt`

### Summary file

Each execution should produce a final summary file named:

`90-summary.txt`

The summary should include:

- evidence directory path;
- readiness HTTP result;
- dashboard HTTP result;
- Argo CD matrix;
- deployment readiness matrix;
- route health matrix;
- Tekton validation matrix;
- UI ChangeRequest detail HTTP results.

### Pass criteria

The runbook passes when:

- `/readyz` returns HTTP `200`;
- dashboard returns HTTP `200`;
- Argo CD is `Synced` and `Healthy` for dev, staging and production;
- deployments are ready in all three namespaces;
- route `/healthz` returns HTTP `200` for all three environments;
- staging and production Tekton validation PipelineRuns are `Succeeded`;
- ChangeRequest detail pages return HTTP `200`;
- no sensitive Secret material is printed;
- the repository working tree remains clean if the check is executed from the repository.

### Failure handling

If one of the checks fails:

1. stop the runbook;
2. preserve the evidence directory;
3. do not rerun destructive commands;
4. identify the failing layer:
   - DevOps Control Plane pod;
   - PostgreSQL dependency;
   - Argo CD Application;
   - application deployment;
   - route health;
   - Tekton PipelineRun;
   - UI rendering;
   - authentication or forwarded groups;
5. open a follow-up remediation task with the evidence directory path.

### Notes for future real multi-cluster onboarding

This runbook validates the current namespace-isolated operational baseline.

When real clusters become available, this matrix must be extended with physical cluster-specific checks.

The future real-cluster checks must validate that:

- staging does not silently fall back to `ocp-dev`;
- production does not silently fall back to `ocp-dev`;
- missing runtime provider fails closed;
- disabled runtime provider fails closed;
- Secret references are allow-listed;
- runtime Secret loader and factories are explicitly enabled only when safe;
- no raw Secret values appear in logs, evidence or UI.

## Post-Phase 15 troubleshooting matrix

Status: Active operational baseline  
Phase reference: 10.10.1  
Last updated: 2026-07-09

### Purpose

This section refreshes the troubleshooting matrix after completion of Phase 15.

The current operational baseline is namespace-isolated on `ocp-dev`:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

The codebase is multi-cluster code-ready, but physical cross-cluster runtime validation remains deferred because no additional OpenShift cluster is currently available.

This troubleshooting matrix focuses on the current validated runtime baseline and on the fail-closed guardrails introduced for future multi-cluster onboarding.

### Troubleshooting principles

During troubleshooting:

1. preserve the evidence directory;
2. avoid destructive commands unless an approved remediation procedure explicitly requires them;
3. do not print Secret values;
4. do not decode or copy raw tokens into notes, tickets or evidence;
5. identify the failing layer before applying remediation;
6. keep `dev`, `staging` and `production` namespace-isolated checks separate;
7. confirm that no operation silently falls back to `ocp-dev` when a non-current target is expected.

### Common failure matrix

#### DevOps Control Plane pod is not ready

Symptoms:

- the DevOps Control Plane pod is not `Running`;
- one or more containers are not ready;
- rollout does not complete.

Check:

- inspect deployment status in namespace `devops-control-plane`;
- inspect pod events;
- inspect container logs;
- verify image reference;
- verify PostgreSQL reachability.

Likely causes:

- image pull issue;
- configuration error;
- PostgreSQL connection issue;
- OAuth proxy sidecar issue;
- resource pressure.

Remediation direction:

- do not restart repeatedly without checking events;
- preserve pod logs and events;
- verify ConfigMap, Secret references and deployment image;
- rollback only if a recent deployment introduced the failure.

#### `/readyz` returns non-200

Symptoms:

- direct `/readyz` returns non-200;
- route `/readyz` returns non-200;
- dashboard may still load partially.

Check:

- direct application port-forward;
- route through OAuth proxy;
- PostgreSQL connectivity;
- application logs.

Likely causes:

- database not reachable;
- application configuration issue;
- readiness handler dependency failure;
- route or OAuth proxy issue.

Remediation direction:

- distinguish direct backend failure from route or proxy failure;
- if direct readiness fails, inspect application and database first;
- if only route readiness fails, inspect OAuth proxy and route configuration.

#### Dashboard returns non-200

Symptoms:

- dashboard HTTP check fails;
- browser shows error;
- UI does not render latest ChangeRequest.

Check:

- backend route or port-forward response;
- authentication headers or OAuth proxy behavior;
- application logs;
- `/readyz`.

Likely causes:

- AuthN/AuthZ issue;
- OAuth proxy issue;
- UI handler error;
- backend dependency issue.

Remediation direction:

- first confirm `/readyz`;
- then test dashboard through direct port-forward with trusted forwarded headers;
- preserve HTML response if present.

#### Dashboard does not show latest ChangeRequest

Symptoms:

- dashboard shows an older ChangeRequest;
- expected latest ChangeRequest is missing.

Check:

- ChangeRequest list through API or UI;
- database records;
- UI dashboard HTML;
- recent commits affecting UI handlers.

Likely causes:

- latest ChangeRequest not present in database;
- UI selection logic issue;
- stale pod image;
- browser cache.

Remediation direction:

- confirm latest ChangeRequest exists in backend data;
- verify running pod image;
- restart rollout only after confirming the image in registry is updated.

#### `Environments / Namespaces` is missing in UI

Symptoms:

- topbar shows old static environment text;
- namespace mapping is not visible.

Check:

- running pod image;
- UI HTML from dashboard;
- `internal/api/ui_handlers.go` in the expected commit;
- rollout status.

Likely causes:

- stale image deployed;
- latest tag not pushed;
- rollout restarted before push;
- browser cache.

Remediation direction:

- confirm image was pushed before rollout;
- inspect pod image;
- force rollout only after push is confirmed.

#### Argo CD Application is OutOfSync or Degraded

Symptoms:

- sync status is not `Synced`;
- health status is not `Healthy`;
- deployment readiness may be affected.

Check:

- Application status in `openshift-gitops`;
- target namespace resources;
- GitOps repository revision;
- Kustomize overlay path.

Likely causes:

- GitOps manifest issue;
- namespace mismatch;
- failed sync;
- missing resource or invalid overlay;
- target namespace permissions.

Remediation direction:

- inspect Argo CD Application status first;
- compare expected GitOps path with actual Application path;
- avoid manual cluster drift unless explicitly required for emergency restoration.

#### Deployment is not ready in one namespace

Symptoms:

- ready replicas do not match desired replicas;
- pods are pending, crashing or unavailable;
- route health check fails.

Check:

- Deployment status;
- ReplicaSet status;
- Pod status;
- Pod events;
- container logs;
- image pull status.

Likely causes:

- image pull issue;
- invalid deployment manifest;
- insufficient resources;
- failing application health endpoint;
- namespace-specific configuration issue.

Remediation direction:

- isolate the affected namespace;
- avoid assuming all environments are impacted;
- check `devops-ci-demo`, `devops-ci-staging` and `devops-ci-production` separately.

#### Route `/healthz` fails

Symptoms:

- route health check returns non-200;
- deployment may still be ready.

Check:

- Route host;
- Service endpoints;
- Pod readiness;
- application logs;
- TLS route configuration.

Likely causes:

- route misconfiguration;
- service selector mismatch;
- application endpoint failure;
- DNS or ingress issue.

Remediation direction:

- test service and pod locally if possible;
- confirm route points to the correct service;
- preserve route host evidence.

#### Tekton PipelineRun failed

Symptoms:

- PipelineRun condition status is not `True`;
- reason is not `Succeeded`;
- failed task count is greater than zero;
- UI validation evidence shows failure.

Check:

- PipelineRun conditions;
- TaskRun status;
- TaskRun logs;
- Git revision;
- validation path;
- namespace where PipelineRun was created.

Likely causes:

- invalid GitOps path;
- missing Task or Pipeline;
- Git clone issue;
- RBAC issue in namespace;
- workspace or parameter issue.

Remediation direction:

- inspect the PipelineRun in the target Tekton namespace;
- confirm validation path is environment-specific;
- do not rerun blindly before understanding failed TaskRuns.

#### UI does not show Tekton validation evidence

Symptoms:

- ChangeRequest detail page loads;
- Tekton validation card is missing;
- raw evidence exists but is not rendered.

Check:

- evidence records for the ChangeRequest;
- evidence type;
- latest validation evidence;
- UI detail HTML;
- backend logs.

Likely causes:

- validation evidence not collected;
- wrong ChangeRequest selected;
- evidence type mismatch;
- stale pod image;
- UI rendering issue.

Remediation direction:

- verify check-validation was executed;
- inspect evidence list for the ChangeRequest;
- confirm UI is running the commit that includes Tekton validation card support.

#### Wrong environment or namespace mapping

Symptoms:

- action executes against unexpected namespace;
- UI shows unexpected environment mapping;
- evidence references the wrong namespace.

Check:

- Environment Catalog;
- Cluster Registry;
- TechnicalRuntimeTarget resolution;
- ChangeRequest target environment;
- runtime evidence payload.

Likely causes:

- stale environment catalog file;
- wrong targetEnvironment in ChangeRequest;
- fallback configuration;
- deployment running previous ConfigMap.

Remediation direction:

- verify target environment first;
- verify resolved namespace;
- verify pod configuration;
- do not proceed if the target does not match the expected namespace.

#### Runtime provider missing

Symptoms:

- technical action fails before runtime client construction;
- error references missing runtime provider for cluster;
- simulated external target fails as expected.

Expected behavior:

- this is fail-closed behavior.

Likely causes:

- cluster exists in target model but no runtime provider is configured;
- real cluster onboarding is incomplete;
- provider registry does not include the target cluster.

Remediation direction:

- do not force fallback to `ocp-dev`;
- configure provider only through the approved onboarding contract;
- verify Secret references, RBAC and factories before enablement.

#### Runtime provider disabled

Symptoms:

- provider selection fails with disabled provider error.

Expected behavior:

- this is fail-closed behavior.

Likely causes:

- target provider intentionally disabled;
- onboarding not approved;
- readiness gates not satisfied.

Remediation direction:

- keep disabled until prerequisites are complete;
- do not enable provider only to bypass the error;
- complete readiness gates first.

#### Secret loading fails

Symptoms:

- Secret reference validation fails;
- allow-list rejects cluster or Secret reference;
- getter not configured;
- required key missing.

Expected behavior:

- this is fail-closed behavior.

Likely causes:

- runtime Secret loader disabled;
- Secret reference not allow-listed;
- Secret missing required key;
- Kubernetes Secret getter not configured.

Remediation direction:

- do not print or decode Secret values into logs;
- verify only names, namespaces and key names;
- update allow-list only through approved change;
- keep evidence sanitized.

#### Runtime factory fails

Symptoms:

- factory returns not configured;
- API URL missing;
- base URL missing;
- token value missing;
- kubeconfig unsupported;
- raw CA unsupported.

Expected behavior:

- this is fail-closed behavior.

Likely causes:

- factory flags disabled;
- required runtime configuration missing;
- unsupported credential material;
- real cluster onboarding incomplete.

Remediation direction:

- do not bypass disabled-by-default flags;
- enable only the exact required factory;
- validate configuration with unit tests before runtime enablement;
- preserve fail-closed semantics.

### Escalation criteria

Escalate when:

- `/readyz` remains non-200 after dependency checks;
- Argo CD is degraded for multiple environments;
- deployments are unavailable in more than one namespace;
- Tekton repeatedly fails for the same validation path;
- evidence contains unexpected sensitive data;
- target environment resolution appears unsafe;
- a provider unexpectedly falls back to `ocp-dev`;
- rollback is required.

### Evidence to preserve

For every incident, preserve:

- evidence directory path;
- pod list;
- pod events;
- readiness output;
- route health output;
- Argo CD Application status;
- deployment status;
- PipelineRun and TaskRun status;
- UI HTTP response status;
- relevant sanitized logs;
- command timestamps where available.

### Final rule

If the system cannot prove the intended target environment and namespace, stop the operation.

For multi-cluster readiness, an explicit fail-closed error is safer than an implicit fallback to the wrong cluster or namespace.

## Post-Phase 15 incident triage quick-reference

Status: Active operational baseline  
Phase reference: 10.10.2  
Last updated: 2026-07-09

### Purpose

This section provides a quick-reference triage model for operators handling incidents on the current DevOps Control Plane runtime baseline.

The goal is to reduce ambiguity during troubleshooting by classifying incidents by layer, expected evidence and escalation criteria.

The current runtime baseline is namespace-isolated on `ocp-dev`:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

Physical multi-cluster validation remains deferred. Multi-cluster code readiness is completed and guarded by fail-closed behavior.

### Triage priority levels

Use the following priority levels during operational triage.

Priority 1 — Control Plane unavailable:

- DevOps Control Plane pod is not ready;
- `/readyz` is not HTTP 200;
- PostgreSQL is unavailable;
- users cannot access the platform UI or APIs.

Required action:

- stop normal validation workflows;
- preserve evidence;
- investigate platform availability first;
- escalate if service cannot be restored quickly.

Priority 2 — Runtime automation unavailable:

- Tekton validation cannot start or complete;
- Argo CD Applications are degraded or out of sync;
- deployment readiness fails in one or more namespaces;
- route health fails for an application environment.

Required action:

- isolate the affected environment;
- preserve PipelineRun, TaskRun, Argo CD and deployment evidence;
- do not assume all environments are impacted.

Priority 3 — Evidence or UI inconsistency:

- UI loads but does not show expected runtime evidence;
- dashboard does not show latest ChangeRequest;
- Tekton validation card is missing;
- evidence appears stale or incomplete.

Required action:

- verify backend evidence records;
- verify ChangeRequest number;
- verify running image and rollout status;
- avoid rerunning automation until the mismatch is understood.

Priority 4 — Guardrail or configuration failure:

- runtime provider is missing;
- runtime provider is disabled;
- Secret reference is not allow-listed;
- runtime factory is disabled;
- unsupported kubeconfig or raw CA input is rejected.

Required action:

- treat the failure as expected fail-closed behavior;
- do not bypass the guardrail;
- complete the required onboarding or configuration checklist before retrying.

### Layer-based triage map

Platform layer:

- DevOps Control Plane Deployment;
- OAuth proxy;
- PostgreSQL;
- `/readyz`;
- `/livez`;
- application logs.

Environment runtime layer:

- Environment Catalog;
- namespace mapping;
- Kubernetes Deployment;
- Service;
- Route;
- runtime evidence.

GitOps layer:

- Argo CD Application;
- Git repository revision;
- Kustomize overlay path;
- sync status;
- health status.

Tekton layer:

- PipelineRun;
- TaskRun;
- validation path;
- pipeline parameters;
- git revision;
- failed task count.

Security and guardrail layer:

- RBAC;
- Secret references;
- Secret reference allow-list;
- runtime Secret loader;
- runtime client factories;
- provider registry.

UI and evidence layer:

- dashboard;
- ChangeRequest detail;
- Tekton validation card;
- latest runtime evidence;
- sanitized raw evidence view.

### Minimum evidence package

For every incident, collect at least the following evidence:

- incident timestamp;
- affected environment;
- affected namespace;
- ChangeRequest number, if applicable;
- evidence directory path;
- DevOps Control Plane pod status;
- `/readyz` HTTP status;
- dashboard HTTP status, if UI is affected;
- Argo CD Application status, if deployment is affected;
- Deployment readiness, if application runtime is affected;
- Route `/healthz` HTTP status, if application runtime is affected;
- PipelineRun and TaskRun status, if Tekton is affected;
- sanitized application logs or controller logs, if needed;
- final `git status --short` output if the check is executed from the repository.

### Evidence naming convention

Use a timestamped directory under `/tmp`.

Recommended pattern:

`/tmp/dcp-incident-YYYYMMDD-HHMMSS`

Recommended files:

- `00-git-status-before.txt`
- `01-dcp-pods.txt`
- `02-readyz-http.txt`
- `03-dashboard-http.txt`
- `10-argocd-status.txt`
- `20-deployment-status.txt`
- `30-route-health.txt`
- `40-tekton-pipelinerun.txt`
- `41-tekton-taskruns.txt`
- `50-ui-change-detail-http.txt`
- `90-summary.txt`
- `99-git-status-after.txt`

### Stop conditions

Stop the troubleshooting flow and escalate when:

- target environment or namespace cannot be proven;
- evidence contains unexpected sensitive data;
- a technical action appears to target the wrong namespace;
- a runtime provider unexpectedly resolves to the wrong cluster;
- a disabled provider or missing provider error is bypassed manually;
- a Secret value is printed or copied into evidence;
- production-like namespace actions are not explicitly understood.

### Expected fail-closed outcomes

The following outcomes are expected and should not be treated as defects by themselves:

- unknown environment rejected;
- disabled environment rejected;
- missing runtime provider rejected;
- disabled runtime provider rejected;
- Secret reference not allow-listed rejected;
- Secret getter not configured rejected;
- runtime Secret loader disabled;
- runtime factory disabled;
- unsupported kubeconfig rejected;
- unsupported raw CA rejected;
- missing token value rejected;
- missing API URL or Argo CD base URL rejected.

These outcomes protect the platform from unsafe execution.

### Unsafe remediation examples

Avoid the following actions:

- enabling a runtime provider only to bypass an error;
- enabling all runtime factories at once;
- adding broad Secret access;
- using cluster-admin for convenience;
- manually changing GitOps-managed resources without documenting drift;
- decoding tokens into terminal output;
- copying raw Secret values into tickets or evidence files;
- forcing a fallback to `ocp-dev` for an environment that should target another cluster.

### Safe remediation examples

Preferred remediation actions:

- verify Environment Catalog entry;
- verify Cluster Registry entry;
- verify RBAC in the target namespace;
- verify Argo CD Application health;
- verify PipelineRun and TaskRun status;
- verify Secret reference names without decoding values;
- verify allow-list entries;
- verify factory flags;
- preserve sanitized evidence;
- create a follow-up change for configuration updates.

### Escalation handoff template

When escalating, include:

- summary of the incident;
- priority level;
- affected environment;
- affected namespace;
- affected ChangeRequest;
- evidence directory;
- exact failing check;
- expected result;
- actual result;
- suspected layer;
- remediation already attempted;
- confirmation that no raw Secret values were exposed.

### Final triage rule

If an operator cannot prove that the operation targets the expected environment, namespace and runtime provider, the operator must stop the operation.

A clear fail-closed error is safer than an implicit action against the wrong cluster or namespace.
