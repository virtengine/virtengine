package keeper

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
	"github.com/virtengine/virtengine/x/veid/types"
)

func TestEvidencePayloadCutoverSanitizesMappedScope(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	payload := json.RawMessage(`{"ciphertext":"legacy-secret","nonce":"legacy-nonce"}`)
	row := mustJSON(t, map[string]any{
		"scope_id": "scope-a", "scope_type": "document", "status": "pending", "encrypted_payload": payload,
		"unknown": map[string]any{"nested_ciphertext": "must-drop"}, "evidence_storage_backend": "vault",
		"evidence_storage_ref": "secret-ref", "evidence_metadata": map[string]any{"opening": "secret"},
	})
	key := types.ScopeKey([]byte("account"), "scope-a")
	ctx.KVStore(k.skey).Set(key, row)
	ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
	migration, resolver := signedManifest(t, k, ctx, []EvidenceMigrationEntry{{LegacyEnvelopeHash: digestRaw(payload), Reference: ref}})
	if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
		t.Fatal(err)
	}
	entry := sanitizeCutoverEntry(t, k, ctx, "scope", key, ref.ObjectCommitment)
	manifest := signedCutoverManifest(t, k, ctx, []EvidencePayloadCutoverEntry{entry}, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "cutover-1", 4, resolver)
	report, err := k.CutoverEvidencePayloads(ctx, manifest, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if report.Scanned != 1 || report.Sanitized != 1 || report.Deleted != 0 || report.ScopeSources != 1 || report.AlreadyCutover != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
	stored := ctx.KVStore(k.skey).Get(key)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(stored, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["encrypted_payload"]) != "null" || string(fields["scope_type"]) != `"document"` || string(fields["status"]) != `"pending"` || bytes.Contains(stored, []byte("legacy-secret")) {
		t.Fatalf("row was not safely sanitized: %s", stored)
	}
	for _, removed := range []string{"unknown", "evidence_storage_backend", "evidence_storage_ref", "evidence_metadata"} {
		if _, found := fields[removed]; found {
			t.Fatalf("sanitization retained %s: %s", removed, stored)
		}
	}
	if _, found := k.GetEvidenceObjectRef(ctx, ref.ObjectCommitment); !found {
		t.Fatal("mapped reference was removed")
	}
	rows, _ := collectLegacyEvidenceRows(ctx, k)
	if len(rows) != 0 {
		t.Fatal("sanitized row still classified as legacy")
	}
	replayed, err := k.CutoverEvidencePayloads(ctx, manifest, resolver)
	if err != nil || replayed != report {
		t.Fatalf("exact replay failed: %+v %v", replayed, err)
	}
}

func TestEvidencePayloadCutoverCanonicalEntryOrdering(t *testing.T) {
	first := EvidencePayloadCutoverEntry{SourceKind: "scope", SourceKeyDigest: testDigest("a"), SourceRowDigest: testDigest("row-a"), Action: EvidencePayloadCutoverActionDelete, AuthorityRecordDigest: testDigest("authority-a")}
	second := EvidencePayloadCutoverEntry{SourceKind: "social_scope", SourceKeyDigest: testDigest("b"), SourceRowDigest: testDigest("row-b"), Action: EvidencePayloadCutoverActionDelete, AuthorityRecordDigest: testDigest("authority-b")}
	base := EvidencePayloadCutoverManifest{Version: 1, ChainID: "chain", CutoverID: "cutover", TargetHeight: 10, SourceSnapshotDigest: testDigest("snapshot"), PrerequisiteMigrationManifestDigest: testDigest("migration"), PreviousCutoverManifestDigest: EvidencePayloadCutoverGenesisMarker, Entries: []EvidencePayloadCutoverEntry{first, second}}
	digest, err := base.ComputeDigest()
	if err != nil {
		t.Fatal(err)
	}
	base.Entries = []EvidencePayloadCutoverEntry{second, first}
	reordered, err := base.ComputeDigest()
	if err != nil || reordered != digest {
		t.Fatalf("entry order changed digest: %s %s %v", digest, reordered, err)
	}
}

func TestEvidencePayloadCutoverSharedPrefixAndMixedActions(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	store := ctx.KVStore(k.skey)
	evidenceKey := append(append([]byte(nil), types.PrefixEvidenceRecord...), []byte("evidence")...)
	evidenceRecord := []byte(`{"evidence_id":"ev","evidence_type":"document","account_address":"account","scope_id":"scope","content_hash":"hash","envelope_hash":"hash","status":"pending"}`)
	store.Set(evidenceKey, evidenceRecord)
	socialPayload := json.RawMessage(`{"ciphertext":"social-secret"}`)
	socialKey := append(append([]byte(nil), types.PrefixSocialMediaScope...), []byte("social")...)
	store.Set(socialKey, mustJSON(t, map[string]any{"version": 1, "scope_id": "social", "account_address": "account", "provider": "google", "encrypted_payload": socialPayload, "status": "verified", "keep": "drop"}))
	brokenKey := types.ScopeKey([]byte("z"), "broken")
	brokenRow := []byte(`{"record_id":"broken","encrypted_payload":{}}`)
	store.Set(brokenKey, brokenRow)
	ref := migrationTestRef(t, contracts.CommitmentDomainSocialScope)
	migration, resolver := signedManifest(t, k, ctx, []EvidenceMigrationEntry{{LegacyEnvelopeHash: digestRaw(socialPayload), Reference: ref}})
	if migrationReport, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil || migrationReport.EvidenceRecordRowsSkipped != 1 {
		t.Fatalf("migration setup failed: %+v %v", migrationReport, err)
	}
	rows, _ := collectLegacyEvidenceRows(ctx, k)
	entries := make([]EvidencePayloadCutoverEntry, 0, len(rows))
	for _, row := range rows {
		if row.sourceKind == "social_scope" && row.classification == "legacy" {
			entries = append(entries, sanitizeCutoverEntry(t, k, ctx, row.sourceKind, row.key, ref.ObjectCommitment))
		} else {
			entries = append(entries, quarantineCutoverEntry(t, k, ctx, row))
		}
	}
	manifest := signedCutoverManifest(t, k, ctx, entries, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "mixed", 5, resolver)
	report, err := k.CutoverEvidencePayloads(ctx, manifest, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if report.Scanned != 2 || report.Sanitized != 1 || report.Deleted != 1 || report.EvidenceRecordsSkipped != 1 || report.ScopeSources != 1 || report.SocialScopeSources != 1 {
		t.Fatalf("unexpected mixed report: %+v", report)
	}
	if !bytes.Equal(store.Get(evidenceKey), evidenceRecord) || store.Get(brokenKey) != nil {
		t.Fatal("cutover changed shared evidence record or retained deleted source")
	}
	var social map[string]json.RawMessage
	if json.Unmarshal(store.Get(socialKey), &social) != nil || string(social["encrypted_payload"]) != "null" || string(social["status"]) != `"verified"` {
		t.Fatalf("social row was not safely sanitized: %s", store.Get(socialKey))
	}
	if _, found := social["keep"]; found {
		t.Fatalf("social sanitizer retained unknown field: %s", store.Get(socialKey))
	}
	quarantineKey := append(append([]byte(nil), types.PrefixEvidenceObjectRef...), []byte("quarantine/"+evidencePayloadCutoverSourceKeyDigest("scope", brokenKey))...)
	if store.Get(quarantineKey) == nil {
		t.Fatal("delete removed quarantine authority")
	}
	changed := manifest
	changed.SignerKeyEpoch++
	resignCutoverManifest(t, &changed, resolver.private)
	if _, err := k.CutoverEvidencePayloads(ctx, changed, resolver); err == nil {
		t.Fatal("changed replay accepted")
	}
}

func TestEvidencePayloadCutoverDeletesQuarantinedRows(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	store := ctx.KVStore(k.skey)
	malformedKey := types.ScopeKey([]byte("b"), "ambiguous")
	ambiguousKey := append(append([]byte(nil), types.PrefixSocialMediaScope...), []byte("ambiguous")...)
	store.Set(malformedKey, []byte(`{"record_id":"ambiguous-scope","encrypted_payload":{}}`))
	store.Set(ambiguousKey, []byte(`{"scope_id":"ambiguous-social","encrypted_payload":{}}`))
	migration, resolver := signedManifest(t, k, ctx, nil)
	if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
		t.Fatal(err)
	}
	rows, _ := collectLegacyEvidenceRows(ctx, k)
	entries := make([]EvidencePayloadCutoverEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, quarantineCutoverEntry(t, k, ctx, row))
	}
	manifest := signedCutoverManifest(t, k, ctx, entries, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "delete-quarantine", 4, resolver)
	report, err := k.CutoverEvidencePayloads(ctx, manifest, resolver)
	if err != nil || report.Deleted != 2 || report.Sanitized != 0 {
		t.Fatalf("quarantine deletion failed: %+v %v", report, err)
	}
	for _, key := range [][]byte{malformedKey, ambiguousKey} {
		if store.Get(key) != nil {
			t.Fatalf("payload source retained at %x", key)
		}
	}
	for _, entry := range entries {
		key := append(append([]byte(nil), types.PrefixEvidenceObjectRef...), []byte("quarantine/"+entry.SourceKeyDigest)...)
		if store.Get(key) == nil {
			t.Fatal("quarantine authority was deleted")
		}
	}
}

func TestEvidencePayloadCutoverRejectsMalformedScopeDelete(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	store := ctx.KVStore(k.skey)
	key := types.ScopeKey([]byte("account"), "malformed")
	source := []byte("malformed-payload")
	store.Set(key, source)
	migration, resolver := signedManifest(t, k, ctx, nil)
	if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
		t.Fatal(err)
	}
	rows, _ := collectEvidencePayloadCutoverRows(ctx, k)
	entry := quarantineCutoverEntry(t, k, ctx, rows[0])
	manifest := signedCutoverManifest(t, k, ctx, []EvidencePayloadCutoverEntry{entry}, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "malformed-scope", 4, resolver)
	if _, err := k.CutoverEvidencePayloads(ctx, manifest, resolver); err == nil {
		t.Fatal("malformed scope without a provable payload marker was deleted")
	}
	if !bytes.Equal(store.Get(key), source) || cutoverBookkeepingExists(ctx, k, manifest.CutoverID) {
		t.Fatal("rejected malformed scope delete was not atomic")
	}
}

func TestEvidencePayloadCutoverRejectsManifestAttacks(t *testing.T) {
	type fixture struct {
		k        Keeper
		ctx      sdk.Context
		manifest EvidencePayloadCutoverManifest
		resolver migrationResolver
	}
	newFixture := func(t *testing.T) fixture {
		k, ctx, _ := newEvidenceMigrationKeeper(t)
		payload := json.RawMessage(`{"ciphertext":"attack-target"}`)
		key := types.ScopeKey([]byte("a"), "target")
		ctx.KVStore(k.skey).Set(key, mustJSON(t, map[string]any{"scope_id": "target", "encrypted_payload": payload}))
		ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
		migration, resolver := signedManifest(t, k, ctx, []EvidenceMigrationEntry{{LegacyEnvelopeHash: digestRaw(payload), Reference: ref}})
		if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
			t.Fatal(err)
		}
		entry := sanitizeCutoverEntry(t, k, ctx, "scope", key, ref.ObjectCommitment)
		manifest := signedCutoverManifest(t, k, ctx, []EvidencePayloadCutoverEntry{entry}, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "attacks", 6, resolver)
		return fixture{k: k, ctx: ctx, manifest: manifest, resolver: resolver}
	}
	tests := map[string]func(*EvidencePayloadCutoverManifest){
		"chain":     func(m *EvidencePayloadCutoverManifest) { m.ChainID = "wrong-chain" },
		"height":    func(m *EvidencePayloadCutoverManifest) { m.TargetHeight++ },
		"signer-id": func(m *EvidencePayloadCutoverManifest) { m.SignerKeyID = "attacker" },
		"prerequisite": func(m *EvidencePayloadCutoverManifest) {
			m.PrerequisiteMigrationManifestDigest = testDigest("wrong-migration")
		},
		"predecessor": func(m *EvidencePayloadCutoverManifest) {
			m.PreviousCutoverManifestDigest = testDigest("wrong-predecessor")
		},
		"snapshot":   func(m *EvidencePayloadCutoverManifest) { m.SourceSnapshotDigest = testDigest("wrong-snapshot") },
		"key-digest": func(m *EvidencePayloadCutoverManifest) { m.Entries[0].SourceKeyDigest = testDigest("wrong-key") },
		"row-digest": func(m *EvidencePayloadCutoverManifest) { m.Entries[0].SourceRowDigest = testDigest("wrong-row") },
		"action": func(m *EvidencePayloadCutoverManifest) {
			m.Entries[0].Action = EvidencePayloadCutoverActionDelete
			m.Entries[0].ObjectCommitment = ""
		},
		"authority": func(m *EvidencePayloadCutoverManifest) {
			m.Entries[0].AuthorityRecordDigest = testDigest("wrong-authority")
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			mutate(&f.manifest)
			resignCutoverManifest(t, &f.manifest, f.resolver.private)
			if _, err := f.k.CutoverEvidencePayloads(f.ctx, f.manifest, f.resolver); err == nil {
				t.Fatal("mutated manifest accepted")
			}
		})
	}
	t.Run("signature", func(t *testing.T) {
		f := newFixture(t)
		f.manifest.Signature[0] ^= 0xff
		if _, err := f.k.CutoverEvidencePayloads(f.ctx, f.manifest, f.resolver); err == nil {
			t.Fatal("invalid signature accepted")
		}
	})
	t.Run("epoch-rollback", func(t *testing.T) {
		f := newFixture(t)
		first, err := f.k.CutoverEvidencePayloads(f.ctx, f.manifest, f.resolver)
		if err != nil || first.Sanitized != 1 {
			t.Fatal(err)
		}
		next := signedCutoverManifest(t, f.k, f.ctx, nil, f.manifest.PrerequisiteMigrationManifestDigest, f.manifest.ManifestDigest, "next", 5, f.resolver)
		if _, err := f.k.CutoverEvidencePayloads(f.ctx, next, f.resolver); err == nil {
			t.Fatal("signer epoch rollback accepted")
		}
	})
}

func TestEvidencePayloadCutoverRejectsEntrySetErrors(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	key := types.ScopeKey([]byte("a"), "broken")
	ctx.KVStore(k.skey).Set(key, []byte("broken"))
	migration, resolver := signedManifest(t, k, ctx, nil)
	if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
		t.Fatal(err)
	}
	rows, _ := collectLegacyEvidenceRows(ctx, k)
	entry := quarantineCutoverEntry(t, k, ctx, rows[0])
	for name, entries := range map[string][]EvidencePayloadCutoverEntry{
		"missing":   nil,
		"extra":     {entry, {SourceKind: "scope", SourceKeyDigest: testDigest("extra-key"), SourceRowDigest: testDigest("extra-row"), Action: EvidencePayloadCutoverActionDelete, AuthorityRecordDigest: testDigest("extra-authority")}},
		"duplicate": {entry, entry},
	} {
		t.Run(name, func(t *testing.T) {
			manifest := signedCutoverManifest(t, k, ctx, entries, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "entry-"+name, 4, resolver)
			if _, err := k.CutoverEvidencePayloads(ctx, manifest, resolver); err == nil {
				t.Fatal("invalid entry set accepted")
			}
		})
	}
}

func TestEvidencePayloadCutoverRejectsWrongAuthorityClass(t *testing.T) {
	t.Run("mapped-delete", func(t *testing.T) {
		k, ctx, _ := newEvidenceMigrationKeeper(t)
		payload := json.RawMessage(`{"ciphertext":"mapped"}`)
		key := types.ScopeKey([]byte("a"), "mapped")
		ctx.KVStore(k.skey).Set(key, mustJSON(t, map[string]any{"scope_id": "mapped", "encrypted_payload": payload}))
		ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
		migration, resolver := signedManifest(t, k, ctx, []EvidenceMigrationEntry{{LegacyEnvelopeHash: digestRaw(payload), Reference: ref}})
		if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
			t.Fatal(err)
		}
		row, _ := classifyLegacyEvidenceRow("scope", key, ctx.KVStore(k.skey).Get(key))
		entry := EvidencePayloadCutoverEntry{SourceKind: "scope", SourceKeyDigest: evidencePayloadCutoverSourceKeyDigest("scope", key), SourceRowDigest: row.rowDigest, Action: EvidencePayloadCutoverActionDelete, AuthorityRecordDigest: testDigest("no-quarantine")}
		manifest := signedCutoverManifest(t, k, ctx, []EvidencePayloadCutoverEntry{entry}, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "mapped-delete", 4, resolver)
		if _, err := k.CutoverEvidencePayloads(ctx, manifest, resolver); err == nil {
			t.Fatal("mapped delete accepted")
		}
	})
	t.Run("quarantined-sanitize", func(t *testing.T) {
		k, ctx, _ := newEvidenceMigrationKeeper(t)
		key := types.ScopeKey([]byte("a"), "broken")
		ctx.KVStore(k.skey).Set(key, []byte("broken"))
		migration, resolver := signedManifest(t, k, ctx, nil)
		if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
			t.Fatal(err)
		}
		ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
		if err := k.SetEvidenceObjectRef(ctx, ref); err != nil {
			t.Fatal(err)
		}
		entry := sanitizeCutoverEntry(t, k, ctx, "scope", key, ref.ObjectCommitment)
		manifest := signedCutoverManifest(t, k, ctx, []EvidencePayloadCutoverEntry{entry}, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "quarantine-sanitize", 4, resolver)
		if _, err := k.CutoverEvidencePayloads(ctx, manifest, resolver); err == nil {
			t.Fatal("quarantined sanitize accepted")
		}
	})
}

func TestEvidencePayloadCutoverRollsBackLateConflict(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	store := ctx.KVStore(k.skey)
	payload := json.RawMessage(`{"ciphertext":"must-rollback"}`)
	firstKey := types.ScopeKey([]byte("a"), "mapped")
	secondKey := types.ScopeKey([]byte("z"), "broken")
	firstRow := mustJSON(t, map[string]any{"scope_id": "mapped", "encrypted_payload": payload})
	store.Set(firstKey, firstRow)
	store.Set(secondKey, []byte("broken"))
	ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
	migration, resolver := signedManifest(t, k, ctx, []EvidenceMigrationEntry{{LegacyEnvelopeHash: digestRaw(payload), Reference: ref}})
	if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
		t.Fatal(err)
	}
	rows, _ := collectLegacyEvidenceRows(ctx, k)
	entries := []EvidencePayloadCutoverEntry{sanitizeCutoverEntry(t, k, ctx, "scope", firstKey, ref.ObjectCommitment), quarantineCutoverEntry(t, k, ctx, rows[1])}
	entries[1].AuthorityRecordDigest = testDigest("late-conflict")
	manifest := signedCutoverManifest(t, k, ctx, entries, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "rollback", 4, resolver)
	if _, err := k.CutoverEvidencePayloads(ctx, manifest, resolver); err == nil {
		t.Fatal("late conflict accepted")
	}
	if !bytes.Equal(store.Get(firstKey), firstRow) || store.Get(secondKey) == nil || cutoverBookkeepingExists(ctx, k, manifest.CutoverID) || store.Get(keyEvidencePayloadCutoverLatest) != nil {
		t.Fatal("failed cutover leaked row mutation or bookkeeping")
	}
}

func TestEvidencePayloadCutoverMapsStaleLegacyQuarantineBeforeSanitize(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	store := ctx.KVStore(k.skey)
	payload := json.RawMessage(`{"ciphertext":"mapped-later"}`)
	key := types.ScopeKey([]byte("account"), "mapped-later")
	store.Set(key, mustJSON(t, map[string]any{"scope_id": "mapped-later", "scope_type": "document", "encrypted_payload": payload}))

	firstMigration, resolver := signedManifest(t, k, ctx, nil)
	if _, err := k.MigrateEvidenceObjects(ctx, firstMigration, resolver); err != nil {
		t.Fatal(err)
	}
	keyDigest := evidencePayloadCutoverSourceKeyDigest("scope", key)
	quarantineKey := append(append([]byte(nil), types.PrefixEvidenceObjectRef...), []byte("quarantine/"+keyDigest)...)
	quarantine := append([]byte(nil), store.Get(quarantineKey)...)
	if len(quarantine) == 0 {
		t.Fatal("first migration did not quarantine unmapped legacy row")
	}

	ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
	secondMigration := newSignedManifest(t, k, ctx, []EvidenceMigrationEntry{{LegacyEnvelopeHash: digestRaw(payload), Reference: ref}}, "upgrade-2", firstMigration.ManifestDigest, 4, resolver)
	if _, err := k.MigrateEvidenceObjects(ctx, secondMigration, resolver); err != nil {
		t.Fatal(err)
	}
	row, _ := classifyLegacyEvidenceRow("scope", key, store.Get(key))
	deleteEntry := quarantineCutoverEntry(t, k, ctx, row)
	deleteManifest := signedCutoverManifest(t, k, ctx, []EvidencePayloadCutoverEntry{deleteEntry}, secondMigration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "stale-delete", 5, resolver)
	if _, err := k.CutoverEvidencePayloads(ctx, deleteManifest, resolver); err == nil {
		t.Fatal("legacy row with stale quarantine was deleted")
	}
	if store.Get(key) == nil || !bytes.Equal(store.Get(quarantineKey), quarantine) {
		t.Fatal("rejected delete changed source or quarantine")
	}

	sanitizeEntry := sanitizeCutoverEntry(t, k, ctx, "scope", key, ref.ObjectCommitment)
	sanitizeManifest := signedCutoverManifest(t, k, ctx, []EvidencePayloadCutoverEntry{sanitizeEntry}, secondMigration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "stale-sanitize", 5, resolver)
	report, err := k.CutoverEvidencePayloads(ctx, sanitizeManifest, resolver)
	if err != nil || report.Sanitized != 1 {
		t.Fatalf("mapped row with exact stale quarantine was not sanitized: %+v %v", report, err)
	}
	var sanitized map[string]json.RawMessage
	if json.Unmarshal(store.Get(key), &sanitized) != nil || string(sanitized["encrypted_payload"]) != "null" || !bytes.Equal(store.Get(quarantineKey), quarantine) {
		t.Fatalf("sanitize did not retain source metadata and historical quarantine: %s", store.Get(key))
	}
}

func TestEvidencePayloadCutoverRejectsSharedPrefixEvidenceShapes(t *testing.T) {
	shapes := map[string][]byte{
		"canonical": []byte(`{"evidence_id":"ev","evidence_type":"document","account_address":"account","scope_id":"scope","content_hash":"hash","envelope_hash":"hash","status":"pending"}`),
		"partial":   []byte(`{"evidence_id":"ev","scope_id":"scope"}`),
		"old":       []byte(`{"evidence_type":"document","scope_id":"scope","encrypted_payload":{}}`),
		"extended":  []byte(`{"version":1,"scope_id":"scope","account_address":"account","provider":"google","encrypted_payload":{},"evidence_id":"ev","evidence_type":"document","content_hash":"hash","envelope_hash":"hash","status":"pending","confidence":9000}`),
		"decision":  []byte(`{"decision_reason":"legacy review","encrypted_payload":{}}`),
		"verified":  []byte(`{"verified_at":"2026-08-02T00:00:00Z","encrypted_payload":{}}`),
	}
	for name, source := range shapes {
		t.Run(name, func(t *testing.T) {
			k, ctx, _ := newEvidenceMigrationKeeper(t)
			store := ctx.KVStore(k.skey)
			key := append(append([]byte(nil), types.PrefixSocialMediaScope...), []byte(name)...)
			store.Set(key, source)
			migration, resolver := signedManifest(t, k, ctx, nil)
			if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
				t.Fatal(err)
			}
			rows, _ := collectEvidencePayloadCutoverRows(ctx, k)
			var entry EvidencePayloadCutoverEntry
			if len(rows) == 0 {
				entry = EvidencePayloadCutoverEntry{SourceKind: "social_scope", SourceKeyDigest: evidencePayloadCutoverSourceKeyDigest("social_scope", key), SourceRowDigest: evidencePayloadCutoverDigest(source), Action: EvidencePayloadCutoverActionDelete, AuthorityRecordDigest: testDigest("canonical-authority")}
			} else {
				entry = quarantineCutoverEntry(t, k, ctx, rows[0])
			}
			manifest := signedCutoverManifest(t, k, ctx, []EvidencePayloadCutoverEntry{entry}, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "evidence-"+name, 4, resolver)
			if _, err := k.CutoverEvidencePayloads(ctx, manifest, resolver); err == nil {
				t.Fatal("evidence-like shared-prefix row was deletable")
			}
			if !bytes.Equal(store.Get(key), source) || cutoverBookkeepingExists(ctx, k, manifest.CutoverID) {
				t.Fatal("rejected evidence-like delete was not atomic")
			}
		})
	}
}

func TestEvidencePayloadCutoverRejectsNestedDataInAllowedFields(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	payload := json.RawMessage(`{"ciphertext":"nested"}`)
	key := types.ScopeKey([]byte("account"), "nested")
	source := []byte(`{"scope_id":{"ciphertext":"hidden"},"scope_type":"document","encrypted_payload":{"ciphertext":"nested"}}`)
	ctx.KVStore(k.skey).Set(key, source)
	ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
	migration, resolver := signedManifest(t, k, ctx, []EvidenceMigrationEntry{{LegacyEnvelopeHash: digestRaw(payload), Reference: ref}})
	// Bind the migration to the actual raw payload representation.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(source, &fields); err != nil {
		t.Fatal(err)
	}
	migration.Entries[0].LegacyEnvelopeHash = digestRaw(fields["encrypted_payload"])
	resignManifest(t, &migration, resolver.private)
	if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
		t.Fatal(err)
	}
	entry := sanitizeCutoverEntry(t, k, ctx, "scope", key, ref.ObjectCommitment)
	manifest := signedCutoverManifest(t, k, ctx, []EvidencePayloadCutoverEntry{entry}, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "nested-field", 4, resolver)
	if _, err := k.CutoverEvidencePayloads(ctx, manifest, resolver); err == nil {
		t.Fatal("nested payload under allowlisted scalar field was accepted")
	}
	if !bytes.Equal(ctx.KVStore(k.skey).Get(key), source) {
		t.Fatal("rejected nested payload changed source row")
	}
}

func TestEvidencePayloadCutoverSnapshotBindsSkippedRecordIdentity(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	store := ctx.KVStore(k.skey)
	firstKey := append(append([]byte(nil), types.PrefixEvidenceRecord...), []byte("first")...)
	secondKey := append(append([]byte(nil), types.PrefixEvidenceRecord...), []byte("second")...)
	firstValue := []byte(`{"evidence_id":"ev","evidence_type":"document","account_address":"account","scope_id":"scope","content_hash":"hash","envelope_hash":"hash","status":"pending"}`)
	secondValue := []byte(`{"evidence_id":"ev","evidence_type":"document","account_address":"account","scope_id":"scope","content_hash":"hash","envelope_hash":"hash","status":"verified"}`)
	store.Set(firstKey, firstValue)
	rows, skipped := collectEvidencePayloadCutoverRows(ctx, k)
	firstDigest := computeEvidencePayloadCutoverSnapshotDigest(rows, skipped)
	store.Set(firstKey, secondValue)
	rows, skipped = collectEvidencePayloadCutoverRows(ctx, k)
	changedBytes := computeEvidencePayloadCutoverSnapshotDigest(rows, skipped)
	if changedBytes == firstDigest {
		t.Fatal("skipped evidence record byte change did not change snapshot")
	}
	store.Delete(firstKey)
	store.Set(secondKey, secondValue)
	rows, skipped = collectEvidencePayloadCutoverRows(ctx, k)
	if changedKey := computeEvidencePayloadCutoverSnapshotDigest(rows, skipped); changedKey == changedBytes {
		t.Fatal("skipped evidence record key change did not change snapshot")
	}
}

func TestEvidencePayloadCutoverRejectsMalformedSignerEpochState(t *testing.T) {
	k, ctx, _ := newEvidenceMigrationKeeper(t)
	payload := json.RawMessage(`{"ciphertext":"epoch"}`)
	key := types.ScopeKey([]byte("account"), "epoch")
	ctx.KVStore(k.skey).Set(key, mustJSON(t, map[string]any{"scope_id": "epoch", "encrypted_payload": payload}))
	ref := migrationTestRef(t, contracts.CommitmentDomainDocument)
	migration, resolver := signedManifest(t, k, ctx, []EvidenceMigrationEntry{{LegacyEnvelopeHash: digestRaw(payload), Reference: ref}})
	if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
		t.Fatal(err)
	}
	entry := sanitizeCutoverEntry(t, k, ctx, "scope", key, ref.ObjectCommitment)
	manifest := signedCutoverManifest(t, k, ctx, []EvidencePayloadCutoverEntry{entry}, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "bad-epoch", 4, resolver)
	epochKey := evidencePayloadCutoverBookkeepingKey(prefixEvidencePayloadCutoverSignerEpoch, manifest.SignerKeyID)
	ctx.KVStore(k.skey).Set(epochKey, []byte{1, 2, 3})
	if _, err := k.CutoverEvidencePayloads(ctx, manifest, resolver); err == nil {
		t.Fatal("malformed signer epoch state was accepted")
	}
}

func TestEvidencePayloadCutoverRejectsCorruptReplayReport(t *testing.T) {
	corruptions := map[string]func(store storetypes.KVStore, key []byte){
		"missing": func(store storetypes.KVStore, key []byte) { store.Delete(key) },
		"malformed": func(store storetypes.KVStore, key []byte) {
			store.Set(key, []byte(`{`))
		},
		"zero": func(store storetypes.KVStore, key []byte) {
			store.Set(key, []byte(`{}`))
		},
		"implausible": func(store storetypes.KVStore, key []byte) {
			store.Set(key, []byte(`{"scanned":2,"sanitized":1,"scope_sources":1}`))
		},
		"payload-field": func(store storetypes.KVStore, key []byte) {
			store.Set(key, []byte(`{"scanned":1,"deleted":1,"scope_sources":1,"ciphertext":"secret"}`))
		},
	}
	for name, corrupt := range corruptions {
		t.Run(name, func(t *testing.T) {
			k, ctx, _ := newEvidenceMigrationKeeper(t)
			store := ctx.KVStore(k.skey)
			key := types.ScopeKey([]byte("account"), "replay")
			store.Set(key, []byte(`{"record_id":"replay","encrypted_payload":{}}`))
			migration, resolver := signedManifest(t, k, ctx, nil)
			if _, err := k.MigrateEvidenceObjects(ctx, migration, resolver); err != nil {
				t.Fatal(err)
			}
			rows, _ := collectEvidencePayloadCutoverRows(ctx, k)
			entry := quarantineCutoverEntry(t, k, ctx, rows[0])
			manifest := signedCutoverManifest(t, k, ctx, []EvidencePayloadCutoverEntry{entry}, migration.ManifestDigest, EvidencePayloadCutoverGenesisMarker, "replay-corrupt", 4, resolver)
			if _, err := k.CutoverEvidencePayloads(ctx, manifest, resolver); err != nil {
				t.Fatal(err)
			}
			reportKey := evidencePayloadCutoverBookkeepingKey(prefixEvidencePayloadCutoverReport, manifest.CutoverID)
			corrupt(store, reportKey)
			if _, err := k.CutoverEvidencePayloads(ctx, manifest, resolver); err == nil {
				t.Fatal("corrupt stored replay report was accepted")
			}
		})
	}
}

func sanitizeCutoverEntry(t *testing.T, k Keeper, ctx sdk.Context, sourceKind string, key []byte, commitment string) EvidencePayloadCutoverEntry {
	t.Helper()
	row := ctx.KVStore(k.skey).Get(key)
	rowDigest := sha256.Sum256(row)
	referenceBytes := ctx.KVStore(k.skey).Get(evidenceObjectRefKey(commitment))
	return EvidencePayloadCutoverEntry{SourceKind: sourceKind, SourceKeyDigest: evidencePayloadCutoverSourceKeyDigest(sourceKind, key), SourceRowDigest: hex.EncodeToString(rowDigest[:]), Action: EvidencePayloadCutoverActionSanitize, AuthorityRecordDigest: evidencePayloadCutoverDigest(referenceBytes), ObjectCommitment: commitment}
}

func signedCutoverManifest(t *testing.T, k Keeper, ctx sdk.Context, entries []EvidencePayloadCutoverEntry, prerequisite, previous, cutoverID string, epoch uint64, resolver migrationResolver) EvidencePayloadCutoverManifest {
	t.Helper()
	rows, skipped := collectEvidencePayloadCutoverRows(ctx, k)
	manifest := EvidencePayloadCutoverManifest{Version: EvidencePayloadCutoverManifestVersion, ChainID: ctx.ChainID(), CutoverID: cutoverID, TargetHeight: ctx.BlockHeight(), SourceSnapshotDigest: computeEvidencePayloadCutoverSnapshotDigest(rows, skipped), PrerequisiteMigrationManifestDigest: prerequisite, PreviousCutoverManifestDigest: previous, SignerKeyID: "governance", SignerKeyEpoch: epoch, Entries: entries}
	resignCutoverManifest(t, &manifest, resolver.private)
	return manifest
}

func resignCutoverManifest(t *testing.T, manifest *EvidencePayloadCutoverManifest, privateKey ed25519.PrivateKey) {
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

func digestRaw(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func quarantineCutoverEntry(t *testing.T, k Keeper, ctx sdk.Context, row legacyEvidenceRow) EvidencePayloadCutoverEntry {
	t.Helper()
	keyDigest := evidencePayloadCutoverSourceKeyDigest(row.sourceKind, row.key)
	authority := ctx.KVStore(k.skey).Get(append(append([]byte(nil), types.PrefixEvidenceObjectRef...), []byte("quarantine/"+keyDigest)...))
	return EvidencePayloadCutoverEntry{SourceKind: row.sourceKind, SourceKeyDigest: keyDigest, SourceRowDigest: row.rowDigest, Action: EvidencePayloadCutoverActionDelete, AuthorityRecordDigest: evidencePayloadCutoverDigest(authority)}
}

func cutoverBookkeepingExists(ctx sdk.Context, k Keeper, cutoverID string) bool {
	return ctx.KVStore(k.skey).Get(evidencePayloadCutoverBookkeepingKey(prefixEvidencePayloadCutoverConsumed, cutoverID)) != nil
}
