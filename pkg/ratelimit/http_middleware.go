package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
)

// HTTPMiddleware provides rate limiting for HTTP endpoints
type HTTPMiddleware struct {
	limiter RateLimiter
	logger  zerolog.Logger

	// Prometheus metrics
	requestsTotal   *prometheus.CounterVec
	requestsBlocked *prometheus.CounterVec
	requestLatency  *prometheus.HistogramVec
}

// HTTPMiddlewareConfig configures the HTTP middleware
type HTTPMiddlewareConfig struct {
	// IdentifierExtractor extracts the identifier from the request
	// If nil, uses default (IP address)
	IdentifierExtractor func(*http.Request) string

	// UserExtractor extracts the user ID from the request
	// If nil, no user-based rate limiting is applied
	UserExtractor func(*http.Request) string

	// SkipPaths is a list of paths to skip rate limiting
	SkipPaths []string

	// OnRateLimited is called when a request is rate limited
	OnRateLimited func(http.ResponseWriter, *http.Request, *RateLimitResult)
}

// NewHTTPMiddleware creates a new HTTP rate limiting middleware
func NewHTTPMiddleware(limiter RateLimiter, logger zerolog.Logger) *HTTPMiddleware {
	return &HTTPMiddleware{
		limiter: limiter,
		logger:  logger.With().Str("component", "http-middleware").Logger(),
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "virtengine_http_requests_total",
				Help: "Total HTTP requests processed by rate limiter",
			},
			[]string{"method", "path", "status"},
		),
		requestsBlocked: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "virtengine_http_requests_blocked_total",
				Help: "Total HTTP requests blocked by rate limiter",
			},
			[]string{"method", "path", "reason"},
		),
		requestLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "virtengine_http_ratelimit_check_duration_seconds",
				Help:    "Duration of rate limit checks",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"limit_type"},
		),
	}
}

// Middleware returns an HTTP middleware function
func (m *HTTPMiddleware) Middleware(config HTTPMiddlewareConfig) func(http.Handler) http.Handler {
	if config.IdentifierExtractor == nil {
		config.IdentifierExtractor = extractIPAddress
	}

	if config.OnRateLimited == nil {
		config.OnRateLimited = m.defaultOnRateLimited
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Skip rate limiting for certain paths
			if m.shouldSkip(r.URL.Path, config.SkipPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract identifiers
			ipAddr := config.IdentifierExtractor(r)
			userID := ""
			if config.UserExtractor != nil {
				userID = config.UserExtractor(r)
			}

			// Check IP-based rate limit
			allowed, result, err := m.checkIPLimit(ctx, ipAddr, r.URL.Path)
			if err != nil {
				m.logger.Error().Err(err).Str("ip", ipAddr).Msg("rate limit check failed")
				// On error, allow the request but log
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				m.handleRateLimited(w, r, result, config.OnRateLimited)
				return
			}

			// Check user-based rate limit if user is identified
			if userID != "" {
				allowed, result, err := m.checkUserLimit(ctx, userID, r.URL.Path)
				if err != nil {
					m.logger.Error().Err(err).Str("user", userID).Msg("user rate limit check failed")
					// On error, allow the request but log
					next.ServeHTTP(w, r)
					return
				}

				if !allowed {
					m.handleRateLimited(w, r, result, config.OnRateLimited)
					return
				}
			}

			// Add rate limit headers to response
			m.addRateLimitHeaders(w, result)

			// Record metrics
			m.requestsTotal.WithLabelValues(r.Method, r.URL.Path, "allowed").Inc()

			next.ServeHTTP(w, r)
		})
	}
}

// checkIPLimit checks IP-based rate limit
func (m *HTTPMiddleware) checkIPLimit(ctx context.Context, ipAddr string, endpoint string) (bool, *RateLimitResult, error) {
	timer := prometheus.NewTimer(m.requestLatency.WithLabelValues(string(LimitTypeIP)))
	defer timer.ObserveDuration()

	return m.limiter.AllowEndpoint(ctx, endpoint, ipAddr, LimitTypeIP)
}

// checkUserLimit checks user-based rate limit
func (m *HTTPMiddleware) checkUserLimit(ctx context.Context, userID string, endpoint string) (bool, *RateLimitResult, error) {
	timer := prometheus.NewTimer(m.requestLatency.WithLabelValues(string(LimitTypeUser)))
	defer timer.ObserveDuration()

	return m.limiter.AllowEndpoint(ctx, endpoint, userID, LimitTypeUser)
}

// handleRateLimited handles a rate-limited request
func (m *HTTPMiddleware) handleRateLimited(w http.ResponseWriter, r *http.Request, result *RateLimitResult, handler func(http.ResponseWriter, *http.Request, *RateLimitResult)) {
	m.logger.Warn().
		Str("ip", extractIPAddress(r)).
		Str("path", r.URL.Path).
		Str("method", r.Method).
		Str("limit_type", string(result.LimitType)).
		Int("remaining", result.Remaining).
		Int("retry_after", result.RetryAfter).
		Msg("request rate limited")

	m.requestsBlocked.WithLabelValues(r.Method, r.URL.Path, string(result.LimitType)).Inc()

	handler(w, r, result)
}

// defaultOnRateLimited is the default rate limit handler
func (m *HTTPMiddleware) defaultOnRateLimited(w http.ResponseWriter, r *http.Request, result *RateLimitResult) {
	m.addRateLimitHeaders(w, result)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)

	response := map[string]interface{}{
		"error":       "rate_limit_exceeded",
		"message":     "Too many requests. Please try again later.",
		"limit":       result.Limit,
		"remaining":   result.Remaining,
		"retry_after": result.RetryAfter,
		"reset_at":    result.ResetAt,
	}

	//nolint:errchkjson // HTTP response encoding
	_ = json.NewEncoder(w).Encode(response)
}

// addRateLimitHeaders adds standard rate limit headers to the response
func (m *HTTPMiddleware) addRateLimitHeaders(w http.ResponseWriter, result *RateLimitResult) {
	if result == nil {
		return
	}

	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetAt.Unix()))

	if !result.Allowed {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", result.RetryAfter))
	}
}

// shouldSkip checks if rate limiting should be skipped for a path
func (m *HTTPMiddleware) shouldSkip(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// extractIPAddress extracts the IP address from the request
func extractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			// Get the first IP (client's real IP)
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// ExtractUserFromJWT extracts user ID from JWT token in Authorization header
func ExtractUserFromJWT(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	// Remove "Bearer " prefix
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		// In production, you would validate and parse the JWT here
		// For now, we just return a hash of the token as the user ID
		return fmt.Sprintf("jwt:%s", hashString(token))
	}

	return ""
}

// ExtractUserFromHeader extracts user ID from a custom header
func ExtractUserFromHeader(headerName string) func(*http.Request) string {
	return func(r *http.Request) string {
		return r.Header.Get(headerName)
	}
}

// ExtractUserFromQuery extracts user ID from a query parameter
func ExtractUserFromQuery(paramName string) func(*http.Request) string {
	return func(r *http.Request) string {
		return r.URL.Query().Get(paramName)
	}
}

// hashString creates a simple hash of a string (for user identification)
func hashString(s string) string {
	if len(s) > 32 {
		return s[:32]
	}
	return s
}

// WrapHandler wraps an http.Handler with rate limiting
func (m *HTTPMiddleware) WrapHandler(handler http.Handler, config HTTPMiddlewareConfig) http.Handler {
	return m.Middleware(config)(handler)
}

// WrapHandlerFunc wraps an http.HandlerFunc with rate limiting
func (m *HTTPMiddleware) WrapHandlerFunc(handler http.HandlerFunc, config HTTPMiddlewareConfig) http.HandlerFunc {
	return m.Middleware(config)(handler).ServeHTTP
}

// HealthCheckMiddleware creates a middleware that skips rate limiting for health checks
func HealthCheckMiddleware(healthCheckPaths ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range healthCheckPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
