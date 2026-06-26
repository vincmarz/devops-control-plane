package app

import (
	"context"
	"errors"
	"github.com/vincmarz/devops-control-plane/internal/domain"
	"testing"
)

type validateFakeStore struct {
	change         domain.ChangeRequest
	markStepCalled bool
	markedStatus   string
}

func (f *validateFakeStore) List(ctx context.Context) ([]domain.ChangeRequest, error) {
	return nil, nil
}
func (f *validateFakeStore) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, nil
}
func (f *validateFakeStore) Get(ctx context.Context, id string) (domain.ChangeRequest, error) {
	if f.change.ChangeNumber == "" {
		return domain.ChangeRequest{}, errors.New("not found")
	}
	return f.change, nil
}
func (f *validateFakeStore) Events(ctx context.Context, id string) ([]domain.ChangeEvent, error) {
	return nil, nil
}
func (f *validateFakeStore) TransitionLifecycle(ctx context.Context, id, a, actor, msg string) (map[string]any, error) {
	return nil, nil
}
func (f *validateFakeStore) MarkStep(ctx context.Context, id, status string) (map[string]any, error) {
	f.markStepCalled = true
	f.markedStatus = status
	return map[string]any{"status": status}, nil
}
func TestChangeServiceValidate(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-2026-0005", ApplicationName: "demo-go-color-app"}}
	var got string
	svc := NewChangeService(store, WithTektonRunPipeline(func(ctx context.Context, ch domain.ChangeRequest) (string, string, string, error) {
		got = ch.ChangeNumber
		return "devops-cp-validate-chg-2026-0005-abcde", "devops-ci-demo", "uid-123", nil
	}))
	r, err := svc.Validate(context.Background(), "CHG-2026-0005")
	if err != nil {
		t.Fatal(err)
	}
	if got != "CHG-2026-0005" {
		t.Fatalf("got %s", got)
	}
	if store.markedStatus != "ValidationRunning" {
		t.Fatalf("status %s", store.markedStatus)
	}
	if r["tekton"] == nil {
		t.Fatalf("missing tekton info")
	}
}
func TestChangeServiceValidatePropagatesTektonError(t *testing.T) {
	store := &validateFakeStore{change: domain.ChangeRequest{ChangeNumber: "CHG-2026-0005"}}
	svc := NewChangeService(store, WithTektonRunPipeline(func(ctx context.Context, ch domain.ChangeRequest) (string, string, string, error) {
		return "", "", "", errors.New("tekton failed")
	}))
	_, err := svc.Validate(context.Background(), "CHG-2026-0005")
	if err == nil {
		t.Fatal("want error")
	}
	if store.markStepCalled {
		t.Fatal("MarkStep called")
	}
}
