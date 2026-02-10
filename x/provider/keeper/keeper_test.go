package keeper_test

import (
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	"github.com/virtengine/virtengine/sdk/go/testutil"

	"github.com/virtengine/virtengine/testutil/state"
	"github.com/virtengine/virtengine/x/provider/keeper"
)

func TestProviderCreate(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	foundProv, found := keeper.Get(ctx, owner)
	require.True(t, found)
	require.Equal(t, prov, foundProv)
}

func TestProviderDuplicate(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	err = keeper.Create(ctx, prov)
	require.EqualError(t, err, types.ErrProviderExists.Error())
}

func TestProviderGetNonExisting(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	foundProv, found := keeper.Get(ctx, owner)
	require.False(t, found)
	require.Equal(t, types.Provider{}, foundProv)
}

func TestProviderDeleteExisting(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// Verify provider exists before deletion
	foundProv, found := keeper.Get(ctx, owner)
	require.True(t, found)
	require.Equal(t, prov, foundProv)

	// Delete the provider
	keeper.Delete(ctx, owner)

	// Verify provider no longer exists after deletion
	_, found = keeper.Get(ctx, owner)
	require.False(t, found, "provider should not exist after deletion")
}

func TestProviderDeleteNonExisting(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// Deleting non-existent provider should be a no-op (no panic)
	require.NotPanics(t, func() {
		keeper.Delete(ctx, owner)
	})
}

func TestProviderUpdateNonExisting(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Update(ctx, prov)
	require.EqualError(t, err, types.ErrProviderNotFound.Error())
}

func TestProviderUpdateExisting(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	prov.HostURI = "virtengine.domain.com"
	err = keeper.Update(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	foundProv, found := keeper.Get(ctx, owner)
	require.True(t, found)
	require.Equal(t, prov, foundProv)
}

func TestWithProviders(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)
	prov2 := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	err = keeper.Create(ctx, prov2)
	require.NoError(t, err)

	count := 0

	keeper.WithProviders(ctx, func(provider types.Provider) bool {
		if !reflect.DeepEqual(provider, prov) && !reflect.DeepEqual(provider, prov2) {
			require.Fail(t, "unknown provider")
		}
		count++
		return false
	})

	require.Equal(t, 2, count)
}

func TestWithProvidersBreak(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)
	prov2 := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	err = keeper.Create(ctx, prov2)
	require.NoError(t, err)

	count := 0

	keeper.WithProviders(ctx, func(provider types.Provider) bool {
		if !reflect.DeepEqual(provider, prov) && !reflect.DeepEqual(provider, prov2) {
			require.Fail(t, "unknown provider")
		}
		count++
		return true
	})

	require.Equal(t, 1, count)
}

func TestKeeperCoder(t *testing.T) {
	_, keeper := setupKeeper(t)
	codec := keeper.Codec()
	require.NotNil(t, codec)
}

// ============= Public Key Storage Tests =============

func TestSetProviderPublicKey(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	// Create provider first
	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// Test setting Ed25519 public key
	testPubKey := make([]byte, 32) // Ed25519 public key size
	for i := range testPubKey {
		testPubKey[i] = byte(i)
	}

	err = keeper.SetProviderPublicKey(ctx, owner, testPubKey, types.PublicKeyTypeEd25519)
	require.NoError(t, err)

	// Verify public key was stored
	storedKey, found := keeper.GetProviderPublicKey(ctx, owner)
	require.True(t, found)
	require.Equal(t, testPubKey, storedKey)

	// Verify full record
	record, found := keeper.GetProviderPublicKeyRecord(ctx, owner)
	require.True(t, found)
	require.Equal(t, testPubKey, record.PublicKey)
	require.Equal(t, types.PublicKeyTypeEd25519, record.KeyType)
	require.Equal(t, uint32(0), record.RotationCount)
}

func TestSetProviderPublicKeyX25519(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// Test X25519 key (32 bytes)
	testPubKey := make([]byte, 32)
	for i := range testPubKey {
		testPubKey[i] = byte(i + 10)
	}

	err = keeper.SetProviderPublicKey(ctx, owner, testPubKey, types.PublicKeyTypeX25519)
	require.NoError(t, err)

	record, found := keeper.GetProviderPublicKeyRecord(ctx, owner)
	require.True(t, found)
	require.Equal(t, types.PublicKeyTypeX25519, record.KeyType)
}

func TestSetProviderPublicKeySecp256k1(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// Test secp256k1 compressed key (33 bytes)
	testPubKey := make([]byte, 33)
	testPubKey[0] = 0x02 // Compressed format prefix
	for i := 1; i < len(testPubKey); i++ {
		testPubKey[i] = byte(i)
	}

	err = keeper.SetProviderPublicKey(ctx, owner, testPubKey, types.PublicKeyTypeSecp256k1)
	require.NoError(t, err)

	record, found := keeper.GetProviderPublicKeyRecord(ctx, owner)
	require.True(t, found)
	require.Equal(t, types.PublicKeyTypeSecp256k1, record.KeyType)
}

func TestSetProviderPublicKeyNonExistentProvider(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	testPubKey := make([]byte, 32)
	err = keeper.SetProviderPublicKey(ctx, owner, testPubKey, types.PublicKeyTypeEd25519)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestSetProviderPublicKeyInvalidType(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	testPubKey := make([]byte, 32)
	err = keeper.SetProviderPublicKey(ctx, owner, testPubKey, "invalid_type")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid public key type")
}

func TestSetProviderPublicKeyInvalidLength(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// Ed25519 requires 32 bytes, try 16
	testPubKey := make([]byte, 16)
	err = keeper.SetProviderPublicKey(ctx, owner, testPubKey, types.PublicKeyTypeEd25519)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid public key")
}

func TestGetProviderPublicKeyNonExistent(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// Provider doesn't exist, should return false
	key, found := keeper.GetProviderPublicKey(ctx, owner)
	require.False(t, found)
	require.Nil(t, key)
}

func TestGetProviderPublicKeyProviderExistsNoKey(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// Provider exists but no key set
	key, found := keeper.GetProviderPublicKey(ctx, owner)
	require.False(t, found)
	require.Nil(t, key)
}

func TestDeleteProviderPublicKey(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// Set a key
	testPubKey := make([]byte, 32)
	err = keeper.SetProviderPublicKey(ctx, owner, testPubKey, types.PublicKeyTypeEd25519)
	require.NoError(t, err)

	// Verify key exists
	_, found := keeper.GetProviderPublicKey(ctx, owner)
	require.True(t, found)

	// Delete the key
	keeper.DeleteProviderPublicKey(ctx, owner)

	// Verify key is gone
	_, found = keeper.GetProviderPublicKey(ctx, owner)
	require.False(t, found)
}

func TestDeleteProviderPublicKeyNonExistent(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// Should not panic on non-existent
	require.NotPanics(t, func() {
		keeper.DeleteProviderPublicKey(ctx, owner)
	})
}

func TestPublicKeyRotationCount(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// First key set
	key1 := make([]byte, 32)
	for i := range key1 {
		key1[i] = byte(i)
	}
	err = keeper.SetProviderPublicKey(ctx, owner, key1, types.PublicKeyTypeEd25519)
	require.NoError(t, err)

	record, found := keeper.GetProviderPublicKeyRecord(ctx, owner)
	require.True(t, found)
	require.Equal(t, uint32(0), record.RotationCount)

	// Second key set (update)
	key2 := make([]byte, 32)
	for i := range key2 {
		key2[i] = byte(i + 100)
	}
	err = keeper.SetProviderPublicKey(ctx, owner, key2, types.PublicKeyTypeEd25519)
	require.NoError(t, err)

	record, found = keeper.GetProviderPublicKeyRecord(ctx, owner)
	require.True(t, found)
	require.Equal(t, uint32(1), record.RotationCount)
	require.Equal(t, key2, record.PublicKey)

	// Third key set
	key3 := make([]byte, 32)
	for i := range key3 {
		key3[i] = byte(i + 200)
	}
	err = keeper.SetProviderPublicKey(ctx, owner, key3, types.PublicKeyTypeEd25519)
	require.NoError(t, err)

	record, found = keeper.GetProviderPublicKeyRecord(ctx, owner)
	require.True(t, found)
	require.Equal(t, uint32(2), record.RotationCount)
}

func TestRotateProviderPublicKeyFirstTime(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	err := keeper.Create(ctx, prov)
	require.NoError(t, err)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	// First rotation (no existing key) should work without signature
	newKey := make([]byte, 32)
	for i := range newKey {
		newKey[i] = byte(i)
	}
	err = keeper.RotateProviderPublicKey(ctx, owner, newKey, types.PublicKeyTypeEd25519, nil)
	require.NoError(t, err)

	storedKey, found := keeper.GetProviderPublicKey(ctx, owner)
	require.True(t, found)
	require.Equal(t, newKey, storedKey)
}

func TestRotateProviderPublicKeyNonExistentProvider(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov := testutil.Provider(t)

	owner, err := sdk.AccAddressFromBech32(prov.Owner)
	require.NoError(t, err)

	newKey := make([]byte, 32)
	err = keeper.RotateProviderPublicKey(ctx, owner, newKey, types.PublicKeyTypeEd25519, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestWithProviderPublicKeys(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov1 := testutil.Provider(t)
	prov2 := testutil.Provider(t)

	err := keeper.Create(ctx, prov1)
	require.NoError(t, err)
	err = keeper.Create(ctx, prov2)
	require.NoError(t, err)

	owner1, _ := sdk.AccAddressFromBech32(prov1.Owner)
	owner2, _ := sdk.AccAddressFromBech32(prov2.Owner)

	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	for i := range key1 {
		key1[i] = byte(i)
		key2[i] = byte(i + 50)
	}

	err = keeper.SetProviderPublicKey(ctx, owner1, key1, types.PublicKeyTypeEd25519)
	require.NoError(t, err)
	err = keeper.SetProviderPublicKey(ctx, owner2, key2, types.PublicKeyTypeEd25519)
	require.NoError(t, err)

	count := 0
	keeper.WithProviderPublicKeys(ctx, func(addr sdk.AccAddress, record types.ProviderPublicKeyRecord) bool {
		count++
		require.NotEmpty(t, record.PublicKey)
		require.Equal(t, types.PublicKeyTypeEd25519, record.KeyType)
		return false
	})
	require.Equal(t, 2, count)
}

func TestWithProviderPublicKeysBreak(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	prov1 := testutil.Provider(t)
	prov2 := testutil.Provider(t)

	err := keeper.Create(ctx, prov1)
	require.NoError(t, err)
	err = keeper.Create(ctx, prov2)
	require.NoError(t, err)

	owner1, _ := sdk.AccAddressFromBech32(prov1.Owner)
	owner2, _ := sdk.AccAddressFromBech32(prov2.Owner)

	key1 := make([]byte, 32)
	key2 := make([]byte, 32)

	err = keeper.SetProviderPublicKey(ctx, owner1, key1, types.PublicKeyTypeEd25519)
	require.NoError(t, err)
	err = keeper.SetProviderPublicKey(ctx, owner2, key2, types.PublicKeyTypeEd25519)
	require.NoError(t, err)

	count := 0
	keeper.WithProviderPublicKeys(ctx, func(addr sdk.AccAddress, record types.ProviderPublicKeyRecord) bool {
		count++
		return true // Stop after first
	})
	require.Equal(t, 1, count)
}

func setupKeeper(t testing.TB) (sdk.Context, keeper.IKeeper) {
	t.Helper()

	suite := state.SetupTestSuite(t)

	return suite.Context(), suite.ProviderKeeper()
}
