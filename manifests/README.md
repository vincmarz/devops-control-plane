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
