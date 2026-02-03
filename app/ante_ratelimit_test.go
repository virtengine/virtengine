package app

import (
	"testing"

	"cosmossdk.io/log"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	apptypes "github.com/virtengine/virtengine/app/types"
)

// MockTx implements sdk.Tx for testing
type MockTx struct {
	msgs    []sdk.Msg
	signers []sdk.AccAddress
}

func (m MockTx) GetMsgs() []sdk.Msg {
	return m.msgs
}

func (m MockTx) GetMsgsV2() ([]proto.Message, error) {
	// Return empty for testing - not used by rate limiter
	return nil, nil
}

// MockSigVerifiableTx combines MockTx with signer capability
// Implements signing.SigVerifiableTx interface
type MockSigVerifiableTx struct {
	MockTx
}

func (m MockSigVerifiableTx) GetSigners() ([][]byte, error) {
	signers := make([][]byte, len(m.signers))
	for i, addr := range m.signers {
		signers[i] = addr.Bytes()
	}
	return signers, nil
}

func (m MockSigVerifiableTx) GetPubKeys() ([]cryptotypes.PubKey, error) {
	return nil, nil
}

func (m MockSigVerifiableTx) GetSignaturesV2() ([]signing.SignatureV2, error) {
	return nil, nil
}

// Test helper to create a test context
func createTestContext(blockHeight int64) sdk.Context {
	return sdk.Context{}.WithBlockHeight(blockHeight).WithEventManager(sdk.NewEventManager())
}

// Test helper to create test addresses
func createTestAddresses(count int) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, count)
	for i := 0; i < count; i++ {
		addrs[i] = sdk.AccAddress([]byte{byte(i + 1), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	}
	return addrs
}

func TestNewRateLimitDecorator(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	decorator := NewRateLimitDecorator(params, nil)

	assert.NotNil(t, decorator.store)
	assert.NotNil(t, decorator.logger)
	assert.NotNil(t, decorator.metrics)
}

func TestRateLimitDecorator_SimulationSkip(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	ctx := createTestContext(1)
	tx := MockTx{}

	nextCalled := false
	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		nextCalled = true
		return ctx, nil
	}

	// Simulation should skip rate limiting
	_, err := decorator.AnteHandle(ctx, tx, true, nextHandler)
	require.NoError(t, err)
	assert.True(t, nextCalled, "next handler should be called during simulation")
}

func TestRateLimitDecorator_DisabledRateLimiting(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	params.Enabled = false
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	ctx := createTestContext(1)
	addrs := createTestAddresses(1)
	tx := MockSigVerifiableTx{
		MockTx: MockTx{signers: addrs},
	}

	nextCalled := false
	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		nextCalled = true
		return ctx, nil
	}

	// Disabled rate limiting should pass through
	_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
	require.NoError(t, err)
	assert.True(t, nextCalled, "next handler should be called when rate limiting is disabled")
}

func TestRateLimitDecorator_UnderLimit(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	params.MaxTxPerBlockPerAccount = 5
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	ctx := createTestContext(1)
	addrs := createTestAddresses(1)
	tx := MockSigVerifiableTx{
		MockTx: MockTx{signers: addrs},
	}

	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}

	// First few transactions should pass
	for i := 0; i < 4; i++ {
		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.NoError(t, err, "transaction %d should pass", i+1)
	}
}

func TestRateLimitDecorator_AtLimit(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	params.MaxTxPerBlockPerAccount = 3
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	ctx := createTestContext(1)
	addrs := createTestAddresses(1)
	tx := MockSigVerifiableTx{
		MockTx: MockTx{signers: addrs},
	}

	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}

	// First 3 transactions should pass
	for i := 0; i < 3; i++ {
		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.NoError(t, err, "transaction %d should pass", i+1)
	}

	// 4th transaction should be rejected
	_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account rate limited")
}

func TestRateLimitDecorator_DifferentBlockResets(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	params.MaxTxPerBlockPerAccount = 2
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	addrs := createTestAddresses(1)
	tx := MockSigVerifiableTx{
		MockTx: MockTx{signers: addrs},
	}

	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}

	// Block 1: Use up the limit
	ctx1 := createTestContext(1)
	for i := 0; i < 2; i++ {
		_, err := decorator.AnteHandle(ctx1, tx, false, nextHandler)
		require.NoError(t, err)
	}

	// Block 1: Should be at limit
	_, err := decorator.AnteHandle(ctx1, tx, false, nextHandler)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account rate limited")

	// Block 2: Counter should reset
	ctx2 := createTestContext(2)
	_, err = decorator.AnteHandle(ctx2, tx, false, nextHandler)
	require.NoError(t, err, "counter should reset for new block")
}

func TestRateLimitDecorator_MultipleAccounts(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	params.MaxTxPerBlockPerAccount = 2
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	ctx := createTestContext(1)
	addrs := createTestAddresses(2)

	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}

	// Account 1 uses up limit
	tx1 := MockSigVerifiableTx{MockTx: MockTx{signers: []sdk.AccAddress{addrs[0]}}}
	for i := 0; i < 2; i++ {
		_, err := decorator.AnteHandle(ctx, tx1, false, nextHandler)
		require.NoError(t, err)
	}

	// Account 1 should be rate limited
	_, err := decorator.AnteHandle(ctx, tx1, false, nextHandler)
	require.Error(t, err)

	// Account 2 should still be able to transact
	tx2 := MockSigVerifiableTx{MockTx: MockTx{signers: []sdk.AccAddress{addrs[1]}}}
	_, err = decorator.AnteHandle(ctx, tx2, false, nextHandler)
	require.NoError(t, err, "different account should not be affected")
}

func TestRateLimitDecorator_ExemptAddresses(t *testing.T) {
	addrs := createTestAddresses(2)

	params := apptypes.DefaultRateLimitParams()
	params.MaxTxPerBlockPerAccount = 1
	params.ExemptAddresses = []string{addrs[0].String()}
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	ctx := createTestContext(1)

	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}

	// Exempt address should not be rate limited
	tx1 := MockSigVerifiableTx{MockTx: MockTx{signers: []sdk.AccAddress{addrs[0]}}}
	for i := 0; i < 10; i++ {
		_, err := decorator.AnteHandle(ctx, tx1, false, nextHandler)
		require.NoError(t, err, "exempt address should not be rate limited")
	}

	// Non-exempt address should be rate limited
	tx2 := MockSigVerifiableTx{MockTx: MockTx{signers: []sdk.AccAddress{addrs[1]}}}
	_, err := decorator.AnteHandle(ctx, tx2, false, nextHandler)
	require.NoError(t, err) // First tx passes

	_, err = decorator.AnteHandle(ctx, tx2, false, nextHandler)
	require.Error(t, err) // Second tx is blocked
}

func TestRateLimitDecorator_BlockLimit(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	params.MaxTotalTxPerBlock = 3
	params.MaxTxPerBlockPerAccount = 10 // High account limit
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	ctx := createTestContext(1)
	addrs := createTestAddresses(5)

	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}

	// First 3 transactions from different accounts should pass
	for i := 0; i < 3; i++ {
		tx := MockSigVerifiableTx{MockTx: MockTx{signers: []sdk.AccAddress{addrs[i]}}}
		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.NoError(t, err)
	}

	// 4th transaction should be blocked by block limit
	tx := MockSigVerifiableTx{MockTx: MockTx{signers: []sdk.AccAddress{addrs[3]}}}
	_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "block rate limited")
}

func TestRateLimitDecorator_GetMetrics(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	params.MaxTxPerBlockPerAccount = 1
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	ctx := createTestContext(1)
	addrs := createTestAddresses(1)
	tx := MockSigVerifiableTx{MockTx: MockTx{signers: addrs}}

	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}

	// First tx passes
	_, _ = decorator.AnteHandle(ctx, tx, false, nextHandler)

	// Second tx blocked
	_, _ = decorator.AnteHandle(ctx, tx, false, nextHandler)

	metrics := decorator.GetMetrics()
	assert.Equal(t, uint64(1), metrics.TotalBlocked)
	assert.Equal(t, uint64(1), metrics.AccountBlocked)
}

func TestRateLimitDecorator_UpdateParams(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	newParams := apptypes.RateLimitParams{
		Enabled:                 true,
		MaxTxPerBlockPerAccount: 20,
		MaxVEIDTxPerBlockGlobal: 200,
		MaxTotalTxPerBlock:      10000,
		ExemptAddresses:         []string{},
	}

	err := decorator.UpdateParams(newParams)
	require.NoError(t, err)
}

func TestRateLimitDecorator_UpdateParamsInvalid(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	invalidParams := apptypes.RateLimitParams{
		Enabled:                 true,
		MaxTxPerBlockPerAccount: 0, // Invalid
		MaxVEIDTxPerBlockGlobal: 100,
		MaxTotalTxPerBlock:      5000,
		ExemptAddresses:         []string{},
	}

	err := decorator.UpdateParams(invalidParams)
	require.Error(t, err)
}

func TestRateLimitDecorator_EnableDisable(t *testing.T) {
	params := apptypes.DefaultRateLimitParams()
	params.MaxTxPerBlockPerAccount = 1
	decorator := NewRateLimitDecorator(params, log.NewNopLogger())

	ctx := createTestContext(1)
	addrs := createTestAddresses(1)
	tx := MockSigVerifiableTx{MockTx: MockTx{signers: addrs}}

	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}

	// First tx passes
	_, _ = decorator.AnteHandle(ctx, tx, false, nextHandler)

	// Second tx would be blocked
	_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
	require.Error(t, err)

	// Disable rate limiting
	decorator.DisableRateLimiting()

	// Now tx should pass (new block to reset counters)
	ctx = createTestContext(2)
	_, err = decorator.AnteHandle(ctx, tx, false, nextHandler)
	require.NoError(t, err)

	// Fill up again
	_, err = decorator.AnteHandle(ctx, tx, false, nextHandler)
	require.NoError(t, err, "should pass when disabled")

	// Re-enable
	decorator.EnableRateLimiting()

	// New block, should be limited again
	ctx = createTestContext(3)
	_, err = decorator.AnteHandle(ctx, tx, false, nextHandler)
	require.NoError(t, err) // First passes

	_, err = decorator.AnteHandle(ctx, tx, false, nextHandler)
	require.Error(t, err) // Second blocked
}

func TestRateLimitParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  apptypes.RateLimitParams
		wantErr bool
	}{
		{
			name:    "valid default params",
			params:  apptypes.DefaultRateLimitParams(),
			wantErr: false,
		},
		{
			name: "zero max tx per account",
			params: apptypes.RateLimitParams{
				Enabled:                 true,
				MaxTxPerBlockPerAccount: 0,
				MaxVEIDTxPerBlockGlobal: 100,
				MaxTotalTxPerBlock:      5000,
			},
			wantErr: true,
		},
		{
			name: "zero max veid tx",
			params: apptypes.RateLimitParams{
				Enabled:                 true,
				MaxTxPerBlockPerAccount: 10,
				MaxVEIDTxPerBlockGlobal: 0,
				MaxTotalTxPerBlock:      5000,
			},
			wantErr: true,
		},
		{
			name: "zero max total tx",
			params: apptypes.RateLimitParams{
				Enabled:                 true,
				MaxTxPerBlockPerAccount: 10,
				MaxVEIDTxPerBlockGlobal: 100,
				MaxTotalTxPerBlock:      0,
			},
			wantErr: true,
		},
		{
			name: "invalid exempt address",
			params: apptypes.RateLimitParams{
				Enabled:                 true,
				MaxTxPerBlockPerAccount: 10,
				MaxVEIDTxPerBlockGlobal: 100,
				MaxTotalTxPerBlock:      5000,
				ExemptAddresses:         []string{"invalid-address"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTransientRateLimitStore_ResetForBlock(t *testing.T) {
	store := apptypes.NewTransientRateLimitStore(apptypes.DefaultRateLimitParams())

	addr := sdk.AccAddress([]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

	// Block 1: Add some counts
	store.ResetForBlock(1)
	store.IncrementAccountTxCount(addr)
	store.IncrementAccountTxCount(addr)
	store.IncrementVEIDTxCount()
	store.IncrementTotalTxCount()

	assert.Equal(t, uint64(2), store.GetAccountTxCount(addr))
	assert.Equal(t, uint64(1), store.GetVEIDTxCount())
	assert.Equal(t, uint64(1), store.GetTotalTxCount())

	// Block 2: Counters should reset
	store.ResetForBlock(2)
	assert.Equal(t, uint64(0), store.GetAccountTxCount(addr))
	assert.Equal(t, uint64(0), store.GetVEIDTxCount())
	assert.Equal(t, uint64(0), store.GetTotalTxCount())
}

func TestIsVEIDTypeURL(t *testing.T) {
	tests := []struct {
		typeURL  string
		expected bool
	}{
		{"/virtengine.veid.v1.MsgUploadScope", true},
		{"/virtengine.veid.v1.MsgRevokeScope", true},
		{"/virtengine.veid.v1.MsgRequestVerification", true},
		{"/virtengine.veid.v1.MsgUpdateVerificationStatus", true},
		{"/virtengine.veid.v1.MsgUpdateScore", true},
		{"/virtengine.veid.v1.MsgRebindWallet", true},
		{"/cosmos.bank.v1beta1.MsgSend", false},
		{"/virtengine.market.v1.MsgCreateOrder", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.typeURL, func(t *testing.T) {
			result := isVEIDTypeURL(tt.typeURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}
