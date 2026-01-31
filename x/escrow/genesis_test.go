package escrow_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	eid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	emodule "github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	types "github.com/virtengine/virtengine/sdk/go/node/escrow/v1"

	"github.com/virtengine/virtengine/x/escrow"
)

type GenesisTestSuite struct {
	suite.Suite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

// Test: DefaultGenesisState returns valid state
func (s *GenesisTestSuite) TestDefaultGenesisState() {
	genesis := escrow.DefaultGenesisState()

	s.Require().NotNil(genesis)
	s.Require().Empty(genesis.Accounts)
	s.Require().Empty(genesis.Payments)
}

// Test: ValidateGenesis with default state
func (s *GenesisTestSuite) TestValidateGenesis_Default() {
	genesis := escrow.DefaultGenesisState()
	err := escrow.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with valid state
func (s *GenesisTestSuite) TestValidateGenesis_Valid() {
	accountID := eid.Account{
		Scope: eid.ScopeDeployment,
		XID:   "test-account-1",
	}

	paymentID := eid.Payment{
		AID: accountID,
		XID: "payment-1",
	}

	genesis := &types.GenesisState{
		Accounts: []etypes.Account{
			{
				ID: accountID,
				State: etypes.AccountState{
					Owner: "cosmos1abcdefg",
					State: etypes.StateOpen,
					Funds: []etypes.Balance{
						{
							Denom:  "uve",
							Amount: sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(1000)),
						},
					},
				},
			},
		},
		Payments: []etypes.Payment{
			{
				ID: paymentID,
				State: etypes.PaymentState{
					Owner:   "cosmos1provider",
					State:   etypes.StateOpen,
					Rate:    sdk.NewDecCoin("uve", sdkmath.NewInt(10)),
					Balance: sdk.NewDecCoin("uve", sdkmath.NewInt(100)),
				},
			},
		},
	}

	err := escrow.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with invalid account
func (s *GenesisTestSuite) TestValidateGenesis_InvalidAccount() {
	genesis := &types.GenesisState{
		Accounts: []etypes.Account{
			{
				// Empty/invalid account ID
				ID: eid.Account{},
				State: etypes.AccountState{
					Owner: "cosmos1abcdefg",
					State: etypes.StateOpen,
				},
			},
		},
	}

	err := escrow.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with duplicate accounts
func (s *GenesisTestSuite) TestValidateGenesis_DuplicateAccounts() {
	accountID := eid.Account{
		Scope: eid.ScopeDeployment,
		XID:   "duplicate-account",
	}

	genesis := &types.GenesisState{
		Accounts: []etypes.Account{
			{
				ID: accountID,
				State: etypes.AccountState{
					Owner: "cosmos1abcdefg",
					State: etypes.StateOpen,
				},
			},
			{
				ID: accountID, // Duplicate
				State: etypes.AccountState{
					Owner: "cosmos1xyz",
					State: etypes.StateOpen,
				},
			},
		},
	}

	err := escrow.ValidateGenesis(genesis)
	s.Require().Error(err)
	s.Require().ErrorIs(err, emodule.ErrAccountExists)
}

// Test: ValidateGenesis with payment referencing non-existent account
func (s *GenesisTestSuite) TestValidateGenesis_OrphanPayment() {
	accountID := eid.Account{
		Scope: eid.ScopeDeployment,
		XID:   "non-existent-account",
	}

	paymentID := eid.Payment{
		AID: accountID,
		XID: "payment-1",
	}

	genesis := &types.GenesisState{
		// No accounts
		Payments: []etypes.Payment{
			{
				ID: paymentID,
				State: etypes.PaymentState{
					Owner:   "cosmos1provider",
					State:   etypes.StateOpen,
					Balance: sdk.NewDecCoin("uve", sdkmath.NewInt(100)),
					Rate:    sdk.NewDecCoin("uve", sdkmath.NewInt(10)),
				},
			},
		},
	}

	err := escrow.ValidateGenesis(genesis)
	s.Require().Error(err)
	s.Require().ErrorIs(err, emodule.ErrAccountNotFound)
}

// Test: ValidateGenesis with duplicate payments
func (s *GenesisTestSuite) TestValidateGenesis_DuplicatePayments() {
	accountID := eid.Account{
		Scope: eid.ScopeDeployment,
		XID:   "test-account",
	}

	paymentID := eid.Payment{
		AID: accountID,
		XID: "duplicate-payment",
	}

	genesis := &types.GenesisState{
		Accounts: []etypes.Account{
			{
				ID: accountID,
				State: etypes.AccountState{
					Owner: "cosmos1abcdefg",
					State: etypes.StateOpen,
				},
			},
		},
		Payments: []etypes.Payment{
			{
				ID: paymentID,
				State: etypes.PaymentState{
					Owner:   "cosmos1provider1",
					State:   etypes.StateOpen,
					Balance: sdk.NewDecCoin("uve", sdkmath.NewInt(100)),
					Rate:    sdk.NewDecCoin("uve", sdkmath.NewInt(10)),
				},
			},
			{
				ID: paymentID, // Duplicate
				State: etypes.PaymentState{
					Owner:   "cosmos1provider2",
					State:   etypes.StateOpen,
					Balance: sdk.NewDecCoin("uve", sdkmath.NewInt(200)),
					Rate:    sdk.NewDecCoin("uve", sdkmath.NewInt(20)),
				},
			},
		},
	}

	err := escrow.ValidateGenesis(genesis)
	s.Require().Error(err)
	s.Require().ErrorIs(err, emodule.ErrPaymentExists)
}

// Test: ValidateGenesis - payment state mismatch with account state
func (s *GenesisTestSuite) TestValidateGenesis_PaymentStateAccountMismatch() {
	accountID := eid.Account{
		Scope: eid.ScopeDeployment,
		XID:   "closed-account",
	}

	paymentID := eid.Payment{
		AID: accountID,
		XID: "payment-1",
	}

	genesis := &types.GenesisState{
		Accounts: []etypes.Account{
			{
				ID: accountID,
				State: etypes.AccountState{
					Owner: "cosmos1abcdefg",
					State: etypes.StateClosed, // Account is closed
				},
			},
		},
		Payments: []etypes.Payment{
			{
				ID: paymentID,
				State: etypes.PaymentState{
					Owner:   "cosmos1provider",
					State:   etypes.StateOpen, // Payment is still open
					Balance: sdk.NewDecCoin("uve", sdkmath.NewInt(100)),
					Rate:    sdk.NewDecCoin("uve", sdkmath.NewInt(10)),
				},
			},
		},
	}

	err := escrow.ValidateGenesis(genesis)
	s.Require().Error(err)
	s.Require().ErrorIs(err, emodule.ErrInvalidPayment)
}

// Test: Account states
func (s *GenesisTestSuite) TestAccountStates() {
	states := []etypes.State{
		etypes.StateOpen,
		etypes.StateOverdrawn,
		etypes.StateClosed,
	}

	for _, state := range states {
		accountID := eid.Account{
			Scope: eid.ScopeDeployment,
			XID:   "state-test-" + state.String(),
		}

		genesis := &types.GenesisState{
			Accounts: []etypes.Account{
				{
					ID: accountID,
					State: etypes.AccountState{
						Owner: "cosmos1abcdefg",
						State: state,
					},
				},
			},
		}

		err := escrow.ValidateGenesis(genesis)
		s.Require().NoError(err, "state %s should be valid", state.String())
	}
}

// Test: Payment states with matching account states
func (s *GenesisTestSuite) TestPaymentStatesMatching() {
	testCases := []struct {
		name         string
		accountState etypes.State
		paymentState etypes.State
		expectError  bool
	}{
		{
			name:         "both open",
			accountState: etypes.StateOpen,
			paymentState: etypes.StateOpen,
			expectError:  false,
		},
		{
			name:         "both overdrawn",
			accountState: etypes.StateOverdrawn,
			paymentState: etypes.StateOverdrawn,
			expectError:  false,
		},
		{
			name:         "account closed, payment closed",
			accountState: etypes.StateClosed,
			paymentState: etypes.StateClosed,
			expectError:  false,
		},
		{
			name:         "account open, payment overdrawn",
			accountState: etypes.StateOpen,
			paymentState: etypes.StateOverdrawn,
			expectError:  true,
		},
		{
			name:         "account overdrawn, payment open",
			accountState: etypes.StateOverdrawn,
			paymentState: etypes.StateOpen,
			expectError:  true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			accountID := eid.Account{
				Scope: eid.ScopeDeployment,
				XID:   "match-test-" + tc.name,
			}

			paymentID := eid.Payment{
				AID: accountID,
				XID: "payment-1",
			}

			genesis := &types.GenesisState{
				Accounts: []etypes.Account{
					{
						ID: accountID,
						State: etypes.AccountState{
							Owner: "cosmos1abcdefg",
							State: tc.accountState,
						},
					},
				},
				Payments: []etypes.Payment{
					{
						ID: paymentID,
						State: etypes.PaymentState{
							Owner:   "cosmos1provider",
							State:   tc.paymentState,
							Balance: sdk.NewDecCoin("uve", sdkmath.NewInt(100)),
							Rate:    sdk.NewDecCoin("uve", sdkmath.NewInt(10)),
						},
					},
				},
			}

			err := escrow.ValidateGenesis(genesis)
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// Test: Multiple accounts and payments
func (s *GenesisTestSuite) TestValidateGenesis_MultipleAccountsAndPayments() {
	accountID1 := eid.Account{Scope: eid.ScopeDeployment, XID: "account-1"}
	accountID2 := eid.Account{Scope: eid.ScopeDeployment, XID: "account-2"}

	genesis := &types.GenesisState{
		Accounts: []etypes.Account{
			{
				ID: accountID1,
				State: etypes.AccountState{
					Owner: "cosmos1owner1",
					State: etypes.StateOpen,
					Funds: []etypes.Balance{
						{Denom: "uve", Amount: sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(1000))},
					},
				},
			},
			{
				ID: accountID2,
				State: etypes.AccountState{
					Owner: "cosmos1owner2",
					State: etypes.StateOpen,
					Funds: []etypes.Balance{
						{Denom: "uve", Amount: sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(2000))},
					},
				},
			},
		},
		Payments: []etypes.Payment{
			{
				ID: eid.Payment{AID: accountID1, XID: "p1"},
				State: etypes.PaymentState{
					Owner:   "cosmos1provider1",
					State:   etypes.StateOpen,
					Balance: sdk.NewDecCoin("uve", sdkmath.NewInt(100)),
					Rate:    sdk.NewDecCoin("uve", sdkmath.NewInt(10)),
				},
			},
			{
				ID: eid.Payment{AID: accountID1, XID: "p2"},
				State: etypes.PaymentState{
					Owner:   "cosmos1provider2",
					State:   etypes.StateOpen,
					Balance: sdk.NewDecCoin("uve", sdkmath.NewInt(200)),
					Rate:    sdk.NewDecCoin("uve", sdkmath.NewInt(20)),
				},
			},
			{
				ID: eid.Payment{AID: accountID2, XID: "p1"},
				State: etypes.PaymentState{
					Owner:   "cosmos1provider3",
					State:   etypes.StateOpen,
					Balance: sdk.NewDecCoin("uve", sdkmath.NewInt(300)),
					Rate:    sdk.NewDecCoin("uve", sdkmath.NewInt(30)),
				},
			},
		},
	}

	err := escrow.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: Account ID validation
func TestAccountIDValidation(t *testing.T) {
	tests := []struct {
		name        string
		scope       eid.Scope
		xid         string
		expectError bool
	}{
		{
			name:        "valid account ID",
			scope:       eid.ScopeDeployment,
			xid:         "abc123",
			expectError: false,
		},
		{
			name:        "invalid scope",
			scope:       eid.ScopeInvalid,
			xid:         "abc123",
			expectError: true,
		},
		{
			name:        "empty XID",
			scope:       eid.ScopeDeployment,
			xid:         "",
			expectError: true,
		},
		{
			name:        "both invalid",
			scope:       eid.ScopeInvalid,
			xid:         "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			accountID := eid.Account{
				Scope: tc.scope,
				XID:   tc.xid,
			}

			genesis := &types.GenesisState{
				Accounts: []etypes.Account{
					{
						ID: accountID,
						State: etypes.AccountState{
							Owner: "cosmos1abcdefg",
							State: etypes.StateOpen,
						},
					},
				},
			}

			err := escrow.ValidateGenesis(genesis)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
