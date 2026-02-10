package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "config"

	// StoreKey is the store key string for config module
	StoreKey = ModuleName

	// RouterKey is the message route for config module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for config module
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// PrefixApprovedClient is the prefix for approved client storage
	// Key: PrefixApprovedClient | client_id -> ApprovedClient
	PrefixApprovedClient = []byte{0x01}

	// PrefixApprovedClientByStatus is the prefix for approved client index by status
	// Key: PrefixApprovedClientByStatus | status | client_id -> []byte{1}
	PrefixApprovedClientByStatus = []byte{0x02}

	// PrefixParams is the prefix for module parameters
	PrefixParams = []byte{0x03}

	// PrefixClientAuditLog is the prefix for client audit log
	// Key: PrefixClientAuditLog | client_id | timestamp -> AuditEntry
	PrefixClientAuditLog = []byte{0x04}
)

// ApprovedClientKey returns the store key for an approved client
func ApprovedClientKey(clientID string) []byte {
	clientIDBytes := []byte(clientID)
	key := make([]byte, 0, len(PrefixApprovedClient)+len(clientIDBytes))
	key = append(key, PrefixApprovedClient...)
	key = append(key, clientIDBytes...)
	return key
}

// ApprovedClientPrefixKey returns the prefix for all approved clients
func ApprovedClientPrefixKey() []byte {
	return PrefixApprovedClient
}

// ApprovedClientByStatusKey returns the store key for approved client status index
func ApprovedClientByStatusKey(status ClientStatus, clientID string) []byte {
	statusBytes := []byte(status)
	clientIDBytes := []byte(clientID)
	key := make([]byte, 0, len(PrefixApprovedClientByStatus)+len(statusBytes)+1+len(clientIDBytes))
	key = append(key, PrefixApprovedClientByStatus...)
	key = append(key, statusBytes...)
	key = append(key, byte('/'))
	key = append(key, clientIDBytes...)
	return key
}

// ApprovedClientByStatusPrefixKey returns the prefix for all clients with a status
func ApprovedClientByStatusPrefixKey(status ClientStatus) []byte {
	statusBytes := []byte(status)
	key := make([]byte, 0, len(PrefixApprovedClientByStatus)+len(statusBytes)+1)
	key = append(key, PrefixApprovedClientByStatus...)
	key = append(key, statusBytes...)
	key = append(key, byte('/'))
	return key
}

// ParamsKey returns the store key for module parameters
func ParamsKey() []byte {
	return PrefixParams
}

// ClientAuditLogKey returns the store key for a client audit log entry
func ClientAuditLogKey(clientID string, timestamp int64) []byte {
	clientIDBytes := []byte(clientID)
	key := make([]byte, 0, len(PrefixClientAuditLog)+len(clientIDBytes)+1+8)
	key = append(key, PrefixClientAuditLog...)
	key = append(key, clientIDBytes...)
	key = append(key, byte('/'))
	key = append(key, encodeInt64(timestamp)...)
	return key
}

// ClientAuditLogPrefixKey returns the prefix for all audit logs of a client
func ClientAuditLogPrefixKey(clientID string) []byte {
	clientIDBytes := []byte(clientID)
	key := make([]byte, 0, len(PrefixClientAuditLog)+len(clientIDBytes)+1)
	key = append(key, PrefixClientAuditLog...)
	key = append(key, clientIDBytes...)
	key = append(key, byte('/'))
	return key
}

// encodeInt64 encodes an int64 as big-endian bytes for proper ordering
func encodeInt64(n int64) []byte {
	return []byte{
		byte(n >> 56),
		byte(n >> 48),
		byte(n >> 40),
		byte(n >> 32),
		byte(n >> 24),
		byte(n >> 16),
		byte(n >> 8),
		byte(n),
	}
}
