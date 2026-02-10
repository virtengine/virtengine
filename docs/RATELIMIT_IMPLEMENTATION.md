# Rate Limiting Implementation Guide

This guide provides step-by-step instructions for implementing comprehensive rate limiting in VirtEngine.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Quick Start](#quick-start)
4. [Detailed Integration](#detailed-integration)
5. [Configuration](#configuration)
6. [Monitoring](#monitoring)
7. [Testing](#testing)
8. [Deployment](#deployment)

---

## Overview

The rate limiting system provides:

- **Multi-layer protection**: IP, user, endpoint, and global rate limiting
- **Redis-backed storage**: Distributed rate limiting across multiple nodes
- **Graceful degradation**: Automatic rate limit adjustment under load
- **Bypass detection**: Automatic detection and banning of abusive clients
- **Comprehensive monitoring**: Prometheus metrics and alerting
- **Easy integration**: Simple HTTP and gRPC middleware

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Client Requests                       │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│            CDN/WAF (Cloudflare/AWS/Azure)                │
│              Network-level Protection                    │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│                HTTP/gRPC Middleware                      │
│          IP + User + Endpoint Rate Limiting              │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│              Redis Rate Limiter Core                     │
│        Token Bucket + Ban Management + Metrics           │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│                 VirtEngine Application                   │
│          Transaction Rate Limiting (Existing)            │
└─────────────────────────────────────────────────────────┘
```

---

## Prerequisites

### Required

- **Redis 6.0+**: For distributed rate limiting storage
- **Go 1.25+**: VirtEngine build requirement

### Optional

- **Prometheus**: For metrics collection
- **Grafana**: For metrics visualization
- **CDN/WAF**: Cloudflare, AWS WAF, or Azure Front Door

---

## Quick Start

### 1. Install Redis

```bash
# Ubuntu/Debian
sudo apt-get install redis-server

# macOS
brew install redis

# Docker
docker run -d -p 6379:6379 redis:7-alpine

# Start Redis
redis-server
```

### 2. Add to Your Server Code

Update `sdk/go/cli/server.go`:

```go
import (
    "github.com/virtengine/virtengine/pkg/ratelimit"
)

func startAPIServer(...) error {
    // Initialize rate limiting
    ctx := context.Background()
    rateLimitIntegration, err := ratelimit.QuickSetup(
        ctx,
        "redis://localhost:6379/0",
        sctx.Logger,
    )
    if err != nil {
        return fmt.Errorf("failed to setup rate limiting: %w", err)
    }
    defer rateLimitIntegration.Close()

    // Start monitoring
    go rateLimitIntegration.StartMonitor(ctx)

    // Wrap HTTP router
    httpConfig := ratelimit.HTTPMiddlewareConfig{
        UserExtractor: ratelimit.ExtractUserFromJWT,
        SkipPaths:     []string{"/health", "/metrics"},
    }
    apiSrv.Router = rateLimitIntegration.WrapHTTPRouter(apiSrv.Router, httpConfig)

    // Add gRPC interceptors
    grpcConfig := ratelimit.GRPCInterceptorConfig{
        UserExtractor: ratelimit.ExtractUserFromAuthToken,
        SkipMethods:   []string{"/cosmos.base.tendermint.v1beta1.Service/"},
    }

    grpcSrv = grpc.NewServer(
        grpc.UnaryInterceptor(rateLimitIntegration.GetGRPCUnaryInterceptor(grpcConfig)),
        grpc.StreamInterceptor(rateLimitIntegration.GetGRPCStreamInterceptor(grpcConfig)),
    )

    // Register admin endpoints
    rateLimitIntegration.RegisterAdminEndpoints(apiSrv.Router, "/admin/ratelimit")

    return nil
}
```

### 3. Configure Environment

Create `.env` file:

```bash
# Redis Configuration
REDIS_URL=redis://localhost:6379/0
REDIS_PREFIX=virtengine:ratelimit

# Rate Limiting
RATELIMIT_ENABLED=true
RATELIMIT_IP_REQUESTS_PER_SECOND=10
RATELIMIT_IP_REQUESTS_PER_MINUTE=300
RATELIMIT_USER_REQUESTS_PER_SECOND=50
RATELIMIT_USER_REQUESTS_PER_MINUTE=1000

# Bypass Detection
RATELIMIT_BYPASS_DETECTION_ENABLED=true
RATELIMIT_BYPASS_MAX_ATTEMPTS_PER_MINUTE=100
RATELIMIT_BAN_DURATION=3600

# Graceful Degradation
RATELIMIT_DEGRADATION_ENABLED=true
```

### 4. Run and Test

```bash
# Start the server
go run ./cmd/virtengine start

# Test rate limiting
for i in {1..15}; do
  curl http://localhost:1317/market/orders
  sleep 0.1
done

# You should see 429 responses after hitting the limit
```

---

## Detailed Integration

### HTTP Server Integration

#### Full Configuration

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/gorilla/mux"
    "github.com/rs/zerolog"
    "github.com/virtengine/virtengine/pkg/ratelimit"
)

func setupHTTPServer() error {
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

    // Create rate limit configuration
    config := ratelimit.RateLimitConfig{
        RedisURL:    "redis://localhost:6379/0",
        RedisPrefix: "virtengine:ratelimit",
        Enabled:     true,
        IPLimits: ratelimit.LimitRules{
            RequestsPerSecond: 10,
            RequestsPerMinute: 300,
            RequestsPerHour:   10000,
            BurstSize:         20,
        },
        UserLimits: ratelimit.LimitRules{
            RequestsPerSecond: 50,
            RequestsPerMinute: 1000,
            RequestsPerHour:   50000,
            BurstSize:         100,
        },
        EndpointLimits: map[string]ratelimit.LimitRules{
            "/api/veid/*": {
                RequestsPerSecond: 5,
                RequestsPerMinute: 100,
                RequestsPerHour:   1000,
                BurstSize:         10,
            },
        },
        WhitelistedIPs:   []string{"127.0.0.1", "10.0.0.0/8"},
        WhitelistedUsers: []string{"admin", "service-account"},
        BypassDetection: ratelimit.BypassDetectionConfig{
            Enabled:                    true,
            MaxFailedAttemptsPerMinute: 100,
            BanDuration:                time.Hour,
            AlertThreshold:             50,
        },
        GracefulDegradation: ratelimit.DegradationConfig{
            Enabled: true,
            LoadThresholds: []ratelimit.LoadThreshold{
                {
                    LoadPercentage:      80,
                    RateLimitMultiplier: 0.7,
                    Priority:            []string{"/api/veid/*"},
                },
                {
                    LoadPercentage:      90,
                    RateLimitMultiplier: 0.5,
                },
            },
        },
    }

    // Create integration
    ctx := context.Background()
    integration, err := ratelimit.QuickSetupWithConfig(ctx, config, logger)
    if err != nil {
        return err
    }
    defer integration.Close()

    // Start monitoring
    go integration.StartMonitor(ctx)

    // Create router
    router := mux.NewRouter()

    // Configure middleware
    middlewareConfig := ratelimit.HTTPMiddlewareConfig{
        IdentifierExtractor: ratelimit.extractIPAddress, // Default
        UserExtractor:       extractUserFromRequest,     // Custom
        SkipPaths:           []string{"/health", "/metrics", "/docs"},
        OnRateLimited:       customRateLimitHandler,     // Optional
    }

    // Wrap router
    router = integration.WrapHTTPRouter(router, middlewareConfig)

    // Register endpoints
    router.HandleFunc("/api/market/orders", handleOrders).Methods("GET")
    router.HandleFunc("/api/veid/verify", handleVerify).Methods("POST")

    // Register admin endpoints
    integration.RegisterAdminEndpoints(router, "/admin/ratelimit")

    // Start server
    log.Println("Starting server on :8080")
    return http.ListenAndServe(":8080", router)
}

// Custom user extractor
func extractUserFromRequest(r *http.Request) string {
    // Extract from JWT, session, or other auth mechanism
    if user := r.Header.Get("X-User-ID"); user != "" {
        return user
    }
    return ""
}

// Custom rate limit handler
func customRateLimitHandler(w http.ResponseWriter, r *http.Request, result *ratelimit.RateLimitResult) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
    w.Header().Set("X-RateLimit-Remaining", "0")
    w.Header().Set("Retry-After", fmt.Sprintf("%d", result.RetryAfter))
    w.WriteHeader(http.StatusTooManyRequests)

    json.NewEncoder(w).Encode(map[string]interface{}{
        "error":       "rate_limit_exceeded",
        "message":     "You have exceeded the rate limit. Please slow down.",
        "limit":       result.Limit,
        "retry_after": result.RetryAfter,
        "reset_at":    result.ResetAt,
    })
}
```

### gRPC Server Integration

```go
import (
    "google.golang.org/grpc"
    "github.com/virtengine/virtengine/pkg/ratelimit"
)

func setupGRPCServer() error {
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

    ctx := context.Background()
    integration, err := ratelimit.QuickSetup(ctx, "redis://localhost:6379/0", logger)
    if err != nil {
        return err
    }

    // Configure interceptor
    interceptorConfig := ratelimit.GRPCInterceptorConfig{
        UserExtractor: extractUserFromMetadata,
        SkipMethods: []string{
            "/cosmos.base.tendermint.v1beta1.Service/GetNodeInfo",
            "/cosmos.base.tendermint.v1beta1.Service/GetSyncing",
        },
    }

    // Create gRPC server with rate limiting
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(
            integration.GetGRPCUnaryInterceptor(interceptorConfig),
        ),
        grpc.StreamInterceptor(
            integration.GetGRPCStreamInterceptor(interceptorConfig),
        ),
    )

    // Register services
    // ...

    return nil
}

func extractUserFromMetadata(ctx context.Context) string {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return ""
    }

    if users := md.Get("user-id"); len(users) > 0 {
        return users[0]
    }

    return ""
}
```

---

## Configuration

### Environment Variables

```bash
# Redis
REDIS_URL=redis://localhost:6379/0
REDIS_PREFIX=virtengine:ratelimit
REDIS_PASSWORD=your_password
REDIS_TLS=false

# Rate Limiting
RATELIMIT_ENABLED=true

# IP Limits
RATELIMIT_IP_PER_SECOND=10
RATELIMIT_IP_PER_MINUTE=300
RATELIMIT_IP_PER_HOUR=10000
RATELIMIT_IP_PER_DAY=100000
RATELIMIT_IP_BURST=20

# User Limits
RATELIMIT_USER_PER_SECOND=50
RATELIMIT_USER_PER_MINUTE=1000
RATELIMIT_USER_PER_HOUR=50000
RATELIMIT_USER_PER_DAY=500000
RATELIMIT_USER_BURST=100

# Bypass Detection
RATELIMIT_BYPASS_ENABLED=true
RATELIMIT_BYPASS_MAX_ATTEMPTS=100
RATELIMIT_BYPASS_BAN_DURATION=3600
RATELIMIT_BYPASS_ALERT_THRESHOLD=50

# Graceful Degradation
RATELIMIT_DEGRADATION_ENABLED=true
RATELIMIT_DEGRADATION_THRESHOLD_1=80
RATELIMIT_DEGRADATION_MULTIPLIER_1=0.7
RATELIMIT_DEGRADATION_THRESHOLD_2=90
RATELIMIT_DEGRADATION_MULTIPLIER_2=0.5

# Monitoring
RATELIMIT_METRICS_INTERVAL=30s
RATELIMIT_ALERTS_ENABLED=true
RATELIMIT_ALERT_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

### Configuration File (YAML)

```yaml
# config/ratelimit.yaml
redis:
  url: redis://localhost:6379/0
  prefix: virtengine:ratelimit
  password: ""
  tls: false

enabled: true

limits:
  ip:
    requests_per_second: 10
    requests_per_minute: 300
    requests_per_hour: 10000
    requests_per_day: 100000
    burst_size: 20

  user:
    requests_per_second: 50
    requests_per_minute: 1000
    requests_per_hour: 50000
    requests_per_day: 500000
    burst_size: 100

  endpoints:
    "/api/veid/*":
      requests_per_second: 5
      requests_per_minute: 100
      requests_per_hour: 1000
      burst_size: 10

    "/api/market/*":
      requests_per_second: 20
      requests_per_minute: 600
      requests_per_hour: 20000
      burst_size: 40

whitelist:
  ips:
    - "127.0.0.1"
    - "10.0.0.0/8"
    - "172.16.0.0/12"
  users:
    - "admin"
    - "service-account"
    - "monitoring"

bypass_detection:
  enabled: true
  max_failed_attempts_per_minute: 100
  ban_duration: 3600  # seconds
  alert_threshold: 50

graceful_degradation:
  enabled: true
  read_only_mode: true
  thresholds:
    - load_percentage: 80
      rate_limit_multiplier: 0.7
      priority:
        - "/api/veid/*"
    - load_percentage: 90
      rate_limit_multiplier: 0.5
      priority:
        - "/api/veid/verify"
    - load_percentage: 95
      rate_limit_multiplier: 0.3
      priority: []

monitoring:
  metrics_interval: 30s
  enable_alerts: true
  alert_webhook_url: ""
  alert_thresholds:
    blocked_requests_per_minute: 1000
    bypass_attempts_per_minute: 100
    banned_identifiers: 50
    load_percentage: 80.0
    top_blocked_ip_requests: 500
```

---

## Monitoring

### Prometheus Metrics

The rate limiter exposes the following Prometheus metrics:

```
# Total requests processed
virtengine_ratelimit_total_requests

# Allowed requests
virtengine_ratelimit_allowed_requests

# Blocked requests
virtengine_ratelimit_blocked_requests

# Bypass attempts
virtengine_ratelimit_bypass_attempts

# Banned identifiers
virtengine_ratelimit_banned_identifiers

# Current system load (0-100)
virtengine_ratelimit_current_load

# HTTP-specific metrics
virtengine_http_requests_total{method, path, status}
virtengine_http_requests_blocked_total{method, path, reason}
virtengine_http_ratelimit_check_duration_seconds{limit_type}

# gRPC-specific metrics
virtengine_grpc_requests_total{method, status}
virtengine_grpc_requests_blocked_total{method, reason}
virtengine_grpc_ratelimit_check_duration_seconds{limit_type}

# Alerts triggered
virtengine_ratelimit_alerts_triggered{severity, type}
```

### Grafana Dashboard

Import the provided Grafana dashboard JSON:

```json
{
  "dashboard": {
    "title": "VirtEngine Rate Limiting",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(virtengine_ratelimit_total_requests[5m])"
          }
        ]
      },
      {
        "title": "Blocked Requests",
        "targets": [
          {
            "expr": "rate(virtengine_ratelimit_blocked_requests[5m])"
          }
        ]
      },
      {
        "title": "Current Load",
        "targets": [
          {
            "expr": "virtengine_ratelimit_current_load"
          }
        ]
      }
    ]
  }
}
```

### Alerts

```yaml
# prometheus-alerts.yaml
groups:
  - name: ratelimiting
    rules:
      - alert: HighBlockRate
        expr: rate(virtengine_ratelimit_blocked_requests[5m]) > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High rate of blocked requests"
          description: "{{ $value }} requests per second are being blocked"

      - alert: PotentialDDoS
        expr: virtengine_ratelimit_bypass_attempts > 1000
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Potential DDoS attack detected"
          description: "{{ $value }} bypass attempts detected"

      - alert: HighSystemLoad
        expr: virtengine_ratelimit_current_load > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High system load"
          description: "System load is at {{ $value }}%"
```

---

## Testing

### Unit Tests

```bash
# Run all tests
go test ./pkg/ratelimit/...

# Run with coverage
go test -cover ./pkg/ratelimit/...

# Run specific test
go test -run TestRedisRateLimiter ./pkg/ratelimit/
```

### Integration Tests

```bash
# Start test Redis
docker run -d --name redis-test -p 6379:6379 redis:7-alpine

# Run integration tests
go test -tags=integration ./pkg/ratelimit/...

# Cleanup
docker stop redis-test && docker rm redis-test
```

### Load Testing

```bash
# Install vegeta
go install github.com/tsenart/vegeta@latest

# Test rate limiting
echo "GET http://localhost:1317/market/orders" | \
  vegeta attack -rate=100 -duration=30s | \
  vegeta report

# Expected output: ~70% 429 responses at 100 req/s
```

---

## Deployment

### Production Checklist

- [ ] Redis cluster configured with replication
- [ ] Rate limits tuned based on capacity testing
- [ ] Monitoring and alerting configured
- [ ] CDN/WAF configured (see [CDN_WAF.md](./RATELIMIT_CDN_WAF.md))
- [ ] Whitelists configured for critical services
- [ ] Incident response plan documented
- [ ] Client documentation published
- [ ] Load testing completed

### Redis Production Setup

```bash
# Redis Cluster (recommended for production)
redis-cli --cluster create \
  127.0.0.1:7000 127.0.0.1:7001 127.0.0.1:7002 \
  127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 \
  --cluster-replicas 1

# Update config
REDIS_URL=redis://localhost:7000?cluster=true
```

### Kubernetes Deployment

```yaml
# redis-deployment.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
spec:
  serviceName: redis
  replicas: 3
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
        volumeMounts:
        - name: data
          mountPath: /data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi

---
apiVersion: v1
kind: Service
metadata:
  name: redis
spec:
  selector:
    app: redis
  ports:
  - port: 6379
  clusterIP: None
```

---

## Troubleshooting

### Issue: Rate limiting not working

**Check:**
1. Redis connection: `redis-cli ping`
2. Configuration: `RATELIMIT_ENABLED=true`
3. Logs: Check for initialization errors

### Issue: Too many false positives

**Solution:**
1. Increase rate limits
2. Add to whitelist
3. Check if behind load balancer (IP extraction)

### Issue: High Redis memory usage

**Solution:**
1. Reduce TTLs in configuration
2. Implement Redis eviction policy
3. Scale Redis cluster

### Issue: Slow API responses

**Check:**
1. Redis latency: `redis-cli --latency`
2. Network latency to Redis
3. Consider caching rate limit results

---

## Next Steps

1. Review [Client Guide](./RATELIMIT_CLIENT_GUIDE.md)
2. Configure [CDN/WAF](./RATELIMIT_CDN_WAF.md)
3. Set up monitoring dashboards
4. Run load tests
5. Deploy to production

For questions or issues, please file an issue on GitHub or contact the security team.
