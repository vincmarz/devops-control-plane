#!/usr/bin/env bash
set -euo pipefail

NS="${NS:-devops-control-plane}"
APP="${APP:-devops-control-plane}"
SA="${SA:-devops-control-plane-oauth-proxy}"
TLS_SECRET="${TLS_SECRET:-devops-control-plane-oauth-proxy-tls}"
COOKIE_SECRET_NAME="${COOKIE_SECRET_NAME:-devops-control-plane-oauth-proxy-cookie}"
OAUTH_IMAGE="${OAUTH_IMAGE:-registry.redhat.io/openshift4/ose-oauth-proxy:latest}"

info() { printf '\n== %s ==\n' "$*"; }
fail() { echo "ERROR: $*" >&2; exit 1; }

info "Confirm AUTH_ENABLED=false"
AUTH_ENABLED="$(oc -n "$NS" get configmap ${APP}-config -o jsonpath='{.data.AUTH_ENABLED}')"
echo "AUTH_ENABLED=${AUTH_ENABLED}"
[[ "$AUTH_ENABLED" == "false" ]] || fail "AUTH_ENABLED must remain false for this staged rollout"

info "Add serving-cert annotation to Service while keeping current http port"
oc -n "$NS" annotate service "$APP" \
  service.beta.openshift.io/serving-cert-secret-name="$TLS_SECRET" \
  --overwrite

info "Wait for serving cert Secret"
for i in $(seq 1 60); do
  if oc -n "$NS" get secret "$TLS_SECRET" >/dev/null 2>&1; then
    KEYS="$(oc -n "$NS" get secret "$TLS_SECRET" -o go-template='{{range $k,$v := .data}}{{printf "%s\n" $k}}{{end}}' | sort | tr '\n' ' ')"
    echo "TLS_SECRET_KEYS=${KEYS}"
    echo "$KEYS" | grep -q 'tls.crt' && echo "$KEYS" | grep -q 'tls.key' && break
  fi
  sleep 2
  [[ "$i" -lt 60 ]] || fail "TLS secret ${TLS_SECRET} not ready"
done

info "Patch Deployment with OAuth Proxy sidecar and Pod ServiceAccount"
oc -n "$NS" patch deployment "$APP" --type=strategic -p "$(cat <<EOF
{
  "spec": {
    "template": {
      "metadata": {
        "annotations": {
          "devops-control-plane/oauth-proxy": "enabled"
        }
      },
      "spec": {
        "serviceAccountName": "${SA}",
        "containers": [
          {
            "name": "oauth-proxy",
            "image": "${OAUTH_IMAGE}",
            "imagePullPolicy": "IfNotPresent",
            "ports": [
              {
                "name": "oauth-proxy",
                "containerPort": 8443,
                "protocol": "TCP"
              }
            ],
            "env": [
              {
                "name": "OAUTH_PROXY_COOKIE_SECRET",
                "valueFrom": {
                  "secretKeyRef": {
                    "name": "${COOKIE_SECRET_NAME}",
                    "key": "cookie-secret"
                  }
                }
              }
            ],
            "args": [
              "--provider=openshift",
              "--openshift-service-account=${SA}",
              "--https-address=:8443",
              "--upstream=http://localhost:8080",
              "--tls-cert=/etc/tls/private/tls.crt",
              "--tls-key=/etc/tls/private/tls.key",
              "--cookie-secret=\$(OAUTH_PROXY_COOKIE_SECRET)",
              "--set-xauthrequest=true",
              "--pass-user-bearer-token=false",
              "--pass-access-token=false",
              "--pass-basic-auth=false"
            ],
            "readinessProbe": {
              "httpGet": {
                "scheme": "HTTPS",
                "path": "/oauth/healthz",
                "port": "oauth-proxy"
              },
              "initialDelaySeconds": 5,
              "periodSeconds": 10
            },
            "livenessProbe": {
              "httpGet": {
                "scheme": "HTTPS",
                "path": "/oauth/healthz",
                "port": "oauth-proxy"
              },
              "initialDelaySeconds": 10,
              "periodSeconds": 30
            },
            "volumeMounts": [
              {
                "name": "oauth-proxy-tls",
                "mountPath": "/etc/tls/private",
                "readOnly": true
              }
            ]
          }
        ],
        "volumes": [
          {
            "name": "oauth-proxy-tls",
            "secret": {
              "secretName": "${TLS_SECRET}"
            }
          }
        ]
      }
    }
  }
}
EOF
)"

info "Wait for Deployment rollout with both containers"
oc -n "$NS" rollout status deployment/"$APP" --timeout=180s

info "Verify new Pod containers"
POD="$(oc -n "$NS" get pod -l app="$APP" -o jsonpath='{.items[0].metadata.name}')"
echo "POD=${POD}"
oc -n "$NS" get pod "$POD" -o jsonpath='{range .spec.containers[*]}{.name}{"\n"}{end}'

info "Switch Service to OAuth Proxy port only"
oc -n "$NS" patch service "$APP" --type=json -p='[
  {"op":"replace","path":"/spec/ports","value":[{"name":"https","port":8443,"protocol":"TCP","targetPort":"oauth-proxy"}]}
]'

info "Switch Route to reencrypt and target https"
oc -n "$NS" patch route "$APP" --type=merge -p='{"spec":{"port":{"targetPort":"https"},"tls":{"termination":"reencrypt","insecureEdgeTerminationPolicy":"Redirect"}}}'

info "Live-patch apply completed"
