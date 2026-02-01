// Package edugain provides EduGAIN federation integration.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
package edugain

import (
	"errors"
	"time"
)

// ============================================================================
// Configuration
// ============================================================================

// Config holds the configuration for the EduGAIN service
type Config struct {
	// Enabled indicates if EduGAIN integration is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// SPEntityID is the Service Provider entity ID
	SPEntityID string `json:"sp_entity_id" yaml:"sp_entity_id"`

	// SPDisplayName is the Service Provider display name
	SPDisplayName string `json:"sp_display_name" yaml:"sp_display_name"`

	// MetadataURL is the EduGAIN federation metadata URL
	MetadataURL string `json:"metadata_url" yaml:"metadata_url"`

	// MetadataRefreshInterval is how often to refresh metadata
	MetadataRefreshInterval time.Duration `json:"metadata_refresh_interval" yaml:"metadata_refresh_interval"`

	// AssertionConsumerServiceURL is the ACS endpoint URL
	AssertionConsumerServiceURL string `json:"assertion_consumer_service_url" yaml:"assertion_consumer_service_url"`

	// SingleLogoutServiceURL is the SLO endpoint URL
	SingleLogoutServiceURL string `json:"single_logout_service_url" yaml:"single_logout_service_url"`

	// SessionDuration is how long sessions remain valid
	SessionDuration time.Duration `json:"session_duration" yaml:"session_duration"`

	// ClockSkew is the allowed clock skew for time validation
	ClockSkew time.Duration `json:"clock_skew" yaml:"clock_skew"`

	// AssertionMaxAge is the maximum age of a SAML assertion
	AssertionMaxAge time.Duration `json:"assertion_max_age" yaml:"assertion_max_age"`

	// ReplayWindowDuration is how long to track assertion IDs
	ReplayWindowDuration time.Duration `json:"replay_window_duration" yaml:"replay_window_duration"`

	// RequireMFA requires all authentications to use REFEDS MFA profile
	RequireMFA bool `json:"require_mfa" yaml:"require_mfa"`

	// RequireEncryptedAssertions requires encrypted assertions
	RequireEncryptedAssertions bool `json:"require_encrypted_assertions" yaml:"require_encrypted_assertions"`

	// AllowedAffiliations restricts which affiliations can authenticate
	// Empty means all affiliations are allowed
	AllowedAffiliations []AffiliationType `json:"allowed_affiliations" yaml:"allowed_affiliations"`

	// AllowedInstitutions restricts which IdPs can be used
	// Empty means all federation IdPs are allowed
	AllowedInstitutions []string `json:"allowed_institutions" yaml:"allowed_institutions"`

	// BlockedInstitutions prevents specific IdPs from being used
	BlockedInstitutions []string `json:"blocked_institutions" yaml:"blocked_institutions"`

	// TrustedFederations restricts which federations are trusted
	// Empty means all EduGAIN federations are trusted
	TrustedFederations []string `json:"trusted_federations" yaml:"trusted_federations"`

	// SPCertificatePath is the path to SP signing certificate (PEM)
	SPCertificatePath string `json:"sp_certificate_path" yaml:"sp_certificate_path"`

	// SPPrivateKeyPath is the path to SP private key (PEM)
	SPPrivateKeyPath string `json:"sp_private_key_path" yaml:"sp_private_key_path"`

	// EncryptionCertificatePath is the path to encryption certificate (PEM)
	EncryptionCertificatePath string `json:"encryption_certificate_path" yaml:"encryption_certificate_path"`

	// EncryptionPrivateKeyPath is the path to encryption private key (PEM)
	EncryptionPrivateKeyPath string `json:"encryption_private_key_path" yaml:"encryption_private_key_path"`

	// MetadataCertificatePath is the path to metadata signing certificate (PEM)
	// Used to verify federation metadata signature
	MetadataCertificatePath string `json:"metadata_certificate_path" yaml:"metadata_certificate_path"`

	// PreferredBinding is the preferred SAML binding
	PreferredBinding string `json:"preferred_binding" yaml:"preferred_binding"`

	// NameIDFormat is the preferred NameID format
	NameIDFormat string `json:"name_id_format" yaml:"name_id_format"`

	// StoreRawAssertions determines if raw XML is stored (for audit)
	StoreRawAssertions bool `json:"store_raw_assertions" yaml:"store_raw_assertions"`

	// VEIDIntegration contains VEID integration settings
	VEIDIntegration VEIDIntegrationConfig `json:"veid_integration" yaml:"veid_integration"`

	// SessionStorage contains session storage settings
	SessionStorage SessionStorageConfig `json:"session_storage" yaml:"session_storage"`

	// MetadataCache contains metadata caching settings
	MetadataCache MetadataCacheConfig `json:"metadata_cache" yaml:"metadata_cache"`

	// Logging contains logging configuration
	Logging LoggingConfig `json:"logging" yaml:"logging"`
}

// VEIDIntegrationConfig contains VEID integration settings
type VEIDIntegrationConfig struct {
	// Enabled indicates if VEID integration is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// CreateScopeOnAuth creates a VEID scope on successful authentication
	CreateScopeOnAuth bool `json:"create_scope_on_auth" yaml:"create_scope_on_auth"`

	// RequireExistingIdentity requires an existing VEID identity
	RequireExistingIdentity bool `json:"require_existing_identity" yaml:"require_existing_identity"`

	// ScoreWeight is the identity score weight for EduGAIN scopes
	ScoreWeight uint32 `json:"score_weight" yaml:"score_weight"`

	// EnrichFromAttributes populates VEID from eduPerson attributes
	EnrichFromAttributes bool `json:"enrich_from_attributes" yaml:"enrich_from_attributes"`

	// HashSensitiveData hashes PII before VEID storage
	HashSensitiveData bool `json:"hash_sensitive_data" yaml:"hash_sensitive_data"`
}

// SessionStorageConfig contains session storage settings
type SessionStorageConfig struct {
	// Type is the storage type (memory, redis, database)
	Type string `json:"type" yaml:"type"`

	// RedisURL is the Redis connection URL (for redis type)
	RedisURL string `json:"redis_url" yaml:"redis_url"`

	// DatabaseDSN is the database connection string (for database type)
	DatabaseDSN string `json:"database_dsn" yaml:"database_dsn"`

	// EncryptionKey is the key for encrypting session data
	// SECURITY: Should be loaded from secure storage
	EncryptionKey string `json:"encryption_key" yaml:"encryption_key"`

	// MaxSessions is the maximum sessions per wallet
	MaxSessions int `json:"max_sessions" yaml:"max_sessions"`

	// CleanupInterval is how often to cleanup expired sessions
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

// MetadataCacheConfig contains metadata caching settings
type MetadataCacheConfig struct {
	// Type is the cache type (memory, file, redis)
	Type string `json:"type" yaml:"type"`

	// FilePath is the file path for file cache
	FilePath string `json:"file_path" yaml:"file_path"`

	// RedisURL is the Redis URL for redis cache
	RedisURL string `json:"redis_url" yaml:"redis_url"`

	// MaxAge is the maximum cache age
	MaxAge time.Duration `json:"max_age" yaml:"max_age"`

	// StaleWhileRevalidate allows serving stale cache while refreshing
	StaleWhileRevalidate bool `json:"stale_while_revalidate" yaml:"stale_while_revalidate"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	// Level is the log level (debug, info, warn, error)
	Level string `json:"level" yaml:"level"`

	// RedactPII redacts personally identifiable information
	RedactPII bool `json:"redact_pii" yaml:"redact_pii"`

	// LogAssertions logs SAML assertions (for debugging)
	// SECURITY: Should be false in production
	LogAssertions bool `json:"log_assertions" yaml:"log_assertions"`

	// LogMetadataRefresh logs metadata refresh events
	LogMetadataRefresh bool `json:"log_metadata_refresh" yaml:"log_metadata_refresh"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Enabled:                 true,
		SPEntityID:              "https://virtengine.com/saml/metadata",
		SPDisplayName:           "VirtEngine",
		MetadataURL:             EduGAINProductionMetadataURL,
		MetadataRefreshInterval: DefaultMetadataRefreshInterval,
		SessionDuration:         DefaultSessionDuration,
		ClockSkew:               DefaultClockSkew,
		AssertionMaxAge:         DefaultAssertionMaxAge,
		ReplayWindowDuration:    DefaultReplayWindowDuration,
		RequireMFA:              false,
		PreferredBinding:        SAMLBindingHTTPPOST,
		NameIDFormat:            NameIDFormatPersistent,
		StoreRawAssertions:      false,
		VEIDIntegration: VEIDIntegrationConfig{
			Enabled:              true,
			CreateScopeOnAuth:    true,
			ScoreWeight:          15, // Higher than SSO metadata but lower than ID documents
			EnrichFromAttributes: true,
			HashSensitiveData:    true,
		},
		SessionStorage: SessionStorageConfig{
			Type:            "memory",
			MaxSessions:     5,
			CleanupInterval: 1 * time.Hour,
		},
		MetadataCache: MetadataCacheConfig{
			Type:                 "memory",
			MaxAge:               24 * time.Hour,
			StaleWhileRevalidate: true,
		},
		Logging: LoggingConfig{
			Level:              "info",
			RedactPII:          true,
			LogAssertions:      false,
			LogMetadataRefresh: true,
		},
	}
}

// TestConfig returns a configuration suitable for testing
func TestConfig() Config {
	cfg := DefaultConfig()
	cfg.MetadataURL = EduGAINTestMetadataURL
	cfg.SessionDuration = 1 * time.Hour
	cfg.Logging.LogAssertions = true
	cfg.StoreRawAssertions = true
	return cfg
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil // Skip validation if disabled
	}

	if c.SPEntityID == "" {
		return errors.New("sp_entity_id is required")
	}

	if c.MetadataURL == "" {
		return errors.New("metadata_url is required")
	}

	if c.AssertionConsumerServiceURL == "" {
		return errors.New("assertion_consumer_service_url is required")
	}

	if c.MetadataRefreshInterval < 5*time.Minute {
		return errors.New("metadata_refresh_interval must be at least 5 minutes")
	}

	if c.SessionDuration < 1*time.Minute {
		return errors.New("session_duration must be at least 1 minute")
	}

	if c.SessionDuration > 24*time.Hour {
		return errors.New("session_duration must be at most 24 hours")
	}

	if c.AssertionMaxAge < 1*time.Minute {
		return errors.New("assertion_max_age must be at least 1 minute")
	}

	// Validate preferred binding
	switch c.PreferredBinding {
	case SAMLBindingHTTPRedirect, SAMLBindingHTTPPOST, SAMLBindingHTTPArtifact:
		// Valid bindings
	default:
		return errors.New("invalid preferred_binding")
	}

	// Validate NameID format
	switch c.NameIDFormat {
	case NameIDFormatPersistent, NameIDFormatTransient, NameIDFormatEmailAddress, NameIDFormatUnspecified:
		// Valid formats
	default:
		return errors.New("invalid name_id_format")
	}

	// Validate session storage type
	switch c.SessionStorage.Type {
	case "memory", "redis", "database":
		// Valid types
	default:
		return errors.New("invalid session_storage.type")
	}

	// Validate cache type
	switch c.MetadataCache.Type {
	case "memory", "file", "redis":
		// Valid types
	default:
		return errors.New("invalid metadata_cache.type")
	}

	// Validate allowed affiliations
	for _, aff := range c.AllowedAffiliations {
		if !IsValidAffiliationType(aff) {
			return errors.New("invalid affiliation type in allowed_affiliations: " + string(aff))
		}
	}

	return nil
}

// IsInstitutionAllowed checks if an institution is allowed
func (c *Config) IsInstitutionAllowed(entityID string) bool {
	// Check blocked list first
	for _, blocked := range c.BlockedInstitutions {
		if blocked == entityID {
			return false
		}
	}

	// If allowed list is specified, check it
	if len(c.AllowedInstitutions) > 0 {
		for _, allowed := range c.AllowedInstitutions {
			if allowed == entityID {
				return true
			}
		}
		return false
	}

	// No restrictions
	return true
}

// IsFederationTrusted checks if a federation is trusted
func (c *Config) IsFederationTrusted(federationName string) bool {
	if len(c.TrustedFederations) == 0 {
		return true // All federations trusted
	}
	for _, trusted := range c.TrustedFederations {
		if trusted == federationName {
			return true
		}
	}
	return false
}

// IsAffiliationAllowed checks if an affiliation is allowed
func (c *Config) IsAffiliationAllowed(affiliation AffiliationType) bool {
	if len(c.AllowedAffiliations) == 0 {
		return true // All affiliations allowed
	}
	for _, allowed := range c.AllowedAffiliations {
		if allowed == affiliation {
			return true
		}
	}
	return false
}

