package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/api"
	"github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/config"
	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type updateGitOpsChangeStore struct {
	change         domain.ChangeRequest
	markStepCalled bool
	markedStatus   string
}

func (s *updateGitOpsChangeStore) List(context.Context) ([]domain.ChangeRequest, error) {
	return nil, nil
}
func (s *updateGitOpsChangeStore) Create(context.Context, domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, errors.New("not implemented")
}
func (s *updateGitOpsChangeStore) Get(context.Context, string) (domain.ChangeRequest, error) {
	return s.change, nil
}
func (s *updateGitOpsChangeStore) Events(context.Context, string) ([]domain.ChangeEvent, error) {
	return nil, nil
}
func (s *updateGitOpsChangeStore) TransitionLifecycle(context.Context, string, string, string, string) (map[string]any, error) {
	return nil, errors.New("not implemented")
}
func (s *updateGitOpsChangeStore) MarkStep(_ context.Context, _ string, status string) (map[string]any, error) {
	s.markStepCalled = true
	s.markedStatus = status
	return map[string]any{"runtimeStatus": status}, nil
}

type updateGitOpsRuntimeStore struct {
	state domain.ChangeRuntimeState
}

func (s *updateGitOpsRuntimeStore) Get(context.Context, string) (domain.ChangeRuntimeState, error) {
	return s.state, nil
}
func (s *updateGitOpsRuntimeStore) UpsertSource(context.Context, string, domain.SourceRuntimeState) error {
	return nil
}
func (s *updateGitOpsRuntimeStore) UpsertArtifact(context.Context, string, domain.ArtifactRuntimeState) error {
	return nil
}
func (s *updateGitOpsRuntimeStore) UpsertGitOps(context.Context, string, domain.GitOpsRuntimeState) error {
	return nil
}
func (s *updateGitOpsRuntimeStore) UpsertTekton(context.Context, string, domain.TektonRuntimeState) error {
	return nil
}
func (s *updateGitOpsRuntimeStore) UpsertArgoCD(context.Context, string, domain.ArgoCDRuntimeState) error {
	return nil
}
func (s *updateGitOpsRuntimeStore) UpsertRuntime(context.Context, string, domain.RuntimeObservationState) error {
	return nil
}

type updateGitOpsProvider struct{}

func (updateGitOpsProvider) Provider() string    { return "github" }
func (updateGitOpsProvider) ProviderRef() string { return "github-public" }
func (updateGitOpsProvider) CreateBranch(context.Context, app.GitRepositoryTarget, string, string) error {
	return nil
}
func (updateGitOpsProvider) CreateOrUpdateFile(_ context.Context, _ app.GitRepositoryTarget, branch, filePath, _, _ string) (app.GitFileUpdateResult, error) {
	return app.GitFileUpdateResult{FilePath: filePath, Branch: branch, CommitSHA: "file-commit"}, nil
}
func (updateGitOpsProvider) OpenMergeRequest(context.Context, app.GitRepositoryTarget, string, string, string, string) (int, string, error) {
	return 3, "https://example/pr/3", nil
}
func (updateGitOpsProvider) MergeRequest(context.Context, app.GitRepositoryTarget, string, string, string) (int, string, string, error) {
	return 3, "https://example/pr/3", "merge-sha", nil
}

func newUpdateGitOpsHTTPServer(t *testing.T, digest string) (*httptest.Server, *updateGitOpsChangeStore) {
	t.Helper()
	changeStore := &updateGitOpsChangeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0075", ApplicationName: "demo-go-color-app", TargetEnvironment: "dev"}}
	runtimeStore := &updateGitOpsRuntimeStore{state: domain.ChangeRuntimeState{Artifact: domain.ArtifactRuntimeState{ImageRepository: "image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app", ImageDigest: digest}}}
	registry, err := app.NewGitProviderRegistry([]app.GitProvider{updateGitOpsProvider{}})
	if err != nil {
		t.Fatal(err)
	}
	service := app.NewChangeService(
		changeStore,
		app.WithChangeRuntimeStateStore(runtimeStore),
		app.WithGitOpsBindingResolverFunc(func(string) (app.RepositoryBinding, error) {
			return app.RepositoryBinding{Provider: "github", ProviderRef: "github-public", Role: app.RepositoryRoleGitOps, ProjectPath: "vincmarz/demo-app-gitops", RepositoryURL: "https://github.com/vincmarz/demo-app-gitops.git", DefaultBranch: "main"}, nil
		}),
		app.WithGitProviderResolver(registry),
		app.WithGitOpsImageKustomizationPath("apps/demo-go-color-app/kustomization.yaml"),
	)
	deps := api.Dependencies{Config: config.Config{}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Services: app.Services{Changes: service}}
	server := httptest.NewServer(api.NewRouter(deps))
	t.Cleanup(server.Close)
	return server, changeStore
}

func TestUpdateGitOpsAsOperatorReturnsAccepted(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_OPERATOR", "cp-operators")
	server, store := newUpdateGitOpsHTTPServer(t, "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
	response := doRequest(t, http.MethodPost, server.URL+"/api/v1/changes/CHG-2026-0075/update-gitops", "", operatorHeaders())
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("status=%d", response.StatusCode)
	}
	if store.markedStatus != "GitOpsUpdated" {
		t.Fatalf("marked=%q", store.markedStatus)
	}
}

func TestUpdateGitOpsFailureReturnsNeutralizedError(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_OPERATOR", "cp-operators")
	server, _ := newUpdateGitOpsHTTPServer(t, "not-a-digest")
	response := doRequest(t, http.MethodPost, server.URL+"/api/v1/changes/CHG-2026-0075/update-gitops", "", operatorHeaders())
	defer response.Body.Close()
	if response.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d", response.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	errorPayload, _ := payload["error"].(map[string]any)
	if errorPayload["code"] != "GITOPS_UPDATE_FAILED" {
		t.Fatalf("error=%#v", errorPayload)
	}
}

func TestUpdateGitOpsViewerIsForbidden(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_VIEWER", "cp-viewers")
	server, store := newUpdateGitOpsHTTPServer(t, "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
	response := doRequest(t, http.MethodPost, server.URL+"/api/v1/changes/CHG-2026-0075/update-gitops", "", viewerHeaders())
	defer response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status=%d", response.StatusCode)
	}
	if store.markStepCalled {
		t.Fatal("UpdateGitOps reached the service for a viewer")
	}
}
