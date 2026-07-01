#!/usr/bin/env bash
set -euo pipefail

NS="${NS:-devops-control-plane}"
APP="${APP:-devops-control-plane}"
SA="${SA:-devops-control-plane-oauth-proxy}"
COOKIE_SECRET_NAME="${COOKIE_SECRET_NAME:-devops-control-plane-oauth-proxy-cookie}"

info() { printf '\n== %s ==\n' "$*"; }

info "Apply OAuth Proxy ServiceAccount only"
cat <<EOF | oc apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${SA}
  namespace: ${NS}
  labels:
    app.kubernetes.io/name: ${APP}
    app.kubernetes.io/component: oauth-proxy
  annotations:
    serviceaccounts.openshift.io/oauth-redirectreference.primary: '{"kind":"OAuthRedirectReference","apiVersion":"v1","reference":{"kind":"Route","name":"${APP}"}}'
EOF

info "Apply namespaced RoleBinding for existing Tekton runtime Role only"
cat <<EOF | oc apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: devops-control-plane-oauth-proxy-tekton-runtime
  namespace: ${NS}
  labels:
    app.kubernetes.io/name: ${APP}
    app.kubernetes.io/component: oauth-proxy
subjects:
  - kind: ServiceAccount
    name: ${SA}
    namespace: ${NS}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: devops-control-plane-tekton-runtime
EOF

info "Ensure auth-delegator grant"
oc adm policy add-role-to-user \
  system:auth-delegator \
  system:serviceaccount:${NS}:${SA}

info "Verify auth delegation"
oc auth can-i create tokenreviews.authentication.k8s.io \
  --as=system:serviceaccount:${NS}:${SA}

oc auth can-i create subjectaccessreviews.authorization.k8s.io \
  --as=system:serviceaccount:${NS}:${SA}

info "Ensure OAuth Proxy cookie secret exists without printing value"
if oc -n "$NS" get secret "$COOKIE_SECRET_NAME" >/dev/null 2>&1; then
  echo "Cookie secret already exists: ${COOKIE_SECRET_NAME}"
else
  set +x
  COOKIE_SECRET="$(openssl rand -base64 48 | tr -dc 'A-Za-z0-9' | head -c 32)"
  oc create secret generic "$COOKIE_SECRET_NAME" \
    -n "$NS" \
    --from-literal=cookie-secret="${COOKIE_SECRET}" \
    --dry-run=client -o yaml \
    | oc apply -f -
  unset COOKIE_SECRET
fi

info "Cookie secret keys, values are not printed"
oc -n "$NS" get secret "$COOKIE_SECRET_NAME" -o go-template='{{range $k,$v := .data}}{{printf "%s\n" $k}}{{end}}'
