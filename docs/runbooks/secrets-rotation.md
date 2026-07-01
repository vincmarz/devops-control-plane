# Secrets Management and Rotation Runbook

## Scope

This runbook describes how to inventory, validate, rotate and troubleshoot the runtime secrets used by the DevOps Control Plane.

The intended audience includes operators who are new to the project. Commands are intentionally explicit and avoid printing secret values.

## Security rule

Never print tokens, passwords or connection strings in clear text.

Allowed checks:

- list secret key names;
- print base64 string length;
- print HTTP status codes;
- print non-sensitive metadata returned by APIs;
- store temporary tokens in files under `/tmp` with restricted permissions and delete them after use.

Forbidden checks:

- echo token values;
- paste token values into chat, tickets or documentation;
- commit real tokens or real database URLs;
- add real secret manifests to Git.

## Runtime secret inventory

The application secret is `devops-control-plane-secrets` in namespace `devops-control-plane`.

Expected keys:

```text
ARGOCD_AUTH_TOKEN
DATABASE_URL
GITLAB_TOKEN
KUBERNETES_TOKEN
```

List the keys without printing values:

```bash
oc get secret devops-control-plane-secrets \
  -n devops-control-plane \
  -o go-template='{{range $k,$v := .data}}{{printf "%s\n" $k}}{{end}}' \
  | sort
```

Check value lengths without decoding or printing secrets:

```bash
for K in ARGOCD_AUTH_TOKEN DATABASE_URL GITLAB_TOKEN KUBERNETES_TOKEN; do
  LEN="$(
    oc get secret devops-control-plane-secrets \
      -n devops-control-plane \
      -o jsonpath="{.data.${K}}" 2>/dev/null \
    | wc -c
  )"
  echo "${K}|base64_chars=${LEN}"
done
```

Validated baseline during Phase 9.2:

```text
ARGOCD_AUTH_TOKEN|base64_chars=344
DATABASE_URL|base64_chars=184
GITLAB_TOKEN|base64_chars=68
KUBERNETES_TOKEN|base64_chars=68
```

## Deployment loading model

The deployment loads configuration through `envFrom`:

```text
devops-control-plane-config
devops-control-plane-secrets
```

Check it with:

```bash
oc get deploy devops-control-plane \
  -n devops-control-plane \
  -o jsonpath='{range .spec.template.spec.containers[0].envFrom[*]}{.configMapRef.name}{.secretRef.name}{"\n"}{end}'
```

The deployment also mounts the app-dedicated trust bundle:

```text
ConfigMap: dcp-app-trust-bundle
Mount: /etc/dcp-trust/ca-bundle.crt
```

## Token health checks

### Argo CD token health check

This check validates `ARGOCD_AUTH_TOKEN` without printing it.

```bash
ARGOCD_HOST="$(oc get route openshift-gitops-server -n openshift-gitops -o jsonpath='{.spec.host}')"

ARGOCD_TOKEN="$(
  oc get secret devops-control-plane-secrets \
    -n devops-control-plane \
    -o jsonpath='{.data.ARGOCD_AUTH_TOKEN}' \
  | base64 -d
)"

HTTP_CODE="$(
  curl -k -sS \
    -o /tmp/argocd-token-check.json \
    -w '%{http_code}' \
    -H "Authorization: Bearer ${ARGOCD_TOKEN}" \
    "https://${ARGOCD_HOST}/api/v1/applications/demo-go-color-app"
)"

unset ARGOCD_TOKEN

echo "ARGOCD_TOKEN_HTTP_CODE=${HTTP_CODE}"
```

Expected result:

```text
ARGOCD_TOKEN_HTTP_CODE=200
```

If the result is not `200`, inspect only the error response:

```bash
python3 -m json.tool /tmp/argocd-token-check.json | sed -n '1,120p'
```

Cleanup:

```bash
rm -f /tmp/argocd-token-check.json
```

Common failure observed during Phase 9.1:

```text
HTTP 401
invalid session: token has invalid claims: token is expired
```

In that case rotate `ARGOCD_AUTH_TOKEN`.

### GitLab token health check

This check validates `GITLAB_TOKEN` against GitLab project `1` without printing it.

```bash
GITLAB_HOST="$(oc get route gitlab-lab -n devops-gitlab -o jsonpath='{.spec.host}')"

GITLAB_TOKEN="$(
  oc get secret devops-control-plane-secrets \
    -n devops-control-plane \
    -o jsonpath='{.data.GITLAB_TOKEN}' \
  | base64 -d
)"

HTTP_CODE="$(
  curl -k -sS \
    -o /tmp/gitlab-token-check.json \
    -w '%{http_code}' \
    -H "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
    "https://${GITLAB_HOST}/api/v4/projects/1"
)"

unset GITLAB_TOKEN

echo "GITLAB_TOKEN_HTTP_CODE=${HTTP_CODE}"
```

Expected result:

```text
GITLAB_TOKEN_HTTP_CODE=200
```

Optional non-sensitive metadata inspection:

```bash
python3 -m json.tool /tmp/gitlab-token-check.json | sed -n '1,80p'
```

Cleanup:

```bash
rm -f /tmp/gitlab-token-check.json
```

### Kubernetes token health check

This check validates `KUBERNETES_TOKEN` without printing it.

```bash
KUBE_API="$(oc get configmap devops-control-plane-config \
  -n devops-control-plane \
  -o jsonpath='{.data.KUBERNETES_API_URL}')"

KUBE_TOKEN="$(
  oc get secret devops-control-plane-secrets \
    -n devops-control-plane \
    -o jsonpath='{.data.KUBERNETES_TOKEN}' \
  | base64 -d
)"

HTTP_CODE="$(
  curl -k -sS \
    -o /tmp/kubernetes-token-check.json \
    -w '%{http_code}' \
    -H "Authorization: Bearer ${KUBE_TOKEN}" \
    "${KUBE_API}/api"
)"

unset KUBE_TOKEN

echo "KUBERNETES_TOKEN_HTTP_CODE=${HTTP_CODE}"
```

Expected result:

```text
KUBERNETES_TOKEN_HTTP_CODE=200
```

Cleanup:

```bash
rm -f /tmp/kubernetes-token-check.json
```

Validated Phase 9.2 token health baseline:

```text
ARGOCD_TOKEN_HTTP_CODE=200
GITLAB_TOKEN_HTTP_CODE=200
KUBERNETES_TOKEN_HTTP_CODE=200
```

## Rotation runbook: Argo CD token

### When to rotate

Rotate `ARGOCD_AUTH_TOKEN` when:

- Argo CD API returns `401`;
- the token is expired;
- the token owner/account is changed;
- a suspected leak occurs;
- as part of periodic security maintenance.

### Important note about the admin account

In the lab, the Argo CD `admin` account can log in to the UI but may not generate API tokens from the UI if it lacks the `apiKey` capability.

Observed error:

```text
Unable to generate new token: failed to update account with new token: account 'admin' does not have apiKey capability
```

Do not change Argo CD/OpenShift GitOps configuration during this phase. Use a session token through `/api/v1/session` instead.

### Generate a session token without printing it

```bash
ARGOCD_HOST="$(oc get route openshift-gitops-server -n openshift-gitops -o jsonpath='{.spec.host}')"

ARGOCD_ADMIN_PASSWORD="$(
  oc get secret openshift-gitops-cluster \
    -n openshift-gitops \
    -o jsonpath='{.data.admin\.password}' \
  | base64 -d
)"

HTTP_CODE="$(
  ARGOCD_ADMIN_PASSWORD="$ARGOCD_ADMIN_PASSWORD" \
  python3 -c 'import json, os; print(json.dumps({"username":"admin","password":os.environ["ARGOCD_ADMIN_PASSWORD"]}))' \
  | curl -k -sS \
      -o /tmp/argocd-session.json \
      -w '%{http_code}' \
      -H 'Content-Type: application/json' \
      -d @- \
      "https://${ARGOCD_HOST}/api/v1/session"
)"

unset ARGOCD_ADMIN_PASSWORD

echo "ARGOCD_SESSION_HTTP_CODE=${HTTP_CODE}"
```

Expected result:

```text
ARGOCD_SESSION_HTTP_CODE=200
```

Save the token to a temporary file without printing it:

```bash
python3 -c '
import json
from pathlib import Path

data = json.loads(Path("/tmp/argocd-session.json").read_text())
token = data.get("token", "")
if not token:
    raise SystemExit("missing token in Argo CD session response")

Path("/tmp/argocd-token.txt").write_text(token)
print("token_saved")
'

chmod 600 /tmp/argocd-token.txt
rm -f /tmp/argocd-session.json
```

Validate the token:

```bash
HTTP_CODE="$(
  curl -k -sS \
    -o /tmp/argocd-app-check.json \
    -w '%{http_code}' \
    -H "Authorization: Bearer $(cat /tmp/argocd-token.txt)" \
    "https://${ARGOCD_HOST}/api/v1/applications/demo-go-color-app"
)"

echo "ARGOCD_APP_HTTP_CODE=${HTTP_CODE}"
```

Expected result:

```text
ARGOCD_APP_HTTP_CODE=200
```

Patch the application Secret:

```bash
ARGOCD_TOKEN_B64="$(base64 -w0 /tmp/argocd-token.txt)"

oc patch secret devops-control-plane-secrets \
  -n devops-control-plane \
  --type merge \
  -p "{\"data\":{\"ARGOCD_AUTH_TOKEN\":\"${ARGOCD_TOKEN_B64}\"}}"

unset ARGOCD_TOKEN_B64
rm -f /tmp/argocd-token.txt
rm -f /tmp/argocd-app-check.json
```

Restart the application:

```bash
oc rollout restart deploy/devops-control-plane -n devops-control-plane

oc rollout status deploy/devops-control-plane \
  -n devops-control-plane \
  --timeout=300s
```

Post-rotation test:

```bash
APP_ROUTE="$(oc get route devops-control-plane -n devops-control-plane -o jsonpath='{.spec.host}')"

curl -k -sS "https://${APP_ROUTE}/api/v1/applications/demo-go-color-app" \
  | python3 -m json.tool \
  | sed -n '1,120p'
```

Expected functional result:

```text
syncStatus: Synced
healthStatus: Healthy
```

## Rotation runbook: GitLab token

### When to rotate

Rotate `GITLAB_TOKEN` when:

- GitLab API returns `401` or `403`;
- the token is expired;
- the token owner changes;
- scopes are changed;
- a suspected leak occurs.

### Rotation procedure

1. Generate a new GitLab token with the minimum required scopes for project API operations.
2. Store it in a temporary file without printing it:

```bash
read -s -p "New GITLAB_TOKEN: " NEW_GITLAB_TOKEN
echo
printf '%s' "$NEW_GITLAB_TOKEN" > /tmp/gitlab-token.txt
chmod 600 /tmp/gitlab-token.txt
unset NEW_GITLAB_TOKEN
```

3. Validate the new token:

```bash
GITLAB_HOST="$(oc get route gitlab-lab -n devops-gitlab -o jsonpath='{.spec.host}')"

HTTP_CODE="$(
  curl -k -sS \
    -o /tmp/gitlab-token-check.json \
    -w '%{http_code}' \
    -H "PRIVATE-TOKEN: $(cat /tmp/gitlab-token.txt)" \
    "https://${GITLAB_HOST}/api/v4/projects/1"
)"

echo "NEW_GITLAB_TOKEN_HTTP_CODE=${HTTP_CODE}"
```

Expected result:

```text
NEW_GITLAB_TOKEN_HTTP_CODE=200
```

4. Patch the application Secret:

```bash
GITLAB_TOKEN_B64="$(base64 -w0 /tmp/gitlab-token.txt)"

oc patch secret devops-control-plane-secrets \
  -n devops-control-plane \
  --type merge \
  -p "{\"data\":{\"GITLAB_TOKEN\":\"${GITLAB_TOKEN_B64}\"}}"

unset GITLAB_TOKEN_B64
rm -f /tmp/gitlab-token.txt
rm -f /tmp/gitlab-token-check.json
```

5. Restart and retest relevant GitLab workflows.

```bash
oc rollout restart deploy/devops-control-plane -n devops-control-plane

oc rollout status deploy/devops-control-plane \
  -n devops-control-plane \
  --timeout=300s
```

## Rotation runbook: Kubernetes token

### Current state

The current MVP uses an explicit `KUBERNETES_TOKEN` stored in the application Secret.

This is acceptable for the lab but should be replaced in a later phase by an in-cluster ServiceAccount token strategy.

Related roadmap item:

```text
9.6 — Rimozione token Kubernetes statico
```

### When to rotate

Rotate `KUBERNETES_TOKEN` when:

- Kubernetes API returns `401`;
- the token duration expires;
- ServiceAccount permissions change;
- a suspected leak occurs.

### Validate ServiceAccount permissions

```bash
oc auth can-i list pods -n devops-ci-demo \
  --as=system:serviceaccount:devops-control-plane:devops-control-plane

oc auth can-i get deployments -n devops-ci-demo \
  --as=system:serviceaccount:devops-control-plane:devops-control-plane

oc auth can-i create pipelineruns.tekton.dev -n devops-ci-demo \
  --as=system:serviceaccount:devops-control-plane:devops-control-plane
```

Expected result:

```text
yes
yes
yes
```

### Generate a new token without printing it

```bash
oc create token devops-control-plane \
  -n devops-control-plane \
  --duration=24h > /tmp/dcp-kubernetes-token.txt

chmod 600 /tmp/dcp-kubernetes-token.txt
wc -c /tmp/dcp-kubernetes-token.txt
```

### Validate the new token

```bash
KUBE_API="$(oc get configmap devops-control-plane-config \
  -n devops-control-plane \
  -o jsonpath='{.data.KUBERNETES_API_URL}')"

HTTP_CODE="$(
  curl -k -sS \
    -o /tmp/kubernetes-token-check.json \
    -w '%{http_code}' \
    -H "Authorization: Bearer $(cat /tmp/dcp-kubernetes-token.txt)" \
    "${KUBE_API}/api"
)"

echo "NEW_KUBERNETES_TOKEN_HTTP_CODE=${HTTP_CODE}"
```

Expected result:

```text
NEW_KUBERNETES_TOKEN_HTTP_CODE=200
```

### Patch Secret

```bash
KUBE_TOKEN_B64="$(base64 -w0 /tmp/dcp-kubernetes-token.txt)"

oc patch secret devops-control-plane-secrets \
  -n devops-control-plane \
  --type merge \
  -p "{\"data\":{\"KUBERNETES_TOKEN\":\"${KUBE_TOKEN_B64}\"}}"

unset KUBE_TOKEN_B64
rm -f /tmp/dcp-kubernetes-token.txt
rm -f /tmp/kubernetes-token-check.json
```

Restart:

```bash
oc rollout restart deploy/devops-control-plane -n devops-control-plane

oc rollout status deploy/devops-control-plane \
  -n devops-control-plane \
  --timeout=300s
```

## DATABASE_URL handling

`DATABASE_URL` contains PostgreSQL connection information. Treat it as secret because it typically contains username and password.

Do not print it.

Allowed check:

```bash
oc get secret devops-control-plane-secrets \
  -n devops-control-plane \
  -o jsonpath='{.data.DATABASE_URL}' \
  | wc -c
```

Rotation of `DATABASE_URL` requires coordination with PostgreSQL credentials and application restart.

## Post-rotation universal checklist

After any secret rotation:

```bash
oc rollout restart deploy/devops-control-plane -n devops-control-plane

oc rollout status deploy/devops-control-plane \
  -n devops-control-plane \
  --timeout=300s
```

Check readiness:

```bash
APP_ROUTE="$(oc get route devops-control-plane -n devops-control-plane -o jsonpath='{.spec.host}')"

curl -k -sS "https://${APP_ROUTE}/readyz" \
  | python3 -m json.tool \
  | sed -n '1,120p'
```

Check Argo CD integration:

```bash
curl -k -sS "https://${APP_ROUTE}/api/v1/applications/demo-go-color-app" \
  | python3 -m json.tool \
  | sed -n '1,160p'
```

Expected application status:

```text
syncStatus: Synced
healthStatus: Healthy
```

## Troubleshooting

### HTTP 401

Meaning:

```text
Token missing, expired, revoked or invalid for the target API.
```

Action:

```text
Rotate the affected token and restart the deployment.
```

### HTTP 403

Meaning:

```text
Token is valid but does not have sufficient permissions.
```

Action:

```text
Review scopes, RBAC or ServiceAccount permissions.
```

### TLS certificate errors

Meaning:

```text
CA bundle is missing, incorrect or not mounted.
```

Action:

```bash
POD="$(oc get pod -n devops-control-plane -l app=devops-control-plane -o jsonpath='{.items[0].metadata.name}')"

oc exec -n devops-control-plane "$POD" -- sh -c '
echo "ARGOCD_CA_FILE=${ARGOCD_CA_FILE}"
echo "ARGOCD_INSECURE_TLS=${ARGOCD_INSECURE_TLS}"
wc -c /etc/dcp-trust/ca-bundle.crt
'
```

## Current Phase 9.2 result

The Phase 9.2 baseline validated:

```text
ARGOCD_AUTH_TOKEN: HTTP 200
GITLAB_TOKEN: HTTP 200
KUBERNETES_TOKEN: HTTP 200
```

No secret values were printed during the checks.
