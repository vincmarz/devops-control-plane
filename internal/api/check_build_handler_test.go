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

type checkBuildChangeStore struct {
	change         domain.ChangeRequest
	markStepCalled bool
	markedStatus   string
}

func (s *checkBuildChangeStore) List(context.Context) ([]domain.ChangeRequest, error) {
	return nil, nil
}
func (s *checkBuildChangeStore) Create(context.Context, domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, errors.New("not implemented")
}
func (s *checkBuildChangeStore) Get(context.Context, string) (domain.ChangeRequest, error) {
	return s.change, nil
}
func (s *checkBuildChangeStore) Events(context.Context, string) ([]domain.ChangeEvent, error) {
	return nil, nil
}
func (s *checkBuildChangeStore) TransitionLifecycle(context.Context, string, string, string, string) (map[string]any, error) {
	return nil, errors.New("not implemented")
}
func (s *checkBuildChangeStore) MarkStep(_ context.Context, _ string, status string) (map[string]any, error) {
	s.markStepCalled = true
	s.markedStatus = status
	return map[string]any{"runtimeStatus": status}, nil
}

type checkBuildRuntimeStore struct {
	state    domain.ChangeRuntimeState
	artifact domain.ArtifactRuntimeState
}

func (s *checkBuildRuntimeStore) Get(context.Context, string) (domain.ChangeRuntimeState, error) {
	return s.state, nil
}
func (s *checkBuildRuntimeStore) UpsertSource(context.Context, string, domain.SourceRuntimeState) error {
	return nil
}
func (s *checkBuildRuntimeStore) UpsertArtifact(_ context.Context, _ string, state domain.ArtifactRuntimeState) error {
	s.artifact = state
	return nil
}
func (s *checkBuildRuntimeStore) UpsertGitOps(context.Context, string, domain.GitOpsRuntimeState) error {
	return nil
}
func (s *checkBuildRuntimeStore) UpsertTekton(context.Context, string, domain.TektonRuntimeState) error {
	return nil
}
func (s *checkBuildRuntimeStore) UpsertArgoCD(context.Context, string, domain.ArgoCDRuntimeState) error {
	return nil
}
func (s *checkBuildRuntimeStore) UpsertRuntime(context.Context, string, domain.RuntimeObservationState) error {
	return nil
}

func checkBuildDigestValue() string {
	return "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
}

func newCheckBuildHTTPServer(t *testing.T, status app.TektonBuildStatusResult) (*httptest.Server, *checkBuildChangeStore) {
	t.Helper()
	changeStore := &checkBuildChangeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0071", ApplicationName: "demo-go-color-app", TargetEnvironment: "dev"}}
	runtimeStore := &checkBuildRuntimeStore{state: domain.ChangeRuntimeState{Artifact: domain.ArtifactRuntimeState{Namespace: "devops-ci-demo", PipelineRunName: "build-run", PipelineRunUID: "build-uid", SourceCommitSHA: "1111111111111111111111111111111111111111"}}}
	service := app.NewChangeService(
		changeStore,
		app.WithChangeRuntimeStateStore(runtimeStore),
		app.WithTektonCheckBuild(func(context.Context, domain.ChangeRequest, string, string) (app.TektonBuildStatusResult, error) {
			return status, nil
		}),
	)
	deps := api.Dependencies{Config: config.Config{}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Services: app.Services{Changes: service}}
	server := httptest.NewServer(api.NewRouter(deps))
	t.Cleanup(server.Close)
	return server, changeStore
}

func TestCheckBuildAsOperatorReturnsAccepted(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_OPERATOR", "cp-operators")
	server, store := newCheckBuildHTTPServer(t, app.TektonBuildStatusResult{UID: "build-uid", Status: "True", Reason: "Succeeded", SourceCommit: "1111111111111111111111111111111111111111", ImageDigest: checkBuildDigestValue()})
	response := doRequest(t, http.MethodPost, server.URL+"/api/v1/changes/CHG-2026-0071/check-build", "", operatorHeaders())
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("status=%d", response.StatusCode)
	}
	if store.markedStatus != "BuildSucceeded" {
		t.Fatalf("marked=%q", store.markedStatus)
	}
}

func TestCheckBuildFailureReturnsNeutralizedError(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_OPERATOR", "cp-operators")
	server, _ := newCheckBuildHTTPServer(t, app.TektonBuildStatusResult{UID: "build-uid", Status: "True", SourceCommit: "1111111111111111111111111111111111111111"})
	response := doRequest(t, http.MethodPost, server.URL+"/api/v1/changes/CHG-2026-0071/check-build", "", operatorHeaders())
	defer response.Body.Close()
	if response.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d", response.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	errorPayload, _ := payload["error"].(map[string]any)
	if errorPayload["code"] != "TEKTON_CHECK_BUILD_FAILED" {
		t.Fatalf("error=%#v", errorPayload)
	}
}

func TestCheckBuildViewerIsForbidden(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_GROUP_VIEWER", "cp-viewers")
	server, store := newCheckBuildHTTPServer(t, app.TektonBuildStatusResult{UID: "build-uid", Status: "True", ImageDigest: checkBuildDigestValue()})
	response := doRequest(t, http.MethodPost, server.URL+"/api/v1/changes/CHG-2026-0071/check-build", "", viewerHeaders())
	defer response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status=%d", response.StatusCode)
	}
	if store.markStepCalled {
		t.Fatal("CheckBuild reached the service for a viewer")
	}
}
