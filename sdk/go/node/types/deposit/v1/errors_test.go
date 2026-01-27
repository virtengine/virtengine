package v1_test

import (
	"testing"

	sdkerrors "cosmossdk.io/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
	attr "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
)

func TestErrorGRPCStatusCodes(t *testing.T) {
	tests := []struct {
		name             string
		err              *sdkerrors.Error
		expectedGRPCCode codes.Code
		expectedABCICode uint32
	}{
		{
			name:             "invalid_depositor",
			err:              v1.ErrInvalidDepositor,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: attr.ErrLast,
		},
		{
			name:             "invalid_deposit_source",
			err:              v1.ErrInvalidDepositSource,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: attr.ErrLast + 1,
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

