package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "encryption"

	// StoreKey is the store key string for encryption module
	StoreKey = ModuleName

	// RouterKey is the message route for encryption module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for encryption module
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// PrefixRecipientKey is the prefix for recipient public key storage
	// Key: PrefixRecipientKey | address | fingerprint -> RecipientKeyRecord
	PrefixRecipientKey = []byte{0x01}

	// PrefixKeyByFingerprint is the prefix for key lookup by fingerprint
	// Key: PrefixKeyByFingerprint | fingerprint -> address
	PrefixKeyByFingerprint = []byte{0x02}

	// PrefixParams is the prefix for module parameters
	PrefixParams = []byte{0x03}

	// PrefixEnvelopeLog is the prefix for envelope validation log (optional audit)
	// Key: PrefixEnvelopeLog | hash -> EnvelopeLogEntry
	PrefixEnvelopeLog = []byte{0x04}

	// PrefixActiveKey is the prefix for active key lookup per address
	// Key: PrefixActiveKey | address -> fingerprint
	PrefixActiveKey = []byte{0x05}
)

// RecipientKeyKey returns the store key for a recipient's public key
func RecipientKeyKey(address []byte, fingerprint []byte) []byte {
	key := make([]byte, 0, len(PrefixRecipientKey)+len(address)+len(fingerprint))
	key = append(key, PrefixRecipientKey...)
	key = append(key, address...)
	key = append(key, fingerprint...)
	return key
}

// RecipientKeyPrefix returns the prefix for all keys owned by an address.
func RecipientKeyPrefix(address []byte) []byte {
	key := make([]byte, 0, len(PrefixRecipientKey)+len(address))
	key = append(key, PrefixRecipientKey...)
	key = append(key, address...)
	return key
}

// ActiveKeyKey returns the store key for an address's active key fingerprint.
func ActiveKeyKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixActiveKey)+len(address))
	key = append(key, PrefixActiveKey...)
	key = append(key, address...)
	return key
}

// KeyByFingerprintKey returns the store key for looking up address by key fingerprint
func KeyByFingerprintKey(fingerprint []byte) []byte {
	key := make([]byte, 0, len(PrefixKeyByFingerprint)+len(fingerprint))
	key = append(key, PrefixKeyByFingerprint...)
	key = append(key, fingerprint...)
	return key
}

// ParamsKey returns the store key for module parameters
func ParamsKey() []byte {
	return PrefixParams
}

// EnvelopeLogKey returns the store key for an envelope log entry
func EnvelopeLogKey(envelopeHash []byte) []byte {
	key := make([]byte, 0, len(PrefixEnvelopeLog)+len(envelopeHash))
	key = append(key, PrefixEnvelopeLog...)
	key = append(key, envelopeHash...)
	return key
}
