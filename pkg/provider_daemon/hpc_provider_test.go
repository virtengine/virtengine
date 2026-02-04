// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-21C: HPC Provider Tests
package provider_daemon

import (
	"testing"
	"time"
)

func TestHPCProviderConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  HPCProviderConfig
		wantErr bool
	}{
		{
			name:    "default config is valid",
			config:  DefaultHPCProviderConfig(),
			wantErr: false,
		},
		{
			name: "invalid chain buffer size",
			config: func() HPCProviderConfig {
				c := DefaultHPCProviderConfig()
				c.Chain.Enabled = true
				c.Chain.SubscriptionBufferSize = 0
				return c
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHPCComponentHealth(t *testing.T) {
	health := HPCComponentHealth{
		Name:      "test_component",
		Healthy:   true,
		Message:   "all good",
		LastCheck: time.Now(),
		Details: map[string]interface{}{
			"key": "value",
		},
	}

	if health.Name != "test_component" {
		t.Errorf("Expected name 'test_component', got %s", health.Name)
	}

	if !health.Healthy {
		t.Error("Expected healthy to be true")
	}
}

func TestDefaultHPCProviderConfigs(t *testing.T) {
	// Test chain subscriber config
	chainConfig := DefaultHPCChainSubscriberConfig()
	if !chainConfig.Enabled {
		t.Error("Expected chain subscriber enabled by default")
	}
	if chainConfig.SubscriptionBufferSize != 100 {
		t.Errorf("Expected buffer size 100, got %d", chainConfig.SubscriptionBufferSize)
	}

	// Test credential config
	credConfig := DefaultHPCCredentialConfig()
	if !credConfig.RequireEncryption {
		t.Error("Expected encryption required by default")
	}

	// Test provider config
	providerConfig := DefaultHPCProviderConfig()
	if providerConfig.HPC.Enabled {
		t.Error("Expected HPC disabled by default")
	}
}

func TestProviderCredentialSigner(t *testing.T) {
	credConfig := DefaultHPCCredentialManagerConfig()
	credConfig.AllowUnencrypted = true

	credManager, err := NewHPCCredentialManager(credConfig)
	if err != nil {
		t.Fatalf("Failed to create credential manager: %v", err)
	}

	// Unlock with empty passphrase (allowed since AllowUnencrypted is true)
	if err := credManager.Unlock(""); err != nil {
		t.Fatalf("Failed to unlock: %v", err)
	}

	// Generate signing key
	if err := credManager.GenerateSigningKey(); err != nil {
		t.Fatalf("Failed to generate signing key: %v", err)
	}

	signer := &providerCredentialSigner{
		credManager:     credManager,
		providerAddress: "test-provider",
	}

	// Test signing
	data := []byte("test data to sign")
	sig, err := signer.Sign(data)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}
	if len(sig) == 0 {
		t.Error("Expected non-empty signature")
	}

	// Test provider address
	if signer.GetProviderAddress() != "test-provider" {
		t.Errorf("Expected provider address 'test-provider', got %s", signer.GetProviderAddress())
	}

	// Test set provider address
	signer.SetProviderAddress("new-provider")
	if signer.GetProviderAddress() != "new-provider" {
		t.Errorf("Expected provider address 'new-provider', got %s", signer.GetProviderAddress())
	}
}

func TestHPCProviderHealth_Structure(t *testing.T) {
	health := &HPCProviderHealth{
		Overall:   true,
		Message:   "healthy",
		LastCheck: time.Now(),
		Components: []HPCComponentHealth{
			{Name: "test", Healthy: true, Message: "ok"},
		},
		ActiveJobs:     5,
		PendingRecords: 10,
	}

	if !health.Overall {
		t.Error("Expected overall healthy")
	}

	if len(health.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(health.Components))
	}

	if health.ActiveJobs != 5 {
		t.Errorf("Expected 5 active jobs, got %d", health.ActiveJobs)
	}
}

func TestHPCChainSubscriberConfig_Defaults(t *testing.T) {
	config := DefaultHPCChainSubscriberConfig()

	if !config.Enabled {
		t.Error("Expected enabled by default")
	}

	if config.SubscriptionBufferSize != 100 {
		t.Errorf("Expected buffer size 100, got %d", config.SubscriptionBufferSize)
	}

	if config.ReconnectInterval != 10*time.Second {
		t.Errorf("Expected reconnect interval 10s, got %v", config.ReconnectInterval)
	}

	if config.MaxReconnectAttempts != 0 {
		t.Errorf("Expected max reconnect attempts 0 (infinite), got %d", config.MaxReconnectAttempts)
	}
}

func TestHPCCredentialConfig_Defaults(t *testing.T) {
	config := DefaultHPCCredentialConfig()

	if !config.RequireEncryption {
		t.Error("Expected encryption required by default")
	}

	if !config.AutoRotateCredentials {
		t.Error("Expected auto rotate enabled by default")
	}

	if config.RotationCheckInterval != 24*time.Hour {
		t.Errorf("Expected rotation check interval 24h, got %v", config.RotationCheckInterval)
	}
}
