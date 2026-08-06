package app

import (
	"context"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

const checkBuildDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func checkBuildRuntimeState() domain.ChangeRuntimeState {
	return domain.ChangeRuntimeState{
		Artifact: domain.ArtifactRuntimeState{
			Provider:        "tekton",
			Namespace:       "devops-ci-demo",
			PipelineName:    "go-build-and-push",
			PipelineRunName: "devops-cp-build-chg-2026-0066-abcde",
			PipelineRunUID:  "build-uid",
			SourceCommitSHA: "1111111111111111111111111111111111111111",
			ImageRepository: "image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app",
			ImageTag:        "chg-2026-0066-1111111",
			Status:          "Unknown",
			Reason:          "Running",
		},
	}
}

func newCheckBuildService(store *validateFakeStore, runtimeStore *sourceRuntimeStateStoreFake, fn TektonCheckBuildFunc) *ChangeService {
	return NewChangeService(store, WithChangeRuntimeStateStore(runtimeStore), WithTektonCheckBuild(fn))
}

func TestCheckBuildSucceededPersistsDigestBeforeMarkStep(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0066"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: checkBuildRuntimeState()}
	service := newCheckBuildService(store, runtimeStore, func(_ context.Context, _ domain.ChangeRequest, namespace string, pipelineRunName string) (TektonBuildStatusResult, error) {
		if namespace != "devops-ci-demo" || pipelineRunName != "devops-cp-build-chg-2026-0066-abcde" {
			t.Fatalf("namespace=%q name=%q", namespace, pipelineRunName)
		}
		return TektonBuildStatusResult{UID: "build-uid", Status: "True", Reason: "Succeeded", SourceCommit: "1111111111111111111111111111111111111111", ImageDigest: checkBuildDigest}, nil
	})

	if _, err := service.CheckBuild(context.Background(), store.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	if !runtimeStore.artifactCalled {
		t.Fatal("artifact state was not persisted")
	}
	if runtimeStore.artifact.ImageDigest != checkBuildDigest {
		t.Fatalf("digest=%q", runtimeStore.artifact.ImageDigest)
	}
	if !store.markStepCalled || store.markedStatus != "BuildSucceeded" {
		t.Fatalf("MarkStep=called:%v status:%q", store.markStepCalled, store.markedStatus)
	}
}

func TestCheckBuildFailed(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0067"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: checkBuildRuntimeState()}
	service := newCheckBuildService(store, runtimeStore, func(context.Context, domain.ChangeRequest, string, string) (TektonBuildStatusResult, error) {
		return TektonBuildStatusResult{UID: "build-uid", Status: "False", Reason: "Failed", Message: "buildah failed"}, nil
	})
	if _, err := service.CheckBuild(context.Background(), store.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	if store.markedStatus != "BuildFailed" {
		t.Fatalf("status=%q", store.markedStatus)
	}
	if runtimeStore.artifact.ImageDigest != "" {
		t.Fatalf("unexpected digest=%q", runtimeStore.artifact.ImageDigest)
	}
}

func TestCheckBuildRunning(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0068"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: checkBuildRuntimeState()}
	service := newCheckBuildService(store, runtimeStore, func(context.Context, domain.ChangeRequest, string, string) (TektonBuildStatusResult, error) {
		return TektonBuildStatusResult{UID: "build-uid", Status: "Unknown", Reason: "Running"}, nil
	})
	if _, err := service.CheckBuild(context.Background(), store.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	if store.markedStatus != "BuildRunning" {
		t.Fatalf("status=%q", store.markedStatus)
	}
}

func TestCheckBuildFailsClosed(t *testing.T) {
	cases := []struct {
		name   string
		result TektonBuildStatusResult
		want   string
	}{
		{name: "succeeded without digest", result: TektonBuildStatusResult{UID: "build-uid", Status: "True", SourceCommit: "1111111111111111111111111111111111111111"}, want: "valid image digest"},
		{name: "uid mismatch", result: TektonBuildStatusResult{UID: "other-uid", Status: "True", ImageDigest: checkBuildDigest}, want: "UID does not match"},
		{name: "commit mismatch", result: TektonBuildStatusResult{UID: "build-uid", Status: "True", SourceCommit: "2222222222222222222222222222222222222222", ImageDigest: checkBuildDigest}, want: "source commit does not match"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0069"}}
			runtimeStore := &sourceRuntimeStateStoreFake{current: checkBuildRuntimeState()}
			result := tc.result
			service := newCheckBuildService(store, runtimeStore, func(context.Context, domain.ChangeRequest, string, string) (TektonBuildStatusResult, error) {
				return result, nil
			})
			_, err := service.CheckBuild(context.Background(), store.change.ChangeNumber)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v want=%q", err, tc.want)
			}
			if store.markStepCalled {
				t.Fatal("MarkStep called after fail-closed")
			}
		})
	}
}

func TestCheckBuildRequiresPipelineRunName(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0070"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: domain.ChangeRuntimeState{}}
	service := newCheckBuildService(store, runtimeStore, func(context.Context, domain.ChangeRequest, string, string) (TektonBuildStatusResult, error) {
		t.Fatal("check build reached Tekton without a PipelineRun name")
		return TektonBuildStatusResult{}, nil
	})
	if _, err := service.CheckBuild(context.Background(), store.change.ChangeNumber); err == nil || !strings.Contains(err.Error(), "PipelineRun name") {
		t.Fatalf("err=%v", err)
	}
}
