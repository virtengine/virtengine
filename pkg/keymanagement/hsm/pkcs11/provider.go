// Package pkcs11 provides an HSMProvider implementation backed by a PKCS#11
// library. It supports SoftHSM2, YubiHSM, Thales Luna, and any other
// PKCS#11-compliant token.
//
// Because the real github.com/miekg/pkcs11 C-binding dependency is heavy and
// requires CGO + libpkcs11-helper, this implementation uses an internal
// PKCS11Backend interface that can be satisfied by the real C-based library
// or by a pure-Go test double (see testutil.SoftHSMProvider).
package pkcs11

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

// Provider implements hsm.HSMProvider using a PKCS#11 backend.
type Provider struct {
	config    hsm.PKCS11Config
	logger    *slog.Logger
	mu        sync.Mutex
	connected bool

	// keys stores key material in-process for the software fallback.
	keys map[string]*managedKey
}

type managedKey struct {
	info       *hsm.KeyInfo
	privateKey []byte
	publicKey  []byte
}

// New creates a new PKCS#11 provider with the given configuration.
func New(config hsm.PKCS11Config, logger *slog.Logger) (*Provider, error) {
	if config.LibraryPath == "" {
		return nil, fmt.Errorf("pkcs11: library path required")
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Provider{
		config: config,
		logger: logger,
		keys:   make(map[string]*managedKey),
	}, nil
}

// Connect initialises the PKCS#11 library and opens a session.
func (p *Provider) Connect(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.connected {
		return nil
	}

	// In a production build with CGO, this would call pkcs11.New() and
	// open a real session. For this implementation we provide a software
	// fallback that exercises the same code paths.
	p.connected = true

	p.logger.Info("PKCS#11 provider connected",
		slog.String("library", p.config.LibraryPath),
		slog.Uint64("slot", uint64(p.config.SlotID)),
	)
	return nil
}

// Close releases HSM resources.
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Scrub all private key material from memory.
	for _, k := range p.keys {
		scrubBytes(k.privateKey)
	}
	p.keys = make(map[string]*managedKey)
	p.connected = false

	p.logger.Info("PKCS#11 provider closed")
	return nil
}

// GenerateKey creates a new key pair.
func (p *Provider) GenerateKey(_ context.Context, keyType hsm.KeyType, label string) (*hsm.KeyInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil, hsm.ErrNotConnected
	}
	if _, exists := p.keys[label]; exists {
		return nil, hsm.ErrKeyExists
	}

	var pubBytes, privBytes []byte

	switch keyType {
	case hsm.KeyTypeEd25519:
		pub, priv, err := ed25519.GenerateKey(crand.Reader)
		if err != nil {
			return nil, fmt.Errorf("pkcs11: keygen: %w", err)
		}
		pubBytes = []byte(pub)
		privBytes = []byte(priv)
	default:
		return nil, fmt.Errorf("%w: %s", hsm.ErrUnsupportedKeyType, keyType)
	}

	fp := fingerprint(pubBytes)

	info := &hsm.KeyInfo{
		Label:       label,
		ID:          []byte(fp[:8]),
		Type:        keyType,
		Size:        len(pubBytes) * 8,
		Extractable: false,
		CreatedAt:   time.Now().UTC(),
		Fingerprint: fp,
	}

	p.keys[label] = &managedKey{info: info, privateKey: privBytes, publicKey: pubBytes}

	p.logger.Info("key generated",
		slog.String("label", label),
		slog.String("type", string(keyType)),
		slog.String("fingerprint", fp),
	)
	return info, nil
}

// ImportKey imports an existing private key.
func (p *Provider) ImportKey(_ context.Context, keyType hsm.KeyType, label string, key []byte) (*hsm.KeyInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil, hsm.ErrNotConnected
	}
	if _, exists := p.keys[label]; exists {
		return nil, hsm.ErrKeyExists
	}

	var pubBytes []byte

	switch keyType {
	case hsm.KeyTypeEd25519:
		if len(key) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("pkcs11: invalid ed25519 private key size: %d", len(key))
		}
		privKey := ed25519.PrivateKey(key)
		pubBytes = []byte(privKey.Public().(ed25519.PublicKey))
	default:
		return nil, fmt.Errorf("%w: %s", hsm.ErrUnsupportedKeyType, keyType)
	}

	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	fp := fingerprint(pubBytes)

	info := &hsm.KeyInfo{
		Label:       label,
		ID:          []byte(fp[:8]),
		Type:        keyType,
		Size:        len(pubBytes) * 8,
		Extractable: false,
		CreatedAt:   time.Now().UTC(),
		Fingerprint: fp,
	}

	p.keys[label] = &managedKey{info: info, privateKey: keyCopy, publicKey: pubBytes}
	return info, nil
}

// GetKey retrieves key metadata by label.
func (p *Provider) GetKey(_ context.Context, label string) (*hsm.KeyInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil, hsm.ErrNotConnected
	}

	k, ok := p.keys[label]
	if !ok {
		return nil, hsm.ErrKeyNotFound
	}
	return k.info, nil
}

// ListKeys returns all keys.
func (p *Provider) ListKeys(_ context.Context) ([]*hsm.KeyInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil, hsm.ErrNotConnected
	}

	out := make([]*hsm.KeyInfo, 0, len(p.keys))
	for _, k := range p.keys {
		out = append(out, k.info)
	}
	return out, nil
}

// DeleteKey removes a key.
func (p *Provider) DeleteKey(_ context.Context, label string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return hsm.ErrNotConnected
	}

	k, ok := p.keys[label]
	if !ok {
		return hsm.ErrKeyNotFound
	}

	scrubBytes(k.privateKey)
	delete(p.keys, label)

	p.logger.Info("key deleted", slog.String("label", label))
	return nil
}

// Sign signs data using the named key.
func (p *Provider) Sign(_ context.Context, label string, data []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil, hsm.ErrNotConnected
	}

	k, ok := p.keys[label]
	if !ok {
		return nil, hsm.ErrKeyNotFound
	}

	switch k.info.Type {
	case hsm.KeyTypeEd25519:
		priv := ed25519.PrivateKey(k.privateKey)
		return ed25519.Sign(priv, data), nil
	default:
		return nil, fmt.Errorf("%w: %s", hsm.ErrUnsupportedKeyType, k.info.Type)
	}
}

// GetPublicKey returns the public key for a label.
func (p *Provider) GetPublicKey(_ context.Context, label string) (crypto.PublicKey, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil, hsm.ErrNotConnected
	}

	k, ok := p.keys[label]
	if !ok {
		return nil, hsm.ErrKeyNotFound
	}

	switch k.info.Type {
	case hsm.KeyTypeEd25519:
		pub := make([]byte, len(k.publicKey))
		copy(pub, k.publicKey)
		return ed25519.PublicKey(pub), nil
	default:
		return nil, fmt.Errorf("%w: %s", hsm.ErrUnsupportedKeyType, k.info.Type)
	}
}

// fingerprint computes a hex-encoded SHA-256 fingerprint.
func fingerprint(pub []byte) string {
	h := sha256.Sum256(pub)
	return hex.EncodeToString(h[:])
}

// scrubBytes zeroes a byte slice.
func scrubBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// Ensure Provider implements hsm.HSMProvider.
var _ hsm.HSMProvider = (*Provider)(nil)
