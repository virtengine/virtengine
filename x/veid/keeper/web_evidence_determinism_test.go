package keeper

// Consensus-determinism regression tests for the VEID web-evidence helpers.
//
// These helpers live in consensus policy code, so every traversal of a Go map
// must be order-independent: Go randomises map iteration order, and any helper
// that derives a digest or a mutation sequence from an unsorted traversal is a
// cross-node divergence hazard. Each helper below now collects keys, sorts
// them, and iterates the sorted slice; these tests pin that property so a
// future refactor back to a bare `range` over a map fails loudly.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

const (
	// webEvidenceMetadataDigestPinValue is sha256 of the canonical JSON
	// {"domain":"VEID_WEB_EVIDENCE_ATTESTATION_METADATA_V1","version":"1",
	//  "fields":[{"name":"a_key","value":"first"},{"name":"m_key","value":"middle"},
	//            {"name":"z_key","value":"last"}]}
	// i.e. the same metadata map with keys sorted ascending by name. If this
	// constant changes, the on-chain metadata digest domain or canonical form
	// changed and that is a state-breaking decision, not a refactor.

	// Field keys and values of the metadata fixtures above, named once so
	// goconst's three-occurrence threshold stays satisfied.
	webEvidenceKeyA                   = "a_key"
	webEvidenceKeyM                   = "m_key"
	webEvidenceKeyZ                   = "z_key"
	webEvidenceValueFirst             = "first"
	webEvidenceValueMid               = "middle"
	webEvidenceValueLast              = "last"
	webEvidenceMetadataDigestPinValue = "5882e75c41d3830a70f4b5f71e928ab6e1a93023c8321b5dc753f023a2c6529a"
)

// pinnedMetadata returns a metadata map whose keys are deliberately not in
// sorted order.
func pinnedMetadata() map[string]string {
	return map[string]string{
		webEvidenceKeyZ: webEvidenceValueLast,
		webEvidenceKeyA: webEvidenceValueFirst,
		webEvidenceKeyM: webEvidenceValueMid,
	}
}

func TestWebEvidenceMetadataDigestPinsCanonicalOrder(t *testing.T) {
	digest, err := webEvidenceMetadataDigest(pinnedMetadata())
	if err != nil {
		t.Fatalf("webEvidenceMetadataDigest returned error: %v", err)
	}
	if digest != webEvidenceMetadataDigestPinValue {
		t.Fatalf("metadata digest changed:\n got %s\nwant %s\n(the canonical metadata digest is consensus-visible; "+
			"only change this pin if a state migration is intended)", digest, webEvidenceMetadataDigestPinValue)
	}

	// Independently rebuild the canonical envelope (fields sorted by name) and
	// confirm the digest is exactly sha256 of that byte string. This catches a
	// regression where iteration order leaks into the marshalled slice.
	canonical := `{"domain":"VEID_WEB_EVIDENCE_ATTESTATION_METADATA_V1","version":"1","fields":[` +
		`{"name":"a_key","value":"first"},{"name":"m_key","value":"middle"},{"name":"z_key","value":"last"}]}`
	sum := sha256.Sum256([]byte(canonical))
	if want := hex.EncodeToString(sum[:]); digest != want {
		t.Fatalf("digest is not sha256 of the name-sorted canonical envelope:\n got %s\nwant %s", digest, want)
	}
}

// TestWebEvidenceMetadataDigestStableAcrossIterations exercises the helper many
// times and across maps built in different orders. Go randomises map iteration
// order per range statement, so an unsorted traversal would corrupt the
// marshalled field order and produce differing digests within a handful of
// runs; 512 iterations makes that failure essentially certain.
func TestWebEvidenceMetadataDigestStableAcrossIterations(t *testing.T) {
	const iterations = 512

	baseline, err := webEvidenceMetadataDigest(pinnedMetadata())
	if err != nil {
		t.Fatalf("baseline digest: %v", err)
	}

	for i := 0; i < iterations; i++ {
		// Rebuild the map fresh each iteration so insertion order and the
		// runtime's map seed vary.
		rebuilt := map[string]string{}
		rebuilt[webEvidenceKeyZ] = webEvidenceValueLast
		rebuilt[webEvidenceKeyA] = webEvidenceValueFirst
		rebuilt[webEvidenceKeyM] = webEvidenceValueMid

		digest, err := webEvidenceMetadataDigest(rebuilt)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if digest != baseline {
			t.Fatalf("iteration %d produced a different digest: got %s want %s "+
				"(map iteration order leaked into the digest)", i, digest, baseline)
		}
	}

	// A wider map: all 64 keys, inserted in reverse order.
	wide := map[string]string{}
	for i := 0; i < 64; i++ {
		wide[fmt.Sprintf("key_%02d", i)] = fmt.Sprintf("value_%02d", i)
	}
	wideDigest, err := webEvidenceMetadataDigest(wide)
	if err != nil {
		t.Fatalf("wide digest: %v", err)
	}
	for i := 0; i < 64; i++ {
		shuffled := map[string]string{}
		for j := 63; j >= 0; j-- {
			shuffled[fmt.Sprintf("key_%02d", j)] = fmt.Sprintf("value_%02d", j)
		}
		got, err := webEvidenceMetadataDigest(shuffled)
		if err != nil {
			t.Fatalf("wide iteration %d: %v", i, err)
		}
		if got != wideDigest {
			t.Fatalf("wide map iteration %d produced a different digest: got %s want %s", i, got, wideDigest)
		}
	}
}

func TestWebEvidenceMetadataDigestRejectsEmptyKey(t *testing.T) {
	if _, err := webEvidenceMetadataDigest(map[string]string{"": "value"}); err == nil {
		t.Fatal("expected an empty metadata key to be rejected")
	}
}

// TestWebEvidenceStorageMatchesIsOrderIndependent checks that the storage
// comparison helper is a pure set comparison: identical maps match regardless
// of the (randomised) order in which they are traversed, and any difference in
// keys or values fails.
func TestWebEvidenceStorageMatchesIsOrderIndependent(t *testing.T) {
	stored := map[string]string{webEvidenceKeyZ: webEvidenceValueLast, webEvidenceKeyA: webEvidenceValueFirst, webEvidenceKeyM: webEvidenceValueMid}

	const iterations = 256
	for i := 0; i < iterations; i++ {
		msg := map[string]string{}
		msg[webEvidenceKeyM] = webEvidenceValueMid
		msg[webEvidenceKeyZ] = webEvidenceValueLast
		msg[webEvidenceKeyA] = webEvidenceValueFirst
		if !webEvidenceStorageMatches("backend", "ref", stored, "backend", "ref", msg) {
			t.Fatalf("iteration %d: equal metadata maps did not match", i)
		}
	}

	cases := []struct {
		name string
		msg  map[string]string
	}{
		{"differing value", map[string]string{webEvidenceKeyZ: webEvidenceValueLast, webEvidenceKeyA: "CHANGED", webEvidenceKeyM: webEvidenceValueMid}},
		{"missing key", map[string]string{webEvidenceKeyZ: webEvidenceValueLast, webEvidenceKeyA: webEvidenceValueFirst}},
		{"extra key", map[string]string{webEvidenceKeyZ: webEvidenceValueLast, webEvidenceKeyA: webEvidenceValueFirst, webEvidenceKeyM: webEvidenceValueMid, "extra": "x"}},
		{"renamed key", map[string]string{webEvidenceKeyZ: webEvidenceValueLast, webEvidenceKeyA: webEvidenceValueFirst, "m_other": webEvidenceValueMid}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if webEvidenceStorageMatches("backend", "ref", stored, "backend", "ref", tc.msg) {
				t.Fatalf("%s: expected mismatch", tc.name)
			}
		})
	}

	if webEvidenceStorageMatches("backend", "ref", stored, "other", "ref", stored) {
		t.Fatal("expected a differing backend to fail the comparison")
	}
	if webEvidenceStorageMatches("backend", "ref", stored, "backend", "other", stored) {
		t.Fatal("expected a differing ref to fail the comparison")
	}
}

// TestValidatePayloadFreeEvidenceJSONSortsForbiddenFieldError asserts the
// forbidden-field scan reports the alphabetically first offending key. Before
// the traversal was sorted, two offending keys made the returned key (and
// therefore the error string) nondeterministic.
func TestValidatePayloadFreeEvidenceJSONSortsForbiddenFieldError(t *testing.T) {
	// Both keys contain the forbidden substring "backend_uri"; the sorted
	// first is "alpha_backend_uri".
	payload := []byte(`{"encrypted_payload":"x","zeta_backend_uri":"a","alpha_backend_uri":"b"}`)
	err := validatePayloadFreeEvidenceJSON(payload)
	if err == nil {
		t.Fatal("expected payload-free evidence JSON validation to reject forbidden fields")
	}
	want := "evidence reference contains forbidden field alpha_backend_uri"
	if err.Error() != want {
		t.Fatalf("nondeterministic/wrong forbidden-field error:\n got %q\nwant %q", err.Error(), want)
	}

	// Repeated calls must return the identical error string.
	for i := 0; i < 256; i++ {
		if got := validatePayloadFreeEvidenceJSON(payload); got == nil || got.Error() != want {
			t.Fatalf("iteration %d: unstable forbidden-field error: %v", i, got)
		}
	}

	if err := validatePayloadFreeEvidenceJSON([]byte(`{"commitment":"abc","size_bytes":12}`)); err != nil {
		t.Fatalf("clean payload-free evidence JSON was rejected: %v", err)
	}
	if err := validatePayloadFreeEvidenceJSON([]byte(`not json`)); err == nil {
		t.Fatal("expected malformed JSON to be rejected")
	}
}

// TestValidateEvidencePayloadCutoverDeleteSourceIsOrderIndependent asserts the
// evidence-record-like scan rejects a social-scope delete source for every
// offending field position and returns a stable error. The scan's error string
// is constant, so the sorted traversal exists to keep the map-iteration
// property structural; this test pins that detection is complete (a match at
// any key position is caught) and repeatable.
func TestValidateEvidencePayloadCutoverDeleteSourceIsOrderIndependent(t *testing.T) {
	const want = "social scope delete source looks evidence-record-like"

	// Keys must match an evidenceFields entry exactly after lowercasing; two
	// case variants of the same field exercise distinct traversal positions.
	value := []byte(`{"encrypted_payload":{"ciphertext":"x"},"EVIDENCE_ID":"1","Evidence_Type":"doc"}`)
	err := validateEvidencePayloadCutoverDeleteSource("social_scope", value)
	if err == nil {
		t.Fatal("expected an evidence-record-like delete source to be rejected")
	}
	if err.Error() != want {
		t.Fatalf("unexpected error: got %q want %q", err.Error(), want)
	}
	for i := 0; i < 256; i++ {
		if got := validateEvidencePayloadCutoverDeleteSource("social_scope", value); got == nil || got.Error() != want {
			t.Fatalf("iteration %d: unstable evidence-field error: %v", i, got)
		}
	}

	// A single evidence-like field must be caught regardless of its position in
	// the JSON object, including when it is the last key.
	for _, offending := range []string{
		`{"encrypted_payload":{"ciphertext":"x"},"confidence":"0.9"}`,
		`{"confidence":"0.9","encrypted_payload":{"ciphertext":"x"}}`,
		`{"encrypted_payload":{"ciphertext":"x"},"Override":"true"}`,
		`{"Override":"true","encrypted_payload":{"ciphertext":"x"}}`,
	} {
		if got := validateEvidencePayloadCutoverDeleteSource("social_scope", []byte(offending)); got == nil || got.Error() != want {
			t.Fatalf("offending social_scope delete source not rejected: %s -> %v", offending, got)
		}
	}

	// A non-social_scope source kind is not subject to the evidence-field check.
	if err := validateEvidencePayloadCutoverDeleteSource("other", value); err != nil {
		t.Fatalf("non-social_scope delete source should pass: %v", err)
	}
	// A clean social-scope delete source passes.
	clean := []byte(`{"encrypted_payload":{"ciphertext":"x"}}`)
	if err := validateEvidencePayloadCutoverDeleteSource("social_scope", clean); err != nil {
		t.Fatalf("clean social_scope delete source should pass: %v", err)
	}
	// Missing payload marker is rejected deterministically.
	if err := validateEvidencePayloadCutoverDeleteSource("social_scope", []byte(`{"other":"x"}`)); err == nil {
		t.Fatal("expected a delete source without an encrypted payload marker to be rejected")
	}
	// Malformed JSON is rejected.
	if err := validateEvidencePayloadCutoverDeleteSource("social_scope", []byte(`not json`)); err == nil {
		t.Fatal("expected malformed JSON to be rejected")
	}
}
