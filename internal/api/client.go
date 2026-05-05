package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/FreeSign-io/fcli/internal/version"
	"github.com/hashicorp/go-retryablehttp"
)

// Client is a typed wrapper around the FreeSign HTTP API.
type Client struct {
	baseURL    string
	token      string
	httpClient *retryablehttp.Client
}

// Options configures a Client.
type Options struct {
	BaseURL string        // e.g. https://freesign.io
	Token   string        // raw token (with or without "Bearer " prefix or "api_" prefix)
	Timeout time.Duration // per-request; default 60s
	// MaxRetries on retryable failures (5xx, 429); default 3.
	MaxRetries int
	// HTTPClient overrides the underlying *http.Client (useful in tests).
	HTTPClient *http.Client
}

// New builds a Client.
func New(opts Options) (*Client, error) {
	if opts.BaseURL == "" {
		return nil, errors.New("base URL required")
	}

	rc := retryablehttp.NewClient()
	rc.RetryMax = opts.MaxRetries
	if opts.MaxRetries <= 0 {
		rc.RetryMax = 3
	}
	rc.RetryWaitMin = 500 * time.Millisecond
	rc.RetryWaitMax = 5 * time.Second
	rc.Logger = nil // silence default INFO logs; CLI is verbose-flag controlled
	if opts.HTTPClient != nil {
		rc.HTTPClient = opts.HTTPClient
	}
	if opts.Timeout > 0 {
		rc.HTTPClient.Timeout = opts.Timeout
	} else {
		rc.HTTPClient.Timeout = 60 * time.Second
	}

	return &Client{
		baseURL:    strings.TrimRight(opts.BaseURL, "/"),
		token:      strings.TrimSpace(strings.TrimPrefix(opts.Token, "Bearer ")),
		httpClient: rc,
	}, nil
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string { return c.baseURL }

// do executes an authenticated JSON request and decodes the response into out.
// out may be nil if no body is expected.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", fmt.Sprintf("fcli/%s", version.Version))
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{Status: resp.StatusCode, Body: respBody}
		_ = json.Unmarshal(respBody, apiErr) // best-effort decode of {message, code}
		return apiErr
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w (body: %s)", err, truncate(respBody, 200))
		}
	}
	return nil
}

// PutBytes uploads raw bytes (e.g. a PDF) to a presigned URL. No auth header
// (the URL is pre-signed); we don't go through retryablehttp since the
// presigned URL has a short TTL and large uploads should fail fast.
func (c *Client) PutBytes(ctx context.Context, presignedURL, contentType string, body io.Reader, length int64) error {
	if _, err := url.Parse(presignedURL); err != nil {
		return fmt.Errorf("invalid presigned URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, presignedURL, body)
	if err != nil {
		return fmt.Errorf("build PUT: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if length > 0 {
		req.ContentLength = length
	}
	req.Header.Set("User-Agent", fmt.Sprintf("fcli/%s", version.Version))

	resp, err := c.httpClient.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload PUT: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return &APIError{Status: resp.StatusCode, Body: respBody, Message: fmt.Sprintf("upload failed: HTTP %d", resp.StatusCode)}
	}
	return nil
}

// IsAuthError reports whether err is a 401/403 from the API.
func IsAuthError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == http.StatusUnauthorized || apiErr.Status == http.StatusForbidden
	}
	return false
}

// IsNotFound reports whether err is a 404.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == http.StatusNotFound
	}
	return false
}

// IsRateLimited reports whether err is a 429.
func IsRateLimited(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == http.StatusTooManyRequests
	}
	return false
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
