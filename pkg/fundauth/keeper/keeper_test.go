package keeper

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/fundauth"
)

type testStore struct {
	key      *storetypes.KVStoreKey
	ctx      sdk.Context
	database dbm.DB
	stores   storetypes.CommitMultiStore
}

func newTestStore(t *testing.T) testStore {
	t.Helper()
	key := storetypes.NewKVStoreKey("fundauth-consumer-test")
	database := dbm.NewMemDB()
	return loadTestStore(t, database, key)
}

func loadTestStore(t *testing.T, database dbm.DB, key *storetypes.KVStoreKey) testStore {
	t.Helper()
	stores := rootmulti.NewStore(database, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stores.MountStoreWithDB(key, storetypes.StoreTypeIAVL, database)
	require.NoError(t, stores.LoadLatestVersion())
	return testStore{
		key:      key,
		ctx:      sdk.NewContext(stores, tmproto.Header{Height: 1}, false, log.NewNopLogger()),
		database: database,
		stores:   stores,
	}
}

func testDigest(value byte) fundauth.Digest {
	var digest fundauth.Digest
	digest[0] = value
	return digest
}

func newTestKeeper(t *testing.T, key storetypes.StoreKey) *Keeper {
	t.Helper()
	consumer, err := NewKeeper(key)
	require.NoError(t, err)
	return consumer
}

func TestWithAuthorizationSuccessReplayAndScoping(t *testing.T) {
	state := newTestStore(t)
	consumer := newTestKeeper(t, state.key)
	nonce := testDigest(1)
	auth := testDigest(2)
	var calls atomic.Int64
	callback := func(context.Context) error { calls.Add(1); return nil }

	require.NoError(t, consumer.WithAuthorization(sdk.WrapSDKContext(state.ctx), "account-a", nonce, auth, callback))
	require.ErrorIs(t, consumer.WithAuthorization(sdk.WrapSDKContext(state.ctx), "account-a", nonce, auth, callback), ErrAuthorizationReplay)
	require.ErrorIs(t, consumer.WithAuthorization(sdk.WrapSDKContext(state.ctx), "account-a", nonce, testDigest(3), callback), ErrAuthorizationReplay)
	require.NoError(t, consumer.WithAuthorization(sdk.WrapSDKContext(state.ctx), "account-b", nonce, auth, callback))
	require.EqualValues(t, 2, calls.Load())

	stored, found, err := consumer.AuthorizationDigest(sdk.WrapSDKContext(state.ctx), "account-a", nonce)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, auth, stored)
}

func TestWithAuthorizationPersistsAcrossKeeperReconstruction(t *testing.T) {
	state := newTestStore(t)
	nonce := testDigest(4)
	auth := testDigest(5)
	require.NoError(t, newTestKeeper(t, state.key).WithAuthorization(state.ctx, "account", nonce, auth, func(context.Context) error { return nil }))
	state.stores.Commit()

	restartedState := loadTestStore(t, state.database, storetypes.NewKVStoreKey(state.key.Name()))
	restarted := newTestKeeper(t, restartedState.key)
	require.ErrorIs(t, restarted.WithAuthorization(restartedState.ctx, "account", nonce, auth, func(context.Context) error {
		t.Fatal("replayed callback executed")
		return nil
	}), ErrAuthorizationReplay)
}

type outerContextKey struct{}

func TestWithAuthorizationPreservesOuterContextAndCachesSDKWrites(t *testing.T) {
	state := newTestStore(t)
	callbackKey := []byte("context-write")
	deadline := time.Now().Add(time.Minute).Round(0)
	outer := context.WithValue(sdk.WrapSDKContext(state.ctx), outerContextKey{}, "outer-value")
	outer, cancel := context.WithDeadline(outer, deadline)
	defer cancel()

	require.NoError(t, newTestKeeper(t, state.key).WithAuthorization(outer, "account", testDigest(16), testDigest(17), func(ctx context.Context) error {
		require.Equal(t, "outer-value", ctx.Value(outerContextKey{}))
		callbackDeadline, ok := ctx.Deadline()
		require.True(t, ok)
		require.Equal(t, deadline, callbackDeadline)
		sdk.UnwrapSDKContext(ctx).KVStore(state.key).Set(callbackKey, []byte("committed"))
		return nil
	}))
	require.Equal(t, []byte("committed"), state.ctx.KVStore(state.key).Get(callbackKey))
}

func TestWithAuthorizationCallbackRollback(t *testing.T) {
	state := newTestStore(t)
	consumer := newTestKeeper(t, state.key)
	nonce := testDigest(6)
	auth := testDigest(7)
	callbackKey := []byte("protected-write")
	callbackErr := errors.New("callback failed")

	err := consumer.WithAuthorization(sdk.WrapSDKContext(state.ctx), "account", nonce, auth, func(ctx context.Context) error {
		sdk.UnwrapSDKContext(ctx).KVStore(state.key).Set(callbackKey, []byte("not committed"))
		return callbackErr
	})
	require.ErrorIs(t, err, callbackErr)
	require.Nil(t, state.ctx.KVStore(state.key).Get(callbackKey))
	_, found, err := consumer.AuthorizationDigest(state.ctx, "account", nonce)
	require.NoError(t, err)
	require.False(t, found)

	require.NoError(t, consumer.WithAuthorization(state.ctx, "account", nonce, auth, func(ctx context.Context) error {
		sdk.UnwrapSDKContext(ctx).KVStore(state.key).Set(callbackKey, []byte("committed"))
		return nil
	}))
	require.Equal(t, []byte("committed"), state.ctx.KVStore(state.key).Get(callbackKey))
}

func TestWithAuthorizationCancellationRollback(t *testing.T) {
	state := newTestStore(t)
	consumer := newTestKeeper(t, state.key)
	nonce := testDigest(8)
	auth := testDigest(9)
	callbackKey := []byte("canceled-write")
	ctx, cancel := context.WithCancel(sdk.WrapSDKContext(state.ctx))

	err := consumer.WithAuthorization(ctx, "account", nonce, auth, func(callbackCtx context.Context) error {
		sdk.UnwrapSDKContext(callbackCtx).KVStore(state.key).Set(callbackKey, []byte("not committed"))
		cancel()
		require.ErrorIs(t, callbackCtx.Err(), context.Canceled)
		select {
		case <-callbackCtx.Done():
		default:
			t.Fatal("callback context cancellation was not observable")
		}
		return nil
	})
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, state.ctx.KVStore(state.key).Get(callbackKey))
	require.NoError(t, consumer.WithAuthorization(state.ctx, "account", nonce, auth, func(context.Context) error { return nil }))
}

func TestWithAuthorizationConcurrentAcrossKeepersExactlyOnce(t *testing.T) {
	state := newTestStore(t)
	reconstructedKey := storetypes.NewKVStoreKey(state.key.Name())
	reconstructed := newTestKeeper(t, reconstructedKey)
	consumers := []*Keeper{newTestKeeper(t, state.key), newTestKeeper(t, state.key)}
	require.Same(t, consumers[0].mu, consumers[1].mu)
	require.Same(t, consumers[0].mu, reconstructed.mu)
	var calls atomic.Int64
	var successes atomic.Int64
	start := make(chan struct{})
	results := make(chan error, 32)
	var wait sync.WaitGroup
	for worker := range 32 {
		wait.Add(1)
		go func(consumer *Keeper) {
			defer wait.Done()
			<-start
			err := consumer.WithAuthorization(state.ctx, "account", testDigest(10), testDigest(11), func(context.Context) error {
				calls.Add(1)
				return nil
			})
			if err == nil {
				successes.Add(1)
				return
			}
			results <- err
		}(consumers[worker%len(consumers)])
	}
	close(start)
	wait.Wait()
	close(results)
	for err := range results {
		require.ErrorIs(t, err, ErrAuthorizationReplay)
	}
	require.EqualValues(t, 1, calls.Load())
	require.EqualValues(t, 1, successes.Load())
}

func TestAuthorizationDigestRejectsCorruptState(t *testing.T) {
	state := newTestStore(t)
	consumer := newTestKeeper(t, state.key)
	nonce := testDigest(18)
	state.ctx.KVStore(state.key).Set(authorizationKey("account", nonce), []byte("corrupt"))

	_, _, err := consumer.AuthorizationDigest(state.ctx, "account", nonce)
	require.ErrorIs(t, err, ErrCorruptState)
}

func TestKeeperRejectsMalformedInputsAndContexts(t *testing.T) {
	state := newTestStore(t)
	consumer := newTestKeeper(t, state.key)
	validNonce := testDigest(12)
	validAuth := testDigest(13)
	callback := func(context.Context) error { return nil }

	_, err := NewKeeper(nil)
	require.ErrorIs(t, err, ErrInvalidInput)
	var typedNilKey *storetypes.KVStoreKey
	_, err = NewKeeper(typedNilKey)
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = NewKeeper(storetypes.NewTransientStoreKey("transient"))
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = NewKeeper(storetypes.NewMemoryStoreKey("memory"))
	require.ErrorIs(t, err, ErrInvalidInput)
	require.False(t, (*Keeper)(nil).KeeperRequired())
	require.ErrorIs(t, consumer.WithAuthorization(state.ctx, "", validNonce, validAuth, callback), ErrInvalidInput)
	require.ErrorIs(t, consumer.WithAuthorization(state.ctx, " \t", validNonce, validAuth, callback), ErrInvalidInput)
	require.ErrorIs(t, consumer.WithAuthorization(state.ctx, string(make([]byte, maxAccountIDLength+1)), validNonce, validAuth, callback), ErrInvalidInput)
	require.ErrorIs(t, consumer.WithAuthorization(state.ctx, "account", fundauth.Digest{}, validAuth, callback), ErrInvalidInput)
	require.ErrorIs(t, consumer.WithAuthorization(state.ctx, "account", validNonce, fundauth.Digest{}, callback), ErrInvalidInput)
	require.ErrorIs(t, consumer.WithAuthorization(state.ctx, "account", validNonce, validAuth, nil), ErrInvalidInput)
	require.ErrorIs(t, consumer.WithAuthorization(nil, "account", validNonce, validAuth, callback), ErrInvalidSDKContext)
	require.ErrorIs(t, consumer.WithAuthorization(context.Background(), "account", validNonce, validAuth, callback), ErrInvalidSDKContext)

	canceled, cancel := context.WithCancel(sdk.WrapSDKContext(state.ctx))
	cancel()
	require.ErrorIs(t, consumer.WithAuthorization(canceled, "account", validNonce, validAuth, callback), context.Canceled)
	_, _, err = consumer.AuthorizationDigest(context.Background(), "account", validNonce)
	require.ErrorIs(t, err, ErrInvalidSDKContext)
}

func TestAuthorizationKeyIsLengthDelimited(t *testing.T) {
	nonce := testDigest(14)
	require.NotEqual(t, authorizationKey("ab", nonce), authorizationKey("a", nonce))
	require.NotEqual(t, authorizationKey("ab", nonce), authorizationKey("ab", testDigest(15)))
}
