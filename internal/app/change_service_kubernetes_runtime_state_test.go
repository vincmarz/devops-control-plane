package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func readyKubernetesRuntimeEvidence() map[string]any {
	return map[string]any{
		"namespace": "devops-ci-staging",
		"deployment": map[string]any{
			"name": "demo-go-color-app",
			"namespace": "devops-ci-staging",
			"generation": 4,
			"observedGeneration": 4,
			"desiredReplicas": 1,
			"readyReplicas": 1,
			"availableReplicas": 1,
			"updatedReplicas": 1,
		},
		"pods": []map[string]any{
			{"name": "demo-go-color-app-abc", "phase": "Running", "ready": true, "restartCount": 0},
		},
	}
}

func runtimeStateEvidenceService(t *testing.T, changeStore ChangeStore, evidenceStore EvidenceStore, runtimeStore ChangeRuntimeStateStore, runtimeEvidence map[string]any) *ChangeService {
	t.Helper()
	return NewChangeService(
		changeStore,
		WithEvidenceStore(evidenceStore),
		WithChangeRuntimeStateStore(runtimeStore),
		WithTechnicalRuntimeTargetResolverFunc(func(context.Context, domain.ChangeRequest) (TechnicalRuntimeTarget, error) {
			return TechnicalRuntimeTarget{
				TargetEnvironment: "staging",
				ClusterName: "ocp-dev",
				KubernetesNamespace: "devops-ci-staging",
				ArgoCDApplicationName: "demo-go-color-app-staging",
			}, nil
		}),
		WithDeploymentEvidenceCollector(func(context.Context, domain.ChangeRequest, TechnicalRuntimeTarget) (domain.Evidence, error) {
			return domain.Evidence{EvidenceType: "deployment", Name: "deployment-evidence", Payload: map[string]any{}}, nil
		}),
		WithKubernetesRuntimeEvidenceCollector(func(context.Context, domain.ChangeRequest) (map[string]any, error) {
			return runtimeEvidence, nil
		}),
	)
}

func TestCollectEvidencePersistsReadyKubernetesRuntimeState(t *testing.T) {
	changeStore := &collectEvidenceFakeChangeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-40", ApplicationName: "demo-go-color-app", TargetEnvironment: "staging"}}
	evidenceStore := &collectEvidenceFakeEvidenceStore{}
	runtimeStore := &sourceRuntimeStateStoreFake{}
	service := runtimeStateEvidenceService(t, changeStore, evidenceStore, runtimeStore, readyKubernetesRuntimeEvidence())

	if _, err := service.CollectEvidence(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	if !runtimeStore.runtimeCalled {
		t.Fatal("Kubernetes runtime state was not persisted")
	}
	got := runtimeStore.runtime
	if got.ClusterName != "ocp-dev" || got.Namespace != "devops-ci-staging" {
		t.Fatalf("runtime target identity = %#v", got)
	}
	if got.ResourceKind != "Deployment" || got.ResourceName != "demo-go-color-app" {
		t.Fatalf("runtime resource identity = %#v", got)
	}
	if got.Status != "Ready" || got.Reason != "ReplicasAvailable" {
		t.Fatalf("runtime status = %#v", got)
	}
	if !evidenceStore.createCalled || changeStore.markedStatus != "EvidenceCollected" {
		t.Fatalf("evidence flow = create:%v status:%q", evidenceStore.createCalled, changeStore.markedStatus)
	}
}

func TestCollectEvidencePersistsProgressingKubernetesRuntimeState(t *testing.T) {
	changeStore := &collectEvidenceFakeChangeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-41", ApplicationName: "demo-go-color-app", TargetEnvironment: "staging"}}
	evidenceStore := &collectEvidenceFakeEvidenceStore{}
	runtimeStore := &sourceRuntimeStateStoreFake{}
	evidence := readyKubernetesRuntimeEvidence()
	deployment := evidence["deployment"].(map[string]any)
	deployment["readyReplicas"] = 0
	deployment["availableReplicas"] = 0
	pods := evidence["pods"].([]map[string]any)
	pods[0]["ready"] = false
	service := runtimeStateEvidenceService(t, changeStore, evidenceStore, runtimeStore, evidence)

	if _, err := service.CollectEvidence(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	got := runtimeStore.runtime
	if got.Status != "Progressing" || got.Reason != "ReplicasPending" {
		t.Fatalf("runtime status = %#v", got)
	}
}

func TestCollectEvidenceDoesNotSaveEvidenceOrMarkStepWhenRuntimePersistenceFails(t *testing.T) {
	changeStore := &collectEvidenceFakeChangeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-42", ApplicationName: "demo-go-color-app", TargetEnvironment: "staging"}}
	evidenceStore := &collectEvidenceFakeEvidenceStore{}
	runtimeStore := &sourceRuntimeStateStoreFake{runtimeErr: errors.New("database unavailable")}
	service := runtimeStateEvidenceService(t, changeStore, evidenceStore, runtimeStore, readyKubernetesRuntimeEvidence())

	_, err := service.CollectEvidence(context.Background(), changeStore.change.ChangeNumber)
	if err == nil || !strings.Contains(err.Error(), "persist Kubernetes runtime state after collecting evidence") {
		t.Fatalf("error = %v", err)
	}
	if evidenceStore.createCalled {
		t.Fatal("EvidenceStore.Create was called after runtime persistence failure")
	}
	if changeStore.markedStatus != "" {
		t.Fatalf("MarkStep was called with %q", changeStore.markedStatus)
	}
}

func TestCollectEvidenceWithoutKubernetesCollectorDoesNotPersistRuntimeState(t *testing.T) {
	changeStore := &collectEvidenceFakeChangeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-43", ApplicationName: "demo-go-color-app", TargetEnvironment: "staging"}}
	evidenceStore := &collectEvidenceFakeEvidenceStore{}
	runtimeStore := &sourceRuntimeStateStoreFake{}
	service := NewChangeService(
		changeStore,
		WithEvidenceStore(evidenceStore),
		WithChangeRuntimeStateStore(runtimeStore),
		WithTechnicalRuntimeTargetResolverFunc(func(context.Context, domain.ChangeRequest) (TechnicalRuntimeTarget, error) {
			return TechnicalRuntimeTarget{ClusterName: "ocp-dev", KubernetesNamespace: "devops-ci-staging", ArgoCDApplicationName: "demo-go-color-app-staging"}, nil
		}),
		WithDeploymentEvidenceCollector(func(context.Context, domain.ChangeRequest, TechnicalRuntimeTarget) (domain.Evidence, error) {
			return domain.Evidence{Payload: map[string]any{}}, nil
		}),
	)

	if _, err := service.CollectEvidence(context.Background(), changeStore.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	if runtimeStore.runtimeCalled {
		t.Fatal("runtime state was persisted without Kubernetes evidence")
	}
}
