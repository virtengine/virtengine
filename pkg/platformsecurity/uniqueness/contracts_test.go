package uniqueness

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func digestFixture(label string) string { return digest([]byte("fixture:" + label)) }

func custodyFixture(t *testing.T) (CustodyNodeSet, map[string]ed25519.PrivateKey) {
	t.Helper()
	nodes := make([]CustodyNodeIdentity, 3)
	keys := make(map[string]ed25519.PrivateKey)
	for index := range nodes {
		seed := make([]byte, ed25519.SeedSize)
		seed[0] = byte(index + 1)
		privateKey := ed25519.NewKeyFromSeed(seed)
		keyID := "key-" + string(rune('a'+index))
		keys[keyID] = privateKey
		nodes[index] = CustodyNodeIdentity{
			Version: Version1, NodeID: "node-" + string(rune('a'+index)), SigningKeyID: keyID, SigningKeyEpoch: 7,
			OperatorCommitment: digestFixture("operator-" + keyID), FailureDomainCommitment: digestFixture("zone-" + keyID),
			SigningPublicKey: privateKey.Public().(ed25519.PublicKey), EncryptionCommitment: digestFixture("enc-" + keyID),
			ShareCommitment: digestFixture("share-" + keyID), EndpointCommitment: digestFixture("endpoint-" + keyID),
			NodeKeyBindingCommitment: digestFixture("binding-" + keyID), State: NodeActive,
		}
	}
	return CustodyNodeSet{Version: Version1, Threshold: 2, Nodes: nodes}, keys
}

func keyEpochFixture(t *testing.T, nodes CustodyNodeSet) ThresholdKeyEpoch {
	t.Helper()
	participantDigest, err := nodes.Digest()
	if err != nil {
		t.Fatal(err)
	}
	return ThresholdKeyEpoch{Version: Version1, ProfileID: "profile", Epoch: 7, Threshold: 2, ParticipantSetDigest: participantDigest, PublicCommitment: digestFixture("public-7"), TransformDigest: digestFixture("transform-7"), ActivationCoordinate: 100, ExpiryCoordinate: 200, State: KeyEpochActive}
}

func profileFixture() CancellableTemplateProfile {
	return CancellableTemplateProfile{Version: Version1, ProfileID: "profile", Purpose: UniquenessPurpose, AlgorithmDigest: digestFixture("algorithm"), ProfileDigest: digestFixture("profile"), VersionDigest: digestFixture("version"), DistanceThreshold: 0, TransformKeyEpoch: 7, OutputRules: OutputCommitmentRules{Version: Version1, Algorithm: "sha256-commitment", OutputBytes: 32, DomainBinding: true}, State: FixtureOnlyState, ExternalReviewBlock: ExternalReviewRequired}
}

func payloadFixture(t *testing.T, nodes CustodyNodeSet, now time.Time) CustodyReceiptPayload {
	t.Helper()
	participantDigest, err := nodes.Digest()
	if err != nil {
		t.Fatal(err)
	}
	return CustodyReceiptPayload{Version: Version1, ProgramDigest: digestFixture("program"), PolicyDigest: digestFixture("policy"), ProfileDigest: digestFixture("profile"), ModelDigest: digestFixture("model"), RuntimeDigest: digestFixture("runtime"), EvidenceDigest: digestFixture("evidence"), RequestDigest: digestFixture("request"), Decision: OutcomeFinalUnique, Reason: ReasonFinalUnique, KeyEpoch: 7, ParticipantSetDigest: participantDigest, Threshold: nodes.Threshold, Freshness: Freshness{IssuedAt: now.Unix() - 1, ExpiresAt: now.Unix() + 60, Nonce: "nonce"}, AppealReference: digestFixture("appeal"), ScopedNullifier: digestFixture("nullifier")}
}

func TestStrictPublicValidationAndNodeAuthority(t *testing.T) {
	nodes, _ := custodyFixture(t)
	if err := nodes.Validate(); err != nil {
		t.Fatal(err)
	}
	first, _ := nodes.CanonicalBytes()
	reversed := nodes
	reversed.Nodes = append([]CustodyNodeIdentity(nil), nodes.Nodes...)
	reversed.Nodes[0], reversed.Nodes[2] = reversed.Nodes[2], reversed.Nodes[0]
	second, _ := reversed.CanonicalBytes()
	if string(first) != string(second) {
		t.Fatal("node set canonical bytes depend on input order")
	}
	epoch := keyEpochFixture(t, nodes)
	if err := epoch.ValidateWithNodeSet(nodes); err != nil {
		t.Fatal(err)
	}

	cases := []func(*CustodyNodeSet){
		func(set *CustodyNodeSet) { set.Nodes[1].NodeID = "Node-A" },
		func(set *CustodyNodeSet) { set.Nodes[1].NodeID = set.Nodes[0].NodeID },
		func(set *CustodyNodeSet) { set.Nodes[1].SigningKeyID = set.Nodes[0].SigningKeyID },
		func(set *CustodyNodeSet) { set.Nodes[1].SigningPublicKey = set.Nodes[0].SigningPublicKey },
		func(set *CustodyNodeSet) { set.Nodes[1].OperatorCommitment = set.Nodes[0].OperatorCommitment },
		func(set *CustodyNodeSet) { set.Nodes[1].FailureDomainCommitment = set.Nodes[0].FailureDomainCommitment },
		func(set *CustodyNodeSet) { set.Nodes[1].ShareCommitment = "ABC" },
	}
	for index, mutate := range cases {
		candidate, _ := custodyFixture(t)
		mutate(&candidate)
		if err := candidate.Validate(); err == nil {
			t.Fatalf("node attack %d accepted", index)
		}
	}
	mismatch := epoch
	mismatch.ParticipantSetDigest = digestFixture("other")
	if err := mismatch.ValidateWithNodeSet(nodes); err == nil {
		t.Fatal("participant-set substitution accepted")
	}
	mismatch = epoch
	mismatch.Threshold++
	if err := mismatch.ValidateWithNodeSet(nodes); err == nil {
		t.Fatal("threshold substitution accepted")
	}
}

func TestCanonicalEncodingsBindEveryField(t *testing.T) {
	nodes, _ := custodyFixture(t)
	now := time.Unix(1000, 0)
	payload := payloadFixture(t, nodes, now)
	first, err := payload.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	mutations := []func(*CustodyReceiptPayload){
		func(p *CustodyReceiptPayload) { p.ProgramDigest = digestFixture("other-program") },
		func(p *CustodyReceiptPayload) { p.Reason = ReasonCancelled },
		func(p *CustodyReceiptPayload) { p.KeyEpoch++ },
		func(p *CustodyReceiptPayload) { p.ParticipantSetDigest = digestFixture("other-set") },
		func(p *CustodyReceiptPayload) { p.Threshold++ },
		func(p *CustodyReceiptPayload) { p.Freshness.Nonce = "other-nonce" },
		func(p *CustodyReceiptPayload) { p.AppealReference = digestFixture("other-appeal") },
	}
	for index, mutate := range mutations {
		candidate := payload
		mutate(&candidate)
		encoded, err := candidate.CanonicalBytes()
		if err != nil || string(encoded) == string(first) {
			t.Fatalf("field mutation %d not bound: %v", index, err)
		}
	}
	bad := payload
	bad.ProgramDigest = "program"
	if _, err := bad.CanonicalBytes(); err == nil {
		t.Fatal("non-digest public data accepted")
	}
	bad = payload
	bad.Reason = ReasonCode("free text")
	if _, err := bad.CanonicalBytes(); err == nil {
		t.Fatal("arbitrary reason accepted")
	}
	bad = payload
	bad.AppealReference = "ticket text"
	if _, err := bad.CanonicalBytes(); err == nil {
		t.Fatal("arbitrary appeal text accepted")
	}
}

func TestQuorumAttestationAuthorityFreshnessAndTamper(t *testing.T) {
	now := time.Unix(1000, 0)
	nodes, keys := custodyFixture(t)
	epoch := keyEpochFixture(t, nodes)
	payload := payloadFixture(t, nodes, now)
	attestor := DeterministicQuorumAttestorFixture{Nodes: nodes, PrivateKeys: keys}
	attestation, err := attestor.Attest(context.Background(), payload, epoch)
	if err != nil {
		t.Fatal(err)
	}
	resolver := DeterministicCustodyKeyResolverFixture{Epochs: map[uint64]CustodyNodeSet{7: nodes}}
	verify := func(p CustodyReceiptPayload, a QuorumAttestation) error {
		return VerifyQuorumAttestation(context.Background(), p, a, resolver, now, 2*time.Minute, time.Second)
	}
	if err := verify(payload, attestation); err != nil {
		t.Fatal(err)
	}
	attacks := []func(*CustodyReceiptPayload, *QuorumAttestation){
		func(p *CustodyReceiptPayload, _ *QuorumAttestation) { p.PolicyDigest = digestFixture("other-policy") },
		func(_ *CustodyReceiptPayload, a *QuorumAttestation) { a.Signatures = a.Signatures[:1] },
		func(_ *CustodyReceiptPayload, a *QuorumAttestation) { a.Signatures[0].NodeID = "unknown" },
		func(_ *CustodyReceiptPayload, a *QuorumAttestation) { a.Signatures[1].NodeID = a.Signatures[0].NodeID },
		func(_ *CustodyReceiptPayload, a *QuorumAttestation) { a.Signatures[0].SigningKeyEpoch++ },
		func(_ *CustodyReceiptPayload, a *QuorumAttestation) {
			a.ParticipantSetDigest = digestFixture("other-set")
		},
		func(_ *CustodyReceiptPayload, a *QuorumAttestation) { a.Threshold++ },
		func(_ *CustodyReceiptPayload, a *QuorumAttestation) { a.Signatures[0].Signature[0] ^= 1 },
	}
	for index, attack := range attacks {
		p := payload
		a := attestation
		a.Signatures = append([]NodeSignature(nil), attestation.Signatures...)
		for i := range a.Signatures {
			a.Signatures[i].Signature = append([]byte(nil), a.Signatures[i].Signature...)
		}
		attack(&p, &a)
		if err := verify(p, a); err == nil {
			t.Fatalf("attestation attack %d accepted", index)
		}
	}
	for name, freshness := range map[string]Freshness{
		"expired":   {IssuedAt: 800, ExpiresAt: 900, Nonce: "nonce"},
		"future":    {IssuedAt: 1010, ExpiresAt: 1050, Nonce: "nonce"},
		"oversized": {IssuedAt: 900, ExpiresAt: 1100, Nonce: "nonce"},
	} {
		t.Run(name, func(t *testing.T) {
			p := payload
			p.Freshness = freshness
			if err := verify(p, attestation); err == nil {
				t.Fatal("invalid freshness accepted")
			}
		})
	}
}

func signRotation(t *testing.T, rotation CompromiseRotation, nodes CustodyNodeSet, keys map[string]ed25519.PrivateKey) CompromiseRotation {
	t.Helper()
	value, err := rotation.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range nodes.Nodes[:nodes.Threshold] {
		rotation.Approvals = append(rotation.Approvals, NodeSignature{NodeID: node.NodeID, SigningKeyID: node.SigningKeyID, SigningKeyEpoch: rotation.OldEpoch, Signature: ed25519.Sign(keys[node.SigningKeyID], value)})
	}
	return rotation
}

func TestCompromiseRotationRequiresVerifiedFrozenAuthority(t *testing.T) {
	nodes, keys := custodyFixture(t)
	setDigest, _ := nodes.Digest()
	frozen := CompromiseRotation{Version: Version1, OldEpoch: 7, NewEpoch: 8, CompromiseDigest: digestFixture("incident"), ParticipantSetDigest: setDigest, Threshold: nodes.Threshold, FreezeCoordinate: 120, ActivationCoordinate: 130, ReenrollmentRequired: true, State: CompromiseFrozen}
	approved := frozen
	approved.State = CompromiseApproved
	complete := approved
	complete.State = CompromiseComplete
	if err := ValidateCompromiseTransition(frozen, approved); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCompromiseTransition(approved, complete); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyCompromiseRotation(complete, nodes); err == nil {
		t.Fatal("unsigned rotation verified")
	}
	signed := signRotation(t, complete, nodes, keys)
	verified, err := VerifyCompromiseRotation(signed, nodes)
	if err != nil {
		t.Fatal(err)
	}
	if err := verified.CanIssue(7, 120); err == nil {
		t.Fatal("old epoch issued after freeze")
	}
	if err := verified.CanIssue(8, 130); err != nil {
		t.Fatal(err)
	}
	attacks := []func(*CompromiseRotation){
		func(r *CompromiseRotation) { r.Approvals = r.Approvals[:1] },
		func(r *CompromiseRotation) { r.Approvals[1].NodeID = r.Approvals[0].NodeID },
		func(r *CompromiseRotation) { r.Approvals[0].NodeID = "unknown" },
		func(r *CompromiseRotation) { r.Approvals[0].Signature[0] ^= 1 },
		func(r *CompromiseRotation) { r.ActivationCoordinate++ },
	}
	for index, attack := range attacks {
		candidate := signed
		candidate.Approvals = append([]NodeSignature(nil), signed.Approvals...)
		for i := range candidate.Approvals {
			candidate.Approvals[i].Signature = append([]byte(nil), candidate.Approvals[i].Signature...)
		}
		attack(&candidate)
		if _, err := VerifyCompromiseRotation(candidate, nodes); err == nil {
			t.Fatalf("rotation attack %d accepted", index)
		}
	}
	mutated := approved
	mutated.FreezeCoordinate++
	if err := ValidateCompromiseTransition(frozen, mutated); err == nil {
		t.Fatal("field mutation accepted")
	}
	if err := ValidateCompromiseTransition(complete, approved); err == nil {
		t.Fatal("state rollback accepted")
	}
}

func TestPublicSerializationHasNoCandidateOrBiometricLeakage(t *testing.T) {
	values := []any{EnrollmentRequest{}, EnrollmentRecord{}, CustodyReceiptPayload{}, QuorumAttestation{}, CandidateSearchResult{}, profileFixture()}
	for _, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(encoded))
		for _, forbidden := range []string{"record_digest", "distance_fixed", "account_id", "subject_id", "global_identifier", "raw_template", "embedding", "template_bytes"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("%T serialization leaks %q", value, forbidden)
			}
		}
	}
	resultType := reflect.TypeOf(CandidateSearchResult{})
	for index := 0; index < resultType.NumField(); index++ {
		name := strings.ToLower(resultType.Field(index).Name)
		if strings.Contains(name, "record") || strings.Contains(name, "distance") {
			t.Fatalf("candidate boundary exports %q", name)
		}
	}
}
