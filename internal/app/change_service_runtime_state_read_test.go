package app

import (
	"context"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func TestGetRuntimeStateReturnsPersistedSections(t *testing.T) {
	runtimeStore := &sourceRuntimeStateStoreFake{current: domain.ChangeRuntimeState{
		ChangeRequestID: "change-id",
		Source:          domain.SourceRuntimeState{Provider: "github", Branch: "change/CHG-50"},
		Artifact:        domain.ArtifactRuntimeState{Provider: "tekton", ImageDigest: "sha256:artifact"},
		GitOps:          domain.GitOpsRuntimeState{Provider: "github", Revision: "main"},
		Tekton:          domain.TektonRuntimeState{PipelineRunName: "validation-run", Status: "True"},
		ArgoCD:          domain.ArgoCDRuntimeState{ApplicationName: "demo-go-color-app", SyncStatus: "Synced"},
		Runtime:         domain.RuntimeObservationState{ClusterName: "ocp-dev", Namespace: "devops-ci-demo", Status: "Ready"},
	}}
	service := NewChangeService(&runtimeStateReadChangeStore{}, WithChangeRuntimeStateStore(runtimeStore))

	state, err := service.GetRuntimeState(context.Background(), "CHG-50")
	if err != nil {
		t.Fatal(err)
	}
	if !runtimeStore.getCalled || runtimeStore.idOrNumber != "CHG-50" {
		t.Fatalf("runtime store call = called:%v id:%q", runtimeStore.getCalled, runtimeStore.idOrNumber)
	}
	if state.Source.Provider != "github" || state.Artifact.ImageDigest != "sha256:artifact" || state.GitOps.Revision != "main" || state.Tekton.PipelineRunName != "validation-run" {
		t.Fatalf("repository and validation state = %#v", state)
	}
	if state.ArgoCD.ApplicationName != "demo-go-color-app" || state.Runtime.Status != "Ready" {
		t.Fatalf("deployment state = %#v", state)
	}
}

func TestGetRuntimeStateRejectsEmptyIdentifier(t *testing.T) {
	service := NewChangeService(&runtimeStateReadChangeStore{}, WithChangeRuntimeStateStore(&sourceRuntimeStateStoreFake{}))
	_, err := service.GetRuntimeState(context.Background(), "   ")
	if err == nil || !strings.Contains(err.Error(), "change id or number is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestGetRuntimeStateRequiresConfiguredStore(t *testing.T) {
	service := NewChangeService(&runtimeStateReadChangeStore{})
	_, err := service.GetRuntimeState(context.Background(), "CHG-50")
	if err == nil || !strings.Contains(err.Error(), "change runtime state store is not configured") {
		t.Fatalf("error = %v", err)
	}
}

type runtimeStateReadChangeStore struct{}

func (runtimeStateReadChangeStore) List(context.Context) ([]domain.ChangeRequest, error) {
	return nil, nil
}
func (runtimeStateReadChangeStore) Create(context.Context, domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, nil
}
func (runtimeStateReadChangeStore) Get(context.Context, string) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, nil
}
func (runtimeStateReadChangeStore) Events(context.Context, string) ([]domain.ChangeEvent, error) {
	return nil, nil
}
func (runtimeStateReadChangeStore) TransitionLifecycle(context.Context, string, string, string, string) (map[string]any, error) {
	return nil, nil
}
func (runtimeStateReadChangeStore) MarkStep(context.Context, string, string) (map[string]any, error) {
	return nil, nil
}
