package api_test

import (
	"context"
	"encoding/json"
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

type runtimeStateStoreFake struct {
	state domain.ChangeRuntimeState
	err   error
	id    string
}

func (f *runtimeStateStoreFake) Get(_ context.Context, id string) (domain.ChangeRuntimeState, error) {
	f.id = id
	return f.state, f.err
}
func (f *runtimeStateStoreFake) UpsertSource(context.Context, string, domain.SourceRuntimeState) error {
	return nil
}
func (f *runtimeStateStoreFake) UpsertArtifact(context.Context, string, domain.ArtifactRuntimeState) error {
	return nil
}
func (f *runtimeStateStoreFake) UpsertGitOps(context.Context, string, domain.GitOpsRuntimeState) error {
	return nil
}
func (f *runtimeStateStoreFake) UpsertTekton(context.Context, string, domain.TektonRuntimeState) error {
	return nil
}
func (f *runtimeStateStoreFake) UpsertArgoCD(context.Context, string, domain.ArgoCDRuntimeState) error {
	return nil
}
func (f *runtimeStateStoreFake) UpsertRuntime(context.Context, string, domain.RuntimeObservationState) error {
	return nil
}

func newRuntimeStateServer(t *testing.T, runtimeStore app.ChangeRuntimeStateStore) *httptest.Server {
	t.Helper()
	deps := api.Dependencies{
		Config:   config.Config{},
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Services: app.Services{Changes: app.NewChangeService(&fakeChangeStore{}, app.WithChangeRuntimeStateStore(runtimeStore))},
	}
	server := httptest.NewServer(api.NewRouter(deps))
	t.Cleanup(server.Close)
	return server
}

func TestGetChangeRuntimeStateReturnsAllSections(t *testing.T) {
	store := &runtimeStateStoreFake{state: domain.ChangeRuntimeState{
		ChangeRequestID: "change-id",
		Source:          domain.SourceRuntimeState{Provider: "github"},
		Artifact:        domain.ArtifactRuntimeState{Provider: "tekton", ImageDigest: "sha256:artifact"},
		GitOps:          domain.GitOpsRuntimeState{Revision: "main"},
		Tekton:          domain.TektonRuntimeState{PipelineRunName: "validation-run"},
		ArgoCD:          domain.ArgoCDRuntimeState{ApplicationName: "demo-go-color-app"},
		Runtime:         domain.RuntimeObservationState{ClusterName: "ocp-dev", Status: "Ready"},
	}}
	server := newRuntimeStateServer(t, store)

	response := doRequest(t, http.MethodGet, server.URL+"/api/v1/changes/CHG-50/runtime-state", "", nil)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	var envelope struct {
		Data domain.ChangeRuntimeState `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	if store.id != "CHG-50" || envelope.Data.ChangeRequestID != "change-id" {
		t.Fatalf("runtime state response = id:%q data:%#v", store.id, envelope.Data)
	}
	if envelope.Data.Source.Provider != "github" || envelope.Data.Artifact.ImageDigest != "sha256:artifact" || envelope.Data.Runtime.Status != "Ready" {
		t.Fatalf("runtime sections = %#v", envelope.Data)
	}
}

func TestGetChangeRuntimeStateReturnsEmptySections(t *testing.T) {
	store := &runtimeStateStoreFake{state: domain.ChangeRuntimeState{ChangeRequestID: "change-id"}}
	server := newRuntimeStateServer(t, store)

	response := doRequest(t, http.MethodGet, server.URL+"/api/v1/changes/CHG-51/runtime-state", "", nil)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	for _, section := range []string{"source", "artifact", "gitops", "tekton", "argocd", "runtime"} {
		if _, ok := envelope.Data[section]; !ok {
			t.Fatalf("missing section %q in %#v", section, envelope.Data)
		}
	}
}
