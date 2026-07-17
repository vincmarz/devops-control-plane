package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type ChangeRuntimeStateRepository struct {
	db *DB
}

func NewChangeRuntimeStateRepository(db *DB) *ChangeRuntimeStateRepository {
	return &ChangeRuntimeStateRepository{db: db}
}

func (r *ChangeRuntimeStateRepository) Get(ctx context.Context, idOrNumber string) (domain.ChangeRuntimeState, error) {
	idOrNumber = strings.TrimSpace(idOrNumber)
	if idOrNumber == "" {
		return domain.ChangeRuntimeState{}, errors.New("change id or number is required")
	}

	var state domain.ChangeRuntimeState
	var exists bool
	var rawSource, rawGitOps, rawTekton, rawArgoCD, rawRuntime []byte
	var createdAt, updatedAt any

	err := r.db.Pool.QueryRow(ctx, `
        SELECT
            cr.id::text,
            crs.change_request_id IS NOT NULL,
            COALESCE(crs.source_state, '{}'::jsonb),
            COALESCE(crs.gitops_state, '{}'::jsonb),
            COALESCE(crs.tekton_state, '{}'::jsonb),
            COALESCE(crs.argocd_state, '{}'::jsonb),
            COALESCE(crs.runtime_state, '{}'::jsonb),
            crs.created_at,
            crs.updated_at
        FROM change_requests cr
        LEFT JOIN change_runtime_states crs ON crs.change_request_id = cr.id
        WHERE cr.id::text = $1 OR cr.change_number = $1
    `, idOrNumber).Scan(
		&state.ChangeRequestID,
		&exists,
		&rawSource,
		&rawGitOps,
		&rawTekton,
		&rawArgoCD,
		&rawRuntime,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ChangeRuntimeState{}, fmt.Errorf("change not found: %s", idOrNumber)
		}
		return domain.ChangeRuntimeState{}, fmt.Errorf("get change runtime state: %w", err)
	}

	if err := decodeRuntimeStateSection(rawSource, &state.Source); err != nil {
		return domain.ChangeRuntimeState{}, fmt.Errorf("decode source runtime state: %w", err)
	}
	if err := decodeRuntimeStateSection(rawGitOps, &state.GitOps); err != nil {
		return domain.ChangeRuntimeState{}, fmt.Errorf("decode GitOps runtime state: %w", err)
	}
	if err := decodeRuntimeStateSection(rawTekton, &state.Tekton); err != nil {
		return domain.ChangeRuntimeState{}, fmt.Errorf("decode Tekton runtime state: %w", err)
	}
	if err := decodeRuntimeStateSection(rawArgoCD, &state.ArgoCD); err != nil {
		return domain.ChangeRuntimeState{}, fmt.Errorf("decode Argo CD runtime state: %w", err)
	}
	if err := decodeRuntimeStateSection(rawRuntime, &state.Runtime); err != nil {
		return domain.ChangeRuntimeState{}, fmt.Errorf("decode runtime observation state: %w", err)
	}

	if exists {
		if err := r.db.Pool.QueryRow(ctx, `
            SELECT created_at, updated_at
            FROM change_runtime_states
            WHERE change_request_id = $1::uuid
        `, state.ChangeRequestID).Scan(&state.CreatedAt, &state.UpdatedAt); err != nil {
			return domain.ChangeRuntimeState{}, fmt.Errorf("load change runtime state timestamps: %w", err)
		}
	}
	return state, nil
}

func (r *ChangeRuntimeStateRepository) UpsertSource(ctx context.Context, idOrNumber string, state domain.SourceRuntimeState) error {
	return r.upsertSection(ctx, idOrNumber, "source_state", state)
}

func (r *ChangeRuntimeStateRepository) UpsertGitOps(ctx context.Context, idOrNumber string, state domain.GitOpsRuntimeState) error {
	return r.upsertSection(ctx, idOrNumber, "gitops_state", state)
}

func (r *ChangeRuntimeStateRepository) UpsertTekton(ctx context.Context, idOrNumber string, state domain.TektonRuntimeState) error {
	return r.upsertSection(ctx, idOrNumber, "tekton_state", state)
}

func (r *ChangeRuntimeStateRepository) UpsertArgoCD(ctx context.Context, idOrNumber string, state domain.ArgoCDRuntimeState) error {
	return r.upsertSection(ctx, idOrNumber, "argocd_state", state)
}

func (r *ChangeRuntimeStateRepository) UpsertRuntime(ctx context.Context, idOrNumber string, state domain.RuntimeObservationState) error {
	return r.upsertSection(ctx, idOrNumber, "runtime_state", state)
}

func (r *ChangeRuntimeStateRepository) upsertSection(ctx context.Context, idOrNumber, column string, state any) error {
	idOrNumber = strings.TrimSpace(idOrNumber)
	if idOrNumber == "" {
		return errors.New("change id or number is required")
	}
	allowed := map[string]bool{
		"source_state":  true,
		"gitops_state":  true,
		"tekton_state":  true,
		"argocd_state":  true,
		"runtime_state": true,
	}
	if !allowed[column] {
		return fmt.Errorf("unsupported runtime state section %q", column)
	}

	rawState, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", column, err)
	}

	query := fmt.Sprintf(`
        INSERT INTO change_runtime_states (change_request_id, %s)
        SELECT id, $2::jsonb
        FROM change_requests
        WHERE id::text = $1 OR change_number = $1
        ON CONFLICT (change_request_id)
        DO UPDATE SET
            %s = EXCLUDED.%s,
            updated_at = now()
    `, column, column, column)

	commandTag, err := r.db.Pool.Exec(ctx, query, idOrNumber, string(rawState))
	if err != nil {
		return fmt.Errorf("upsert %s: %w", column, err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("change not found: %s", idOrNumber)
	}
	return nil
}

func decodeRuntimeStateSection(raw []byte, target any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, target)
}
