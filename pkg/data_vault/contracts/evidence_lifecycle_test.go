package contracts

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

type receiptKeyResolver map[string]ed25519.PublicKey

func (r receiptKeyResolver) ResolveDeletionReceiptKey(kind DeletionReceiptKind, keyID string, keyEpoch uint64) (ed25519.PublicKey, error) {
	key, found := r[string(kind)+"/"+keyID]
	if !found || keyEpoch != 1 {
		return nil, errors.New("key not found")
	}
	return key, nil
}

type receiptReplayStore struct {
	consumed map[string]string
	fail     bool
}

func (s *receiptReplayStore) ConsumeDeletionReceipts(storage, kms DeletionReceiptReplay, apply func() error) error {
	if s.fail {
		return errors.New("durable store unavailable")
	}
	if s.consumed == nil {
		s.consumed = make(map[string]string)
	}
	for _, receipt := range []DeletionReceiptReplay{storage, kms} {
		if _, found := s.consumed[receipt.OperationID]; found {
			return errors.New("receipt operation replayed")
		}
		for _, digest := range s.consumed {
			if digest == receipt.ReceiptDigest {
				return errors.New("receipt digest replayed")
			}
		}
	}
	if err := apply(); err != nil {
		return err
	}
	s.consumed[storage.OperationID] = storage.ReceiptDigest
	s.consumed[kms.OperationID] = kms.ReceiptDigest
	return nil
}

func TestEvidenceCommitmentsAreRandomizedDomainSeparatedAndPayloadFree(t *testing.T) {
	fields := testEvidenceFields()
	first, firstOpening, err := CreateEvidenceObjectRef(rand.Reader, "identity-scope", fields)
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := CreateEvidenceObjectRef(rand.Reader, "identity-scope", fields)
	if err != nil {
		t.Fatal(err)
	}
	otherDomain, _, err := CreateEvidenceObjectRef(bytes.NewReader(firstOpening), "social-scope", fields)
	if err != nil {
		t.Fatal(err)
	}
	if first.ObjectCommitment == second.ObjectCommitment {
		t.Fatal("same evidence reused a commitment")
	}
	if first.ObjectCommitment == otherDomain.ObjectCommitment {
		t.Fatal("commitment domains were not separated")
	}
	if err := VerifyEvidenceObjectRef(first, firstOpening); err != nil {
		t.Fatal(err)
	}
	if err := VerifyEvidenceObjectRef(first, make([]byte, CommitmentOpeningSize)); err == nil {
		t.Fatal("zero opening accepted")
	}
	if err := VerifyEvidenceObjectRef(first, firstOpening[:31]); err == nil {
		t.Fatal("malformed opening accepted")
	}
	if _, _, err := CreateEvidenceObjectRef(bytes.NewReader(make([]byte, CommitmentOpeningSize)), "identity-scope", fields); err == nil {
		t.Fatal("zero random opening accepted")
	}
	for _, domain := range []string{"https://example.test/evidence", "customer-secret", "custom"} {
		if _, _, err := CreateEvidenceObjectRef(rand.Reader, domain, fields); err == nil {
			t.Fatalf("unsupported commitment domain %q accepted", domain)
		}
	}

	bz, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(bz))
	for _, forbidden := range []string{"ciphertext", "opening", "backend_ref", "backend_uri", "wrapped_key", "nonce"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("serialized reference contains %q: %s", forbidden, bz)
		}
	}
}

func TestDeletionReceiptsRequireIndependentMatchingClaims(t *testing.T) {
	ref, _, err := CreateEvidenceObjectRef(rand.Reader, "identity-scope", testEvidenceFields())
	if err != nil {
		t.Fatal(err)
	}
	ref, err = TransitionRetention(ref, RetentionDeletionScheduled, 11, 101)
	if err != nil {
		t.Fatal(err)
	}
	ref, err = TransitionRetention(ref, RetentionDeletionPending, 12, 102)
	if err != nil {
		t.Fatal(err)
	}

	storagePublic, storagePrivate, _ := ed25519.GenerateKey(rand.Reader)
	kmsPublic, kmsPrivate, _ := ed25519.GenerateKey(rand.Reader)
	resolver := receiptKeyResolver{
		string(ReceiptStorageDeletion) + "/storage": storagePublic,
		string(ReceiptKMSDestruction) + "/kms":      kmsPublic,
	}
	authorization := digest("authorization")
	storage := testReceipt(ReceiptStorageDeletion, "storage", ref, authorization, 13, 103)
	kms := testReceipt(ReceiptKMSDestruction, "kms", ref, authorization, 14, 104)
	signReceipt(t, &storage, storagePrivate)
	signReceipt(t, &kms, kmsPrivate)

	firstSignBytes, _ := storage.CanonicalSignBytes()
	expectedLegacyBytes := canonicalValues("virtengine/evidence-deletion-receipt/v1", fmt.Sprint(storage.Version), string(storage.Kind),
		storage.TargetCommitment, storage.AuthorizationDigest, storage.PolicyDigest, storage.ProfileDigest, storage.OperationID,
		fmt.Sprint(storage.KeyEpoch), fmt.Sprint(storage.CompletedHeight), fmt.Sprint(storage.CompletedUnix),
		storage.SignerKeyID, fmt.Sprint(storage.SignerKeyEpoch))
	if !bytes.Equal(firstSignBytes, expectedLegacyBytes) {
		t.Fatal("legacy deletion receipt canonical bytes changed")
	}
	storage.Signature = append([]byte(nil), storage.Signature...)
	secondSignBytes, _ := storage.CanonicalSignBytes()
	if !bytes.Equal(firstSignBytes, secondSignBytes) {
		t.Fatal("signature changed canonical sign bytes")
	}
	replayStore := &receiptReplayStore{}
	resolution := deletionResolution(ref, authorization, replayStore)
	resolved, err := ResolveDeletion(ref, resolution, []DeletionReceipt{storage, kms}, resolver)
	if err != nil || resolved.State != RetentionDeleted {
		t.Fatalf("valid receipts resolved to %s", resolved.State)
	}
	if replayed, err := ResolveDeletion(ref, resolution, []DeletionReceipt{storage, kms}, resolver); err == nil || replayed.State != RetentionDeletionUnresolved {
		t.Fatal("durable receipt replay resolved deletion")
	}

	forged := kms
	forged.Signature = append([]byte(nil), kms.Signature...)
	forged.Signature[0] ^= 1
	cases := []struct {
		name     string
		receipts []DeletionReceipt
		hold     bool
	}{
		{"incomplete", []DeletionReceipt{storage}, false},
		{"duplicate kind", []DeletionReceipt{storage, storage}, false},
		{"forged", []DeletionReceipt{storage, forged}, false},
		{"legal hold", []DeletionReceipt{storage, kms}, true},
	}
	replayed := kms
	replayed.OperationID = storage.OperationID
	signReceipt(t, &replayed, kmsPrivate)
	cases = append(cases, struct {
		name     string
		receipts []DeletionReceipt
		hold     bool
	}{"replayed operation", []DeletionReceipt{storage, replayed}, false})
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			context := deletionResolution(ref, authorization, &receiptReplayStore{})
			context.LegalHold = test.hold
			result, err := ResolveDeletion(ref, context, test.receipts, resolver)
			if err == nil {
				t.Fatal("invalid deletion resolution returned no error")
			}
			expected := RetentionDeletionUnresolved
			if test.hold {
				expected = RetentionDeletionPending
			}
			if result.State != expected {
				t.Fatalf("false claim resolved to %s", result.State)
			}
		})
	}
	mismatch := storage
	mismatch.ProfileDigest = digest("wrong-profile")
	signReceipt(t, &mismatch, storagePrivate)
	if result, err := ResolveDeletion(ref, deletionResolution(ref, authorization, &receiptReplayStore{}), []DeletionReceipt{mismatch, kms}, resolver); err == nil || result.State != RetentionDeletionUnresolved {
		t.Fatal("mismatched receipt resolved deletion")
	}
	stale := storage
	stale.CompletedHeight = ref.UpdatedHeight - 1
	signReceipt(t, &stale, storagePrivate)
	if result, err := ResolveDeletion(ref, deletionResolution(ref, authorization, &receiptReplayStore{}), []DeletionReceipt{stale, kms}, resolver); err == nil || result.State != RetentionDeletionUnresolved {
		t.Fatal("stale deletion receipt resolved deletion")
	}
	future := storage
	future.CompletedHeight = 16
	signReceipt(t, &future, storagePrivate)
	if result, err := ResolveDeletion(ref, deletionResolution(ref, authorization, &receiptReplayStore{}), []DeletionReceipt{future, kms}, resolver); err == nil || result.State != RetentionDeletionUnresolved {
		t.Fatal("future deletion receipt resolved deletion")
	}
	aliasedResolver := receiptKeyResolver{
		string(ReceiptStorageDeletion) + "/storage": storagePublic,
		string(ReceiptKMSDestruction) + "/kms":      storagePublic,
	}
	aliasedKMS := kms
	signReceipt(t, &aliasedKMS, storagePrivate)
	if result, err := ResolveDeletion(ref, deletionResolution(ref, authorization, &receiptReplayStore{}), []DeletionReceipt{storage, aliasedKMS}, aliasedResolver); err == nil || result.State != RetentionDeletionUnresolved {
		t.Fatal("aliased deletion authority keys resolved deletion")
	}
	failingStore := &receiptReplayStore{fail: true}
	if result, err := ResolveDeletion(ref, deletionResolution(ref, authorization, failingStore), []DeletionReceipt{storage, kms}, resolver); err == nil || result.State != RetentionDeletionUnresolved || len(failingStore.consumed) != 0 {
		t.Fatal("consumer failure did not roll back deletion resolution")
	}
	callbackStore := &receiptReplayStore{}
	callbackFailure := deletionResolution(ref, authorization, callbackStore)
	callbackFailure.ApplyResolved = func(EvidenceObjectRef) error { return errors.New("state write failed") }
	if result, err := ResolveDeletion(ref, callbackFailure, []DeletionReceipt{storage, kms}, resolver); err == nil || result.State != RetentionDeletionUnresolved || len(callbackStore.consumed) != 0 {
		t.Fatal("protected callback failure did not roll back replay consumption")
	}
}

func TestRetentionTransitionTableBlocksLegalHoldResolution(t *testing.T) {
	ref, _, err := CreateEvidenceObjectRef(rand.Reader, "identity-scope", testEvidenceFields())
	if err != nil {
		t.Fatal(err)
	}
	hold, err := TransitionRetention(ref, RetentionLegalHold, 11, 101)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := TransitionRetention(hold, RetentionDeleted, 12, 102); err == nil {
		t.Fatal("legal hold transitioned directly to deleted")
	}
	resolution := deletionResolution(hold, digest("authorization"), &receiptReplayStore{})
	resolution.LegalHold = true
	if resolved, err := ResolveDeletion(hold, resolution, nil, nil); err == nil || resolved.State != RetentionLegalHold {
		t.Fatal("deletion resolution changed legal hold state")
	}
	deleted := ref
	deleted.State = RetentionDeleted
	if resolved, err := ResolveDeletion(deleted, deletionResolution(deleted, digest("authorization"), &receiptReplayStore{}), nil, nil); err == nil || resolved.State != RetentionDeleted {
		t.Fatal("deletion resolution downgraded deleted state")
	}
	if CanTransitionRetention(RetentionDeleted, RetentionActive) {
		t.Fatal("deleted state was not terminal")
	}
}

func testEvidenceFields() EvidenceObjectFields {
	return EvidenceObjectFields{
		EvidenceDigest: digest("evidence"), RetentionPolicyDigest: digest("retention"),
		PolicyDigest: digest("policy"), ProfileDigest: digest("profile"), KeyEpoch: 7,
		CreatedHeight: 10, CreatedUnix: 100, ExpiresHeight: 1000, ExpiresUnix: 10000,
	}
}

func testReceipt(kind DeletionReceiptKind, keyID string, ref EvidenceObjectRef, authorization string, height, unixTime int64) DeletionReceipt {
	return DeletionReceipt{
		Version: EvidenceObjectRefVersion, Kind: kind, TargetCommitment: ref.ObjectCommitment,
		AuthorizationDigest: authorization, PolicyDigest: ref.PolicyDigest, ProfileDigest: ref.ProfileDigest,
		OperationID: "operation-" + keyID, KeyEpoch: ref.KeyEpoch, CompletedHeight: height,
		CompletedUnix: unixTime, SignerKeyID: keyID, SignerKeyEpoch: 1,
	}
}

func deletionResolution(ref EvidenceObjectRef, authorization string, replayStore DeletionReceiptReplayConsumer) DeletionResolutionContext {
	return DeletionResolutionContext{
		AuthorizationDigest: authorization, PolicyDigest: ref.PolicyDigest, ProfileDigest: ref.ProfileDigest,
		CurrentHeight: 15, CurrentUnix: 105, ReplayConsumer: replayStore,
		ApplyResolved: func(EvidenceObjectRef) error { return nil },
	}
}

func signReceipt(t *testing.T, receipt *DeletionReceipt, privateKey ed25519.PrivateKey) {
	t.Helper()
	signBytes, err := receipt.CanonicalSignBytes()
	if err != nil {
		t.Fatal(err)
	}
	receipt.Signature = ed25519.Sign(privateKey, signBytes)
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
