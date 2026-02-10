# VEID Verification Service Deployment Guide

This document describes how to deploy and scale the VEID verification shared infrastructure.

## Overview

The VEID verification infrastructure consists of several components:

- **Signer Service**: Issues verifiable attestations with key rotation support
- **Key Storage**: Secure storage for signing keys (Memory, File, Vault, HSM)
- **Nonce Store**: Replay protection for attestations
- **Rate Limiter**: Abuse prevention and traffic control
- **Audit Logger**: Append-only audit trail
- **Metrics Collector**: Prometheus metrics for monitoring

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Verification Service                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Signer    │  │  Rate       │  │   Audit Logger          │  │
│  │   Service   │  │  Limiter    │  │                         │  │
│  └──────┬──────┘  └──────┬──────┘  └────────────┬────────────┘  │
│         │                │                      │               │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌───────────▼─────────────┐  │
│  │   Key       │  │   Redis     │  │   Audit Storage         │  │
│  │   Storage   │  │             │  │   (File/Redis/ES)       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
│         │                                                       │
│  ┌──────▼──────────────────────────────────────────────────┐   │
│  │   Nonce Store (Memory/Redis)                             │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Go 1.25.5+ for building
- Redis 6.0+ (for production rate limiting and nonce storage)
- HashiCorp Vault (optional, for secure key storage)
- Prometheus (for metrics collection)
- Grafana (for visualization, optional)

## Configuration

### Signer Service Configuration

```yaml
# config/verification-signer.yaml
signer:
  signer_id: "veid-signer-prod-1"
  signer_name: "VEID Production Signer"
  validator_address: "virtenginevaloper1..."
  
  key_storage_type: "vault"  # memory, file, vault, hsm
  key_storage_config:
    address: "https://vault.internal:8200"
    token: "${VAULT_TOKEN}"
    mount_path: "secret"
    key_path: "veid/signer/keys"
  
  default_algorithm: "Ed25519Signature2020"
  
  key_policy:
    max_key_age_seconds: 7776000  # 90 days
    rotation_overlap_seconds: 604800  # 7 days
    min_rotation_notice_seconds: 259200  # 3 days
    max_pending_keys: 2
    require_successor_key: true
    allow_emergency_revocation: true
    key_algorithms:
      - "Ed25519Signature2020"
      - "EcdsaSecp256k1Signature2019"
    min_key_strength: 256
  
  audit_log_enabled: true
  metrics_enabled: true
  service_endpoint: "https://signer.veid.virtengine.io"
```

### Rate Limiter Configuration

```yaml
# config/verification-ratelimit.yaml
ratelimit:
  redis_url: "redis://redis:6379/1"
  redis_prefix: "virtengine:veid:ratelimit"
  enabled: true
  
  verification_limits:
    email_verification:
      requests_per_hour: 10
      requests_per_day: 50
      max_failures_per_hour: 5
      cooldown_minutes: 60
    
    sms_verification:
      requests_per_hour: 5
      requests_per_day: 20
      max_failures_per_hour: 3
      cooldown_minutes: 120
    
    facial_verification:
      requests_per_hour: 20
      requests_per_day: 100
      max_failures_per_hour: 10
      cooldown_minutes: 30
  
  abuse_scoring:
    enabled: true
    score_threshold_for_captcha: 40
    score_threshold_for_block: 80
    score_decay_per_hour: 5
```

### Nonce Store Configuration

```yaml
# config/verification-nonce.yaml
nonce:
  backend: "redis"  # memory, redis
  
  policy:
    nonce_window_seconds: 3600
    require_timestamp_binding: true
    max_clock_skew_seconds: 300
    max_nonces_per_issuer: 10000
    require_issuer_binding: true
    track_nonce_history: true
  
  cleanup_interval: "5m"
  
  redis:
    url: "redis://redis:6379/2"
    prefix: "virtengine:veid:nonce"
    pool_size: 10
```

## Deployment

### Docker Compose (Development)

```yaml
version: '3.8'

services:
  verification-signer:
    image: virtengine/verification-signer:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - SIGNER_ID=dev-signer-1
      - KEY_STORAGE_TYPE=memory
      - REDIS_URL=redis://redis:6379
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  redis_data:
```

### Kubernetes (Production)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: verification-signer
  namespace: virtengine-veid
spec:
  replicas: 3
  selector:
    matchLabels:
      app: verification-signer
  template:
    metadata:
      labels:
        app: verification-signer
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
    spec:
      containers:
      - name: signer
        image: virtengine/verification-signer:v1.0.0
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: SIGNER_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: VAULT_ADDR
          value: "https://vault.virtengine.io:8200"
        - name: VAULT_TOKEN
          valueFrom:
            secretKeyRef:
              name: vault-token
              key: token
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: verification-signer
  namespace: virtengine-veid
spec:
  selector:
    app: verification-signer
  ports:
  - port: 8080
    targetPort: 8080
    name: http
  - port: 9090
    targetPort: 9090
    name: metrics
```

### Vault Configuration

```bash
# Enable the KV secrets engine
vault secrets enable -path=secret kv-v2

# Create a policy for the signer
vault policy write veid-signer - <<EOF
path "secret/data/veid/signer/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF

# Create a role for the signer (if using Kubernetes auth)
vault write auth/kubernetes/role/veid-signer \
  bound_service_account_names=verification-signer \
  bound_service_account_namespaces=virtengine-veid \
  policies=veid-signer \
  ttl=1h
```

## Scaling Recommendations

### Horizontal Scaling

The verification signer service is designed for horizontal scaling:

| Component | Recommendation |
|-----------|----------------|
| Signer Service | 3+ replicas behind load balancer |
| Redis | Redis Cluster with 3+ nodes |
| Vault | HA deployment with 3+ nodes |
| Metrics | Prometheus with Thanos for HA |

### Resource Sizing

| Load Level | CPU | Memory | Redis Memory |
|------------|-----|--------|--------------|
| Low (<100 req/s) | 0.5 vCPU | 256 MB | 1 GB |
| Medium (100-1000 req/s) | 2 vCPU | 1 GB | 4 GB |
| High (1000+ req/s) | 4 vCPU | 2 GB | 8 GB |

### Performance Tuning

1. **Redis Connection Pooling**: Set pool size based on expected concurrency
2. **Nonce Cleanup**: Adjust cleanup interval based on nonce volume
3. **Audit Buffer**: Configure buffer size to batch writes efficiently
4. **Key Caching**: Active keys are cached in memory for performance

## Monitoring

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `signer_sign_requests_total` | Total signing requests | - |
| `signer_sign_latency_seconds` | Signing latency | p99 > 100ms |
| `signer_key_age_seconds` | Active key age | > 80 days |
| `nonce_store_size` | Number of tracked nonces | > 80% capacity |
| `ratelimit_blocked_total` | Blocked requests | Spike detection |
| `service_health` | Service health status | < 1 |

### Grafana Dashboard

Import the verification dashboard from `deploy/grafana/veid-verification.json`.

### Alerts

```yaml
groups:
- name: veid-verification
  rules:
  - alert: VerificationSignerUnhealthy
    expr: virtengine_verification_service_health{service="signer"} < 1
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "Verification signer unhealthy"
      
  - alert: VerificationKeyNearExpiry
    expr: virtengine_verification_signer_key_age_seconds > 6912000  # 80 days
    for: 1h
    labels:
      severity: warning
    annotations:
      summary: "Signing key approaching expiration"
      
  - alert: VerificationHighErrorRate
    expr: rate(virtengine_verification_signer_errors_total[5m]) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High verification error rate"
```

## Security Considerations

1. **Key Storage**: Always use Vault or HSM in production
2. **Network**: Signer should only be accessible from internal network
3. **TLS**: Use TLS for all connections (Vault, Redis, API)
4. **Secrets**: Never log or expose private keys
5. **Rotation**: Automate key rotation before expiry
6. **Audit**: Enable audit logging and monitor for anomalies

## Troubleshooting

### Common Issues

1. **No active key available**
   - Check key storage connectivity
   - Verify key generation succeeded
   - Check for key expiration

2. **Nonce validation failures**
   - Verify clock synchronization
   - Check Redis connectivity
   - Review nonce window settings

3. **High latency**
   - Check Redis latency
   - Review key storage performance
   - Monitor CPU and memory usage

4. **Rate limit false positives**
   - Review rate limit thresholds
   - Check for abuse score misconfiguration
   - Verify IP detection is accurate
