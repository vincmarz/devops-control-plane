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

type startBuildChangeStore struct {
	change         domain.ChangeRequest
	markStepCalled bool
	markedStatus   string
}

func (s *startBuildChangeStore) List(context.Context) ([]domain.ChangeRequest, error) {
	return nil, nil
}
func (s *startBuildChangeStore) Create(context.Context, domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, errors.New("not implemented")
}
func (s *startBuildChangeStore) Get(context.Context, string) (domain.ChangeRequest, error) {
	return s.change, nil
}
func (s *startBuildChangeStore) Events(context.Context, string) ([]domain.ChangeEvent, error) {
	return nil, nil
}
func (s *startBuildChangeStore) TransitionLifecycle(context.Context, string, string, string, string) (map[string]any, error) {
	return nil, errors.New("not implemented")
}
func (s *startBuildChangeStore) MarkStep(_ context.Context, _ string, status string) (map[string]any, error) {
	s.markStepCalled = true
	s.markedStatus = status
	return map[string]any{"runtimeStatus": status}, nil
}

type startBuildRuntimeStore struct {
	state       domain.ChangeRuntimeState
	artifact    domain.ArtifactRuntimeState
	artifactErr error
}

func (s *startBuildRuntimeStore) Get(context.Context, string) (domain.ChangeRuntimeState, error) {
	return s.state, nil
}
func (s *startBuildRuntimeStore) UpsertSource(context.Context, string, domain.SourceRuntimeState) error {
	return nil
}
func (s *startBuildRuntimeStore) UpsertArtifact(_ context.Context, _ string, state domain.ArtifactRuntimeState) error {
	s.artifact = state
	return s.artifactErr
}
func (s *startBuildRuntimeStore) UpsertGitOps(context.Context, string, domain.GitOpsRuntimeState) error {
	return nil
}
func (s *startBuildRuntimeStore) UpsertTekton(context.Context, string, domain.TektonRuntimeState) error {
	return nil
}
func (s *startBuildRuntimeStore) UpsertArgoCD(context.Context, string, domain.ArgoCDRuntimeState) error {
	return nil
}
func (s *startBuildRuntimeStore) UpsertRuntime(context.Context, string, domain.RuntimeObservationState) error {
	return nil
}

func newStartBuildHTTPServer(t *testing.T, runtimeStore *startBuildRuntimeStore) (*httptest.Server, *startBuildChangeStore) {
	t.Helper()
	changeStore := &startBuildChangeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0065", ApplicationName: "demo-go-color-app", TargetEnvironment: "dev"}}
	binding := app.RepositoryBinding{Provider: "github", ProviderRef: "github-public", Role: app.RepositoryRoleSource, ProjectPath: "org/app", RepositoryURL: "https://github.com/org/app.git", DefaultBranch: "main", WorkflowEnabled: true}
	service := app.NewChangeService(
		changeStore,
		app.WithGitSourceBindingResolverFunc(func(string) (app.RepositoryBinding, error) { return binding, nil }),
		app.WithChangeRuntimeStateStore(runtimeStore),
		app.WithTektonStartBuild(func(context.Context, domain.ChangeRequest, app.TektonStartBuildRequest) (app.TektonBuildRunResult, error) {
			return app.TektonBuildRunResult{Namespace: "devops-ci-demo", PipelineName: "go-build-and-push", PipelineRunName: "build-run", PipelineRunUID: "build-uid"}, nil
		}, "image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app"),
	)
	deps := api.Dependencies{Config: config.Config{}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Services: app.Services{Changes: service}}
	server := httptest.NewServer(api.NewRouter(deps))
	t.Cleanup(server.Close)
	return server, changeStore
}

func validStartBuildRuntimeState() domain.ChangeRuntimeState {
	return domain.ChangeRuntimeState{Source: domain.SourceRuntimeState{Provider: "github", ProviderRef: "github-public", RepositoryURL: "https://github.com/org/app.git", ProposalState: "merged", MergeCommitSHA: "abcdef0123456789abcdef0123456789abcdef01"}}
}

func TestStartBuildAsOperatorReturnsAccepted(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_OPERATOR", "cp-operators")
	server, store := newStartBuildHTTPServer(t, &startBuildRuntimeStore{state: validStartBuildRuntimeState()})

	response := doRequest(t, http.MethodPost, server.URL+"/api/v1/changes/CHG-2026-0065/start-build", "", operatorHeaders())
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusAccepted)
	}
	if !store.markStepCalled || store.markedStatus != "BuildRunning" {
		t.Fatalf("MarkStep = called:%v status:%q", store.markStepCalled, store.markedStatus)
	}
}

func TestStartBuildFailureReturnsNeutralizedError(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_OPERATOR", "cp-operators")
	runtimeStore := &startBuildRuntimeStore{state: validStartBuildRuntimeState(), artifactErr: errors.New("database unavailable")}
	server, _ := newStartBuildHTTPServer(t, runtimeStore)

	response := doRequest(t, http.MethodPost, server.URL+"/api/v1/changes/CHG-2026-0065/start-build", "", operatorHeaders())
	defer response.Body.Close()
	if response.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusUnprocessableEntity)
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	errorPayload, _ := payload["error"].(map[string]any)
	if errorPayload["code"] != "TEKTON_START_BUILD_FAILED" {
		t.Fatalf("error = %#v", errorPayload)
	}
	if errorPayload["message"] != "Unable to start Tekton application build for ChangeRequest" {
		t.Fatalf("message = %#v", errorPayload["message"])
	}
}

func TestStartBuildViewerIsForbidden(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_VIEWER", "cp-viewers")
	server, store := newStartBuildHTTPServer(t, &startBuildRuntimeStore{state: validStartBuildRuntimeState()})

	response := doRequest(t, http.MethodPost, server.URL+"/api/v1/changes/CHG-2026-0065/start-build", "", viewerHeaders())
	defer response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
	if store.markStepCalled {
		t.Fatal("StartBuild reached the service for a viewer")
	}
}
