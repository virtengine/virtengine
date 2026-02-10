package cloud

import (
	"context"
	"crypto"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

// GCPCloudHSMProvider implements hsm.HSMProvider for Google Cloud HSM via
// Cloud KMS. In production this would use the Cloud KMS Go SDK.
type GCPCloudHSMProvider struct {
	config    hsm.CloudConfig
	logger    *slog.Logger
	mu        sync.Mutex
	connected bool
	keys      map[string]*cloudKey
}

// NewGCPCloudHSMProvider creates a new GCP Cloud HSM provider.
func NewGCPCloudHSMProvider(config hsm.CloudConfig, logger *slog.Logger) (*GCPCloudHSMProvider, error) {
	if config.ProjectID == "" {
		return nil, fmt.Errorf("gcp cloudhsm: project_id is required")
	}
	if config.KeyRingName == "" {
		return nil, fmt.Errorf("gcp cloudhsm: key_ring_name is required")
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &GCPCloudHSMProvider{
		config: config,
		logger: logger,
		keys:   make(map[string]*cloudKey),
	}, nil
}

func (p *GCPCloudHSMProvider) Connect(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connected = true
	p.logger.Info("GCP Cloud HSM connected",
		slog.String("project", p.config.ProjectID),
		slog.String("key_ring", p.config.KeyRingName),
	)
	return nil
}

func (p *GCPCloudHSMProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, k := range p.keys {
		scrub(k.privateKey)
	}
	p.keys = make(map[string]*cloudKey)
	p.connected = false
	return nil
}

func (p *GCPCloudHSMProvider) GenerateKey(_ context.Context, keyType hsm.KeyType, label string) (*hsm.KeyInfo, error) {
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

func (p *GCPCloudHSMProvider) ImportKey(_ context.Context, keyType hsm.KeyType, label string, key []byte) (*hsm.KeyInfo, error) {
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

func (p *GCPCloudHSMProvider) GetKey(_ context.Context, label string) (*hsm.KeyInfo, error) {
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

func (p *GCPCloudHSMProvider) ListKeys(_ context.Context) ([]*hsm.KeyInfo, error) {
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

func (p *GCPCloudHSMProvider) DeleteKey(_ context.Context, label string) error {
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

func (p *GCPCloudHSMProvider) Sign(_ context.Context, label string, data []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.connected {
		return nil, hsm.ErrNotConnected
	}
	return signWithKey(p.keys, label, data)
}

func (p *GCPCloudHSMProvider) GetPublicKey(_ context.Context, label string) (crypto.PublicKey, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.connected {
		return nil, hsm.ErrNotConnected
	}
	return getPublicKey(p.keys, label)
}

var _ hsm.HSMProvider = (*GCPCloudHSMProvider)(nil)
