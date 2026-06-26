package app

import (
	"context"
	"errors"
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
