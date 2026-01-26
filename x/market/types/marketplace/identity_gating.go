// Package marketplace provides types for the marketplace on-chain module.
//
// VE-301: Marketplace gating: identity score requirement enforcement
// This file implements identity gating logic that checks VEID score before order placement.
package marketplace

import (
	"fmt"
)

// GatingCheckType represents the type of gating check
type GatingCheckType string

const (
	// GatingCheckIdentityScore checks the identity score requirement
	GatingCheckIdentityScore GatingCheckType = "identity_score"

	// GatingCheckIdentityStatus checks the identity status requirement
	GatingCheckIdentityStatus GatingCheckType = "identity_status"

	// GatingCheckEmailVerified checks email verification
	GatingCheckEmailVerified GatingCheckType = "email_verified"

	// GatingCheckDomainVerified checks domain verification
	GatingCheckDomainVerified GatingCheckType = "domain_verified"

	// GatingCheckMFAEnabled checks if MFA is enabled
	GatingCheckMFAEnabled GatingCheckType = "mfa_enabled"

	// GatingCheckProviderIdentity checks provider identity requirements
	GatingCheckProviderIdentity GatingCheckType = "provider_identity"
)

// GatingFailureReason provides structured information about gating failures
type GatingFailureReason struct {
	// CheckType is the type of check that failed
	CheckType GatingCheckType `json:"check_type"`

	// RequiredValue is what was required
	RequiredValue string `json:"required_value"`

	// ActualValue is what the user has
	ActualValue string `json:"actual_value"`

	// Message is a human-readable message
	Message string `json:"message"`

	// RequiredSteps describes what steps the user needs to take
	RequiredSteps []string `json:"required_steps,omitempty"`

	// DocumentationURL points to help documentation
	DocumentationURL string `json:"documentation_url,omitempty"`
}

// NewGatingFailureReason creates a new gating failure reason
func NewGatingFailureReason(checkType GatingCheckType, required, actual, message string) *GatingFailureReason {
	return &GatingFailureReason{
		CheckType:     checkType,
		RequiredValue: required,
		ActualValue:   actual,
		Message:       message,
		RequiredSteps: make([]string, 0),
	}
}

// WithSteps adds required steps to the failure reason
func (r *GatingFailureReason) WithSteps(steps ...string) *GatingFailureReason {
	r.RequiredSteps = append(r.RequiredSteps, steps...)
	return r
}

// WithDocumentation adds a documentation URL
func (r *GatingFailureReason) WithDocumentation(url string) *GatingFailureReason {
	r.DocumentationURL = url
	return r
}

// Error returns the error message
func (r *GatingFailureReason) Error() string {
	return r.Message
}

// IdentityGatingError is an error returned when identity gating fails
type IdentityGatingError struct {
	// Reasons contains all the gating check failures
	Reasons []*GatingFailureReason `json:"reasons"`

	// OfferingID is the offering that was being ordered
	OfferingID OfferingID `json:"offering_id"`

	// CustomerAddress is the customer's address
	CustomerAddress string `json:"customer_address"`
}

// NewIdentityGatingError creates a new identity gating error
func NewIdentityGatingError(offeringID OfferingID, customerAddress string) *IdentityGatingError {
	return &IdentityGatingError{
		Reasons:         make([]*GatingFailureReason, 0),
		OfferingID:      offeringID,
		CustomerAddress: customerAddress,
	}
}

// AddReason adds a failure reason
func (e *IdentityGatingError) AddReason(reason *GatingFailureReason) {
	e.Reasons = append(e.Reasons, reason)
}

// HasErrors returns true if there are any gating errors
func (e *IdentityGatingError) HasErrors() bool {
	return len(e.Reasons) > 0
}

// Error returns the error message
func (e *IdentityGatingError) Error() string {
	if len(e.Reasons) == 0 {
		return ""
	}
	if len(e.Reasons) == 1 {
		return e.Reasons[0].Message
	}
	return fmt.Sprintf("identity gating failed: %d checks failed", len(e.Reasons))
}

// CustomerIdentityInfo holds the customer's identity information for gating checks
type CustomerIdentityInfo struct {
	// Score is the customer's VEID identity score (0-100)
	Score uint32 `json:"score"`

	// Status is the customer's identity status
	Status string `json:"status"`

	// EmailVerified indicates if email is verified
	EmailVerified bool `json:"email_verified"`

	// DomainVerified indicates if domain is verified
	DomainVerified bool `json:"domain_verified"`

	// MFAEnabled indicates if MFA is enabled
	MFAEnabled bool `json:"mfa_enabled"`

	// HasEnrolledFactors indicates if the customer has MFA factors enrolled
	HasEnrolledFactors bool `json:"has_enrolled_factors"`

	// Tier is the calculated identity tier
	Tier int `json:"tier"`
}

// ProviderIdentitySettings holds provider-level identity requirements
type ProviderIdentitySettings struct {
	// RequireVerifiedIdentityForAll requires verified identity for all offerings
	RequireVerifiedIdentityForAll bool `json:"require_verified_identity_for_all"`

	// MinimumScoreForAll is the minimum score required for all offerings
	MinimumScoreForAll uint32 `json:"minimum_score_for_all"`

	// RequireMFAForAll requires MFA for all offerings
	RequireMFAForAll bool `json:"require_mfa_for_all"`
}

// DefaultProviderIdentitySettings returns default provider identity settings
func DefaultProviderIdentitySettings() ProviderIdentitySettings {
	return ProviderIdentitySettings{
		RequireVerifiedIdentityForAll: false,
		MinimumScoreForAll:            0,
		RequireMFAForAll:              false,
	}
}

// IdentityGatingChecker performs identity gating checks
type IdentityGatingChecker struct {
	// offering is the offering being ordered
	offering *Offering

	// providerSettings are the provider's identity settings
	providerSettings *ProviderIdentitySettings

	// customerInfo is the customer's identity info
	customerInfo *CustomerIdentityInfo
}

// NewIdentityGatingChecker creates a new identity gating checker
func NewIdentityGatingChecker(offering *Offering, customerInfo *CustomerIdentityInfo) *IdentityGatingChecker {
	return &IdentityGatingChecker{
		offering:     offering,
		customerInfo: customerInfo,
	}
}

// WithProviderSettings sets the provider identity settings
func (c *IdentityGatingChecker) WithProviderSettings(settings *ProviderIdentitySettings) *IdentityGatingChecker {
	c.providerSettings = settings
	return c
}

// Check performs all identity gating checks
func (c *IdentityGatingChecker) Check() *IdentityGatingError {
	gatingErr := NewIdentityGatingError(c.offering.ID, "")

	// Check offering-level identity requirements
	c.checkOfferingRequirements(gatingErr)

	// Check provider-level identity requirements
	c.checkProviderRequirements(gatingErr)

	return gatingErr
}

// checkOfferingRequirements checks the offering's identity requirements
func (c *IdentityGatingChecker) checkOfferingRequirements(gatingErr *IdentityGatingError) {
	req := c.offering.IdentityRequirement

	// Check identity score
	if c.customerInfo.Score < req.MinScore {
		reason := NewGatingFailureReason(
			GatingCheckIdentityScore,
			fmt.Sprintf("%d", req.MinScore),
			fmt.Sprintf("%d", c.customerInfo.Score),
			fmt.Sprintf("Identity score %d is below minimum required score of %d", c.customerInfo.Score, req.MinScore),
		).WithSteps(
			"Complete identity verification to improve your score",
			"Upload a valid government-issued ID document",
			"Complete facial verification",
		).WithDocumentation("/docs/identity-verification")

		gatingErr.AddReason(reason)
	}

	// Check identity status
	if req.RequiredStatus != "" && c.customerInfo.Status != req.RequiredStatus {
		reason := NewGatingFailureReason(
			GatingCheckIdentityStatus,
			req.RequiredStatus,
			c.customerInfo.Status,
			fmt.Sprintf("Identity status '%s' does not meet required status '%s'", c.customerInfo.Status, req.RequiredStatus),
		).WithSteps(
			"Complete identity verification to achieve verified status",
		)

		gatingErr.AddReason(reason)
	}

	// Check email verification
	if req.RequireVerifiedEmail && !c.customerInfo.EmailVerified {
		reason := NewGatingFailureReason(
			GatingCheckEmailVerified,
			"true",
			"false",
			"Email verification is required for this offering",
		).WithSteps(
			"Verify your email address in account settings",
		).WithDocumentation("/docs/email-verification")

		gatingErr.AddReason(reason)
	}

	// Check domain verification
	if req.RequireVerifiedDomain && !c.customerInfo.DomainVerified {
		reason := NewGatingFailureReason(
			GatingCheckDomainVerified,
			"true",
			"false",
			"Domain verification is required for this offering",
		).WithSteps(
			"Add a DNS TXT record to verify domain ownership",
			"Or use HTTP well-known verification",
		).WithDocumentation("/docs/domain-verification")

		gatingErr.AddReason(reason)
	}

	// Check MFA requirement
	if (req.RequireMFA || c.offering.RequireMFAForOrders) && !c.customerInfo.MFAEnabled {
		reason := NewGatingFailureReason(
			GatingCheckMFAEnabled,
			"true",
			"false",
			"Multi-factor authentication must be enabled for this offering",
		).WithSteps(
			"Enable MFA in your account security settings",
			"Enroll at least one authentication factor (TOTP, FIDO2, or SMS)",
		).WithDocumentation("/docs/mfa-setup")

		gatingErr.AddReason(reason)
	}
}

// checkProviderRequirements checks provider-level identity requirements
func (c *IdentityGatingChecker) checkProviderRequirements(gatingErr *IdentityGatingError) {
	if c.providerSettings == nil {
		return
	}

	// Check provider's minimum score requirement
	if c.customerInfo.Score < c.providerSettings.MinimumScoreForAll {
		reason := NewGatingFailureReason(
			GatingCheckProviderIdentity,
			fmt.Sprintf("%d", c.providerSettings.MinimumScoreForAll),
			fmt.Sprintf("%d", c.customerInfo.Score),
			fmt.Sprintf("Provider requires minimum identity score of %d", c.providerSettings.MinimumScoreForAll),
		).WithSteps(
			"Complete additional identity verification to improve your score",
		)

		gatingErr.AddReason(reason)
	}

	// Check provider's verified identity requirement
	if c.providerSettings.RequireVerifiedIdentityForAll && c.customerInfo.Status != "verified" {
		reason := NewGatingFailureReason(
			GatingCheckProviderIdentity,
			"verified",
			c.customerInfo.Status,
			"Provider requires verified identity for all offerings",
		).WithSteps(
			"Complete full identity verification",
		)

		gatingErr.AddReason(reason)
	}

	// Check provider's MFA requirement
	if c.providerSettings.RequireMFAForAll && !c.customerInfo.MFAEnabled {
		reason := NewGatingFailureReason(
			GatingCheckProviderIdentity,
			"mfa_enabled",
			"mfa_disabled",
			"Provider requires MFA for all offerings",
		).WithSteps(
			"Enable MFA in your account security settings",
		)

		gatingErr.AddReason(reason)
	}
}

// ValidateOrderCreation validates that an order can be created based on identity gating
func ValidateOrderCreation(
	offering *Offering,
	customerInfo *CustomerIdentityInfo,
	providerSettings *ProviderIdentitySettings,
) error {
	checker := NewIdentityGatingChecker(offering, customerInfo)

	if providerSettings != nil {
		checker.WithProviderSettings(providerSettings)
	}

	result := checker.Check()

	if result.HasErrors() {
		return result
	}

	return nil
}

// GatingCheckResult represents the result of a gating check
type GatingCheckResult struct {
	// Passed indicates if all checks passed
	Passed bool `json:"passed"`

	// CustomerScore is the customer's identity score
	CustomerScore uint32 `json:"customer_score"`

	// CustomerStatus is the customer's identity status
	CustomerStatus string `json:"customer_status"`

	// FailedChecks contains details of failed checks
	FailedChecks []*GatingFailureReason `json:"failed_checks,omitempty"`

	// CheckedAt is when the checks were performed
	CheckedAt int64 `json:"checked_at"`
}

// NewPassedGatingResult creates a result indicating all checks passed
func NewPassedGatingResult(score uint32, status string, checkedAt int64) *GatingCheckResult {
	return &GatingCheckResult{
		Passed:         true,
		CustomerScore:  score,
		CustomerStatus: status,
		FailedChecks:   nil,
		CheckedAt:      checkedAt,
	}
}

// NewFailedGatingResult creates a result indicating checks failed
func NewFailedGatingResult(score uint32, status string, failures []*GatingFailureReason, checkedAt int64) *GatingCheckResult {
	return &GatingCheckResult{
		Passed:         false,
		CustomerScore:  score,
		CustomerStatus: status,
		FailedChecks:   failures,
		CheckedAt:      checkedAt,
	}
}
