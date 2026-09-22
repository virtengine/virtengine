package ratelimit

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// fakeLimiter is an in-memory RateLimiter for offline unit tests.
type fakeLimiter struct {
	mu            sync.Mutex
	allowEndpoint func(endpoint, identifier string, limitType LimitType) (bool, *RateLimitResult, error)
	metrics       *Metrics
	metricsErr    error
	banErr        error
	updateErr     error
	banned        map[string]bool
	closed        bool
}

func newFakeLimiter() *fakeLimiter {
	return &fakeLimiter{
		banned: map[string]bool{},
		metrics: &Metrics{
			TotalRequests:   100,
			AllowedRequests: 90,
			BlockedRequests: 10,
			ByLimitType:     map[LimitType]*LimitTypeMetrics{},
		},
	}
}

func allowResult(id string, lt LimitType) *RateLimitResult {
	return &RateLimitResult{
		Allowed:    true,
		Limit:      10,
		Remaining:  9,
		ResetAt:    time.Now().Add(time.Minute),
		LimitType:  lt,
		Identifier: id,
	}
}

func denyResult(id string, lt LimitType) *RateLimitResult {
	return &RateLimitResult{
		Allowed:    false,
		Limit:      10,
		Remaining:  0,
		RetryAfter: 30,
		ResetAt:    time.Now().Add(time.Minute),
		LimitType:  lt,
		Identifier: id,
	}
}

func (f *fakeLimiter) Allow(_ context.Context, key string, limitType LimitType) (bool, *RateLimitResult, error) {
	return f.AllowEndpoint(context.Background(), "test", key, limitType)
}

func (f *fakeLimiter) AllowEndpoint(_ context.Context, endpoint, identifier string, limitType LimitType) (bool, *RateLimitResult, error) {
	if f.allowEndpoint != nil {
		return f.allowEndpoint(endpoint, identifier, limitType)
	}
	return true, allowResult(identifier, limitType), nil
}

func (f *fakeLimiter) RecordBypassAttempt(_ context.Context, _ string, _ string) error {
	return nil
}

func (f *fakeLimiter) GetMetrics(_ context.Context) (*Metrics, error) {
	if f.metricsErr != nil {
		return nil, f.metricsErr
	}
	return f.metrics, nil
}

func (f *fakeLimiter) UpdateConfig(_ RateLimitConfig) error {
	return f.updateErr
}

func (f *fakeLimiter) IsWhitelisted(_ string, _ LimitType) bool {
	return false
}

func (f *fakeLimiter) Ban(_ context.Context, identifier string, _ time.Duration, _ string) error {
	if f.banErr != nil {
		return f.banErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.banned[identifier] = true
	return nil
}

func (f *fakeLimiter) IsBanned(_ context.Context, identifier string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.banned[identifier], nil
}

func (f *fakeLimiter) GetCurrentLoad(_ context.Context) (float64, error) {
	return 12.5, nil
}

func (f *fakeLimiter) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func testLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}

// Shared constructors: promauto panics on duplicate registration, so each
// constructor is called exactly once per test binary and shared by subtests.
var (
	sharedHTTPMiddleware  = NewHTTPMiddleware(newFakeLimiter(), testLogger())
	sharedGRPCInterceptor = NewGRPCInterceptor(newFakeLimiter(), testLogger())
	sharedMonitor         = NewMonitor(newFakeLimiter(), testLogger(), MonitorConfig{
		MetricsInterval: 10 * time.Millisecond,
		EnableAlerts:    true,
		AlertThresholds: AlertThresholds{
			BlockedRequestsPerMinute: 5,
			BypassAttemptsPerMinute:  5,
			BannedIdentifiersCount:   2,
			LoadPercentage:           50.0,
			TopBlockedIPRequests:     3,
		},
	})
)

// mwWithFake clones the shared middleware (reusing its promauto metrics)
// with a controllable fake limiter.
func mwWithFake(f *fakeLimiter) *HTTPMiddleware {
	m := *sharedHTTPMiddleware
	m.limiter = f
	return &m
}

// icWithFake clones the shared interceptor (reusing its promauto metrics)
// with a controllable fake limiter.
func icWithFake(f *fakeLimiter) *GRPCInterceptor {
	i := *sharedGRPCInterceptor
	i.limiter = f
	return &i
}

func okHandler(msg string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(msg))
	})
}

func TestHTTPMiddlewareAllow(t *testing.T) {
	h := sharedHTTPMiddleware.WrapHandler(okHandler("hello"), HTTPMiddlewareConfig{})
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "hello", rec.Body.String())
	assert.Equal(t, "10", rec.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"))
	assert.Empty(t, rec.Header().Get("Retry-After"))
}

func TestHTTPMiddlewareSkipPaths(t *testing.T) {
	// Skip logic bypasses the limiter entirely, so the shared middleware
	// (single promauto registration) exercises it without Redis.
	h := sharedHTTPMiddleware.WrapHandler(okHandler("skipped"), HTTPMiddlewareConfig{
		SkipPaths: []string{"/health"},
	})
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "skipped", rec.Body.String())
}

func TestHTTPMiddlewareBlocked(t *testing.T) {
	mw := mwWithFake(&fakeLimiter{
		allowEndpoint: func(_, _ string, _ LimitType) (bool, *RateLimitResult, error) {
			return false, denyResult("10.0.0.9", LimitTypeIP), nil
		},
	})
	h := mw.Middleware(HTTPMiddlewareConfig{})(okHandler("never"))
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.RemoteAddr = "10.0.0.9:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "30", rec.Header().Get("Retry-After"))
	assert.Contains(t, rec.Body.String(), "rate_limit_exceeded")
}

func TestHTTPMiddlewareLimiterErrorPassesThrough(t *testing.T) {
	mw := mwWithFake(&fakeLimiter{
		allowEndpoint: func(_, _ string, _ LimitType) (bool, *RateLimitResult, error) {
			return false, nil, errors.New("limiter down")
		},
	})
	h := mw.Middleware(HTTPMiddlewareConfig{})(okHandler("passthrough"))
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.RemoteAddr = "10.0.0.9:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHTTPMiddlewareUserLimit(t *testing.T) {
	mw := mwWithFake(&fakeLimiter{
		allowEndpoint: func(_, id string, lt LimitType) (bool, *RateLimitResult, error) {
			if lt == LimitTypeUser {
				return false, denyResult(id, lt), nil
			}
			return true, allowResult(id, lt), nil
		},
	})

	t.Run("user blocked", func(t *testing.T) {
		h := mw.Middleware(HTTPMiddlewareConfig{
			UserExtractor: ExtractUserFromHeader("X-User-ID"),
		})(okHandler("never"))
		req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		req.RemoteAddr = "10.0.0.1:1"
		req.Header.Set("X-User-ID", "user-123")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	})

	t.Run("user limiter error passes through", func(t *testing.T) {
		errLimiter := &fakeLimiter{allowEndpoint: func(_, _ string, lt LimitType) (bool, *RateLimitResult, error) {
			if lt == LimitTypeUser {
				return false, nil, errors.New("user limiter down")
			}
			return true, allowResult("x", lt), nil
		}}
		mw2 := mwWithFake(errLimiter)
		h := mw2.Middleware(HTTPMiddlewareConfig{
			UserExtractor: ExtractUserFromHeader("X-User-ID"),
		})(okHandler("passthrough"))
		req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		req.RemoteAddr = "10.0.0.1:1"
		req.Header.Set("X-User-ID", "user-123")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("custom handler and extractor", func(t *testing.T) {
		called := false
		h := mw.Middleware(HTTPMiddlewareConfig{
			IdentifierExtractor: func(_ *http.Request) string { return "custom-id" },
			UserExtractor:       ExtractUserFromQuery("user"),
			OnRateLimited: func(w http.ResponseWriter, _ *http.Request, _ *RateLimitResult) {
				called = true
				w.WriteHeader(http.StatusForbidden)
			},
		})(okHandler("never"))
		req := httptest.NewRequest(http.MethodGet, "/api/data?user=u1", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.True(t, called)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestExtractors(t *testing.T) {
	t.Run("X-Forwarded-For first IP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", " 1.1.1.1, 2.2.2.2")
		assert.Equal(t, "1.1.1.1", extractIPAddress(req))
	})
	t.Run("X-Real-IP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-IP", "3.3.3.3")
		assert.Equal(t, "3.3.3.3", extractIPAddress(req))
	})
	t.Run("RemoteAddr with port", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "4.4.4.4:5678"
		assert.Equal(t, "4.4.4.4", extractIPAddress(req))
	})
	t.Run("RemoteAddr without port", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "no-port-here"
		// "no-port-here" contains no colon... SplitHostPort fails -> as-is
		assert.Equal(t, "no-port-here", extractIPAddress(req))
	})
	t.Run("JWT extractor", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, "", ExtractUserFromJWT(req))
		req.Header.Set("Authorization", "Basic abc")
		assert.Equal(t, "", ExtractUserFromJWT(req))
		req.Header.Set("Authorization", "Bearer mytoken123")
		assert.Equal(t, "jwt:mytoken123", ExtractUserFromJWT(req))
		req.Header.Set("Authorization", "Bearer a-very-long-token-value-that-exceeds-thirty-two-chars")
		assert.Equal(t, "jwt:a-very-long-token-value-that-exc", ExtractUserFromJWT(req))
	})
	t.Run("header and query extractors", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?user=q1", nil)
		req.Header.Set("X-User", "h1")
		assert.Equal(t, "h1", ExtractUserFromHeader("X-User")(req))
		assert.Equal(t, "q1", ExtractUserFromQuery("user")(req))
		assert.Equal(t, "", ExtractUserFromQuery("missing")(req))
	})
	t.Run("hashString", func(t *testing.T) {
		assert.Equal(t, "short", hashString("short"))
		assert.Equal(t, strings.Repeat("a", 32), hashString(strings.Repeat("a", 64)))
	})
}

func TestWrapHelpers(t *testing.T) {
	h := sharedHTTPMiddleware.WrapHandler(okHandler("wrapped"), HTTPMiddlewareConfig{})
	req := httptest.NewRequest(http.MethodGet, "/w", nil)
	req.RemoteAddr = "5.5.5.5:1"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	hf := sharedHTTPMiddleware.WrapHandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("func"))
	}, HTTPMiddlewareConfig{})
	rec2 := httptest.NewRecorder()
	hf.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/w", nil))
	assert.Equal(t, "func", rec2.Body.String())

	health := HealthCheckMiddleware("/healthz")(okHandler("next"))
	reqH := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recH := httptest.NewRecorder()
	health.ServeHTTP(recH, reqH)
	assert.Equal(t, "next", recH.Body.String())

	reqO := httptest.NewRequest(http.MethodGet, "/other", nil)
	recO := httptest.NewRecorder()
	health.ServeHTTP(recO, reqO)
	assert.Equal(t, "next", recO.Body.String())
}

func TestAddRateLimitHeadersNil(t *testing.T) {
	rec := httptest.NewRecorder()
	sharedHTTPMiddleware.addRateLimitHeaders(rec, nil)
	assert.Empty(t, rec.Header().Get("X-RateLimit-Limit"))
	rec2 := httptest.NewRecorder()
	sharedHTTPMiddleware.addRateLimitHeaders(rec2, allowResult("x", LimitTypeIP))
	assert.Empty(t, rec2.Header().Get("Retry-After"))
}

func TestShouldSkipHTTP(t *testing.T) {
	assert.True(t, sharedHTTPMiddleware.shouldSkip("/health/live", []string{"/health"}))
	assert.False(t, sharedHTTPMiddleware.shouldSkip("/api", []string{"/health"}))
	assert.False(t, sharedHTTPMiddleware.shouldSkip("/api", nil))
}

// ---- gRPC interceptor ----

func unaryHandlerOK(_ context.Context, _ interface{}) (interface{}, error) {
	return "ok-response", nil
}

func TestGRPCUnaryAllow(t *testing.T) {
	ic := sharedGRPCInterceptor.UnaryServerInterceptor(GRPCInterceptorConfig{})
	resp, err := ic(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, unaryHandlerOK)
	require.NoError(t, err)
	assert.Equal(t, "ok-response", resp)
}

func TestGRPCUnarySkip(t *testing.T) {
	ic := sharedGRPCInterceptor.UnaryServerInterceptor(GRPCInterceptorConfig{
		SkipMethods: []string{"/svc/Skipped"},
	})
	resp, err := ic(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Skipped.Foo"}, unaryHandlerOK)
	require.NoError(t, err)
	assert.Equal(t, "ok-response", resp)
}

func TestGRPCUnaryBlockedCode(t *testing.T) {
	ic := icWithFake(&fakeLimiter{
		allowEndpoint: func(_, _ string, _ LimitType) (bool, *RateLimitResult, error) {
			return false, denyResult("peer", LimitTypeIP), nil
		},
	})
	interceptor := ic.UnaryServerInterceptor(GRPCInterceptorConfig{})
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, unaryHandlerOK)
	require.Error(t, err)
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
	assert.Contains(t, err.Error(), "rate limit exceeded")
}

func TestGRPCUnaryLimiterError(t *testing.T) {
	ic := icWithFake(&fakeLimiter{
		allowEndpoint: func(_, _ string, _ LimitType) (bool, *RateLimitResult, error) {
			return false, nil, errors.New("down")
		},
	})
	interceptor := ic.UnaryServerInterceptor(GRPCInterceptorConfig{})
	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, unaryHandlerOK)
	require.NoError(t, err)
	assert.Equal(t, "ok-response", resp)
}

func TestGRPCUnaryUserPaths(t *testing.T) {
	mk := func(lt LimitType, err error) *GRPCInterceptor {
		return icWithFake(&fakeLimiter{
			allowEndpoint: func(_, _ string, got LimitType) (bool, *RateLimitResult, error) {
				if got == lt {
					if err != nil {
						return false, nil, err
					}
					return false, denyResult("u", got), nil
				}
				return true, allowResult("x", got), nil
			},
		})
	}
	userCfg := GRPCInterceptorConfig{UserExtractor: func(context.Context) string { return "user-1" }}

	t.Run("user blocked", func(t *testing.T) {
		ic := mk(LimitTypeUser, nil).UnaryServerInterceptor(userCfg)
		_, err := ic(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/m"}, unaryHandlerOK)
		require.Error(t, err)
		assert.Equal(t, codes.ResourceExhausted, status.Code(err))
	})
	t.Run("user limiter error passes through", func(t *testing.T) {
		ic := mk(LimitTypeUser, errors.New("down")).UnaryServerInterceptor(userCfg)
		resp, err := ic(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/m"}, unaryHandlerOK)
		require.NoError(t, err)
		assert.Equal(t, "ok-response", resp)
	})
}

type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (f *fakeServerStream) Context() context.Context { return f.ctx }

func TestGRPCStreamAllow(t *testing.T) {
	ic := sharedGRPCInterceptor.StreamServerInterceptor(GRPCInterceptorConfig{})
	var gotCtx context.Context
	ss := &fakeServerStream{ctx: context.Background()}
	err := ic(nil, ss, &grpc.StreamServerInfo{FullMethod: "/svc/Stream"}, func(_ interface{}, stream grpc.ServerStream) error {
		gotCtx = stream.Context()
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, gotCtx)
}

func TestGRPCStreamVariants(t *testing.T) {
	t.Run("skip", func(t *testing.T) {
		ic := sharedGRPCInterceptor.StreamServerInterceptor(GRPCInterceptorConfig{
			SkipMethods: []string{"/svc/Skip"},
		})
		ss := &fakeServerStream{ctx: context.Background()}
		err := ic(nil, ss, &grpc.StreamServerInfo{FullMethod: "/svc/Skip.It"}, func(_ interface{}, _ grpc.ServerStream) error {
			return nil
		})
		require.NoError(t, err)
	})
	t.Run("blocked", func(t *testing.T) {
		ic := icWithFake(&fakeLimiter{
			allowEndpoint: func(_, _ string, _ LimitType) (bool, *RateLimitResult, error) {
				return false, denyResult("p", LimitTypeIP), nil
			},
		})
		interceptor := ic.StreamServerInterceptor(GRPCInterceptorConfig{})
		ss := &fakeServerStream{ctx: context.Background()}
		err := interceptor(nil, ss, &grpc.StreamServerInfo{FullMethod: "/m"}, func(_ interface{}, _ grpc.ServerStream) error {
			return nil
		})
		require.Error(t, err)
		assert.Equal(t, codes.ResourceExhausted, status.Code(err))
	})
	t.Run("limiter error passes through", func(t *testing.T) {
		ic := icWithFake(&fakeLimiter{
			allowEndpoint: func(_, _ string, _ LimitType) (bool, *RateLimitResult, error) {
				return false, nil, errors.New("down")
			},
		})
		interceptor := ic.StreamServerInterceptor(GRPCInterceptorConfig{})
		ss := &fakeServerStream{ctx: context.Background()}
		called := false
		err := interceptor(nil, ss, &grpc.StreamServerInfo{FullMethod: "/m"}, func(_ interface{}, _ grpc.ServerStream) error {
			called = true
			return nil
		})
		require.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("user blocked", func(t *testing.T) {
		ic := icWithFake(&fakeLimiter{
			allowEndpoint: func(_, _ string, lt LimitType) (bool, *RateLimitResult, error) {
				if lt == LimitTypeUser {
					return false, denyResult("u", lt), nil
				}
				return true, allowResult("x", lt), nil
			},
		})
		interceptor := ic.StreamServerInterceptor(GRPCInterceptorConfig{
			UserExtractor: func(context.Context) string { return "u1" },
		})
		ss := &fakeServerStream{ctx: context.Background()}
		err := interceptor(nil, ss, &grpc.StreamServerInfo{FullMethod: "/m"}, func(_ interface{}, _ grpc.ServerStream) error {
			return nil
		})
		require.Error(t, err)
	})
	t.Run("user limiter error passes through", func(t *testing.T) {
		ic := icWithFake(&fakeLimiter{
			allowEndpoint: func(_, _ string, lt LimitType) (bool, *RateLimitResult, error) {
				if lt == LimitTypeUser {
					return false, nil, errors.New("down")
				}
				return true, allowResult("x", lt), nil
			},
		})
		interceptor := ic.StreamServerInterceptor(GRPCInterceptorConfig{
			UserExtractor: func(context.Context) string { return "u1" },
		})
		ss := &fakeServerStream{ctx: context.Background()}
		err := interceptor(nil, ss, &grpc.StreamServerInfo{FullMethod: "/m"}, func(_ interface{}, _ grpc.ServerStream) error {
			return nil
		})
		require.NoError(t, err)
	})
}

func TestGRPCPeerAndMetadataExtractors(t *testing.T) {
	assert.Equal(t, "unknown", extractPeerAddress(context.Background()))
	pctx := peer.NewContext(context.Background(), &peer.Peer{
		Addr: &net.TCPAddr{IP: net.ParseIP("9.9.9.9"), Port: 1234},
	})
	assert.Equal(t, "9.9.9.9", extractPeerAddress(pctx))

	extract := ExtractUserFromGRPCMetadata("x-user-id")
	assert.Equal(t, "", extract(context.Background()))
	mdctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-user-id", "md-user"))
	assert.Equal(t, "md-user", extract(mdctx))
	emptyMD := metadata.NewIncomingContext(context.Background(), metadata.Pairs("other", "v"))
	assert.Equal(t, "", extract(emptyMD))

	assert.Equal(t, "", ExtractUserFromAuthToken(context.Background()))
	assert.Equal(t, "", ExtractUserFromAuthToken(metadata.NewIncomingContext(context.Background(), metadata.Pairs("other", "v"))))
	assert.Equal(t, "", ExtractUserFromAuthToken(metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Basic abc"))))
	got := ExtractUserFromAuthToken(metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer tok123")))
	assert.Equal(t, "jwt:tok123", got)

	assert.True(t, sharedGRPCInterceptor.shouldSkip("/a/b", []string{"/a"}))
	assert.False(t, sharedGRPCInterceptor.shouldSkip("/b", []string{"/a"}))

	assert.Equal(t, context.Background(), sharedGRPCInterceptor.addRateLimitMetadata(context.Background(), nil))
	out := sharedGRPCInterceptor.addRateLimitMetadata(context.Background(), allowResult("x", LimitTypeIP))
	assert.NotNil(t, out)
}

func TestRateLimitInfo(t *testing.T) {
	assert.Nil(t, RateLimitInfo(context.Background()))
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-ratelimit-limit", "25",
		"x-ratelimit-remaining", "7",
		"x-ratelimit-reset", "1700000000",
	))
	info := RateLimitInfo(ctx)
	require.NotNil(t, info)
	assert.Equal(t, 25, info.Limit)
	assert.Equal(t, 7, info.Remaining)
	assert.Equal(t, int64(1700000000), info.ResetAt.Unix())
}

func TestChainInterceptors(t *testing.T) {
	order := []string{}
	mkUnary := func(name string) grpc.UnaryServerInterceptor {
		return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			order = append(order, name+":in")
			resp, err := handler(ctx, req)
			order = append(order, name+":out")
			return resp, err
		}
	}
	chained := ChainUnaryInterceptors(mkUnary("a"), mkUnary("b"))
	resp, err := chained(context.Background(), nil, &grpc.UnaryServerInfo{}, unaryHandlerOK)
	require.NoError(t, err)
	assert.Equal(t, "ok-response", resp)
	assert.Equal(t, []string{"a:in", "b:in", "b:out", "a:out"}, order)

	sorder := []string{}
	mkStream := func(name string) grpc.StreamServerInterceptor {
		return func(srv interface{}, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			sorder = append(sorder, name)
			return handler(srv, ss)
		}
	}
	schained := ChainStreamInterceptors(mkStream("x"), mkStream("y"))
	ss := &fakeServerStream{ctx: context.Background()}
	require.NoError(t, schained(nil, ss, &grpc.StreamServerInfo{}, func(_ interface{}, _ grpc.ServerStream) error {
		return nil
	}))
	assert.Equal(t, []string{"x", "y"}, sorder)

	w := &rateLimitedServerStream{ctx: context.Background()}
	assert.Equal(t, w.ctx, w.Context())
}

// ---- Monitor ----

func TestMonitorCollectMetrics(t *testing.T) {
	require.NoError(t, sharedMonitor.collectMetrics(context.Background()))

	bad := &Monitor{limiter: &fakeLimiter{metricsErr: errors.New("metrics down")}, logger: testLogger(), config: MonitorConfig{}}
	assert.Error(t, bad.collectMetrics(context.Background()))
}

func TestMonitorAlertThresholds(t *testing.T) {
	// Drain any leftovers first.
	select {
	case <-sharedMonitor.GetAlerts():
	default:
	}
	sharedMonitor.checkAlertThresholds(&Metrics{
		TotalRequests:     1000,
		AllowedRequests:   100,
		BlockedRequests:   50,
		BypassAttempts:    50,
		BannedIdentifiers: 10,
		CurrentLoad:       90.0,
		TopBlockedIPs:     []string{"1.2.3.4", "5.6.7.8"},
		ByLimitType:       map[LimitType]*LimitTypeMetrics{LimitTypeIP: {Blocked: 100}},
	})

	titles := map[string]bool{}
	timeout := time.After(2 * time.Second)
	for len(titles) < 5 {
		select {
		case a := <-sharedMonitor.GetAlerts():
			titles[a.Title] = true
		case <-timeout:
			t.Fatalf("timed out waiting for alerts, got %v", titles)
		}
	}
	for _, want := range []string{"High Rate Limit Blocks", "Potential DDoS Attack", "High Number of Banned Identifiers", "High System Load", "High Block Rate for IP"} {
		assert.True(t, titles[want], "missing alert %q", want)
	}
}

func TestMonitorNoAlertsBelowThreshold(t *testing.T) {
	// Drain leftovers.
	select {
	case <-sharedMonitor.GetAlerts():
	default:
	}
	sharedMonitor.checkAlertThresholds(&Metrics{
		BlockedRequests:   1,
		BypassAttempts:    0,
		BannedIdentifiers: 0,
		CurrentLoad:       1.0,
	})
	select {
	case a := <-sharedMonitor.GetAlerts():
		t.Fatalf("unexpected alert: %v", a.Title)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestMonitorSendAlertFullChannel(t *testing.T) {
	m := &Monitor{logger: testLogger(), alertChan: make(chan Alert, 1)}
	m.alertChan <- Alert{Title: "fill"}
	// Channel full -> must not block, alert dropped.
	done := make(chan struct{})
	go func() { m.sendAlert(Alert{Title: "dropped"}); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("sendAlert blocked on full channel")
	}
	assert.Equal(t, "fill", (<-m.alertChan).Title)
	// Normal path increments without panic.
	m2 := &Monitor{logger: testLogger(), alertChan: make(chan Alert, 10)}
	m2.metrics = sharedMonitor.metrics
	m2.sendAlert(Alert{Severity: "info", Title: "t"})
	assert.Equal(t, "t", (<-m2.alertChan).Title)
}

func TestMonitorHandleAlertAndSeverity(t *testing.T) {
	assert.Equal(t, "error", sharedMonitor.getSeverityLevel("critical").String())
	assert.Equal(t, "warn", sharedMonitor.getSeverityLevel("warning").String())
	assert.Equal(t, "info", sharedMonitor.getSeverityLevel("info").String())
	assert.Equal(t, "debug", sharedMonitor.getSeverityLevel("whatever").String())

	sharedMonitor.handleAlert(Alert{Severity: "critical", Title: "c", Message: "m"})
	withHook := &Monitor{logger: testLogger(), config: MonitorConfig{AlertWebhookURL: "http://example.com/hook"}}
	withHook.handleAlert(Alert{Severity: "info", Title: "i", Message: "m"})
}

func TestMonitorStartStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := sharedMonitor.Start(ctx)
	assert.Error(t, err)
}

func TestMonitorStartCollects(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := sharedMonitor.Start(ctx)
	assert.Error(t, err)
}

func TestDefaultConfigs(t *testing.T) {
	cfg := DefaultConfig()
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "redis://localhost:6379/0", cfg.RedisURL)
	assert.Greater(t, cfg.IPLimits.RequestsPerSecond, 0)

	th := DefaultAlertThresholds()
	assert.Equal(t, uint64(1000), th.BlockedRequestsPerMinute)
	assert.Equal(t, 80.0, th.LoadPercentage)
	assert.Equal(t, 1, min(1, 2))
	assert.Equal(t, 2, min(3, 2))
}

// ---- ServerIntegration handlers (constructed directly with fake limiter) ----

func testIntegration() *ServerIntegration {
	return &ServerIntegration{
		limiter:         newFakeLimiter(),
		httpMiddleware:  sharedHTTPMiddleware,
		grpcInterceptor: sharedGRPCInterceptor,
		monitor:         sharedMonitor,
		logger:          testLogger(),
	}
}

func TestIntegrationMetricsEndpoint(t *testing.T) {
	s := testIntegration()
	r := mux.NewRouter()
	s.RegisterMetricsEndpoint(r, "/metrics")
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "total_requests")

	s.limiter.(*fakeLimiter).metricsErr = errors.New("down")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	assert.Equal(t, http.StatusInternalServerError, rec2.Code)
}

func TestIntegrationBanUnban(t *testing.T) {
	s := testIntegration()
	r := mux.NewRouter()
	s.RegisterAdminEndpoints(r, "/admin")

	doPost := func(path, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
		return rec
	}

	t.Run("ban ok", func(t *testing.T) {
		rec := doPost("/admin/ban", `{"identifier":"bad-ip","duration":60,"reason":"spam"}`)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "banned")
		banned, _ := s.limiter.IsBanned(context.Background(), "bad-ip")
		assert.True(t, banned)
	})
	t.Run("ban invalid json", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, doPost("/admin/ban", `{oops`).Code)
	})
	t.Run("ban missing identifier", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, doPost("/admin/ban", `{"duration":60}`).Code)
	})
	t.Run("ban limiter error", func(t *testing.T) {
		s.limiter.(*fakeLimiter).banErr = errors.New("ban failed")
		assert.Equal(t, http.StatusInternalServerError, doPost("/admin/ban", `{"identifier":"x"}`).Code)
		s.limiter.(*fakeLimiter).banErr = nil
	})
	t.Run("unban ok", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, doPost("/admin/unban", `{"identifier":"bad-ip"}`).Code)
	})
	t.Run("unban invalid json", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, doPost("/admin/unban", `{oops`).Code)
	})
	t.Run("unban missing identifier", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, doPost("/admin/unban", `{}`).Code)
	})
	t.Run("get config", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/config", nil))
		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("update config ok", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/config", strings.NewReader(`{"enabled":true}`)))
		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("update config invalid", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/config", strings.NewReader(`{oops`)))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
	t.Run("update config limiter error", func(t *testing.T) {
		s.limiter.(*fakeLimiter).updateErr = errors.New("nope")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/config", strings.NewReader(`{}`)))
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestIntegrationAccessors(t *testing.T) {
	s := testIntegration()
	assert.NotNil(t, s.GetLimiter())

	r := mux.NewRouter()
	out := s.WrapHTTPRouter(r, HTTPMiddlewareConfig{})
	assert.NotNil(t, out)

	assert.NotNil(t, s.GetGRPCUnaryInterceptor(GRPCInterceptorConfig{}))
	assert.NotNil(t, s.GetGRPCStreamInterceptor(GRPCInterceptorConfig{}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.Error(t, s.StartMonitor(ctx))
	assert.NoError(t, s.Close())

	assert.Equal(t, time.Duration(0), parseDuration(0))
	assert.Equal(t, 90*time.Second, parseDuration(90))
	assert.Error(t, readJSON(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{bad`)), &struct{}{}))
	rec := httptest.NewRecorder()
	assert.NoError(t, writeJSON(rec, map[string]string{"a": "b"}))
	assert.Contains(t, rec.Body.String(), `"a"`)
}

func TestQuickSetupError(t *testing.T) {
	_, err := QuickSetupWithConfig(context.Background(), RateLimitConfig{RedisURL: "://invalid-url"}, testLogger())
	assert.Error(t, err)
}

// ---- RedisRateLimiter offline paths (struct built directly, no client) ----

func offlineRedisLimiter(cfg RateLimitConfig) *RedisRateLimiter {
	return &RedisRateLimiter{
		config: cfg,
		logger: testLogger(),
		metrics: &metricsTracker{
			byLimitType:       make(map[LimitType]*LimitTypeMetrics),
			blockedIPCounts:   make(map[string]uint64),
			blockedUserCounts: make(map[string]uint64),
		},
	}
}

func TestRedisLimiterAllowDisabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = false
	r := offlineRedisLimiter(cfg)
	allowed, result, err := r.Allow(context.Background(), "any-key", LimitTypeIP)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, "any-key", result.Identifier)
}

func TestRedisLimiterAllowWhitelisted(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WhitelistedIPs = []string{"10.1.2.3", "192.168.0.0/16"}
	cfg.WhitelistedUsers = []string{"admin"}
	r := offlineRedisLimiter(cfg)

	allowed, _, err := r.Allow(context.Background(), "10.1.2.3", LimitTypeIP)
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, _, err = r.Allow(context.Background(), "192.168.5.5", LimitTypeIP)
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, _, err = r.Allow(context.Background(), "admin", LimitTypeUser)
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestRedisLimiterIsWhitelisted(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WhitelistedIPs = []string{"10.1.2.3", "192.168.0.0/16"}
	cfg.WhitelistedUsers = []string{"admin"}
	r := offlineRedisLimiter(cfg)

	assert.True(t, r.IsWhitelisted("10.1.2.3", LimitTypeIP))
	assert.True(t, r.IsWhitelisted("192.168.1.1", LimitTypeIP))
	assert.False(t, r.IsWhitelisted("11.0.0.1", LimitTypeIP))
	assert.True(t, r.IsWhitelisted("admin", LimitTypeUser))
	assert.False(t, r.IsWhitelisted("bob", LimitTypeUser))
	assert.False(t, r.IsWhitelisted("x", LimitTypeGlobal))
	assert.False(t, r.IsWhitelisted("x", LimitTypeEndpoint))
}

func TestRedisLimiterUpdateConfig(t *testing.T) {
	r := offlineRedisLimiter(DefaultConfig())
	cfg := DefaultConfig()
	cfg.RedisPrefix = "test:prefix"
	require.NoError(t, r.UpdateConfig(cfg))
	assert.Equal(t, "test:prefix", r.config.RedisPrefix)
	assert.Equal(t, cfg.IPLimits, r.getLimitsForType(LimitTypeIP))
	assert.Equal(t, cfg.UserLimits, r.getLimitsForType(LimitTypeUser))
	assert.Equal(t, cfg.GlobalLimits, r.getLimitsForType(LimitTypeGlobal))
	assert.Equal(t, cfg.IPLimits, r.getLimitsForType(LimitTypeEndpoint))
	assert.Equal(t, cfg.IPLimits, r.getLimitsForType("bogus"))
}

func TestRedisLimiterEndpointLimits(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EndpointLimits = map[string]LimitRules{
		"/api/exact": {RequestsPerSecond: 5},
		"/api/v1/*":  {RequestsPerSecond: 7},
	}
	r := offlineRedisLimiter(cfg)

	exact := r.getEndpointLimits("/api/exact")
	require.NotNil(t, exact)
	assert.Equal(t, 5, exact.RequestsPerSecond)

	wild := r.getEndpointLimits("/api/v1/users")
	require.NotNil(t, wild)
	assert.Equal(t, 7, wild.RequestsPerSecond)

	assert.Nil(t, r.getEndpointLimits("/other"))
	assert.True(t, r.matchesPattern("/a/b", "/a/*"))
	assert.False(t, r.matchesPattern("/b", "/a/*"))
	assert.True(t, r.matchesPattern("/same", "/same"))
	assert.False(t, r.matchesPattern("/a", "/b"))
}

func TestRedisLimiterIPPatterns(t *testing.T) {
	r := offlineRedisLimiter(DefaultConfig())
	assert.True(t, r.matchesIPPattern("1.2.3.4", "1.2.3.4"))
	assert.True(t, r.matchesIPPattern("10.0.5.5", "10.0.0.0/8"))
	assert.False(t, r.matchesIPPattern("11.0.0.1", "10.0.0.0/8"))
	assert.False(t, r.matchesIPPattern("1.2.3.4", "not-a-cidr/99"))
	assert.False(t, r.matchesIPPattern("not-an-ip", "10.0.0.0/8"))
	assert.False(t, r.matchesIPPattern("1.2.3.4", "5.6.7.8"))
}

func TestRedisLimiterDegradationMultiplier(t *testing.T) {
	r := offlineRedisLimiter(DefaultConfig())
	// DefaultConfig enables degradation: 80->0.7, 90->0.5, 95->0.3.
	assert.Equal(t, 0.3, r.getDegradationMultiplier(95, "/api"))
	assert.Equal(t, 1.0, r.getDegradationMultiplier(95, "/veid/verify"))
	assert.Equal(t, 1.0, r.getDegradationMultiplier(85, "/veid/other"))
	assert.Equal(t, 1.0, r.getDegradationMultiplier(10, "/api"))

	cfg := DefaultConfig()
	cfg.GracefulDegradation = DegradationConfig{
		Enabled: true,
		LoadThresholds: []LoadThreshold{
			{LoadPercentage: 50, RateLimitMultiplier: 0.5},
			{LoadPercentage: 90, RateLimitMultiplier: 0.1, Priority: []string{"/api/priority/*"}},
		},
	}
	r2 := offlineRedisLimiter(cfg)
	assert.Equal(t, 1.0, r2.getDegradationMultiplier(10, "/api"))
	assert.Equal(t, 0.5, r2.getDegradationMultiplier(60, "/api"))
	assert.Equal(t, 0.1, r2.getDegradationMultiplier(95, "/api"))
	assert.Equal(t, 1.0, r2.getDegradationMultiplier(95, "/api/priority/job"))

	scaled := r2.applyMultiplier(LimitRules{RequestsPerSecond: 100, RequestsPerMinute: 1000, RequestsPerHour: 10000, RequestsPerDay: 100000, BurstSize: 50}, 0.5)
	assert.Equal(t, 50, scaled.RequestsPerSecond)
	assert.Equal(t, 500, scaled.RequestsPerMinute)
	assert.Equal(t, 25, scaled.BurstSize)
	same := r2.applyMultiplier(scaled, 1.0)
	assert.Equal(t, scaled, same)
}

func TestRedisLimiterRecordMetric(t *testing.T) {
	r := offlineRedisLimiter(DefaultConfig())
	r.recordMetric(true, LimitTypeIP, "1.1.1.1")
	r.recordMetric(false, LimitTypeIP, "2.2.2.2")
	r.recordMetric(false, LimitTypeIP, "2.2.2.2")
	r.recordMetric(false, LimitTypeUser, "u1")
	r.recordMetric(false, LimitTypeGlobal, "g")

	assert.Equal(t, uint64(5), r.metrics.totalRequests)
	assert.Equal(t, uint64(1), r.metrics.allowedRequests)
	assert.Equal(t, uint64(4), r.metrics.blockedRequests)
	assert.Equal(t, uint64(2), r.metrics.blockedIPCounts["2.2.2.2"])
	assert.Equal(t, uint64(1), r.metrics.blockedUserCounts["u1"])
	assert.Equal(t, uint64(3), r.metrics.byLimitType[LimitTypeIP].Requests)
	assert.Equal(t, uint64(2), r.metrics.byLimitType[LimitTypeIP].Blocked)
	assert.Equal(t, uint64(1), r.metrics.byLimitType[LimitTypeIP].Allowed)

	top := r.getTopBlocked(map[string]uint64{"a": 3, "b": 10, "c": 1}, 2)
	assert.Equal(t, []string{"b", "a"}, top)
	assert.Empty(t, r.getTopBlocked(map[string]uint64{}, 5))
}

func TestRedisLimiterCloseOffline(t *testing.T) {
	r := offlineRedisLimiter(DefaultConfig())
	r.client = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	assert.NoError(t, r.Close())
}

// ---- Minimal RESP stub: exercises the Redis-backed paths without Redis ----

type respStub struct {
	ln   net.Listener
	addr string
}

func startRESPStub(t *testing.T) *respStub {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	s := &respStub{ln: ln, addr: ln.Addr().String()}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go s.serve(conn)
		}
	}()
	t.Cleanup(func() { _ = ln.Close() })
	return s
}

func (s *respStub) serve(conn net.Conn) {
	defer conn.Close()
	rd := bufio.NewReader(conn)
	for {
		args, err := readRESPArray(rd)
		if err != nil {
			return
		}
		if len(args) == 0 {
			return
		}
		s.handle(conn, args)
	}
}

func readRESPArray(rd *bufio.Reader) ([]string, error) {
	line, err := rd.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	if !strings.HasPrefix(line, "*") {
		return nil, errors.New("expected array")
	}
	n, err := strconv.Atoi(strings.TrimPrefix(line, "*"))
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, n)
	for i := 0; i < n; i++ {
		hdr, err := rd.ReadString('\n')
		if err != nil {
			return nil, err
		}
		hdr = strings.TrimSuffix(strings.TrimSuffix(hdr, "\n"), "\r")
		if !strings.HasPrefix(hdr, "$") {
			return nil, errors.New("expected bulk string")
		}
		ln, err := strconv.Atoi(strings.TrimPrefix(hdr, "$"))
		if err != nil {
			return nil, err
		}
		if ln < 0 {
			args = append(args, "")
			continue
		}
		buf := make([]byte, ln+2)
		if _, err := io.ReadFull(rd, buf); err != nil {
			return nil, err
		}
		args = append(args, string(buf[:ln]))
	}
	return args, nil
}

func writeStr(conn net.Conn, s string) {
	_, _ = fmt.Fprintf(conn, "%s\r\n", s)
}

func (s *respStub) handle(conn net.Conn, args []string) {
	switch strings.ToUpper(args[0]) {
	case "HELLO":
		// Minimal HELLO 2 response (map) so go-redis accepts the connection.
		_, _ = conn.Write([]byte("%6\r\n$6\r\nserver\r\n$5\r\nredis\r\n$7\r\nversion\r\n$5\r\n7.0.0\r\n$5\r\nproto\r\n:2\r\n$4\r\nmode\r\n$10\r\nstandalone\r\n$4\r\nrole\r\n$6\r\nmaster\r\n$2\r\nid\r\n:1\r\n"))
	case "PING":
		writeStr(conn, "+PONG")
	case "QUIT", "CLIENT":
		writeStr(conn, "+OK")
	case "SET":
		writeStr(conn, "+OK")
	case "EXPIRE":
		writeStr(conn, ":1")
	case "INCR":
		writeStr(conn, ":1")
	case "EXISTS":
		if len(args) > 1 && strings.Contains(args[1], "banned") {
			writeStr(conn, ":1")
		} else {
			writeStr(conn, ":0")
		}
	case "GET":
		writeStr(conn, "$-1")
	case "SCAN":
		// Empty key list, cursor 0 (done).
		_, _ = conn.Write([]byte("*2\r\n$1\r\n0\r\n*0\r\n"))
	case "EVAL":
		key := ""
		if len(args) > 3 {
			key = args[3]
		}
		now := time.Now().Unix()
		if strings.Contains(key, "exhaust") {
			// Deny with a reset time in the past (covers retryAfter clamp).
			_, _ = fmt.Fprintf(conn, "*3\r\n:0\r\n:0\r\n:%d\r\n", now-30)
			return
		}
		_, _ = fmt.Fprintf(conn, "*3\r\n:1\r\n:9\r\n:%d\r\n", now+60)
	default:
		writeStr(conn, "+OK")
	}
}

func stubBackedLimiter(t *testing.T, mutate func(*RateLimitConfig)) *RedisRateLimiter {
	t.Helper()
	stub := startRESPStub(t)
	cfg := DefaultConfig()
	cfg.RedisURL = "redis://" + stub.addr + "/0"
	if mutate != nil {
		mutate(&cfg)
	}
	r, err := NewRedisRateLimiter(context.Background(), cfg, testLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close() })
	return r
}

func TestRedisLimiterStubAllowEndpoint(t *testing.T) {
	r := stubBackedLimiter(t, nil)
	allowed, result, err := r.AllowEndpoint(context.Background(), "/market/ticker", "trader-1", LimitTypeIP)
	require.NoError(t, err)
	assert.True(t, allowed)
	require.NotNil(t, result)

	// Endpoint without specific limits passes through general result.
	allowed, _, err = r.AllowEndpoint(context.Background(), "/no/limits", "trader-2", LimitTypeIP)
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestRedisLimiterStubDenied(t *testing.T) {
	r := stubBackedLimiter(t, nil)
	allowed, result, err := r.Allow(context.Background(), "exhaust-me", LimitTypeIP)
	require.NoError(t, err)
	assert.False(t, allowed)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.RetryAfter) // past reset clamped
	assert.Equal(t, 10, result.Limit)     // per-second window first
}

func TestRedisLimiterStubBanned(t *testing.T) {
	r := stubBackedLimiter(t, nil)
	allowed, result, err := r.Allow(context.Background(), "banned-user", LimitTypeIP)
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 3600, result.RetryAfter)

	banned, err := r.IsBanned(context.Background(), "banned-user")
	require.NoError(t, err)
	assert.True(t, banned)
	banned, err = r.IsBanned(context.Background(), "clean-user")
	require.NoError(t, err)
	assert.False(t, banned)
}

func TestRedisLimiterStubBanAndBypass(t *testing.T) {
	r := stubBackedLimiter(t, func(c *RateLimitConfig) {
		c.BypassDetection.MaxFailedAttemptsPerMinute = 0
		c.BypassDetection.AlertThreshold = 0
	})
	require.NoError(t, r.Ban(context.Background(), "spammer", time.Minute, "test"))
	require.NoError(t, r.Ban(context.Background(), "forever", 0, "permanent"))
	// INCR returns 1 > thresholds of 0: warn + auto-ban + alert branches.
	require.NoError(t, r.RecordBypassAttempt(context.Background(), "spammer", "burst"))
}

func TestRedisLimiterStubMetrics(t *testing.T) {
	r := stubBackedLimiter(t, nil)
	m, err := r.GetMetrics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, m.BannedIdentifiers)
	assert.Equal(t, float64(0), m.CurrentLoad)

	load, err := r.GetCurrentLoad(context.Background())
	require.NoError(t, err)
	assert.Equal(t, float64(0), load)
}

func TestRedisLimiterNewInvalidURL(t *testing.T) {
	_, err := NewRedisRateLimiter(context.Background(), RateLimitConfig{RedisURL: "://bad"}, testLogger())
	assert.Error(t, err)
	_, err = NewRedisRateLimiter(context.Background(), RateLimitConfig{RedisURL: "redis://127.0.0.1:1/0"}, testLogger())
	assert.Error(t, err)
}
