package types

import (
	"encoding/binary"
	"fmt"
)

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

	// PrefixSupportRequest is the prefix for support requests
	// Key: PrefixSupportRequest | request_id -> SupportRequest
	PrefixSupportRequest = []byte{0x10}

	// PrefixSupportRequestBySubmitter is the prefix for submitter index
	// Key: PrefixSupportRequestBySubmitter | submitter_addr | "/" | request_id -> bool
	PrefixSupportRequestBySubmitter = []byte{0x11}

	// PrefixSupportRequestByStatus is the prefix for status index
	// Key: PrefixSupportRequestByStatus | status | "/" | request_id -> bool
	PrefixSupportRequestByStatus = []byte{0x12}

	// PrefixSupportResponse is the prefix for support responses
	// Key: PrefixSupportResponse | request_id | "/" | sequence -> SupportResponse
	PrefixSupportResponse = []byte{0x13}

	// PrefixSupportResponseByRequest is the prefix for response lookup by request
	// Key: PrefixSupportResponseByRequest | request_id | "/" | sequence -> bool
	PrefixSupportResponseByRequest = []byte{0x14}

	// PrefixSupportRequestSequence is the prefix for submitter-scoped ticket sequences
	// Key: PrefixSupportRequestSequence | submitter_addr -> uint64
	PrefixSupportRequestSequence = []byte{0x15}

	// PrefixSupportTicketNumber is the prefix for global ticket number sequence
	// Key: PrefixSupportTicketNumber -> uint64
	PrefixSupportTicketNumber = []byte{0x16}

	// PrefixSupportResponseSequence is the prefix for response sequences
	// Key: PrefixSupportResponseSequence | request_id -> uint64
	PrefixSupportResponseSequence = []byte{0x17}

	// PrefixSupportEventSequence is the prefix for support event sequence
	// Key: PrefixSupportEventSequence -> uint64
	PrefixSupportEventSequence = []byte{0x18}

	// PrefixSupportEventCheckpoint is the prefix for event checkpoints
	// Key: PrefixSupportEventCheckpoint | subscriber_id -> SupportEventCheckpoint
	PrefixSupportEventCheckpoint = []byte{0x19}

	// PrefixSupportArchiveQueue is the prefix for archive queue items
	// Key: PrefixSupportArchiveQueue | archive_unix_be | "/" | request_id -> RetentionQueueEntry
	PrefixSupportArchiveQueue = []byte{0x1A}

	// PrefixSupportPurgeQueue is the prefix for purge queue items
	// Key: PrefixSupportPurgeQueue | purge_unix_be | "/" | request_id -> RetentionQueueEntry
	PrefixSupportPurgeQueue = []byte{0x1B}

	// PrefixSupportArchiveQueueByRequest is the prefix for archive queue request index
	// Key: PrefixSupportArchiveQueueByRequest | request_id -> archive_unix_be
	PrefixSupportArchiveQueueByRequest = []byte{0x1C}

	// PrefixSupportPurgeQueueByRequest is the prefix for purge queue request index
	// Key: PrefixSupportPurgeQueueByRequest | request_id -> purge_unix_be
	PrefixSupportPurgeQueueByRequest = []byte{0x1D}
)

const supportQueueTimeLength = 8

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

// SupportRequestKey returns the store key for a support request
func SupportRequestKey(requestID string) []byte {
	key := make([]byte, 0, len(PrefixSupportRequest)+len(requestID))
	key = append(key, PrefixSupportRequest...)
	key = append(key, []byte(requestID)...)
	return key
}

// SupportRequestBySubmitterKey returns the index key for a submitter's request
func SupportRequestBySubmitterKey(submitter []byte, requestID string) []byte {
	key := make([]byte, 0, len(PrefixSupportRequestBySubmitter)+len(submitter)+len(requestID)+1)
	key = append(key, PrefixSupportRequestBySubmitter...)
	key = append(key, submitter...)
	key = append(key, '/')
	key = append(key, []byte(requestID)...)
	return key
}

// SupportRequestBySubmitterPrefixKey returns the prefix for a submitter's requests
func SupportRequestBySubmitterPrefixKey(submitter []byte) []byte {
	key := make([]byte, 0, len(PrefixSupportRequestBySubmitter)+len(submitter)+1)
	key = append(key, PrefixSupportRequestBySubmitter...)
	key = append(key, submitter...)
	key = append(key, '/')
	return key
}

// SupportRequestByStatusKey returns the index key for status lookup
func SupportRequestByStatusKey(status SupportStatus, requestID string) []byte {
	key := make([]byte, 0, len(PrefixSupportRequestByStatus)+len(status.String())+len(requestID)+1)
	key = append(key, PrefixSupportRequestByStatus...)
	key = append(key, []byte(status.String())...)
	key = append(key, '/')
	key = append(key, []byte(requestID)...)
	return key
}

// SupportRequestByStatusPrefixKey returns the prefix for status lookup
func SupportRequestByStatusPrefixKey(status SupportStatus) []byte {
	key := make([]byte, 0, len(PrefixSupportRequestByStatus)+len(status.String())+1)
	key = append(key, PrefixSupportRequestByStatus...)
	key = append(key, []byte(status.String())...)
	key = append(key, '/')
	return key
}

// SupportRequestSequenceKey returns the sequence key for a submitter
func SupportRequestSequenceKey(submitter []byte) []byte {
	key := make([]byte, 0, len(PrefixSupportRequestSequence)+len(submitter))
	key = append(key, PrefixSupportRequestSequence...)
	key = append(key, submitter...)
	return key
}

// SupportTicketNumberKey returns the global ticket number sequence key
func SupportTicketNumberKey() []byte {
	return PrefixSupportTicketNumber
}

// SupportResponseKey returns the store key for a support response
func SupportResponseKey(requestID string, sequence uint64) []byte {
	seq := []byte(fmt.Sprintf("%d", sequence))
	key := make([]byte, 0, len(PrefixSupportResponse)+len(requestID)+len(seq)+1)
	key = append(key, PrefixSupportResponse...)
	key = append(key, []byte(requestID)...)
	key = append(key, '/')
	key = append(key, seq...)
	return key
}

// SupportResponseByRequestKey returns the index key for response lookup
func SupportResponseByRequestKey(requestID string, sequence uint64) []byte {
	seq := []byte(fmt.Sprintf("%d", sequence))
	key := make([]byte, 0, len(PrefixSupportResponseByRequest)+len(requestID)+len(seq)+1)
	key = append(key, PrefixSupportResponseByRequest...)
	key = append(key, []byte(requestID)...)
	key = append(key, '/')
	key = append(key, seq...)
	return key
}

// SupportResponseByRequestPrefixKey returns the prefix for responses by request
func SupportResponseByRequestPrefixKey(requestID string) []byte {
	key := make([]byte, 0, len(PrefixSupportResponseByRequest)+len(requestID)+1)
	key = append(key, PrefixSupportResponseByRequest...)
	key = append(key, []byte(requestID)...)
	key = append(key, '/')
	return key
}

// SupportResponseSequenceKey returns the sequence key for responses of a request
func SupportResponseSequenceKey(requestID string) []byte {
	key := make([]byte, 0, len(PrefixSupportResponseSequence)+len(requestID))
	key = append(key, PrefixSupportResponseSequence...)
	key = append(key, []byte(requestID)...)
	return key
}

// SupportEventSequenceKey returns the support event sequence key
func SupportEventSequenceKey() []byte {
	return PrefixSupportEventSequence
}

// SupportEventCheckpointKey returns the event checkpoint key
func SupportEventCheckpointKey(subscriberID string) []byte {
	key := make([]byte, 0, len(PrefixSupportEventCheckpoint)+len(subscriberID))
	key = append(key, PrefixSupportEventCheckpoint...)
	key = append(key, []byte(subscriberID)...)
	return key
}

// SupportArchiveQueueKey returns the archive queue key
func SupportArchiveQueueKey(archiveAt int64, requestID string) []byte {
	archive := EncodeSupportQueueTime(archiveAt)
	key := make([]byte, 0, len(PrefixSupportArchiveQueue)+len(archive)+len(requestID)+1)
	key = append(key, PrefixSupportArchiveQueue...)
	key = append(key, archive...)
	key = append(key, '/')
	key = append(key, []byte(requestID)...)
	return key
}

// SupportArchiveQueuePrefixKey returns the prefix for archive queue items
func SupportArchiveQueuePrefixKey(archiveAt int64) []byte {
	archive := EncodeSupportQueueTime(archiveAt)
	key := make([]byte, 0, len(PrefixSupportArchiveQueue)+len(archive)+1)
	key = append(key, PrefixSupportArchiveQueue...)
	key = append(key, archive...)
	key = append(key, '/')
	return key
}

// SupportPurgeQueueKey returns the purge queue key
func SupportPurgeQueueKey(purgeAt int64, requestID string) []byte {
	purge := EncodeSupportQueueTime(purgeAt)
	key := make([]byte, 0, len(PrefixSupportPurgeQueue)+len(purge)+len(requestID)+1)
	key = append(key, PrefixSupportPurgeQueue...)
	key = append(key, purge...)
	key = append(key, '/')
	key = append(key, []byte(requestID)...)
	return key
}

// SupportPurgeQueuePrefixKey returns the prefix for purge queue items
func SupportPurgeQueuePrefixKey(purgeAt int64) []byte {
	purge := EncodeSupportQueueTime(purgeAt)
	key := make([]byte, 0, len(PrefixSupportPurgeQueue)+len(purge)+1)
	key = append(key, PrefixSupportPurgeQueue...)
	key = append(key, purge...)
	key = append(key, '/')
	return key
}

// SupportArchiveQueueByRequestKey returns the archive queue request index key.
func SupportArchiveQueueByRequestKey(requestID string) []byte {
	key := make([]byte, 0, len(PrefixSupportArchiveQueueByRequest)+len(requestID))
	key = append(key, PrefixSupportArchiveQueueByRequest...)
	key = append(key, []byte(requestID)...)
	return key
}

// SupportPurgeQueueByRequestKey returns the purge queue request index key.
func SupportPurgeQueueByRequestKey(requestID string) []byte {
	key := make([]byte, 0, len(PrefixSupportPurgeQueueByRequest)+len(requestID))
	key = append(key, PrefixSupportPurgeQueueByRequest...)
	key = append(key, []byte(requestID)...)
	return key
}

// EncodeSupportQueueTime encodes a unix timestamp into sortable bytes.
func EncodeSupportQueueTime(unix int64) []byte {
	bz := make([]byte, supportQueueTimeLength)
	binary.BigEndian.PutUint64(bz, uint64(unix)^(uint64(1)<<63))
	return bz
}

// DecodeSupportQueueTime decodes sortable queue time bytes.
func DecodeSupportQueueTime(bz []byte) (int64, error) {
	if len(bz) != supportQueueTimeLength {
		return 0, fmt.Errorf("invalid queue time length: %d", len(bz))
	}
	return int64(binary.BigEndian.Uint64(bz) ^ (uint64(1) << 63)), nil
}

func parseSupportQueueKey(key []byte, prefix []byte) (int64, string, error) {
	expectedMin := len(prefix) + supportQueueTimeLength + 2
	if len(key) < expectedMin {
		return 0, "", fmt.Errorf("invalid queue key length: %d", len(key))
	}
	if string(key[:len(prefix)]) != string(prefix) {
		return 0, "", fmt.Errorf("invalid queue key prefix")
	}
	timeOffset := len(prefix)
	requestOffset := timeOffset + supportQueueTimeLength
	if key[requestOffset] != '/' {
		return 0, "", fmt.Errorf("invalid queue key separator")
	}
	unix, err := DecodeSupportQueueTime(key[timeOffset:requestOffset])
	if err != nil {
		return 0, "", err
	}
	requestID := string(key[requestOffset+1:])
	if requestID == "" {
		return 0, "", fmt.Errorf("queue key missing request id")
	}
	return unix, requestID, nil
}

// ParseSupportArchiveQueueKey parses an archive queue key.
func ParseSupportArchiveQueueKey(key []byte) (int64, string, error) {
	return parseSupportQueueKey(key, PrefixSupportArchiveQueue)
}

// ParseSupportPurgeQueueKey parses a purge queue key.
func ParseSupportPurgeQueueKey(key []byte) (int64, string, error) {
	return parseSupportQueueKey(key, PrefixSupportPurgeQueue)
}
