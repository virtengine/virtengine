// Package v1 contains the enclave module types for CLI usage.
// These types mirror x/enclave/types for use in the SDK CLI.
package v1

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "enclave"
)

// TEEType represents the type of Trusted Execution Environment
type TEEType string

const (
	// TEETypeSGX is Intel SGX
	TEETypeSGX TEEType = "SGX"

	// TEETypeSEVSNP is AMD SEV-SNP
	TEETypeSEVSNP TEEType = "SEV-SNP"

	// TEETypeNitro is AWS Nitro Enclaves
	TEETypeNitro TEEType = "NITRO"

	// TEETypeTrustZone is ARM TrustZone
	TEETypeTrustZone TEEType = "TRUSTZONE"
)

// AllTEETypes returns all valid TEE types
func AllTEETypes() []TEEType {
	return []TEEType{TEETypeSGX, TEETypeSEVSNP, TEETypeNitro, TEETypeTrustZone}
}

// IsValidTEEType checks if a TEE type is valid
func IsValidTEEType(teeType TEEType) bool {
	for _, t := range AllTEETypes() {
		if t == teeType {
			return true
		}
	}
	return false
}
