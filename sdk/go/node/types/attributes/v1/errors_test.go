package v1_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
)

func TestErrorGRPCStatusCodes(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode codes.Code
	}{
		{
			name:         "attributes_duplicate_keys_returns_invalid_argument",
			err:          v1.ErrAttributesDuplicateKeys,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "invalid_attribute_key_returns_invalid_argument",
			err:          v1.ErrInvalidAttributeKey,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, ok := status.FromError(tt.err)
			require.True(t, ok, "error should be convertible to gRPC status")
			require.Equal(t, tt.expectedCode, st.Code(), "gRPC status code mismatch")
		})
	}
}

