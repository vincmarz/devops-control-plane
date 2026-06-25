package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/vincmarz/devops-control-plane/internal/config"
	"github.com/vincmarz/devops-control-plane/internal/database"
	"github.com/vincmarz/devops-control-plane/internal/logging"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up or down")
	upPath := flag.String("up", "migrations/000001_init.up.sql", "path to up migration sql file")
	downPath := flag.String("down", "migrations/000001_init.down.sql", "path to down migration sql file")
	flag.Parse()

	cfg := config.Load()
	logger := logging.New(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	path := *upPath
	if *direction == "down" {
		path = *downPath
	}

	logger.Info("applying migration", "direction", *direction, "path", path)
	if err := db.ExecSQLFile(ctx, path); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}

	logger.Info("migration completed", "direction", *direction, "path", path)
}
