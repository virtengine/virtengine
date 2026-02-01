package v1beta4_test

import (
	"testing"

	sdkerrors "cosmossdk.io/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1beta4 "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

func TestErrorGRPCStatusCodes(t *testing.T) {
	tests := []struct {
		name             string
		err              *sdkerrors.Error
		expectedGRPCCode codes.Code
		expectedABCICode uint32
	}{
		{
			name:             "invalid_provider_uri",
			err:              v1beta4.ErrInvalidProviderURI,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 1,
		},
		{
			name:             "not_abs_provider_uri",
			err:              v1beta4.ErrNotAbsProviderURI,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 2,
		},
		{
			name:             "provider_not_found",
			err:              v1beta4.ErrProviderNotFound,
			expectedGRPCCode: codes.NotFound,
			expectedABCICode: 3,
		},
		{
			name:             "provider_exists",
			err:              v1beta4.ErrProviderExists,
			expectedGRPCCode: codes.AlreadyExists,
			expectedABCICode: 4,
		},
		{
			name:             "invalid_address",
			err:              v1beta4.ErrInvalidAddress,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 5,
		},
		{
			name:             "attributes",
			err:              v1beta4.ErrAttributes,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 6,
		},
		{
			name:             "incompatible_attributes",
			err:              v1beta4.ErrIncompatibleAttributes,
			expectedGRPCCode: codes.FailedPrecondition,
			expectedABCICode: 7,
		},
		{
			name:             "invalid_info_website",
			err:              v1beta4.ErrInvalidInfoWebsite,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 8,
		},
		{
			name:             "internal",
			err:              v1beta4.ErrInternal,
			expectedGRPCCode: codes.Internal,
			expectedABCICode: 9,
		},
		{
			name:             "provider_has_active_leases",
			err:              v1beta4.ErrProviderHasActiveLeases,
			expectedGRPCCode: codes.FailedPrecondition,
			expectedABCICode: 24,
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

