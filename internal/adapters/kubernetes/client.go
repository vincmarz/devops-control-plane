package kubernetes

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = defaultTimeoutSeconds
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &Client{apiURL: apiURL, token: token, httpClient: &http.Client{Timeout: time.Duration(timeout) * time.Second, Transport: transport}}
	for _, opt := range opts {
		opt(client)
	}
	return client, nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.do(ctx, http.MethodGet, "/api/v1", nil)
	return err
}

func (c *Client) CollectRuntimeEvidence(ctx context.Context, namespace string, applicationName string) (map[string]any, error) {
	namespace = strings.TrimSpace(namespace)
	applicationName = strings.TrimSpace(applicationName)
	if namespace == "" {
		return nil, errors.New("kubernetes namespace is required")
	}
	if applicationName == "" {
		return nil, errors.New("application name is required")
	}

	deployment, err := c.getDeployment(ctx, namespace, applicationName)
	if err != nil {
		return nil, err
	}
	selector := deploymentSelector(deployment)
	pods, err := c.listPods(ctx, namespace, selector, applicationName)
	if err != nil {
		return nil, err
	}
	service, err := c.getService(ctx, namespace, applicationName)
	if err != nil {
		service = map[string]any{"name": applicationName, "namespace": namespace, "error": err.Error()}
	}
	route, err := c.getRoute(ctx, namespace, applicationName)
	if err != nil {
		route = map[string]any{"name": applicationName, "namespace": namespace, "error": err.Error()}
	}

	return map[string]any{
		"namespace":  namespace,
		"deployment": deployment,
		"pods":       pods,
		"service":    service,
		"route":      route,
	}, nil
}

func (c *Client) getDeployment(ctx context.Context, namespace string, name string) (map[string]any, error) {
	obj, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s", namespace, name), nil)
	if err != nil {
		return nil, fmt.Errorf("get deployment %s/%s: %w", namespace, name, err)
	}
	metadata, _ := obj["metadata"].(map[string]any)
	spec, _ := obj["spec"].(map[string]any)
	status, _ := obj["status"].(map[string]any)
	selector, _ := spec["selector"].(map[string]any)
	matchLabels, _ := selector["matchLabels"].(map[string]any)
	return map[string]any{
		"name":               stringValue(metadata, "name"),
		"namespace":          stringValue(metadata, "namespace"),
		"generation":         intValue(metadata, "generation"),
		"observedGeneration": intValue(status, "observedGeneration"),
		"desiredReplicas":    intValue(spec, "replicas"),
		"replicas":           intValue(status, "replicas"),
		"readyReplicas":      intValue(status, "readyReplicas"),
		"availableReplicas":  intValue(status, "availableReplicas"),
		"updatedReplicas":    intValue(status, "updatedReplicas"),
		"selector":           matchLabels,
	}, nil
}

func (c *Client) listPods(ctx context.Context, namespace string, selector string, applicationName string) ([]map[string]any, error) {
	query := url.Values{}
	if selector != "" {
		query.Set("labelSelector", selector)
	}
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods", namespace)
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	list, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("list pods in %s: %w", namespace, err)
	}
	items, _ := list["items"].([]any)
	pods := make([]map[string]any, 0)
	for _, raw := range items {
		pod, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		metadata, _ := pod["metadata"].(map[string]any)
		podName := stringValue(metadata, "name")
		if !strings.HasPrefix(podName, applicationName+"-") {
			continue
		}
		status, _ := pod["status"].(map[string]any)
		spec, _ := pod["spec"].(map[string]any)
		containerStatuses, _ := status["containerStatuses"].([]any)
		containers := make([]map[string]any, 0)
		ready := true
		totalRestarts := 0
		for _, rawContainer := range containerStatuses {
			container, ok := rawContainer.(map[string]any)
			if !ok {
				continue
			}
			containerReady := boolValue(container, "ready")
			if !containerReady {
				ready = false
			}
			restartCount := intValue(container, "restartCount")
			totalRestarts += restartCount
			containers = append(containers, map[string]any{
				"name":         stringValue(container, "name"),
				"ready":        containerReady,
				"restartCount": restartCount,
				"image":        stringValue(container, "image"),
				"imageID":      stringValue(container, "imageID"),
			})
		}
		if len(containerStatuses) == 0 {
			ready = false
		}
		pods = append(pods, map[string]any{
			"name":         podName,
			"namespace":    stringValue(metadata, "namespace"),
			"phase":        stringValue(status, "phase"),
			"nodeName":     stringValue(spec, "nodeName"),
			"podIP":        stringValue(status, "podIP"),
			"ready":        ready,
			"restartCount": totalRestarts,
			"containers":   containers,
		})
	}
	return pods, nil
}

func (c *Client) getService(ctx context.Context, namespace string, name string) (map[string]any, error) {
	obj, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/namespaces/%s/services/%s", namespace, name), nil)
	if err != nil {
		return nil, fmt.Errorf("get service %s/%s: %w", namespace, name, err)
	}
	metadata, _ := obj["metadata"].(map[string]any)
	spec, _ := obj["spec"].(map[string]any)
	ports, _ := spec["ports"].([]any)
	outPorts := make([]map[string]any, 0)
	for _, rawPort := range ports {
		port, ok := rawPort.(map[string]any)
		if !ok {
			continue
		}
		outPorts = append(outPorts, map[string]any{
			"name":       stringValue(port, "name"),
			"port":       intValue(port, "port"),
			"targetPort": anyValue(port, "targetPort"),
			"protocol":   stringValue(port, "protocol"),
		})
	}
	return map[string]any{"name": stringValue(metadata, "name"), "namespace": stringValue(metadata, "namespace"), "type": stringValue(spec, "type"), "clusterIP": stringValue(spec, "clusterIP"), "ports": outPorts}, nil
}

func (c *Client) getRoute(ctx context.Context, namespace string, name string) (map[string]any, error) {
	obj, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/apis/route.openshift.io/v1/namespaces/%s/routes/%s", namespace, name), nil)
	if err != nil {
		return nil, fmt.Errorf("get route %s/%s: %w", namespace, name, err)
	}
	metadata, _ := obj["metadata"].(map[string]any)
	spec, _ := obj["spec"].(map[string]any)
	to, _ := spec["to"].(map[string]any)
	tls, _ := spec["tls"].(map[string]any)
	return map[string]any{"name": stringValue(metadata, "name"), "namespace": stringValue(metadata, "namespace"), "host": stringValue(spec, "host"), "to": stringValue(to, "name"), "tlsTermination": stringValue(tls, "termination")}, nil
}

func deploymentSelector(deployment map[string]any) string {
	selector, _ := deployment["selector"].(map[string]any)
	parts := make([]string, 0, len(selector))
	for key, value := range selector {
		parts = append(parts, fmt.Sprintf("%s=%s", key, fmt.Sprint(value)))
	}
	return strings.Join(parts, ",")
}

func stringValue(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	value, _ := values[key].(string)
	return value
}

func boolValue(values map[string]any, key string) bool {
	if values == nil {
		return false
	}
	value, _ := values[key].(bool)
	return value
}

func intValue(values map[string]any, key string) int {
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
	case json.Number:
		parsed, _ := strconv.Atoi(value.String())
		return parsed
	default:
		return 0
	}
}

func anyValue(values map[string]any, key string) any {
	if values == nil {
		return nil
	}
	return values[key]
}

func (c *Client) do(ctx context.Context, method string, path string, payload any) (map[string]any, error) {
	if c == nil {
		return nil, errors.New("kubernetes client is nil")
	}
	var body io.Reader
	if payload != nil {
		return nil, errors.New("kubernetes client payloads are not implemented")
	}
	req, err := http.NewRequestWithContext(ctx, method, c.apiURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute kubernetes request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read kubernetes response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kubernetes API %s %s returned status %d: %s", req.Method, req.URL.Path, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode kubernetes response: %w", err)
	}
	return out, nil
}
