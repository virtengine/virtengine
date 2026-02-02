package module_test

import (
	"testing"

	sdkerrors "cosmossdk.io/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/virtengine/virtengine/sdk/go/node/escrow/module"
)

func TestErrorGRPCStatusCodes(t *testing.T) {
	tests := []struct {
		name             string
		err              *sdkerrors.Error
		expectedGRPCCode codes.Code
		expectedABCICode uint32
	}{
		{
			name:             "account_exists",
			err:              module.ErrAccountExists,
			expectedGRPCCode: codes.AlreadyExists,
			expectedABCICode: 1,
		},
		{
			name:             "account_closed",
			err:              module.ErrAccountClosed,
			expectedGRPCCode: codes.FailedPrecondition,
			expectedABCICode: 2,
		},
		{
			name:             "account_not_found",
			err:              module.ErrAccountNotFound,
			expectedGRPCCode: codes.NotFound,
			expectedABCICode: 3,
		},
		{
			name:             "account_overdrawn",
			err:              module.ErrAccountOverdrawn,
			expectedGRPCCode: codes.FailedPrecondition,
			expectedABCICode: 4,
		},
		{
			name:             "invalid_denomination",
			err:              module.ErrInvalidDenomination,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 5,
		},
		{
			name:             "payment_exists",
			err:              module.ErrPaymentExists,
			expectedGRPCCode: codes.AlreadyExists,
			expectedABCICode: 6,
		},
		{
			name:             "payment_closed",
			err:              module.ErrPaymentClosed,
			expectedGRPCCode: codes.FailedPrecondition,
			expectedABCICode: 7,
		},
		{
			name:             "payment_not_found",
			err:              module.ErrPaymentNotFound,
			expectedGRPCCode: codes.NotFound,
			expectedABCICode: 8,
		},
		{
			name:             "payment_rate_zero",
			err:              module.ErrPaymentRateZero,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 9,
		},
		{
			name:             "invalid_payment",
			err:              module.ErrInvalidPayment,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 10,
		},
		{
			name:             "invalid_settlement",
			err:              module.ErrInvalidSettlement,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 11,
		},
		{
			name:             "invalid_id",
			err:              module.ErrInvalidID,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 12,
		},
		{
			name:             "invalid_account",
			err:              module.ErrInvalidAccount,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 13,
		},
		{
			name:             "invalid_account_depositor",
			err:              module.ErrInvalidAccountDepositor,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 14,
		},
		{
			name:             "unauthorized_deposit_scope",
			err:              module.ErrUnauthorizedDepositScope,
			expectedGRPCCode: codes.PermissionDenied,
			expectedABCICode: 15,
		},
		{
			name:             "invalid_deposit",
			err:              module.ErrInvalidDeposit,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 16,
		},
		{
			name:             "invalid_authz_scope",
			err:              module.ErrInvalidAuthzScope,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 17,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, ok := status.FromError(tt.err)
			require.True(t, ok, "error should be convertible to gRPC status")
			require.Equal(t, tt.expectedGRPCCode, st.Code(), "gRPC status code mismatch")
			require.Equal(t, tt.expectedABCICode, tt.err.ABCICode(), "ABCI error code mismatch")
		})
	}
}
