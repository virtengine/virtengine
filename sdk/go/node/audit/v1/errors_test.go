package v1_test

import (
	"testing"

	sdkerrors "cosmossdk.io/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
)

func TestErrorGRPCStatusCodes(t *testing.T) {
	tests := []struct {
		name             string
		err              *sdkerrors.Error
		expectedGRPCCode codes.Code
		expectedABCICode uint32
	}{
		{
			name:             "provider_not_found",
			err:              v1.ErrProviderNotFound,
			expectedGRPCCode: codes.NotFound,
			expectedABCICode: 1,
		},
		{
			name:             "invalid_address",
			err:              v1.ErrInvalidAddress,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 2,
		},
		{
			name:             "attribute_not_found",
			err:              v1.ErrAttributeNotFound,
			expectedGRPCCode: codes.NotFound,
			expectedABCICode: 3,
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
