package types

const (
	ModuleName = "veidregistry"
	// StoreKey must not share a prefix with the existing veid module store key.
	StoreKey     = "vreg"
	RouterKey    = ModuleName
	QuerierRoute = ModuleName
)

var (
	PrefixVerifierVersion    = []byte{0x01}
	PrefixActiveVerifier     = []byte{0x02}
	PrefixValidatorReadiness = []byte{0x03}
	PrefixParams             = []byte{0x04}
)

func VerifierVersionKey(verifierID string) []byte {
	return append(append([]byte{}, PrefixVerifierVersion...), []byte(verifierID)...)
}

func VerifierVersionPrefixKey() []byte {
	return PrefixVerifierVersion
}

func ActiveVerifierKey() []byte {
	return PrefixActiveVerifier
}

func ValidatorReadinessKey(verifierID, validatorAddress string) []byte {
	key := append(append([]byte{}, PrefixValidatorReadiness...), []byte(verifierID)...)
	key = append(key, byte('/'))
	key = append(key, []byte(validatorAddress)...)
	return key
}

func ValidatorReadinessPrefixKey(verifierID string) []byte {
	key := append([]byte{}, PrefixValidatorReadiness...)
	key = append(key, []byte(verifierID)...)
	key = append(key, byte('/'))
	return key
}

func ParamsKey() []byte {
	return PrefixParams
}
