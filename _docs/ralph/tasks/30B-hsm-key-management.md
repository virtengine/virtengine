# Task 30B: HSM Integration for Production Key Management

**vibe-kanban ID:** `4e4ff48d-7cd9-440d-b7e1-e993ffeea276`

## Problem Statement

Production validators and providers MUST NOT store signing keys in software for mainnet. Current key management uses file-based keyring which is unsuitable for:
- Validator consensus keys (single point of failure)
- Provider signing keys (compromises provider reputation)
- Encryption keys (compromises data confidentiality)

HSM integration provides:
- Hardware-backed key protection
- Non-exportable keys
- Audit logging
- Compliance requirements (SOC2, PCI-DSS)

## Acceptance Criteria

### AC-1: PKCS#11 Provider
- [ ] PKCS#11 wrapper implementation
- [ ] Session management (login/logout)
- [ ] Key generation (Ed25519, Secp256k1)
- [ ] Key import (existing keys)
- [ ] Signing operations
- [ ] Works with SoftHSM2 (testing)

### AC-2: Cloud HSM Adapters
- [ ] AWS CloudHSM client
- [ ] GCP Cloud HSM client
- [ ] Azure Dedicated HSM client
- [ ] Configuration and auto-discovery

### AC-3: Ledger Integration
- [ ] Ledger Nano S/X support
- [ ] Cosmos derivation paths
- [ ] Transaction signing
- [ ] Key display/verification

### AC-4: Provider Daemon Integration
- [ ] HSM key support in key_manager.go
- [ ] Remote signing for provider operations
- [ ] Graceful fallback handling
- [ ] Key backup/recovery tooling

### AC-5: Validator Integration
- [ ] TMKms signer compatibility
- [ ] Cosmos Priv Validator interface
- [ ] Key migration from file to HSM

### AC-6: CLI Tooling
- [ ] `virtengine hsm init` command
- [ ] `virtengine hsm keygen` command
- [ ] `virtengine hsm migrate` command
- [ ] `virtengine hsm status` command
- [ ] `virtengine hsm backup` command

## Technical Requirements

### Package Structure

```
pkg/keymanagement/
├── hsm/
│   ├── doc.go
│   ├── interfaces.go           # HSMProvider, Signer interfaces
│   ├── manager.go              # HSMManager orchestrator
│   ├── config.go               # Configuration types
│   ├── pkcs11/
│   │   ├── provider.go         # PKCS#11 provider implementation
│   │   ├── session.go          # Session management
│   │   ├── key.go              # Key operations
│   │   ├── sign.go             # Signing operations
│   │   └── provider_test.go
│   ├── cloud/
│   │   ├── aws_cloudhsm.go     # AWS CloudHSM via PKCS#11
│   │   ├── gcp_cloudhsm.go     # GCP Cloud HSM
│   │   ├── azure_hsm.go        # Azure Dedicated HSM
│   │   └── cloud_test.go
│   ├── ledger/
│   │   ├── signer.go           # Ledger signer implementation
│   │   ├── transport.go        # USB HID transport
│   │   └── ledger_test.go
│   └── testutil/
│       └── softhsm.go          # SoftHSM2 test helper
├── keyring.go                  # Keyring abstraction
└── keyring_hsm.go              # HSM-backed keyring
```

### HSM Provider Interface

```go
// pkg/keymanagement/hsm/interfaces.go
package hsm

import (
    "context"
    "crypto"
)

// KeyType represents the cryptographic algorithm
type KeyType string

const (
    KeyTypeEd25519   KeyType = "ed25519"
    KeyTypeSecp256k1 KeyType = "secp256k1"
    KeyTypeX25519    KeyType = "x25519"
)

// KeyInfo contains metadata about a key
type KeyInfo struct {
    Label       string
    ID          []byte
    Type        KeyType
    Size        int
    Extractable bool
    CreatedAt   time.Time
}

// HSMProvider is the interface for HSM backends
type HSMProvider interface {
    // Connect establishes connection to the HSM
    Connect(ctx context.Context) error
    
    // Close releases HSM resources
    Close() error
    
    // GenerateKey creates a new key in the HSM
    GenerateKey(ctx context.Context, keyType KeyType, label string) (*KeyInfo, error)
    
    // ImportKey imports an existing key into the HSM
    ImportKey(ctx context.Context, keyType KeyType, label string, key []byte) (*KeyInfo, error)
    
    // GetKey retrieves key info by label
    GetKey(ctx context.Context, label string) (*KeyInfo, error)
    
    // ListKeys returns all keys in the HSM
    ListKeys(ctx context.Context) ([]*KeyInfo, error)
    
    // DeleteKey removes a key from the HSM
    DeleteKey(ctx context.Context, label string) error
    
    // Sign signs data using the specified key
    Sign(ctx context.Context, label string, data []byte) ([]byte, error)
    
    // GetPublicKey returns the public key for a key pair
    GetPublicKey(ctx context.Context, label string) (crypto.PublicKey, error)
}

// Signer provides signing operations for a specific key
type Signer interface {
    crypto.Signer
    Label() string
    KeyInfo() *KeyInfo
}
```

### PKCS#11 Implementation

```go
// pkg/keymanagement/hsm/pkcs11/provider.go
package pkcs11

import (
    "context"
    "fmt"
    "sync"

    "github.com/miekg/pkcs11"
    "github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

// Config holds PKCS#11 configuration
type Config struct {
    LibraryPath string `yaml:"library" json:"library"`
    SlotID      uint   `yaml:"slot" json:"slot"`
    PIN         string `yaml:"pin" json:"pin"`
    TokenLabel  string `yaml:"token_label" json:"token_label"`
}

// Provider implements HSMProvider using PKCS#11
type Provider struct {
    config  Config
    ctx     *pkcs11.Ctx
    session pkcs11.SessionHandle
    mu      sync.Mutex
    logger  *slog.Logger
}

// NewProvider creates a new PKCS#11 provider
func NewProvider(config Config, logger *slog.Logger) (*Provider, error) {
    if config.LibraryPath == "" {
        return nil, fmt.Errorf("pkcs11 library path required")
    }
    
    return &Provider{
        config: config,
        logger: logger,
    }, nil
}

// Connect initializes the PKCS#11 library and opens a session
func (p *Provider) Connect(ctx context.Context) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // Initialize library
    p.ctx = pkcs11.New(p.config.LibraryPath)
    if err := p.ctx.Initialize(); err != nil {
        return fmt.Errorf("failed to initialize pkcs11: %w", err)
    }
    
    // Find slot
    slots, err := p.ctx.GetSlotList(true)
    if err != nil {
        return fmt.Errorf("failed to get slots: %w", err)
    }
    
    if int(p.config.SlotID) >= len(slots) {
        return fmt.Errorf("slot %d not found", p.config.SlotID)
    }
    
    // Open session
    session, err := p.ctx.OpenSession(
        slots[p.config.SlotID],
        pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION,
    )
    if err != nil {
        return fmt.Errorf("failed to open session: %w", err)
    }
    p.session = session
    
    // Login
    if p.config.PIN != "" {
        if err := p.ctx.Login(session, pkcs11.CKU_USER, p.config.PIN); err != nil {
            p.ctx.CloseSession(session)
            return fmt.Errorf("failed to login: %w", err)
        }
    }
    
    p.logger.Info("connected to HSM",
        slog.Uint64("slot", uint64(p.config.SlotID)),
        slog.String("library", p.config.LibraryPath),
    )
    
    return nil
}

// GenerateKey creates a new key pair in the HSM
func (p *Provider) GenerateKey(
    ctx context.Context,
    keyType hsm.KeyType,
    label string,
) (*hsm.KeyInfo, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    var pubTemplate, privTemplate []*pkcs11.Attribute
    var mechanism *pkcs11.Mechanism
    
    switch keyType {
    case hsm.KeyTypeEd25519:
        mechanism = pkcs11.NewMechanism(pkcs11.CKM_EC_EDWARDS_KEY_PAIR_GEN, nil)
        pubTemplate = []*pkcs11.Attribute{
            pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
            pkcs11.NewAttribute(pkcs11.CKA_LABEL, label),
            pkcs11.NewAttribute(pkcs11.CKA_VERIFY, true),
            pkcs11.NewAttribute(pkcs11.CKA_EC_PARAMS, ed25519OID),
        }
        privTemplate = []*pkcs11.Attribute{
            pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
            pkcs11.NewAttribute(pkcs11.CKA_LABEL, label),
            pkcs11.NewAttribute(pkcs11.CKA_SIGN, true),
            pkcs11.NewAttribute(pkcs11.CKA_SENSITIVE, true),
            pkcs11.NewAttribute(pkcs11.CKA_EXTRACTABLE, false),
        }
        
    case hsm.KeyTypeSecp256k1:
        mechanism = pkcs11.NewMechanism(pkcs11.CKM_EC_KEY_PAIR_GEN, nil)
        pubTemplate = []*pkcs11.Attribute{
            pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
            pkcs11.NewAttribute(pkcs11.CKA_LABEL, label),
            pkcs11.NewAttribute(pkcs11.CKA_VERIFY, true),
            pkcs11.NewAttribute(pkcs11.CKA_EC_PARAMS, secp256k1OID),
        }
        privTemplate = []*pkcs11.Attribute{
            pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
            pkcs11.NewAttribute(pkcs11.CKA_LABEL, label),
            pkcs11.NewAttribute(pkcs11.CKA_SIGN, true),
            pkcs11.NewAttribute(pkcs11.CKA_SENSITIVE, true),
            pkcs11.NewAttribute(pkcs11.CKA_EXTRACTABLE, false),
        }
        
    default:
        return nil, fmt.Errorf("unsupported key type: %s", keyType)
    }
    
    pubHandle, privHandle, err := p.ctx.GenerateKeyPair(
        p.session, []*pkcs11.Mechanism{mechanism},
        pubTemplate, privTemplate,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to generate key pair: %w", err)
    }
    
    p.logger.Info("generated key pair",
        slog.String("label", label),
        slog.String("type", string(keyType)),
    )
    
    return &hsm.KeyInfo{
        Label:       label,
        ID:          uint64ToBytes(uint64(pubHandle)),
        Type:        keyType,
        Extractable: false,
        CreatedAt:   time.Now(),
    }, nil
}

// Sign signs data using the specified key
func (p *Provider) Sign(
    ctx context.Context,
    label string,
    data []byte,
) ([]byte, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // Find private key
    keyHandle, err := p.findKey(label, pkcs11.CKO_PRIVATE_KEY)
    if err != nil {
        return nil, err
    }
    
    // Determine mechanism from key type
    keyInfo, err := p.getKeyInfo(keyHandle)
    if err != nil {
        return nil, err
    }
    
    var mechanism *pkcs11.Mechanism
    switch keyInfo.Type {
    case hsm.KeyTypeEd25519:
        mechanism = pkcs11.NewMechanism(pkcs11.CKM_EDDSA, nil)
    case hsm.KeyTypeSecp256k1:
        mechanism = pkcs11.NewMechanism(pkcs11.CKM_ECDSA, nil)
    default:
        return nil, fmt.Errorf("unsupported key type for signing: %s", keyInfo.Type)
    }
    
    // Sign
    if err := p.ctx.SignInit(p.session, []*pkcs11.Mechanism{mechanism}, keyHandle); err != nil {
        return nil, fmt.Errorf("sign init failed: %w", err)
    }
    
    signature, err := p.ctx.Sign(p.session, data)
    if err != nil {
        return nil, fmt.Errorf("sign failed: %w", err)
    }
    
    p.logger.Debug("signed data",
        slog.String("label", label),
        slog.Int("data_len", len(data)),
        slog.Int("sig_len", len(signature)),
    )
    
    return signature, nil
}
```

### Ledger Signer

```go
// pkg/keymanagement/hsm/ledger/signer.go
package ledger

import (
    "context"
    "fmt"

    ledger "github.com/cosmos/ledger-cosmos-go"
    "github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

// Config holds Ledger configuration
type Config struct {
    DerivationPath string `yaml:"derivation_path" json:"derivation_path"`
    HRP            string `yaml:"hrp" json:"hrp"` // Human-readable prefix
}

// Signer implements HSMProvider for Ledger devices
type Signer struct {
    config Config
    device ledger.LedgerCosmos
    logger *slog.Logger
}

// NewSigner creates a new Ledger signer
func NewSigner(config Config, logger *slog.Logger) (*Signer, error) {
    if config.DerivationPath == "" {
        config.DerivationPath = "m/44'/118'/0'/0/0" // Cosmos default
    }
    if config.HRP == "" {
        config.HRP = "ve" // VirtEngine prefix
    }
    
    return &Signer{
        config: config,
        logger: logger,
    }, nil
}

// Connect establishes connection to the Ledger device
func (s *Signer) Connect(ctx context.Context) error {
    device, err := ledger.NewLedgerCosmos()
    if err != nil {
        return fmt.Errorf("failed to connect to ledger: %w", err)
    }
    
    s.device = device
    
    // Verify app is open
    version, err := device.GetVersion()
    if err != nil {
        return fmt.Errorf("failed to get ledger version: %w", err)
    }
    
    s.logger.Info("connected to ledger",
        slog.String("app", version.AppName),
        slog.Int("major", int(version.Major)),
        slog.Int("minor", int(version.Minor)),
    )
    
    return nil
}

// GetPublicKey returns the public key for the configured derivation path
func (s *Signer) GetPublicKey(ctx context.Context, label string) (crypto.PublicKey, error) {
    path, err := parsePath(s.config.DerivationPath)
    if err != nil {
        return nil, err
    }
    
    pubKey, addr, err := s.device.GetAddressPubKeySECP256K1(path, s.config.HRP)
    if err != nil {
        return nil, fmt.Errorf("failed to get public key: %w", err)
    }
    
    s.logger.Debug("got public key from ledger",
        slog.String("address", addr),
        slog.String("path", s.config.DerivationPath),
    )
    
    return secp256k1.PubKey{Key: pubKey}, nil
}

// Sign signs data using the Ledger device (requires user confirmation)
func (s *Signer) Sign(ctx context.Context, label string, data []byte) ([]byte, error) {
    path, err := parsePath(s.config.DerivationPath)
    if err != nil {
        return nil, err
    }
    
    s.logger.Info("please confirm signing on ledger device")
    
    sig, err := s.device.SignSECP256K1(path, data, 0)
    if err != nil {
        return nil, fmt.Errorf("ledger signing failed: %w", err)
    }
    
    return sig, nil
}
```

### HSM-Backed Keyring

```go
// pkg/keymanagement/keyring_hsm.go
package keymanagement

import (
    "context"

    "github.com/cosmos/cosmos-sdk/crypto/keyring"
    "github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

// HSMKeyring wraps an HSM provider to implement the Cosmos SDK keyring interface
type HSMKeyring struct {
    provider hsm.HSMProvider
    backend  keyring.Backend
}

// NewHSMKeyring creates a keyring backed by an HSM
func NewHSMKeyring(provider hsm.HSMProvider) (*HSMKeyring, error) {
    return &HSMKeyring{
        provider: provider,
        backend:  "hsm",
    }, nil
}

// Key returns the public key and address for the given key name
func (k *HSMKeyring) Key(uid string) (*keyring.Record, error) {
    ctx := context.Background()
    
    pubKey, err := k.provider.GetPublicKey(ctx, uid)
    if err != nil {
        return nil, err
    }
    
    return keyring.NewLocalRecord(uid, pubKey, nil)
}

// Sign signs the given message using the HSM
func (k *HSMKeyring) Sign(uid string, msg []byte) ([]byte, crypto.PublicKey, error) {
    ctx := context.Background()
    
    sig, err := k.provider.Sign(ctx, uid, msg)
    if err != nil {
        return nil, nil, err
    }
    
    pubKey, err := k.provider.GetPublicKey(ctx, uid)
    if err != nil {
        return nil, nil, err
    }
    
    return sig, pubKey, nil
}
```

### CLI Commands

```go
// cmd/virtengine/cmd/hsm/init.go
package hsm

import (
    "github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize HSM configuration",
        Long: `Initialize HSM configuration for key management.
        
Supports:
- PKCS#11 backends (SoftHSM2, CloudHSM, Luna)
- Ledger hardware wallets
- Azure Managed HSM`,
        RunE: runInit,
    }
    
    cmd.Flags().String("backend", "pkcs11", "HSM backend (pkcs11, ledger, azure)")
    cmd.Flags().String("library", "", "PKCS#11 library path")
    cmd.Flags().Int("slot", 0, "PKCS#11 slot ID")
    cmd.Flags().String("config", "", "Output config file path")
    
    return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
    backend, _ := cmd.Flags().GetString("backend")
    
    switch backend {
    case "pkcs11":
        return initPKCS11(cmd)
    case "ledger":
        return initLedger(cmd)
    case "azure":
        return initAzure(cmd)
    default:
        return fmt.Errorf("unknown backend: %s", backend)
    }
}
```

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `pkg/keymanagement/hsm/interfaces.go` | Provider interfaces | 100 |
| `pkg/keymanagement/hsm/manager.go` | HSM manager | 200 |
| `pkg/keymanagement/hsm/config.go` | Configuration | 150 |
| `pkg/keymanagement/hsm/pkcs11/provider.go` | PKCS#11 provider | 500 |
| `pkg/keymanagement/hsm/pkcs11/session.go` | Session mgmt | 200 |
| `pkg/keymanagement/hsm/pkcs11/key.go` | Key operations | 300 |
| `pkg/keymanagement/hsm/pkcs11/sign.go` | Signing | 150 |
| `pkg/keymanagement/hsm/cloud/aws_cloudhsm.go` | AWS adapter | 300 |
| `pkg/keymanagement/hsm/cloud/gcp_cloudhsm.go` | GCP adapter | 250 |
| `pkg/keymanagement/hsm/cloud/azure_hsm.go` | Azure adapter | 250 |
| `pkg/keymanagement/hsm/ledger/signer.go` | Ledger signer | 400 |
| `pkg/keymanagement/keyring_hsm.go` | HSM keyring | 300 |
| `cmd/virtengine/cmd/hsm/*.go` | CLI commands | 500 |
| `*_test.go` | Test files | 800 |

**Total Estimated:** 4,400 lines

## Validation Checklist

- [ ] PKCS#11 works with SoftHSM2
- [ ] Key generation creates non-extractable keys
- [ ] Signing operations work
- [ ] AWS CloudHSM integration tested
- [ ] Ledger signing requires user confirmation
- [ ] CLI commands functional
- [ ] Provider daemon integration working
- [ ] Audit logging captures all operations
- [ ] PIN/passwords never logged
- [ ] Session cleanup on errors
- [ ] E2E tests with mock HSM

## Dependencies

- None (foundation component)

## Security Considerations

1. **PIN/Password Handling**
   - Never log PIN values
   - Clear from memory after use
   - Use secure input methods

2. **Session Management**
   - Timeout inactive sessions
   - Clean up on errors
   - One session per operation (optional)

3. **Key Operations**
   - All keys created as non-extractable
   - Audit log all operations
   - Rate limit signing operations

4. **Error Handling**
   - Don't leak HSM state in errors
   - Graceful degradation
   - Alert on failures
