// Package keeper provides VEID module keeper implementation.
//
// This file implements geographic restriction rules for VEID compliance.
//
// Task Reference: VE-3032 - Add Geographic Restriction Rules for VEID
package keeper

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// boolTrue is the string representation of boolean true for event attributes
const boolTrue = "true"

// ============================================================================
// Geographic Restriction Parameters
// ============================================================================

// GetGeoRestrictionParams retrieves the geo restriction parameters
func (k Keeper) GetGeoRestrictionParams(ctx sdk.Context) types.GeoRestrictionParams {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GeoRestrictionParamsKey())
	if bz == nil {
		return types.DefaultGeoRestrictionParams()
	}

	var params types.GeoRestrictionParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultGeoRestrictionParams()
	}
	return params
}

// SetGeoRestrictionParams sets the geo restriction parameters
func (k Keeper) SetGeoRestrictionParams(ctx sdk.Context, params types.GeoRestrictionParams) error {
	if err := params.Validate(); err != nil {
		return types.ErrGeoParamsInvalid.Wrap(err.Error())
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.GeoRestrictionParamsKey(), bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGeoRestrictionParamsUpdated,
			sdk.NewAttribute(types.AttributeKeyGeoEnabled, boolToString(params.Enabled)),
		),
	)

	return nil
}

// ============================================================================
// Geographic Policy CRUD Operations
// ============================================================================

// CreateGeoPolicy creates a new geographic restriction policy
func (k Keeper) CreateGeoPolicy(ctx sdk.Context, policy *types.GeoRestrictionPolicy) error {
	// Check if geo restrictions are enabled
	params := k.GetGeoRestrictionParams(ctx)
	if !params.Enabled {
		return types.ErrGeoRestrictionDisabled
	}

	// Validate policy
	if err := policy.Validate(); err != nil {
		return err
	}

	// Check if policy already exists
	if _, found := k.GetGeoPolicy(ctx, policy.PolicyID); found {
		return types.ErrGeoPolicyAlreadyExists.Wrapf("policy_id: %s", policy.PolicyID)
	}

	// Check max policies limit
	count := k.GetGeoPolicyCount(ctx)
	if count >= types.MaxGeoRestrictionPolicies {
		return types.ErrMaxGeoPoliciesExceeded.Wrapf("limit: %d", types.MaxGeoRestrictionPolicies)
	}

	// Store policy
	if err := k.setGeoPolicy(ctx, policy); err != nil {
		return err
	}

	// Update blocked country index
	k.updateBlockedCountryIndex(ctx, policy, nil)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGeoPolicyCreated,
			sdk.NewAttribute(types.AttributeKeyPolicyID, policy.PolicyID),
			sdk.NewAttribute(types.AttributeKeyPolicyName, policy.Name),
			sdk.NewAttribute(types.AttributeKeyEnforcementLevel, policy.EnforcementLevel.String()),
		),
	)

	return nil
}

// UpdateGeoPolicy updates an existing geographic restriction policy
func (k Keeper) UpdateGeoPolicy(ctx sdk.Context, policy *types.GeoRestrictionPolicy) error {
	// Validate policy
	if err := policy.Validate(); err != nil {
		return err
	}

	// Check if policy exists
	oldPolicy, found := k.GetGeoPolicy(ctx, policy.PolicyID)
	if !found {
		return types.ErrGeoPolicyNotFound.Wrapf("policy_id: %s", policy.PolicyID)
	}

	// Update timestamp
	policy.UpdatedAt = ctx.BlockTime()

	// Store updated policy
	if err := k.setGeoPolicy(ctx, policy); err != nil {
		return err
	}

	// Update blocked country index (remove old, add new)
	k.updateBlockedCountryIndex(ctx, policy, oldPolicy)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGeoPolicyUpdated,
			sdk.NewAttribute(types.AttributeKeyPolicyID, policy.PolicyID),
			sdk.NewAttribute(types.AttributeKeyPolicyName, policy.Name),
			sdk.NewAttribute(types.AttributeKeyPolicyStatus, policy.Status.String()),
		),
	)

	return nil
}

// DeleteGeoPolicy removes a geographic restriction policy
func (k Keeper) DeleteGeoPolicy(ctx sdk.Context, policyID string) error {
	policy, found := k.GetGeoPolicy(ctx, policyID)
	if !found {
		return types.ErrGeoPolicyNotFound.Wrapf("policy_id: %s", policyID)
	}

	store := ctx.KVStore(k.skey)

	// Remove from main store
	store.Delete(types.GeoPolicyKey(policyID))

	// Remove from priority index
	store.Delete(types.GeoPolicyByPriorityKey(policy.Priority, policyID))

	// Update blocked country index (remove only)
	k.updateBlockedCountryIndex(ctx, nil, policy)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGeoPolicyDeleted,
			sdk.NewAttribute(types.AttributeKeyPolicyID, policyID),
		),
	)

	return nil
}

// GetGeoPolicy retrieves a geographic restriction policy by ID
func (k Keeper) GetGeoPolicy(ctx sdk.Context, policyID string) (*types.GeoRestrictionPolicy, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GeoPolicyKey(policyID))
	if bz == nil {
		return nil, false
	}

	var policy types.GeoRestrictionPolicy
	if err := json.Unmarshal(bz, &policy); err != nil {
		return nil, false
	}

	return &policy, true
}

// setGeoPolicy stores a geo restriction policy
func (k Keeper) setGeoPolicy(ctx sdk.Context, policy *types.GeoRestrictionPolicy) error {
	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	// Store main policy
	store.Set(types.GeoPolicyKey(policy.PolicyID), bz)

	// Store priority index
	store.Set(types.GeoPolicyByPriorityKey(policy.Priority, policy.PolicyID), []byte{0x01})

	return nil
}

// GetGeoPolicyCount returns the total number of geo policies
func (k Keeper) GetGeoPolicyCount(ctx sdk.Context) int {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.GeoPolicyPrefixKey())
	defer iterator.Close()

	count := 0
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return count
}

// GetAllGeoPolicies returns all geographic restriction policies
func (k Keeper) GetAllGeoPolicies(ctx sdk.Context, activeOnly bool) []types.GeoRestrictionPolicy {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.GeoPolicyPrefixKey())
	defer iterator.Close()

	var policies []types.GeoRestrictionPolicy
	for ; iterator.Valid(); iterator.Next() {
		var policy types.GeoRestrictionPolicy
		if err := json.Unmarshal(iterator.Value(), &policy); err != nil {
			continue
		}
		if activeOnly && !policy.IsActive() {
			continue
		}
		policies = append(policies, policy)
	}

	// Sort by priority
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority < policies[j].Priority
	})

	return policies
}

// GetPoliciesByPriority returns policies ordered by priority
func (k Keeper) GetPoliciesByPriority(ctx sdk.Context) []types.GeoRestrictionPolicy {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.GeoPolicyByPriorityPrefixKey())
	defer iterator.Close()

	var policies []types.GeoRestrictionPolicy
	for ; iterator.Valid(); iterator.Next() {
		// Extract policy ID from key
		key := iterator.Key()
		// Key format: prefix (1) + priority (4) + separator (1) + policy_id
		if len(key) < 6 {
			continue
		}
		policyID := string(key[6:])

		policy, found := k.GetGeoPolicy(ctx, policyID)
		if found && policy.IsActive() {
			policies = append(policies, *policy)
		}
	}

	return policies
}

// ============================================================================
// Geographic Compliance Checking
// ============================================================================

// CheckGeoCompliance checks if a user's location passes geo restriction rules
func (k Keeper) CheckGeoCompliance(ctx sdk.Context, address string, location *types.GeoLocation) (*types.GeoCheckResult, error) {
	params := k.GetGeoRestrictionParams(ctx)
	if !params.Enabled {
		// Geo restrictions disabled - always allow
		result := types.NewGeoCheckResult(address, location.Country, ctx.BlockTime())
		result.IsAllowed = true
		return result, nil
	}

	// Validate location
	if err := location.Validate(); err != nil {
		return nil, err
	}

	// Check minimum confidence
	if location.Confidence < params.MinConfidenceScore {
		return nil, types.ErrGeoRestrictionInvalid.Wrapf(
			"location confidence %d below minimum %d",
			location.Confidence, params.MinConfidenceScore,
		)
	}

	// Initialize result
	result := types.NewGeoCheckResult(address, location.Country, ctx.BlockTime())
	result.Region = location.Region

	// Check global blocked countries first
	country := types.NormalizeCountryCode(location.Country)
	for _, blocked := range params.GlobalBlockedCountries {
		if types.NormalizeCountryCode(blocked) == country {
			result.IsAllowed = false
			result.BlockReason = "country is globally blocked"
			result.EnforcementLevel = types.EnforcementHardBlock
			result.AllowsOverride = false
			return result, nil
		}
	}

	// Get applicable policies ordered by priority
	policies := k.GetPoliciesByPriority(ctx)
	result.EvaluatedPolicies = int32(len(policies))

	for _, policy := range policies {
		matched, allowed, reason := k.evaluatePolicy(ctx, &policy, location)
		if matched {
			result.MatchedPolicyID = policy.PolicyID
			result.MatchedPolicyName = policy.Name
			result.EnforcementLevel = policy.EnforcementLevel
			result.AllowsOverride = policy.EnforcementLevel.AllowsOverride() && params.AllowOverrideWithMFA

			if !allowed {
				result.IsAllowed = false
				result.BlockReason = reason
				break
			}
		}
	}

	// Cache the result
	k.cacheGeoCheckResult(ctx, address, result)

	// Emit event
	if !result.IsAllowed {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeGeoCheckFailed,
				sdk.NewAttribute(types.AttributeKeyAddress, address),
				sdk.NewAttribute(types.AttributeKeyCountry, location.Country),
				sdk.NewAttribute(types.AttributeKeyPolicyID, result.MatchedPolicyID),
				sdk.NewAttribute(types.AttributeKeyBlockReason, result.BlockReason),
			),
		)
	}

	return result, nil
}

// evaluatePolicy evaluates a single policy against a location
// Returns: matched (bool), allowed (bool), reason (string)
func (k Keeper) evaluatePolicy(ctx sdk.Context, policy *types.GeoRestrictionPolicy, location *types.GeoLocation) (bool, bool, string) {
	country := types.NormalizeCountryCode(location.Country)
	region := types.NormalizeRegionCode(location.Region)

	// Check if policy applies to scopes/markets (if applicable)
	// For now, we check all policies - scope/market filtering can be added in caller

	// Allowlist mode: if allowed lists exist, location must be in them
	if policy.HasAllowlist() {
		// Check country allowlist
		if len(policy.AllowedCountries) > 0 {
			allowed := false
			for _, c := range policy.AllowedCountries {
				if types.NormalizeCountryCode(c) == country {
					allowed = true
					break
				}
			}
			if !allowed {
				return true, false, "country not in allowed list"
			}
		}

		// Check region allowlist (if region specified)
		if len(policy.AllowedRegions) > 0 && region != "" {
			allowed := false
			for _, r := range policy.AllowedRegions {
				if types.NormalizeRegionCode(r) == region {
					allowed = true
					break
				}
			}
			if !allowed {
				return true, false, "region not in allowed list"
			}
		}
	}

	// Blocklist mode: check if location is blocked
	if policy.HasBlocklist() {
		// Check country blocklist
		for _, c := range policy.BlockedCountries {
			if types.NormalizeCountryCode(c) == country {
				return true, false, "country is blocked"
			}
		}

		// Check region blocklist
		if region != "" {
			for _, r := range policy.BlockedRegions {
				if types.NormalizeRegionCode(r) == region {
					return true, false, "region is blocked"
				}
			}
		}
	}

	// If we got here and the policy has restrictions, it matched and allowed
	if policy.HasAllowlist() || policy.HasBlocklist() {
		return true, true, ""
	}

	// No restrictions in this policy - doesn't match
	return false, true, ""
}

// GetApplicablePolicies returns policies applicable to a specific scope and/or market
func (k Keeper) GetApplicablePolicies(ctx sdk.Context, scopeType, marketID string) []types.GeoRestrictionPolicy {
	allPolicies := k.GetPoliciesByPriority(ctx)

	var applicable []types.GeoRestrictionPolicy
	for _, policy := range allPolicies {
		scopeMatches := policy.AppliesToScope(scopeType) || scopeType == ""
		marketMatches := policy.AppliesToMarket(marketID) || marketID == ""

		if scopeMatches && marketMatches {
			applicable = append(applicable, policy)
		}
	}

	return applicable
}

// ============================================================================
// Country Code Validation
// ============================================================================

// ValidateCountryCode validates an ISO 3166-1 alpha-2 country code
func (k Keeper) ValidateCountryCode(ctx sdk.Context, code string) error {
	return types.ValidateCountryCode(code)
}

// ValidateRegionCode validates an ISO 3166-2 region code
func (k Keeper) ValidateRegionCode(ctx sdk.Context, code string) error {
	return types.ValidateRegionCode(code)
}

// ============================================================================
// Blocked Country Index
// ============================================================================

// GetBlockedCountries returns all countries blocked by any active policy
func (k Keeper) GetBlockedCountries(ctx sdk.Context) []string {
	params := k.GetGeoRestrictionParams(ctx)
	countrySet := make(map[string]bool)

	// Add global blocked countries
	for _, c := range params.GlobalBlockedCountries {
		countrySet[types.NormalizeCountryCode(c)] = true
	}

	// Add countries from all active policies
	policies := k.GetAllGeoPolicies(ctx, true)
	for _, policy := range policies {
		for _, c := range policy.BlockedCountries {
			countrySet[types.NormalizeCountryCode(c)] = true
		}
	}

	// Convert to sorted slice
	countries := make([]string, 0, len(countrySet))
	for c := range countrySet {
		countries = append(countries, c)
	}
	sort.Strings(countries)

	return countries
}

// GetPoliciesBlockingCountry returns all policies that block a specific country
func (k Keeper) GetPoliciesBlockingCountry(ctx sdk.Context, countryCode string) []types.GeoRestrictionPolicy {
	normalizedCode := types.NormalizeCountryCode(countryCode)
	policies := k.GetAllGeoPolicies(ctx, true)

	var blocking []types.GeoRestrictionPolicy
	for _, policy := range policies {
		for _, c := range policy.BlockedCountries {
			if types.NormalizeCountryCode(c) == normalizedCode {
				blocking = append(blocking, policy)
				break
			}
		}
	}

	return blocking
}

// IsCountryBlocked checks if a country is blocked by any active policy
func (k Keeper) IsCountryBlocked(ctx sdk.Context, countryCode string) bool {
	normalizedCode := types.NormalizeCountryCode(countryCode)

	// Check global blocklist
	params := k.GetGeoRestrictionParams(ctx)
	for _, c := range params.GlobalBlockedCountries {
		if types.NormalizeCountryCode(c) == normalizedCode {
			return true
		}
	}

	// Check policies
	policies := k.GetAllGeoPolicies(ctx, true)
	for _, policy := range policies {
		for _, c := range policy.BlockedCountries {
			if types.NormalizeCountryCode(c) == normalizedCode {
				return true
			}
		}
	}

	return false
}

// updateBlockedCountryIndex updates the blocked country index when policies change
func (k Keeper) updateBlockedCountryIndex(ctx sdk.Context, newPolicy, oldPolicy *types.GeoRestrictionPolicy) {
	store := ctx.KVStore(k.skey)

	// Remove old policy's countries from index
	if oldPolicy != nil {
		for _, c := range oldPolicy.BlockedCountries {
			code := types.NormalizeCountryCode(c)
			key := types.BlockedCountryIndexKey(code)
			bz := store.Get(key)
			if bz != nil {
				var policyIDs []string
				if err := json.Unmarshal(bz, &policyIDs); err == nil {
					// Remove old policy ID
					var updated []string
					for _, id := range policyIDs {
						if id != oldPolicy.PolicyID {
							updated = append(updated, id)
						}
					}
					if len(updated) == 0 {
						store.Delete(key)
					} else {
						if newBz, err := json.Marshal(updated); err == nil {
							store.Set(key, newBz)
						}
					}
				}
			}
		}
	}

	// Add new policy's countries to index
	if newPolicy != nil && newPolicy.IsActive() {
		for _, c := range newPolicy.BlockedCountries {
			code := types.NormalizeCountryCode(c)
			key := types.BlockedCountryIndexKey(code)
			var policyIDs []string
			bz := store.Get(key)
			if bz != nil {
				_ = json.Unmarshal(bz, &policyIDs)
			}
			// Add new policy ID if not already present
			found := false
			for _, id := range policyIDs {
				if id == newPolicy.PolicyID {
					found = true
					break
				}
			}
			if !found {
				policyIDs = append(policyIDs, newPolicy.PolicyID)
				if newBz, err := json.Marshal(policyIDs); err == nil {
					store.Set(key, newBz)
				}
			}
		}
	}
}

// ============================================================================
// Result Caching
// ============================================================================

// GetCachedGeoCheckResult retrieves a cached geo check result
func (k Keeper) GetCachedGeoCheckResult(ctx sdk.Context, address string) (*types.GeoCheckResult, bool) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, false
	}

	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GeoCheckResultKey(addr.Bytes()))
	if bz == nil {
		return nil, false
	}

	var result types.GeoCheckResult
	if err := json.Unmarshal(bz, &result); err != nil {
		return nil, false
	}

	return &result, true
}

// cacheGeoCheckResult stores a geo check result in cache
func (k Keeper) cacheGeoCheckResult(ctx sdk.Context, address string, result *types.GeoCheckResult) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(result)
	if err != nil {
		return
	}
	store.Set(types.GeoCheckResultKey(addr.Bytes()), bz)
}

// InvalidateGeoCheckCache invalidates cached geo check results for an address
func (k Keeper) InvalidateGeoCheckCache(ctx sdk.Context, address string) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return
	}

	store := ctx.KVStore(k.skey)
	store.Delete(types.GeoCheckResultKey(addr.Bytes()))
}

// ============================================================================
// Helper Functions
// ============================================================================

// boolToString converts bool to string for event attributes
func boolToString(b bool) string {
	if b {
		return boolTrue
	}
	return "false"
}

// CheckGeoComplianceForScope checks geo compliance for a specific verification scope
func (k Keeper) CheckGeoComplianceForScope(ctx sdk.Context, address string, location *types.GeoLocation, scopeType string) (*types.GeoCheckResult, error) {
	params := k.GetGeoRestrictionParams(ctx)
	if !params.Enabled {
		result := types.NewGeoCheckResult(address, location.Country, ctx.BlockTime())
		result.IsAllowed = true
		return result, nil
	}

	// Validate location
	if err := location.Validate(); err != nil {
		return nil, err
	}

	result := types.NewGeoCheckResult(address, location.Country, ctx.BlockTime())
	result.Region = location.Region

	// Get policies applicable to this scope
	policies := k.GetApplicablePolicies(ctx, scopeType, "")
	result.EvaluatedPolicies = int32(len(policies))

	// Check global blocklist
	country := types.NormalizeCountryCode(location.Country)
	for _, blocked := range params.GlobalBlockedCountries {
		if types.NormalizeCountryCode(blocked) == country {
			result.IsAllowed = false
			result.BlockReason = "country is globally blocked"
			result.EnforcementLevel = types.EnforcementHardBlock
			return result, nil
		}
	}

	// Evaluate applicable policies
	for _, policy := range policies {
		matched, allowed, reason := k.evaluatePolicy(ctx, &policy, location)
		if matched {
			result.MatchedPolicyID = policy.PolicyID
			result.MatchedPolicyName = policy.Name
			result.EnforcementLevel = policy.EnforcementLevel
			result.AllowsOverride = policy.EnforcementLevel.AllowsOverride() && params.AllowOverrideWithMFA

			if !allowed {
				result.IsAllowed = false
				result.BlockReason = reason
				break
			}
		}
	}

	return result, nil
}

// CheckIPGeoMatch checks if IP geolocation matches document country
func (k Keeper) CheckIPGeoMatch(ctx sdk.Context, documentCountry, ipCountry string) (*types.GeoCheckResult, error) {
	params := k.GetGeoRestrictionParams(ctx)

	docCountry := types.NormalizeCountryCode(documentCountry)
	ipCtry := types.NormalizeCountryCode(ipCountry)

	result := &types.GeoCheckResult{
		Country:    docCountry,
		IPCountry:  ipCtry,
		IsAllowed:  true,
		CheckedAt:  ctx.BlockTime(),
		IPMismatch: docCountry != ipCtry,
	}

	if params.RequireIPVerification && result.IPMismatch {
		result.IsAllowed = false
		result.BlockReason = "IP geolocation does not match document country"
		result.EnforcementLevel = types.EnforcementSoftBlock
		result.AllowsOverride = params.AllowOverrideWithMFA
	}

	return result, nil
}

// WithGeoPolicies iterates over all geo restriction policies
func (k Keeper) WithGeoPolicies(ctx sdk.Context, fn func(policy types.GeoRestrictionPolicy) bool) {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.GeoPolicyPrefixKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var policy types.GeoRestrictionPolicy
		if err := json.Unmarshal(iterator.Value(), &policy); err != nil {
			continue
		}
		if fn(policy) {
			break
		}
	}
}

// SetGeoPolicyStatus updates the status of a geo policy
func (k Keeper) SetGeoPolicyStatus(ctx sdk.Context, policyID string, status types.PolicyStatus, updatedBy string) error {
	policy, found := k.GetGeoPolicy(ctx, policyID)
	if !found {
		return types.ErrGeoPolicyNotFound.Wrapf("policy_id: %s", policyID)
	}

	if !status.IsValid() {
		return types.ErrGeoRestrictionInvalid.Wrap("invalid status")
	}

	oldStatus := policy.Status
	policy.Status = status
	policy.UpdatedAt = ctx.BlockTime()
	policy.UpdatedBy = updatedBy

	if err := k.setGeoPolicy(ctx, policy); err != nil {
		return err
	}

	// Update blocked country index if activation state changed
	if oldStatus.IsEnforceable() != status.IsEnforceable() {
		if status.IsEnforceable() {
			k.updateBlockedCountryIndex(ctx, policy, nil)
		} else {
			k.updateBlockedCountryIndex(ctx, nil, policy)
		}
	}

	return nil
}

// Ensure unused import doesn't cause errors
var (
	_ = strings.ToUpper
	_ = time.Now
)
