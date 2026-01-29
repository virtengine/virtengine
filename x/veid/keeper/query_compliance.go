// Package keeper provides VEID module keeper implementation.
//
// This file implements compliance query handlers.
//
// Task Reference: VE-3021 - KYC/AML Compliance Interface
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// QueryComplianceStatus returns the compliance status for an account
func (k Keeper) QueryComplianceStatus(ctx sdk.Context, req *types.QueryComplianceStatusRequest) (*types.QueryComplianceStatusResponse, error) {
	if req.Address == "" {
		return nil, types.ErrInvalidAddress.Wrap("address cannot be empty")
	}

	// Validate address format
	if _, err := sdk.AccAddressFromBech32(req.Address); err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid address format")
	}

	record, found := k.GetComplianceRecord(ctx, req.Address)
	if !found {
		return &types.QueryComplianceStatusResponse{
			Record: nil,
			Found:  false,
		}, nil
	}

	// Check if expired and update status if needed
	now := ctx.BlockTime().Unix()
	if record.IsExpired(now) && record.Status != types.ComplianceStatusExpired && record.Status != types.ComplianceStatusBlocked {
		record.Status = types.ComplianceStatusExpired
		record.UpdatedAt = now
		// Update in store
		_ = k.SetComplianceRecord(ctx, record)
	}

	return &types.QueryComplianceStatusResponse{
		Record: record,
		Found:  true,
	}, nil
}

// QueryComplianceParams returns the compliance configuration
func (k Keeper) QueryComplianceParams(ctx sdk.Context, req *types.QueryComplianceParamsRequest) (*types.QueryComplianceParamsResponse, error) {
	params := k.GetComplianceParams(ctx)
	return &types.QueryComplianceParamsResponse{
		Params: params,
	}, nil
}

// QueryComplianceProvider returns a specific compliance provider
func (k Keeper) QueryComplianceProvider(ctx sdk.Context, req *types.QueryComplianceProviderRequest) (*types.QueryComplianceProviderResponse, error) {
	if req.ProviderID == "" {
		return nil, types.ErrNotComplianceProvider.Wrap("provider_id cannot be empty")
	}

	provider, found := k.GetComplianceProvider(ctx, req.ProviderID)
	if !found {
		return &types.QueryComplianceProviderResponse{
			Provider: nil,
			Found:    false,
		}, nil
	}

	return &types.QueryComplianceProviderResponse{
		Provider: provider,
		Found:    true,
	}, nil
}

// QueryComplianceProviders returns all compliance providers
func (k Keeper) QueryComplianceProviders(ctx sdk.Context, req *types.QueryComplianceProvidersRequest) (*types.QueryComplianceProvidersResponse, error) {
	providers := k.GetAllComplianceProviders(ctx, req.ActiveOnly)

	return &types.QueryComplianceProvidersResponse{
		Providers: providers,
	}, nil
}

// QueryBlockedAddresses returns all blocked addresses
func (k Keeper) QueryBlockedAddresses(ctx sdk.Context) ([]string, error) {
	return k.GetBlockedAddresses(ctx), nil
}

// QueryIsCompliant checks if an address is currently compliant
func (k Keeper) QueryIsCompliant(ctx sdk.Context, address string) (bool, string, error) {
	if address == "" {
		return false, "", types.ErrInvalidAddress.Wrap("address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(address); err != nil {
		return false, "", types.ErrInvalidAddress.Wrap("invalid address format")
	}

	record, found := k.GetComplianceRecord(ctx, address)
	if !found {
		return false, "no compliance record", nil
	}

	now := ctx.BlockTime().Unix()

	// Check expiry
	if record.IsExpired(now) {
		return false, "compliance check expired", nil
	}

	// Check status
	switch record.Status {
	case types.ComplianceStatusCleared:
		return true, "cleared", nil
	case types.ComplianceStatusBlocked:
		return false, "blocked", nil
	case types.ComplianceStatusFlagged:
		return false, "flagged for review", nil
	case types.ComplianceStatusPending:
		return false, "check pending", nil
	case types.ComplianceStatusExpired:
		return false, "expired", nil
	default:
		return false, "unknown status", nil
	}
}
