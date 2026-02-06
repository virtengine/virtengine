package hsm

import (
	"context"
	"crypto"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"
)

// Manager coordinates HSM providers, health-checks, and audit logging.
type Manager struct {
	config   Config
	provider HSMProvider
	logger   *slog.Logger
	mu       sync.RWMutex
	closed   bool
}

// NewManager creates a Manager with the given configuration. Call Connect to
// initialise the underlying provider.
func NewManager(config Config, logger *slog.Logger) (*Manager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("hsm manager: %w", err)
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Manager{
		config: config,
		logger: logger,
	}, nil
}

// SetProvider replaces the current HSM provider. The caller is responsible for
// closing the previous provider if one was set.
func (m *Manager) SetProvider(p HSMProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.provider = p
}

// Connect initialises the underlying HSM provider and connects to the device.
func (m *Manager) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.provider == nil {
		return fmt.Errorf("hsm manager: no provider configured")
	}

	ctx, cancel := context.WithTimeout(ctx, m.config.ConnectionTimeout)
	defer cancel()

	if err := m.provider.Connect(ctx); err != nil {
		return fmt.Errorf("hsm manager: connect: %w", err)
	}

	m.logger.Info("HSM connected", slog.String("backend", string(m.config.Backend)))
	return nil
}

// Close shuts down the HSM provider.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}
	m.closed = true

	if m.provider == nil {
		return nil
	}
	return m.provider.Close()
}

// GenerateKey creates a new key pair via the HSM.
func (m *Manager) GenerateKey(ctx context.Context, keyType KeyType, label string) (*KeyInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.ensureProvider(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, m.config.OperationTimeout)
	defer cancel()

	info, err := m.provider.GenerateKey(ctx, keyType, label)
	if err != nil {
		return nil, fmt.Errorf("hsm manager: generate key: %w", err)
	}

	m.audit("generate_key", label, string(keyType))
	return info, nil
}

// ImportKey imports a private key into the HSM.
func (m *Manager) ImportKey(ctx context.Context, keyType KeyType, label string, key []byte) (*KeyInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.ensureProvider(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, m.config.OperationTimeout)
	defer cancel()

	info, err := m.provider.ImportKey(ctx, keyType, label, key)
	if err != nil {
		return nil, fmt.Errorf("hsm manager: import key: %w", err)
	}

	m.audit("import_key", label, string(keyType))
	return info, nil
}

// GetKey retrieves key metadata by label.
func (m *Manager) GetKey(ctx context.Context, label string) (*KeyInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.ensureProvider(); err != nil {
		return nil, err
	}

	return m.provider.GetKey(ctx, label)
}

// ListKeys returns all keys in the HSM.
func (m *Manager) ListKeys(ctx context.Context) ([]*KeyInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.ensureProvider(); err != nil {
		return nil, err
	}

	return m.provider.ListKeys(ctx)
}

// DeleteKey removes a key from the HSM.
func (m *Manager) DeleteKey(ctx context.Context, label string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.ensureProvider(); err != nil {
		return err
	}

	if err := m.provider.DeleteKey(ctx, label); err != nil {
		return fmt.Errorf("hsm manager: delete key: %w", err)
	}

	m.audit("delete_key", label, "")
	return nil
}

// Sign signs data using the named key.
func (m *Manager) Sign(ctx context.Context, label string, data []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.ensureProvider(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, m.config.OperationTimeout)
	defer cancel()

	sig, err := m.provider.Sign(ctx, label, data)
	if err != nil {
		return nil, fmt.Errorf("hsm manager: sign: %w", err)
	}

	m.audit("sign", label, "")
	return sig, nil
}

// GetPublicKey returns the public key for a label.
func (m *Manager) GetPublicKey(ctx context.Context, label string) (crypto.PublicKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.ensureProvider(); err != nil {
		return nil, err
	}

	return m.provider.GetPublicKey(ctx, label)
}

// Provider returns the underlying provider for advanced usage.
func (m *Manager) Provider() HSMProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.provider
}

func (m *Manager) ensureProvider() error {
	if m.provider == nil {
		return fmt.Errorf("hsm manager: no provider configured")
	}
	if m.closed {
		return fmt.Errorf("hsm manager: manager is closed")
	}
	return nil
}

func (m *Manager) audit(operation, label, extra string) {
	if !m.config.AuditLog {
		return
	}
	m.logger.Info("HSM audit",
		slog.String("op", operation),
		slog.String("label", label),
		slog.String("extra", extra),
		slog.Time("time", time.Now().UTC()),
	)
}

// keySigner wraps an HSMProvider + label to implement the Signer interface.
type keySigner struct {
	provider HSMProvider
	label    string
	info     *KeyInfo
}

// NewSigner creates a Signer for the given key label. The provider must already
// be connected.
func NewSigner(provider HSMProvider, label string, info *KeyInfo) Signer {
	return &keySigner{provider: provider, label: label, info: info}
}

func (s *keySigner) Label() string     { return s.label }
func (s *keySigner) KeyInfo() *KeyInfo { return s.info }

func (s *keySigner) Public() crypto.PublicKey {
	pk, err := s.provider.GetPublicKey(context.Background(), s.label)
	if err != nil {
		return nil
	}
	return pk
}

func (s *keySigner) Sign(_ io.Reader, digest []byte, _ crypto.SignerOpts) ([]byte, error) {
	return s.provider.Sign(context.Background(), s.label, digest)
}
