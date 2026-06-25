# DevOps Control Plane

DevOps Control Plane is an API-first Go backend intended to orchestrate GitOps change workflows across GitLab, Tekton/OpenShift Pipelines, Argo CD/OpenShift GitOps and OpenShift/Kubernetes runtime evidence.

## MVP principles

- Git is the source of truth.
- Argo CD is the GitOps reconciliation engine.
- Tekton is the validation engine.
- PostgreSQL stores ChangeRequest history and sanitized evidence.
- API/workflow stability comes before a complete Web UI.
- Secrets must never be stored in Git, logs, API responses or evidence.

## Initial skeleton

```text
cmd/devops-control-plane/main.go
internal/config
internal/logging
internal/api
internal/domain
internal/app
internal/workflow
internal/adapters
internal/database
migrations
manifests
pipelines
docs/adr
```

## Run locally

```bash
cp .env.example .env
# edit .env with local values if needed
make run
```

Health endpoints:

```bash
curl -s http://localhost:8080/healthz | jq .
curl -s http://localhost:8080/readyz | jq .
```

## Build

```bash
make build
```

## Test

```bash
make test
```

## Security note

Do not commit real `.env`, tokens, kubeconfig files, private keys or Kubernetes Secret manifests containing real values.
