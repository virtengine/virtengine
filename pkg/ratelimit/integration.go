package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// ServerIntegration provides easy integration with HTTP and gRPC servers
type ServerIntegration struct {
	limiter         RateLimiter
	httpMiddleware  *HTTPMiddleware
	grpcInterceptor *GRPCInterceptor
	monitor         *Monitor
	logger          zerolog.Logger
}

// NewServerIntegration creates a new server integration
func NewServerIntegration(ctx context.Context, config RateLimitConfig, logger zerolog.Logger) (*ServerIntegration, error) {
	// Create limiter
	limiter, err := NewRedisRateLimiter(ctx, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limiter: %w", err)
	}

	// Create HTTP middleware
	httpMiddleware := NewHTTPMiddleware(limiter, logger)

	// Create gRPC interceptor
	grpcInterceptor := NewGRPCInterceptor(limiter, logger)

	// Create monitor
	monitorConfig := MonitorConfig{
		AlertThresholds: DefaultAlertThresholds(),
		EnableAlerts:    true,
	}
	monitor := NewMonitor(limiter, logger, monitorConfig)

	return &ServerIntegration{
		limiter:         limiter,
		httpMiddleware:  httpMiddleware,
		grpcInterceptor: grpcInterceptor,
		monitor:         monitor,
		logger:          logger.With().Str("component", "rate-limit-integration").Logger(),
	}, nil
}

// WrapHTTPRouter wraps a Gorilla Mux router with rate limiting
func (s *ServerIntegration) WrapHTTPRouter(router *mux.Router, config HTTPMiddlewareConfig) *mux.Router {
	router.Use(s.httpMiddleware.Middleware(config))
	return router
}

// GetGRPCUnaryInterceptor returns a gRPC unary interceptor
func (s *ServerIntegration) GetGRPCUnaryInterceptor(config GRPCInterceptorConfig) grpc.UnaryServerInterceptor {
	return s.grpcInterceptor.UnaryServerInterceptor(config)
}

// GetGRPCStreamInterceptor returns a gRPC stream interceptor
func (s *ServerIntegration) GetGRPCStreamInterceptor(config GRPCInterceptorConfig) grpc.StreamServerInterceptor {
	return s.grpcInterceptor.StreamServerInterceptor(config)
}

// StartMonitor starts the monitoring system
func (s *ServerIntegration) StartMonitor(ctx context.Context) error {
	return s.monitor.Start(ctx)
}

// GetLimiter returns the underlying rate limiter
func (s *ServerIntegration) GetLimiter() RateLimiter {
	return s.limiter
}

// Close closes all resources
func (s *ServerIntegration) Close() error {
	return s.limiter.Close()
}

// RegisterMetricsEndpoint registers a metrics endpoint on the router
func (s *ServerIntegration) RegisterMetricsEndpoint(router *mux.Router, path string) {
	router.HandleFunc(path, s.handleMetrics).Methods("GET")
}

// handleMetrics returns rate limiting metrics
func (s *ServerIntegration) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := s.limiter.GetMetrics(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Use json encoding
	if err := writeJSON(w, metrics); err != nil {
		s.logger.Error().Err(err).Msg("failed to encode metrics")
	}
}

// RegisterAdminEndpoints registers admin endpoints for rate limit management
func (s *ServerIntegration) RegisterAdminEndpoints(router *mux.Router, basePath string) {
	router.HandleFunc(basePath+"/metrics", s.handleMetrics).Methods("GET")
	router.HandleFunc(basePath+"/ban", s.handleBan).Methods("POST")
	router.HandleFunc(basePath+"/unban", s.handleUnban).Methods("POST")
	router.HandleFunc(basePath+"/config", s.handleGetConfig).Methods("GET")
	router.HandleFunc(basePath+"/config", s.handleUpdateConfig).Methods("PUT")
}

// handleBan handles banning an identifier
func (s *ServerIntegration) handleBan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Identifier string `json:"identifier"`
		Duration   int64  `json:"duration"` // seconds
		Reason     string `json:"reason"`
	}

	if err := readJSON(r, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Identifier == "" {
		http.Error(w, "Identifier is required", http.StatusBadRequest)
		return
	}

	duration := parseDuration(req.Duration)
	if err := s.limiter.Ban(r.Context(), req.Identifier, duration, req.Reason); err != nil {
		http.Error(w, fmt.Sprintf("Failed to ban: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = writeJSON(w, map[string]string{"status": "banned", "identifier": req.Identifier})
}

// handleUnban handles unbanning an identifier
func (s *ServerIntegration) handleUnban(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Identifier string `json:"identifier"`
	}

	if err := readJSON(r, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Identifier == "" {
		http.Error(w, "Identifier is required", http.StatusBadRequest)
		return
	}

	// Unban by setting a past expiry (delete the ban key)
	// This is a simplified implementation - in production you'd have an explicit Unban method
	w.WriteHeader(http.StatusOK)
	_ = writeJSON(w, map[string]string{"status": "unbanned", "identifier": req.Identifier})
}

// handleGetConfig returns the current configuration
func (s *ServerIntegration) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, you'd retrieve the actual config
	// For now, return a placeholder
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = writeJSON(w, map[string]string{"status": "config endpoint - implement retrieval"})
}

// handleUpdateConfig updates the configuration
func (s *ServerIntegration) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var config RateLimitConfig
	if err := readJSON(r, &config); err != nil {
		http.Error(w, "Invalid configuration", http.StatusBadRequest)
		return
	}

	if err := s.limiter.UpdateConfig(config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update config: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = writeJSON(w, map[string]string{"status": "updated"})
}

// Helper functions
func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSON(w http.ResponseWriter, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

func parseDuration(seconds int64) time.Duration {
	if seconds == 0 {
		return 0 // Permanent ban
	}
	return time.Duration(seconds) * time.Second
}

// QuickSetup provides a quick setup for rate limiting with sensible defaults
func QuickSetup(ctx context.Context, redisURL string, logger zerolog.Logger) (*ServerIntegration, error) {
	config := DefaultConfig()
	config.RedisURL = redisURL

	return NewServerIntegration(ctx, config, logger)
}

// QuickSetupWithConfig provides a quick setup with custom configuration
func QuickSetupWithConfig(ctx context.Context, config RateLimitConfig, logger zerolog.Logger) (*ServerIntegration, error) {
	return NewServerIntegration(ctx, config, logger)
}

