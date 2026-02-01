// Package govdata provides government data source integration for identity verification.
//
// VE-909: Government data integration for authoritative identity verification
// SECURITY-004: Real government verification APIs with liveness, fraud, and compliance
//
// # Overview
//
// This package implements privacy-preserving integration with government data sources
// including DMV (Department of Motor Vehicles), passport authorities, and vital records
// offices. The integration follows a verification-only model where actual government
// data is never stored - only verification results and timestamps are retained.
//
// # Implemented Government API Adapters
//
// Real government verification APIs are integrated:
//
//   - AAMVA (US): American Association of Motor Vehicle Administrators DLDV API
//     for driver's license verification across all 50 US states and territories
//
//   - DVS (Australia): Document Verification Service for Australian documents
//     across 8 states/territories
//
//   - eIDAS (EU): European electronic identification framework for 27 EU member
//     states plus 3 EEA countries using SAML-based authentication
//
//   - GOV.UK (UK): UK Government Verify service supporting LOA_1 and LOA_2
//     assurance levels
//
//   - PCTF (Canada): Pan-Canadian Trust Framework for 13 provinces/territories
//
// # Liveness Detection Integration
//
// Liveness detection prevents presentation attacks during identity verification:
//
//   - Passive: Background analysis without user interaction
//   - Active: User-prompted challenges (blink, smile, turn head)
//   - Hybrid: Combination of both for maximum security
//
// Liveness results contribute to VEID identity scoring and trigger fraud detection
// when spoofing is detected.
//
// # Fraud Detection Hooks
//
// Comprehensive fraud detection is integrated:
//
//   - Velocity checking: Detect unusual verification patterns
//   - Blacklist checking: Block known fraudulent documents
//   - Verification analysis: Detect fake documents and stolen identities
//   - Liveness spoofing detection: Photo, screen, mask, and deepfake detection
//
// Critical fraud is reported to the x/fraud blockchain module for on-chain tracking.
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
// # GDPR/CCPA Compliance
//
// The privacy module provides:
//   - Data minimization enforcement (GDPR Article 5)
//   - Subject access requests (GDPR Article 15)
//   - Right to erasure (GDPR Article 17)
//   - Data portability (GDPR Article 20)
//   - PII hashing and anonymization
//   - Jurisdiction-aware retention policies
//
// # Cost Management
//
// API cost tracking and budget management:
//   - Per-adapter cost configuration
//   - Daily and monthly budget limits
//   - Real-time cost tracking and alerts
//   - Cost estimation before verification
//
// # Supported Document Types
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
//   - Liveness detection adds biometric assurance
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
//
// Combined verification with liveness:
//
//	livenessConfig := govdata.LivenessConfig{
//	    Mode:                "hybrid",
//	    MinConfidence:       0.9,
//	    RequireBothPassive:  true,
//	}
//
//	result, err := service.VerifyDocumentWithLiveness(ctx, docReq, livenessReq, livenessConfig)
package govdata

