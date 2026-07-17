//go:build integration

package database_test

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/vincmarz/devops-control-plane/internal/database"
	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping PostgreSQL integration test")
	}

	ctx := context.Background()

	db, err := database.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(db.Close)

	if _, err := db.Pool.Exec(ctx, `DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatalf("reset public schema: %v", err)
	}

	migrationNames := []string{
		"000001_init.up.sql",
		"000002_change_runtime_states.up.sql",
	}
	for _, migrationName := range migrationNames {
		migration, err := filepath.Abs(filepath.Join("..", "..", "migrations", migrationName))
		if err != nil {
			t.Fatalf("resolve migration path: %v", err)
		}
		if _, err := os.Stat(migration); err != nil {
			t.Fatalf("migration file not found at %s: %v", migration, err)
		}
		if err := db.ExecSQLFile(ctx, migration); err != nil {
			t.Fatalf("apply migration %s: %v", migration, err)
		}
	}

	return db
}

func TestChangeRequestLifecycle_Integration(t *testing.T) {
	db := setupTestDB(t)
	repo := database.NewChangeRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, domain.CreateChangeRequest{
		Title:             "Bump payment-service to v1.4.2",
		ApplicationName:   "payment-service",
		TargetEnvironment: "dev",
		ChangeType:        "image-update",
		RiskLevel:         "medium",
		RequestedBy:       "alice",
		Description:       "Routine version bump",
		Payload:           map[string]any{"image": "payment-service:v1.4.2", "replicas": float64(3)},
	})
	if err != nil {
		t.Fatalf("create change: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected a generated ID, got empty string")
	}
	if created.Status != domain.ChangeStatusDraft {
		t.Fatalf("new change: expected status %q, got %q", domain.ChangeStatusDraft, created.Status)
	}
	wantPrefix := "CHG-" + strconv.Itoa(time.Now().UTC().Year()) + "-"
	if !strings.HasPrefix(created.ChangeNumber, wantPrefix) {
		t.Fatalf("unexpected change number %q, want prefix %q", created.ChangeNumber, wantPrefix)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Payload["image"] != "payment-service:v1.4.2" {
		t.Fatalf("payload did not round-trip through jsonb: %#v", got.Payload)
	}
	if _, err := repo.Get(ctx, created.ChangeNumber); err != nil {
		t.Fatalf("get by change number %q: %v", created.ChangeNumber, err)
	}

	steps := []struct {
		action     string
		wantStatus string
	}{
		{"submit", domain.ChangeStatusSubmitted},
		{"approve", domain.ChangeStatusApproved},
		{"start-execution", domain.ChangeStatusExecuting},
		{"complete-execution", domain.ChangeStatusExecuted},
		{"close", domain.ChangeStatusClosed},
	}
	for _, s := range steps {
		if _, err := repo.TransitionLifecycle(ctx, created.ID, s.action, "operator", "via integration test"); err != nil {
			t.Fatalf("transition %q: %v", s.action, err)
		}
		cur, err := repo.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("get after %q: %v", s.action, err)
		}
		if cur.Status != s.wantStatus {
			t.Fatalf("after %q: expected status %q, got %q", s.action, s.wantStatus, cur.Status)
		}
	}

	final, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("final get: %v", err)
	}
	if final.SubmittedAt == nil {
		t.Error("SubmittedAt not set after submit")
	}
	if final.ApprovedAt == nil {
		t.Error("ApprovedAt not set after approve")
	}
	if final.ExecutedAt == nil {
		t.Error("ExecutedAt not set after complete-execution")
	}
	if final.ClosedAt == nil {
		t.Error("ClosedAt not set after close")
	}

	events, err := repo.Events(ctx, created.ID)
	if err != nil {
		t.Fatalf("events: %v", err)
	}
	if len(events) < len(steps) {
		t.Fatalf("expected at least %d audit events, got %d", len(steps), len(events))
	}
}

func TestInvalidLifecycleTransition_Integration(t *testing.T) {
	db := setupTestDB(t)
	repo := database.NewChangeRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, domain.CreateChangeRequest{
		Title:             "Config tweak",
		ApplicationName:   "billing",
		TargetEnvironment: "dev",
		ChangeType:        "config",
		RiskLevel:         "low",
		RequestedBy:       "bob",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := repo.TransitionLifecycle(ctx, created.ID, "approve", "bob", ""); err == nil {
		t.Fatal("expected an error approving a draft change, got nil")
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != domain.ChangeStatusDraft {
		t.Fatalf("status changed after a rejected transition: got %q, want %q", got.Status, domain.ChangeStatusDraft)
	}
}
