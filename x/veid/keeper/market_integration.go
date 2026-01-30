package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Market Integration Keeper Methods (VE-3028)
// ============================================================================

// marketRequirementsStore is the stored format of market requirements
type marketRequirementsStore struct {
	MarketType              string            `json:"market_type"`
	MinTrustScore           string            `json:"min_trust_score"`
	RequiredScopes          []string          `json:"required_scopes"`
	RequiredLevels          map[string]string `json:"required_levels"`
	AllowDelegation         bool              `json:"allow_delegation"`
	MaxDelegationAgeSecs    int64             `json:"max_delegation_age_secs"`
	RequireActiveIdentity   bool              `json:"require_active_identity"`
	RequireUnlockedIdentity bool              `json:"require_unlocked_identity"`
	ProviderMinTrustScore   *string           `json:"provider_min_trust_score,omitempty"`
	ProviderRequiredScopes  []string          `json:"provider_required_scopes,omitempty"`
	ProviderRequireDomain   bool              `json:"provider_require_domain"`
	ProviderRequireStake    bool              `json:"provider_require_stake"`
	CreatedAt               int64             `json:"created_at"`
	UpdatedAt               int64             `json:"updated_at"`
	Authority               string            `json:"authority"`
}

// requirementsToStore converts MarketVEIDRequirements to stored format
func requirementsToStore(r *types.MarketVEIDRequirements) *marketRequirementsStore {
	store := &marketRequirementsStore{
		MarketType:              string(r.MarketType),
		MinTrustScore:           r.MinTrustScore.String(),
		RequiredScopes:          make([]string, len(r.RequiredScopes)),
		RequiredLevels:          make(map[string]string),
		AllowDelegation:         r.AllowDelegation,
		MaxDelegationAgeSecs:    int64(r.MaxDelegationAge.Seconds()),
		RequireActiveIdentity:   r.RequireActiveIdentity,
		RequireUnlockedIdentity: r.RequireUnlockedIdentity,
		CreatedAt:               r.CreatedAt.Unix(),
		UpdatedAt:               r.UpdatedAt.Unix(),
		Authority:               r.Authority,
	}

	for i, scope := range r.RequiredScopes {
		store.RequiredScopes[i] = string(scope)
	}

	for scope, status := range r.RequiredLevels {
		store.RequiredLevels[string(scope)] = string(status)
	}

	if r.ProviderRequirements != nil {
		providerScore := r.ProviderRequirements.MinTrustScore.String()
		store.ProviderMinTrustScore = &providerScore
		store.ProviderRequiredScopes = make([]string, len(r.ProviderRequirements.RequiredScopes))
		for i, scope := range r.ProviderRequirements.RequiredScopes {
			store.ProviderRequiredScopes[i] = string(scope)
		}
		store.ProviderRequireDomain = r.ProviderRequirements.RequireDomainVerification
		store.ProviderRequireStake = r.ProviderRequirements.RequireActiveStake
	}

	return store
}

// requirementsFromStore converts stored format back to MarketVEIDRequirements
func requirementsFromStore(store *marketRequirementsStore) (*types.MarketVEIDRequirements, error) {
	minScore, err := sdkmath.LegacyNewDecFromStr(store.MinTrustScore)
	if err != nil {
		return nil, fmt.Errorf("failed to parse min trust score: %w", err)
	}

	r := &types.MarketVEIDRequirements{
		MarketType:              types.MarketType(store.MarketType),
		MinTrustScore:           minScore,
		RequiredScopes:          make([]types.ScopeType, len(store.RequiredScopes)),
		RequiredLevels:          make(map[types.ScopeType]types.VerificationStatus),
		AllowDelegation:         store.AllowDelegation,
		MaxDelegationAge:        time.Duration(store.MaxDelegationAgeSecs) * time.Second,
		RequireActiveIdentity:   store.RequireActiveIdentity,
		RequireUnlockedIdentity: store.RequireUnlockedIdentity,
		CreatedAt:               time.Unix(store.CreatedAt, 0).UTC(),
		UpdatedAt:               time.Unix(store.UpdatedAt, 0).UTC(),
		Authority:               store.Authority,
	}

	for i, scope := range store.RequiredScopes {
		r.RequiredScopes[i] = types.ScopeType(scope)
	}

	for scope, status := range store.RequiredLevels {
		r.RequiredLevels[types.ScopeType(scope)] = types.VerificationStatus(status)
	}

	if store.ProviderMinTrustScore != nil {
		providerScore, err := sdkmath.LegacyNewDecFromStr(*store.ProviderMinTrustScore)
		if err != nil {
			return nil, fmt.Errorf("failed to parse provider min trust score: %w", err)
		}

		r.ProviderRequirements = &types.ProviderVEIDRequirements{
			MinTrustScore:             providerScore,
			RequiredScopes:            make([]types.ScopeType, len(store.ProviderRequiredScopes)),
			RequireDomainVerification: store.ProviderRequireDomain,
			RequireActiveStake:        store.ProviderRequireStake,
		}

		for i, scope := range store.ProviderRequiredScopes {
			r.ProviderRequirements.RequiredScopes[i] = types.ScopeType(scope)
		}
	}

	return r, nil
}

// ============================================================================
// Market Requirements CRUD
// ============================================================================

// SetMarketRequirements configures VEID requirements for a market type.
// Only the module authority (governance) can set these requirements.
func (k Keeper) SetMarketRequirements(ctx sdk.Context, requirements *types.MarketVEIDRequirements) error {
	if err := requirements.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.MarketRequirementsKey(requirements.MarketType)

	storeData := requirementsToStore(requirements)
	bz, err := json.Marshal(storeData)
	if err != nil {
		return fmt.Errorf("failed to marshal market requirements: %w", err)
	}

	store.Set(key, bz)

	k.Logger(ctx).Info("Market VEID requirements set",
		"market_type", requirements.MarketType,
		"min_trust_score", requirements.MinTrustScore.String(),
		"required_scopes", len(requirements.RequiredScopes),
		"allow_delegation", requirements.AllowDelegation,
	)

	return nil
}

// GetMarketRequirements retrieves VEID requirements for a market type.
func (k Keeper) GetMarketRequirements(ctx sdk.Context, marketType types.MarketType) (*types.MarketVEIDRequirements, bool) {
	store := ctx.KVStore(k.skey)
	key := types.MarketRequirementsKey(marketType)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var storeData marketRequirementsStore
	if err := json.Unmarshal(bz, &storeData); err != nil {
		k.Logger(ctx).Error("Failed to unmarshal market requirements", "error", err)
		return nil, false
	}

	requirements, err := requirementsFromStore(&storeData)
	if err != nil {
		k.Logger(ctx).Error("Failed to convert market requirements from store", "error", err)
		return nil, false
	}

	return requirements, true
}

// DeleteMarketRequirements removes VEID requirements for a market type.
func (k Keeper) DeleteMarketRequirements(ctx sdk.Context, marketType types.MarketType) {
	store := ctx.KVStore(k.skey)
	key := types.MarketRequirementsKey(marketType)
	store.Delete(key)
}

// WithMarketRequirements iterates over all market requirements.
func (k Keeper) WithMarketRequirements(ctx sdk.Context, fn func(*types.MarketVEIDRequirements) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixMarketVEIDRequirements)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var storeData marketRequirementsStore
		if err := json.Unmarshal(iter.Value(), &storeData); err != nil {
			continue
		}

		requirements, err := requirementsFromStore(&storeData)
		if err != nil {
			continue
		}

		if fn(requirements) {
			return
		}
	}
}

// ============================================================================
// Participant Validation
// ============================================================================

// ValidateParticipant checks if an address meets VEID requirements for a market type.
// Returns true if the participant meets all requirements.
func (k Keeper) ValidateParticipant(ctx sdk.Context, address sdk.AccAddress, marketType types.MarketType) (bool, error) {
	result, err := k.checkParticipantEligibility(ctx, address, marketType, false)
	if err != nil {
		return false, err
	}
	return result.Eligible, nil
}

// ValidateProvider checks if an address meets VEID requirements for a provider in a market type.
// Applies additional provider-specific requirements.
func (k Keeper) ValidateProvider(ctx sdk.Context, address sdk.AccAddress, marketType types.MarketType) (bool, error) {
	result, err := k.checkParticipantEligibility(ctx, address, marketType, true)
	if err != nil {
		return false, err
	}
	return result.Eligible, nil
}

// GetParticipantStatus returns detailed VEID status for a marketplace participant.
func (k Keeper) GetParticipantStatus(ctx sdk.Context, address sdk.AccAddress, marketType types.MarketType) (*types.ParticipantVEIDStatus, error) {
	status := types.NewParticipantVEIDStatus(address.String())

	// Get identity record
	identity, found := k.GetIdentityRecord(ctx, address)
	if !found {
		status.EligibilityReason = "no identity record found"
		return status, nil
	}

	// Populate basic status
	status.IsVerified = identity.CurrentScore >= types.ThresholdBasic
	status.TrustScore = sdkmath.LegacyNewDec(int64(identity.CurrentScore))
	status.Tier = identity.Tier
	status.IsLocked = identity.Locked
	status.LastVerifiedAt = identity.LastVerifiedAt

	// Check expiration
	if identity.LastVerifiedAt != nil {
		params := k.GetParams(ctx)
		expirationDuration := time.Duration(params.VerificationExpiryDays) * 24 * time.Hour
		if ctx.BlockTime().After(identity.LastVerifiedAt.Add(expirationDuration)) {
			status.IsExpired = true
		}
	}

	// Get verified scopes
	k.WithScopes(ctx, address, func(scope types.IdentityScope) bool {
		if scope.Status == types.VerificationStatusVerified {
			status.VerifiedScopes = append(status.VerifiedScopes, scope.ScopeType)
		}
		return false
	})

	// Check requirements if available
	requirements, found := k.GetMarketRequirements(ctx, marketType)
	if found {
		status.MeetsRequirements = true

		// Check trust score
		if status.TrustScore.LT(requirements.MinTrustScore) {
			status.MeetsRequirements = false
			status.EligibilityReason = "trust score below minimum requirement"
		}

		// Check required scopes
		for _, reqScope := range requirements.RequiredScopes {
			found := false
			for _, verifiedScope := range status.VerifiedScopes {
				if verifiedScope == reqScope {
					found = true
					break
				}
			}
			if !found {
				status.MissingScopes = append(status.MissingScopes, reqScope)
				status.MeetsRequirements = false
			}
		}

		if len(status.MissingScopes) > 0 {
			status.EligibilityReason = "missing required verification scopes"
		}

		// Check locked/expired
		if requirements.RequireUnlockedIdentity && status.IsLocked {
			status.MeetsRequirements = false
			status.EligibilityReason = "identity is locked"
		}
		if requirements.RequireActiveIdentity && status.IsExpired {
			status.MeetsRequirements = false
			status.EligibilityReason = "identity verification has expired"
		}
	} else {
		// Apply default requirements
		status.MeetsRequirements = status.IsVerified && !status.IsLocked && !status.IsExpired
	}

	return status, nil
}

// ============================================================================
// Eligibility Checks
// ============================================================================

// CheckOrderEligibility performs a detailed eligibility check for order creation.
func (k Keeper) CheckOrderEligibility(ctx sdk.Context, tenantAddress sdk.AccAddress, marketType types.MarketType) (*types.MarketEligibilityResult, error) {
	return k.checkParticipantEligibility(ctx, tenantAddress, marketType, false)
}

// CheckBidEligibility performs a detailed eligibility check for bid submission.
func (k Keeper) CheckBidEligibility(ctx sdk.Context, providerAddress sdk.AccAddress, marketType types.MarketType) (*types.MarketEligibilityResult, error) {
	return k.checkParticipantEligibility(ctx, providerAddress, marketType, true)
}

// CheckLeaseEligibility performs a detailed eligibility check for lease signing.
func (k Keeper) CheckLeaseEligibility(ctx sdk.Context, tenantAddress, providerAddress sdk.AccAddress, marketType types.MarketType) (*types.MarketEligibilityResult, error) {
	// Check tenant eligibility
	tenantResult, err := k.checkParticipantEligibility(ctx, tenantAddress, marketType, false)
	if err != nil {
		return nil, err
	}

	if !tenantResult.Eligible {
		tenantResult.Reason = fmt.Sprintf("tenant not eligible: %s", tenantResult.Reason)
		return tenantResult, nil
	}

	// Check provider eligibility
	providerResult, err := k.checkParticipantEligibility(ctx, providerAddress, marketType, true)
	if err != nil {
		return nil, err
	}

	if !providerResult.Eligible {
		providerResult.Reason = fmt.Sprintf("provider not eligible: %s", providerResult.Reason)
		return providerResult, nil
	}

	// Both eligible
	return types.NewMarketEligibilityResult(true, "both parties meet VEID requirements", ctx.BlockTime()), nil
}

// checkParticipantEligibility is the core eligibility checking logic.
func (k Keeper) checkParticipantEligibility(ctx sdk.Context, address sdk.AccAddress, marketType types.MarketType, isProvider bool) (*types.MarketEligibilityResult, error) {
	result := types.NewMarketEligibilityResult(false, "", ctx.BlockTime())

	// Validate market type
	if !types.IsValidMarketType(marketType) {
		result.Reason = fmt.Sprintf("invalid market type: %s", marketType)
		return result, nil
	}

	// Get requirements (or use defaults)
	requirements, found := k.GetMarketRequirements(ctx, marketType)
	if !found {
		// Use default requirements
		requirements = types.NewMarketVEIDRequirements(marketType, ctx.BlockTime())
		requirements.MinTrustScore = sdkmath.LegacyNewDec(int64(k.getDefaultVEIDLevel(marketType).MinScore()))
	}
	result.Requirements = requirements

	// Get participant status
	status, err := k.GetParticipantStatus(ctx, address, marketType)
	if err != nil {
		return nil, err
	}
	result.ParticipantStatus = status

	// Check identity exists
	if !status.IsVerified {
		result.AddValidationError("no verified identity")
		result.Reason = "participant does not have a verified identity"
		return result, nil
	}

	// Determine which requirements to use
	minScore := requirements.MinTrustScore
	requiredScopes := requirements.RequiredScopes

	if isProvider && requirements.ProviderRequirements != nil {
		// Apply stricter provider requirements
		if requirements.ProviderRequirements.MinTrustScore.GT(minScore) {
			minScore = requirements.ProviderRequirements.MinTrustScore
		}
		// Combine scopes
		scopeMap := make(map[types.ScopeType]bool)
		for _, s := range requiredScopes {
			scopeMap[s] = true
		}
		for _, s := range requirements.ProviderRequirements.RequiredScopes {
			scopeMap[s] = true
		}
		requiredScopes = make([]types.ScopeType, 0, len(scopeMap))
		for s := range scopeMap {
			requiredScopes = append(requiredScopes, s)
		}

		// Check domain verification for providers if required
		if requirements.ProviderRequirements.RequireDomainVerification {
			hasDomain := false
			for _, scope := range status.VerifiedScopes {
				if scope == types.ScopeTypeDomainVerify {
					hasDomain = true
					break
				}
			}
			if !hasDomain {
				result.AddValidationError("domain verification required for providers")
			}
		}
	}

	// Check trust score
	if status.TrustScore.LT(minScore) {
		result.AddValidationError(fmt.Sprintf("trust score %s below required %s",
			status.TrustScore.String(), minScore.String()))
	}

	// Check required scopes
	for _, reqScope := range requiredScopes {
		found := false
		for _, verifiedScope := range status.VerifiedScopes {
			if verifiedScope == reqScope {
				found = true
				break
			}
		}
		if !found {
			result.AddValidationError(fmt.Sprintf("missing required scope: %s", reqScope))
		}
	}

	// Check locked status
	if requirements.RequireUnlockedIdentity && status.IsLocked {
		result.AddValidationError("identity is locked")
	}

	// Check expiration
	if requirements.RequireActiveIdentity && status.IsExpired {
		result.AddValidationError("identity verification has expired")
	}

	// Check delegation if participant is using delegated identity
	if status.IsDelegated {
		if !requirements.AllowDelegation {
			result.AddValidationError("delegation not allowed for this market type")
		} else if requirements.MaxDelegationAge > 0 && status.DelegationID != "" {
			// Check delegation age
			delegation, found := k.GetDelegation(ctx, status.DelegationID)
			if !found {
				result.AddValidationError("delegation record not found")
			} else {
				delegationAge := ctx.BlockTime().Sub(delegation.CreatedAt)
				if delegationAge > requirements.MaxDelegationAge {
					result.AddValidationError(fmt.Sprintf("delegation is too old: %v > max %v",
						delegationAge.Truncate(time.Hour), requirements.MaxDelegationAge.Truncate(time.Hour)))
				}
			}
		}
	}

	// Determine final result
	if len(result.ValidationErrors) == 0 {
		result.Eligible = true
		result.Reason = "all VEID requirements met"
	} else {
		result.Reason = fmt.Sprintf("failed %d validation checks", len(result.ValidationErrors))
	}

	return result, nil
}

// ============================================================================
// Delegation in Market Context
// ============================================================================

// CheckDelegationForMarket validates if a delegation can be used for marketplace participation.
func (k Keeper) CheckDelegationForMarket(ctx sdk.Context, delegationID string, marketType types.MarketType) (bool, error) {
	delegation, found := k.GetDelegation(ctx, delegationID)
	if !found {
		return false, types.ErrDelegationNotFound
	}

	// Check delegation is active
	if delegation.Status != types.DelegationActive {
		switch delegation.Status {
		case types.DelegationExpired:
			return false, types.ErrDelegationExpired
		case types.DelegationRevoked:
			return false, types.ErrDelegationRevoked
		case types.DelegationExhausted:
			return false, types.ErrDelegationExhausted
		default:
			return false, types.ErrInvalidDelegation.Wrap("delegation is not active")
		}
	}

	// Check delegation has required permission
	hasProvePermission := false
	for _, perm := range delegation.Permissions {
		if perm == types.PermissionProveIdentity || perm == types.PermissionSignOnBehalf {
			hasProvePermission = true
			break
		}
	}
	if !hasProvePermission {
		return false, types.ErrDelegationPermissionDenied.Wrap("requires prove_identity or sign_on_behalf permission")
	}

	// Check market requirements for delegation
	requirements, found := k.GetMarketRequirements(ctx, marketType)
	if found {
		if !requirements.AllowDelegation {
			return false, types.ErrDelegationNotAllowed
		}

		// Check delegation age
		if requirements.MaxDelegationAge > 0 {
			delegationAge := ctx.BlockTime().Sub(delegation.CreatedAt)
			if delegationAge > requirements.MaxDelegationAge {
				return false, types.ErrDelegationTooOld.Wrapf(
					"delegation age %v exceeds maximum %v",
					delegationAge, requirements.MaxDelegationAge,
				)
			}
		}
	}

	// Check delegator's VEID status
	delegatorAddr, err := sdk.AccAddressFromBech32(delegation.DelegatorAddress)
	if err != nil {
		return false, types.ErrInvalidAddress.Wrap(err.Error())
	}

	eligible, err := k.ValidateParticipant(ctx, delegatorAddr, marketType)
	if err != nil {
		return false, err
	}
	if !eligible {
		return false, types.ErrMarketVEIDNotMet.Wrap("delegator does not meet VEID requirements")
	}

	return true, nil
}

// GetParticipantStatusWithDelegation returns participant status considering delegation.
func (k Keeper) GetParticipantStatusWithDelegation(
	ctx sdk.Context,
	address sdk.AccAddress,
	delegationID string,
	marketType types.MarketType,
) (*types.ParticipantVEIDStatus, error) {
	// If no delegation, return regular status
	if delegationID == "" {
		return k.GetParticipantStatus(ctx, address, marketType)
	}

	// Get delegation
	delegation, found := k.GetDelegation(ctx, delegationID)
	if !found {
		return nil, types.ErrDelegationNotFound
	}

	// Verify delegate matches address
	if delegation.DelegateAddress != address.String() {
		return nil, types.ErrInvalidDelegation.Wrap("delegation does not belong to this address")
	}

	// Get delegator's status
	delegatorAddr, err := sdk.AccAddressFromBech32(delegation.DelegatorAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	status, err := k.GetParticipantStatus(ctx, delegatorAddr, marketType)
	if err != nil {
		return nil, err
	}

	// Mark as delegated
	status.IsDelegated = true
	status.DelegatorAddress = delegation.DelegatorAddress
	status.DelegationID = delegationID

	// The address field should reflect the actual participant
	status.Address = address.String()

	return status, nil
}
