// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"time"

	"github.com/virtengine/virtengine/pkg/payment"
)

// ============================================================================
// Main Configuration
// ============================================================================

// Config contains the configuration for the off-ramp service
type Config struct {
	// Enabled enables off-ramp functionality
	Enabled bool `json:"enabled"`

	// DefaultProvider is the default payout provider
	DefaultProvider ProviderType `json:"default_provider"`

	// PayPalConfig is configuration for PayPal Payouts
	PayPalConfig PayPalConfig `json:"paypal_config,omitempty"`

	// ACHConfig is configuration for ACH transfers
	ACHConfig ACHConfig `json:"ach_config,omitempty"`

	// KYCConfig is configuration for KYC verification
	KYCConfig KYCConfig `json:"kyc_config"`

	// AMLConfig is configuration for AML screening
	AMLConfig AMLConfig `json:"aml_config"`

	// LimitsConfig contains payout limit configuration
	LimitsConfig LimitsConfig `json:"limits_config"`

	// ReconciliationConfig contains reconciliation settings
	ReconciliationConfig ReconciliationConfig `json:"reconciliation_config"`

	// WebhookConfig contains webhook settings
	WebhookConfig WebhookConfig `json:"webhook_config"`

	// RetryConfig contains retry/backoff settings
	RetryConfig RetryConfig `json:"retry_config"`

	// ConversionConfig contains conversion settings
	ConversionConfig ConversionConfig `json:"conversion_config"`

	// SupportedCurrencies lists supported fiat currencies
	SupportedCurrencies []payment.Currency `json:"supported_currencies"`

	// QuoteValiditySeconds is how long quotes are valid
	QuoteValiditySeconds int `json:"quote_validity_seconds"`

	// EnableSandbox enables sandbox/test mode
	EnableSandbox bool `json:"enable_sandbox"`

	// EnableLogging enables debug logging
	EnableLogging bool `json:"enable_logging"`
}

// ============================================================================
// PayPal Configuration
// ============================================================================

// PayPalConfig contains PayPal-specific configuration
type PayPalConfig struct {
	// ClientID is the PayPal client ID
	ClientID string `json:"client_id"`

	// ClientSecret is the PayPal client secret
	ClientSecret string `json:"client_secret"`

	// WebhookID is the PayPal webhook ID
	WebhookID string `json:"webhook_id"`

	// Environment is "sandbox" or "live"
	Environment string `json:"environment"`

	// BaseURL is the PayPal API base URL
	BaseURL string `json:"base_url,omitempty"`

	// EmailSubject is the email subject for payout notifications
	EmailSubject string `json:"email_subject,omitempty"`

	// EmailMessage is the email message for payout notifications
	EmailMessage string `json:"email_message,omitempty"`

	// SenderBatchIDPrefix is a prefix for batch IDs
	SenderBatchIDPrefix string `json:"sender_batch_id_prefix,omitempty"`
}

// GetBaseURL returns the appropriate base URL
func (c PayPalConfig) GetBaseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	if c.Environment == "live" {
		return "https://api-m.paypal.com"
	}
	return "https://api-m.sandbox.paypal.com"
}

// ============================================================================
// ACH Configuration
// ============================================================================

// ACHConfig contains ACH-specific configuration
type ACHConfig struct {
	// Provider is the ACH provider (e.g., "stripe", "plaid")
	Provider string `json:"provider"`

	// SecretKey is the provider secret key
	SecretKey string `json:"secret_key"`

	// WebhookSecret is the webhook signing secret
	WebhookSecret string `json:"webhook_secret"`

	// SourceAccountID is the source account ID
	SourceAccountID string `json:"source_account_id"`

	// ProcessingDays is the expected processing time in days
	ProcessingDays int `json:"processing_days"`

	// EnableSameDayACH enables same-day ACH (higher fees)
	EnableSameDayACH bool `json:"enable_same_day_ach"`

	// Environment is "sandbox" or "live"
	Environment string `json:"environment"`
}

// ============================================================================
// KYC Configuration
// ============================================================================

// KYCConfig contains KYC verification configuration
type KYCConfig struct {
	// Enabled enables KYC verification requirement
	Enabled bool `json:"enabled"`

	// RequiredLevel is the minimum verification level required
	RequiredLevel KYCVerificationLevel `json:"required_level"`

	// Provider is the KYC provider (e.g., "veid", "jumio", "onfido")
	Provider string `json:"provider"`

	// RevalidationDays is how often to revalidate KYC (0 = never)
	RevalidationDays int `json:"revalidation_days"`

	// AllowPendingVerification allows payouts during pending verification
	AllowPendingVerification bool `json:"allow_pending_verification"`

	// MaxPendingAmount is the max amount for pending verification
	MaxPendingAmount int64 `json:"max_pending_amount"`
}

// ============================================================================
// AML Configuration
// ============================================================================

// AMLConfig contains AML screening configuration
type AMLConfig struct {
	// Enabled enables AML screening
	Enabled bool `json:"enabled"`

	// Provider is the AML screening provider
	Provider string `json:"provider"`

	// APIURL is the AML provider API URL
	APIURL string `json:"api_url,omitempty"`

	// APIKey is the AML provider API key
	APIKey string `json:"api_key,omitempty"`

	// ScreenSanctions enables sanctions list screening
	ScreenSanctions bool `json:"screen_sanctions"`

	// ScreenPEP enables politically exposed persons screening
	ScreenPEP bool `json:"screen_pep"`

	// ScreenAdverseMedia enables adverse media screening
	ScreenAdverseMedia bool `json:"screen_adverse_media"`

	// RiskThreshold is the risk score threshold (0-100)
	RiskThreshold int `json:"risk_threshold"`

	// AutoApproveBelow auto-approves scores below this threshold
	AutoApproveBelow int `json:"auto_approve_below"`
}

// ============================================================================
// Limits Configuration
// ============================================================================

// LimitsConfig contains payout limit configuration
type LimitsConfig struct {
	// DefaultDailyLimit is the default daily limit (minor units)
	DefaultDailyLimit int64 `json:"default_daily_limit"`

	// DefaultMonthlyLimit is the default monthly limit (minor units)
	DefaultMonthlyLimit int64 `json:"default_monthly_limit"`

	// DefaultPerTransactionLimit is the default per-transaction limit
	DefaultPerTransactionLimit int64 `json:"default_per_transaction_limit"`

	// MinPayoutAmount is the minimum payout amount
	MinPayoutAmount map[payment.Currency]int64 `json:"min_payout_amount"`

	// MaxPayoutAmount is the maximum payout amount
	MaxPayoutAmount map[payment.Currency]int64 `json:"max_payout_amount"`

	// TierLimits defines limits by verification tier
	TierLimits map[KYCVerificationLevel]TierLimit `json:"tier_limits,omitempty"`
}

// TierLimit defines limits for a verification tier
type TierLimit struct {
	// DailyLimit is the daily limit for this tier
	DailyLimit int64 `json:"daily_limit"`

	// MonthlyLimit is the monthly limit for this tier
	MonthlyLimit int64 `json:"monthly_limit"`

	// PerTransactionLimit is the per-transaction limit
	PerTransactionLimit int64 `json:"per_transaction_limit"`
}

// ============================================================================
// Reconciliation Configuration
// ============================================================================

// ReconciliationConfig contains reconciliation settings
type ReconciliationConfig struct {
	// Enabled enables automatic reconciliation
	Enabled bool `json:"enabled"`

	// IntervalMinutes is how often to run reconciliation
	IntervalMinutes int `json:"interval_minutes"`

	// MaxAgeDays is the max age of payouts to reconcile
	MaxAgeDays int `json:"max_age_days"`

	// DiscrepancyThreshold is the amount threshold for flagging
	DiscrepancyThreshold int64 `json:"discrepancy_threshold"`

	// AutoResolveMatches auto-resolves matching records
	AutoResolveMatches bool `json:"auto_resolve_matches"`

	// AlertOnMismatch enables alerts on mismatches
	AlertOnMismatch bool `json:"alert_on_mismatch"`
}

// ============================================================================
// Webhook Configuration
// ============================================================================

// WebhookConfig contains webhook handling configuration
type WebhookConfig struct {
	// Enabled enables webhook handling
	Enabled bool `json:"enabled"`

	// Path is the webhook endpoint path
	Path string `json:"path"`

	// SignatureVerification enables signature checking
	SignatureVerification bool `json:"signature_verification"`

	// ToleranceSeconds is the timestamp tolerance
	ToleranceSeconds int `json:"tolerance_seconds"`

	// MaxRetries for webhook processing
	MaxRetries int `json:"max_retries"`
}

// ============================================================================
// Retry Configuration
// ============================================================================

// RetryConfig contains retry/backoff settings
type RetryConfig struct {
	// MaxAttempts is the maximum retry attempts
	MaxAttempts int `json:"max_attempts"`

	// InitialDelay is the initial retry delay
	InitialDelay time.Duration `json:"initial_delay"`

	// MaxDelay is the maximum retry delay
	MaxDelay time.Duration `json:"max_delay"`

	// BackoffFactor is the exponential backoff factor
	BackoffFactor float64 `json:"backoff_factor"`

	// RetryableErrors lists error codes that are retryable
	RetryableErrors []string `json:"retryable_errors,omitempty"`
}

// ============================================================================
// Conversion Configuration
// ============================================================================

// ConversionConfig contains crypto-to-fiat conversion settings
type ConversionConfig struct {
	// FeePercent is the conversion fee percentage
	FeePercent float64 `json:"fee_percent"`

	// FeeMinimum is the minimum fee in minor units
	FeeMinimum map[payment.Currency]int64 `json:"fee_minimum"`

	// FeeMaximum is the maximum fee in minor units
	FeeMaximum map[payment.Currency]int64 `json:"fee_maximum"`

	// SlippagePercent is the maximum slippage allowed
	SlippagePercent float64 `json:"slippage_percent"`

	// PriceFeedSource is the source for conversion rates
	PriceFeedSource string `json:"price_feed_source"`
}

// ============================================================================
// Default Configuration
// ============================================================================

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Enabled:         true,
		DefaultProvider: ProviderPayPal,
		PayPalConfig: PayPalConfig{
			Environment:         "sandbox",
			EmailSubject:        "VirtEngine Payout",
			EmailMessage:        "You have received a payout from VirtEngine",
			SenderBatchIDPrefix: "VE_",
		},
		ACHConfig: ACHConfig{
			Provider:         "stripe",
			ProcessingDays:   3,
			EnableSameDayACH: false,
			Environment:      "sandbox",
		},
		KYCConfig: KYCConfig{
			Enabled:                  true,
			RequiredLevel:            KYCLevelBasic,
			Provider:                 "veid",
			RevalidationDays:         365,
			AllowPendingVerification: false,
		},
		AMLConfig: AMLConfig{
			Enabled:            true,
			ScreenSanctions:    true,
			ScreenPEP:          true,
			ScreenAdverseMedia: false,
			RiskThreshold:      70,
			AutoApproveBelow:   30,
		},
		LimitsConfig: LimitsConfig{
			DefaultDailyLimit:          100000000, // $1,000,000.00
			DefaultMonthlyLimit:        500000000, // $5,000,000.00
			DefaultPerTransactionLimit: 10000000,  // $100,000.00
			MinPayoutAmount: map[payment.Currency]int64{
				payment.CurrencyUSD: 100,   // $1.00
				payment.CurrencyEUR: 100,   // €1.00
				payment.CurrencyGBP: 100,   // £1.00
			},
			MaxPayoutAmount: map[payment.Currency]int64{
				payment.CurrencyUSD: 10000000, // $100,000.00
				payment.CurrencyEUR: 10000000, // €100,000.00
				payment.CurrencyGBP: 10000000, // £100,000.00
			},
			TierLimits: map[KYCVerificationLevel]TierLimit{
				KYCLevelBasic: {
					DailyLimit:          100000,   // $1,000.00
					MonthlyLimit:        500000,   // $5,000.00
					PerTransactionLimit: 50000,    // $500.00
				},
				KYCLevelEnhanced: {
					DailyLimit:          1000000,  // $10,000.00
					MonthlyLimit:        5000000,  // $50,000.00
					PerTransactionLimit: 500000,   // $5,000.00
				},
				KYCLevelFull: {
					DailyLimit:          10000000,  // $100,000.00
					MonthlyLimit:        50000000,  // $500,000.00
					PerTransactionLimit: 5000000,   // $50,000.00
				},
			},
		},
		ReconciliationConfig: ReconciliationConfig{
			Enabled:              true,
			IntervalMinutes:      60,
			MaxAgeDays:           30,
			DiscrepancyThreshold: 100, // $1.00
			AutoResolveMatches:   true,
			AlertOnMismatch:      true,
		},
		WebhookConfig: WebhookConfig{
			Enabled:               true,
			Path:                  "/webhooks/offramp",
			SignatureVerification: true,
			ToleranceSeconds:      300,
			MaxRetries:            3,
		},
		RetryConfig: RetryConfig{
			MaxAttempts:   3,
			InitialDelay:  1 * time.Second,
			MaxDelay:      30 * time.Second,
			BackoffFactor: 2.0,
			RetryableErrors: []string{
				"TIMEOUT",
				"RATE_LIMIT",
				"SERVICE_UNAVAILABLE",
			},
		},
		ConversionConfig: ConversionConfig{
			FeePercent:      1.5,
			FeeMinimum: map[payment.Currency]int64{
				payment.CurrencyUSD: 25,  // $0.25
				payment.CurrencyEUR: 25,
				payment.CurrencyGBP: 25,
			},
			FeeMaximum: map[payment.Currency]int64{
				payment.CurrencyUSD: 10000, // $100.00
				payment.CurrencyEUR: 10000,
				payment.CurrencyGBP: 10000,
			},
			SlippagePercent: 0.5,
			PriceFeedSource: "coingecko",
		},
		SupportedCurrencies: []payment.Currency{
			payment.CurrencyUSD,
			payment.CurrencyEUR,
			payment.CurrencyGBP,
		},
		QuoteValiditySeconds: 60,
		EnableSandbox:        true,
		EnableLogging:        false,
	}
}

// Validate validates the configuration
func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if !c.DefaultProvider.IsValid() {
		return ErrProviderNotConfigured
	}

	switch c.DefaultProvider {
	case ProviderPayPal:
		if c.PayPalConfig.ClientID == "" || c.PayPalConfig.ClientSecret == "" {
			return ErrProviderNotConfigured
		}
	case ProviderACH:
		if c.ACHConfig.SecretKey == "" {
			return ErrProviderNotConfigured
		}
	}

	if len(c.SupportedCurrencies) == 0 {
		return payment.ErrInvalidCurrency
	}

	return nil
}

// IsCurrencySupported checks if a currency is supported
func (c Config) IsCurrencySupported(currency payment.Currency) bool {
	for _, curr := range c.SupportedCurrencies {
		if curr == currency {
			return true
		}
	}
	return false
}

// GetLimitsForTier returns limits for a verification tier
func (c Config) GetLimitsForTier(tier KYCVerificationLevel) TierLimit {
	if limits, ok := c.LimitsConfig.TierLimits[tier]; ok {
		return limits
	}
	// Return default limits
	return TierLimit{
		DailyLimit:          c.LimitsConfig.DefaultDailyLimit,
		MonthlyLimit:        c.LimitsConfig.DefaultMonthlyLimit,
		PerTransactionLimit: c.LimitsConfig.DefaultPerTransactionLimit,
	}
}

// ValidatePayoutAmount validates a payout amount
func (c Config) ValidatePayoutAmount(amount payment.Amount) error {
	if !c.IsCurrencySupported(amount.Currency) {
		return payment.ErrInvalidCurrency
	}

	if min, ok := c.LimitsConfig.MinPayoutAmount[amount.Currency]; ok && amount.Value < min {
		return ErrPayoutAmountBelowMinimum
	}

	if max, ok := c.LimitsConfig.MaxPayoutAmount[amount.Currency]; ok && amount.Value > max {
		return ErrPayoutAmountAboveMaximum
	}

	return nil
}
