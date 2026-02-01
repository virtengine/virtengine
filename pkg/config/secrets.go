// Copyright 2024-2025 VirtEngine Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package config provides secure configuration management for VirtEngine secrets.
package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

const stringTrue = "true"

// SecretConfig holds all secret configuration values for the application.
// Use struct tags to specify environment variable names and requirements:
//   - env: the environment variable name
//   - required: if "true", the secret must be set
//   - desc: human-readable description of the secret
type SecretConfig struct {
	// Database connection string
	DatabaseURL string `env:"DATABASE_URL" required:"true" desc:"Database connection string"`

	// AI Provider API Keys
	OpenAIAPIKey    string `env:"OPENAI_API_KEY" required:"false" desc:"OpenAI API key for AI services"`
	AnthropicAPIKey string `env:"ANTHROPIC_API_KEY" required:"false" desc:"Anthropic API key for Claude"`

	// Payment Provider Keys
	StripeSecretKey string `env:"STRIPE_SECRET_KEY" required:"false" desc:"Stripe secret key for payments"`
	AdyenAPIKey     string `env:"ADYEN_API_KEY" required:"false" desc:"Adyen API key for payments"`

	// Infrastructure Keys
	WaldurToken string `env:"WALDUR_TOKEN" required:"false" desc:"Waldur API token for infrastructure"`

	// Security Keys
	JWTSecret     string `env:"JWT_SECRET" required:"true" desc:"Secret for JWT token signing"`
	EncryptionKey string `env:"ENCRYPTION_KEY" required:"true" desc:"Master encryption key (32 bytes hex)"`

	// Identity Verification API Keys
	GovUKAPIKey      string `env:"GOVUK_API_KEY" required:"false" desc:"GOV.UK Verify API key"`
	EIDASAPIKey      string `env:"EIDAS_API_KEY" required:"false" desc:"eIDAS API key for EU identity"`
	AAMVAClientSecret string `env:"AAMVA_CLIENT_SECRET" required:"false" desc:"AAMVA client secret for US license verification"`

	// Provider Keys
	AWSAccessKeyID     string `env:"AWS_ACCESS_KEY_ID" required:"false" desc:"AWS access key ID"`
	AWSSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY" required:"false" desc:"AWS secret access key"`
	AzureClientSecret  string `env:"AZURE_CLIENT_SECRET" required:"false" desc:"Azure client secret"`
	GCPServiceAccount  string `env:"GCP_SERVICE_ACCOUNT_KEY" required:"false" desc:"GCP service account key JSON"`

	// Monitoring Keys
	DatadogAPIKey   string `env:"DATADOG_API_KEY" required:"false" desc:"Datadog API key for monitoring"`
	SentryDSN       string `env:"SENTRY_DSN" required:"false" desc:"Sentry DSN for error tracking"`
	PagerDutyAPIKey string `env:"PAGERDUTY_API_KEY" required:"false" desc:"PagerDuty API key for alerting"`

	// Blockchain Keys
	ValidatorPrivateKey string `env:"VALIDATOR_PRIVATE_KEY" required:"false" desc:"Validator signing key (sensitive)"`
	CosmosRPCAuth       string `env:"COSMOS_RPC_AUTH" required:"false" desc:"Cosmos RPC authentication token"`
}

// MissingSecretError indicates a required secret is not set.
type MissingSecretError struct {
	Name        string
	EnvVar      string
	Description string
}

func (e *MissingSecretError) Error() string {
	return fmt.Sprintf("missing required secret %s (env: %s): %s", e.Name, e.EnvVar, e.Description)
}

// MissingSecretsError aggregates multiple missing secret errors.
type MissingSecretsError struct {
	Errors []*MissingSecretError
}

func (e *MissingSecretsError) Error() string {
	if len(e.Errors) == 0 {
		return "no missing secrets"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d required secrets are missing:\n", len(e.Errors)))
	for _, err := range e.Errors {
		sb.WriteString("  - ")
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}

// LoadSecrets loads secret configuration from environment variables.
// It populates all fields that have corresponding environment variables set.
// Call ValidateSecrets() after loading to check for missing required secrets.
func LoadSecrets() (*SecretConfig, error) {
	cfg := &SecretConfig{}

	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}

		envValue := os.Getenv(envTag)
		if envValue != "" {
			v.Field(i).SetString(envValue)
		}
	}

	return cfg, nil
}

// ValidateSecrets checks that all required secrets are set.
// Returns a MissingSecretsError if any required secrets are missing.
func (c *SecretConfig) ValidateSecrets() error {
	var missing []*MissingSecretError

	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		requiredTag := field.Tag.Get("required")
		if requiredTag != stringTrue {
			continue
		}

		value := v.Field(i).String()
		if value == "" {
			missing = append(missing, &MissingSecretError{
				Name:        field.Name,
				EnvVar:      field.Tag.Get("env"),
				Description: field.Tag.Get("desc"),
			})
		}
	}

	if len(missing) > 0 {
		return &MissingSecretsError{Errors: missing}
	}
	return nil
}

// RedactedString returns a string representation of the config with all
// secret values redacted for safe logging. Only shows which secrets are set.
func (c *SecretConfig) RedactedString() string {
	var sb strings.Builder
	sb.WriteString("SecretConfig{\n")

	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).String()

		status := "[NOT SET]"
		if value != "" {
			status = "[SET - REDACTED]"
		}

		envTag := field.Tag.Get("env")
		requiredTag := field.Tag.Get("required")
		reqMarker := ""
		if requiredTag == stringTrue {
			reqMarker = " (required)"
		}

		sb.WriteString(fmt.Sprintf("  %s (%s)%s: %s\n", field.Name, envTag, reqMarker, status))
	}

	sb.WriteString("}")
	return sb.String()
}

// Clear zeros out all secret values from memory.
// This should be called when the secrets are no longer needed to minimize
// the window where sensitive data exists in memory.
//
// Note: Go's garbage collector may still have copies of the data.
// For maximum security, consider using specialized secret storage libraries
// that allocate secrets in mlock'd memory regions.
func (c *SecretConfig) Clear() {
	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String {
			// Set to empty string - Go's GC will eventually reclaim the memory.
			// We cannot safely zero string memory as strings may point to
			// read-only memory (e.g., string literals or interned strings).
			field.SetString("")
		}
	}
}

// secureBytes is a helper type for secrets that need to be zeroed.
// Use this for secrets loaded from external sources (files, network)
// where you control the allocation and can safely zero the memory.
//
//nolint:unused // Reserved for future secure memory handling
type secureBytes []byte

// Clear zeros out the byte slice contents.
//
//nolint:unused // Reserved for future secure memory handling
func (s secureBytes) Clear() {
	for i := range s {
		s[i] = 0
	}
}

// IsSet returns true if the specified secret field is set (non-empty).
func (c *SecretConfig) IsSet(fieldName string) bool {
	v := reflect.ValueOf(c).Elem()
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return false
	}
	return field.String() != ""
}

// GetRequiredFields returns a list of all field names marked as required.
func GetRequiredFields() []string {
	var required []string
	t := reflect.TypeOf(SecretConfig{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("required") == stringTrue {
			required = append(required, field.Name)
		}
	}
	return required
}

// GetEnvVarName returns the environment variable name for a given field.
func GetEnvVarName(fieldName string) string {
	t := reflect.TypeOf(SecretConfig{})
	field, ok := t.FieldByName(fieldName)
	if !ok {
		return ""
	}
	return field.Tag.Get("env")
}

