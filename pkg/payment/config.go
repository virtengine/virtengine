// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-906: Payment gateway integration for fiat-to-crypto onramp
package payment

import (
	"strings"
	"time"
)

// ============================================================================
// Configuration
// ============================================================================

// Config contains the configuration for the payment service
type Config struct {
	// Gateway is the primary payment gateway to use
	Gateway GatewayType `json:"gateway"`

	// StripeConfig is configuration for Stripe gateway
	StripeConfig StripeConfig `json:"stripe_config,omitempty"`

	// AdyenConfig is configuration for Adyen gateway
	AdyenConfig AdyenConfig `json:"adyen_config,omitempty"`

	// PayPalConfig is configuration for PayPal gateway
	PayPalConfig PayPalConfig `json:"paypal_config,omitempty"`

	// ACHConfig is configuration for ACH gateway
	ACHConfig ACHConfig `json:"ach_config,omitempty"`

	// WebhookConfig is configuration for webhook handling
	WebhookConfig WebhookConfig `json:"webhook_config"`

	// RateLimitConfig is configuration for rate limiting
	RateLimitConfig RateLimitConfig `json:"rate_limit_config"`

	// ConversionConfig is for fiat-to-crypto conversion
	ConversionConfig ConversionConfig `json:"conversion_config"`

	// Timeouts and retry settings
	RequestTimeout     time.Duration `json:"request_timeout"`
	RetryMaxAttempts   int           `json:"retry_max_attempts"`
	RetryInitialDelay  time.Duration `json:"retry_initial_delay"`
	RetryMaxDelay      time.Duration `json:"retry_max_delay"`
	RetryBackoffFactor float64       `json:"retry_backoff_factor"`

	// SupportedCurrencies is the list of supported fiat currencies
	SupportedCurrencies []Currency `json:"supported_currencies"`

	// MinAmount is the minimum transaction amount (in minor units)
	MinAmount map[Currency]int64 `json:"min_amount"`

	// MaxAmount is the maximum transaction amount (in minor units)
	MaxAmount map[Currency]int64 `json:"max_amount"`

	// DefaultStatementDescriptor for card statements
	DefaultStatementDescriptor string `json:"default_statement_descriptor"`

	// EnableSandbox enables sandbox/test mode
	EnableSandbox bool `json:"enable_sandbox"`

	// EnableLogging enables debug logging
	EnableLogging bool `json:"enable_logging"`
}

// StripeConfig contains Stripe-specific configuration
type StripeConfig struct {
	// SecretKey is the Stripe secret key (starts with sk_)
	SecretKey string `json:"secret_key"`

	// PublishableKey is the Stripe publishable key (starts with pk_)
	PublishableKey string `json:"publishable_key"`

	// WebhookSecret is the webhook signing secret
	WebhookSecret string `json:"webhook_secret"`

	// APIVersion is the Stripe API version to use
	APIVersion string `json:"api_version,omitempty"`

	// ConnectAccountID for Stripe Connect (optional)
	ConnectAccountID string `json:"connect_account_id,omitempty"`

	// EnablePaymentIntentsMigration uses PaymentIntents API
	EnablePaymentIntentsMigration bool `json:"enable_payment_intents_migration"`
}

// AdyenConfig contains Adyen-specific configuration
type AdyenConfig struct {
	// APIKey is the Adyen API key
	APIKey string `json:"api_key"`

	// MerchantAccount is the Adyen merchant account
	MerchantAccount string `json:"merchant_account"`

	// ClientKey is for client-side encryption
	ClientKey string `json:"client_key"`

	// HMACKey is for webhook signature verification
	HMACKey string `json:"hmac_key"`

	// Environment is "live" or "test"
	Environment string `json:"environment"`

	// APIPrefix is the API URL prefix for the environment
	APIPrefix string `json:"api_prefix,omitempty"`

	// LiveEndpointURLPrefix for live mode
	LiveEndpointURLPrefix string `json:"live_endpoint_url_prefix,omitempty"`
}

// PayPalConfig contains PayPal-specific configuration
type PayPalConfig struct {
	// ClientID is the PayPal client ID
	ClientID string `json:"client_id"`

	// ClientSecret is the PayPal client secret
	ClientSecret string `json:"client_secret"`

	// WebhookID is the PayPal webhook ID for signature verification
	WebhookID string `json:"webhook_id"`

	// Environment is "live" or "sandbox"
	Environment string `json:"environment"`

	// BaseURL overrides the API base URL (optional)
	BaseURL string `json:"base_url,omitempty"`
}

// GetBaseURL returns the PayPal API base URL
func (c PayPalConfig) GetBaseURL() string {
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	if strings.EqualFold(c.Environment, "live") {
		return "https://api-m.paypal.com"
	}
	return "https://api-m.sandbox.paypal.com"
}

// ACHConfig contains ACH-specific configuration
type ACHConfig struct {
	// Provider identifies the ACH provider (stripe, dwolla, custom)
	Provider string `json:"provider,omitempty"`

	// SecretKey is the API key for the ACH provider
	SecretKey string `json:"secret_key"`

	// WebhookSecret is the webhook signing secret
	WebhookSecret string `json:"webhook_secret"`

	// Environment is "live" or "sandbox"
	Environment string `json:"environment"`

	// BaseURL overrides the API base URL (optional)
	BaseURL string `json:"base_url,omitempty"`

	// NACHAOriginID is the origin ID for NACHA file generation
	NACHAOriginID string `json:"nacha_origin_id,omitempty"`

	// NACHACompanyName is the company name for NACHA file generation
	NACHACompanyName string `json:"nacha_company_name,omitempty"`

	// RetryMaxAttempts is the maximum retry attempts for ACH requests
	RetryMaxAttempts int `json:"retry_max_attempts,omitempty"`

	// RetryInitialDelay is the initial retry delay
	RetryInitialDelay time.Duration `json:"retry_initial_delay,omitempty"`

	// RetryMaxDelay is the maximum retry delay
	RetryMaxDelay time.Duration `json:"retry_max_delay,omitempty"`

	// RetryBackoffFactor is the exponential backoff factor
	RetryBackoffFactor float64 `json:"retry_backoff_factor,omitempty"`
}

// GetBaseURL returns the ACH API base URL
func (c ACHConfig) GetBaseURL() string {
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	if strings.EqualFold(c.Provider, "dwolla") {
		return "https://api.dwolla.com"
	}
	return "https://api.stripe.com/v1"
}

// WebhookConfig contains webhook handling configuration
type WebhookConfig struct {
	// Enabled enables webhook handling
	Enabled bool `json:"enabled"`

	// Path is the webhook endpoint path
	Path string `json:"path"`

	// SignatureVerification enables signature checking
	SignatureVerification bool `json:"signature_verification"`

	// ToleranceSeconds is the timestamp tolerance for signatures
	ToleranceSeconds int `json:"tolerance_seconds"`

	// MaxRetries for webhook delivery
	MaxRetries int `json:"max_retries"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	// Enabled enables rate limiting
	Enabled bool `json:"enabled"`

	// MaxRequestsPerMinute is the max requests per minute per customer
	MaxRequestsPerMinute int `json:"max_requests_per_minute"`

	// MaxPaymentsPerHour is the max payments per hour per customer
	MaxPaymentsPerHour int `json:"max_payments_per_hour"`

	// MaxRefundsPerDay is the max refunds per day
	MaxRefundsPerDay int `json:"max_refunds_per_day"`

	// BurstSize is the token bucket burst size
	BurstSize int `json:"burst_size"`
}

// ConversionConfig contains fiat-to-crypto conversion configuration
type ConversionConfig struct {
	// Enabled enables fiat-to-crypto conversion
	Enabled bool `json:"enabled"`

	// CryptoDenom is the target cryptocurrency denomination
	CryptoDenom string `json:"crypto_denom"`

	// PriceFeedSource is the source for conversion rates.
	// Supported values: "coingecko", "chainlink", "pyth", "median", "weighted"
	// - "coingecko": Use CoinGecko as primary source (free, rate-limited)
	// - "chainlink": Use Chainlink oracle as primary source (decentralized)
	// - "pyth": Use Pyth network as primary source (high-frequency)
	// - "median": Use median price across all sources
	// - "weighted": Use weighted average based on source confidence
	PriceFeedSource string `json:"price_feed_source"`

	// ConversionFeePercent is the fee percentage for conversion
	ConversionFeePercent float64 `json:"conversion_fee_percent"`

	// QuoteValiditySeconds is how long quotes are valid
	QuoteValiditySeconds int `json:"quote_validity_seconds"`

	// MinSlippagePercent is the minimum slippage tolerance
	MinSlippagePercent float64 `json:"min_slippage_percent"`

	// CoinGeckoAPIKey is the optional CoinGecko Pro API key
	CoinGeckoAPIKey string `json:"coingecko_api_key,omitempty"`

	// ChainlinkRPCURL is the Ethereum RPC URL for Chainlink feeds
	ChainlinkRPCURL string `json:"chainlink_rpc_url,omitempty"`

	// PythHermesURL is the Pyth Hermes API URL
	PythHermesURL string `json:"pyth_hermes_url,omitempty"`

	// CacheTTLSeconds is how long to cache prices (default: 30)
	CacheTTLSeconds int `json:"cache_ttl_seconds,omitempty"`

	// MaxPriceDeviation is the max allowed deviation between sources (0-1)
	MaxPriceDeviation float64 `json:"max_price_deviation,omitempty"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Gateway: GatewayStripe,
		StripeConfig: StripeConfig{
			APIVersion:                    "2024-06-20",
			EnablePaymentIntentsMigration: true,
		},
		AdyenConfig: AdyenConfig{
			Environment: "test",
		},
		PayPalConfig: PayPalConfig{
			Environment: "sandbox",
		},
		ACHConfig: ACHConfig{
			Provider:           "stripe",
			Environment:        "sandbox",
			RetryMaxAttempts:   3,
			RetryInitialDelay:  200 * time.Millisecond,
			RetryMaxDelay:      2 * time.Second,
			RetryBackoffFactor: 2.0,
		},
		WebhookConfig: WebhookConfig{
			Enabled:               true,
			Path:                  "/webhooks/payment",
			SignatureVerification: true,
			ToleranceSeconds:      300,
			MaxRetries:            3,
		},
		RateLimitConfig: RateLimitConfig{
			Enabled:              true,
			MaxRequestsPerMinute: 60,
			MaxPaymentsPerHour:   10,
			MaxRefundsPerDay:     5,
			BurstSize:            5,
		},
		ConversionConfig: ConversionConfig{
			Enabled:              true,
			CryptoDenom:          "uve",
			PriceFeedSource:      "coingecko",
			ConversionFeePercent: 1.5,
			QuoteValiditySeconds: 60,
			MinSlippagePercent:   0.5,
		},
		RequestTimeout:     30 * time.Second,
		RetryMaxAttempts:   3,
		RetryInitialDelay:  100 * time.Millisecond,
		RetryMaxDelay:      2 * time.Second,
		RetryBackoffFactor: 2.0,
		SupportedCurrencies: []Currency{
			CurrencyUSD,
			CurrencyEUR,
			CurrencyGBP,
		},
		MinAmount: map[Currency]int64{
			CurrencyUSD: 100, // $1.00
			CurrencyEUR: 100, // €1.00
			CurrencyGBP: 100, // £1.00
		},
		MaxAmount: map[Currency]int64{
			CurrencyUSD: 10000000, // $100,000.00
			CurrencyEUR: 10000000, // €100,000.00
			CurrencyGBP: 10000000, // £100,000.00
		},
		DefaultStatementDescriptor: "VIRTENGINE",
		EnableSandbox:              false,
		EnableLogging:              false,
	}
}

// Validate validates the configuration
func (c Config) Validate() error {
	if !c.Gateway.IsValid() {
		return ErrGatewayNotConfigured
	}

	switch c.Gateway {
	case GatewayStripe:
		if c.StripeConfig.SecretKey == "" {
			return ErrGatewayNotConfigured
		}
	case GatewayAdyen:
		if c.AdyenConfig.APIKey == "" || c.AdyenConfig.MerchantAccount == "" {
			return ErrGatewayNotConfigured
		}
	case GatewayPayPal:
		if c.PayPalConfig.ClientID == "" || c.PayPalConfig.ClientSecret == "" {
			return ErrGatewayNotConfigured
		}
	case GatewayACH:
		if c.ACHConfig.SecretKey == "" {
			return ErrGatewayNotConfigured
		}
	}

	if len(c.SupportedCurrencies) == 0 {
		return ErrInvalidCurrency
	}

	for _, curr := range c.SupportedCurrencies {
		if !curr.IsValid() {
			return ErrInvalidCurrency
		}
	}

	return nil
}

// IsCurrencySupported checks if a currency is supported
func (c Config) IsCurrencySupported(currency Currency) bool {
	for _, curr := range c.SupportedCurrencies {
		if curr == currency {
			return true
		}
	}
	return false
}

// ValidateAmount checks if an amount is within limits
func (c Config) ValidateAmount(amount Amount) error {
	if !c.IsCurrencySupported(amount.Currency) {
		return ErrInvalidCurrency
	}

	if min, ok := c.MinAmount[amount.Currency]; ok && amount.Value < min {
		return ErrAmountBelowMinimum
	}

	if max, ok := c.MaxAmount[amount.Currency]; ok && amount.Value > max {
		return ErrAmountAboveMaximum
	}

	return nil
}
