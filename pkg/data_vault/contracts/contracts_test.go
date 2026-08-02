package contracts

import (
	"errors"
	"testing"
)

type restartProbe struct {
	evidence RestartEvidence
	err      error
}

func (p restartProbe) ProbeRestart(VaultKMSProfile) (RestartEvidence, error) {
	return p.evidence, p.err
}

type restoreProbe struct {
	evidence RestoreEvidence
	err      error
}

func (p restoreProbe) ProbeRestore(VaultKMSProfile, RestoreManifest) (RestoreEvidence, error) {
	return p.evidence, p.err
}

type holdVerifier struct{ err error }

func (v holdVerifier) VerifyHoldAuthority(LegalHoldAuthority) error { return v.err }

type erasureProbe struct {
	evidence ErasureEvidence
	err      error
}

func (p erasureProbe) ProbeErasure(VaultKMSProfile, ErasureTombstone) (ErasureEvidence, error) {
	return p.evidence, p.err
}

func fixtureProfile() VaultKMSProfile {
	return VaultKMSProfile{
		Version: Version1, ID: "vault-fixture", State: ProfileFixtureOnly,
		BlobBackend: "postgres", MetadataBackend: "postgres", KMSProvider: "vault-dev-server",
		Dependency88DHash: "sha256:88d",
	}
}

func TestValidateRestart(t *testing.T) {
	complete := RestartEvidence{true, true, true, true}
	cases := []struct {
		name    string
		profile VaultKMSProfile
		probe   RestartProbe
		wantErr bool
	}{
		{"complete", fixtureProfile(), restartProbe{evidence: complete}, false},
		{"missing digest", withProfile(func(p *VaultKMSProfile) { p.Dependency88DHash = "" }), restartProbe{evidence: complete}, true},
		{"memory backend", withProfile(func(p *VaultKMSProfile) { p.BlobBackend = "memory" }), restartProbe{evidence: complete}, true},
		{"production state capped", withProfile(func(p *VaultKMSProfile) { p.State = ProfileProduction }), restartProbe{evidence: complete}, true},
		{"unknown state", withProfile(func(p *VaultKMSProfile) { p.State = "future" }), restartProbe{evidence: complete}, true},
		{"unknown version", withProfile(func(p *VaultKMSProfile) { p.Version = 2 }), restartProbe{evidence: complete}, true},
		{"lost wrapped keys", fixtureProfile(), restartProbe{evidence: RestartEvidence{true, false, true, true}}, true},
		{"probe failure", fixtureProfile(), restartProbe{err: errors.New("offline")}, true},
		{"missing probe", fixtureProfile(), nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidateRestart(tc.profile, tc.probe); (got != nil) != tc.wantErr {
				t.Fatalf("ValidateRestart() error = %v, wantErr %v", got, tc.wantErr)
			}
		})
	}
}

func TestValidateRestore(t *testing.T) {
	manifest := RestoreManifest{
		Version: Version1, ManifestDigest: "manifest", ObjectVersionIDs: []string{"object-v1", "object-v2"},
		WrappedKeyIDs: []string{"key-1", "key-2"}, AuditCheckpoint: "audit", ConsentLinkDigest: "consent",
	}
	complete := RestoreEvidence{
		ObjectVersionIDs: []string{"object-v2", "object-v1"}, WrappedKeyIDs: []string{"key-2", "key-1"},
		AuditCheckpoint: "audit", ConsentLinkDigest: "consent",
	}
	cases := []struct {
		name     string
		manifest RestoreManifest
		evidence RestoreEvidence
		wantErr  bool
	}{
		{"complete", manifest, complete, false},
		{"omitted object", manifest, withRestore(complete, func(e *RestoreEvidence) { e.ObjectVersionIDs = []string{"object-v1"} }), true},
		{"omitted key", manifest, withRestore(complete, func(e *RestoreEvidence) { e.WrappedKeyIDs = []string{"key-1"} }), true},
		{"substituted key", manifest, withRestore(complete, func(e *RestoreEvidence) { e.WrappedKeyIDs[0] = "new-key" }), true},
		{"regenerated key", manifest, withRestore(complete, func(e *RestoreEvidence) { e.GeneratedKeyCount = 1 }), true},
		{"audit mismatch", manifest, withRestore(complete, func(e *RestoreEvidence) { e.AuditCheckpoint = "other" }), true},
		{"duplicate manifest key", withManifest(manifest, func(m *RestoreManifest) { m.WrappedKeyIDs = []string{"key-1", "key-1"} }), complete, true},
		{"unknown manifest version", withManifest(manifest, func(m *RestoreManifest) { m.Version = 9 }), complete, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRestore(fixtureProfile(), tc.manifest, restoreProbe{evidence: tc.evidence})
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateRestore() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateHold(t *testing.T) {
	authority := LegalHoldAuthority{
		Version: Version1, HoldID: "hold-1", State: HoldActive, AuthorityType: "x/gov",
		PolicyDigest: "policy", EvidenceDigest: "evidence", Approvals: 3, Threshold: 2,
	}
	cases := []struct {
		name      string
		authority LegalHoldAuthority
		verifier  HoldAuthorityVerifier
		wantErr   bool
	}{
		{"authorized", authority, holdVerifier{}, false},
		{"group authorized", withHold(authority, func(h *LegalHoldAuthority) { h.AuthorityType = "x/group" }), holdVerifier{}, false},
		{"operator denied", withHold(authority, func(h *LegalHoldAuthority) { h.AuthorityType = "operator" }), holdVerifier{}, true},
		{"unilateral denied", withHold(authority, func(h *LegalHoldAuthority) { h.Threshold = 1; h.Approvals = 1 }), holdVerifier{}, true},
		{"below threshold", withHold(authority, func(h *LegalHoldAuthority) { h.Approvals = 1 }), holdVerifier{}, true},
		{"unknown state", withHold(authority, func(h *LegalHoldAuthority) { h.State = "future" }), holdVerifier{}, true},
		{"unknown version", withHold(authority, func(h *LegalHoldAuthority) { h.Version = 2 }), holdVerifier{}, true},
		{"verifier denied", authority, holdVerifier{err: errors.New("bad signature")}, true},
		{"missing verifier", authority, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidateHold(tc.authority, tc.verifier); (got != nil) != tc.wantErr {
				t.Fatalf("ValidateHold() error = %v, wantErr %v", got, tc.wantErr)
			}
		})
	}
}

func TestValidateErasure(t *testing.T) {
	released := LegalHoldAuthority{
		Version: Version1, HoldID: "hold-1", State: HoldReleased, AuthorityType: "x/gov",
		PolicyDigest: "policy", EvidenceDigest: "evidence", Approvals: 2, Threshold: 2,
	}
	tombstone := ErasureTombstone{
		Version: Version1, ObjectID: "object-1", AuthorizationDigest: "authorization",
		StorageReceipts: []DestructionReceipt{{TargetID: "blob", Digest: "deleted"}},
		KeyReceipts:     []DestructionReceipt{{TargetID: "dek", Digest: "destroyed"}},
		Holds:           []LegalHoldAuthority{released},
	}
	complete := ErasureEvidence{StorageDeleted: true, KeysDestroyed: true, Undecryptable: true}
	cases := []struct {
		name      string
		tombstone ErasureTombstone
		evidence  ErasureEvidence
		wantErr   bool
	}{
		{"proven", tombstone, complete, false},
		{"active hold", withTombstone(tombstone, func(e *ErasureTombstone) { e.Holds[0].State = HoldActive }), complete, true},
		{"missing deletion receipt", withTombstone(tombstone, func(e *ErasureTombstone) { e.StorageReceipts = nil }), complete, true},
		{"missing destruction receipt", withTombstone(tombstone, func(e *ErasureTombstone) { e.KeyReceipts = nil }), complete, true},
		{"empty receipt digest", withTombstone(tombstone, func(e *ErasureTombstone) { e.KeyReceipts[0].Digest = "" }), complete, true},
		{"still decryptable", tombstone, ErasureEvidence{StorageDeleted: true, KeysDestroyed: true}, true},
		{"storage remains", tombstone, ErasureEvidence{KeysDestroyed: true, Undecryptable: true}, true},
		{"unknown tombstone version", withTombstone(tombstone, func(e *ErasureTombstone) { e.Version = 2 }), complete, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateErasure(fixtureProfile(), tc.tombstone, erasureProbe{evidence: tc.evidence})
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateErasure() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestVersionedSupportingContractsDenyUnknownState(t *testing.T) {
	consent := ConsentDecision{Version: Version1, SubjectID: "subject", Purpose: "verify", Scope: "veid", PolicyDigest: "policy", State: "future"}
	if err := consent.Validate(); err == nil {
		t.Fatal("unknown consent state accepted")
	}
	rotation := RotationState{Version: Version1, RotationID: "rotation", FromKey: "old", ToKey: "new", Status: "future", TotalObjects: 1}
	if err := rotation.Validate(); err == nil {
		t.Fatal("unknown rotation state accepted")
	}
	key := WrappedKeyMetadata{Version: 2, ObjectID: "object", WrappedKeyID: "key", KEKVersion: "kek", CiphertextHash: "hash"}
	if err := key.Validate(); err == nil {
		t.Fatal("unknown wrapped-key version accepted")
	}
}

func withProfile(change func(*VaultKMSProfile)) VaultKMSProfile {
	value := fixtureProfile()
	change(&value)
	return value
}

func withRestore(value RestoreEvidence, change func(*RestoreEvidence)) RestoreEvidence {
	value.ObjectVersionIDs = append([]string(nil), value.ObjectVersionIDs...)
	value.WrappedKeyIDs = append([]string(nil), value.WrappedKeyIDs...)
	change(&value)
	return value
}

func withManifest(value RestoreManifest, change func(*RestoreManifest)) RestoreManifest {
	value.ObjectVersionIDs = append([]string(nil), value.ObjectVersionIDs...)
	value.WrappedKeyIDs = append([]string(nil), value.WrappedKeyIDs...)
	change(&value)
	return value
}

func withHold(value LegalHoldAuthority, change func(*LegalHoldAuthority)) LegalHoldAuthority {
	change(&value)
	return value
}

func withTombstone(value ErasureTombstone, change func(*ErasureTombstone)) ErasureTombstone {
	value.StorageReceipts = append([]DestructionReceipt(nil), value.StorageReceipts...)
	value.KeyReceipts = append([]DestructionReceipt(nil), value.KeyReceipts...)
	value.Holds = append([]LegalHoldAuthority(nil), value.Holds...)
	change(&value)
	return value
}
