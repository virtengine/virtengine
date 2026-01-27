package v1_test

import (
	"testing"

	sdkerrors "cosmossdk.io/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/virtengine/virtengine/sdk/go/node/cert/v1"
)

func TestErrorGRPCStatusCodes(t *testing.T) {
	tests := []struct {
		name             string
		err              *sdkerrors.Error
		expectedGRPCCode codes.Code
		expectedABCICode uint32
	}{
		{
			name:             "certificate_not_found",
			err:              v1.ErrCertificateNotFound,
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
			name:             "certificate_exists",
			err:              v1.ErrCertificateExists,
			expectedGRPCCode: codes.AlreadyExists,
			expectedABCICode: 3,
		},
		{
			name:             "certificate_already_revoked",
			err:              v1.ErrCertificateAlreadyRevoked,
			expectedGRPCCode: codes.FailedPrecondition,
			expectedABCICode: 4,
		},
		{
			name:             "invalid_serial_number",
			err:              v1.ErrInvalidSerialNumber,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 5,
		},
		{
			name:             "invalid_certificate_value",
			err:              v1.ErrInvalidCertificateValue,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 6,
		},
		{
			name:             "invalid_pubkey_value",
			err:              v1.ErrInvalidPubkeyValue,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 7,
		},
		{
			name:             "invalid_state",
			err:              v1.ErrInvalidState,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 8,
		},
		{
			name:             "invalid_key_size",
			err:              v1.ErrInvalidKeySize,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 9,
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

