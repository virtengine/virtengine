// Package waldur provides a wrapper around the official Waldur go-client
// with VirtEngine-specific configuration, error handling, and retry logic.
//
// VE-2024: Waldur API integration using official Go client
package waldur

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	client "github.com/waldur/go-client"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/virtengine/virtengine/pkg/security"
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

const (
	usersMePath    = "/users/me/"
	apiUsersMePath = "/api/users/me/"
)

// Config holds the Waldur client configuration
type Config struct {
	// BaseURL is the Waldur API base URL (e.g., "https://waldur.example.com/api")
	BaseURL string

	// Token is the API authentication token
	Token string

	// AuthScheme controls the Authorization scheme when Token does not include one.
	// Supported values are "auto", "token", and "bearer".
	AuthScheme string

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

	// VersionPaths controls API base path negotiation when BaseURL points at the site root.
	// When empty, "/api/" and "/api/v1/" are attempted after the provided BaseURL.
	VersionPaths []string
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
		AuthScheme:         "auto",
		VersionPaths:       []string{"/api/", "/api/v1/"},
	}
}

// Client wraps the official Waldur go-client with VirtEngine-specific functionality
type Client struct {
	mu         sync.RWMutex
	config     Config
	httpClient *http.Client
	api        *client.ClientWithResponses
	authHeader string

	// Rate limiting
	rateLimiter *rateLimiter

	// Metrics
	requestCount   int64
	errorCount     int64
	lastRequestAt  time.Time
	lastResponseAt time.Time

	resolvedBaseURL string
	negotiated      bool
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

	// Create HTTP client with timeout
	httpClient := security.NewSecureHTTPClient(security.WithTimeout(cfg.Timeout))
	httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)

	baseURL := normalizeBaseURL(cfg.BaseURL)
	authHeader := formatAuthHeader(cfg.Token, cfg.AuthScheme)

	c := &Client{
		config:          cfg,
		httpClient:      httpClient,
		rateLimiter:     newRateLimiter(cfg.RateLimitPerSecond),
		authHeader:      authHeader,
		resolvedBaseURL: baseURL,
	}

	api, err := c.newAPIClient(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create waldur client: %w", err)
	}
	c.api = api
	return c, nil
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
	if err := c.ensureNegotiated(ctx); err != nil {
		return err
	}

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
			//nolint:gosec // G115: attempt is bounded by MaxRetries configuration (typically < 10)
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
		if errors.Is(err, ErrTimeout) {
			continue
		}

		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
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
	if err := c.ensureNegotiated(ctx); err != nil {
		return nil, 0, err
	}

	targetURL := path
	if !isAbsoluteURL(path) {
		targetURL = joinBaseURL(c.resolvedBaseURL, path)
	}

	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, targetURL, bytes.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, targetURL, nil)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, classifyTransportError(err)
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
	return c.doWithRetry(ctx, func() error {
		respBody, statusCode, err := c.doRequest(ctx, http.MethodGet, usersMePath, nil)
		if err != nil {
			return fmt.Errorf("health check failed: %w", err)
		}
		if statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}
		return nil
	})
}

// GetCurrentUser retrieves the current authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (*UserInfo, error) {
	var user *UserInfo

	err := c.doWithRetry(ctx, func() error {
		respBody, statusCode, err := c.doRequest(ctx, http.MethodGet, usersMePath, nil)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		var resp struct {
			Uuid      *string `json:"uuid"`
			Username  *string `json:"username"`
			Email     *string `json:"email"`
			FirstName *string `json:"first_name"`
			LastName  *string `json:"last_name"`
		}
		if err := json.Unmarshal(respBody, &resp); err != nil {
			return ErrInvalidResponse
		}

		user = &UserInfo{
			Username: safeString(resp.Username),
		}
		if resp.Uuid != nil {
			user.UUID = *resp.Uuid
		}
		if resp.Email != nil {
			user.Email = string(*resp.Email)
		}
		if resp.FirstName != nil {
			user.FirstName = *resp.FirstName
		}
		if resp.LastName != nil {
			user.LastName = *resp.LastName
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

func (c *Client) ensureNegotiated(ctx context.Context) error {
	c.mu.RLock()
	if c.negotiated {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.negotiated {
		return nil
	}

	for _, candidate := range baseURLCandidates(c.config.BaseURL, c.config.VersionPaths) {
		statusCode, err := c.probeUsersMe(ctx, candidate)
		if err != nil {
			continue
		}
		if statusCode == http.StatusOK || statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
			api, buildErr := c.newAPIClient(candidate)
			if buildErr != nil {
				return buildErr
			}
			c.api = api
			c.resolvedBaseURL = candidate
			c.negotiated = true
			return nil
		}
	}

	c.negotiated = true
	return nil
}

func (c *Client) newAPIClient(baseURL string) (*client.ClientWithResponses, error) {
	return client.NewClientWithResponses(
		baseURL,
		client.WithHTTPClient(c.httpClient),
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			rewriteRequestPath(req, baseURL)
			req.Header.Set("Authorization", c.authHeader)
			if c.config.UserAgent != "" {
				req.Header.Set("User-Agent", c.config.UserAgent)
			}
			return nil
		}),
	)
}

func (c *Client) probeUsersMe(ctx context.Context, baseURL string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, joinBaseURL(baseURL, usersMePath), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Accept", "application/json")
	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func formatAuthHeader(token string, scheme string) string {
	trimmed := strings.TrimSpace(token)
	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "token "):
		return "Token " + strings.TrimSpace(trimmed[6:])
	case strings.HasPrefix(lower, "bearer "):
		return "Bearer " + strings.TrimSpace(trimmed[7:])
	}

	switch strings.ToLower(strings.TrimSpace(scheme)) {
	case "", "auto", "token":
		return "Token " + trimmed
	case "bearer":
		return "Bearer " + trimmed
	default:
		return strings.TrimSpace(scheme) + " " + trimmed
	}
}

func normalizeBaseURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return strings.TrimSpace(raw)
	}
	if parsed.Path == "" {
		parsed.Path = "/"
	} else if !strings.HasSuffix(parsed.Path, "/") {
		parsed.Path += "/"
	}
	return parsed.String()
}

func joinBaseURL(baseURL string, path string) string {
	if isAbsoluteURL(path) {
		return path
	}
	base := strings.TrimRight(baseURL, "/")
	relative := strings.TrimSpace(path)
	if relative == "" {
		return base
	}
	if !strings.HasPrefix(relative, "/") {
		relative = "/" + relative
	}
	return base + relative
}

func baseURLCandidates(rawBaseURL string, versionPaths []string) []string {
	base := normalizeBaseURL(rawBaseURL)
	candidates := []string{base}

	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return uniqueStrings(candidates)
	}

	paths := versionPaths
	if len(paths) == 0 {
		paths = DefaultConfig().VersionPaths
	}
	originalPath := strings.TrimRight(parsed.Path, "/")
	for _, versionPath := range paths {
		versionPath = "/" + strings.Trim(versionPath, "/")
		switch {
		case originalPath == "":
			parsedCopy := *parsed
			parsedCopy.Path = versionPath
			candidates = append(candidates, normalizeBaseURL(parsedCopy.String()))
		case originalPath != versionPath:
			if !strings.HasSuffix(originalPath, versionPath) {
				rootCopy := *parsed
				rootCopy.Path = versionPath
				candidates = append(candidates, normalizeBaseURL(rootCopy.String()))

				joinedCopy := *parsed
				joinedCopy.Path = strings.TrimRight(originalPath, "/") + versionPath
				candidates = append(candidates, normalizeBaseURL(joinedCopy.String()))
			}
		}
	}

	return uniqueStrings(candidates)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func isAbsoluteURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func rewriteRequestPath(req *http.Request, baseURL string) {
	basePath := strings.TrimSuffix((&url.URL{Path: strings.TrimSpace(mustParseURL(baseURL).Path)}).EscapedPath(), "/")
	if basePath == "" || basePath == "/" {
		return
	}
	if strings.HasPrefix(req.URL.Path, basePath+"/") || req.URL.Path == basePath {
		return
	}
	req.URL.Path = basePath + "/" + strings.TrimPrefix(req.URL.Path, "/")
	req.URL.RawPath = req.URL.EscapedPath()
}

func mustParseURL(raw string) *url.URL {
	parsed, err := url.Parse(raw)
	if err != nil {
		return &url.URL{}
	}
	return parsed
}
