package provider_daemon

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	portalauth "github.com/virtengine/virtengine/pkg/provider_daemon/auth"
)

// RateLimitConfig configures per-user rate limiting.
type RateLimitConfig struct {
	RequestsPerMinute int
}

type rateWindow struct {
	start time.Time
	count int
}

// RateLimiter enforces a simple fixed-window rate limit.
type PortalRateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	state  map[string]*rateWindow
}

// NewPortalRateLimiter creates a new rate limiter.
func NewPortalRateLimiter(limit int, window time.Duration) *PortalRateLimiter {
	if limit <= 0 {
		return nil
	}
	return &PortalRateLimiter{
		limit:  limit,
		window: window,
		state:  make(map[string]*rateWindow),
	}
}

// Allow checks if the key can proceed and returns remaining + reset time.
func (r *PortalRateLimiter) Allow(key string) (bool, int, time.Time) {
	if r == nil || r.limit <= 0 {
		return true, 0, time.Time{}
	}

	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	window := r.state[key]
	if window == nil || now.Sub(window.start) >= r.window {
		window = &rateWindow{start: now}
		r.state[key] = window
	}

	if window.count >= r.limit {
		reset := window.start.Add(r.window)
		return false, 0, reset
	}

	window.count++
	remaining := r.limit - window.count
	reset := window.start.Add(r.window)
	return true, remaining, reset
}

func (s *PortalAPIServer) rateLimitMiddleware() func(http.Handler) http.Handler {
	if s.rateLimiter == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := principalFromContext(r.Context())
			if key == "" {
				key = clientIP(r.RemoteAddr)
			}

			allowed, remaining, reset := s.rateLimiter.Allow(key)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(s.rateLimiter.limit))
			if !reset.IsZero() {
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
			}
			if allowed {
				w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Remaining", "0")
			writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
		})
	}
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func principalFromContext(ctx context.Context) string {
	return portalauth.FromContext(ctx).Address
}

func authFromContext(ctx context.Context) portalauth.AuthContext {
	return portalauth.FromContext(ctx)
}

func withAuth(ctx context.Context, auth portalauth.AuthContext) context.Context {
	if auth.Address == "" {
		return ctx
	}
	return portalauth.WithAuth(ctx, auth)
}

func parseLimit(r *http.Request, fallback, max int) int {
	val := strings.TrimSpace(r.URL.Query().Get("limit"))
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed <= 0 {
		return fallback
	}
	if max > 0 && parsed > max {
		return max
	}
	return parsed
}

func parseCursor(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get("cursor"))
}

func parseTimeRange(r *http.Request) (time.Time, time.Time, error) {
	startRaw := strings.TrimSpace(r.URL.Query().Get("start"))
	endRaw := strings.TrimSpace(r.URL.Query().Get("end"))
	if startRaw == "" || endRaw == "" {
		return time.Time{}, time.Time{}, errors.New("start and end are required")
	}
	start, err := time.Parse(time.RFC3339, startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid start time")
	}
	end, err := time.Parse(time.RFC3339, endRaw)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid end time")
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, errors.New("end must be after start")
	}
	return start, end, nil
}

func parseInterval(r *http.Request, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("interval"))
	if raw == "" {
		return fallback, nil
	}
	interval, err := time.ParseDuration(raw)
	if err != nil {
		return 0, errors.New("invalid interval")
	}
	if interval <= 0 {
		return 0, errors.New("interval must be positive")
	}
	return interval, nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func errorsIsNotFound(err error) bool {
	return err != nil && errors.Is(err, ErrPortalNotFound)
}
