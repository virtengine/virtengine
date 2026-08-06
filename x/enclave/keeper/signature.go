package keeper

import (
	"crypto/ed25519"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

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

		pub, err := crypto.UnmarshalPubkey(pubKey)
		if err != nil {
			return fmt.Errorf("invalid secp256k1 public key: %w", err)
		}
		uncompressed := crypto.FromECDSAPub(pub)
		if !crypto.VerifySignature(uncompressed, payload, signature) {
			return fmt.Errorf("secp256k1 signature verification failed")
		}
		return nil

	default:
		return fmt.Errorf(
			"unsupported signing public key length: %d (expected %d for Ed25519 or %d for secp256k1)",
			len(pubKey), ed25519.PublicKeySize, Secp256k1UncompressedPubKeySize,
		)
	}
}
