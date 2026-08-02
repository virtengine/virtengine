package keeper

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
	"github.com/virtengine/virtengine/x/veid/types"
)

type migrationResolver struct {
	key     ed25519.PublicKey
	private ed25519.PrivateKey
}

func (r migrationResolver) ResolveEvidenceMigrationKey(keyID string, keyEpoch uint64) (ed25519.PublicKey, error) {
	if keyID != "governance" || keyEpoch == 0 {
		return nil, errors.New("key not found")
	}
	return r.key, nil
}

func TestEvidenceObjectStoreIsPayloadFree(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
	if err := k.SetEvidenceObjectRef(ctx, ref); err != nil {
		t.Fatal(err)
	}
	stored, found := k.GetEvidenceObjectRef(ctx, ref.ObjectCommitment)
	if !found || stored != ref {
		t.Fatal("evidence reference did not round trip")
	}
	bz := ctx.KVStore(k.skey).Get(evidenceObjectRefKey(ref.ObjectCommitment))
	for _, forbidden := range [][]byte{[]byte("ciphertext"), []byte("opening"), []byte("backend_ref")} {
		if bytes.Contains(bytes.ToLower(bz), forbidden) {
			t.Fatalf("consensus reference contains %q", forbidden)
		}
	}
}

func TestEvidenceObjectStoreEnforcesRetentionTransitions(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
	if err := k.SetEvidenceObjectRef(ctx, ref); err != nil {
		t.Fatal(err)
	}
	if err := k.SetEvidenceObjectRef(ctx, ref); err != nil {
		t.Fatalf("exact retry failed: %v", err)
	}
	scheduled, err := contracts.TransitionRetention(ref, contracts.RetentionDeletionScheduled, 11, 101)
	if err != nil {
		t.Fatal(err)
	}
	if err := k.SetEvidenceObjectRef(ctx, scheduled); err != nil {
		t.Fatal(err)
	}
	stale := scheduled
	stale.State = contracts.RetentionDeletionPending
	stale.UpdatedHeight--
	if err := k.SetEvidenceObjectRef(ctx, stale); err == nil {
		t.Fatal("stale update accepted")
	}
	changed := scheduled
	changed.PolicyDigest = testDigest("changed-policy")
	if err := k.SetEvidenceObjectRef(ctx, changed); err == nil {
		t.Fatal("commitment-bound field change accepted")
	}
	hold, err := contracts.TransitionRetention(scheduled, contracts.RetentionLegalHold, 12, 102)
	if err != nil {
		t.Fatal(err)
	}
	if err := k.SetEvidenceObjectRef(ctx, hold); err != nil {
		t.Fatal(err)
	}
	bypass := hold
	bypass.State = contracts.RetentionActive
	bypass.UpdatedHeight++
	bypass.UpdatedUnix++
	if err := k.SetEvidenceObjectRef(ctx, bypass); err == nil {
		t.Fatal("legal hold bypass accepted")
	}
	deleted := hold
	deleted.State = contracts.RetentionDeleted
	deleted.UpdatedHeight++
	deleted.UpdatedUnix++
	ctx.KVStore(k.skey).Set(evidenceObjectRefKey(deleted.ObjectCommitment), mustJSON(t, deleted))
	if err := k.SetEvidenceObjectRef(ctx, bypass); err == nil {
		t.Fatal("deleted state overwrite accepted")
	}
}

func TestEvidenceObjectMigrationDeterministicIdempotentAndTruthful(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	legacyPayload := json.RawMessage(`{"ciphertext":"legacy-secret","nonce":"legacy-nonce"}`)
	legacyRow := mustJSON(t, struct {
		ScopeID          string          `json:"scope_id"`
		EncryptedPayload json.RawMessage `json:"encrypted_payload"`
	}{ScopeID: "scope-a", EncryptedPayload: legacyPayload})
	legacyKey := types.ScopeKey([]byte("account"), "scope-a")
	ctx.KVStore(k.skey).Set(legacyKey, legacyRow)
	legacyHash := sha256.Sum256(legacyPayload)
	ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
	manifest, resolver := signedManifest(t, k, ctx, []EvidenceMigrationEntry{{LegacyEnvelopeHash: hex.EncodeToString(legacyHash[:]), Reference: ref}})

	first, err := k.MigrateEvidenceObjects(ctx, manifest, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if first.Scanned != 1 || first.ReferenceCreated != 1 || first.LegacyRowsSanitized != 0 || first.LegacyRowsPendingCutover != 1 {
		t.Fatalf("untruthful first report: %+v", first)
	}
	if !bytes.Equal(ctx.KVStore(k.skey).Get(legacyKey), legacyRow) {
		t.Fatal("migration silently rewrote legacy row")
	}
	if _, found := k.GetEvidenceObjectRef(ctx, ref.ObjectCommitment); !found {
		t.Fatal("mapped reference missing")
	}

	second, err := k.MigrateEvidenceObjects(ctx, manifest, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if second != first {
		t.Fatalf("migration was not idempotent: %+v", second)
	}
}

func TestEvidenceObjectMigrationQuarantinesMissingMappingWithoutSecrets(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	payload := json.RawMessage(`{"ciphertext":"must-remain-legacy"}`)
	row := mustJSON(t, struct {
		Version          uint32          `json:"version"`
		ScopeID          string          `json:"scope_id"`
		AccountAddress   string          `json:"account_address"`
		Provider         string          `json:"provider"`
		EncryptedPayload json.RawMessage `json:"encrypted_payload"`
	}{Version: 1, ScopeID: "social-a", AccountAddress: "account", Provider: "google", EncryptedPayload: payload})
	ctx.KVStore(k.skey).Set(append(append([]byte(nil), types.PrefixSocialMediaScope...), []byte("social-a")...), row)
	manifest, resolver := signedManifest(t, k, ctx, nil)
	report, err := k.MigrateEvidenceObjects(ctx, manifest, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if report.MissingMapping != 1 || report.Quarantined != 1 || report.MissingMappingQuarantined != 1 || report.LegacyRowsPendingCutover != 1 || report.LegacyRowsSanitized != 0 {
		t.Fatalf("unexpected quarantine report: %+v", report)
	}
	quarantinePrefix := append(append([]byte(nil), types.PrefixEvidenceObjectRef...), []byte("quarantine/")...)
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), quarantinePrefix)
	defer iterator.Close()
	if !iterator.Valid() {
		t.Fatal("quarantine record missing")
	}
	if bytes.Contains(iterator.Value(), []byte("must-remain-legacy")) || bytes.Contains(bytes.ToLower(iterator.Value()), []byte("ciphertext")) {
		t.Fatal("quarantine record contains legacy payload")
	}
}

func TestEvidenceObjectMigrationQuarantinesMalformedAndAmbiguousRows(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	payload := json.RawMessage(`{"ciphertext":"valid"}`)
	validRow := mustJSON(t, struct {
		ScopeID          string          `json:"scope_id"`
		EncryptedPayload json.RawMessage `json:"encrypted_payload"`
	}{ScopeID: "valid", EncryptedPayload: payload})
	ctx.KVStore(k.skey).Set(types.ScopeKey([]byte("a"), "valid"), validRow)
	ctx.KVStore(k.skey).Set(types.ScopeKey([]byte("a"), "broken"), []byte("not-json"))
	ctx.KVStore(k.skey).Set(append(append([]byte(nil), types.PrefixSocialMediaScope...), []byte("ambiguous")...), []byte(`{"scope_id":"ambiguous","encrypted_payload":{}}`))
	manifest, resolver := signedManifest(t, k, ctx, nil)
	report, err := k.MigrateEvidenceObjects(ctx, manifest, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if report.Scanned != 3 || report.Quarantined != 3 || report.Ambiguous != 1 || report.LegacyRowsPendingCutover != 3 {
		t.Fatalf("unexpected malformed-row report: %+v", report)
	}
}

func TestEvidenceObjectMigrationRejectsContextAndReplayChanges(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	manifest, resolver := signedManifest(t, k, ctx, nil)
	for name, mutate := range map[string]func(*EvidenceMigrationManifest){
		"chain":    func(value *EvidenceMigrationManifest) { value.ChainID = "wrong-chain" },
		"height":   func(value *EvidenceMigrationManifest) { value.TargetHeight++ },
		"snapshot": func(value *EvidenceMigrationManifest) { value.SourceSnapshotDigest = testDigest("wrong-snapshot") },
	} {
		t.Run(name, func(t *testing.T) {
			changed := manifest
			mutate(&changed)
			resignManifest(t, &changed, resolver.private)
			if _, err := k.MigrateEvidenceObjects(ctx, changed, resolver); err == nil {
				t.Fatal("mismatched migration context accepted")
			}
		})
	}
	first, err := k.MigrateEvidenceObjects(ctx, manifest, resolver)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := k.MigrateEvidenceObjects(ctx, manifest, resolver)
	if err != nil || replayed != first {
		t.Fatal("exact manifest replay did not return stored report")
	}
	changed := manifest
	changed.PreviousManifestDigest = testDigest("different-predecessor")
	resignManifest(t, &changed, resolver.private)
	if _, err := k.MigrateEvidenceObjects(ctx, changed, resolver); err == nil {
		t.Fatal("changed manifest replay accepted")
	}
	changedSigner := manifest
	changedSigner.SignerKeyEpoch++
	resignManifest(t, &changedSigner, resolver.private)
	if _, err := k.MigrateEvidenceObjects(ctx, changedSigner, resolver); err == nil {
		t.Fatal("manifest replay with changed signer metadata accepted")
	}
	next := newSignedManifest(t, k, ctx, nil, "upgrade-2", manifest.ManifestDigest, 2, resolver)
	if _, err := k.MigrateEvidenceObjects(ctx, next, resolver); err == nil {
		t.Fatal("signer epoch rollback accepted")
	}
}

func TestEvidenceObjectMigrationRejectsMalformedSignerEpochState(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	manifest, resolver := signedManifest(t, k, ctx, nil)
	epochKey := evidenceMigrationUpgradeKey(types.PrefixEvidenceMigrationSignerEpoch, manifest.SignerKeyID)
	ctx.KVStore(k.skey).Set(epochKey, []byte{1, 2, 3})
	if _, err := k.MigrateEvidenceObjects(ctx, manifest, resolver); err == nil {
		t.Fatal("malformed signer epoch floor was accepted")
	}
	store := ctx.KVStore(k.skey)
	if store.Get(evidenceMigrationUpgradeKey(types.PrefixEvidenceMigrationConsumed, manifest.UpgradeID)) != nil ||
		store.Get(evidenceMigrationUpgradeKey(types.PrefixEvidenceMigrationReport, manifest.UpgradeID)) != nil ||
		store.Get(types.KeyEvidenceMigrationLatest) != nil {
		t.Fatal("malformed signer epoch state allowed migration writes")
	}
}

func TestEvidenceObjectMigrationDistinguishesSharedPrefixRows(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	store := ctx.KVStore(k.skey)
	evidenceRecord := []byte(`{"evidence_id":"ev","evidence_type":"document","account_address":"account","scope_id":"scope","content_hash":"hash","envelope_hash":"hash","status":"pending"}`)
	store.Set(append(append([]byte(nil), types.PrefixEvidenceRecord...), []byte("evidence")...), evidenceRecord)
	payload := json.RawMessage(`{"ciphertext":"social"}`)
	social := mustJSON(t, map[string]any{"version": 1, "scope_id": "social", "account_address": "account", "provider": "google", "encrypted_payload": payload})
	store.Set(append(append([]byte(nil), types.PrefixSocialMediaScope...), []byte("social")...), social)
	store.Set(append(append([]byte(nil), types.PrefixSocialMediaScope...), []byte("ambiguous")...), []byte(`{"scope_id":"maybe","encrypted_payload":{}}`))
	payloadHash := sha256.Sum256(payload)
	ref := migrationTestRef(t, contracts.CommitmentDomainSocialScope)
	manifest, resolver := signedManifest(t, k, ctx, []EvidenceMigrationEntry{{LegacyEnvelopeHash: hex.EncodeToString(payloadHash[:]), Reference: ref}})
	report, err := k.MigrateEvidenceObjects(ctx, manifest, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if report.EvidenceRecordRowsSkipped != 1 || report.Ambiguous != 1 || report.ReferenceCreated != 1 || report.Quarantined != 1 {
		t.Fatalf("shared prefix rows misclassified: %+v", report)
	}
}

func TestEvidenceObjectMigrationRollsBackWritesOnConflict(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	ctx.KVStore(k.skey).Set(types.ScopeKey([]byte("a"), "broken"), []byte("not-json"))
	payload := json.RawMessage(`{"ciphertext":"valid"}`)
	ctx.KVStore(k.skey).Set(types.ScopeKey([]byte("z"), "valid"), mustJSON(t, map[string]any{"scope_id": "valid", "encrypted_payload": payload}))
	ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
	if err := k.SetEvidenceObjectRef(ctx, ref); err != nil {
		t.Fatal(err)
	}
	mapped, err := contracts.TransitionRetention(ref, contracts.RetentionDeletionScheduled, 11, 101)
	if err != nil {
		t.Fatal(err)
	}
	payloadHash := sha256.Sum256(payload)
	manifest, resolver := signedManifest(t, k, ctx, []EvidenceMigrationEntry{{LegacyEnvelopeHash: hex.EncodeToString(payloadHash[:]), Reference: mapped}})
	if _, err := k.MigrateEvidenceObjects(ctx, manifest, resolver); err == nil {
		t.Fatal("conflicting migration accepted")
	}
	if ctx.KVStore(k.skey).Get(evidenceMigrationUpgradeKey(types.PrefixEvidenceMigrationConsumed, manifest.UpgradeID)) != nil {
		t.Fatal("failed migration consumed manifest")
	}
	quarantineKey := append(append([]byte(nil), types.PrefixEvidenceObjectRef...), []byte("quarantine/")...)
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), quarantineKey)
	defer iterator.Close()
	if iterator.Valid() {
		t.Fatal("failed migration committed quarantine write")
	}
}

func newEvidenceMigrationKeeper(t *testing.T) (Keeper, sdk.Context, store.CommitMultiStore) {
	t.Helper()
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	database := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(database, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, database)
	if err := stateStore.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}
	ctx := sdk.NewContext(stateStore, cmtproto.Header{ChainID: "chain-test", Time: time.Unix(100, 0).UTC(), Height: 10}, false, log.NewNopLogger())
	return Keeper{cdc: codec.NewProtoCodec(registry), skey: storeKey, authority: "authority"}, ctx, stateStore
}

func migrationTestRef(t *testing.T, domain string) contracts.EvidenceObjectRef {
	t.Helper()
	digest := func(value string) string { sum := sha256.Sum256([]byte(value)); return hex.EncodeToString(sum[:]) }
	ref, _, err := contracts.CreateEvidenceObjectRef(rand.Reader, domain, contracts.EvidenceObjectFields{
		EvidenceDigest: digest("evidence"), RetentionPolicyDigest: digest("retention"),
		PolicyDigest: digest("policy"), ProfileDigest: digest("profile"), KeyEpoch: 1,
		CreatedHeight: 10, CreatedUnix: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func signedManifest(t *testing.T, k Keeper, ctx sdk.Context, entries []EvidenceMigrationEntry) (EvidenceMigrationManifest, migrationResolver) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	resolver := migrationResolver{key: publicKey, private: privateKey}
	return newSignedManifest(t, k, ctx, entries, "upgrade-1", EvidenceMigrationGenesisMarker, 3, resolver), resolver
}

func newSignedManifest(t *testing.T, k Keeper, ctx sdk.Context, entries []EvidenceMigrationEntry, upgradeID, previous string, epoch uint64, resolver migrationResolver) EvidenceMigrationManifest {
	t.Helper()
	rows, _ := collectLegacyEvidenceRows(ctx, k)
	manifest := EvidenceMigrationManifest{
		Version: EvidenceMigrationManifestVersion, ChainID: ctx.ChainID(), UpgradeID: upgradeID,
		TargetHeight: ctx.BlockHeight(), SourceSnapshotDigest: computeLegacyEvidenceSnapshotDigest(rows),
		PreviousManifestDigest: previous, Entries: entries, SignerKeyID: "governance", SignerKeyEpoch: epoch,
	}
	resignManifest(t, &manifest, resolver.private)
	return manifest
}

func resignManifest(t *testing.T, manifest *EvidenceMigrationManifest, privateKey ed25519.PrivateKey) {
	t.Helper()
	var err error
	manifest.ManifestDigest, err = manifest.ComputeDigest()
	if err != nil {
		t.Fatal(err)
	}
	signBytes, err := manifest.CanonicalSignBytes()
	if err != nil {
		t.Fatal(err)
	}
	manifest.Signature = ed25519.Sign(privateKey, signBytes)
}

func testDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	bz, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return bz
}
