// Package ledger provides an hsm.HSMProvider implementation for Ledger
// hardware wallets. In production this delegates to the Cosmos Ledger Go
// library; the current implementation provides a software mock suitable for
// unit tests and CI.
package ledger

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	crand "crypto/rand"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

// DefaultDerivationPath is the Cosmos default HD derivation path.
const DefaultDerivationPath = "m/44'/118'/0'/0/0"

// DefaultHRP is the VirtEngine Bech32 human-readable prefix.
const DefaultHRP = "ve"

// Signer implements hsm.HSMProvider for Ledger devices.
type Signer struct {
	config    hsm.LedgerConfig
	logger    *slog.Logger
	mu        sync.Mutex
	connected bool
	keys      map[string]*ledgerKey
}

type ledgerKey struct {
	info       *hsm.KeyInfo
	privateKey []byte
	publicKey  []byte
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
func (s *Signer) Connect(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	for _, k := range s.keys {
		scrubBytes(k.privateKey)
	}
	s.keys = make(map[string]*ledgerKey)
	s.connected = false
	return nil
}

// GenerateKey generates a new key on the Ledger. For real Ledger devices,
// keys are derived from the seed via the derivation path. This mock generates
// a random key.
func (s *Signer) GenerateKey(_ context.Context, keyType hsm.KeyType, label string) (*hsm.KeyInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return nil, hsm.ErrNotConnected
	}
	if _, ok := s.keys[label]; ok {
		return nil, hsm.ErrKeyExists
	}

	switch keyType {
	case hsm.KeyTypeEd25519, hsm.KeyTypeSecp256k1:
		// For mock, generate ed25519 regardless
	default:
		return nil, fmt.Errorf("%w: %s", hsm.ErrUnsupportedKeyType, keyType)
	}

	pub, priv, err := ed25519.GenerateKey(crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ledger: keygen: %w", err)
	}

	h := sha256.Sum256([]byte(pub))
	fp := hex.EncodeToString(h[:])

	info := &hsm.KeyInfo{
		Label:       label,
		ID:          []byte(fp[:8]),
		Type:        keyType,
		Size:        256,
		Extractable: false,
		CreatedAt:   time.Now().UTC(),
		Fingerprint: fp,
	}

	s.keys[label] = &ledgerKey{info: info, privateKey: []byte(priv), publicKey: []byte(pub)}
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

// DeleteKey removes a key.
func (s *Signer) DeleteKey(_ context.Context, label string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.connected {
		return hsm.ErrNotConnected
	}
	k, ok := s.keys[label]
	if !ok {
		return hsm.ErrKeyNotFound
	}
	scrubBytes(k.privateKey)
	delete(s.keys, label)
	return nil
}

// Sign signs data using the Ledger device. On a real Ledger this would
// prompt the user for confirmation.
func (s *Signer) Sign(_ context.Context, label string, data []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.connected {
		return nil, hsm.ErrNotConnected
	}
	k, ok := s.keys[label]
	if !ok {
		return nil, hsm.ErrKeyNotFound
	}

	s.logger.Debug("Ledger signing (mock; real device would prompt user)")

	priv := ed25519.PrivateKey(k.privateKey)
	return ed25519.Sign(priv, data), nil
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
	pub := make([]byte, len(k.publicKey))
	copy(pub, k.publicKey)
	return ed25519.PublicKey(pub), nil
}

func scrubBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

var _ hsm.HSMProvider = (*Signer)(nil)
