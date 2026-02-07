package hsm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, BackendSoftHSM, cfg.Backend)
	assert.NotNil(t, cfg.PKCS11)
	assert.Equal(t, 30*time.Second, cfg.ConnectionTimeout)
	assert.Equal(t, 10*time.Second, cfg.OperationTimeout)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.True(t, cfg.AuditLog)
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr string
	}{
		{
			name:   "valid default",
			modify: func(_ *Config) {},
		},
		{
			name:    "empty backend",
			modify:  func(c *Config) { c.Backend = "" },
			wantErr: "backend is required",
		},
		{
			name:    "pkcs11 without config",
			modify:  func(c *Config) { c.Backend = BackendPKCS11; c.PKCS11 = nil },
			wantErr: "pkcs11 config required",
		},
		{
			name: "pkcs11 empty library",
			modify: func(c *Config) {
				c.Backend = BackendPKCS11
				c.PKCS11 = &PKCS11Config{}
			},
			wantErr: "library path is required",
		},
		{
			name:    "ledger without config",
			modify:  func(c *Config) { c.Backend = BackendLedger; c.PKCS11 = nil; c.Ledger = nil },
			wantErr: "ledger config required",
		},
		{
			name:    "cloud without config",
			modify:  func(c *Config) { c.Backend = BackendAWSCloudHSM; c.PKCS11 = nil; c.Cloud = nil },
			wantErr: "cloud config required",
		},
		{
			name: "cloud empty provider",
			modify: func(c *Config) {
				c.Backend = BackendAWSCloudHSM
				c.PKCS11 = nil
				c.Cloud = &CloudConfig{}
			},
			wantErr: "cloud provider is required",
		},
		{
			name:    "zero connection timeout",
			modify:  func(c *Config) { c.ConnectionTimeout = 0 },
			wantErr: "connection_timeout must be positive",
		},
		{
			name:    "zero operation timeout",
			modify:  func(c *Config) { c.OperationTimeout = 0 },
			wantErr: "operation_timeout must be positive",
		},
		{
			name: "unsupported backend",
			modify: func(c *Config) {
				c.Backend = "invalid"
				c.PKCS11 = nil
			},
			wantErr: "unsupported backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(&cfg)
			err := cfg.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
