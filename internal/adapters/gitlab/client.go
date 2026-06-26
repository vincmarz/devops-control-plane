package gitlab

import (
	"context"
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

	client := &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
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
