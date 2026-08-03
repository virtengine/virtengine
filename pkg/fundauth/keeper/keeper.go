package keeper

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"sync"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/fundauth"
)

const (
	authorizationKeyDomain = "virtengine/fundauth/authorization/v1"
	maxAccountIDLength     = 4096
)

var (
	ErrInvalidInput        = errors.New("invalid fund authorization consumer input")
	ErrInvalidSDKContext   = errors.New("invalid Cosmos SDK context")
	ErrAuthorizationReplay = errors.New("fund authorization nonce already consumed")
	ErrCorruptState        = errors.New("corrupt fund authorization consumer state")
)

// Keeper durably consumes fund authorizations in a Cosmos SDK KV store.
type Keeper struct {
	storeKey *storetypes.KVStoreKey
	mu       *sync.Mutex
}

var keeperLocks sync.Map

var _ fundauth.AtomicAuthorizationConsumer = (*Keeper)(nil)

// NewKeeper constructs an authorization consumer backed by storeKey.
func NewKeeper(storeKey storetypes.StoreKey) (*Keeper, error) {
	kvStoreKey, ok := storeKey.(*storetypes.KVStoreKey)
	if !ok || kvStoreKey == nil {
		return nil, fmt.Errorf("%w: store key must be a nonnil durable KV store key", ErrInvalidInput)
	}
	lock, _ := keeperLocks.LoadOrStore(kvStoreKey.Name(), &sync.Mutex{})
	return &Keeper{storeKey: kvStoreKey, mu: lock.(*sync.Mutex)}, nil
}

func (keeper *Keeper) KeeperRequired() bool {
	return keeper != nil && keeper.storeKey != nil && keeper.mu != nil
}

// WithAuthorization atomically runs protected and records the consumed nonce.
func (keeper *Keeper) WithAuthorization(ctx context.Context, accountID string, nonceDigest, authDigest fundauth.Digest, protected func(context.Context) error) error {
	if !keeper.KeeperRequired() {
		return fmt.Errorf("%w: uninitialized keeper", ErrInvalidInput)
	}
	if err := validateAuthorizationInput(accountID, nonceDigest, authDigest, protected); err != nil {
		return err
	}
	if ctx == nil {
		return fmt.Errorf("%w: nil context", ErrInvalidSDKContext)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	sdkCtx, err := unwrapSDKContext(ctx)
	if err != nil {
		return err
	}
	key := authorizationKey(accountID, nonceDigest)

	keeper.mu.Lock()
	defer keeper.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}
	cacheCtx, write := sdkCtx.CacheContext()
	store := cacheCtx.KVStore(keeper.storeKey)
	if store.Has(key) {
		return ErrAuthorizationReplay
	}
	if err := protected(&protectedContext{Context: ctx, sdkContext: sdk.WrapSDKContext(cacheCtx)}); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	store.Set(key, authDigest[:])
	write()
	return nil
}

// AuthorizationDigest returns the digest recorded for an account and nonce.
func (keeper *Keeper) AuthorizationDigest(ctx context.Context, accountID string, nonceDigest fundauth.Digest) (fundauth.Digest, bool, error) {
	var digest fundauth.Digest
	if !keeper.KeeperRequired() {
		return digest, false, fmt.Errorf("%w: uninitialized keeper", ErrInvalidInput)
	}
	if strings.TrimSpace(accountID) == "" || len(accountID) > maxAccountIDLength || isZeroDigest(nonceDigest) {
		return digest, false, fmt.Errorf("%w: account ID and nonce digest are required", ErrInvalidInput)
	}
	if ctx == nil {
		return digest, false, fmt.Errorf("%w: nil context", ErrInvalidSDKContext)
	}
	if err := ctx.Err(); err != nil {
		return digest, false, err
	}
	sdkCtx, err := unwrapSDKContext(ctx)
	if err != nil {
		return digest, false, err
	}
	keeper.mu.Lock()
	defer keeper.mu.Unlock()
	value := sdkCtx.KVStore(keeper.storeKey).Get(authorizationKey(accountID, nonceDigest))
	if value == nil {
		return digest, false, nil
	}
	if len(value) != len(digest) {
		return digest, false, fmt.Errorf("%w: authorization digest length %d", ErrCorruptState, len(value))
	}
	copy(digest[:], value)
	if isZeroDigest(digest) {
		return fundauth.Digest{}, false, fmt.Errorf("%w: zero authorization digest", ErrCorruptState)
	}
	return digest, true, nil
}

func validateAuthorizationInput(accountID string, nonceDigest, authDigest fundauth.Digest, protected func(context.Context) error) error {
	if strings.TrimSpace(accountID) == "" || len(accountID) > maxAccountIDLength {
		return fmt.Errorf("%w: account ID is required", ErrInvalidInput)
	}
	if isZeroDigest(nonceDigest) {
		return fmt.Errorf("%w: nonce digest is required", ErrInvalidInput)
	}
	if isZeroDigest(authDigest) {
		return fmt.Errorf("%w: authorization digest is required", ErrInvalidInput)
	}
	if protected == nil {
		return fmt.Errorf("%w: callback is required", ErrInvalidInput)
	}
	return nil
}

func authorizationKey(accountID string, nonceDigest fundauth.Digest) []byte {
	domain := []byte(authorizationKeyDomain)
	key := make([]byte, 0, 4+len(domain)+4+len(accountID)+4+len(nonceDigest))
	key = appendLengthPrefixed(key, domain)
	key = appendLengthPrefixed(key, []byte(accountID))
	key = appendLengthPrefixed(key, nonceDigest[:])
	return key
}

func appendLengthPrefixed(dst, value []byte) []byte {
	dst = binary.BigEndian.AppendUint32(dst, uint32(len(value)))
	return append(dst, value...)
}

func isZeroDigest(digest fundauth.Digest) bool {
	return digest == fundauth.Digest{}
}

type protectedContext struct {
	context.Context
	sdkContext context.Context
}

func (ctx *protectedContext) Value(key any) any {
	if value := ctx.sdkContext.Value(key); value != nil {
		return value
	}
	return ctx.Context.Value(key)
}

func unwrapSDKContext(ctx context.Context) (sdkCtx sdk.Context, err error) {
	if direct, ok := ctx.(sdk.Context); ok {
		return direct, nil
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%w: %v", ErrInvalidSDKContext, recovered)
		}
	}()
	return sdk.UnwrapSDKContext(ctx), nil
}
