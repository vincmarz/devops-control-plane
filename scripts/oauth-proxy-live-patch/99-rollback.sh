#!/usr/bin/env bash
set -euo pipefail

NS="${NS:-devops-control-plane}"
APP="${APP:-devops-control-plane}"

info() { printf '\n== %s ==\n' "$*"; }

info "Restore Service baseline http 8080 -> http"
oc -n "$NS" patch service "$APP" --type=json -p='[
  {"op":"remove","path":"/metadata/annotations/service.beta.openshift.io~1serving-cert-secret-name"},
  {"op":"replace","path":"/spec/ports","value":[{"name":"http","port":8080,"protocol":"TCP","targetPort":"http"}]}
]' || \
oc -n "$NS" patch service "$APP" --type=json -p='[
  {"op":"replace","path":"/spec/ports","value":[{"name":"http","port":8080,"protocol":"TCP","targetPort":"http"}]}
]'

info "Restore Route baseline edge -> http"
oc -n "$NS" patch route "$APP" --type=merge -p='{"spec":{"port":{"targetPort":"http"},"tls":{"termination":"edge","insecureEdgeTerminationPolicy":null}}}'

info "Pause deployment to stop churn"
oc -n "$NS" rollout pause deployment/"$APP" || true

info "Remove OAuth Proxy sidecar and TLS volume; restore application ServiceAccount"
oc -n "$NS" patch deployment "$APP" --type=strategic -p='{
  "spec": {
    "template": {
      "metadata": {
        "annotations": {
          "devops-control-plane/oauth-proxy": null
        }
      },
      "spec": {
        "serviceAccountName": "devops-control-plane",
        "containers": [
          {
            "name": "oauth-proxy",
            "$patch": "delete"
          }
        ],
        "volumes": [
          {
            "name": "oauth-proxy-tls",
            "$patch": "delete"
          }
        ]
      }
    }
  }
}'

info "Resume and wait rollout"
oc -n "$NS" rollout resume deployment/"$APP" || true
oc -n "$NS" rollout status deployment/"$APP" --timeout=180s || true

info "Final baseline verification"
oc -n "$NS" get pods -l app="$APP" -o wide
oc -n "$NS" get deployment "$APP" -o jsonpath='serviceAccountName={.spec.template.spec.serviceAccountName}{" paused="}{.spec.paused}{" replicas="}{.spec.replicas}{"\n"}'
oc -n "$NS" get service "$APP" -o jsonpath='{range .spec.ports[*]}{.name}{" "}{.port}{" -> "}{.targetPort}{"\n"}{end}'
oc -n "$NS" get route "$APP" -o jsonpath='targetPort={.spec.port.targetPort}{" termination="}{.spec.tls.termination}{" insecure="}{.spec.tls.insecureEdgeTerminationPolicy}{"\n"}'
oc -n "$NS" get configmap ${APP}-config -o jsonpath='AUTH_ENABLED={.data.AUTH_ENABLED}{"\n"}'
