package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultServiceAccountTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultServiceAccountCAFile    = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
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

	GitHubAPIURL         string
	GitHubToken          string
	GitHubTimeoutSeconds int
	GitHubInsecureTLS    bool
	GitHubCAFile         string

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

	RuntimeSecretLoaderEnabled            bool
	RuntimeClientFactoriesEnabled         bool
	RuntimeClientFactoryKubernetesEnabled bool
	RuntimeClientFactoryTektonEnabled     bool
	RuntimeClientFactoryArgoCDEnabled     bool
}

func Load() Config {
	kubernetesAPIURL := kubernetesAPIURLFromEnv()
	kubernetesToken := kubernetesTokenFromEnvOrServiceAccount()
	kubernetesCAFile := kubernetesCAFileFromEnvOrServiceAccount()

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

		GitHubAPIURL:         getEnv("GITHUB_API_URL", "https://api.github.com"),
		GitHubToken:          getEnv("GITHUB_TOKEN", ""),
		GitHubTimeoutSeconds: getIntEnv("GITHUB_TIMEOUT_SECONDS", 30),
		GitHubInsecureTLS:    getBoolEnv("GITHUB_INSECURE_TLS", false),
		GitHubCAFile:         getEnv("GITHUB_CA_FILE", ""),

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

		KubernetesAPIURL:      kubernetesAPIURL,
		KubernetesToken:       kubernetesToken,
		KubernetesInsecureTLS: getBoolEnv("KUBERNETES_INSECURE_TLS", false),
		KubernetesCAFile:      kubernetesCAFile,
		KubernetesNamespace:   getEnv("KUBERNETES_NAMESPACE", "devops-ci-demo"),
		EvidenceBasePath:      getEnv("EVIDENCE_BASE_PATH", ""),

		RuntimeSecretLoaderEnabled:            getBoolEnv("DCP_RUNTIME_SECRET_LOADER_ENABLED", false),
		RuntimeClientFactoriesEnabled:         getBoolEnv("DCP_RUNTIME_CLIENT_FACTORIES_ENABLED", false),
		RuntimeClientFactoryKubernetesEnabled: getBoolEnv("DCP_RUNTIME_CLIENT_FACTORY_KUBERNETES_ENABLED", false),
		RuntimeClientFactoryTektonEnabled:     getBoolEnv("DCP_RUNTIME_CLIENT_FACTORY_TEKTON_ENABLED", false),
		RuntimeClientFactoryArgoCDEnabled:     getBoolEnv("DCP_RUNTIME_CLIENT_FACTORY_ARGOCD_ENABLED", false),
	}
}

func kubernetesAPIURLFromEnv() string {
	if value := strings.TrimSpace(os.Getenv("KUBERNETES_API_URL")); value != "" {
		return value
	}

	host := strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_HOST"))
	if host == "" {
		return ""
	}

	port := strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_PORT"))
	if port == "" {
		port = "443"
	}

	return "https://" + host + ":" + port
}

func kubernetesTokenFromEnvOrServiceAccount() string {
	if value := strings.TrimSpace(os.Getenv("KUBERNETES_TOKEN")); value != "" {
		return value
	}

	tokenFile := getEnv("KUBERNETES_TOKEN_FILE", defaultServiceAccountTokenFile)
	return readTrimmedFileIfPresent(tokenFile)
}

func kubernetesCAFileFromEnvOrServiceAccount() string {
	if value := strings.TrimSpace(os.Getenv("KUBERNETES_CA_FILE")); value != "" {
		return value
	}

	caFile := getEnv("KUBERNETES_SERVICEACCOUNT_CA_FILE", defaultServiceAccountCAFile)
	if fileExists(caFile) {
		return caFile
	}

	return ""
}

func readTrimmedFileIfPresent(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(content))
}

func fileExists(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}

	info, err := os.Stat(filepath.Clean(path))
	return err == nil && !info.IsDir()
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
