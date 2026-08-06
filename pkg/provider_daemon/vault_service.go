package provider_daemon

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/virtengine/virtengine/pkg/artifact_store"
	"github.com/virtengine/virtengine/pkg/data_vault"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
)

// VaultServiceConfig configures the data vault service.
type VaultServiceConfig struct {
	Enabled                  bool
	Backend                  string
	ArtifactBackend          artifact_store.ArtifactStore
	BlobCipher               data_vault.BlobCipher
	Environment              string
	Profile                  string
	DevelopmentOnly          bool
	ArtifactPath             string
	KeyStatePath             string
	AuditPath                string
	FixtureWrappingKey       []byte
	RevisionAnchor           data_vault.RevisionAnchor
	UnsafeWindowsDevelopment bool
	InitializeFixtureKeys    bool
	KeyManager               keys.VaultKeyManager
	KeyPersistence           keys.StatePersistence
	AuditStore               data_vault.AuditStore
	ConsentResolver          data_vault.ConsentResolver
	AuditOwner               string
	OrgResolver              data_vault.OrgResolver
	RoleResolver             data_vault.RoleResolver
	RotateOverlap            time.Duration
	AnomalyWindow            time.Duration
	AnomalyThreshold         int
}

// DefaultVaultServiceConfig returns default vault config.
func DefaultVaultServiceConfig() VaultServiceConfig {
	return VaultServiceConfig{
		Enabled:          false,
		Environment:      "production",
		Profile:          "disabled",
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
	if err := validateVaultServiceConfig(cfg); err != nil {
		return nil, err
	}

	backend, err := createVaultBackend(cfg)
	if err != nil {
		return nil, err
	}

	keyMgr := cfg.KeyManager
	if keyMgr == nil {
		keyMgr, err = createVaultKeyManager(cfg)
		if err != nil {
			closeIfPossible(backend)
			return nil, err
		}
	}

	store, err := data_vault.NewEncryptedBlobStoreWithCipher(backend, keyMgr, cfg.BlobCipher)
	if err != nil {
		closeIfPossible(backend)
		closeIfPossible(keyMgr)
		return nil, fmt.Errorf("open vault store: %w", err)
	}
	metrics := data_vault.NewVaultMetrics()

	accessPolicy := data_vault.DefaultAccessPolicy()
	accessControl := data_vault.NewPolicyAccessControl(accessPolicy, cfg.RoleResolver, cfg.OrgResolver)

	auditStore := cfg.AuditStore
	if auditStore == nil && cfg.AuditPath != "" {
		auditStore, err = data_vault.NewFixtureFileAuditStoreWithSecurity(cfg.AuditPath, cfg.Profile, data_vault.FixtureSecurityOptions{
			UnsafeWindowsDevelopment: cfg.UnsafeWindowsDevelopment,
		}, cfg.RevisionAnchor)
		if err != nil {
			_ = store.Close()
			return nil, fmt.Errorf("open vault audit store: %w", err)
		}
	}
	if auditStore == nil && cfg.DevelopmentOnly {
		auditStore = data_vault.NewMemoryAuditStore()
	}
	auditLogger := data_vault.NewAuditLogger(data_vault.DefaultAuditLogConfig(), auditStore)
	auditLogger.RegisterExporter(data_vault.NewVaultAuditExporter(store, cfg.AuditOwner))

	anomalyDetector := data_vault.NewAccessAnomalyDetector(cfg.AnomalyThreshold, cfg.AnomalyWindow, nil)

	vault, err := data_vault.NewVaultService(data_vault.VaultConfig{
		Store:              store,
		AccessControl:      accessControl,
		ConsentResolver:    cfg.ConsentResolver,
		AuditLogger:        auditLogger,
		AuditOwner:         cfg.AuditOwner,
		Metrics:            metrics,
		AnomalyDetector:    anomalyDetector,
		KeyRotationOverlap: cfg.RotateOverlap,
	})
	if err != nil {
		_ = auditLogger.Close()
		_ = store.Close()
		return nil, err
	}
	return vault, nil
}

func closeIfPossible(value any) {
	if closer, ok := value.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}

func createVaultBackend(cfg VaultServiceConfig) (artifact_store.ArtifactStore, error) {
	switch cfg.Backend {
	case "external":
		if cfg.ArtifactBackend == nil {
			return nil, errors.New("external vault backend is required")
		}
		return cfg.ArtifactBackend, nil
	case "memory":
		return artifact_store.NewMemoryBackend(), nil
	case "fixture-filesystem":
		return data_vault.NewFixtureFileArtifactStoreWithSecurity(cfg.ArtifactPath, cfg.Profile, data_vault.FixtureSecurityOptions{
			UnsafeWindowsDevelopment: cfg.UnsafeWindowsDevelopment,
		}, cfg.RevisionAnchor)
	default:
		return nil, fmt.Errorf("unsupported vault backend: %s", cfg.Backend)
	}
}

func createVaultKeyManager(cfg VaultServiceConfig) (keys.VaultKeyManager, error) {
	if cfg.Backend == "memory" {
		manager := keys.NewKeyManager()
		if err := manager.Initialize(); err != nil {
			return nil, fmt.Errorf("init development vault keys: %w", err)
		}
		return manager, nil
	}
	persistence := cfg.KeyPersistence
	if persistence == nil {
		var err error
		persistence, err = keys.NewFixtureFilePersistenceWithSecurity(cfg.KeyStatePath, cfg.FixtureWrappingKey, cfg.Profile, data_vault.FixtureSecurityOptions{
			UnsafeWindowsDevelopment: cfg.UnsafeWindowsDevelopment,
		}, cfg.RevisionAnchor)
		if err != nil {
			return nil, fmt.Errorf("open fixture key custody: %w", err)
		}
	}
	if cfg.InitializeFixtureKeys {
		manager, err := keys.NewUninitializedPersistentKeyManager(persistence)
		if err != nil {
			return nil, fmt.Errorf("create fixture key manager: %w", err)
		}
		if err := manager.Initialize(); err != nil {
			return nil, fmt.Errorf("initialize fixture key manager: %w", err)
		}
		return manager, nil
	}
	manager, err := keys.NewPersistentKeyManager(persistence)
	if err != nil {
		return nil, fmt.Errorf("restore fixture key manager: %w", err)
	}
	return manager, nil
}

func validateVaultServiceConfig(cfg VaultServiceConfig) error {
	if cfg.Environment == "production" || cfg.Profile == "production" {
		violations := make([]string, 0)
		if cfg.Backend != "external" || !isDurableArtifactStore(cfg.ArtifactBackend) {
			violations = append(violations, "production requires a production durable artifact backend")
		}
		if nonExportable, ok := cfg.KeyManager.(keys.NonExportableVaultKeyManager); !ok || !nonExportable.UsesNonExportableKeys() {
			violations = append(violations, "production requires a non-exportable KMS operation interface; prototype key export is forbidden")
		}
		if cfg.BlobCipher == nil || !cfg.BlobCipher.ProductionSafe() {
			violations = append(violations, "production requires a production-safe KMS envelope cipher")
		}
		if cfg.RoleResolver == nil {
			violations = append(violations, "production requires role resolver")
		}
		if cfg.OrgResolver == nil {
			violations = append(violations, "production requires organization resolver")
		}
		if cfg.ConsentResolver == nil {
			violations = append(violations, "production requires consent resolver")
		}
		if !durableAuditStore(cfg.AuditStore) {
			violations = append(violations, "production requires durable audit store")
		}
		if cfg.RevisionAnchor == nil || cfg.RevisionAnchor.Replayable() || cfg.RevisionAnchor.Local() {
			violations = append(violations, "production requires a non-local non-replayable revision anchor")
		}
		if cfg.UnsafeWindowsDevelopment {
			violations = append(violations, "production rejects UnsafeWindowsDevelopment")
		}
		if len(violations) > 0 {
			return errors.New(strings.Join(violations, "; "))
		}
		return nil
	}
	if cfg.Backend == "memory" {
		if cfg.Environment != "development" || cfg.Profile != "development" || !cfg.DevelopmentOnly {
			return errors.New("memory vault requires explicit development environment, development profile, and DevelopmentOnly")
		}
		return nil
	}
	if cfg.Backend == "fixture-filesystem" {
		if cfg.Profile != "fixture" && cfg.Profile != "development" {
			return errors.New("fixture filesystem vault requires fixture or development profile")
		}
		if cfg.ArtifactPath == "" || (cfg.KeyPersistence == nil && (cfg.KeyStatePath == "" || len(cfg.FixtureWrappingKey) != 32)) {
			return errors.New("fixture filesystem vault requires artifact path and durable key custody inputs")
		}
		if cfg.AuditStore == nil && cfg.AuditPath == "" {
			return errors.New("fixture filesystem vault requires durable audit store or audit path")
		}
		if cfg.RevisionAnchor == nil || cfg.RevisionAnchor.Replayable() {
			return errors.New("fixture filesystem vault requires a non-replayable revision anchor")
		}
		if runtime.GOOS == "windows" && !cfg.UnsafeWindowsDevelopment {
			return errors.New("fixture filesystem vault on Windows requires explicit UnsafeWindowsDevelopment")
		}
		return nil
	}
	return fmt.Errorf("unsupported vault backend: %s", cfg.Backend)
}

func durableAuditStore(store data_vault.AuditStore) bool {
	durable, ok := store.(data_vault.DurableAuditStore)
	return ok && durable.Durable()
}

func isDurableArtifactStore(store artifact_store.ArtifactStore) bool {
	durable, ok := store.(data_vault.DurableVaultArtifactStore)
	return ok && durable.Durable()
}
