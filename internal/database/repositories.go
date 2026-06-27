package database

import "context"

type RepositorySet struct {
	DB        *DB
	Changes   *ChangeRepository
	Evidences *EvidenceRepository
}

func NewRepositorySet(db *DB) *RepositorySet {
	return &RepositorySet{
		DB:        db,
		Changes:   NewChangeRepository(db),
		Evidences: NewEvidenceRepository(db),
	}
}

func (r *RepositorySet) Ping(ctx context.Context) error {
	return r.DB.Ping(ctx)
}
