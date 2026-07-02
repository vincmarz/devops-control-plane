# DevOps Control Plane - OpenShift deployment MVP

This directory contains the first OpenShift deployment skeleton for the DevOps Control Plane.

## Files

- `namespace.yaml`: creates the `devops-control-plane` namespace.
- `serviceaccount.yaml`: service account used by the Control Plane pod.
- `configmap.yaml`: non-sensitive runtime configuration.
- `secret-template.yaml`: placeholder application secret template. Replace placeholders locally before applying. Do not commit real values.
- `postgresql-secret-template.yaml`: placeholder PostgreSQL secret template. Replace placeholders locally before applying. Do not commit real values.
- `postgresql-pvc.yaml`: persistent storage for PostgreSQL.
- `postgresql-deployment.yaml`: PostgreSQL database used by the Control Plane MVP deployment.
- `postgresql-service.yaml`: in-cluster PostgreSQL service named `postgresql`.
- `role.yaml`: permissions in `devops-ci-demo` for Tekton PipelineRuns and runtime evidence collection.
- `rolebinding.yaml`: binds the `devops-control-plane` service account to the role in `devops-ci-demo`.
- `deployment.yaml`: app deployment and probes.
- `service.yaml`: ClusterIP service on port 8080.
- `route.yaml`: OpenShift edge route.
- `kustomization.yaml`: apply all manifests with `oc apply -k manifests`.

## Sensitive values

The following values must be provided through `devops-control-plane-secrets` and must never be committed with real values:

- `DATABASE_URL`
- `GITLAB_TOKEN`
- `ARGOCD_AUTH_TOKEN`
- `KUBERNETES_TOKEN`
- `POSTGRESQL_PASSWORD` in `postgresql-secret`

## Build image example

From the repository root:

```bash
podman build -t devops-control-plane:latest -f Containerfile .
```

For the OpenShift internal registry, tag and push to:

```bash
image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane:latest
```

Alternatively, adapt the deployment image to the registry used by the lab.

## Apply manifests

Create the PostgreSQL Secret from `postgresql-secret-template.yaml` and the application Secret from `secret-template.yaml` without committing real values, then apply the non-sensitive manifests:

```bash
oc apply -k manifests
```

The files `secret-template.yaml` and `postgresql-secret-template.yaml` are intentionally not included in `kustomization.yaml`.

## Validate

```bash
oc get pods,svc,route -n devops-control-plane
oc rollout status deploy/devops-control-plane -n devops-control-plane
oc get route devops-control-plane -n devops-control-plane
```

Expected probes:

- Liveness: `/healthz`
- Readiness: `/readyz`

## Lab-only notes

- `GITLAB_INSECURE_TLS=true`, `ARGOCD_INSECURE_TLS=true` and `KUBERNETES_INSECURE_TLS=true` are acceptable for the current lab, but must be revisited during security hardening.
- AuthN/AuthZ is intentionally deferred to Phase 9.3.
- The UI environment selector is static until multi-environment support is implemented.

## App-only trust bundle mount

The DevOps Control Plane deployment can mount the namespace-local `config-trusted-cabundle` ConfigMap as a read-only trust bundle at:

```text
/etc/dcp-trust/ca-bundle.crt
```

This is an application-only setting. It does not modify OpenShift cluster-wide objects such as `proxy/cluster`, ingress/router certificates, or managed cluster CA ConfigMaps.

The manifest sets:

```text
ARGOCD_CA_FILE=/etc/dcp-trust/ca-bundle.crt
GITLAB_CA_FILE=/etc/dcp-trust/ca-bundle.crt
KUBERNETES_CA_FILE=
```

The current lab behavior remains unchanged while `ARGOCD_INSECURE_TLS=true`, `GITLAB_INSECURE_TLS=true`, and `KUBERNETES_INSECURE_TLS=true` are enabled. The `*_CA_FILE` values prepare the runtime for a future controlled test with insecure TLS disabled for selected integrations.

## App-dedicated trust bundle for Argo CD TLS verification

The deployment expects an application-owned ConfigMap named `dcp-app-trust-bundle` in the `devops-control-plane` namespace. The ConfigMap must contain the key `ca-bundle.crt` and is mounted read-only at:

```text
/etc/dcp-trust/ca-bundle.crt
```

The application ConfigMap uses:

```text
ARGOCD_INSECURE_TLS=false
ARGOCD_CA_FILE=/etc/dcp-trust/ca-bundle.crt
```

GitLab and Kubernetes integrations intentionally remain in the current lab-safe mode for now:

```text
GITLAB_INSECURE_TLS=true
KUBERNETES_INSECURE_TLS=true
```

### Creating the app-owned trust bundle ConfigMap

Extract the CA certificate from the Argo CD Route chain in read-only mode and create the namespace-local ConfigMap. Do not modify OpenShift cluster-wide objects.

```bash
mkdir -p /tmp/dcp-ingress-ca

ARGOCD_HOST="$(oc get route openshift-gitops-server -n openshift-gitops -o jsonpath='{.spec.host}')"

echo | openssl s_client \
  -connect "${ARGOCD_HOST}:443" \
  -servername "${ARGOCD_HOST}" \
  -showcerts \
  2>/dev/null > /tmp/dcp-ingress-ca/argocd-route-chain.pem
```

The CA selected during the lab validation was the self-signed ingress CA with:

```text
subject=CN = ingress-operator@1772025286
issuer=CN = ingress-operator@1772025286
SHA256 Fingerprint=4D:1B:4E:01:CF:F5:CA:3E:24:E6:11:9A:DA:5F:81:2B:1D:9B:47:5A:D3:03:18:07:4A:91:57:E5:FA:4B:E0:43
```

After writing the selected PEM certificate to `/tmp/dcp-ingress-ca/ingress-ca.pem`, create or update the ConfigMap:

```bash
oc create configmap dcp-app-trust-bundle \
  -n devops-control-plane \
  --from-file=ca-bundle.crt=/tmp/dcp-ingress-ca/ingress-ca.pem \
  --dry-run=client \
  -o yaml > /tmp/dcp-ingress-ca/dcp-app-trust-bundle.yaml

oc apply -f /tmp/dcp-ingress-ca/dcp-app-trust-bundle.yaml
```

Validate the mounted file after rollout:

```bash
POD="$(oc get pod -n devops-control-plane -l app=devops-control-plane -o jsonpath='{.items[0].metadata.name}')"

oc exec -n devops-control-plane "$POD" -- sh -c '
echo "CA bundle path: /etc/dcp-trust/ca-bundle.crt"
wc -c /etc/dcp-trust/ca-bundle.crt
'
```

### Important boundary

This is an application-only trust strategy. It must not modify:

```text
proxy/cluster
ingress/router certificates
cluster-wide CA configuration
OpenShift-managed ConfigMaps
```

### Kubernetes/OpenShift authentication after phase 9.6

`KUBERNETES_TOKEN` is no longer required as a static application Secret value.
The preferred runtime model is to use the ServiceAccount token automatically mounted in the pod:

```text
/var/run/secrets/kubernetes.io/serviceaccount/token
```

The application configuration now follows this order:

1. If `KUBERNETES_TOKEN` is set, it is used as an explicit legacy override.
2. If `KUBERNETES_TOKEN` is empty or missing, the application reads the ServiceAccount token file.
3. If `KUBERNETES_API_URL` is empty, the application builds it from `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT`.
4. If `KUBERNETES_CA_FILE` is empty, the ServiceAccount CA file is used when present.

For production-like deployments, do not create or rotate a static Kubernetes token unless a specific break-glass use case requires it.

