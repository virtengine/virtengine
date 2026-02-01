package keeper_test

import (
	"crypto/sha256"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	_ "github.com/virtengine/virtengine/sdk/go/sdkutil" // Import for init side-effect (Bech32 config)
	"github.com/virtengine/virtengine/x/settlement/types"
)

func TestPayoutRecordValidation(t *testing.T) {
	// Generate valid test addresses using SDK
	_, pubKey1 := createTestKeyPair(t, "provider")
	_, pubKey2 := createTestKeyPair(t, "customer")
	validProvider := sdk.AccAddress(pubKey1.Address()).String()
	validCustomer := sdk.AccAddress(pubKey2.Address()).String()

	testCases := []struct {
		name        string
		payout      types.PayoutRecord
		expectError bool
	}{
		{
			name: "valid payout record",
			payout: types.PayoutRecord{
				PayoutID:       "payout-1",
				InvoiceID:      "inv-1",
				SettlementID:   "settle-1",
				EscrowID:       "escrow-1",
				OrderID:        "order-1",
				Provider:       validProvider,
				Customer:       validCustomer,
				GrossAmount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				PlatformFee:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(50))),
				ValidatorFee:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10))),
				HoldbackAmount: sdk.NewCoins(),
				NetAmount:      sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(940))),
				State:          types.PayoutStatePending,
				IdempotencyKey: "payout-inv-1-settle-1",
			},
			expectError: false,
		},
		{
			name: "empty payout ID",
			payout: types.PayoutRecord{
				PayoutID:     "",
				InvoiceID:    "inv-1",
				SettlementID: "settle-1",
				Provider:     validProvider,
				Customer:     validCustomer,
				GrossAmount:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				NetAmount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(940))),
				State:        types.PayoutStatePending,
			},
			expectError: true,
		},
		{
			name: "missing invoice and settlement",
			payout: types.PayoutRecord{
				PayoutID:    "payout-1",
				InvoiceID:   "",
				SettlementID: "",
				Provider:    validProvider,
				Customer:    validCustomer,
				GrossAmount: sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				NetAmount:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(940))),
				State:       types.PayoutStatePending,
			},
			expectError: true,
		},
		{
			name: "invalid provider address",
			payout: types.PayoutRecord{
				PayoutID:     "payout-1",
				InvoiceID:    "inv-1",
				SettlementID: "settle-1",
				Provider:     "invalid",
				Customer:     validCustomer,
				GrossAmount:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				NetAmount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(940))),
				State:        types.PayoutStatePending,
			},
			expectError: true,
		},
		{
			name: "zero gross amount",
			payout: types.PayoutRecord{
				PayoutID:     "payout-1",
				InvoiceID:    "inv-1",
				SettlementID: "settle-1",
				Provider:     validProvider,
				Customer:     validCustomer,
				GrossAmount:  sdk.NewCoins(),
				NetAmount:    sdk.NewCoins(),
				State:        types.PayoutStatePending,
			},
			expectError: true,
		},
		{
			name: "invalid state",
			payout: types.PayoutRecord{
				PayoutID:     "payout-1",
				InvoiceID:    "inv-1",
				SettlementID: "settle-1",
				Provider:     validProvider,
				Customer:     validCustomer,
				GrossAmount:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				NetAmount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(940))),
				State:        types.PayoutState("invalid"),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.payout.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPayoutStateTransitions(t *testing.T) {
	testCases := []struct {
		name        string
		fromState   types.PayoutState
		toState     types.PayoutState
		shouldAllow bool
	}{
		// From Pending
		{"pending to processing", types.PayoutStatePending, types.PayoutStateProcessing, true},
		{"pending to held", types.PayoutStatePending, types.PayoutStateHeld, true},
		{"pending to cancelled", types.PayoutStatePending, types.PayoutStateCancelled, true},
		{"pending to completed", types.PayoutStatePending, types.PayoutStateCompleted, false},
		{"pending to refunded", types.PayoutStatePending, types.PayoutStateRefunded, false},

		// From Processing
		{"processing to completed", types.PayoutStateProcessing, types.PayoutStateCompleted, true},
		{"processing to failed", types.PayoutStateProcessing, types.PayoutStateFailed, true},
		{"processing to held", types.PayoutStateProcessing, types.PayoutStateHeld, true},
		{"processing to pending", types.PayoutStateProcessing, types.PayoutStatePending, false},

		// From Held
		{"held to processing", types.PayoutStateHeld, types.PayoutStateProcessing, true},
		{"held to refunded", types.PayoutStateHeld, types.PayoutStateRefunded, true},
		{"held to cancelled", types.PayoutStateHeld, types.PayoutStateCancelled, true},
		{"held to completed", types.PayoutStateHeld, types.PayoutStateCompleted, false},

		// From Failed
		{"failed to processing", types.PayoutStateFailed, types.PayoutStateProcessing, true},
		{"failed to completed", types.PayoutStateFailed, types.PayoutStateCompleted, false},

		// From Completed (terminal)
		{"completed to pending", types.PayoutStateCompleted, types.PayoutStatePending, false},
		{"completed to failed", types.PayoutStateCompleted, types.PayoutStateFailed, false},

		// From Refunded (terminal)
		{"refunded to pending", types.PayoutStateRefunded, types.PayoutStatePending, false},
		{"refunded to held", types.PayoutStateRefunded, types.PayoutStateHeld, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.fromState.CanTransitionTo(tc.toState)
			require.Equal(t, tc.shouldAllow, result)
		})
	}
}

func TestPayoutIsTerminal(t *testing.T) {
	terminalStates := []types.PayoutState{
		types.PayoutStateCompleted,
		types.PayoutStateFailed,
		types.PayoutStateRefunded,
		types.PayoutStateCancelled,
	}

	nonTerminalStates := []types.PayoutState{
		types.PayoutStatePending,
		types.PayoutStateProcessing,
		types.PayoutStateHeld,
	}

	for _, state := range terminalStates {
		t.Run(string(state)+"_is_terminal", func(t *testing.T) {
			require.True(t, state.IsTerminal())
		})
	}

	for _, state := range nonTerminalStates {
		t.Run(string(state)+"_is_not_terminal", func(t *testing.T) {
			require.False(t, state.IsTerminal())
		})
	}
}

func TestNewPayoutRecord(t *testing.T) {
	_, pubKey1 := createTestKeyPair(t, "provider")
	_, pubKey2 := createTestKeyPair(t, "customer")
	validProvider := sdk.AccAddress(pubKey1.Address()).String()
	validCustomer := sdk.AccAddress(pubKey2.Address()).String()

	grossAmount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	platformFee := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(50)))
	validatorFee := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))
	holdbackAmount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100)))

	now := time.Now()
	payout := types.NewPayoutRecord(
		"payout-1",
		"inv-1",
		"settle-1",
		"escrow-1",
		"order-1",
		"lease-1",
		validProvider,
		validCustomer,
		grossAmount,
		platformFee,
		validatorFee,
		holdbackAmount,
		now,
		12345,
	)

	require.NotNil(t, payout)
	require.Equal(t, "payout-1", payout.PayoutID)
	require.Equal(t, types.PayoutStatePending, payout.State)
	require.Equal(t, uint32(0), payout.ExecutionAttempts)

	// Check net amount calculation: 1000 - 50 - 10 - 100 = 840
	expectedNet := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(840)))
	require.True(t, payout.NetAmount.Equal(expectedNet))

	// Check idempotency key
	require.Equal(t, "payout-inv-1-settle-1", payout.IdempotencyKey)
}

func TestPayoutMarkProcessing(t *testing.T) {
	_, pubKey1 := createTestKeyPair(t, "provider")
	_, pubKey2 := createTestKeyPair(t, "customer")
	validProvider := sdk.AccAddress(pubKey1.Address()).String()
	validCustomer := sdk.AccAddress(pubKey2.Address()).String()

	payout := &types.PayoutRecord{
		PayoutID:          "payout-1",
		InvoiceID:         "inv-1",
		SettlementID:      "settle-1",
		Provider:          validProvider,
		Customer:          validCustomer,
		GrossAmount:       sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
		NetAmount:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(940))),
		State:             types.PayoutStatePending,
		ExecutionAttempts: 0,
	}

	now := time.Now()
	err := payout.MarkProcessing(now)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateProcessing, payout.State)
	require.Equal(t, uint32(1), payout.ExecutionAttempts)
	require.NotNil(t, payout.ProcessedAt)
	require.NotNil(t, payout.LastAttemptAt)

	// Try marking processing from invalid state
	completedPayout := &types.PayoutRecord{
		State: types.PayoutStateCompleted,
	}
	err = completedPayout.MarkProcessing(now)
	require.Error(t, err)
}

func TestPayoutMarkCompleted(t *testing.T) {
	payout := &types.PayoutRecord{
		PayoutID: "payout-1",
		State:    types.PayoutStateProcessing,
	}

	now := time.Now()
	err := payout.MarkCompleted("tx-hash-123", now)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateCompleted, payout.State)
	require.Equal(t, "tx-hash-123", payout.TxHash)
	require.NotNil(t, payout.CompletedAt)
	require.Empty(t, payout.LastError)

	// Try completing from invalid state
	pendingPayout := &types.PayoutRecord{
		State: types.PayoutStatePending,
	}
	err = pendingPayout.MarkCompleted("tx-hash", now)
	require.Error(t, err)
}

func TestPayoutMarkFailed(t *testing.T) {
	payout := &types.PayoutRecord{
		PayoutID: "payout-1",
		State:    types.PayoutStateProcessing,
	}

	now := time.Now()
	err := payout.MarkFailed("insufficient funds", now)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateFailed, payout.State)
	require.Equal(t, "insufficient funds", payout.LastError)
	require.NotNil(t, payout.LastAttemptAt)
}

func TestPayoutHold(t *testing.T) {
	payout := &types.PayoutRecord{
		PayoutID: "payout-1",
		State:    types.PayoutStatePending,
	}

	now := time.Now()
	err := payout.Hold("dispute-1", "customer complaint", now)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateHeld, payout.State)
	require.Equal(t, "dispute-1", payout.DisputeID)
	require.Equal(t, "customer complaint", payout.HoldReason)

	// Try holding from completed state
	completedPayout := &types.PayoutRecord{
		State: types.PayoutStateCompleted,
	}
	err = completedPayout.Hold("dispute-1", "reason", now)
	require.Error(t, err)
}

func TestPayoutReleaseHold(t *testing.T) {
	payout := &types.PayoutRecord{
		PayoutID:   "payout-1",
		State:      types.PayoutStateHeld,
		DisputeID:  "dispute-1",
		HoldReason: "customer complaint",
	}

	err := payout.ReleaseHold()
	require.NoError(t, err)
	require.Equal(t, types.PayoutStatePending, payout.State)
	require.Empty(t, payout.DisputeID)
	require.Empty(t, payout.HoldReason)

	// Try releasing from non-held state
	pendingPayout := &types.PayoutRecord{
		State: types.PayoutStatePending,
	}
	err = pendingPayout.ReleaseHold()
	require.Error(t, err)
}

func TestPayoutRefund(t *testing.T) {
	payout := &types.PayoutRecord{
		PayoutID: "payout-1",
		State:    types.PayoutStateHeld,
	}

	now := time.Now()
	err := payout.Refund("dispute resolved in customer favor", now)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateRefunded, payout.State)
	require.Equal(t, "dispute resolved in customer favor", payout.HoldReason)

	// Try refunding from pending state
	pendingPayout := &types.PayoutRecord{
		State: types.PayoutStatePending,
	}
	err = pendingPayout.Refund("reason", now)
	require.Error(t, err)
}

func TestPayoutLedgerEntry(t *testing.T) {
	now := time.Now()
	entry := types.NewPayoutLedgerEntry(
		"entry-1",
		"payout-1",
		types.PayoutLedgerEntryCompleted,
		types.PayoutStateProcessing,
		types.PayoutStateCompleted,
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(940))),
		"payout completed successfully",
		"system",
		"tx-hash-123",
		12345,
		now,
	)

	require.NotNil(t, entry)
	require.Equal(t, "entry-1", entry.EntryID)
	require.Equal(t, "payout-1", entry.PayoutID)
	require.Equal(t, types.PayoutLedgerEntryCompleted, entry.EntryType)
	require.Equal(t, types.PayoutStateProcessing, entry.PreviousState)
	require.Equal(t, types.PayoutStateCompleted, entry.NewState)
	require.Equal(t, "system", entry.Initiator)
	require.Equal(t, "tx-hash-123", entry.TransactionHash)
}

func TestTreasuryRecordTypes(t *testing.T) {
	testCases := []struct {
		recordType types.TreasuryRecordType
		expected   string
	}{
		{types.TreasuryRecordPlatformFee, "platform_fee"},
		{types.TreasuryRecordValidatorFee, "validator_fee"},
		{types.TreasuryRecordHoldback, "holdback"},
		{types.TreasuryRecordRefund, "refund"},
		{types.TreasuryRecordWithdrawal, "withdrawal"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.recordType.String())
		})
	}
}

func TestPayoutLedgerEntryTypes(t *testing.T) {
	testCases := []struct {
		entryType types.PayoutLedgerEntryType
		expected  string
	}{
		{types.PayoutLedgerEntryCreated, "created"},
		{types.PayoutLedgerEntryProcessing, "processing"},
		{types.PayoutLedgerEntryCompleted, "completed"},
		{types.PayoutLedgerEntryFailed, "failed"},
		{types.PayoutLedgerEntryHeld, "held"},
		{types.PayoutLedgerEntryReleased, "released"},
		{types.PayoutLedgerEntryRefunded, "refunded"},
		{types.PayoutLedgerEntryCancelled, "cancelled"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.entryType.String())
		})
	}
}

func TestPayoutStoreKeys(t *testing.T) {
	// Test payout key building
	payoutKey := types.PayoutKey("payout-123")
	require.True(t, len(payoutKey) > 0)
	require.Equal(t, types.PrefixPayout[0], payoutKey[0])

	// Test payout by invoice key
	invoiceKey := types.PayoutByInvoiceKey("inv-123")
	require.True(t, len(invoiceKey) > 0)
	require.Equal(t, types.PrefixPayoutByInvoice[0], invoiceKey[0])

	// Test payout by settlement key
	settlementKey := types.PayoutBySettlementKey("settle-123")
	require.True(t, len(settlementKey) > 0)
	require.Equal(t, types.PrefixPayoutBySettlement[0], settlementKey[0])

	// Test payout by provider key
	providerKey := types.PayoutByProviderKey("cosmos1abc", "payout-123")
	require.True(t, len(providerKey) > 0)
	require.Equal(t, types.PrefixPayoutByProvider[0], providerKey[0])

	// Test payout by state key
	stateKey := types.PayoutByStateKey(types.PayoutStatePending, "payout-123")
	require.True(t, len(stateKey) > 0)
	require.Equal(t, types.PrefixPayoutByState[0], stateKey[0])

	// Test idempotency key
	idempotencyKey := types.PayoutIdempotencyKey("payout-inv-1-settle-1")
	require.True(t, len(idempotencyKey) > 0)
	require.Equal(t, types.PrefixPayoutIdempotency[0], idempotencyKey[0])
}

func TestPayoutParamsDefaults(t *testing.T) {
	params := types.DefaultParams()

	require.Equal(t, "0.0", params.PayoutHoldbackRate)
	require.Equal(t, uint32(3), params.MaxPayoutRetries)
	require.Equal(t, uint64(604800), params.DisputeWindowDuration) // 7 days
}

// createTestKeyPair generates a test key pair for testing
//
//nolint:unparam // result 0 (AccAddress) is useful for other test cases
func createTestKeyPair(t *testing.T, name string) (sdk.AccAddress, *ed25519.PubKey) {
	t.Helper()
	// Generate deterministic keys based on name for reproducibility
	seed := sha256.Sum256([]byte(name))
	privKey := ed25519.GenPrivKeyFromSecret(seed[:])
	pubKey := privKey.PubKey().(*ed25519.PubKey)
	return sdk.AccAddress(pubKey.Address()), pubKey
}
