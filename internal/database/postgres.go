package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type DB struct {
	Pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*DB, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	cfg.MaxConns = 5
	cfg.MinConns = 0
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	db := &DB{Pool: pool}
	if err := db.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

func (db *DB) Ping(ctx context.Context) error {
	if db == nil || db.Pool == nil {
		return errors.New("postgres pool is not initialized")
	}
	return db.Pool.Ping(ctx)
}

func (db *DB) Close() {
	if db != nil && db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *DB) ExecSQLFile(ctx context.Context, path string) error {
	if db == nil || db.Pool == nil {
		return errors.New("postgres pool is not initialized")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read sql file %s: %w", path, err)
	}

	sql := strings.TrimSpace(string(content))
	if sql == "" {
		return fmt.Errorf("sql file %s is empty", path)
	}

	_, err = db.Pool.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("execute sql file %s: %w", path, err)
	}

	return nil
}
