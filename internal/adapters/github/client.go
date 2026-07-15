package github

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/vincmarz/devops-control-plane/internal/adapters/tlsutil"
)

const defaultTimeoutSeconds = 30

type Client struct {
	apiURL     string
	token      string
	httpClient *http.Client
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func New(cfg Config, opts ...Option) (*Client, error) {
	apiURL := strings.TrimRight(strings.TrimSpace(cfg.APIURL), "/")
	if apiURL == "" {
		return nil, errors.New("github API URL is required")
	}
	parsed, err := url.Parse(apiURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("github API URL must be an absolute URL")
	}
	token := strings.TrimSpace(cfg.Token)
	if token == "" {
		return nil, errors.New("github token is required")
	}
	timeoutSeconds := cfg.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultTimeoutSeconds
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // Explicit lab-only option.
	} else if strings.TrimSpace(cfg.CAFile) != "" {
		tlsConfig, err := tlsutil.TLSConfigFromCAFile(cfg.CAFile)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig = tlsConfig
	}
	client := &Client{apiURL: apiURL, token: token, httpClient: &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second, Transport: transport}}
	for _, opt := range opts {
		opt(client)
	}
	return client, nil
}

func ParseProjectPath(projectPath string) (string, string, error) {
	parts := strings.Split(strings.Trim(strings.TrimSpace(projectPath), "/"), "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("github projectPath %q must use owner/repository format", projectPath)
	}
	return parts[0], parts[1], nil
}

func (c *Client) CreateBranch(ctx context.Context, projectPath, branch, ref string) (Reference, error) {
	owner, repository, err := ParseProjectPath(projectPath)
	if err != nil {
		return Reference{}, err
	}
	branch = strings.TrimSpace(branch)
	ref = strings.TrimSpace(ref)
	if branch == "" {
		return Reference{}, errors.New("github branch is required")
	}
	if ref == "" {
		return Reference{}, errors.New("github branch ref is required")
	}
	var base Reference
	if err := c.doJSON(ctx, http.MethodGet, c.repositoryPath(owner, repository, "git", "ref", "heads", ref), nil, &base); err != nil {
		return Reference{}, err
	}
	payload := map[string]string{"ref": "refs/heads/" + branch, "sha": base.Object.SHA}
	var created Reference
	if err := c.doJSON(ctx, http.MethodPost, c.repositoryPath(owner, repository, "git", "refs"), payload, &created); err != nil {
		return Reference{}, err
	}
	return created, nil
}

func (c *Client) CreateOrUpdateFile(ctx context.Context, projectPath, branch, filePath, commitMessage, content string) (RepositoryContent, error) {
	owner, repository, err := ParseProjectPath(projectPath)
	if err != nil {
		return RepositoryContent{}, err
	}
	branch = strings.TrimSpace(branch)
	filePath = strings.Trim(strings.TrimSpace(filePath), "/")
	commitMessage = strings.TrimSpace(commitMessage)
	if branch == "" || filePath == "" || commitMessage == "" {
		return RepositoryContent{}, errors.New("github branch, filePath and commitMessage are required")
	}
	endpoint := c.repositoryPath(owner, repository, "contents", filePath)
	var existing RepositoryContent
	query := url.Values{"ref": []string{branch}}
	getErr := c.doJSON(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil, &existing)
	if getErr != nil && !errors.Is(getErr, ErrNotFound) {
		return RepositoryContent{}, getErr
	}
	payload := map[string]string{"message": commitMessage, "content": base64.StdEncoding.EncodeToString([]byte(content)), "branch": branch}
	if getErr == nil && strings.TrimSpace(existing.SHA) != "" {
		payload["sha"] = existing.SHA
	}
	var response struct {
		Content RepositoryContent `json:"content"`
	}
	if err := c.doJSON(ctx, http.MethodPut, endpoint, payload, &response); err != nil {
		return RepositoryContent{}, err
	}
	return response.Content, nil
}

func (c *Client) OpenPullRequest(ctx context.Context, projectPath, sourceBranch, targetBranch, title, description string) (PullRequest, error) {
	owner, repository, err := ParseProjectPath(projectPath)
	if err != nil {
		return PullRequest{}, err
	}
	payload := map[string]string{"head": strings.TrimSpace(sourceBranch), "base": strings.TrimSpace(targetBranch), "title": strings.TrimSpace(title), "body": description}
	var pullRequest PullRequest
	if err := c.doJSON(ctx, http.MethodPost, c.repositoryPath(owner, repository, "pulls"), payload, &pullRequest); err != nil {
		return PullRequest{}, err
	}
	return pullRequest, nil
}

func (c *Client) FindOpenPullRequest(ctx context.Context, projectPath, sourceBranch, targetBranch string) (PullRequest, error) {
	owner, repository, err := ParseProjectPath(projectPath)
	if err != nil {
		return PullRequest{}, err
	}
	query := url.Values{"state": []string{"open"}, "head": []string{owner + ":" + strings.TrimSpace(sourceBranch)}, "base": []string{strings.TrimSpace(targetBranch)}}
	var pullRequests []PullRequest
	if err := c.doJSON(ctx, http.MethodGet, c.repositoryPath(owner, repository, "pulls")+"?"+query.Encode(), nil, &pullRequests); err != nil {
		return PullRequest{}, err
	}
	if len(pullRequests) == 0 {
		return PullRequest{}, errors.New("github open pull request was not found")
	}
	if len(pullRequests) != 1 {
		return PullRequest{}, fmt.Errorf("github open pull request lookup returned %d results", len(pullRequests))
	}
	return pullRequests[0], nil
}

func (c *Client) MergePullRequest(ctx context.Context, projectPath string, number int, commitMessage string) (PullRequestMerge, error) {
	owner, repository, err := ParseProjectPath(projectPath)
	if err != nil {
		return PullRequestMerge{}, err
	}
	if number <= 0 {
		return PullRequestMerge{}, errors.New("github pull request number must be greater than zero")
	}
	payload := map[string]string{"commit_message": strings.TrimSpace(commitMessage), "merge_method": "merge"}
	var merged PullRequestMerge
	if err := c.doJSON(ctx, http.MethodPut, c.repositoryPath(owner, repository, "pulls", fmt.Sprintf("%d", number), "merge"), payload, &merged); err != nil {
		return PullRequestMerge{}, err
	}
	if !merged.Merged {
		return PullRequestMerge{}, fmt.Errorf("github pull request was not merged: %s", merged.Message)
	}
	return merged, nil
}

var ErrNotFound = errors.New("github resource not found")

func (c *Client) repositoryPath(owner, repository string, elements ...string) string {
	values := []string{"repos", owner, repository}
	values = append(values, elements...)
	escaped := make([]string, 0, len(values))
	for _, value := range values {
		escaped = append(escaped, url.PathEscape(value))
	}
	return "/" + path.Join(escaped...)
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, requestBody any, responseBody any) error {
	var body io.Reader
	if requestBody != nil {
		encoded, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.apiURL+endpoint, body)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	content, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("github API %s %s failed with status %d", method, endpoint, response.StatusCode)
	}
	if responseBody != nil && len(content) != 0 {
		if err := json.Unmarshal(content, responseBody); err != nil {
			return err
		}
	}
	return nil
}
