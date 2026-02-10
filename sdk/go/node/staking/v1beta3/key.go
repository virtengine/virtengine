package v1beta3

const (
	// ModuleName is the name of the module
	ModuleName = "astaking"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName
)

func ParamsPrefix() []byte {
	return []byte{0x01}
}
