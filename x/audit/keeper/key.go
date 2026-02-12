package keeper

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	types "github.com/virtengine/virtengine/sdk/go/node/audit/v1"

	"github.com/virtengine/virtengine/util/validation"
)

func ProviderKey(id types.ProviderID) []byte {
	buf := bytes.NewBuffer(types.PrefixProviderID())
	if _, err := buf.Write(address.MustLengthPrefix(id.Owner.Bytes())); err != nil {
		panic(err)
	}

	if _, err := buf.Write(address.MustLengthPrefix(id.Auditor.Bytes())); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func ProviderPrefix(id sdk.Address) []byte {
	buf := bytes.NewBuffer(types.PrefixProviderID())
	if _, err := buf.Write(address.MustLengthPrefix(id.Bytes())); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func ParseIDFromKey(key []byte) types.ProviderID {
	// skip prefix if set

	validation.AssertKeyAtLeastLength(key, len(types.PrefixProviderID())+1)
	if !bytes.HasPrefix(key, types.PrefixProviderID()) {
		panic(fmt.Sprintf("invalid key prefix. expected 0x%s, actual 0x%s", hex.EncodeToString(key[:1]), types.PrefixProviderID()))
	}

	// remove a prefix key
	key = key[len(types.PrefixProviderID()):]

	dataLen := int(key[0])
	key = key[1:]
	validation.AssertKeyAtLeastLength(key, dataLen)

	owner := make([]byte, dataLen)
	copy(owner, key[:dataLen])
	key = key[dataLen:]
	validation.AssertKeyAtLeastLength(key, 1)

	dataLen = int(key[0])
	key = key[1:]
	validation.AssertKeyLength(key, dataLen)
	auditor := make([]byte, dataLen)
	copy(auditor, key[:dataLen])

	return types.ProviderID{
		Owner:   sdk.AccAddress(owner),
		Auditor: sdk.AccAddress(auditor),
	}
}

// Audit log key functions

var (
	auditLogPrefixKey    = []byte{0x10}
	actorIndexPrefixKey  = []byte{0x11}
	moduleIndexPrefixKey = []byte{0x12}
	exportJobPrefixKey   = []byte{0x13}
	paramsKey            = []byte{0x14}
)

// auditLogPrefix returns the prefix for all audit logs
func auditLogPrefix() []byte {
	return auditLogPrefixKey
}

// auditLogKey returns the key for a specific audit log entry
func auditLogKey(id string) []byte {
	key := make([]byte, 1+len(id))
	key[0] = auditLogPrefixKey[0]
	copy(key[1:], []byte(id))
	return key
}

// auditLogActorPrefix returns the prefix for actor index
func auditLogActorPrefix(actor string) []byte {
	key := make([]byte, 1+len(actor))
	key[0] = actorIndexPrefixKey[0]
	copy(key[1:], []byte(actor))
	return key
}

// auditLogActorIndexKey returns the full index key for actor + height + id
func auditLogActorIndexKey(actor string, height int64, id string) []byte {
	prefix := auditLogActorPrefix(actor)
	heightBytes := make([]byte, 8)
	// Safe conversion: height is always positive in SDK context
	binary.BigEndian.PutUint64(heightBytes, uint64(height)) //nolint:gosec

	key := make([]byte, len(prefix)+8+len(id))
	copy(key, prefix)
	copy(key[len(prefix):], heightBytes)
	copy(key[len(prefix)+8:], []byte(id))
	return key
}

// auditLogModulePrefix returns the prefix for module index
func auditLogModulePrefix(module string) []byte {
	key := make([]byte, 1+len(module))
	key[0] = moduleIndexPrefixKey[0]
	copy(key[1:], []byte(module))
	return key
}

// auditLogModuleIndexKey returns the full index key for module + height + id
func auditLogModuleIndexKey(module string, height int64, id string) []byte {
	prefix := auditLogModulePrefix(module)
	heightBytes := make([]byte, 8)
	// Safe conversion: height is always positive in SDK context
	binary.BigEndian.PutUint64(heightBytes, uint64(height)) //nolint:gosec

	key := make([]byte, len(prefix)+8+len(id))
	copy(key, prefix)
	copy(key[len(prefix):], heightBytes)
	copy(key[len(prefix)+8:], []byte(id))
	return key
}

// exportJobPrefix returns the prefix for all export jobs
func exportJobPrefix() []byte {
	return exportJobPrefixKey
}

// exportJobKey returns the key for a specific export job
func exportJobKey(jobID string) []byte {
	key := make([]byte, 1+len(jobID))
	key[0] = exportJobPrefixKey[0]
	copy(key[1:], []byte(jobID))
	return key
}

// auditLogParamsKey returns the key for audit log parameters
func auditLogParamsKey() []byte {
	return paramsKey
}
