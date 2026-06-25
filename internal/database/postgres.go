package database

import (
	"context"
	"errors"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Placeholder struct {
	DatabaseURL string
}

func NewPlaceholder(databaseURL string) *Placeholder {
	return &Placeholder{DatabaseURL: databaseURL}
}

func (p *Placeholder) Ping(ctx context.Context) error {
	if p.DatabaseURL == "" {
		return errors.New("DATABASE_URL not configured")
	}
	// TODO: replace with real PostgreSQL connection using pgx or database/sql driver.
	return nil
}
