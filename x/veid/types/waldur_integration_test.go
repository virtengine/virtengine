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
				"waldur-user-456",
				"cosmos1abc...",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty link ID",
			record: &types.WaldurLinkRecord{
				Version:        types.WaldurIntegrationVersion,
				LinkID:         "",
				AccountAddress: "cosmos1abc...",
				WaldurUserID:   "waldur-user-456",
				LinkedAt:       now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty account address",
			record: &types.WaldurLinkRecord{
				Version:        types.WaldurIntegrationVersion,
				LinkID:         "link-123",
				AccountAddress: "",
				WaldurUserID:   "waldur-user-456",
				LinkedAt:       now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty waldur user ID",
			record: &types.WaldurLinkRecord{
				Version:        types.WaldurIntegrationVersion,
				LinkID:         "link-123",
				AccountAddress: "cosmos1abc...",
				WaldurUserID:   "",
				LinkedAt:       now,
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
				"waldur-user-456",
				"cosmos1abc...",
				[]types.ScopeType{types.ScopeTypeSelfie},
				now,
				3600, // TTL seconds
			),
			wantErr: false,
		},
		{
			name: "invalid - empty request ID",
			request: &types.WaldurUploadRequest{
				RequestID:       "",
				WaldurUserID:    "waldur-user-456",
				AccountAddress:  "cosmos1abc...",
				RequestedScopes: []types.ScopeType{types.ScopeTypeSelfie},
				CreatedAt:       now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty account address",
			request: &types.WaldurUploadRequest{
				RequestID:       "upload-123",
				WaldurUserID:    "waldur-user-456",
				AccountAddress:  "",
				RequestedScopes: []types.ScopeType{types.ScopeTypeSelfie},
				CreatedAt:       now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty waldur user ID",
			request: &types.WaldurUploadRequest{
				RequestID:       "upload-123",
				WaldurUserID:    "",
				AccountAddress:  "cosmos1abc...",
				RequestedScopes: []types.ScopeType{types.ScopeTypeSelfie},
				CreatedAt:       now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty requested scopes",
			request: &types.WaldurUploadRequest{
				RequestID:       "upload-123",
				WaldurUserID:    "waldur-user-456",
				AccountAddress:  "cosmos1abc...",
				RequestedScopes: []types.ScopeType{},
				CreatedAt:       now,
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

func TestWaldurUploadRequest_IsExpired(t *testing.T) {
	now := time.Now()

	// Not expired
	request := types.NewWaldurUploadRequest(
		"upload-123",
		"waldur-user-456",
		"cosmos1abc...",
		[]types.ScopeType{types.ScopeTypeSelfie},
		now,
		3600, // TTL seconds
	)

	assert.False(t, request.IsExpired(now), "Request should not be expired")
	assert.False(t, request.IsExpired(now.Add(30*time.Minute)), "Request should not be expired after 30 minutes")

	// Expired
	assert.True(t, request.IsExpired(now.Add(2*time.Hour)), "Request should be expired after 2 hours")
}

func TestWaldurLinkRecord_Unlink(t *testing.T) {
	now := time.Now()

	record := types.NewWaldurLinkRecord(
		"link-123",
		"waldur-user-456",
		"cosmos1abc...",
		now,
	)

	assert.True(t, record.IsActive, "New record should be active")

	// Unlink the record
	unlinkTime := now.Add(1 * time.Hour)
	record.Unlink(unlinkTime, "user requested")

	assert.False(t, record.IsActive, "Record should not be active after unlink")
	require.NotNil(t, record.UnlinkedAt, "UnlinkedAt should be set")
	assert.True(t, record.UnlinkedAt.Equal(unlinkTime), "UnlinkedAt should be set to unlink time")
	assert.Equal(t, "user requested", record.UnlinkReason, "UnlinkReason should be set")
}

func TestWaldurLinkRecord_RecordSync(t *testing.T) {
	now := time.Now()

	record := types.NewWaldurLinkRecord(
		"link-123",
		"waldur-user-456",
		"cosmos1abc...",
		now,
	)

	assert.Nil(t, record.LastSyncAt, "LastSyncAt should be nil initially")

	// Record a sync
	syncTime := now.Add(1 * time.Hour)
	record.RecordSync(syncTime)

	require.NotNil(t, record.LastSyncAt, "LastSyncAt should be set")
	assert.True(t, record.LastSyncAt.Equal(syncTime), "LastSyncAt should be set to sync time")
}

func TestWaldurCallback_Creation(t *testing.T) {
	now := time.Now()

	callback := types.NewWaldurCallback(
		"callback-123",
		types.WaldurCallbackVerificationComplete,
		"waldur-user-456",
		"cosmos1abc...",
		types.WaldurStateVerified,
		now,
	)

	require.NotNil(t, callback)
	assert.Equal(t, "callback-123", callback.CallbackID)
	assert.Equal(t, types.WaldurCallbackVerificationComplete, callback.Type)
	assert.Equal(t, "waldur-user-456", callback.WaldurUserID)
	assert.Equal(t, "cosmos1abc...", callback.AccountAddress)
	assert.Equal(t, types.WaldurStateVerified, callback.State)
	assert.True(t, callback.Timestamp.Equal(now))
}

func TestWaldurIdentityStatus_Creation(t *testing.T) {
	now := time.Now()

	status := types.NewWaldurIdentityStatus(
		"cosmos1abc...",
		types.WaldurStateVerified,
		85,
		"verified",
		now,
	)

	require.NotNil(t, status)
	assert.Equal(t, types.WaldurIntegrationVersion, status.Version)
	assert.Equal(t, "cosmos1abc...", status.AccountAddress)
	assert.Equal(t, types.WaldurStateVerified, status.State)
	assert.Equal(t, uint32(85), status.Score)
	assert.Equal(t, "verified", status.Tier)
	assert.True(t, status.LastUpdatedAt.Equal(now))
}
