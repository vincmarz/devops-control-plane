package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type ChangeRepository struct {
	db *DB
}

func NewChangeRepository(db *DB) *ChangeRepository {
	return &ChangeRepository{db: db}
}

func (r *ChangeRepository) List(ctx context.Context) ([]domain.ChangeRequest, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			id::text,
			change_number,
			application_name,
			change_type,
			status,
			COALESCE(requested_by, ''),
			COALESCE(description, ''),
			COALESCE(request_payload, '{}'::jsonb),
			created_at,
			updated_at,
			completed_at
		FROM change_requests
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, fmt.Errorf("list change_requests: %w", err)
	}
	defer rows.Close()

	changes := make([]domain.ChangeRequest, 0)
	for rows.Next() {
		change, err := scanChange(rows)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate change_requests: %w", err)
	}
	return changes, nil
}

func (r *ChangeRepository) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	if req.ApplicationName == "" || req.ChangeType == "" {
		return domain.ChangeRequest{}, errors.New("applicationName and changeType are required")
	}

	payload, err := json.Marshal(req.Payload)
	if err != nil {
		return domain.ChangeRequest{}, fmt.Errorf("marshal request payload: %w", err)
	}

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return domain.ChangeRequest{}, fmt.Errorf("begin create change transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	year := time.Now().UTC().Year()
	var nextNumber int
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(split_part(change_number, '-', 3)::int), 0) + 1
		FROM change_requests
		WHERE change_number LIKE $1
	`, fmt.Sprintf("CHG-%d-%%", year)).Scan(&nextNumber); err != nil {
		return domain.ChangeRequest{}, fmt.Errorf("generate change number: %w", err)
	}

	changeNumber := fmt.Sprintf("CHG-%d-%04d", year, nextNumber)

	var change domain.ChangeRequest
	var rawPayload []byte
	if err := tx.QueryRow(ctx, `
		INSERT INTO change_requests (
			change_number,
			application_name,
			change_type,
			status,
			requested_by,
			description,
			request_payload
		)
		VALUES ($1, $2, $3, 'Created', $4, $5, $6::jsonb)
		RETURNING
			id::text,
			change_number,
			application_name,
			change_type,
			status,
			COALESCE(requested_by, ''),
			COALESCE(description, ''),
			COALESCE(request_payload, '{}'::jsonb),
			created_at,
			updated_at,
			completed_at
	`, changeNumber, req.ApplicationName, req.ChangeType, req.RequestedBy, req.Description, string(payload)).Scan(
		&change.ID,
		&change.ChangeNumber,
		&change.ApplicationName,
		&change.ChangeType,
		&change.Status,
		&change.RequestedBy,
		&change.Description,
		&rawPayload,
		&change.CreatedAt,
		&change.UpdatedAt,
		&change.CompletedAt,
	); err != nil {
		return domain.ChangeRequest{}, fmt.Errorf("insert change_request: %w", err)
	}
	change.Payload = decodeMap(rawPayload)

	if _, err := tx.Exec(ctx, `
		INSERT INTO change_events (
			change_request_id,
			event_type,
			new_status,
			message,
			source,
			payload
		)
		VALUES ($1::uuid, 'Created', 'Created', 'ChangeRequest created', 'workflow', '{}'::jsonb)
	`, change.ID); err != nil {
		return domain.ChangeRequest{}, fmt.Errorf("insert created event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.ChangeRequest{}, fmt.Errorf("commit create change transaction: %w", err)
	}

	return change, nil
}

func (r *ChangeRepository) Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT
			id::text,
			change_number,
			application_name,
			change_type,
			status,
			COALESCE(requested_by, ''),
			COALESCE(description, ''),
			COALESCE(request_payload, '{}'::jsonb),
			created_at,
			updated_at,
			completed_at
		FROM change_requests
		WHERE id::text = $1 OR change_number = $1
	`, idOrNumber)

	change, err := scanChange(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ChangeRequest{}, fmt.Errorf("change not found: %s", idOrNumber)
		}
		return domain.ChangeRequest{}, err
	}
	return change, nil
}

func (r *ChangeRepository) Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error) {
	change, err := r.Get(ctx, idOrNumber)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			event_type,
			COALESCE(previous_status, ''),
			COALESCE(new_status, ''),
			COALESCE(message, ''),
			COALESCE(error_code, ''),
			COALESCE(source, ''),
			COALESCE(payload, '{}'::jsonb),
			created_at
		FROM change_events
		WHERE change_request_id = $1::uuid
		ORDER BY created_at ASC
	`, change.ID)
	if err != nil {
		return nil, fmt.Errorf("list change events: %w", err)
	}
	defer rows.Close()

	events := make([]domain.ChangeEvent, 0)
	for rows.Next() {
		var event domain.ChangeEvent
		var rawPayload []byte
		if err := rows.Scan(
			&event.EventType,
			&event.PreviousStatus,
			&event.NewStatus,
			&event.Message,
			&event.ErrorCode,
			&event.Source,
			&rawPayload,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan change event: %w", err)
		}
		event.Payload = decodeMap(rawPayload)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate change events: %w", err)
	}
	return events, nil
}

func (r *ChangeRepository) MarkStep(ctx context.Context, idOrNumber string, status string) (map[string]any, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin mark step transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id string
	var changeNumber string
	var previousStatus string
	if err := tx.QueryRow(ctx, `
		SELECT id::text, change_number, status
		FROM change_requests
		WHERE id::text = $1 OR change_number = $1
	`, idOrNumber).Scan(&id, &changeNumber, &previousStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("change not found: %s", idOrNumber)
		}
		return nil, fmt.Errorf("load change for mark step: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE change_requests
		SET status = $2,
		    updated_at = now(),
		    completed_at = CASE WHEN $2 IN ('Completed', 'Failed', 'Cancelled') THEN now() ELSE completed_at END
		WHERE id = $1::uuid
	`, id, status)
	if err != nil {
		return nil, fmt.Errorf("update change status: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO change_events (
			change_request_id,
			event_type,
			previous_status,
			new_status,
			message,
			source,
			payload
		)
		VALUES ($1::uuid, $2, $3, $2, 'Workflow step executed', 'workflow', '{}'::jsonb)
	`, id, status, previousStatus)
	if err != nil {
		return nil, fmt.Errorf("insert mark step event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit mark step transaction: %w", err)
	}

	return map[string]any{
		"id":             id,
		"changeNumber":   changeNumber,
		"previousStatus": previousStatus,
		"status":         status,
	}, nil
}

type changeScanner interface {
	Scan(dest ...any) error
}

func scanChange(row changeScanner) (domain.ChangeRequest, error) {
	var change domain.ChangeRequest
	var rawPayload []byte
	if err := row.Scan(
		&change.ID,
		&change.ChangeNumber,
		&change.ApplicationName,
		&change.ChangeType,
		&change.Status,
		&change.RequestedBy,
		&change.Description,
		&rawPayload,
		&change.CreatedAt,
		&change.UpdatedAt,
		&change.CompletedAt,
	); err != nil {
		return domain.ChangeRequest{}, fmt.Errorf("scan change_request: %w", err)
	}
	change.Payload = decodeMap(rawPayload)
	return change, nil
}

func decodeMap(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{"_decodeError": err.Error()}
	}
	return out
}
