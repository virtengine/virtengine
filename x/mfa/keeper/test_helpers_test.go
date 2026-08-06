package keeper_test

import (
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
