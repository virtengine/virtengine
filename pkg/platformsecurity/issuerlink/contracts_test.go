package issuerlink

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func fixtureDigest(label string) string { return digest([]byte("fixture:" + label)) }

func issuerFixture(t *testing.T, visibility IssuerVisibility) (IssuerProfile, ed25519.PrivateKey, GovernedIssuerKeyResolverFixture) {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	seed[0] = 7
	privateKey := ed25519.NewKeyFromSeed(seed)
	profile := IssuerProfile{
		Version: Version1, ProfileID: "issuer-profile", IssuerCommitment: fixtureDigest("issuer"),
		ProgramDigest: fixtureDigest("program"), ProfileDigest: fixtureDigest("profile"), PolicyDigest: fixtureDigest("policy"),
		SigningKeyID: "issuer-key", KeyEpoch: 4, RetentionPolicyDigest: fixtureDigest("retention"),
		DeletionPolicyDigest: fixtureDigest("deletion"), Visibility: visibility, State: IssuerFixtureOnly,
		AttestationCapability: true, ExternalBlocker: ExternalBlockerRequired,
	}
	resolver := GovernedIssuerKeyResolverFixture{Entries: map[string]IssuerKeyEntry{
		IssuerKeyCoordinate(profile.IssuerCommitment, profile.SigningKeyID, profile.KeyEpoch): {Profile: profile, PublicKey: privateKey.Public().(ed25519.PublicKey)},
	}}
	return profile, privateKey, resolver
}

func programDomain() NullifierDomain {
	return NullifierDomain{Version: Version1, Kind: ProgramDomainKind, ProgramDigest: fixtureDigest("program")}
}

func pairwiseDomain(label string) NullifierDomain {
	return NullifierDomain{Version: Version1, Kind: PairwiseDomainKind, ProgramDigest: fixtureDigest("program"), RelyingPartyDigest: fixtureDigest(label)}
}

func walletKeyFixture(t *testing.T, marker byte) ed25519.PrivateKey {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	seed[0] = marker
	return ed25519.NewKeyFromSeed(seed)
}

func stableInputFixture(t *testing.T, label string) StableInputCommitment {
	t.Helper()
	stable, err := NewStableInputCommitment([]byte("confidential-stable-input-" + label))
	if err != nil {
		t.Fatal(err)
	}
	return stable
}

func requestFixture(t *testing.T, domain NullifierDomain, idempotency string, oldKey, newKey ed25519.PrivateKey, current uint64) WalletLinkRequest {
	t.Helper()
	newCommitment, err := WalletKeyCommitment(newKey.Public().(ed25519.PublicKey))
	if err != nil {
		t.Fatal(err)
	}
	request := WalletLinkRequest{
		Version: Version1, IdempotencyKey: idempotency, IssuerCommitment: fixtureDigest("issuer"), Domain: domain,
		PolicyDigest: fixtureDigest("policy"), ProfileDigest: fixtureDigest("profile"), NewWalletKeyCommitment: newCommitment,
		NewWalletKeyEpoch: 1, ChallengeNonceDigest: fixtureDigest("challenge-" + idempotency), EvidenceDigest: fixtureDigest("evidence-" + idempotency),
		CooldownUntilCoordinate: current + 10, ExpiresAtCoordinate: current + 20, CurrentCoordinate: current,
	}
	if oldKey != nil {
		request.OldWalletKeyCommitment, err = WalletKeyCommitment(oldKey.Public().(ed25519.PublicKey))
		if err != nil {
			t.Fatal(err)
		}
		request.OldWalletKeyEpoch = 1
		request.NewWalletKeyEpoch = 2
	}
	request, err = request.WithComputedDigest()
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func authorizationsFixture(t *testing.T, request WalletLinkRequest, oldKey, newKey ed25519.PrivateKey) WalletAuthorizations {
	t.Helper()
	message, err := request.SigningBytes()
	if err != nil {
		t.Fatal(err)
	}
	auth := WalletAuthorizations{NewWalletPublicKey: newKey.Public().(ed25519.PublicKey), NewWalletSignature: ed25519.Sign(newKey, message)}
	if oldKey != nil {
		auth.OldWalletPublicKey = oldKey.Public().(ed25519.PublicKey)
		auth.OldWalletSignature = ed25519.Sign(oldKey, message)
	}
	return auth
}

func attestationFixture(t *testing.T, request WalletLinkRequest, profile IssuerProfile, key ed25519.PrivateKey) IssuerAttestation {
	t.Helper()
	attestation := IssuerAttestation{
		Version: Version1, RequestDigest: request.RequestDigest, IssuerCommitment: request.IssuerCommitment,
		SigningKeyID: profile.SigningKeyID, SigningKeyEpoch: profile.KeyEpoch, IssuedAtCoordinate: request.CurrentCoordinate,
		ExpiresAtCoordinate: request.ExpiresAtCoordinate, NonceDigest: request.ChallengeNonceDigest,
	}
	attestation, err := SignIssuerAttestationFixture(attestation, key)
	if err != nil {
		t.Fatal(err)
	}
	return attestation
}

func TestThresholdPRFRequiresExactVerifiedCapability(t *testing.T) {
	profile, issuerKey, resolver := issuerFixture(t, VisibilityProgramScoped)
	request := requestFixture(t, programDomain(), "prf-a", nil, walletKeyFixture(t, 1), 100)
	stable := stableInputFixture(t, "person-a")
	authorization, input, err := VerifyIssuerAttestation(context.Background(), request, attestationFixture(t, request, profile, issuerKey), resolver, 100, AttestationPolicy{MaxLifetimeCoordinates: 20}, stable)
	if err != nil {
		t.Fatal(err)
	}
	prf, _ := NewDeterministicThresholdPRFFixture([]byte("issuer-link-fixture-prf-key"))
	first, err := prf.Evaluate(context.Background(), authorization, input)
	if err != nil || !validDigest(first) {
		t.Fatalf("verified PRF evaluation failed: %v", err)
	}
	otherRequest := requestFixture(t, programDomain(), "prf-b", nil, walletKeyFixture(t, 2), 100)
	_, otherInput, err := VerifyIssuerAttestation(context.Background(), otherRequest, attestationFixture(t, otherRequest, profile, issuerKey), resolver, 100, AttestationPolicy{MaxLifetimeCoordinates: 20}, stable)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := prf.Evaluate(context.Background(), authorization, otherInput); err == nil {
		t.Fatal("verified PRF capability authorized a different exact input")
	}
	if _, err := prf.Evaluate(context.Background(), VerifiedNullifierAuthorization{}, input); err == nil {
		t.Fatal("zero-value verified capability was accepted")
	}
	if prf.ProductionReady() || prf.FixtureState() != FixtureOnlyState || !strings.Contains(prf.ReviewStatus(), ProductionUnavailable) {
		t.Fatal("fixture PRF claimed production availability")
	}
}

func TestCanonicalWalletAuthorizationTamper(t *testing.T) {
	oldKey, newKey := walletKeyFixture(t, 3), walletKeyFixture(t, 4)
	request := requestFixture(t, programDomain(), "wallet-auth", oldKey, newKey, 100)
	auth := authorizationsFixture(t, request, oldKey, newKey)
	if err := VerifyWalletAuthorizations(request, auth); err != nil {
		t.Fatal(err)
	}
	for index, mutate := range []func(*WalletAuthorizations){
		func(value *WalletAuthorizations) { value.OldWalletSignature[0] ^= 1 },
		func(value *WalletAuthorizations) { value.NewWalletSignature[0] ^= 1 },
		func(value *WalletAuthorizations) { value.OldWalletSignature = nil },
		func(value *WalletAuthorizations) { value.NewWalletSignature = nil },
	} {
		candidate := auth
		candidate.OldWalletSignature = slicesClone(auth.OldWalletSignature)
		candidate.NewWalletSignature = slicesClone(auth.NewWalletSignature)
		mutate(&candidate)
		if VerifyWalletAuthorizations(request, candidate) == nil {
			t.Fatalf("wallet authorization attack %d accepted", index)
		}
	}
}

func TestGovernedIssuerAttestationNonceLifetimeAndKeyState(t *testing.T) {
	profile, issuerKey, resolver := issuerFixture(t, VisibilityProgramScoped)
	request := requestFixture(t, programDomain(), "attestation", nil, walletKeyFixture(t, 5), 100)
	stable := stableInputFixture(t, "attestation")
	attestation := attestationFixture(t, request, profile, issuerKey)
	if _, _, err := VerifyIssuerAttestation(context.Background(), request, attestation, resolver, 100, AttestationPolicy{MaxLifetimeCoordinates: 20}, stable); err != nil {
		t.Fatal(err)
	}
	wrongNonce := attestation
	wrongNonce.NonceDigest = fixtureDigest("wrong-nonce")
	wrongNonce, _ = SignIssuerAttestationFixture(wrongNonce, issuerKey)
	if _, _, err := VerifyIssuerAttestation(context.Background(), request, wrongNonce, resolver, 100, AttestationPolicy{MaxLifetimeCoordinates: 20}, stable); err == nil {
		t.Fatal("attestation with a different challenge nonce was accepted")
	}
	if _, _, err := VerifyIssuerAttestation(context.Background(), request, attestation, resolver, 100, AttestationPolicy{MaxLifetimeCoordinates: 19}, stable); err == nil {
		t.Fatal("attestation exceeding governed lifetime was accepted")
	}
	beyondRequest := attestation
	beyondRequest.ExpiresAtCoordinate++
	beyondRequest, _ = SignIssuerAttestationFixture(beyondRequest, issuerKey)
	if _, _, err := VerifyIssuerAttestation(context.Background(), request, beyondRequest, resolver, 100, AttestationPolicy{MaxLifetimeCoordinates: 30}, stable); err == nil {
		t.Fatal("attestation outliving its request was accepted")
	}
	revoked := profile
	revoked.State = IssuerRevoked
	revokedResolver := GovernedIssuerKeyResolverFixture{Entries: map[string]IssuerKeyEntry{
		IssuerKeyCoordinate(profile.IssuerCommitment, profile.SigningKeyID, profile.KeyEpoch): {Profile: revoked, PublicKey: issuerKey.Public().(ed25519.PublicKey)},
	}}
	if _, _, err := VerifyIssuerAttestation(context.Background(), request, attestation, revokedResolver, 100, AttestationPolicy{MaxLifetimeCoordinates: 20}, stable); err == nil {
		t.Fatal("revoked issuer key was accepted")
	}
}

func TestPublicSerializationAndRegistryShape(t *testing.T) {
	values := []any{IssuerProfile{}, WalletLinkRequest{}, IssuerAttestation{}, IssuerLinkRecord{}, NullifierDomain{}, NullifierInput{}, StableInputCommitment{}, IssuerLinkProjection{}, DeletionTombstone{}}
	for _, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(encoded))
		for _, forbidden := range []string{"issuer_subject", "subject_id", "wallet_address", "account_id", "email", "phone", "identity_attributes", "raw_evidence", "backend_uri", "key_share", "stable_input"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("%T serialization leaks forbidden field %q", value, forbidden)
			}
		}
	}
	if _, ok := reflect.TypeOf(LinkRegistration{}).FieldByName("Nullifier"); ok {
		t.Fatal("registration still exposes a caller-supplied nullifier")
	}
	registryType := reflect.TypeOf((*ConfidentialIssuerLinkRegistry)(nil)).Elem()
	if registryType.NumMethod() != 1 || registryType.Method(0).Name != "Lookup" || registryType.Method(0).Type.Out(0) != reflect.TypeOf(IssuerLinkProjection{}) {
		t.Fatal("confidential registry does not expose only minimal projection lookup")
	}
}

func slicesClone(value []byte) []byte { return append([]byte(nil), value...) }
