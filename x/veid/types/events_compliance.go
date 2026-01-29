// Package types provides VEID module types.
//
// This file defines compliance-related events for the VEID module.
//
// Task Reference: VE-3021 - KYC/AML Compliance Interface
package types

// Event types for compliance operations
const (
	// EventTypeComplianceChecked is emitted when a compliance check is performed
	EventTypeComplianceChecked = "compliance_checked"

	// EventTypeComplianceCleared is emitted when an identity is cleared
	EventTypeComplianceCleared = "compliance_cleared"

	// EventTypeComplianceFlagged is emitted when an identity is flagged
	EventTypeComplianceFlagged = "compliance_flagged"

	// EventTypeComplianceBlocked is emitted when an identity is blocked
	EventTypeComplianceBlocked = "compliance_blocked"

	// EventTypeComplianceAttested is emitted when a validator attests compliance
	EventTypeComplianceAttested = "compliance_attested"

	// EventTypeComplianceExpired is emitted when a compliance record expires
	EventTypeComplianceExpired = "compliance_expired"

	// EventTypeComplianceProviderRegistered is emitted when a provider is registered
	EventTypeComplianceProviderRegistered = "compliance_provider_registered"

	// EventTypeComplianceParamsUpdated is emitted when compliance params are updated
	EventTypeComplianceParamsUpdated = "compliance_params_updated"
)

// Event attribute keys for compliance events
const (
	// AttributeKeyComplianceStatus is the compliance status attribute
	AttributeKeyComplianceStatus = "compliance_status"

	// AttributeKeyComplianceCheckType is the check type attribute
	AttributeKeyComplianceCheckType = "check_type"

	// AttributeKeyComplianceProvider is the provider ID attribute
	AttributeKeyComplianceProvider = "provider_id"

	// AttributeKeyComplianceRiskScore is the risk score attribute
	AttributeKeyComplianceRiskScore = "risk_score"

	// AttributeKeyComplianceCheckPassed is the check passed attribute
	AttributeKeyComplianceCheckPassed = "check_passed"

	// AttributeKeyComplianceMatchScore is the match score attribute
	AttributeKeyComplianceMatchScore = "match_score"

	// AttributeKeyComplianceExpiresAt is the expiry timestamp attribute
	AttributeKeyComplianceExpiresAt = "expires_at"

	// AttributeKeyComplianceAttestationType is the attestation type attribute
	AttributeKeyComplianceAttestationType = "attestation_type"

	// AttributeKeyComplianceAttestationCount is the attestation count attribute
	AttributeKeyComplianceAttestationCount = "attestation_count"

	// AttributeKeyComplianceRestrictedRegion is the restricted region attribute
	AttributeKeyComplianceRestrictedRegion = "restricted_region"
)

// EventComplianceChecked is emitted when a compliance check is performed
type EventComplianceChecked struct {
	AccountAddress string `json:"account_address"`
	CheckType      string `json:"check_type"`
	ProviderID     string `json:"provider_id"`
	Passed         bool   `json:"passed"`
	MatchScore     int32  `json:"match_score"`
	CheckedAt      int64  `json:"checked_at"`
}

// EventComplianceStatusChanged is emitted when compliance status changes
type EventComplianceStatusChanged struct {
	AccountAddress string `json:"account_address"`
	OldStatus      string `json:"old_status"`
	NewStatus      string `json:"new_status"`
	RiskScore      int32  `json:"risk_score"`
	UpdatedAt      int64  `json:"updated_at"`
}

// EventComplianceAttested is emitted when a validator attests compliance
type EventComplianceAttested struct {
	AccountAddress   string `json:"account_address"`
	ValidatorAddress string `json:"validator_address"`
	AttestationType  string `json:"attestation_type"`
	AttestedAt       int64  `json:"attested_at"`
	ExpiresAt        int64  `json:"expires_at"`
	AttestationCount int32  `json:"attestation_count"`
}

// EventComplianceExpired is emitted when a compliance record expires
type EventComplianceExpired struct {
	AccountAddress string `json:"account_address"`
	ExpiredAt      int64  `json:"expired_at"`
	LastCheckedAt  int64  `json:"last_checked_at"`
}

// EventComplianceProviderRegistered is emitted when a provider is registered
type EventComplianceProviderRegistered struct {
	ProviderID      string   `json:"provider_id"`
	ProviderName    string   `json:"provider_name"`
	ProviderAddress string   `json:"provider_address"`
	CheckTypes      []string `json:"check_types"`
	RegisteredAt    int64    `json:"registered_at"`
}

// EventComplianceParamsUpdated is emitted when compliance params are updated
type EventComplianceParamsUpdated struct {
	RequireSanctionCheck    bool     `json:"require_sanction_check"`
	RequirePEPCheck         bool     `json:"require_pep_check"`
	RiskScoreThreshold      int32    `json:"risk_score_threshold"`
	MinAttestationsRequired int32    `json:"min_attestations_required"`
	RestrictedCountries     []string `json:"restricted_countries"`
	UpdatedAt               int64    `json:"updated_at"`
}

// ============================================================================
// Geographic Restriction Events (VE-3032)
// ============================================================================

// Event types for geographic restriction operations
const (
	// EventTypeGeoPolicyCreated is emitted when a geo restriction policy is created
	EventTypeGeoPolicyCreated = "geo_policy_created"

	// EventTypeGeoPolicyUpdated is emitted when a geo restriction policy is updated
	EventTypeGeoPolicyUpdated = "geo_policy_updated"

	// EventTypeGeoPolicyDeleted is emitted when a geo restriction policy is deleted
	EventTypeGeoPolicyDeleted = "geo_policy_deleted"

	// EventTypeGeoCheckFailed is emitted when a geographic compliance check fails
	EventTypeGeoCheckFailed = "geo_check_failed"

	// EventTypeGeoRestrictionParamsUpdated is emitted when geo restriction params are updated
	EventTypeGeoRestrictionParamsUpdated = "geo_restriction_params_updated"
)

// Event attribute keys for geographic restriction events
const (
	// AttributeKeyPolicyID is the policy ID attribute
	AttributeKeyPolicyID = "policy_id"

	// AttributeKeyPolicyName is the policy name attribute
	AttributeKeyPolicyName = "policy_name"

	// AttributeKeyPolicyStatus is the policy status attribute
	AttributeKeyPolicyStatus = "policy_status"

	// AttributeKeyEnforcementLevel is the enforcement level attribute
	AttributeKeyEnforcementLevel = "enforcement_level"

	// AttributeKeyCountry is the country code attribute
	AttributeKeyCountry = "country"

	// AttributeKeyRegion is the region code attribute
	AttributeKeyRegion = "region"

	// AttributeKeyBlockReason is the block reason attribute
	AttributeKeyBlockReason = "block_reason"

	// AttributeKeyGeoEnabled is the geo restriction enabled attribute
	AttributeKeyGeoEnabled = "geo_enabled"

	// AttributeKeyAddress is the account address attribute
	AttributeKeyAddress = "address"
)
