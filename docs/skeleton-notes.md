# Go Skeleton Notes

This skeleton is intentionally API-first and adapter-oriented.

## What works now

- Go HTTP server with standard library router.
- Structured JSON logging with `log/slog`.
- `/healthz` endpoint.
- `/readyz` endpoint with placeholder DB readiness.
- Placeholder Application API.
- In-memory ChangeRequest service.
- Placeholder workflow action endpoints.
- Placeholder Evidence API.
- Initial migrations and OpenShift manifests.

## What is intentionally placeholder

- Real PostgreSQL connection and repositories.
- GitLab API adapter.
- Argo CD API adapter.
- Tekton/Kubernetes API adapter.
- Real workflow orchestration.
- Real evidence persistence.

## Next recommended implementation steps

1. Add PostgreSQL driver and real DB connection.
2. Implement migrations execution strategy.
3. Replace in-memory ChangeService with PostgreSQL repository.
4. Implement Argo CD read-only application discovery.
5. Implement GitLab GetFile/ListCommits/CreateBranch/CommitFiles.
6. Implement Tekton PipelineRun creation and polling.
7. Implement evidence persistence.
