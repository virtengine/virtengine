package migrate

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	eid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	"github.com/virtengine/virtengine/sdk/go/node/escrow/v1beta3"
	deposit "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
)

func TestAccountIDFromV1beta3(t *testing.T) {
	// Get the prefix directly from v1beta3 module
	prefix := v1beta3.AccountKeyPrefix()

	// Sample data
	akashAddr := "akash1keydahz9uv8fs8u4lk6q3cluaprjpnm7dd3cf0"
	dseq := "1000"

	// Helper function to create valid account keys
	createAccountKey := func(scope string, xid string) []byte {
		return append(prefix, []byte("/"+scope+"/"+xid)...)
	}

	t.Run("valid deployment account", func(t *testing.T) {
		key := createAccountKey("deployment", akashAddr+"/"+dseq)
		expected := eid.Account{
			Scope: eid.ScopeDeployment,
			XID:   akashAddr + "/" + dseq,
		}

		result := AccountIDFromV1beta3(key)
		require.Equal(t, expected, result)
	})

	t.Run("valid bid account", func(t *testing.T) {
		key := createAccountKey("bid", akashAddr+"/"+dseq)
		expected := eid.Account{
			Scope: eid.ScopeBid,
			XID:   akashAddr + "/" + dseq,
		}

		result := AccountIDFromV1beta3(key)
		require.Equal(t, expected, result)
	})

	t.Run("prefix check", func(t *testing.T) {
		invalidPrefix := []byte("/wrong_prefix/deployment/" + akashAddr + "/" + dseq)
		require.Panics(t, func() {
			AccountIDFromV1beta3(invalidPrefix)
		})
	})

	t.Run("separator check", func(t *testing.T) {
		// Missing separator after prefix
		invalidKey := append(bytes.Clone(prefix), []byte("deployment/"+akashAddr+"/"+dseq)...)
		require.Panics(t, func() {
			AccountIDFromV1beta3(invalidKey)
		})
	})
	t.Run("empty XID after separator", func(t *testing.T) {
		// Key has prefix and separator but no XID parts
		shortKey := append(bytes.Clone(prefix), '/')
		require.Panics(t, func() {
			AccountIDFromV1beta3(shortKey)
		})
	})

	t.Run("invalid scope", func(t *testing.T) {
		// Scope is not "deployment" or "bid"
		invalidScope := createAccountKey("invalid", akashAddr+"/"+dseq)
		require.Panics(t, func() {
			AccountIDFromV1beta3(invalidScope)
		})
	})

	t.Run("invalid parts count", func(t *testing.T) {
		// Not enough parts
		tooFewParts := createAccountKey("deployment", akashAddr)
		require.Panics(t, func() {
			AccountIDFromV1beta3(tooFewParts)
		})

		// Too many parts that don't match 3 or 6
		tooManyParts := createAccountKey("deployment", akashAddr+"/"+dseq+"/extra/random")
		require.Panics(t, func() {
			AccountIDFromV1beta3(tooManyParts)
		})
	})

	t.Run("complex XID with 6 parts", func(t *testing.T) {
		// Valid XID with 6 parts
		complexXID := akashAddr + "/1/2/3/4"
		key := createAccountKey("bid", complexXID)
		expected := eid.Account{
			Scope: eid.ScopeBid,
			XID:   complexXID,
		}

		result := AccountIDFromV1beta3(key)
		require.Equal(t, expected, result)
	})
}

func TestAccountFromV1beta3(t *testing.T) {
	// Create a codec for marshaling/unmarshaling
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Helper function to create a valid account key
	createAccountKey := func(scope string, xid string) []byte {
		prefix := v1beta3.AccountKeyPrefix()
		return append(prefix, []byte("/"+scope+"/"+xid)...)
	}

	// Helper function to create test account data
	createTestAccount := func(owner, depositor string, balance, funds sdk.DecCoin, state v1beta3.Account_State) []byte {
		account := v1beta3.Account{
			ID: v1beta3.AccountID{
				Scope: "deployment",
				XID:   "akash1test/1000",
			},
			Owner:       owner,
			Depositor:   depositor,
			Balance:     balance,
			Funds:       funds,
			State:       state,
			Transferred: sdk.NewDecCoin("uve", sdkmath.NewInt(0)),
			SettledAt:   1000,
		}
		data, err := cdc.Marshal(&account)
		require.NoError(t, err)
		return data
	}

	// Test cases for different scenarios
	testCases := []struct {
		name             string
		owner            string
		depositor        string
		balance          sdk.DecCoin
		funds            sdk.DecCoin
		state            v1beta3.Account_State
		expectedDeposits int
		expectedSources  []deposit.Source
		expectedOwners   []string
		expectedBalances []sdk.DecCoin
	}{
		{
			name:             "funds_only_depositor_different_from_owner",
			owner:            "akash1owner",
			depositor:        "akash1depositor",
			balance:          sdk.NewDecCoin("uve", sdkmath.NewInt(0)),
			funds:            sdk.NewDecCoin("uve", sdkmath.NewInt(1000)),
			state:            v1beta3.AccountOpen,
			expectedDeposits: 1,
			expectedSources:  []deposit.Source{deposit.SourceGrant},
			expectedOwners:   []string{"akash1depositor"},
			expectedBalances: []sdk.DecCoin{sdk.NewDecCoin("uve", sdkmath.NewInt(1000))},
		},
		{
			name:             "balance_only_depositor_same_as_owner",
			owner:            "akash1owner",
			depositor:        "akash1owner",
			balance:          sdk.NewDecCoin("uve", sdkmath.NewInt(2000)),
			funds:            sdk.NewDecCoin("uve", sdkmath.NewInt(0)),
			state:            v1beta3.AccountOpen,
			expectedDeposits: 1,
			expectedSources:  []deposit.Source{deposit.SourceBalance},
			expectedOwners:   []string{"akash1owner"},
			expectedBalances: []sdk.DecCoin{sdk.NewDecCoin("uve", sdkmath.NewInt(2000))},
		},
		{
			name:             "both_funds_and_balance_present",
			owner:            "akash1owner",
			depositor:        "akash1depositor",
			balance:          sdk.NewDecCoin("uve", sdkmath.NewInt(1500)),
			funds:            sdk.NewDecCoin("uve", sdkmath.NewInt(500)),
			state:            v1beta3.AccountOpen,
			expectedDeposits: 2,
			expectedSources:  []deposit.Source{deposit.SourceGrant, deposit.SourceBalance},
			expectedOwners:   []string{"akash1depositor", "akash1owner"},
			expectedBalances: []sdk.DecCoin{sdk.NewDecCoin("uve", sdkmath.NewInt(500)), sdk.NewDecCoin("uve", sdkmath.NewInt(1500))},
		},
		{
			name:             "zero_balance_zero_funds",
			owner:            "akash1owner",
			depositor:        "akash1depositor",
			balance:          sdk.NewDecCoin("uve", sdkmath.NewInt(0)),
			funds:            sdk.NewDecCoin("uve", sdkmath.NewInt(0)),
			state:            v1beta3.AccountOpen,
			expectedDeposits: 0,
			expectedSources:  []deposit.Source{},
			expectedOwners:   []string{},
			expectedBalances: []sdk.DecCoin{},
		},
		{
			name:             "negative_balance_zero_funds",
			owner:            "akash1owner",
			depositor:        "akash1owner",
			balance:          sdk.DecCoin{Denom: "uve", Amount: sdkmath.LegacyNewDec(-100)},
			funds:            sdk.NewDecCoin("uve", sdkmath.NewInt(0)),
			state:            v1beta3.AccountOpen,
			expectedDeposits: 0,
			expectedSources:  []deposit.Source{},
			expectedOwners:   []string{},
			expectedBalances: []sdk.DecCoin{},
		},
		{
			name:             "zero_balance_negative_funds",
			owner:            "akash1owner",
			depositor:        "akash1depositor",
			balance:          sdk.NewDecCoin("uve", sdkmath.NewInt(0)),
			funds:            sdk.DecCoin{Denom: "uve", Amount: sdkmath.LegacyNewDec(-50)},
			state:            v1beta3.AccountOpen,
			expectedDeposits: 0,
			expectedSources:  []deposit.Source{},
			expectedOwners:   []string{},
			expectedBalances: []sdk.DecCoin{},
		},
		{
			name:             "small_positive_amounts",
			owner:            "akash1owner",
			depositor:        "akash1depositor",
			balance:          sdk.NewDecCoinFromDec("uve", sdkmath.LegacyNewDecWithPrec(1, 1)), // 0.1
			funds:            sdk.NewDecCoinFromDec("uve", sdkmath.LegacyNewDecWithPrec(1, 2)), // 0.01
			state:            v1beta3.AccountOpen,
			expectedDeposits: 2,
			expectedSources:  []deposit.Source{deposit.SourceGrant, deposit.SourceBalance},
			expectedOwners:   []string{"akash1depositor", "akash1owner"},
			expectedBalances: []sdk.DecCoin{sdk.NewDecCoinFromDec("uve", sdkmath.LegacyNewDecWithPrec(1, 2)), sdk.NewDecCoinFromDec("uve", sdkmath.LegacyNewDecWithPrec(1, 1))},
		},
		{
			name:             "large_amounts",
			owner:            "akash1owner",
			depositor:        "akash1depositor",
			balance:          sdk.NewDecCoin("uve", sdkmath.NewInt(1000000000)),
			funds:            sdk.NewDecCoin("uve", sdkmath.NewInt(500000000)),
			state:            v1beta3.AccountOpen,
			expectedDeposits: 2,
			expectedSources:  []deposit.Source{deposit.SourceGrant, deposit.SourceBalance},
			expectedOwners:   []string{"akash1depositor", "akash1owner"},
			expectedBalances: []sdk.DecCoin{sdk.NewDecCoin("uve", sdkmath.NewInt(500000000)), sdk.NewDecCoin("uve", sdkmath.NewInt(1000000000))},
		},
		{
			name:             "different_denomination",
			owner:            "akash1owner",
			depositor:        "akash1depositor",
			balance:          sdk.NewDecCoin("ibc/123", sdkmath.NewInt(1000)),
			funds:            sdk.NewDecCoin("ibc/123", sdkmath.NewInt(500)),
			state:            v1beta3.AccountOpen,
			expectedDeposits: 2,
			expectedSources:  []deposit.Source{deposit.SourceGrant, deposit.SourceBalance},
			expectedOwners:   []string{"akash1depositor", "akash1owner"},
			expectedBalances: []sdk.DecCoin{sdk.NewDecCoin("ibc/123", sdkmath.NewInt(500)), sdk.NewDecCoin("ibc/123", sdkmath.NewInt(1000))},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test account data
			accountData := createTestAccount(tc.owner, tc.depositor, tc.balance, tc.funds, tc.state)

			// Create account key
			key := createAccountKey("deployment", "akash1test/1000")

			// Call the migration function
			result := AccountFromV1beta3(cdc, key, accountData)

			// Verify basic account structure
			require.Equal(t, tc.owner, result.State.Owner)
			require.Equal(t, etypes.State(tc.state), result.State.State)
			require.Equal(t, int64(1000), result.State.SettledAt)

			// Verify deposits length
			require.Len(t, result.State.Deposits, tc.expectedDeposits, "Expected %d deposits, got %d", tc.expectedDeposits, len(result.State.Deposits))

			// Verify deposits content
			for i, deposit := range result.State.Deposits {
				require.Equal(t, tc.expectedOwners[i], deposit.Owner, "Deposit %d owner mismatch", i)
				require.Equal(t, tc.expectedSources[i], deposit.Source, "Deposit %d source mismatch", i)
				require.Equal(t, tc.expectedBalances[i], deposit.Balance, "Deposit %d balance mismatch", i)
				require.Equal(t, int64(0), deposit.Height, "Deposit %d height should be 0", i)
			}

			// Verify funds calculation (Balance + Funds)
			expectedTotalAmount := tc.balance.Amount.Add(tc.funds.Amount)
			require.Len(t, result.State.Funds, 1, "Expected 1 fund entry")
			require.Equal(t, tc.balance.Denom, result.State.Funds[0].Denom, "Fund denom mismatch")
			require.True(t, expectedTotalAmount.Equal(result.State.Funds[0].Amount),
				"Fund amount mismatch: expected %s, got %s", expectedTotalAmount.String(), result.State.Funds[0].Amount.String())
		})
	}
}
