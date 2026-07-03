#!/usr/bin/env bash
set -Eeuo pipefail

# DevOps Control Plane — Operability smoke-test script
# Phase 10.3.3
#
# Purpose:
#   Execute a read-only smoke test against the DevOps Control Plane runtime on OpenShift.
#
# Safety:
#   - Does not print Secret values.
#   - Does not decode Secret values.
#   - Does not apply, patch, edit, delete or restart resources.
#   - Writes evidence files under EVIDENCE_DIR.
#
# Exit codes:
#   0 = all mandatory checks passed
#   1 = one or more mandatory checks failed
#   2 = invalid local prerequisites or missing required tools/context

DCP_NS="${DCP_NS:-devops-control-plane}"
APP_LABEL="${APP_LABEL:-app=devops-control-plane}"
PG_LABEL="${PG_LABEL:-app=postgresql}"
DCP_ROUTE_NAME="${DCP_ROUTE_NAME:-devops-control-plane}"
DCP_SERVICE_NAME="${DCP_SERVICE_NAME:-devops-control-plane}"
PG_SERVICE_NAME="${PG_SERVICE_NAME:-postgresql}"
BACKEND_PORT="${BACKEND_PORT:-8080}"
LOCAL_PORT="${LOCAL_PORT:-18080}"
TEST_USER="${TEST_USER:-operability-admin}"
TEST_GROUP="${TEST_GROUP:-devops-control-plane-admins}"
EVIDENCE_DIR="${EVIDENCE_DIR:-/tmp/dcp-operability-smoke-test-$(date +%Y%m%d-%H%M%S)}"
SKIP_TOP="${SKIP_TOP:-false}"
SKIP_PG="${SKIP_PG:-false}"
SKIP_CAN_I="${SKIP_CAN_I:-true}"
CURL_INSECURE="${CURL_INSECURE:-true}"

mkdir -p "$EVIDENCE_DIR"

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0
PF_PID=""
APP_ROUTE=""
DCP_POD=""
PG_POD=""

log() {
  printf '%s %s\n' "$(date -Is)" "$*"
}

info() {
  log "INFO  $*" | tee -a "$EVIDENCE_DIR/00-run.log"
}

warn() {
  WARN_COUNT=$((WARN_COUNT + 1))
  log "WARN  $*" | tee -a "$EVIDENCE_DIR/00-run.log"
}

pass() {
  PASS_COUNT=$((PASS_COUNT + 1))
  log "PASS  $*" | tee -a "$EVIDENCE_DIR/00-run.log"
}

fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  log "FAIL  $*" | tee -a "$EVIDENCE_DIR/00-run.log"
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" | tee -a "$EVIDENCE_DIR/00-run.log"
    exit 2
  fi
}

cleanup() {
  if [ -n "${PF_PID:-}" ]; then
    if ps -p "$PF_PID" >/dev/null 2>&1; then
      kill "$PF_PID" >/dev/null 2>&1 || true
      # Reap the background port-forward process to avoid an interactive shell
      # job-control "Terminated" message after a successful run.
      wait "$PF_PID" >/dev/null 2>&1 || true
      sleep 1
    fi
  fi
}
trap cleanup EXIT

http_status() {
  local url="$1"
  local output_file="$2"
  shift 2

  local insecure_args=()
  if [ "$CURL_INSECURE" = "true" ]; then
    insecure_args=(-k)
  fi

  curl --noproxy "*" "${insecure_args[@]}" -sS -o "$output_file" -w "%{http_code}" "$@" "$url"
}

check_status_equals() {
  local name="$1"
  local actual="$2"
  local expected="$3"

  if [ "$actual" = "$expected" ]; then
    pass "$name returned HTTP $actual"
  else
    fail "$name returned HTTP $actual, expected $expected"
  fi
}

check_status_in() {
  local name="$1"
  local actual="$2"
  shift 2
  local expected

  for expected in "$@"; do
    if [ "$actual" = "$expected" ]; then
      pass "$name returned HTTP $actual"
      return 0
    fi
  done

  fail "$name returned HTTP $actual, expected one of: $*"
}

info "Starting DevOps Control Plane operability smoke test"
info "Evidence directory: $EVIDENCE_DIR"

need_cmd oc
need_cmd curl
need_cmd grep
need_cmd sed
need_cmd sort
need_cmd tail
need_cmd tee

if ! oc whoami >/dev/null 2>&1; then
  echo "Not logged into OpenShift or current context unavailable" | tee -a "$EVIDENCE_DIR/00-run.log"
  exit 2
fi

cat > "$EVIDENCE_DIR/00-context.txt" <<EOF
DCP_NS=$DCP_NS
APP_LABEL=$APP_LABEL
PG_LABEL=$PG_LABEL
DCP_ROUTE_NAME=$DCP_ROUTE_NAME
DCP_SERVICE_NAME=$DCP_SERVICE_NAME
PG_SERVICE_NAME=$PG_SERVICE_NAME
BACKEND_PORT=$BACKEND_PORT
LOCAL_PORT=$LOCAL_PORT
TEST_USER=$TEST_USER
TEST_GROUP=$TEST_GROUP
EVIDENCE_DIR=$EVIDENCE_DIR
SKIP_TOP=$SKIP_TOP
SKIP_PG=$SKIP_PG
SKIP_CAN_I=$SKIP_CAN_I
EOF

oc whoami > "$EVIDENCE_DIR/01-oc-whoami.txt" 2>&1 || true
oc project > "$EVIDENCE_DIR/02-oc-project.txt" 2>&1 || true

info "Checking namespace"
if oc get ns "$DCP_NS" -o wide > "$EVIDENCE_DIR/03-namespace.txt" 2>&1; then
  if grep -q "Active" "$EVIDENCE_DIR/03-namespace.txt"; then
    pass "Namespace $DCP_NS is Active"
  else
    fail "Namespace $DCP_NS is not Active"
  fi
else
  fail "Namespace $DCP_NS not found or not accessible"
fi

info "Collecting workload inventory"
oc get deploy -n "$DCP_NS" -o wide > "$EVIDENCE_DIR/04-deployments.txt" 2>&1 || true
oc get pods -n "$DCP_NS" -o wide > "$EVIDENCE_DIR/05-pods.txt" 2>&1 || true
oc get svc -n "$DCP_NS" -o wide > "$EVIDENCE_DIR/06-services.txt" 2>&1 || true
oc get route -n "$DCP_NS" -o wide > "$EVIDENCE_DIR/07-routes.txt" 2>&1 || true
oc get pvc -n "$DCP_NS" -o wide > "$EVIDENCE_DIR/08-pvc.txt" 2>&1 || true

if oc get deploy "$DCP_SERVICE_NAME" -n "$DCP_NS" -o jsonpath='{.status.availableReplicas}' > "$EVIDENCE_DIR/09-dcp-available-replicas.txt" 2>&1; then
  DCP_AVAILABLE="$(cat "$EVIDENCE_DIR/09-dcp-available-replicas.txt")"
  if [ "${DCP_AVAILABLE:-0}" -ge 1 ] 2>/dev/null; then
    pass "Deployment $DCP_SERVICE_NAME has available replicas: $DCP_AVAILABLE"
  else
    fail "Deployment $DCP_SERVICE_NAME has no available replicas"
  fi
else
  fail "Deployment $DCP_SERVICE_NAME not found"
fi

if oc get deploy "$PG_SERVICE_NAME" -n "$DCP_NS" -o jsonpath='{.status.availableReplicas}' > "$EVIDENCE_DIR/10-postgresql-available-replicas.txt" 2>&1; then
  PG_AVAILABLE="$(cat "$EVIDENCE_DIR/10-postgresql-available-replicas.txt")"
  if [ "${PG_AVAILABLE:-0}" -ge 1 ] 2>/dev/null; then
    pass "Deployment $PG_SERVICE_NAME has available replicas: $PG_AVAILABLE"
  else
    fail "Deployment $PG_SERVICE_NAME has no available replicas"
  fi
else
  fail "Deployment $PG_SERVICE_NAME not found"
fi

DCP_POD="$(oc get pod -n "$DCP_NS" -l "$APP_LABEL" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
PG_POD="$(oc get pod -n "$DCP_NS" -l "$PG_LABEL" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"

echo "DCP_POD=$DCP_POD" > "$EVIDENCE_DIR/11-runtime-pods.txt"
echo "PG_POD=$PG_POD" >> "$EVIDENCE_DIR/11-runtime-pods.txt"

if [ -z "$DCP_POD" ]; then
  fail "DCP pod not found with label $APP_LABEL"
else
  oc get pod "$DCP_POD" -n "$DCP_NS" -o wide > "$EVIDENCE_DIR/12-dcp-pod-wide.txt" 2>&1 || true
  oc get pod "$DCP_POD" -n "$DCP_NS" -o jsonpath='{range .status.containerStatuses[*]}{.name}{" ready="}{.ready}{" restarts="}{.restartCount}{" image="}{.image}{"\n"}{end}' > "$EVIDENCE_DIR/13-dcp-container-statuses.txt" 2>&1 || true

  if grep -q 'devops-control-plane ready=true' "$EVIDENCE_DIR/13-dcp-container-statuses.txt" && grep -q 'oauth-proxy ready=true' "$EVIDENCE_DIR/13-dcp-container-statuses.txt"; then
    pass "DCP application and oauth-proxy containers are ready"
  else
    fail "One or more DCP containers are not ready"
  fi

  if awk -F'restarts=' '{print $2}' "$EVIDENCE_DIR/13-dcp-container-statuses.txt" | awk '{print $1}' | grep -vqE '^[0]+$'; then
    warn "DCP pod has one or more container restarts"
  else
    pass "DCP pod restart count is 0 for all containers"
  fi
fi

if [ -z "$PG_POD" ]; then
  fail "PostgreSQL pod not found with label $PG_LABEL"
else
  oc get pod "$PG_POD" -n "$DCP_NS" -o wide > "$EVIDENCE_DIR/14-postgresql-pod-wide.txt" 2>&1 || true
  oc get pod "$PG_POD" -n "$DCP_NS" -o jsonpath='{range .status.containerStatuses[*]}{.name}{" ready="}{.ready}{" restarts="}{.restartCount}{" image="}{.image}{"\n"}{end}' > "$EVIDENCE_DIR/15-postgresql-container-statuses.txt" 2>&1 || true

  if grep -q 'postgresql ready=true' "$EVIDENCE_DIR/15-postgresql-container-statuses.txt"; then
    pass "PostgreSQL container is ready"
  else
    fail "PostgreSQL container is not ready"
  fi

  if awk -F'restarts=' '{print $2}' "$EVIDENCE_DIR/15-postgresql-container-statuses.txt" | awk '{print $1}' | grep -vqE '^[0]+$'; then
    warn "PostgreSQL pod has one or more container restarts"
  else
    pass "PostgreSQL pod restart count is 0"
  fi
fi

info "Checking services, route and PVC"
if oc get svc "$DCP_SERVICE_NAME" -n "$DCP_NS" > "$EVIDENCE_DIR/16-dcp-service.txt" 2>&1; then
  pass "Service $DCP_SERVICE_NAME is present"
else
  fail "Service $DCP_SERVICE_NAME is missing"
fi

if oc get svc "$PG_SERVICE_NAME" -n "$DCP_NS" > "$EVIDENCE_DIR/17-postgresql-service.txt" 2>&1; then
  pass "Service $PG_SERVICE_NAME is present"
else
  fail "Service $PG_SERVICE_NAME is missing"
fi

if oc get route "$DCP_ROUTE_NAME" -n "$DCP_NS" > "$EVIDENCE_DIR/18-dcp-route.txt" 2>&1; then
  pass "Route $DCP_ROUTE_NAME is present"
  APP_ROUTE="$(oc get route "$DCP_ROUTE_NAME" -n "$DCP_NS" -o jsonpath='{.spec.host}')"
  echo "APP_ROUTE=$APP_ROUTE" > "$EVIDENCE_DIR/19-route-host.txt"
else
  fail "Route $DCP_ROUTE_NAME is missing"
fi

if oc get pvc postgresql-data -n "$DCP_NS" > "$EVIDENCE_DIR/20-postgresql-pvc.txt" 2>&1; then
  if grep -q 'Bound' "$EVIDENCE_DIR/20-postgresql-pvc.txt"; then
    pass "PVC postgresql-data is Bound"
  else
    fail "PVC postgresql-data is not Bound"
  fi
else
  fail "PVC postgresql-data is missing"
fi

info "Collecting events and logs"
oc get events -n "$DCP_NS" --sort-by=.lastTimestamp > "$EVIDENCE_DIR/21-events.txt" 2>&1 || true
tail -n 80 "$EVIDENCE_DIR/21-events.txt" > "$EVIDENCE_DIR/22-events-tail.txt" 2>/dev/null || true

if grep -Eiq 'BackOff|Failed|FailedMount|FailedScheduling|Unhealthy|OOMKilled|CrashLoopBackOff|ImagePullBackOff|ErrImagePull' "$EVIDENCE_DIR/21-events.txt"; then
  fail "Namespace has relevant warning/error events"
else
  pass "No relevant warning/error namespace events detected"
fi

if [ -n "$DCP_POD" ]; then
  oc logs "$DCP_POD" -n "$DCP_NS" -c devops-control-plane --tail=120 > "$EVIDENCE_DIR/23-dcp-app-logs-tail.txt" 2>&1 || true
  oc logs "$DCP_POD" -n "$DCP_NS" -c oauth-proxy --tail=120 > "$EVIDENCE_DIR/24-oauth-proxy-logs-tail.txt" 2>&1 || true
  grep -Ei 'error|panic|fatal|timeout|refused|denied|unauthorized|forbidden|x509|tls' "$EVIDENCE_DIR/23-dcp-app-logs-tail.txt" > "$EVIDENCE_DIR/25-dcp-app-log-warnings.txt" 2>/dev/null || true
  grep -Ei 'error|panic|fatal|timeout|refused|denied|unauthorized|forbidden|x509|tls' "$EVIDENCE_DIR/24-oauth-proxy-logs-tail.txt" > "$EVIDENCE_DIR/26-oauth-proxy-log-warnings.txt" 2>/dev/null || true

  if grep -Eiq 'panic|fatal|x509|connection refused|database|could not|timeout' "$EVIDENCE_DIR/25-dcp-app-log-warnings.txt"; then
    fail "DCP application logs contain relevant error patterns"
  else
    pass "DCP application logs do not contain relevant fatal/error patterns"
  fi

  if grep -Eiq 'panic|fatal|x509|connection refused|upstream unavailable|timeout' "$EVIDENCE_DIR/26-oauth-proxy-log-warnings.txt"; then
    fail "OAuth Proxy logs contain relevant error patterns"
  else
    pass "OAuth Proxy logs do not contain relevant fatal/error patterns"
  fi
fi

if [ -n "$PG_POD" ]; then
  oc logs "$PG_POD" -n "$DCP_NS" --tail=120 > "$EVIDENCE_DIR/27-postgresql-logs-tail.txt" 2>&1 || true
  grep -Ei 'error|fatal|panic|could not|timeout|refused|denied|recovery' "$EVIDENCE_DIR/27-postgresql-logs-tail.txt" > "$EVIDENCE_DIR/28-postgresql-log-warnings.txt" 2>/dev/null || true

  if grep -Eiq 'fatal|panic|could not|timeout|refused|denied|recovery' "$EVIDENCE_DIR/28-postgresql-log-warnings.txt"; then
    fail "PostgreSQL logs contain relevant error patterns"
  else
    pass "PostgreSQL logs do not contain relevant fatal/error patterns"
  fi
fi

info "Collecting sanitized configuration"
oc get configmap devops-control-plane-config -n "$DCP_NS" -o json > "$EVIDENCE_DIR/29-configmap.raw.json" 2>/dev/null || true
if command -v jq >/dev/null 2>&1 && [ -s "$EVIDENCE_DIR/29-configmap.raw.json" ]; then
  jq -r '.data | to_entries[] | select(.key | test("TOKEN|PASSWORD|SECRET|DATABASE_URL") | not) | .key + "=" + .value' "$EVIDENCE_DIR/29-configmap.raw.json" | sort > "$EVIDENCE_DIR/30-configmap-sanitized.txt"
else
  warn "jq not available or ConfigMap unavailable; sanitized ConfigMap output not generated"
fi
rm -f "$EVIDENCE_DIR/29-configmap.raw.json"

oc get secret devops-control-plane-secrets -n "$DCP_NS" -o json > "$EVIDENCE_DIR/31-secret.raw.json" 2>/dev/null || true
if command -v jq >/dev/null 2>&1 && [ -s "$EVIDENCE_DIR/31-secret.raw.json" ]; then
  jq -r '.data | keys[]' "$EVIDENCE_DIR/31-secret.raw.json" | sort > "$EVIDENCE_DIR/32-secret-keys-only.txt"
  jq -r '.data | to_entries[] | .key + " base64_length=" + (.value | length | tostring)' "$EVIDENCE_DIR/31-secret.raw.json" | sort > "$EVIDENCE_DIR/33-secret-key-lengths-only.txt"
  if grep -q '^KUBERNETES_TOKEN$' "$EVIDENCE_DIR/32-secret-keys-only.txt"; then
    warn "KUBERNETES_TOKEN still present in application Secret"
  else
    pass "KUBERNETES_TOKEN not present in application Secret"
  fi
else
  warn "jq not available or Secret unavailable; Secret key-only inventory not generated"
fi
rm -f "$EVIDENCE_DIR/31-secret.raw.json"

if [ "$SKIP_TOP" != "true" ]; then
  info "Collecting pod metrics"
  if oc adm top pod -n "$DCP_NS" > "$EVIDENCE_DIR/34-top-pods.txt" 2>&1; then
    pass "Pod metrics available through oc adm top"
  else
    warn "Pod metrics unavailable through oc adm top"
  fi
fi

if [ -n "$APP_ROUTE" ]; then
  info "Testing Route endpoints"
  ROUTE_READYZ_CODE="$(http_status "https://$APP_ROUTE/readyz" "$EVIDENCE_DIR/35-route-readyz-body.txt")"
  echo "route readyz HTTP $ROUTE_READYZ_CODE" > "$EVIDENCE_DIR/35-route-readyz-status.txt"
  check_status_equals "Route /readyz" "$ROUTE_READYZ_CODE" "200"

  ROUTE_LIVEZ_CODE="$(http_status "https://$APP_ROUTE/livez" "$EVIDENCE_DIR/36-route-livez-body.txt")"
  echo "route livez HTTP $ROUTE_LIVEZ_CODE" > "$EVIDENCE_DIR/36-route-livez-status.txt"
  check_status_equals "Route /livez" "$ROUTE_LIVEZ_CODE" "200"

  ROUTE_API_ANON_CODE="$(http_status "https://$APP_ROUTE/api/v1/changes" "$EVIDENCE_DIR/37-route-api-changes-anonymous-body.txt")"
  echo "route api changes anonymous HTTP $ROUTE_API_ANON_CODE" > "$EVIDENCE_DIR/37-route-api-changes-anonymous-status.txt"
  check_status_equals "Route anonymous /api/v1/changes" "$ROUTE_API_ANON_CODE" "403"

  ROUTE_UI_ANON_CODE="$(http_status "https://$APP_ROUTE/ui/dashboard" "$EVIDENCE_DIR/38-route-ui-dashboard-anonymous-body.txt")"
  echo "route ui dashboard anonymous HTTP $ROUTE_UI_ANON_CODE" > "$EVIDENCE_DIR/38-route-ui-dashboard-anonymous-status.txt"
  check_status_equals "Route anonymous /ui/dashboard" "$ROUTE_UI_ANON_CODE" "403"
fi

if [ -n "$DCP_POD" ]; then
  info "Testing backend endpoints through port-forward"
  oc port-forward -n "$DCP_NS" pod/"$DCP_POD" "$LOCAL_PORT:$BACKEND_PORT" > "$EVIDENCE_DIR/39-port-forward.log" 2>&1 &
  PF_PID=$!
  echo "PF_PID=$PF_PID" > "$EVIDENCE_DIR/40-port-forward-pid.txt"
  sleep 3

  if ! ps -p "$PF_PID" >/dev/null 2>&1; then
    fail "Port-forward failed to start"
  else
    BACKEND_READYZ_CODE="$(http_status "http://127.0.0.1:$LOCAL_PORT/readyz" "$EVIDENCE_DIR/41-backend-readyz-body.txt")"
    echo "backend readyz HTTP $BACKEND_READYZ_CODE" > "$EVIDENCE_DIR/41-backend-readyz-status.txt"
    check_status_equals "Backend /readyz" "$BACKEND_READYZ_CODE" "200"

    BACKEND_LIVEZ_CODE="$(http_status "http://127.0.0.1:$LOCAL_PORT/livez" "$EVIDENCE_DIR/42-backend-livez-body.txt")"
    echo "backend livez HTTP $BACKEND_LIVEZ_CODE" > "$EVIDENCE_DIR/42-backend-livez-status.txt"
    check_status_equals "Backend /livez" "$BACKEND_LIVEZ_CODE" "200"

    BACKEND_API_ANON_CODE="$(http_status "http://127.0.0.1:$LOCAL_PORT/api/v1/changes" "$EVIDENCE_DIR/43-backend-api-changes-anonymous-body.txt")"
    echo "backend api changes anonymous HTTP $BACKEND_API_ANON_CODE" > "$EVIDENCE_DIR/43-backend-api-changes-anonymous-status.txt"
    check_status_in "Backend anonymous /api/v1/changes" "$BACKEND_API_ANON_CODE" "401" "403"

    BACKEND_API_ADMIN_CODE="$(http_status "http://127.0.0.1:$LOCAL_PORT/api/v1/changes" "$EVIDENCE_DIR/44-backend-api-changes-admin-body.json" -H "X-Forwarded-User: $TEST_USER" -H "X-Forwarded-Groups: $TEST_GROUP")"
    echo "backend api changes admin HTTP $BACKEND_API_ADMIN_CODE" > "$EVIDENCE_DIR/44-backend-api-changes-admin-status.txt"
    check_status_equals "Backend admin-header /api/v1/changes" "$BACKEND_API_ADMIN_CODE" "200"
  fi
fi

if [ "$SKIP_PG" != "true" ] && [ -n "$PG_POD" ]; then
  info "Testing PostgreSQL connectivity and counts"
  if oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select now() as db_time, current_database() as database_name, current_user as user_name;"' > "$EVIDENCE_DIR/45-postgresql-connectivity.txt" 2>&1; then
    pass "PostgreSQL connectivity query succeeded"
  else
    fail "PostgreSQL connectivity query failed"
  fi

  if oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as change_requests_count from change_requests;"' > "$EVIDENCE_DIR/46-postgresql-change-requests-count.txt" 2>&1; then
    pass "PostgreSQL change_requests count query succeeded"
  else
    fail "PostgreSQL change_requests count query failed"
  fi

  if oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as change_events_count from change_events;"' > "$EVIDENCE_DIR/47-postgresql-change-events-count.txt" 2>&1; then
    pass "PostgreSQL change_events count query succeeded"
  else
    fail "PostgreSQL change_events count query failed"
  fi

  if oc exec -n "$DCP_NS" "$PG_POD" -- bash -lc 'psql "$POSTGRESQL_DATABASE" -c "select count(*) as evidences_count from evidences;"' > "$EVIDENCE_DIR/48-postgresql-evidences-count.txt" 2>&1; then
    pass "PostgreSQL evidences count query succeeded"
  else
    fail "PostgreSQL evidences count query failed"
  fi
fi

info "Checking NetworkPolicy and RBAC inventory"
oc get networkpolicy -n "$DCP_NS" -o wide > "$EVIDENCE_DIR/49-networkpolicies.txt" 2>&1 || true
if grep -q 'postgresql-ingress-from-devops-control-plane' "$EVIDENCE_DIR/49-networkpolicies.txt"; then
  pass "PostgreSQL ingress NetworkPolicy is present"
else
  warn "PostgreSQL ingress NetworkPolicy not found"
fi

oc get rolebinding -n "$DCP_NS" -o wide > "$EVIDENCE_DIR/50-rolebindings-dcp-namespace.txt" 2>&1 || true
oc get rolebinding -n devops-ci-demo -o wide > "$EVIDENCE_DIR/51-rolebindings-devops-ci-demo.txt" 2>&1 || true
oc get clusterrolebinding > "$EVIDENCE_DIR/52-clusterrolebindings-all.txt" 2>&1 || true
grep -E 'devops-control-plane|oauth-proxy|group-reader|auth-delegator' "$EVIDENCE_DIR/52-clusterrolebindings-all.txt" > "$EVIDENCE_DIR/53-clusterrolebindings-filtered.txt" 2>/dev/null || true

if grep -q 'devops-control-plane-group-reader' "$EVIDENCE_DIR/53-clusterrolebindings-filtered.txt"; then
  pass "ClusterRoleBinding devops-control-plane-group-reader is present"
else
  fail "ClusterRoleBinding devops-control-plane-group-reader is missing"
fi

if [ "$SKIP_CAN_I" != "true" ]; then
  DCP_SA="system:serviceaccount:$DCP_NS:devops-control-plane-oauth-proxy"
  oc auth can-i create pipelineruns.tekton.dev -n devops-ci-demo --as "$DCP_SA" > "$EVIDENCE_DIR/54-can-i-create-pipelineruns.txt" 2>&1 || true
  oc auth can-i get deployments -n devops-ci-demo --as "$DCP_SA" > "$EVIDENCE_DIR/55-can-i-get-deployments.txt" 2>&1 || true
  oc auth can-i list pods -n devops-ci-demo --as "$DCP_SA" > "$EVIDENCE_DIR/56-can-i-list-pods.txt" 2>&1 || true
  oc auth can-i get secrets -n "$DCP_NS" --as "$DCP_SA" > "$EVIDENCE_DIR/57-can-i-get-secrets-dcp.txt" 2>&1 || true

  grep -qx 'yes' "$EVIDENCE_DIR/54-can-i-create-pipelineruns.txt" && pass "can-i create PipelineRuns: yes" || fail "can-i create PipelineRuns: not yes"
  grep -qx 'yes' "$EVIDENCE_DIR/55-can-i-get-deployments.txt" && pass "can-i get deployments: yes" || fail "can-i get deployments: not yes"
  grep -qx 'yes' "$EVIDENCE_DIR/56-can-i-list-pods.txt" && pass "can-i list pods: yes" || fail "can-i list pods: not yes"
  grep -qx 'no' "$EVIDENCE_DIR/57-can-i-get-secrets-dcp.txt" && pass "can-i get DCP secrets: no" || fail "can-i get DCP secrets: not no"
fi

{
  echo "=== DevOps Control Plane operability smoke-test summary ==="
  echo "Timestamp: $(date -Is)"
  echo "Namespace: $DCP_NS"
  echo "DCP pod: ${DCP_POD:-not-found}"
  echo "PostgreSQL pod: ${PG_POD:-not-found}"
  echo "Route: ${APP_ROUTE:-not-found}"
  echo
  echo "Results:"
  echo "PASS=$PASS_COUNT"
  echo "WARN=$WARN_COUNT"
  echo "FAIL=$FAIL_COUNT"
  echo
  echo "Route smoke:"
  cat "$EVIDENCE_DIR/35-route-readyz-status.txt" 2>/dev/null || true
  cat "$EVIDENCE_DIR/36-route-livez-status.txt" 2>/dev/null || true
  cat "$EVIDENCE_DIR/37-route-api-changes-anonymous-status.txt" 2>/dev/null || true
  cat "$EVIDENCE_DIR/38-route-ui-dashboard-anonymous-status.txt" 2>/dev/null || true
  echo
  echo "Backend smoke:"
  cat "$EVIDENCE_DIR/41-backend-readyz-status.txt" 2>/dev/null || true
  cat "$EVIDENCE_DIR/42-backend-livez-status.txt" 2>/dev/null || true
  cat "$EVIDENCE_DIR/43-backend-api-changes-anonymous-status.txt" 2>/dev/null || true
  cat "$EVIDENCE_DIR/44-backend-api-changes-admin-status.txt" 2>/dev/null || true
  echo
  echo "Evidence directory:"
  echo "$EVIDENCE_DIR"
} | tee "$EVIDENCE_DIR/99-summary.txt"

if [ "$FAIL_COUNT" -eq 0 ]; then
  info "Smoke test completed successfully"
  exit 0
fi

info "Smoke test completed with failures"
exit 1
