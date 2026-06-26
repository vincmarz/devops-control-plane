package gitlab

import (
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

// Config contiene la configurazione minima per parlare con GitLab REST API v4.
type Config struct {
	BaseURL        string
	Token          string
	TimeoutSeconds int
	InsecureTLS    bool
}

// Client implementa un client minimale per GitLab REST API v4.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Option permette di personalizzare il client, soprattutto nei test.
type Option func(*Client)

// WithHTTPClient consente di iniettare un client HTTP custom, ad esempio httptest.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// New crea un client GitLab REST API v4.
func New(cfg Config, opts ...Option) (*Client, error) {
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		return nil, errors.New("gitlab base URL is required")
	}
	baseURL = strings.TrimRight(baseURL, "/")

	token := strings.TrimSpace(cfg.Token)
	if token == "" {
		return nil, errors.New("gitlab token is required")
	}

	timeoutSeconds := cfg.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultTimeoutSeconds
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // Lab-only option for self-signed GitLab routes.
	}

	client := &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout:   time.Duration(timeoutSeconds) * time.Second,
			Transport: transport,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// Ping valida la raggiungibilita' di GitLab usando un endpoint autenticato.
func (c *Client) Ping(ctx context.Context) error {
	var currentUser struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	}
	return c.doJSON(ctx, http.MethodGet, "/api/v4/user", nil, &currentUser)
}

// CreateBranch crea un branch GitLab partendo da un ref esistente.
func (c *Client) CreateBranch(ctx context.Context, projectID int, branch string, ref string) (Branch, error) {
	branch = strings.TrimSpace(branch)
	ref = strings.TrimSpace(ref)

	if projectID <= 0 {
		return Branch{}, errors.New("gitlab project ID must be greater than zero")
	}
	if branch == "" {
		return Branch{}, errors.New("gitlab branch is required")
	}
	if ref == "" {
		return Branch{}, errors.New("gitlab ref is required")
	}

	form := url.Values{}
	form.Set("branch", branch)
	form.Set("ref", ref)

	path := fmt.Sprintf("/api/v4/projects/%d/repository/branches", projectID)

	var created Branch
	if err := c.doForm(ctx, http.MethodPost, path, form, &created); err != nil {
		return Branch{}, err
	}

	return created, nil
}

// CreateOrUpdateFile crea un file GitLab oppure lo aggiorna se esiste gia'.
func (c *Client) CreateOrUpdateFile(ctx context.Context, projectID int, branch string, filePath string, commitMessage string, content string) (RepositoryFile, error) {
	branch = strings.TrimSpace(branch)
	filePath = strings.TrimSpace(filePath)
	commitMessage = strings.TrimSpace(commitMessage)

	if projectID <= 0 {
		return RepositoryFile{}, errors.New("gitlab project ID must be greater than zero")
	}
	if branch == "" {
		return RepositoryFile{}, errors.New("gitlab branch is required")
	}
	if filePath == "" {
		return RepositoryFile{}, errors.New("gitlab file path is required")
	}
	if commitMessage == "" {
		return RepositoryFile{}, errors.New("gitlab commit message is required")
	}

	form := url.Values{}
	form.Set("branch", branch)
	form.Set("commit_message", commitMessage)
	form.Set("content", content)

	path := fmt.Sprintf("/api/v4/projects/%d/repository/files/%s", projectID, url.PathEscape(filePath))

	var created RepositoryFile
	if err := c.doForm(ctx, http.MethodPost, path, form, &created); err != nil {
		if !isFileAlreadyExistsError(err) {
			return RepositoryFile{}, err
		}
		var updated RepositoryFile
		if updateErr := c.doForm(ctx, http.MethodPut, path, form, &updated); updateErr != nil {
			return RepositoryFile{}, updateErr
		}
		return updated, nil
	}

	return created, nil
}

func isFileAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "already exists") || strings.Contains(message, "file with this name")
}

func (c *Client) doForm(ctx context.Context, method string, path string, form url.Values, out any) error {
	body := strings.NewReader(form.Encode())
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req, out)
}

func (c *Client) doJSON(ctx context.Context, method string, path string, body io.Reader, out any) error {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	if c == nil {
		return nil, errors.New("gitlab client is nil")
	}
	if strings.TrimSpace(c.baseURL) == "" {
		return nil, errors.New("gitlab client base URL is empty")
	}
	if strings.TrimSpace(c.token) == "" {
		return nil, errors.New("gitlab client token is empty")
	}

	requestURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("create gitlab request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute gitlab request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read gitlab response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gitlab API %s %s returned status %d: %s", req.Method, req.URL.Path, resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	if out == nil || len(responseBody) == 0 {
		return nil
	}

	if err := json.Unmarshal(responseBody, out); err != nil {
		return fmt.Errorf("decode gitlab response: %w", err)
	}

	return nil
}
