// Package jwt provides JWT signing and verification methods for the Cosmos SDK.
// It includes functionality for creating and verifying JWT tokens using the ES256K
// signing method.
package jwt

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/golang-jwt/jwt/v5"
)

type signer struct{}

var _ jwt.SigningMethod = (*signer)(nil)

var (
	SigningMethodES256K *signer
)

func (s *signer) Sign(signingString string, key interface{}) ([]byte, error) {
	switch key := key.(type) {
	case SignerI:
		res, _, err := key.SignByAddress(key.GetAddress(), []byte(signingString), signing.SignMode_SIGN_MODE_DIRECT)
		if err != nil {
			return nil, err
		}
		return res, nil
	default:
		return nil, fmt.Errorf("%w: ES256K sign expects cryptotypes.LedgerPrivKey", jwt.ErrInvalidKeyType)
	}
}

func (s *signer) Verify(signingString string, sig []byte, key interface{}) error {
	switch key := key.(type) {
	case VerifyI:
		if !key.Pubkey().VerifySignature([]byte(signingString), sig) {
			return jwt.ErrTokenSignatureInvalid
		}

		return nil
	default:
		return fmt.Errorf("%w: ES256K verify expects cryptotypes.PubKey", jwt.ErrInvalidKeyType)
	}
}

func (s *signer) Alg() string {
	return "ES256K"
}

func init() {
	SigningMethodES256K = &signer{}

	jwt.RegisterSigningMethod(SigningMethodES256K.Alg(), func() jwt.SigningMethod {
		return SigningMethodES256K
	})
}
