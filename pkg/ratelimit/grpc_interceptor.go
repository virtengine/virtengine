package ratelimit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// GRPCInterceptor provides rate limiting for gRPC endpoints
type GRPCInterceptor struct {
	limiter RateLimiter
	logger  zerolog.Logger

	// Prometheus metrics
	requestsTotal   *prometheus.CounterVec
	requestsBlocked *prometheus.CounterVec
	requestLatency  *prometheus.HistogramVec
}

// GRPCInterceptorConfig configures the gRPC interceptor
type GRPCInterceptorConfig struct {
	// IdentifierExtractor extracts the identifier from the context
	// If nil, uses default (IP address from peer)
	IdentifierExtractor func(context.Context) string

	// UserExtractor extracts the user ID from the context
	// If nil, no user-based rate limiting is applied
	UserExtractor func(context.Context) string

	// SkipMethods is a list of methods to skip rate limiting
	SkipMethods []string
}

// NewGRPCInterceptor creates a new gRPC rate limiting interceptor
func NewGRPCInterceptor(limiter RateLimiter, logger zerolog.Logger) *GRPCInterceptor {
	return &GRPCInterceptor{
		limiter: limiter,
		logger:  logger.With().Str("component", "grpc-interceptor").Logger(),
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "virtengine_grpc_requests_total",
				Help: "Total gRPC requests processed by rate limiter",
			},
			[]string{"method", "status"},
		),
		requestsBlocked: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "virtengine_grpc_requests_blocked_total",
				Help: "Total gRPC requests blocked by rate limiter",
			},
			[]string{"method", "reason"},
		),
		requestLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "virtengine_grpc_ratelimit_check_duration_seconds",
				Help:    "Duration of rate limit checks",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"limit_type"},
		),
	}
}

// UnaryServerInterceptor returns a gRPC unary server interceptor
func (i *GRPCInterceptor) UnaryServerInterceptor(config GRPCInterceptorConfig) grpc.UnaryServerInterceptor {
	if config.IdentifierExtractor == nil {
		config.IdentifierExtractor = extractPeerAddress
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip rate limiting for certain methods
		if i.shouldSkip(info.FullMethod, config.SkipMethods) {
			return handler(ctx, req)
		}

		// Extract identifiers
		ipAddr := config.IdentifierExtractor(ctx)
		userID := ""
		if config.UserExtractor != nil {
			userID = config.UserExtractor(ctx)
		}

		// Check IP-based rate limit
		allowed, result, err := i.checkIPLimit(ctx, ipAddr, info.FullMethod)
		if err != nil {
			i.logger.Error().Err(err).Str("ip", ipAddr).Str("method", info.FullMethod).Msg("rate limit check failed")
			// On error, allow the request but log
			return handler(ctx, req)
		}

		if !allowed {
			return nil, i.handleRateLimited(ctx, info.FullMethod, result)
		}

		// Check user-based rate limit if user is identified
		if userID != "" {
			allowed, result, err := i.checkUserLimit(ctx, userID, info.FullMethod)
			if err != nil {
				i.logger.Error().Err(err).Str("user", userID).Str("method", info.FullMethod).Msg("user rate limit check failed")
				// On error, allow the request but log
				return handler(ctx, req)
			}

			if !allowed {
				return nil, i.handleRateLimited(ctx, info.FullMethod, result)
			}
		}

		// Add rate limit metadata to response
		ctx = i.addRateLimitMetadata(ctx, result)

		// Record metrics
		i.requestsTotal.WithLabelValues(info.FullMethod, "allowed").Inc()

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor
func (i *GRPCInterceptor) StreamServerInterceptor(config GRPCInterceptorConfig) grpc.StreamServerInterceptor {
	if config.IdentifierExtractor == nil {
		config.IdentifierExtractor = extractPeerAddress
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// Skip rate limiting for certain methods
		if i.shouldSkip(info.FullMethod, config.SkipMethods) {
			return handler(srv, ss)
		}

		// Extract identifiers
		ipAddr := config.IdentifierExtractor(ctx)
		userID := ""
		if config.UserExtractor != nil {
			userID = config.UserExtractor(ctx)
		}

		// Check IP-based rate limit
		allowed, result, err := i.checkIPLimit(ctx, ipAddr, info.FullMethod)
		if err != nil {
			i.logger.Error().Err(err).Str("ip", ipAddr).Str("method", info.FullMethod).Msg("rate limit check failed")
			// On error, allow the request but log
			return handler(srv, ss)
		}

		if !allowed {
			return i.handleRateLimited(ctx, info.FullMethod, result)
		}

		// Check user-based rate limit if user is identified
		if userID != "" {
			allowed, result, err := i.checkUserLimit(ctx, userID, info.FullMethod)
			if err != nil {
				i.logger.Error().Err(err).Str("user", userID).Str("method", info.FullMethod).Msg("user rate limit check failed")
				// On error, allow the request but log
				return handler(srv, ss)
			}

			if !allowed {
				return i.handleRateLimited(ctx, info.FullMethod, result)
			}
		}

		// Record metrics
		i.requestsTotal.WithLabelValues(info.FullMethod, "allowed").Inc()

		// Wrap stream to add rate limit headers
		wrappedStream := &rateLimitedServerStream{
			ServerStream: ss,
			ctx:          i.addRateLimitMetadata(ctx, result),
		}

		return handler(srv, wrappedStream)
	}
}

// checkIPLimit checks IP-based rate limit
func (i *GRPCInterceptor) checkIPLimit(ctx context.Context, ipAddr string, method string) (bool, *RateLimitResult, error) {
	timer := prometheus.NewTimer(i.requestLatency.WithLabelValues(string(LimitTypeIP)))
	defer timer.ObserveDuration()

	return i.limiter.AllowEndpoint(ctx, method, ipAddr, LimitTypeIP)
}

// checkUserLimit checks user-based rate limit
func (i *GRPCInterceptor) checkUserLimit(ctx context.Context, userID string, method string) (bool, *RateLimitResult, error) {
	timer := prometheus.NewTimer(i.requestLatency.WithLabelValues(string(LimitTypeUser)))
	defer timer.ObserveDuration()

	return i.limiter.AllowEndpoint(ctx, method, userID, LimitTypeUser)
}

// handleRateLimited handles a rate-limited request
func (i *GRPCInterceptor) handleRateLimited(ctx context.Context, method string, result *RateLimitResult) error {
	ipAddr := extractPeerAddress(ctx)

	i.logger.Warn().
		Str("ip", ipAddr).
		Str("method", method).
		Str("limit_type", string(result.LimitType)).
		Int("remaining", result.Remaining).
		Int("retry_after", result.RetryAfter).
		Msg("gRPC request rate limited")

	i.requestsBlocked.WithLabelValues(method, string(result.LimitType)).Inc()

	// Add rate limit metadata to error
	md := metadata.Pairs(
		"x-ratelimit-limit", fmt.Sprintf("%d", result.Limit),
		"x-ratelimit-remaining", fmt.Sprintf("%d", result.Remaining),
		"x-ratelimit-reset", fmt.Sprintf("%d", result.ResetAt.Unix()),
		"retry-after", fmt.Sprintf("%d", result.RetryAfter),
	)

	if err := grpc.SetHeader(ctx, md); err != nil {
		i.logger.Error().Err(err).Msg("failed to set rate limit headers")
	}

	return status.Errorf(codes.ResourceExhausted,
		"rate limit exceeded: %d requests allowed, retry after %d seconds",
		result.Limit, result.RetryAfter)
}

// addRateLimitMetadata adds rate limit metadata to the context
func (i *GRPCInterceptor) addRateLimitMetadata(ctx context.Context, result *RateLimitResult) context.Context {
	if result == nil {
		return ctx
	}

	md := metadata.Pairs(
		"x-ratelimit-limit", fmt.Sprintf("%d", result.Limit),
		"x-ratelimit-remaining", fmt.Sprintf("%d", result.Remaining),
		"x-ratelimit-reset", fmt.Sprintf("%d", result.ResetAt.Unix()),
	)

	if err := grpc.SetHeader(ctx, md); err != nil {
		i.logger.Error().Err(err).Msg("failed to set rate limit headers")
	}

	return ctx
}

// shouldSkip checks if rate limiting should be skipped for a method
func (i *GRPCInterceptor) shouldSkip(method string, skipMethods []string) bool {
	for _, skipMethod := range skipMethods {
		if strings.HasPrefix(method, skipMethod) {
			return true
		}
	}
	return false
}

// extractPeerAddress extracts the peer address from the context
func extractPeerAddress(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "unknown"
	}

	// Extract IP from peer address
	addr := p.Addr.String()
	// addr is typically "ip:port", so split it
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}

	return addr
}

// ExtractUserFromGRPCMetadata extracts user ID from gRPC metadata
func ExtractUserFromGRPCMetadata(key string) func(context.Context) string {
	return func(ctx context.Context) string {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return ""
		}

		values := md.Get(key)
		if len(values) == 0 {
			return ""
		}

		return values[0]
	}
}

// ExtractUserFromAuthToken extracts user ID from authorization token in metadata
func ExtractUserFromAuthToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	authValues := md.Get("authorization")
	if len(authValues) == 0 {
		return ""
	}

	auth := authValues[0]
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		// In production, you would validate and parse the JWT here
		// For now, we just return a hash of the token as the user ID
		return fmt.Sprintf("jwt:%s", hashString(token))
	}

	return ""
}

// rateLimitedServerStream wraps grpc.ServerStream to provide custom context
type rateLimitedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *rateLimitedServerStream) Context() context.Context {
	return s.ctx
}

// ChainUnaryInterceptors chains multiple unary interceptors
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chain
			chain = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return interceptor(currentCtx, currentReq, info, next)
			}
		}
		return chain(ctx, req)
	}
}

// ChainStreamInterceptors chains multiple stream interceptors
func ChainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chain
			chain = func(currentSrv interface{}, currentSS grpc.ServerStream) error {
				return interceptor(currentSrv, currentSS, info, next)
			}
		}
		return chain(srv, ss)
	}
}

// RateLimitInfo extracts rate limit information from gRPC metadata
func RateLimitInfo(ctx context.Context) *RateLimitResult {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	result := &RateLimitResult{}

	if values := md.Get("x-ratelimit-limit"); len(values) > 0 {
		fmt.Sscanf(values[0], "%d", &result.Limit)
	}

	if values := md.Get("x-ratelimit-remaining"); len(values) > 0 {
		fmt.Sscanf(values[0], "%d", &result.Remaining)
	}

	if values := md.Get("x-ratelimit-reset"); len(values) > 0 {
		var resetUnix int64
		fmt.Sscanf(values[0], "%d", &resetUnix)
		result.ResetAt = time.Unix(resetUnix, 0)
	}

	return result
}

