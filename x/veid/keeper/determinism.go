package keeper

import (
	"fmt"
	"hash"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

const randomnessByteLen = 32

// resolveRandomBytes returns caller-provided randomness when present, otherwise
// derives deterministic bytes from the keeper's RandomSource.
func (k Keeper) resolveRandomBytes(
	ctx sdk.Context,
	provided []byte,
	purpose string,
	extra ...[]byte,
) ([]byte, error) {
	if len(provided) > 0 {
		if len(provided) != randomnessByteLen {
			return nil, types.ErrProofGenerationFailed.Wrapf("%s must be %d bytes", purpose, randomnessByteLen)
		}
		return provided, nil
	}

	source := k.randSource
	if source == nil {
		source = DeterministicRandomSource{}
	}

	derived, err := source.Bytes(ctx, purpose, randomnessByteLen, extra...)
	if err != nil {
		return nil, types.ErrProofGenerationFailed.Wrapf("failed to derive %s", purpose)
	}

	return derived, nil
}

// appendClaimsDeterministically writes claim key/value pairs into the hash in
// a stable, sorted order to avoid map iteration nondeterminism.
func appendClaimsDeterministically(h hash.Hash, claims map[string]interface{}) {
	if len(claims) == 0 {
		return
	}

	keys := make([]string, 0, len(claims))
	for k := range claims {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		h.Write([]byte(key))
		fmt.Fprintf(h, "%v", claims[key])
	}
}
