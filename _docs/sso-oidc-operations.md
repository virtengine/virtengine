# SSO/OIDC Verification Service Operations Runbook

## Overview

This runbook documents the operational procedures for the VEID SSO/OIDC Verification Service. The service enables users to link their blockchain accounts with verified SSO identities from providers like Google, Microsoft, and GitHub.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   User Wallet   │────>│  SSO Service     │────>│  OIDC Provider  │
│   (Client)      │     │  (Off-Chain)     │     │  (Google/MS)    │
└────────┬────────┘     └────────┬─────────┘     └─────────────────┘
         │                       │
         │                       v
         │              ┌──────────────────┐
         │              │  Signer Service  │
         │              │  (HSM/Vault)     │
         │              └────────┬─────────┘
         │                       │
         v                       v
┌─────────────────────────────────────────────────┐
│              VirtEngine Chain                    │
│  ┌─────────────┐  ┌─────────────────────────┐  │
│  │  x/veid     │  │  SSO Linkage Records    │  │
│  │  Keeper     │──│  Nonce Tracking         │  │
│  │             │  │  Attestations           │  │
│  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

## Component Overview

| Component | Location | Purpose |
|-----------|----------|---------|
| OIDC Verifier | `pkg/verification/oidc/` | Token validation, JWKS caching |
| SSO Service | `pkg/verification/sso/` | Challenge management, attestation creation |
| SSO Keeper | `x/veid/keeper/sso_linkage.go` | On-chain linkage storage, nonce tracking |
| SSO Types | `x/veid/types/sso_*.go` | Message types, attestation schema |
| Admin CLI | `sdk/go/cli/sso.go` | Configuration management |

## Configuration

### Environment Variables

```bash
# SSO Service Configuration
SSO_ENABLED=true
SSO_CHALLENGE_TTL_SECONDS=600
SSO_ATTESTATION_VALIDITY_DAYS=365
SSO_MAX_CHALLENGES_PER_ACCOUNT=3

# OIDC Configuration
OIDC_JWKS_CACHE_TTL_SECONDS=3600
OIDC_JWKS_REFRESH_INTERVAL_SECONDS=1800
OIDC_HTTP_TIMEOUT_SECONDS=30
OIDC_MAX_CLOCK_SKEW_SECONDS=300

# Rate Limiting
SSO_RATE_LIMIT_ENABLED=true
SSO_RATE_LIMIT_MAX_REQUESTS=10
SSO_RATE_LIMIT_WINDOW_SECONDS=300
```

### Issuer Configuration

Create an issuer configuration file (`sso-issuers.yaml`):

```yaml
issuers:
  - issuer: "https://accounts.google.com"
    client_id: "${GOOGLE_CLIENT_ID}"
    client_secret: "${GOOGLE_CLIENT_SECRET}"
    provider_type: "google"
    score_weight: 250  # 2.5%
    enabled: true
    required_scopes:
      - "openid"
      - "email"
      - "profile"
    require_email_verified: true

  - issuer: "https://login.microsoftonline.com/common/v2.0"
    client_id: "${MICROSOFT_CLIENT_ID}"
    client_secret: "${MICROSOFT_CLIENT_SECRET}"
    provider_type: "microsoft"
    score_weight: 300  # 3.0%
    enabled: true
    required_scopes:
      - "openid"
      - "email"
      - "profile"
    require_email_verified: true
    allowed_tenants:
      - "${ALLOWED_TENANT_ID}"

  - issuer: "https://token.actions.githubusercontent.com"
    client_id: "${GITHUB_CLIENT_ID}"
    client_secret: "${GITHUB_CLIENT_SECRET}"
    provider_type: "github"
    score_weight: 200  # 2.0%
    enabled: true
```

## Deployment

### Prerequisites

1. **Signer Service**: HSM or Vault-based signing service must be deployed and configured
2. **Redis** (recommended): For production challenge storage
3. **Monitoring**: Prometheus/Grafana for metrics

### Deployment Steps

1. **Configure Secrets**
   ```bash
   # Store OIDC client credentials securely
   vault kv put secret/sso/google \
     client_id="your-google-client-id" \
     client_secret="your-google-client-secret"

   vault kv put secret/sso/microsoft \
     client_id="your-microsoft-client-id" \
     client_secret="your-microsoft-client-secret"
   ```

2. **Deploy SSO Service**
   ```bash
   # Using Docker Compose
   docker-compose -f docker-compose.sso.yaml up -d

   # Using Kubernetes
   kubectl apply -f deploy/sso-service.yaml
   ```

3. **Verify Deployment**
   ```bash
   # Health check
   curl http://sso-service:8080/health

   # Check issuer connectivity
   virtengine sso issuer list
   ```

## Operations

### Adding a New Issuer

1. **Register OAuth Application with Provider**
   - Google: [Google Cloud Console](https://console.cloud.google.com)
   - Microsoft: [Azure Portal](https://portal.azure.com)
   - Generic OIDC: Provider's developer console

2. **Configure Redirect URI**
   ```
   https://your-domain.com/api/v1/sso/callback
   ```

3. **Add Issuer to Configuration**
   ```bash
   virtengine sso issuer add https://issuer.example.com \
     --client-id="your-client-id" \
     --provider-type=oidc \
     --score-weight=150 \
     --enabled=true
   ```

4. **Validate Configuration**
   ```bash
   virtengine sso config validate sso-issuers.yaml
   ```

5. **Restart SSO Service**
   ```bash
   kubectl rollout restart deployment/sso-service
   ```

### Removing an Issuer

1. **Disable the Issuer First**
   ```bash
   virtengine sso issuer update https://issuer.example.com --enabled=false
   ```

2. **Monitor for Active Sessions** (wait 24 hours if possible)

3. **Remove from Configuration**
   ```bash
   virtengine sso issuer remove https://issuer.example.com
   ```

### Revoking a Linkage

```bash
# Via CLI (requires authorized signer)
virtengine tx veid revoke-sso-linkage \
  --linkage-id="linkage-uuid" \
  --reason="user request" \
  --from=admin \
  --chain-id=virtengine-1
```

## Key Rotation

### Attestation Signing Key Rotation

The signer service handles key rotation automatically. Manual rotation:

1. **Generate New Key**
   ```bash
   virtengine signer key rotate \
     --reason="scheduled rotation" \
     --overlap-hours=24
   ```

2. **Monitor Rotation Progress**
   ```bash
   virtengine signer key status
   ```

3. **Verify New Key Active**
   ```bash
   virtengine signer key list --state=active
   ```

### JWKS Rotation

JWKS rotation is handled automatically by providers. To force refresh:

```bash
# Force JWKS refresh for an issuer
virtengine sso issuer refresh-jwks https://accounts.google.com
```

## Monitoring

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `sso_challenge_created_total` | Total challenges created | - |
| `sso_challenge_completed_total` | Completed verifications | - |
| `sso_challenge_failed_total` | Failed verifications | > 10% of total |
| `sso_attestation_created_total` | Attestations issued | - |
| `sso_linkage_created_total` | On-chain linkages created | - |
| `sso_nonce_used_total` | Nonces consumed | - |
| `oidc_token_validation_duration_seconds` | Token validation time | p99 > 1s |
| `oidc_jwks_cache_hit_total` | JWKS cache hits | hit rate < 90% |
| `oidc_jwks_fetch_errors_total` | JWKS fetch failures | > 0 sustained |

### Alerts

```yaml
# Prometheus alert rules
groups:
  - name: sso-alerts
    rules:
      - alert: SSOVerificationFailureRateHigh
        expr: rate(sso_challenge_failed_total[5m]) / rate(sso_challenge_created_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "SSO verification failure rate is high"

      - alert: JWKSFetchErrors
        expr: rate(oidc_jwks_fetch_errors_total[5m]) > 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "JWKS fetch errors detected"

      - alert: SSOServiceUnhealthy
        expr: up{job="sso-service"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "SSO service is down"
```

### Logging

Structured logging is used for all SSO operations. Key log events:

| Event | Level | Fields |
|-------|-------|--------|
| challenge_created | INFO | challenge_id, account, provider |
| challenge_completed | INFO | challenge_id, account, provider, linkage_id |
| challenge_failed | WARN | challenge_id, account, provider, error |
| token_verified | DEBUG | issuer, subject_hash |
| jwks_refreshed | DEBUG | issuer, key_count |
| nonce_used | INFO | nonce_hash, account |

## Troubleshooting

### Common Issues

#### 1. JWKS Fetch Failures

**Symptoms**: Token verification fails, JWKS cache misses

**Causes**:
- Network connectivity issues
- Provider outage
- DNS resolution failure

**Resolution**:
```bash
# Check connectivity to provider
curl -v https://accounts.google.com/.well-known/openid-configuration

# Force JWKS refresh
virtengine sso issuer refresh-jwks https://accounts.google.com

# Check JWKS cache status
virtengine sso issuer status https://accounts.google.com
```

#### 2. Token Validation Failures

**Symptoms**: Valid tokens being rejected

**Causes**:
- Clock skew between nodes
- Expired tokens
- Audience mismatch

**Resolution**:
```bash
# Check node time synchronization
timedatectl status

# Verify client ID configuration
virtengine sso issuer show https://accounts.google.com

# Check token claims (decode without verifying for debugging)
echo $TOKEN | cut -d. -f2 | base64 -d | jq
```

#### 3. Nonce Replay Detection

**Symptoms**: Legitimate verifications rejected as replays

**Causes**:
- User clicked "verify" multiple times
- Network retry caused duplicate submission
- Nonce not properly generated

**Resolution**:
```bash
# Check nonce status
virtengine query veid sso-nonce <nonce-hash>

# Verify nonce was used for which linkage
virtengine query veid sso-linkage <linkage-id>
```

#### 4. Rate Limiting

**Symptoms**: 429 errors, verification requests rejected

**Causes**:
- Too many verification attempts
- Potential abuse

**Resolution**:
```bash
# Check rate limit status for account
virtengine query veid rate-limit-status <account-address>

# If legitimate, consider increasing limits
# Update configuration and restart service
```

### Debug Mode

Enable debug logging:

```bash
export SSO_LOG_LEVEL=debug
export OIDC_LOG_LEVEL=debug
```

## Security Considerations

### Secrets Management

- **Never log** client secrets or tokens
- Use **Vault or HSM** for signing keys
- Rotate client secrets **annually**
- Monitor for **credential exposure**

### Attack Prevention

| Attack | Mitigation |
|--------|-----------|
| Token replay | Nonce tracking, short TTL |
| CSRF | State parameter validation |
| Session fixation | Unique challenge per request |
| Denial of service | Rate limiting, max challenges |
| Linkage hijacking | Account signature verification |

### Audit Requirements

All SSO operations are logged to the audit service. Retain logs for:
- Linkage creation: **7 years** (compliance)
- Verification attempts: **90 days**
- Failed attempts: **1 year**

## Maintenance

### Daily Tasks

- [ ] Check alert dashboard
- [ ] Verify JWKS cache health
- [ ] Monitor error rates

### Weekly Tasks

- [ ] Review failed verification patterns
- [ ] Check disk space for audit logs
- [ ] Verify backup integrity

### Monthly Tasks

- [ ] Review and update issuer policies
- [ ] Audit rate limit thresholds
- [ ] Test disaster recovery procedures

### Quarterly Tasks

- [ ] Rotate attestation signing keys
- [ ] Review access controls
- [ ] Security audit of configuration

## Disaster Recovery

### Service Failure

1. **Identify** the failure mode
2. **Restart** the service
3. **Verify** JWKS caches are rebuilt
4. **Monitor** for cascading failures

### Key Compromise

1. **Immediately rotate** the compromised key
2. **Revoke** all attestations signed with the key
3. **Issue** new attestations with the new key
4. **Notify** affected users
5. **Investigate** the compromise

### Provider Outage

1. **Disable** the affected issuer
2. **Notify** users of temporary unavailability
3. **Monitor** provider status page
4. **Re-enable** when provider recovers

## Support

For issues not covered in this runbook:

- **Internal**: #virtengine-sso Slack channel
- **On-call**: Page via PagerDuty
- **Documentation**: [Internal Wiki](https://wiki.internal/veid/sso)
