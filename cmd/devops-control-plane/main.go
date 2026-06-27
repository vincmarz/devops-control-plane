package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	argocdadapter "github.com/vincmarz/devops-control-plane/internal/adapters/argocd"
	gitlabadapter "github.com/vincmarz/devops-control-plane/internal/adapters/gitlab"
	tektonadapter "github.com/vincmarz/devops-control-plane/internal/adapters/tekton"
	"github.com/vincmarz/devops-control-plane/internal/api"
	"github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/config"
	"github.com/vincmarz/devops-control-plane/internal/database"
	"github.com/vincmarz/devops-control-plane/internal/domain"
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

	changeServiceOptions := []app.ChangeServiceOption{app.WithEvidenceStore(repositories.Evidences)}
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
			app.WithGitMergeRequest(
				func(ctx context.Context, projectID int, sourceBranch string, targetBranch string, mergeCommitMessage string) (int, string, string, error) {
					mr, err := gitLabClient.FindOpenMergeRequest(ctx, projectID, sourceBranch, targetBranch)
					if err != nil {
						return 0, "", "", err
					}
					merged, err := gitLabClient.MergeMergeRequest(ctx, projectID, mr.IID, mr.SHA, mergeCommitMessage)
					return merged.IID, merged.WebURL, merged.MergeCommitSHA, err
				},
			),
		)
		logger.Info("gitlab integration enabled", "baseURL", cfg.GitLabBaseURL, "projectID", cfg.GitLabProjectID, "defaultBranch", cfg.GitLabDefaultBranch, "insecureTLS", cfg.GitLabInsecureTLS)
	} else {
		logger.Info("gitlab integration disabled; create-branch endpoint will require GitLab configuration")
	}

	if cfg.KubernetesAPIURL != "" || cfg.KubernetesToken != "" {
		tektonClient, err := tektonadapter.New(tektonadapter.Config{APIURL: cfg.KubernetesAPIURL, Token: cfg.KubernetesToken, TimeoutSeconds: cfg.TektonTimeoutSeconds, InsecureTLS: cfg.KubernetesInsecureTLS})
		if err != nil {
			logger.Error("tekton client initialization failed", "error", err)
			os.Exit(1)
		}
		changeServiceOptions = append(changeServiceOptions, app.WithTektonRunPipeline(func(ctx context.Context, change domain.ChangeRequest) (string, string, string, error) {
			revision := cfg.TektonGitRevision
			if cfg.TektonGitRevisionTemplate != "" {
				revision = strings.ReplaceAll(cfg.TektonGitRevisionTemplate, "{changeNumber}", change.ChangeNumber)
			}

			ref, err := tektonClient.CreatePipelineRun(ctx, tektonadapter.CreatePipelineRunRequest{
				Namespace:          cfg.TektonNamespace,
				PipelineName:       cfg.TektonPipelineName,
				ServiceAccountName: cfg.TektonServiceAccount,
				GenerateName:       "devops-cp-validate-" + strings.ToLower(change.ChangeNumber) + "-",
				ChangeNumber:       change.ChangeNumber,
				GitURL:             cfg.TektonGitURL,
				GitRevision:        revision,
				ValidationPath:     cfg.TektonValidationPath,
				Image:              cfg.TektonImage,
				WorkspacePVC:       cfg.TektonWorkspacePVC,
				DockerConfigSecret: cfg.TektonDockerConfigSecret,
			})
			return ref.Name, ref.Namespace, ref.UID, err
		}))

		changeServiceOptions = append(changeServiceOptions, app.WithTektonCheckValidation(func(ctx context.Context, change domain.ChangeRequest) (app.TektonValidationResult, error) {
			status, err := tektonClient.FindLatestPipelineRunByChange(ctx, cfg.TektonNamespace, change.ChangeNumber)
			return app.TektonValidationResult{PipelineRunName: status.Name, Namespace: status.Namespace, UID: status.UID, Status: status.Status, Reason: status.Reason, Message: status.Message}, err
		}))
		logger.Info("tekton integration enabled", "namespace", cfg.TektonNamespace, "pipeline", cfg.TektonPipelineName, "serviceAccount", cfg.TektonServiceAccount, "apiURL", cfg.KubernetesAPIURL, "insecureTLS", cfg.KubernetesInsecureTLS)
	} else {
		logger.Info("tekton integration disabled; validate endpoint will require Kubernetes API configuration")
	}

	applicationService := app.NewApplicationService()
	if cfg.ArgoCDBaseURL != "" || cfg.ArgoCDAuthToken != "" {
		argoCDClient, err := argocdadapter.New(argocdadapter.Config{BaseURL: cfg.ArgoCDBaseURL, AuthToken: cfg.ArgoCDAuthToken, TimeoutSeconds: cfg.ArgoCDTimeoutSeconds, InsecureTLS: cfg.ArgoCDInsecureTLS})
		if err != nil {
			logger.Error("argocd client initialization failed", "error", err)
			os.Exit(1)
		}
		applicationService = app.NewApplicationService(argoCDClient)
		changeServiceOptions = append(changeServiceOptions, app.WithArgoCDCheckDeployment(func(ctx context.Context, change domain.ChangeRequest) (app.ArgoCDDeploymentResult, error) {
			argoApp, err := argoCDClient.GetApplication(ctx, change.ApplicationName)
			return app.ArgoCDDeploymentResult{ApplicationName: argoApp.Name, Project: argoApp.Project, SyncStatus: argoApp.SyncStatus, HealthStatus: argoApp.HealthStatus, Revision: argoApp.CurrentRevision}, err
		}))
		changeServiceOptions = append(changeServiceOptions, app.WithDeploymentEvidenceCollector(func(ctx context.Context, change domain.ChangeRequest) (domain.Evidence, error) {
			argoApp, err := argoCDClient.GetApplication(ctx, change.ApplicationName)
			if err != nil {
				return domain.Evidence{}, err
			}
			return domain.Evidence{
				ChangeNumber: change.ChangeNumber,
				EvidenceType: "deployment",
				Name:         "deployment-evidence",
				Summary:      "Post-deployment evidence for " + change.ApplicationName,
				Sanitized:    true,
				ExternalRef:  argoApp.CurrentRevision,
				Payload: map[string]any{
					"change": map[string]any{
						"changeNumber":      change.ChangeNumber,
						"applicationName":   change.ApplicationName,
						"targetEnvironment": change.TargetEnvironment,
						"lifecycleStatus":   change.Status,
					},
					"argocd": map[string]any{
						"applicationName": argoApp.Name,
						"project":         argoApp.Project,
						"syncStatus":      argoApp.SyncStatus,
						"healthStatus":    argoApp.HealthStatus,
						"revision":        argoApp.CurrentRevision,
						"repoUrl":         argoApp.RepoURL,
						"targetRevision":  argoApp.TargetRevision,
						"path":            argoApp.Path,
						"conditions":      argoApp.Conditions,
					},
				},
			}, nil
		}))
		logger.Info("argocd integration enabled", "baseURL", cfg.ArgoCDBaseURL, "insecureTLS", cfg.ArgoCDInsecureTLS)
	} else {
		logger.Info("argocd integration disabled; check-deployment endpoint will require Argo CD configuration")
	}

	services := app.Services{
		Applications: applicationService,
		Changes:      app.NewChangeService(repositories.Changes, changeServiceOptions...),
		Evidence:     app.NewEvidenceService(repositories.Evidences),
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
