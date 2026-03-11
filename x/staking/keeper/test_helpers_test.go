package keeper

import (
	"context"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type mockStakeKeeper struct {
	stakes             map[string]int64
	validatorPayoutBPS map[string]int64
	slashCalls         []mockSlashCall
	rewardDistCalls    []mockRewardDistribution
}

type mockSlashCall struct {
	ValidatorAddress string
	Fraction         string
	InfractionHeight int64
	Amount           int64
}

type mockRewardDistribution struct {
	ValidatorAddress string
	Epoch            uint64
	ValidatorAmount  int64
	DelegatorAmount  int64
	Denom            string
}

func newMockStakeKeeper() *mockStakeKeeper {
	return &mockStakeKeeper{
		stakes:             make(map[string]int64),
		validatorPayoutBPS: make(map[string]int64),
	}
}

func (m *mockStakeKeeper) SetStake(validatorAddr string, stake int64) {
	m.stakes[validatorAddr] = stake
}

func (m *mockStakeKeeper) SetValidatorPayoutBPS(validatorAddr string, payoutBPS int64) {
	m.validatorPayoutBPS[validatorAddr] = payoutBPS
}

func (m *mockStakeKeeper) GetAllValidators(_ sdk.Context) []sdk.AccAddress {
	validators := make([]string, 0, len(m.stakes))
	for validatorAddr := range m.stakes {
		validators = append(validators, validatorAddr)
	}
	sort.Strings(validators)

	out := make([]sdk.AccAddress, 0, len(validators))
	for _, validatorAddr := range validators {
		addr, err := sdk.AccAddressFromBech32(validatorAddr)
		if err != nil {
			continue
		}
		out = append(out, addr)
	}
	return out
}

func (m *mockStakeKeeper) GetValidatorStake(_ sdk.Context, validatorAddr sdk.AccAddress) int64 {
	return m.stakes[validatorAddr.String()]
}

func (m *mockStakeKeeper) GetTotalStake(_ sdk.Context) int64 {
	total := int64(0)
	for _, stake := range m.stakes {
		total += stake
	}
	return total
}

func (m *mockStakeKeeper) SlashDelegations(_ sdk.Context, validatorAddr string, fraction sdkmath.LegacyDec, infractionHeight int64) error {
	stake := m.stakes[validatorAddr]
	amount := fraction.MulInt64(stake).TruncateInt64()
	if amount > stake {
		amount = stake
	}
	m.stakes[validatorAddr] = stake - amount
	m.slashCalls = append(m.slashCalls, mockSlashCall{
		ValidatorAddress: validatorAddr,
		Fraction:         fraction.String(),
		InfractionHeight: infractionHeight,
		Amount:           amount,
	})
	return nil
}

func (m *mockStakeKeeper) DistributeValidatorRewardsToDelegators(_ sdk.Context, validatorAddr string, epoch uint64, validatorReward sdk.Coins) (sdk.Coins, sdk.Coins, error) {
	if validatorReward.IsZero() {
		return sdk.NewCoins(), sdk.NewCoins(), nil
	}

	denom := validatorReward[0].Denom
	total := validatorReward.AmountOf(denom)
	payoutBPS := m.validatorPayoutBPS[validatorAddr]
	if payoutBPS <= 0 {
		payoutBPS = 10000
	}
	if payoutBPS > 10000 {
		payoutBPS = 10000
	}

	validatorAmount := total.MulRaw(payoutBPS).QuoRaw(10000)
	delegatorAmount := total.Sub(validatorAmount)

	m.rewardDistCalls = append(m.rewardDistCalls, mockRewardDistribution{
		ValidatorAddress: validatorAddr,
		Epoch:            epoch,
		ValidatorAmount:  validatorAmount.Int64(),
		DelegatorAmount:  delegatorAmount.Int64(),
		Denom:            denom,
	})

	validatorCoins := sdk.NewCoins()
	if validatorAmount.IsPositive() {
		validatorCoins = sdk.NewCoins(sdk.NewCoin(denom, validatorAmount))
	}
	delegatorCoins := sdk.NewCoins()
	if delegatorAmount.IsPositive() {
		delegatorCoins = sdk.NewCoins(sdk.NewCoin(denom, delegatorAmount))
	}

	return validatorCoins, delegatorCoins, nil
}

type mockBankKeeper struct {
	minted           []sdk.Coins
	moduleTransfers  []mockModuleTransfer
	accountTransfers []mockAccountTransfer
	accountToModule  []mockAccountToModuleTransfer
}

type mockModuleTransfer struct {
	SenderModule    string
	RecipientModule string
	Amount          sdk.Coins
}

type mockAccountTransfer struct {
	SenderModule string
	Recipient    string
	Amount       sdk.Coins
}

type mockAccountToModuleTransfer struct {
	Sender    string
	Recipient string
	Amount    sdk.Coins
}

func (m *mockBankKeeper) SendCoins(_ context.Context, _ sdk.AccAddress, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	m.accountTransfers = append(m.accountTransfers, mockAccountTransfer{
		SenderModule: senderModule,
		Recipient:    recipientAddr.String(),
		Amount:       amt,
	})
	return nil
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	m.accountToModule = append(m.accountToModule, mockAccountToModuleTransfer{
		Sender:    senderAddr.String(),
		Recipient: recipientModule,
		Amount:    amt,
	})
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToModule(_ context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	m.moduleTransfers = append(m.moduleTransfers, mockModuleTransfer{
		SenderModule:    senderModule,
		RecipientModule: recipientModule,
		Amount:          amt,
	})
	return nil
}

func (m *mockBankKeeper) SpendableCoins(_ context.Context, _ sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins()
}

func (m *mockBankKeeper) MintCoins(_ context.Context, _ string, amt sdk.Coins) error {
	m.minted = append(m.minted, amt)
	return nil
}
