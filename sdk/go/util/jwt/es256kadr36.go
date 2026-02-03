// Package jwt provides JWT signing and verification methods for the Cosmos SDK.
// It includes functionality for creating and verifying JWT tokens using the ES256K
// signing method.
package jwt

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/golang-jwt/jwt/v5"

	offchain "github.com/virtengine/virtengine/sdk/go/node/types/offchain/sign"
)

// signerADR36 this contraption is to get JWT sign work with web base wallets like Keplr/Leap
// `signArbitrary` there does not sign provided data directly for chain security reasons,
// rather wraps data in fake tx and signs it. Such an approach does not allow to validate
// JWT token on the provider side with purely ES256K signature
type signerADR36 struct {
	cdc *codec.LegacyAmino
}

var _ jwt.SigningMethod = (*signerADR36)(nil)

var (
	SigningMethodES256KADR36 *signerADR36
)

func (s *signerADR36) Sign(signingString string, key interface{}) ([]byte, error) {
	switch key := key.(type) {
	case SignerI:
		//nolint:staticcheck // SA1019: using deprecated StdSignBytes for ADR-36 compatibility
		signBytes := offchain.StdSignBytes(s.cdc, "", 0, 0, 0, legacytx.StdFee{}, []sdk.Msg{
			&offchain.MsgSignData{
				Signer: key.GetAddress().String(),
				Data:   []byte(signingString),
			},
		}, "")

		res, _, err := key.SignByAddress(key.GetAddress(), signBytes, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
		if err != nil {
			return nil, err
		}
		return res, nil
	default:
		return nil, fmt.Errorf("%w: ES256KADR36 sign expects cryptotypes.LedgerPrivKey", jwt.ErrInvalidKeyType)
	}
}

func (s *signerADR36) Verify(signingString string, sig []byte, key interface{}) error {
	switch key := key.(type) {
	case VerifyI:
		//nolint:staticcheck // SA1019: using deprecated StdSignBytes for ADR-36 compatibility
		signBytes := offchain.StdSignBytes(s.cdc, "", 0, 0, 0, legacytx.StdFee{}, []sdk.Msg{
			&offchain.MsgSignData{
				Signer: key.GetAddress().String(),
				Data:   []byte(signingString),
			},
		}, "")

		if !key.Pubkey().VerifySignature(signBytes, sig) {
			return jwt.ErrTokenSignatureInvalid
		}

		return nil
	default:
		return fmt.Errorf("%w: ES256KADR36 verify expects cryptotypes.PubKey", jwt.ErrInvalidKeyType)
	}
}

func (s *signerADR36) Alg() string {
	return "ES256KADR36"
}

func init() {
	cdc := codec.NewLegacyAmino()
	cdc.RegisterConcrete(&offchain.MsgSignData{}, "sign/MsgSignData", nil)

	SigningMethodES256KADR36 = &signerADR36{
		cdc: cdc,
	}

	jwt.RegisterSigningMethod(SigningMethodES256KADR36.Alg(), func() jwt.SigningMethod {
		return SigningMethodES256KADR36
	})
}
