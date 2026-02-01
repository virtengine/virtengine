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

package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadSecrets(t *testing.T) {
	// Save original env and restore after test
	origEnv := map[string]string{
		"DATABASE_URL":    os.Getenv("DATABASE_URL"),
		"JWT_SECRET":      os.Getenv("JWT_SECRET"),
		"ENCRYPTION_KEY":  os.Getenv("ENCRYPTION_KEY"),
		"OPENAI_API_KEY":  os.Getenv("OPENAI_API_KEY"),
		"STRIPE_SECRET_KEY": os.Getenv("STRIPE_SECRET_KEY"),
	}
	defer func() {
		for k, v := range origEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	// Set test values
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/testdb")
	os.Setenv("JWT_SECRET", "test-jwt-secret-12345")
	os.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	os.Setenv("OPENAI_API_KEY", "sk-test-openai-key")
	os.Setenv("STRIPE_SECRET_KEY", "sk_test_stripe_key")

	cfg, err := LoadSecrets()
	if err != nil {
		t.Fatalf("LoadSecrets() error = %v", err)
	}

	// Verify loaded values
	if cfg.DatabaseURL != "postgres://localhost:5432/testdb" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://localhost:5432/testdb")
	}
	if cfg.JWTSecret != "test-jwt-secret-12345" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "test-jwt-secret-12345")
	}
	if cfg.EncryptionKey != "0123456789abcdef0123456789abcdef" {
		t.Errorf("EncryptionKey = %q, want %q", cfg.EncryptionKey, "0123456789abcdef0123456789abcdef")
	}
	if cfg.OpenAIAPIKey != "sk-test-openai-key" {
		t.Errorf("OpenAIAPIKey = %q, want %q", cfg.OpenAIAPIKey, "sk-test-openai-key")
	}
	if cfg.StripeSecretKey != "sk_test_stripe_key" {
		t.Errorf("StripeSecretKey = %q, want %q", cfg.StripeSecretKey, "sk_test_stripe_key")
	}
}

func TestValidateSecrets_AllRequired(t *testing.T) {
	cfg := &SecretConfig{
		DatabaseURL:   "postgres://localhost:5432/testdb",
		JWTSecret:     "test-jwt-secret",
		EncryptionKey: "test-encryption-key",
	}

	err := cfg.ValidateSecrets()
	if err != nil {
		t.Errorf("ValidateSecrets() error = %v, want nil", err)
	}
}

func TestValidateSecrets_MissingRequired(t *testing.T) {
	cfg := &SecretConfig{
		DatabaseURL: "postgres://localhost:5432/testdb",
		// JWTSecret and EncryptionKey are missing
	}

	err := cfg.ValidateSecrets()
	if err == nil {
		t.Fatal("ValidateSecrets() error = nil, want error for missing required secrets")
	}

	missingErr, ok := err.(*MissingSecretsError)
	if !ok {
		t.Fatalf("error type = %T, want *MissingSecretsError", err)
	}

	if len(missingErr.Errors) != 2 {
		t.Errorf("missing errors count = %d, want 2", len(missingErr.Errors))
	}

	// Check that the error message mentions the missing fields
	errStr := err.Error()
	if !strings.Contains(errStr, "JWTSecret") {
		t.Errorf("error message should mention JWTSecret: %s", errStr)
	}
	if !strings.Contains(errStr, "EncryptionKey") {
		t.Errorf("error message should mention EncryptionKey: %s", errStr)
	}
}

func TestValidateSecrets_AllMissing(t *testing.T) {
	cfg := &SecretConfig{}

	err := cfg.ValidateSecrets()
	if err == nil {
		t.Fatal("ValidateSecrets() error = nil, want error")
	}

	missingErr, ok := err.(*MissingSecretsError)
	if !ok {
		t.Fatalf("error type = %T, want *MissingSecretsError", err)
	}

	// Should have 3 required fields missing: DatabaseURL, JWTSecret, EncryptionKey
	if len(missingErr.Errors) != 3 {
		t.Errorf("missing errors count = %d, want 3", len(missingErr.Errors))
	}
}

func TestRedactedString(t *testing.T) {
	cfg := &SecretConfig{
		DatabaseURL:     "postgres://user:password@localhost:5432/db",
		JWTSecret:       "super-secret-jwt-key",
		EncryptionKey:   "encryption-key-value",
		OpenAIAPIKey:    "sk-openai-key",
		StripeSecretKey: "", // Not set
	}

	redacted := cfg.RedactedString()

	// Should NOT contain actual secret values
	if strings.Contains(redacted, "postgres://user:password") {
		t.Error("RedactedString() contains database URL password")
	}
	if strings.Contains(redacted, "super-secret-jwt-key") {
		t.Error("RedactedString() contains JWT secret")
	}
	if strings.Contains(redacted, "sk-openai-key") {
		t.Error("RedactedString() contains OpenAI key")
	}

	// Should contain field names and status
	if !strings.Contains(redacted, "DatabaseURL") {
		t.Error("RedactedString() should contain field name DatabaseURL")
	}
	if !strings.Contains(redacted, "[SET - REDACTED]") {
		t.Error("RedactedString() should contain [SET - REDACTED] for set values")
	}
	if !strings.Contains(redacted, "[NOT SET]") {
		t.Error("RedactedString() should contain [NOT SET] for unset values")
	}
	if !strings.Contains(redacted, "(required)") {
		t.Error("RedactedString() should indicate required fields")
	}
}

func TestClear(t *testing.T) {
	cfg := &SecretConfig{
		DatabaseURL:     "postgres://localhost:5432/testdb",
		JWTSecret:       "secret-jwt-key",
		EncryptionKey:   "encryption-key",
		OpenAIAPIKey:    "sk-openai-key",
		StripeSecretKey: "sk_stripe_key",
		WaldurToken:     "waldur-token",
	}

	// Verify secrets are set before clear
	if cfg.DatabaseURL == "" {
		t.Fatal("DatabaseURL should be set before Clear()")
	}

	cfg.Clear()

	// Verify all secrets are cleared
	if cfg.DatabaseURL != "" {
		t.Errorf("DatabaseURL = %q after Clear(), want empty", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "" {
		t.Errorf("JWTSecret = %q after Clear(), want empty", cfg.JWTSecret)
	}
	if cfg.EncryptionKey != "" {
		t.Errorf("EncryptionKey = %q after Clear(), want empty", cfg.EncryptionKey)
	}
	if cfg.OpenAIAPIKey != "" {
		t.Errorf("OpenAIAPIKey = %q after Clear(), want empty", cfg.OpenAIAPIKey)
	}
	if cfg.StripeSecretKey != "" {
		t.Errorf("StripeSecretKey = %q after Clear(), want empty", cfg.StripeSecretKey)
	}
	if cfg.WaldurToken != "" {
		t.Errorf("WaldurToken = %q after Clear(), want empty", cfg.WaldurToken)
	}
}

func TestIsSet(t *testing.T) {
	cfg := &SecretConfig{
		DatabaseURL: "postgres://localhost:5432/testdb",
		JWTSecret:   "", // Not set
	}

	if !cfg.IsSet("DatabaseURL") {
		t.Error("IsSet(DatabaseURL) = false, want true")
	}
	if cfg.IsSet("JWTSecret") {
		t.Error("IsSet(JWTSecret) = true, want false")
	}
	if cfg.IsSet("NonExistentField") {
		t.Error("IsSet(NonExistentField) = true, want false")
	}
}

func TestGetRequiredFields(t *testing.T) {
	required := GetRequiredFields()

	// Should contain the known required fields
	expectedRequired := []string{"DatabaseURL", "JWTSecret", "EncryptionKey"}
	for _, exp := range expectedRequired {
		found := false
		for _, r := range required {
			if r == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetRequiredFields() missing %s", exp)
		}
	}
}

func TestGetEnvVarName(t *testing.T) {
	tests := []struct {
		fieldName string
		wantEnv   string
	}{
		{"DatabaseURL", "DATABASE_URL"},
		{"JWTSecret", "JWT_SECRET"},
		{"EncryptionKey", "ENCRYPTION_KEY"},
		{"OpenAIAPIKey", "OPENAI_API_KEY"},
		{"NonExistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			got := GetEnvVarName(tt.fieldName)
			if got != tt.wantEnv {
				t.Errorf("GetEnvVarName(%s) = %q, want %q", tt.fieldName, got, tt.wantEnv)
			}
		})
	}
}

func TestMissingSecretError(t *testing.T) {
	err := &MissingSecretError{
		Name:        "JWTSecret",
		EnvVar:      "JWT_SECRET",
		Description: "Secret for JWT token signing",
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "JWTSecret") {
		t.Error("error should contain field name")
	}
	if !strings.Contains(errStr, "JWT_SECRET") {
		t.Error("error should contain env var name")
	}
	if !strings.Contains(errStr, "Secret for JWT token signing") {
		t.Error("error should contain description")
	}
}

func TestMissingSecretsError_Single(t *testing.T) {
	err := &MissingSecretsError{
		Errors: []*MissingSecretError{
			{Name: "JWTSecret", EnvVar: "JWT_SECRET", Description: "test"},
		},
	}

	errStr := err.Error()
	// Single error should use the simpler format
	if strings.Contains(errStr, "required secrets are missing") {
		t.Error("single error should not use plural format")
	}
}

func TestMissingSecretsError_Empty(t *testing.T) {
	err := &MissingSecretsError{Errors: nil}
	if err.Error() != "no missing secrets" {
		t.Errorf("empty error = %q, want 'no missing secrets'", err.Error())
	}
}

func TestLoadSecrets_EmptyEnv(t *testing.T) {
	// Save and clear relevant env vars
	envVars := []string{"DATABASE_URL", "JWT_SECRET", "ENCRYPTION_KEY", "OPENAI_API_KEY"}
	saved := make(map[string]string)
	for _, v := range envVars {
		saved[v] = os.Getenv(v)
		os.Unsetenv(v)
	}
	defer func() {
		for k, v := range saved {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	cfg, err := LoadSecrets()
	if err != nil {
		t.Fatalf("LoadSecrets() error = %v", err)
	}

	// All fields should be empty
	if cfg.DatabaseURL != "" {
		t.Errorf("DatabaseURL = %q, want empty", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "" {
		t.Errorf("JWTSecret = %q, want empty", cfg.JWTSecret)
	}
}

