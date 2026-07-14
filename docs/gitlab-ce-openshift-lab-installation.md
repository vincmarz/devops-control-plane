# GitLab CE Installation on OpenShift

## 1. Purpose

This document describes the verified GitLab Community Edition installation used by the DevOps Control Plane development and validation environment.

GitLab CE is operational, persistent, exposed through an OpenShift Route, and integrated with the DevOps Control Plane through the GitLab REST API. The current deployment provides a validated development baseline for SCM workflow testing. It is not a highly available or production-grade GitLab architecture.

## 2. Deployment classification

The GitLab integration is implemented and validated. The current deployment architecture is classified as a development and validation baseline because it uses one replica, a `Recreate` strategy, `ReadWriteOnce` persistent volumes, and no demonstrated high-availability or disaster-recovery configuration.

- OpenShift cluster: `ocp-dev`
- Namespace: `devops-gitlab`
- Deployment: `gitlab-lab`
- Image: `docker.io/gitlab/gitlab-ce:18.11.6-ce.0`
- Replicas: `1`
- Strategy: `Recreate`
- Revision history limit: `10`

## 3. Verified runtime state

At verification time:

- Pod: `gitlab-lab-746dbdd74f-zxx4z`
- Phase: `Running`
- Restarts: `0`
- Node: `worker-01.ocp4.mim.lan`
- Observed start time: `2026-06-26T13:50:49Z`

The pod name is runtime-specific and must not be hard-coded in operational procedures.

## 4. Network exposure

### 4.1 Service

- Name: `gitlab-lab`
- Type: `ClusterIP`
- Observed Cluster IP: `172.30.243.125`
- Port: `80/TCP`
- Target port: `http`

### 4.2 Route

- Name: `gitlab-lab`
- Host: `gitlab-devops-gitlab.apps.ocp4.mim.lan`
- URL: `https://gitlab-devops-gitlab.apps.ocp4.mim.lan`
- Target service: `gitlab-lab`
- TLS termination: `edge`
- Insecure traffic policy: `Redirect`

```text
HTTPS client
-> OpenShift Router
-> edge TLS termination
-> Service gitlab-lab:80
-> GitLab pod:80
```

## 5. TLS and trust

The Route uses the wildcard certificate for `*.apps.ocp4.mim.lan`, issued by the internal Ingress Operator CA. TLS 1.3 was observed.

The hostname is covered by the certificate, but the internal CA was not trusted by the observed bastion and browser clients. `curl -k` is acceptable only for lab diagnostics. The proper remediation is to distribute the internal CA or use a certificate signed by a trusted CA.

## 6. Initial administrator credential

The initial administrator password is injected through:

- Environment variable: `GITLAB_ROOT_PASSWORD`
- Secret: `gitlab-lab-root-password`
- Key: `GITLAB_ROOT_PASSWORD`

The Secret value must never be included in documentation, logs, or evidence.

## 7. Persistent storage

All PVCs are `Bound`, use access mode `RWO`, and use storage class `ocp-rbd-rook`.

### 7.1 Configuration

- PVC: `gitlab-config`
- Capacity: `1Gi`
- Mount: `/etc/gitlab`
- Observed use: approximately `1%`

### 7.2 Logs

- PVC: `gitlab-logs`
- Capacity: `5Gi`
- Mount: `/var/log/gitlab`
- Observed use: approximately `8%`

### 7.3 Application data

- PVC: `gitlab-data`
- Capacity: `30Gi`
- Mount: `/var/opt/gitlab`
- Observed use: approximately `2%`

PVC persistence protects data across pod recreation but does not replace backup and restore procedures.

## 8. Container resources

- CPU request: `1`
- CPU limit: `3`
- Memory request: `4Gi`
- Memory limit: `8Gi`

Memory use, `OOMKilled` events, CPU throttling, log growth, and PVC capacity should be monitored.

## 9. Health probes

### 9.1 Startup probe

- Endpoint: `http://127.0.0.1:80/-/health`
- Initial delay: `60s`
- Period: `15s`
- Timeout: `10s`
- Failure threshold: `80`

### 9.2 Readiness probe

- Endpoint: `http://127.0.0.1:80/-/readiness`
- Initial delay: `30s`
- Period: `30s`
- Timeout: `10s`
- Failure threshold: `20`

### 9.3 Liveness probe

- Endpoint: `http://127.0.0.1:80/-/liveness`
- Initial delay: `300s`
- Period: `60s`
- Timeout: `10s`
- Failure threshold: `10`

Manual checks returned:

```text
/-/health     -> GitLab OK
/-/readiness  -> status ok, master check ok
/-/liveness   -> status ok
```

## 10. Observed component versions

- GitLab: `18.11.6`
- GitLab Shell: `14.50.0`
- GitLab Workhorse: `18.11.6`
- GitLab API: `v4`
- GitLab KAS: `18.11.6`
- Ruby: `3.3.10p183`
- Rails: `7.2.3.1`
- PostgreSQL: `17.8`
- Redis: `7.2.11`

These values must be rechecked after upgrades or reinstallation.

## 11. Observed inventory

- Projects: `1`
- Users: `2`
- Groups: `1`
- Merge Requests: `4`
- Group: `devops-lab`
- Project: `devops-lab/demo-go-color-app-gitops`

## 12. Role in the DevOps Control Plane

GitLab CE is used as an operational SCM provider through REST APIs. The tested baseline covers:

- branch creation;
- file creation or update;
- Merge Request creation;
- open Merge Request lookup;
- Merge Request merge.

GitHub Actions, not GitLab CI, provides the quality gate for the `devops-control-plane` repository. The absence of `.gitlab-ci.yml` does not indicate an absent GitLab SCM integration.

## 13. Known limitations

- Development and validation deployment, not production-grade
- Single replica and no application high availability
- `Recreate` rollout strategy
- `RWO` persistent volumes
- Internal Ingress CA not trusted by the observed clients
- Sign-up enabled
- Backup and restore not yet formalized
- PVC monitoring and alerts not yet formalized
- Effective Instance Runner state not yet verified
- No demonstrated automatic synchronization with the GitHub repository consumed by Argo CD

## 14. Essential checks

```bash
oc -n devops-gitlab get deployment gitlab-lab
oc -n devops-gitlab get pod
oc -n devops-gitlab get service gitlab-lab
oc -n devops-gitlab get route gitlab-lab
oc -n devops-gitlab get pvc
```

## 15. Remaining work

- Define and test GitLab backup and restore
- Formalize maintenance and upgrade procedures
- Verify registered runners and health
- Reassess whether sign-up should remain enabled
- Distribute the internal CA or use a trusted certificate
- Define alerts for PVC capacity, memory, restarts, and probes
- Decide the target relationship between the GitLab SCM repository and the GitOps runtime repository
