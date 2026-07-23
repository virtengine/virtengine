package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	providerkeeper "github.com/virtengine/virtengine/x/provider/keeper"
)

// reservationEligibilityAdapter deliberately certifies only committed provider
// registration. Optional benchmark, attestation, and collateral profiles remain
// fail-closed until their authoritative Task 90B/87A policy readers are wired.
type reservationEligibilityAdapter struct{ provider providerkeeper.IKeeper }

func newReservationEligibilityAdapter(provider providerkeeper.IKeeper) reservationEligibilityAdapter {
	return reservationEligibilityAdapter{provider: provider}
}

func (a reservationEligibilityAdapter) IsProviderEligible(ctx sdk.Context, provider sdk.AccAddress) bool {
	return a.provider != nil && a.provider.IsProvider(ctx, provider)
}

func (reservationEligibilityAdapter) HasCurrentBenchmark(sdk.Context, string) bool   { return false }
func (reservationEligibilityAdapter) HasCurrentAttestation(sdk.Context, string) bool { return false }
func (reservationEligibilityAdapter) HasSufficientCollateral(sdk.Context, string, string) bool {
	return false
}
