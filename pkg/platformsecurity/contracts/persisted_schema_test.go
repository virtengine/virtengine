package contracts

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

const persistedSchemaFixture = "testdata/t5-persisted-schema-inventory-v1.json"

func loadPersistedSchemaFixture(t *testing.T) PersistedSchemaInventory {
	t.Helper()
	data, err := os.ReadFile(persistedSchemaFixture)
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := ParsePersistedSchemaInventory(data)
	if err != nil {
		t.Fatal(err)
	}
	return inventory
}

func clonePersistedSchemaFixture(t *testing.T) PersistedSchemaInventory {
	t.Helper()
	inventory := loadPersistedSchemaFixture(t)
	inventory.Entries = append([]PersistedSchemaEntry(nil), inventory.Entries...)
	for index := range inventory.Entries {
		inventory.Entries[index].EvidencePaths = append([]string(nil), inventory.Entries[index].EvidencePaths...)
	}
	inventory.NonPersistentSurfaces = append([]string(nil), inventory.NonPersistentSurfaces...)
	inventory.Blockers = append([]string(nil), inventory.Blockers...)
	return inventory
}

func entryByID(t *testing.T, inventory *PersistedSchemaInventory, id string) *PersistedSchemaEntry {
	t.Helper()
	for index := range inventory.Entries {
		if inventory.Entries[index].ID == id {
			return &inventory.Entries[index]
		}
	}
	t.Fatalf("entry %q not found", id)
	return nil
}

func TestPersistedSchemaFixtureAndCanonicalDigest(t *testing.T) {
	inventory := loadPersistedSchemaFixture(t)
	if inventory.PayloadHead != PersistedSchemaPayloadHead {
		t.Fatalf("payload_head = %q, want inventoried code boundary %q", inventory.PayloadHead, PersistedSchemaPayloadHead)
	}
	digest, err := inventory.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	const expected = "c2ea1ee7426f671415d6ab4a246dce2f8990fed673b8fa408d3d0b5b0b452d14"
	if actual := hex.EncodeToString(digest[:]); actual != expected {
		t.Fatalf("canonical digest = %s, want %s", actual, expected)
	}
}

func TestPersistedSchemaParserRejectsUnknownAndTrailingJSON(t *testing.T) {
	data, err := os.ReadFile(persistedSchemaFixture)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["unknown"] = true
	unknown, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	for name, candidate := range map[string][]byte{"unknown": unknown, "trailing": append(data, []byte("\n{}")...)} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParsePersistedSchemaInventory(candidate); err == nil {
				t.Fatal("malformed JSON accepted")
			}
		})
	}
}

func TestPersistedSchemaEntryMutationsFailClosed(t *testing.T) {
	tests := []struct {
		name  string
		apply func(*PersistedSchemaInventory)
	}{
		{"omission", func(value *PersistedSchemaInventory) { value.Entries = value.Entries[1:] }},
		{"duplicate", func(value *PersistedSchemaInventory) { value.Entries = append(value.Entries, value.Entries[0]) }},
		{"unknown", func(value *PersistedSchemaInventory) { value.Entries[0].ID = "unknown-entry" }},
		{"wrong class", func(value *PersistedSchemaInventory) { value.Entries[0].StorageClass = StorageConsensus }},
		{"wrong disposition", func(value *PersistedSchemaInventory) { value.Entries[0].MigrationDisposition = MigrationNone }},
		{"missing blocker", func(value *PersistedSchemaInventory) { entryByID(t, value, "fundauth-replay-store").Blocker = "" }},
		{"missing migration ID", func(value *PersistedSchemaInventory) { entryByID(t, value, "veid-evidence-reference").MigrationID = "" }},
		{"bad owner path", func(value *PersistedSchemaInventory) { value.Entries[0].OwnerPath = "../outside.go" }},
		{"bad evidence path", func(value *PersistedSchemaInventory) { value.Entries[0].EvidencePaths[0] = "C:\\outside.go" }},
		{"bad payload SHA", func(value *PersistedSchemaInventory) { value.PayloadHead = strings.ToUpper(value.PayloadHead) }},
		{"duplicate evidence", func(value *PersistedSchemaInventory) {
			value.Entries[0].EvidencePaths = append(value.Entries[0].EvidencePaths, value.Entries[0].EvidencePaths[0])
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inventory := clonePersistedSchemaFixture(t)
			test.apply(&inventory)
			if err := inventory.Validate(); err == nil {
				t.Fatal("invalid inventory accepted")
			}
		})
	}
}

func TestPersistedSchemaGlobalBlockerSetFailsClosed(t *testing.T) {
	tests := []struct {
		name  string
		apply func(*PersistedSchemaInventory)
	}{
		{"omission", func(value *PersistedSchemaInventory) { value.Blockers = value.Blockers[1:] }},
		{"mutation", func(value *PersistedSchemaInventory) { value.Blockers[0] += " changed" }},
		{"duplicate", func(value *PersistedSchemaInventory) { value.Blockers = append(value.Blockers, value.Blockers[0]) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inventory := clonePersistedSchemaFixture(t)
			test.apply(&inventory)
			if err := inventory.Validate(); err == nil {
				t.Fatal("invalid global blocker set accepted")
			}
		})
	}
}

func TestPersistedSchemaPayloadCutoverBookkeepingFailsClosed(t *testing.T) {
	const expectedFormat = "D4 05 || UTF-8 cutoverID -> signed replay digest; D4 06 || cutoverID -> JSON EvidencePayloadCutoverReport; singleton D4 07 -> latest cutover manifest digest; D4 08 || UTF-8 signerKeyID -> BE64 epoch floor"
	inventory := clonePersistedSchemaFixture(t)
	if actual := entryByID(t, &inventory, "veid-payload-cutover-bookkeeping").KeyOrFormat; actual != expectedFormat {
		t.Fatalf("cutover bookkeeping format = %q, want %q", actual, expectedFormat)
	}

	tests := []struct {
		name  string
		apply func(*PersistedSchemaEntry)
	}{
		{"class", func(entry *PersistedSchemaEntry) { entry.StorageClass = StorageOffChain }},
		{"disposition", func(entry *PersistedSchemaEntry) { entry.MigrationDisposition = MigrationNone }},
		{"migration ID", func(entry *PersistedSchemaEntry) { entry.MigrationID = "changed" }},
		{"D4 format", func(entry *PersistedSchemaEntry) { entry.KeyOrFormat += " changed" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := clonePersistedSchemaFixture(t)
			test.apply(entryByID(t, &candidate, "veid-payload-cutover-bookkeeping"))
			if err := candidate.Validate(); err == nil {
				t.Fatal("invalid cutover bookkeeping contract accepted")
			}
		})
	}
}

func TestPersistedSchemaEvidencePathsExistAtPayloadHead(t *testing.T) {
	inventory := loadPersistedSchemaFixture(t)
	assertPersistedSchemaPayloadObjects(t, inventory)
}

func assertPersistedSchemaPayloadObjects(t *testing.T, inventory PersistedSchemaInventory) {
	t.Helper()
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Fatalf("git is required to verify persisted-schema payload objects: %v", err)
	}
	assertGitCommand(t, repoRoot, "cat-file", "-e", inventory.PayloadHead+"^{commit}")
	assertGitCommand(t, repoRoot, "merge-base", "--is-ancestor", PersistedSchemaEpochBase, inventory.PayloadHead)
	for _, entry := range inventory.Entries {
		paths := append([]string{entry.OwnerPath}, entry.EvidencePaths...)
		for _, repositoryPath := range paths {
			object := inventory.PayloadHead + ":" + repositoryPath
			assertGitCommand(t, repoRoot, "cat-file", "-e", object)
			if objectType := gitCommandOutput(t, repoRoot, "cat-file", "-t", object); objectType != "blob" {
				t.Errorf("entry %q payload object %q has type %q, want blob", entry.ID, repositoryPath, objectType)
			}
			size, err := strconv.ParseInt(gitCommandOutput(t, repoRoot, "cat-file", "-s", object), 10, 64)
			if err != nil {
				t.Fatalf("entry %q payload object %q has invalid size: %v", entry.ID, repositoryPath, err)
			}
			if size == 0 {
				t.Errorf("entry %q payload blob %q is empty", entry.ID, repositoryPath)
			}
			t.Logf("verified entry %q payload blob %q (%d bytes)", entry.ID, repositoryPath, size)
		}
	}
}

func assertGitCommand(t *testing.T, repoRoot string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v: %s", strings.Join(args, " "), err, output)
	}
}

func gitCommandOutput(t *testing.T, repoRoot string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v: %s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

func TestPersistedSchemaAuditedEntriesAreExact(t *testing.T) {
	inventory := loadPersistedSchemaFixture(t)
	tests := []struct {
		id          string
		owner       string
		class       StorageClass
		domain      string
		format      string
		disposition MigrationDisposition
		migrationID string
		evidence    []string
		blocker     string
	}{
		{
			"data-vault-artifact-index", "pkg/data_vault/fixture_artifact_store.go", StorageFixtureOnly, "data-vault artifact index",
			"fixture envelope v1 namespace/revision/checksum; index.json maps artifacts/blob_metadata/legal_holds/erasure_intents/mutations; objects/<sha256>; .fixture-artifact.lock",
			MigrationNewStore, "", []string{"pkg/data_vault/fixture_artifact_store.go", "pkg/data_vault/fixture_erasure.go"}, "",
		},
		{
			"data-vault-erasure-operation-contract", "pkg/data_vault/erasure_coordinator.go", StorageOffChain, "backend-neutral erasure operation",
			"ErasureOperation journal + storage/KMS receipts + replay transaction", MigrationIntegrationRequired, "",
			[]string{"pkg/data_vault/erasure_coordinator.go", "pkg/data_vault/contracts/evidence_lifecycle.go"},
			"production erasure requires durable transactional ErasureOperationStore; production storage deletion and KMS destruction adapters; independent receipt signers/key resolver; consent, hold, backup and finalization authorities; authenticated restore-manifest authority",
		},
		{
			"veid-evidence-reference", "x/veid/keeper/evidence_object_store.go", StorageConsensus, "VEID evidence object reference",
			"0xD4 || ASCII(\"reference/\") || objectCommitment -> JSON EvidenceObjectRef v1", MigrationRequiredUnwired, "veid-evidence-object-v1",
			[]string{"pkg/data_vault/contracts/evidence_lifecycle.go", "x/veid/keeper/evidence_object_store.go", "x/veid/types/keys.go"},
			"T4 upgrade owner must register the VEID evidence migration",
		},
		{
			"veid-evidence-quarantine", "x/veid/keeper/evidence_object_migration.go", StorageConsensus, "VEID evidence object quarantine",
			"0xD4 || ASCII(\"quarantine/\") || sourceKeyDigest -> JSON evidenceMigrationQuarantine v1", MigrationRequiredUnwired, "veid-evidence-object-v1",
			[]string{"x/veid/keeper/evidence_object_migration.go", "x/veid/types/keys.go"}, "T4 upgrade owner must register the VEID evidence migration",
		},
		{
			"veid-evidence-migration-bookkeeping", "x/veid/keeper/evidence_object_migration.go", StorageConsensus, "VEID evidence object migration",
			"D4 01 || UTF-8 upgradeID -> replay digest; D4 02 || upgradeID -> JSON report; singleton D4 03 -> manifest digest; D4 04 || UTF-8 signerKeyID -> BE64 epoch floor",
			MigrationRequiredUnwired, "veid-evidence-object-v1", []string{"x/veid/keeper/evidence_object_migration.go", "x/veid/types/keys.go"},
			"T4 upgrade owner must register the VEID evidence migration",
		},
		{
			"veid-payload-cutover-bookkeeping", "x/veid/keeper/evidence_payload_cutover.go", StorageConsensus, "VEID legacy payload cutover",
			"D4 05 || UTF-8 cutoverID -> signed replay digest; D4 06 || cutoverID -> JSON EvidencePayloadCutoverReport; singleton D4 07 -> latest cutover manifest digest; D4 08 || UTF-8 signerKeyID -> BE64 epoch floor",
			MigrationRequiredUnwired, "veid-legacy-payload-cutover-v1", []string{"x/veid/keeper/evidence_payload_cutover.go", "x/veid/keeper/evidence_payload_cutover_test.go"},
			"T4 must register signed cutover after evidence reference migration",
		},
		{
			"veid-legacy-payload-cutover", "x/veid/keeper/evidence_object_migration.go", StorageConsensus, "legacy VEID evidence payload cutover",
			"legacy PrefixScope 0x02 rows and classified shared-prefix 0x9A rows retain encrypted_payload until sanitize/delete cutover", MigrationRequiredUnwired, "veid-legacy-payload-cutover-v1",
			[]string{"x/veid/keeper/evidence_object_migration.go", "x/veid/keeper/evidence_object_migration_test.go"},
			"separate T4 migration must sanitize/delete legacy PrefixScope 0x02 rows and classified shared-prefix 0x9A rows",
		},
	}
	for _, test := range tests {
		entry := entryByID(t, &inventory, test.id)
		if entry.OwnerPath != test.owner || entry.StorageClass != test.class || entry.SchemaVersion != "v1" || entry.Domain != test.domain ||
			entry.KeyOrFormat != test.format || entry.MigrationDisposition != test.disposition || entry.MigrationID != test.migrationID ||
			!slices.Equal(entry.EvidencePaths, test.evidence) || entry.Blocker != test.blocker {
			t.Errorf("entry %q does not match the exact audit contract: %+v", test.id, *entry)
		}
	}
	for _, entry := range inventory.Entries {
		if entry.ID == "data-vault-erasure-journal" {
			t.Fatal("separate fixture erasure journal entry must not be present")
		}
	}
}

func TestPersistedSchemaNonPersistentSurfaceSetFailsClosed(t *testing.T) {
	tests := []struct {
		name  string
		apply func(*PersistedSchemaInventory)
	}{
		{"omission", func(value *PersistedSchemaInventory) { value.NonPersistentSurfaces = value.NonPersistentSurfaces[1:] }},
		{"unknown", func(value *PersistedSchemaInventory) { value.NonPersistentSurfaces[0] = "unknown-contracts" }},
		{"duplicate", func(value *PersistedSchemaInventory) {
			value.NonPersistentSurfaces = append(value.NonPersistentSurfaces, value.NonPersistentSurfaces[0])
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inventory := clonePersistedSchemaFixture(t)
			test.apply(&inventory)
			if err := inventory.Validate(); err == nil {
				t.Fatal("invalid non-persistent surface set accepted")
			}
		})
	}
}

func TestPersistedSchemaConsensusOwnershipIsExactAndProductionOwned(t *testing.T) {
	inventory := loadPersistedSchemaFixture(t)
	var consensus []string
	for _, entry := range inventory.Entries {
		if entry.StorageClass != StorageConsensus {
			continue
		}
		consensus = append(consensus, entry.ID)
		if nonProductionConsensusOwnerPath(entry.OwnerPath) {
			t.Fatalf("consensus entry %q has non-production owner %q", entry.ID, entry.OwnerPath)
		}
	}
	slices.Sort(consensus)
	want := []string{"fundauth-replay-store", "veid-evidence-migration-bookkeeping", "veid-evidence-quarantine", "veid-evidence-reference", "veid-legacy-payload-cutover", "veid-payload-cutover-bookkeeping"}
	if !slices.Equal(consensus, want) {
		t.Fatalf("consensus entries = %v, want %v", consensus, want)
	}
}
