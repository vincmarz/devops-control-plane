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
            title,
            application_name,
            target_environment,
            change_type,
            status,
            COALESCE(runtime_status, ''),
            risk_level,
            requested_by,
            COALESCE(description, ''),
            COALESCE(request_payload, '{}'::jsonb),
            created_at,
            updated_at,
            submitted_at,
            approved_at,
            rejected_at,
            executed_at,
            closed_at,
            cancelled_at,
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
	if req.Title == "" {
		return domain.ChangeRequest{}, errors.New("title is required")
	}
	if req.ApplicationName == "" {
		return domain.ChangeRequest{}, errors.New("applicationName is required")
	}
	if req.TargetEnvironment == "" {
		return domain.ChangeRequest{}, errors.New("targetEnvironment is required")
	}
	if req.ChangeType == "" {
		return domain.ChangeRequest{}, errors.New("changeType is required")
	}
	if req.RiskLevel == "" {
		return domain.ChangeRequest{}, errors.New("riskLevel is required")
	}
	if req.RequestedBy == "" {
		return domain.ChangeRequest{}, errors.New("requestedBy is required")
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
            title,
            application_name,
            target_environment,
            change_type,
            status,
            runtime_status,
            risk_level,
            requested_by,
            description,
            request_payload
        )
        VALUES ($1, $2, $3, $4, $5, 'draft', '', $6, $7, $8, $9::jsonb)
        RETURNING
            id::text,
            change_number,
            title,
            application_name,
            target_environment,
            change_type,
            status,
            COALESCE(runtime_status, ''),
            risk_level,
            requested_by,
            COALESCE(description, ''),
            COALESCE(request_payload, '{}'::jsonb),
            created_at,
            updated_at,
            submitted_at,
            approved_at,
            rejected_at,
            executed_at,
            closed_at,
            cancelled_at,
            completed_at
    `, changeNumber, req.Title, req.ApplicationName, req.TargetEnvironment, req.ChangeType, req.RiskLevel, req.RequestedBy, req.Description, string(payload)).Scan(
		&change.ID,
		&change.ChangeNumber,
		&change.Title,
		&change.ApplicationName,
		&change.TargetEnvironment,
		&change.ChangeType,
		&change.Status,
		&change.RuntimeStatus,
		&change.RiskLevel,
		&change.RequestedBy,
		&change.Description,
		&rawPayload,
		&change.CreatedAt,
		&change.UpdatedAt,
		&change.SubmittedAt,
		&change.ApprovedAt,
		&change.RejectedAt,
		&change.ExecutedAt,
		&change.ClosedAt,
		&change.CancelledAt,
		&change.CompletedAt,
	); err != nil {
		return domain.ChangeRequest{}, fmt.Errorf("insert change_request: %w", err)
	}
	change.Payload = decodeMap(rawPayload)

	if _, err := tx.Exec(ctx, `
        INSERT INTO change_events (
            change_request_id,
            event_type,
            previous_status,
            new_status,
            message,
            source,
            payload
        )
        VALUES (
            $1::uuid,
            $2,
            NULL,
            $3,
            'ChangeRequest created',
            'workflow',
            '{}'::jsonb
        )
    `, change.ID, domain.ChangeEventCreated, domain.ChangeStatusDraft); err != nil {
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
            title,
            application_name,
            target_environment,
            change_type,
            status,
            COALESCE(runtime_status, ''),
            risk_level,
            requested_by,
            COALESCE(description, ''),
            COALESCE(request_payload, '{}'::jsonb),
            created_at,
            updated_at,
            submitted_at,
            approved_at,
            rejected_at,
            executed_at,
            closed_at,
            cancelled_at,
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

func (r *ChangeRepository) TransitionLifecycle(ctx context.Context, idOrNumber string, action string, actor string, message string) (map[string]any, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin lifecycle transition transaction: %w", err)
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
		return nil, fmt.Errorf("load change for lifecycle transition: %w", err)
	}

	targetStatus, eventType, timestampColumn, markCompleted, err := lifecycleTransition(action, previousStatus)
	if err != nil {
		return nil, err
	}

	timestampSetClause := ""
	if timestampColumn != "" {
		timestampSetClause = fmt.Sprintf(",\n\t\t    %s = now()", timestampColumn)
	}

	updateSQL := fmt.Sprintf(`
		UPDATE change_requests
		SET status = $2,
		    updated_at = now()%s%s
		WHERE id = $1::uuid
	`, timestampSetClause, completedAtClause(markCompleted))

	if _, err := tx.Exec(ctx, updateSQL, id, targetStatus); err != nil {
		return nil, fmt.Errorf("update change lifecycle status: %w", err)
	}

	if message == "" {
		message = fmt.Sprintf("ChangeRequest lifecycle transition: %s", action)
	}

	eventPayload := map[string]any{
		"action":         action,
		"actor":          actor,
		"previousStatus": previousStatus,
		"newStatus":      targetStatus,
	}
	rawEventPayload, err := json.Marshal(eventPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal lifecycle transition payload: %w", err)
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
		VALUES (
			$1::uuid,
			$2,
			$3,
			$4,
			$5,
			'workflow',
			$6::jsonb
		)
	`, id, eventType, previousStatus, targetStatus, message, string(rawEventPayload))
	if err != nil {
		return nil, fmt.Errorf("insert lifecycle transition event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit lifecycle transition transaction: %w", err)
	}

	return map[string]any{
		"id":             id,
		"changeNumber":   changeNumber,
		"action":         action,
		"actor":          actor,
		"previousStatus": previousStatus,
		"status":         targetStatus,
	}, nil
}

func lifecycleTransition(action string, currentStatus string) (targetStatus string, eventType string, timestampColumn string, markCompleted bool, err error) {
	switch action {
	case "submit":
		return requireLifecycleStatus(action, currentStatus, domain.ChangeStatusDraft, domain.ChangeStatusSubmitted, domain.ChangeEventSubmitted, "submitted_at", false)
	case "approve":
		return requireLifecycleStatus(action, currentStatus, domain.ChangeStatusSubmitted, domain.ChangeStatusApproved, domain.ChangeEventApproved, "approved_at", false)
	case "reject":
		return requireLifecycleStatus(action, currentStatus, domain.ChangeStatusSubmitted, domain.ChangeStatusRejected, domain.ChangeEventRejected, "rejected_at", true)
	case "start-execution":
		return requireLifecycleStatus(action, currentStatus, domain.ChangeStatusApproved, domain.ChangeStatusExecuting, domain.ChangeEventExecutionStarted, "", false)
	case "complete-execution":
		return requireLifecycleStatus(action, currentStatus, domain.ChangeStatusExecuting, domain.ChangeStatusExecuted, domain.ChangeEventExecutionCompleted, "executed_at", false)
	case "fail-execution":
		return requireLifecycleStatus(action, currentStatus, domain.ChangeStatusExecuting, domain.ChangeStatusFailed, domain.ChangeEventExecutionFailed, "executed_at", true)
	case "close":
		if currentStatus != domain.ChangeStatusExecuted && currentStatus != domain.ChangeStatusFailed {
			return "", "", "", false, fmt.Errorf("invalid lifecycle transition %q from status %q", action, currentStatus)
		}
		return domain.ChangeStatusClosed, domain.ChangeEventClosed, "closed_at", true, nil
	case "cancel":
		if currentStatus != domain.ChangeStatusDraft && currentStatus != domain.ChangeStatusSubmitted && currentStatus != domain.ChangeStatusApproved {
			return "", "", "", false, fmt.Errorf("invalid lifecycle transition %q from status %q", action, currentStatus)
		}
		return domain.ChangeStatusCancelled, domain.ChangeEventCancelled, "cancelled_at", true, nil
	default:
		return "", "", "", false, fmt.Errorf("unsupported lifecycle action %q", action)
	}
}

func requireLifecycleStatus(action string, currentStatus string, requiredStatus string, targetStatus string, eventType string, timestampColumn string, markCompleted bool) (string, string, string, bool, error) {
	if currentStatus != requiredStatus {
		return "", "", "", false, fmt.Errorf("invalid lifecycle transition %q from status %q: expected %q", action, currentStatus, requiredStatus)
	}
	return targetStatus, eventType, timestampColumn, markCompleted, nil
}

func completedAtClause(markCompleted bool) string {
	if markCompleted {
		return ", completed_at = now()"
	}
	return ""
}

// MarkStep registra uno step tecnico del workflow.
//
// Nota importante:
// questo metodo NON modifica change_requests.status.
// Lo stato governato resta draft/submitted/approved/executing/executed/...
// Lo step tecnico viene salvato in:
// - change_requests.runtime_status
// - change_events con event_type = technical_step_completed
func (r *ChangeRepository) MarkStep(ctx context.Context, idOrNumber string, step string) (map[string]any, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin mark step transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id string
	var changeNumber string
	var lifecycleStatus string
	var previousRuntimeStatus string

	if err := tx.QueryRow(ctx, `
        SELECT
            id::text,
            change_number,
            status,
            COALESCE(runtime_status, '')
        FROM change_requests
        WHERE id::text = $1 OR change_number = $1
    `, idOrNumber).Scan(&id, &changeNumber, &lifecycleStatus, &previousRuntimeStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("change not found: %s", idOrNumber)
		}
		return nil, fmt.Errorf("load change for mark step: %w", err)
	}

	_, err = tx.Exec(ctx, `
        UPDATE change_requests
        SET runtime_status = $2,
            updated_at = now()
        WHERE id = $1::uuid
    `, id, step)
	if err != nil {
		return nil, fmt.Errorf("update change runtime status: %w", err)
	}

	eventPayload := map[string]any{
		"step":                  step,
		"previousRuntimeStatus": previousRuntimeStatus,
		"lifecycleStatus":       lifecycleStatus,
	}

	rawEventPayload, err := json.Marshal(eventPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal technical step payload: %w", err)
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
        VALUES (
            $1::uuid,
            $2,
            $3,
            $4,
            'Technical workflow step completed',
            'workflow',
            $5::jsonb
        )
    `, id, domain.ChangeEventTechnicalStepCompleted, lifecycleStatus, lifecycleStatus, string(rawEventPayload))
	if err != nil {
		return nil, fmt.Errorf("insert mark step event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit mark step transaction: %w", err)
	}

	return map[string]any{
		"id":                    id,
		"changeNumber":          changeNumber,
		"status":                lifecycleStatus,
		"runtimeStatus":         step,
		"previousRuntimeStatus": previousRuntimeStatus,
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
		&change.Title,
		&change.ApplicationName,
		&change.TargetEnvironment,
		&change.ChangeType,
		&change.Status,
		&change.RuntimeStatus,
		&change.RiskLevel,
		&change.RequestedBy,
		&change.Description,
		&rawPayload,
		&change.CreatedAt,
		&change.UpdatedAt,
		&change.SubmittedAt,
		&change.ApprovedAt,
		&change.RejectedAt,
		&change.ExecutedAt,
		&change.ClosedAt,
		&change.CancelledAt,
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
