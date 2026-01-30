// Package waldur provides a wrapper around the official Waldur go-client
// with VirtEngine-specific configuration, error handling, and retry logic.
//
// VE-2024: Waldur API integration using official Go client
package waldur

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	client "github.com/waldur/go-client"
)

// Waldur-specific errors
var (
	// ErrNotConfigured is returned when the client is not configured
	ErrNotConfigured = errors.New("waldur client not configured")

	// ErrInvalidToken is returned when the API token is invalid
	ErrInvalidToken = errors.New("invalid API token")

	// ErrUnauthorized is returned when the API returns 401
	ErrUnauthorized = errors.New("unauthorized: check API token")

	// ErrForbidden is returned when the API returns 403
	ErrForbidden = errors.New("forbidden: insufficient permissions")

	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrConflict is returned when there's a resource conflict
	ErrConflict = errors.New("resource conflict")

	// ErrRateLimited is returned when rate limited
	ErrRateLimited = errors.New("rate limited")

	// ErrServerError is returned for 5xx errors
	ErrServerError = errors.New("waldur server error")

	// ErrTimeout is returned when a request times out
	ErrTimeout = errors.New("request timeout")

	// ErrInvalidResponse is returned when the response cannot be parsed
	ErrInvalidResponse = errors.New("invalid response from waldur")
)

// Config holds the Waldur client configuration
type Config struct {
	// BaseURL is the Waldur API base URL (e.g., "https://waldur.example.com/api")
	BaseURL string

	// Token is the API authentication token
	Token string

	// Timeout is the HTTP request timeout
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryWaitMin is the minimum wait time between retries
	RetryWaitMin time.Duration

	// RetryWaitMax is the maximum wait time between retries
	RetryWaitMax time.Duration

	// RateLimitPerSecond is the maximum requests per second (0 = unlimited)
	RateLimitPerSecond int

	// UserAgent is the User-Agent header value
	UserAgent string
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Timeout:            30 * time.Second,
		MaxRetries:         3,
		RetryWaitMin:       1 * time.Second,
		RetryWaitMax:       30 * time.Second,
		RateLimitPerSecond: 10,
		UserAgent:          "VirtEngine-Provider-Daemon/1.0",
	}
}

// Client wraps the official Waldur go-client with VirtEngine-specific functionality
type Client struct {
	mu         sync.RWMutex
	config     Config
	httpClient *http.Client
	auth       *client.TokenAuth
	api        *client.ClientWithResponses

	// Rate limiting
	rateLimiter *rateLimiter

	// Metrics
	requestCount   int64
	errorCount     int64
	lastRequestAt  time.Time
	lastResponseAt time.Time
}

// rateLimiter implements a simple token bucket rate limiter
type rateLimiter struct {
	mu           sync.Mutex
	tokens       float64
	maxTokens    float64
	refillRate   float64
	lastRefillAt time.Time
}

func newRateLimiter(rps int) *rateLimiter {
	if rps <= 0 {
		return nil
	}
	return &rateLimiter{
		tokens:       float64(rps),
		maxTokens:    float64(rps),
		refillRate:   float64(rps),
		lastRefillAt: time.Now(),
	}
}

func (r *rateLimiter) Wait(ctx context.Context) error {
	if r == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(r.lastRefillAt).Seconds()
	r.tokens += elapsed * r.refillRate
	if r.tokens > r.maxTokens {
		r.tokens = r.maxTokens
	}
	r.lastRefillAt = now

	// Wait if no tokens available
	if r.tokens < 1 {
		waitDuration := time.Duration((1 - r.tokens) / r.refillRate * float64(time.Second))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
			r.tokens = 0
		}
	} else {
		r.tokens--
	}

	return nil
}

// NewClient creates a new Waldur client with the given configuration
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("%w: base URL is required", ErrNotConfigured)
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("%w: API token is required", ErrNotConfigured)
	}

	// Apply defaults for unset values
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultConfig().Timeout
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = DefaultConfig().MaxRetries
	}
	if cfg.RetryWaitMin == 0 {
		cfg.RetryWaitMin = DefaultConfig().RetryWaitMin
	}
	if cfg.RetryWaitMax == 0 {
		cfg.RetryWaitMax = DefaultConfig().RetryWaitMax
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = DefaultConfig().UserAgent
	}

	// Create token auth
	auth, err := client.NewTokenAuth(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth: %w", err)
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}

	// Create Waldur API client
	api, err := client.NewClientWithResponses(
		cfg.BaseURL,
		client.WithHTTPClient(httpClient),
		client.WithRequestEditorFn(auth.Intercept),
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("User-Agent", cfg.UserAgent)
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create waldur client: %w", err)
	}

	return &Client{
		config:      cfg,
		httpClient:  httpClient,
		auth:        auth,
		api:         api,
		rateLimiter: newRateLimiter(cfg.RateLimitPerSecond),
	}, nil
}

// API returns the underlying Waldur API client for direct access
func (c *Client) API() *client.ClientWithResponses {
	return c.api
}

// Config returns the current configuration
func (c *Client) Config() Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// Metrics returns client metrics
func (c *Client) Metrics() ClientMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return ClientMetrics{
		RequestCount:   c.requestCount,
		ErrorCount:     c.errorCount,
		LastRequestAt:  c.lastRequestAt,
		LastResponseAt: c.lastResponseAt,
	}
}

// ClientMetrics holds client metrics
type ClientMetrics struct {
	RequestCount   int64
	ErrorCount     int64
	LastRequestAt  time.Time
	LastResponseAt time.Time
}

// doWithRetry executes a function with retry logic
func (c *Client) doWithRetry(ctx context.Context, fn func() error) error {
	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	c.mu.Lock()
	c.requestCount++
	c.lastRequestAt = time.Now()
	c.mu.Unlock()

	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff with exponential increase
			waitTime := c.config.RetryWaitMin * time.Duration(1<<uint(attempt-1))
			if waitTime > c.config.RetryWaitMax {
				waitTime = c.config.RetryWaitMax
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
			}
		}

		err := fn()
		c.mu.Lock()
		c.lastResponseAt = time.Now()
		c.mu.Unlock()

		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on certain errors
		if errors.Is(err, ErrUnauthorized) ||
			errors.Is(err, ErrForbidden) ||
			errors.Is(err, ErrNotFound) ||
			errors.Is(err, ErrConflict) ||
			errors.Is(err, context.Canceled) {
			break
		}

		// Retry on rate limit and server errors
		if errors.Is(err, ErrRateLimited) || errors.Is(err, ErrServerError) {
			continue
		}
	}

	c.mu.Lock()
	c.errorCount++
	c.mu.Unlock()

	return lastErr
}

// mapHTTPError maps HTTP status codes to VirtEngine errors
func mapHTTPError(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusConflict:
		return ErrConflict
	case http.StatusTooManyRequests:
		return ErrRateLimited
	}

	if statusCode >= 500 {
		if len(body) > 0 {
			return fmt.Errorf("%w: %s", ErrServerError, string(body))
		}
		return ErrServerError
	}

	if statusCode >= 400 {
		if len(body) > 0 {
			return fmt.Errorf("waldur error (%d): %s", statusCode, string(body))
		}
		return fmt.Errorf("waldur error: status %d", statusCode)
	}

	return nil
}

// doRequest performs a raw HTTP request to the Waldur API.
// VE-2D: Added for provider offerings API which may not be in the generated client.
func (c *Client) doRequest(ctx context.Context, method, path string, body []byte) ([]byte, int, error) {
	url := c.config.BaseURL + path

	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.config.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// HealthCheck verifies the client can connect to Waldur
func (c *Client) HealthCheck(ctx context.Context) error {
	var healthErr error

	err := c.doWithRetry(ctx, func() error {
		resp, err := c.api.UsersMeRetrieveWithResponse(ctx, &client.UsersMeRetrieveParams{})
		if err != nil {
			return fmt.Errorf("health check failed: %w", err)
		}

		if resp.StatusCode() != http.StatusOK {
			healthErr = mapHTTPError(resp.StatusCode(), resp.Body)
			return healthErr
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// GetCurrentUser retrieves the current authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (*UserInfo, error) {
	var user *UserInfo

	err := c.doWithRetry(ctx, func() error {
		resp, err := c.api.UsersMeRetrieveWithResponse(ctx, &client.UsersMeRetrieveParams{})
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		user = &UserInfo{
			Username: safeString(resp.JSON200.Username),
		}
		if resp.JSON200.Uuid != nil {
			user.UUID = resp.JSON200.Uuid.String()
		}
		if resp.JSON200.Email != nil {
			user.Email = string(*resp.JSON200.Email)
		}
		if resp.JSON200.FirstName != nil {
			user.FirstName = *resp.JSON200.FirstName
		}
		if resp.JSON200.LastName != nil {
			user.LastName = *resp.JSON200.LastName
		}

		return nil
	})

	return user, err
}

// UserInfo contains user information
type UserInfo struct {
	UUID      string
	Username  string
	Email     string
	FirstName string
	LastName  string
}

// safeString safely dereferences a string pointer
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// safeInt safely dereferences an int pointer
func safeInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// ptr returns a pointer to the value
func ptr[T any](v T) *T {
	return &v
}
