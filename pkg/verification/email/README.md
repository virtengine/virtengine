# Email Verification Service

This package provides email OTP/link verification with signed attestations for the VEID identity verification module.

## Features

- **OTP Verification**: 6-digit one-time passwords with configurable TTL (60-3600 seconds)
- **Magic Link Verification**: Secure token-based verification links (24-hour validity)
- **Multiple Email Providers**: Support for AWS SES, SendGrid, and mock provider for testing
- **Signed Attestations**: Ed25519 signed verification attestations for on-chain storage
- **Rate Limiting**: Configurable rate limiting per account/IP
- **Audit Logging**: Comprehensive audit trail for all verification events
- **Prometheus Metrics**: Delivery success/failure, verification latency, attestation counts
- **Webhook Support**: Process delivery status webhooks from email providers

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     Email Verification Service                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌───────────────┐     ┌──────────────┐     ┌────────────────────────┐ │
│  │  Initiate     │────▶│   Cache      │────▶│  Email Provider        │ │
│  │  Verification │     │  (Redis)     │     │  (SES/SendGrid/Mock)   │ │
│  └───────────────┘     └──────────────┘     └────────────────────────┘ │
│         │                    │                        │                 │
│         ▼                    ▼                        ▼                 │
│  ┌───────────────┐     ┌──────────────┐     ┌────────────────────────┐ │
│  │   Verify      │────▶│  Challenge   │────▶│  Template Manager      │ │
│  │   Challenge   │     │  Storage     │     │  (HTML/Text templates) │ │
│  └───────────────┘     └──────────────┘     └────────────────────────┘ │
│         │                                                               │
│         ▼                                                               │
│  ┌───────────────┐     ┌──────────────┐     ┌────────────────────────┐ │
│  │   Create      │────▶│   Signer     │────▶│   Audit Logger         │ │
│  │  Attestation  │     │  (Ed25519)   │     │                        │ │
│  └───────────────┘     └──────────────┘     └────────────────────────┘ │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## Configuration

### Basic Configuration

```go
config := email.DefaultConfig()
config.Provider = "ses"                          // ses, sendgrid, or mock
config.FromAddress = "noreply@virtengine.io"
config.FromName = "VirtEngine Identity"
config.BaseURL = "https://verify.virtengine.io"
config.OTPLength = 6
config.OTPTTLSeconds = 600                       // 10 minutes
config.MaxAttempts = 5
config.MaxResends = 3
```

### AWS SES Configuration

```go
sesConfig := email.SESConfig{
    Region:     "us-east-1",
    FromDomain: "virtengine.io",
}
provider, err := email.NewSESProvider(sesConfig, logger)
```

### SendGrid Configuration

```go
sgConfig := email.SendGridConfig{
    APIKey:     os.Getenv("SENDGRID_API_KEY"),
    FromDomain: "virtengine.io",
}
provider, err := email.NewSendGridProvider(sgConfig, logger)
```

## Usage

### Initialize Service

```go
import (
    "github.com/virtengine/virtengine/pkg/verification/email"
    "github.com/virtengine/virtengine/pkg/cache"
)

// Create cache
redisCache, err := cache.NewRedisCache[string, *email.EmailChallenge](redisURL)
if err != nil {
    return err
}

// Create service
service, err := email.NewService(
    ctx,
    config,
    logger,
    email.WithCache(redisCache),
    email.WithProvider(provider),
    email.WithSigner(signerService),
    email.WithAuditLogger(auditLogger),
    email.WithRateLimiter(rateLimiter),
)
if err != nil {
    return err
}
defer service.Close()
```

### Initiate Verification

```go
resp, err := service.InitiateVerification(ctx, &email.InitiateRequest{
    AccountAddress: "virtengine1xyz...",
    Email:          "user@example.com",
    Method:         email.MethodOTP,  // or email.MethodMagicLink
    IPAddress:      clientIP,
    UserAgent:      userAgent,
})
if err != nil {
    return err
}

// Response contains ChallengeID, MaskedEmail, ExpiresAt
fmt.Printf("Challenge ID: %s, Masked: %s\n", resp.ChallengeID, resp.MaskedEmail)
```

### Verify Challenge

```go
resp, err := service.VerifyChallenge(ctx, &email.VerifyRequest{
    ChallengeID:    challengeID,
    Secret:         "123456",  // OTP code or magic link token
    AccountAddress: "virtengine1xyz...",
})
if err != nil {
    return err
}

if resp.Verified {
    // Attestation created and stored
    fmt.Printf("Attestation ID: %s\n", resp.AttestationID)
}
```

### Resend Verification

```go
resp, err := service.ResendVerification(ctx, &email.ResendRequest{
    ChallengeID:    challengeID,
    AccountAddress: "virtengine1xyz...",
    Email:          "user@example.com",  // Required for resend
})
if err != nil {
    if errors.Is(err, email.ErrResendCooldown) {
        // User must wait before resending
    }
    return err
}
```

### Process Webhooks

```go
// For AWS SES (via SNS)
err := service.ProcessWebhook(ctx, "ses", snsPayload)

// For SendGrid
err := service.ProcessWebhook(ctx, "sendgrid", webhookPayload)
```

## Metrics

The service exposes Prometheus metrics with the prefix `veid_email_verification_`:

| Metric | Type | Description |
|--------|------|-------------|
| `challenges_created_total` | Counter | Total challenges created by method |
| `emails_sent_total` | Counter | Emails sent by provider, template, status |
| `verification_attempts_total` | Counter | Verification attempts by result |
| `attestations_created_total` | Counter | Successful attestations created |
| `active_challenges` | Gauge | Currently active challenges |
| `verification_latency_seconds` | Histogram | Time from initiation to verification |
| `email_send_latency_seconds` | Histogram | Email delivery latency by provider |
| `rate_limit_hits_total` | Counter | Rate limit rejections |

### Grafana Dashboard

Import the dashboard JSON from `_docs/grafana/email-verification-dashboard.json` (if available) or create panels for:

1. **Email Delivery Health**
   - Delivery success rate (%)
   - Bounce/complaint rates
   - Send latency P50/P95/P99

2. **Verification Flow**
   - Initiation rate (per minute)
   - Verification success rate
   - Average time to verify

3. **Security**
   - Rate limit hits
   - Max attempts exceeded
   - Failed verification by IP

## Security Considerations

### OTP Storage

- OTPs are **never stored in plaintext**
- SHA256 hash is stored in cache
- Original OTP is only in the email

### Email Address Privacy

- Only the SHA256 hash of the email is stored on-chain
- Domain hash is stored separately for organizational detection
- Masked email (e.g., `t***@example.com`) shown in responses

### Rate Limiting

Default limits (configurable):

- Per account: 5 verifications per hour
- Per IP: 10 verifications per hour
- Global: 1000 verifications per minute

### Replay Protection

- Each challenge has a unique nonce
- Challenges are single-use (consumed on verification)
- TTL prevents delayed replay

## Operational Runbook

### High Bounce Rate

1. Check SES/SendGrid console for bounce reasons
2. Review email domain reputation
3. Enable bounce webhook processing if not configured
4. Consider implementing email validation at initiation

### Rate Limit Spikes

1. Check for abuse patterns (same IP/account)
2. Review audit logs for suspicious activity
3. Temporarily increase limits if legitimate traffic spike
4. Consider IP-based blocking for persistent abusers

### Low Verification Rate

1. Check email delivery success rate
2. Review OTP TTL (may need to increase)
3. Check for spam folder issues
4. Review email template for clarity

### Cache Failures

1. Check Redis connectivity
2. Review cache memory usage
3. Verify TTL settings aren't causing early eviction
4. Consider cache replication for high availability

## Integration Tests

Run tests with:

```bash
go test ./pkg/verification/email/... -v
```

For integration tests with actual email providers (sandbox mode):

```bash
SENDGRID_API_KEY=xxx go test ./pkg/verification/email/... -v -tags=integration
```

## Dependencies

- `pkg/cache` - Redis cache interface
- `pkg/verification/signer` - Attestation signing service
- `pkg/verification/audit` - Audit logging
- `pkg/verification/ratelimit` - Rate limiting
- `x/veid/types` - VEID attestation types
