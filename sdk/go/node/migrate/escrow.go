package migrate

import (
	"bytes"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dv1beta3 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta3"
	eid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	ev1 "github.com/virtengine/virtengine/sdk/go/node/escrow/v1"
	"github.com/virtengine/virtengine/sdk/go/node/escrow/v1beta3"
	deposit "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
)

func AccountV1beta3Prefix() []byte {
	return v1beta3.AccountKeyPrefix()
}

func PaymentV1beta3Prefix() []byte {
	return v1beta3.PaymentKeyPrefix()
}

func AccountIDFromV1beta3(key []byte) eid.Account {
	prefix := v1beta3.AccountKeyPrefix()

	if len(key) < len(prefix)+1 {
		panic("invalid escrow.v1beta3 key")
	}

	if !bytes.Equal(prefix, key[:len(prefix)]) {
		panic("invalid escrow.v1beta3 account prefix")
	}

	key = key[len(prefix):]
	if key[0] != '/' {
		panic("invalid escrow.v1beta3 account separator")
	}

	key = key[1:]

	parts := strings.Split(string(key), "/")
	if len(parts) != 3 && len(parts) != 6 {
		panic("invalid escrow.v1beta3 account xid")
	}

	scope := parts[0]
	if scope != "deployment" && scope != "bid" {
		panic(fmt.Sprintf("invalid escrow account scope \"%s\"", scope))
	}
	return eid.Account{
		Scope: eid.Scope(eid.Scope_value[scope]),
		XID:   strings.Join(parts[1:], "/"),
	}
}

func PaymentIDFromV1beta3(key []byte) eid.Payment {
	prefix := v1beta3.PaymentKeyPrefix()

	if len(key) < len(prefix)+1 {
		panic("invalid escrow.v1beta3 payment key")
	}

	if !bytes.Equal(prefix, key[:len(prefix)]) {
		panic("invalid escrow.v1beta3 payment prefix")
	}

	key = key[len(prefix):]
	if key[0] != '/' {
		panic("invalid escrow.v1beta3 payment separator")
	}

	key = key[1:]

	parts := strings.Split(string(key), "/")
	if len(parts) != 6 {
		panic("invalid escrow.v1beta3 payment xid")
	}

	return eid.Payment{
		AID: eid.Account{
			Scope: eid.Scope(eid.Scope_value[parts[0]]),
			XID:   strings.Join(parts[1:3], "/"),
		},
		XID: strings.Join(parts[3:], "/"),
	}
}

func AccountFromV1beta3(cdc codec.BinaryCodec, key []byte, val []byte) etypes.Account {
	id := AccountIDFromV1beta3(key)

	var from v1beta3.Account
	cdc.MustUnmarshal(val, &from)

	deposits := make([]etypes.Depositor, 0)

	if from.Funds.IsPositive() {
		deposits = append(deposits, etypes.Depositor{
			Owner:   from.Depositor,
			Height:  0,
			Balance: from.Funds,
			Source:  deposit.SourceGrant,
		})
	}

	if from.Balance.IsPositive() {
		deposits = append(deposits, etypes.Depositor{
			Owner:   from.Owner,
			Height:  0,
			Balance: from.Balance,
			Source:  deposit.SourceBalance,
		})
	}

	to := etypes.Account{
		ID: id,
		State: etypes.AccountState{
			Owner: from.Owner,
			State: etypes.State(from.State),
			Funds: []etypes.Balance{
				{
					Denom:  from.Balance.Denom,
					Amount: from.Balance.Add(from.Funds).Amount,
				},
			},
			Transferred: sdk.DecCoins{
				from.Transferred,
			},
			SettledAt: from.SettledAt,
			Deposits:  deposits,
		},
	}

	return to
}

func PaymentFromV1beta3(cdc codec.BinaryCodec, key []byte, val []byte) etypes.Payment {
	id := PaymentIDFromV1beta3(key)

	var from v1beta3.FractionalPayment
	cdc.MustUnmarshal(val, &from)

	to := etypes.Payment{
		ID: id,
		State: etypes.PaymentState{
			Owner:     from.Owner,
			State:     etypes.State(from.State),
			Rate:      from.Rate,
			Balance:   from.Balance,
			Unsettled: sdk.NewDecCoin(from.Balance.Denom, sdkmath.ZeroInt()),
			Withdrawn: from.Withdrawn,
		},
	}

	return to
}

func DepositAuthorizationFromV1beta3(from dv1beta3.DepositDeploymentAuthorization) ev1.DepositAuthorization {
	return ev1.DepositAuthorization{
		Scopes: ev1.DepositAuthorizationScopes{
			ev1.DepositScopeDeployment,
		},
		SpendLimit: from.SpendLimit,
	}
}
