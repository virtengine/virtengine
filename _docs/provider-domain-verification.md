# Provider Domain Verification

## Overview

VirtEngine supports DNS-based domain verification for providers, similar to how email services (Gmail for Work, Zoho Mail) verify domain ownership. This ensures that providers can prove ownership of their domains before accepting marketplace orders.

## How It Works

### 1. Generate Verification Token

A provider generates a unique verification token for their domain:

```bash
virtengine tx provider generate-domain-token <domain> --from <provider-key>
```

Example:
```bash
virtengine tx provider generate-domain-token provider.example.com --from myprovider
```

Response:
```json
{
  "token": "a1b2c3d4e5f6...",
  "expires_at": 1738195200
}
```

### 2. Add DNS TXT Record

The provider must add a TXT record to their DNS configuration:

**Record Name:** `_virtengine-verification.<domain>`  
**Record Type:** TXT  
**Record Value:** `<token>` (from step 1)

Example for `provider.example.com`:
```
_virtengine-verification.provider.example.com    TXT    "a1b2c3d4e5f6..."
```

### 3. Verify Domain

Once the DNS record is propagated, verify the domain:

```bash
virtengine tx provider verify-domain --from <provider-key>
```

The system will:
- Query the DNS TXT record at `_virtengine-verification.<domain>`
- Check if the token matches
- Mark the domain as verified if successful

### 4. Verification Status

Check verification status:

```bash
virtengine query provider verification-status <provider-address>
```

Response:
```json
{
  "domain": "provider.example.com",
  "status": "verified",
  "verified_at": "2026-01-30T12:00:00Z",
  "expires_at": "2026-02-06T12:00:00Z"
}
```

## Token Expiration

- Verification tokens expire after **7 days** if not verified
- After expiration, generate a new token and update the DNS record
- Verified domains remain valid until the token expiration date

## Domain Verification Status

| Status | Description |
|--------|-------------|
| `pending` | Token generated, awaiting DNS verification |
| `verified` | Domain successfully verified |
| `failed` | Verification attempt failed (DNS lookup or token mismatch) |
| `expired` | Verification token expired without successful verification |

## Security Considerations

### Token Uniqueness
- Each token is cryptographically random (32 bytes, hex-encoded)
- Tokens are unique per provider and cannot be reused

### Rate Limiting
- Verification attempts are subject to rate limiting
- Multiple failed attempts may trigger cooldown periods

### Domain Validation
- Domain format is validated before token generation
- Prevents malicious domain patterns
- Maximum domain length: 253 characters
- Each label must be 1-63 characters

## Integration with Provider Registration

Domain verification is **optional** during provider registration but recommended for:
- Enhanced reputation and trust
- Priority in marketplace order matching
- Access to premium features (future)

Providers can register without domain verification, but they will be marked as "unverified" until verification is complete.

## DNS Configuration Examples

### Cloudflare
1. Go to DNS settings for your domain
2. Click "Add record"
3. Type: TXT
4. Name: `_virtengine-verification`
5. Content: `<your-token>`
6. TTL: Auto
7. Click "Save"

### Route 53 (AWS)
```bash
aws route53 change-resource-record-sets \
  --hosted-zone-id <zone-id> \
  --change-batch '{
    "Changes": [{
      "Action": "CREATE",
      "ResourceRecordSet": {
        "Name": "_virtengine-verification.provider.example.com",
        "Type": "TXT",
        "TTL": 300,
        "ResourceRecords": [{"Value": "\"<your-token>\""}]
      }
    }]
  }'
```

### Google Cloud DNS
```bash
gcloud dns record-sets create _virtengine-verification.provider.example.com \
  --zone=<zone-name> \
  --type=TXT \
  --ttl=300 \
  --rrdatas="<your-token>"
```

## Troubleshooting

### Verification Fails

1. **Check DNS propagation:**
   ```bash
   dig TXT _virtengine-verification.provider.example.com
   ```

2. **Verify token matches:**
   - Ensure no extra spaces or quotes
   - Token should be exactly as provided

3. **Wait for DNS propagation:**
   - Can take 5-60 minutes depending on TTL
   - Some providers may take longer

### Token Expired

Generate a new token:
```bash
virtengine tx provider generate-domain-token <domain> --from <provider-key>
```

Update your DNS record with the new token and verify again.

## API Reference

### Messages

#### MsgGenerateDomainVerificationToken
Generates a verification token for a provider's domain.

**Request:**
```json
{
  "owner": "virtengine1...",
  "domain": "provider.example.com"
}
```

**Response:**
```json
{
  "token": "a1b2c3d4e5f6...",
  "expires_at": 1738195200
}
```

#### MsgVerifyProviderDomain
Verifies a provider's domain via DNS TXT record lookup.

**Request:**
```json
{
  "owner": "virtengine1..."
}
```

**Response:**
```json
{
  "verified": true
}
```

### Events

#### EventProviderDomainVerificationStarted
Emitted when domain verification is initiated.

```json
{
  "owner": "virtengine1...",
  "domain": "provider.example.com",
  "token": "a1b2c3d4e5f6..."
}
```

#### EventProviderDomainVerified
Emitted when domain is successfully verified.

```json
{
  "owner": "virtengine1...",
  "domain": "provider.example.com"
}
```

## Implementation Details

### Storage

Domain verification records are stored in the provider module state:

**Key:** `DomainVerificationPrefix + ProviderAddress`  
**Value:** JSON-encoded `DomainVerificationRecord`

```go
type DomainVerificationRecord struct {
    ProviderAddress string
    Domain          string
    Token           string
    Status          DomainVerificationStatus
    GeneratedAt     int64
    VerifiedAt      int64
    ExpiresAt       int64
}
```

### DNS Query

The system uses Go's `net.LookupTXT()` to query DNS records. The expected record format:

```
_virtengine-verification.<domain>    TXT    "<token>"
```

### Keeper Methods

- `GenerateDomainVerificationToken(ctx, providerAddr, domain)` - Generates token
- `VerifyProviderDomain(ctx, providerAddr)` - Performs DNS verification
- `GetDomainVerificationRecord(ctx, providerAddr)` - Retrieves record
- `IsDomainVerified(ctx, providerAddr)` - Checks if domain is verified
- `DeleteDomainVerificationRecord(ctx, providerAddr)` - Removes record

## Future Enhancements

- DNSSEC validation for enhanced security
- Support for multiple domains per provider
- Automatic re-verification before expiration
- Domain verification badges in marketplace UI
- Integration with provider reputation scoring
