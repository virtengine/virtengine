package keeper

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
	"github.com/virtengine/virtengine/x/veid/types"
)

const (
	EvidenceMigrationManifestVersion uint32 = 1
	EvidenceMigrationGenesisMarker          = "genesis"
)

type EvidenceMigrationEntry struct {
	LegacyEnvelopeHash string                      `json:"legacy_envelope_hash"`
	Reference          contracts.EvidenceObjectRef `json:"reference"`
}

type EvidenceMigrationManifest struct {
	Version                uint32                   `json:"version"`
	ChainID                string                   `json:"chain_id"`
	UpgradeID              string                   `json:"upgrade_id"`
	TargetHeight           int64                    `json:"target_height"`
	SourceSnapshotDigest   string                   `json:"source_snapshot_digest"`
	PreviousManifestDigest string                   `json:"previous_manifest_digest"`
	ManifestDigest         string                   `json:"manifest_digest"`
	Entries                []EvidenceMigrationEntry `json:"entries"`
	SignerKeyID            string                   `json:"signer_key_id"`
	SignerKeyEpoch         uint64                   `json:"signer_key_epoch"`
	Signature              []byte                   `json:"signature"`
}

type EvidenceMigrationKeyResolver interface {
	ResolveEvidenceMigrationKey(keyID string, keyEpoch uint64) (ed25519.PublicKey, error)
}

type EvidenceMigrationReport struct {
	Scanned                   uint64 `json:"scanned"`
	ReferenceCreated          uint64 `json:"reference_created"`
	AlreadyMigrated           uint64 `json:"already_migrated"`
	MissingMapping            uint64 `json:"missing_mapping"`
	Quarantined               uint64 `json:"quarantined"`
	MissingMappingQuarantined uint64 `json:"missing_mapping_quarantined"`
	LegacyRowsSanitized       uint64 `json:"legacy_rows_sanitized"`
	LegacyRowsPendingCutover  uint64 `json:"legacy_rows_pending_cutover"`
	Ambiguous                 uint64 `json:"ambiguous"`
	EvidenceRecordRowsSkipped uint64 `json:"evidence_record_rows_skipped"`
}

type evidenceMigrationQuarantine struct {
	Version            uint32 `json:"version"`
	SourceKind         string `json:"source_kind"`
	SourceKeyDigest    string `json:"source_key_digest"`
	LegacyEnvelopeHash string `json:"legacy_envelope_hash"`
	SourceRowDigest    string `json:"source_row_digest"`
	Classification     string `json:"classification"`
	Availability       string `json:"availability"`
	State              string `json:"state"`
}

type legacyEvidenceRow struct {
	sourceKind     string
	key            []byte
	envelopeHash   string
	rowDigest      string
	classification string
}

func (m EvidenceMigrationManifest) ComputeDigest() (string, error) {
	if m.Version != EvidenceMigrationManifestVersion || m.ChainID == "" || m.UpgradeID == "" || m.TargetHeight <= 0 {
		return "", errors.New("incomplete evidence migration context")
	}
	if err := validateSHA256Digest(m.SourceSnapshotDigest, "source snapshot digest"); err != nil {
		return "", err
	}
	if m.PreviousManifestDigest != EvidenceMigrationGenesisMarker {
		if err := validateSHA256Digest(m.PreviousManifestDigest, "previous manifest digest"); err != nil {
			return "", err
		}
	}
	entries := append([]EvidenceMigrationEntry(nil), m.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].LegacyEnvelopeHash < entries[j].LegacyEnvelopeHash })
	hash := sha256.New()
	writeMigrationValue(hash, "virtengine/evidence-migration-manifest/v1")
	writeMigrationValue(hash, fmt.Sprint(m.Version))
	writeMigrationValue(hash, m.ChainID)
	writeMigrationValue(hash, m.UpgradeID)
	writeMigrationValue(hash, fmt.Sprint(m.TargetHeight))
	writeMigrationValue(hash, m.SourceSnapshotDigest)
	writeMigrationValue(hash, m.PreviousManifestDigest)
	for _, entry := range entries {
		if err := validateMigrationEntry(entry); err != nil {
			return "", err
		}
		bz, err := json.Marshal(entry)
		if err != nil {
			return "", fmt.Errorf("marshal migration entry: %w", err)
		}
		writeMigrationBytes(hash, bz)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (m EvidenceMigrationManifest) CanonicalSignBytes() ([]byte, error) {
	digest, err := m.ComputeDigest()
	if err != nil {
		return nil, err
	}
	if m.ManifestDigest != digest || m.SignerKeyID == "" || m.SignerKeyEpoch == 0 {
		return nil, errors.New("invalid evidence migration manifest metadata")
	}
	hash := sha256.New()
	writeMigrationValue(hash, "virtengine/evidence-migration-manifest-signature/v1")
	writeMigrationValue(hash, fmt.Sprint(m.Version))
	writeMigrationValue(hash, m.ChainID)
	writeMigrationValue(hash, m.UpgradeID)
	writeMigrationValue(hash, fmt.Sprint(m.TargetHeight))
	writeMigrationValue(hash, m.SourceSnapshotDigest)
	writeMigrationValue(hash, m.PreviousManifestDigest)
	writeMigrationValue(hash, m.ManifestDigest)
	writeMigrationValue(hash, m.SignerKeyID)
	writeMigrationValue(hash, fmt.Sprint(m.SignerKeyEpoch))
	return hash.Sum(nil), nil
}

func (m EvidenceMigrationManifest) Validate(resolver EvidenceMigrationKeyResolver) error {
	if resolver == nil {
		return errors.New("evidence migration key resolver is required")
	}
	if len(m.Signature) != ed25519.SignatureSize {
		return errors.New("invalid evidence migration manifest signature size")
	}
	seen := make(map[string]struct{}, len(m.Entries))
	for _, entry := range m.Entries {
		if _, duplicate := seen[entry.LegacyEnvelopeHash]; duplicate {
			return errors.New("duplicate legacy envelope hash in migration manifest")
		}
		seen[entry.LegacyEnvelopeHash] = struct{}{}
	}
	signBytes, err := m.CanonicalSignBytes()
	if err != nil {
		return err
	}
	publicKey, err := resolver.ResolveEvidenceMigrationKey(m.SignerKeyID, m.SignerKeyEpoch)
	if err != nil {
		return fmt.Errorf("resolve evidence migration key: %w", err)
	}
	if len(publicKey) != ed25519.PublicKeySize || !ed25519.Verify(publicKey, signBytes, m.Signature) {
		return errors.New("invalid evidence migration manifest signature")
	}
	return nil
}

func (k Keeper) MigrateEvidenceObjects(ctx sdk.Context, manifest EvidenceMigrationManifest, resolver EvidenceMigrationKeyResolver) (EvidenceMigrationReport, error) {
	if err := manifest.Validate(resolver); err != nil {
		return EvidenceMigrationReport{}, err
	}
	replayDigest, err := manifestReplayDigest(manifest)
	if err != nil {
		return EvidenceMigrationReport{}, err
	}
	if manifest.ChainID != ctx.ChainID() || manifest.TargetHeight != ctx.BlockHeight() || manifest.UpgradeID == "" {
		return EvidenceMigrationReport{}, errors.New("evidence migration manifest does not match execution context")
	}
	store := ctx.KVStore(k.skey)
	consumedKey := evidenceMigrationUpgradeKey(types.PrefixEvidenceMigrationConsumed, manifest.UpgradeID)
	if consumed := store.Get(consumedKey); consumed != nil {
		if string(consumed) != replayDigest {
			return EvidenceMigrationReport{}, errors.New("different manifest already consumed for upgrade")
		}
		var report EvidenceMigrationReport
		if err := json.Unmarshal(store.Get(evidenceMigrationUpgradeKey(types.PrefixEvidenceMigrationReport, manifest.UpgradeID)), &report); err != nil {
			return EvidenceMigrationReport{}, fmt.Errorf("decode stored evidence migration report: %w", err)
		}
		return report, nil
	}
	epochKey := evidenceMigrationUpgradeKey(types.PrefixEvidenceMigrationSignerEpoch, manifest.SignerKeyID)
	if floor := store.Get(epochKey); floor != nil {
		if len(floor) != 8 {
			return EvidenceMigrationReport{}, errors.New("evidence migration signer key epoch state is corrupt")
		}
		if manifest.SignerKeyEpoch < binary.BigEndian.Uint64(floor) {
			return EvidenceMigrationReport{}, errors.New("evidence migration signer key epoch rollback")
		}
	}
	latest := store.Get(types.KeyEvidenceMigrationLatest)
	if latest == nil {
		if manifest.PreviousManifestDigest != EvidenceMigrationGenesisMarker {
			return EvidenceMigrationReport{}, errors.New("first evidence migration must use genesis predecessor")
		}
	} else if manifest.PreviousManifestDigest != string(latest) {
		return EvidenceMigrationReport{}, errors.New("evidence migration predecessor mismatch")
	}
	rows, evidenceRowsSkipped := collectLegacyEvidenceRows(ctx, k)
	snapshotDigest := computeLegacyEvidenceSnapshotDigest(rows)
	if manifest.SourceSnapshotDigest != snapshotDigest {
		return EvidenceMigrationReport{}, errors.New("evidence migration source snapshot mismatch")
	}
	mappings := make(map[string]contracts.EvidenceObjectRef, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		mappings[entry.LegacyEnvelopeHash] = entry.Reference
	}

	cacheCtx, commit := ctx.CacheContext()
	report := EvidenceMigrationReport{Scanned: uint64(len(rows)), EvidenceRecordRowsSkipped: evidenceRowsSkipped}
	for _, row := range rows {
		if row.classification != "legacy" {
			if err := setEvidenceMigrationQuarantine(cacheCtx, k, row); err != nil {
				return EvidenceMigrationReport{}, err
			}
			report.Quarantined++
			report.LegacyRowsPendingCutover++
			if row.classification == "ambiguous" {
				report.Ambiguous++
			}
			continue
		}
		ref, mapped := mappings[row.envelopeHash]
		if !mapped {
			if err := setEvidenceMigrationQuarantine(cacheCtx, k, row); err != nil {
				return EvidenceMigrationReport{}, err
			}
			report.MissingMapping++
			report.Quarantined++
			report.MissingMappingQuarantined++
			report.LegacyRowsPendingCutover++
			continue
		}
		if existing, found := k.GetEvidenceObjectRef(cacheCtx, ref.ObjectCommitment); found {
			if existing != ref {
				return EvidenceMigrationReport{}, errors.New("existing evidence commitment has different reference metadata")
			}
			report.AlreadyMigrated++
		} else {
			if err := k.SetEvidenceObjectRef(cacheCtx, ref); err != nil {
				return EvidenceMigrationReport{}, err
			}
			report.ReferenceCreated++
		}
		report.LegacyRowsPendingCutover++
	}
	reportBytes, err := json.Marshal(report)
	if err != nil {
		return EvidenceMigrationReport{}, fmt.Errorf("marshal evidence migration report: %w", err)
	}
	cacheStore := cacheCtx.KVStore(k.skey)
	cacheStore.Set(consumedKey, []byte(replayDigest))
	cacheStore.Set(evidenceMigrationUpgradeKey(types.PrefixEvidenceMigrationReport, manifest.UpgradeID), reportBytes)
	epochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(epochBytes, manifest.SignerKeyEpoch)
	cacheStore.Set(epochKey, epochBytes)
	cacheStore.Set(types.KeyEvidenceMigrationLatest, []byte(manifest.ManifestDigest))
	commit()
	return report, nil
}

func collectLegacyEvidenceRows(ctx sdk.Context, k Keeper) ([]legacyEvidenceRow, uint64) {
	store := ctx.KVStore(k.skey)
	rows := make([]legacyEvidenceRow, 0)
	var evidenceRowsSkipped uint64
	for _, source := range []struct {
		kind   string
		prefix []byte
	}{{"scope", types.PrefixScope}, {"social_scope", types.PrefixSocialMediaScope}} {
		iterator := storetypes.KVStorePrefixIterator(store, source.prefix)
		for ; iterator.Valid(); iterator.Next() {
			row, classification := classifyLegacyEvidenceRow(source.kind, iterator.Key(), iterator.Value())
			if classification == "evidence_record" {
				evidenceRowsSkipped++
				continue
			}
			if classification != "none" {
				rows = append(rows, row)
			}
		}
		iterator.Close()
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].sourceKind != rows[j].sourceKind {
			return rows[i].sourceKind < rows[j].sourceKind
		}
		return string(rows[i].key) < string(rows[j].key)
	})
	return rows, evidenceRowsSkipped
}

func setEvidenceMigrationQuarantine(ctx sdk.Context, k Keeper, row legacyEvidenceRow) error {
	keyDigest := sha256.Sum256(append(append([]byte(row.sourceKind), 0), row.key...))
	record := evidenceMigrationQuarantine{
		Version: EvidenceMigrationManifestVersion, SourceKind: row.sourceKind,
		SourceKeyDigest: hex.EncodeToString(keyDigest[:]), LegacyEnvelopeHash: row.envelopeHash,
		SourceRowDigest: row.rowDigest, Classification: row.classification,
		Availability: "unavailable", State: string(contracts.RetentionDeletionUnresolved),
	}
	bz, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal evidence migration quarantine: %w", err)
	}
	if err := validatePayloadFreeEvidenceJSON(bz); err != nil {
		return err
	}
	key := append(append([]byte(nil), types.PrefixEvidenceObjectRef...), []byte("quarantine/"+record.SourceKeyDigest)...)
	ctx.KVStore(k.skey).Set(key, bz)
	return nil
}

func validateMigrationEntry(entry EvidenceMigrationEntry) error {
	if err := validateSHA256Digest(entry.LegacyEnvelopeHash, "legacy envelope hash"); err != nil {
		return err
	}
	if err := entry.Reference.Validate(); err != nil {
		return fmt.Errorf("validate mapped evidence reference: %w", err)
	}
	return nil
}

func classifyLegacyEvidenceRow(sourceKind string, key, value []byte) (legacyEvidenceRow, string) {
	rowHash := sha256.Sum256(value)
	row := legacyEvidenceRow{
		sourceKind: sourceKind, key: append([]byte(nil), key...), rowDigest: hex.EncodeToString(rowHash[:]),
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(value, &fields); err != nil || fields == nil {
		row.classification = "malformed"
		row.envelopeHash = row.rowDigest
		return row, row.classification
	}
	evidenceRecord := hasJSONFields(fields, "evidence_id", "evidence_type", "account_address", "scope_id", "content_hash", "envelope_hash", "status")
	legacyScope := hasJSONFields(fields, "scope_id", "encrypted_payload")
	if sourceKind == "social_scope" {
		socialScope := hasJSONFields(fields, "version", "scope_id", "account_address", "provider", "encrypted_payload")
		switch {
		case evidenceRecord && socialScope:
			row.classification = "ambiguous"
			row.envelopeHash = row.rowDigest
			return row, row.classification
		case evidenceRecord:
			return row, "evidence_record"
		case socialScope:
			legacyScope = true
		default:
			row.classification = "ambiguous"
			row.envelopeHash = row.rowDigest
			return row, row.classification
		}
	}
	if !legacyScope {
		row.classification = "malformed"
		row.envelopeHash = row.rowDigest
		return row, row.classification
	}
	payload := fields["encrypted_payload"]
	if len(payload) == 0 || string(payload) == "null" {
		return row, "none"
	}
	payloadHash := sha256.Sum256(payload)
	row.classification = "legacy"
	row.envelopeHash = hex.EncodeToString(payloadHash[:])
	return row, row.classification
}

func hasJSONFields(fields map[string]json.RawMessage, names ...string) bool {
	for _, name := range names {
		if _, found := fields[name]; !found {
			return false
		}
	}
	return true
}

func computeLegacyEvidenceSnapshotDigest(rows []legacyEvidenceRow) string {
	hash := sha256.New()
	writeMigrationValue(hash, "virtengine/evidence-migration-source-snapshot/v1")
	for _, row := range rows {
		writeMigrationValue(hash, row.sourceKind)
		writeMigrationBytes(hash, row.key)
		writeMigrationValue(hash, row.rowDigest)
		writeMigrationValue(hash, row.classification)
		writeMigrationValue(hash, row.envelopeHash)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func manifestReplayDigest(manifest EvidenceMigrationManifest) (string, error) {
	signBytes, err := manifest.CanonicalSignBytes()
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	writeMigrationValue(hash, "virtengine/evidence-migration-replay/v1")
	writeMigrationBytes(hash, signBytes)
	writeMigrationBytes(hash, manifest.Signature)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func evidenceMigrationUpgradeKey(prefix []byte, identifier string) []byte {
	key := make([]byte, 0, len(prefix)+len(identifier))
	key = append(key, prefix...)
	return append(key, identifier...)
}

func validateSHA256Digest(value, name string) error {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size || hex.EncodeToString(decoded) != value {
		return fmt.Errorf("%s must be lowercase SHA-256 hex", name)
	}
	return nil
}

func writeMigrationValue(writer interface{ Write([]byte) (int, error) }, value string) {
	writeMigrationBytes(writer, []byte(value))
}

func writeMigrationBytes(writer interface{ Write([]byte) (int, error) }, value []byte) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = writer.Write(size[:])
	_, _ = writer.Write(value)
}
