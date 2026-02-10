package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// CheckProviderEligibility enforces VEID and compliance gating for providers.
func (k Keeper) CheckProviderEligibility(ctx sdk.Context, providerAddr sdk.AccAddress) error {
	if k.veidKeeper == nil {
		return nil
	}

	score, scoreFound := k.veidKeeper.GetIdentityScore(ctx, providerAddr)
	status, statusFound := k.veidKeeper.GetIdentityStatus(ctx, providerAddr)
	if !scoreFound || !statusFound {
		return marketplace.ErrProviderIdentityMissing
	}

	if status != string(veidtypes.AccountStatusVerified) {
		return marketplace.ErrProviderIdentityNotVerified
	}

	minScore := veidtypes.GetMinimumScoreForTier(veidtypes.TierStandard)
	if score < minScore {
		return marketplace.ErrProviderInsufficientScore.Wrapf("score %d below minimum %d", score, minScore)
	}

	complianceCleared, complianceFound := k.veidKeeper.IsComplianceCleared(ctx, providerAddr)
	if !complianceFound || !complianceCleared {
		return marketplace.ErrProviderComplianceIncomplete
	}

	return nil
}
