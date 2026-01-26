package types

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"golang.org/x/crypto/ocsp"
)

// ============================================================================
// Certificate Revocation Checking (VE-925)
// ============================================================================

// RevocationStatus represents the revocation status of a certificate
type RevocationStatus uint8

const (
	// RevocationStatusUnknown indicates the revocation status is unknown
	RevocationStatusUnknown RevocationStatus = 0
	// RevocationStatusGood indicates the certificate is not revoked
	RevocationStatusGood RevocationStatus = 1
	// RevocationStatusRevoked indicates the certificate has been revoked
	RevocationStatusRevoked RevocationStatus = 2
	// RevocationStatusCheckFailed indicates the revocation check failed
	RevocationStatusCheckFailed RevocationStatus = 3
)

// String returns the string representation of a RevocationStatus
func (s RevocationStatus) String() string {
	switch s {
	case RevocationStatusUnknown:
		return "unknown"
	case RevocationStatusGood:
		return "good"
	case RevocationStatusRevoked:
		return "revoked"
	case RevocationStatusCheckFailed:
		return "check_failed"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// RevocationCheckResult represents the result of a revocation check
type RevocationCheckResult struct {
	// Status is the revocation status
	Status RevocationStatus `json:"status"`

	// Method indicates which method was used (crl, ocsp)
	Method string `json:"method"`

	// CheckedAt is when the check was performed
	CheckedAt int64 `json:"checked_at"`

	// NextUpdate is when the revocation data should be refreshed
	NextUpdate int64 `json:"next_update,omitempty"`

	// RevocationTime is when the certificate was revoked (if revoked)
	RevocationTime int64 `json:"revocation_time,omitempty"`

	// RevocationReason is the reason for revocation (if revoked)
	RevocationReason string `json:"revocation_reason,omitempty"`

	// Error contains any error message from the check
	Error string `json:"error,omitempty"`

	// ResponseHash is a hash of the response for caching
	ResponseHash string `json:"response_hash,omitempty"`
}

// RevocationChecker handles certificate revocation checking
type RevocationChecker struct {
	httpClient *http.Client
	// crlCache caches CRL responses
	crlCache map[string]*CRLCacheEntry
	// ocspCache caches OCSP responses
	ocspCache map[string]*OCSPCacheEntry
}

// CRLCacheEntry represents a cached CRL
type CRLCacheEntry struct {
	CRL        *x509.RevocationList
	FetchedAt  time.Time
	NextUpdate time.Time
}

// OCSPCacheEntry represents a cached OCSP response
type OCSPCacheEntry struct {
	Response   *ocsp.Response
	FetchedAt  time.Time
	NextUpdate time.Time
}

// NewRevocationChecker creates a new revocation checker
func NewRevocationChecker(timeout time.Duration) *RevocationChecker {
	return &RevocationChecker{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		crlCache:  make(map[string]*CRLCacheEntry),
		ocspCache: make(map[string]*OCSPCacheEntry),
	}
}

// CheckRevocation checks if a certificate has been revoked using available methods
func (c *RevocationChecker) CheckRevocation(ctx context.Context, cert, issuer *x509.Certificate) *RevocationCheckResult {
	// Try OCSP first as it's more efficient
	if len(cert.OCSPServer) > 0 {
		result := c.CheckOCSP(ctx, cert, issuer)
		if result.Status != RevocationStatusCheckFailed {
			return result
		}
	}

	// Fall back to CRL
	if len(cert.CRLDistributionPoints) > 0 {
		result := c.CheckCRL(ctx, cert)
		if result.Status != RevocationStatusCheckFailed {
			return result
		}
	}

	// No revocation information available
	return &RevocationCheckResult{
		Status:    RevocationStatusUnknown,
		Method:    "none",
		CheckedAt: time.Now().Unix(),
		Error:     "no revocation checking endpoints available",
	}
}

// CheckOCSP performs an OCSP check for a certificate
func (c *RevocationChecker) CheckOCSP(ctx context.Context, cert, issuer *x509.Certificate) *RevocationCheckResult {
	now := time.Now()
	result := &RevocationCheckResult{
		Method:    "ocsp",
		CheckedAt: now.Unix(),
	}

	if len(cert.OCSPServer) == 0 {
		result.Status = RevocationStatusCheckFailed
		result.Error = "no OCSP servers available"
		return result
	}

	// Try each OCSP server
	var lastError error
	for _, ocspServer := range cert.OCSPServer {
		// Check cache first
		cacheKey := fmt.Sprintf("%s:%s", ocspServer, cert.SerialNumber.String())
		if cached, ok := c.ocspCache[cacheKey]; ok {
			if now.Before(cached.NextUpdate) {
				return ocspResponseToResult(cached.Response, result)
			}
		}

		// Create OCSP request
		ocspReq, err := ocsp.CreateRequest(cert, issuer, nil)
		if err != nil {
			lastError = fmt.Errorf("failed to create OCSP request: %v", err)
			continue
		}

		// Make OCSP request
		resp, err := c.makeOCSPRequest(ctx, ocspServer, ocspReq)
		if err != nil {
			lastError = fmt.Errorf("OCSP request failed: %v", err)
			continue
		}

		// Parse OCSP response
		ocspResp, err := ocsp.ParseResponseForCert(resp, cert, issuer)
		if err != nil {
			lastError = fmt.Errorf("failed to parse OCSP response: %v", err)
			continue
		}

		// Cache the response
		c.ocspCache[cacheKey] = &OCSPCacheEntry{
			Response:   ocspResp,
			FetchedAt:  now,
			NextUpdate: ocspResp.NextUpdate,
		}

		return ocspResponseToResult(ocspResp, result)
	}

	result.Status = RevocationStatusCheckFailed
	if lastError != nil {
		result.Error = lastError.Error()
	} else {
		result.Error = "all OCSP servers failed"
	}
	return result
}

// makeOCSPRequest makes an HTTP OCSP request
func (c *RevocationChecker) makeOCSPRequest(ctx context.Context, server string, requestBytes []byte) ([]byte, error) {
	// Try POST first (RFC 6960 recommends POST for larger requests)
	req, err := http.NewRequestWithContext(ctx, "POST", server, bytes.NewReader(requestBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/ocsp-request")
	req.Header.Set("Accept", "application/ocsp-response")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Try GET as fallback
		return c.makeOCSPRequestGET(ctx, server, requestBytes)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OCSP server returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// makeOCSPRequestGET makes an HTTP GET OCSP request
func (c *RevocationChecker) makeOCSPRequestGET(ctx context.Context, server string, requestBytes []byte) ([]byte, error) {
	encoded := base64.StdEncoding.EncodeToString(requestBytes)
	url := fmt.Sprintf("%s/%s", server, encoded)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/ocsp-response")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OCSP server returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// ocspResponseToResult converts an OCSP response to a RevocationCheckResult
func ocspResponseToResult(resp *ocsp.Response, result *RevocationCheckResult) *RevocationCheckResult {
	result.NextUpdate = resp.NextUpdate.Unix()

	switch resp.Status {
	case ocsp.Good:
		result.Status = RevocationStatusGood
	case ocsp.Revoked:
		result.Status = RevocationStatusRevoked
		result.RevocationTime = resp.RevokedAt.Unix()
		result.RevocationReason = ocspRevocationReasonToString(resp.RevocationReason)
	case ocsp.Unknown:
		result.Status = RevocationStatusUnknown
	default:
		result.Status = RevocationStatusCheckFailed
		result.Error = fmt.Sprintf("unexpected OCSP status: %d", resp.Status)
	}

	return result
}

// ocspRevocationReasonToString converts OCSP revocation reason to string
func ocspRevocationReasonToString(reason int) string {
	switch reason {
	case ocsp.Unspecified:
		return "unspecified"
	case ocsp.KeyCompromise:
		return "key_compromise"
	case ocsp.CACompromise:
		return "ca_compromise"
	case ocsp.AffiliationChanged:
		return "affiliation_changed"
	case ocsp.Superseded:
		return "superseded"
	case ocsp.CessationOfOperation:
		return "cessation_of_operation"
	case ocsp.CertificateHold:
		return "certificate_hold"
	case ocsp.RemoveFromCRL:
		return "remoVIRTENGINE_from_crl"
	case ocsp.PrivilegeWithdrawn:
		return "privilege_withdrawn"
	case ocsp.AACompromise:
		return "aa_compromise"
	default:
		return fmt.Sprintf("unknown(%d)", reason)
	}
}

// CheckCRL performs a CRL check for a certificate
func (c *RevocationChecker) CheckCRL(ctx context.Context, cert *x509.Certificate) *RevocationCheckResult {
	now := time.Now()
	result := &RevocationCheckResult{
		Method:    "crl",
		CheckedAt: now.Unix(),
	}

	if len(cert.CRLDistributionPoints) == 0 {
		result.Status = RevocationStatusCheckFailed
		result.Error = "no CRL distribution points available"
		return result
	}

	// Try each CRL distribution point
	var lastError error
	for _, crlURL := range cert.CRLDistributionPoints {
		// Check cache first
		if cached, ok := c.crlCache[crlURL]; ok {
			if now.Before(cached.NextUpdate) {
				return checkCRLForCert(cached.CRL, cert, result)
			}
		}

		// Fetch CRL
		crl, err := c.fetchCRL(ctx, crlURL)
		if err != nil {
			lastError = fmt.Errorf("failed to fetch CRL: %v", err)
			continue
		}

		// Cache the CRL
		c.crlCache[crlURL] = &CRLCacheEntry{
			CRL:        crl,
			FetchedAt:  now,
			NextUpdate: crl.NextUpdate,
		}

		return checkCRLForCert(crl, cert, result)
	}

	result.Status = RevocationStatusCheckFailed
	if lastError != nil {
		result.Error = lastError.Error()
	} else {
		result.Error = "all CRL distribution points failed"
	}
	return result
}

// fetchCRL fetches a CRL from a URL
func (c *RevocationChecker) fetchCRL(ctx context.Context, url string) (*x509.RevocationList, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CRL server returned status %d", resp.StatusCode)
	}

	crlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read CRL: %v", err)
	}

	return x509.ParseRevocationList(crlBytes)
}

// checkCRLForCert checks if a certificate is in a CRL
func checkCRLForCert(crl *x509.RevocationList, cert *x509.Certificate, result *RevocationCheckResult) *RevocationCheckResult {
	result.NextUpdate = crl.NextUpdate.Unix()

	for _, revoked := range crl.RevokedCertificateEntries {
		if revoked.SerialNumber.Cmp(cert.SerialNumber) == 0 {
			result.Status = RevocationStatusRevoked
			result.RevocationTime = revoked.RevocationTime.Unix()
			result.RevocationReason = crlRevocationReasonToString(revoked.ReasonCode)
			return result
		}
	}

	result.Status = RevocationStatusGood
	return result
}

// crlRevocationReasonToString converts CRL revocation reason to string
func crlRevocationReasonToString(reason int) string {
	switch reason {
	case 0:
		return "unspecified"
	case 1:
		return "key_compromise"
	case 2:
		return "ca_compromise"
	case 3:
		return "affiliation_changed"
	case 4:
		return "superseded"
	case 5:
		return "cessation_of_operation"
	case 6:
		return "certificate_hold"
	case 8:
		return "remoVIRTENGINE_from_crl"
	case 9:
		return "privilege_withdrawn"
	case 10:
		return "aa_compromise"
	default:
		return fmt.Sprintf("unknown(%d)", reason)
	}
}

// VerifyCRLSignature verifies that a CRL was signed by the given issuer
func VerifyCRLSignature(crl *x509.RevocationList, issuer *x509.Certificate) error {
	return crl.CheckSignatureFrom(issuer)
}

// CreateOCSPRequest creates an OCSP request for a certificate
func CreateOCSPRequest(cert, issuer *x509.Certificate, nonce []byte) ([]byte, error) {
	opts := &ocsp.RequestOptions{
		Hash: crypto.SHA256,
	}

	return ocsp.CreateRequest(cert, issuer, opts)
}

// OCSPRequestWithNonce represents an OCSP request with a nonce extension
type OCSPRequestWithNonce struct {
	TBSRequest       TBSRequest
	OptionalSignature asn1.RawValue `asn1:"explicit,optional,tag:0"`
}

// TBSRequest represents the TBS portion of an OCSP request
type TBSRequest struct {
	Version       int           `asn1:"explicit,default:0,tag:0"`
	RequestorName asn1.RawValue `asn1:"explicit,optional,tag:1"`
	RequestList   []Request
	Extensions    []Extension `asn1:"explicit,optional,tag:2"`
}

// Request represents a single request in an OCSP request
type Request struct {
	CertID          CertID
	SingleExtensions []Extension `asn1:"explicit,optional,tag:0"`
}

// CertID identifies a certificate in an OCSP request
type CertID struct {
	HashAlgorithm  asn1.RawValue
	IssuerNameHash []byte
	IssuerKeyHash  []byte
	SerialNumber   *big.Int
}

// Extension represents an ASN.1 extension
type Extension struct {
	OID      asn1.ObjectIdentifier
	Critical bool `asn1:"optional"`
	Value    []byte
}

// ClearCaches clears both CRL and OCSP caches
func (c *RevocationChecker) ClearCaches() {
	c.crlCache = make(map[string]*CRLCacheEntry)
	c.ocspCache = make(map[string]*OCSPCacheEntry)
}

// ClearExpiredCaches removes expired entries from both caches
func (c *RevocationChecker) ClearExpiredCaches() {
	now := time.Now()

	for key, entry := range c.crlCache {
		if now.After(entry.NextUpdate) {
			delete(c.crlCache, key)
		}
	}

	for key, entry := range c.ocspCache {
		if now.After(entry.NextUpdate) {
			delete(c.ocspCache, key)
		}
	}
}
