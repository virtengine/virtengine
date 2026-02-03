// Package govdata provides government data source integration for identity verification.
//
// VE-909: Government data integration for authoritative identity verification
package govdata

import (
	"errors"
	"time"
)

// ============================================================================
// Configuration
// ============================================================================

// Config holds the configuration for the government data service
type Config struct {
	// Enabled indicates if government data integration is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// ServiceID is the unique service identifier
	ServiceID string `json:"service_id" yaml:"service_id"`

	// DefaultTimeout is the default request timeout
	DefaultTimeout time.Duration `json:"default_timeout" yaml:"default_timeout"`

	// MaxRetries is the maximum retry count for failed requests
	MaxRetries int `json:"max_retries" yaml:"max_retries"`

	// RetryBackoff is the retry backoff duration
	RetryBackoff time.Duration `json:"retry_backoff" yaml:"retry_backoff"`

	// RateLimits contains rate limiting configuration
	RateLimits RateLimitConfig `json:"rate_limits" yaml:"rate_limits"`

	// Jurisdictions lists enabled jurisdictions
	Jurisdictions []string `json:"jurisdictions" yaml:"jurisdictions"`

	// DefaultRetention contains default retention settings
	DefaultRetention RetentionPolicy `json:"default_retention" yaml:"default_retention"`

	// RequireConsent requires explicit user consent
	RequireConsent bool `json:"require_consent" yaml:"require_consent"`

	// ConsentDuration is how long consent remains valid
	ConsentDuration time.Duration `json:"consent_duration" yaml:"consent_duration"`

	// VEIDIntegration contains VEID integration settings
	VEIDIntegration VEIDIntegrationConfig `json:"veid_integration" yaml:"veid_integration"`

	// Audit contains audit configuration
	Audit AuditConfig `json:"audit" yaml:"audit"`

	// Adapters contains per-adapter configuration
	Adapters map[string]AdapterConfig `json:"adapters" yaml:"adapters"`

	// TLS contains TLS configuration for secure communication
	TLS TLSConfig `json:"tls" yaml:"tls"`

	// HealthCheck contains health check configuration
	HealthCheck HealthCheckConfig `json:"health_check" yaml:"health_check"`

	// Metrics contains metrics configuration
	Metrics MetricsConfig `json:"metrics" yaml:"metrics"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	// Enabled indicates if rate limiting is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// RequestsPerMinute is the max requests per minute per wallet
	RequestsPerMinute int `json:"requests_per_minute" yaml:"requests_per_minute"`

	// RequestsPerHour is the max requests per hour per wallet
	RequestsPerHour int `json:"requests_per_hour" yaml:"requests_per_hour"`

	// RequestsPerDay is the max requests per day per wallet
	RequestsPerDay int `json:"requests_per_day" yaml:"requests_per_day"`

	// BurstSize is the allowed burst size
	BurstSize int `json:"burst_size" yaml:"burst_size"`
}

// VEIDIntegrationConfig contains VEID integration settings
type VEIDIntegrationConfig struct {
	// Enabled indicates if VEID integration is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// BaseScoreContribution is the base score contribution for verified docs
	BaseScoreContribution float64 `json:"base_score_contribution" yaml:"base_score_contribution"`

	// MultiSourceBonus is the bonus for multi-source verification
	MultiSourceBonus float64 `json:"multi_source_bonus" yaml:"multi_source_bonus"`

	// GovernmentSourceWeight is the weight for government sources
	GovernmentSourceWeight float64 `json:"government_source_weight" yaml:"government_source_weight"`

	// VerificationFreshnessDecay is how quickly verification freshness decays
	VerificationFreshnessDecay time.Duration `json:"verification_freshness_decay" yaml:"verification_freshness_decay"`

	// MinConfidenceThreshold is the minimum confidence for score contribution
	MinConfidenceThreshold float64 `json:"min_confidence_threshold" yaml:"min_confidence_threshold"`

	// ScopeExpiryDuration is how long verification scopes remain valid
	ScopeExpiryDuration time.Duration `json:"scope_expiry_duration" yaml:"scope_expiry_duration"`
}

// AuditConfig contains audit configuration
type AuditConfig struct {
	// Enabled indicates if audit logging is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// LogPath is the path to audit logs
	LogPath string `json:"log_path" yaml:"log_path"`

	// EncryptLogs indicates if logs should be encrypted
	EncryptLogs bool `json:"encrypt_logs" yaml:"encrypt_logs"`

	// RetentionDays is how long to retain audit logs
	RetentionDays int `json:"retention_days" yaml:"retention_days"`

	// CompressLogs indicates if old logs should be compressed
	CompressLogs bool `json:"compress_logs" yaml:"compress_logs"`

	// ExportEnabled indicates if audit export is enabled
	ExportEnabled bool `json:"export_enabled" yaml:"export_enabled"`

	// ExportFormat is the export format (json, csv)
	ExportFormat string `json:"export_format" yaml:"export_format"`

	// AlertOnFailure indicates if alerts should be sent on failures
	AlertOnFailure bool `json:"alert_on_failure" yaml:"alert_on_failure"`
}

// AdapterConfig contains per-adapter configuration
type AdapterConfig struct {
	// Type is the adapter type
	Type DataSourceType `json:"type" yaml:"type"`

	// Jurisdiction is the jurisdiction served
	Jurisdiction string `json:"jurisdiction" yaml:"jurisdiction"`

	// Enabled indicates if adapter is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Endpoint is the API endpoint
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	// APIKey is the API key (should be from secret store)
	APIKey string `json:"api_key" yaml:"api_key"`

	// APISecret is the API secret (should be from secret store)
	APISecret string `json:"api_secret" yaml:"api_secret"`

	// Timeout is the request timeout
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// RateLimit is the adapter-specific rate limit
	RateLimit int `json:"rate_limit" yaml:"rate_limit"`

	// SupportedDocuments lists supported document types
	SupportedDocuments []DocumentType `json:"supported_documents" yaml:"supported_documents"`

	// HealthCheckEndpoint is the health check endpoint
	HealthCheckEndpoint string `json:"health_check_endpoint" yaml:"health_check_endpoint"`

	// Headers contains custom headers
	Headers map[string]string `json:"headers" yaml:"headers"`
}

// TLSConfig contains TLS configuration
type TLSConfig struct {
	// Enabled indicates if TLS is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// CertPath is the path to the certificate
	CertPath string `json:"cert_path" yaml:"cert_path"`

	// KeyPath is the path to the private key
	KeyPath string `json:"key_path" yaml:"key_path"`

	// CAPath is the path to the CA certificate
	CAPath string `json:"ca_path" yaml:"ca_path"`

	// InsecureSkipVerify disables certificate verification (NEVER use in production)
	InsecureSkipVerify bool `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`

	// MinVersion is the minimum TLS version
	MinVersion string `json:"min_version" yaml:"min_version"`
}

// HealthCheckConfig contains health check configuration
type HealthCheckConfig struct {
	// Enabled indicates if health checks are enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Interval is the health check interval
	Interval time.Duration `json:"interval" yaml:"interval"`

	// Timeout is the health check timeout
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// FailureThreshold is failures before marking unhealthy
	FailureThreshold int `json:"failure_threshold" yaml:"failure_threshold"`

	// RecoveryThreshold is successes before marking healthy
	RecoveryThreshold int `json:"recovery_threshold" yaml:"recovery_threshold"`
}

// MetricsConfig contains metrics configuration
type MetricsConfig struct {
	// Enabled indicates if metrics are enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Namespace is the metrics namespace
	Namespace string `json:"namespace" yaml:"namespace"`

	// Subsystem is the metrics subsystem
	Subsystem string `json:"subsystem" yaml:"subsystem"`
}

// ============================================================================
// Default Configuration
// ============================================================================

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Enabled:        false,
		ServiceID:      "govdata-service",
		DefaultTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryBackoff:   time.Second,
		RateLimits: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 10,
			RequestsPerHour:   100,
			RequestsPerDay:    500,
			BurstSize:         5,
		},
		Jurisdictions: []string{},
		DefaultRetention: RetentionPolicy{
			ResultRetentionDays:   90,
			AuditLogRetentionDays: 365,
			ConsentRetentionDays:  365 * 7, // 7 years
			AutoPurge:             true,
		},
		RequireConsent:  true,
		ConsentDuration: 365 * 24 * time.Hour, // 1 year
		VEIDIntegration: VEIDIntegrationConfig{
			Enabled:                    true,
			BaseScoreContribution:      0.25,
			MultiSourceBonus:           0.1,
			GovernmentSourceWeight:     1.5,
			VerificationFreshnessDecay: 180 * 24 * time.Hour, // 6 months
			MinConfidenceThreshold:     0.7,
			ScopeExpiryDuration:        365 * 24 * time.Hour, // 1 year
		},
		Audit: AuditConfig{
			Enabled:        true,
			LogPath:        "/var/log/govdata/audit.log",
			EncryptLogs:    true,
			RetentionDays:  365 * 7, // 7 years
			CompressLogs:   true,
			ExportEnabled:  true,
			ExportFormat:   "json",
			AlertOnFailure: true,
		},
		Adapters: make(map[string]AdapterConfig),
		TLS: TLSConfig{
			Enabled:            true,
			MinVersion:         "1.2",
			InsecureSkipVerify: false,
		},
		HealthCheck: HealthCheckConfig{
			Enabled:           true,
			Interval:          60 * time.Second,
			Timeout:           10 * time.Second,
			FailureThreshold:  3,
			RecoveryThreshold: 2,
		},
		Metrics: MetricsConfig{
			Enabled:   true,
			Namespace: "virtengine",
			Subsystem: "govdata",
		},
	}
}

// ============================================================================
// Validation
// ============================================================================

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ServiceID == "" {
		return errors.New("service_id is required")
	}

	if c.DefaultTimeout <= 0 {
		return errors.New("default_timeout must be positive")
	}

	if c.MaxRetries < 0 {
		return errors.New("max_retries cannot be negative")
	}

	if c.RateLimits.Enabled {
		if c.RateLimits.RequestsPerMinute <= 0 {
			return errors.New("requests_per_minute must be positive when rate limiting is enabled")
		}
	}

	if c.RequireConsent && c.ConsentDuration <= 0 {
		return errors.New("consent_duration must be positive when consent is required")
	}

	if c.VEIDIntegration.Enabled {
		if c.VEIDIntegration.MinConfidenceThreshold < 0 || c.VEIDIntegration.MinConfidenceThreshold > 1 {
			return errors.New("min_confidence_threshold must be between 0 and 1")
		}
	}

	if c.Audit.Enabled && c.Audit.RetentionDays <= 0 {
		return errors.New("audit retention_days must be positive when audit is enabled")
	}

	// Validate adapters
	for name, adapter := range c.Adapters {
		if err := adapter.Validate(); err != nil {
			return errors.New("adapter " + name + ": " + err.Error())
		}
	}

	return nil
}

// Validate validates adapter configuration
func (c *AdapterConfig) Validate() error {
	if !c.Type.IsValid() {
		return errors.New("invalid adapter type")
	}

	if c.Jurisdiction == "" {
		return errors.New("jurisdiction is required")
	}

	if c.Enabled && c.Endpoint == "" {
		return errors.New("endpoint is required when adapter is enabled")
	}

	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}

	return nil
}

// ============================================================================
// Jurisdiction Configuration
// ============================================================================

// JurisdictionConfig contains jurisdiction-specific configuration
type JurisdictionConfig struct {
	// Jurisdictions maps jurisdiction codes to configurations
	Jurisdictions map[string]Jurisdiction `json:"jurisdictions" yaml:"jurisdictions"`
}

// DefaultJurisdictions returns default jurisdiction configurations
func DefaultJurisdictions() JurisdictionConfig {
	return JurisdictionConfig{
		Jurisdictions: map[string]Jurisdiction{
			"US": {
				Code:    "US",
				Name:    "United States",
				Country: "US",
				SupportedDocuments: []DocumentType{
					DocumentTypeDriversLicense,
					DocumentTypeStateID,
					DocumentTypePassport,
					DocumentTypeBirthCertificate,
					DocumentTypeTaxID,
				},
				DataSources: []DataSourceType{
					DataSourceDMV,
					DataSourcePassport,
					DataSourceVitalRecords,
					DataSourceTaxAuthority,
				},
				RetentionPolicy: RetentionPolicy{
					ResultRetentionDays:   90,
					AuditLogRetentionDays: 365 * 7,
					ConsentRetentionDays:  365 * 7,
					AutoPurge:             true,
				},
				GDPRApplicable:  false,
				CCPAApplicable:  true,
				Active:          true,
				RequiresConsent: true,
			},
			"US-CA": {
				Code:        "US-CA",
				Name:        "California, United States",
				Country:     "US",
				Subdivision: "CA",
				SupportedDocuments: []DocumentType{
					DocumentTypeDriversLicense,
					DocumentTypeStateID,
				},
				DataSources: []DataSourceType{
					DataSourceDMV,
				},
				RetentionPolicy: RetentionPolicy{
					ResultRetentionDays:   90,
					AuditLogRetentionDays: 365 * 7,
					ConsentRetentionDays:  365 * 7,
					AutoPurge:             true,
				},
				GDPRApplicable:  false,
				CCPAApplicable:  true,
				Active:          true,
				RequiresConsent: true,
			},
			"EU": {
				Code:    "EU",
				Name:    "European Union",
				Country: "EU",
				SupportedDocuments: []DocumentType{
					DocumentTypeNationalID,
					DocumentTypePassport,
					DocumentTypeResidencePermit,
				},
				DataSources: []DataSourceType{
					DataSourceNationalRegistry,
					DataSourcePassport,
					DataSourceImmigration,
				},
				RetentionPolicy: RetentionPolicy{
					ResultRetentionDays:   30, // GDPR minimum
					AuditLogRetentionDays: 365,
					ConsentRetentionDays:  365,
					AutoPurge:             true,
				},
				GDPRApplicable:  true,
				CCPAApplicable:  false,
				Active:          true,
				RequiresConsent: true,
			},
			"GB": {
				Code:    "GB",
				Name:    "United Kingdom",
				Country: "GB",
				SupportedDocuments: []DocumentType{
					DocumentTypePassport,
					DocumentTypeDriversLicense,
					DocumentTypeNationalID,
				},
				DataSources: []DataSourceType{
					DataSourcePassport,
					DataSourceDMV,
					DataSourceNationalRegistry,
				},
				RetentionPolicy: RetentionPolicy{
					ResultRetentionDays:   30,
					AuditLogRetentionDays: 365,
					ConsentRetentionDays:  365,
					AutoPurge:             true,
				},
				GDPRApplicable:  true, // UK GDPR
				CCPAApplicable:  false,
				Active:          true,
				RequiresConsent: true,
			},
			"AU": {
				Code:    "AU",
				Name:    "Australia",
				Country: "AU",
				SupportedDocuments: []DocumentType{
					DocumentTypePassport,
					DocumentTypeDriversLicense,
					DocumentTypeBirthCertificate,
				},
				DataSources: []DataSourceType{
					DataSourcePassport,
					DataSourceDMV,
					DataSourceVitalRecords,
				},
				RetentionPolicy: RetentionPolicy{
					ResultRetentionDays:   90,
					AuditLogRetentionDays: 365 * 7,
					ConsentRetentionDays:  365 * 7,
					AutoPurge:             true,
				},
				GDPRApplicable:  false,
				CCPAApplicable:  false,
				Active:          true,
				RequiresConsent: true,
			},
		},
	}
}
