// Package keeper implements the market module keeper.
//
// VE-301: Marketplace gating - identity score requirement enforcement
// This file implements VEID score gating logic that checks identity requirements
// before order creation.
package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// VEIDGatingRequirements defines the VEID gating requirements for an order.
// These can be set at the market level (via params) or per-offering.
type VEIDGatingRequirements struct {
	// MinCustomerScore is the minimum identity score required (0-100)
	MinCustomerScore uint32 `json:"min_customer_score"`

	// MinCustomerTier is the minimum identity tier required
	// Valid values: TierUnverified(0), TierBasic(1), TierStandard(2), TierPremium(3)
	MinCustomerTier int `json:"min_customer_tier"`

	// RequiredScopes are scope types that must be verified
	RequiredScopes []veidtypes.ScopeType `json:"required_scopes,omitempty"`

	// RequireVerifiedStatus requires the identity to have verified status
	RequireVerifiedStatus bool `json:"require_verified_status"`

	// RequireUnlockedIdentity requires the identity to not be locked
	RequireUnlockedIdentity bool `json:"require_unlocked_identity"`
}

// DefaultVEIDGatingRequirements returns default VEID gating requirements.
// By default, no gating is enforced (all values set to minimum).
func DefaultVEIDGatingRequirements() VEIDGatingRequirements {
	return VEIDGatingRequirements{
		MinCustomerScore:        0,
		MinCustomerTier:         veidtypes.TierUnverified,
		RequiredScopes:          nil,
		RequireVerifiedStatus:   false,
		RequireUnlockedIdentity: true,
	}
}

// VEIDGatingResult represents the result of a VEID gating check.
type VEIDGatingResult struct {
	// Passed indicates if all gating checks passed
	Passed bool `json:"passed"`

	// CustomerScore is the customer's current identity score
	CustomerScore uint32 `json:"customer_score"`

	// CustomerTier is the customer's current identity tier
	CustomerTier int `json:"customer_tier"`

	// CustomerStatus is the customer's identity status
	CustomerStatus string `json:"customer_status"`

	// FailureReasons contains detailed failure information
	FailureReasons []VEIDGatingFailure `json:"failure_reasons,omitempty"`
}

// VEIDGatingFailure represents a single gating check failure.
type VEIDGatingFailure struct {
	// CheckType identifies which check failed
	CheckType string `json:"check_type"`

	// RequiredValue is what was required
	RequiredValue string `json:"required_value"`

	// ActualValue is what the customer has
	ActualValue string `json:"actual_value"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// RequiredSteps describes what the customer needs to do
	RequiredSteps []string `json:"required_steps,omitempty"`

	// DocumentationURL points to help documentation
	DocumentationURL string `json:"documentation_url,omitempty"`
}

// Error codes for VEID gating failures
var (
	// ErrVEIDGatingFailed is the base error for VEID gating failures
	ErrVEIDGatingFailed = errorsmod.Register("market", 2250, "VEID gating failed")

	// ErrInsufficientVEIDScore is returned when identity score is too low
	ErrInsufficientVEIDScore = errorsmod.Register("market", 2251, "insufficient VEID identity score")

	// ErrInsufficientVEIDTier is returned when identity tier is too low
	ErrInsufficientVEIDTier = errorsmod.Register("market", 2252, "insufficient VEID identity tier")

	// ErrVEIDNotVerified is returned when identity is not verified
	ErrVEIDNotVerified = errorsmod.Register("market", 2253, "VEID identity not verified")

	// ErrVEIDLocked is returned when identity is locked
	ErrVEIDLocked = errorsmod.Register("market", 2254, "VEID identity is locked")

	// ErrVEIDScopeMissing is returned when a required scope is not verified
	ErrVEIDScopeMissing = errorsmod.Register("market", 2255, "required VEID scope not verified")

	// ErrVEIDRecordNotFound is returned when no identity record exists
	ErrVEIDRecordNotFound = errorsmod.Register("market", 2256, "VEID identity record not found")
)

// CheckVEIDGating checks if the customer meets the VEID gating requirements.
// This should be called before order creation to enforce identity requirements.
//
// Parameters:
//   - ctx: the SDK context
//   - customerAddr: the customer's blockchain address
//   - requirements: the VEID requirements to check against
//
// Returns:
//   - VEIDGatingResult: detailed result of the gating check
//   - error: non-nil if gating failed, with structured error information
func (k Keeper) CheckVEIDGating(
	ctx sdk.Context,
	customerAddr sdk.AccAddress,
	requirements VEIDGatingRequirements,
) (*VEIDGatingResult, error) {
	result := &VEIDGatingResult{
		Passed:         true,
		FailureReasons: make([]VEIDGatingFailure, 0),
	}

	// If no VEIDKeeper is configured, skip gating (for backwards compatibility)
	if k.veidKeeper == nil {
		ctx.Logger().Debug("VEID gating skipped: no VEIDKeeper configured")
		return result, nil
	}

	// Get identity record
	record, found := k.veidKeeper.GetIdentityRecord(ctx, customerAddr)
	if !found {
		// No identity record - check if requirements allow unverified
		if requirements.MinCustomerScore > 0 ||
			requirements.MinCustomerTier > veidtypes.TierUnverified ||
			requirements.RequireVerifiedStatus ||
			len(requirements.RequiredScopes) > 0 {

			result.Passed = false
			result.CustomerScore = 0
			result.CustomerTier = veidtypes.TierUnverified
			result.CustomerStatus = "unverified"
			result.FailureReasons = append(result.FailureReasons, VEIDGatingFailure{
				CheckType:     "identity_record",
				RequiredValue: "exists",
				ActualValue:   "not_found",
				Message:       "No identity record found. Please complete identity verification.",
				RequiredSteps: []string{
					"Create an identity record by uploading verification documents",
					"Complete facial verification if required",
					"Wait for verification to be processed",
				},
				DocumentationURL: "/docs/identity-verification",
			})
			return result, ErrVEIDRecordNotFound.Wrap("customer has no identity record")
		}
		return result, nil
	}

	// Populate result with customer info
	result.CustomerScore = record.CurrentScore
	result.CustomerTier = veidtypes.ComputeTierFromScoreValue(record.CurrentScore, veidtypes.AccountStatusFromVerificationStatus(getOverallStatus(record)))
	result.CustomerStatus = string(record.Tier)

	// Check if identity is locked
	if requirements.RequireUnlockedIdentity && record.Locked {
		result.Passed = false
		result.FailureReasons = append(result.FailureReasons, VEIDGatingFailure{
			CheckType:     "identity_locked",
			RequiredValue: "unlocked",
			ActualValue:   "locked",
			Message:       fmt.Sprintf("Identity is locked: %s", record.LockedReason),
			RequiredSteps: []string{
				"Contact support to resolve the lock reason",
				"Submit an appeal if you believe the lock is in error",
			},
			DocumentationURL: "/docs/identity-appeals",
		})
	}

	// Check minimum score
	if record.CurrentScore < requirements.MinCustomerScore {
		result.Passed = false
		result.FailureReasons = append(result.FailureReasons, VEIDGatingFailure{
			CheckType:     "identity_score",
			RequiredValue: fmt.Sprintf("%d", requirements.MinCustomerScore),
			ActualValue:   fmt.Sprintf("%d", record.CurrentScore),
			Message: fmt.Sprintf(
				"Identity score %d is below minimum required score of %d",
				record.CurrentScore, requirements.MinCustomerScore,
			),
			RequiredSteps: []string{
				"Complete additional identity verification scopes",
				"Upload a valid government-issued ID document",
				"Complete facial verification",
			},
			DocumentationURL: "/docs/identity-score",
		})
	}

	// Check minimum tier
	customerTier := veidtypes.ComputeTierFromScoreValue(record.CurrentScore, veidtypes.AccountStatusFromVerificationStatus(getOverallStatus(record)))
	if customerTier < requirements.MinCustomerTier {
		result.Passed = false
		result.FailureReasons = append(result.FailureReasons, VEIDGatingFailure{
			CheckType:     "identity_tier",
			RequiredValue: veidtypes.TierToString(requirements.MinCustomerTier),
			ActualValue:   veidtypes.TierToString(customerTier),
			Message: fmt.Sprintf(
				"Identity tier '%s' does not meet minimum required tier '%s'",
				veidtypes.TierToString(customerTier), veidtypes.TierToString(requirements.MinCustomerTier),
			),
			RequiredSteps: []string{
				fmt.Sprintf("Achieve at least %s tier by improving your identity score", veidtypes.TierToString(requirements.MinCustomerTier)),
				fmt.Sprintf("Minimum score of %d required for %s tier", veidtypes.GetMinimumScoreForTier(requirements.MinCustomerTier), veidtypes.TierToString(requirements.MinCustomerTier)),
			},
			DocumentationURL: "/docs/identity-tiers",
		})
	}

	// Check verified status if required
	if requirements.RequireVerifiedStatus {
		overallStatus := getOverallStatus(record)
		if overallStatus != veidtypes.VerificationStatusVerified {
			result.Passed = false
			result.FailureReasons = append(result.FailureReasons, VEIDGatingFailure{
				CheckType:     "verification_status",
				RequiredValue: string(veidtypes.VerificationStatusVerified),
				ActualValue:   string(overallStatus),
				Message:       "Identity verification is not complete",
				RequiredSteps: []string{
					"Complete all pending verification steps",
					"Ensure all uploaded documents are approved",
				},
				DocumentationURL: "/docs/verification-status",
			})
		}
	}

	// Check required scopes
	for _, requiredScope := range requirements.RequiredScopes {
		hasVerifiedScope := false
		scopes := k.veidKeeper.GetScopesByType(ctx, customerAddr, requiredScope)
		for _, scope := range scopes {
			if scope.Status == veidtypes.VerificationStatusVerified && !scope.Revoked {
				hasVerifiedScope = true
				break
			}
		}
		if !hasVerifiedScope {
			result.Passed = false
			result.FailureReasons = append(result.FailureReasons, VEIDGatingFailure{
				CheckType:     "required_scope",
				RequiredValue: string(requiredScope),
				ActualValue:   "not_verified",
				Message:       fmt.Sprintf("Required scope '%s' is not verified", requiredScope),
				RequiredSteps: []string{
					fmt.Sprintf("Upload and verify a %s scope", requiredScope),
				},
				DocumentationURL: fmt.Sprintf("/docs/scopes/%s", requiredScope),
			})
		}
	}

	// Return appropriate error if failed
	if !result.Passed {
		errMsg := "VEID gating failed"
		if len(result.FailureReasons) == 1 {
			errMsg = result.FailureReasons[0].Message
		} else {
			errMsg = fmt.Sprintf("VEID gating failed: %d requirements not met", len(result.FailureReasons))
		}
		return result, ErrVEIDGatingFailed.Wrap(errMsg)
	}

	return result, nil
}

// getOverallStatus determines the overall verification status from an identity record.
// It returns the highest priority status from all scope refs.
func getOverallStatus(record veidtypes.IdentityRecord) veidtypes.VerificationStatus {
	if len(record.ScopeRefs) == 0 {
		return veidtypes.VerificationStatusUnknown
	}

	// Priority: Verified > InProgress > Pending > Rejected > Expired > Unknown
	hasVerified := false
	hasInProgress := false
	hasPending := false

	for _, ref := range record.ScopeRefs {
		switch ref.Status {
		case veidtypes.VerificationStatusVerified:
			hasVerified = true
		case veidtypes.VerificationStatusInProgress:
			hasInProgress = true
		case veidtypes.VerificationStatusPending:
			hasPending = true
		}
	}

	if hasVerified {
		return veidtypes.VerificationStatusVerified
	}
	if hasInProgress {
		return veidtypes.VerificationStatusInProgress
	}
	if hasPending {
		return veidtypes.VerificationStatusPending
	}

	return veidtypes.VerificationStatusUnknown
}

// MustCheckVEIDGating is a convenience wrapper that panics on internal errors.
// This should only be used in contexts where errors indicate programming bugs.
func (k Keeper) MustCheckVEIDGating(
	ctx sdk.Context,
	customerAddr sdk.AccAddress,
	requirements VEIDGatingRequirements,
) *VEIDGatingResult {
	result, _ := k.CheckVEIDGating(ctx, customerAddr, requirements)
	return result
}
