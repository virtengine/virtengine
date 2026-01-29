# VirtEngine Rate Limiting Package

Comprehensive rate limiting and DDoS protection for VirtEngine APIs.

## Features

- **Multi-layer Rate Limiting**
  - IP-based rate limiting
  - Per-user rate limiting
  - Endpoint-specific rate limiting
  - Global system-wide rate limiting

- **Distributed & Scalable**
  - Redis-backed storage
  - Supports multiple nodes
  - Token bucket algorithm
  - Configurable time windows (second, minute, hour, day)

- **DDoS Protection**
  - Automatic bypass detection
  - Auto-ban malicious actors
  - Graceful degradation under load
  - Whitelist/blacklist support

- **Easy Integration**
  - HTTP middleware for REST APIs
  - gRPC interceptors
  - One-line setup
  - Configurable extractors

- **Monitoring & Alerting**
  - Prometheus metrics
  - Real-time alerts
  - Comprehensive logging
  - Admin API endpoints

## Quick Start

```go
import "github.com/virtengine/virtengine/pkg/ratelimit"

// Initialize
ctx := context.Background()
integration, err := ratelimit.QuickSetup(
    ctx,
    "redis://localhost:6379/0",
    logger,
)
if err != nil {
    return err
}
defer integration.Close()

// HTTP Server
router := mux.NewRouter()
config := ratelimit.HTTPMiddlewareConfig{
    UserExtractor: ratelimit.ExtractUserFromJWT,
    SkipPaths:     []string{"/health"},
}
router = integration.WrapHTTPRouter(router, config)

// gRPC Server
grpcConfig := ratelimit.GRPCInterceptorConfig{
    UserExtractor: ratelimit.ExtractUserFromAuthToken,
}
server := grpc.NewServer(
    grpc.UnaryInterceptor(integration.GetGRPCUnaryInterceptor(grpcConfig)),
)

// Start Monitoring
go integration.StartMonitor(ctx)
```

## Default Rate Limits

### Anonymous Users (IP-based)
- 10 requests/second
- 300 requests/minute
- 10,000 requests/hour
- Burst: 20 requests

### Authenticated Users
- 50 requests/second
- 1,000 requests/minute
- 50,000 requests/hour
- Burst: 100 requests

### Endpoint-Specific Limits

| Endpoint | Requests/Second | Requests/Minute |
|----------|----------------|-----------------|
| `/veid/*` | 5 | 100 |
| `/market/*` | 20 | 600 |
| Default | 10 | 300 |

## Architecture

```
Client → CDN/WAF → HTTP/gRPC Middleware → Redis Limiter → Backend
```

## Configuration

### Environment Variables

```bash
REDIS_URL=redis://localhost:6379/0
RATELIMIT_ENABLED=true
RATELIMIT_IP_PER_SECOND=10
RATELIMIT_USER_PER_SECOND=50
RATELIMIT_BYPASS_ENABLED=true
RATELIMIT_DEGRADATION_ENABLED=true
```

### Programmatic Configuration

```go
config := ratelimit.RateLimitConfig{
    RedisURL: "redis://localhost:6379/0",
    Enabled:  true,
    IPLimits: ratelimit.LimitRules{
        RequestsPerSecond: 10,
        RequestsPerMinute: 300,
        BurstSize:         20,
    },
    BypassDetection: ratelimit.BypassDetectionConfig{
        Enabled:                    true,
        MaxFailedAttemptsPerMinute: 100,
        BanDuration:                time.Hour,
    },
}

integration, err := ratelimit.QuickSetupWithConfig(ctx, config, logger)
```

## API Response Headers

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1672531200
Retry-After: 60  (on 429 only)
```

## Prometheus Metrics

```
virtengine_ratelimit_total_requests
virtengine_ratelimit_allowed_requests
virtengine_ratelimit_blocked_requests
virtengine_ratelimit_bypass_attempts
virtengine_ratelimit_banned_identifiers
virtengine_ratelimit_current_load
```

## Testing

```bash
# Run tests
go test ./pkg/ratelimit/...

# With coverage
go test -cover ./pkg/ratelimit/...

# Benchmarks
go test -bench=. ./pkg/ratelimit/...
```

## Documentation

- [Implementation Guide](../../docs/RATELIMIT_IMPLEMENTATION.md) - Complete setup guide
- [Client Guide](../../docs/RATELIMIT_CLIENT_GUIDE.md) - For API consumers
- [CDN/WAF Setup](../../docs/RATELIMIT_CDN_WAF.md) - Network-level protection

## Examples

### Custom User Extractor

```go
config := ratelimit.HTTPMiddlewareConfig{
    UserExtractor: func(r *http.Request) string {
        // Extract from custom header
        if userID := r.Header.Get("X-User-ID"); userID != "" {
            return userID
        }
        // Extract from session
        if session := getSession(r); session != nil {
            return session.UserID
        }
        return ""
    },
}
```

### Custom Rate Limit Handler

```go
config := ratelimit.HTTPMiddlewareConfig{
    OnRateLimited: func(w http.ResponseWriter, r *http.Request, result *ratelimit.RateLimitResult) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusTooManyRequests)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error":   "rate_limit_exceeded",
            "message": "Please slow down",
            "retry_after": result.RetryAfter,
        })
    },
}
```

### Admin API

```go
// Register admin endpoints
integration.RegisterAdminEndpoints(router, "/admin/ratelimit")

// GET  /admin/ratelimit/metrics  - Get metrics
// POST /admin/ratelimit/ban      - Ban identifier
// POST /admin/ratelimit/unban    - Unban identifier
// GET  /admin/ratelimit/config   - Get config
// PUT  /admin/ratelimit/config   - Update config
```

### Manual Rate Limit Check

```go
limiter := integration.GetLimiter()

// Check IP limit
allowed, result, err := limiter.Allow(ctx, "192.168.1.1", ratelimit.LimitTypeIP)
if !allowed {
    log.Printf("Rate limited: retry after %d seconds", result.RetryAfter)
}

// Check user limit
allowed, result, err = limiter.Allow(ctx, "user123", ratelimit.LimitTypeUser)

// Check endpoint limit
allowed, result, err = limiter.AllowEndpoint(ctx, "/api/orders", "user123", ratelimit.LimitTypeUser)

// Ban an identifier
err = limiter.Ban(ctx, "malicious-ip", time.Hour, "DDoS attack")

// Check if banned
banned, err := limiter.IsBanned(ctx, "malicious-ip")
```

## Performance

Benchmarks on a standard development machine:

```
BenchmarkAllow-8           100000    10523 ns/op
BenchmarkAllowEndpoint-8    50000    21045 ns/op
BenchmarkIsBanned-8        200000     5234 ns/op
```

Redis operations are the primary bottleneck. Consider:
- Using Redis cluster for horizontal scaling
- Co-locating Redis with application servers
- Using Redis Cluster with read replicas

## Graceful Degradation

Under high load, the system automatically reduces rate limits:

```
Load < 80%:  Normal limits
Load 80-90%: 70% of normal limits (priority endpoints unaffected)
Load 90-95%: 50% of normal limits
Load > 95%:  30% of normal limits (read-only mode optional)
```

## Security Considerations

1. **IP Spoofing**: Always trust X-Forwarded-For from CDN/WAF, not directly from clients
2. **Bypass Detection**: Enabled by default, auto-bans after 100 failed attempts/minute
3. **Whitelisting**: Use carefully, regularly audit whitelist
4. **Redis Security**: Use authentication, TLS, and network isolation
5. **Metrics Exposure**: Protect `/admin/ratelimit/*` endpoints with authentication

## Troubleshooting

### High Redis Memory Usage

```bash
# Check memory usage
redis-cli INFO memory

# Set eviction policy
redis-cli CONFIG SET maxmemory-policy allkeys-lru
redis-cli CONFIG SET maxmemory 1gb
```

### Rate Limits Not Working

```bash
# Check Redis connection
redis-cli -u redis://localhost:6379 PING

# Check configuration
curl http://localhost:1317/admin/ratelimit/config

# Check logs
grep "rate-limit" /var/log/virtengine.log
```

### False Positives

Add to whitelist:
```go
config.WhitelistedIPs = []string{"10.0.0.0/8", "trusted-ip"}
config.WhitelistedUsers = []string{"admin", "monitoring"}
```

## Contributing

Contributions welcome! Please:

1. Add tests for new features
2. Update documentation
3. Follow Go best practices
4. Run `go fmt` and `go vet`

## License

Copyright © 2024 VirtEngine. All rights reserved.

See [LICENSE](../../LICENSE) for details.

## Support

- **Issues**: https://github.com/virtengine/virtengine/issues
- **Email**: security@virtengine.com
- **Discord**: https://discord.gg/virtengine
