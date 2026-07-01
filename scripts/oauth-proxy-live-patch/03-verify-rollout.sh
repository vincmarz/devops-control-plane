#!/usr/bin/env bash
set -euo pipefail

NS="${NS:-devops-control-plane}"
APP="${APP:-devops-control-plane}"

info() { printf '\n== %s ==\n' "$*"; }

info "Pods"
oc -n "$NS" get pods -l app="$APP" -o wide

POD="$(oc -n "$NS" get pod -l app="$APP" -o jsonpath='{.items[0].metadata.name}')"
echo "POD=${POD}"

info "Containers"
oc -n "$NS" get pod "$POD" -o jsonpath='{range .spec.containers[*]}{.name}{"\n"}{end}'

info "Container statuses"
oc -n "$NS" get pod "$POD" -o jsonpath='{range .status.containerStatuses[*]}{.name}{" ready="}{.ready}{" restartCount="}{.restartCount}{"\n"}{end}'

info "Deployment"
oc -n "$NS" get deployment "$APP" -o jsonpath='serviceAccountName={.spec.template.spec.serviceAccountName}{" paused="}{.spec.paused}{" replicas="}{.spec.replicas}{"\n"}'

info "AUTH_ENABLED"
oc -n "$NS" get configmap ${APP}-config -o jsonpath='AUTH_ENABLED={.data.AUTH_ENABLED}{"\n"}'

info "Service"
oc -n "$NS" get service "$APP" -o jsonpath='{range .spec.ports[*]}{.name}{" "}{.port}{" -> "}{.targetPort}{"\n"}{end}'

info "Route"
oc -n "$NS" get route "$APP" -o jsonpath='targetPort={.spec.port.targetPort}{" termination="}{.spec.tls.termination}{" insecure="}{.spec.tls.insecureEdgeTerminationPolicy}{"\n"}'

info "OAuth Proxy logs tail"
oc -n "$NS" logs "$POD" -c oauth-proxy --tail=80 || true

info "Application logs tail"
oc -n "$NS" logs "$POD" -c devops-control-plane --tail=80 || true

info "Route curl headers, OAuth redirect/challenge is expected without browser session"
APP_ROUTE="$(oc -n "$NS" get route "$APP" -o jsonpath='{.spec.host}')"
echo "https://${APP_ROUTE}"
curl -k -I "https://${APP_ROUTE}/ui/dashboard" || true
curl -k -I "https://${APP_ROUTE}/readyz" || true
