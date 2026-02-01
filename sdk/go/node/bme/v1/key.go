package v1

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "bme"

	// StoreKey is the store key string for bme
	StoreKey = ModuleName

	// RouterKey is the message route for bme
	RouterKey = ModuleName
)

// ParamsPrefix returns the prefix for storing module params
func ParamsPrefix() []byte {
	return []byte{0x01, 0x00}
}

// StatePrefix returns the prefix for storing vault state
func StatePrefix() []byte {
	return []byte{0x02, 0x00}
}

// LedgerRecordPrefix returns the prefix for storing ledger records
func LedgerRecordPrefix() []byte {
	return []byte{0x03, 0x00}
}

// LedgerPendingRecordPrefix returns the prefix for storing pending ledger records
func LedgerPendingRecordPrefix() []byte {
	return []byte{0x03, 0x01}
}

