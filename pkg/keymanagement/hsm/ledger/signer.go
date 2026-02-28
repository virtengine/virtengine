// Package ledger provides an hsm.HSMProvider implementation for Ledger
// hardware wallets.
package ledger

import (
	"context"
	"crypto"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	sdksecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
	vecrypto "github.com/virtengine/virtengine/x/encryption/crypto"
)

// DefaultDerivationPath is the Cosmos default HD derivation path.
const DefaultDerivationPath = "m/44'/118'/0'/0/0"

// DefaultHRP is the VirtEngine Bech32 human-readable prefix.
const DefaultHRP = "ve"

type ledgerWalletClient interface {
	Connect(ctx context.Context) error
	Disconnect() error
	GetPublicKey(ctx context.Context, hdPath string) ([]byte, error)
	SignTransaction(ctx context.Context, hdPath string, message []byte) (*vecrypto.LedgerSignature, error)
}

var newLedgerWalletClient = func(config *vecrypto.LedgerWalletConfig) ledgerWalletClient {
	return vecrypto.NewLedgerWallet(config)
}

// Signer implements hsm.HSMProvider for Ledger devices.
type Signer struct {
	config    hsm.LedgerConfig
	logger    *slog.Logger
	mu        sync.Mutex
	connected bool
	wallet    ledgerWalletClient
	keys      map[string]*ledgerKey
}

type ledgerKey struct {
	info      *hsm.KeyInfo
	publicKey []byte
}

// NewSigner creates a new Ledger signer.
func NewSigner(config hsm.LedgerConfig, logger *slog.Logger) (*Signer, error) {
	if config.DerivationPath == "" {
		config.DerivationPath = DefaultDerivationPath
	}
	if config.HRP == "" {
		config.HRP = DefaultHRP
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Signer{
		config: config,
		logger: logger,
		keys:   make(map[string]*ledgerKey),
	}, nil
}

// Connect connects to the Ledger device.
func (s *Signer) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.wallet == nil {
		cfg := vecrypto.DefaultLedgerWalletConfig()
		cfg.DefaultHDPath = s.config.DerivationPath
		cfg.HRPPrefix = s.config.HRP
		s.wallet = newLedgerWalletClient(cfg)
	}

	if err := s.wallet.Connect(ctx); err != nil {
		return err
	}

	s.connected = true
	s.logger.Info("Ledger device connected",
		slog.String("derivation_path", s.config.DerivationPath),
		slog.String("hrp", s.config.HRP),
	)
	return nil
}

// Close disconnects from the Ledger device.
func (s *Signer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.wallet != nil {
		if err := s.wallet.Disconnect(); err != nil {
			return err
		}
	}

	s.keys = make(map[string]*ledgerKey)
	s.connected = false
	return nil
}

// GenerateKey derives the configured Ledger secp256k1 public key and registers
// it under the requested label without exporting private material.
func (s *Signer) GenerateKey(ctx context.Context, keyType hsm.KeyType, label string) (*hsm.KeyInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return nil, hsm.ErrNotConnected
	}
	if _, ok := s.keys[label]; ok {
		return nil, hsm.ErrKeyExists
	}
	if keyType != hsm.KeyTypeSecp256k1 {
		return nil, fmt.Errorf("%w: %s", hsm.ErrUnsupportedKeyType, keyType)
	}

	pub, err := s.wallet.GetPublicKey(ctx, s.config.DerivationPath)
	if err != nil {
		return nil, fmt.Errorf("ledger: public key lookup failed: %w", err)
	}

	h := sha256.Sum256(pub)
	fp := hex.EncodeToString(h[:])
	info := &hsm.KeyInfo{
		Label:       label,
		ID:          []byte(fp[:8]),
		Type:        keyType,
		Size:        len(pub) * 8,
		Extractable: false,
		CreatedAt:   time.Now().UTC(),
		Fingerprint: fp,
	}

	s.keys[label] = &ledgerKey{
		info:      info,
		publicKey: cloneBytes(pub),
	}

	return info, nil
}

// ImportKey is not supported on Ledger devices.
func (s *Signer) ImportKey(_ context.Context, _ hsm.KeyType, _ string, _ []byte) (*hsm.KeyInfo, error) {
	return nil, fmt.Errorf("ledger: import key not supported on Ledger devices")
}

// GetKey retrieves key metadata.
func (s *Signer) GetKey(_ context.Context, label string) (*hsm.KeyInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.connected {
		return nil, hsm.ErrNotConnected
	}
	k, ok := s.keys[label]
	if !ok {
		return nil, hsm.ErrKeyNotFound
	}
	return k.info, nil
}

// ListKeys returns all keys.
func (s *Signer) ListKeys(_ context.Context) ([]*hsm.KeyInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.connected {
		return nil, hsm.ErrNotConnected
	}
	out := make([]*hsm.KeyInfo, 0, len(s.keys))
	for _, k := range s.keys {
		out = append(out, k.info)
	}
	return out, nil
}

// DeleteKey removes a key label mapping.
func (s *Signer) DeleteKey(_ context.Context, label string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.connected {
		return hsm.ErrNotConnected
	}
	if _, ok := s.keys[label]; !ok {
		return hsm.ErrKeyNotFound
	}
	delete(s.keys, label)
	return nil
}

// Sign signs data using the Ledger device.
func (s *Signer) Sign(ctx context.Context, label string, data []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.connected {
		return nil, hsm.ErrNotConnected
	}
	if _, ok := s.keys[label]; !ok {
		return nil, hsm.ErrKeyNotFound
	}

	signature, err := s.wallet.SignTransaction(ctx, s.config.DerivationPath, data)
	if err != nil {
		return nil, fmt.Errorf("ledger: sign failed: %w", err)
	}

	return cloneBytes(signature.Signature), nil
}

// GetPublicKey returns the public key for a label.
func (s *Signer) GetPublicKey(_ context.Context, label string) (crypto.PublicKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.connected {
		return nil, hsm.ErrNotConnected
	}
	k, ok := s.keys[label]
	if !ok {
		return nil, hsm.ErrKeyNotFound
	}

	return &sdksecp256k1.PubKey{Key: cloneBytes(k.publicKey)}, nil
}

func cloneBytes(data []byte) []byte {
	if data == nil {
		return nil
	}

	cloned := make([]byte, len(data))
	copy(cloned, data)
	return cloned
}

var _ hsm.HSMProvider = (*Signer)(nil)
