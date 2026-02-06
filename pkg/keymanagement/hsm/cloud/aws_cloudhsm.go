// Package cloud provides HSMProvider implementations for cloud-based HSM
// services (AWS CloudHSM, GCP Cloud HSM, Azure Dedicated HSM).
//
// Each adapter translates the hsm.HSMProvider interface into the
// cloud-specific API calls. In production these would use the respective
// cloud SDKs; the implementations here provide the structural scaffold
// with a software fallback for testing.
package cloud

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

// AWSCloudHSMProvider implements hsm.HSMProvider for AWS CloudHSM.
// In production this would use the AWS CloudHSM PKCS#11 library or the
// AWS SDK. The current implementation provides a software fallback.
type AWSCloudHSMProvider struct {
	config    hsm.CloudConfig
	logger    *slog.Logger
	mu        sync.Mutex
	connected bool
	keys      map[string]*cloudKey
}

type cloudKey struct {
	info       *hsm.KeyInfo
	privateKey []byte
	publicKey  []byte
}

// NewAWSCloudHSMProvider creates a new AWS CloudHSM provider.
func NewAWSCloudHSMProvider(config hsm.CloudConfig, logger *slog.Logger) (*AWSCloudHSMProvider, error) {
	if config.Region == "" {
		return nil, fmt.Errorf("aws cloudhsm: region is required")
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &AWSCloudHSMProvider{
		config: config,
		logger: logger,
		keys:   make(map[string]*cloudKey),
	}, nil
}

func (p *AWSCloudHSMProvider) Connect(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connected = true
	p.logger.Info("AWS CloudHSM connected", slog.String("region", p.config.Region))
	return nil
}

func (p *AWSCloudHSMProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, k := range p.keys {
		scrub(k.privateKey)
	}
	p.keys = make(map[string]*cloudKey)
	p.connected = false
	return nil
}

func (p *AWSCloudHSMProvider) GenerateKey(_ context.Context, keyType hsm.KeyType, label string) (*hsm.KeyInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.connected {
		return nil, hsm.ErrNotConnected
	}
	if _, ok := p.keys[label]; ok {
		return nil, hsm.ErrKeyExists
	}
	pub, priv, err := generateKey(keyType)
	if err != nil {
		return nil, err
	}
	info := makeKeyInfo(label, keyType, pub)
	p.keys[label] = &cloudKey{info: info, privateKey: priv, publicKey: pub}
	return info, nil
}

func (p *AWSCloudHSMProvider) ImportKey(_ context.Context, keyType hsm.KeyType, label string, key []byte) (*hsm.KeyInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.connected {
		return nil, hsm.ErrNotConnected
	}
	if _, ok := p.keys[label]; ok {
		return nil, hsm.ErrKeyExists
	}
	pub, err := extractPublicKey(keyType, key)
	if err != nil {
		return nil, err
	}
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	info := makeKeyInfo(label, keyType, pub)
	p.keys[label] = &cloudKey{info: info, privateKey: keyCopy, publicKey: pub}
	return info, nil
}

func (p *AWSCloudHSMProvider) GetKey(_ context.Context, label string) (*hsm.KeyInfo, error) {
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

func (p *AWSCloudHSMProvider) ListKeys(_ context.Context) ([]*hsm.KeyInfo, error) {
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

func (p *AWSCloudHSMProvider) DeleteKey(_ context.Context, label string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.connected {
		return hsm.ErrNotConnected
	}
	k, ok := p.keys[label]
	if !ok {
		return hsm.ErrKeyNotFound
	}
	scrub(k.privateKey)
	delete(p.keys, label)
	return nil
}

func (p *AWSCloudHSMProvider) Sign(_ context.Context, label string, data []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.connected {
		return nil, hsm.ErrNotConnected
	}
	return signWithKey(p.keys, label, data)
}

func (p *AWSCloudHSMProvider) GetPublicKey(_ context.Context, label string) (crypto.PublicKey, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.connected {
		return nil, hsm.ErrNotConnected
	}
	return getPublicKey(p.keys, label)
}

var _ hsm.HSMProvider = (*AWSCloudHSMProvider)(nil)

// --- helpers shared across cloud adapters ---

func generateKey(keyType hsm.KeyType) (pub, priv []byte, err error) {
	switch keyType {
	case hsm.KeyTypeEd25519:
		pubKey, privKey, err := ed25519.GenerateKey(crand.Reader)
		if err != nil {
			return nil, nil, err
		}
		return []byte(pubKey), []byte(privKey), nil
	default:
		return nil, nil, fmt.Errorf("%w: %s", hsm.ErrUnsupportedKeyType, keyType)
	}
}

func extractPublicKey(keyType hsm.KeyType, key []byte) ([]byte, error) {
	switch keyType {
	case hsm.KeyTypeEd25519:
		if len(key) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid ed25519 key size: %d", len(key))
		}
		priv := ed25519.PrivateKey(key)
		return []byte(priv.Public().(ed25519.PublicKey)), nil
	default:
		return nil, fmt.Errorf("%w: %s", hsm.ErrUnsupportedKeyType, keyType)
	}
}

func makeKeyInfo(label string, keyType hsm.KeyType, pub []byte) *hsm.KeyInfo {
	h := sha256.Sum256(pub)
	fp := hex.EncodeToString(h[:])
	return &hsm.KeyInfo{
		Label:       label,
		ID:          []byte(fp[:8]),
		Type:        keyType,
		Size:        len(pub) * 8,
		Extractable: false,
		CreatedAt:   time.Now().UTC(),
		Fingerprint: fp,
	}
}

func signWithKey(keys map[string]*cloudKey, label string, data []byte) ([]byte, error) {
	k, ok := keys[label]
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

func getPublicKey(keys map[string]*cloudKey, label string) (crypto.PublicKey, error) {
	k, ok := keys[label]
	if !ok {
		return nil, hsm.ErrKeyNotFound
	}
	pub := make([]byte, len(k.publicKey))
	copy(pub, k.publicKey)
	switch k.info.Type {
	case hsm.KeyTypeEd25519:
		return ed25519.PublicKey(pub), nil
	default:
		return nil, fmt.Errorf("%w: %s", hsm.ErrUnsupportedKeyType, k.info.Type)
	}
}

func scrub(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
