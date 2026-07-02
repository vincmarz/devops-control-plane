#!/usr/bin/env bash
set -euo pipefail

NS="${NS:-devops-control-plane}"
APP="${APP:-devops-control-plane}"
SA="${SA:-devops-control-plane-oauth-proxy}"
TARGET_NS="${TARGET_NS:-devops-ci-demo}"
ROLE_NAME="${ROLE_NAME:-devops-control-plane-tekton-runtime}"
COOKIE_SECRET_NAME="${COOKIE_SECRET_NAME:-devops-control-plane-oauth-proxy-cookie}"

info() { printf '
== %s ==
' "$*"; }

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

info "Apply least-privilege runtime Role in target namespace"
cat <<EOF | oc apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: ${ROLE_NAME}
  namespace: ${TARGET_NS}
  labels:
    app.kubernetes.io/name: ${APP}
    app.kubernetes.io/component: runtime-rbac
rules:
  - apiGroups:
      - tekton.dev
    resources:
      - pipelineruns
    verbs:
      - create
      - get
      - list
  - apiGroups:
      - tekton.dev
    resources:
      - taskruns
    verbs:
      - get
      - list
  - apiGroups:
      - apps
    resources:
      - deployments
    verbs:
      - get
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - get
      - list
  - apiGroups:
      - ""
    resources:
      - services
    verbs:
      - get
  - apiGroups:
      - route.openshift.io
    resources:
      - routes
    verbs:
      - get
EOF

info "Apply RoleBinding in target namespace for OAuth Proxy runtime ServiceAccount"
cat <<EOF | oc apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ${ROLE_NAME}
  namespace: ${TARGET_NS}
  labels:
    app.kubernetes.io/name: ${APP}
    app.kubernetes.io/component: runtime-rbac
subjects:
  - kind: ServiceAccount
    name: ${SA}
    namespace: ${NS}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ${ROLE_NAME}
EOF

info "Remove obsolete dangling RoleBinding from application namespace if present"
oc delete rolebinding devops-control-plane-oauth-proxy-tekton-runtime   -n "${NS}"   --ignore-not-found=true

info "Ensure auth-delegator grant"
oc adm policy add-role-to-user   system:auth-delegator   system:serviceaccount:${NS}:${SA}

info "Verify auth delegation"
oc auth can-i create tokenreviews.authentication.k8s.io   --as=system:serviceaccount:${NS}:${SA}

oc auth can-i create subjectaccessreviews.authorization.k8s.io   --as=system:serviceaccount:${NS}:${SA}

info "Verify target namespace runtime permissions"
oc auth can-i create pipelineruns.tekton.dev   -n "${TARGET_NS}"   --as=system:serviceaccount:${NS}:${SA}

oc auth can-i get deployments.apps   -n "${TARGET_NS}"   --as=system:serviceaccount:${NS}:${SA}

oc auth can-i list pods   -n "${TARGET_NS}"   --as=system:serviceaccount:${NS}:${SA}

info "Ensure OAuth Proxy cookie secret exists without printing value"
if oc -n "$NS" get secret "$COOKIE_SECRET_NAME" >/dev/null 2>&1; then
  echo "Cookie secret already exists: ${COOKIE_SECRET_NAME}"
else
  set +x
  COOKIE_SECRET="$(openssl rand -base64 48 | tr -dc 'A-Za-z0-9' | head -c 32)"
  oc create secret generic "$COOKIE_SECRET_NAME"     -n "$NS"     --from-literal=cookie-secret="${COOKIE_SECRET}"     --dry-run=client -o yaml     | oc apply -f -
  unset COOKIE_SECRET
fi

info "Cookie secret keys, values are not printed"
oc -n "$NS" get secret "$COOKIE_SECRET_NAME" -o go-template='{{range $k,$v := .data}}{{printf "%s
" $k}}{{end}}'
