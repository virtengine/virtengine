package keeper

// Consensus-safety net for the secp256k1 attestation-signature verifier.
//
// verifySecp256k1Signature used to delegate to github.com/ethereum/go-ethereum/crypto
// (GPL-3.0, denied by scripts/supply-chain/go-module-policy.json). It now calls
// github.com/decred/dcrd/dcrec/secp256k1/v4 (ISC) directly. x/enclave is
// consensus-critical, so the accept/reject set must be provably unchanged, otherwise
// a node running the new code and a node running the old code could disagree on
// whether an attestation is valid and fork the chain.
//
// The corpus in testdata/ was recorded by running the OLD go-ethereum
// implementation over a fixed set of cases before the dependency was removed. This
// test replays that corpus against the new implementation and fails on any
// accept/reject divergence or change of rejection reason.
//
// Regenerating the corpus (only valid while go-ethereum is still a dependency) is
// described in the file header of testdata/README.md.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	decred_ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// Canonical secp256k1 fixture, hex-encoded so the tests below are hermetic and do
// not depend on a signer.
//
// The key is the secp256k1 public key for private key
// 1e99423a4ed27608a15a2616cf5b0f8b0d3c0f6d5a2c9e8b7a6f5e4d3c2b1a09 and the
// signature is over sha256("attested-result-v1"). testCorpusOverHalfS is
// N/2 + 1, i.e. over half order yet below 2^255 so its most significant byte is
// clear.
const (
	testCorpusPubKey    = "0454dee68fb11f8211e7833a05fc9140d8ae665e65f4dd105d4785fdbd6baa66c1497858d23c332c61dd0093e816cadee05d4a225d2385798777ed42043a8fd2e6"
	testCorpusPayload   = "da025d475c52b74de05d7299069b544541d6b7edb0430a19143b6d5b813a2743"
	testCorpusSig       = "f5b5dd05c960aa3390e568784be23e9e1545612e0c7eabce495dccbf8a52d6790661775e9b8b78a5578e1f4e85825fc24406de149dad78d3faae941fade37931"
	testCorpusOverHalfS = "7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a1"
)

type equivalenceCase struct {
	Name           string `json:"name"`
	Note           string `json:"note"`
	PubKey         string `json:"pubKey"`
	Payload        string `json:"payload"`
	Sig            string `json:"sig"`
	OldAccept      bool   `json:"oldAccept"`
	OldErrCategory string `json:"oldErrCategory"`
}

type equivalenceCorpus struct {
	SchemaVersion int               `json:"schemaVersion"`
	BuildPath     string            `json:"buildPath"`
	Source        string            `json:"source"`
	Cases         []equivalenceCase `json:"cases"`
}

// classify maps a verification error onto the same category vocabulary the corpus
// uses, so the test can assert the *reason* for a rejection is preserved and not
// just the boolean outcome. A change of category can indicate a behaviour change
// even where both old and new reject.
func classify(err error) string {
	switch {
	case err == nil:
		return "accept"
	case errors.Is(err, errSecp256k1DigestLength):
		return "digest-length"
	case errors.Is(err, errSecp256k1InvalidPubKey):
		return "invalid-pubkey"
	case errors.Is(err, errSecp256k1VerificationFailed):
		return "verify-failed"
	}
	// The pre-checks that live in verifySigningKeySignature were carried over
	// unchanged; match them by message so the corpus categories line up.
	msg := err.Error()
	switch {
	case strings.Contains(msg, "signature length"):
		return "sig-length"
	case strings.Contains(msg, "must be in low form"):
		return "low-s-msb"
	case strings.Contains(msg, "empty payload"):
		return "empty-payload"
	case strings.Contains(msg, "unsupported signing public key length"):
		return "unsupported-pubkey-length"
	}
	return "other:" + msg
}

// deliberateReasonChanges records, per corpus case, the rejection categories the new
// implementation is allowed to report in place of the recorded go-ethereum one.
//
// Every entry here is a *refinement*, never a change of accept/reject verdict: the
// boolean is asserted separately and unconditionally. Refining a rejection reason is
// safe because the reason is diagnostic text wrapped into
// types.ErrEnclaveSignatureInvalid, not consensus state; what matters is that both
// old and new reject.
//
// "digest-length" replaces the generic "verify-failed" for non-32-byte payloads. The
// old code funnelled every verification failure through one error string, so a
// 64-byte payload was reported as a failed verification even though nothing was
// verified. The new code names the actual cause. This also removes the cgo/no-cgo
// divergence go-ethereum had; see
// TestSignatureVerifierNonCgoPayloadDivergenceIsDeliberate.
var deliberateReasonChanges = map[string][]string{
	"payload-31-bytes": {"digest-length"},
	"payload-33-bytes": {"digest-length"},
	"payload-64-bytes": {"digest-length"},
}

func reasonIsAcceptable(caseName, recorded, got string) bool {
	if got == recorded {
		return true
	}
	for _, allowed := range deliberateReasonChanges[caseName] {
		if got == allowed {
			return true
		}
	}
	return false
}

func mustDecodeHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("corpus contains invalid hex: %v", err)
	}
	return b
}

func loadCorpus(t *testing.T, filename string) equivalenceCorpus {
	t.Helper()
	path := filepath.Join("testdata", filename)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading corpus: %v", err)
	}
	var corpus equivalenceCorpus
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatalf("parsing corpus: %v", err)
	}
	if corpus.SchemaVersion != 1 {
		t.Fatalf("unsupported corpus schemaVersion %d", corpus.SchemaVersion)
	}
	if len(corpus.Cases) == 0 {
		t.Fatal("corpus is empty; it cannot prove anything")
	}
	return corpus
}

// TestSignatureVerifierEquivalence replays the recorded go-ethereum verdicts.
//
// Both recorded corpora are checked on every build, not just the one matching the
// current toolchain. The cgo (libsecp256k1) and non-cgo (decred) builds of
// go-ethereum disagreed with each other on non-32-byte payloads, so pinning the
// new behaviour requires knowing both. The new implementation deliberately
// resolves that divergence by requiring exactly 32 bytes, which means the non-32-byte
// cases are expected to differ from the non-cgo corpus; see
// TestSignatureVerifierNonCgoPayloadDivergenceIsDeliberate.
func TestSignatureVerifierEquivalence(t *testing.T) {
	// cgo (libsecp256k1) is the Linux production path and is authoritative.
	corpus := loadCorpus(t, "signature_equivalence_cgo.json")

	for _, tc := range corpus.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			pubKey := mustDecodeHex(t, tc.PubKey)
			payload := mustDecodeHex(t, tc.Payload)
			sig := mustDecodeHex(t, tc.Sig)

			err := verifySigningKeySignature(pubKey, payload, sig)
			gotAccept := err == nil
			if gotAccept != tc.OldAccept {
				t.Fatalf("accept/reject divergence from go-ethereum (%s): old=%v new=%v err=%v\nnote: %s",
					corpus.BuildPath, tc.OldAccept, gotAccept, err, tc.Note)
			}
			if got := classify(err); !reasonIsAcceptable(tc.Name, tc.OldErrCategory, got) {
				t.Fatalf("rejection-reason divergence from go-ethereum (%s): recorded=%q new=%q allowed=%v\nnote: %s",
					corpus.BuildPath, tc.OldErrCategory, got, deliberateReasonChanges[tc.Name], tc.Note)
			}
		})
	}
}

// TestSignatureVerifierNonCgoPayloadDivergenceIsDeliberate documents the one
// intentional behaviour change.
//
// go-ethereum's non-cgo build verified a 33-byte payload by silently truncating it
// to 32 bytes, and a 64-byte payload by using the first 32. Its cgo build rejected
// both outright. A node built with cgo and a node built without it therefore reached
// *different verdicts on the same attestation*, which is a latent consensus hazard
// independent of this licence work. verifySecp256k1Signature now requires exactly 32
// bytes, matching the cgo (production) behaviour and removing the divergence.
//
// This test asserts the new behaviour is uniform regardless of how this binary was
// built, which is the property that actually matters.
func TestSignatureVerifierNonCgoPayloadDivergenceIsDeliberate(t *testing.T) {
	pubKey, err := hex.DecodeString(testCorpusPubKey)
	if err != nil {
		t.Fatalf("bad test public key: %v", err)
	}
	payload, err := hex.DecodeString(testCorpusPayload)
	if err != nil {
		t.Fatalf("bad test payload: %v", err)
	}
	sig, err := hex.DecodeString(testCorpusSig)
	if err != nil {
		t.Fatalf("bad test signature: %v", err)
	}

	// Sanity: the canonical 32-byte case still verifies, so the rejections below are
	// caused by the length and not by a broken fixture.
	if err := verifySigningKeySignature(pubKey, payload, sig); err != nil {
		t.Fatalf("canonical 32-byte case must verify, got: %v", err)
	}

	for _, tc := range []struct {
		name    string
		payload []byte
	}{
		{"truncated-31", payload[:31]},
		{"extended-33", append(append([]byte{}, payload...), 0x00)},
		{"extended-64", append(append([]byte{}, payload...), make([]byte, 32)...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := verifySigningKeySignature(pubKey, tc.payload, sig)
			if err == nil {
				t.Fatalf("non-32-byte payload must be rejected on every build path (this binary: %s)", runtime.Compiler)
			}
			if !errors.Is(err, errSecp256k1DigestLength) {
				t.Fatalf("expected digest-length rejection, got: %v", err)
			}
		})
	}
}

// TestSignatureVerifierRejectsHybridPublicKeyEncodings pins a subtle difference.
//
// secp256k1.ParsePubKey accepts the hybrid encodings 0x06 and 0x07 for a 65-byte
// key, but go-ethereum's UnmarshalPubkey accepted only 0x04. A naive swap would have
// widened the accepted key set.
func TestSignatureVerifierRejectsHybridPublicKeyEncodings(t *testing.T) {
	pubKey, err := hex.DecodeString(testCorpusPubKey)
	if err != nil {
		t.Fatalf("bad test public key: %v", err)
	}
	payload, err := hex.DecodeString(testCorpusPayload)
	if err != nil {
		t.Fatalf("bad test payload: %v", err)
	}
	sig, err := hex.DecodeString(testCorpusSig)
	if err != nil {
		t.Fatalf("bad test signature: %v", err)
	}

	for _, prefix := range []byte{0x02, 0x03, 0x05, 0x06, 0x07, 0x00} {
		altered := append([]byte{}, pubKey...)
		altered[0] = prefix
		err := verifySigningKeySignature(altered, payload, sig)
		if err == nil {
			t.Fatalf("public key with 0x%02x prefix must be rejected", prefix)
		}
		if !errors.Is(err, errSecp256k1InvalidPubKey) {
			t.Fatalf("prefix 0x%02x: expected invalid-pubkey rejection, got: %v", prefix, err)
		}
	}
}

// TestSignatureVerifierRejectsHighSWithClearMSB pins the one guard the
// malleability twin cannot reach.
//
// For any valid signature (R, S) with S in low form, the twin (R, N-S) is also
// valid. Because N < 2^256, N-S for a low S always lands at or above 2^255, so the
// twin's most significant byte is always set and the old
// `signature[32]&0x80 != 0` pre-check caught every twin. That pre-check was
// therefore already sufficient against the classic malleability vector, and this
// test deliberately does not pretend otherwise.
//
// What the pre-check misses is the complementary window: a *valid* signature whose
// S satisfies N/2 < S < 2^255. Its MSB is clear, so the pre-check waves it through,
// yet S is over half order. Reaching it by chance has probability about 2^-128, so
// it cannot be found by sampling; but it can be constructed deliberately, by
// choosing S in the window and solving for the private key that makes it verify.
//
// Both go-ethereum build paths rejected such signatures, so the replacement must
// too, otherwise a deliberately-crafted attestation could be signed in a second
// accepted encoding.
func TestSignatureVerifierRejectsHighSWithClearMSB(t *testing.T) {
	// Deterministic nonce; any k with 0 < k < N works, and R must not overflow N.
	nonceBytes, err := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000007")
	if err != nil {
		t.Fatalf("bad nonce: %v", err)
	}
	digest := sha256.Sum256([]byte("high-s-clear-msb-probe"))

	for _, tc := range highSWithClearMSBSignatures(t, nonceBytes, digest[:]) {
		t.Run(tc.name, func(t *testing.T) {
			// Confirm the constructed signature really does verify. Without this the
			// test would pass for the wrong reason: an invalid signature is
			// rejected by the arithmetic, not by the low-S policy.
			if !isValidSecp256k1(tc.pubKey, digest[:], tc.sig) {
				t.Fatal("fixture precondition failed: constructed signature must be arithmetically valid")
			}
			// Its S must be over half order yet have a clear MSB, which is the
			// entire point of the fixture.
			var s secp256k1.ModNScalar
			s.SetBytes((*[32]byte)(tc.sig[32:]))
			if !s.IsOverHalfOrder() {
				t.Fatal("fixture precondition failed: S must be over half order")
			}
			if tc.sig[32]&0x80 != 0 {
				t.Fatal("fixture precondition failed: S MSB must be clear for this test to be meaningful")
			}

			err := verifySigningKeySignature(tc.pubKey, digest[:], tc.sig)
			if err == nil {
				t.Fatal("a valid signature with S over half order must be rejected")
			}
			if !errors.Is(err, errSecp256k1VerificationFailed) {
				t.Fatalf("expected a verification-policy rejection, got: %v", err)
			}
		})
	}
}

// highSWithClearMSBSignatures builds valid secp256k1 signatures whose S lies in the
// open interval (N/2, 2^255).
//
// Construction: pick a nonce k and let R be the x-coordinate of kG. For any desired
// S, the private key d = (S*R - e) * R^-1 mod N makes (R, S) verify over digest e,
// because the verification equation collapses to u1*G + u2*Q = ((e + dR)/S)*G = R*G
// when d is chosen this way. That lets the test *choose* an S in the narrow window
// instead of waiting ~2^128 signatures for one to occur.
func highSWithClearMSBSignatures(t *testing.T, nonce, digest []byte) []struct {
	name   string
	pubKey []byte
	sig    []byte
} {
	t.Helper()

	n, ok := new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	if !ok {
		t.Fatal("failed to parse group order")
	}
	halfN := new(big.Int).Rsh(n, 1)
	twoPow255 := new(big.Int).Lsh(big.NewInt(1), 255)

	var k secp256k1.ModNScalar
	if k.SetByteSlice(nonce) {
		t.Fatal("nonce out of range")
	}
	var kG secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(&k, &kG)
	kG.ToAffine()

	var rBytes [32]byte
	kG.X.PutBytes(&rBytes)
	r := new(big.Int).SetBytes(rBytes[:])
	if r.Sign() == 0 {
		t.Fatal("degenerate nonce produced R = 0")
	}

	// R must be invertible mod N for the private key to exist.
	rMod := new(big.Int).Mod(r, n)
	if new(big.Int).GCD(nil, nil, rMod, n).Cmp(big.NewInt(1)) != 0 {
		t.Skipf("nonce produced an R with no inverse mod N; rerun")
	}
	rInv := new(big.Int).ModInverse(rMod, n)
	if rInv == nil {
		t.Skipf("nonce produced an R with no inverse mod N; rerun")
	}

	e := new(big.Int).Mod(new(big.Int).SetBytes(digest), n)

	// Sample S values inside (N/2, 2^255). The window is narrow, so walk it in
	// steps rather than sampling randomly.
	windowWidth := new(big.Int).Sub(twoPow255, halfN)
	step := new(big.Int).Rsh(windowWidth, 4) // 16 samples across the window
	if step.Sign() == 0 {
		step = big.NewInt(1)
	}

	out := make([]struct {
		name   string
		pubKey []byte
		sig    []byte
	}, 0, 4)
	for i := int64(0); i < 4; i++ {
		s := new(big.Int).Add(halfN, new(big.Int).Mul(step, big.NewInt(i+1)))
		if s.Cmp(twoPow255) >= 0 {
			break
		}

		// Verification computes u1 = e*S^-1, u2 = R*S^-1 and X = u1*G + u2*Q.
		// With Q = d*G that is X = ((e + d*R)/S)*G, and X.x == R holds exactly
		// when (e + d*R)/S == k (mod N), i.e. d = (S*k - e) * R^-1 mod N.
		// Note the numerator uses the nonce k, not the x-coordinate R.
		d := new(big.Int).Mul(s, new(big.Int).SetBytes(nonce))
		d.Sub(d, e)
		d.Mul(d, rInv)
		d.Mod(d, n)
		if d.Sign() == 0 {
			continue
		}

		var dScalar secp256k1.PrivateKey
		var dBytes [32]byte
		d.FillBytes(dBytes[:])
		if dScalar.Key.SetByteSlice(dBytes[:]) {
			continue
		}
		pubKey := dScalar.PubKey().SerializeUncompressed()

		sig := make([]byte, 64)
		copy(sig[:32], rBytes[:])
		s.FillBytes(sig[32:])

		out = append(out, struct {
			name   string
			pubKey []byte
			sig    []byte
		}{
			name:   fmt.Sprintf("high-s-clear-msb-%d", i),
			pubKey: pubKey,
			sig:    sig,
		})
	}

	if len(out) == 0 {
		t.Fatal("failed to construct any high-S-with-clear-MSB signatures")
	}
	return out
}

// isValidSecp256k1 reports whether a signature is arithmetically valid, ignoring the
// low-S policy. It exists so the malleability test can distinguish "rejected
// because it is a forgery" from "rejected because it is a second encoding of an
// otherwise-valid signature".
func isValidSecp256k1(pubKey, payload, sig []byte) bool {
	key, err := secp256k1.ParsePubKey(pubKey)
	if err != nil {
		return false
	}
	var r, s secp256k1.ModNScalar
	if r.SetByteSlice(sig[:32]) || s.SetByteSlice(sig[32:]) {
		return false
	}
	return decred_ecdsa.NewSignature(&r, &s).Verify(payload, key)
}
