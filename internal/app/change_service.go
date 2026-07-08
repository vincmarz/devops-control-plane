package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type ChangeStore interface {
	List(ctx context.Context) ([]domain.ChangeRequest, error)
	Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error)
	Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error)
	Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error)
	TransitionLifecycle(ctx context.Context, idOrNumber string, action string, actor string, message string) (map[string]any, error)
	MarkStep(ctx context.Context, idOrNumber string, status string) (map[string]any, error)
}

// GitCreateBranchFunc rappresenta la porta applicativa minima per creare un branch Git.
//
// Nota architetturale:
// Il package app non importa l'adapter GitLab concreto. Il main fa da composition root
// e passa una funzione che incapsula l'adapter reale.
type GitCreateBranchFunc func(ctx context.Context, projectID int, branch string, ref string) error

// TektonRunPipelineFunc rappresenta la porta applicativa minima per avviare una PipelineRun Tekton.
type TektonRunPipelineFunc func(ctx context.Context, change domain.ChangeRequest) (string, string, string, error)

type TektonValidationResult struct {
	PipelineRunName string
	Namespace       string
	UID             string
	Status          string
	Reason          string
	Message         string
	PipelineName    string
	GitURL          string
	GitRevision     string
	ValidationPath  string
	TaskRuns        []TektonTaskRunResult
}

type TektonTaskRunResult struct {
	Name             string
	Namespace        string
	PipelineTaskName string
	TaskName         string
	Status           string
	Reason           string
	Message          string
	StartTime        string
	CompletionTime   string
}

// TektonCheckValidationFunc rappresenta la porta applicativa minima per leggere lo stato di una PipelineRun Tekton.
type TektonCheckValidationFunc func(ctx context.Context, change domain.ChangeRequest) (TektonValidationResult, error)

type ArgoCDDeploymentResult struct {
	ApplicationName string
	Project         string
	SyncStatus      string
	HealthStatus    string
	Revision        string
	RuntimeStatus   string
}

// ArgoCDCheckDeploymentFunc rappresenta la porta applicativa minima per leggere lo stato deployment da Argo CD.
type ArgoCDCheckDeploymentFunc func(ctx context.Context, change domain.ChangeRequest) (ArgoCDDeploymentResult, error)

type EvidenceStore interface {
	Create(ctx context.Context, changeID string, evidence domain.Evidence) (domain.Evidence, error)
}

// DeploymentEvidenceCollectorFunc rappresenta la porta applicativa minima per raccogliere evidenze post-deployment.
type DeploymentEvidenceCollectorFunc func(ctx context.Context, change domain.ChangeRequest) (domain.Evidence, error)

// KubernetesRuntimeEvidenceCollectorFunc rappresenta la porta applicativa minima per raccogliere evidenze runtime Kubernetes/OpenShift.
type KubernetesRuntimeEvidenceCollectorFunc func(ctx context.Context, change domain.ChangeRequest) (map[string]any, error)

// TechnicalRuntimeTargetResolverFunc represents the application-level runtime target
// resolution hook used before technical workflow execution.
//
// The resolver is intentionally optional in this phase so tests and deployments
// without runtime target selection preserve the previous behavior.
type TechnicalRuntimeTargetResolverFunc func(ctx context.Context, change domain.ChangeRequest) (TechnicalRuntimeTarget, error)

// RuntimeClientProviderSelectorFunc represents the application-level provider
// selection hook used after resolving a TechnicalRuntimeTarget.
//
// The selector remains metadata-only in this phase. It prepares the workflow for
// future concrete Kubernetes, Tekton and Argo CD client selection.
type RuntimeClientProviderSelectorFunc func(ctx context.Context, target TechnicalRuntimeTarget) (RuntimeClientProviderSelection, error)

// GitCreateOrUpdateFileFunc rappresenta la porta applicativa minima per creare o aggiornare un file Git.
type GitCreateOrUpdateFileFunc func(ctx context.Context, projectID int, branch string, filePath string, commitMessage string, content string) error

// GitOpenMergeRequestFunc rappresenta la porta applicativa minima per aprire una Merge Request Git.
type GitOpenMergeRequestFunc func(ctx context.Context, projectID int, sourceBranch string, targetBranch string, title string, description string) (int, string, error)

// GitMergeRequestFunc rappresenta la porta applicativa minima per mergiare una Merge Request Git aperta.
type GitMergeRequestFunc func(ctx context.Context, projectID int, sourceBranch string, targetBranch string, mergeCommitMessage string) (int, string, string, error)

type ChangeServiceOption func(*ChangeService)

func WithTektonRunPipeline(fn TektonRunPipelineFunc) ChangeServiceOption {
	return func(s *ChangeService) { s.tektonRunPipeline = fn }
}

func WithTektonCheckValidation(fn TektonCheckValidationFunc) ChangeServiceOption {
	return func(s *ChangeService) { s.tektonCheckValidation = fn }
}

func WithArgoCDCheckDeployment(fn ArgoCDCheckDeploymentFunc) ChangeServiceOption {
	return func(s *ChangeService) { s.argocdCheckDeployment = fn }
}

func WithEvidenceStore(store EvidenceStore) ChangeServiceOption {
	return func(s *ChangeService) { s.evidenceStore = store }
}

func WithDeploymentEvidenceCollector(fn DeploymentEvidenceCollectorFunc) ChangeServiceOption {
	return func(s *ChangeService) { s.deploymentEvidenceCollector = fn }
}

func WithKubernetesRuntimeEvidenceCollector(fn KubernetesRuntimeEvidenceCollectorFunc) ChangeServiceOption {
	return func(s *ChangeService) { s.kubernetesRuntimeEvidenceCollector = fn }
}

// WithTechnicalRuntimeTargetResolver wires the technical runtime target resolver used
// by technical workflows before calling concrete runtime adapters.

// KubernetesRuntimeClientProviderResolver is the minimal runtime client
// resolution contract used by ChangeService. It is satisfied by both the
// default Kubernetes runtime client provider registry and the factory-aware
// registry wrapper prepared for real multi-cluster clients.
type KubernetesRuntimeClientProviderResolver interface {
	Resolve(context.Context, RuntimeClientProviderSelection) (KubernetesRuntimeEvidenceClient, error)
}

// WithKubernetesRuntimeClientProviderRegistry wires the Kubernetes runtime
// client provider registry used by collect-evidence preparation.
func WithKubernetesRuntimeClientProviderRegistry(registry KubernetesRuntimeClientProviderResolver) ChangeServiceOption {
	return func(s *ChangeService) { s.kubernetesRuntimeClientProviderRegistry = registry }
}

func WithTechnicalRuntimeTargetResolver(resolver TechnicalRuntimeTargetResolver) ChangeServiceOption {
	return func(s *ChangeService) {
		s.technicalRuntimeTargetResolver = resolver.ResolveChange
	}
}

// WithTechnicalRuntimeTargetResolverFunc wires a custom technical runtime target
// resolver. Tests can use this option to verify workflow behavior without
// depending on runtime catalog files.
func WithTechnicalRuntimeTargetResolverFunc(fn TechnicalRuntimeTargetResolverFunc) ChangeServiceOption {
	return func(s *ChangeService) { s.technicalRuntimeTargetResolver = fn }
}

// WithRuntimeClientProviderRegistry wires the runtime provider registry used to
// select the provider associated with a resolved TechnicalRuntimeTarget.
func WithRuntimeClientProviderRegistry(registry RuntimeClientProviderRegistry) ChangeServiceOption {
	return func(s *ChangeService) {
		s.runtimeClientProviderSelector = func(ctx context.Context, target TechnicalRuntimeTarget) (RuntimeClientProviderSelection, error) {
			return registry.Select(target)
		}
	}
}

// WithRuntimeClientProviderSelectorFunc wires a custom provider selection hook.
// Tests can use this option to validate workflow preparation without relying on
// runtime configuration files.
func WithRuntimeClientProviderSelectorFunc(fn RuntimeClientProviderSelectorFunc) ChangeServiceOption {
	return func(s *ChangeService) { s.runtimeClientProviderSelector = fn }
}

// WithRuntimeClientSecretRefsRegistry wires the optional Secret reference
// registry used to enrich RuntimeClientProviderSelection metadata.
//
// The registry contains references only. No Secret values are read by this
// option.
func WithRuntimeClientSecretRefsRegistry(registry RuntimeClientSecretRefsRegistry) ChangeServiceOption {
	return func(s *ChangeService) { s.runtimeClientSecretRefsRegistry = registry }
}

func WithGitCreateBranch(fn GitCreateBranchFunc, projectID int, defaultBranch string) ChangeServiceOption {
	return func(s *ChangeService) {
		s.gitCreateBranch = fn
		s.gitProjectID = projectID
		s.gitDefaultBranch = strings.TrimSpace(defaultBranch)
	}
}

func WithGitCreateOrUpdateFile(fn GitCreateOrUpdateFileFunc) ChangeServiceOption {
	return func(s *ChangeService) {
		s.gitCreateOrUpdateFile = fn
	}
}

func WithGitOpenMergeRequest(fn GitOpenMergeRequestFunc) ChangeServiceOption {
	return func(s *ChangeService) {
		s.gitOpenMergeRequest = fn
	}
}

func WithGitMergeRequest(fn GitMergeRequestFunc) ChangeServiceOption {
	return func(s *ChangeService) {
		s.gitMergeRequest = fn
	}
}

type ChangeService struct {
	store ChangeStore

	tektonRunPipeline                       TektonRunPipelineFunc
	tektonCheckValidation                   TektonCheckValidationFunc
	argocdCheckDeployment                   ArgoCDCheckDeploymentFunc
	evidenceStore                           EvidenceStore
	deploymentEvidenceCollector             DeploymentEvidenceCollectorFunc
	kubernetesRuntimeEvidenceCollector      KubernetesRuntimeEvidenceCollectorFunc
	kubernetesRuntimeClientProviderRegistry KubernetesRuntimeClientProviderResolver
	technicalRuntimeTargetResolver          TechnicalRuntimeTargetResolverFunc
	runtimeClientProviderSelector           RuntimeClientProviderSelectorFunc
	runtimeClientSecretRefsRegistry         RuntimeClientSecretRefsRegistry

	gitCreateBranch       GitCreateBranchFunc
	gitCreateOrUpdateFile GitCreateOrUpdateFileFunc
	gitOpenMergeRequest   GitOpenMergeRequestFunc
	gitMergeRequest       GitMergeRequestFunc
	gitProjectID          int
	gitDefaultBranch      string
}

func NewChangeService(store ChangeStore, opts ...ChangeServiceOption) *ChangeService {
	service := &ChangeService{
		store:            store,
		gitDefaultBranch: "main",
	}
	for _, opt := range opts {
		opt(service)
	}
	if service.gitDefaultBranch == "" {
		service.gitDefaultBranch = "main"
	}
	return service
}

func (s *ChangeService) resolveTechnicalRuntimeTarget(ctx context.Context, change domain.ChangeRequest) (TechnicalRuntimeTarget, error) {
	if s.technicalRuntimeTargetResolver == nil {
		return TechnicalRuntimeTarget{}, nil
	}
	return s.technicalRuntimeTargetResolver(ctx, change)
}

func (s *ChangeService) resolveRuntimeClientProviderSelection(ctx context.Context, change domain.ChangeRequest) (RuntimeClientProviderSelection, error) {
	target, err := s.resolveTechnicalRuntimeTarget(ctx, change)
	if err != nil {
		return RuntimeClientProviderSelection{}, err
	}

	var selection RuntimeClientProviderSelection
	if s.runtimeClientProviderSelector == nil {
		selection = RuntimeClientProviderSelection{Target: target}
	} else {
		selection, err = s.runtimeClientProviderSelector(ctx, target)
		if err != nil {
			return RuntimeClientProviderSelection{}, err
		}
	}

	refs, ok := s.runtimeClientSecretRefsRegistry.Resolve(selection.Provider.ClusterName)
	if ok {
		selection.SecretRefsConfigured = true
		selection.SecretRefs = refs
	}

	return selection, nil
}

func (s *ChangeService) collectKubernetesRuntimeEvidence(ctx context.Context, change domain.ChangeRequest, selection RuntimeClientProviderSelection) (map[string]any, error) {
	if s.kubernetesRuntimeClientProviderRegistry != nil {
		client, err := s.kubernetesRuntimeClientProviderRegistry.Resolve(ctx, selection)
		if err != nil {
			return nil, err
		}
		namespace := strings.TrimSpace(selection.Target.KubernetesNamespace)
		if namespace == "" {
			return nil, fmt.Errorf("technical runtime target for change %s does not define kubernetesNamespace", change.ChangeNumber)
		}
		return client.CollectRuntimeEvidence(ctx, namespace, change.ApplicationName)
	}

	if s.kubernetesRuntimeEvidenceCollector == nil {
		return nil, nil
	}
	return s.kubernetesRuntimeEvidenceCollector(ctx, change)
}

func (s *ChangeService) List(ctx context.Context) ([]domain.ChangeRequest, error) {
	return s.store.List(ctx)
}

func (s *ChangeService) Create(ctx context.Context, req domain.CreateChangeRequest) (domain.ChangeRequest, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.ApplicationName = strings.TrimSpace(req.ApplicationName)
	req.TargetEnvironment = strings.TrimSpace(req.TargetEnvironment)
	req.ChangeType = strings.TrimSpace(req.ChangeType)
	req.RiskLevel = strings.TrimSpace(req.RiskLevel)
	req.RequestedBy = strings.TrimSpace(req.RequestedBy)

	if req.Title == "" {
		return domain.ChangeRequest{}, errors.New("title is required")
	}
	if req.ApplicationName == "" {
		return domain.ChangeRequest{}, errors.New("applicationName is required")
	}
	if req.ChangeType == "" {
		return domain.ChangeRequest{}, errors.New("changeType is required")
	}
	if req.RequestedBy == "" {
		return domain.ChangeRequest{}, errors.New("requestedBy is required")
	}

	environmentCatalog := DefaultEnvironmentCatalog()
	if req.TargetEnvironment == "" {
		req.TargetEnvironment = environmentCatalog.DefaultEnvironment()
	}
	if err := environmentCatalog.ValidateCreateTargetEnvironment(req.TargetEnvironment); err != nil {
		return domain.ChangeRequest{}, err
	}

	if _, err := DefaultEnvironmentClusterResolver().ResolveEnabledTarget(req.TargetEnvironment); err != nil {
		return domain.ChangeRequest{}, err
	}
	if req.RiskLevel == "" {
		req.RiskLevel = domain.RiskMedium
	}

	if !isAllowedRiskLevel(req.RiskLevel) {
		return domain.ChangeRequest{}, errors.New("riskLevel must be one of: low, medium, high, critical")
	}

	return s.store.Create(ctx, req)
}

func (s *ChangeService) Get(ctx context.Context, idOrNumber string) (domain.ChangeRequest, error) {
	return s.store.Get(ctx, idOrNumber)
}

func (s *ChangeService) Events(ctx context.Context, idOrNumber string) ([]domain.ChangeEvent, error) {
	return s.store.Events(ctx, idOrNumber)
}

// TransitionLifecycle applica una transizione governata della ChangeRequest.
//
// Nota importante:
// Questo metodo lavora sul lifecycle status della change, ad esempio:
// draft -> submitted -> approved -> executing -> executed -> closed.
//
// Non deve essere usato per step tecnici come create-branch, update-files o sync,
// che restano tracciati tramite MarkStep/runtime_status.
func (s *ChangeService) TransitionLifecycle(ctx context.Context, idOrNumber string, action string, actor string, message string) (map[string]any, error) {
	idOrNumber = strings.TrimSpace(idOrNumber)
	action = strings.TrimSpace(action)
	actor = strings.TrimSpace(actor)
	message = strings.TrimSpace(message)

	if idOrNumber == "" {
		return nil, errors.New("change id or number is required")
	}
	if action == "" {
		return nil, errors.New("lifecycle action is required")
	}
	if actor == "" {
		return nil, errors.New("actor is required")
	}

	return s.store.TransitionLifecycle(ctx, idOrNumber, action, actor, message)
}

// Validate esegue lo step tecnico Tekton validate e poi marca
// la ChangeRequest con runtime_status ValidationRunning.
func (s *ChangeService) Validate(ctx context.Context, idOrNumber string) (map[string]any, error) {
	idOrNumber = strings.TrimSpace(idOrNumber)
	if idOrNumber == "" {
		return nil, errors.New("change id or number is required")
	}
	if s.tektonRunPipeline == nil {
		return nil, errors.New("tekton run pipeline client is not configured")
	}
	change, err := s.store.Get(ctx, idOrNumber)
	if err != nil {
		return nil, err
	}
	if _, err := s.resolveRuntimeClientProviderSelection(ctx, change); err != nil {
		return nil, err
	}

	pipelineRunName, namespace, uid, err := s.tektonRunPipeline(ctx, change)
	if err != nil {
		return nil, fmt.Errorf("run tekton validation pipeline for %q: %w", change.ChangeNumber, err)
	}
	result, err := s.store.MarkStep(ctx, idOrNumber, "ValidationRunning")
	if err != nil {
		return nil, err
	}
	result["tekton"] = map[string]any{"pipelineRunName": pipelineRunName, "namespace": namespace, "uid": uid}
	return result, nil
}

// CheckValidation legge lo stato reale della PipelineRun Tekton piu recente e aggiorna il runtime_status.
func (s *ChangeService) CheckValidation(ctx context.Context, idOrNumber string) (map[string]any, error) {
	idOrNumber = strings.TrimSpace(idOrNumber)
	if idOrNumber == "" {
		return nil, errors.New("change id or number is required")
	}
	if s.tektonCheckValidation == nil {
		return nil, errors.New("tekton check validation client is not configured")
	}
	change, err := s.store.Get(ctx, idOrNumber)
	if err != nil {
		return nil, err
	}
	if _, err := s.resolveRuntimeClientProviderSelection(ctx, change); err != nil {
		return nil, err
	}

	validation, err := s.tektonCheckValidation(ctx, change)
	if err != nil {
		return nil, fmt.Errorf("check tekton validation pipeline for %q: %w", change.ChangeNumber, err)
	}
	runtimeStatus := "ValidationRunning"
	if validation.Status == "True" {
		runtimeStatus = "ValidationSucceeded"
	} else if validation.Status == "False" {
		runtimeStatus = "ValidationFailed"
	}
	result, err := s.store.MarkStep(ctx, idOrNumber, runtimeStatus)
	if err != nil {
		return nil, err
	}
	tektonPayload := map[string]any{"pipelineRunName": validation.PipelineRunName, "namespace": validation.Namespace, "uid": validation.UID, "status": validation.Status, "reason": validation.Reason, "message": validation.Message}
	taskRunsPayload, diagnosticsPayload := tektonTaskRunDiagnostics(validation.TaskRuns)
	result["tekton"] = tektonPayload
	result["taskRuns"] = taskRunsPayload
	result["diagnostics"] = diagnosticsPayload

	if s.evidenceStore != nil && (runtimeStatus == "ValidationSucceeded" || runtimeStatus == "ValidationFailed") {
		evidence := domain.Evidence{
			ChangeNumber: change.ChangeNumber,
			EvidenceType: "validation",
			Name:         "tekton-validation-evidence",
			Summary:      "Tekton GitOps validation result for " + change.ChangeNumber,
			Sanitized:    true,
			ExternalRef:  validation.PipelineRunName,
			Payload: map[string]any{
				"change": map[string]any{
					"changeNumber":      change.ChangeNumber,
					"applicationName":   change.ApplicationName,
					"targetEnvironment": change.TargetEnvironment,
					"lifecycleStatus":   change.Status,
				},
				"tekton":      tektonPayload,
				"taskRuns":    taskRunsPayload,
				"diagnostics": diagnosticsPayload,
				"gitops": map[string]any{
					"pipelineName":    validation.PipelineName,
					"tektonNamespace": validation.Namespace,
					"repoURL":         validation.GitURL,
					"revision":        validation.GitRevision,
					"validationPath":  validation.ValidationPath,
				},
			},
		}
		savedEvidence, err := s.evidenceStore.Create(ctx, change.ID, evidence)
		if err != nil {
			return nil, err
		}
		result["evidence"] = map[string]any{
			"id":           savedEvidence.ID,
			"evidenceType": savedEvidence.EvidenceType,
			"name":         savedEvidence.Name,
			"summary":      savedEvidence.Summary,
			"sanitized":    savedEvidence.Sanitized,
			"createdAt":    savedEvidence.CreatedAt,
		}
	}
	return result, nil
}

func tektonTaskRunDiagnostics(taskRuns []TektonTaskRunResult) ([]map[string]any, map[string]any) {
	taskRunsPayload := make([]map[string]any, 0, len(taskRuns))
	failedTasks := make([]string, 0)
	for _, taskRun := range taskRuns {
		pipelineTaskName := strings.TrimSpace(taskRun.PipelineTaskName)
		if pipelineTaskName == "" {
			pipelineTaskName = strings.TrimSpace(taskRun.TaskName)
		}
		if pipelineTaskName == "" {
			pipelineTaskName = strings.TrimSpace(taskRun.Name)
		}
		taskRunsPayload = append(taskRunsPayload, map[string]any{
			"name":             taskRun.Name,
			"namespace":        taskRun.Namespace,
			"pipelineTaskName": pipelineTaskName,
			"taskName":         taskRun.TaskName,
			"status":           taskRun.Status,
			"reason":           taskRun.Reason,
			"message":          taskRun.Message,
			"startTime":        taskRun.StartTime,
			"completionTime":   taskRun.CompletionTime,
		})
		if taskRun.Status == "False" {
			failedTasks = append(failedTasks, pipelineTaskName)
		}
	}

	summary := "Tekton validation completed without failed TaskRuns"
	if len(failedTasks) == 1 {
		summary = fmt.Sprintf("Tekton validation failed in task %s", failedTasks[0])
	} else if len(failedTasks) > 1 {
		summary = fmt.Sprintf("Tekton validation failed in %d tasks: %s", len(failedTasks), strings.Join(failedTasks, ", "))
	}

	return taskRunsPayload, map[string]any{
		"failedTaskCount": len(failedTasks),
		"failedTasks":     failedTasks,
		"summary":         summary,
	}
}

// CheckDeployment legge lo stato reale della application Argo CD e aggiorna il runtime_status.
func (s *ChangeService) CheckDeployment(ctx context.Context, idOrNumber string) (map[string]any, error) {
	idOrNumber = strings.TrimSpace(idOrNumber)
	if idOrNumber == "" {
		return nil, errors.New("change id or number is required")
	}
	if s.argocdCheckDeployment == nil {
		return nil, errors.New("argocd check deployment client is not configured")
	}

	change, err := s.store.Get(ctx, idOrNumber)
	if err != nil {
		return nil, err
	}
	if _, err := s.resolveRuntimeClientProviderSelection(ctx, change); err != nil {
		return nil, err
	}

	deployment, err := s.argocdCheckDeployment(ctx, change)
	if err != nil {
		return nil, fmt.Errorf("check argocd deployment for %q: %w", change.ChangeNumber, err)
	}

	runtimeStatus := mapArgoCDDeploymentRuntimeStatus(deployment.SyncStatus, deployment.HealthStatus)
	deployment.RuntimeStatus = runtimeStatus

	result, err := s.store.MarkStep(ctx, idOrNumber, runtimeStatus)
	if err != nil {
		return nil, err
	}
	result["argocd"] = map[string]any{
		"applicationName": deployment.ApplicationName,
		"project":         deployment.Project,
		"syncStatus":      deployment.SyncStatus,
		"healthStatus":    deployment.HealthStatus,
		"revision":        deployment.Revision,
	}
	return result, nil
}

func mapArgoCDDeploymentRuntimeStatus(syncStatus string, healthStatus string) string {
	syncStatus = strings.TrimSpace(syncStatus)
	healthStatus = strings.TrimSpace(healthStatus)

	if healthStatus == "Degraded" {
		return "DeploymentDegraded"
	}
	if syncStatus == "OutOfSync" {
		return "DeploymentOutOfSync"
	}
	if syncStatus == "Synced" && healthStatus == "Healthy" {
		return "DeploymentSyncedHealthy"
	}
	if syncStatus == "Synced" && healthStatus == "Progressing" {
		return "DeploymentProgressing"
	}
	return "DeploymentUnknown"
}

func (s *ChangeService) CreateBranch(ctx context.Context, idOrNumber string) (map[string]any, error) {
	idOrNumber = strings.TrimSpace(idOrNumber)
	if idOrNumber == "" {
		return nil, errors.New("change id or number is required")
	}
	if s.gitCreateBranch == nil {
		return nil, errors.New("git create branch client is not configured")
	}
	if s.gitProjectID <= 0 {
		return nil, errors.New("git project ID must be configured")
	}

	change, err := s.store.Get(ctx, idOrNumber)
	if err != nil {
		return nil, err
	}

	branchName := fmt.Sprintf("change/%s", change.ChangeNumber)
	ref := s.gitDefaultBranch
	if strings.TrimSpace(ref) == "" {
		ref = "main"
	}

	if err := s.gitCreateBranch(ctx, s.gitProjectID, branchName, ref); err != nil {
		return nil, fmt.Errorf("create git branch %q from ref %q: %w", branchName, ref, err)
	}

	result, err := s.store.MarkStep(ctx, idOrNumber, "BranchCreated")
	if err != nil {
		return nil, err
	}
	result["git"] = map[string]any{
		"projectID": s.gitProjectID,
		"branch":    branchName,
		"ref":       ref,
	}
	return result, nil
}

// UpdateFiles esegue lo step tecnico GitLab update-files e poi marca
// la ChangeRequest con runtime_status CommitCreated.
func (s *ChangeService) UpdateFiles(ctx context.Context, idOrNumber string) (map[string]any, error) {
	idOrNumber = strings.TrimSpace(idOrNumber)
	if idOrNumber == "" {
		return nil, errors.New("change id or number is required")
	}
	if s.gitCreateOrUpdateFile == nil {
		return nil, errors.New("git create or update file client is not configured")
	}
	if s.gitProjectID <= 0 {
		return nil, errors.New("git project ID must be configured")
	}

	change, err := s.store.Get(ctx, idOrNumber)
	if err != nil {
		return nil, err
	}

	branchName := fmt.Sprintf("change/%s", change.ChangeNumber)
	filePath := fmt.Sprintf("manifests/%s-control-plane.yaml", strings.ToLower(change.ChangeNumber))
	commitMessage := fmt.Sprintf("Add generated manifest for %s", change.ChangeNumber)
	content := generatedChangeConfigMap(change)

	if err := s.gitCreateOrUpdateFile(ctx, s.gitProjectID, branchName, filePath, commitMessage, content); err != nil {
		return nil, fmt.Errorf("create or update git file %q on branch %q: %w", filePath, branchName, err)
	}

	result, err := s.store.MarkStep(ctx, idOrNumber, "CommitCreated")
	if err != nil {
		return nil, err
	}
	result["git"] = map[string]any{
		"projectID": s.gitProjectID,
		"branch":    branchName,
		"filePath":  filePath,
	}
	return result, nil
}

func generatedChangeConfigMap(change domain.ChangeRequest) string {
	return fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s-control-plane
data:
  changeNumber: %s
  applicationName: %s
  targetEnvironment: %s
  managedBy: devops-control-plane
`, strings.ToLower(change.ChangeNumber), change.ChangeNumber, change.ApplicationName, change.TargetEnvironment)
}

// OpenMergeRequest esegue lo step tecnico GitLab open-merge-request e poi marca
// la ChangeRequest con runtime_status MergeRequestOpened.
func (s *ChangeService) OpenMergeRequest(ctx context.Context, idOrNumber string) (map[string]any, error) {
	idOrNumber = strings.TrimSpace(idOrNumber)
	if idOrNumber == "" {
		return nil, errors.New("change id or number is required")
	}
	if s.gitOpenMergeRequest == nil {
		return nil, errors.New("git open merge request client is not configured")
	}
	if s.gitProjectID <= 0 {
		return nil, errors.New("git project ID must be configured")
	}

	change, err := s.store.Get(ctx, idOrNumber)
	if err != nil {
		return nil, err
	}

	sourceBranch := fmt.Sprintf("change/%s", change.ChangeNumber)
	targetBranch := s.gitDefaultBranch
	if strings.TrimSpace(targetBranch) == "" {
		targetBranch = "main"
	}
	title := fmt.Sprintf("%s - GitOps change for %s", change.ChangeNumber, change.ApplicationName)
	description := fmt.Sprintf("Merge Request generated by DevOps Control Plane for %s.", change.ChangeNumber)

	mrIID, mrWebURL, err := s.gitOpenMergeRequest(ctx, s.gitProjectID, sourceBranch, targetBranch, title, description)
	if err != nil {
		return nil, fmt.Errorf("open git merge request from %q to %q: %w", sourceBranch, targetBranch, err)
	}

	result, err := s.store.MarkStep(ctx, idOrNumber, "MergeRequestOpened")
	if err != nil {
		return nil, err
	}
	result["git"] = map[string]any{
		"projectID":       s.gitProjectID,
		"sourceBranch":    sourceBranch,
		"targetBranch":    targetBranch,
		"mergeRequestIID": mrIID,
		"mergeRequestURL": mrWebURL,
	}
	return result, nil
}

// MergeRequest esegue lo step tecnico GitLab merge-request e poi marca
// la ChangeRequest con runtime_status MergeRequestMerged.
func (s *ChangeService) MergeRequest(ctx context.Context, idOrNumber string) (map[string]any, error) {
	idOrNumber = strings.TrimSpace(idOrNumber)
	if idOrNumber == "" {
		return nil, errors.New("change id or number is required")
	}
	if s.gitMergeRequest == nil {
		return nil, errors.New("git merge request client is not configured")
	}
	if s.gitProjectID <= 0 {
		return nil, errors.New("git project ID must be configured")
	}

	change, err := s.store.Get(ctx, idOrNumber)
	if err != nil {
		return nil, err
	}

	sourceBranch := fmt.Sprintf("change/%s", change.ChangeNumber)
	targetBranch := s.gitDefaultBranch
	if strings.TrimSpace(targetBranch) == "" {
		targetBranch = "main"
	}
	mergeCommitMessage := fmt.Sprintf("Merge %s via DevOps Control Plane", change.ChangeNumber)

	mrIID, mrWebURL, mergeCommitSHA, err := s.gitMergeRequest(ctx, s.gitProjectID, sourceBranch, targetBranch, mergeCommitMessage)
	if err != nil {
		return nil, fmt.Errorf("merge git merge request from %q to %q: %w", sourceBranch, targetBranch, err)
	}

	result, err := s.store.MarkStep(ctx, idOrNumber, "MergeRequestMerged")
	if err != nil {
		return nil, err
	}
	result["git"] = map[string]any{
		"projectID":       s.gitProjectID,
		"sourceBranch":    sourceBranch,
		"targetBranch":    targetBranch,
		"mergeRequestIID": mrIID,
		"mergeRequestURL": mrWebURL,
		"mergeCommitSHA":  mergeCommitSHA,
	}
	return result, nil
}

// CollectEvidence raccoglie e persiste evidenze post-deployment e marca lo step tecnico EvidenceCollected.
func (s *ChangeService) CollectEvidence(ctx context.Context, idOrNumber string) (map[string]any, error) {
	idOrNumber = strings.TrimSpace(idOrNumber)
	if idOrNumber == "" {
		return nil, errors.New("change id or number is required")
	}
	if s.evidenceStore == nil {
		return nil, errors.New("evidence store is not configured")
	}
	if s.deploymentEvidenceCollector == nil {
		return nil, errors.New("deployment evidence collector is not configured")
	}

	change, err := s.store.Get(ctx, idOrNumber)
	if err != nil {
		return nil, err
	}
	providerSelection, err := s.resolveRuntimeClientProviderSelection(ctx, change)
	if err != nil {
		return nil, err
	}

	evidence, err := s.deploymentEvidenceCollector(ctx, change)
	if err != nil {
		return nil, fmt.Errorf("collect deployment evidence for %q: %w", change.ChangeNumber, err)
	}
	if evidence.EvidenceType == "" {
		evidence.EvidenceType = "deployment"
	}
	if evidence.Name == "" {
		evidence.Name = "deployment-evidence"
	}
	if evidence.ChangeNumber == "" {
		evidence.ChangeNumber = change.ChangeNumber
	}
	evidence.Sanitized = true

	if s.kubernetesRuntimeEvidenceCollector != nil {
		runtimeEvidence, err := s.collectKubernetesRuntimeEvidence(ctx, change, providerSelection)
		if err != nil {
			return nil, fmt.Errorf("collect kubernetes runtime evidence for %q: %w", change.ChangeNumber, err)
		}
		if evidence.Payload == nil {
			evidence.Payload = map[string]any{}
		}
		evidence.Payload["kubernetes"] = runtimeEvidence
	}
	if evidence.Payload == nil {
		evidence.Payload = map[string]any{}
	}
	evidence.Payload["diagnostics"] = deploymentEvidenceDiagnostics(evidence.Payload)

	savedEvidence, err := s.evidenceStore.Create(ctx, change.ID, evidence)
	if err != nil {
		return nil, err
	}

	result, err := s.store.MarkStep(ctx, idOrNumber, "EvidenceCollected")
	if err != nil {
		return nil, err
	}
	result["evidence"] = map[string]any{
		"id":           savedEvidence.ID,
		"evidenceType": savedEvidence.EvidenceType,
		"name":         savedEvidence.Name,
		"summary":      savedEvidence.Summary,
		"sanitized":    savedEvidence.Sanitized,
		"createdAt":    savedEvidence.CreatedAt,
	}
	return result, nil
}

func deploymentEvidenceDiagnostics(payload map[string]any) map[string]any {
	argocd, _ := payload["argocd"].(map[string]any)
	kubernetesPayload, _ := payload["kubernetes"].(map[string]any)

	syncStatus := strings.TrimSpace(stringFromMap(argocd, "syncStatus"))
	healthStatus := strings.TrimSpace(stringFromMap(argocd, "healthStatus"))
	appName := strings.TrimSpace(stringFromMap(argocd, "applicationName"))
	if appName == "" {
		if change, ok := payload["change"].(map[string]any); ok {
			appName = strings.TrimSpace(stringFromMap(change, "applicationName"))
		}
	}
	if appName == "" {
		appName = "application"
	}

	deployment, _ := kubernetesPayload["deployment"].(map[string]any)
	desiredReplicas := intFromMap(deployment, "desiredReplicas")
	readyReplicas := intFromMap(deployment, "readyReplicas")
	availableReplicas := intFromMap(deployment, "availableReplicas")
	updatedReplicas := intFromMap(deployment, "updatedReplicas")
	generation := intFromMap(deployment, "generation")
	observedGeneration := intFromMap(deployment, "observedGeneration")
	deploymentReady := desiredReplicas > 0 && readyReplicas >= desiredReplicas && availableReplicas >= desiredReplicas && updatedReplicas >= desiredReplicas
	generationObserved := generation == 0 || observedGeneration >= generation

	podsRaw, _ := kubernetesPayload["pods"].([]map[string]any)
	if podsRaw == nil {
		if genericPods, ok := kubernetesPayload["pods"].([]any); ok {
			podsRaw = make([]map[string]any, 0, len(genericPods))
			for _, rawPod := range genericPods {
				if pod, ok := rawPod.(map[string]any); ok {
					podsRaw = append(podsRaw, pod)
				}
			}
		}
	}
	podsReady := 0
	totalRestarts := 0
	for _, pod := range podsRaw {
		if boolFromMap(pod, "ready") {
			podsReady++
		}
		totalRestarts += intFromMap(pod, "restartCount")
	}

	route, _ := kubernetesPayload["route"].(map[string]any)
	routeAvailable := strings.TrimSpace(stringFromMap(route, "host")) != "" && strings.TrimSpace(stringFromMap(route, "error")) == ""
	service, _ := kubernetesPayload["service"].(map[string]any)
	serviceAvailable := strings.TrimSpace(stringFromMap(service, "clusterIP")) != "" && strings.TrimSpace(stringFromMap(service, "error")) == ""

	warnings := argocdConditionWarnings(argocd)
	if !generationObserved {
		warnings = append(warnings, "Deployment observedGeneration is behind generation")
	}
	if totalRestarts > 0 {
		warnings = append(warnings, fmt.Sprintf("Pods reported %d total restarts", totalRestarts))
	}
	if !routeAvailable {
		warnings = append(warnings, "Route is not available or has no host")
	}
	if !serviceAvailable {
		warnings = append(warnings, "Service is not available or has no ClusterIP")
	}

	summary := fmt.Sprintf("Application %s status: Argo CD %s/%s, replicas %d/%d ready, pods %d/%d ready", appName, valueOrUnknown(syncStatus), valueOrUnknown(healthStatus), readyReplicas, desiredReplicas, podsReady, len(podsRaw))
	if syncStatus == "Synced" && healthStatus == "Healthy" && deploymentReady {
		summary = fmt.Sprintf("Application %s is Synced/Healthy with %d/%d replicas ready", appName, readyReplicas, desiredReplicas)
	}
	if len(warnings) > 0 {
		summary = fmt.Sprintf("%s; warnings: %d", summary, len(warnings))
	}

	return map[string]any{
		"argocdSynced":       syncStatus == "Synced",
		"argocdHealthy":      healthStatus == "Healthy",
		"deploymentReady":    deploymentReady,
		"generationObserved": generationObserved,
		"readyReplicas":      fmt.Sprintf("%d/%d", readyReplicas, desiredReplicas),
		"availableReplicas":  availableReplicas,
		"updatedReplicas":    updatedReplicas,
		"podsReady":          fmt.Sprintf("%d/%d", podsReady, len(podsRaw)),
		"totalRestarts":      totalRestarts,
		"serviceAvailable":   serviceAvailable,
		"routeAvailable":     routeAvailable,
		"warnings":           warnings,
		"summary":            summary,
	}
}

func argocdConditionWarnings(argocd map[string]any) []string {
	warnings := make([]string, 0)
	conditionsRaw := argocd["conditions"]
	switch conditions := conditionsRaw.(type) {
	case []map[string]any:
		for _, condition := range conditions {
			warnings = appendConditionWarning(warnings, condition)
		}
	case []any:
		for _, rawCondition := range conditions {
			if condition, ok := rawCondition.(map[string]any); ok {
				warnings = appendConditionWarning(warnings, condition)
			}
		}
	}
	return warnings
}

func appendConditionWarning(warnings []string, condition map[string]any) []string {
	conditionType := strings.TrimSpace(stringFromMap(condition, "type"))
	message := strings.TrimSpace(stringFromMap(condition, "message"))
	if conditionType == "" && message == "" {
		return warnings
	}
	if message == "" {
		return append(warnings, conditionType)
	}
	return append(warnings, fmt.Sprintf("%s: %s", conditionType, message))
}

func stringFromMap(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	value, _ := values[key].(string)
	return value
}

func boolFromMap(values map[string]any, key string) bool {
	if values == nil {
		return false
	}
	value, _ := values[key].(bool)
	return value
}

func intFromMap(values map[string]any, key string) int {
	if values == nil {
		return 0
	}
	switch value := values[key].(type) {
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func valueOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "Unknown"
	}
	return value
}

// MarkStep registra uno step tecnico del workflow.
// Nota importante:
// Non deve modificare lo stato governato della ChangeRequest.
// Gli step tecnici come BranchCreated, CommitCreated o SyncRunning devono
// essere tracciati come eventi/runtime status, non come lifecycle status.
func (s *ChangeService) MarkStep(ctx context.Context, idOrNumber string, status string) (map[string]any, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return nil, errors.New("workflow step status is required")
	}
	return s.store.MarkStep(ctx, idOrNumber, status)
}

func isAllowedRiskLevel(riskLevel string) bool {
	switch riskLevel {
	case domain.RiskLow, domain.RiskMedium, domain.RiskHigh, domain.RiskCritical:
		return true
	default:
		return false
	}
}
