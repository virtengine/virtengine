package offramp

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ProfileState is the independently reviewable state of a payout support row.
type ProfileState string

const (
	ProfileUnsupported                        ProfileState = "unsupported"
	ProfileEngineeringIncomplete              ProfileState = "engineering_incomplete"
	ProfileEngineeringCompleteExternalBlocked ProfileState = "engineering_complete_external_blocked"
	ProfileCertifiedEnabled                   ProfileState = "certified_enabled"
	ProfilePaused                             ProfileState = "paused"
)

// Environment identifies the partner environment selected by a profile.
type Environment string

const (
	EnvironmentSandbox    Environment = "sandbox"
	EnvironmentProduction Environment = "production"
)

// ExecutionMode controls runtime profile gating.
type ExecutionMode string

const (
	ExecutionModeLegacy     ExecutionMode = "legacy_test"
	ExecutionModeSandbox    ExecutionMode = "engineering_sandbox"
	ExecutionModeProduction ExecutionMode = "production"
)

// PayoutCorridor is one exact jurisdiction, currency, and payout rail row.
type PayoutCorridor struct {
	ID            string            `json:"id"`
	Jurisdiction  string            `json:"jurisdiction"`
	Currency      string            `json:"currency"`
	Rail          string            `json:"rail"`
	MinimumAmount sdkmath.LegacyDec `json:"minimum_amount"`
	MaximumAmount sdkmath.LegacyDec `json:"maximum_amount"`
	DailyLimit    sdkmath.LegacyDec `json:"daily_limit"`
	QuoteTTL      time.Duration     `json:"quote_ttl"`
	Finality      string            `json:"finality"`
}

// BeneficiaryRequirements declares the non-sensitive beneficiary contract.
type BeneficiaryRequirements struct {
	TokenizedReferenceRequired bool     `json:"tokenized_reference_required"`
	ReferencePrefix            string   `json:"reference_prefix"`
	RequiredFields             []string `json:"required_fields,omitempty"`
	ProhibitedRawFields        []string `json:"prohibited_raw_fields,omitempty"`
}

// DecisionRequirements declares required compliance decisions.
type DecisionRequirements struct {
	KYCRequired       bool `json:"kyc_required"`
	SanctionsRequired bool `json:"sanctions_required"`
}

// SecretReference names a credential location without containing credential
// material. Ref must use an explicit secret-store reference scheme.
type SecretReference struct {
	Purpose string `json:"purpose"`
	Ref     string `json:"ref"`
	Version string `json:"version"`
	Scope   string `json:"scope"`
}

// WebhookKeyReference identifies one accepted webhook signing key version.
type WebhookKeyReference struct {
	KeyID     string `json:"key_id"`
	Version   string `json:"version"`
	SecretRef string `json:"secret_ref"`
}

// WebhookProfile pins the webhook protocol and rotation key set.
type WebhookProfile struct {
	Version   string                `json:"version"`
	Algorithm string                `json:"algorithm"`
	Keys      []WebhookKeyReference `json:"keys"`
}

// ApprovalEvidence records an approval reference and its accountable owner.
// It intentionally does not store contract or credential contents.
type ApprovalEvidence struct {
	Reference string `json:"reference"`
	Owner     string `json:"owner"`
}

// ProfileEvidence contains all external approvals required before production.
type ProfileEvidence struct {
	Contract            ApprovalEvidence `json:"contract"`
	Legal               ApprovalEvidence `json:"legal"`
	DPA                 ApprovalEvidence `json:"dpa"`
	Compliance          ApprovalEvidence `json:"compliance"`
	Custody             ApprovalEvidence `json:"custody"`
	Banking             ApprovalEvidence `json:"banking"`
	WebhookRegistration ApprovalEvidence `json:"webhook_registration"`
	Corridor            ApprovalEvidence `json:"corridor"`
}

// ProfileOwners names accountable operational owners for the support row.
type ProfileOwners struct {
	Engineering string `json:"engineering"`
	Operations  string `json:"operations"`
	Compliance  string `json:"compliance"`
	Security    string `json:"security"`
}

// PayoutProfile is a versioned fiat payout support-matrix row.
type PayoutProfile struct {
	ID                      string                  `json:"id"`
	State                   ProfileState            `json:"state"`
	Provider                string                  `json:"provider"`
	APIVersion              string                  `json:"api_version"`
	Environment             Environment             `json:"environment"`
	Corridors               []PayoutCorridor        `json:"corridors"`
	BeneficiaryRequirements BeneficiaryRequirements `json:"beneficiary_requirements"`
	DecisionRequirements    DecisionRequirements    `json:"decision_requirements"`
	CredentialSecretRefs    []SecretReference       `json:"credential_secret_refs,omitempty"`
	Webhook                 WebhookProfile          `json:"webhook"`
	Evidence                ProfileEvidence         `json:"evidence"`
	Owners                  ProfileOwners           `json:"owners"`
}

// ProfiledAdapter exposes the exact runtime support profile used by an adapter.
type ProfiledAdapter interface {
	Profile() PayoutProfile
}

// ProfileAuthorizer verifies that an executable profile originated from the
// trusted deployment support matrix. Self-asserted approval references cannot
// promote a caller-constructed profile to production.
type ProfileAuthorizer interface {
	AuthorizePayoutProfile(profile PayoutProfile) error
}

// ComplianceAuthorizer independently resolves a recorded compliance decision.
// Request fields are references only and cannot self-assert approval.
type ComplianceAuthorizer interface {
	AuthorizePayout(ctx context.Context, decision ComplianceDecision, sender string, beneficiaryReference string, corridorID string, at time.Time) error
}

// TestOnlyAdapter marks deterministic doubles that must never enter a
// production bridge.
type TestOnlyAdapter interface {
	IsTestOnly() bool
}

var (
	profileTokenPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/@+-]{0,255}$`)
	externalIDPattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:@+-]{0,255}$`)
	secretRefPattern    = regexp.MustCompile(`^(vault|kms|secret|env)://[A-Za-z0-9][A-Za-z0-9._/@:+-]*$`)
)

// Validate validates the support row without elevating its certification state.
func (p PayoutProfile) Validate() error {
	if !profileTokenPattern.MatchString(p.ID) || !profileTokenPattern.MatchString(p.Provider) {
		return fmt.Errorf("%w: invalid profile or provider identifier", ErrInvalidRequest)
	}
	if !validExactVersion(p.APIVersion) || !validExactVersion(p.Webhook.Version) {
		return fmt.Errorf("%w: exact API and webhook versions are required", ErrInvalidRequest)
	}
	switch p.State {
	case ProfileUnsupported, ProfileEngineeringIncomplete, ProfileEngineeringCompleteExternalBlocked, ProfileCertifiedEnabled, ProfilePaused:
	default:
		return fmt.Errorf("%w: unknown profile state", ErrInvalidRequest)
	}
	if p.Environment != EnvironmentSandbox && p.Environment != EnvironmentProduction {
		return fmt.Errorf("%w: unknown environment", ErrInvalidRequest)
	}
	if len(p.Corridors) == 0 {
		return fmt.Errorf("%w: at least one corridor is required", ErrInvalidRequest)
	}
	seen := make(map[string]struct{}, len(p.Corridors))
	for _, corridor := range p.Corridors {
		if err := validateCorridor(corridor); err != nil {
			return err
		}
		key := corridor.Jurisdiction + "|" + corridor.Currency + "|" + corridor.Rail
		if _, exists := seen[key]; exists {
			return fmt.Errorf("%w: duplicate corridor", ErrInvalidRequest)
		}
		seen[key] = struct{}{}
	}
	if !p.BeneficiaryRequirements.TokenizedReferenceRequired ||
		!externalIDPattern.MatchString(p.BeneficiaryRequirements.ReferencePrefix) ||
		!slices.Contains(p.BeneficiaryRequirements.RequiredFields, "beneficiary_reference") ||
		len(p.BeneficiaryRequirements.ProhibitedRawFields) == 0 {
		return fmt.Errorf("%w: tokenized beneficiary references are mandatory", ErrInvalidRequest)
	}
	if !p.DecisionRequirements.KYCRequired || !p.DecisionRequirements.SanctionsRequired {
		return fmt.Errorf("%w: KYC and sanctions decisions are mandatory", ErrInvalidRequest)
	}
	if !strings.EqualFold(p.Webhook.Algorithm, "HMAC-SHA256") || len(p.Webhook.Keys) == 0 {
		return fmt.Errorf("%w: HMAC-SHA256 webhook keys are required", ErrInvalidRequest)
	}
	keyVersions := make(map[string]struct{}, len(p.Webhook.Keys))
	for _, key := range p.Webhook.Keys {
		if !profileTokenPattern.MatchString(key.KeyID) || !validExactVersion(key.Version) || !secretRefPattern.MatchString(key.SecretRef) {
			return fmt.Errorf("%w: invalid webhook key reference", ErrInvalidRequest)
		}
		identity := key.KeyID + "|" + key.Version
		if _, exists := keyVersions[identity]; exists {
			return fmt.Errorf("%w: duplicate webhook key version", ErrInvalidRequest)
		}
		keyVersions[identity] = struct{}{}
	}
	for _, ref := range p.CredentialSecretRefs {
		if !profileTokenPattern.MatchString(ref.Purpose) || !validExactVersion(ref.Version) || !secretRefPattern.MatchString(ref.Ref) {
			return fmt.Errorf("%w: invalid credential secret reference", ErrInvalidRequest)
		}
		if ref.Scope != string(EnvironmentSandbox) && ref.Scope != string(EnvironmentProduction) {
			return fmt.Errorf("%w: invalid credential scope", ErrInvalidRequest)
		}
	}
	if p.State == ProfileCertifiedEnabled {
		return p.validateProductionEvidence()
	}
	return nil
}

// ValidateForExecution enforces runtime state and environment without treating
// engineering conformance as production certification.
func (p PayoutProfile) ValidateForExecution(mode ExecutionMode, allowExternalBlocked bool) error {
	if err := p.Validate(); err != nil {
		return err
	}
	switch mode {
	case ExecutionModeProduction:
		if p.Environment != EnvironmentProduction || p.State != ProfileCertifiedEnabled {
			return ErrProfileNotExecutable
		}
		if _, err := findCredentialRef(p, "api", EnvironmentProduction); err != nil {
			return ErrProfileNotExecutable
		}
		return p.validateProductionEvidence()
	case ExecutionModeSandbox:
		if p.Environment != EnvironmentSandbox {
			return ErrProfileNotExecutable
		}
		if p.State == ProfileEngineeringCompleteExternalBlocked && allowExternalBlocked {
			if _, err := findCredentialRef(p, "api", EnvironmentSandbox); err != nil {
				return ErrProfileNotExecutable
			}
			return nil
		}
		return ErrProfileNotExecutable
	default:
		return ErrProfileNotExecutable
	}
}

// Corridor returns the exact supported row, failing closed on partial matches.
func (p PayoutProfile) Corridor(jurisdiction string, currency string, rail string) (PayoutCorridor, error) {
	for _, corridor := range p.Corridors {
		if corridor.Jurisdiction == jurisdiction && corridor.Currency == currency && corridor.Rail == rail {
			return corridor, nil
		}
	}
	return PayoutCorridor{}, ErrUnsupportedCorridor
}

func (p PayoutProfile) validateProductionEvidence() error {
	if p.Environment != EnvironmentProduction {
		return fmt.Errorf("%w: certified profile must select production", ErrProfileNotExecutable)
	}
	if strings.TrimSpace(p.Owners.Engineering) == "" || strings.TrimSpace(p.Owners.Operations) == "" ||
		strings.TrimSpace(p.Owners.Compliance) == "" || strings.TrimSpace(p.Owners.Security) == "" {
		return fmt.Errorf("%w: all production owners are required", ErrProfileNotExecutable)
	}
	approvals := []ApprovalEvidence{
		p.Evidence.Contract, p.Evidence.Legal, p.Evidence.DPA, p.Evidence.Compliance,
		p.Evidence.Custody, p.Evidence.Banking, p.Evidence.WebhookRegistration, p.Evidence.Corridor,
	}
	for _, approval := range approvals {
		if strings.TrimSpace(approval.Reference) == "" || strings.TrimSpace(approval.Owner) == "" {
			return fmt.Errorf("%w: complete external approval evidence is required", ErrProfileNotExecutable)
		}
	}
	if !slices.ContainsFunc(p.CredentialSecretRefs, func(ref SecretReference) bool {
		return ref.Purpose == "api" && ref.Scope == string(EnvironmentProduction)
	}) {
		return fmt.Errorf("%w: production API credential reference is required", ErrProfileNotExecutable)
	}
	for _, key := range p.Webhook.Keys {
		if key.SecretRef == "" {
			return fmt.Errorf("%w: production webhook credential reference is required", ErrProfileNotExecutable)
		}
	}
	return nil
}

func validateCorridor(c PayoutCorridor) error {
	if !profileTokenPattern.MatchString(c.ID) || !isUpperToken(c.Jurisdiction) || !isUpperToken(c.Currency) || !profileTokenPattern.MatchString(c.Rail) {
		return fmt.Errorf("%w: invalid corridor identity", ErrInvalidRequest)
	}
	if c.MinimumAmount.IsNil() || !c.MinimumAmount.IsPositive() || c.MaximumAmount.IsNil() || c.DailyLimit.IsNil() {
		return fmt.Errorf("%w: positive corridor limits are required", ErrInvalidRequest)
	}
	if c.MaximumAmount.LT(c.MinimumAmount) || c.DailyLimit.LT(c.MaximumAmount) {
		return fmt.Errorf("%w: corridor limits are inconsistent", ErrInvalidRequest)
	}
	if c.QuoteTTL <= 0 || c.QuoteTTL > 30*time.Minute || strings.TrimSpace(c.Finality) == "" {
		return fmt.Errorf("%w: quote TTL and finality are required", ErrInvalidRequest)
	}
	return nil
}

func validateCompliance(decision ComplianceDecision, requirements DecisionRequirements, now time.Time) error {
	if decision.Revoked || strings.TrimSpace(decision.Reference) == "" {
		return ErrComplianceRequired
	}
	if requirements.KYCRequired && decision.KYCDecision != "approved" {
		return ErrComplianceRequired
	}
	if requirements.SanctionsRequired && decision.SanctionsDecision != "approved" {
		return ErrComplianceRequired
	}
	if decision.ValidUntil.IsZero() || !decision.ValidUntil.After(now) {
		return ErrComplianceRequired
	}
	return nil
}

func validateAmount(c PayoutCorridor, amount sdkmath.LegacyDec) error {
	if amount.IsNil() || amount.LT(c.MinimumAmount) || amount.GT(c.MaximumAmount) {
		return ErrLimitExceeded
	}
	return nil
}

func validExactVersion(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && value != "latest" && profileTokenPattern.MatchString(value)
}

func isUpperToken(value string) bool {
	return value != "" && value == strings.ToUpper(value) && profileTokenPattern.MatchString(value)
}

func findCredentialRef(profile PayoutProfile, purpose string, environment Environment) (SecretReference, error) {
	for _, ref := range profile.CredentialSecretRefs {
		if ref.Purpose == purpose && ref.Scope == string(environment) {
			return ref, nil
		}
	}
	return SecretReference{}, errors.New("credential secret reference not configured")
}

func clonePayoutProfile(profile PayoutProfile) PayoutProfile {
	profile.Corridors = slices.Clone(profile.Corridors)
	profile.BeneficiaryRequirements.RequiredFields = slices.Clone(profile.BeneficiaryRequirements.RequiredFields)
	profile.BeneficiaryRequirements.ProhibitedRawFields = slices.Clone(profile.BeneficiaryRequirements.ProhibitedRawFields)
	profile.CredentialSecretRefs = slices.Clone(profile.CredentialSecretRefs)
	profile.Webhook.Keys = slices.Clone(profile.Webhook.Keys)
	return profile
}
