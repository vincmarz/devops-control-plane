package database

import "context"

type RepositorySet struct {
	DB      *DB
	Changes *ChangeRepository
}

func NewRepositorySet(db *DB) *RepositorySet {
	return &RepositorySet{
		DB:      db,
		Changes: NewChangeRepository(db),
	}
}

func (r *RepositorySet) Ping(ctx context.Context) error {
	return r.DB.Ping(ctx)
}
