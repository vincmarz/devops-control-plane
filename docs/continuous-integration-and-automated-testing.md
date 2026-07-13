# DevOps Control Plane - Continuous Integration and Automated Testing

Status: Current baseline  
Owner: Vincenzo Marzario  
Language: English  
Last updated: 2026-07-13

## 1. Purpose

This document describes the current Continuous Integration (CI) and automated testing baseline of the DevOps Control Plane project.

The CI baseline provides automated quality gates for changes submitted through GitHub pull requests or pushed to the `main` branch. It verifies source formatting, static analysis, unit tests, race conditions, code coverage, PostgreSQL integration behavior, HTTP endpoint behavior, lifecycle concurrency, and secure-by-default TLS configuration.

This document is the authoritative Markdown source for the CI and automated testing capabilities introduced after the initial final technical guide was generated.

The CI pipeline complements, but does not replace:

- local developer validation;
- OpenShift runtime smoke tests;
- operational health checks;
- maintenance runbooks;
- real-cluster onboarding validation;
- physical multi-cluster validation.

## 2. Current CI baseline

The current CI baseline is implemented in:

```text
.github/workflows/ci.yml
```

The workflow is named:

```text
CI
```

The CI job runs on:

```text
ubuntu-latest
```

The pipeline uses:

```text
Go 1.22
PostgreSQL 16
```

The current quality gates are:

1. repository checkout;
2. Go toolchain setup with dependency caching;
3. `gofmt` validation;
4. `go vet ./...`;
5. unit tests with race detection and atomic coverage;
6. PostgreSQL integration tests;
7. coverage artifact upload.

The workflow has been exercised through pull requests and has already identified formatting issues that were corrected before merge.

## 3. GitHub Actions workflow

The GitHub Actions workflow performs the following sequence:

```text
Checkout
   |
   v
Set up Go 1.22
   |
   v
Check formatting with gofmt
   |
   v
Run go vet
   |
   v
Run unit tests with race detector and coverage
   |
   v
Run PostgreSQL integration tests
   |
   v
Upload coverage artifact
```

A disposable PostgreSQL service container is created for each workflow run. The integration test suite can therefore initialize or reset its test schema without affecting a persistent environment.

## 4. Trigger model

The CI workflow runs on:

```yaml
push:
  branches: [main]
pull_request:
```

This means:

- every pull request is validated before merge;
- every push to `main` is validated;
- changes from test or feature branches are checked through their pull requests;
- the merged state of `main` is checked again after integration.

The pull request model provides a traceable quality gate between implementation and the protected project baseline.

## 5. Formatting gate

The formatting step runs:

```bash
unformatted="$(gofmt -l .)"
```

If one or more files are not `gofmt`-clean, the workflow:

- prints the affected paths;
- exits with a non-zero status;
- blocks successful completion of the CI job.

The formatting gate has already detected formatting differences in newly added tests. The affected files were corrected through dedicated style commits before the related pull requests were merged.

Local equivalent:

```bash
gofmt -l .
```

Expected result:

```text
no output
```

## 6. Static analysis with go vet

The static analysis step runs:

```bash
go vet ./...
```

`go vet` checks Go packages for suspicious constructs and common correctness problems that may compile successfully but still indicate defects.

This gate applies to the complete repository package tree.

A successful CI run requires `go vet ./...` to complete without errors.

## 7. Unit tests, race detector, and coverage

The main test step runs:

```bash
go test ./... -race -covermode=atomic -coverprofile=coverage.out
```

This step provides three controls.

### 7.1 Unit and package tests

All standard tests under the repository package tree are executed, except tests that require the `integration` build tag.

This keeps the main test suite fast and hermetic while still covering application, API, configuration, adapter, and database logic that does not require a live PostgreSQL service.

### 7.2 Race detection

The `-race` option enables the Go race detector.

This helps identify unsafe concurrent access to shared memory during automated tests.

The race detector complements the database-level concurrency test described later. The two controls address different layers:

- the Go race detector checks in-process memory access;
- the PostgreSQL concurrency test checks lifecycle transition atomicity across concurrent transactions.

### 7.3 Atomic coverage

Coverage is generated with:

```text
-covermode=atomic
-coverprofile=coverage.out
```

Atomic coverage mode is compatible with concurrent test execution and the race detector.

The resulting file is uploaded as a GitHub Actions artifact named:

```text
coverage
```

The current workflow produces and preserves the coverage artifact, but it does not yet enforce a minimum coverage percentage.

## 8. PostgreSQL integration testing

The workflow starts a disposable PostgreSQL 16 service container.

CI-only service configuration:

```text
POSTGRES_USER=dcp
POSTGRES_PASSWORD=dcp
POSTGRES_DB=dcp_test
```

The test connection is exposed through:

```text
TEST_DATABASE_URL=postgres://dcp:dcp@localhost:5432/dcp_test?sslmode=disable
```

These values are disposable CI credentials for the isolated GitHub Actions service container. They are not production or OpenShift runtime credentials.

The PostgreSQL health check uses:

```bash
pg_isready -U dcp
```

The integration suite runs separately from the main unit-test step:

```bash
go test -tags=integration ./internal/database/... -run Integration -v
```

Primary integration test source:

```text
internal/database/change_repository_integration_test.go
```

The integration coverage includes:

- a complete ChangeRequest lifecycle against PostgreSQL;
- persistence behavior across lifecycle transitions;
- invalid lifecycle transition rejection;
- database-backed repository behavior using the real PostgreSQL engine.

Separating integration tests through the `integration` build tag keeps the default local and CI unit suite independent from external database availability.

## 9. HTTP end-to-end test coverage

The automated test baseline includes HTTP-level tests that exercise handlers, middleware, routing, authorization behavior, request validation, and application service integration.

### 9.1 Health and readiness endpoints

Source:

```text
internal/api/health_handlers_test.go
```

Covered scenarios include:

- health endpoint success;
- readiness success when PostgreSQL is healthy;
- readiness failure when PostgreSQL is unavailable.

Relevant tests:

```text
TestHealthzEndpoint
TestReadyzWhenDatabaseHealthy
TestReadyzWhenDatabaseDown
```

These tests validate both the public HTTP behavior and the dependency-aware readiness contract.

### 9.2 Authenticated ChangeRequest API

Source:

```text
internal/api/change_handlers_test.go
```

Covered scenarios include:

- authenticated viewer listing changes;
- authentication required for protected endpoints;
- viewer denied on mutating operations;
- operator allowed to create a ChangeRequest;
- malformed JSON rejection;
- domain validation error propagation.

Relevant tests:

```text
TestListChangesAuthenticatedViewer
TestListChangesRequiresAuthentication
TestCreateChangeForbiddenForViewer
TestCreateChangeAsOperator
TestCreateChangeRejectsMalformedJSON
TestCreateChangeValidationError
```

### 9.3 Lifecycle transition routes

Source:

```text
internal/api/change_lifecycle_handlers_test.go
```

Covered scenarios include:

- approval by an approver;
- approval denied for a viewer;
- invalid lifecycle transition rejection;
- missing lifecycle actor rejection.

Relevant tests:

```text
TestApproveChangeAsApprover
TestApproveChangeForbiddenForViewer
TestApproveChangeInvalidTransition
TestSubmitChangeMissingActor
```

### 9.4 Existing authentication and UI behavior tests

The CI suite also executes previously existing tests for:

- authentication middleware;
- viewer, operator, approver, and admin roles;
- public health endpoint exceptions;
- OpenShift group resolution fallback;
- technical action visibility;
- dashboard environment summaries;
- latest Tekton validation evidence selection.

These tests ensure that the new HTTP end-to-end coverage remains compatible with the established AuthN/AuthZ and UI behavior baseline.

## 10. Lifecycle concurrency invariant

Concurrent lifecycle transitions require database-level serialization.

The repository implementation now uses:

```sql
SELECT ... FOR UPDATE
```

in the lifecycle transition path.

Implementation source:

```text
internal/database/change_repository.go
```

Concurrency test source:

```text
internal/database/change_repository_concurrency_test.go
```

Primary invariant test:

```text
TestConcurrentApproveOnlyOneWins
```

The test repeatedly issues concurrent approval attempts against the same ChangeRequest and verifies that exactly one approval succeeds.

The invariant is:

```text
A ChangeRequest lifecycle transition must be atomic.
Concurrent approval attempts must not both succeed.
```

This protects the lifecycle state machine from a read-modify-write race in which multiple transactions could otherwise observe the same initial state and apply conflicting transitions.

The control combines:

- Go concurrency in the test;
- PostgreSQL row-level locking;
- lifecycle transition validation;
- repeated concurrent attempts to make the test more reliable.

## 11. TLS secure-by-default invariants

The test baseline explicitly protects the requirement that TLS certificate verification is enabled by default.

Disabling TLS verification requires an explicit opt-in configuration.

### 11.1 Configuration defaults

Source:

```text
internal/config/tls_defaults_test.go
```

Relevant tests:

```text
TestLoadDefaultsInsecureTLSFlagsToFalse
TestLoadParsesInsecureTLSFlagsWhenExplicitlyEnabled
```

The default behavior is tested for:

```text
ARGOCD_INSECURE_TLS
GITLAB_INSECURE_TLS
KUBERNETES_INSECURE_TLS
```

When these variables are absent or empty, the corresponding configuration flags must remain `false`, which keeps TLS certificate verification enabled.

The insecure mode is accepted only when explicitly configured.

### 11.2 Argo CD adapter

Source:

```text
internal/adapters/argocd/tls_test.go
```

Relevant tests:

```text
TestNewVerifiesTLSByDefault
TestNewDisablesTLSVerificationOnlyWhenRequested
```

### 11.3 GitLab adapter

Source:

```text
internal/adapters/gitlab/tls_test.go
```

Relevant tests:

```text
TestNewVerifiesTLSByDefault
TestNewDisablesTLSVerificationOnlyWhenRequested
```

### 11.4 Kubernetes adapter

Source:

```text
internal/adapters/kubernetes/tls_test.go
```

Relevant tests:

```text
TestNewVerifiesTLSByDefault
TestNewDisablesTLSVerificationOnlyWhenRequested
```

### 11.5 Tekton adapter

Source:

```text
internal/adapters/tekton/tls_test.go
```

Relevant tests:

```text
TestNewVerifiesTLSByDefault
TestNewDisablesTLSVerificationOnlyWhenRequested
```

For each adapter, the tested invariant is:

```text
Default client construction must not set InsecureSkipVerify=true.
InsecureSkipVerify=true is accepted only after explicit InsecureTLS opt-in.
```

These tests reduce the risk that a future refactoring silently weakens transport security.

## 12. Pull request validation model

The CI baseline was introduced and expanded incrementally through GitHub pull requests.

Relevant pull requests:

```text
PR #1 - GitHub Actions workflow and PostgreSQL integration test
PR #2 - Health and readiness HTTP end-to-end tests
PR #3 - Authenticated ChangeRequest API end-to-end tests
PR #4 - Lifecycle transition HTTP end-to-end tests
PR #5 - Concurrent approval serialization and SELECT FOR UPDATE
PR #6 - Configuration and Argo CD TLS invariants
PR #7 - GitLab, Kubernetes, and Tekton TLS invariants
```

The workflow detected formatting differences in newly added tests. Dedicated `gofmt` commits corrected these issues before merge.

This provides evidence that CI is acting as an effective quality gate rather than only as a passive workflow definition.

## 13. Local developer validation

Before opening or merging a pull request, developers should run the local equivalent of the CI checks where possible.

Minimum checks:

```bash
gofmt -l .
go vet ./...
go test ./...
```

Expected `gofmt` result:

```text
no output
```

The PostgreSQL integration tests additionally require a reachable test database and the `TEST_DATABASE_URL` environment variable.

Example:

```bash
go test -tags=integration ./internal/database/... -run Integration -v
```

The current baseline was also validated locally with:

```bash
go test ./...
```

All tested packages completed successfully.

## 14. Security considerations

The CI baseline follows these security rules:

- production credentials must not be committed to the repository;
- CI PostgreSQL credentials are disposable and limited to the service container;
- test tokens such as `test-token` are placeholders used only inside isolated tests;
- TLS verification remains enabled by default;
- insecure TLS requires explicit opt-in;
- CI output must not print real Secret values;
- coverage artifacts must not contain credentials;
- pull requests must not include raw tokens, kubeconfig files, or private keys.

The pipeline complements the existing Secret reference, RBAC, evidence sanitization, and fail-closed security models.

## 15. Current limitations

The current CI baseline does not yet:

- deploy the DevOps Control Plane to OpenShift;
- run runtime smoke tests against `ocp-dev`;
- run Argo CD integration tests against a live Argo CD instance;
- run Tekton integration tests against a live Tekton control plane;
- perform physical multi-cluster validation;
- run browser-based UI tests;
- build or publish a production container image;
- enforce a minimum coverage threshold;
- execute scheduled nightly or periodic test workflows;
- replace operational health-check and maintenance runbooks.

These limitations must be considered when interpreting a successful CI result.

A green CI run demonstrates source-level, package-level, HTTP-level, PostgreSQL integration, concurrency, and TLS invariant validation. It does not prove that the active OpenShift runtime is healthy.

## 16. Evidence and relevant commits

Initial CI and PostgreSQL integration:

```text
eb799f2 ci: add GitHub Actions workflow and Postgres integration test
716568b ci: restore full content of Postgres integration test (was committed empty)
aa166b7 style: gofmt existing files to satisfy CI formatting check
a9c8eb9 Merge pull request #1 from vincmarz/ci/add-workflow-and-integration-tests
```

HTTP end-to-end tests:

```text
cc58bf1 test: add HTTP end-to-end tests for health and readiness endpoints
7e27099 Merge pull request #2 from vincmarz/test/http-health-e2e
087665f test: add HTTP end-to-end tests for authenticated changes API
c82f13e Merge pull request #3 from vincmarz/test/http-changes-api
a28e576 test: add HTTP end-to-end tests for lifecycle transition routes
1cd2bb1 Merge pull request #4 from vincmarz/test/http-lifecycle
```

Lifecycle concurrency:

```text
3673340 fix: serialize lifecycle transitions with SELECT FOR UPDATE; add concurrency test
9f6e4aa style: gofmt concurrency test to satisfy CI formatting check
1c92bd4 Merge pull request #5 from vincmarz/test/concurrency-approve
```

TLS secure-by-default invariants:

```text
5343de6 test: assert TLS verification is secure-by-default in config and argocd client
ea84314 Merge pull request #6 from vincmarz/test/security-tls-invariants
805e213 test: assert TLS secure-by-default in gitlab, kubernetes and tekton clients
f1d8fb1 style: gofmt adapter TLS tests to satisfy CI formatting check
ca4bef3 Merge pull request #7 from vincmarz/test/security-tls-adapters
```

Current merged baseline:

```text
ca4bef3 Merge pull request #7 from vincmarz/test/security-tls-adapters
```

## 17. Definition of Done

The current CI and automated testing baseline is considered complete when:

- `.github/workflows/ci.yml` exists and is active;
- the workflow runs on pull requests and pushes to `main`;
- `gofmt` validation is enforced;
- `go vet ./...` is enforced;
- unit tests run with the race detector;
- atomic coverage is generated and uploaded;
- PostgreSQL 16 integration tests run against a disposable service container;
- health and readiness HTTP tests are present;
- authenticated ChangeRequest API tests are present;
- lifecycle HTTP tests are present;
- concurrent approval is serialized with `SELECT FOR UPDATE`;
- the concurrent approval invariant is covered by an automated test;
- TLS verification is tested as secure-by-default for configuration, Argo CD, GitLab, Kubernetes, and Tekton;
- local `gofmt -l .` returns no output;
- local `go test ./...` succeeds;
- the CI baseline and its limitations are documented in this file.
