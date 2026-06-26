package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gitlabadapter "github.com/vincmarz/devops-control-plane/internal/adapters/gitlab"
	"github.com/vincmarz/devops-control-plane/internal/api"
	"github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/config"
	"github.com/vincmarz/devops-control-plane/internal/database"
	"github.com/vincmarz/devops-control-plane/internal/logging"
)

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.LogLevel)
	slog.SetDefault(logger)

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startupCancel()

	db, err := database.Open(startupCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repositories := database.NewRepositorySet(db)

	changeServiceOptions := []app.ChangeServiceOption{}
	if cfg.GitLabBaseURL != "" || cfg.GitLabToken != "" || cfg.GitLabProjectID > 0 {
		gitLabClient, err := gitlabadapter.New(gitlabadapter.Config{
			BaseURL:        cfg.GitLabBaseURL,
			Token:          cfg.GitLabToken,
			TimeoutSeconds: cfg.GitLabTimeoutSeconds,
			InsecureTLS:    cfg.GitLabInsecureTLS,
		})
		if err != nil {
			logger.Error("gitlab client initialization failed", "error", err)
			os.Exit(1)
		}
		if cfg.GitLabProjectID <= 0 {
			logger.Error("gitlab project ID must be configured when GitLab integration is enabled")
			os.Exit(1)
		}
		changeServiceOptions = append(changeServiceOptions,
			app.WithGitCreateBranch(
				func(ctx context.Context, projectID int, branch string, ref string) error {
					_, err := gitLabClient.CreateBranch(ctx, projectID, branch, ref)
					return err
				},
				cfg.GitLabProjectID,
				cfg.GitLabDefaultBranch,
			),
			app.WithGitCreateOrUpdateFile(
				func(ctx context.Context, projectID int, branch string, filePath string, commitMessage string, content string) error {
					_, err := gitLabClient.CreateOrUpdateFile(ctx, projectID, branch, filePath, commitMessage, content)
					return err
				},
			),
			app.WithGitOpenMergeRequest(
				func(ctx context.Context, projectID int, sourceBranch string, targetBranch string, title string, description string) (int, string, error) {
					mr, err := gitLabClient.OpenMergeRequest(ctx, projectID, sourceBranch, targetBranch, title, description)
					return mr.IID, mr.WebURL, err
				},
			),
		)
		logger.Info("gitlab integration enabled", "baseURL", cfg.GitLabBaseURL, "projectID", cfg.GitLabProjectID, "defaultBranch", cfg.GitLabDefaultBranch, "insecureTLS", cfg.GitLabInsecureTLS)
	} else {
		logger.Info("gitlab integration disabled; create-branch endpoint will require GitLab configuration")
	}

	services := app.Services{
		Applications: app.NewApplicationService(),
		Changes:      app.NewChangeService(repositories.Changes, changeServiceOptions...),
		Evidence:     app.NewEvidenceService(),
		DB:           db,
	}

	router := api.NewRouter(api.Dependencies{
		Config:   cfg,
		Logger:   logger,
		Services: services,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("starting DevOps Control Plane", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	logger.Info("shutting down DevOps Control Plane")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
