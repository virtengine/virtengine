// Package edugain provides EduGAIN federation integration.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
package edugain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// ============================================================================
// Attribute Mapper Implementation
// ============================================================================

// Attribute OID constants (eduPerson schema)
const (
	// eduPerson attributes
	OIDEduPersonPrincipalName      = "urn:oid:1.3.6.1.4.1.5923.1.1.1.6"
	OIDEduPersonAffiliation        = "urn:oid:1.3.6.1.4.1.5923.1.1.1.1"
	OIDEduPersonScopedAffiliation  = "urn:oid:1.3.6.1.4.1.5923.1.1.1.9"
	OIDEduPersonEntitlement        = "urn:oid:1.3.6.1.4.1.5923.1.1.1.7"
	OIDEduPersonAssurance          = "urn:oid:1.3.6.1.4.1.5923.1.1.1.11"
	OIDEduPersonTargetedID         = "urn:oid:1.3.6.1.4.1.5923.1.1.1.10"
	OIDEduPersonUniqueID           = "urn:oid:1.3.6.1.4.1.5923.1.1.1.13"
	OIDEduPersonOrcid              = "urn:oid:1.3.6.1.4.1.5923.1.1.1.16"

	// SCHAC attributes
	OIDSchacHomeOrganization       = "urn:oid:1.3.6.1.4.1.25178.1.2.9"
	OIDSchacHomeOrganizationType   = "urn:oid:1.3.6.1.4.1.25178.1.2.10"
	OIDSchacPersonalUniqueCode     = "urn:oid:1.3.6.1.4.1.25178.1.2.14"
	OIDSchacUserStatus             = "urn:oid:1.3.6.1.4.1.25178.1.2.19"
	OIDSchacCountryOfCitizenship   = "urn:oid:1.3.6.1.4.1.25178.1.2.5"

	// Standard attributes
	OIDDisplayName  = "urn:oid:2.16.840.1.113730.3.1.241"
	OIDGivenName    = "urn:oid:2.5.4.42"
	OIDSurname      = "urn:oid:2.5.4.4"
	OIDEmail        = "urn:oid:0.9.2342.19200300.100.1.3"
	OIDCommonName   = "urn:oid:2.5.4.3"
)

// Attribute friendly names
const (
	FriendlyEduPersonPrincipalName     = "eduPersonPrincipalName"
	FriendlyEduPersonAffiliation       = "eduPersonAffiliation"
	FriendlyEduPersonScopedAffiliation = "eduPersonScopedAffiliation"
	FriendlyEduPersonEntitlement       = "eduPersonEntitlement"
	FriendlyEduPersonAssurance         = "eduPersonAssurance"
	FriendlyEduPersonTargetedID        = "eduPersonTargetedID"
	FriendlyEduPersonUniqueID          = "eduPersonUniqueId"
	FriendlyEduPersonOrcid             = "eduPersonOrcid"
	FriendlySchacHomeOrganization      = "schacHomeOrganization"
	FriendlySchacHomeOrganizationType  = "schacHomeOrganizationType"
	FriendlySchacPersonalUniqueCode    = "schacPersonalUniqueCode"
	FriendlySchacUserStatus            = "schacUserStatus"
	FriendlySchacCountryOfCitizenship  = "schacCountryOfCitizenship"
	FriendlyDisplayName                = "displayName"
	FriendlyGivenName                  = "givenName"
	FriendlySurname                    = "sn"
	FriendlyEmail                      = "mail"
	FriendlyCommonName                 = "cn"
)

// attributeMapper implements the AttributeMapper interface
type attributeMapper struct {
	oidToFriendly map[string]string
	friendlyToOID map[string]string
}

// newAttributeMapper creates a new attribute mapper
func newAttributeMapper() AttributeMapper {
	m := &attributeMapper{
		oidToFriendly: make(map[string]string),
		friendlyToOID: make(map[string]string),
	}

	// Map OIDs to friendly names
	mappings := map[string]string{
		OIDEduPersonPrincipalName:     FriendlyEduPersonPrincipalName,
		OIDEduPersonAffiliation:       FriendlyEduPersonAffiliation,
		OIDEduPersonScopedAffiliation: FriendlyEduPersonScopedAffiliation,
		OIDEduPersonEntitlement:       FriendlyEduPersonEntitlement,
		OIDEduPersonAssurance:         FriendlyEduPersonAssurance,
		OIDEduPersonTargetedID:        FriendlyEduPersonTargetedID,
		OIDEduPersonUniqueID:          FriendlyEduPersonUniqueID,
		OIDEduPersonOrcid:             FriendlyEduPersonOrcid,
		OIDSchacHomeOrganization:      FriendlySchacHomeOrganization,
		OIDSchacHomeOrganizationType:  FriendlySchacHomeOrganizationType,
		OIDSchacPersonalUniqueCode:    FriendlySchacPersonalUniqueCode,
		OIDSchacUserStatus:            FriendlySchacUserStatus,
		OIDSchacCountryOfCitizenship:  FriendlySchacCountryOfCitizenship,
		OIDDisplayName:                FriendlyDisplayName,
		OIDGivenName:                  FriendlyGivenName,
		OIDSurname:                    FriendlySurname,
		OIDEmail:                      FriendlyEmail,
		OIDCommonName:                 FriendlyCommonName,
	}

	for oid, friendly := range mappings {
		m.oidToFriendly[oid] = friendly
		m.friendlyToOID[friendly] = oid
	}

	return m
}

// MapAttributes maps raw SAML attributes to UserAttributes
func (m *attributeMapper) MapAttributes(rawAttrs map[string][]string) (*UserAttributes, error) {
	attrs := &UserAttributes{
		Raw: rawAttrs,
	}

	// Normalize attribute names (OID to friendly)
	normalized := m.normalizeAttributes(rawAttrs)

	// Map eduPerson attributes
	if val := m.getFirst(normalized, FriendlyEduPersonPrincipalName); val != "" {
		attrs.EduPerson.PrincipalName = val
		attrs.EduPerson.EPPN = val
	}

	if vals := m.getAll(normalized, FriendlyEduPersonAffiliation); len(vals) > 0 {
		attrs.EduPerson.Affiliation = m.parseAffiliations(vals)
	}

	if vals := m.getAll(normalized, FriendlyEduPersonScopedAffiliation); len(vals) > 0 {
		attrs.EduPerson.ScopedAffiliation = vals
	}

	if vals := m.getAll(normalized, FriendlyEduPersonEntitlement); len(vals) > 0 {
		attrs.EduPerson.Entitlement = vals
	}

	if vals := m.getAll(normalized, FriendlyEduPersonAssurance); len(vals) > 0 {
		attrs.EduPerson.Assurance = vals
	}

	if val := m.getFirst(normalized, FriendlyEduPersonTargetedID); val != "" {
		attrs.EduPerson.TargetedID = val
	}

	if val := m.getFirst(normalized, FriendlyEduPersonUniqueID); val != "" {
		attrs.EduPerson.UniqueID = val
	}

	if val := m.getFirst(normalized, FriendlyEduPersonOrcid); val != "" {
		attrs.EduPerson.OrcidID = val
	}

	// Map SCHAC attributes
	if val := m.getFirst(normalized, FriendlySchacHomeOrganization); val != "" {
		attrs.Schac.HomeOrganization = val
	}

	if val := m.getFirst(normalized, FriendlySchacHomeOrganizationType); val != "" {
		attrs.Schac.HomeOrganizationType = val
	}

	if val := m.getFirst(normalized, FriendlySchacPersonalUniqueCode); val != "" {
		attrs.Schac.PersonalUniqueCode = val
	}

	if val := m.getFirst(normalized, FriendlySchacUserStatus); val != "" {
		attrs.Schac.UserStatus = val
	}

	if vals := m.getAll(normalized, FriendlySchacCountryOfCitizenship); len(vals) > 0 {
		attrs.Schac.CountryOfCitizenship = vals
	}

	// Map standard attributes
	if val := m.getFirst(normalized, FriendlyDisplayName); val != "" {
		attrs.DisplayName = val
	}

	if val := m.getFirst(normalized, FriendlyGivenName); val != "" {
		attrs.GivenName = val
	}

	if val := m.getFirst(normalized, FriendlySurname); val != "" {
		attrs.Surname = val
	}

	if val := m.getFirst(normalized, FriendlyEmail); val != "" {
		attrs.Email = val
	}

	if val := m.getFirst(normalized, FriendlyCommonName); val != "" {
		attrs.CommonName = val
	}

	return attrs, nil
}

// GetRequiredAttributes returns required attribute OIDs/names
func (m *attributeMapper) GetRequiredAttributes() []string {
	return []string{
		OIDEduPersonPrincipalName, // Primary identifier
	}
}

// GetOptionalAttributes returns optional attribute OIDs/names
func (m *attributeMapper) GetOptionalAttributes() []string {
	return []string{
		OIDEduPersonAffiliation,
		OIDEduPersonScopedAffiliation,
		OIDEduPersonEntitlement,
		OIDEduPersonTargetedID,
		OIDSchacHomeOrganization,
		OIDDisplayName,
		OIDGivenName,
		OIDSurname,
		OIDEmail,
	}
}

// ValidateAttributes validates required attributes are present
func (m *attributeMapper) ValidateAttributes(attrs *UserAttributes) error {
	// eduPersonPrincipalName is required
	if attrs.EduPerson.PrincipalName == "" {
		return ErrMissingRequiredAttribute
	}

	// Validate scope in principal name
	if !strings.Contains(attrs.EduPerson.PrincipalName, "@") {
		return ErrMissingRequiredAttribute
	}

	return nil
}

// HashSensitiveData hashes PII fields
func (m *attributeMapper) HashSensitiveData(attrs *UserAttributes) *UserAttributes {
	// Create a copy to avoid modifying original
	hashed := *attrs
	hashed.EduPerson = attrs.EduPerson
	hashed.Schac = attrs.Schac

	// Hash eduPerson fields
	if attrs.EduPerson.PrincipalName != "" {
		hashed.EduPerson.PrincipalNameHash = hashString(attrs.EduPerson.PrincipalName)
	}
	if attrs.EduPerson.EPPN != "" {
		hashed.EduPerson.EPPNHash = hashString(attrs.EduPerson.EPPN)
	}
	if attrs.EduPerson.TargetedID != "" {
		hashed.EduPerson.TargetedIDHash = hashString(attrs.EduPerson.TargetedID)
	}

	// Hash SCHAC fields
	if attrs.Schac.PersonalUniqueCode != "" {
		hashed.Schac.PersonalUniqueCodeHash = hashString(attrs.Schac.PersonalUniqueCode)
	}

	// Hash email
	if attrs.Email != "" {
		hashed.EmailHash = hashString(attrs.Email)
	}

	// Clear plaintext sensitive fields (optional, depending on policy)
	// hashed.Email = ""
	// hashed.EduPerson.PrincipalName = ""

	return &hashed
}

// ExtractAffiliations extracts affiliation types from attributes
func (m *attributeMapper) ExtractAffiliations(attrs *UserAttributes) []AffiliationType {
	return attrs.EduPerson.Affiliation
}

// normalizeAttributes normalizes attribute names to friendly names
func (m *attributeMapper) normalizeAttributes(rawAttrs map[string][]string) map[string][]string {
	normalized := make(map[string][]string)

	for name, values := range rawAttrs {
		// Try to find friendly name
		if friendly, ok := m.oidToFriendly[name]; ok {
			normalized[friendly] = values
		} else {
			// Keep original name (might already be friendly)
			normalized[name] = values
		}
	}

	return normalized
}

// getFirst gets the first value for an attribute
func (m *attributeMapper) getFirst(attrs map[string][]string, name string) string {
	if vals, ok := attrs[name]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// getAll gets all values for an attribute
func (m *attributeMapper) getAll(attrs map[string][]string, name string) []string {
	if vals, ok := attrs[name]; ok {
		return vals
	}
	return nil
}

// parseAffiliations parses affiliation strings to types
func (m *attributeMapper) parseAffiliations(values []string) []AffiliationType {
	var affiliations []AffiliationType

	for _, val := range values {
		// Handle scoped affiliations (e.g., "student@university.edu")
		val = strings.Split(val, "@")[0]
		val = strings.ToLower(strings.TrimSpace(val))

		var aff AffiliationType
		switch val {
		case "student":
			aff = AffiliationStudent
		case "faculty":
			aff = AffiliationFaculty
		case "staff":
			aff = AffiliationStaff
		case "employee":
			aff = AffiliationEmployee
		case "member":
			aff = AffiliationMember
		case "affiliate":
			aff = AffiliationAffiliate
		case "alum":
			aff = AffiliationAlum
		case "library-walk-in":
			aff = AffiliationLibraryWalkIn
		default:
			continue // Skip unknown affiliations
		}

		// Avoid duplicates
		found := false
		for _, existing := range affiliations {
			if existing == aff {
				found = true
				break
			}
		}
		if !found {
			affiliations = append(affiliations, aff)
		}
	}

	return affiliations
}

// ============================================================================
// Attribute Score Computation
// ============================================================================

// ComputeAttributeScore computes an identity score contribution from attributes
func ComputeAttributeScore(attrs *UserAttributes) uint32 {
	var score uint32

	// Base score for having verified eduPerson principal name
	if attrs.EduPerson.PrincipalName != "" {
		score += 5
	}

	// Additional points for affiliations
	for _, aff := range attrs.EduPerson.Affiliation {
		switch aff {
		case AffiliationFaculty:
			score += 4 // Faculty/researchers get higher trust
		case AffiliationEmployee, AffiliationStaff:
			score += 3
		case AffiliationStudent:
			score += 2
		case AffiliationMember, AffiliationAffiliate:
			score += 1
		case AffiliationAlum:
			score += 1
		}
	}

	// Cap at maximum affiliation contribution
	if score > 15 {
		score = 15
	}

	// Points for verified organization
	if attrs.Schac.HomeOrganization != "" {
		score += 2
	}

	// Points for assurance level
	for _, assurance := range attrs.EduPerson.Assurance {
		if strings.Contains(assurance, "IAP/high") {
			score += 3
		} else if strings.Contains(assurance, "IAP/medium") {
			score += 2
		} else if strings.Contains(assurance, "IAP/low") {
			score += 1
		}
	}

	// Points for verified email (implied by EPPN scope)
	if attrs.Email != "" {
		score += 1
	}

	// Cap total score contribution
	if score > 20 {
		score = 20
	}

	return score
}

// ============================================================================
// Helper Functions
// ============================================================================

// hashString hashes a string using SHA-256
func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
