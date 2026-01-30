package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateIdentityWalletParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateIdentityWalletParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid params",
			params: CreateIdentityWalletParams{
				AccountAddress:   "cosmos1test",
				BindingSignature: []byte("signature"),
				BindingPubKey:    []byte("pubkey"),
			},
			wantErr: false,
		},
		{
			name: "empty account address",
			params: CreateIdentityWalletParams{
				AccountAddress:   "",
				BindingSignature: []byte("signature"),
				BindingPubKey:    []byte("pubkey"),
			},
			wantErr: true,
			errMsg:  "account address cannot be empty",
		},
		{
			name: "empty binding signature",
			params: CreateIdentityWalletParams{
				AccountAddress:   "cosmos1test",
				BindingSignature: nil,
				BindingPubKey:    []byte("pubkey"),
			},
			wantErr: true,
			errMsg:  "binding_signature cannot be empty",
		},
		{
			name: "empty binding pub key",
			params: CreateIdentityWalletParams{
				AccountAddress:   "cosmos1test",
				BindingSignature: []byte("signature"),
				BindingPubKey:    nil,
			},
			wantErr: true,
			errMsg:  "binding_pub_key cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateIdentityWalletParams_Validate(t *testing.T) {
	score := uint32(75)
	invalidScore := uint32(150)
	status := AccountStatusVerified
	invalidStatus := AccountStatus("invalid")

	tests := []struct {
		name    string
		params  UpdateIdentityWalletParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid params",
			params: UpdateIdentityWalletParams{
				AccountAddress:   "cosmos1test",
				Score:            &score,
				Status:           &status,
				ModelVersion:     "v1.0.0",
				ValidatorAddress: "cosmosvaloper1test",
			},
			wantErr: false,
		},
		{
			name: "empty account address",
			params: UpdateIdentityWalletParams{
				AccountAddress: "",
				Score:          &score,
			},
			wantErr: true,
			errMsg:  "account address cannot be empty",
		},
		{
			name: "score exceeds maximum",
			params: UpdateIdentityWalletParams{
				AccountAddress: "cosmos1test",
				Score:          &invalidScore,
			},
			wantErr: true,
			errMsg:  "exceeds maximum",
		},
		{
			name: "invalid status",
			params: UpdateIdentityWalletParams{
				AccountAddress: "cosmos1test",
				Status:         &invalidStatus,
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRevokeScopeParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  RevokeScopeParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid params",
			params: RevokeScopeParams{
				AccountAddress: "cosmos1test",
				ScopeID:        "scope_123",
				Reason:         "user requested",
				UserSignature:  []byte("signature"),
			},
			wantErr: false,
		},
		{
			name: "empty account address",
			params: RevokeScopeParams{
				AccountAddress: "",
				ScopeID:        "scope_123",
				UserSignature:  []byte("signature"),
			},
			wantErr: true,
			errMsg:  "account address cannot be empty",
		},
		{
			name: "empty scope id",
			params: RevokeScopeParams{
				AccountAddress: "cosmos1test",
				ScopeID:        "",
				UserSignature:  []byte("signature"),
			},
			wantErr: true,
			errMsg:  "scope_id cannot be empty",
		},
		{
			name: "empty signature",
			params: RevokeScopeParams{
				AccountAddress: "cosmos1test",
				ScopeID:        "scope_123",
				UserSignature:  nil,
			},
			wantErr: true,
			errMsg:  "user_signature cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestToggleScopeConsentParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  ToggleScopeConsentParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid grant consent",
			params: ToggleScopeConsentParams{
				AccountAddress: "cosmos1test",
				ScopeID:        "scope_123",
				GrantConsent:   true,
				Purpose:        "identity verification",
				UserSignature:  []byte("signature"),
			},
			wantErr: false,
		},
		{
			name: "valid revoke consent",
			params: ToggleScopeConsentParams{
				AccountAddress: "cosmos1test",
				ScopeID:        "scope_123",
				GrantConsent:   false,
				UserSignature:  []byte("signature"),
			},
			wantErr: false,
		},
		{
			name: "grant consent without purpose",
			params: ToggleScopeConsentParams{
				AccountAddress: "cosmos1test",
				ScopeID:        "scope_123",
				GrantConsent:   true,
				Purpose:        "",
				UserSignature:  []byte("signature"),
			},
			wantErr: true,
			errMsg:  "purpose is required when granting consent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWalletQueryFilter_Matches(t *testing.T) {
	now := time.Now()
	activeStatus := WalletStatusActive
	suspendedStatus := WalletStatusSuspended
	minScore := uint32(50)
	maxScore := uint32(75)
	basicTier := IdentityTierBasic
	hasActive := true
	idDocType := ScopeTypeIDDocument

	// Create test wallet
	wallet := &IdentityWallet{
		WalletID:       "wallet_test",
		AccountAddress: "cosmos1test",
		Status:         WalletStatusActive,
		CurrentScore:   60,
		Tier:           IdentityTierBasic,
		ScopeRefs: []ScopeReference{
			{
				ScopeID:        "scope_1",
				ScopeType:      ScopeTypeIDDocument,
				Status:         ScopeRefStatusActive,
				EnvelopeHash:   make([]byte, 32),
				AddedAt:        now.Add(-time.Hour),
				ConsentGranted: true,
			},
		},
	}

	tests := []struct {
		name    string
		filter  WalletQueryFilter
		matches bool
	}{
		{
			name:    "empty filter matches all",
			filter:  WalletQueryFilter{},
			matches: true,
		},
		{
			name: "status filter matches",
			filter: WalletQueryFilter{
				Status: &activeStatus,
			},
			matches: true,
		},
		{
			name: "status filter doesn't match",
			filter: WalletQueryFilter{
				Status: &suspendedStatus,
			},
			matches: false,
		},
		{
			name: "min score filter matches",
			filter: WalletQueryFilter{
				MinScore: &minScore,
			},
			matches: true,
		},
		{
			name: "max score filter matches",
			filter: WalletQueryFilter{
				MaxScore: &maxScore,
			},
			matches: true,
		},
		{
			name: "tier filter matches",
			filter: WalletQueryFilter{
				Tier: &basicTier,
			},
			matches: true,
		},
		{
			name: "has active scopes filter matches",
			filter: WalletQueryFilter{
				HasActiveScopes: &hasActive,
			},
			matches: true,
		},
		{
			name: "scope type filter matches",
			filter: WalletQueryFilter{
				ScopeType: &idDocType,
			},
			matches: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.Matches(wallet, now)
			require.Equal(t, tt.matches, result)
		})
	}
}

func TestWalletBindingData_ComputeBindingHash(t *testing.T) {
	binding := WalletBindingData{
		WalletID:       "wallet_test",
		AccountAddress: "cosmos1test",
		BindingPubKey:  []byte("test_pubkey"),
		BoundAt:        time.Now(),
	}

	hash := binding.ComputeBindingHash()
	require.Len(t, hash, 32)

	// Same input produces same hash
	hash2 := binding.ComputeBindingHash()
	require.Equal(t, hash, hash2)
}

func TestIdentityWallet_GetVerificationState(t *testing.T) {
	now := time.Now()

	wallet := &IdentityWallet{
		WalletID:       "wallet_test",
		AccountAddress: "cosmos1test",
		CurrentScore:   75,
		ScoreStatus:    AccountStatusVerified,
		Tier:           IdentityTierStandard,
		ScopeRefs: []ScopeReference{
			{
				ScopeID:        "scope_id",
				ScopeType:      ScopeTypeIDDocument,
				Status:         ScopeRefStatusActive,
				EnvelopeHash:   make([]byte, 32),
				AddedAt:        now.Add(-time.Hour),
				ConsentGranted: true,
			},
			{
				ScopeID:        "scope_selfie",
				ScopeType:      ScopeTypeSelfie,
				Status:         ScopeRefStatusActive,
				EnvelopeHash:   make([]byte, 32),
				AddedAt:        now.Add(-time.Hour),
				ConsentGranted: false,
			},
		},
		VerificationHistory: []VerificationHistoryEntry{
			{
				EntryID:      "vhe_1",
				Timestamp:    now.Add(-time.Minute),
				NewScore:     75,
				NewStatus:    AccountStatusVerified,
				ModelVersion: "v1.0.0",
			},
		},
		ConsentSettings: ConsentSettings{
			ScopeConsents: map[string]ScopeConsent{
				"scope_id": {
					ScopeID: "scope_id",
					Granted: true,
				},
			},
		},
	}

	state := wallet.GetVerificationState(now)

	require.Equal(t, uint32(75), state.Score)
	require.Equal(t, AccountStatusVerified, state.Status)
	require.Equal(t, IdentityTierStandard, state.Tier)
	require.Equal(t, 2, state.ActiveScopeCount)
	require.Equal(t, 2, state.TotalScopeCount)
	require.True(t, state.HasRequiredScopes)
	require.Equal(t, "v1.0.0", state.ModelVersion)
}

func TestIdentityWallet_GetEligibility(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name             string
		score            uint32
		status           AccountStatus
		scopes           []ScopeReference
		canAccessBasic   bool
		canAccessStandard bool
		canAccessPremium bool
	}{
		{
			name:             "unverified - no access",
			score:            0,
			status:           AccountStatusUnknown,
			canAccessBasic:   false,
			canAccessStandard: false,
			canAccessPremium: false,
		},
		{
			name:             "basic tier - basic access",
			score:            50,
			status:           AccountStatusVerified,
			canAccessBasic:   true,
			canAccessStandard: false,
			canAccessPremium: false,
		},
		{
			name:             "standard tier - basic and standard access",
			score:            70,
			status:           AccountStatusVerified,
			canAccessBasic:   true,
			canAccessStandard: true,
			canAccessPremium: false,
		},
		{
			name:             "premium tier - all access",
			score:            85,
			status:           AccountStatusVerified,
			canAccessBasic:   true,
			canAccessStandard: true,
			canAccessPremium: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := &IdentityWallet{
				CurrentScore: tt.score,
				ScoreStatus:  tt.status,
				ScopeRefs:    tt.scopes,
			}

			eligibility := wallet.GetEligibility(now)
			require.Equal(t, tt.canAccessBasic, eligibility.CanAccessBasic)
			require.Equal(t, tt.canAccessStandard, eligibility.CanAccessStandard)
			require.Equal(t, tt.canAccessPremium, eligibility.CanAccessPremium)
		})
	}
}

func TestNewWalletEvent(t *testing.T) {
	now := time.Now()
	blockHeight := int64(1000)

	event := NewWalletEvent(
		WalletEventCreated,
		"wallet_123",
		"cosmos1test",
		blockHeight,
		now,
	)

	require.Equal(t, WalletEventCreated, event.EventType)
	require.Equal(t, "wallet_123", event.WalletID)
	require.Equal(t, "cosmos1test", event.AccountAddress)
	require.Equal(t, blockHeight, event.BlockHeight)
	require.Equal(t, now, event.Timestamp)
	require.NotNil(t, event.Details)

	// Test AddDetail
	event.AddDetail("key", "value")
	require.Equal(t, "value", event.Details["key"])
}
