package treasury

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestLedgerReconciliation(t *testing.T) {
	ledger := NewLedger()

	ledger.RecordConversion(LedgerEntry{
		FromAsset:    "UVE",
		ToAsset:      "USDC",
		InputAmount:  sdkmath.NewInt(1000),
		OutputAmount: sdkmath.NewInt(2000),
		FeeAmount:    sdkmath.NewInt(10),
		FeeAsset:     "USDC",
	})

	balances := ledger.Balances()
	require.Equal(t, sdkmath.NewInt(-1000), balances["UVE"])
	require.Equal(t, sdkmath.NewInt(1990), balances["USDC"])

	report := ledger.Reconcile(map[string]sdkmath.Int{
		"UVE":  sdkmath.NewInt(-1000),
		"USDC": sdkmath.NewInt(1990),
	})
	require.True(t, report.Balanced)

	report = ledger.Reconcile(map[string]sdkmath.Int{
		"UVE":  sdkmath.NewInt(-1000),
		"USDC": sdkmath.NewInt(1980),
	})
	require.False(t, report.Balanced)
}
