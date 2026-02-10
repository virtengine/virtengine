package types

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================================
// Enhanced Eligibility Types (VE-3033)
// Provides detailed eligibility checking with remediation hints
// ============================================================================

// EligibilityCheckType represents the type of eligibility check performed
type EligibilityCheckType string

const (
	// EligibilityCheckTypeTierRequirement checks if tier meets minimum requirement
	EligibilityCheckTypeTierRequirement EligibilityCheckType = "tier_requirement"

	// EligibilityCheckTypeScoreThreshold checks if score meets minimum threshold
	EligibilityCheckTypeScoreThreshold EligibilityCheckType = "score_threshold"

	// EligibilityCheckTypeScopeRequired checks if required scopes are present
	EligibilityCheckTypeScopeRequired EligibilityCheckType = "scope_required"

	// EligibilityCheckTypeAccountStatus checks if account status is verified
	EligibilityCheckTypeAccountStatus EligibilityCheckType = "account_status"

	// EligibilityCheckTypeMFARequired checks if MFA is enabled when required
	EligibilityCheckTypeMFARequired EligibilityCheckType = "mfa_required"

	// EligibilityCheckTypeVerificationAge checks if verification is not expired
	EligibilityCheckTypeVerificationAge EligibilityCheckType = "verification_age"

	// EligibilityCheckTypeLivenessRequired checks if liveness detection is verified
	EligibilityCheckTypeLivenessRequired EligibilityCheckType = "liveness_required"

	// EligibilityCheckTypeBiometricRequired checks if biometric data is verified
	EligibilityCheckTypeBiometricRequired EligibilityCheckType = "biometric_required"
)

// AllEligibilityCheckTypes returns all valid eligibility check types
func AllEligibilityCheckTypes() []EligibilityCheckType {
	return []EligibilityCheckType{
		EligibilityCheckTypeTierRequirement,
		EligibilityCheckTypeScoreThreshold,
		EligibilityCheckTypeScopeRequired,
		EligibilityCheckTypeAccountStatus,
		EligibilityCheckTypeMFARequired,
		EligibilityCheckTypeVerificationAge,
		EligibilityCheckTypeLivenessRequired,
		EligibilityCheckTypeBiometricRequired,
	}
}

// EligibilityCheck represents a single eligibility check with detailed results
type EligibilityCheck struct {
	// CheckType identifies the type of check performed
	CheckType EligibilityCheckType `json:"check_type"`

	// Passed indicates if the check passed
	Passed bool `json:"passed"`

	// Reason provides a human-readable explanation
	Reason string `json:"reason"`

	// CurrentValue is the current value being checked (varies by check type)
	CurrentValue interface{} `json:"current_value,omitempty"`

	// RequiredValue is the required value to pass (varies by check type)
	RequiredValue interface{} `json:"required_value,omitempty"`

	// Weight indicates the importance of this check (0-100)
	Weight int `json:"weight,omitempty"`
}

// NewEligibilityCheck creates a new eligibility check result
func NewEligibilityCheck(checkType EligibilityCheckType, passed bool, reason string) EligibilityCheck {
	return EligibilityCheck{
		CheckType: checkType,
		Passed:    passed,
		Reason:    reason,
	}
}

// WithValues adds current and required values to the check
func (c EligibilityCheck) WithValues(current, required interface{}) EligibilityCheck {
	c.CurrentValue = current
	c.RequiredValue = required
	return c
}

// WithWeight adds a weight to the check
func (c EligibilityCheck) WithWeight(weight int) EligibilityCheck {
	c.Weight = weight
	return c
}

// RemediationHint provides guidance on how to address an eligibility failure
type RemediationHint struct {
	// Issue describes the specific issue preventing eligibility
	Issue string `json:"issue"`

	// Action describes the recommended action to resolve the issue
	Action string `json:"action"`

	// DocumentsNeeded lists any documents required for remediation
	DocumentsNeeded []string `json:"documents_needed,omitempty"`

	// EstimatedTime is the estimated time to complete remediation
	EstimatedTime time.Duration `json:"estimated_time,omitempty"`

	// Priority indicates the priority of this remediation (1 = highest)
	Priority int `json:"priority,omitempty"`

	// Category groups related remediation steps
	Category string `json:"category,omitempty"`
}

// NewRemediationHint creates a new remediation hint
func NewRemediationHint(issue, action string) RemediationHint {
	return RemediationHint{
		Issue:  issue,
		Action: action,
	}
}

// WithDocuments adds required documents to the hint
func (h RemediationHint) WithDocuments(docs ...string) RemediationHint {
	h.DocumentsNeeded = docs
	return h
}

// WithEstimatedTime adds an estimated completion time
func (h RemediationHint) WithEstimatedTime(d time.Duration) RemediationHint {
	h.EstimatedTime = d
	return h
}

// WithPriority sets the priority of this remediation step
func (h RemediationHint) WithPriority(priority int) RemediationHint {
	h.Priority = priority
	return h
}

// WithCategory sets the category for this remediation step
func (h RemediationHint) WithCategory(category string) RemediationHint {
	h.Category = category
	return h
}

// EnhancedEligibilityResult provides comprehensive eligibility information
type EnhancedEligibilityResult struct {
	// IsEligible indicates overall eligibility status
	IsEligible bool `json:"is_eligible"`

	// AccountAddress is the address that was checked
	AccountAddress string `json:"account_address"`

	// CurrentTier is the account's current identity tier
	CurrentTier IdentityTier `json:"current_tier"`

	// RequiredTier is the minimum tier required for eligibility
	RequiredTier IdentityTier `json:"required_tier"`

	// CurrentScore is the account's current identity score
	CurrentScore uint32 `json:"current_score"`

	// RequiredScore is the minimum score required for eligibility
	RequiredScore uint32 `json:"required_score"`

	// VerifiedScopes lists scopes that are currently verified
	VerifiedScopes []ScopeType `json:"verified_scopes,omitempty"`

	// MissingScopes lists scopes that are required but not verified
	MissingScopes []ScopeType `json:"missing_scopes,omitempty"`

	// Checks contains all eligibility checks performed
	Checks []EligibilityCheck `json:"checks"`

	// FailedChecks contains only the checks that failed
	FailedChecks []EligibilityCheck `json:"failed_checks,omitempty"`

	// Remediation provides hints for resolving failed checks
	Remediation []RemediationHint `json:"remediation,omitempty"`

	// NextSteps provides ordered steps to achieve eligibility
	NextSteps []string `json:"next_steps,omitempty"`

	// Summary provides a human-readable summary of eligibility status
	Summary string `json:"summary"`

	// CheckedAt is when the eligibility check was performed
	CheckedAt time.Time `json:"checked_at"`

	// Context provides additional context for the eligibility check
	Context EligibilityContext `json:"context,omitempty"`
}

// EligibilityContext provides additional context for eligibility checks
type EligibilityContext struct {
	// OfferingType is the offering type being checked (if applicable)
	OfferingType OfferingType `json:"offering_type,omitempty"`

	// MarketType is the market type being checked (if applicable)
	MarketType MarketType `json:"market_type,omitempty"`

	// RequiresMFA indicates if MFA is required
	RequiresMFA bool `json:"requires_mfa"`

	// MFAEnabled indicates if MFA is currently enabled
	MFAEnabled bool `json:"mfa_enabled"`

	// LastVerifiedAt is when the account was last verified
	LastVerifiedAt *time.Time `json:"last_verified_at,omitempty"`

	// VerificationExpiresAt is when the current verification expires
	VerificationExpiresAt *time.Time `json:"verification_expires_at,omitempty"`
}

// NewEnhancedEligibilityResult creates a new enhanced eligibility result
func NewEnhancedEligibilityResult(address string, checkedAt time.Time) *EnhancedEligibilityResult {
	return &EnhancedEligibilityResult{
		AccountAddress: address,
		Checks:         make([]EligibilityCheck, 0),
		FailedChecks:   make([]EligibilityCheck, 0),
		Remediation:    make([]RemediationHint, 0),
		NextSteps:      make([]string, 0),
		VerifiedScopes: make([]ScopeType, 0),
		MissingScopes:  make([]ScopeType, 0),
		CheckedAt:      checkedAt,
	}
}

// AddCheck adds an eligibility check to the result
func (r *EnhancedEligibilityResult) AddCheck(check EligibilityCheck) {
	r.Checks = append(r.Checks, check)
	if !check.Passed {
		r.FailedChecks = append(r.FailedChecks, check)
	}
}

// AddRemediation adds a remediation hint to the result
func (r *EnhancedEligibilityResult) AddRemediation(hint RemediationHint) {
	r.Remediation = append(r.Remediation, hint)
}

// AddNextStep adds a next step to the result
func (r *EnhancedEligibilityResult) AddNextStep(step string) {
	r.NextSteps = append(r.NextSteps, step)
}

// SetScopes sets the verified and missing scopes
func (r *EnhancedEligibilityResult) SetScopes(verified, missing []ScopeType) {
	r.VerifiedScopes = verified
	r.MissingScopes = missing
}

// Finalize computes the final eligibility status and summary
func (r *EnhancedEligibilityResult) Finalize() {
	// Eligible if no checks failed
	r.IsEligible = len(r.FailedChecks) == 0

	// Generate summary
	if r.IsEligible {
		r.Summary = fmt.Sprintf("Account %s is eligible with %s tier (score: %d)",
			r.AccountAddress, r.CurrentTier, r.CurrentScore)
	} else {
		failedCount := len(r.FailedChecks)
		r.Summary = fmt.Sprintf("Account %s is not eligible: %d check(s) failed",
			r.AccountAddress, failedCount)
	}
}

// PassedCheckCount returns the number of passed checks
func (r *EnhancedEligibilityResult) PassedCheckCount() int {
	return len(r.Checks) - len(r.FailedChecks)
}

// TotalCheckCount returns the total number of checks
func (r *EnhancedEligibilityResult) TotalCheckCount() int {
	return len(r.Checks)
}

// GetFailureReasons returns a list of failure reasons
func (r *EnhancedEligibilityResult) GetFailureReasons() []string {
	reasons := make([]string, 0, len(r.FailedChecks))
	for _, check := range r.FailedChecks {
		reasons = append(reasons, check.Reason)
	}
	return reasons
}

// ============================================================================
// Tier Requirements
// ============================================================================

// TierRequirements defines the requirements for achieving a specific tier
type TierRequirements struct {
	// Tier is the target tier
	Tier IdentityTier `json:"tier"`

	// MinScore is the minimum score required
	MinScore uint32 `json:"min_score"`

	// MaxScore is the maximum score for this tier (exclusive for next tier)
	MaxScore uint32 `json:"max_score"`

	// RequiredScopes lists the scope types required for this tier
	RequiredScopes []ScopeType `json:"required_scopes"`

	// OptionalScopes lists scopes that can contribute to this tier
	OptionalScopes []ScopeType `json:"optional_scopes,omitempty"`

	// RequiresMFA indicates if MFA is required for this tier
	RequiresMFA bool `json:"requires_mfa"`

	// Description provides a human-readable description of the tier
	Description string `json:"description"`

	// Benefits lists the benefits of achieving this tier
	Benefits []string `json:"benefits,omitempty"`
}

// GetTierRequirements returns the requirements for a specific tier
func GetTierRequirements(tier IdentityTier) *TierRequirements {
	switch tier {
	case IdentityTierUnverified:
		return &TierRequirements{
			Tier:           IdentityTierUnverified,
			MinScore:       0,
			MaxScore:       ThresholdBasic,
			RequiredScopes: []ScopeType{},
			RequiresMFA:    false,
			Description:    "Unverified accounts with no identity verification",
			Benefits:       []string{"Browse marketplace", "View public resources"},
		}
	case IdentityTierBasic:
		return &TierRequirements{
			Tier:           IdentityTierBasic,
			MinScore:       ThresholdBasic,
			MaxScore:       ThresholdStandard,
			RequiredScopes: []ScopeType{ScopeTypeEmailProof},
			OptionalScopes: []ScopeType{ScopeTypeSMSProof},
			RequiresMFA:    false,
			Description:    "Basic verification with email confirmation",
			Benefits: []string{
				"Access basic compute marketplace",
				"Create limited orders",
				"Basic storage access",
			},
		}
	case IdentityTierStandard:
		return &TierRequirements{
			Tier:           IdentityTierStandard,
			MinScore:       ThresholdStandard,
			MaxScore:       ThresholdPremium,
			RequiredScopes: []ScopeType{ScopeTypeEmailProof, ScopeTypeSelfie, ScopeTypeIDDocument},
			OptionalScopes: []ScopeType{ScopeTypeSMSProof, ScopeTypeDomainVerify},
			RequiresMFA:    true,
			Description:    "Standard verification with ID document and selfie",
			Benefits: []string{
				"Full marketplace access",
				"Become a provider",
				"Access HPC and GPU resources",
				"Higher order limits",
			},
		}
	case IdentityTierPremium:
		return &TierRequirements{
			Tier:     IdentityTierPremium,
			MinScore: ThresholdPremium,
			MaxScore: MaxScore + 1, // No upper limit
			RequiredScopes: []ScopeType{
				ScopeTypeEmailProof,
				ScopeTypeSelfie,
				ScopeTypeIDDocument,
				ScopeTypeFaceVideo,
			},
			OptionalScopes: []ScopeType{ScopeTypeBiometric, ScopeTypeDomainVerify, ScopeTypeADSSO},
			RequiresMFA:    true,
			Description:    "Premium verification with liveness detection",
			Benefits: []string{
				"Access TEE marketplace",
				"Validator eligibility",
				"Premium support",
				"No order limits",
				"Enterprise features",
			},
		}
	default:
		return nil
	}
}

// GetAllTierRequirements returns requirements for all tiers
func GetAllTierRequirements() []*TierRequirements {
	return []*TierRequirements{
		GetTierRequirements(IdentityTierUnverified),
		GetTierRequirements(IdentityTierBasic),
		GetTierRequirements(IdentityTierStandard),
		GetTierRequirements(IdentityTierPremium),
	}
}

// GetRequiredScopesForTier returns the scopes required for a tier
func GetRequiredScopesForTier(tier IdentityTier) []ScopeType {
	req := GetTierRequirements(tier)
	if req == nil {
		return []ScopeType{}
	}
	return req.RequiredScopes
}

// ============================================================================
// Failure Explanation
// ============================================================================

// FailureExplanation provides a detailed human-readable explanation of a failure
type FailureExplanation struct {
	// Title is a short summary of the failure
	Title string `json:"title"`

	// Description provides detailed explanation
	Description string `json:"description"`

	// FailedChecks lists the specific checks that failed
	FailedChecks []string `json:"failed_checks"`

	// Steps lists ordered steps to resolve the failure
	Steps []string `json:"steps"`

	// EstimatedTime is the estimated time to resolve
	EstimatedTime time.Duration `json:"estimated_time,omitempty"`

	// Severity indicates how critical this failure is (low, medium, high, critical)
	Severity string `json:"severity"`
}

// NewFailureExplanation creates a failure explanation from an eligibility result
func NewFailureExplanation(result *EnhancedEligibilityResult) *FailureExplanation {
	if result.IsEligible {
		return nil
	}

	explanation := &FailureExplanation{
		FailedChecks: make([]string, 0, len(result.FailedChecks)),
		Steps:        make([]string, 0),
	}

	// Build title
	if len(result.FailedChecks) == 1 {
		explanation.Title = fmt.Sprintf("Eligibility check failed: %s", result.FailedChecks[0].Reason)
	} else {
		explanation.Title = fmt.Sprintf("%d eligibility checks failed", len(result.FailedChecks))
	}

	// Build description
	descParts := make([]string, 0, len(result.FailedChecks))
	for _, check := range result.FailedChecks {
		explanation.FailedChecks = append(explanation.FailedChecks, string(check.CheckType))
		descParts = append(descParts, check.Reason)
	}
	explanation.Description = strings.Join(descParts, "; ")

	// Add steps from next steps
	explanation.Steps = append(explanation.Steps, result.NextSteps...)

	// Calculate severity based on number and type of failed checks
	explanation.Severity = calculateSeverity(result.FailedChecks)

	// Estimate time
	for _, hint := range result.Remediation {
		explanation.EstimatedTime += hint.EstimatedTime
	}

	return explanation
}

// calculateSeverity determines the severity based on failed checks
func calculateSeverity(failedChecks []EligibilityCheck) string {
	if len(failedChecks) == 0 {
		return "low"
	}

	// Check for critical failures
	for _, check := range failedChecks {
		switch check.CheckType {
		case EligibilityCheckTypeAccountStatus:
			return "critical"
		case EligibilityCheckTypeTierRequirement, EligibilityCheckTypeScoreThreshold:
			if len(failedChecks) > 2 {
				return "high"
			}
			return "medium"
		}
	}

	if len(failedChecks) > 3 {
		return "high"
	}
	if len(failedChecks) > 1 {
		return "medium"
	}
	return "low"
}

// ============================================================================
// Remediation Templates
// ============================================================================

// Standard remediation hints for common issues
var (
	// RemediationEmailVerification is the standard hint for email verification
	RemediationEmailVerification = RemediationHint{
		Issue:           "Email address not verified",
		Action:          "Complete email verification by clicking the link sent to your registered email",
		DocumentsNeeded: []string{},
		EstimatedTime:   5 * time.Minute,
		Priority:        1,
		Category:        "basic_verification",
	}

	// RemediationIDDocument is the standard hint for ID document upload
	RemediationIDDocument = RemediationHint{
		Issue:  "Government-issued ID not verified",
		Action: "Upload a clear photo of your valid government-issued ID (passport, driver's license, or national ID)",
		DocumentsNeeded: []string{
			"Valid passport", "OR Driver's license", "OR National ID card",
		},
		EstimatedTime: 15 * time.Minute,
		Priority:      2,
		Category:      "document_verification",
	}

	// RemediationSelfie is the standard hint for selfie verification
	RemediationSelfie = RemediationHint{
		Issue:           "Selfie verification not completed",
		Action:          "Take a clear selfie photo with good lighting for face verification",
		DocumentsNeeded: []string{},
		EstimatedTime:   5 * time.Minute,
		Priority:        3,
		Category:        "biometric_verification",
	}

	// RemediationLiveness is the standard hint for liveness detection
	RemediationLiveness = RemediationHint{
		Issue:           "Liveness detection not completed",
		Action:          "Complete the video verification process to prove you are a real person",
		DocumentsNeeded: []string{},
		EstimatedTime:   10 * time.Minute,
		Priority:        4,
		Category:        "biometric_verification",
	}

	// RemediationMFA is the standard hint for MFA enablement
	RemediationMFA = RemediationHint{
		Issue:           "Multi-factor authentication not enabled",
		Action:          "Enable MFA using an authenticator app or hardware security key",
		DocumentsNeeded: []string{},
		EstimatedTime:   5 * time.Minute,
		Priority:        2,
		Category:        "security",
	}

	// RemediationScoreImprovement is the standard hint for score improvement
	RemediationScoreImprovement = RemediationHint{
		Issue:           "Identity score below required threshold",
		Action:          "Complete additional verification steps to improve your identity score",
		DocumentsNeeded: []string{},
		EstimatedTime:   30 * time.Minute,
		Priority:        1,
		Category:        "score_improvement",
	}
)

// GetRemediationForScope returns the remediation hint for a missing scope
func GetRemediationForScope(scopeType ScopeType) RemediationHint {
	switch scopeType {
	case ScopeTypeEmailProof:
		return RemediationEmailVerification
	case ScopeTypeIDDocument:
		return RemediationIDDocument
	case ScopeTypeSelfie:
		return RemediationSelfie
	case ScopeTypeFaceVideo:
		return RemediationLiveness
	case ScopeTypeSMSProof:
		return RemediationHint{
			Issue:           "Phone number not verified",
			Action:          "Verify your phone number via SMS code",
			DocumentsNeeded: []string{},
			EstimatedTime:   5 * time.Minute,
			Priority:        2,
			Category:        "basic_verification",
		}
	case ScopeTypeDomainVerify:
		return RemediationHint{
			Issue:  "Domain ownership not verified",
			Action: "Add a DNS TXT record to verify domain ownership",
			DocumentsNeeded: []string{
				"Access to domain DNS settings",
			},
			EstimatedTime: 15 * time.Minute,
			Priority:      3,
			Category:      "enterprise_verification",
		}
	case ScopeTypeBiometric:
		return RemediationHint{
			Issue:           "Biometric data not verified",
			Action:          "Complete biometric verification (fingerprint or voice)",
			DocumentsNeeded: []string{},
			EstimatedTime:   10 * time.Minute,
			Priority:        4,
			Category:        "biometric_verification",
		}
	case ScopeTypeADSSO:
		return RemediationHint{
			Issue:  "Active Directory SSO not configured",
			Action: "Configure enterprise SSO through Azure AD, SAML, or LDAP",
			DocumentsNeeded: []string{
				"Enterprise admin access",
				"SSO provider credentials",
			},
			EstimatedTime: 60 * time.Minute,
			Priority:      5,
			Category:      "enterprise_verification",
		}
	default:
		return RemediationHint{
			Issue:           fmt.Sprintf("Scope %s not verified", scopeType),
			Action:          "Complete the required verification step",
			DocumentsNeeded: []string{},
			EstimatedTime:   15 * time.Minute,
			Priority:        5,
			Category:        "verification",
		}
	}
}
