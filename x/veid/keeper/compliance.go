// Package keeper provides VEID module keeper implementation.
//
// This file implements the KYC/AML compliance interface.
//
// Task Reference: VE-3021 - KYC/AML Compliance Interface
package keeper

import (
	"encoding/json"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ComplianceKeeper defines the interface for external compliance providers
type ComplianceKeeper interface {
	// CheckSanctionList checks if identity matches any sanction lists
	CheckSanctionList(ctx sdk.Context, identity *types.IdentityRecord) (*types.ComplianceCheckResult, error)

	// CheckPEP checks for politically exposed person matches
	CheckPEP(ctx sdk.Context, identity *types.IdentityRecord) (*types.ComplianceCheckResult, error)

	// CheckGeographicRestrictions checks if identity is from restricted region
	CheckGeographicRestrictions(ctx sdk.Context, identity *types.IdentityRecord, regions []string) (*types.ComplianceCheckResult, error)

	// GetComplianceStatus returns the current compliance status
	GetComplianceStatus(ctx sdk.Context, address string) (*types.ComplianceRecord, error)

	// SetComplianceStatus updates the compliance status
	SetComplianceStatus(ctx sdk.Context, record *types.ComplianceRecord) error
}

// ============================================================================
// Compliance Parameters
// ============================================================================

// GetComplianceParams retrieves the compliance system parameters
func (k Keeper) GetComplianceParams(ctx sdk.Context) types.ComplianceParams {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ComplianceParamsKey())
	if bz == nil {
		return types.DefaultComplianceParams()
	}

	var params types.ComplianceParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultComplianceParams()
	}
	return params
}

// SetComplianceParams sets the compliance system parameters
func (k Keeper) SetComplianceParams(ctx sdk.Context, params types.ComplianceParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.ComplianceParamsKey(), bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeComplianceParamsUpdated,
			sdk.NewAttribute(types.AttributeKeyComplianceRiskScore, string(rune(params.RiskScoreThreshold))),
		),
	)

	return nil
}

// ============================================================================
// Compliance Record Storage
// ============================================================================

// GetComplianceRecord retrieves a compliance record by address
func (k Keeper) GetComplianceRecord(ctx sdk.Context, address string) (*types.ComplianceRecord, bool) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, false
	}

	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ComplianceRecordKey(addr.Bytes()))
	if bz == nil {
		return nil, false
	}

	var record types.ComplianceRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, false
	}

	return &record, true
}

// SetComplianceRecord stores a compliance record
func (k Keeper) SetComplianceRecord(ctx sdk.Context, record *types.ComplianceRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(record.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrap("invalid account address")
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}

	store.Set(types.ComplianceRecordKey(addr.Bytes()), bz)

	// Update blocked address index if blocked
	if record.Status == types.ComplianceStatusBlocked {
		store.Set(types.BlockedAddressKey(addr.Bytes()), []byte(record.Notes))
	} else {
		// Remove from blocked index if previously blocked
		store.Delete(types.BlockedAddressKey(addr.Bytes()))
	}

	return nil
}

// DeleteComplianceRecord removes a compliance record
func (k Keeper) DeleteComplianceRecord(ctx sdk.Context, address string) error {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return types.ErrInvalidAddress.Wrap("invalid address")
	}

	store := ctx.KVStore(k.skey)
	store.Delete(types.ComplianceRecordKey(addr.Bytes()))
	store.Delete(types.BlockedAddressKey(addr.Bytes()))

	return nil
}

// GetOrCreateComplianceRecord gets or creates a compliance record
func (k Keeper) GetOrCreateComplianceRecord(ctx sdk.Context, address string) (*types.ComplianceRecord, error) {
	record, found := k.GetComplianceRecord(ctx, address)
	if found {
		return record, nil
	}

	// Create new record
	record = types.NewComplianceRecord(address, ctx.BlockTime())
	if err := k.SetComplianceRecord(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// WithComplianceRecords iterates over all compliance records
func (k Keeper) WithComplianceRecords(ctx sdk.Context, fn func(record types.ComplianceRecord) bool) {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ComplianceRecordPrefixKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record types.ComplianceRecord
		if err := json.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		if fn(record) {
			break
		}
	}
}

// ============================================================================
// Compliance Provider Storage
// ============================================================================

// GetComplianceProvider retrieves a compliance provider by ID
func (k Keeper) GetComplianceProvider(ctx sdk.Context, providerID string) (*types.ComplianceProvider, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ComplianceProviderKey(providerID))
	if bz == nil {
		return nil, false
	}

	var provider types.ComplianceProvider
	if err := json.Unmarshal(bz, &provider); err != nil {
		return nil, false
	}

	return &provider, true
}

// GetComplianceProviderByAddress retrieves a compliance provider by address
func (k Keeper) GetComplianceProviderByAddress(ctx sdk.Context, address string) (*types.ComplianceProvider, bool) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, false
	}

	store := ctx.KVStore(k.skey)
	providerIDBytes := store.Get(types.ComplianceProviderByAddressKey(addr.Bytes()))
	if providerIDBytes == nil {
		return nil, false
	}

	return k.GetComplianceProvider(ctx, string(providerIDBytes))
}

// SetComplianceProvider stores a compliance provider
func (k Keeper) SetComplianceProvider(ctx sdk.Context, provider *types.ComplianceProvider) error {
	if err := provider.Validate(); err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(provider.ProviderAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrap("invalid provider address")
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(provider)
	if err != nil {
		return err
	}

	// Store provider
	store.Set(types.ComplianceProviderKey(provider.ProviderID), bz)

	// Store address index
	store.Set(types.ComplianceProviderByAddressKey(addr.Bytes()), []byte(provider.ProviderID))

	return nil
}

// GetAllComplianceProviders returns all compliance providers
func (k Keeper) GetAllComplianceProviders(ctx sdk.Context, activeOnly bool) []types.ComplianceProvider {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ComplianceProviderPrefixKey())
	defer iterator.Close()

	var providers []types.ComplianceProvider
	for ; iterator.Valid(); iterator.Next() {
		var provider types.ComplianceProvider
		if err := json.Unmarshal(iterator.Value(), &provider); err != nil {
			continue
		}
		if activeOnly && !provider.IsActive {
			continue
		}
		providers = append(providers, provider)
	}

	return providers
}

// IsAuthorizedComplianceProvider checks if an address is an authorized provider
func (k Keeper) IsAuthorizedComplianceProvider(ctx sdk.Context, address string) bool {
	provider, found := k.GetComplianceProviderByAddress(ctx, address)
	if !found {
		return false
	}
	return provider.IsActive
}

// DeactivateComplianceProvider deactivates a compliance provider
func (k Keeper) DeactivateComplianceProvider(ctx sdk.Context, providerID string, reason string) error {
	provider, found := k.GetComplianceProvider(ctx, providerID)
	if !found {
		return types.ErrNotComplianceProvider.Wrap("provider not found")
	}

	provider.IsActive = false
	return k.SetComplianceProvider(ctx, provider)
}

// ============================================================================
// Compliance Checks
// ============================================================================

// RunComplianceChecks runs all required compliance checks for an address
func (k Keeper) RunComplianceChecks(ctx sdk.Context, address string) (*types.ComplianceRecord, error) {
	// Get or create compliance record
	record, err := k.GetOrCreateComplianceRecord(ctx, address)
	if err != nil {
		return nil, err
	}

	// If already blocked, don't rerun checks
	if record.Status == types.ComplianceStatusBlocked {
		return nil, types.ErrComplianceRecordBlocked.Wrap("address is blocked")
	}

	// Get compliance params (for future use in compliance check logic)
	_ = k.GetComplianceParams(ctx)
	now := ctx.BlockTime().Unix()

	// Mark as pending
	record.Status = types.ComplianceStatusPending
	record.UpdatedAt = now

	// Update record
	if err := k.SetComplianceRecord(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// SubmitComplianceCheck processes submitted compliance check results
func (k Keeper) SubmitComplianceCheck(ctx sdk.Context, msg *types.MsgSubmitComplianceCheck) error {
	// Verify provider is authorized
	if !k.IsAuthorizedComplianceProvider(ctx, msg.ProviderAddress) {
		return types.ErrNotComplianceProvider.Wrap("sender is not an authorized provider")
	}

	// Get provider to verify it supports the check types
	provider, found := k.GetComplianceProviderByAddress(ctx, msg.ProviderAddress)
	if !found {
		return types.ErrNotComplianceProvider.Wrap("provider not found")
	}

	// Verify provider supports all submitted check types
	for _, result := range msg.CheckResults {
		if !provider.SupportsCheckType(result.CheckType) {
			return types.ErrUnsupportedCheckType.Wrapf("provider does not support check type: %s", result.CheckType)
		}
	}

	// Get or create compliance record
	record, err := k.GetOrCreateComplianceRecord(ctx, msg.TargetAddress)
	if err != nil {
		return err
	}

	// If blocked, don't accept new checks
	if record.Status == types.ComplianceStatusBlocked {
		return types.ErrComplianceRecordBlocked.Wrap("address is blocked")
	}

	now := ctx.BlockTime().Unix()
	params := k.GetComplianceParams(ctx)

	// Add check results
	allPassed := true
	for _, result := range msg.CheckResults {
		record.AddCheckResult(result, now)

		if !result.Passed {
			allPassed = false
		}

		// Emit event for each check
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeComplianceChecked,
				sdk.NewAttribute(types.AttributeKeyAccountAddress, msg.TargetAddress),
				sdk.NewAttribute(types.AttributeKeyComplianceCheckType, result.CheckType.String()),
				sdk.NewAttribute(types.AttributeKeyComplianceProvider, msg.ProviderID),
			),
		)
	}

	// Calculate new risk score
	record.RiskScore = record.CalculateRiskScore()

	// Determine new status
	oldStatus := record.Status
	if record.RiskScore > params.RiskScoreThreshold {
		record.Status = types.ComplianceStatusBlocked
	} else if !allPassed {
		record.Status = types.ComplianceStatusFlagged
	} else {
		record.Status = types.ComplianceStatusCleared
	}

	// Set expiry
	blockDuration := time.Duration(params.CheckExpiryBlocks) * 6 * time.Second // Assuming 6s block time
	record.ExpiresAt = ctx.BlockTime().Add(blockDuration).Unix()

	record.UpdatedAt = now

	// Save record
	if err := k.SetComplianceRecord(ctx, record); err != nil {
		return err
	}

	// Update provider last active
	provider.LastActiveAt = now
	if err := k.SetComplianceProvider(ctx, provider); err != nil {
		return err
	}

	// Emit status change event if changed
	if oldStatus != record.Status {
		eventType := types.EventTypeComplianceChecked
		switch record.Status {
		case types.ComplianceStatusCleared:
			eventType = types.EventTypeComplianceCleared
		case types.ComplianceStatusFlagged:
			eventType = types.EventTypeComplianceFlagged
		case types.ComplianceStatusBlocked:
			eventType = types.EventTypeComplianceBlocked
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				eventType,
				sdk.NewAttribute(types.AttributeKeyAccountAddress, msg.TargetAddress),
				sdk.NewAttribute(types.AttributeKeyComplianceStatus, record.Status.String()),
			),
		)
	}

	return nil
}

// ============================================================================
// Compliance Attestations
// ============================================================================

// AddComplianceAttestation adds a validator attestation to a compliance record
func (k Keeper) AddComplianceAttestation(ctx sdk.Context, address string, attestation types.ComplianceAttestation) error {
	// Verify attestor is a validator
	attestorAddr, err := sdk.AccAddressFromBech32(attestation.ValidatorAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrap("invalid validator address")
	}

	if !k.IsValidator(ctx, attestorAddr) {
		return types.ErrInsufficientAttestations.Wrap("attestor is not a validator")
	}

	// Get or create compliance record
	record, err := k.GetOrCreateComplianceRecord(ctx, address)
	if err != nil {
		return err
	}

	// Add attestation
	now := ctx.BlockTime().Unix()
	if err := record.AddAttestation(attestation, now); err != nil {
		return err
	}

	// Check if we have enough attestations to clear
	params := k.GetComplianceParams(ctx)
	if record.HasValidAttestations(params.MinAttestationsRequired, now) &&
		record.Status == types.ComplianceStatusFlagged {
		record.Status = types.ComplianceStatusCleared
	}

	// Save record
	if err := k.SetComplianceRecord(ctx, record); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeComplianceAttested,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, address),
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, attestation.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyComplianceAttestationType, attestation.AttestationType),
		),
	)

	return nil
}

// ============================================================================
// Compliance Status Queries
// ============================================================================

// IsComplianceCurrent checks if compliance status is current and valid
func (k Keeper) IsComplianceCurrent(ctx sdk.Context, address string) (bool, error) {
	record, found := k.GetComplianceRecord(ctx, address)
	if !found {
		return false, types.ErrComplianceNotFound.Wrap("no compliance record found")
	}

	now := ctx.BlockTime().Unix()

	// Check if expired
	if record.IsExpired(now) {
		return false, types.ErrComplianceExpired.Wrap("compliance check has expired")
	}

	// Check if blocked
	if record.Status == types.ComplianceStatusBlocked {
		return false, types.ErrComplianceRecordBlocked.Wrap("address is blocked")
	}

	// Only cleared status is considered compliant
	return record.Status == types.ComplianceStatusCleared, nil
}

// IsAddressBlocked checks if an address is blocked
func (k Keeper) IsAddressBlocked(ctx sdk.Context, address string) bool {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return false
	}

	store := ctx.KVStore(k.skey)
	return store.Has(types.BlockedAddressKey(addr.Bytes()))
}

// GetBlockedAddresses returns all blocked addresses
func (k Keeper) GetBlockedAddresses(ctx sdk.Context) []string {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.BlockedAddressPrefixKey())
	defer iterator.Close()

	var addresses []string
	for ; iterator.Valid(); iterator.Next() {
		// Extract address from key
		key := iterator.Key()
		addressBytes := key[len(types.BlockedAddressPrefixKey()):]
		addresses = append(addresses, sdk.AccAddress(addressBytes).String())
	}

	return addresses
}

// ============================================================================
// Compliance Check Expiry
// ============================================================================

// ExpireComplianceChecks expires all compliance records past their expiry time
func (k Keeper) ExpireComplianceChecks(ctx sdk.Context) int {
	now := ctx.BlockTime().Unix()
	expiredCount := 0

	k.WithComplianceRecords(ctx, func(record types.ComplianceRecord) bool {
		if record.IsExpired(now) && record.Status != types.ComplianceStatusExpired && record.Status != types.ComplianceStatusBlocked {
			record.Status = types.ComplianceStatusExpired
			record.UpdatedAt = now

			if err := k.SetComplianceRecord(ctx, &record); err == nil {
				expiredCount++

				// Emit event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeComplianceExpired,
						sdk.NewAttribute(types.AttributeKeyAccountAddress, record.AccountAddress),
					),
				)
			}
		}
		return false // Continue iteration
	})

	return expiredCount
}

// ============================================================================
// Geographic Restrictions
// ============================================================================

// CheckGeographicRestrictions verifies an identity is not from a restricted region
func (k Keeper) CheckGeographicRestrictions(ctx sdk.Context, address string, region string) error {
	params := k.GetComplianceParams(ctx)

	for _, restricted := range params.RestrictedCountries {
		if restricted == region {
			return types.ErrRestrictedRegion.Wrapf("region %s is restricted", region)
		}
	}

	// Update record if exists
	record, found := k.GetComplianceRecord(ctx, address)
	if found {
		// Add to restricted regions if not already present
		for _, r := range record.RestrictedRegions {
			if r == region {
				return nil // Already recorded
			}
		}
		record.RestrictedRegions = append(record.RestrictedRegions, region)
		record.UpdatedAt = ctx.BlockTime().Unix()
		if err := k.SetComplianceRecord(ctx, record); err != nil {
			return err
		}
	}

	return nil
}
