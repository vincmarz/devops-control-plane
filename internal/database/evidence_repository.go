package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type EvidenceRepository struct {
	db *DB
}

func NewEvidenceRepository(db *DB) *EvidenceRepository {
	return &EvidenceRepository{db: db}
}

func (r *EvidenceRepository) List(ctx context.Context, idOrNumber string, evidenceType string) ([]domain.Evidence, error) {
	rows, err := r.db.Pool.Query(ctx, `
        SELECT
            e.id::text,
            cr.change_number,
            e.evidence_type,
            COALESCE(e.name, ''),
            COALESCE(e.summary, ''),
            COALESCE(e.content, ''),
            COALESCE(e.payload, '{}'::jsonb),
            COALESCE(e.external_ref, ''),
            e.sanitized,
            e.created_at
        FROM evidences e
        JOIN change_requests cr ON cr.id = e.change_request_id
        WHERE (cr.id::text = $1 OR cr.change_number = $1)
          AND ($2 = '' OR e.evidence_type = $2)
        ORDER BY e.created_at DESC
    `, idOrNumber, evidenceType)
	if err != nil {
		return nil, fmt.Errorf("list evidences: %w", err)
	}
	defer rows.Close()

	evidences := make([]domain.Evidence, 0)
	for rows.Next() {
		evidence, err := scanEvidence(rows)
		if err != nil {
			return nil, err
		}
		evidences = append(evidences, evidence)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate evidences: %w", err)
	}
	return evidences, nil
}

func (r *EvidenceRepository) Create(ctx context.Context, changeID string, evidence domain.Evidence) (domain.Evidence, error) {
	rawPayload, err := json.Marshal(evidence.Payload)
	if err != nil {
		return domain.Evidence{}, fmt.Errorf("marshal evidence payload: %w", err)
	}

	var saved domain.Evidence
	var rawSavedPayload []byte
	if err := r.db.Pool.QueryRow(ctx, `
        INSERT INTO evidences (
            change_request_id,
            evidence_type,
            name,
            summary,
            content,
            payload,
            external_ref,
            sanitized
        )
        VALUES ($1::uuid, $2, $3, $4, $5, $6::jsonb, $7, $8)
        RETURNING
            id::text,
            $9::text,
            evidence_type,
            COALESCE(name, ''),
            COALESCE(summary, ''),
            COALESCE(content, ''),
            COALESCE(payload, '{}'::jsonb),
            COALESCE(external_ref, ''),
            sanitized,
            created_at
    `,
		changeID,
		evidence.EvidenceType,
		evidence.Name,
		evidence.Summary,
		evidence.Content,
		string(rawPayload),
		evidence.ExternalRef,
		evidence.Sanitized,
		evidence.ChangeNumber,
	).Scan(
		&saved.ID,
		&saved.ChangeNumber,
		&saved.EvidenceType,
		&saved.Name,
		&saved.Summary,
		&saved.Content,
		&rawSavedPayload,
		&saved.ExternalRef,
		&saved.Sanitized,
		&saved.CreatedAt,
	); err != nil {
		return domain.Evidence{}, fmt.Errorf("insert evidence: %w", err)
	}
	saved.Payload = decodeMap(rawSavedPayload)
	return saved, nil
}

type evidenceScanner interface {
	Scan(dest ...any) error
}

func scanEvidence(row evidenceScanner) (domain.Evidence, error) {
	var evidence domain.Evidence
	var rawPayload []byte
	if err := row.Scan(
		&evidence.ID,
		&evidence.ChangeNumber,
		&evidence.EvidenceType,
		&evidence.Name,
		&evidence.Summary,
		&evidence.Content,
		&rawPayload,
		&evidence.ExternalRef,
		&evidence.Sanitized,
		&evidence.CreatedAt,
	); err != nil {
		return domain.Evidence{}, fmt.Errorf("scan evidence: %w", err)
	}
	evidence.Payload = decodeMap(rawPayload)
	return evidence, nil
}
