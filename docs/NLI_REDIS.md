# NLI Redis Session Store and Rate Limiting

This document describes Redis-backed session storage and distributed rate limiting for the NLI service.

## Configuration

The NLI configuration (`pkg/nli.Config`) supports Redis-backed sessions and distributed rate limiting:

```json
{
  "session_store": {
    "backend": "redis",
    "redis_url": "redis://localhost:6379/0",
    "redis_prefix": "virtengine:nli:session:",
    "session_ttl": "30m",
    "max_history_length": 20,
    "redis_max_memory_mb": 512,
    "redis_eviction_policy": "allkeys-lru"
  },
  "distributed_rate_limiter": {
    "enabled": true,
    "redis_url": "redis://localhost:6379/0",
    "redis_prefix": "virtengine:nli:ratelimit",
    "requests_per_minute": 60,
    "requests_per_second": 5,
    "burst_size": 10,
    "redis_max_memory_mb": 512,
    "redis_eviction_policy": "allkeys-lru"
  },
  "metrics_namespace": "virtengine"
}
```

### TTLs
- `session_store.session_ttl` controls session expiration for Redis entries.

### Redis Eviction Policy
- `redis_max_memory_mb` applies a Redis `maxmemory` setting (MB).
- `redis_eviction_policy` applies `maxmemory-policy` (e.g. `allkeys-lru`, `volatile-ttl`).

These settings are applied on service startup using Redis `CONFIG SET`. Ensure Redis ACLs allow configuration changes; otherwise leave them unset and configure Redis externally.

## Metrics

Prometheus metrics are exported under the `nli` subsystem and `metrics_namespace` namespace:

- `virtengine_nli_active_sessions`
- `virtengine_nli_session_operations_total{operation}`
- `virtengine_nli_session_operation_duration_seconds{operation}`
- `virtengine_nli_ratelimit_hits_total`
- `virtengine_nli_ratelimit_passes_total`
- `virtengine_nli_requests_total{status}`
- `virtengine_nli_request_duration_seconds{intent}`
- `virtengine_nli_intents_classified_total{intent}`
- `virtengine_nli_errors_total{type}`

Use the global metrics handler from `pkg/observability` to expose metrics.

## Rate Limiting

Distributed rate limiting uses `pkg/ratelimit` with user-based limits keyed by NLI session ID. The limiter is initialized by `NewDistributedRateLimiter`.

## Load Testing

Run the NLI load test (short mode skips):

```bash
go test -v ./tests/load/... -run TestNLIBurstLoad
```

Run the NLI benchmark:

```bash
go test -bench=BenchmarkNLIBurst -benchtime=30s ./tests/load/...
```
