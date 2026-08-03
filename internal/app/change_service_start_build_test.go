package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

const startBuildImage = "image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app:latest"

func startBuildBinding() RepositoryBinding {
	return RepositoryBinding{
		Provider:        "github",
		ProviderRef:     "github-public",
		Role:            RepositoryRoleSource,
		ProjectPath:     "org/app",
		RepositoryURL:   "https://github.com/org/app.git",
		DefaultBranch:   "main",
		WorkflowEnabled: true,
	}
}

func startBuildSource() domain.SourceRuntimeState {
	return domain.SourceRuntimeState{
		Provider:       "github",
		ProviderRef:    "github-public",
		RepositoryURL:  "https://github.com/org/app.git",
		CommitSHA:      "1111111111111111111111111111111111111111",
		ProposalState:  "merged",
		MergeCommitSHA: "2222222222222222222222222222222222222222",
	}
}

func newStartBuildService(store *validateFakeStore, runtimeStore *sourceRuntimeStateStoreFake, start TektonStartBuildFunc) *ChangeService {
	return NewChangeService(
		store,
		WithGitSourceBindingResolverFunc(func(string) (RepositoryBinding, error) { return startBuildBinding(), nil }),
		WithChangeRuntimeStateStore(runtimeStore),
		WithTektonStartBuild(start, startBuildImage),
	)
}

func successfulStartBuild(_ context.Context, _ domain.ChangeRequest, _ TektonStartBuildRequest) (TektonBuildRunResult, error) {
	return TektonBuildRunResult{
		Namespace:       "devops-ci-demo",
		PipelineName:    "go-build-and-push",
		PipelineRunName: "devops-cp-build-chg-2026-0059-abcde",
		PipelineRunUID:  "uid-build",
	}, nil
}

func TestStartBuildUsesMergedCommitAndPersistsArtifactBeforeMarkStep(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0059", ApplicationName: "demo-go-color-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: domain.ChangeRuntimeState{Source: startBuildSource()}}
	var request TektonStartBuildRequest
	service := newStartBuildService(store, runtimeStore, func(ctx context.Context, change domain.ChangeRequest, got TektonStartBuildRequest) (TektonBuildRunResult, error) {
		request = got
		return successfulStartBuild(ctx, change, got)
	})

	result, err := service.StartBuild(context.Background(), store.change.ChangeNumber)
	if err != nil {
		t.Fatal(err)
	}
	if request.GitRevision != startBuildSource().MergeCommitSHA {
		t.Fatalf("GitRevision = %q", request.GitRevision)
	}
	if request.GitURL != startBuildBinding().RepositoryURL {
		t.Fatalf("GitURL = %q", request.GitURL)
	}
	wantImage := "image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app:chg-2026-0059-2222222"
	if request.Image != wantImage {
		t.Fatalf("Image = %q, want %q", request.Image, wantImage)
	}
	if !runtimeStore.artifactCalled {
		t.Fatal("artifact state was not persisted")
	}
	artifact := runtimeStore.artifact
	if artifact.PipelineRunName != "devops-cp-build-chg-2026-0059-abcde" || artifact.PipelineRunUID != "uid-build" {
		t.Fatalf("artifact PipelineRun = %#v", artifact)
	}
	if artifact.SourceCommitSHA != startBuildSource().MergeCommitSHA || artifact.SourceRevision != startBuildSource().MergeCommitSHA {
		t.Fatalf("artifact source = %#v", artifact)
	}
	if artifact.ImageRepository != "image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app" || artifact.ImageTag != "chg-2026-0059-2222222" {
		t.Fatalf("artifact image = %#v", artifact)
	}
	if artifact.Status != "Unknown" || artifact.Reason != "Running" || artifact.ImageDigest != "" {
		t.Fatalf("artifact status = %#v", artifact)
	}
	if !store.markStepCalled || store.markedStatus != "BuildRunning" {
		t.Fatalf("MarkStep = called:%v status:%q", store.markStepCalled, store.markedStatus)
	}
	if result["artifact"] == nil {
		t.Fatal("artifact result is missing")
	}
}

func TestStartBuildFallsBackToSourceCommit(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0060", ApplicationName: "demo-go-color-app"}}
	source := startBuildSource()
	source.ProposalState = "open"
	source.MergeCommitSHA = ""
	runtimeStore := &sourceRuntimeStateStoreFake{current: domain.ChangeRuntimeState{Source: source}}
	var request TektonStartBuildRequest
	service := newStartBuildService(store, runtimeStore, func(ctx context.Context, change domain.ChangeRequest, got TektonStartBuildRequest) (TektonBuildRunResult, error) {
		request = got
		return successfulStartBuild(ctx, change, got)
	})

	if _, err := service.StartBuild(context.Background(), store.change.ChangeNumber); err != nil {
		t.Fatal(err)
	}
	if request.GitRevision != source.CommitSHA {
		t.Fatalf("GitRevision = %q", request.GitRevision)
	}
}

func TestStartBuildFailsClosedBeforeTekton(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*domain.SourceRuntimeState)
		want   string
	}{
		{name: "missing commit", mutate: func(source *domain.SourceRuntimeState) { source.CommitSHA = ""; source.MergeCommitSHA = "" }, want: "immutable commit SHA"},
		{name: "provider mismatch", mutate: func(source *domain.SourceRuntimeState) { source.Provider = "gitlab" }, want: "provider does not match"},
		{name: "provider ref mismatch", mutate: func(source *domain.SourceRuntimeState) { source.ProviderRef = "other" }, want: "providerRef does not match"},
		{name: "repository mismatch", mutate: func(source *domain.SourceRuntimeState) { source.RepositoryURL = "https://github.com/other/app.git" }, want: "repositoryURL does not match"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0061", ApplicationName: "demo-go-color-app"}}
			source := startBuildSource()
			tc.mutate(&source)
			runtimeStore := &sourceRuntimeStateStoreFake{current: domain.ChangeRuntimeState{Source: source}}
			tektonCalled := false
			service := newStartBuildService(store, runtimeStore, func(context.Context, domain.ChangeRequest, TektonStartBuildRequest) (TektonBuildRunResult, error) {
				tektonCalled = true
				return TektonBuildRunResult{}, nil
			})

			_, err := service.StartBuild(context.Background(), store.change.ChangeNumber)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
			if tektonCalled || runtimeStore.artifactCalled || store.markStepCalled {
				t.Fatalf("side effects: tekton=%v artifact=%v mark=%v", tektonCalled, runtimeStore.artifactCalled, store.markStepCalled)
			}
		})
	}
}

func TestStartBuildRequiresRuntimeStateStoreAndBindingResolver(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0062", ApplicationName: "demo-go-color-app"}}
	service := NewChangeService(store, WithTektonStartBuild(successfulStartBuild, startBuildImage))
	if _, err := service.StartBuild(context.Background(), store.change.ChangeNumber); err == nil || !strings.Contains(err.Error(), "runtime state store") {
		t.Fatalf("runtime store error = %v", err)
	}

	service = NewChangeService(store, WithChangeRuntimeStateStore(&sourceRuntimeStateStoreFake{}), WithTektonStartBuild(successfulStartBuild, startBuildImage))
	if _, err := service.StartBuild(context.Background(), store.change.ChangeNumber); err == nil || !strings.Contains(err.Error(), "source binding resolver") {
		t.Fatalf("binding resolver error = %v", err)
	}
}

func TestStartBuildPersistenceFailureIncludesPipelineRunIdentity(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ID: "change-id", ChangeNumber: "CHG-2026-0063", ApplicationName: "demo-go-color-app"}}
	runtimeStore := &sourceRuntimeStateStoreFake{current: domain.ChangeRuntimeState{Source: startBuildSource()}, artifactErr: errors.New("database unavailable")}
	service := newStartBuildService(store, runtimeStore, successfulStartBuild)

	_, err := service.StartBuild(context.Background(), store.change.ChangeNumber)
	if err == nil || !strings.Contains(err.Error(), "devops-cp-build-chg-2026-0059-abcde") || !strings.Contains(err.Error(), "uid-build") {
		t.Fatalf("error = %v", err)
	}
	if store.markStepCalled {
		t.Fatal("MarkStep was called after artifact persistence failure")
	}
}

func TestDeterministicBuildImageTagAndRepositoryParsing(t *testing.T) {
	tag, err := deterministicBuildImageTag("CHG-2026-0064", "abcdef0123456789")
	if err != nil {
		t.Fatal(err)
	}
	if tag != "chg-2026-0064-abcdef0" {
		t.Fatalf("tag = %q", tag)
	}
	repository, err := splitImageRepository(startBuildImage)
	if err != nil {
		t.Fatal(err)
	}
	if repository != "image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app" {
		t.Fatalf("repository = %q", repository)
	}
}
