package database

import "context"

type RepositorySet struct {
	DB *DB
}

func NewRepositorySet(db *DB) *RepositorySet {
	return &RepositorySet{DB: db}
}

func (r *RepositorySet) Ping(ctx context.Context) error {
	return r.DB.Ping(ctx)
}

// TODO: implement PostgreSQL repositories for:
// - applications
// - change_requests
// - change_events
// - evidences
