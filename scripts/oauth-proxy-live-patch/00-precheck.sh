#!/usr/bin/env bash
set -euo pipefail

NS="${NS:-devops-control-plane}"
APP="${APP:-devops-control-plane}"

info() { printf '\n== %s ==\n' "$*"; }
fail() { echo "ERROR: $*" >&2; exit 1; }

info "Current pods"
oc -n "$NS" get pods -l app="$APP" -o wide

info "Deployment baseline"
oc -n "$NS" get deployment "$APP" -o jsonpath='serviceAccountName={.spec.template.spec.serviceAccountName}{" paused="}{.spec.paused}{" replicas="}{.spec.replicas}{"\n"}'

info "AUTH_ENABLED"
AUTH_ENABLED="$(oc -n "$NS" get configmap ${APP}-config -o jsonpath='{.data.AUTH_ENABLED}')"
echo "AUTH_ENABLED=${AUTH_ENABLED}"
[[ "$AUTH_ENABLED" == "false" ]] || fail "AUTH_ENABLED must be false before staged OAuth Proxy rollout"

info "Service"
oc -n "$NS" get service "$APP" -o jsonpath='{range .spec.ports[*]}{.name}{" "}{.port}{" -> "}{.targetPort}{"\n"}{end}'

info "Route"
oc -n "$NS" get route "$APP" -o jsonpath='targetPort={.spec.port.targetPort}{" termination="}{.spec.tls.termination}{" insecure="}{.spec.tls.insecureEdgeTerminationPolicy}{"\n"}'

info "Application Secret lengths, values are not printed"
oc -n "$NS" get secret ${APP}-secrets -o go-template='{{range $k,$v := .data}}{{printf "%s base64_len=%d\n" $k (len $v)}}{{end}}'

info "PostgreSQL Secret lengths, values are not printed"
oc -n "$NS" get secret postgresql-secret -o go-template='{{range $k,$v := .data}}{{printf "%s base64_len=%d\n" $k (len $v)}}{{end}}'

info "Trust bundle validation"
TMP_CA="/tmp/dcp-app-trust-bundle-precheck.crt"
oc -n "$NS" get configmap dcp-app-trust-bundle -o jsonpath='{.data.ca-bundle\.crt}' > "$TMP_CA"
wc -c "$TMP_CA"
BEGIN_COUNT="$(grep -c 'BEGIN CERTIFICATE' "$TMP_CA" || true)"
echo "BEGIN_CERTIFICATE_COUNT=${BEGIN_COUNT}"
FIRST_LINE="$(head -n 1 "$TMP_CA" || true)"
echo "FIRST_LINE=${FIRST_LINE}"
[[ "$BEGIN_COUNT" -ge 1 ]] || fail "dcp-app-trust-bundle does not contain a PEM certificate"
[[ "$FIRST_LINE" == "-----BEGIN CERTIFICATE-----" ]] || fail "dcp-app-trust-bundle first line is not a PEM certificate header"
openssl x509 -in "$TMP_CA" -noout -subject -issuer -fingerprint -sha256 >/tmp/dcp-app-trust-bundle-precheck.openssl
cat /tmp/dcp-app-trust-bundle-precheck.openssl

info "Forbidden live mutation reminder"
echo "This procedure must not apply kind: Secret, PostgreSQL resources, or dcp-app-trust-bundle."
