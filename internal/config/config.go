package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr string
	LogLevel string

	DatabaseURL string

	ArgoCDBaseURL          string
	ArgoCDAuthToken        string
	ArgoCDInsecureTLS      bool
	ArgoCDCAFile           string
	ArgoCDTimeoutSeconds   int
	ArgoCDPollIntervalSecs int

	GitLabBaseURL        string
	GitLabToken          string
	GitLabProjectID      int
	GitLabDefaultBranch  string
	GitLabTimeoutSeconds int
	GitLabInsecureTLS    bool
	GitLabCAFile         string

	TektonNamespace           string
	TektonPipelineName        string
	TektonServiceAccount      string
	TektonTimeoutSeconds      int
	TektonPollIntervalSecs    int
	TektonGitURL              string
	TektonGitRevision         string
	TektonGitRevisionTemplate string
	TektonValidationPath      string
	TektonImage               string
	TektonWorkspacePVC        string
	TektonDockerConfigSecret  string

	KubernetesAPIURL      string
	KubernetesToken       string
	KubernetesInsecureTLS bool
	KubernetesCAFile      string
	KubernetesNamespace   string
	EvidenceBasePath      string
}

func Load() Config {
	return Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		ArgoCDBaseURL:          getEnv("ARGOCD_BASE_URL", ""),
		ArgoCDAuthToken:        getEnv("ARGOCD_AUTH_TOKEN", ""),
		ArgoCDInsecureTLS:      getBoolEnv("ARGOCD_INSECURE_TLS", false),
		ArgoCDCAFile:           getEnv("ARGOCD_CA_FILE", ""),
		ArgoCDTimeoutSeconds:   getIntEnv("ARGOCD_TIMEOUT_SECONDS", 30),
		ArgoCDPollIntervalSecs: getIntEnv("ARGOCD_POLL_INTERVAL_SECONDS", 5),

		GitLabBaseURL:        getEnv("GITLAB_BASE_URL", ""),
		GitLabToken:          getEnv("GITLAB_TOKEN", ""),
		GitLabProjectID:      getIntEnv("GITLAB_PROJECT_ID", 0),
		GitLabDefaultBranch:  getEnv("GITLAB_DEFAULT_BRANCH", "main"),
		GitLabTimeoutSeconds: getIntEnv("GITLAB_TIMEOUT_SECONDS", 30),
		GitLabInsecureTLS:    getBoolEnv("GITLAB_INSECURE_TLS", false),
		GitLabCAFile:         getEnv("GITLAB_CA_FILE", ""),

		TektonNamespace:           getEnv("TEKTON_NAMESPACE", "devops-ci-demo"),
		TektonPipelineName:        getEnv("TEKTON_PIPELINE_NAME", "go-build-and-push"),
		TektonServiceAccount:      getEnv("TEKTON_SERVICE_ACCOUNT", "pipeline"),
		TektonTimeoutSeconds:      getIntEnv("TEKTON_TIMEOUT_SECONDS", 900),
		TektonPollIntervalSecs:    getIntEnv("TEKTON_POLL_INTERVAL_SECONDS", 5),
		TektonGitURL:              getEnv("TEKTON_GIT_URL", "https://github.com/vincmarz/demo-go-color-app.git"),
		TektonGitRevision:         getEnv("TEKTON_GIT_REVISION", "main"),
		TektonGitRevisionTemplate: getEnv("TEKTON_GIT_REVISION_TEMPLATE", ""),
		TektonValidationPath:      getEnv("TEKTON_VALIDATION_PATH", ""),
		TektonImage:               getEnv("TEKTON_IMAGE", "image-registry.openshift-image-registry.svc:5000/devops-ci-demo/demo-go-color-app:latest"),
		TektonWorkspacePVC:        getEnv("TEKTON_WORKSPACE_PVC", "pipeline-workspace"),
		TektonDockerConfigSecret:  getEnv("TEKTON_DOCKERCONFIG_SECRET", "pipeline-registry-dockerconfig"),

		KubernetesAPIURL:      getEnv("KUBERNETES_API_URL", ""),
		KubernetesToken:       getEnv("KUBERNETES_TOKEN", ""),
		KubernetesInsecureTLS: getBoolEnv("KUBERNETES_INSECURE_TLS", false),
		KubernetesCAFile:      getEnv("KUBERNETES_CA_FILE", ""),
		KubernetesNamespace:   getEnv("KUBERNETES_NAMESPACE", "devops-ci-demo"),
		EvidenceBasePath:      getEnv("EVIDENCE_BASE_PATH", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getBoolEnv(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
