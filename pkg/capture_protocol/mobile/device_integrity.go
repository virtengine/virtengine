// Package mobile implements native mobile capture specifications for iOS and Android.
// VE-900/VE-4F: Device integrity checks - SafetyNet/DeviceCheck/App Attest integration
//
// This file implements device attestation and integrity verification using
// platform-specific attestation APIs (Google Play Integrity, Apple DeviceCheck/App Attest).
package mobile

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// ============================================================================
// Device Attestation Types
// ============================================================================

// AttestationProvider represents the attestation provider
type AttestationProvider string

const (
	// AttestationProviderPlayIntegrity is Google Play Integrity API (replaces SafetyNet)
	AttestationProviderPlayIntegrity AttestationProvider = "play_integrity"

	// AttestationProviderSafetyNet is Google SafetyNet (deprecated)
	AttestationProviderSafetyNet AttestationProvider = "safetynet"

	// AttestationProviderDeviceCheck is Apple DeviceCheck
	AttestationProviderDeviceCheck AttestationProvider = "devicecheck"

	// AttestationProviderAppAttest is Apple App Attest
	AttestationProviderAppAttest AttestationProvider = "app_attest"

	// AttestationProviderNone indicates no attestation
	AttestationProviderNone AttestationProvider = "none"
)

// DeviceIntegrityStatus represents the overall integrity status
type DeviceIntegrityStatus string

const (
	// IntegrityStatusPassed indicates all integrity checks passed
	IntegrityStatusPassed DeviceIntegrityStatus = "passed"

	// IntegrityStatusPartial indicates some checks passed
	IntegrityStatusPartial DeviceIntegrityStatus = "partial"

	// IntegrityStatusFailed indicates integrity checks failed
	IntegrityStatusFailed DeviceIntegrityStatus = "failed"

	// IntegrityStatusUnavailable indicates attestation is not available
	IntegrityStatusUnavailable DeviceIntegrityStatus = "unavailable"

	// IntegrityStatusError indicates an error during attestation
	IntegrityStatusError DeviceIntegrityStatus = "error"
)

// ============================================================================
// Device Attestation Result
// ============================================================================

// DeviceAttestationResult contains the result of device attestation
type DeviceAttestationResult struct {
	// Provider is the attestation provider used
	Provider AttestationProvider `json:"provider"`

	// Status is the overall integrity status
	Status DeviceIntegrityStatus `json:"status"`

	// Timestamp is when attestation was performed
	Timestamp time.Time `json:"timestamp"`

	// Expiry is when the attestation expires (if applicable)
	Expiry *time.Time `json:"expiry,omitempty"`

	// Token is the attestation token (for server verification)
	Token string `json:"token,omitempty"`

	// TokenHash is SHA256 of the token (for logging without exposing token)
	TokenHash string `json:"token_hash,omitempty"`

	// PlayIntegrityVerdict is the Play Integrity result (Android)
	PlayIntegrityVerdict *PlayIntegrityVerdict `json:"play_integrity_verdict,omitempty"`

	// SafetyNetResult is the SafetyNet result (Android, deprecated)
	SafetyNetResult *SafetyNetResult `json:"safetynet_result,omitempty"`

	// AppAttestResult is the App Attest result (iOS)
	AppAttestResult *AppAttestResult `json:"app_attest_result,omitempty"`

	// DeviceCheckResult is the DeviceCheck result (iOS)
	DeviceCheckResult *DeviceCheckResult `json:"devicecheck_result,omitempty"`

	// Checks contains individual check results
	Checks []IntegrityCheck `json:"checks"`

	// Error contains any error message
	Error string `json:"error,omitempty"`
}

// IsValid returns true if the attestation is valid and not expired
func (r *DeviceAttestationResult) IsValid() bool {
	if r.Status != IntegrityStatusPassed {
		return false
	}
	if r.Expiry != nil && time.Now().After(*r.Expiry) {
		return false
	}
	return true
}

// IntegrityCheck represents a single integrity check result
type IntegrityCheck struct {
	// Name is the check name
	Name string `json:"name"`

	// Passed indicates if the check passed
	Passed bool `json:"passed"`

	// Required indicates if this check is required
	Required bool `json:"required"`

	// Details contains check-specific details
	Details string `json:"details,omitempty"`

	// Severity indicates the severity if failed
	Severity IntegrityCheckSeverity `json:"severity"`
}

// IntegrityCheckSeverity represents check severity
type IntegrityCheckSeverity string

const (
	SeverityCritical IntegrityCheckSeverity = "critical" // Fails attestation
	SeverityHigh     IntegrityCheckSeverity = "high"     // Warning, may fail
	SeverityMedium   IntegrityCheckSeverity = "medium"   // Warning
	SeverityLow      IntegrityCheckSeverity = "low"      // Informational
)

// ============================================================================
// Google Play Integrity API
// ============================================================================

// PlayIntegrityVerdict contains the Google Play Integrity verdict
type PlayIntegrityVerdict struct {
	// RequestPackageName is the requesting package name
	RequestPackageName string `json:"requestPackageName"`

	// AppIntegrity contains app integrity verdict
	AppIntegrity AppIntegrityVerdict `json:"appIntegrity"`

	// DeviceIntegrity contains device integrity verdict
	DeviceIntegrity DeviceIntegrityVerdict `json:"deviceIntegrity"`

	// AccountDetails contains account/licensing details
	AccountDetails AccountDetailsVerdict `json:"accountDetails"`

	// RequestHash is the hash of the request
	RequestHash string `json:"requestHash,omitempty"`

	// TimestampMillis is the verdict timestamp
	TimestampMillis int64 `json:"timestampMillis"`
}

// AppIntegrityVerdict contains app integrity information
type AppIntegrityVerdict struct {
	// AppRecognitionVerdict is the app recognition result
	// Values: UNRECOGNIZED_VERSION, UNEVALUATED, PLAY_RECOGNIZED
	AppRecognitionVerdict string `json:"appRecognitionVerdict"`

	// PackageName is the app package name
	PackageName string `json:"packageName,omitempty"`

	// CertificateSha256Digest are the signing certificate digests
	CertificateSha256Digest []string `json:"certificateSha256Digest,omitempty"`

	// VersionCode is the app version code
	VersionCode int64 `json:"versionCode,omitempty"`
}

// DeviceIntegrityVerdict contains device integrity information
type DeviceIntegrityVerdict struct {
	// DeviceRecognitionVerdict contains device integrity labels
	// Values: MEETS_BASIC_INTEGRITY, MEETS_DEVICE_INTEGRITY, MEETS_STRONG_INTEGRITY
	DeviceRecognitionVerdict []string `json:"deviceRecognitionVerdict"`
}

// AccountDetailsVerdict contains account information
type AccountDetailsVerdict struct {
	// AppLicensingVerdict is the licensing result
	// Values: LICENSED, UNLICENSED, UNEVALUATED
	AppLicensingVerdict string `json:"appLicensingVerdict"`
}

// MeetsBasicIntegrity checks if device meets basic integrity
func (v *DeviceIntegrityVerdict) MeetsBasicIntegrity() bool {
	for _, verdict := range v.DeviceRecognitionVerdict {
		if verdict == "MEETS_BASIC_INTEGRITY" {
			return true
		}
	}
	return false
}

// MeetsDeviceIntegrity checks if device meets device integrity
func (v *DeviceIntegrityVerdict) MeetsDeviceIntegrity() bool {
	for _, verdict := range v.DeviceRecognitionVerdict {
		if verdict == "MEETS_DEVICE_INTEGRITY" {
			return true
		}
	}
	return false
}

// MeetsStrongIntegrity checks if device meets strong integrity
func (v *DeviceIntegrityVerdict) MeetsStrongIntegrity() bool {
	for _, verdict := range v.DeviceRecognitionVerdict {
		if verdict == "MEETS_STRONG_INTEGRITY" {
			return true
		}
	}
	return false
}

// IsAppRecognized checks if the app is recognized by Play Store
func (v *AppIntegrityVerdict) IsAppRecognized() bool {
	return v.AppRecognitionVerdict == "PLAY_RECOGNIZED"
}

// IsLicensed checks if the app is licensed
func (v *AccountDetailsVerdict) IsLicensed() bool {
	return v.AppLicensingVerdict == "LICENSED"
}

// ============================================================================
// Google SafetyNet (Deprecated - for legacy support)
// ============================================================================

// SafetyNetResult contains the SafetyNet attestation result
type SafetyNetResult struct {
	// Nonce is the nonce used in the request
	Nonce string `json:"nonce"`

	// TimestampMs is when attestation was generated
	TimestampMs int64 `json:"timestampMs"`

	// ApkPackageName is the app package name
	ApkPackageName string `json:"apkPackageName"`

	// ApkDigestSha256 is the SHA256 of the APK
	ApkDigestSha256 string `json:"apkDigestSha256"`

	// CtsProfileMatch indicates CTS profile compatibility
	CtsProfileMatch bool `json:"ctsProfileMatch"`

	// BasicIntegrity indicates basic device integrity
	BasicIntegrity bool `json:"basicIntegrity"`

	// Advice contains remediation advice
	Advice string `json:"advice,omitempty"`

	// EvaluationType indicates evaluation type (BASIC, HARDWARE_BACKED)
	EvaluationType string `json:"evaluationType,omitempty"`
}

// ============================================================================
// Apple App Attest
// ============================================================================

// AppAttestResult contains the Apple App Attest result
type AppAttestResult struct {
	// KeyID is the App Attest key identifier
	KeyID string `json:"keyId"`

	// AttestationObject is the attestation object (CBOR-encoded)
	AttestationObject []byte `json:"attestationObject,omitempty"`

	// Verified indicates if the attestation was verified
	Verified bool `json:"verified"`

	// AppID is the app identifier
	AppID string `json:"appId"`

	// TeamID is the Apple Team ID
	TeamID string `json:"teamId"`

	// ReceiptType is the receipt type
	ReceiptType string `json:"receiptType"` // "production" or "sandbox"

	// CreationDate is when the attestation was created
	CreationDate time.Time `json:"creationDate"`

	// Risk assessment
	RiskMetric float64 `json:"riskMetric,omitempty"`
}

// ============================================================================
// Apple DeviceCheck
// ============================================================================

// DeviceCheckResult contains the Apple DeviceCheck result
type DeviceCheckResult struct {
	// DeviceToken is the DeviceCheck token
	DeviceToken string `json:"deviceToken,omitempty"`

	// Bit0 is the first persistent bit
	Bit0 bool `json:"bit0"`

	// Bit1 is the second persistent bit
	Bit1 bool `json:"bit1"`

	// LastUpdateTime is when bits were last updated
	LastUpdateTime time.Time `json:"lastUpdateTime,omitempty"`

	// TransactionID is the transaction identifier
	TransactionID string `json:"transactionId,omitempty"`
}

// ============================================================================
// Device Attestation Interface
// ============================================================================

// DeviceAttestationProvider defines the interface for device attestation
type DeviceAttestationProvider interface {
	// GetProvider returns the attestation provider type
	GetProvider() AttestationProvider

	// IsAvailable checks if attestation is available on this device
	IsAvailable() bool

	// Attest performs device attestation
	Attest(challenge []byte) (*DeviceAttestationResult, error)

	// GetKeyID returns the attestation key ID (for App Attest)
	GetKeyID() (string, error)

	// GenerateAssertion generates an assertion for a request (App Attest)
	GenerateAssertion(clientData []byte) ([]byte, error)
}

// ============================================================================
// Device Integrity Checker
// ============================================================================

// DeviceIntegrityChecker performs device integrity verification
type DeviceIntegrityChecker struct {
	// Platform is the device platform
	platform Platform

	// Config is the device requirements
	config DeviceRequirements

	// Provider is the attestation provider
	provider DeviceAttestationProvider

	// LastAttestation is the last attestation result
	lastAttestation *DeviceAttestationResult

	// Cache duration
	cacheDuration time.Duration
}

// NewDeviceIntegrityChecker creates a new device integrity checker
func NewDeviceIntegrityChecker(
	platform Platform,
	config DeviceRequirements,
	provider DeviceAttestationProvider,
) *DeviceIntegrityChecker {
	return &DeviceIntegrityChecker{
		platform:      platform,
		config:        config,
		provider:      provider,
		cacheDuration: 5 * time.Minute,
	}
}

// CheckIntegrity performs a full device integrity check
func (c *DeviceIntegrityChecker) CheckIntegrity(challenge []byte) (*DeviceAttestationResult, error) {
	// Check if we have a valid cached attestation
	if c.lastAttestation != nil && c.lastAttestation.IsValid() {
		// Check if cache is still fresh
		if time.Since(c.lastAttestation.Timestamp) < c.cacheDuration {
			return c.lastAttestation, nil
		}
	}

	// Perform new attestation
	result, err := c.attestDevice(challenge)
	if err != nil {
		return &DeviceAttestationResult{
			Provider:  c.provider.GetProvider(),
			Status:    IntegrityStatusError,
			Timestamp: time.Now(),
			Error:     err.Error(),
		}, err
	}

	// Cache result
	c.lastAttestation = result

	return result, nil
}

// attestDevice performs device attestation
func (c *DeviceIntegrityChecker) attestDevice(challenge []byte) (*DeviceAttestationResult, error) {
	if c.provider == nil {
		return &DeviceAttestationResult{
			Provider:  AttestationProviderNone,
			Status:    IntegrityStatusUnavailable,
			Timestamp: time.Now(),
			Error:     "no attestation provider configured",
		}, nil
	}

	if !c.provider.IsAvailable() {
		return &DeviceAttestationResult{
			Provider:  c.provider.GetProvider(),
			Status:    IntegrityStatusUnavailable,
			Timestamp: time.Now(),
			Error:     "attestation not available on this device",
		}, nil
	}

	return c.provider.Attest(challenge)
}

// EvaluateAttestation evaluates an attestation result against requirements
func (c *DeviceIntegrityChecker) EvaluateAttestation(result *DeviceAttestationResult) (bool, []IntegrityCheck) {
	var checks []IntegrityCheck
	allPassed := true

	// Check basic attestation status
	checks = append(checks, IntegrityCheck{
		Name:     "attestation_status",
		Passed:   result.Status == IntegrityStatusPassed,
		Required: true,
		Severity: SeverityCritical,
	})

	if result.Status != IntegrityStatusPassed {
		allPassed = false
	}

	// Platform-specific checks
	switch c.platform {
	case PlatformAndroid:
		androidChecks := c.evaluateAndroidAttestation(result)
		checks = append(checks, androidChecks...)
		for _, check := range androidChecks {
			if check.Required && !check.Passed {
				allPassed = false
			}
		}

	case PlatformIOS:
		iosChecks := c.evaluateIOSAttestation(result)
		checks = append(checks, iosChecks...)
		for _, check := range iosChecks {
			if check.Required && !check.Passed {
				allPassed = false
			}
		}
	}

	// Rooted/jailbroken check
	checks = append(checks, IntegrityCheck{
		Name:     "not_rooted",
		Passed:   !c.isRootedOrJailbroken(result),
		Required: !c.config.AllowRooted && !c.config.AllowJailbroken,
		Severity: SeverityCritical,
	})

	// Emulator check
	isEmulator := c.isEmulator(result)
	checks = append(checks, IntegrityCheck{
		Name:     "not_emulator",
		Passed:   !isEmulator,
		Required: !c.config.AllowEmulator,
		Severity: SeverityCritical,
	})

	if !c.config.AllowEmulator && isEmulator {
		allPassed = false
	}

	return allPassed, checks
}

// evaluateAndroidAttestation evaluates Android-specific attestation
func (c *DeviceIntegrityChecker) evaluateAndroidAttestation(result *DeviceAttestationResult) []IntegrityCheck {
	var checks []IntegrityCheck

	if result.PlayIntegrityVerdict != nil {
		verdict := result.PlayIntegrityVerdict

		// Basic integrity
		checks = append(checks, IntegrityCheck{
			Name:     "basic_integrity",
			Passed:   verdict.DeviceIntegrity.MeetsBasicIntegrity(),
			Required: true,
			Severity: SeverityCritical,
		})

		// Device integrity
		checks = append(checks, IntegrityCheck{
			Name:     "device_integrity",
			Passed:   verdict.DeviceIntegrity.MeetsDeviceIntegrity(),
			Required: true,
			Severity: SeverityHigh,
		})

		// Strong integrity (optional)
		checks = append(checks, IntegrityCheck{
			Name:     "strong_integrity",
			Passed:   verdict.DeviceIntegrity.MeetsStrongIntegrity(),
			Required: false,
			Severity: SeverityMedium,
		})

		// App recognition
		checks = append(checks, IntegrityCheck{
			Name:     "app_recognized",
			Passed:   verdict.AppIntegrity.IsAppRecognized(),
			Required: true,
			Severity: SeverityHigh,
		})

		// App licensed
		checks = append(checks, IntegrityCheck{
			Name:     "app_licensed",
			Passed:   verdict.AccountDetails.IsLicensed(),
			Required: false,
			Severity: SeverityLow,
		})
	}

	if result.SafetyNetResult != nil {
		sn := result.SafetyNetResult

		checks = append(checks, IntegrityCheck{
			Name:     "safetynet_basic",
			Passed:   sn.BasicIntegrity,
			Required: true,
			Severity: SeverityCritical,
		})

		checks = append(checks, IntegrityCheck{
			Name:     "safetynet_cts",
			Passed:   sn.CtsProfileMatch,
			Required: true,
			Severity: SeverityHigh,
		})
	}

	return checks
}

// evaluateIOSAttestation evaluates iOS-specific attestation
func (c *DeviceIntegrityChecker) evaluateIOSAttestation(result *DeviceAttestationResult) []IntegrityCheck {
	var checks []IntegrityCheck

	if result.AppAttestResult != nil {
		attest := result.AppAttestResult

		checks = append(checks, IntegrityCheck{
			Name:     "app_attest_verified",
			Passed:   attest.Verified,
			Required: true,
			Severity: SeverityCritical,
		})

		checks = append(checks, IntegrityCheck{
			Name:     "production_environment",
			Passed:   attest.ReceiptType == "production",
			Required: true,
			Severity: SeverityHigh,
		})

		// Low risk check
		checks = append(checks, IntegrityCheck{
			Name:     "low_risk",
			Passed:   attest.RiskMetric < 0.3,
			Required: false,
			Severity: SeverityMedium,
		})
	}

	return checks
}

// isRootedOrJailbroken checks if the device appears rooted/jailbroken
func (c *DeviceIntegrityChecker) isRootedOrJailbroken(result *DeviceAttestationResult) bool {
	switch c.platform {
	case PlatformAndroid:
		if result.PlayIntegrityVerdict != nil {
			// No basic integrity = likely rooted
			return !result.PlayIntegrityVerdict.DeviceIntegrity.MeetsBasicIntegrity()
		}
		if result.SafetyNetResult != nil {
			return !result.SafetyNetResult.BasicIntegrity
		}
	case PlatformIOS:
		// For iOS, App Attest failure in production typically indicates jailbreak
		if result.AppAttestResult != nil {
			return !result.AppAttestResult.Verified && result.AppAttestResult.ReceiptType == "production"
		}
	}
	return false
}

// isEmulator checks if running on an emulator
func (c *DeviceIntegrityChecker) isEmulator(result *DeviceAttestationResult) bool {
	switch c.platform {
	case PlatformAndroid:
		if result.PlayIntegrityVerdict != nil {
			// No device integrity typically indicates emulator
			return !result.PlayIntegrityVerdict.DeviceIntegrity.MeetsDeviceIntegrity()
		}
	case PlatformIOS:
		// iOS simulator won't have App Attest at all
		if result.AppAttestResult == nil && result.Provider == AttestationProviderAppAttest {
			return true
		}
	}
	return false
}

// ============================================================================
// Attestation Token Utilities
// ============================================================================

// GenerateAttestationChallenge generates a challenge for attestation
func GenerateAttestationChallenge(sessionID string, deviceFingerprint string, timestamp int64) []byte {
	h := sha256.New()
	h.Write([]byte(sessionID))
	h.Write([]byte(deviceFingerprint))
	h.Write(int64ToBytes(timestamp))
	return h.Sum(nil)
}

// VerifyAttestationTimestamp verifies the attestation is recent
func VerifyAttestationTimestamp(attestTimestamp int64, maxAgeSeconds int64) bool {
	now := time.Now().Unix()
	age := now - attestTimestamp
	return age >= 0 && age <= maxAgeSeconds
}

// HashToken hashes an attestation token for logging
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// ============================================================================
// Serialization Helpers
// ============================================================================

// MarshalAttestationResult marshals an attestation result to JSON
func MarshalAttestationResult(result *DeviceAttestationResult) ([]byte, error) {
	return json.Marshal(result)
}

// UnmarshalAttestationResult unmarshals an attestation result from JSON
func UnmarshalAttestationResult(data []byte) (*DeviceAttestationResult, error) {
	var result DeviceAttestationResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attestation result: %w", err)
	}
	return &result, nil
}
