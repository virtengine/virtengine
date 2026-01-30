package errors

// Error code allocation for VirtEngine modules.
//
// All VirtEngine modules use error codes starting from 100 to avoid conflicts with:
// - Cosmos SDK core modules (codes 1-50)
// - IBC-Go modules (codes 1-50)
// - CosmWasm modules (codes 1-50)
//
// Each module is allocated a range of 100 error codes.

// ModuleCodeRange defines the error code range for a module.
type ModuleCodeRange struct {
	Module      string
	StartCode   uint32
	EndCode     uint32
	Description string
}

// AllModuleRanges contains all allocated error code ranges.
var AllModuleRanges = []ModuleCodeRange{
	// x/ blockchain modules (on-chain consensus)
	{Module: "veid", StartCode: 1000, EndCode: 1099, Description: "Identity verification and ML scoring"},
	{Module: "mfa", StartCode: 1200, EndCode: 1299, Description: "Multi-factor authentication"},
	{Module: "encryption", StartCode: 1300, EndCode: 1399, Description: "Encryption and key management"},
	{Module: "market", StartCode: 1400, EndCode: 1499, Description: "Marketplace orders and bids"},
	{Module: "escrow", StartCode: 1500, EndCode: 1599, Description: "Payment escrow"},
	{Module: "roles", StartCode: 1600, EndCode: 1699, Description: "Role-based access control"},
	{Module: "hpc", StartCode: 1700, EndCode: 1799, Description: "High-performance computing"},
	{Module: "provider", StartCode: 1800, EndCode: 1899, Description: "Provider registration and management"},
	{Module: "deployment", StartCode: 1900, EndCode: 1999, Description: "Deployment management"},
	{Module: "cert", StartCode: 2000, EndCode: 2099, Description: "Certificate management"},
	{Module: "audit", StartCode: 2100, EndCode: 2199, Description: "Audit logging"},
	{Module: "settlement", StartCode: 2200, EndCode: 2299, Description: "Payment settlement"},
	{Module: "benchmark", StartCode: 2300, EndCode: 2399, Description: "Provider benchmarking"},
	{Module: "staking", StartCode: 2400, EndCode: 2499, Description: "Staking and rewards"},
	{Module: "delegation", StartCode: 2500, EndCode: 2599, Description: "Stake delegation"},
	{Module: "fraud", StartCode: 2600, EndCode: 2699, Description: "Fraud detection"},
	{Module: "review", StartCode: 2700, EndCode: 2799, Description: "Provider reviews"},
	{Module: "enclave", StartCode: 2800, EndCode: 2899, Description: "Trusted execution environments"},
	{Module: "config", StartCode: 2900, EndCode: 2999, Description: "On-chain configuration"},
	{Module: "take", StartCode: 3000, EndCode: 3099, Description: "Fee distribution"},
	{Module: "marketplace", StartCode: 3100, EndCode: 3199, Description: "Marketplace integration"},

	// pkg/ off-chain services (not consensus-critical)
	{Module: "provider_daemon", StartCode: 100, EndCode: 199, Description: "Provider daemon service"},
	{Module: "inference", StartCode: 200, EndCode: 299, Description: "ML inference service"},
	{Module: "workflow", StartCode: 300, EndCode: 399, Description: "Workflow engine"},
	{Module: "benchmark_daemon", StartCode: 400, EndCode: 499, Description: "Benchmark daemon"},
	{Module: "enclave_runtime", StartCode: 500, EndCode: 599, Description: "Enclave runtime"},
	{Module: "waldur", StartCode: 600, EndCode: 699, Description: "Waldur integration"},
	{Module: "govdata", StartCode: 700, EndCode: 799, Description: "Government data verification"},
	{Module: "edugain", StartCode: 800, EndCode: 899, Description: "EduGAIN federation"},
	{Module: "nli", StartCode: 900, EndCode: 999, Description: "Natural language interface"},
	{Module: "artifact_store", StartCode: 3200, EndCode: 3299, Description: "Artifact storage"},
	{Module: "capture_protocol", StartCode: 3300, EndCode: 3399, Description: "Identity capture protocol"},
	{Module: "payment", StartCode: 3400, EndCode: 3499, Description: "Payment processing"},
	{Module: "dex", StartCode: 3500, EndCode: 3599, Description: "DEX integration"},
	{Module: "jira", StartCode: 3600, EndCode: 3699, Description: "JIRA integration"},
	{Module: "slurm_adapter", StartCode: 3700, EndCode: 3799, Description: "SLURM adapter"},
	{Module: "ood_adapter", StartCode: 3800, EndCode: 3899, Description: "Open OnDemand adapter"},
	{Module: "moab_adapter", StartCode: 3900, EndCode: 3999, Description: "MOAB adapter"},
	{Module: "sre", StartCode: 4000, EndCode: 4099, Description: "SRE tooling"},
	{Module: "observability", StartCode: 4100, EndCode: 4199, Description: "Observability"},
	{Module: "ratelimit", StartCode: 4200, EndCode: 4299, Description: "Rate limiting"},
}

// GetModuleRange returns the error code range for a module.
func GetModuleRange(module string) (ModuleCodeRange, bool) {
	for _, r := range AllModuleRanges {
		if r.Module == module {
			return r, true
		}
	}
	return ModuleCodeRange{}, false
}

// ValidateCode checks if an error code is within the allocated range for a module.
func ValidateCode(module string, code uint32) bool {
	r, ok := GetModuleRange(module)
	if !ok {
		return false
	}
	return code >= r.StartCode && code <= r.EndCode
}

// Common error code categories within each module range.
// Modules should follow this pattern for consistency:
const (
	// 00-09: Invalid input/validation errors
	CodeInvalidInput = 0
	CodeInvalidAddress = 1
	CodeInvalidParams = 2

	// 10-19: Not found errors
	CodeNotFound = 10
	CodeResourceNotFound = 11

	// 20-29: Already exists/conflict errors
	CodeAlreadyExists = 20
	CodeConflict = 21

	// 30-39: Unauthorized/permission errors
	CodeUnauthorized = 30
	CodeForbidden = 31
	CodeInsufficientPermissions = 32

	// 40-49: State/lifecycle errors
	CodeInvalidState = 40
	CodeExpired = 41
	CodeRevoked = 42
	CodeLocked = 43

	// 50-59: External service errors
	CodeExternalServiceError = 50
	CodeTimeout = 51
	CodeUnavailable = 52

	// 60-69: Internal errors
	CodeInternal = 60
	CodeInferenceFailed = 61
	CodeEncryptionFailed = 62
	CodeDecryptionFailed = 63

	// 70-79: Verification/validation errors
	CodeVerificationFailed = 70
	CodeSignatureInvalid = 71
	CodeChallengeInvalid = 72

	// 80-89: Rate limiting/quota errors
	CodeRateLimitExceeded = 80
	CodeQuotaExceeded = 81
	CodeMaxRetriesExceeded = 82

	// 90-99: Reserved for future use
)
