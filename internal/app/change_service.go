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
}

// TektonCheckValidationFunc rappresenta la porta applicativa minima per leggere lo stato di una PipelineRun Tekton.
type TektonCheckValidationFunc func(ctx context.Context, change domain.ChangeRequest) (TektonValidationResult, error)

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

	tektonRunPipeline     TektonRunPipelineFunc
	tektonCheckValidation TektonCheckValidationFunc

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

	if req.TargetEnvironment == "" {
		req.TargetEnvironment = "dev"
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
	result["tekton"] = map[string]any{"pipelineRunName": validation.PipelineRunName, "namespace": validation.Namespace, "uid": validation.UID, "status": validation.Status, "reason": validation.Reason, "message": validation.Message}
	return result, nil
}

// CreateBranch esegue lo step tecnico GitLab create-branch e poi marca
// la ChangeRequest con runtime_status BranchCreated.
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
