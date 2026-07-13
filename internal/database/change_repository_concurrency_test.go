//go:build integration

package database_test

import (
	"context"
	"sync"
	"testing"

	"github.com/vincmarz/devops-control-plane/internal/database"
	"github.com/vincmarz/devops-control-plane/internal/domain"
)

// TestConcurrentApproveOnlyOneWins repeatedly fires concurrent approve
// transitions against a freshly submitted change. The lifecycle contract
// requires that exactly one wins: submitted -> approved may happen only once,
// producing exactly one approval event.
//
// A single concurrent burst is unreliable because the read-modify-write window
// is tiny on localhost, so the scenario is repeated many times to make any
// atomicity defect surface. Against a correct (row-locked) implementation this
// passes deterministically; against an unguarded read-modify-write it will,
// across enough iterations, observe more than one winner.
//
// Relies on setupTestDB from change_repository_integration_test.go; runs only
// under the `integration` build tag against a disposable PostgreSQL.
func TestConcurrentApproveOnlyOneWins(t *testing.T) {
	db := setupTestDB(t)
	repo := database.NewChangeRepository(db)
	ctx := context.Background()

	const iterations = 300
	const workers = 8

	for iter := 0; iter < iterations; iter++ {
		created, err := repo.Create(ctx, domain.CreateChangeRequest{
			Title:             "Concurrent approval",
			ApplicationName:   "app",
			TargetEnvironment: "dev",
			ChangeType:        "config",
			RiskLevel:         "low",
			RequestedBy:       "alice",
		})
		if err != nil {
			t.Fatalf("iter %d: create: %v", iter, err)
		}
		if _, err := repo.TransitionLifecycle(ctx, created.ID, "submit", "alice", ""); err != nil {
			t.Fatalf("iter %d: submit: %v", iter, err)
		}

		var wg sync.WaitGroup
		start := make(chan struct{})
		results := make(chan error, workers)
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start // release together to maximize contention
				_, err := repo.TransitionLifecycle(ctx, created.ID, "approve", "approver", "")
				results <- err
			}()
		}
		close(start)
		wg.Wait()
		close(results)

		successes := 0
		for err := range results {
			if err == nil {
				successes++
			}
		}
		if successes != 1 {
			t.Fatalf("iter %d: expected exactly 1 successful approve, got %d — concurrent lifecycle transitions are not atomic", iter, successes)
		}

		final, err := repo.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("iter %d: get: %v", iter, err)
		}
		if final.Status != domain.ChangeStatusApproved {
			t.Fatalf("iter %d: expected final status %q, got %q", iter, domain.ChangeStatusApproved, final.Status)
		}

		events, err := repo.Events(ctx, created.ID)
		if err != nil {
			t.Fatalf("iter %d: events: %v", iter, err)
		}
		approveEvents := 0
		for _, e := range events {
			if e.EventType == domain.ChangeEventApproved {
				approveEvents++
			}
		}
		if approveEvents != 1 {
			t.Fatalf("iter %d: expected exactly 1 approval event, got %d — duplicate approvals recorded", iter, approveEvents)
		}
	}
}
