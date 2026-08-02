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
	"strings"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
	"github.com/virtengine/virtengine/x/veid/types"
)

const (
	EvidencePayloadCutoverManifestVersion uint32 = 1
	EvidencePayloadCutoverGenesisMarker          = "genesis"

	EvidencePayloadCutoverActionSanitize = "sanitize"
	EvidencePayloadCutoverActionDelete   = "delete"
)

// These local keys reserve new children of the D4 evidence-object namespace without
// changing keys.go: 05=cutover consumed, 06=report, 07=latest, 08=signer epoch.
var (
	prefixEvidencePayloadCutoverConsumed    = []byte{0xD4, 0x05}
	prefixEvidencePayloadCutoverReport      = []byte{0xD4, 0x06}
	keyEvidencePayloadCutoverLatest         = []byte{0xD4, 0x07}
	prefixEvidencePayloadCutoverSignerEpoch = []byte{0xD4, 0x08}
)

type EvidencePayloadCutoverEntry struct {
	SourceKind            string `json:"source_kind"`
	SourceKeyDigest       string `json:"source_key_digest"`
	SourceRowDigest       string `json:"source_row_digest"`
	Action                string `json:"action"`
	AuthorityRecordDigest string `json:"authority_record_digest"`
	// For sanitize, this signed commitment associates the source row with the
	// exact stored reference bytes bound by AuthorityRecordDigest.
	ObjectCommitment string `json:"object_commitment,omitempty"`
}

type EvidencePayloadCutoverManifest struct {
	Version                             uint32                        `json:"version"`
	ChainID                             string                        `json:"chain_id"`
	CutoverID                           string                        `json:"cutover_id"`
	TargetHeight                        int64                         `json:"target_height"`
	SourceSnapshotDigest                string                        `json:"source_snapshot_digest"`
	PrerequisiteMigrationManifestDigest string                        `json:"prerequisite_evidence_migration_manifest_digest"`
	PreviousCutoverManifestDigest       string                        `json:"previous_cutover_manifest_digest"`
	ManifestDigest                      string                        `json:"manifest_digest"`
	SignerKeyID                         string                        `json:"signer_key_id"`
	SignerKeyEpoch                      uint64                        `json:"signer_key_epoch"`
	Entries                             []EvidencePayloadCutoverEntry `json:"entries"`
	Signature                           []byte                        `json:"signature"`
}

type EvidencePayloadCutoverReport struct {
	Scanned                uint64 `json:"scanned"`
	Sanitized              uint64 `json:"sanitized"`
	Deleted                uint64 `json:"deleted"`
	EvidenceRecordsSkipped uint64 `json:"evidence_records_skipped"`
	AlreadyCutover         uint64 `json:"already_cutover"`
	ScopeSources           uint64 `json:"scope_sources"`
	SocialScopeSources     uint64 `json:"social_scope_sources"`
}

type evidencePayloadCutoverSkippedRecord struct {
	sourceKind     string
	keyDigest      string
	rowDigest      string
	classification string
}

func (m EvidencePayloadCutoverManifest) ComputeDigest() (string, error) {
	if m.Version != EvidencePayloadCutoverManifestVersion || m.ChainID == "" || m.CutoverID == "" || m.TargetHeight <= 0 {
		return "", errors.New("incomplete evidence payload cutover context")
	}
	if err := validateSHA256Digest(m.SourceSnapshotDigest, "source snapshot digest"); err != nil {
		return "", err
	}
	if err := validateSHA256Digest(m.PrerequisiteMigrationManifestDigest, "prerequisite migration manifest digest"); err != nil {
		return "", err
	}
	if m.PreviousCutoverManifestDigest != EvidencePayloadCutoverGenesisMarker {
		if err := validateSHA256Digest(m.PreviousCutoverManifestDigest, "previous cutover manifest digest"); err != nil {
			return "", err
		}
	}
	entries := canonicalEvidencePayloadCutoverEntries(m.Entries)
	hash := sha256.New()
	writeMigrationValue(hash, "virtengine/evidence-payload-cutover-manifest/v1")
	writeMigrationValue(hash, fmt.Sprint(m.Version))
	writeMigrationValue(hash, m.ChainID)
	writeMigrationValue(hash, m.CutoverID)
	writeMigrationValue(hash, fmt.Sprint(m.TargetHeight))
	writeMigrationValue(hash, m.SourceSnapshotDigest)
	writeMigrationValue(hash, m.PrerequisiteMigrationManifestDigest)
	writeMigrationValue(hash, m.PreviousCutoverManifestDigest)
	for _, entry := range entries {
		if err := validateEvidencePayloadCutoverEntry(entry); err != nil {
			return "", err
		}
		writeMigrationValue(hash, entry.SourceKind)
		writeMigrationValue(hash, entry.SourceKeyDigest)
		writeMigrationValue(hash, entry.SourceRowDigest)
		writeMigrationValue(hash, entry.Action)
		writeMigrationValue(hash, entry.AuthorityRecordDigest)
		writeMigrationValue(hash, entry.ObjectCommitment)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (m EvidencePayloadCutoverManifest) CanonicalSignBytes() ([]byte, error) {
	digest, err := m.ComputeDigest()
	if err != nil {
		return nil, err
	}
	if m.ManifestDigest != digest || m.SignerKeyID == "" || m.SignerKeyEpoch == 0 {
		return nil, errors.New("invalid evidence payload cutover manifest metadata")
	}
	hash := sha256.New()
	writeMigrationValue(hash, "virtengine/evidence-payload-cutover-signature/v1")
	writeMigrationValue(hash, fmt.Sprint(m.Version))
	writeMigrationValue(hash, m.ChainID)
	writeMigrationValue(hash, m.CutoverID)
	writeMigrationValue(hash, fmt.Sprint(m.TargetHeight))
	writeMigrationValue(hash, m.SourceSnapshotDigest)
	writeMigrationValue(hash, m.PrerequisiteMigrationManifestDigest)
	writeMigrationValue(hash, m.PreviousCutoverManifestDigest)
	writeMigrationValue(hash, m.ManifestDigest)
	writeMigrationValue(hash, m.SignerKeyID)
	writeMigrationValue(hash, fmt.Sprint(m.SignerKeyEpoch))
	return hash.Sum(nil), nil
}

func (m EvidencePayloadCutoverManifest) Validate(resolver EvidenceMigrationKeyResolver) error {
	if resolver == nil {
		return errors.New("evidence payload cutover key resolver is required")
	}
	if len(m.Signature) != ed25519.SignatureSize {
		return errors.New("invalid evidence payload cutover signature size")
	}
	seen := make(map[string]struct{}, len(m.Entries))
	for _, entry := range m.Entries {
		identity := entry.SourceKind + "\x00" + entry.SourceKeyDigest
		if _, duplicate := seen[identity]; duplicate {
			return errors.New("duplicate evidence payload cutover entry")
		}
		seen[identity] = struct{}{}
	}
	signBytes, err := m.CanonicalSignBytes()
	if err != nil {
		return err
	}
	publicKey, err := resolver.ResolveEvidenceMigrationKey(m.SignerKeyID, m.SignerKeyEpoch)
	if err != nil {
		return fmt.Errorf("resolve evidence payload cutover key: %w", err)
	}
	if len(publicKey) != ed25519.PublicKeySize || !ed25519.Verify(publicKey, signBytes, m.Signature) {
		return errors.New("invalid evidence payload cutover signature")
	}
	return nil
}

func (k Keeper) CutoverEvidencePayloads(ctx sdk.Context, manifest EvidencePayloadCutoverManifest, resolver EvidenceMigrationKeyResolver) (EvidencePayloadCutoverReport, error) {
	if err := manifest.Validate(resolver); err != nil {
		return EvidencePayloadCutoverReport{}, err
	}
	replayDigest, err := evidencePayloadCutoverReplayDigest(manifest)
	if err != nil {
		return EvidencePayloadCutoverReport{}, err
	}
	if manifest.ChainID != ctx.ChainID() || manifest.TargetHeight != ctx.BlockHeight() || manifest.CutoverID == "" {
		return EvidencePayloadCutoverReport{}, errors.New("evidence payload cutover manifest does not match execution context")
	}
	store := ctx.KVStore(k.skey)
	consumedKey := evidencePayloadCutoverBookkeepingKey(prefixEvidencePayloadCutoverConsumed, manifest.CutoverID)
	if consumed := store.Get(consumedKey); consumed != nil {
		if string(consumed) != replayDigest {
			return EvidencePayloadCutoverReport{}, errors.New("different manifest already consumed for cutover")
		}
		reportBytes := store.Get(evidencePayloadCutoverBookkeepingKey(prefixEvidencePayloadCutoverReport, manifest.CutoverID))
		report, err := validateStoredEvidencePayloadCutoverReport(reportBytes)
		if err != nil {
			return EvidencePayloadCutoverReport{}, err
		}
		return report, nil
	}
	if latestMigration := store.Get(types.KeyEvidenceMigrationLatest); latestMigration == nil || manifest.PrerequisiteMigrationManifestDigest != string(latestMigration) {
		return EvidencePayloadCutoverReport{}, errors.New("evidence payload cutover prerequisite migration mismatch")
	}
	if latestCutover := store.Get(keyEvidencePayloadCutoverLatest); latestCutover == nil {
		if manifest.PreviousCutoverManifestDigest != EvidencePayloadCutoverGenesisMarker {
			return EvidencePayloadCutoverReport{}, errors.New("first evidence payload cutover must use genesis predecessor")
		}
	} else if manifest.PreviousCutoverManifestDigest != string(latestCutover) {
		return EvidencePayloadCutoverReport{}, errors.New("evidence payload cutover predecessor mismatch")
	}
	epochKey := evidencePayloadCutoverBookkeepingKey(prefixEvidencePayloadCutoverSignerEpoch, manifest.SignerKeyID)
	if floor := store.Get(epochKey); floor != nil {
		if len(floor) != 8 {
			return EvidencePayloadCutoverReport{}, errors.New("corrupt evidence payload cutover signer key epoch")
		}
		if manifest.SignerKeyEpoch < binary.BigEndian.Uint64(floor) {
			return EvidencePayloadCutoverReport{}, errors.New("evidence payload cutover signer key epoch rollback")
		}
	}
	rows, skippedRecords := collectEvidencePayloadCutoverRows(ctx, k)
	if manifest.SourceSnapshotDigest != computeEvidencePayloadCutoverSnapshotDigest(rows, skippedRecords) {
		return EvidencePayloadCutoverReport{}, errors.New("evidence payload cutover source snapshot mismatch")
	}
	entries := make(map[string]EvidencePayloadCutoverEntry, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		entries[entry.SourceKind+"\x00"+entry.SourceKeyDigest] = entry
	}
	if len(entries) != len(rows) {
		return EvidencePayloadCutoverReport{}, errors.New("evidence payload cutover entry count does not match source rows")
	}

	cacheCtx, commit := ctx.CacheContext()
	report := EvidencePayloadCutoverReport{Scanned: uint64(len(rows)), EvidenceRecordsSkipped: uint64(len(skippedRecords))}
	for _, row := range rows {
		keyDigest := evidencePayloadCutoverSourceKeyDigest(row.sourceKind, row.key)
		entry, found := entries[row.sourceKind+"\x00"+keyDigest]
		if !found {
			return EvidencePayloadCutoverReport{}, errors.New("missing evidence payload cutover entry for source row")
		}
		if entry.SourceRowDigest != row.rowDigest {
			return EvidencePayloadCutoverReport{}, errors.New("evidence payload cutover source row digest mismatch")
		}
		switch row.sourceKind {
		case "scope":
			report.ScopeSources++
		case "social_scope":
			report.SocialScopeSources++
		default:
			return EvidencePayloadCutoverReport{}, errors.New("unsupported evidence payload cutover source kind")
		}
		if err := applyEvidencePayloadCutoverEntry(cacheCtx, k, row, entry); err != nil {
			return EvidencePayloadCutoverReport{}, err
		}
		if entry.Action == EvidencePayloadCutoverActionSanitize {
			report.Sanitized++
		} else {
			report.Deleted++
		}
	}
	reportBytes, err := json.Marshal(report)
	if err != nil {
		return EvidencePayloadCutoverReport{}, fmt.Errorf("marshal evidence payload cutover report: %w", err)
	}
	if err := validatePayloadFreeEvidenceJSON(reportBytes); err != nil {
		return EvidencePayloadCutoverReport{}, err
	}
	cacheStore := cacheCtx.KVStore(k.skey)
	cacheStore.Set(consumedKey, []byte(replayDigest))
	cacheStore.Set(evidencePayloadCutoverBookkeepingKey(prefixEvidencePayloadCutoverReport, manifest.CutoverID), reportBytes)
	cacheStore.Set(keyEvidencePayloadCutoverLatest, []byte(manifest.ManifestDigest))
	epochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(epochBytes, manifest.SignerKeyEpoch)
	cacheStore.Set(epochKey, epochBytes)
	commit()
	return report, nil
}

func applyEvidencePayloadCutoverEntry(ctx sdk.Context, k Keeper, row legacyEvidenceRow, entry EvidencePayloadCutoverEntry) error {
	store := ctx.KVStore(k.skey)
	quarantineKey := append(append([]byte(nil), types.PrefixEvidenceObjectRef...), []byte("quarantine/"+entry.SourceKeyDigest)...)
	quarantineBytes := store.Get(quarantineKey)
	switch entry.Action {
	case EvidencePayloadCutoverActionSanitize:
		if row.classification != "legacy" || entry.ObjectCommitment == "" {
			return errors.New("sanitize requires a mapped legacy row")
		}
		if quarantineBytes != nil {
			if err := validateEvidencePayloadCutoverQuarantine(quarantineBytes, row, entry, "legacy"); err != nil {
				return fmt.Errorf("sanitize quarantine mismatch: %w", err)
			}
		}
		referenceBytes := store.Get(evidenceObjectRefKey(entry.ObjectCommitment))
		if referenceBytes == nil || evidencePayloadCutoverDigest(referenceBytes) != entry.AuthorityRecordDigest {
			return errors.New("sanitize authority reference mismatch")
		}
		var ref contracts.EvidenceObjectRef
		if json.Unmarshal(referenceBytes, &ref) != nil || ref.Validate() != nil || ref.ObjectCommitment != entry.ObjectCommitment || ref.EvidenceDigest == "" {
			return errors.New("sanitize authority reference is invalid")
		}
		sanitized, err := sanitizeLegacyEvidenceRow(row.sourceKind, store.Get(row.key))
		if err != nil {
			return err
		}
		store.Set(row.key, sanitized)
		return nil
	case EvidencePayloadCutoverActionDelete:
		if row.classification == "legacy" {
			return errors.New("legacy rows cannot be deleted without persisted reverse mapping authority")
		}
		if entry.ObjectCommitment != "" || quarantineBytes == nil || evidencePayloadCutoverDigest(quarantineBytes) != entry.AuthorityRecordDigest {
			return errors.New("delete requires exact quarantine authority")
		}
		if err := validateEvidencePayloadCutoverQuarantine(quarantineBytes, row, entry, row.classification); err != nil {
			return err
		}
		if err := validateEvidencePayloadCutoverDeleteSource(row.sourceKind, store.Get(row.key)); err != nil {
			return err
		}
		store.Delete(row.key)
		return nil
	default:
		return errors.New("unsupported evidence payload cutover action")
	}
}

func sanitizeLegacyEvidenceRow(sourceKind string, value []byte) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(value, &fields); err != nil || fields == nil {
		return nil, errors.New("decode legacy evidence row for sanitization")
	}
	payload, found := fields["encrypted_payload"]
	if !found || len(payload) == 0 || string(payload) == "null" {
		return nil, errors.New("legacy evidence row has no payload to sanitize")
	}
	allowed, found := evidencePayloadCutoverSanitizeFields[sourceKind]
	if !found {
		return nil, errors.New("unsupported legacy evidence source kind for sanitization")
	}
	safeFields := make(map[string]json.RawMessage, len(allowed)+1)
	for key, kind := range allowed {
		if value, exists := fields[key]; exists {
			if err := validateEvidencePayloadCutoverField(key, kind, value); err != nil {
				return nil, err
			}
			safeFields[key] = value
		}
	}
	safeFields["encrypted_payload"] = json.RawMessage("null")
	sanitized, err := json.Marshal(safeFields)
	if err != nil {
		return nil, fmt.Errorf("marshal sanitized legacy evidence row: %w", err)
	}
	if err := validateSanitizedEvidencePayloadJSON(sanitized, payload); err != nil {
		return nil, err
	}
	return sanitized, nil
}

type evidencePayloadCutoverFieldKind uint8

const (
	cutoverString evidencePayloadCutoverFieldKind = iota + 1
	cutoverOptionalString
	cutoverUnsigned
	cutoverBoolean
)

var evidencePayloadCutoverSanitizeFields = map[string]map[string]evidencePayloadCutoverFieldKind{
	"scope": {
		"scope_id": cutoverString, "account_address": cutoverString, "scope_type": cutoverString,
		"content_hash": cutoverString, "envelope_hash": cutoverString, "status": cutoverString,
		"created_at": cutoverString, "updated_at": cutoverString, "uploaded_at": cutoverString,
		"verified_at": cutoverOptionalString, "expires_at": cutoverOptionalString,
		"revoked": cutoverBoolean, "version": cutoverUnsigned,
	},
	"social_scope": {
		"version": cutoverUnsigned, "scope_id": cutoverString, "account_address": cutoverString,
		"provider": cutoverString, "profile_name_hash": cutoverString, "email_hash": cutoverString,
		"username_hash": cutoverString, "org_hash": cutoverString, "account_created_at": cutoverOptionalString,
		"account_age_days": cutoverUnsigned, "is_verified": cutoverBoolean, "friend_count_range": cutoverString,
		"status": cutoverString, "created_at": cutoverString, "updated_at": cutoverString, "evidence_hash": cutoverString,
	},
}

func validateEvidencePayloadCutoverField(name string, kind evidencePayloadCutoverFieldKind, raw json.RawMessage) error {
	switch kind {
	case cutoverString, cutoverOptionalString:
		if kind == cutoverOptionalString && string(raw) == "null" {
			return nil
		}
		var value string
		if json.Unmarshal(raw, &value) != nil || value == "" || len(value) > 512 {
			return fmt.Errorf("sanitized field %s must be a bounded nonempty string", name)
		}
		for _, char := range value {
			if char < 0x20 {
				return fmt.Errorf("sanitized field %s contains control data", name)
			}
		}
	case cutoverUnsigned:
		var value uint64
		if json.Unmarshal(raw, &value) != nil {
			return fmt.Errorf("sanitized field %s must be an unsigned integer", name)
		}
	case cutoverBoolean:
		var value bool
		if json.Unmarshal(raw, &value) != nil {
			return fmt.Errorf("sanitized field %s must be boolean", name)
		}
	default:
		return fmt.Errorf("sanitized field %s has unknown schema type", name)
	}
	return nil
}

func validateEvidencePayloadCutoverQuarantine(value []byte, row legacyEvidenceRow, entry EvidencePayloadCutoverEntry, classification string) error {
	var quarantine evidenceMigrationQuarantine
	if json.Unmarshal(value, &quarantine) != nil || quarantine.SourceKind != row.sourceKind || quarantine.SourceKeyDigest != entry.SourceKeyDigest || quarantine.SourceRowDigest != row.rowDigest || quarantine.LegacyEnvelopeHash != row.envelopeHash || quarantine.Classification != classification {
		return errors.New("quarantine does not describe source row")
	}
	return nil
}

func validateEvidencePayloadCutoverDeleteSource(sourceKind string, value []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(value, &fields); err != nil || fields == nil {
		return errors.New("delete source must be a JSON object")
	}
	payload, found := fields["encrypted_payload"]
	if !found || len(payload) == 0 || string(payload) == "null" {
		return errors.New("delete source must contain a non-null encrypted payload marker")
	}
	if sourceKind != "social_scope" {
		return nil
	}
	evidenceFields := map[string]struct{}{
		"evidence_id": {}, "evidence_type": {},
		"content_hash": {}, "envelope_hash": {}, "recipient_key_ids": {}, "algorithm_id": {},
		"confidence": {}, "provenance_hash": {}, "decision_reason": {},
		"verified_at": {}, "verifier_key_id": {}, "override": {},
	}
	for field := range fields {
		if _, exists := evidenceFields[strings.ToLower(field)]; exists {
			return errors.New("social scope delete source looks evidence-record-like")
		}
	}
	var record types.EvidenceRecord
	if json.Unmarshal(value, &record) == nil && record.Validate() == nil {
		return errors.New("social scope delete source is a valid evidence record")
	}
	return nil
}

func validateSanitizedEvidencePayloadJSON(value, originalPayload []byte) error {
	if len(originalPayload) > 0 && bytesContains(value, originalPayload) {
		return errors.New("sanitized legacy evidence row contains original payload")
	}
	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return errors.New("decode sanitized legacy evidence row")
	}
	if err := scanSanitizedEvidencePayloadJSON(decoded); err != nil {
		return err
	}
	return nil
}

func scanSanitizedEvidencePayloadJSON(value any) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			lower := strings.ToLower(key)
			if lower == "encrypted_payload" {
				if child != nil {
					return errors.New("sanitized encrypted_payload must be null")
				}
				continue
			}
			for _, forbidden := range []string{"ciphertext", "plaintext", "payload", "secret", "storage_ref", "backend", "metadata", "opening", "nonce", "wrapped_key"} {
				if strings.Contains(lower, forbidden) {
					return errors.New("sanitized legacy evidence row contains forbidden field " + key)
				}
			}
			if err := scanSanitizedEvidencePayloadJSON(child); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range typed {
			if err := scanSanitizedEvidencePayloadJSON(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func bytesContains(value, fragment []byte) bool {
	return len(fragment) <= len(value) && strings.Contains(string(value), string(fragment))
}

func validateEvidencePayloadCutoverEntry(entry EvidencePayloadCutoverEntry) error {
	if entry.SourceKind != "scope" && entry.SourceKind != "social_scope" {
		return errors.New("invalid evidence payload cutover source kind")
	}
	if err := validateSHA256Digest(entry.SourceKeyDigest, "source key digest"); err != nil {
		return err
	}
	if err := validateSHA256Digest(entry.SourceRowDigest, "source row digest"); err != nil {
		return err
	}
	if err := validateSHA256Digest(entry.AuthorityRecordDigest, "authority record digest"); err != nil {
		return err
	}
	if entry.Action != EvidencePayloadCutoverActionSanitize && entry.Action != EvidencePayloadCutoverActionDelete {
		return errors.New("invalid evidence payload cutover action")
	}
	if entry.Action == EvidencePayloadCutoverActionSanitize {
		if err := validateSHA256Digest(entry.ObjectCommitment, "object commitment"); err != nil {
			return err
		}
	} else if entry.ObjectCommitment != "" {
		return errors.New("delete entry must not name an object commitment")
	}
	return nil
}

func canonicalEvidencePayloadCutoverEntries(entries []EvidencePayloadCutoverEntry) []EvidencePayloadCutoverEntry {
	canonical := append([]EvidencePayloadCutoverEntry(nil), entries...)
	sort.Slice(canonical, func(i, j int) bool {
		if canonical[i].SourceKind != canonical[j].SourceKind {
			return canonical[i].SourceKind < canonical[j].SourceKind
		}
		return canonical[i].SourceKeyDigest < canonical[j].SourceKeyDigest
	})
	return canonical
}

func collectEvidencePayloadCutoverRows(ctx sdk.Context, k Keeper) ([]legacyEvidenceRow, []evidencePayloadCutoverSkippedRecord) {
	store := ctx.KVStore(k.skey)
	rows := make([]legacyEvidenceRow, 0)
	skipped := make([]evidencePayloadCutoverSkippedRecord, 0)
	for _, source := range []struct {
		kind   string
		prefix []byte
	}{{"scope", types.PrefixScope}, {"social_scope", types.PrefixSocialMediaScope}} {
		iterator := storetypes.KVStorePrefixIterator(store, source.prefix)
		for ; iterator.Valid(); iterator.Next() {
			row, classification := classifyLegacyEvidenceRow(source.kind, iterator.Key(), iterator.Value())
			if classification == "evidence_record" {
				skipped = append(skipped, evidencePayloadCutoverSkippedRecord{sourceKind: source.kind, keyDigest: evidencePayloadCutoverSourceKeyDigest(source.kind, iterator.Key()), rowDigest: evidencePayloadCutoverDigest(iterator.Value()), classification: classification})
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
	sort.Slice(skipped, func(i, j int) bool {
		if skipped[i].sourceKind != skipped[j].sourceKind {
			return skipped[i].sourceKind < skipped[j].sourceKind
		}
		return skipped[i].keyDigest < skipped[j].keyDigest
	})
	return rows, skipped
}

func computeEvidencePayloadCutoverSnapshotDigest(rows []legacyEvidenceRow, skipped []evidencePayloadCutoverSkippedRecord) string {
	hash := sha256.New()
	writeMigrationValue(hash, "virtengine/evidence-payload-cutover-source-snapshot/v2")
	for _, row := range rows {
		writeMigrationValue(hash, row.sourceKind)
		writeMigrationValue(hash, evidencePayloadCutoverSourceKeyDigest(row.sourceKind, row.key))
		writeMigrationValue(hash, row.rowDigest)
		writeMigrationValue(hash, row.classification)
		writeMigrationValue(hash, row.envelopeHash)
	}
	for _, record := range skipped {
		writeMigrationValue(hash, record.sourceKind)
		writeMigrationValue(hash, record.keyDigest)
		writeMigrationValue(hash, record.rowDigest)
		writeMigrationValue(hash, record.classification)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func validateStoredEvidencePayloadCutoverReport(value []byte) (EvidencePayloadCutoverReport, error) {
	if len(value) == 0 {
		return EvidencePayloadCutoverReport{}, errors.New("stored evidence payload cutover report is missing")
	}
	if err := validateSanitizedEvidencePayloadJSON(value, nil); err != nil {
		return EvidencePayloadCutoverReport{}, fmt.Errorf("stored evidence payload cutover report is not payload-free: %w", err)
	}
	var report EvidencePayloadCutoverReport
	if err := json.Unmarshal(value, &report); err != nil {
		return EvidencePayloadCutoverReport{}, fmt.Errorf("decode stored evidence payload cutover report: %w", err)
	}
	if report == (EvidencePayloadCutoverReport{}) || report.Scanned != report.Sanitized+report.Deleted+report.AlreadyCutover || report.ScopeSources+report.SocialScopeSources != report.Scanned {
		return EvidencePayloadCutoverReport{}, errors.New("stored evidence payload cutover report has implausible counts")
	}
	return report, nil
}

func evidencePayloadCutoverSourceKeyDigest(sourceKind string, key []byte) string {
	// Evidence migration quarantine records use the same source identifier.
	digest := sha256.Sum256(append(append([]byte(sourceKind), 0), key...))
	return hex.EncodeToString(digest[:])
}

func evidencePayloadCutoverReplayDigest(manifest EvidencePayloadCutoverManifest) (string, error) {
	signBytes, err := manifest.CanonicalSignBytes()
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	writeMigrationValue(hash, "virtengine/evidence-payload-cutover-replay/v1")
	writeMigrationBytes(hash, signBytes)
	writeMigrationBytes(hash, manifest.Signature)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func evidencePayloadCutoverDigest(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func evidencePayloadCutoverBookkeepingKey(prefix []byte, identifier string) []byte {
	key := make([]byte, 0, len(prefix)+len(identifier))
	key = append(key, prefix...)
	return append(key, identifier...)
}
