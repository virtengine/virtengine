package hsm

import (
	"errors"
	"fmt"
	"time"
)

// BackendType identifies the HSM backend.
type BackendType string

const (
	// BackendPKCS11 uses a PKCS#11 compatible hardware or software HSM.
	BackendPKCS11 BackendType = "pkcs11"

	// BackendLedger uses a Ledger hardware wallet.
	BackendLedger BackendType = "ledger"

	// BackendAWSCloudHSM uses AWS CloudHSM via PKCS#11.
	BackendAWSCloudHSM BackendType = "aws_cloudhsm"

	// BackendGCPCloudHSM uses Google Cloud HSM via Cloud KMS.
	BackendGCPCloudHSM BackendType = "gcp_cloudhsm"

	// BackendAzureHSM uses Azure Dedicated HSM / Managed HSM.
	BackendAzureHSM BackendType = "azure_hsm"

	// BackendSoftHSM uses SoftHSM2 for testing.
	BackendSoftHSM BackendType = "softhsm"
)

// Config holds HSM configuration.
type Config struct {
	// Backend selects the HSM backend.
	Backend BackendType `json:"backend" yaml:"backend"`

	// PKCS11 holds PKCS#11-specific configuration.
	PKCS11 *PKCS11Config `json:"pkcs11,omitempty" yaml:"pkcs11,omitempty"`

	// Ledger holds Ledger-specific configuration.
	Ledger *LedgerConfig `json:"ledger,omitempty" yaml:"ledger,omitempty"`

	// Cloud holds cloud HSM configuration.
	Cloud *CloudConfig `json:"cloud,omitempty" yaml:"cloud,omitempty"`

	// ConnectionTimeout is the timeout for connecting to the HSM.
	ConnectionTimeout time.Duration `json:"connection_timeout" yaml:"connection_timeout"`

	// OperationTimeout is the timeout for individual HSM operations.
	OperationTimeout time.Duration `json:"operation_timeout" yaml:"operation_timeout"`

	// MaxRetries is the maximum retry count for transient failures.
	MaxRetries int `json:"max_retries" yaml:"max_retries"`

	// AuditLog enables audit logging of all HSM operations.
	AuditLog bool `json:"audit_log" yaml:"audit_log"`
}

// PKCS11Config holds PKCS#11 backend settings.
type PKCS11Config struct {
	// LibraryPath is the path to the PKCS#11 shared library.
	LibraryPath string `json:"library" yaml:"library"`

	// SlotID is the token slot number.
	SlotID uint `json:"slot" yaml:"slot"`

	// PIN is the user PIN for the token.
	PIN string `json:"pin" yaml:"pin"`

	// TokenLabel is the token label (used for slot discovery).
	TokenLabel string `json:"token_label" yaml:"token_label"`
}

// LedgerConfig holds Ledger device configuration.
type LedgerConfig struct {
	// DerivationPath is the HD derivation path (e.g. m/44'/118'/0'/0/0).
	DerivationPath string `json:"derivation_path" yaml:"derivation_path"`

	// HRP is the Bech32 human-readable prefix.
	HRP string `json:"hrp" yaml:"hrp"`
}

// CloudConfig holds cloud HSM settings.
type CloudConfig struct {
	// Provider is the cloud provider (aws, gcp, azure).
	Provider string `json:"provider" yaml:"provider"`

	// Region is the cloud region.
	Region string `json:"region" yaml:"region"`

	// ClusterID is the HSM cluster identifier (AWS).
	ClusterID string `json:"cluster_id,omitempty" yaml:"cluster_id,omitempty"`

	// KeyVaultName is the Azure Key Vault name.
	KeyVaultName string `json:"key_vault_name,omitempty" yaml:"key_vault_name,omitempty"`

	// KeyRingName is the GCP KMS key ring name.
	KeyRingName string `json:"key_ring_name,omitempty" yaml:"key_ring_name,omitempty"`

	// ProjectID is the GCP project identifier.
	ProjectID string `json:"project_id,omitempty" yaml:"project_id,omitempty"`

	// CredentialsFile is the path to a service account or credentials file.
	CredentialsFile string `json:"credentials_file,omitempty" yaml:"credentials_file,omitempty"`
}

// DefaultConfig returns the default HSM configuration suitable for development
// with SoftHSM2.
func DefaultConfig() Config {
	return Config{
		Backend: BackendSoftHSM,
		PKCS11: &PKCS11Config{
			LibraryPath: "/usr/lib/softhsm/libsofthsm2.so",
			SlotID:      0,
		},
		ConnectionTimeout: 30 * time.Second,
		OperationTimeout:  10 * time.Second,
		MaxRetries:        3,
		AuditLog:          true,
	}
}

// Validate checks the configuration for required fields.
func (c *Config) Validate() error {
	if c.Backend == "" {
		return errors.New("hsm: backend is required")
	}

	switch c.Backend {
	case BackendPKCS11, BackendSoftHSM:
		if c.PKCS11 == nil {
			return errors.New("hsm: pkcs11 config required for pkcs11/softhsm backend")
		}
		if c.PKCS11.LibraryPath == "" {
			return errors.New("hsm: pkcs11 library path is required")
		}
	case BackendLedger:
		if c.Ledger == nil {
			return errors.New("hsm: ledger config required for ledger backend")
		}
	case BackendAWSCloudHSM, BackendGCPCloudHSM, BackendAzureHSM:
		if c.Cloud == nil {
			return errors.New("hsm: cloud config required for cloud HSM backend")
		}
		if c.Cloud.Provider == "" {
			return errors.New("hsm: cloud provider is required")
		}
	default:
		return fmt.Errorf("hsm: unsupported backend: %s", c.Backend)
	}

	if c.ConnectionTimeout <= 0 {
		return errors.New("hsm: connection_timeout must be positive")
	}
	if c.OperationTimeout <= 0 {
		return errors.New("hsm: operation_timeout must be positive")
	}

	return nil
}
