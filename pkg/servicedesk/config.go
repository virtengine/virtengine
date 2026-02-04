package servicedesk

import (
	"fmt"
	"time"

	"cosmossdk.io/errors"
)

// Config holds the bridge service configuration
type Config struct {
	// Enabled enables the bridge service
	Enabled bool `json:"enabled"`

	// JiraConfig is the Jira configuration
	JiraConfig *JiraConfig `json:"jira,omitempty"`

	// WaldurConfig is the Waldur configuration
	WaldurConfig *WaldurConfig `json:"waldur,omitempty"`

	// MappingSchema is the field mapping schema
	MappingSchema *MappingSchema `json:"mapping_schema,omitempty"`

	// SyncConfig is the sync configuration
	SyncConfig SyncConfig `json:"sync"`

	// RetryConfig is the retry configuration
	RetryConfig RetryConfig `json:"retry"`

	// WebhookConfig is the webhook configuration
	WebhookConfig WebhookServerConfig `json:"webhook"`

	// AuditConfig is the audit configuration
	AuditConfig AuditConfig `json:"audit"`

	// Decryption config for encrypted payloads
	Decryption *DecryptionConfig `json:"decryption,omitempty"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:       false,
		MappingSchema: DefaultMappingSchema(),
		SyncConfig:    DefaultSyncConfig(),
		RetryConfig:   DefaultRetryConfig(),
		WebhookConfig: DefaultWebhookServerConfig(),
		AuditConfig:   DefaultAuditConfig(),
		Decryption:    nil,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.JiraConfig == nil && c.WaldurConfig == nil {
		return fmt.Errorf("at least one service desk must be configured")
	}

	if c.JiraConfig != nil {
		if err := c.JiraConfig.Validate(); err != nil {
			return fmt.Errorf("jira config: %w", err)
		}
	}

	if c.WaldurConfig != nil {
		if err := c.WaldurConfig.Validate(); err != nil {
			return fmt.Errorf("waldur config: %w", err)
		}
	}

	if err := c.SyncConfig.Validate(); err != nil {
		return fmt.Errorf("sync config: %w", err)
	}

	if err := c.RetryConfig.Validate(); err != nil {
		return fmt.Errorf("retry config: %w", err)
	}

	if c.Decryption != nil {
		if _, err := c.Decryption.LoadPrivateKey(); err != nil {
			return fmt.Errorf("decryption config: %w", err)
		}
	}

	return nil
}

// JiraConfig holds Jira configuration
type JiraConfig struct {
	// BaseURL is the Jira instance URL
	BaseURL string `json:"base_url"`

	// Username is the Jira username (for basic auth)
	Username string `json:"username"`

	// APIToken is the API token (CRITICAL: never log)
	APIToken string `json:"-"`

	// ProjectKey is the Jira project key
	ProjectKey string `json:"project_key"`

	// IssueType is the default issue type
	IssueType string `json:"issue_type"`

	// WebhookSecret is the webhook secret (CRITICAL: never log)
	WebhookSecret string `json:"-"`

	// Timeout is the API timeout
	Timeout time.Duration `json:"timeout"`
}

// Validate validates the Jira configuration
func (c *JiraConfig) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.APIToken == "" {
		return fmt.Errorf("api_token is required")
	}
	if c.ProjectKey == "" {
		return fmt.Errorf("project_key is required")
	}
	return nil
}

// WaldurConfig holds Waldur configuration
type WaldurConfig struct {
	// BaseURL is the Waldur API URL
	BaseURL string `json:"base_url"`

	// Token is the API token (CRITICAL: never log)
	Token string `json:"-"`

	// OrganizationUUID is the Waldur organization UUID
	OrganizationUUID string `json:"organization_uuid"`

	// ProjectUUID is the Waldur project UUID
	ProjectUUID string `json:"project_uuid"`

	// WebhookSecret is the webhook secret (CRITICAL: never log)
	WebhookSecret string `json:"-"`

	// Timeout is the API timeout
	Timeout time.Duration `json:"timeout"`
}

// Validate validates the Waldur configuration
func (c *WaldurConfig) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if c.Token == "" {
		return fmt.Errorf("token is required")
	}
	if c.OrganizationUUID == "" {
		return fmt.Errorf("organization_uuid is required")
	}
	return nil
}

// SyncConfig holds sync behavior configuration
type SyncConfig struct {
	// SyncInterval is the interval between sync cycles
	SyncInterval time.Duration `json:"sync_interval"`

	// BatchSize is the maximum events per sync batch
	BatchSize int `json:"batch_size"`

	// ConflictResolution is the default conflict resolution strategy
	ConflictResolution ConflictResolution `json:"conflict_resolution"`

	// EnableInbound enables inbound sync (external to on-chain)
	EnableInbound bool `json:"enable_inbound"`

	// EnableOutbound enables outbound sync (on-chain to external)
	EnableOutbound bool `json:"enable_outbound"`

	// SyncAttachments enables attachment sync
	SyncAttachments bool `json:"sync_attachments"`
}

// DefaultSyncConfig returns a default sync configuration
func DefaultSyncConfig() SyncConfig {
	return SyncConfig{
		SyncInterval:       30 * time.Second,
		BatchSize:          50,
		ConflictResolution: ConflictResolutionOnChainWins,
		EnableInbound:      true,
		EnableOutbound:     true,
		SyncAttachments:    true,
	}
}

// Validate validates the sync configuration
func (c *SyncConfig) Validate() error {
	if c.SyncInterval < 5*time.Second {
		return fmt.Errorf("sync_interval must be at least 5 seconds")
	}
	if c.BatchSize < 1 {
		return fmt.Errorf("batch_size must be at least 1")
	}
	if c.BatchSize > 500 {
		return fmt.Errorf("batch_size must not exceed 500")
	}
	return nil
}

// RetryConfig holds retry behavior configuration
type RetryConfig struct {
	// MaxRetries is the maximum retry attempts
	MaxRetries int `json:"max_retries"`

	// InitialBackoff is the initial backoff duration
	InitialBackoff time.Duration `json:"initial_backoff"`

	// MaxBackoff is the maximum backoff duration
	MaxBackoff time.Duration `json:"max_backoff"`

	// BackoffMultiplier is the backoff multiplier
	BackoffMultiplier float64 `json:"backoff_multiplier"`

	// RetryableStatusCodes are HTTP status codes that should be retried
	RetryableStatusCodes []int `json:"retryable_status_codes"`
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:           5,
		InitialBackoff:       1 * time.Second,
		MaxBackoff:           5 * time.Minute,
		BackoffMultiplier:    2.0,
		RetryableStatusCodes: []int{429, 500, 502, 503, 504},
	}
}

// Validate validates the retry configuration
func (c *RetryConfig) Validate() error {
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}
	if c.InitialBackoff < 0 {
		return fmt.Errorf("initial_backoff cannot be negative")
	}
	if c.MaxBackoff < c.InitialBackoff {
		return fmt.Errorf("max_backoff must be >= initial_backoff")
	}
	if c.BackoffMultiplier < 1 {
		return fmt.Errorf("backoff_multiplier must be >= 1")
	}
	return nil
}

// CalculateBackoff calculates the backoff duration for a given attempt
func (c *RetryConfig) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return c.InitialBackoff
	}

	backoff := float64(c.InitialBackoff)
	for i := 0; i < attempt; i++ {
		backoff *= c.BackoffMultiplier
		if backoff > float64(c.MaxBackoff) {
			backoff = float64(c.MaxBackoff)
			break
		}
	}
	return time.Duration(backoff)
}

// IsRetryable checks if an HTTP status code is retryable
func (c *RetryConfig) IsRetryable(statusCode int) bool {
	for _, code := range c.RetryableStatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

// WebhookServerConfig holds webhook server configuration
type WebhookServerConfig struct {
	// Enabled enables the webhook server
	Enabled bool `json:"enabled"`

	// ListenAddr is the address to listen on
	ListenAddr string `json:"listen_addr"`

	// PathPrefix is the webhook path prefix
	PathPrefix string `json:"path_prefix"`

	// RequireSignature requires signature verification
	RequireSignature bool `json:"require_signature"`

	// AllowedIPs are allowed source IP addresses
	AllowedIPs []string `json:"allowed_ips,omitempty"`

	// RateLimitPerSecond is the rate limit per second
	RateLimitPerSecond int `json:"rate_limit_per_second"`
}

// DefaultWebhookServerConfig returns a default webhook server configuration
func DefaultWebhookServerConfig() WebhookServerConfig {
	return WebhookServerConfig{
		Enabled:            true,
		ListenAddr:         ":8480",
		PathPrefix:         "/webhooks",
		RequireSignature:   true,
		RateLimitPerSecond: 100,
	}
}

// AuditConfig holds audit configuration
type AuditConfig struct {
	// Enabled enables audit logging
	Enabled bool `json:"enabled"`

	// LogLevel is the audit log level
	LogLevel string `json:"log_level"`

	// RetentionDays is the audit log retention period
	RetentionDays int `json:"retention_days"`

	// LogSensitive enables logging of sensitive operations
	LogSensitive bool `json:"log_sensitive"`
}

// DefaultAuditConfig returns a default audit configuration
func DefaultAuditConfig() AuditConfig {
	return AuditConfig{
		Enabled:       true,
		LogLevel:      "info",
		RetentionDays: 90,
		LogSensitive:  false,
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	cfg := DefaultConfig()

	// This is a placeholder - actual implementation would read from env
	// and secrets management system (Vault, etc.)

	return cfg, nil
}

// package-level errors using cosmossdk error pattern
var (
	// ErrConfigInvalid indicates an invalid configuration
	ErrConfigInvalid = errors.Register("servicedesk", 1, "invalid configuration")

	// ErrSyncFailed indicates a sync operation failed
	ErrSyncFailed = errors.Register("servicedesk", 2, "sync failed")

	// ErrExternalAPIError indicates an external API error
	ErrExternalAPIError = errors.Register("servicedesk", 3, "external API error")

	// ErrConflict indicates a sync conflict
	ErrConflict = errors.Register("servicedesk", 4, "sync conflict detected")

	// ErrTicketNotFound indicates the ticket was not found
	ErrTicketNotFound = errors.Register("servicedesk", 5, "ticket not found")

	// ErrSignatureInvalid indicates an invalid callback signature
	ErrSignatureInvalid = errors.Register("servicedesk", 6, "invalid signature")

	// ErrRateLimited indicates rate limiting
	ErrRateLimited = errors.Register("servicedesk", 7, "rate limited")

	// ErrAttachmentFailed indicates attachment sync failed
	ErrAttachmentFailed = errors.Register("servicedesk", 8, "attachment sync failed")

	// ErrAuditFailed indicates audit logging failed
	ErrAuditFailed = errors.Register("servicedesk", 9, "audit logging failed")
)
