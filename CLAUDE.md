# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

An API-first Go backend that orchestrates GitOps change workflows across Git providers (GitLab/GitHub), Tekton/OpenShift Pipelines, Argo CD/OpenShift GitOps, and OpenShift/Kubernetes runtime evidence. Standard library only for HTTP; the sole runtime dependencies are `pgx/v5` and `yaml.v3`.

## Commands

```bash
make run          # go run ./cmd/devops-control-plane (needs DATABASE_URL; copy .env.example → .env)
make build        # binary into bin/
make test         # go test ./...
make fmt          # go fmt ./...  — CI fails if `gofmt -l .` is non-empty
make vet          # go vet ./...
```

Single test / package:

```bash
go test ./internal/app -run TestChangeServiceValidate -v
go test ./... -race                      # CI runs with -race -covermode=atomic
```

Reproducing the CI `test` job locally (`.github/workflows/ci.yml`, Go 1.22) — these four steps in order are the whole gate:

```bash
test -z "$(gofmt -l .)" && go vet ./... && \
  go test ./... -race -covermode=atomic -coverprofile=coverage.out && \
  go test -tags=integration ./internal/database/... -run Integration -v
```

Integration tests are behind the `integration` build tag and require a real PostgreSQL:

```bash
TEST_DATABASE_URL='postgres://dcp:dcp@localhost:5432/dcp_test?sslmode=disable' \
  go test -tags=integration ./internal/database/... -run Integration -v
```

Without the tag, DB-backed tests are excluded entirely; without `TEST_DATABASE_URL` they skip. These tests reset schema, so only point them at a disposable database.

Migrations (each numbered migration has its own Make target — `make migrate-up` applies **only** `000001_init`):

```bash
make migrate-up                    # 000001_init
make migrate-runtime-state-up      # 000002_change_runtime_states
make migrate-artifact-state-up     # 000003_artifact_runtime_state
# equivalently: go run ./cmd/devops-control-plane-migrate -direction up -up migrations/<file>.sql
```

Container build & deploy (bastion / OpenShift dev):

- **Build in a temporary directory**, never in the repo root (disk-space constraint). Clone the target commit into `$(mktemp -d /tmp/devops-cp-<sha>.XXXXXX)` and build there.
- Build and push an **immutable digest**, do not deploy by tag:
  ```bash
  podman build -t <registry>/devops-control-plane/devops-control-plane:<sha> -f Containerfile .
  podman push ... --digestfile /tmp/devops-control-plane-<sha>-push-digest.txt
  ```
- Deploy by **digest**, on the application container addressed **by name** — `containers[0]` is the oauth-proxy, not the app:
  ```bash
  oc set image deploy/devops-control-plane \
    devops-control-plane=image-registry.openshift-image-registry.svc:5000/devops-control-plane/devops-control-plane@sha256:<digest> \
    -n devops-control-plane
  oc rollout status deploy/devops-control-plane -n devops-control-plane
  ```
- **Do NOT run `oc apply -k manifests`**: `deployment.yaml`/`service.yaml`/`route.yaml` are not aligned with the evolved live runtime configuration and applying them regresses the deployment. Secret templates, `configmap-clusters.yaml`, the oauth-proxy overlays and the NetworkPolicies are all deliberately outside `kustomization.yaml` too — the live oauth-proxy configuration is managed by the staged scripts in `scripts/oauth-proxy-live-patch/` (`00-precheck` → `01-prepare-prereqs` → `02-apply-live-patches` → `03-verify-rollout`, with `99-rollback`), documented in `docs/runbooks/oauth-proxy-live-patch-rollout.md`.
- The route is `reencrypt` with a self-signed cert: external readiness checks need `curl -k` (a transient `504` right after rollout is normal until the pod/oauth-proxy are ready). `scripts/operability/health-check.sh` is the scripted probe.

## Architecture

Layering is `api → app → (domain, database, adapters)`. **The `app` package never imports a concrete adapter.** Anything crossing that line is injected from the composition root.

### Composition root: `cmd/devops-control-plane/main.go`

`main.go` builds the entire object graph and is the only place adapters are constructed. It assembles a `[]app.ChangeServiceOption` list, and each integration is enabled only if its config is present:

- GitLab provider if `GITLAB_BASE_URL`/`GITLAB_TOKEN` set; GitHub provider if `GITHUB_TOKEN` set; both registered in an `app.GitProviderRegistry` keyed by `providerRef` (`"gitlab-lab"`, `"github-public"`).
- Kubernetes + Tekton clients if `KUBERNETES_API_URL`/`KUBERNETES_TOKEN` set (token falls back to the pod ServiceAccount token file).
- Argo CD client if `ARGOCD_BASE_URL`/`ARGOCD_AUTH_TOKEN` set.

When an integration is absent the server still starts and the corresponding endpoints fail with a clear error. Adding a new capability means: define a `...Func` port type + `With...` option in `internal/app`, implement the adapter in `internal/adapters/<system>`, and wire the closure in `main.go`.

`package main` carries real adapter code, not just `main.go`: the `*_runtime_client_adapter.go` / `*_runtime_client_factory_adapter.go` / `runtime_client_factory_builders.go` files bridge concrete adapters onto the `app` port types, which is what keeps `internal/app` adapter-free. The wiring is itself covered by ~15 `main_*_wiring_test.go` files, and a new integration is expected to land with one — but read the next section before touching them.

### Source-text assertion tests — read before refactoring

About 20 test files across `cmd/devops-control-plane`, `internal/app` and `internal/api` do **not** exercise behavior: they `os.ReadFile` a sibling source file and assert that specific literal substrings are present (or absent, or appear exactly N times). All the `main_*_wiring_test.go` files work this way, as do `git_error_contract_test.go`, `ui_provider_neutral_test.go`, `build_pipeline_contract_test.go` and the `change_service_*_wiring_test.go` set. Find them with:

```bash
grep -rln 'os.ReadFile("' --include='*_test.go' internal cmd
```

Consequences to keep in mind:

- A pure refactor — renaming a local variable, reordering arguments, extracting a helper, even changing whitespace inside a matched expression — **fails these tests without any behavior change**. The failure message names the missing string, so treat it as "the pinned wiring changed", then decide whether to restore the shape or update the expected literal deliberately.
- They pin intent that is otherwise invisible: provider-neutral error codes (`GIT_*`, never `GITLAB_*`), "exactly one Git provider registry", flags-off returning `nil`. When updating one, preserve the *intent* rather than pasting in whatever the new source says.
- New behavior deserves a real behavioral test with fake closures (see below); reach for a source-text test only to pin a cross-layer invariant that has no runtime observable.

### Ports as function types

`internal/app/change_service.go` (~2k lines, the core of the system) declares its dependencies as named func types — `TektonRunPipelineFunc`, `ArgoCDCheckDeploymentFunc`, `DeploymentEvidenceCollectorFunc`, `KubernetesRuntimeEvidenceCollectorFunc`, etc. — plus small interfaces (`ChangeStore`, `EvidenceStore`). Tests construct `ChangeService` with fake closures rather than fake servers; follow that pattern.

`GitProvider` (`internal/app/git_provider.go`) is the exception: a real interface, provider-neutral by design. GitLab merge requests and GitHub pull requests both map onto `OpenMergeRequest`/`MergeRequest`, and API responses must stay provider-neutral (see `internal/api/ui_provider_neutral_test.go`, `git_error_contract_test.go`).

### Configuration model: YAML catalogs with in-code fallbacks

Four file-backed registries, each read from an env-var path at startup and each falling back to a hardcoded in-code baseline if the file is missing or invalid:

| Registry | Env var | Purpose |
|---|---|---|
| `ApplicationCatalog` | `APPLICATION_CATALOG_FILE` | logical app → repository bindings (`role: source` / `gitops`, `consumedBy: [argocd, tekton]`, plus per-app `buildImage` / `kustomizationPath`) |
| `EnvironmentCatalog` | `ENVIRONMENT_CATALOG_FILE` | environments, namespaces, Argo CD app name, enablement flags |
| `ClusterRegistry` | `CLUSTER_REGISTRY_FILE` | clusters, API URLs, CA/token Secret *references* |
| `RuntimeClientSecretRefs` | `DCP_RUNTIME_CLIENT_SECRET_REFS_FILE` | per-cluster Secret references (references only, never values) |

`config.Load()` in `internal/config` only handles scalar env vars (see `.env.example` for the full scalar surface); the catalogs load themselves via `DefaultXxx()` and fall back to `DefaultApplicationCatalogFallback()` / `DefaultClusterRegistryFallback()` / `DefaultEnvironmentCatalogFallback()` (the secret-refs registry's baseline is inline in `DefaultRuntimeClientSecretRefsRegistry()`). In tests prefer the explicit constructors (`NewEnvironmentCatalog`, `NewTechnicalRuntimeTargetResolver`, …) so behavior does not depend on ambient env vars or files.

Per-application repository bindings **override** the global env defaults, resolved per change via `resolveBuildImage` / `resolveKustomizationPath` in `change_service.go`: the `source` binding's `buildImage` overrides `TEKTON_BUILD_IMAGE` in `StartBuild`, the `gitops` binding's `kustomizationPath` overrides `GITOPS_IMAGE_KUSTOMIZATION_PATH` in `UpdateGitOps`. Both fall back to the global value when the binding leaves the field empty, so prefer extending the catalog over adding another global env var when something becomes per-application.

### Environment → cluster → runtime client resolution

Every technical action resolves through this chain, and each step can reject:

```
targetEnvironment
  → EnvironmentClusterResolver   (EnvironmentCatalog + ClusterRegistry)
  → TechnicalRuntimeTarget       (namespaces, Argo CD app, git branch, validation path; .Validate())
  → RuntimeClientProviderRegistry.Select  (metadata-only provider entry per cluster)
  → factory-aware provider registry       (resolves a concrete client for the cluster)
```

The system is **fail-closed by design**: only `dev`/`ocp-dev` is enabled in the baseline; `staging` and `production` are configured but disabled, and disabled environments or clusters are rejected at resolution time. `internal/app/multicluster_readiness_failclosed_test.go` guards this — do not weaken it casually.

The concrete per-cluster client factories are gated behind disabled-by-default flags (`DCP_RUNTIME_CLIENT_FACTORIES_ENABLED` plus one per capability: `..._KUBERNETES_ENABLED`, `..._TEKTON_ENABLED`, `..._ARGOCD_ENABLED`, and `DCP_RUNTIME_SECRET_LOADER_ENABLED`). With flags off, `buildRuntime*ClientFactory` returns `nil, nil` and the factory-aware registries fall back to the current-cluster clients. `runtime_non_regression_factories_disabled` tests assert that off means unchanged behavior.

### HTTP layer

`internal/api/router.go` uses `net/http.ServeMux` with two hand-rolled subrouters (`changeSubrouter`, `applicationSubrouter`) that split the path and switch on `(len(parts), parts[1], method)`. Adding an endpoint means adding a case there **and** classifying it in `requiredRolesForRequest`/`requiredRolesForAction` in `auth_middleware.go` — an unclassified route returns 403 when auth is enabled. The whole chain is `requestID → logging → auth → mux`.

Every response goes through `writeJSON`/`writeError` in `response.go` and uses the same envelope — `{data, meta, error}`, never a bare payload. `APIError` is deliberately operator-facing: `code`, `message`, plus optional `technicalMessage`, `suggestedAction` and a `recoverable` flag. Adapter failures must be translated into a **provider-neutral** code/message pair — `GIT_CREATE_BRANCH_FAILED` / "Unable to create Git branch for ChangeRequest", never `GITLAB_*` or a message naming GitLab — and `git_error_contract_test.go` enforces exactly that by scanning `change_handlers.go` for the required and forbidden literals.

Auth (`AUTH_ENABLED`, off by default) trusts identity headers set by an OAuth proxy (`X-Forwarded-User`, `X-Forwarded-Groups`), optionally resolving OpenShift groups via the API, and maps groups to four roles (`viewer` < `operator`/`approver` < `admin`). Read actions need any role; technical actions (`validate`, `start-build`, `update-gitops`, …) need `operator`/`admin`; lifecycle actions (`submit`, `approve`, `merge-request`, …) need `approver`/`admin`.

The `/ui/*` server-rendered dashboard lives entirely in `internal/api/ui_handlers.go` as one large `html/template` string with a `template.FuncMap`. It is a view over the same services — no separate frontend build.

### Change lifecycle vs. technical steps

Two orthogonal state axes on a ChangeRequest:

- **Lifecycle status** (`domain.ChangeStatus*`: draft → submitted → approved → executing → executed → closed, plus rejected/failed/cancelled), transitioned via `TransitionLifecycle` with `SELECT ... FOR UPDATE` in `internal/database/change_repository.go`. Concurrency behavior is covered by `change_repository_concurrency_test.go` (integration tag).
- **Runtime state** (`domain.ChangeRuntimeState`: source / artifact / gitops / tekton / argocd / runtime sections), persisted separately and driven by the technical step endpoints: `create-branch` → `update-files` → `open-merge-request` → `merge-request` → `start-build` → `check-build` → `update-gitops` → `validate` → `check-validation` → `check-deployment` → `collect-evidence`.

**GitLab merge race condition (PR #62/#63):** GitLab computes `detailed_merge_status` asynchronously. A `PUT /merge` issued immediately after opening the merge request returns **HTTP 405 Method Not Allowed** because the MR is still `checking`/`unchecked`. `internal/adapters/gitlab` must therefore call `Client.WaitUntilMergeable` (poll `detailed_merge_status==mergeable`, legacy `merge_status==can_be_merged` fallback, timeout, fail-closed) **before** `MergeMergeRequest`. This is the real root cause of the 405 (empirically: a bare `PUT /merge` returns 405 immediately after MR creation and 200 once mergeable). Omitting the stale `sha` (PR #62) is complementary but was not the fix — do not remove the mergeable-wait in refactors.

`internal/workflow` is a placeholder (`Engine.Run` is a TODO); orchestration currently lives in `ChangeService` and is driven by explicit API calls, not a background engine.

### Security invariants baked into tests

- TLS is secure-by-default in every adapter; `InsecureTLS`/`CAFile` handling is shared in `internal/adapters/tlsutil` and each adapter has a `tls_test.go` asserting the invariants. Keep them passing when touching client construction.
- Secrets are never stored in Git, logs, API responses, or evidence. Provider selections expose only `SafeSummary()`; the secret-refs model carries references, not values. Evidence payloads are marked `Sanitized`.

## Repository conventions

- **Documentation is English-only** (`docs/documentation-language-policy.md`). The single declared exception is `docs/final-technical-guide/final-technical-guide-it.md`. Some older files still contain Italian — `docs/adr/README.md`, and comments in `change_service.go`, `internal/database/change_repository.go` and `internal/adapters/gitlab/client.go`. New content must be English; leave the existing Italian alone unless the task is that migration (tracked in `docs/italian-documentation-inventory.md`, with `scripts/docs/inventory-italian-content.py` to re-scan).
- New docs must be registered in `docs/README.md` under the right section; new ADRs also in `docs/adr/README.md`. Accepted ADRs are not edited retroactively — supersede them with a new ADR.
- Filenames: lowercase kebab-case.
- `docs/` mixes living reference (`00`–`13`, design docs, runbooks) with point-in-time `*-results.md` validation evidence — the latter describe the state at a past commit and should not be read as current behavior.
- `main` is protected: changes land via pull request after the `test` check passes.
- Never commit real `.env`, tokens, kubeconfigs, or Secret manifests with real values.
- Change delivery: prefer **self-contained patches or ZIP packages** over large inline here-docs or monolithic scripts (here-docs risk truncation). Apply files explicitly by name (no wildcards).
- The Italian final technical guide (`-it` suffix) is the **source** for generating the DOCX via the dedicated script in `docs/final-technical-guide/scripts/`; keep it in sync when the English guide changes.
- Frozen ChangeRequests (e.g. `CHG-2026-0060`, `CHG-2026-0061`) are diagnostic/history and must not be advanced; validate new behavior with a **new** ChangeRequest.

## Finding your way around docs

`docs/README.md` is the index and is kept current — start there rather than globbing `docs/`. The fastest paths in: `docs/05-architecture.md` for the component model, `docs/13-api-design.md` for the API surface, `docs/adr/README.md` for why the design is the way it is, `docs/environment-configuration-model.md` and `docs/multi-environment-model.md` for the catalog/cluster model, and `docs/runbooks/` for anything operational (oauth-proxy, secrets rotation, backup/restore, health checks).
