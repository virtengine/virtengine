package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/mfa/keeper"
	"github.com/virtengine/virtengine/x/mfa/types"
)

func enrollActiveFactor(k keeper.Keeper, ctx sdk.Context, address sdk.AccAddress, factorType types.FactorType, factorID string) error {
	publicIdentifier := []byte("factor-key")
	if factorType == types.FactorTypeVEID {
		publicIdentifier = nil
	}

	enrollment := &types.FactorEnrollment{
		AccountAddress:   address.String(),
		FactorType:       factorType,
		FactorID:         factorID,
		PublicIdentifier: publicIdentifier,
		Status:           types.EnrollmentStatusActive,
		EnrolledAt:       ctx.BlockTime().Unix(),
	}
	if factorType == types.FactorTypeVEID {
		enrollment.Metadata = &types.FactorMetadata{VEIDThreshold: 50}
	}
	if factorType == types.FactorTypeFIDO2 {
		enrollment.Metadata = &types.FactorMetadata{
			FIDO2Info: &types.FIDO2CredentialInfo{
				CredentialID: []byte(factorID),
				PublicKey:    []byte("public-key"),
			},
		}
	}

	return k.EnrollFactor(ctx, enrollment)
}

func requiredTxConfig(txType types.SensitiveTransactionType, factors ...types.FactorType) *types.SensitiveTxConfig {
	return &types.SensitiveTxConfig{
		TransactionType: txType,
		Enabled:         true,
		RequiredFactorCombinations: []types.FactorCombination{
			{Factors: factors},
		},
		Description: txType.String() + " requires MFA",
	}
}

func validSession(address sdk.AccAddress, txType types.SensitiveTransactionType, now time.Time, factors ...types.FactorType) *types.AuthorizationSession {
	return &types.AuthorizationSession{
		AccountAddress:  address.String(),
		TransactionType: txType,
		VerifiedFactors: factors,
		CreatedAt:       now.Unix(),
		ExpiresAt:       now.Add(time.Hour).Unix(),
	}
}
