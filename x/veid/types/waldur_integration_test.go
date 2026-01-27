package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Waldur Integration Tests (VE-226: Waldur Integration Interface)
// ============================================================================

func TestWaldurLinkRecord_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		record  *types.WaldurLinkRecord
		wantErr bool
	}{
		{
			name: "valid record",
			record: types.NewWaldurLinkRecord(
				"link-123",
				"cosmos1abc...",
				"waldur-user-456",
				"waldur-org-789",
				"https://waldur.example.com",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty link ID",
			record: &types.WaldurLinkRecord{
				Version:       types.WaldurLinkVersion,
				LinkID:        "",
				VeidAddress:   "cosmos1abc...",
				WaldurUserID:  "waldur-user-456",
				WaldurOrgID:   "waldur-org-789",
				WaldurBaseURL: "https://waldur.example.com",
				LinkedAt:      now,
				State:         types.WaldurStatePending,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty veid address",
			record: &types.WaldurLinkRecord{
				Version:       types.WaldurLinkVersion,
				LinkID:        "link-123",
				VeidAddress:   "",
				WaldurUserID:  "waldur-user-456",
				WaldurOrgID:   "waldur-org-789",
				WaldurBaseURL: "https://waldur.example.com",
				LinkedAt:      now,
				State:         types.WaldurStatePending,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty waldur user ID",
			record: &types.WaldurLinkRecord{
				Version:       types.WaldurLinkVersion,
				LinkID:        "link-123",
				VeidAddress:   "cosmos1abc...",
				WaldurUserID:  "",
				WaldurOrgID:   "waldur-org-789",
				WaldurBaseURL: "https://waldur.example.com",
				LinkedAt:      now,
				State:         types.WaldurStatePending,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty waldur org ID",
			record: &types.WaldurLinkRecord{
				Version:       types.WaldurLinkVersion,
				LinkID:        "link-123",
				VeidAddress:   "cosmos1abc...",
				WaldurUserID:  "waldur-user-456",
				WaldurOrgID:   "",
				WaldurBaseURL: "https://waldur.example.com",
				LinkedAt:      now,
				State:         types.WaldurStatePending,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty base URL",
			record: &types.WaldurLinkRecord{
				Version:       types.WaldurLinkVersion,
				LinkID:        "link-123",
				VeidAddress:   "cosmos1abc...",
				WaldurUserID:  "waldur-user-456",
				WaldurOrgID:   "waldur-org-789",
				WaldurBaseURL: "",
				LinkedAt:      now,
				State:         types.WaldurStatePending,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid state",
			record: &types.WaldurLinkRecord{
				Version:       types.WaldurLinkVersion,
				LinkID:        "link-123",
				VeidAddress:   "cosmos1abc...",
				WaldurUserID:  "waldur-user-456",
				WaldurOrgID:   "waldur-org-789",
				WaldurBaseURL: "https://waldur.example.com",
				LinkedAt:      now,
				State:         "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWaldurVerificationStates(t *testing.T) {
	// Test all states are valid
	for _, state := range types.AllWaldurVerificationStates() {
		assert.True(t, types.IsValidWaldurVerificationState(state), "AllWaldurVerificationStates returned invalid state: %s", state)
	}

	// Test invalid state
	assert.False(t, types.IsValidWaldurVerificationState("invalid"), "IsValidWaldurVerificationState should return false for invalid state")
}

func TestWaldurUploadRequest_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		request *types.WaldurUploadRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: types.NewWaldurUploadRequest(
				"upload-123",
				"cosmos1abc...",
				"waldur-org-789",
				"https://waldur.example.com",
				"veid-identity-hash-abc",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty request ID",
			request: &types.WaldurUploadRequest{
				Version:          types.WaldurUploadVersion,
				RequestID:        "",
				VeidAddress:      "cosmos1abc...",
				WaldurOrgID:      "waldur-org-789",
				WaldurBaseURL:    "https://waldur.example.com",
				VeidIdentityHash: "veid-hash",
				CreatedAt:        now,
				Status:           types.WaldurUploadStatusPending,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty veid address",
			request: &types.WaldurUploadRequest{
				Version:          types.WaldurUploadVersion,
				RequestID:        "upload-123",
				VeidAddress:      "",
				WaldurOrgID:      "waldur-org-789",
				WaldurBaseURL:    "https://waldur.example.com",
				VeidIdentityHash: "veid-hash",
				CreatedAt:        now,
				Status:           types.WaldurUploadStatusPending,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty waldur org ID",
			request: &types.WaldurUploadRequest{
				Version:          types.WaldurUploadVersion,
				RequestID:        "upload-123",
				VeidAddress:      "cosmos1abc...",
				WaldurOrgID:      "",
				WaldurBaseURL:    "https://waldur.example.com",
				VeidIdentityHash: "veid-hash",
				CreatedAt:        now,
				Status:           types.WaldurUploadStatusPending,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty base URL",
			request: &types.WaldurUploadRequest{
				Version:          types.WaldurUploadVersion,
				RequestID:        "upload-123",
				VeidAddress:      "cosmos1abc...",
				WaldurOrgID:      "waldur-org-789",
				WaldurBaseURL:    "",
				VeidIdentityHash: "veid-hash",
				CreatedAt:        now,
				Status:           types.WaldurUploadStatusPending,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty identity hash",
			request: &types.WaldurUploadRequest{
				Version:          types.WaldurUploadVersion,
				RequestID:        "upload-123",
				VeidAddress:      "cosmos1abc...",
				WaldurOrgID:      "waldur-org-789",
				WaldurBaseURL:    "https://waldur.example.com",
				VeidIdentityHash: "",
				CreatedAt:        now,
				Status:           types.WaldurUploadStatusPending,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid status",
			request: &types.WaldurUploadRequest{
				Version:          types.WaldurUploadVersion,
				RequestID:        "upload-123",
				VeidAddress:      "cosmos1abc...",
				WaldurOrgID:      "waldur-org-789",
				WaldurBaseURL:    "https://waldur.example.com",
				VeidIdentityHash: "veid-hash",
				CreatedAt:        now,
				Status:           "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWaldurUploadStatuses(t *testing.T) {
	// Test all statuses are valid
	for _, status := range types.AllWaldurUploadStatuses() {
		assert.True(t, types.IsValidWaldurUploadStatus(status), "AllWaldurUploadStatuses returned invalid status: %s", status)
	}

	// Test invalid status
	assert.False(t, types.IsValidWaldurUploadStatus("invalid"), "IsValidWaldurUploadStatus should return false for invalid status")
}

func TestWaldurUploadResponse_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		response *types.WaldurUploadResponse
		wantErr  bool
	}{
		{
			name: "valid success response",
			response: types.NewWaldurUploadResponse(
				"upload-123",
				true,
				"waldur-identity-789",
				"",
				now,
			),
			wantErr: false,
		},
		{
			name: "valid failure response",
			response: types.NewWaldurUploadResponse(
				"upload-123",
				false,
				"",
				"Upload failed due to validation error",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty request ID",
			response: &types.WaldurUploadResponse{
				Version:         types.WaldurResponseVersion,
				RequestID:       "",
				Success:         true,
				WaldurIdentity:  "waldur-id-123",
				ReceivedAt:      now,
				ExpiresAt:       now.Add(24 * time.Hour),
				VerificationURL: "https://waldur.example.com/verify",
			},
			wantErr: true,
		},
		{
			name: "invalid - success without waldur identity",
			response: &types.WaldurUploadResponse{
				Version:        types.WaldurResponseVersion,
				RequestID:      "upload-123",
				Success:        true,
				WaldurIdentity: "",
				ReceivedAt:     now,
			},
			wantErr: true,
		},
		{
			name: "invalid - failure without error message",
			response: &types.WaldurUploadResponse{
				Version:      types.WaldurResponseVersion,
				RequestID:    "upload-123",
				Success:      false,
				ErrorMessage: "",
				ReceivedAt:   now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.response.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWaldurCallback_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		callback *types.WaldurCallback
		wantErr  bool
	}{
		{
			name: "valid callback",
			callback: types.NewWaldurCallback(
				"callback-123",
				"upload-123",
				types.WaldurCallbackTypeVerified,
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty callback ID",
			callback: &types.WaldurCallback{
				Version:      types.WaldurCallbackVersion,
				CallbackID:   "",
				RequestID:    "upload-123",
				CallbackType: types.WaldurCallbackTypeVerified,
				ReceivedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty request ID",
			callback: &types.WaldurCallback{
				Version:      types.WaldurCallbackVersion,
				CallbackID:   "callback-123",
				RequestID:    "",
				CallbackType: types.WaldurCallbackTypeVerified,
				ReceivedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid callback type",
			callback: &types.WaldurCallback{
				Version:      types.WaldurCallbackVersion,
				CallbackID:   "callback-123",
				RequestID:    "upload-123",
				CallbackType: "invalid",
				ReceivedAt:   now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.callback.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWaldurCallbackTypes(t *testing.T) {
	// Test all callback types are valid
	for _, callbackType := range types.AllWaldurCallbackTypes() {
		assert.True(t, types.IsValidWaldurCallbackType(callbackType), "AllWaldurCallbackTypes returned invalid type: %s", callbackType)
	}

	// Test invalid type
	assert.False(t, types.IsValidWaldurCallbackType("invalid"), "IsValidWaldurCallbackType should return false for invalid type")
}

func TestWaldurIdentityStatus_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		status  *types.WaldurIdentityStatus
		wantErr bool
	}{
		{
			name: "valid status",
			status: types.NewWaldurIdentityStatus(
				"status-123",
				"cosmos1abc...",
				"waldur-identity-456",
				types.WaldurStateVerified,
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty status ID",
			status: &types.WaldurIdentityStatus{
				Version:               types.WaldurStatusVersion,
				StatusID:              "",
				VeidAddress:           "cosmos1abc...",
				WaldurIdentity:        "waldur-id-456",
				VerificationState:     types.WaldurStateVerified,
				LastVerifiedAt:        now,
				VerificationExpiresAt: now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - empty veid address",
			status: &types.WaldurIdentityStatus{
				Version:               types.WaldurStatusVersion,
				StatusID:              "status-123",
				VeidAddress:           "",
				WaldurIdentity:        "waldur-id-456",
				VerificationState:     types.WaldurStateVerified,
				LastVerifiedAt:        now,
				VerificationExpiresAt: now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - empty waldur identity",
			status: &types.WaldurIdentityStatus{
				Version:               types.WaldurStatusVersion,
				StatusID:              "status-123",
				VeidAddress:           "cosmos1abc...",
				WaldurIdentity:        "",
				VerificationState:     types.WaldurStateVerified,
				LastVerifiedAt:        now,
				VerificationExpiresAt: now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid state",
			status: &types.WaldurIdentityStatus{
				Version:               types.WaldurStatusVersion,
				StatusID:              "status-123",
				VeidAddress:           "cosmos1abc...",
				WaldurIdentity:        "waldur-id-456",
				VerificationState:     "invalid",
				LastVerifiedAt:        now,
				VerificationExpiresAt: now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.status.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWaldurLinkRecord_UpdateState(t *testing.T) {
	now := time.Now()

	record := types.NewWaldurLinkRecord(
		"link-123",
		"cosmos1abc...",
		"waldur-user-456",
		"waldur-org-789",
		"https://waldur.example.com",
		now,
	)

	assert.Equal(t, types.WaldurStatePending, record.State, "New record should have pending state")

	// Update to verified
	updateTime := now.Add(1 * time.Hour)
	record.UpdateState(types.WaldurStateVerified, updateTime)

	assert.Equal(t, types.WaldurStateVerified, record.State, "State should be verified")
	require.NotNil(t, record.LastUpdatedAt, "LastUpdatedAt should be set")
	assert.True(t, record.LastUpdatedAt.Equal(updateTime), "LastUpdatedAt should be set to update time")

	// Update to revoked
	revokeTime := now.Add(2 * time.Hour)
	record.UpdateState(types.WaldurStateRevoked, revokeTime)

	assert.Equal(t, types.WaldurStateRevoked, record.State, "State should be revoked")
}

func TestWaldurUploadRequest_Complete(t *testing.T) {
	now := time.Now()

	request := types.NewWaldurUploadRequest(
		"upload-123",
		"cosmos1abc...",
		"waldur-org-789",
		"https://waldur.example.com",
		"veid-identity-hash-abc",
		now,
	)

	assert.Equal(t, types.WaldurUploadStatusPending, request.Status, "New request should have pending status")

	// Complete with success
	completeTime := now.Add(1 * time.Hour)
	request.Complete(true, completeTime)

	assert.Equal(t, types.WaldurUploadStatusCompleted, request.Status, "Status should be completed")
	require.NotNil(t, request.CompletedAt, "CompletedAt should be set")
	assert.True(t, request.CompletedAt.Equal(completeTime), "CompletedAt should be set to complete time")
}

func TestWaldurUploadRequest_Fail(t *testing.T) {
	now := time.Now()

	request := types.NewWaldurUploadRequest(
		"upload-123",
		"cosmos1abc...",
		"waldur-org-789",
		"https://waldur.example.com",
		"veid-identity-hash-abc",
		now,
	)

	// Fail the request
	failTime := now.Add(1 * time.Hour)
	request.Fail("Network timeout", failTime)

	assert.Equal(t, types.WaldurUploadStatusFailed, request.Status, "Status should be failed")
	assert.Equal(t, "Network timeout", request.ErrorMessage, "Error message should be set")
	require.NotNil(t, request.CompletedAt, "CompletedAt should be set even on failure")
}

func TestWaldurIdentityStatus_IsExpired(t *testing.T) {
	now := time.Now()

	// Not expired
	status := types.NewWaldurIdentityStatus(
		"status-123",
		"cosmos1abc...",
		"waldur-identity-456",
		types.WaldurStateVerified,
		now,
	)
	status.VerificationExpiresAt = now.Add(24 * time.Hour)

	assert.False(t, status.IsExpired(now), "Status should not be expired")

	// Expired
	status.VerificationExpiresAt = now.Add(-1 * time.Hour)

	assert.True(t, status.IsExpired(now), "Status should be expired")
}
