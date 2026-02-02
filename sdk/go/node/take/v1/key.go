package v1

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "take"

	// StoreKey is the store key string for deployment
	StoreKey = ModuleName

	// RouterKey is the message route for deployment
	RouterKey = ModuleName
)

func ParamsPrefix() []byte {
	return []byte{0x01}
}
