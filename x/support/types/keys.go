package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "support"

	// StoreKey is the store key string for support module
	StoreKey = ModuleName

	// RouterKey is the message route for support module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for support module
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// PrefixExternalRef is the prefix for external ticket reference storage
	// Key: PrefixExternalRef | resource_type | "/" | resource_id -> ExternalTicketRef
	PrefixExternalRef = []byte{0x01}

	// PrefixExternalRefByOwner is the prefix for owner-based index
	// Key: PrefixExternalRefByOwner | owner_address | "/" | resource_type | "/" | resource_id -> bool
	PrefixExternalRefByOwner = []byte{0x02}

	// PrefixParams is the prefix for module parameters
	PrefixParams = []byte{0x03}
)

// ExternalRefKey returns the store key for an external ticket reference
func ExternalRefKey(resourceType ResourceType, resourceID string) []byte {
	key := make([]byte, 0, len(PrefixExternalRef)+len(resourceType)+len(resourceID)+1)
	key = append(key, PrefixExternalRef...)
	key = append(key, []byte(resourceType)...)
	key = append(key, '/')
	key = append(key, []byte(resourceID)...)
	return key
}

// ExternalRefPrefixKey returns the prefix for external refs of a resource type
func ExternalRefPrefixKey(resourceType ResourceType) []byte {
	key := make([]byte, 0, len(PrefixExternalRef)+len(resourceType)+1)
	key = append(key, PrefixExternalRef...)
	key = append(key, []byte(resourceType)...)
	key = append(key, '/')
	return key
}

// ExternalRefByOwnerKey returns the store key for owner-based index
func ExternalRefByOwnerKey(ownerAddr []byte, resourceType ResourceType, resourceID string) []byte {
	key := make([]byte, 0, len(PrefixExternalRefByOwner)+len(ownerAddr)+len(resourceType)+len(resourceID)+2)
	key = append(key, PrefixExternalRefByOwner...)
	key = append(key, ownerAddr...)
	key = append(key, '/')
	key = append(key, []byte(resourceType)...)
	key = append(key, '/')
	key = append(key, []byte(resourceID)...)
	return key
}

// ExternalRefByOwnerPrefixKey returns the prefix for an owner's refs
func ExternalRefByOwnerPrefixKey(ownerAddr []byte) []byte {
	key := make([]byte, 0, len(PrefixExternalRefByOwner)+len(ownerAddr)+1)
	key = append(key, PrefixExternalRefByOwner...)
	key = append(key, ownerAddr...)
	key = append(key, '/')
	return key
}

// ParamsKey returns the store key for module parameters
func ParamsKey() []byte {
	return PrefixParams
}
