package servicediscovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	options    *clientOptions

	serviceID string
	stopCh    chan struct{}
	mu        sync.Mutex
}

func NewClient(baseURL string, opts ...ClientOption) *Client {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: options.timeout,
		},
		options: options,
	}
}

func (c *Client) Register(ctx context.Context, req *RegisterRequest) (*Service, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/services/register", body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var service Service
	if err := json.NewDecoder(resp.Body).Decode(&service); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.serviceID = service.ID
	c.mu.Unlock()

	return &service, nil
}

func (c *Client) AutoRegister(ctx context.Context, opts ...RegisterOption) (*Service, error) {
	regOpts := &registerOptions{}
	for _, opt := range opts {
		opt(regOpts)
	}

	req := &RegisterRequest{
		Name:        c.getServiceName(regOpts.name),
		Host:        c.getHost(regOpts.host),
		Port:        c.getPort(regOpts.port),
		Protocol:    c.getProtocol(regOpts.protocol),
		BasePath:    c.getBasePath(regOpts.basePath),
		Routes:      regOpts.routes,
		HealthCheck: c.getHealthCheck(regOpts.healthCheck),
		Tags:        c.getTags(regOpts.tags),
		Metadata:    c.getMetadata(regOpts.metadata),
	}

	return c.Register(ctx, req)
}

func (c *Client) Get(ctx context.Context, id string) (*Service, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/services/"+id, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrServiceNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var service Service
	if err := json.NewDecoder(resp.Body).Decode(&service); err != nil {
		return nil, err
	}

	return &service, nil
}

func (c *Client) List(ctx context.Context) ([]*Service, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/services/list", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var listResp ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	return listResp.Services, nil
}

func (c *Client) Search(ctx context.Context, route, name, tag string) ([]*Service, error) {
	params := url.Values{}
	if route != "" {
		params.Set("route", route)
	}
	if name != "" {
		params.Set("name", name)
	}
	if tag != "" {
		params.Set("tag", tag)
	}

	path := "/api/v1/services/search"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var listResp ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	return listResp.Services, nil
}

func (c *Client) Update(ctx context.Context, id string, req *UpdateRequest) (*Service, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, http.MethodPut, "/api/v1/services/"+id+"/update", body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrServiceNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var service Service
	if err := json.NewDecoder(resp.Body).Decode(&service); err != nil {
		return nil, err
	}

	return &service, nil
}

func (c *Client) Unregister(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/api/v1/services/"+id+"/unregister", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return ErrServiceNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *Client) Heartbeat(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/api/v1/services/"+id+"/heartbeat", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return ErrServiceNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *Client) StartHeartbeat(ctx context.Context, id string, interval time.Duration) {
	c.mu.Lock()
	if c.stopCh != nil {
		close(c.stopCh)
	}
	c.stopCh = make(chan struct{})
	stopCh := c.stopCh
	c.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-stopCh:
				return
			case <-ticker.C:
				_ = c.Heartbeat(ctx, id)
			}
		}
	}()
}

func (c *Client) StopHeartbeat() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopCh != nil {
		close(c.stopCh)
		c.stopCh = nil
	}
}

func (c *Client) GetServiceID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.serviceID
}

func (c *Client) doRequest(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	var lastErr error

	for i := 0; i <= c.options.retries; i++ {
		var reqBody io.Reader
		if body != nil {
			reqBody = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
		if err != nil {
			return nil, err
		}

		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if i < c.options.retries {
				time.Sleep(c.options.retryDelay)
				continue
			}
			return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
		}

		return resp, nil
	}

	return nil, lastErr
}

func (c *Client) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return fmt.Errorf("%s", errResp.Error)
	}
	return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

func (c *Client) getServiceName(override string) string {
	if override != "" {
		return override
	}
	if name := os.Getenv("SERVICE_NAME"); name != "" {
		return name
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Base(exe)
	}
	return "unknown"
}

func (c *Client) getHost(override string) string {
	if override != "" {
		return override
	}
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		return podIP
	}
	if hostname := os.Getenv("HOSTNAME"); hostname != "" {
		return hostname
	}
	if host, err := os.Hostname(); err == nil {
		return host
	}
	return "localhost"
}

func (c *Client) getPort(override int) int {
	if override != 0 {
		return override
	}
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			return p
		}
	}
	return 8080
}

func (c *Client) getProtocol(override string) string {
	if override != "" {
		return override
	}
	if protocol := os.Getenv("SERVICE_PROTOCOL"); protocol != "" {
		return protocol
	}
	return "http"
}

func (c *Client) getBasePath(override string) string {
	if override != "" {
		return override
	}
	return os.Getenv("SERVICE_BASE_PATH")
}

func (c *Client) getHealthCheck(override string) string {
	if override != "" {
		return override
	}
	if hc := os.Getenv("SERVICE_HEALTH_CHECK"); hc != "" {
		return hc
	}
	return "/health"
}

func (c *Client) getTags(override []string) []string {
	if len(override) > 0 {
		return override
	}
	if tags := os.Getenv("SERVICE_TAGS"); tags != "" {
		return strings.Split(tags, ",")
	}
	return nil
}

func (c *Client) getMetadata(override map[string]string) map[string]string {
	if len(override) > 0 {
		return override
	}

	metadata := make(map[string]string)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "SERVICE_META_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimPrefix(parts[0], "SERVICE_META_")
				metadata[strings.ToLower(key)] = parts[1]
			}
		}
	}

	if len(metadata) == 0 {
		return nil
	}
	return metadata
}
