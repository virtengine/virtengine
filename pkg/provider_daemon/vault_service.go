package provider_daemon

import (
	"fmt"
	"time"

	"github.com/virtengine/virtengine/pkg/artifact_store"
	"github.com/virtengine/virtengine/pkg/data_vault"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
)

// VaultServiceConfig configures the data vault service.
type VaultServiceConfig struct {
	Enabled          bool
	Backend          string
	AuditOwner       string
	OrgResolver      data_vault.OrgResolver
	RotateOverlap    time.Duration
	AnomalyWindow    time.Duration
	AnomalyThreshold int
}

// DefaultVaultServiceConfig returns default vault config.
func DefaultVaultServiceConfig() VaultServiceConfig {
	return VaultServiceConfig{
		Enabled:          true,
		Backend:          "memory",
		AuditOwner:       "audit-system",
		RotateOverlap:    24 * time.Hour,
		AnomalyWindow:    10 * time.Minute,
		AnomalyThreshold: 5,
	}
}

// NewVaultService constructs a data vault service.
func NewVaultService(cfg VaultServiceConfig) (data_vault.VaultService, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	backend, err := createVaultBackend(cfg.Backend)
	if err != nil {
		return nil, err
	}

	keyMgr := keys.NewKeyManager()
	if err := keyMgr.Initialize(); err != nil {
		return nil, fmt.Errorf("init vault keys: %w", err)
	}

	store := data_vault.NewEncryptedBlobStore(backend, keyMgr)
	metrics := data_vault.NewVaultMetrics()

	accessPolicy := data_vault.DefaultAccessPolicy()
	for scope, policy := range accessPolicy.ScopePolicies {
		policy.AllowedRoles = nil
		accessPolicy.ScopePolicies[scope] = policy
	}
	accessControl := data_vault.NewPolicyAccessControl(accessPolicy, nil, cfg.OrgResolver)

	auditLogger := data_vault.NewAuditLogger(data_vault.DefaultAuditLogConfig(), nil)
	auditLogger.RegisterExporter(data_vault.NewVaultAuditExporter(store, cfg.AuditOwner))

	anomalyDetector := data_vault.NewAccessAnomalyDetector(cfg.AnomalyThreshold, cfg.AnomalyWindow, nil)

	return data_vault.NewVaultService(data_vault.VaultConfig{
		Store:              store,
		AccessControl:      accessControl,
		AuditLogger:        auditLogger,
		AuditOwner:         cfg.AuditOwner,
		Metrics:            metrics,
		AnomalyDetector:    anomalyDetector,
		KeyRotationOverlap: cfg.RotateOverlap,
	})
}

func createVaultBackend(backend string) (artifact_store.ArtifactStore, error) {
	switch backend {
	case "", "memory":
		return artifact_store.NewMemoryBackend(), nil
	default:
		return nil, fmt.Errorf("unsupported vault backend: %s", backend)
	}
}
