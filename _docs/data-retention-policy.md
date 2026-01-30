# Data Retention & Archival Policy

## Overview

VirtEngine implements a comprehensive data retention and archival system to manage the lifecycle of identity artifacts in compliance with global data protection regulations (GDPR, HIPAA, CCPA) while maintaining blockchain efficiency and storage costs.

## Architecture

### Three-Tier Storage Model

```
┌─────────────────────────────────────────────────────────┐
│                 VirtEngine Blockchain                    │
│  ┌──────────────────────────────────────────────────┐  │
│  │  On-Chain: Hashes, Metadata, Retention Policies  │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────┬───────────────────────────────────────┘
                  │
    ┌─────────────┴─────────────┐
    │                           │
┌───▼──────────────┐  ┌────────▼────────────┐
│  Hot Storage     │  │  Cold Archive       │
│  (Active Data)   │  │  (Historical Data)  │
│                  │  │                     │
│  - IPFS          │  │  - S3 Glacier       │
│  - Waldur        │  │  - Azure Archive    │
│  - Fast Access   │  │  - Local Tape       │
│  - Age < 90 days │  │  - Age > 90 days    │
└──────────────────┘  └─────────────────────┘
```

## Data Retention Policies

### Retention Types

VirtEngine supports multiple retention policy types:

1. **Duration-Based**: Retains data for a specific time period
   ```go
   policy := NewRetentionPolicyDuration(
       policyID,
       durationSeconds,  // e.g., 90 * 24 * 60 * 60 for 90 days
       createdAt,
       createdAtBlock,
       deleteOnExpiry,
   )
   ```

2. **Block-Count-Based**: Retains data for a specific number of blocks
   ```go
   policy := NewRetentionPolicyBlockCount(
       policyID,
       blockCount,  // e.g., 12960 blocks (~90 days)
       createdAt,
       createdAtBlock,
       deleteOnExpiry,
   )
   ```

3. **Indefinite**: Retains data indefinitely (e.g., verification records)
   ```go
   policy := NewRetentionPolicyIndefinite(policyID, createdAt, createdAtBlock)
   ```

4. **Until-Revoked**: Retains data until explicitly revoked by user
   ```go
   policy := NewRetentionPolicyUntilRevoked(policyID, createdAt, createdAtBlock)
   ```

### Default Retention Policies by Artifact Type

| Artifact Type | Storage Location | Default Retention | Delete After Verification |
|---------------|------------------|-------------------|---------------------------|
| Raw Image | Off-chain only | 7 days | Yes |
| Processed Image | Off-chain only | 1 day | Yes |
| Face Embedding | Hash on-chain, encrypted off-chain | 365 days | No |
| Document Hash | On-chain | Indefinite | No |
| Biometric Hash | On-chain | Indefinite | No |
| Verification Record | On-chain | Indefinite | No |
| OCR Data | Off-chain only | 7 days | Yes |

## Archival Process

### Eligibility Criteria

An artifact becomes eligible for archival when ALL conditions are met:

1. **Minimum Age**: Artifact is at least 90 days old (configurable)
2. **Low Access**: Accessed fewer than 5 times total
3. **No Recent Access**: Not accessed in the last 30 days
4. **Has Retention Policy**: A valid retention policy exists
5. **Not Excluded**: Artifact type is not in exclusion list

### Archival Workflow

```
┌─────────────────────┐
│  BeginBlocker       │ Checks for eligible artifacts every 100 blocks
│  (every 100 blocks) │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Eligibility Check   │ Apply archival criteria
│  - Age > 90 days    │
│  - Access count < 5 │
│  - Last access > 30d│
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Archive Artifact   │
│  1. Encrypt with    │ Archive-specific encryption key (supports rotation)
│     archive key     │
│  2. Upload to cold  │ S3 Glacier, Azure Archive, or local filesystem
│     storage         │
│  3. Verify checksum │ SHA-256 integrity verification
│  4. Update index    │ Mark as archived on-chain
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Cleanup Hot        │ Remove from hot storage, emit event
│  Storage            │
└─────────────────────┘
```

### Archival Configuration

Default archival settings (can be customized via governance):

```go
type ArchivalConfig struct {
    Enabled                      bool   // true
    AutoArchive                  bool   // true
    ArchivalCheckIntervalBlocks  int64  // 100 (~10 min)
    DefaultArchiveTier           string // "standard"
    MaxArchivesPerBlock          uint32 // 10
    MinAgeForArchivalBlocks      int64  // 12960 (~90 days)
    MinAgeForArchivalSeconds     int64  // 90 * 24 * 60 * 60
    RestoreTTLSeconds            int64  // 24 * 60 * 60 (24 hours)
    EnableIntegrityChecks        bool   // true
    IntegrityCheckIntervalBlocks int64  // 14400 (~daily)
}
```

## Storage Tiers

### Standard Archive
- **Retrieval Time**: Minutes
- **Cost**: Moderate
- **Use Case**: Occasionally accessed historical data

### Glacier Archive
- **Retrieval Time**: 1-5 hours
- **Cost**: Low
- **Use Case**: Rarely accessed compliance data

### Deep Archive
- **Retrieval Time**: 12-48 hours
- **Cost**: Very low
- **Use Case**: Long-term regulatory retention

### Local Archive
- **Retrieval Time**: Immediate
- **Cost**: Hardware dependent
- **Use Case**: Development, testing, on-premise deployments

## Data Retrieval

### Restoration Process

```
┌─────────────────────┐
│  Restore Request    │ User requests archived artifact
│  - Archive ID       │
│  - Restore tier     │ expedited/standard/bulk
│  - Duration         │ How long to keep restored (default: 24h)
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Authorization      │ Verify user owns artifact or has access rights
│  Check              │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Initiate Restore   │ Backend initiates restoration
│  (Async)            │ May take minutes to hours depending on tier
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Poll Status        │ Client polls restore status
│                     │ Progress: 0-100%
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Retrieve Data      │ Once restored, data available for TTL period
│  (Temporary)        │ Default: 24 hours
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Auto-Cleanup       │ After TTL expires, remove from hot storage
└─────────────────────┘
```

### Restore Command Examples

```bash
# Restore an archive with standard tier
virtengine tx veid restore-archive \
    --archive-id "abc123..." \
    --restore-tier standard \
    --duration 86400 \
    --from mykey

# Check restore status
virtengine query veid archive-status abc123...

# Retrieve restored data
virtengine query veid get-archived-artifact abc123...
```

## Compliance Features

### Regulatory Compliance

#### GDPR (General Data Protection Regulation)
- **Right to Erasure**: Archives can be deleted on user request
  ```go
  ComplianceFlags: &ArchiveComplianceFlags{
      RegulationTags: []string{"GDPR"},
  }
  ```
- **Data Portability**: Archived data can be exported
- **Purpose Limitation**: Tracked via metadata

#### HIPAA (Health Insurance Portability and Accountability Act)
- **6-Year Minimum Retention**: Enforced for health-related data
  ```go
  ComplianceFlags: &ArchiveComplianceFlags{
      ComplianceRetentionMinDays: 2190, // 6 years
      RegulationTags: []string{"HIPAA"},
  }
  ```
- **Encryption Required**: All archives encrypted with AES-256-GCM
- **Access Logging**: All retrievals logged for audit trail

#### CCPA (California Consumer Privacy Act)
- **Consumer Data Deletion**: Right to delete personal information
- **Disclosure Tracking**: Access patterns tracked
- **Opt-Out Mechanisms**: Users can revoke consent

### Legal Hold

Archives can be placed under legal hold to prevent deletion:

```go
ComplianceFlags: &ArchiveComplianceFlags{
    LegalHold:       true,
    LegalHoldReason: "Pending litigation case #12345",
    LegalHoldSetAt:  &time.Now(),
}
```

Legal hold prevents deletion even if retention policy has expired.

### Compliance Validation

Automatic compliance checks run periodically:

```bash
# Check compliance status for an account
virtengine query veid compliance-audit <account-address>
```

Output:
```json
{
  "rules_version": 1,
  "embedding_envelopes": {
    "total": 50,
    "by_type": {"face_embedding": 25, "document_hash": 25},
    "revoked": 5,
    "expired": 3
  },
  "verification_records": {"total": 100},
  "compliance": {
    "raw_biometrics_on_chain": false,
    "derived_features_only": true
  }
}
```

## Archive Encryption

### Encryption Architecture

Archives use a separate encryption layer from hot storage to support:
- **Key Rotation**: Archive keys can be rotated without re-encrypting hot storage
- **Compliance**: Different encryption policies for different retention periods
- **Recovery**: Independent recovery mechanisms

### Encryption Format

```go
type ArchiveEncryptionEnvelope struct {
    Version              string    // Envelope version
    EncryptionAlgorithm  string    // "AES-256-GCM"
    ArchiveKeyID         string    // Key ID for rotation
    Ciphertext           []byte    // Encrypted data
    Nonce                []byte    // Unique nonce
    AuthTag              []byte    // Authentication tag
    Timestamp            time.Time // Encryption timestamp
    OriginalHash         []byte    // SHA-256 of plaintext
}
```

### Key Management

- **Archive Key Hierarchy**: Separate from hot storage keys
- **Key Rotation**: Supported via `ArchiveKeyID`
- **Key Storage**: Hardware Security Modules (HSM) recommended for production
- **Key Derivation**: BIP-39/BIP-44 compatible for validator keys

## Integrity Verification

### Checksum Verification

Every archive includes:
1. **Content Hash**: SHA-256 of original unencrypted data
2. **Integrity Checksum**: SHA-256 of encrypted archive data
3. **Merkle Root**: For chunked archives

### Periodic Integrity Checks

Automatic integrity verification runs daily:

```bash
# Verify integrity of an archive
virtengine query veid verify-archive-integrity <archive-id>
```

### Corruption Detection

If corruption is detected:
1. Alert emitted via events
2. Admin notification sent
3. Restore from redundant copy (if configured)
4. Mark archive as failed in index

## Monitoring & Metrics

### Archive Metrics

Query archival system metrics:

```bash
virtengine query veid archive-metrics
```

Output:
```json
{
  "total_archives": 1000,
  "total_bytes": 1073741824,
  "archives_by_tier": {
    "standard": 300,
    "glacier": 500,
    "deep_archive": 200
  },
  "archives_by_status": {
    "archived": 950,
    "restoring": 10,
    "restored": 40
  },
  "expired_archives": 50,
  "legal_hold_archives": 5,
  "last_purge_time": "2026-01-30T00:00:00Z",
  "last_integrity_check_time": "2026-01-30T02:00:00Z"
}
```

### Operational Metrics

Key metrics for operators:

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| Archival Rate | Archives created per hour | > 1000/hr |
| Restore Queue | Pending restore requests | > 100 |
| Failed Archives | Archives that failed integrity checks | > 10 |
| Storage Usage | Total bytes in archives | > 1 TB |
| Legal Holds | Archives under legal hold | Monitor only |

## Cost Optimization

### Storage Cost Tiers

Approximate costs (AWS S3 pricing as reference):

| Tier | Storage Cost | Retrieval Cost | Best For |
|------|-------------|----------------|----------|
| Standard | $0.023/GB/mo | $0.01/GB | Frequent access (< 90 days) |
| Glacier | $0.004/GB/mo | $0.10/GB | Rare access (1-2x/year) |
| Deep Archive | $0.001/GB/mo | $0.20/GB | Compliance only (never accessed) |

### Cost Optimization Strategies

1. **Aggressive Archival**: Archive after 30 days instead of 90 days
2. **Tier Promotion**: Move to deeper tiers after 1 year
3. **Lifecycle Policies**: Automatically transition between tiers
4. **Compression**: Enable compression for large artifacts
5. **Deduplication**: Avoid storing duplicate artifacts

## Operations Guide

### Enable Archival

Archival is enabled by default. To disable:

```bash
virtengine tx gov submit-proposal param-change \
    --title "Disable Archival" \
    --description "Temporarily disable archival" \
    --changes '[{"subspace":"veid","key":"ArchivalConfig","value":{"enabled":false}}]' \
    --from mykey
```

### Manual Archival

Archive a specific artifact manually:

```bash
virtengine tx veid archive-artifact \
    --content-hash "abc123..." \
    --archive-tier glacier \
    --from mykey
```

### Purge Expired Archives

Manually trigger purge of expired archives:

```bash
virtengine tx veid purge-expired-archives --from mykey
```

### Configure Backend

Set archive backend (requires governance):

```bash
# Configure S3 Glacier backend
virtengine tx gov submit-proposal param-change \
    --title "Configure S3 Archive Backend" \
    --description "Use AWS S3 Glacier for archival" \
    --changes '[{
        "subspace": "veid",
        "key": "ArchiveBackend",
        "value": {
            "backend_type": "s3",
            "bucket": "virtengine-archives",
            "region": "us-east-1"
        }
    }]' \
    --from mykey
```

## Disaster Recovery

### Backup Strategy

1. **Redundant Archives**: Store archives in multiple regions/providers
2. **Metadata Backup**: On-chain metadata is backed up with blockchain state
3. **Key Backup**: Archive encryption keys backed up in HSM or secure vault

### Recovery Procedures

#### Archive Corruption

```bash
# 1. Detect corruption
virtengine query veid verify-archive-integrity <archive-id>

# 2. Restore from backup
virtengine tx veid restore-from-backup \
    --archive-id <archive-id> \
    --backup-source <backup-location> \
    --from mykey

# 3. Verify integrity
virtengine query veid verify-archive-integrity <archive-id>
```

#### Lost Archive Keys

```bash
# Recover key from HSM backup
virtengine keys recover-archive-key \
    --key-id <archive-key-id> \
    --recovery-source hsm
```

## Security Considerations

### Access Control

- **Owner-Only**: Only artifact owner can restore by default
- **Delegation**: Owners can delegate access to specific addresses
- **Rate Limiting**: Prevent abuse with per-account quotas
- **Audit Logging**: All restore operations logged

### Threat Mitigations

| Threat | Mitigation |
|--------|------------|
| Unauthorized Access | Strong access control, audit logging |
| Data Leakage | End-to-end encryption, secure key management |
| Archive Tampering | Integrity checksums, immutable on-chain index |
| Key Compromise | Key rotation, HSM storage |
| Compliance Violation | Automated compliance checks, legal holds |

## Best Practices

### For Validators

1. **Configure Cold Storage**: Set up S3 Glacier or equivalent
2. **Enable Monitoring**: Track archival metrics and alerts
3. **Test Restoration**: Periodically test restore workflow
4. **Backup Keys**: Secure backup of archive encryption keys
5. **Compliance Audit**: Regular compliance audits

### For Users

1. **Understand Retention**: Know how long your data is retained
2. **Request Deletion**: Exercise right to erasure if needed
3. **Export Data**: Request data export before archival
4. **Monitor Access**: Review access logs periodically

### For Developers

1. **Handle Async Restore**: Restoration is asynchronous
2. **Implement Polling**: Poll restore status before retrieval
3. **Cache Restored Data**: Avoid repeated restore requests
4. **Respect TTL**: Re-archive after restore TTL expires

## Troubleshooting

### Common Issues

#### Archive Creation Failed

```
Error: archive creation failed: encryption error
```

**Solution**: Verify archive encryption key is configured correctly

#### Restore Timeout

```
Error: restore request timed out after 1 hour
```

**Solution**: Glacier tier restores can take 1-5 hours. Use expedited tier for faster restore.

#### Integrity Check Failed

```
Error: integrity verification failed: checksum mismatch
```

**Solution**: Archive may be corrupted. Restore from backup or contact support.

#### Legal Hold Prevents Deletion

```
Error: cannot delete archive: legal hold active
```

**Solution**: Remove legal hold via governance or wait for legal hold to be lifted.

## References

- [GDPR Documentation](https://gdpr.eu/)
- [HIPAA Compliance Guide](https://www.hhs.gov/hipaa/index.html)
- [CCPA Overview](https://oag.ca.gov/privacy/ccpa)
- [AWS S3 Glacier Pricing](https://aws.amazon.com/s3/pricing/)
- [Azure Archive Storage](https://azure.microsoft.com/en-us/pricing/details/storage/blobs/)

## Changelog

- **2026-01-30**: Initial implementation (DATA-001)
- Support for S3 Glacier, Azure Archive, Local filesystem
- GDPR, HIPAA, CCPA compliance features
- Automated archival with configurable eligibility criteria
- Legal hold and compliance retention policies
