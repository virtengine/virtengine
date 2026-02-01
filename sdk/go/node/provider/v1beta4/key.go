package v1beta4

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "provider"

	// StoreKey is the store key string for provider
	StoreKey = ModuleName

	// RouterKey is the message route for provider
	RouterKey = ModuleName

	// Supported public key types for provider cryptographic operations
	PublicKeyTypeEd25519   = "ed25519"
	PublicKeyTypeX25519    = "x25519"
	PublicKeyTypeSecp256k1 = "secp256k1"
)

// ProviderPrefix returns the key prefix for provider storage
func ProviderPrefix() []byte {
	return []byte{0x01}
}

// ProviderPublicKeyPrefix returns the key prefix for provider public key storage
func ProviderPublicKeyPrefix() []byte {
	return []byte{0x02}
}

// DomainVerificationPrefix returns the key prefix for domain verification storage
func DomainVerificationPrefix() []byte {
	return []byte{0x03}
}

