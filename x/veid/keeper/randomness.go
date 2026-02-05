package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RandomSource exposes deterministic random byte generation.
// Implementations MUST be deterministic for the same (ctx, purpose, extra) inputs
// to preserve consensus safety.
type RandomSource interface {
	Bytes(ctx sdk.Context, purpose string, size int, extra ...[]byte) ([]byte, error)
}

// DeterministicRandomSource derives pseudo-random bytes from transaction context.
// The derivation is deterministic across validators for the same block/tx and
// purpose inputs, making it safe for consensus logic.
type DeterministicRandomSource struct{}

// Bytes returns a deterministic byte slice of the requested size using ctx fields,
// purpose label, and optional extra domain separators.
func (DeterministicRandomSource) Bytes(ctx sdk.Context, purpose string, size int, extra ...[]byte) ([]byte, error) {
	if size <= 0 {
		return nil, errors.New("random byte size must be positive")
	}

	seedHasher := sha256.New()
	seedHasher.Write([]byte(purpose))
	seedHasher.Write([]byte(ctx.ChainID()))

	// Block height (consensus deterministic)
	height := ctx.BlockHeight()
	if height < 0 {
		height = 0
	}
	var heightBuf [8]byte
	binary.BigEndian.PutUint64(heightBuf[:], uint64(height))
	seedHasher.Write(heightBuf[:])

	// Block time (consensus deterministic)
	blockTime := ctx.BlockTime().UTC().UnixNano()
	if blockTime < 0 {
		blockTime = 0
	}
	var timeBuf [8]byte
	binary.BigEndian.PutUint64(timeBuf[:], uint64(blockTime))
	seedHasher.Write(timeBuf[:])

	// Transaction bytes (deterministic within the block)
	if tx := ctx.TxBytes(); len(tx) > 0 {
		seedHasher.Write(tx)
	}

	for _, b := range extra {
		seedHasher.Write(b)
	}

	seed := seedHasher.Sum(nil)
	return expandSeed(seed, size), nil
}

// expandSeed deterministically expands a seed to the requested size using
// counter-mode SHA256.
func expandSeed(seed []byte, size int) []byte {
	out := make([]byte, 0, size)
	var counter byte

	for len(out) < size {
		blockHasher := sha256.New()
		blockHasher.Write(seed)
		blockHasher.Write([]byte{counter})
		out = append(out, blockHasher.Sum(nil)...)
		counter++
	}

	return out[:size]
}

// RandomnessInputs carries caller-supplied randomness for nonce/salt injection.
// When provided, these values are used verbatim; otherwise the keeper's RandomSource
// supplies deterministic bytes derived from context.
type RandomnessInputs struct {
	Nonce          []byte
	CommitmentSalt []byte
	ScoreSalt      []byte
}
