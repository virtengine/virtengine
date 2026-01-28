// Package govdata provides government data source integration for identity verification.
//
// VE-909: Government data integration for authoritative identity verification
//
// # Overview
//
// This package implements privacy-preserving integration with government data sources
// including DMV (Department of Motor Vehicles), passport authorities, and vital records
// offices. The integration follows a verification-only model where actual government
// data is never stored - only verification results and timestamps are retained.
//
// # Privacy-Preserving Design
//
// The package follows strict privacy principles:
//   - Never stores raw government data
//   - Only stores verification results (match/no-match, confidence scores)
//   - All requests are logged for audit compliance
//   - Supports data minimization (verify only what's needed)
//   - Implements consent tracking
//
// # Supported Data Sources
//
// Government data sources are integrated through adapters:
//   - DMV: Driver's license and state ID verification
//   - Passport: Passport number and issuance verification
//   - Vital Records: Birth certificate verification
//   - National ID: National identity card systems
//   - Tax Authority: Tax ID verification (SSN/TIN)
//
// # Multi-Jurisdiction Support
//
// The package supports multiple jurisdictions through:
//   - Jurisdiction registry with adapter configuration
//   - Automatic routing based on document country/state
//   - Configurable data retention per jurisdiction
//   - GDPR/CCPA compliance controls
//
// # VEID Integration
//
// Government verification results contribute to VEID identity scoring:
//   - Verified documents increase trust score
//   - Multi-source verification provides higher confidence
//   - Verification freshness affects scoring weight
//
// # Audit Requirements
//
// All government data access is logged with:
//   - Requester identity and authorization
//   - Timestamp and request type
//   - Verification result (without sensitive data)
//   - Retention policy compliance tracking
//
// # Security
//
// The package implements:
//   - Encrypted communication with all data sources
//   - Request signing and verification
//   - Rate limiting per requester
//   - IP allowlisting for government APIs
//   - HSM support for key management
//
// # Usage
//
// Basic verification flow:
//
//	cfg := govdata.DefaultConfig()
//	cfg.Jurisdictions = []string{"US", "EU"}
//	
//	service, err := govdata.NewService(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	req := &govdata.VerificationRequest{
//	    WalletAddress: "ve1...",
//	    DocumentType:  govdata.DocumentTypeDriversLicense,
//	    Jurisdiction:  "US-CA",
//	    Fields: govdata.VerificationFields{
//	        DocumentNumber: "D1234567",
//	        FirstName:      "John",
//	        LastName:       "Doe",
//	        DateOfBirth:    time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
//	    },
//	}
//	
//	result, err := service.VerifyDocument(ctx, req)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	if result.Status == govdata.VerificationStatusVerified {
//	    fmt.Println("Document verified with confidence:", result.Confidence)
//	}
package govdata
