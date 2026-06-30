package tekton

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTimeoutSeconds = 30

type Config struct {
	APIURL         string
	Token          string
	TimeoutSeconds int
	InsecureTLS    bool
}
type Client struct {
	apiURL     string
	token      string
	httpClient *http.Client
}
type Option func(*Client)

func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.httpClient = h
		}
	}
}
func New(cfg Config, opts ...Option) (*Client, error) {
	apiURL := strings.TrimRight(strings.TrimSpace(cfg.APIURL), "/")
	if apiURL == "" {
		return nil, errors.New("kubernetes API URL is required")
	}
	token := strings.TrimSpace(cfg.Token)
	if token == "" {
		return nil, errors.New("kubernetes bearer token is required")
	}
	to := cfg.TimeoutSeconds
	if to <= 0 {
		to = defaultTimeoutSeconds
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureTLS {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	c := &Client{apiURL: apiURL, token: token, httpClient: &http.Client{Timeout: time.Duration(to) * time.Second, Transport: tr}}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.do(ctx, http.MethodGet, "/apis/tekton.dev/v1", nil)
	return err
}

type CreatePipelineRunRequest struct {
	Namespace          string
	PipelineName       string
	ServiceAccountName string
	GenerateName       string
	ChangeNumber       string
	GitURL             string
	GitRevision        string
	ValidationPath     string
	Image              string
	WorkspacePVC       string
	DockerConfigSecret string
}

func (c *Client) CreatePipelineRun(ctx context.Context, r CreatePipelineRunRequest) (PipelineRunRef, error) {
	r.Namespace = strings.TrimSpace(r.Namespace)
	r.PipelineName = strings.TrimSpace(r.PipelineName)
	r.ServiceAccountName = strings.TrimSpace(r.ServiceAccountName)
	r.GenerateName = strings.TrimSpace(r.GenerateName)
	r.ChangeNumber = strings.TrimSpace(r.ChangeNumber)
	r.GitURL = strings.TrimSpace(r.GitURL)
	r.GitRevision = strings.TrimSpace(r.GitRevision)
	r.ValidationPath = strings.TrimSpace(r.ValidationPath)
	r.Image = strings.TrimSpace(r.Image)
	r.WorkspacePVC = strings.TrimSpace(r.WorkspacePVC)
	r.DockerConfigSecret = strings.TrimSpace(r.DockerConfigSecret)
	if r.Namespace == "" {
		return PipelineRunRef{}, errors.New("tekton namespace is required")
	}
	if r.PipelineName == "" {
		return PipelineRunRef{}, errors.New("tekton pipeline name is required")
	}
	if r.ServiceAccountName == "" {
		return PipelineRunRef{}, errors.New("tekton service account is required")
	}
	if r.ChangeNumber == "" {
		return PipelineRunRef{}, errors.New("change number is required")
	}
	if r.GitURL == "" {
		return PipelineRunRef{}, errors.New("tekton git URL is required")
	}
	if r.Image == "" {
		return PipelineRunRef{}, errors.New("tekton image is required")
	}
	if r.WorkspacePVC == "" {
		return PipelineRunRef{}, errors.New("tekton workspace PVC is required")
	}
	if r.DockerConfigSecret == "" {
		return PipelineRunRef{}, errors.New("tekton dockerconfig secret is required")
	}
	if r.GenerateName == "" {
		r.GenerateName = "devops-cp-validate-"
	}
	if r.GitRevision == "" {
		r.GitRevision = "main"
	}
	params := []map[string]string{
		{"name": "GIT_URL", "value": r.GitURL},
		{"name": "GIT_REVISION", "value": r.GitRevision},
		{"name": "IMAGE", "value": r.Image},
	}
	validationPath := r.ValidationPath
	if validationPath != "" {
		params = append(params, map[string]string{"name": "VALIDATION_PATH", "value": validationPath})
	}

	labels := map[string]string{
		"app.kubernetes.io/managed-by":       "devops-control-plane",
		"devops-control-plane/change-number": r.ChangeNumber,
	}
	if validationPath != "" {
		labels["devops-control-plane/validation-type"] = "gitops"
	}

	payload := map[string]any{
		"apiVersion": "tekton.dev/v1",
		"kind":       "PipelineRun",
		"metadata": map[string]any{
			"generateName": r.GenerateName,
			"namespace":    r.Namespace,
			"labels":       labels,
		},
		"spec": map[string]any{
			"pipelineRef": map[string]string{"name": r.PipelineName},
			"taskRunTemplate": map[string]string{
				"serviceAccountName": r.ServiceAccountName,
			},
			"params": params,
			"workspaces": []map[string]any{
				{"name": "shared-workspace", "persistentVolumeClaim": map[string]string{"claimName": r.WorkspacePVC}},
				{"name": "dockerconfig", "secret": map[string]string{"secretName": r.DockerConfigSecret}},
			},
		},
	}

	created, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/apis/tekton.dev/v1/namespaces/%s/pipelineruns", r.Namespace), payload)
	if err != nil {
		return PipelineRunRef{}, err
	}
	m, _ := created["metadata"].(map[string]any)
	name, _ := m["name"].(string)
	ns, _ := m["namespace"].(string)
	uid, _ := m["uid"].(string)
	if name == "" {
		return PipelineRunRef{}, errors.New("tekton API response did not include PipelineRun name")
	}
	if ns == "" {
		ns = r.Namespace
	}
	return PipelineRunRef{Name: name, Namespace: ns, UID: uid}, nil
}

// FindLatestPipelineRunByChange cerca la PipelineRun Tekton piu' recente associata a una ChangeRequest.
func (c *Client) FindLatestPipelineRunByChange(ctx context.Context, namespace string, changeNumber string) (PipelineRunStatus, error) {
	namespace = strings.TrimSpace(namespace)
	changeNumber = strings.TrimSpace(changeNumber)
	if namespace == "" {
		return PipelineRunStatus{}, errors.New("tekton namespace is required")
	}
	if changeNumber == "" {
		return PipelineRunStatus{}, errors.New("change number is required")
	}
	query := url.Values{}
	query.Set("labelSelector", "devops-control-plane/change-number="+changeNumber)
	path := fmt.Sprintf("/apis/tekton.dev/v1/namespaces/%s/pipelineruns?%s", namespace, query.Encode())
	list, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return PipelineRunStatus{}, err
	}
	items, _ := list["items"].([]any)
	if len(items) == 0 {
		return PipelineRunStatus{}, fmt.Errorf("tekton PipelineRun not found for change %q in namespace %q", changeNumber, namespace)
	}
	var latest map[string]any
	latestCreation := ""
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		metadata, _ := item["metadata"].(map[string]any)
		creation, _ := metadata["creationTimestamp"].(string)
		if latest == nil || creation > latestCreation {
			latest = item
			latestCreation = creation
		}
	}
	if latest == nil {
		return PipelineRunStatus{}, errors.New("tekton API response did not include valid PipelineRun items")
	}
	metadata, _ := latest["metadata"].(map[string]any)
	statusBlock, _ := latest["status"].(map[string]any)
	conditions, _ := statusBlock["conditions"].([]any)
	result := PipelineRunStatus{Name: stringValue(metadata, "name"), Namespace: stringValue(metadata, "namespace"), UID: stringValue(metadata, "uid"), CreationTimestamp: stringValue(metadata, "creationTimestamp"), CompletionTime: stringValue(statusBlock, "completionTime")}
	if result.Namespace == "" {
		result.Namespace = namespace
	}
	for _, rawCondition := range conditions {
		condition, ok := rawCondition.(map[string]any)
		if !ok {
			continue
		}
		if stringValue(condition, "type") == "Succeeded" {
			result.Status = stringValue(condition, "status")
			result.Reason = stringValue(condition, "reason")
			result.Message = stringValue(condition, "message")
			break
		}
	}
	if result.Status == "" {
		result.Status = "Unknown"
		result.Reason = "Pending"
		result.Message = "PipelineRun has no Succeeded condition yet"
	}
	return result, nil
}

// ListTaskRunsByPipelineRun returns TaskRun status details associated with a PipelineRun.
func (c *Client) ListTaskRunsByPipelineRun(ctx context.Context, namespace string, pipelineRunName string) ([]TaskRunStatus, error) {
	namespace = strings.TrimSpace(namespace)
	pipelineRunName = strings.TrimSpace(pipelineRunName)
	if namespace == "" {
		return nil, errors.New("tekton namespace is required")
	}
	if pipelineRunName == "" {
		return nil, errors.New("tekton PipelineRun name is required")
	}

	query := url.Values{}
	query.Set("labelSelector", "tekton.dev/pipelineRun="+pipelineRunName)
	path := fmt.Sprintf("/apis/tekton.dev/v1/namespaces/%s/taskruns?%s", namespace, query.Encode())
	list, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	items, _ := list["items"].([]any)
	results := make([]TaskRunStatus, 0, len(items))
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		metadata, _ := item["metadata"].(map[string]any)
		labels, _ := metadata["labels"].(map[string]any)
		statusBlock, _ := item["status"].(map[string]any)
		conditions, _ := statusBlock["conditions"].([]any)

		tr := TaskRunStatus{
			Name:             stringValue(metadata, "name"),
			Namespace:        stringValue(metadata, "namespace"),
			PipelineTaskName: stringValue(labels, "tekton.dev/pipelineTask"),
			TaskName:         stringValue(labels, "tekton.dev/task"),
			StartTime:        stringValue(statusBlock, "startTime"),
			CompletionTime:   stringValue(statusBlock, "completionTime"),
		}
		if tr.Namespace == "" {
			tr.Namespace = namespace
		}
		for _, rawCondition := range conditions {
			condition, ok := rawCondition.(map[string]any)
			if !ok {
				continue
			}
			if stringValue(condition, "type") == "Succeeded" {
				tr.Status = stringValue(condition, "status")
				tr.Reason = stringValue(condition, "reason")
				tr.Message = stringValue(condition, "message")
				break
			}
		}
		if tr.Status == "" {
			tr.Status = "Unknown"
			tr.Reason = "Pending"
			tr.Message = "TaskRun has no Succeeded condition yet"
		}
		results = append(results, tr)
	}
	return results, nil
}

func stringValue(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	value, _ := values[key].(string)
	return value
}

func (c *Client) do(ctx context.Context, method, path string, payload any) (map[string]any, error) {
	if c == nil {
		return nil, errors.New("tekton client is nil")
	}
	if strings.TrimSpace(c.apiURL) == "" {
		return nil, errors.New("tekton client API URL is empty")
	}
	if strings.TrimSpace(c.token) == "" {
		return nil, errors.New("tekton client token is empty")
	}
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal tekton request: %w", err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.apiURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create tekton request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute tekton request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read tekton response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tekton API %s %s returned status %d: %s", req.Method, req.URL.Path, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode tekton response: %w", err)
	}
	return out, nil
}
