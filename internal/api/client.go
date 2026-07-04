package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPClient interface {
	Get(ctx context.Context, path string, out any) error
	GetWithStatus(ctx context.Context, path string, out any) (int, error)
	Post(ctx context.Context, path string, body, out any) error
	PostStatus(ctx context.Context, path string, body, out any) (int, error)
	BaseURL() string
}

const defaultTimeout = 30 * time.Second

var _ HTTPClient = (*Client)(nil)

type RefreshFunc func(ctx context.Context) (newToken string, err error)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	refresh    RefreshFunc
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (c *Client) WithToken(token string) *Client {
	clone := *c
	clone.token = token
	return &clone
}

func (c *Client) WithAutoRefresh(fn RefreshFunc) *Client {
	clone := *c
	clone.refresh = fn
	return &clone
}

func (c *Client) Get(ctx context.Context, path string, out any) error {
	_, err := c.doWithRefresh(ctx, http.MethodGet, path, nil, out)
	return err
}

func (c *Client) GetWithStatus(ctx context.Context, path string, out any) (int, error) {
	return c.doWithRefresh(ctx, http.MethodGet, path, nil, out)
}

func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	_, err := c.doWithRefresh(ctx, http.MethodPost, path, body, out)
	return err
}

func (c *Client) doWithRefresh(ctx context.Context, method, path string, body, out any) (int, error) {
	status, err := c.do(ctx, method, path, body, out)
	if err == nil || c.refresh == nil || !IsUnauthorized(err) {
		return status, err
	}

	newToken, rErr := c.refresh(ctx)
	if rErr != nil {
		return status, &APIError{
			Status:  http.StatusUnauthorized,
			Message: fmt.Sprintf("session expired, run `ipriver login`: %s", rErr),
		}
	}

	c.token = newToken
	return c.do(ctx, method, path, body, out)
}

func (c *Client) PostStatus(ctx context.Context, path string, body, out any) (int, error) {
	return c.doWithRefresh(ctx, http.MethodPost, path, body, out)
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) (int, error) {
	u := strings.TrimRight(c.baseURL, "/") + "/" + strings.TrimLeft(path, "/")

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return resp.StatusCode, parseAPIError(resp.StatusCode, respBody)
	}

	if out == nil || len(respBody) == 0 {
		return resp.StatusCode, nil
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}

	return resp.StatusCode, nil
}

type APIError struct {
	Status  int
	Message string
	Code    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("API request failed with status %d", e.Status)
}

func parseAPIError(status int, body []byte) error {
	var payload struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		Message          string `json:"message"`
		Code             int    `json:"code"`
	}
	_ = json.Unmarshal(body, &payload)

	msg := payload.Message
	if msg == "" {
		msg = payload.ErrorDescription
	}
	if msg == "" {
		msg = payload.Error
	}
	if msg == "" {
		msg = strings.TrimSpace(string(body))
	}

	code := payload.Error
	return &APIError{Status: status, Message: msg, Code: code}
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func IsUnauthorized(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized
}
