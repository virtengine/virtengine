package keeper

import (
	"encoding/binary"
	"sync/atomic"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/sdkutil"
)

var benchmarkAddrCounter uint64

const benchmarkAddrLen = 20

func bech32AddrBenchmark(t *testing.T) string {
	t.Helper()
	counter := atomic.AddUint64(&benchmarkAddrCounter, 1)
	addr := make([]byte, benchmarkAddrLen)
	binary.BigEndian.PutUint64(addr[len(addr)-8:], counter)
	return sdk.MustBech32ifyAddressBytes(sdkutil.Bech32PrefixAccAddr, addr)
}
