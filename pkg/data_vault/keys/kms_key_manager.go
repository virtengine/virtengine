package keys

import (
	"errors"
	"time"
)

// NonExportableVaultKeyManager identifies a key manager whose private key
// material cannot be retrieved by the caller.  Production vault wiring must
// require this interface instead of accepting the legacy KeyManager.
type NonExportableVaultKeyManager interface {
	VaultKeyManager
	UsesNonExportableKeys() bool
}

// KMSKeyRecord is the public lifecycle metadata returned by a KMS-backed key
// catalogue. Private key material is intentionally not part of this type.
type KMSKeyRecord struct {
	ID           string
	Scope        Scope
	Version      uint32
	CreatedAt    time.Time
	ActivatedAt  *time.Time
	DeprecatedAt *time.Time
	RevokedAt    *time.Time
	Status       KeyStatus
}

// KMSKeyCatalog is implemented by a provider-specific KMS key catalogue.
// It performs lifecycle changes remotely and never exports private keys.
type KMSKeyCatalog interface {
	ActiveKey(Scope) (KMSKeyRecord, error)
	Key(Scope, string) (KMSKeyRecord, error)
	Rotate(Scope, time.Duration) error
	CompleteRotation(Scope) error
	RotationStatus(Scope) (*KeyRotation, error)
	List(Scope) ([]KMSKeyRecord, error)
}

// KMSKeyManager adapts a remote KMS key catalogue to VaultKeyManager. It
// deliberately returns zeroed key material; encryption is performed through
// a KMS envelope cipher rather than by callers holding private keys.
type KMSKeyManager struct{ catalog KMSKeyCatalog }

func NewKMSKeyManager(catalog KMSKeyCatalog) (*KMSKeyManager, error) {
	if catalog == nil {
		return nil, errors.New("KMS key catalogue is required")
	}
	return &KMSKeyManager{catalog: catalog}, nil
}

func (*KMSKeyManager) UsesNonExportableKeys() bool { return true }

func (m *KMSKeyManager) GetActiveKey(scope Scope) (*KeyInfo, error) {
	record, err := m.catalog.ActiveKey(scope)
	if err != nil {
		return nil, err
	}
	return kmsRecordToKeyInfo(record)
}
func (m *KMSKeyManager) GetKey(scope Scope, id string) (*KeyInfo, error) {
	record, err := m.catalog.Key(scope, id)
	if err != nil {
		return nil, err
	}
	return kmsRecordToKeyInfo(record)
}
func (m *KMSKeyManager) RotateKey(scope Scope, overlap time.Duration) error {
	return m.catalog.Rotate(scope, overlap)
}
func (m *KMSKeyManager) CompleteRotation(scope Scope) error { return m.catalog.CompleteRotation(scope) }
func (m *KMSKeyManager) GetRotationStatus(scope Scope) (*KeyRotation, error) {
	return m.catalog.RotationStatus(scope)
}
func (m *KMSKeyManager) ListKeys(scope Scope) ([]*KeyInfo, error) {
	records, err := m.catalog.List(scope)
	if err != nil {
		return nil, err
	}
	result := make([]*KeyInfo, 0, len(records))
	for _, record := range records {
		key, err := kmsRecordToKeyInfo(record)
		if err != nil {
			return nil, err
		}
		result = append(result, key)
	}
	return result, nil
}

func kmsRecordToKeyInfo(record KMSKeyRecord) (*KeyInfo, error) {
	if record.ID == "" || record.Scope == "" || record.Version == 0 || record.Status == "" {
		return nil, errors.New("invalid KMS key record")
	}
	return &KeyInfo{ID: record.ID, Scope: record.Scope, Version: record.Version, CreatedAt: record.CreatedAt, ActivatedAt: record.ActivatedAt, DeprecatedAt: record.DeprecatedAt, RevokedAt: record.RevokedAt, Status: record.Status}, nil
}
