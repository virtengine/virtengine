// Package edugain provides EduGAIN federation integration.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
package edugain

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// VEID Integrator Implementation
// ============================================================================

// veidIntegrator implements the VEIDIntegrator interface
type veidIntegrator struct {
	config           Config
	scopes           map[string]*VEIDScope // scopeID -> scope
	veidAssociations map[string]string     // scopeID -> veidID
	mu               sync.RWMutex
}

// newVEIDIntegrator creates a new VEID integrator
func newVEIDIntegrator(config Config) VEIDIntegrator {
	return &veidIntegrator{
		config:           config,
		scopes:           make(map[string]*VEIDScope),
		veidAssociations: make(map[string]string),
	}
}

// CreateScope creates a VEID scope from EduGAIN session
func (v *veidIntegrator) CreateScope(ctx context.Context, session *Session) (*VEIDScope, error) {
	if session == nil {
		return nil, ErrVEIDIntegrationFailed
	}

	// Generate scope ID
	scopeID := fmt.Sprintf("edugain-%s-%d", hashString(session.WalletAddress+session.InstitutionID)[:16], time.Now().UnixNano())

	// Compute score contribution
	scoreContribution := v.ComputeScoreContribution(&session.Attributes)

	// Determine expiry based on session
	expiresAt := session.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	now := time.Now()
	scope := &VEIDScope{
		ID:                scopeID,
		WalletAddress:     session.WalletAddress,
		InstitutionID:     session.InstitutionID,
		InstitutionName:   session.InstitutionName,
		Federation:        parseFederationFromEntityID(session.InstitutionID),
		PrincipalNameHash: session.Attributes.EduPerson.PrincipalNameHash,
		HomeOrganization:  session.Attributes.Schac.HomeOrganization,
		Affiliations:      session.Attributes.EduPerson.Affiliation,
		IsMFA:             session.IsMFA,
		ScoreContribution: scoreContribution,
		AuthnInstant:      session.AuthnInstant,
		ExpiresAt:         expiresAt,
		CreatedAt:         now,
		UpdatedAt:         now,
		Status:            "active",
	}

	// Store scope
	v.scopes[scopeID] = scope

	return scope, nil
}

// EnrichIdentity enriches an existing VEID identity
func (v *veidIntegrator) EnrichIdentity(ctx context.Context, session *Session, veidID string) error {
	if session == nil || veidID == "" {
		return ErrVEIDIntegrationFailed
	}

	// In a full implementation, this would call the VEID keeper to:
	// 1. Look up the existing identity
	// 2. Add or update the EduGAIN scope
	// 3. Recalculate the identity score

	// For now, create a scope associated with this VEID
	scope, err := v.CreateScope(ctx, session)
	if err != nil {
		return err
	}

	// Store the VEID association in the scope's status metadata
	// In production, this would be stored in the VEID module's state
	v.mu.Lock()
	v.veidAssociations[scope.ID] = veidID
	v.mu.Unlock()

	return nil
}

// GetExistingScope returns an existing EduGAIN scope for a wallet
func (v *veidIntegrator) GetExistingScope(ctx context.Context, walletAddress string) (*VEIDScope, error) {
	// Find scope by wallet address
	for _, scope := range v.scopes {
		if scope.WalletAddress == walletAddress && scope.Status == "active" {
			return scope, nil
		}
	}
	return nil, fmt.Errorf("no existing scope for wallet")
}

// UpdateScope updates an existing EduGAIN scope
func (v *veidIntegrator) UpdateScope(ctx context.Context, scope *VEIDScope, session *Session) error {
	if scope == nil || session == nil {
		return ErrVEIDIntegrationFailed
	}

	// Update scope with new session data
	scope.AuthnInstant = session.AuthnInstant
	scope.IsMFA = session.IsMFA
	scope.Affiliations = session.Attributes.EduPerson.Affiliation
	scope.ScoreContribution = v.ComputeScoreContribution(&session.Attributes)
	scope.UpdatedAt = time.Now()

	// Extend expiry
	if session.ExpiresAt.After(scope.ExpiresAt) {
		scope.ExpiresAt = session.ExpiresAt
	}

	return nil
}

// RevokeScope revokes an EduGAIN scope
func (v *veidIntegrator) RevokeScope(ctx context.Context, scopeID string) error {
	scope, ok := v.scopes[scopeID]
	if !ok {
		return fmt.Errorf("scope not found: %s", scopeID)
	}

	scope.Status = "revoked"
	scope.UpdatedAt = time.Now()

	return nil
}

// ComputeScoreContribution computes identity score contribution
func (v *veidIntegrator) ComputeScoreContribution(attrs *UserAttributes) uint32 {
	if attrs == nil {
		return 0
	}

	// Base score from configuration
	baseScore := v.config.VEIDIntegration.ScoreWeight

	// Add attribute-based score
	attrScore := ComputeAttributeScore(attrs)

	// Total contribution (capped at weight)
	total := baseScore + attrScore/2
	if total > baseScore+10 {
		total = baseScore + 10
	}

	return total
}

// ============================================================================
// VEID Scope Types for On-Chain Storage
// ============================================================================

// VEIDScopeData represents the on-chain storage format for EduGAIN scope
// This would be used when integrating with the x/veid module
type VEIDScopeData struct {
	// Version is the scope data version
	Version uint32 `json:"version" protobuf:"varint,1,opt,name=version"`

	// Type is always "edugain" for this scope
	Type string `json:"type" protobuf:"bytes,2,opt,name=type"`

	// InstitutionIDHash is SHA-256 hash of the IdP entity ID
	InstitutionIDHash string `json:"institution_id_hash" protobuf:"bytes,3,opt,name=institution_id_hash"`

	// FederationHash is SHA-256 hash of the federation name
	FederationHash string `json:"federation_hash" protobuf:"bytes,4,opt,name=federation_hash"`

	// PrincipalNameHash is SHA-256 hash of eduPersonPrincipalName
	PrincipalNameHash string `json:"principal_name_hash" protobuf:"bytes,5,opt,name=principal_name_hash"`

	// HomeOrganizationHash is SHA-256 hash of schacHomeOrganization
	HomeOrganizationHash string `json:"home_organization_hash" protobuf:"bytes,6,opt,name=home_organization_hash"`

	// Affiliations are the verified affiliation types
	Affiliations []string `json:"affiliations" protobuf:"bytes,7,rep,name=affiliations"`

	// AssuranceLevels are the eduPersonAssurance values
	AssuranceLevels []string `json:"assurance_levels" protobuf:"bytes,8,rep,name=assurance_levels"`

	// IsMFA indicates if MFA (REFEDS profile) was used
	IsMFA bool `json:"is_mfa" protobuf:"varint,9,opt,name=is_mfa"`

	// AuthnInstant is when authentication occurred (Unix timestamp)
	AuthnInstant int64 `json:"authn_instant" protobuf:"varint,10,opt,name=authn_instant"`

	// ExpiresAt is when the scope expires (Unix timestamp)
	ExpiresAt int64 `json:"expires_at" protobuf:"varint,11,opt,name=expires_at"`

	// ScoreContribution is the identity score contribution
	ScoreContribution uint32 `json:"score_contribution" protobuf:"varint,12,opt,name=score_contribution"`
}

// ConvertToScopeData converts a VEIDScope to on-chain storage format
func ConvertToScopeData(scope *VEIDScope) *VEIDScopeData {
	affiliations := make([]string, len(scope.Affiliations))
	for i, aff := range scope.Affiliations {
		affiliations[i] = string(aff)
	}

	return &VEIDScopeData{
		Version:              1,
		Type:                 "edugain",
		InstitutionIDHash:    hashString(scope.InstitutionID),
		FederationHash:       hashString(scope.Federation),
		PrincipalNameHash:    scope.PrincipalNameHash,
		HomeOrganizationHash: hashString(scope.HomeOrganization),
		Affiliations:         affiliations,
		IsMFA:                scope.IsMFA,
		AuthnInstant:         scope.AuthnInstant.Unix(),
		ExpiresAt:            scope.ExpiresAt.Unix(),
		ScoreContribution:    scope.ScoreContribution,
	}
}

// ============================================================================
// VEID Integration Helpers
// ============================================================================

// ScopeTypeEduGAIN is the scope type for EduGAIN verification
const ScopeTypeEduGAIN = "edugain"

// EduGAINScopeWeight is the base weight for EduGAIN scopes
// This should match x/veid/types/scope.go ScopeTypeWeight
const EduGAINScopeWeight uint32 = 15

// ValidateForVEID validates that a session can be used to create a VEID scope
func ValidateForVEID(session *Session) error {
	if session == nil {
		return fmt.Errorf("session is nil")
	}

	if session.Status != SessionStatusActive {
		return fmt.Errorf("session is not active: %s", session.Status)
	}

	if session.WalletAddress == "" {
		return fmt.Errorf("wallet address is required")
	}

	if session.Attributes.EduPerson.PrincipalName == "" &&
		session.Attributes.EduPerson.PrincipalNameHash == "" {
		return fmt.Errorf("eduPersonPrincipalName is required")
	}

	return nil
}

// ComputeVEIDScoreContribution computes the score contribution for VEID
func ComputeVEIDScoreContribution(scope *VEIDScope) uint32 {
	if scope == nil {
		return 0
	}

	// Base contribution from EduGAIN scope
	score := EduGAINScopeWeight

	// Bonus for MFA
	if scope.IsMFA {
		score += 3
	}

	// Bonus for certain affiliations
	for _, aff := range scope.Affiliations {
		switch aff {
		case AffiliationFaculty:
			score += 2
		case AffiliationEmployee, AffiliationStaff:
			score += 1
		}
	}

	// Cap at reasonable maximum
	if score > 25 {
		score = 25
	}

	return score
}
