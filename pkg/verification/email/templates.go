// Package email provides email template rendering for verification emails.
package email

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/errors"
)

// ============================================================================
// Template Types
// ============================================================================

// TemplateType identifies the type of email template
type TemplateType string

const (
	// TemplateOTPVerification is the OTP verification email template
	TemplateOTPVerification TemplateType = "otp_verification"

	// TemplateMagicLink is the magic link verification email template
	TemplateMagicLink TemplateType = "magic_link"

	// TemplateVerificationSuccess is the verification success confirmation email
	TemplateVerificationSuccess TemplateType = "verification_success"

	// TemplateVerificationReminder is the verification reminder email
	TemplateVerificationReminder TemplateType = "verification_reminder"
)

// TemplateData contains data for rendering email templates
type TemplateData struct {
	// OTP is the one-time password (for OTP template)
	OTP string `json:"otp,omitempty"`

	// VerificationLink is the magic link URL (for magic link template)
	VerificationLink string `json:"verification_link,omitempty"`

	// ExpiresIn is how long until the verification expires (human-readable)
	ExpiresIn string `json:"expires_in,omitempty"`

	// ExpiresAt is when the verification expires
	ExpiresAt time.Time `json:"expires_at,omitempty"`

	// AccountAddress is the user's account address
	AccountAddress string `json:"account_address,omitempty"`

	// MaskedEmail is the masked email address
	MaskedEmail string `json:"masked_email,omitempty"`

	// ProductName is the product name
	ProductName string `json:"product_name,omitempty"`

	// SupportEmail is the support email address
	SupportEmail string `json:"support_email,omitempty"`

	// CompanyName is the company name
	CompanyName string `json:"company_name,omitempty"`

	// Year is the current year (for copyright)
	Year int `json:"year,omitempty"`

	// Locale is the user's preferred locale
	Locale string `json:"locale,omitempty"`

	// IPAddress is the IP address that initiated the request
	IPAddress string `json:"ip_address,omitempty"`

	// UserAgent is the user agent that initiated the request
	UserAgent string `json:"user_agent,omitempty"`

	// AttemptsRemaining is the number of verification attempts remaining
	AttemptsRemaining uint32 `json:"attempts_remaining,omitempty"`

	// Custom contains additional custom data
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// ============================================================================
// Template Manager
// ============================================================================

// TemplateManager manages email templates
type TemplateManager struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
	defaults  TemplateDefaults
}

// TemplateDefaults contains default values for templates
type TemplateDefaults struct {
	ProductName  string
	CompanyName  string
	SupportEmail string
	Year         int
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(defaults TemplateDefaults) *TemplateManager {
	if defaults.Year == 0 {
		defaults.Year = time.Now().Year()
	}
	if defaults.ProductName == "" {
		defaults.ProductName = "VirtEngine"
	}
	if defaults.CompanyName == "" {
		defaults.CompanyName = "VirtEngine Foundation"
	}

	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
		defaults:  defaults,
	}

	// Load built-in templates
	tm.loadBuiltinTemplates()

	return tm
}

// loadBuiltinTemplates loads the built-in email templates
func (tm *TemplateManager) loadBuiltinTemplates() {
	// OTP Verification HTML Template
	otpHTML := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Email Verification</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .container { background: #f9f9f9; border-radius: 8px; padding: 30px; }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 24px; font-weight: bold; color: #7C3AED; }
        .otp-box { background: #7C3AED; color: white; font-size: 32px; letter-spacing: 8px; padding: 20px; text-align: center; border-radius: 8px; margin: 20px 0; font-family: monospace; }
        .expires { color: #666; font-size: 14px; text-align: center; }
        .security-notice { background: #FEF3CD; border-left: 4px solid #FFC107; padding: 15px; margin: 20px 0; font-size: 14px; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">{{.ProductName}}</div>
        </div>
        
        <h2>Verify Your Email Address</h2>
        
        <p>Use the following verification code to complete your email verification:</p>
        
        <div class="otp-box">{{.OTP}}</div>
        
        <p class="expires">This code expires {{.ExpiresIn}}</p>
        
        <div class="security-notice">
            <strong>Security Notice:</strong> Never share this code with anyone. {{.ProductName}} will never ask for your verification code via phone, email, or any other channel.
        </div>
        
        <p>If you didn't request this verification, you can safely ignore this email.</p>
        
        <div class="footer">
            <p>© {{.Year}} {{.CompanyName}}. All rights reserved.</p>
            <p>Need help? Contact us at {{.SupportEmail}}</p>
        </div>
    </div>
</body>
</html>`

	// OTP Verification Text Template
	otpText := `{{.ProductName}} - Email Verification

Your verification code is: {{.OTP}}

This code expires {{.ExpiresIn}}.

SECURITY NOTICE: Never share this code with anyone. {{.ProductName}} will never ask for your verification code.

If you didn't request this verification, you can safely ignore this email.

---
© {{.Year}} {{.CompanyName}}
Need help? Contact us at {{.SupportEmail}}`

	// Magic Link HTML Template
	magicLinkHTML := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Email Verification</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .container { background: #f9f9f9; border-radius: 8px; padding: 30px; }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 24px; font-weight: bold; color: #7C3AED; }
        .button { display: inline-block; background: #7C3AED; color: white !important; text-decoration: none; padding: 15px 30px; border-radius: 8px; font-size: 16px; font-weight: bold; margin: 20px 0; }
        .button:hover { background: #6D28D9; }
        .link-box { background: #f0f0f0; padding: 15px; border-radius: 4px; word-break: break-all; font-size: 12px; margin: 20px 0; }
        .expires { color: #666; font-size: 14px; text-align: center; }
        .security-notice { background: #FEF3CD; border-left: 4px solid #FFC107; padding: 15px; margin: 20px 0; font-size: 14px; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">{{.ProductName}}</div>
        </div>
        
        <h2>Verify Your Email Address</h2>
        
        <p>Click the button below to verify your email address:</p>
        
        <p style="text-align: center;">
            <a href="{{.VerificationLink}}" class="button">Verify Email</a>
        </p>
        
        <p class="expires">This link expires {{.ExpiresIn}}</p>
        
        <p>If the button doesn't work, copy and paste this link into your browser:</p>
        <div class="link-box">{{.VerificationLink}}</div>
        
        <div class="security-notice">
            <strong>Security Notice:</strong> Only click this link if you initiated this verification. If you didn't request this, you can safely ignore this email.
        </div>
        
        <div class="footer">
            <p>© {{.Year}} {{.CompanyName}}. All rights reserved.</p>
            <p>Need help? Contact us at {{.SupportEmail}}</p>
        </div>
    </div>
</body>
</html>`

	// Magic Link Text Template
	magicLinkText := `{{.ProductName}} - Email Verification

Click the link below to verify your email address:

{{.VerificationLink}}

This link expires {{.ExpiresIn}}.

SECURITY NOTICE: Only click this link if you initiated this verification. If you didn't request this, you can safely ignore this email.

---
© {{.Year}} {{.CompanyName}}
Need help? Contact us at {{.SupportEmail}}`

	// Verification Success HTML Template
	successHTML := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Email Verified</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .container { background: #f9f9f9; border-radius: 8px; padding: 30px; }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 24px; font-weight: bold; color: #7C3AED; }
        .success-icon { font-size: 64px; text-align: center; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">{{.ProductName}}</div>
        </div>
        
        <div class="success-icon">✅</div>
        
        <h2 style="text-align: center;">Email Verified Successfully!</h2>
        
        <p>Your email address <strong>{{.MaskedEmail}}</strong> has been successfully verified and linked to your identity.</p>
        
        <p>Your verification attestation has been recorded on-chain and contributes to your identity score.</p>
        
        <div class="footer">
            <p>© {{.Year}} {{.CompanyName}}. All rights reserved.</p>
            <p>Need help? Contact us at {{.SupportEmail}}</p>
        </div>
    </div>
</body>
</html>`

	// Verification Success Text Template
	successText := `{{.ProductName}} - Email Verified Successfully!

Your email address {{.MaskedEmail}} has been successfully verified and linked to your identity.

Your verification attestation has been recorded on-chain and contributes to your identity score.

---
© {{.Year}} {{.CompanyName}}
Need help? Contact us at {{.SupportEmail}}`

	// Parse and store templates
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.templates[string(TemplateOTPVerification)+"_html"] = template.Must(template.New("otp_html").Parse(otpHTML))
	tm.templates[string(TemplateOTPVerification)+"_text"] = template.Must(template.New("otp_text").Parse(otpText))
	tm.templates[string(TemplateMagicLink)+"_html"] = template.Must(template.New("magic_link_html").Parse(magicLinkHTML))
	tm.templates[string(TemplateMagicLink)+"_text"] = template.Must(template.New("magic_link_text").Parse(magicLinkText))
	tm.templates[string(TemplateVerificationSuccess)+"_html"] = template.Must(template.New("success_html").Parse(successHTML))
	tm.templates[string(TemplateVerificationSuccess)+"_text"] = template.Must(template.New("success_text").Parse(successText))
}

// Render renders an email template
func (tm *TemplateManager) Render(templateType TemplateType, data TemplateData) (subject, textBody, htmlBody string, err error) {
	// Apply defaults
	if data.ProductName == "" {
		data.ProductName = tm.defaults.ProductName
	}
	if data.CompanyName == "" {
		data.CompanyName = tm.defaults.CompanyName
	}
	if data.SupportEmail == "" {
		data.SupportEmail = tm.defaults.SupportEmail
	}
	if data.Year == 0 {
		data.Year = tm.defaults.Year
	}

	tm.mu.RLock()
	htmlTemplate := tm.templates[string(templateType)+"_html"]
	textTemplate := tm.templates[string(templateType)+"_text"]
	tm.mu.RUnlock()

	if htmlTemplate == nil || textTemplate == nil {
		return "", "", "", errors.Wrapf(ErrTemplateError, "template not found: %s", templateType)
	}

	// Render HTML
	var htmlBuf bytes.Buffer
	if err := htmlTemplate.Execute(&htmlBuf, data); err != nil {
		return "", "", "", errors.Wrapf(ErrTemplateError, "failed to render HTML: %v", err)
	}
	htmlBody = htmlBuf.String()

	// Render text
	var textBuf bytes.Buffer
	if err := textTemplate.Execute(&textBuf, data); err != nil {
		return "", "", "", errors.Wrapf(ErrTemplateError, "failed to render text: %v", err)
	}
	textBody = textBuf.String()

	// Generate subject
	subject = tm.generateSubject(templateType, data)

	return subject, textBody, htmlBody, nil
}

// generateSubject generates the email subject for a template type
func (tm *TemplateManager) generateSubject(templateType TemplateType, data TemplateData) string {
	productName := data.ProductName
	if productName == "" {
		productName = tm.defaults.ProductName
	}

	switch templateType {
	case TemplateOTPVerification:
		return fmt.Sprintf("[%s] Your verification code is %s", productName, data.OTP)
	case TemplateMagicLink:
		return fmt.Sprintf("[%s] Verify your email address", productName)
	case TemplateVerificationSuccess:
		return fmt.Sprintf("[%s] Email verified successfully", productName)
	case TemplateVerificationReminder:
		return fmt.Sprintf("[%s] Complete your email verification", productName)
	default:
		return fmt.Sprintf("[%s] Email Verification", productName)
	}
}

// RenderEmail renders a complete email for sending
func (tm *TemplateManager) RenderEmail(templateType TemplateType, data TemplateData, to string) (*Email, error) {
	subject, textBody, htmlBody, err := tm.Render(templateType, data)
	if err != nil {
		return nil, err
	}

	return &Email{
		To:       to,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
		Tags:     []string{string(templateType), "verification"},
		Metadata: map[string]string{
			"template_type": string(templateType),
		},
	}, nil
}

// FormatExpiryDuration formats a duration as a human-readable string
func FormatExpiryDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("in %d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "in 1 minute"
		}
		return fmt.Sprintf("in %d minutes", mins)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "in 1 hour"
		}
		return fmt.Sprintf("in %d hours", hours)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "in 1 day"
	}
	return fmt.Sprintf("in %d days", days)
}

// FormatExpiryTime formats an expiry time as a human-readable string
func FormatExpiryTime(expiresAt time.Time) string {
	d := time.Until(expiresAt)
	if d < 0 {
		return "expired"
	}
	return FormatExpiryDuration(d)
}

// ============================================================================
// Subject Line Helpers
// ============================================================================

// OTPSubjectLine generates a subject line for OTP emails
func OTPSubjectLine(otp string, productName string) string {
	// Some users prefer seeing the OTP in the subject for convenience
	return fmt.Sprintf("[%s] Your verification code is %s", productName, otp)
}

// MagicLinkSubjectLine generates a subject line for magic link emails
func MagicLinkSubjectLine(productName string) string {
	return fmt.Sprintf("[%s] Verify your email address", productName)
}

// SanitizeOTPForSubject ensures the OTP is safe for email subjects
func SanitizeOTPForSubject(otp string) string {
	// Remove any characters that might cause issues in email subjects
	var safe strings.Builder
	for _, r := range otp {
		if r >= '0' && r <= '9' {
			safe.WriteRune(r)
		}
	}
	return safe.String()
}
