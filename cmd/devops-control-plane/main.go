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
	githubadapter "github.com/vincmarz/devops-control-plane/internal/adapters/github"
	gitlabadapter "github.com/vincmarz/devops-control-plane/internal/adapters/gitlab"
	kubernetesadapter "github.com/vincmarz/devops-control-plane/internal/adapters/kubernetes"
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

	changeServiceOptions := []app.ChangeServiceOption{
		app.WithTechnicalRuntimeTargetResolver(app.DefaultTechnicalRuntimeTargetResolver(cfg.TektonPipelineName)), app.WithRuntimeClientProviderRegistry(app.DefaultRuntimeClientProviderRegistry()), app.WithRuntimeClientSecretRefsRegistry(app.DefaultRuntimeClientSecretRefsRegistry()), app.WithEvidenceStore(repositories.Evidences)}
	applicationCatalog := app.DefaultApplicationCatalog()
	gitOpsRepositoryTargetResolver := app.NewGitOpsRepositoryTargetResolver(applicationCatalog.ResolveGitOpsBinding)
	changeServiceOptions = append(changeServiceOptions, app.WithGitSourceBindingResolverFunc(applicationCatalog.ResolveSourceBinding))
	gitProviders := make([]app.GitProvider, 0, 2)

	if cfg.GitLabBaseURL != "" || cfg.GitLabToken != "" {
		gitLabClient, err := gitlabadapter.New(gitlabadapter.Config{BaseURL: cfg.GitLabBaseURL, Token: cfg.GitLabToken, TimeoutSeconds: cfg.GitLabTimeoutSeconds, InsecureTLS: cfg.GitLabInsecureTLS, CAFile: cfg.GitLabCAFile})
		if err != nil {
			logger.Error("gitlab client initialization failed", "error", err)
			os.Exit(1)
		}
		gitLabProvider, err := gitlabadapter.NewProvider("gitlab-lab", gitLabClient)
		if err != nil {
			logger.Error("gitlab provider initialization failed", "error", err)
			os.Exit(1)
		}
		gitProviders = append(gitProviders, gitLabProvider)
		logger.Info("gitlab provider enabled", "baseURL", cfg.GitLabBaseURL, "providerRef", gitLabProvider.ProviderRef(), "insecureTLS", cfg.GitLabInsecureTLS)
	}

	if cfg.GitHubToken != "" {
		gitHubClient, err := githubadapter.New(githubadapter.Config{APIURL: cfg.GitHubAPIURL, Token: cfg.GitHubToken, TimeoutSeconds: cfg.GitHubTimeoutSeconds, InsecureTLS: cfg.GitHubInsecureTLS, CAFile: cfg.GitHubCAFile})
		if err != nil {
			logger.Error("github client initialization failed", "error", err)
			os.Exit(1)
		}
		gitHubProvider, err := githubadapter.NewProvider("github-public", gitHubClient)
		if err != nil {
			logger.Error("github provider initialization failed", "error", err)
			os.Exit(1)
		}
		gitProviders = append(gitProviders, gitHubProvider)
		logger.Info("github provider enabled", "apiURL", cfg.GitHubAPIURL, "providerRef", gitHubProvider.ProviderRef(), "insecureTLS", cfg.GitHubInsecureTLS)
	}

	if len(gitProviders) > 0 {
		gitProviderRegistry, err := app.NewGitProviderRegistry(gitProviders)
		if err != nil {
			logger.Error("git provider registry initialization failed", "error", err)
			os.Exit(1)
		}
		changeServiceOptions = append(changeServiceOptions, app.WithGitProviderResolver(gitProviderRegistry))
	} else {
		logger.Info("git provider integration disabled; Git workflow endpoints require a registered provider")
	}

	var runtimeSecretValueLoader app.RuntimeSecretValueLoader = app.EmptyRuntimeSecretValueLoader{}

	kubernetesRuntimeClientFactory, err := buildRuntimeKubernetesClientFactory(cfg)
	if err != nil {
		logger.Error("runtime kubernetes client factory initialization failed", "error", err)
		os.Exit(1)
	}
	tektonRuntimeClientFactory, err := buildRuntimeTektonClientFactory(cfg)
	if err != nil {
		logger.Error("runtime tekton client factory initialization failed", "error", err)
		os.Exit(1)
	}
	argoCDRuntimeClientFactory, err := buildRuntimeArgoCDClientFactory(cfg)
	if err != nil {
		logger.Error("runtime argocd client factory initialization failed", "error", err)
		os.Exit(1)
	}

	if cfg.KubernetesAPIURL != "" || cfg.KubernetesToken != "" {
		kubernetesRuntimeClient, err := kubernetesadapter.New(kubernetesadapter.Config{APIURL: cfg.KubernetesAPIURL, Token: cfg.KubernetesToken, TimeoutSeconds: cfg.TektonTimeoutSeconds, InsecureTLS: cfg.KubernetesInsecureTLS, CAFile: cfg.KubernetesCAFile})
		if err != nil {
			logger.Error("kubernetes client initialization failed", "error", err)
			os.Exit(1)
		}
		runtimeSecretValueLoader, err = buildRuntimeSecretValueLoader(
			cfg,
			app.KubernetesSecretValueLoaderConfig{},
			kubernetesRuntimeClient,
		)
		if err != nil {
			logger.Error("runtime secret value loader initialization failed", "error", err)
			os.Exit(1)
		}
		changeServiceOptions = append(changeServiceOptions, app.WithKubernetesRuntimeEvidenceCollector(func(ctx context.Context, change domain.ChangeRequest) (map[string]any, error) {
			return kubernetesRuntimeClient.CollectRuntimeEvidence(ctx, cfg.KubernetesNamespace, change.ApplicationName)
		}))
		changeServiceOptions = append(changeServiceOptions, app.WithKubernetesRuntimeClientProviderRegistry(
			app.NewKubernetesRuntimeClientProviderFactoryAwareRegistry(
				app.DefaultKubernetesRuntimeClientProviderRegistry(kubernetesRuntimeClient),
				kubernetesRuntimeClientFactory,
				runtimeSecretValueLoader,
			),
		))

		tektonClient, err := tektonadapter.New(tektonadapter.Config{APIURL: cfg.KubernetesAPIURL, Token: cfg.KubernetesToken, TimeoutSeconds: cfg.TektonTimeoutSeconds, InsecureTLS: cfg.KubernetesInsecureTLS, CAFile: cfg.KubernetesCAFile})
		if err != nil {
			logger.Error("tekton client initialization failed", "error", err)
			os.Exit(1)
		}
		tektonRuntimeClientProviderRegistry := app.NewTektonRuntimeClientProviderFactoryAwareRegistry(
			app.DefaultTektonRuntimeClientProviderRegistry(currentTektonRuntimeClient{client: tektonClient}),
			tektonRuntimeClientFactory,
			runtimeSecretValueLoader,
		)
		changeServiceOptions = append(changeServiceOptions, app.WithTektonRunPipeline(func(ctx context.Context, change domain.ChangeRequest) (string, string, string, error) {
			gitOpsTarget, err := gitOpsRepositoryTargetResolver.Resolve(change.ApplicationName, app.GitOpsConsumerTekton)
			if err != nil {
				return "", "", "", err
			}
			revision, err := app.ResolveTektonGitRevision(gitOpsTarget)
			if err != nil {
				return "", "", "", err
			}

			target, err := app.DefaultTechnicalRuntimeTargetResolver(cfg.TektonPipelineName).Resolve(change.TargetEnvironment)
			if err != nil {
				return "", "", "", err
			}
			validationPath := target.ValidationPath
			if validationPath == "" {
				validationPath = cfg.TektonValidationPath
			}
			selection, err := app.DefaultRuntimeClientProviderRegistry().Select(target)
			if err != nil {
				return "", "", "", err
			}
			tektonRuntimeClient, err := tektonRuntimeClientProviderRegistry.Resolve(ctx, selection)
			if err != nil {
				return "", "", "", err
			}
			ref, err := tektonRuntimeClient.CreatePipelineRun(ctx, app.TektonRuntimePipelineRunRequest{
				Namespace:          target.TektonNamespace,
				PipelineName:       target.TektonPipelineName,
				ServiceAccountName: cfg.TektonServiceAccount,
				GenerateName:       "devops-cp-validate-" + strings.ToLower(change.ChangeNumber) + "-",
				ChangeNumber:       change.ChangeNumber,
				GitURL:             gitOpsTarget.RepositoryURL,
				GitRevision:        revision,
				ValidationPath:     validationPath,
				Image:              cfg.TektonImage,
				WorkspacePVC:       cfg.TektonWorkspacePVC,
				DockerConfigSecret: cfg.TektonDockerConfigSecret,
			})
			return ref.Name, ref.Namespace, ref.UID, err
		}))

		changeServiceOptions = append(changeServiceOptions, app.WithTektonCheckValidation(func(ctx context.Context, change domain.ChangeRequest) (app.TektonValidationResult, error) {
			gitOpsTarget, err := gitOpsRepositoryTargetResolver.Resolve(change.ApplicationName, app.GitOpsConsumerTekton)
			if err != nil {
				return app.TektonValidationResult{}, err
			}
			target, err := app.DefaultTechnicalRuntimeTargetResolver(cfg.TektonPipelineName).Resolve(change.TargetEnvironment)
			if err != nil {
				return app.TektonValidationResult{}, err
			}
			selection, err := app.DefaultRuntimeClientProviderRegistry().Select(target)
			if err != nil {
				return app.TektonValidationResult{}, err
			}
			tektonRuntimeClient, err := tektonRuntimeClientProviderRegistry.Resolve(ctx, selection)
			if err != nil {
				return app.TektonValidationResult{}, err
			}
			status, err := tektonRuntimeClient.FindLatestPipelineRunByChange(ctx, target.TektonNamespace, change.ChangeNumber)
			revision, err := app.ResolveTektonGitRevision(gitOpsTarget)
			if err != nil {
				return app.TektonValidationResult{}, err
			}
			validationPath := target.ValidationPath
			if validationPath == "" {
				validationPath = cfg.TektonValidationPath
			}
			result := app.TektonValidationResult{PipelineRunName: status.PipelineRunName, Namespace: status.Namespace, UID: status.UID, Status: status.Status, Reason: status.Reason, Message: status.Message, PipelineName: target.TektonPipelineName, GitURL: gitOpsTarget.RepositoryURL, GitRevision: revision, ValidationPath: validationPath}
			if err != nil {
				return result, err
			}
			taskRuns, taskRunErr := tektonRuntimeClient.ListTaskRunsByPipelineRun(ctx, target.TektonNamespace, status.PipelineRunName)
			if taskRunErr == nil {
				for _, taskRun := range taskRuns {
					result.TaskRuns = append(result.TaskRuns, app.TektonTaskRunResult{Name: taskRun.Name, Namespace: taskRun.Namespace, PipelineTaskName: taskRun.PipelineTaskName, TaskName: taskRun.TaskName, Status: taskRun.Status, Reason: taskRun.Reason, Message: taskRun.Message, StartTime: taskRun.StartTime, CompletionTime: taskRun.CompletionTime})
				}
			}
			return result, nil
		}))
		logger.Info("tekton integration enabled", "namespace", cfg.TektonNamespace, "pipeline", cfg.TektonPipelineName, "serviceAccount", cfg.TektonServiceAccount, "apiURL", cfg.KubernetesAPIURL, "insecureTLS", cfg.KubernetesInsecureTLS)
	} else {
		logger.Info("tekton integration disabled; validate endpoint will require Kubernetes API configuration")
	}

	applicationService := app.NewApplicationService()
	if cfg.ArgoCDBaseURL != "" || cfg.ArgoCDAuthToken != "" {
		argoCDClient, err := argocdadapter.New(argocdadapter.Config{BaseURL: cfg.ArgoCDBaseURL, AuthToken: cfg.ArgoCDAuthToken, TimeoutSeconds: cfg.ArgoCDTimeoutSeconds, InsecureTLS: cfg.ArgoCDInsecureTLS, CAFile: cfg.ArgoCDCAFile})
		if err != nil {
			logger.Error("argocd client initialization failed", "error", err)
			os.Exit(1)
		}
		applicationService = app.NewApplicationService(argoCDClient)
		argoCDRuntimeClientProviderRegistry := app.NewArgoCDRuntimeClientProviderFactoryAwareRegistry(
			app.DefaultArgoCDRuntimeClientProviderRegistry(currentArgoCDRuntimeClient{client: argoCDClient}),
			argoCDRuntimeClientFactory,
			runtimeSecretValueLoader,
		)
		changeServiceOptions = append(changeServiceOptions, app.WithArgoCDCheckDeployment(func(ctx context.Context, change domain.ChangeRequest) (app.ArgoCDDeploymentResult, error) {
			gitOpsTarget, err := gitOpsRepositoryTargetResolver.Resolve(change.ApplicationName, app.GitOpsConsumerArgoCD)
			if err != nil {
				return app.ArgoCDDeploymentResult{}, err
			}
			target, err := app.DefaultTechnicalRuntimeTargetResolver(cfg.TektonPipelineName).Resolve(change.TargetEnvironment)
			if err != nil {
				return app.ArgoCDDeploymentResult{}, err
			}
			selection, err := app.DefaultRuntimeClientProviderRegistry().Select(target)
			if err != nil {
				return app.ArgoCDDeploymentResult{}, err
			}
			argoCDRuntimeClient, err := argoCDRuntimeClientProviderRegistry.Resolve(ctx, selection)
			if err != nil {
				return app.ArgoCDDeploymentResult{}, err
			}
			deployment, err := argoCDRuntimeClient.CheckDeployment(ctx, target.ArgoCDApplicationName)
			if err != nil {
				return app.ArgoCDDeploymentResult{}, err
			}
			deployment.GitOpsProvider = gitOpsTarget.Provider
			deployment.GitOpsProviderRef = gitOpsTarget.ProviderRef
			deployment.GitOpsProjectPath = gitOpsTarget.ProjectPath
			deployment.DeclaredRepositoryURL = gitOpsTarget.RepositoryURL
			deployment.DeclaredDefaultBranch = gitOpsTarget.DefaultBranch
			return deployment, nil
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
