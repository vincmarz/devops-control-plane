package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type createBranchFakeStore struct {
	change domain.ChangeRequest

	getCalled      bool
	markStepCalled bool
	markedID       string
	markedStatus   string
}

func (f *createBranchFakeStore) List(ctx context.Context) ([]domain.ChangeRequest, error) {
	return nil, nil
}
func (f *createBranchFakeStore) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, nil
}
func (f *createBranchFakeStore) Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error) {
	f.getCalled = true
	if f.change.ChangeNumber == "" {
		return domain.ChangeRequest{}, errors.New("not found")
	}
	return f.change, nil
}
func (f *createBranchFakeStore) Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error) {
	return nil, nil
}
func (f *createBranchFakeStore) TransitionLifecycle(ctx context.Context, idOrNumber string, action string, actor string, message string) (map[string]any, error) {
	return nil, nil
}
func (f *createBranchFakeStore) MarkStep(ctx context.Context, idOrNumber string, status string) (map[string]any, error) {
	f.markStepCalled = true
	f.markedID = idOrNumber
	f.markedStatus = status
	return map[string]any{"status": status}, nil
}

func TestChangeServiceCreateBranch(t *testing.T) {
	store := &createBranchFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-2026-0003"}}

	var gotProjectID int
	var gotBranch string
	var gotRef string
	service := NewChangeService(store, WithGitCreateBranch(func(ctx context.Context, projectID int, branch string, ref string) error {
		gotProjectID = projectID
		gotBranch = branch
		gotRef = ref
		return nil
	}, 1, "main"))

	result, err := service.CreateBranch(context.Background(), "CHG-2026-0003")
	if err != nil {
		t.Fatalf("CreateBranch returned error: %v", err)
	}
	if !store.getCalled {
		t.Fatal("store.Get was not called")
	}
	if !store.markStepCalled {
		t.Fatal("store.MarkStep was not called")
	}
	if store.markedStatus != "BranchCreated" {
		t.Fatalf("marked status = %q, want BranchCreated", store.markedStatus)
	}
	if gotProjectID != 1 {
		t.Fatalf("projectID = %d, want 1", gotProjectID)
	}
	if gotBranch != "change/CHG-2026-0003" {
		t.Fatalf("branch = %q, want change/CHG-2026-0003", gotBranch)
	}
	if gotRef != "main" {
		t.Fatalf("ref = %q, want main", gotRef)
	}
	gitInfo, ok := result["git"].(map[string]any)
	if !ok {
		t.Fatalf("result git info missing or invalid: %#v", result["git"])
	}
	if gitInfo["branch"] != "change/CHG-2026-0003" {
		t.Fatalf("git branch = %v", gitInfo["branch"])
	}
}

func TestChangeServiceCreateBranchRequiresGitClient(t *testing.T) {
	store := &createBranchFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-2026-0003"}}
	service := NewChangeService(store)

	_, err := service.CreateBranch(context.Background(), "CHG-2026-0003")
	if err == nil {
		t.Fatal("CreateBranch returned nil error, want missing git client error")
	}
}

func TestChangeServiceCreateBranchPropagatesGitError(t *testing.T) {
	store := &createBranchFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-2026-0003"}}
	wantErr := errors.New("gitlab failed")
	service := NewChangeService(store, WithGitCreateBranch(func(ctx context.Context, projectID int, branch string, ref string) error {
		return wantErr
	}, 1, "main"))

	_, err := service.CreateBranch(context.Background(), "CHG-2026-0003")
	if err == nil {
		t.Fatal("CreateBranch returned nil error, want git error")
	}
	if store.markStepCalled {
		t.Fatal("store.MarkStep was called even if GitLab create branch failed")
	}
}

func TestChangeServiceUpdateFiles(t *testing.T) {
	store := &createBranchFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-2026-0003", ApplicationName: "demo-go-color-app", TargetEnvironment: "dev"}}

	var gotProjectID int
	var gotBranch string
	var gotFilePath string
	var gotCommitMessage string
	var gotContent string
	service := NewChangeService(
		store,
		WithGitCreateBranch(func(ctx context.Context, projectID int, branch string, ref string) error { return nil }, 1, "main"),
		WithGitCreateOrUpdateFile(func(ctx context.Context, projectID int, branch string, filePath string, commitMessage string, content string) error {
			gotProjectID = projectID
			gotBranch = branch
			gotFilePath = filePath
			gotCommitMessage = commitMessage
			gotContent = content
			return nil
		}),
	)

	result, err := service.UpdateFiles(context.Background(), "CHG-2026-0003")
	if err != nil {
		t.Fatalf("UpdateFiles returned error: %v", err)
	}
	if gotProjectID != 1 {
		t.Fatalf("projectID = %d, want 1", gotProjectID)
	}
	if gotBranch != "change/CHG-2026-0003" {
		t.Fatalf("branch = %q", gotBranch)
	}
	if gotFilePath != "manifests/chg-2026-0003-control-plane.yaml" {
		t.Fatalf("filePath = %q", gotFilePath)
	}
	if gotCommitMessage != "Add generated manifest for CHG-2026-0003" {
		t.Fatalf("commitMessage = %q", gotCommitMessage)
	}
	for _, expected := range []string{"changeNumber: CHG-2026-0003", "applicationName: demo-go-color-app", "targetEnvironment: dev", "managedBy: devops-control-plane"} {
		if !strings.Contains(gotContent, expected) {
			t.Fatalf("content does not contain %q: %s", expected, gotContent)
		}
	}
	if store.markedStatus != "CommitCreated" {
		t.Fatalf("marked status = %q, want CommitCreated", store.markedStatus)
	}
	gitInfo, ok := result["git"].(map[string]any)
	if !ok {
		t.Fatalf("result git info missing or invalid: %#v", result["git"])
	}
	if gitInfo["filePath"] != "manifests/chg-2026-0003-control-plane.yaml" {
		t.Fatalf("git filePath = %v", gitInfo["filePath"])
	}
}

func TestChangeServiceUpdateFilesPropagatesGitError(t *testing.T) {
	store := &createBranchFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-2026-0003"}}
	service := NewChangeService(
		store,
		WithGitCreateBranch(func(ctx context.Context, projectID int, branch string, ref string) error { return nil }, 1, "main"),
		WithGitCreateOrUpdateFile(func(ctx context.Context, projectID int, branch string, filePath string, commitMessage string, content string) error {
			return errors.New("gitlab file failed")
		}),
	)

	_, err := service.UpdateFiles(context.Background(), "CHG-2026-0003")
	if err == nil {
		t.Fatal("UpdateFiles returned nil error, want git error")
	}
	if store.markStepCalled {
		t.Fatal("store.MarkStep was called even if GitLab update file failed")
	}
}

func TestChangeServiceOpenMergeRequest(t *testing.T) {
	store := &createBranchFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-2026-0003", ApplicationName: "demo-go-color-app", TargetEnvironment: "dev"}}

	var gotProjectID int
	var gotSourceBranch string
	var gotTargetBranch string
	var gotTitle string
	var gotDescription string
	service := NewChangeService(
		store,
		WithGitCreateBranch(func(ctx context.Context, projectID int, branch string, ref string) error { return nil }, 1, "main"),
		WithGitCreateOrUpdateFile(func(ctx context.Context, projectID int, branch string, filePath string, commitMessage string, content string) error {
			return nil
		}),
		WithGitOpenMergeRequest(func(ctx context.Context, projectID int, sourceBranch string, targetBranch string, title string, description string) (int, string, error) {
			gotProjectID = projectID
			gotSourceBranch = sourceBranch
			gotTargetBranch = targetBranch
			gotTitle = title
			gotDescription = description
			return 1, "https://gitlab.example.local/group/project/-/merge_requests/1", nil
		}),
	)

	result, err := service.OpenMergeRequest(context.Background(), "CHG-2026-0003")
	if err != nil {
		t.Fatalf("OpenMergeRequest returned error: %v", err)
	}
	if gotProjectID != 1 {
		t.Fatalf("projectID = %d, want 1", gotProjectID)
	}
	if gotSourceBranch != "change/CHG-2026-0003" {
		t.Fatalf("sourceBranch = %q", gotSourceBranch)
	}
	if gotTargetBranch != "main" {
		t.Fatalf("targetBranch = %q", gotTargetBranch)
	}
	if gotTitle != "CHG-2026-0003 - GitOps change for demo-go-color-app" {
		t.Fatalf("title = %q", gotTitle)
	}
	if !strings.Contains(gotDescription, "CHG-2026-0003") {
		t.Fatalf("description does not contain change number: %q", gotDescription)
	}
	if store.markedStatus != "MergeRequestOpened" {
		t.Fatalf("marked status = %q, want MergeRequestOpened", store.markedStatus)
	}
	gitInfo, ok := result["git"].(map[string]any)
	if !ok {
		t.Fatalf("result git info missing or invalid: %#v", result["git"])
	}
	if gitInfo["mergeRequestIID"] != 1 {
		t.Fatalf("mergeRequestIID = %v", gitInfo["mergeRequestIID"])
	}
}

func TestChangeServiceOpenMergeRequestPropagatesGitError(t *testing.T) {
	store := &createBranchFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-2026-0003", ApplicationName: "demo-go-color-app"}}
	service := NewChangeService(
		store,
		WithGitCreateBranch(func(ctx context.Context, projectID int, branch string, ref string) error { return nil }, 1, "main"),
		WithGitOpenMergeRequest(func(ctx context.Context, projectID int, sourceBranch string, targetBranch string, title string, description string) (int, string, error) {
			return 0, "", errors.New("gitlab mr failed")
		}),
	)

	_, err := service.OpenMergeRequest(context.Background(), "CHG-2026-0003")
	if err == nil {
		t.Fatal("OpenMergeRequest returned nil error, want git error")
	}
	if store.markStepCalled {
		t.Fatal("store.MarkStep was called even if GitLab open merge request failed")
	}
}
