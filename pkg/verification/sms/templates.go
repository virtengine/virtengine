// Package sms provides localized templates for SMS verification messages.
//
// Task Reference: VE-4C - SMS Verification Delivery + Anti-Fraud
package sms

import (
	"fmt"
	"strings"
	"sync"
)

// ============================================================================
// Template Types
// ============================================================================

// TemplateType identifies the type of SMS template
type TemplateType string

const (
	// TemplateOTPVerification is the OTP verification template
	TemplateOTPVerification TemplateType = "otp_verification"

	// TemplateOTPResend is the OTP resend template
	TemplateOTPResend TemplateType = "otp_resend"

	// TemplateVerificationSuccess is the success notification template
	TemplateVerificationSuccess TemplateType = "verification_success"
)

// Template represents an SMS template
type Template struct {
	// Type is the template type
	Type TemplateType `json:"type"`

	// Locale is the locale code (e.g., "en", "es", "fr")
	Locale string `json:"locale"`

	// Body is the template body with placeholders
	Body string `json:"body"`

	// MaxLength is the maximum message length
	MaxLength int `json:"max_length"`
}

// TemplateData contains data for template rendering
type TemplateData struct {
	// OTP is the one-time password
	OTP string `json:"otp,omitempty"`

	// ExpiresIn is the human-readable expiry time
	ExpiresIn string `json:"expires_in,omitempty"`

	// ExpiresMinutes is the expiry time in minutes
	ExpiresMinutes int `json:"expires_minutes,omitempty"`

	// ProductName is the product/service name
	ProductName string `json:"product_name,omitempty"`

	// AccountAddress is the account address (truncated)
	AccountAddress string `json:"account_address,omitempty"`

	// Extra contains additional template variables
	Extra map[string]string `json:"extra,omitempty"`
}

// ============================================================================
// Template Manager
// ============================================================================

// TemplateManager manages SMS templates for different locales
type TemplateManager struct {
	mu        sync.RWMutex
	templates map[string]map[TemplateType]*Template
	defaults  TemplateDefaults
}

// TemplateDefaults contains default template values
type TemplateDefaults struct {
	ProductName    string
	SupportContact string
	ExpiresMinutes int
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(defaults TemplateDefaults) *TemplateManager {
	if defaults.ProductName == "" {
		defaults.ProductName = "VirtEngine"
	}
	if defaults.ExpiresMinutes == 0 {
		defaults.ExpiresMinutes = 5
	}

	tm := &TemplateManager{
		templates: make(map[string]map[TemplateType]*Template),
		defaults:  defaults,
	}

	// Register built-in templates
	tm.registerBuiltinTemplates()

	return tm
}

// registerBuiltinTemplates registers the built-in templates for all supported locales
func (tm *TemplateManager) registerBuiltinTemplates() {
	// English (default)
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "en",
		Body:      "Your {{product_name}} verification code is: {{otp}}. Valid for {{expires_in}}. Do not share this code.",
		MaxLength: 160,
	})
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPResend,
		Locale:    "en",
		Body:      "Your new {{product_name}} code is: {{otp}}. Valid for {{expires_in}}. Do not share this code.",
		MaxLength: 160,
	})
	tm.RegisterTemplate(&Template{
		Type:      TemplateVerificationSuccess,
		Locale:    "en",
		Body:      "Your {{product_name}} phone verification is complete. If you did not request this, contact support immediately.",
		MaxLength: 160,
	})

	// Spanish
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "es",
		Body:      "Tu código de verificación de {{product_name}} es: {{otp}}. Válido por {{expires_in}}. No compartas este código.",
		MaxLength: 160,
	})
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPResend,
		Locale:    "es",
		Body:      "Tu nuevo código de {{product_name}} es: {{otp}}. Válido por {{expires_in}}. No lo compartas.",
		MaxLength: 160,
	})

	// French
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "fr",
		Body:      "Votre code de vérification {{product_name}} est: {{otp}}. Valide pendant {{expires_in}}. Ne partagez pas ce code.",
		MaxLength: 160,
	})
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPResend,
		Locale:    "fr",
		Body:      "Votre nouveau code {{product_name}} est: {{otp}}. Valide pendant {{expires_in}}. Ne le partagez pas.",
		MaxLength: 160,
	})

	// German
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "de",
		Body:      "Ihr {{product_name}} Bestätigungscode lautet: {{otp}}. Gültig für {{expires_in}}. Teilen Sie diesen Code nicht.",
		MaxLength: 160,
	})

	// Portuguese
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "pt",
		Body:      "Seu código de verificação {{product_name}} é: {{otp}}. Válido por {{expires_in}}. Não compartilhe este código.",
		MaxLength: 160,
	})

	// Japanese
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "ja",
		Body:      "{{product_name}}の確認コード: {{otp}}。{{expires_in}}有効。このコードを共有しないでください。",
		MaxLength: 160,
	})

	// Chinese (Simplified)
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "zh",
		Body:      "您的{{product_name}}验证码是：{{otp}}。{{expires_in}}内有效。请勿分享此验证码。",
		MaxLength: 160,
	})

	// Korean
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "ko",
		Body:      "{{product_name}} 인증 코드: {{otp}}. {{expires_in}} 동안 유효합니다. 이 코드를 공유하지 마세요.",
		MaxLength: 160,
	})

	// Hindi
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "hi",
		Body:      "आपका {{product_name}} सत्यापन कोड है: {{otp}}। {{expires_in}} के लिए वैध। इस कोड को साझा न करें।",
		MaxLength: 160,
	})

	// Arabic
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "ar",
		Body:      "رمز التحقق الخاص بـ {{product_name}}: {{otp}}. صالح لمدة {{expires_in}}. لا تشارك هذا الرمز.",
		MaxLength: 160,
	})

	// Italian
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "it",
		Body:      "Il tuo codice di verifica {{product_name}} è: {{otp}}. Valido per {{expires_in}}. Non condividere questo codice.",
		MaxLength: 160,
	})

	// Russian
	tm.RegisterTemplate(&Template{
		Type:      TemplateOTPVerification,
		Locale:    "ru",
		Body:      "Ваш код подтверждения {{product_name}}: {{otp}}. Действителен {{expires_in}}. Не сообщайте этот код.",
		MaxLength: 160,
	})
}

// RegisterTemplate registers a template
func (tm *TemplateManager) RegisterTemplate(t *Template) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.templates[t.Locale] == nil {
		tm.templates[t.Locale] = make(map[TemplateType]*Template)
	}
	tm.templates[t.Locale][t.Type] = t
}

// GetTemplate returns a template for the given type and locale
func (tm *TemplateManager) GetTemplate(templateType TemplateType, locale string) *Template {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// Try exact locale
	if localeTemplates, ok := tm.templates[locale]; ok {
		if t, ok := localeTemplates[templateType]; ok {
			return t
		}
	}

	// Try base locale (e.g., "en-US" -> "en")
	if len(locale) > 2 && strings.Contains(locale, "-") {
		baseLocale := strings.Split(locale, "-")[0]
		if localeTemplates, ok := tm.templates[baseLocale]; ok {
			if t, ok := localeTemplates[templateType]; ok {
				return t
			}
		}
	}

	// Fall back to English
	if localeTemplates, ok := tm.templates["en"]; ok {
		if t, ok := localeTemplates[templateType]; ok {
			return t
		}
	}

	return nil
}

// RenderMessage renders an SMS message using a template
func (tm *TemplateManager) RenderMessage(templateType TemplateType, data TemplateData, locale string) (string, error) {
	template := tm.GetTemplate(templateType, locale)
	if template == nil {
		return "", fmt.Errorf("template not found: %s/%s", templateType, locale)
	}

	// Set defaults
	if data.ProductName == "" {
		data.ProductName = tm.defaults.ProductName
	}
	if data.ExpiresIn == "" && data.ExpiresMinutes > 0 {
		data.ExpiresIn = FormatExpiryTime(data.ExpiresMinutes)
	}
	if data.ExpiresIn == "" {
		data.ExpiresIn = FormatExpiryTime(tm.defaults.ExpiresMinutes)
	}

	// Render template
	body := template.Body
	body = strings.ReplaceAll(body, "{{otp}}", data.OTP)
	body = strings.ReplaceAll(body, "{{product_name}}", data.ProductName)
	body = strings.ReplaceAll(body, "{{expires_in}}", data.ExpiresIn)
	body = strings.ReplaceAll(body, "{{account_address}}", data.AccountAddress)

	// Replace extra variables
	for key, value := range data.Extra {
		body = strings.ReplaceAll(body, "{{"+key+"}}", value)
	}

	// Truncate if necessary
	if template.MaxLength > 0 && len(body) > template.MaxLength {
		body = body[:template.MaxLength-3] + "..."
	}

	return body, nil
}

// GetSupportedLocales returns the list of supported locales
func (tm *TemplateManager) GetSupportedLocales() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	locales := make([]string, 0, len(tm.templates))
	for locale := range tm.templates {
		locales = append(locales, locale)
	}
	return locales
}

// FormatExpiryTime formats the expiry time for display
func FormatExpiryTime(minutes int) string {
	if minutes < 1 {
		return "1 minute"
	}
	if minutes == 1 {
		return "1 minute"
	}
	if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	}
	hours := minutes / 60
	if hours == 1 {
		return "1 hour"
	}
	return fmt.Sprintf("%d hours", hours)
}

// ============================================================================
// Region-Based Rate Limits
// ============================================================================

// RegionRateLimits defines rate limits per region
type RegionRateLimits struct {
	mu     sync.RWMutex
	limits map[string]*RegionLimit
}

// RegionLimit defines rate limits for a specific region
type RegionLimit struct {
	// CountryCode is the ISO country code
	CountryCode string `json:"country_code"`

	// MaxRequestsPerPhonePerHour overrides global limit
	MaxRequestsPerPhonePerHour int `json:"max_requests_per_phone_per_hour"`

	// MaxRequestsPerAccountPerDay overrides global limit
	MaxRequestsPerAccountPerDay int `json:"max_requests_per_account_per_day"`

	// CooldownSeconds is the cooldown between requests
	CooldownSeconds int64 `json:"cooldown_seconds"`

	// RequireCarrierLookup requires carrier lookup for this region
	RequireCarrierLookup bool `json:"require_carrier_lookup"`

	// BlockVoIP blocks VoIP numbers for this region
	BlockVoIP bool `json:"block_voip"`

	// RiskMultiplier multiplies the base risk score
	RiskMultiplier float64 `json:"risk_multiplier"`
}

// NewRegionRateLimits creates a new region rate limits manager
func NewRegionRateLimits() *RegionRateLimits {
	rrl := &RegionRateLimits{
		limits: make(map[string]*RegionLimit),
	}

	// Set default limits for high-risk regions
	rrl.SetLimit(&RegionLimit{
		CountryCode:                 "NG", // Nigeria - high fraud risk
		MaxRequestsPerPhonePerHour:  2,
		MaxRequestsPerAccountPerDay: 5,
		CooldownSeconds:             120,
		RequireCarrierLookup:        true,
		BlockVoIP:                   true,
		RiskMultiplier:              1.5,
	})

	rrl.SetLimit(&RegionLimit{
		CountryCode:                 "PH", // Philippines
		MaxRequestsPerPhonePerHour:  2,
		MaxRequestsPerAccountPerDay: 5,
		CooldownSeconds:             90,
		RequireCarrierLookup:        true,
		BlockVoIP:                   true,
		RiskMultiplier:              1.3,
	})

	rrl.SetLimit(&RegionLimit{
		CountryCode:                 "IN", // India
		MaxRequestsPerPhonePerHour:  3,
		MaxRequestsPerAccountPerDay: 8,
		CooldownSeconds:             60,
		RequireCarrierLookup:        true,
		BlockVoIP:                   true,
		RiskMultiplier:              1.2,
	})

	// Low-risk regions with relaxed limits
	rrl.SetLimit(&RegionLimit{
		CountryCode:                 "US",
		MaxRequestsPerPhonePerHour:  5,
		MaxRequestsPerAccountPerDay: 15,
		CooldownSeconds:             30,
		RequireCarrierLookup:        true,
		BlockVoIP:                   true,
		RiskMultiplier:              1.0,
	})

	rrl.SetLimit(&RegionLimit{
		CountryCode:                 "CA",
		MaxRequestsPerPhonePerHour:  5,
		MaxRequestsPerAccountPerDay: 15,
		CooldownSeconds:             30,
		RequireCarrierLookup:        true,
		BlockVoIP:                   true,
		RiskMultiplier:              1.0,
	})

	rrl.SetLimit(&RegionLimit{
		CountryCode:                 "GB",
		MaxRequestsPerPhonePerHour:  5,
		MaxRequestsPerAccountPerDay: 15,
		CooldownSeconds:             30,
		RequireCarrierLookup:        false,
		BlockVoIP:                   true,
		RiskMultiplier:              1.0,
	})

	return rrl
}

// GetLimit returns the rate limit for a region
func (rrl *RegionRateLimits) GetLimit(countryCode string) *RegionLimit {
	rrl.mu.RLock()
	defer rrl.mu.RUnlock()

	if limit, ok := rrl.limits[strings.ToUpper(countryCode)]; ok {
		return limit
	}

	// Return default
	return &RegionLimit{
		CountryCode:                 countryCode,
		MaxRequestsPerPhonePerHour:  3,
		MaxRequestsPerAccountPerDay: 10,
		CooldownSeconds:             60,
		RequireCarrierLookup:        true,
		BlockVoIP:                   true,
		RiskMultiplier:              1.0,
	}
}

// SetLimit sets the rate limit for a region
func (rrl *RegionRateLimits) SetLimit(limit *RegionLimit) {
	rrl.mu.Lock()
	defer rrl.mu.Unlock()
	rrl.limits[strings.ToUpper(limit.CountryCode)] = limit
}

// GetAllLimits returns all configured region limits
func (rrl *RegionRateLimits) GetAllLimits() []*RegionLimit {
	rrl.mu.RLock()
	defer rrl.mu.RUnlock()

	limits := make([]*RegionLimit, 0, len(rrl.limits))
	for _, limit := range rrl.limits {
		limits = append(limits, limit)
	}
	return limits
}
