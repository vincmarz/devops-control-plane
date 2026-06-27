package argocd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vincmarz/devops-control-plane/internal/domain"
)

type Client struct {
	baseURL    string
	authToken  string
	httpClient *http.Client
}

func New(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, errors.New("argocd base URL is required")
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("invalid argocd base URL: %w", err)
	}

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // Lab/self-signed OpenShift GitOps route support.
	}

	return &Client{
		baseURL:   baseURL,
		authToken: strings.TrimSpace(cfg.AuthToken),
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/version")
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("argocd ping failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) ListApplications(ctx context.Context) ([]domain.Application, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/v1/applications")
	if err != nil {
		return nil, err
	}
	var payload applicationListResponse
	if err := c.doJSON(req, &payload); err != nil {
		return nil, err
	}
	apps := make([]domain.Application, 0, len(payload.Items))
	for _, item := range payload.Items {
		apps = append(apps, item.toDomain())
	}
	return apps, nil
}

func (c *Client) GetApplication(ctx context.Context, name string) (domain.Application, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Application{}, errors.New("argocd application name is required")
	}
	req, err := c.newRequest(ctx, http.MethodGet, "/api/v1/applications/"+url.PathEscape(name))
	if err != nil {
		return domain.Application{}, err
	}
	var payload applicationResponse
	if err := c.doJSON(req, &payload); err != nil {
		return domain.Application{}, err
	}
	return payload.toDomain(), nil
}

func (c *Client) newRequest(ctx context.Context, method string, path string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
	return req, nil
}

func (c *Client) doJSON(req *http.Request, target any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("argocd API %s %s failed with status %d", req.Method, req.URL.Path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode argocd API response: %w", err)
	}
	return nil
}
