package keeper

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	decred_ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// Rejection categories for secp256k1 attestation signatures.
//
// verifySecp256k1Signature historically delegated to
// github.com/ethereum/go-ethereum/crypto. That module is licensed GPL-3.0, which
// the supply-chain policy in scripts/supply-chain/go-module-policy.json denies, so
// the verification is now performed directly on top of
// github.com/decred/dcrd/dcrec/secp256k1/v4 (ISC), which this repository already
// depends on. The categories exist so the equivalence test can assert not just
// accept/reject parity but parity in *why* a signature was rejected, which is the
// property that actually matters for state-machine determinism: two validators
// reaching different verdicts on the same input would diverge the chain.
//
// The errors are intentionally unexported and never surfaced directly; callers wrap
// them into types.ErrEnclaveSignatureInvalid.
var (
	// errSecp256k1DigestLength rejects a payload that is not exactly 32 bytes.
	errSecp256k1DigestLength = errors.New("secp256k1 payload must be exactly 32 bytes")

	// errSecp256k1InvalidPubKey rejects a public key that is not a canonical
	// uncompressed encoding of a point on secp256k1.
	errSecp256k1InvalidPubKey = errors.New("invalid secp256k1 public key")

	// errSecp256k1VerificationFailed rejects a canonically-encoded signature that
	// does not verify against the public key. This covers R or S outside [1, N-1],
	// an S value that is not in low form, and a genuine verification failure.
	errSecp256k1VerificationFailed = errors.New("secp256k1 signature verification failed")
)

// secp256k1UncompressedFormat is the only accepted public key encoding prefix.
// secp256k1.ParsePubKey also accepts the hybrid encodings 0x06 and 0x07 for a
// 65-byte input, but go-ethereum's UnmarshalPubkey (via its btCurve.Unmarshal)
// accepted only 0x04, so hybrid encodings must stay rejected.
const secp256k1UncompressedFormat = 0x04

// secp256k1DigestSize is the required payload length in bytes.
const secp256k1DigestSize = 32

// verifySecp256k1Signature verifies an ECDSA signature over secp256k1 for a
// 65-byte uncompressed public key and a 64-byte R||S signature.
//
// It reproduces, case for case, the accept/reject behaviour of the
// go-ethereum crypto calls it replaces, so that swapping the implementation does
// not change which attestations are considered valid. The three rules that matter:
//
//   - The public key must be a canonical uncompressed (0x04) encoding of a point
//     on secp256k1 whose coordinates are both less than the field prime P.
//   - The payload must be exactly 32 bytes. go-ethereum's cgo path (libsecp256k1,
//     the Linux production build) rejected any other length outright, while its
//     non-cgo path silently zero-padded or truncated the digest to 32 bytes.
//     Requiring exactly 32 bytes removes that build-dependent divergence. Every
//     caller in this module passes a sha256.Sum256 digest, so the reachable
//     behaviour is unchanged.
//   - R and S must both be canonical scalars in [1, N-1] and S must be in low form,
//     which keeps signature malleability from being consensus-visible.
//
// x/enclave is consensus-critical, so signature_equivalence_test.go pins this
// function's behaviour against a recorded corpus of the pre-change verdicts.
func verifySecp256k1Signature(pubKey []byte, payload []byte, signature []byte) error {
	if len(payload) != secp256k1DigestSize {
		return fmt.Errorf("%w: expected %d, got %d", errSecp256k1DigestLength, secp256k1DigestSize, len(payload))
	}
	if pubKey[0] != secp256k1UncompressedFormat {
		return fmt.Errorf("%w: expected 0x%02x prefix, got 0x%02x", errSecp256k1InvalidPubKey, secp256k1UncompressedFormat, pubKey[0])
	}

	key, err := secp256k1.ParsePubKey(pubKey)
	if err != nil {
		return fmt.Errorf("%w: %w", errSecp256k1InvalidPubKey, err)
	}

	var r, s secp256k1.ModNScalar
	if r.SetByteSlice(signature[:32]) {
		// R >= group order N; not a canonical scalar.
		return errSecp256k1VerificationFailed
	}
	if s.SetByteSlice(signature[32:]) {
		// S >= group order N; not a canonical scalar.
		return errSecp256k1VerificationFailed
	}
	if s.IsOverHalfOrder() {
		// Malleable S. go-ethereum rejected these in both its cgo and non-cgo
		// paths; the explicit low-S pre-check in verifySigningKeySignature only
		// covers the MSB, so this closes the remaining window.
		return errSecp256k1VerificationFailed
	}

	if !decred_ecdsa.NewSignature(&r, &s).Verify(payload, key) {
		return errSecp256k1VerificationFailed
	}
	return nil
}

func verifySigningKeySignature(pubKey []byte, payload []byte, signature []byte) error {
	if len(payload) == 0 {
		return fmt.Errorf("empty payload")
	}

	switch len(pubKey) {
	case ed25519.PublicKeySize:
		if len(signature) != ed25519.SignatureSize {
			return fmt.Errorf("invalid ed25519 signature length: expected %d, got %d", ed25519.SignatureSize, len(signature))
		}
		if !ed25519.Verify(pubKey, payload, signature) {
			return fmt.Errorf("ed25519 signature verification failed")
		}
		return nil

	case Secp256k1UncompressedPubKeySize:
		if len(signature) != Secp256k1SignatureSize {
			return fmt.Errorf("invalid secp256k1 signature length: expected %d, got %d", Secp256k1SignatureSize, len(signature))
		}
		if signature[32]&0x80 != 0 {
			return fmt.Errorf("secp256k1 signature S-value must be in low form")
		}

		return verifySecp256k1Signature(pubKey, payload, signature)

	default:
		return fmt.Errorf(
			"unsupported signing public key length: %d (expected %d for Ed25519 or %d for secp256k1)",
			len(pubKey), ed25519.PublicKeySize, Secp256k1UncompressedPubKeySize,
		)
	}
}
