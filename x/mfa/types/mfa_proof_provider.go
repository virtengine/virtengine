package types

// MFAProofProvider exposes MFA proof data for sensitive transactions.
// Modules can implement this on messages that require MFA validation.
type MFAProofProvider interface {
	GetMFAProof() *MFAProof
	GetDeviceFingerprint() string
}
