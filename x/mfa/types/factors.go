package types

import (
	"fmt"
)

// FactorType represents the type of authentication factor
type FactorType uint8

const (
	// FactorTypeUnspecified represents an unspecified factor type
	FactorTypeUnspecified FactorType = 0
	// FactorTypeTOTP represents Time-based One-Time Password
	FactorTypeTOTP FactorType = 1
	// FactorTypeFIDO2 represents FIDO2/WebAuthn authentication
	FactorTypeFIDO2 FactorType = 2
	// FactorTypeSMS represents SMS OTP authentication
	FactorTypeSMS FactorType = 3
	// FactorTypeEmail represents Email OTP authentication
	FactorTypeEmail FactorType = 4
	// FactorTypeVEID represents VEID identity score threshold
	FactorTypeVEID FactorType = 5
	// FactorTypeTrustedDevice represents trusted browser/device binding
	FactorTypeTrustedDevice FactorType = 6
	// FactorTypeHardwareKey represents X.509/smart card/PIV hardware key authentication (VE-925)
	FactorTypeHardwareKey FactorType = 7
)

// FactorTypeNames maps factor types to human-readable names
var FactorTypeNames = map[FactorType]string{
	FactorTypeUnspecified:   "unspecified",
	FactorTypeTOTP:          "totp",
	FactorTypeFIDO2:         "fido2",
	FactorTypeSMS:           "sms",
	FactorTypeEmail:         "email",
	FactorTypeVEID:          "veid",
	FactorTypeTrustedDevice: "trusted_device",
	FactorTypeHardwareKey:   "hardware_key",
}

// FactorTypeFromString converts a string to FactorType
func FactorTypeFromString(s string) (FactorType, error) {
	for ft, name := range FactorTypeNames {
		if name == s {
			return ft, nil
		}
	}
	return FactorTypeUnspecified, fmt.Errorf("unknown factor type: %s", s)
}

// String returns the string representation of a FactorType
func (ft FactorType) String() string {
	if name, ok := FactorTypeNames[ft]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", ft)
}

// IsValid returns true if the factor type is valid
func (ft FactorType) IsValid() bool {
	return ft >= FactorTypeTOTP && ft <= FactorTypeHardwareKey
}

// RequiresOffChainVerification returns true if the factor requires off-chain verification
func (ft FactorType) RequiresOffChainVerification() bool {
	switch ft {
	case FactorTypeTOTP, FactorTypeFIDO2, FactorTypeSMS, FactorTypeEmail, FactorTypeHardwareKey:
		return true
	case FactorTypeVEID, FactorTypeTrustedDevice:
		return false
	default:
		return false
	}
}

// FactorSecurityLevel represents the security level of a factor
type FactorSecurityLevel uint8

const (
	// FactorSecurityLevelLow represents low security factors (email, SMS)
	FactorSecurityLevelLow FactorSecurityLevel = 1
	// FactorSecurityLevelMedium represents medium security factors (TOTP)
	FactorSecurityLevelMedium FactorSecurityLevel = 2
	// FactorSecurityLevelHigh represents high security factors (FIDO2, VEID)
	FactorSecurityLevelHigh FactorSecurityLevel = 3
)

// GetSecurityLevel returns the security level of a factor type
func (ft FactorType) GetSecurityLevel() FactorSecurityLevel {
	switch ft {
	case FactorTypeFIDO2, FactorTypeVEID, FactorTypeHardwareKey:
		return FactorSecurityLevelHigh
	case FactorTypeTOTP, FactorTypeTrustedDevice:
		return FactorSecurityLevelMedium
	case FactorTypeSMS, FactorTypeEmail:
		return FactorSecurityLevelLow
	default:
		return FactorSecurityLevelLow
	}
}

// FactorInfo provides metadata about a factor type
type FactorInfo struct {
	// Type is the factor type
	Type FactorType
	// Name is the human-readable name
	Name string
	// Description provides details about the factor
	Description string
	// SecurityLevel indicates the factor's security strength
	SecurityLevel FactorSecurityLevel
	// RequiresOffChain indicates if verification happens off-chain
	RequiresOffChain bool
	// CanBeRecoveryFactor indicates if this factor can be used for account recovery
	CanBeRecoveryFactor bool
}

// AllFactorTypes returns all valid factor types
func AllFactorTypes() []FactorType {
	return []FactorType{
		FactorTypeTOTP,
		FactorTypeFIDO2,
		FactorTypeSMS,
		FactorTypeEmail,
		FactorTypeVEID,
		FactorTypeTrustedDevice,
		FactorTypeHardwareKey,
	}
}

// GetFactorInfo returns detailed information about a factor type
func GetFactorInfo(ft FactorType) FactorInfo {
	switch ft {
	case FactorTypeTOTP:
		return FactorInfo{
			Type:                ft,
			Name:                "Time-based OTP",
			Description:         "Time-based One-Time Password using TOTP algorithm (RFC 6238)",
			SecurityLevel:       FactorSecurityLevelMedium,
			RequiresOffChain:    true,
			CanBeRecoveryFactor: true,
		}
	case FactorTypeFIDO2:
		return FactorInfo{
			Type:                ft,
			Name:                "FIDO2/WebAuthn",
			Description:         "Hardware security key or platform authenticator using FIDO2/WebAuthn standard",
			SecurityLevel:       FactorSecurityLevelHigh,
			RequiresOffChain:    true,
			CanBeRecoveryFactor: true,
		}
	case FactorTypeSMS:
		return FactorInfo{
			Type:                ft,
			Name:                "SMS OTP",
			Description:         "One-Time Password sent via SMS to registered phone number",
			SecurityLevel:       FactorSecurityLevelLow,
			RequiresOffChain:    true,
			CanBeRecoveryFactor: true,
		}
	case FactorTypeEmail:
		return FactorInfo{
			Type:                ft,
			Name:                "Email OTP",
			Description:         "One-Time Password sent via email to registered address",
			SecurityLevel:       FactorSecurityLevelLow,
			RequiresOffChain:    true,
			CanBeRecoveryFactor: true,
		}
	case FactorTypeVEID:
		return FactorInfo{
			Type:                ft,
			Name:                "VEID Score",
			Description:         "VirtEngine Identity verification score threshold check",
			SecurityLevel:       FactorSecurityLevelHigh,
			RequiresOffChain:    false,
			CanBeRecoveryFactor: false, // VEID is derived from identity, not a standalone factor
		}
	case FactorTypeTrustedDevice:
		return FactorInfo{
			Type:                ft,
			Name:                "Trusted Device",
			Description:         "Browser or device binding for trusted session management",
			SecurityLevel:       FactorSecurityLevelMedium,
			RequiresOffChain:    false,
			CanBeRecoveryFactor: false,
		}
	case FactorTypeHardwareKey:
		return FactorInfo{
			Type:                ft,
			Name:                "Hardware Key",
			Description:         "X.509 certificate, smart card, or PIV-based hardware key authentication (VE-925)",
			SecurityLevel:       FactorSecurityLevelHigh,
			RequiresOffChain:    true,
			CanBeRecoveryFactor: true,
		}
	default:
		return FactorInfo{
			Type:                ft,
			Name:                "Unknown",
			Description:         "Unknown factor type",
			SecurityLevel:       FactorSecurityLevelLow,
			RequiresOffChain:    false,
			CanBeRecoveryFactor: false,
		}
	}
}

// FactorVerificationMethod represents how a factor is verified
type FactorVerificationMethod uint8

const (
	// VerificationMethodNone indicates no verification method
	VerificationMethodNone FactorVerificationMethod = 0
	// VerificationMethodOTP indicates OTP-based verification
	VerificationMethodOTP FactorVerificationMethod = 1
	// VerificationMethodSignature indicates cryptographic signature verification
	VerificationMethodSignature FactorVerificationMethod = 2
	// VerificationMethodThreshold indicates threshold-based verification (e.g., VEID score)
	VerificationMethodThreshold FactorVerificationMethod = 3
	// VerificationMethodBinding indicates device binding verification
	VerificationMethodBinding FactorVerificationMethod = 4
)

// GetVerificationMethod returns the verification method for a factor type
func (ft FactorType) GetVerificationMethod() FactorVerificationMethod {
	switch ft {
	case FactorTypeTOTP, FactorTypeSMS, FactorTypeEmail:
		return VerificationMethodOTP
	case FactorTypeFIDO2, FactorTypeHardwareKey:
		return VerificationMethodSignature
	case FactorTypeVEID:
		return VerificationMethodThreshold
	case FactorTypeTrustedDevice:
		return VerificationMethodBinding
	default:
		return VerificationMethodNone
	}
}
