package federation

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRequestCanonicalBytesAreDeterministic(t *testing.T) {
	discoveryDigest := [32]byte{1}
	timestamp := time.Unix(1_800_000_000, 0).UTC()
	first, err := NewRequestBinding("provider-1", "orders", "post", "/v1/orders?z=2&a=1", "application/json; charset=utf-8", []byte(`{"id":"order-1"}`), timestamp, "nonce-1", 7, discoveryDigest, "virtengine-1")
	if err != nil {
		t.Fatalf("build request binding: %v", err)
	}
	second := first
	second.Query = "a=1&z=2"
	firstBytes, err := first.CanonicalBytes()
	if err != nil {
		t.Fatalf("canonical bytes: %v", err)
	}
	secondBytes, err := second.CanonicalBytes()
	if err != nil {
		t.Fatalf("second canonical bytes: %v", err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatal("equivalent request targets produced different bytes")
	}
}

func TestSignedDiscoveryValidationAndRollback(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	document, privateKey := fixtureDiscovery(now)
	digest, err := document.Digest()
	if err != nil {
		t.Fatalf("discovery digest: %v", err)
	}
	canonical, _ := document.CanonicalBytes()
	signed := SignedDiscoveryDocument{Document: document, Signature: ed25519.Sign(privateKey, canonical)}
	record := ServiceRecord{
		Version: Version1, ProviderID: document.ProviderID, ServiceID: document.ServiceID, Revision: document.Revision,
		DiscoveryDigest: digest, ActiveKeyEpoch: document.SigningKeyEpoch, State: StateFixtureOnly,
	}
	if err := signed.Verify(record, nil, now); err != nil {
		t.Fatalf("valid discovery rejected: %v", err)
	}
	disabled := record
	disabled.State = StateDisabled
	if err := disabled.Validate(); err != nil {
		t.Fatalf("disabled service record is not structurally valid: %v", err)
	}
	if err := signed.Verify(disabled, nil, now); err == nil {
		t.Fatal("disabled service record authenticated discovery")
	}
	if err := signed.Verify(record, nil, document.ExpiresAt); err == nil {
		t.Fatal("stale discovery accepted")
	}
	if err := signed.Verify(record, &document, now); err == nil {
		t.Fatal("discovery rollback accepted")
	}
	wrongProvider := record
	wrongProvider.ProviderID = "provider-2"
	if err := signed.Verify(wrongProvider, nil, now); err == nil {
		t.Fatal("cross-provider discovery replay accepted")
	}
}

func TestDiscoveryRejectsDuplicatesAndInvalidEpochs(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	document, _ := fixtureDiscovery(now)
	document.Capabilities = append(document.Capabilities, document.Capabilities[0])
	if err := document.Validate(); err == nil {
		t.Fatal("duplicate capability accepted")
	}
	document, _ = fixtureDiscovery(now)
	document.Endpoints = append(document.Endpoints, Endpoint{Name: "duplicate-name", URL: document.Endpoints[0].URL})
	if err := document.Validate(); err == nil {
		t.Fatal("duplicate endpoint accepted")
	}
	document, _ = fixtureDiscovery(now)
	duplicate := document.KeyEpochs[0]
	duplicate.Epoch++
	duplicate.PublicKey = append([]byte(nil), duplicate.PublicKey...)
	document.KeyEpochs = append(document.KeyEpochs, duplicate)
	if err := document.Validate(); err == nil {
		t.Fatal("key epoch without key rotation accepted")
	}
}

func TestTrustProfilesAndOriginPolicy(t *testing.T) {
	pin := [32]byte{1}
	profiles := []ClientTrustProfile{
		{Version: Version1, Mode: TrustNativeSPKI, SPKISHA256: [][32]byte{pin}},
		{Version: Version1, Mode: TrustNativeCA, CABundleSHA256: pin},
		{Version: Version1, Mode: TrustGatewayProxy, GatewayOrigin: "https://gateway.example", GatewayVerificationKey: make([]byte, ed25519.PublicKeySize)},
		{Version: Version1, Mode: TrustBrowserSameOrigin, ApplicationOrigin: "https://app.example"},
		{Version: Version1, Mode: TrustBrowserAppContinuity, ApplicationOrigin: "https://app.example", ApplicationContinuity: pin},
	}
	for _, profile := range profiles {
		if err := profile.Validate(); err != nil {
			t.Fatalf("valid trust profile %q rejected: %v", profile.Mode, err)
		}
	}
	if err := profiles[3].ValidateOrigin("https://app.example", "https://app.example/federation"); err != nil {
		t.Fatalf("same-origin request rejected: %v", err)
	}
	if err := profiles[3].ValidateOrigin("https://evil.example", "https://app.example/federation"); err == nil {
		t.Fatal("cross-origin browser request accepted")
	}
	unsupported := ClientTrustProfile{Version: Version1, Mode: "browser_spki"}
	if err := unsupported.Validate(); err == nil {
		t.Fatal("unsupported trust profile accepted")
	}
	claimsSPKI := profiles[3]
	claimsSPKI.ClaimsBrowserPeerSPKI = true
	if err := claimsSPKI.Validate(); err == nil {
		t.Fatal("browser peer SPKI inspection claim accepted")
	}
	ambiguous := profiles[0]
	ambiguous.CABundleSHA256 = pin
	if err := ambiguous.Validate(); err == nil {
		t.Fatal("ambiguous trust profile accepted")
	}
}

func TestRequestPolicyBindings(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	document, privateKey := fixtureDiscovery(now)
	record := fixtureServiceRecord(t, document)
	binding := fixtureBinding(t, document, now, []byte(`{"id":"order-1"}`), "nonce-bindings")
	policy := fixtureRoutePolicy()
	if err := policy.Authorize(record, binding, document, uint64(len(`{"id":"order-1"}`)), now); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*RequestBinding)
	}{
		{name: "provider", mutate: func(value *RequestBinding) { value.ProviderID = "provider-2" }},
		{name: "service", mutate: func(value *RequestBinding) { value.ServiceID = "billing" }},
		{name: "route", mutate: func(value *RequestBinding) { value.Path = "/v1/other" }},
		{name: "method", mutate: func(value *RequestBinding) { value.Method = "GET" }},
		{name: "chain", mutate: func(value *RequestBinding) { value.ChainID = "other-chain" }},
		{name: "key epoch", mutate: func(value *RequestBinding) { value.SigningKeyEpoch++ }},
		{name: "discovery", mutate: func(value *RequestBinding) { value.DiscoveryDigest[0] ^= 0xff }},
		{name: "expired", mutate: func(value *RequestBinding) { value.Timestamp = now.Add(-3 * time.Minute) }},
		{name: "future", mutate: func(value *RequestBinding) { value.Timestamp = now.Add(31 * time.Second) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := binding
			test.mutate(&candidate)
			if err := policy.Authorize(record, candidate, document, uint64(len(`{"id":"order-1"}`)), now); err == nil {
				t.Fatal("mismatched request accepted")
			}
		})
	}
	canonical, _ := binding.CanonicalBytes()
	signature := ed25519.Sign(privateKey, canonical)
	if err := VerifyAndConsume(context.Background(), &nonceStore{}, policy, record, binding, document, []byte(`{"id":"tampered"}`), now, signature, func() error { return nil }); err == nil {
		t.Fatal("wrong body accepted")
	}
	if err := policy.Authorize(record, binding, document, policy.MaxBodyBytes+1, now); err == nil {
		t.Fatal("oversized body accepted")
	}
	missingCapability := document
	missingCapability.Capabilities = []string{"orders.read"}
	if err := policy.Authorize(record, binding, missingCapability, uint64(len(`{"id":"order-1"}`)), now); err == nil {
		t.Fatal("request without required capability accepted")
	}
	disabled := record
	disabled.State = StateDisabled
	if err := policy.Authorize(disabled, binding, document, uint64(len(`{"id":"order-1"}`)), now); err == nil {
		t.Fatal("disabled service record authorized request")
	}
	oldSeed := sha256.Sum256([]byte("virtengine-federation-retired-fixture-key-v1"))
	oldPrivateKey := ed25519.NewKeyFromSeed(oldSeed[:])
	overlap := document
	overlap.KeyEpochs = append(overlap.KeyEpochs, KeyEpoch{
		Epoch: 6, PublicKey: oldPrivateKey.Public().(ed25519.PublicKey),
		NotBefore: now.Add(-2 * time.Hour), NotAfter: now.Add(30 * time.Minute),
	})
	overlapDigest, err := overlap.Digest()
	if err != nil {
		t.Fatalf("overlapping discovery digest: %v", err)
	}
	retired := binding
	retired.SigningKeyEpoch = 6
	retired.DiscoveryDigest = overlapDigest
	overlapRecord := record
	overlapRecord.DiscoveryDigest = overlapDigest
	if err := policy.Authorize(overlapRecord, retired, overlap, uint64(len(`{"id":"order-1"}`)), now); err == nil {
		t.Fatal("retired overlapping key epoch authorized request")
	}
	unauthenticated := policy
	unauthenticated.Auth = AuthNone
	if err := unauthenticated.Validate(); err == nil {
		t.Fatal("unauthenticated route policy accepted")
	}
}

func TestAtomicNonceConsumeOnSuccess(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	document, privateKey := fixtureDiscovery(now)
	record := fixtureServiceRecord(t, document)
	body := []byte(`{"id":"order-1"}`)
	binding := fixtureBinding(t, document, now, body, "nonce-atomic")
	canonical, _ := binding.CanonicalBytes()
	signature := ed25519.Sign(privateKey, canonical)
	store := &nonceStore{}
	disabled := record
	disabled.State = StateDisabled
	protectedCalled := false
	if err := VerifyAndConsume(context.Background(), store, fixtureRoutePolicy(), disabled, binding, document, body, now, signature, func() error {
		protectedCalled = true
		return nil
	}); err == nil {
		t.Fatal("disabled service record consumed request")
	}
	if store.calls != 0 || protectedCalled {
		t.Fatal("disabled request reached nonce store or protected mutation")
	}
	protectedErr := errors.New("protected mutation failed")
	if err := VerifyAndConsume(context.Background(), store, fixtureRoutePolicy(), record, binding, document, body, now, signature, func() error { return protectedErr }); !errors.Is(err, protectedErr) {
		t.Fatalf("protected failure not returned: %v", err)
	}
	if len(store.used) != 0 {
		t.Fatal("nonce burned when protected mutation failed")
	}
	if err := VerifyAndConsume(context.Background(), store, fixtureRoutePolicy(), record, binding, document, body, now, signature, func() error { return nil }); err != nil {
		t.Fatalf("valid retry rejected: %v", err)
	}
	if err := VerifyAndConsume(context.Background(), store, fixtureRoutePolicy(), record, binding, document, body, now, signature, func() error { return nil }); err == nil {
		t.Fatal("request replay accepted")
	}
	badSignature := append([]byte(nil), signature...)
	badSignature[0] ^= 0xff
	fresh := binding
	fresh.Nonce = "nonce-invalid-proof"
	if err := VerifyAndConsume(context.Background(), store, fixtureRoutePolicy(), record, fresh, document, body, now, badSignature, func() error { return nil }); err == nil {
		t.Fatal("invalid signature accepted")
	}
	if store.calls != 3 {
		t.Fatalf("nonce store called before proof validation: calls=%d", store.calls)
	}
}

func TestVersionStateAndKeyEpochDowngrade(t *testing.T) {
	record := ServiceRecord{
		Version: Version1, ProviderID: "provider-1", ServiceID: "orders", Revision: 1,
		DiscoveryDigest: [32]byte{1}, ActiveKeyEpoch: 1, State: StateFixtureOnly,
	}
	if err := record.Validate(); err != nil {
		t.Fatalf("fixture record rejected: %v", err)
	}
	record.Version++
	if err := record.Validate(); err == nil {
		t.Fatal("unknown service record version accepted")
	}
	record.Version = Version1
	record.State = StateSandbox
	if err := record.Validate(); err == nil {
		t.Fatal("service record exceeded fixture-only cap")
	}
	now := time.Unix(1_800_000_000, 0).UTC()
	publicA, _, _ := ed25519.GenerateKey(bytes.NewReader(bytes.Repeat([]byte{1}, 32)))
	publicB, _, _ := ed25519.GenerateKey(bytes.NewReader(bytes.Repeat([]byte{2}, 32)))
	current := KeyEpoch{Epoch: 5, PublicKey: publicA, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour)}
	next := KeyEpoch{Epoch: 6, PublicKey: publicB, NotBefore: now, NotAfter: now.Add(2 * time.Hour)}
	if err := ValidateKeyEpochTransition(current, next); err != nil {
		t.Fatalf("valid key transition rejected: %v", err)
	}
	next.Epoch = current.Epoch
	if err := ValidateKeyEpochTransition(current, next); err == nil {
		t.Fatal("key epoch downgrade accepted")
	}
}

type goldenVector struct {
	Name            string `json:"name"`
	Domain          string `json:"domain"`
	ProviderID      string `json:"provider_id"`
	ServiceID       string `json:"service_id"`
	Method          string `json:"method"`
	RequestTarget   string `json:"request_target"`
	BodyBase64      string `json:"body_base64"`
	ContentType     string `json:"content_type"`
	TimestampUnix   int64  `json:"timestamp_unix"`
	Nonce           string `json:"nonce"`
	SigningKeyEpoch uint64 `json:"signing_key_epoch"`
	DiscoveryHex    string `json:"discovery_digest_hex"`
	ChainID         string `json:"chain_id"`
	CanonicalBase64 string `json:"canonical_base64"`
	DigestHex       string `json:"digest_hex"`
}

func TestJSONGoldenVector(t *testing.T) {
	encoded, err := os.ReadFile("testdata/request-v1.json")
	if err != nil {
		t.Fatalf("read golden vector: %v", err)
	}
	var vector goldenVector
	if err := json.Unmarshal(encoded, &vector); err != nil {
		t.Fatalf("parse golden vector: %v", err)
	}
	if vector.Name != "request-v1" || vector.Domain != requestDomain {
		t.Fatal("unexpected golden vector identity")
	}
	body, err := base64.StdEncoding.DecodeString(vector.BodyBase64)
	if err != nil {
		t.Fatalf("decode body: %v", err)
	}
	discoveryBytes, err := hex.DecodeString(vector.DiscoveryHex)
	if err != nil || len(discoveryBytes) != sha256.Size {
		t.Fatal("invalid discovery digest")
	}
	var discoveryDigest [32]byte
	copy(discoveryDigest[:], discoveryBytes)
	binding, err := NewRequestBinding(vector.ProviderID, vector.ServiceID, vector.Method, vector.RequestTarget, vector.ContentType, body, time.Unix(vector.TimestampUnix, 0).UTC(), vector.Nonce, vector.SigningKeyEpoch, discoveryDigest, vector.ChainID)
	if err != nil {
		t.Fatalf("build vector binding: %v", err)
	}
	canonical, err := binding.CanonicalBytes()
	if err != nil {
		t.Fatalf("canonical vector: %v", err)
	}
	actualBase64 := base64.StdEncoding.EncodeToString(canonical)
	actualDigest := sha256.Sum256(canonical)
	actualDigestHex := hex.EncodeToString(actualDigest[:])
	if actualBase64 != vector.CanonicalBase64 || actualDigestHex != vector.DigestHex {
		t.Fatalf("golden mismatch\ncanonical_base64=%s\ndigest_hex=%s", actualBase64, actualDigestHex)
	}
	typeScript, err := os.ReadFile("request_v1_vector.ts")
	if err != nil {
		t.Fatalf("read TypeScript vector: %v", err)
	}
	if !strings.Contains(string(typeScript), vector.CanonicalBase64) || !strings.Contains(string(typeScript), vector.DigestHex) {
		t.Fatal("TypeScript vector drifted from language-neutral JSON vector")
	}
}

type nonceStore struct {
	used  map[string]struct{}
	calls int
}

func (store *nonceStore) WithNonce(_ context.Context, scope, nonce string, protected func() error) error {
	store.calls++
	if store.used == nil {
		store.used = make(map[string]struct{})
	}
	key := scope + "\x00" + nonce
	if _, exists := store.used[key]; exists {
		return errors.New("replay")
	}
	if err := protected(); err != nil {
		return err
	}
	store.used[key] = struct{}{}
	return nil
}

func fixtureDiscovery(now time.Time) (DiscoveryDocument, ed25519.PrivateKey) {
	seed := sha256.Sum256([]byte("virtengine-federation-fixture-key-v1"))
	privateKey := ed25519.NewKeyFromSeed(seed[:])
	document := DiscoveryDocument{
		Version: Version1, ProviderID: "provider-1", ServiceID: "orders", Revision: 1,
		Capabilities: []string{"orders.read", "orders.write"},
		Endpoints:    []Endpoint{{Name: "api", URL: "https://provider.example/federation"}},
		KeyEpochs: []KeyEpoch{{
			Epoch: 7, PublicKey: privateKey.Public().(ed25519.PublicKey),
			NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		}},
		IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(30 * time.Minute), SigningKeyEpoch: 7,
	}
	return document, privateKey
}

func fixtureServiceRecord(t *testing.T, document DiscoveryDocument) ServiceRecord {
	t.Helper()
	digest, err := document.Digest()
	if err != nil {
		t.Fatalf("discovery digest: %v", err)
	}
	return ServiceRecord{
		Version: Version1, ProviderID: document.ProviderID, ServiceID: document.ServiceID, Revision: document.Revision,
		DiscoveryDigest: digest, ActiveKeyEpoch: document.SigningKeyEpoch, State: StateFixtureOnly,
	}
}

func fixtureBinding(t *testing.T, document DiscoveryDocument, now time.Time, body []byte, nonce string) RequestBinding {
	t.Helper()
	discoveryDigest, err := document.Digest()
	if err != nil {
		t.Fatalf("discovery digest: %v", err)
	}
	binding, err := NewRequestBinding(document.ProviderID, document.ServiceID, "POST", "/v1/orders?limit=10&sort=asc", "application/json", body, now, nonce, document.SigningKeyEpoch, discoveryDigest, "virtengine-1")
	if err != nil {
		t.Fatalf("request binding: %v", err)
	}
	return binding
}

func fixtureRoutePolicy() RoutePolicy {
	return RoutePolicy{
		Version: Version1, ChainID: "virtengine-1", ServiceID: "orders", Method: "POST", ExactPath: "/v1/orders",
		Auth: AuthEd25519, RequiredCapability: "orders.write", MaxBodyBytes: 1024,
		MaxClockSkew: 30 * time.Second, MaxRequestAge: 2 * time.Minute, ContentTypes: []string{"application/json"},
	}
}
