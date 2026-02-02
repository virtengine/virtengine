package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	encryptioncrypto "github.com/virtengine/virtengine/x/encryption/crypto"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Test Suite
// ============================================================================

type VerificationPipelineTestSuite struct {
	suite.Suite
	keyPair *encryptioncrypto.KeyPair
}

func TestVerificationPipelineTestSuite(t *testing.T) {
	suite.Run(t, new(VerificationPipelineTestSuite))
}

func (s *VerificationPipelineTestSuite) SetupTest() {
	// This will be called before each test
	// In a real test, we would set up proper mocks
}

// ============================================================================
// Verification Request Tests
// ============================================================================

func TestVerificationRequest_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		request *types.VerificationRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &types.VerificationRequest{
				RequestID:      "req123",
				AccountAddress: "cosmos1abc...",
				ScopeIDs:       []string{"scope1", "scope2"},
				RequestedAt:    now,
				RequestedBlock: 100,
				Status:         types.RequestStatusPending,
			},
			wantErr: false,
		},
		{
			name: "empty request ID",
			request: &types.VerificationRequest{
				RequestID:      "",
				AccountAddress: "cosmos1abc...",
				ScopeIDs:       []string{"scope1"},
				RequestedAt:    now,
				RequestedBlock: 100,
				Status:         types.RequestStatusPending,
			},
			wantErr: true,
		},
		{
			name: "empty account address",
			request: &types.VerificationRequest{
				RequestID:      "req123",
				AccountAddress: "",
				ScopeIDs:       []string{"scope1"},
				RequestedAt:    now,
				RequestedBlock: 100,
				Status:         types.RequestStatusPending,
			},
			wantErr: true,
		},
		{
			name: "empty scope IDs",
			request: &types.VerificationRequest{
				RequestID:      "req123",
				AccountAddress: "cosmos1abc...",
				ScopeIDs:       []string{},
				RequestedAt:    now,
				RequestedBlock: 100,
				Status:         types.RequestStatusPending,
			},
			wantErr: true,
		},
		{
			name: "zero requested at",
			request: &types.VerificationRequest{
				RequestID:      "req123",
				AccountAddress: "cosmos1abc...",
				ScopeIDs:       []string{"scope1"},
				RequestedAt:    time.Time{},
				RequestedBlock: 100,
				Status:         types.RequestStatusPending,
			},
			wantErr: true,
		},
		{
			name: "negative block height",
			request: &types.VerificationRequest{
				RequestID:      "req123",
				AccountAddress: "cosmos1abc...",
				ScopeIDs:       []string{"scope1"},
				RequestedAt:    now,
				RequestedBlock: -1,
				Status:         types.RequestStatusPending,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			request: &types.VerificationRequest{
				RequestID:      "req123",
				AccountAddress: "cosmos1abc...",
				ScopeIDs:       []string{"scope1"},
				RequestedAt:    now,
				RequestedBlock: 100,
				Status:         "invalid_status",
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

func TestVerificationRequest_IsRetryable(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		request    *types.VerificationRequest
		maxRetries uint32
		want       bool
	}{
		{
			name: "pending request is retryable",
			request: &types.VerificationRequest{
				Status:     types.RequestStatusPending,
				RetryCount: 0,
			},
			maxRetries: 3,
			want:       true,
		},
		{
			name: "in progress request is retryable",
			request: &types.VerificationRequest{
				Status:     types.RequestStatusInProgress,
				RetryCount: 1,
			},
			maxRetries: 3,
			want:       true,
		},
		{
			name: "completed request is not retryable",
			request: &types.VerificationRequest{
				Status:     types.RequestStatusCompleted,
				RetryCount: 0,
			},
			maxRetries: 3,
			want:       false,
		},
		{
			name: "failed request is not retryable",
			request: &types.VerificationRequest{
				Status:     types.RequestStatusFailed,
				RetryCount: 0,
			},
			maxRetries: 3,
			want:       false,
		},
		{
			name: "max retries exceeded",
			request: &types.VerificationRequest{
				Status:     types.RequestStatusTimeout,
				RetryCount: 3,
			},
			maxRetries: 3,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.request.RequestID = "test"
			tt.request.AccountAddress = "addr"
			tt.request.ScopeIDs = []string{"s1"}
			tt.request.RequestedAt = now
			got := tt.request.IsRetryable(tt.maxRetries)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestVerificationRequest_StatusTransitions(t *testing.T) {
	now := time.Now()
	request := types.NewVerificationRequest("req1", "addr1", []string{"scope1"}, now, 100)

	// Initial status should be pending
	require.Equal(t, types.RequestStatusPending, request.Status)

	// Set in progress
	request.SetInProgress(now)
	require.Equal(t, types.RequestStatusInProgress, request.Status)
	require.NotNil(t, request.LastAttemptAt)

	// Create new request and complete it
	request2 := types.NewVerificationRequest("req2", "addr1", []string{"scope1"}, now, 100)
	request2.SetCompleted()
	require.Equal(t, types.RequestStatusCompleted, request2.Status)

	// Create new request and fail it
	request3 := types.NewVerificationRequest("req3", "addr1", []string{"scope1"}, now, 100)
	request3.SetFailed("test failure")
	require.Equal(t, types.RequestStatusFailed, request3.Status)
	require.Equal(t, "test failure", request3.Metadata["failure_reason"])

	// Create new request and reject it
	request4 := types.NewVerificationRequest("req4", "addr1", []string{"scope1"}, now, 100)
	request4.SetRejected("invalid document")
	require.Equal(t, types.RequestStatusRejected, request4.Status)
	require.Equal(t, "invalid document", request4.Metadata["rejection_reason"])
}

// ============================================================================
// Verification Result Tests
// ============================================================================

func TestVerificationResult_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		result  *types.VerificationResult
		wantErr bool
	}{
		{
			name: "valid result",
			result: &types.VerificationResult{
				RequestID:      "req123",
				AccountAddress: "cosmos1abc...",
				Score:          75,
				Status:         types.VerificationResultStatusSuccess,
				ComputedAt:     now,
				BlockHeight:    100,
			},
			wantErr: false,
		},
		{
			name: "score exceeds max",
			result: &types.VerificationResult{
				RequestID:      "req123",
				AccountAddress: "cosmos1abc...",
				Score:          150,
				Status:         types.VerificationResultStatusSuccess,
				ComputedAt:     now,
				BlockHeight:    100,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			result: &types.VerificationResult{
				RequestID:      "req123",
				AccountAddress: "cosmos1abc...",
				Score:          75,
				Status:         "invalid",
				ComputedAt:     now,
				BlockHeight:    100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVerificationResult_ComputeOverallScore(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		scopeResults []types.ScopeVerificationResult
		wantScore    uint32
	}{
		{
			name:         "no scope results",
			scopeResults: []types.ScopeVerificationResult{},
			wantScore:    0,
		},
		{
			name: "all successful scopes",
			scopeResults: []types.ScopeVerificationResult{
				{ScopeID: "s1", ScopeType: types.ScopeTypeIDDocument, Success: true, Score: 80, Weight: 30},
				{ScopeID: "s2", ScopeType: types.ScopeTypeSelfie, Success: true, Score: 90, Weight: 20},
			},
			wantScore: 84, // (80*30 + 90*20) / 50 = 4200/50 = 84
		},
		{
			name: "mixed success/failure",
			scopeResults: []types.ScopeVerificationResult{
				{ScopeID: "s1", ScopeType: types.ScopeTypeIDDocument, Success: true, Score: 80, Weight: 30},
				{ScopeID: "s2", ScopeType: types.ScopeTypeSelfie, Success: false, Score: 0, Weight: 20},
			},
			wantScore: 80, // Only successful scopes count: 80*30 / 30 = 80
		},
		{
			name: "all failed",
			scopeResults: []types.ScopeVerificationResult{
				{ScopeID: "s1", Success: false, Score: 0, Weight: 30},
				{ScopeID: "s2", Success: false, Score: 0, Weight: 20},
			},
			wantScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.NewVerificationResult("req1", "addr1", now, 100)
			for _, sr := range tt.scopeResults {
				result.AddScopeResult(sr)
			}
			got := result.ComputeOverallScore()
			require.Equal(t, tt.wantScore, got)
		})
	}
}

func TestVerificationResult_DetermineStatus(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		scopeResults []types.ScopeVerificationResult
		wantStatus   types.VerificationResultStatus
	}{
		{
			name:         "no scopes - failed",
			scopeResults: []types.ScopeVerificationResult{},
			wantStatus:   types.VerificationResultStatusFailed,
		},
		{
			name: "all success",
			scopeResults: []types.ScopeVerificationResult{
				{ScopeID: "s1", Success: true},
				{ScopeID: "s2", Success: true},
			},
			wantStatus: types.VerificationResultStatusSuccess,
		},
		{
			name: "partial success",
			scopeResults: []types.ScopeVerificationResult{
				{ScopeID: "s1", Success: true},
				{ScopeID: "s2", Success: false, ReasonCodes: []types.ReasonCode{types.ReasonCodeDecryptError}},
			},
			wantStatus: types.VerificationResultStatusPartial,
		},
		{
			name: "all failed",
			scopeResults: []types.ScopeVerificationResult{
				{ScopeID: "s1", Success: false, ReasonCodes: []types.ReasonCode{types.ReasonCodeDecryptError}},
				{ScopeID: "s2", Success: false, ReasonCodes: []types.ReasonCode{types.ReasonCodeInvalidScope}},
			},
			wantStatus: types.VerificationResultStatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.NewVerificationResult("req1", "addr1", now, 100)
			for _, sr := range tt.scopeResults {
				result.AddScopeResult(sr)
			}
			result.DetermineStatus()
			require.Equal(t, tt.wantStatus, result.Status)
		})
	}
}

// ============================================================================
// Decryption Tests
// ============================================================================

func TestDecryptedScope_ContentHash(t *testing.T) {
	plaintext := []byte("test identity data")
	scope := NewDecryptedScope("scope1", types.ScopeTypeIDDocument, plaintext)

	require.Equal(t, "scope1", scope.ScopeID)
	require.Equal(t, types.ScopeTypeIDDocument, scope.ScopeType)
	require.Equal(t, plaintext, scope.Plaintext)
	require.NotNil(t, scope.ContentHash)
	require.Len(t, scope.ContentHash, 32) // SHA256 produces 32 bytes
}

func TestInMemoryKeyProvider(t *testing.T) {
	// Generate a key pair
	keyPair, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)

	provider := NewInMemoryKeyProvider(keyPair)

	// Get private key
	privateKey, err := provider.GetPrivateKey()
	require.NoError(t, err)
	require.Len(t, privateKey, 32)

	// Get fingerprint
	fingerprint := provider.GetKeyFingerprint()
	require.NotEmpty(t, fingerprint)

	// Close should not error
	err = provider.Close()
	require.NoError(t, err)
}

func TestInMemoryKeyProvider_NilKeyPair(t *testing.T) {
	provider := NewInMemoryKeyProvider(nil)

	_, err := provider.GetPrivateKey()
	require.Error(t, err)

	fingerprint := provider.GetKeyFingerprint()
	require.Empty(t, fingerprint)
}

func TestValidateImageHeader(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		valid bool
	}{
		{
			name:  "JPEG header",
			data:  []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46},
			valid: true,
		},
		{
			name:  "PNG header",
			data:  []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			valid: true,
		},
		{
			name:  "WebP header",
			data:  []byte{'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P'},
			valid: true,
		},
		{
			name:  "invalid header",
			data:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			valid: false,
		},
		{
			name:  "too short",
			data:  []byte{0xFF, 0xD8},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasValidImageHeader(tt.data)
			require.Equal(t, tt.valid, got)
		})
	}
}

func TestValidateVideoHeader(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		valid bool
	}{
		{
			name:  "MP4 ftyp box",
			data:  []byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm'},
			valid: true,
		},
		{
			name:  "WebM EBML header",
			data:  []byte{0x1A, 0x45, 0xDF, 0xA3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			valid: true,
		},
		{
			name:  "AVI RIFF header",
			data:  []byte{'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00, 'A', 'V', 'I', ' '},
			valid: true,
		},
		{
			name:  "invalid header",
			data:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasValidVideoHeader(tt.data)
			require.Equal(t, tt.valid, got)
		})
	}
}

// ============================================================================
// ML Scoring Tests
// ============================================================================

func TestStubMLScorer_Score(t *testing.T) {
	config := DefaultMLScoringConfig()
	scorer := NewStubMLScorer(config)

	// Create test input with ID document and selfie
	input := &ScoringInput{
		AccountAddress: "cosmos1abc...",
		DecryptedScopes: []DecryptedScope{
			{
				ScopeID:     "scope1",
				ScopeType:   types.ScopeTypeIDDocument,
				Plaintext:   []byte("id document data"),
				ContentHash: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			},
			{
				ScopeID:     "scope2",
				ScopeType:   types.ScopeTypeSelfie,
				Plaintext:   []byte("selfie data"),
				ContentHash: []byte{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			},
		},
		RequestTime: time.Now(),
		BlockHeight: 100,
	}

	output, err := scorer.Score(input)
	require.NoError(t, err)
	require.NotNil(t, output)

	// Score should be in valid range
	require.LessOrEqual(t, output.Score, types.MaxScore)

	// Model version should match config
	require.Equal(t, config.ModelVersion, output.ModelVersion)

	// Should have individual scope scores
	require.Len(t, output.ScopeScores, 2)

	// Confidence should be reasonable
	require.Greater(t, output.Confidence, 0.0)
	require.LessOrEqual(t, output.Confidence, 1.0)

	// Input hash should be set
	require.NotNil(t, output.InputHash)
	require.Len(t, output.InputHash, 32)
}

func TestStubMLScorer_InsufficientScopes(t *testing.T) {
	config := DefaultMLScoringConfig()
	config.MinScopesForScoring = 2
	scorer := NewStubMLScorer(config)

	// Create input with only one scope
	input := &ScoringInput{
		AccountAddress: "cosmos1abc...",
		DecryptedScopes: []DecryptedScope{
			{ScopeID: "scope1", ScopeType: types.ScopeTypeIDDocument, ContentHash: make([]byte, 32)},
		},
		RequestTime: time.Now(),
		BlockHeight: 100,
	}

	output, err := scorer.Score(input)
	require.NoError(t, err)
	require.NotNil(t, output)

	// Score should be 0 due to insufficient scopes
	require.Equal(t, uint32(0), output.Score)

	// Should have insufficient scopes reason code
	require.Contains(t, output.ReasonCodes, types.ReasonCodeInsufficientScopes)
}

func TestStubMLScorer_DeterministicScoring(t *testing.T) {
	config := DefaultMLScoringConfig()
	scorer := NewStubMLScorer(config)

	// Same input should produce same output (deterministic)
	contentHash := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	input := &ScoringInput{
		AccountAddress: "cosmos1abc...",
		DecryptedScopes: []DecryptedScope{
			{ScopeID: "scope1", ScopeType: types.ScopeTypeIDDocument, ContentHash: contentHash},
			{ScopeID: "scope2", ScopeType: types.ScopeTypeSelfie, ContentHash: contentHash},
		},
		RequestTime: time.Now(),
		BlockHeight: 100,
	}

	output1, err := scorer.Score(input)
	require.NoError(t, err)

	output2, err := scorer.Score(input)
	require.NoError(t, err)

	// Scores should be identical
	require.Equal(t, output1.Score, output2.Score)
	require.Equal(t, output1.ScopeScores, output2.ScopeScores)
}

func TestStubMLScorer_Health(t *testing.T) {
	config := DefaultMLScoringConfig()
	scorer := NewStubMLScorer(config)

	require.True(t, scorer.IsHealthy())
	require.Equal(t, config.ModelVersion, scorer.GetModelVersion())
	require.NoError(t, scorer.Close())
}

// ============================================================================
// Scoring Input Hash Tests
// ============================================================================

func TestScoringInput_ComputeInputHash(t *testing.T) {
	// Same inputs should produce same hash
	contentHash := make([]byte, 32)
	for i := range contentHash {
		contentHash[i] = byte(i)
	}

	input1 := &ScoringInput{
		AccountAddress: "cosmos1abc...",
		DecryptedScopes: []DecryptedScope{
			{ScopeID: "scope1", ContentHash: contentHash},
		},
		BlockHeight: 100,
	}

	input2 := &ScoringInput{
		AccountAddress: "cosmos1abc...",
		DecryptedScopes: []DecryptedScope{
			{ScopeID: "scope1", ContentHash: contentHash},
		},
		BlockHeight: 100,
	}

	hash1 := input1.ComputeInputHash()
	hash2 := input2.ComputeInputHash()

	require.Equal(t, hash1, hash2)

	// Different inputs should produce different hash
	input3 := &ScoringInput{
		AccountAddress: "cosmos1xyz...",
		DecryptedScopes: []DecryptedScope{
			{ScopeID: "scope1", ContentHash: contentHash},
		},
		BlockHeight: 100,
	}

	hash3 := input3.ComputeInputHash()
	require.NotEqual(t, hash1, hash3)
}

// ============================================================================
// Scope Verification Result Tests
// ============================================================================

func TestScopeVerificationResult_WeightedScore(t *testing.T) {
	tests := []struct {
		name   string
		result *types.ScopeVerificationResult
		want   uint32
	}{
		{
			name: "successful with score",
			result: &types.ScopeVerificationResult{
				Success: true,
				Score:   80,
				Weight:  30,
			},
			want: 24, // (80 * 30) / 100 = 24
		},
		{
			name: "failed scope",
			result: &types.ScopeVerificationResult{
				Success: false,
				Score:   0,
				Weight:  30,
			},
			want: 0,
		},
		{
			name: "high score high weight",
			result: &types.ScopeVerificationResult{
				Success: true,
				Score:   100,
				Weight:  100,
			},
			want: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.WeightedScore()
			require.Equal(t, tt.want, got)
		})
	}
}

// ============================================================================
// Pipeline Configuration Tests
// ============================================================================

func TestDefaultVerificationPipelineConfig(t *testing.T) {
	config := DefaultVerificationPipelineConfig()

	require.Greater(t, config.MaxVerificationTimePerBlock, int64(0))
	require.Greater(t, config.MaxVerificationTimePerRequest, int64(0))
	require.Greater(t, config.MaxRequestsPerBlock, 0)
	require.Greater(t, config.MaxRetries, uint32(0))
	require.Greater(t, config.RetryDelayBlocks, int64(0))
}

func TestDefaultMLScoringConfig(t *testing.T) {
	config := DefaultMLScoringConfig()

	require.NotEmpty(t, config.ModelVersion)
	require.Greater(t, config.MinScopesForScoring, 0)
	require.Greater(t, len(config.RequiredScopeTypes), 0)
	require.Greater(t, config.MaxInferenceTime, int64(0))
}

// ============================================================================
// Reason Code Tests
// ============================================================================

func TestReasonCodes(t *testing.T) {
	// Ensure all reason codes are distinct
	codes := []types.ReasonCode{
		types.ReasonCodeSuccess,
		types.ReasonCodeDecryptError,
		types.ReasonCodeInvalidScope,
		types.ReasonCodeScopeNotFound,
		types.ReasonCodeScopeRevoked,
		types.ReasonCodeScopeExpired,
		types.ReasonCodeMLInferenceError,
		types.ReasonCodeTimeout,
		types.ReasonCodeMaxRetriesExceeded,
		types.ReasonCodeInvalidPayload,
		types.ReasonCodeKeyNotFound,
		types.ReasonCodeInsufficientScopes,
		types.ReasonCodeFaceMismatch,
		types.ReasonCodeDocumentInvalid,
		types.ReasonCodeLivenessCheckFailed,
	}

	seen := make(map[types.ReasonCode]bool)
	for _, code := range codes {
		require.False(t, seen[code], "duplicate reason code: %s", code)
		seen[code] = true
	}
}

// ============================================================================
// Proposer Hook Configuration Tests
// ============================================================================

func TestProposerHookConfig(t *testing.T) {
	// Test default config
	defaultConfig := DefaultProposerHookConfig()
	require.True(t, defaultConfig.Enabled)
	require.Nil(t, defaultConfig.KeyProviderFactory)

	// Test setting config
	customConfig := ProposerHookConfig{
		Enabled: false,
		KeyProviderFactory: func() (ValidatorKeyProvider, error) {
			keyPair, _ := encryptioncrypto.GenerateKeyPair()
			return NewInMemoryKeyProvider(keyPair), nil
		},
	}
	SetProposerHookConfig(customConfig)

	retrievedConfig := GetProposerHookConfig()
	require.False(t, retrievedConfig.Enabled)
	require.NotNil(t, retrievedConfig.KeyProviderFactory)

	// Reset to default
	SetProposerHookConfig(DefaultProposerHookConfig())
}
