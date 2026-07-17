//go:build integration

package database_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/vincmarz/devops-control-plane/internal/database"
	"github.com/vincmarz/devops-control-plane/internal/domain"
)

func TestChangeRuntimeStateRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	changeRepo := database.NewChangeRepository(db)
	runtimeRepo := database.NewChangeRuntimeStateRepository(db)
	ctx := context.Background()

	change, err := changeRepo.Create(ctx, domain.CreateChangeRequest{
		Title:             "Provider-aware runtime persistence",
		ApplicationName:   "demo-go-color-app",
		TargetEnvironment: "dev",
		ChangeType:        "config",
		RiskLevel:         "low",
		RequestedBy:       "integration-test",
	})
	if err != nil {
		t.Fatalf("create change: %v", err)
	}

	empty, err := runtimeRepo.Get(ctx, change.ChangeNumber)
	if err != nil {
		t.Fatalf("get empty runtime state: %v", err)
	}
	if empty.ChangeRequestID != change.ID {
		t.Fatalf("empty state change ID = %q, want %q", empty.ChangeRequestID, change.ID)
	}
	if empty.Source.Provider != "" || empty.GitOps.Revision != "" || empty.Tekton.Status != "" || empty.ArgoCD.SyncStatus != "" {
		t.Fatalf("expected empty runtime state, got %#v", empty)
	}

	source := domain.SourceRuntimeState{
		Provider: "github", ProviderRef: "github-public",
		ProjectPath: "vincmarz/source", RepositoryURL: "https://github.com/vincmarz/source.git",
		DefaultBranch: "main", Branch: "change/CHG-1", CommitSHA: "source-sha",
	}
	if err := runtimeRepo.UpsertSource(ctx, change.ID, source); err != nil {
		t.Fatalf("upsert source: %v", err)
	}
	first, err := runtimeRepo.Get(ctx, change.ID)
	if err != nil {
		t.Fatalf("get after source upsert: %v", err)
	}
	if first.Source.Branch != source.Branch || first.Source.ProviderRef != source.ProviderRef {
		t.Fatalf("source state = %#v", first.Source)
	}
	if first.CreatedAt.IsZero() || first.UpdatedAt.IsZero() {
		t.Fatalf("runtime timestamps were not populated: %#v", first)
	}

	time.Sleep(10 * time.Millisecond)
	gitops := domain.GitOpsRuntimeState{
		Provider: "github", ProviderRef: "github-public",
		ProjectPath: "vincmarz/demo-app-gitops", RepositoryURL: "https://github.com/vincmarz/demo-app-gitops.git",
		DefaultBranch: "main", Revision: "main", CommitSHA: "gitops-sha",
	}
	if err := runtimeRepo.UpsertGitOps(ctx, change.ChangeNumber, gitops); err != nil {
		t.Fatalf("upsert GitOps: %v", err)
	}

	tekton := domain.TektonRuntimeState{
		Namespace: "devops-ci-demo", PipelineName: "validate-gitops",
		PipelineRunName: "devops-cp-validate-chg-1", GitURL: gitops.RepositoryURL,
		GitRevision: "main", ValidationPath: "apps/demo-go-color-app", Status: "Succeeded", Reason: "Succeeded",
	}
	if err := runtimeRepo.UpsertTekton(ctx, change.ID, tekton); err != nil {
		t.Fatalf("upsert Tekton: %v", err)
	}

	argocd := domain.ArgoCDRuntimeState{
		ApplicationName: "demo-go-color-app", Provider: "github", ProviderRef: "github-public",
		DeclaredRepositoryURL: gitops.RepositoryURL, ObservedRepositoryURL: gitops.RepositoryURL,
		DeclaredDefaultBranch: "main", ObservedTargetRevision: "main", ObservedRevision: "argocd-sha",
		SyncStatus: "Synced", HealthStatus: "Healthy", CorrelationStatus: "Matched",
	}
	if err := runtimeRepo.UpsertArgoCD(ctx, change.ID, argocd); err != nil {
		t.Fatalf("upsert Argo CD: %v", err)
	}

	runtime := domain.RuntimeObservationState{
		ClusterName: "ocp-dev", Namespace: "devops-ci-demo",
		ResourceKind: "Deployment", ResourceName: "demo-go-color-app", Status: "Ready",
	}
	if err := runtimeRepo.UpsertRuntime(ctx, change.ID, runtime); err != nil {
		t.Fatalf("upsert runtime: %v", err)
	}

	got, err := runtimeRepo.Get(ctx, change.ChangeNumber)
	if err != nil {
		t.Fatalf("get complete runtime state: %v", err)
	}
	if got.Source.Branch != source.Branch {
		t.Fatalf("source state was overwritten: %#v", got.Source)
	}
	if got.GitOps.Revision != "main" || got.GitOps.CommitSHA != "gitops-sha" {
		t.Fatalf("GitOps state = %#v", got.GitOps)
	}
	if got.Tekton.PipelineRunName != tekton.PipelineRunName || got.Tekton.Status != "Succeeded" {
		t.Fatalf("Tekton state = %#v", got.Tekton)
	}
	if got.ArgoCD.CorrelationStatus != "Matched" || got.ArgoCD.ObservedRevision != "argocd-sha" {
		t.Fatalf("Argo CD state = %#v", got.ArgoCD)
	}
	if got.Runtime.ResourceName != runtime.ResourceName || got.Runtime.Status != "Ready" {
		t.Fatalf("runtime observation = %#v", got.Runtime)
	}
	if !got.UpdatedAt.After(first.UpdatedAt) {
		t.Fatalf("updatedAt did not advance: first=%s current=%s", first.UpdatedAt, got.UpdatedAt)
	}

	missingID := "00000000-0000-0000-0000-000000000000"
	if _, err := runtimeRepo.Get(ctx, missingID); err == nil || !strings.Contains(err.Error(), "change not found") {
		t.Fatalf("get missing change error = %v", err)
	}
	if err := runtimeRepo.UpsertSource(ctx, missingID, source); err == nil || !strings.Contains(err.Error(), "change not found") {
		t.Fatalf("upsert missing change error = %v", err)
	}

	if _, err := db.Pool.Exec(ctx, `DELETE FROM change_requests WHERE id = $1::uuid`, change.ID); err != nil {
		t.Fatalf("delete change: %v", err)
	}
	var count int
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM change_runtime_states WHERE change_request_id = $1::uuid`, change.ID).Scan(&count); err != nil {
		t.Fatalf("count runtime states after cascade: %v", err)
	}
	if count != 0 {
		t.Fatalf("runtime state rows after cascade = %d, want 0", count)
	}
}
