package contracts

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
)

const (
	PersistedSchemaInventoryVersion = "virtengine.platformsecurity.persisted-schema-inventory/v1"
	PersistedSchemaThread           = "T5"
	PersistedSchemaFrozenBaseline   = "79391a3df86d85522b92e0400c6904971ecbe65d"
	PersistedSchemaEpochBase        = "5587c384f634552c3a2dd7181ca49cafa4da1984"
	// PersistedSchemaPayloadHead is the inventoried code boundary. The inventory
	// artifact is metadata committed later as a descendant of this revision.
	PersistedSchemaPayloadHead = "b6980d266eb23d56e23108858cbbf1b2da9dfe8c"
)

type StorageClass string

const (
	StorageConsensus   StorageClass = "consensus"
	StorageOffChain    StorageClass = "off_chain"
	StorageSerialized  StorageClass = "serialized"
	StorageFixtureOnly StorageClass = "fixture_only"
	StorageTransient   StorageClass = "transient"
)

type MigrationDisposition string

const (
	MigrationRequiredUnwired     MigrationDisposition = "required_unwired"
	MigrationIntegrationRequired MigrationDisposition = "integration_required"
	MigrationExplicitReencrypt   MigrationDisposition = "explicit_reencrypt"
	MigrationAdditiveNoRewrite   MigrationDisposition = "additive_no_rewrite"
	MigrationNewStore            MigrationDisposition = "new_store"
	MigrationNone                MigrationDisposition = "none"
)

type PersistedSchemaEntry struct {
	ID                   string               `json:"id"`
	OwnerPath            string               `json:"owner_path"`
	StorageClass         StorageClass         `json:"storage_class"`
	SchemaVersion        string               `json:"schema_version"`
	Domain               string               `json:"domain"`
	KeyOrFormat          string               `json:"key_or_format"`
	MigrationDisposition MigrationDisposition `json:"migration_disposition"`
	MigrationID          string               `json:"migration_id,omitempty"`
	EvidencePaths        []string             `json:"evidence_paths"`
	Blocker              string               `json:"blocker,omitempty"`
}

type PersistedSchemaInventory struct {
	Version               string                 `json:"version"`
	Thread                string                 `json:"thread"`
	FrozenBaseline        string                 `json:"frozen_baseline"`
	IntakeBaseSHA         string                 `json:"intake_base_sha"`
	PayloadHead           string                 `json:"payload_head"`
	Entries               []PersistedSchemaEntry `json:"entries"`
	NonPersistentSurfaces []string               `json:"non_persistent_surfaces"`
	Blockers              []string               `json:"blockers"`
}

type persistedSchemaSpec struct {
	owner       string
	class       StorageClass
	version     string
	domain      string
	format      string
	disposition MigrationDisposition
	migrationID string
	evidence    []string
	blocker     string
}

var requiredPersistedSchemas = map[string]persistedSchemaSpec{
	"data-vault-key-state": {
		"pkg/data_vault/keys/persistence.go", StorageFixtureOnly, "v1", "data-vault key state", "namespace/file envelope", MigrationNewStore, "",
		[]string{"pkg/data_vault/keys/persistence.go"}, "",
	},
	"data-vault-artifact-index": {
		"pkg/data_vault/fixture_artifact_store.go", StorageFixtureOnly, "v1", "data-vault artifact index", "fixture envelope v1 namespace/revision/checksum; index.json maps artifacts/blob_metadata/legal_holds/erasure_intents/mutations; objects/<sha256>; .fixture-artifact.lock", MigrationNewStore, "",
		[]string{"pkg/data_vault/fixture_artifact_store.go", "pkg/data_vault/fixture_erasure.go"}, "",
	},
	"data-vault-metadata-audit-operation": {
		"pkg/data_vault/types.go", StorageOffChain, "v1", "data-vault blob metadata", "BlobMetadata JSON additive field", MigrationAdditiveNoRewrite, "",
		[]string{"pkg/data_vault/types.go", "pkg/data_vault/store.go"}, "",
	},
	"data-vault-audit-chain": {
		"pkg/data_vault/fixture_audit_store.go", StorageFixtureOnly, "v1", "data-vault audit chain", "envelope/hash chain", MigrationNewStore, "",
		[]string{"pkg/data_vault/fixture_audit_store.go", "pkg/data_vault/audit_logger.go"}, "",
	},
	"data-vault-erasure-operation-contract": {
		"pkg/data_vault/erasure_coordinator.go", StorageOffChain, "v1", "backend-neutral erasure operation", "ErasureOperation journal + storage/KMS receipts + replay transaction", MigrationIntegrationRequired, "",
		[]string{"pkg/data_vault/erasure_coordinator.go", "pkg/data_vault/contracts/evidence_lifecycle.go"}, "production erasure requires durable transactional ErasureOperationStore; production storage deletion and KMS destruction adapters; independent receipt signers/key resolver; consent, hold, backup and finalization authorities; authenticated restore-manifest authority",
	},
	"fundauth-replay-store": {
		"pkg/fundauth/keeper/keeper.go", StorageConsensus, "v1", "virtengine/fundauth/authorization/v1", "domain-separated length-prefixed KV key + 32-byte auth digest", MigrationIntegrationRequired, "",
		[]string{"pkg/fundauth/keeper/keeper.go"}, "T4 must mount and wire the fundauth keeper",
	},
	"veid-evidence-reference": {
		"x/veid/keeper/evidence_object_store.go", StorageConsensus, "v1", "VEID evidence object reference", "0xD4 || ASCII(\"reference/\") || objectCommitment -> JSON EvidenceObjectRef v1", MigrationRequiredUnwired, "veid-evidence-object-v1",
		[]string{"pkg/data_vault/contracts/evidence_lifecycle.go", "x/veid/keeper/evidence_object_store.go", "x/veid/types/keys.go"}, "T4 upgrade owner must register the VEID evidence migration",
	},
	"veid-evidence-quarantine": {
		"x/veid/keeper/evidence_object_migration.go", StorageConsensus, "v1", "VEID evidence object quarantine", "0xD4 || ASCII(\"quarantine/\") || sourceKeyDigest -> JSON evidenceMigrationQuarantine v1", MigrationRequiredUnwired, "veid-evidence-object-v1",
		[]string{"x/veid/keeper/evidence_object_migration.go", "x/veid/types/keys.go"}, "T4 upgrade owner must register the VEID evidence migration",
	},
	"veid-evidence-migration-bookkeeping": {
		"x/veid/keeper/evidence_object_migration.go", StorageConsensus, "v1", "VEID evidence object migration", "D4 01 || UTF-8 upgradeID -> replay digest; D4 02 || upgradeID -> JSON report; singleton D4 03 -> manifest digest; D4 04 || UTF-8 signerKeyID -> BE64 epoch floor", MigrationRequiredUnwired, "veid-evidence-object-v1",
		[]string{"x/veid/keeper/evidence_object_migration.go", "x/veid/types/keys.go"}, "T4 upgrade owner must register the VEID evidence migration",
	},
	"veid-legacy-payload-cutover": {
		"x/veid/keeper/evidence_object_migration.go", StorageConsensus, "v1", "legacy VEID evidence payload cutover", "legacy PrefixScope 0x02 rows and classified shared-prefix 0x9A rows retain encrypted_payload until sanitize/delete cutover", MigrationRequiredUnwired, "veid-legacy-payload-cutover-v1",
		[]string{"x/veid/keeper/evidence_object_migration.go", "x/veid/keeper/evidence_object_migration_test.go"}, "separate T4 migration must sanitize/delete legacy PrefixScope 0x02 rows and classified shared-prefix 0x9A rows",
	},
	"encryption-envelope-v2": {
		"x/encryption/types/envelope.go", StorageSerialized, "v2", "VirtEngine encrypted payload", "authenticated envelope", MigrationExplicitReencrypt, "",
		[]string{"x/encryption/crypto/envelope.go", "x/encryption/types/envelope.go"}, "",
	},
	"mfa-otp-verifier-receipt": {
		"x/mfa/keeper/verification.go", StorageTransient, "v1", "virtengine/mfa/otp-verifier-receipt/v1", "signed receipt", MigrationNone, "",
		[]string{"x/mfa/keeper/verification.go"}, "",
	},
}

var requiredNonPersistentSurfaces = []string{
	"organization-contracts",
	"platformsecurity-dependency-manifest",
	"issuerlink-contracts",
	"uniqueness-contracts",
	"biometric-incident-contracts",
	"backendprofile-contracts",
	"federation-contracts",
	"privileged-role-contracts",
	"mfa-prototype-contracts",
}

var requiredPersistedSchemaBlockers = []string{
	"T4 upgrade owner must register the VEID evidence migration",
	"separate T4 migration must sanitize/delete legacy PrefixScope 0x02 rows and classified shared-prefix 0x9A rows",
	"T4 must mount and wire the fundauth keeper",
	"production erasure requires durable transactional ErasureOperationStore; production storage deletion and KMS destruction adapters; independent receipt signers/key resolver; consent, hold, backup and finalization authorities; authenticated restore-manifest authority",
}

func (inventory PersistedSchemaInventory) Validate() error {
	if inventory.Version != PersistedSchemaInventoryVersion || inventory.Thread != PersistedSchemaThread {
		return errors.New("unknown persisted-schema inventory version or thread")
	}
	if inventory.FrozenBaseline != PersistedSchemaFrozenBaseline || inventory.IntakeBaseSHA != PersistedSchemaEpochBase {
		return errors.New("persisted-schema inventory provenance changed")
	}
	if inventory.PayloadHead != PersistedSchemaPayloadHead {
		return errors.New("payload_head must match the inventoried code boundary")
	}
	if err := validatePersistedSchemaEntries(inventory.Entries); err != nil {
		return err
	}
	if !sameStrings(inventory.NonPersistentSurfaces, requiredNonPersistentSurfaces) {
		return errors.New("non_persistent_surfaces must contain the exact required IDs")
	}
	if !sameStrings(inventory.Blockers, requiredPersistedSchemaBlockers) {
		return errors.New("blockers must contain the exact required global blockers")
	}
	return nil
}

func validatePersistedSchemaEntries(entries []PersistedSchemaEntry) error {
	seen := make(map[string]bool, len(entries))
	for _, entry := range entries {
		spec, ok := requiredPersistedSchemas[entry.ID]
		if !ok {
			return fmt.Errorf("unknown persisted-schema entry %q", entry.ID)
		}
		if seen[entry.ID] {
			return fmt.Errorf("duplicate persisted-schema entry %q", entry.ID)
		}
		seen[entry.ID] = true
		if err := validateRepositoryPath(entry.OwnerPath); err != nil {
			return fmt.Errorf("entry %q owner_path: %w", entry.ID, err)
		}
		if len(entry.EvidencePaths) == 0 {
			return fmt.Errorf("entry %q requires evidence paths", entry.ID)
		}
		for _, evidence := range entry.EvidencePaths {
			if err := validateRepositoryPath(evidence); err != nil {
				return fmt.Errorf("entry %q evidence path: %w", entry.ID, err)
			}
		}
		if !sameStrings(entry.EvidencePaths, spec.evidence) {
			return fmt.Errorf("entry %q has incorrect or duplicate evidence", entry.ID)
		}
		if entry.OwnerPath != spec.owner || entry.StorageClass != spec.class || entry.SchemaVersion != spec.version ||
			entry.Domain != spec.domain || entry.KeyOrFormat != spec.format || entry.MigrationDisposition != spec.disposition ||
			entry.MigrationID != spec.migrationID || entry.Blocker != spec.blocker {
			return fmt.Errorf("entry %q does not match the frozen producer specification", entry.ID)
		}
		switch entry.MigrationDisposition {
		case MigrationRequiredUnwired:
			if entry.MigrationID == "" || entry.Blocker == "" {
				return fmt.Errorf("entry %q requires a migration ID and blocker", entry.ID)
			}
		case MigrationIntegrationRequired:
			if entry.Blocker == "" {
				return fmt.Errorf("entry %q requires an integration blocker", entry.ID)
			}
		case MigrationExplicitReencrypt, MigrationAdditiveNoRewrite, MigrationNewStore, MigrationNone:
		default:
			return fmt.Errorf("entry %q has unknown migration disposition %q", entry.ID, entry.MigrationDisposition)
		}
		switch entry.StorageClass {
		case StorageConsensus, StorageOffChain, StorageSerialized, StorageFixtureOnly, StorageTransient:
		default:
			return fmt.Errorf("entry %q has unknown storage class %q", entry.ID, entry.StorageClass)
		}
		if entry.StorageClass == StorageConsensus && nonProductionConsensusOwnerPath(entry.OwnerPath) {
			return fmt.Errorf("consensus entry %q owner_path must identify production code", entry.ID)
		}
	}
	for id := range requiredPersistedSchemas {
		if !seen[id] {
			return fmt.Errorf("required persisted-schema entry %q is missing", id)
		}
	}
	return nil
}

func validateRepositoryPath(value string) error {
	if value == "" || strings.Contains(value, "\\") || strings.Contains(value, ":") || strings.HasPrefix(value, "/") || path.Clean(value) != value || value == "." || strings.HasPrefix(value, "../") {
		return errors.New("must be a clean repository-relative slash path")
	}
	return nil
}

func nonProductionConsensusOwnerPath(value string) bool {
	return strings.HasPrefix(value, "tests/integration/") || strings.Contains(value, "/testdata/") || strings.HasSuffix(value, "_test.go") || strings.Contains(value, "/fixture_")
}

func ParsePersistedSchemaInventory(data []byte) (PersistedSchemaInventory, error) {
	var inventory PersistedSchemaInventory
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&inventory); err != nil {
		return PersistedSchemaInventory{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return PersistedSchemaInventory{}, errors.New("persisted-schema inventory must contain exactly one JSON value")
	}
	if err := inventory.Validate(); err != nil {
		return PersistedSchemaInventory{}, err
	}
	return inventory, nil
}

func (inventory PersistedSchemaInventory) CanonicalDigest() ([sha256.Size]byte, error) {
	if err := inventory.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	canonical, err := json.Marshal(inventory)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	return sha256.Sum256(canonical), nil
}
