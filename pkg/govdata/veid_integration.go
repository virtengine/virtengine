// Package govdata provides government data source integration for identity verification.
//
// VE-909: Government data integration for authoritative identity verification
package govdata

import (
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
	config           VEIDIntegrationConfig
	scopes           map[string]*VEIDScope // scopeID -> scope
	walletScopes     map[string][]string   // walletAddress -> scopeIDs
	veidAssociations map[string]string     // scopeID -> veidID
	mu               sync.RWMutex
}

// newVEIDIntegrator creates a new VEID integrator
func newVEIDIntegrator(config VEIDIntegrationConfig) VEIDIntegrator {
	return &veidIntegrator{
		config:           config,
		scopes:           make(map[string]*VEIDScope),
		walletScopes:     make(map[string][]string),
		veidAssociations: make(map[string]string),
	}
}

// CreateScope creates a VEID scope from government verification
func (v *veidIntegrator) CreateScope(ctx context.Context, verification *VerificationResponse) (*VEIDScope, error) {
	if verification == nil {
		return nil, ErrVEIDIntegrationFailed
	}

	// Only create scopes for successful verifications
	if !verification.Status.IsSuccess() {
		return nil, fmt.Errorf("cannot create scope from unsuccessful verification: %s", verification.Status)
	}

	// Check confidence threshold
	if verification.Confidence < v.config.MinConfidenceThreshold {
		return nil, fmt.Errorf("verification confidence %.2f below threshold %.2f",
			verification.Confidence, v.config.MinConfidenceThreshold)
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	scopeID := generateScopeID()
	now := time.Now()

	// Extract verified field names
	fieldsVerified := make([]string, 0, len(verification.FieldResults))
	for fieldName, result := range verification.FieldResults {
		if result.Match == FieldMatchExact || result.Match == FieldMatchFuzzy {
			fieldsVerified = append(fieldsVerified, fieldName)
		}
	}

	// Compute score contribution
	scoreContribution := v.ComputeScoreContribution(verification)

	// Determine expiry
	expiresAt := now.Add(v.config.ScopeExpiryDuration)
	if verification.DocumentExpiresAt != nil && verification.DocumentExpiresAt.Before(expiresAt) {
		expiresAt = *verification.DocumentExpiresAt
	}

	scope := &VEIDScope{
		ID:                 scopeID,
		DocumentType:       getDocumentTypeFromDataSource(verification.DataSourceType),
		Jurisdiction:       verification.Jurisdiction,
		DataSource:         verification.DataSourceType,
		VerificationStatus: verification.Status,
		Confidence:         verification.Confidence,
		ScoreContribution:  scoreContribution,
		FieldsVerified:     fieldsVerified,
		VerifiedAt:         verification.VerifiedAt,
		ExpiresAt:          expiresAt,
		CreatedAt:          now,
		UpdatedAt:          now,
		Status:             "active",
	}

	v.scopes[scopeID] = scope

	return scope, nil
}

// EnrichIdentity enriches an existing VEID identity with government verification
func (v *veidIntegrator) EnrichIdentity(ctx context.Context, verification *VerificationResponse, veidID string) error {
	if verification == nil || veidID == "" {
		return ErrVEIDIntegrationFailed
	}

	// Create scope
	scope, err := v.CreateScope(ctx, verification)
	if err != nil {
		return err
	}

	// Associate scope with VEID
	v.mu.Lock()
	v.veidAssociations[scope.ID] = veidID
	v.mu.Unlock()

	// In a full implementation, this would:
	// 1. Call the VEID keeper to lookup the existing identity
	// 2. Add the government verification scope
	// 3. Recalculate the identity score with government source bonus

	return nil
}

// GetScopes retrieves VEID scopes for a wallet
func (v *veidIntegrator) GetScopes(ctx context.Context, walletAddress string) ([]VEIDScope, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	scopeIDs, ok := v.walletScopes[walletAddress]
	if !ok {
		return []VEIDScope{}, nil
	}

	var scopes []VEIDScope
	for _, id := range scopeIDs {
		if scope, ok := v.scopes[id]; ok {
			scopes = append(scopes, *scope)
		}
	}

	return scopes, nil
}

// RevokeScope revokes a VEID scope
func (v *veidIntegrator) RevokeScope(ctx context.Context, scopeID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	scope, ok := v.scopes[scopeID]
	if !ok {
		return fmt.Errorf("scope not found: %s", scopeID)
	}

	scope.Status = "revoked"
	scope.UpdatedAt = time.Now()

	return nil
}

// ComputeScoreContribution computes the VEID score contribution from verification
func (v *veidIntegrator) ComputeScoreContribution(verification *VerificationResponse) float64 {
	if verification == nil || !verification.Status.IsSuccess() {
		return 0.0
	}

	// Start with base contribution
	contribution := v.config.BaseScoreContribution

	// Apply confidence factor
	contribution *= verification.Confidence

	// Apply government source weight (government sources are highly trusted)
	contribution *= v.config.GovernmentSourceWeight

	// Bonus for multi-field verification
	if len(verification.FieldResults) > 3 {
		contribution += 0.05
	}

	// Bonus for document validity confirmation
	if verification.DocumentValid {
		contribution += 0.05
	}

	// Apply data source specific weights
	switch verification.DataSourceType {
	case DataSourcePassport:
		contribution *= 1.2 // Passports are highly trusted
	case DataSourceNationalRegistry:
		contribution *= 1.15
	case DataSourceDMV:
		contribution *= 1.1
	case DataSourceTaxAuthority:
		contribution *= 1.1
	case DataSourceVitalRecords:
		contribution *= 1.05
	case DataSourceImmigration:
		contribution *= 1.0
	}

	// Cap contribution at 0.5 (50% of total score from single verification)
	if contribution > 0.5 {
		contribution = 0.5
	}

	return contribution
}

// GetScopeStats returns scope statistics for a wallet
func (v *veidIntegrator) GetScopeStats(ctx context.Context, walletAddress string) (*VEIDScopeStats, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	stats := &VEIDScopeStats{
		DocumentTypesVerified:  []DocumentType{},
		JurisdictionsVerified: []string{},
	}

	scopeIDs, ok := v.walletScopes[walletAddress]
	if !ok {
		return stats, nil
	}

	now := time.Now()
	docTypeSet := make(map[DocumentType]bool)
	jurisdictionSet := make(map[string]bool)
	var lastVerification time.Time

	for _, id := range scopeIDs {
		scope, ok := v.scopes[id]
		if !ok {
			continue
		}

		stats.TotalScopes++

		if scope.Status == "active" && scope.ExpiresAt.After(now) {
			stats.ActiveScopes++
			stats.TotalScoreContribution += scope.ScoreContribution
		} else {
			stats.ExpiredScopes++
		}

		docTypeSet[scope.DocumentType] = true
		jurisdictionSet[scope.Jurisdiction] = true

		if scope.VerifiedAt.After(lastVerification) {
			lastVerification = scope.VerifiedAt
		}
	}

	// Convert sets to slices
	for dt := range docTypeSet {
		stats.DocumentTypesVerified = append(stats.DocumentTypesVerified, dt)
	}
	for j := range jurisdictionSet {
		stats.JurisdictionsVerified = append(stats.JurisdictionsVerified, j)
	}

	if !lastVerification.IsZero() {
		stats.LastVerificationAt = &lastVerification
	}

	return stats, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// getDocumentTypeFromDataSource maps data source to document type
func getDocumentTypeFromDataSource(ds DataSourceType) DocumentType {
	switch ds {
	case DataSourceDMV:
		return DocumentTypeDriversLicense
	case DataSourcePassport:
		return DocumentTypePassport
	case DataSourceVitalRecords:
		return DocumentTypeBirthCertificate
	case DataSourceNationalRegistry:
		return DocumentTypeNationalID
	case DataSourceTaxAuthority:
		return DocumentTypeTaxID
	case DataSourceImmigration:
		return DocumentTypeResidencePermit
	default:
		return DocumentTypeStateID
	}
}

// ============================================================================
// Score Computation with Multi-Source Verification
// ============================================================================

// ComputeMultiSourceScore computes combined score from multiple government verifications
func (v *veidIntegrator) ComputeMultiSourceScore(ctx context.Context, walletAddress string) (float64, error) {
	scopes, err := v.GetScopes(ctx, walletAddress)
	if err != nil {
		return 0.0, err
	}

	if len(scopes) == 0 {
		return 0.0, nil
	}

	now := time.Now()
	var totalScore float64
	activeCount := 0
	jurisdictions := make(map[string]bool)
	dataSources := make(map[DataSourceType]bool)

	for _, scope := range scopes {
		// Skip expired or revoked scopes
		if scope.Status != "active" || scope.ExpiresAt.Before(now) {
			continue
		}

		// Apply freshness decay
		age := now.Sub(scope.VerifiedAt)
		freshnessFactor := 1.0
		if age > v.config.VerificationFreshnessDecay {
			// Linear decay after decay period
			decayRatio := float64(age-v.config.VerificationFreshnessDecay) / float64(v.config.VerificationFreshnessDecay)
			freshnessFactor = 1.0 - (decayRatio * 0.5) // Max 50% decay
			if freshnessFactor < 0.5 {
				freshnessFactor = 0.5
			}
		}

		totalScore += scope.ScoreContribution * freshnessFactor
		activeCount++
		jurisdictions[scope.Jurisdiction] = true
		dataSources[scope.DataSource] = true
	}

	// Apply multi-source bonus
	if len(dataSources) > 1 {
		totalScore += v.config.MultiSourceBonus * float64(len(dataSources)-1)
	}

	// Apply multi-jurisdiction bonus
	if len(jurisdictions) > 1 {
		totalScore += 0.05 * float64(len(jurisdictions)-1)
	}

	// Cap total score at 1.0
	if totalScore > 1.0 {
		totalScore = 1.0
	}

	return totalScore, nil
}

// ============================================================================
// Verification Status Mapping to VEID
// ============================================================================

// MapToVEIDVerificationLevel maps government verification to VEID level
func MapToVEIDVerificationLevel(verification *VerificationResponse) string {
	if verification == nil {
		return "none"
	}

	switch verification.Status {
	case VerificationStatusVerified:
		if verification.Confidence >= 0.95 {
			return "government_verified_high"
		} else if verification.Confidence >= 0.8 {
			return "government_verified_medium"
		}
		return "government_verified_low"

	case VerificationStatusPartialMatch:
		return "government_partial"

	case VerificationStatusExpired:
		return "government_expired"

	case VerificationStatusRevoked:
		return "government_revoked"

	default:
		return "unverified"
	}
}

// GetVerificationTrustLevel returns a numeric trust level from verification
func GetVerificationTrustLevel(verification *VerificationResponse) int {
	if verification == nil || !verification.Status.IsSuccess() {
		return 0
	}

	// Base level from confidence
	level := int(verification.Confidence * 100)

	// Bonus for document validity
	if verification.DocumentValid {
		level += 10
	}

	// Bonus for data source type
	switch verification.DataSourceType {
	case DataSourcePassport:
		level += 20
	case DataSourceNationalRegistry:
		level += 15
	case DataSourceDMV:
		level += 10
	case DataSourceTaxAuthority:
		level += 10
	default:
		level += 5
	}

	// Cap at 100
	if level > 100 {
		level = 100
	}

	return level
}
