# VirtEngine Key Management System

## Overview

The VirtEngine Key Management System provides comprehensive cryptographic key management for blockchain validators, providers, and users. This document covers all aspects of key management including HSM integration, key rotation, backup and recovery, multi-signature support, compromise detection, lifecycle management, and access controls.

## Table of Contents

1. [Architecture](#architecture)
2. [HSM Integration](#hsm-integration)
3. [Key Rotation Procedures](#key-rotation-procedures)
4. [Key Backup and Recovery](#key-backup-and-recovery)
5. [Multi-Signature Support](#multi-signature-support)
6. [Compromise Detection](#compromise-detection)
7. [Key Lifecycle Management](#key-lifecycle-management)
8. [Access Controls and Auditing](#access-controls-and-auditing)
9. [Security Best Practices](#security-best-practices)
10. [API Reference](#api-reference)

---

## Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       Key Management System                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │   HSM       │  │   Backup    │  │  MultiSig   │  │  Lifecycle  │    │
│  │  Provider   │  │  Manager    │  │  Manager    │  │  Manager    │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
│         │                │                │                │            │
│         └────────────────┴────────────────┴────────────────┘            │
│                                  │                                       │
│                    ┌─────────────┴─────────────┐                        │
│                    │       Key Manager         │                        │
│                    └─────────────┬─────────────┘                        │
│                                  │                                       │
│  ┌───────────────────────────────┼───────────────────────────────┐     │
│  │                               │                                │     │
│  │  ┌─────────────┐  ┌──────────┴──────────┐  ┌─────────────┐   │     │
│  │  │ Compromise  │  │   Access Controller  │  │   Audit     │   │     │
│  │  │  Detector   │  │                      │  │   Logger    │   │     │
│  │  └─────────────┘  └─────────────────────┘  └─────────────┘   │     │
│  │                         Security Layer                         │     │
│  └───────────────────────────────────────────────────────────────┘     │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Storage Types

The system supports multiple key storage backends:

| Storage Type | Use Case | Security Level |
|-------------|----------|----------------|
| Memory | Testing only | Low |
| File | Development | Medium |
| HSM | Production validators | High |
| Ledger | Non-custodial production | Very High |

### Supported Key Algorithms

| Algorithm | Type | Use Case |
|-----------|------|----------|
| Ed25519 | EdDSA | Validator signing, default |
| secp256k1 | ECDSA | Ethereum compatibility |
| P256 (NIST) | ECDSA | Enterprise compliance |
| RSA-2048/4096 | RSA | Legacy compatibility |

---

## HSM Integration

### Overview

Hardware Security Modules (HSMs) provide tamper-resistant key storage and cryptographic operations. The VirtEngine HSM integration uses PKCS#11 as the standard interface.

### Supported HSMs

| HSM Type | Description | Production Ready |
|----------|-------------|------------------|
| SoftHSM | Software-based HSM for development | No |
| YubiHSM | Hardware HSM from Yubico | Yes |
| CloudHSM | AWS/Azure/GCP cloud HSM services | Yes |
| Thales Luna | Enterprise HSM | Yes |
| TPM | Trusted Platform Module | Limited |

### Configuration

```go
// HSM Provider Configuration
type HSMProviderConfig struct {
    // Path to PKCS#11 library
    LibraryPath string
    
    // HSM slot number
    SlotID uint
    
    // PIN for authentication
    PIN string
    
    // Maximum concurrent sessions
    MaxSessions int
    
    // Enable session pooling
    SessionPooling bool
}
```

### Usage Example

```go
// Initialize HSM provider
provider := NewPKCS11Provider()
err := provider.Initialize(&HSMProviderConfig{
    LibraryPath: "/usr/lib/softhsm/libsofthsm2.so",
    SlotID:      0,
    PIN:         os.Getenv("HSM_PIN"),
    MaxSessions: 10,
})
if err != nil {
    log.Fatal("Failed to initialize HSM:", err)
}
defer provider.Close()

// Generate a key in HSM
handle, err := provider.GenerateKey("validator-key", HSMKeyTypeEd25519)
if err != nil {
    log.Fatal("Failed to generate key:", err)
}

// Sign with HSM-protected key
message := []byte("transaction data")
signature, err := provider.Sign(handle, message)
```

### SoftHSM Setup (Development)

For local development, use SoftHSM2:

```bash
# Install SoftHSM (Ubuntu/Debian)
sudo apt-get install softhsm2

# Initialize a token
softhsm2-util --init-token --slot 0 --label "VirtEngine" --pin 1234 --so-pin 5678

# Verify installation
softhsm2-util --show-slots
```

### Production HSM Deployment

1. **Install HSM Hardware**: Follow vendor-specific installation
2. **Install PKCS#11 Library**: Vendor provides library
3. **Initialize Tokens**: Create partitions/tokens for VirtEngine
4. **Configure Access**: Set up PINs and access policies
5. **Test Connectivity**: Verify PKCS#11 operations work

---

## Key Rotation Procedures

### Rotation Overview

Key rotation is the process of replacing an active key with a new key while maintaining service continuity. Regular rotation limits the impact of potential key compromise.

### Rotation Policy

```go
// Key Lifecycle Policy
type KeyLifecyclePolicy struct {
    Name                 string
    MaxActiveAgeDays     int  // Days before rotation required
    WarningDaysBeforeExp int  // Days before expiration warning
    ExpirationDays       int  // Days until key expires
    GracePeriodDays      int  // Days to keep old key after rotation
    AutoRotate           bool // Enable automatic rotation
    RequireMFA           bool // Require MFA for rotation
}
```

### Manual Rotation Procedure

```bash
# 1. Generate new key
virtengine keys add new-validator-key --keyring-backend=os

# 2. Register new key with network
virtengine tx staking create-validator \
    --pubkey=$(virtengine tendermint show-validator --home=/new-key-dir)

# 3. Update provider configuration
virtengine provider update --key=new-validator-key

# 4. Archive old key
virtengine keys migrate old-key --archive
```

### Automatic Rotation

The system supports automatic key rotation based on policy:

```go
// Configure automatic rotation
lm := NewKeyLifecycleManager(km)
lm.RegisterPolicy(&KeyLifecyclePolicy{
    Name:             "validator-policy",
    MaxActiveAgeDays: 365,
    AutoRotate:       true,
    RequireMFA:       true,
})

// Start rotation monitor
lm.StartRotationMonitor(ctx)
```

### Rotation States

```
Active ──> Rotating ──> Deactivated ──> Archived ──> Destroyed
            │
            └──> New key becomes Active
```

---

## Key Backup and Recovery

### Backup Methods

| Method | Description | Recovery Shares |
|--------|-------------|-----------------|
| Encrypted Backup | AES-256-GCM encrypted file | N/A |
| Shamir Secret Sharing | Split key into M-of-N shares | Configurable |
| Multi-location Backup | Distribute to multiple sites | N/A |

### Encrypted Backup

All backups use strong encryption:

- **Encryption**: AES-256-GCM
- **Key Derivation**: Argon2id (3 iterations, 64MB memory)
- **Integrity**: SHA-256 checksum

```go
// Create encrypted backup
bm := NewKeyBackupManager(nil, km)
backup, err := bm.CreateBackup("strong-passphrase")
if err != nil {
    log.Fatal(err)
}

// Save backup to secure location
backupJSON, _ := json.Marshal(backup)
// Store securely...
```

### Shamir Secret Sharing

For high-security scenarios, split the backup into shares:

```go
// Configure Shamir sharing (3-of-5)
config := &KeyBackupConfig{
    ShamirThreshold:    3,
    ShamirTotalShares: 5,
}
bm := NewKeyBackupManager(config, km)

// Create shares
shares, err := bm.CreateShamirShares("backup-passphrase")
if err != nil {
    log.Fatal(err)
}

// Distribute shares to different custodians
for i, share := range shares {
    distributeToLocation(i, share)
}

// Recovery requires 3 of 5 shares
recoveredShares := []RecoveryShare{shares[0], shares[2], shares[4]}
backup, err := bm.ReconstructFromShares(recoveredShares)
```

### Backup Verification

```go
// Verify backup integrity
valid, err := bm.VerifyBackup(backup, passphrase)
if !valid {
    log.Error("Backup integrity check failed:", err)
}
```

### Recovery Procedure

1. **Gather Required Materials**
   - Encrypted backup file or M-of-N Shamir shares
   - Backup passphrase
   - MFA device (if required)

2. **Verify Identity**
   - Authenticate with admin credentials
   - Complete MFA challenge

3. **Restore Keys**
   ```go
   restoredKeys, err := bm.RestoreBackup(backup, passphrase)
   if err != nil {
       log.Fatal("Recovery failed:", err)
   }
   ```

4. **Verify Restoration**
   - Test signing with restored keys
   - Verify public key fingerprints match records

---

## Multi-Signature Support

### Overview

Multi-signature (multisig) requires multiple parties to authorize critical operations. This provides defense against single-party compromise.

### Configuration

```go
// Create 2-of-3 multisig
config := &MultiSigConfig{
    Threshold:       2,           // Signatures required
    TotalSigners:    3,           // Total authorized signers
    TimeoutDuration: 24 * time.Hour,
    AllowPartialSig: false,
}

signers := []AuthorizedSigner{
    {PublicKey: "pubkey1", Label: "CEO", Weight: 1},
    {PublicKey: "pubkey2", Label: "CTO", Weight: 1},
    {PublicKey: "pubkey3", Label: "Security", Weight: 1},
}

msm := NewMultiSigManager(km)
key, err := msm.CreateMultiSigKey(config, signers, "Treasury Key")
```

### Weighted Multi-Signature

Weights allow different signers to have different voting power:

```go
// 3-of-5 weight with CEO having 2x weight
signers := []AuthorizedSigner{
    {PublicKey: "ceo-key", Label: "CEO", Weight: 2},
    {PublicKey: "cto-key", Label: "CTO", Weight: 1},
    {PublicKey: "cfo-key", Label: "CFO", Weight: 1},
    {PublicKey: "legal-key", Label: "Legal", Weight: 1},
    {PublicKey: "security-key", Label: "Security", Weight: 1},
}
// Total weight: 6, Threshold: 3
// CEO alone = 2 (not enough)
// CEO + anyone = 3 (approved)
// Any 3 others = 3 (approved)
```

### Signing Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Multi-Signature Workflow                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   1. Initiate          2. Collect           3. Complete         │
│   ┌─────────┐         ┌─────────┐          ┌─────────┐         │
│   │ Create  │ ──────> │ Add     │ ──────>  │ Finalize│         │
│   │ Request │         │ Sigs    │          │ Multisig│         │
│   └─────────┘         └─────────┘          └─────────┘         │
│       │                   │                     │               │
│       v                   v                     v               │
│   Status:             Status:               Status:             │
│   PENDING             PENDING/              COMPLETE            │
│                       THRESHOLD_MET                              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Example: Transaction Signing

```go
// Initiate a multisig operation
op, err := msm.InitiateOperation(
    key.ID,
    transactionBytes,
    "ceo@company.com",
    "Treasury withdrawal of 1000 VE",
)

// Collect signatures from authorized signers
err = msm.AddSignature(op.ID, "ceo-key", ceoSignature)
err = msm.AddSignature(op.ID, "cto-key", ctoSignature)

// Check if threshold is met
op, _ = msm.GetOperation(op.ID)
if op.Status == MultiSigStatusThresholdMet {
    // Complete the operation
    finalOp, err := msm.CompleteOperation(op.ID)
    // Use finalOp.FinalSignature for the transaction
}
```

---

## Compromise Detection

### Detection Indicators

| Indicator | Description | Severity |
|-----------|-------------|----------|
| Rapid Usage | Unusual number of operations per minute | Medium |
| Anomalous Time | Operations outside normal hours | Low |
| Anomalous Location | Operations from unexpected IP | High |
| Failed Verifications | Multiple signature failures | High |
| External Report | Manual report of compromise | Critical |
| Key Leakage | Key material detected externally | Critical |

### Configuration

```go
config := &CompromiseDetectorConfig{
    UsageThresholdPerMinute: 10,
    UsageThresholdPerHour:   100,
    AnomalousTimeWindowStart: 6,   // 6 AM
    AnomalousTimeWindowEnd:   22,  // 10 PM
    FailedVerificationThreshold: 5,
    AutoRevokeOnCritical:    true,
}

detector := NewCompromiseDetector(config, km)
```

### Usage Monitoring

```go
// Record key usage (called automatically by KeyManager)
indicators := detector.RecordKeyUsage(keyID, clientIP, time.Now())

if len(indicators) > 0 {
    // Alert on detected anomalies
    for _, indicator := range indicators {
        alertSecurityTeam(keyID, indicator)
    }
}
```

### Manual Compromise Report

```go
// Report a suspected compromise
err := detector.ReportCompromise(
    keyID,
    IndicatorExternalReport,
    SeverityCritical,
    "Key found in public GitHub repository",
    "security-team",
)
```

### Automatic Response

When `AutoRevokeOnCritical` is enabled:

1. Key is immediately suspended
2. Alert is sent to administrators
3. All active sessions using the key are terminated
4. Audit log entry is created

---

## Key Lifecycle Management

### Key States

```
                    ┌───────────┐
                    │  Created  │
                    └─────┬─────┘
                          │ activate
                          v
      ┌───────────┐ ┌───────────┐
      │ Suspended │←│  Active   │←┐
      └─────┬─────┘ └─────┬─────┘ │
            │             │        │
            │ reactivate  │rotate  │
            └─────────────┘        │
                          │        │
                          v        │
                    ┌───────────┐  │
                    │ Rotating  │──┘
                    └─────┬─────┘
                          │ complete
                          v
                    ┌───────────┐
                    │Deactivated│
                    └─────┬─────┘
                          │ archive
                          v
                    ┌───────────┐
                    │ Archived  │
                    └─────┬─────┘
                          │ destroy
                          v
                    ┌───────────┐
                    │ Destroyed │
                    └───────────┘

Special States:
- Compromised: Key suspected/confirmed compromised
- Expired: Key past expiration date
```

### Lifecycle Policies

```go
// Default policy
policy := &KeyLifecyclePolicy{
    Name:                 "default",
    MaxActiveAgeDays:     365,    // Rotate annually
    WarningDaysBeforeExp: 30,     // Warn 30 days before
    ExpirationDays:       730,    // Expire after 2 years
    GracePeriodDays:      90,     // Keep old key 90 days
    AutoRotate:           false,
    RequireMFA:           true,
}

// High-security policy
hsPolicy := &KeyLifecyclePolicy{
    Name:                 "high-security",
    MaxActiveAgeDays:     90,     // Rotate quarterly
    WarningDaysBeforeExp: 14,
    ExpirationDays:       180,
    GracePeriodDays:      30,
    AutoRotate:           true,
    RequireMFA:           true,
}
```

### Lifecycle Reports

```go
lm := NewKeyLifecycleManager(km)
report := lm.GenerateLifecycleReport()

fmt.Printf("Total Keys: %d\n", report.TotalKeys)
fmt.Printf("Active: %d\n", report.ActiveKeys)
fmt.Printf("Expiring Soon: %d\n", report.ExpiringKeys)
fmt.Printf("Needs Rotation: %d\n", report.RotationNeeded)
```

---

## Access Controls and Auditing

### Permission Model

The system uses Role-Based Access Control (RBAC):

| Permission | Description |
|------------|-------------|
| `key:create` | Generate new keys |
| `key:read` | View key metadata |
| `key:sign` | Sign with keys |
| `key:rotate` | Rotate keys |
| `key:backup` | Create backups |
| `key:restore` | Restore from backup |
| `key:delete` | Delete keys |
| `hsm:access` | Access HSM operations |
| `multisig:initiate` | Start multisig operations |
| `multisig:sign` | Add signatures |
| `multisig:complete` | Complete operations |
| `audit:read` | View audit logs |
| `audit:export` | Export audit data |
| `policy:manage` | Manage policies |
| `admin:*` | All permissions |

### Default Roles

| Role | Permissions |
|------|-------------|
| admin | All permissions |
| operator | key:create, key:read, key:sign, key:rotate, hsm:access |
| auditor | key:read, audit:read, audit:export |
| signer | key:read, key:sign, multisig:sign |

### Session Management

```go
ac := NewAccessController(nil)

// Create a principal
principal := &Principal{
    ID:    "user-123",
    Type:  "user",
    Name:  "Alice Smith",
    Roles: []string{"operator"},
}
ac.CreatePrincipal(principal)

// Create session
session, err := ac.CreateSession("user-123", "192.168.1.100", "Mozilla/5.0...")

// Check permission before operation
err = ac.CheckPermission(session.ID, PermissionKeySign)
if err != nil {
    return errors.New("unauthorized")
}

// Perform operation...

// End session
ac.EndSession(session.ID)
```

### Audit Logging

All key operations are logged with:

- Timestamp (UTC)
- Session ID
- Principal ID and name
- Key ID
- Operation type
- Success/failure
- Error message (if failed)
- Additional metadata
- Hash chain link (integrity)

```go
// Configure audit logging
config := &AuditLogConfig{
    LogFile:        "/var/log/virtengine/keymanagement.audit.log",
    MaxFileSize:    100 * 1024 * 1024, // 100MB
    EnableChaining: true,
    SyncWrites:     true,
}

logger, _ := NewAuditLogger(config)

// Log an operation
logger.LogKeyOperation(
    AuditEventKeySigned,
    session.ID,
    principal.ID,
    principal.Name,
    keyID,
    "sign_transaction",
    true,
    "",
    map[string]interface{}{
        "tx_hash": txHash,
        "amount":  "1000uve",
    },
)
```

### Audit Reports

```go
// Generate compliance report
report := logger.GenerateAuditReport(time.Now().AddDate(0, -1, 0))

fmt.Printf("Period: %s to %s\n", report.StartTime, report.EndTime)
fmt.Printf("Total Events: %d\n", report.TotalEvents)
fmt.Printf("Success Rate: %.2f%%\n", 
    float64(report.SuccessCount)/float64(report.TotalEvents)*100)
```

### Hash Chain Verification

```go
// Verify audit log integrity
valid, errors := logger.VerifyIntegrity()
if !valid {
    for _, err := range errors {
        log.Error("Integrity violation:", err)
    }
    alertSecurityTeam("Audit log tampering detected")
}
```

---

## Security Best Practices

### Key Generation

- ✅ Use HSM for production validator keys
- ✅ Generate keys in secure environment
- ✅ Use Ed25519 or secp256k1 algorithms
- ✅ Store keys encrypted at rest
- ❌ Never generate keys on shared systems
- ❌ Never log key material

### Key Storage

- ✅ Use HSM or hardware wallet for high-value keys
- ✅ Encrypt file-based keys with strong passphrase
- ✅ Limit filesystem permissions (600)
- ✅ Use separate storage from application data
- ❌ Never store unencrypted keys
- ❌ Never commit keys to version control

### Key Access

- ✅ Implement principle of least privilege
- ✅ Use MFA for sensitive operations
- ✅ Rotate access credentials regularly
- ✅ Monitor and alert on anomalous access
- ❌ Never share keys between environments
- ❌ Never expose keys to client applications

### Key Rotation

- ✅ Rotate validator keys annually (minimum)
- ✅ Rotate immediately on suspected compromise
- ✅ Maintain audit trail of rotations
- ✅ Test rotation procedures regularly
- ❌ Never skip rotation schedules
- ❌ Never delete old keys before grace period

### Backup and Recovery

- ✅ Create encrypted backups regularly
- ✅ Use Shamir secret sharing for critical keys
- ✅ Store shares in geographically separate locations
- ✅ Test recovery procedures quarterly
- ❌ Never store backup passphrase with backup
- ❌ Never give all shares to single custodian

---

## API Reference

### KeyManager

```go
type KeyManager interface {
    // GenerateKey creates a new key pair
    GenerateKey(label string) (*ManagedKey, error)
    
    // Sign signs data with the active key
    Sign(data []byte) ([]byte, error)
    
    // Verify verifies a signature
    Verify(data, signature, publicKey []byte) (bool, error)
    
    // GetKey retrieves a key by ID
    GetKey(keyID string) (*ManagedKey, error)
    
    // ListKeys returns all managed keys
    ListKeys() ([]*ManagedKey, error)
    
    // RotateKey rotates to a new key
    RotateKey() (*ManagedKey, error)
}
```

### HSMProvider

```go
type HSMProvider interface {
    Initialize(config *HSMProviderConfig) error
    Close() error
    GenerateKey(label string, keyType HSMKeyType) (*HSMKeyHandle, error)
    Sign(handle *HSMKeyHandle, data []byte) ([]byte, error)
    Verify(handle *HSMKeyHandle, data, signature []byte) (bool, error)
    GetKey(label string) (*HSMKeyHandle, error)
    ListKeys() ([]*HSMKeyHandle, error)
    DeleteKey(label string) error
    ImportKey(label string, privateKey []byte, keyType HSMKeyType) (*HSMKeyHandle, error)
    ExportPublicKey(handle *HSMKeyHandle) ([]byte, error)
}
```

### KeyBackupManager

```go
type KeyBackupManager interface {
    CreateBackup(passphrase string) (*KeyBackup, error)
    RestoreBackup(backup *KeyBackup, passphrase string) ([]*ManagedKey, error)
    VerifyBackup(backup *KeyBackup, passphrase string) (bool, error)
    CreateShamirShares(passphrase string) ([]RecoveryShare, error)
    ReconstructFromShares(shares []RecoveryShare) (*KeyBackup, error)
}
```

### MultiSigManager

```go
type MultiSigManager interface {
    CreateMultiSigKey(config *MultiSigConfig, signers []AuthorizedSigner, description string) (*MultiSigKey, error)
    GetMultiSigKey(id string) (*MultiSigKey, error)
    InitiateOperation(keyID string, message []byte, initiator, description string) (*MultiSigOperation, error)
    AddSignature(opID, signerPublicKey string, signature []byte) error
    CompleteOperation(opID string) (*MultiSigOperation, error)
    CancelOperation(opID string, reason string) error
}
```

### CompromiseDetector

```go
type CompromiseDetector interface {
    RecordKeyUsage(keyID, clientIP string, timestamp time.Time) []CompromiseIndicator
    ReportCompromise(keyID string, indicator CompromiseIndicator, severity Severity, details, reportedBy string) error
    IsKeyCompromised(keyID string) bool
    GetEventsByKey(keyID string) []*CompromiseEvent
    AcknowledgeEvent(eventID, acknowledgedBy string) error
}
```

### KeyLifecycleManager

```go
type KeyLifecycleManager interface {
    RegisterKey(keyID, algorithm, fingerprint, policyName string) (*KeyLifecycleRecord, error)
    ActivateKey(keyID, activatedBy string) error
    RotateKey(oldKeyID, newKeyID, rotatedBy string) error
    CompleteRotation(oldKeyID, completedBy string) error
    SuspendKey(keyID, suspendedBy, reason string) error
    DeactivateKey(keyID, deactivatedBy string) error
    GetRecord(keyID string) (*KeyLifecycleRecord, error)
    GetKeysNeedingRotation() []*KeyLifecycleRecord
    GenerateLifecycleReport() *LifecycleReport
}
```

### AccessController

```go
type AccessController interface {
    CreatePrincipal(principal *Principal) error
    CreateSession(principalID, clientIP, userAgent string) (*Session, error)
    ValidateSession(sessionID string) (*Session, error)
    CheckPermission(sessionID string, permission Permission) error
    EndSession(sessionID string) error
}
```

### AuditLogger

```go
type AuditLogger interface {
    Log(event *AuditEvent) error
    LogKeyOperation(eventType AuditEventType, sessionID, principalID, principalName, keyID, operation string, success bool, errorMsg string, metadata map[string]interface{}) error
    GetEvents(since time.Time, eventType *AuditEventType) []*AuditEvent
    GetEventsByKey(keyID string, since time.Time) []*AuditEvent
    VerifyIntegrity() (bool, []string)
    GenerateAuditReport(since time.Time) *AuditReport
}
```

---

## Troubleshooting

### Common Issues

**HSM Connection Failed**
```
Error: failed to initialize PKCS#11: library not found
Solution: Verify PKCS#11 library path and permissions
```

**Key Generation Failed**
```
Error: HSM slot full
Solution: Delete unused keys or configure additional slots
```

**Backup Restoration Failed**
```
Error: backup decryption failed
Solution: Verify passphrase and backup file integrity
```

**Multi-Sig Timeout**
```
Error: operation timed out waiting for signatures
Solution: Increase timeout or contact missing signers
```

### Getting Help

- **Documentation**: [VirtEngine Docs](https://docs.virtengine.network)
- **GitHub Issues**: [Report bugs](https://github.com/virtengine/virtengine/issues)
- **Discord**: [Community support](https://discord.gg/virtengine)
- **Security Issues**: security@virtengine.network (PGP encrypted)

---

## Appendix: Compliance Notes

### SOC 2 Type II

The key management system supports SOC 2 compliance through:
- Comprehensive audit logging
- Access controls with least privilege
- Encryption of sensitive data
- Regular key rotation

### GDPR

For GDPR compliance:
- Keys used for personal data can be destroyed (right to be forgotten)
- Audit logs track all access to keys
- Access controls limit who can access encryption keys

### PCI DSS

For payment-related deployments:
- HSM support for key protection (Requirement 3.5)
- Key rotation procedures (Requirement 3.6)
- Access logging (Requirement 10)

---

*Document Version: 1.0.0*
*Last Updated: 2025-01-14*
*Maintainer: VirtEngine Security Team*
