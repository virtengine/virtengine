package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// ComputeFactorFingerprint computes a unique fingerprint for a factor enrollment
// based on the factor type and public credential data.
// IMPORTANT: This should NEVER include secret data like TOTP seeds.
func ComputeFactorFingerprint(factorType FactorType, publicCredentialData []byte) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d:", factorType)))
	h.Write(publicCredentialData)
	return hex.EncodeToString(h.Sum(nil))[:32] // Return first 32 hex chars (16 bytes)
}

// ComputeDeviceFingerprint computes a unique fingerprint for a trusted device
// based on device metadata.
func ComputeDeviceFingerprint(deviceName string, devicePublicKey []byte) string {
	h := sha256.New()
	h.Write([]byte(deviceName))
	h.Write(devicePublicKey)
	return hex.EncodeToString(h.Sum(nil))[:32]
}

// ComputeChallengeID computes a unique challenge ID
func ComputeChallengeID(address string, factorType FactorType, timestamp int64) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%d:%d", address, factorType, timestamp)))
	return hex.EncodeToString(h.Sum(nil))[:24] // Return first 24 hex chars (12 bytes)
}

// ComputeSessionID computes a unique session ID
func ComputeSessionID(address string, timestamp int64, nonce []byte) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%d:", address, timestamp)))
	h.Write(nonce)
	return hex.EncodeToString(h.Sum(nil))[:32]
}
