package types

import (
	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = types.ModuleName

	// StoreKey is the store key string for bme
	StoreKey = types.StoreKey

	// RouterKey is the message route for bme
	RouterKey = types.RouterKey
)

// Keys for bme store
// Items are stored with the following key: values
// - 0x01: Params
// - 0x02: State

var (
	ParamsKey = []byte{0x01}
	StateKey  = []byte{0x02}
)
