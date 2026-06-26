package app

import (
	"context"
	"errors"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type lifecycleTransitionFakeStore struct {
	called bool

	gotID      string
	gotAction  string
	gotActor   string
	gotMessage string

	result map[string]any
	err    error
}

func (f *lifecycleTransitionFakeStore) List(ctx context.Context) ([]domain.ChangeRequest, error) {
	return nil, nil
}

func (f *lifecycleTransitionFakeStore) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, nil
}

func (f *lifecycleTransitionFakeStore) Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error) {
	return domain.ChangeRequest{}, nil
}

func (f *lifecycleTransitionFakeStore) Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error) {
	return nil, nil
}

func (f *lifecycleTransitionFakeStore) TransitionLifecycle(ctx context.Context, idOrNumber string, action string, actor string, message string) (map[string]any, error) {
	f.called = true
	f.gotID = idOrNumber
	f.gotAction = action
	f.gotActor = actor
	f.gotMessage = message
	return f.result, f.err
}

func (f *lifecycleTransitionFakeStore) MarkStep(ctx context.Context, idOrNumber string, status string) (map[string]any, error) {
	return nil, nil
}

func TestChangeServiceTransitionLifecycleRequiresActor(t *testing.T) {
	store := &lifecycleTransitionFakeStore{}
	service := NewChangeService(store)

	_, err := service.TransitionLifecycle(context.Background(), "CHG-2026-0001", "submit", "", "missing actor")
	if err == nil {
		t.Fatal("TransitionLifecycle returned nil error, want actor validation error")
	}
	if store.called {
		t.Fatal("store was called even if actor validation failed")
	}
}

func TestChangeServiceTransitionLifecycleRequiresAction(t *testing.T) {
	store := &lifecycleTransitionFakeStore{}
	service := NewChangeService(store)

	_, err := service.TransitionLifecycle(context.Background(), "CHG-2026-0001", "", "vincenzo.marzario", "missing action")
	if err == nil {
		t.Fatal("TransitionLifecycle returned nil error, want action validation error")
	}
	if store.called {
		t.Fatal("store was called even if action validation failed")
	}
}

func TestChangeServiceTransitionLifecycleRequiresChangeIDOrNumber(t *testing.T) {
	store := &lifecycleTransitionFakeStore{}
	service := NewChangeService(store)

	_, err := service.TransitionLifecycle(context.Background(), "", "submit", "vincenzo.marzario", "missing id")
	if err == nil {
		t.Fatal("TransitionLifecycle returned nil error, want id validation error")
	}
	if store.called {
		t.Fatal("store was called even if id validation failed")
	}
}

func TestChangeServiceTransitionLifecycleTrimsAndDelegates(t *testing.T) {
	store := &lifecycleTransitionFakeStore{
		result: map[string]any{"status": domain.ChangeStatusSubmitted},
	}
	service := NewChangeService(store)

	result, err := service.TransitionLifecycle(
		context.Background(),
		"  CHG-2026-0001  ",
		"  submit  ",
		"  vincenzo.marzario  ",
		"  submit della change  ",
	)
	if err != nil {
		t.Fatalf("TransitionLifecycle returned unexpected error: %v", err)
	}
	if !store.called {
		t.Fatal("store was not called")
	}
	if store.gotID != "CHG-2026-0001" {
		t.Fatalf("id = %q, want CHG-2026-0001", store.gotID)
	}
	if store.gotAction != "submit" {
		t.Fatalf("action = %q, want submit", store.gotAction)
	}
	if store.gotActor != "vincenzo.marzario" {
		t.Fatalf("actor = %q, want vincenzo.marzario", store.gotActor)
	}
	if store.gotMessage != "submit della change" {
		t.Fatalf("message = %q, want trimmed message", store.gotMessage)
	}
	if result["status"] != domain.ChangeStatusSubmitted {
		t.Fatalf("result status = %v, want %q", result["status"], domain.ChangeStatusSubmitted)
	}
}

func TestChangeServiceTransitionLifecyclePropagatesStoreError(t *testing.T) {
	wantErr := errors.New("store transition failed")
	store := &lifecycleTransitionFakeStore{err: wantErr}
	service := NewChangeService(store)

	_, err := service.TransitionLifecycle(context.Background(), "CHG-2026-0001", "submit", "vincenzo.marzario", "submit")
	if !errors.Is(err, wantErr) {
		t.Fatalf("TransitionLifecycle error = %v, want %v", err, wantErr)
	}
	if !store.called {
		t.Fatal("store was not called")
	}
}
