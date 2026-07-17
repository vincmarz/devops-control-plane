package database

import "context"

type RepositorySet struct {
	DB            *DB
	Changes       *ChangeRepository
	Evidences     *EvidenceRepository
	RuntimeStates *ChangeRuntimeStateRepository
}

func NewRepositorySet(db *DB) *RepositorySet {
	return &RepositorySet{
		DB:            db,
		Changes:       NewChangeRepository(db),
		Evidences:     NewEvidenceRepository(db),
		RuntimeStates: NewChangeRuntimeStateRepository(db),
	}
}

func (r *RepositorySet) Ping(ctx context.Context) error {
	return r.DB.Ping(ctx)
}
