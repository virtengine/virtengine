// Package edugain provides EduGAIN federation integration.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
package edugain

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// Metadata Service Implementation
// ============================================================================

// metadataService implements the MetadataService interface
type metadataService struct {
	config          Config
	metadata        *FederationMetadata
	stats           *FederationStats
	lastRefresh     time.Time
	nextRefresh     time.Time
	lastError       error
	signingCert     []byte
	refreshCallback MetadataRefreshCallback
	httpClient      *http.Client
	mu              sync.RWMutex
}

// newMetadataService creates a new metadata service
func newMetadataService(config Config) MetadataService {
	return &metadataService{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Load loads metadata from the configured source
func (m *metadataService) Load(ctx context.Context) error {
	return m.ForceRefresh(ctx)
}

// Refresh refreshes metadata (respects cache settings)
func (m *metadataService) Refresh(ctx context.Context) error {
	m.mu.RLock()
	needsRefresh := m.metadata == nil || time.Now().After(m.nextRefresh)
	m.mu.RUnlock()

	if !needsRefresh && m.config.MetadataCache.StaleWhileRevalidate {
		// Async refresh in background
		go func() {
			_ = m.doRefresh(context.Background())
		}()
		return nil
	}

	if needsRefresh {
		return m.doRefresh(ctx)
	}

	return nil
}

// ForceRefresh forces a metadata refresh ignoring cache
func (m *metadataService) ForceRefresh(ctx context.Context) error {
	return m.doRefresh(ctx)
}

// doRefresh performs the actual metadata refresh
func (m *metadataService) doRefresh(ctx context.Context) error {
	m.mu.Lock()
	m.metadata = &FederationMetadata{Status: FederationStatusLoading}
	m.mu.Unlock()

	// Fetch metadata
	data, err := m.fetchMetadata(ctx)
	if err != nil {
		m.mu.Lock()
		m.lastError = err
		m.metadata = &FederationMetadata{Status: FederationStatusError}
		m.mu.Unlock()

		if m.refreshCallback != nil {
			m.refreshCallback(nil, err)
		}
		return fmt.Errorf("failed to fetch metadata: %w", err)
	}

	// Parse metadata
	metadata, err := m.parseMetadata(data)
	if err != nil {
		m.mu.Lock()
		m.lastError = err
		m.metadata = &FederationMetadata{Status: FederationStatusError}
		m.mu.Unlock()

		if m.refreshCallback != nil {
			m.refreshCallback(nil, err)
		}
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Verify signature if certificate is configured
	if len(m.signingCert) > 0 {
		valid, err := m.verifySignature(data)
		if err != nil {
			m.mu.Lock()
			m.lastError = err
			m.metadata = &FederationMetadata{Status: FederationStatusError}
			m.mu.Unlock()

			if m.refreshCallback != nil {
				m.refreshCallback(nil, err)
			}
			return fmt.Errorf("failed to verify signature: %w", err)
		}
		metadata.SignatureValid = valid
		if !valid {
			return ErrInvalidMetadataSignature
		}
	}

	// Apply filters
	metadata = m.applyFilters(metadata)

	// Compute statistics
	stats := m.computeStats(metadata)

	// Update state
	m.mu.Lock()
	m.metadata = metadata
	m.stats = stats
	m.lastRefresh = time.Now()
	m.nextRefresh = time.Now().Add(m.config.MetadataRefreshInterval)
	m.lastError = nil
	m.mu.Unlock()

	if m.refreshCallback != nil {
		m.refreshCallback(metadata, nil)
	}

	return nil
}

// Get returns the current metadata
func (m *metadataService) Get() (*FederationMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.metadata == nil {
		return nil, ErrMetadataNotLoaded
	}

	if m.metadata.Status != FederationStatusActive {
		return nil, fmt.Errorf("metadata not active: %s", m.metadata.Status)
	}

	if !m.metadata.IsValid() {
		return nil, ErrMetadataExpired
	}

	return m.metadata, nil
}

// GetStatus returns the metadata status
func (m *metadataService) GetStatus() MetadataStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := MetadataStatus{
		Status:      FederationStatusUnknown,
		LastRefresh: m.lastRefresh,
		NextRefresh: m.nextRefresh,
	}

	if m.metadata != nil {
		status.Status = m.metadata.Status
		status.ValidUntil = m.metadata.ValidUntil
		status.InstitutionCount = len(m.metadata.Institutions)
		status.SignatureValid = m.metadata.SignatureValid
	}

	if m.lastError != nil {
		status.LastError = m.lastError.Error()
	}

	return status
}

// IsValid returns true if metadata is valid
func (m *metadataService) IsValid() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metadata != nil && m.metadata.IsValid()
}

// FindInstitution finds an institution by entity ID
func (m *metadataService) FindInstitution(entityID string) (*Institution, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.metadata == nil {
		return nil, ErrMetadataNotLoaded
	}

	return m.metadata.FindInstitution(entityID)
}

// SearchInstitutions searches institutions
func (m *metadataService) SearchInstitutions(query InstitutionSearchQuery) (*InstitutionSearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.metadata == nil {
		return nil, ErrMetadataNotLoaded
	}

	var matches []Institution
	for _, inst := range m.metadata.Institutions {
		if m.matchesQuery(&inst, query) {
			matches = append(matches, inst)
		}
	}

	// Apply limit and offset
	totalCount := len(matches)
	if query.Offset > 0 && query.Offset < len(matches) {
		matches = matches[query.Offset:]
	}
	if query.Limit > 0 && query.Limit < len(matches) {
		matches = matches[:query.Limit]
	}

	return &InstitutionSearchResult{
		Institutions: matches,
		TotalCount:   totalCount,
		Query:        query,
	}, nil
}

// GetStats returns federation statistics
func (m *metadataService) GetStats() *FederationStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// SetCertificate sets the certificate for signature verification
func (m *metadataService) SetCertificate(certPEM []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signingCert = certPEM
	return nil
}

// OnRefresh registers a callback for refresh events
func (m *metadataService) OnRefresh(callback MetadataRefreshCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshCallback = callback
}

// fetchMetadata fetches metadata from the configured URL
func (m *metadataService) fetchMetadata(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.config.MetadataURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/xml, application/samlmetadata+xml")
	req.Header.Set("User-Agent", "VirtEngine-EduGAIN/"+EduGAINVersion)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// parseMetadata parses SAML metadata XML
func (m *metadataService) parseMetadata(data []byte) (*FederationMetadata, error) {
	// Parse the entities descriptor
	var entities EntitiesDescriptor
	if err := xml.Unmarshal(data, &entities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	metadata := &FederationMetadata{
		FederationID:  entities.Name,
		Name:          "EduGAIN",
		ValidUntil:    entities.ValidUntil,
		CacheDuration: entities.CacheDuration,
		LastRefresh:   time.Now(),
		Status:        FederationStatusActive,
	}

	// Parse each entity descriptor
	for _, entity := range entities.EntityDescriptors {
		// Only process IdPs (entities with IDPSSODescriptor)
		if entity.IDPSSODescriptor == nil {
			continue
		}

		inst, err := m.parseIdPDescriptor(entity)
		if err != nil {
			// Skip invalid IdPs but continue processing
			continue
		}

		metadata.Institutions = append(metadata.Institutions, *inst)
	}

	metadata.TotalEntities = len(metadata.Institutions)

	return metadata, nil
}

// parseIdPDescriptor parses a single IdP entity descriptor
func (m *metadataService) parseIdPDescriptor(entity EntityDescriptor) (*Institution, error) {
	idp := entity.IDPSSODescriptor

	inst := &Institution{
		EntityID:           entity.EntityID,
		SupportedBindings:  make([]string, 0),
		SSOEndpoints:       make(map[string]string),
		SLOEndpoints:       make(map[string]string),
		Certificates:       make([]string, 0),
		NameIDFormats:      make([]string, 0),
		LastUpdated:        time.Now(),
		MetadataValidUntil: entity.ValidUntil,
	}

	// Parse organization info
	if entity.Organization != nil {
		inst.DisplayName = entity.Organization.DisplayName
		inst.InformationURL = entity.Organization.URL
	}

	// Parse SSO endpoints
	for _, sso := range idp.SingleSignOnServices {
		inst.SSOEndpoints[sso.Binding] = sso.Location
		if !containsString(inst.SupportedBindings, sso.Binding) {
			inst.SupportedBindings = append(inst.SupportedBindings, sso.Binding)
		}
	}

	// Parse SLO endpoints
	for _, slo := range idp.SingleLogoutServices {
		inst.SLOEndpoints[slo.Binding] = slo.Location
	}

	// Parse certificates
	for _, keyDesc := range idp.KeyDescriptors {
		if keyDesc.Use == "signing" || keyDesc.Use == "" {
			if keyDesc.KeyInfo.X509Data != nil {
				inst.Certificates = append(inst.Certificates, keyDesc.KeyInfo.X509Data.Certificate)
			}
		}
	}

	// Parse NameID formats
	inst.NameIDFormats = idp.NameIDFormats

	// Parse extensions for additional info
	if entity.Extensions != nil {
		inst.LogoURL = entity.Extensions.Logo
		inst.PrivacyStatementURL = entity.Extensions.PrivacyStatementURL
		inst.Type = parseInstitutionType(entity.Extensions.EntityCategory)
		inst.Country = entity.Extensions.RegistrationAuthority
		inst.SupportsMFA = containsString(entity.Extensions.EntityAttributes, AuthnContextMFA)
	}

	// Determine federation from registration authority or entity ID
	inst.Federation = parseFederationFromEntityID(inst.EntityID)

	// Default institution type if not set
	if inst.Type == "" {
		inst.Type = InstitutionTypeOther
	}

	return inst, nil
}

// verifySignature verifies the metadata XML signature
func (m *metadataService) verifySignature(data []byte) (bool, error) {
	// This would implement XML-DSig verification
	// For production, use a proper XML signature library
	// Simplified implementation for now
	if len(m.signingCert) == 0 {
		return false, fmt.Errorf("no signing certificate configured")
	}

	// Verify signature using the certificate
	return verifyXMLSignature(data, m.signingCert)
}

// applyFilters applies configuration filters to metadata
func (m *metadataService) applyFilters(metadata *FederationMetadata) *FederationMetadata {
	var filtered []Institution

	for _, inst := range metadata.Institutions {
		// Check institution allowlist/blocklist
		if !m.config.IsInstitutionAllowed(inst.EntityID) {
			continue
		}

		// Check federation trust
		if !m.config.IsFederationTrusted(inst.Federation) {
			continue
		}

		filtered = append(filtered, inst)
	}

	metadata.Institutions = filtered
	metadata.TotalEntities = len(filtered)

	return metadata
}

// computeStats computes federation statistics
func (m *metadataService) computeStats(metadata *FederationMetadata) *FederationStats {
	stats := &FederationStats{
		TotalInstitutions: len(metadata.Institutions),
		ByCountry:         make(map[string]int),
		ByFederation:      make(map[string]int),
		ByType:            make(map[InstitutionType]int),
		LastRefresh:       time.Now(),
	}

	for _, inst := range metadata.Institutions {
		stats.ByCountry[inst.Country]++
		stats.ByFederation[inst.Federation]++
		stats.ByType[inst.Type]++

		if inst.SupportsMFA {
			stats.MFACapableCount++
		}
	}

	return stats
}

// matchesQuery checks if an institution matches a search query
func (m *metadataService) matchesQuery(inst *Institution, query InstitutionSearchQuery) bool {
	// Text search
	if query.Query != "" {
		queryLower := strings.ToLower(query.Query)
		if !strings.Contains(strings.ToLower(inst.DisplayName), queryLower) &&
			!strings.Contains(strings.ToLower(inst.EntityID), queryLower) &&
			!strings.Contains(strings.ToLower(inst.Description), queryLower) {
			return false
		}
	}

	// Country filter
	if query.Country != "" && inst.Country != query.Country {
		return false
	}

	// Federation filter
	if query.Federation != "" && inst.Federation != query.Federation {
		return false
	}

	// Type filter
	if query.Type != "" && inst.Type != query.Type {
		return false
	}

	// MFA filter
	if query.SupportsMFA != nil && inst.SupportsMFA != *query.SupportsMFA {
		return false
	}

	return true
}

// ============================================================================
// XML Types for SAML Metadata
// ============================================================================

// EntitiesDescriptor represents the root element of SAML metadata
type EntitiesDescriptor struct {
	XMLName           xml.Name           `xml:"EntitiesDescriptor"`
	Name              string             `xml:"Name,attr"`
	ValidUntil        time.Time          `xml:"validUntil,attr"`
	CacheDuration     time.Duration      `xml:"cacheDuration,attr"`
	EntityDescriptors []EntityDescriptor `xml:"EntityDescriptor"`
}

// EntityDescriptor represents a single entity in SAML metadata
type EntityDescriptor struct {
	XMLName          xml.Name          `xml:"EntityDescriptor"`
	EntityID         string            `xml:"entityID,attr"`
	ValidUntil       time.Time         `xml:"validUntil,attr"`
	IDPSSODescriptor *IDPSSODescriptor `xml:"IDPSSODescriptor"`
	Organization     *Organization     `xml:"Organization"`
	Extensions       *EntityExtensions `xml:"Extensions"`
}

// IDPSSODescriptor represents an IdP descriptor
type IDPSSODescriptor struct {
	XMLName              xml.Name             `xml:"IDPSSODescriptor"`
	WantAuthnRequestsSigned bool             `xml:"WantAuthnRequestsSigned,attr"`
	ProtocolSupportEnumeration string        `xml:"protocolSupportEnumeration,attr"`
	SingleSignOnServices []SingleSignOnService `xml:"SingleSignOnService"`
	SingleLogoutServices []SingleLogoutService `xml:"SingleLogoutService"`
	KeyDescriptors       []KeyDescriptor       `xml:"KeyDescriptor"`
	NameIDFormats        []string              `xml:"NameIDFormat"`
	Attributes           []Attribute           `xml:"Attribute"`
}

// SingleSignOnService represents an SSO endpoint
type SingleSignOnService struct {
	Binding  string `xml:"Binding,attr"`
	Location string `xml:"Location,attr"`
}

// SingleLogoutService represents an SLO endpoint
type SingleLogoutService struct {
	Binding  string `xml:"Binding,attr"`
	Location string `xml:"Location,attr"`
}

// KeyDescriptor represents a key descriptor
type KeyDescriptor struct {
	Use     string  `xml:"use,attr"`
	KeyInfo KeyInfo `xml:"KeyInfo"`
}

// KeyInfo contains key information
type KeyInfo struct {
	X509Data *X509Data `xml:"X509Data"`
}

// X509Data contains X509 certificate data
type X509Data struct {
	Certificate string `xml:"X509Certificate"`
}

// Attribute represents a SAML attribute
type Attribute struct {
	Name         string `xml:"Name,attr"`
	FriendlyName string `xml:"FriendlyName,attr"`
}

// Organization represents organization info
type Organization struct {
	DisplayName string `xml:"OrganizationDisplayName"`
	Name        string `xml:"OrganizationName"`
	URL         string `xml:"OrganizationURL"`
}

// EntityExtensions represents entity extensions
type EntityExtensions struct {
	Logo                  string   `xml:"Logo"`
	PrivacyStatementURL   string   `xml:"PrivacyStatementURL"`
	EntityCategory        []string `xml:"EntityCategory"`
	EntityAttributes      []string `xml:"EntityAttributes>Attribute>AttributeValue"`
	RegistrationAuthority string   `xml:"RegistrationInfo>registrationAuthority,attr"`
}

// ============================================================================
// Helper Functions
// ============================================================================

// containsString checks if a slice contains a string
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// parseInstitutionType parses institution type from entity category
func parseInstitutionType(categories []string) InstitutionType {
	for _, cat := range categories {
		if strings.Contains(cat, "university") {
			return InstitutionTypeUniversity
		}
		if strings.Contains(cat, "research") {
			return InstitutionTypeResearchInstitute
		}
		if strings.Contains(cat, "school") || strings.Contains(cat, "k-12") {
			return InstitutionTypeSchool
		}
		if strings.Contains(cat, "library") {
			return InstitutionTypeLibrary
		}
		if strings.Contains(cat, "government") {
			return InstitutionTypeGovernment
		}
	}
	return InstitutionTypeOther
}

// parseFederationFromEntityID extracts federation from entity ID
func parseFederationFromEntityID(entityID string) string {
	// Parse domain from entity ID
	if strings.Contains(entityID, ".edu") {
		if strings.Contains(entityID, ".au") {
			return "AAF"
		}
		return "InCommon"
	}
	if strings.Contains(entityID, ".ac.uk") {
		return "UK Access Federation"
	}
	if strings.Contains(entityID, ".edu.au") {
		return "AAF"
	}
	if strings.Contains(entityID, "cern.ch") {
		return "SWITCHaai"
	}
	// Default to EduGAIN for unknown
	return "EduGAIN"
}

// verifyXMLSignature verifies an XML signature
func verifyXMLSignature(data, cert []byte) (bool, error) {
	// VE-2005: Real XML-DSig verification using goxmldsig
	return VerifyXMLSignatureWithCertBytes(data, cert)
}
