package types

import (
	"encoding/binary"
)

const (
	// ModuleName defines the module name.
	ModuleName = "resources"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey defines the module message route.
	RouterKey = ModuleName

	// QuerierRoute defines the querier route.
	QuerierRoute = ModuleName
)

var (
	InventoryKeyPrefix          = []byte{0x01}
	AllocationKeyPrefix         = []byte{0x02}
	AllocationEventKeyPrefix    = []byte{0x03}
	AllocationProviderKeyPrefix = []byte{0x04}
	PendingAllocationKeyPrefix  = []byte{0x05}
	InventorySequenceKeyPrefix  = []byte{0x06}
	AllocationSequenceKeyPrefix = []byte{0x07}
	AllocationEventSeqKeyPrefix = []byte{0x08}
	SlashingEventKeyPrefix      = []byte{0x09}
)

// InventoryKey returns the key for a provider inventory entry.
func InventoryKey(provider string, class ResourceClass, inventoryID string) []byte {
	key := append([]byte(provider), 0x00)
	key = append(key, byte(class))
	key = append(key, 0x00)
	key = append(key, []byte(inventoryID)...)
	return append(InventoryKeyPrefix, key...)
}

// InventoryProviderPrefix returns the prefix for inventories by provider.
func InventoryProviderPrefix(provider string) []byte {
	key := append([]byte(provider), 0x00)
	return append(InventoryKeyPrefix, key...)
}

// AllocationKey returns the key for an allocation.
func AllocationKey(allocationID string) []byte {
	return append(AllocationKeyPrefix, []byte(allocationID)...)
}

// AllocationEventKey returns the key for an allocation event.
func AllocationEventKey(allocationID string, sequence uint64) []byte {
	seq := make([]byte, 8)
	binary.BigEndian.PutUint64(seq, sequence)
	key := append([]byte(allocationID), 0x00)
	key = append(key, seq...)
	return append(AllocationEventKeyPrefix, key...)
}

// AllocationEventPrefix returns the prefix for allocation events.
func AllocationEventPrefix(allocationID string) []byte {
	key := append([]byte(allocationID), 0x00)
	return append(AllocationEventKeyPrefix, key...)
}

// SlashingEventKey returns the key for a slashing event.
func SlashingEventKey(allocationID string, sequence uint64) []byte {
	seq := make([]byte, 8)
	binary.BigEndian.PutUint64(seq, sequence)
	key := append([]byte(allocationID), 0x00)
	key = append(key, seq...)
	return append(SlashingEventKeyPrefix, key...)
}

// AllocationProviderKey indexes allocations by provider.
func AllocationProviderKey(provider string, allocationID string) []byte {
	key := append([]byte(provider), 0x00)
	key = append(key, []byte(allocationID)...)
	return append(AllocationProviderKeyPrefix, key...)
}

// AllocationProviderPrefix returns prefix for provider allocations.
func AllocationProviderPrefix(provider string) []byte {
	key := append([]byte(provider), 0x00)
	return append(AllocationProviderKeyPrefix, key...)
}

// PendingAllocationKey indexes pending allocations by expiry.
func PendingAllocationKey(expiryUnix uint64, allocationID string) []byte {
	seq := make([]byte, 8)
	binary.BigEndian.PutUint64(seq, expiryUnix)
	key := append(seq, 0x00)
	key = append(key, []byte(allocationID)...)
	return append(PendingAllocationKeyPrefix, key...)
}

// PendingAllocationPrefixByTime returns prefix up to time.
func PendingAllocationPrefixByTime(expiryUnix uint64) []byte {
	seq := make([]byte, 8)
	binary.BigEndian.PutUint64(seq, expiryUnix)
	return append(PendingAllocationKeyPrefix, seq...)
}

// SequenceKey returns a key for a sequence.
func SequenceKey(prefix []byte, name string) []byte {
	return append(prefix, []byte(name)...)
}
