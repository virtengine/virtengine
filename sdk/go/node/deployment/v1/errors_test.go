package v1_test

import (
	"testing"

	sdkerrors "cosmossdk.io/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
)

func TestErrorGRPCStatusCodes(t *testing.T) {
	tests := []struct {
		name             string
		err              *sdkerrors.Error
		expectedGRPCCode codes.Code
		expectedABCICode uint32
	}{
		{
			name:             "name_does_not_exist",
			err:              v1.ErrNameDoesNotExist,
			expectedGRPCCode: codes.NotFound,
			expectedABCICode: 1,
		},
		{
			name:             "invalid_request",
			err:              v1.ErrInvalidRequest,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 2,
		},
		{
			name:             "deployment_exists",
			err:              v1.ErrDeploymentExists,
			expectedGRPCCode: codes.AlreadyExists,
			expectedABCICode: 3,
		},
		{
			name:             "deployment_not_found",
			err:              v1.ErrDeploymentNotFound,
			expectedGRPCCode: codes.NotFound,
			expectedABCICode: 4,
		},
		{
			name:             "deployment_closed",
			err:              v1.ErrDeploymentClosed,
			expectedGRPCCode: codes.FailedPrecondition,
			expectedABCICode: 5,
		},
		{
			name:             "owner_account_missing",
			err:              v1.ErrOwnerAcctMissing,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 6,
		},
		{
			name:             "invalid_groups",
			err:              v1.ErrInvalidGroups,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 7,
		},
		{
			name:             "invalid_deployment_id",
			err:              v1.ErrInvalidDeploymentID,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 8,
		},
		{
			name:             "empty_hash",
			err:              v1.ErrEmptyHash,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 9,
		},
		{
			name:             "invalid_hash",
			err:              v1.ErrInvalidHash,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 10,
		},
		{
			name:             "internal",
			err:              v1.ErrInternal,
			expectedGRPCCode: codes.Internal,
			expectedABCICode: 11,
		},
		{
			name:             "invalid_deployment",
			err:              v1.ErrInvalidDeployment,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 12,
		},
		{
			name:             "invalid_group_id",
			err:              v1.ErrInvalidGroupID,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 13,
		},
		{
			name:             "group_not_found",
			err:              v1.ErrGroupNotFound,
			expectedGRPCCode: codes.NotFound,
			expectedABCICode: 14,
		},
		{
			name:             "group_closed",
			err:              v1.ErrGroupClosed,
			expectedGRPCCode: codes.FailedPrecondition,
			expectedABCICode: 15,
		},
		{
			name:             "group_open",
			err:              v1.ErrGroupOpen,
			expectedGRPCCode: codes.FailedPrecondition,
			expectedABCICode: 16,
		},
		{
			name:             "group_paused",
			err:              v1.ErrGroupPaused,
			expectedGRPCCode: codes.FailedPrecondition,
			expectedABCICode: 17,
		},
		{
			name:             "group_not_open",
			err:              v1.ErrGroupNotOpen,
			expectedGRPCCode: codes.FailedPrecondition,
			expectedABCICode: 18,
		},
		{
			name:             "group_spec_invalid",
			err:              v1.ErrGroupSpecInvalid,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 19,
		},
		{
			name:             "invalid_deposit",
			err:              v1.ErrInvalidDeposit,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 20,
		},
		{
			name:             "invalid_id_path",
			err:              v1.ErrInvalidIDPath,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 21,
		},
		{
			name:             "invalid_param",
			err:              v1.ErrInvalidParam,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 22,
		},
		{
			name:             "invalid_escrow_id",
			err:              v1.ErrInvalidEscrowID,
			expectedGRPCCode: codes.InvalidArgument,
			expectedABCICode: 23,
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

