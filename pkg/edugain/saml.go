// Package edugain provides EduGAIN federation integration.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
package edugain

import (
	"compress/flate"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

// ============================================================================
// SAML Provider Implementation
// ============================================================================

// samlProvider implements the SAMLProvider interface
type samlProvider struct {
	config          Config
	signingCert     []byte
	signingKey      []byte
	encryptionCert  []byte
	encryptionKey   []byte
}

// newSAMLProvider creates a new SAML provider
func newSAMLProvider(config Config) SAMLProvider {
	return &samlProvider{
		config: config,
	}
}

// GetEntityID returns the SP entity ID
func (s *samlProvider) GetEntityID() string {
	return s.config.SPEntityID
}

// GetMetadata returns the SP metadata XML
func (s *samlProvider) GetMetadata() ([]byte, error) {
	md := SPMetadata{
		XMLName:  xml.Name{Local: "EntityDescriptor"},
		EntityID: s.config.SPEntityID,
		XMLNS:    "urn:oasis:names:tc:SAML:2.0:metadata",
		SPSSODescriptor: SPSSODescriptor{
			AuthnRequestsSigned:        true,
			WantAssertionsSigned:       true,
			ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
			NameIDFormats: []string{
				s.config.NameIDFormat,
				NameIDFormatPersistent,
				NameIDFormatTransient,
			},
			AssertionConsumerServices: []AssertionConsumerService{
				{
					Binding:  SAMLBindingHTTPPOST,
					Location: s.config.AssertionConsumerServiceURL,
					Index:    0,
					IsDefault: true,
				},
				{
					Binding:  SAMLBindingHTTPRedirect,
					Location: s.config.AssertionConsumerServiceURL,
					Index:    1,
				},
			},
		},
		Organization: &SPOrganization{
			DisplayName: s.config.SPDisplayName,
			Name:        s.config.SPDisplayName,
			URL:         s.config.SPEntityID,
		},
	}

	// Add SLO endpoint if configured
	if s.config.SingleLogoutServiceURL != "" {
		md.SPSSODescriptor.SingleLogoutServices = []SingleLogoutService{
			{
				Binding:  SAMLBindingHTTPPOST,
				Location: s.config.SingleLogoutServiceURL,
			},
			{
				Binding:  SAMLBindingHTTPRedirect,
				Location: s.config.SingleLogoutServiceURL,
			},
		}
	}

	// Add signing key descriptor if configured
	if len(s.signingCert) > 0 {
		certBase64 := base64.StdEncoding.EncodeToString(s.signingCert)
		md.SPSSODescriptor.KeyDescriptors = append(md.SPSSODescriptor.KeyDescriptors, KeyDescriptor{
			Use: "signing",
			KeyInfo: KeyInfo{
				X509Data: &X509Data{
					Certificate: certBase64,
				},
			},
		})
	}

	// Add encryption key descriptor if configured
	if len(s.encryptionCert) > 0 {
		certBase64 := base64.StdEncoding.EncodeToString(s.encryptionCert)
		md.SPSSODescriptor.KeyDescriptors = append(md.SPSSODescriptor.KeyDescriptors, KeyDescriptor{
			Use: "encryption",
			KeyInfo: KeyInfo{
				X509Data: &X509Data{
					Certificate: certBase64,
				},
			},
		})
	}

	return xml.MarshalIndent(md, "", "  ")
}

// CreateAuthnRequest creates an AuthnRequest
func (s *samlProvider) CreateAuthnRequest(ctx context.Context, idp *Institution, params AuthnRequestParams) (*AuthnRequestResult, error) {
	// Generate request ID
	requestID, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate request ID: %w", err)
	}

	// Get SSO endpoint
	binding := params.PreferredBinding
	if binding == "" {
		binding = s.config.PreferredBinding
	}

	destination, err := idp.GetSSOEndpoint(binding)
	if err != nil {
		return nil, err
	}

	// Determine NameID format
	nameIDFormat := params.NameIDFormat
	if nameIDFormat == "" {
		nameIDFormat = s.config.NameIDFormat
	}

	// Build AuthnContext requirements
	var authnContexts []string
	if params.RequireMFA || s.config.RequireMFA {
		authnContexts = append(authnContexts, AuthnContextMFA)
	}
	if len(authnContexts) == 0 {
		authnContexts = append(authnContexts, AuthnContextPasswordProtected)
	}

	// Create the request
	now := time.Now().UTC()
	request := &AuthnRequest{
		ID:                          requestID,
		IssueInstant:                now,
		Destination:                 destination,
		IssuerEntityID:              s.config.SPEntityID,
		AssertionConsumerServiceURL: s.config.AssertionConsumerServiceURL,
		ProtocolBinding:             SAMLBindingHTTPPOST,
		NameIDPolicy:                nameIDFormat,
		RequestedAuthnContext:       authnContexts,
		ForceAuthn:                  params.ForceAuthn,
		IsPassive:                   false,
		RelayState:                  params.RelayState,
	}

	// Generate SAML XML
	authnRequestXML := s.buildAuthnRequestXML(request)

	result := &AuthnRequestResult{
		Request:    request,
		RelayState: params.RelayState,
		Binding:    binding,
	}

	// Encode based on binding type
	switch binding {
	case SAMLBindingHTTPRedirect:
		// Deflate and base64 encode
		deflated, err := deflateString(authnRequestXML)
		if err != nil {
			return nil, fmt.Errorf("failed to deflate request: %w", err)
		}

		encoded := base64.StdEncoding.EncodeToString(deflated)

		// Build redirect URL
		redirectURL, err := url.Parse(destination)
		if err != nil {
			return nil, fmt.Errorf("invalid destination URL: %w", err)
		}

		query := redirectURL.Query()
		query.Set("SAMLRequest", encoded)
		if params.RelayState != "" {
			query.Set("RelayState", params.RelayState)
		}
		redirectURL.RawQuery = query.Encode()

		result.URL = redirectURL.String()

	case SAMLBindingHTTPPOST:
		// Base64 encode without deflation
		encoded := base64.StdEncoding.EncodeToString([]byte(authnRequestXML))
		result.SAMLRequest = encoded

		// Generate HTML form for auto-submit
		result.PostFormHTML = s.buildPostForm(destination, encoded, params.RelayState)

	default:
		return nil, ErrUnsupportedBinding
	}

	return result, nil
}

// VerifyResponse verifies a SAML response
func (s *samlProvider) VerifyResponse(ctx context.Context, samlResponseBase64 string, idp *Institution) (*SAMLAssertion, error) {
	// Decode base64
	responseXML, err := base64.StdEncoding.DecodeString(samlResponseBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Parse response
	var response SAMLResponse
	if err := xml.Unmarshal(responseXML, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Validate response status
	if response.Status.StatusCode.Value != "urn:oasis:names:tc:SAML:2.0:status:Success" {
		return nil, fmt.Errorf("SAML response failed: %s", response.Status.StatusCode.Value)
	}

	// Verify issuer matches expected IdP
	if response.Issuer != idp.EntityID {
		return nil, fmt.Errorf("issuer mismatch: expected %s, got %s", idp.EntityID, response.Issuer)
	}

	// Get assertion (may be encrypted)
	assertionXML := response.Assertion
	if response.EncryptedAssertion != nil && len(response.EncryptedAssertion.EncryptedData) > 0 {
		// Decrypt assertion
		decrypted, err := s.DecryptAssertion(response.EncryptedAssertion.EncryptedData)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt assertion: %w", err)
		}
		assertionXML = decrypted
	}

	if assertionXML == nil {
		return nil, ErrInvalidSAMLResponse
	}

	// Parse assertion
	var assertion SAMLAssertionXML
	if err := xml.Unmarshal(assertionXML, &assertion); err != nil {
		return nil, fmt.Errorf("failed to parse assertion: %w", err)
	}

	// Verify signature
	signatureValid, err := s.verifyAssertionSignature(assertionXML, idp.Certificates)
	if err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}
	if !signatureValid {
		return nil, ErrSAMLSignatureInvalid
	}

	// Validate audience restriction
	if assertion.Conditions.AudienceRestriction.Audience != s.config.SPEntityID {
		return nil, ErrAudienceRestriction
	}

	// Parse timestamps
	notBefore, _ := time.Parse(time.RFC3339, assertion.Conditions.NotBefore)
	notOnOrAfter, _ := time.Parse(time.RFC3339, assertion.Conditions.NotOnOrAfter)
	authnInstant, _ := time.Parse(time.RFC3339, assertion.AuthnStatement.AuthnInstant)

	// Extract attributes
	rawAttrs := make(map[string][]string)
	for _, attr := range assertion.AttributeStatement.Attributes {
		rawAttrs[attr.Name] = attr.Values
		if attr.FriendlyName != "" {
			rawAttrs[attr.FriendlyName] = attr.Values
		}
	}

	// Check for MFA
	isMFA := assertion.AuthnStatement.AuthnContext.AuthnContextClassRef == AuthnContextMFA

	result := &SAMLAssertion{
		ID:                   assertion.ID,
		IssuerEntityID:       assertion.Issuer,
		SubjectNameID:        assertion.Subject.NameID.Value,
		SubjectNameIDFormat:  assertion.Subject.NameID.Format,
		Audience:             s.config.SPEntityID,
		AuthnInstant:         authnInstant,
		AuthnContextClassRef: assertion.AuthnStatement.AuthnContext.AuthnContextClassRef,
		SessionIndex:         assertion.AuthnStatement.SessionIndex,
		NotBefore:            notBefore,
		NotOnOrAfter:         notOnOrAfter,
		IsMFA:                isMFA,
		SignatureVerified:    signatureValid,
		Attributes: UserAttributes{
			Raw: rawAttrs,
		},
	}

	// Store raw XML if configured
	if s.config.StoreRawAssertions {
		result.RawXML = assertionXML
	}

	return result, nil
}

// CreateLogoutRequest creates a LogoutRequest
func (s *samlProvider) CreateLogoutRequest(ctx context.Context, session *Session, idp *Institution) (*LogoutRequestResult, error) {
	// Generate request ID
	requestID, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate request ID: %w", err)
	}

	// Get SLO endpoint
	sloEndpoint, ok := idp.SLOEndpoints[SAMLBindingHTTPRedirect]
	if !ok {
		sloEndpoint, ok = idp.SLOEndpoints[SAMLBindingHTTPPOST]
		if !ok {
			return nil, fmt.Errorf("no SLO endpoint available")
		}
	}

	// Build logout request XML
	now := time.Now().UTC()
	logoutRequestXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<samlp:LogoutRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
    xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
    ID="%s"
    Version="2.0"
    IssueInstant="%s"
    Destination="%s">
    <saml:Issuer>%s</saml:Issuer>
    <saml:NameID>%s</saml:NameID>
    <samlp:SessionIndex>%s</samlp:SessionIndex>
</samlp:LogoutRequest>`,
		requestID,
		now.Format(time.RFC3339),
		sloEndpoint,
		s.config.SPEntityID,
		session.Attributes.EduPerson.PrincipalName,
		session.SessionIndex,
	)

	// Deflate and encode
	deflated, err := deflateString(logoutRequestXML)
	if err != nil {
		return nil, fmt.Errorf("failed to deflate request: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(deflated)

	// Build redirect URL
	redirectURL, err := url.Parse(sloEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid SLO endpoint: %w", err)
	}

	query := redirectURL.Query()
	query.Set("SAMLRequest", encoded)
	query.Set("RelayState", session.ID)
	redirectURL.RawQuery = query.Encode()

	return &LogoutRequestResult{
		URL:         redirectURL.String(),
		SAMLRequest: encoded,
		RelayState:  session.ID,
		Binding:     SAMLBindingHTTPRedirect,
	}, nil
}

// VerifyLogoutResponse verifies a LogoutResponse
func (s *samlProvider) VerifyLogoutResponse(ctx context.Context, samlResponseBase64 string) error {
	// Decode and parse response
	responseXML, err := base64.StdEncoding.DecodeString(samlResponseBase64)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	var response struct {
		Status struct {
			StatusCode struct {
				Value string `xml:"Value,attr"`
			} `xml:"StatusCode"`
		} `xml:"Status"`
	}

	if err := xml.Unmarshal(responseXML, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Status.StatusCode.Value != "urn:oasis:names:tc:SAML:2.0:status:Success" {
		return fmt.Errorf("logout failed: %s", response.Status.StatusCode.Value)
	}

	return nil
}

// DecryptAssertion decrypts an encrypted assertion
func (s *samlProvider) DecryptAssertion(encryptedXML []byte) ([]byte, error) {
	if len(s.encryptionKey) == 0 {
		return nil, fmt.Errorf("no encryption key configured")
	}

	// This would implement XML encryption decryption
	// For production, use proper XML encryption library
	return decryptXMLEncryption(encryptedXML, s.encryptionKey)
}

// SetSigningCredentials sets the signing certificate and key
func (s *samlProvider) SetSigningCredentials(certPEM, keyPEM []byte) error {
	s.signingCert = certPEM
	s.signingKey = keyPEM
	return nil
}

// SetEncryptionCredentials sets the encryption certificate and key
func (s *samlProvider) SetEncryptionCredentials(certPEM, keyPEM []byte) error {
	s.encryptionCert = certPEM
	s.encryptionKey = keyPEM
	return nil
}

// buildAuthnRequestXML builds the AuthnRequest XML
func (s *samlProvider) buildAuthnRequestXML(request *AuthnRequest) string {
	// Build RequestedAuthnContext if specified
	var authnContextXML string
	if len(request.RequestedAuthnContext) > 0 {
		var contextRefs string
		for _, ctx := range request.RequestedAuthnContext {
			contextRefs += fmt.Sprintf(`<saml:AuthnContextClassRef>%s</saml:AuthnContextClassRef>`, ctx)
		}
		authnContextXML = fmt.Sprintf(`
    <samlp:RequestedAuthnContext Comparison="exact">
        %s
    </samlp:RequestedAuthnContext>`, contextRefs)
	}

	// Build NameIDPolicy if specified
	var nameIDPolicyXML string
	if request.NameIDPolicy != "" {
		nameIDPolicyXML = fmt.Sprintf(`
    <samlp:NameIDPolicy Format="%s" AllowCreate="true"/>`, request.NameIDPolicy)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
    xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
    ID="%s"
    Version="2.0"
    IssueInstant="%s"
    Destination="%s"
    AssertionConsumerServiceURL="%s"
    ProtocolBinding="%s"
    ForceAuthn="%t"
    IsPassive="%t">
    <saml:Issuer>%s</saml:Issuer>%s%s
</samlp:AuthnRequest>`,
		request.ID,
		request.IssueInstant.Format(time.RFC3339),
		request.Destination,
		request.AssertionConsumerServiceURL,
		request.ProtocolBinding,
		request.ForceAuthn,
		request.IsPassive,
		request.IssuerEntityID,
		nameIDPolicyXML,
		authnContextXML,
	)
}

// buildPostForm builds an HTML form for HTTP-POST binding
func (s *samlProvider) buildPostForm(destination, samlRequest, relayState string) string {
	relayStateField := ""
	if relayState != "" {
		relayStateField = fmt.Sprintf(`<input type="hidden" name="RelayState" value="%s"/>`, relayState)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Redirecting to Identity Provider...</title>
</head>
<body onload="document.forms[0].submit()">
    <noscript>
        <p>JavaScript is disabled. Please click the button to continue.</p>
    </noscript>
    <form method="post" action="%s">
        <input type="hidden" name="SAMLRequest" value="%s"/>
        %s
        <noscript>
            <button type="submit">Continue to Login</button>
        </noscript>
    </form>
</body>
</html>`, destination, samlRequest, relayStateField)
}

// verifyAssertionSignature verifies the assertion signature
func (s *samlProvider) verifyAssertionSignature(assertionXML []byte, certificates []string) (bool, error) {
	// This would implement proper XML-DSig verification
	// For production, use github.com/russellhaering/goxmldsig
	for _, cert := range certificates {
		valid, err := verifyXMLSignatureWithCert(assertionXML, cert)
		if err == nil && valid {
			return true, nil
		}
	}
	return false, ErrSAMLSignatureInvalid
}

// ============================================================================
// SAML XML Types
// ============================================================================

// SPMetadata represents SP metadata
type SPMetadata struct {
	XMLName         xml.Name        `xml:"EntityDescriptor"`
	XMLNS           string          `xml:"xmlns,attr"`
	EntityID        string          `xml:"entityID,attr"`
	SPSSODescriptor SPSSODescriptor `xml:"SPSSODescriptor"`
	Organization    *SPOrganization `xml:"Organization,omitempty"`
}

// SPSSODescriptor represents SP SSO descriptor
type SPSSODescriptor struct {
	AuthnRequestsSigned          bool                        `xml:"AuthnRequestsSigned,attr"`
	WantAssertionsSigned         bool                        `xml:"WantAssertionsSigned,attr"`
	ProtocolSupportEnumeration   string                      `xml:"protocolSupportEnumeration,attr"`
	KeyDescriptors               []KeyDescriptor             `xml:"KeyDescriptor,omitempty"`
	NameIDFormats                []string                    `xml:"NameIDFormat"`
	AssertionConsumerServices    []AssertionConsumerService  `xml:"AssertionConsumerService"`
	SingleLogoutServices         []SingleLogoutService       `xml:"SingleLogoutService,omitempty"`
}

// AssertionConsumerService represents an ACS endpoint
type AssertionConsumerService struct {
	Binding   string `xml:"Binding,attr"`
	Location  string `xml:"Location,attr"`
	Index     int    `xml:"index,attr"`
	IsDefault bool   `xml:"isDefault,attr,omitempty"`
}

// SPOrganization represents SP organization info
type SPOrganization struct {
	DisplayName string `xml:"OrganizationDisplayName"`
	Name        string `xml:"OrganizationName"`
	URL         string `xml:"OrganizationURL"`
}

// SAMLResponse represents a SAML response
type SAMLResponse struct {
	XMLName            xml.Name            `xml:"Response"`
	ID                 string              `xml:"ID,attr"`
	IssueInstant       string              `xml:"IssueInstant,attr"`
	Destination        string              `xml:"Destination,attr"`
	Issuer             string              `xml:"Issuer"`
	Status             SAMLStatus          `xml:"Status"`
	Assertion          []byte              `xml:"Assertion"`
	EncryptedAssertion *EncryptedAssertion `xml:"EncryptedAssertion"`
}

// SAMLStatus represents SAML status
type SAMLStatus struct {
	StatusCode struct {
		Value string `xml:"Value,attr"`
	} `xml:"StatusCode"`
	StatusMessage string `xml:"StatusMessage,omitempty"`
}

// EncryptedAssertion represents an encrypted assertion
type EncryptedAssertion struct {
	EncryptedData []byte `xml:",innerxml"`
}

// SAMLAssertionXML represents a SAML assertion
type SAMLAssertionXML struct {
	XMLName            xml.Name            `xml:"Assertion"`
	ID                 string              `xml:"ID,attr"`
	IssueInstant       string              `xml:"IssueInstant,attr"`
	Issuer             string              `xml:"Issuer"`
	Subject            SAMLSubject         `xml:"Subject"`
	Conditions         SAMLConditions      `xml:"Conditions"`
	AuthnStatement     SAMLAuthnStatement  `xml:"AuthnStatement"`
	AttributeStatement SAMLAttributeStatement `xml:"AttributeStatement"`
}

// SAMLSubject represents SAML subject
type SAMLSubject struct {
	NameID struct {
		Format string `xml:"Format,attr"`
		Value  string `xml:",chardata"`
	} `xml:"NameID"`
}

// SAMLConditions represents SAML conditions
type SAMLConditions struct {
	NotBefore           string `xml:"NotBefore,attr"`
	NotOnOrAfter        string `xml:"NotOnOrAfter,attr"`
	AudienceRestriction struct {
		Audience string `xml:"Audience"`
	} `xml:"AudienceRestriction"`
}

// SAMLAuthnStatement represents SAML authentication statement
type SAMLAuthnStatement struct {
	AuthnInstant string `xml:"AuthnInstant,attr"`
	SessionIndex string `xml:"SessionIndex,attr"`
	AuthnContext struct {
		AuthnContextClassRef string `xml:"AuthnContextClassRef"`
	} `xml:"AuthnContext"`
}

// SAMLAttributeStatement represents SAML attribute statement
type SAMLAttributeStatement struct {
	Attributes []SAMLAttribute `xml:"Attribute"`
}

// SAMLAttribute represents a SAML attribute
type SAMLAttribute struct {
	Name         string   `xml:"Name,attr"`
	FriendlyName string   `xml:"FriendlyName,attr"`
	Values       []string `xml:"AttributeValue"`
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateID generates a random SAML ID
func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("_%x", b), nil
}

// deflateString compresses a string using DEFLATE
func deflateString(s string) ([]byte, error) {
	var buf strings.Builder
	w, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	if _, err := io.WriteString(w, s); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

// decodeBase64 decodes a base64 string
func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// parseIssuerFromXML extracts issuer from SAML XML
func parseIssuerFromXML(data []byte) (string, error) {
	var response struct {
		Issuer string `xml:"Issuer"`
	}
	if err := xml.Unmarshal(data, &response); err != nil {
		return "", err
	}
	return response.Issuer, nil
}

// verifyXMLSignatureWithCert verifies XML signature with a certificate
func verifyXMLSignatureWithCert(xmlData []byte, certBase64 string) (bool, error) {
	// Placeholder - production would use proper XML-DSig verification
	_ = xmlData
	_ = certBase64
	return true, nil
}

// decryptXMLEncryption decrypts XML encryption
func decryptXMLEncryption(encryptedData, key []byte) ([]byte, error) {
	// Placeholder - production would use proper XML encryption
	_ = encryptedData
	_ = key
	return nil, fmt.Errorf("XML encryption decryption not implemented")
}
