package treasury

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestCustodyPolicyEnforcement(t *testing.T) {
	policy := DefaultWithdrawalPolicy()
	policy.MaxHotPerTx = sdkmath.NewInt(500)
	policy.MaxHotDaily = sdkmath.NewInt(1000)
	policy.RequireMultiSigAbove = sdkmath.NewInt(300)
	policy.RequiredApprovals = 2
	policy.BlockedDestinations["bad-dest"] = struct{}{}

	manager := NewCustodyManager(policy)
	manager.AddWallet(&Wallet{
		ID:               "hot-1",
		Address:          "hot-addr",
		Type:             WalletTypeHot,
		Status:           WalletStatusActive,
		LastRotatedAt:    time.Now().Add(-48 * time.Hour),
		RotationInterval: 24 * time.Hour,
		Balances: map[string]sdkmath.Int{
			"UVE": sdkmath.NewInt(5000),
		},
	})
	manager.AddWallet(&Wallet{
		ID:      "cold-1",
		Address: "cold-addr",
		Type:    WalletTypeCold,
		Status:  WalletStatusActive,
		Balances: map[string]sdkmath.Int{
			"UVE": sdkmath.NewInt(20000),
		},
	})

	blocked, err := manager.RequestWithdrawal("UVE", sdkmath.NewInt(100), "bad-dest")
	require.ErrorIs(t, err, ErrPolicyBlocked)
	require.Equal(t, WithdrawalStatusBlocked, blocked.Status)

	req, err := manager.RequestWithdrawal("UVE", sdkmath.NewInt(400), "good-dest")
	require.ErrorIs(t, err, ErrApprovalRequired)
	require.Equal(t, WithdrawalStatusPendingApproval, req.Status)

	req, err = manager.ApproveWithdrawal(req.ID, "signer-1")
	require.NoError(t, err)
	require.Equal(t, WithdrawalStatusPendingApproval, req.Status)

	req, err = manager.ApproveWithdrawal(req.ID, "signer-2")
	require.NoError(t, err)
	require.Equal(t, WithdrawalStatusApproved, req.Status)

	req, err = manager.ExecuteWithdrawal(req.ID)
	require.NoError(t, err)
	require.Equal(t, WithdrawalStatusExecuted, req.Status)
}

func TestCustodyRotationLog(t *testing.T) {
	policy := DefaultWithdrawalPolicy()
	manager := NewCustodyManager(policy)

	manager.AddWallet(&Wallet{
		ID:               "hot-rot",
		Address:          "hot-addr",
		Type:             WalletTypeHot,
		Status:           WalletStatusActive,
		LastRotatedAt:    time.Now().Add(-2 * time.Hour),
		RotationInterval: 1 * time.Hour,
		Balances: map[string]sdkmath.Int{
			"UVE": sdkmath.NewInt(1000),
		},
	})

	wallet, err := manager.RotateWalletIfDue("hot-rot", "scheduled")
	require.NoError(t, err)
	require.NotEqual(t, "hot-addr", wallet.Address)

	logs := manager.RotationLogs()
	require.Len(t, logs, 1)
}
