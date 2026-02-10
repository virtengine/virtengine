package types

// ============================================================================
// Wallet Query Types
// ============================================================================

// QueryIdentityWalletRequest is the request for QueryIdentityWallet
type QueryIdentityWalletRequest struct {
	// AccountAddress is the account to query the wallet for
	AccountAddress string `json:"account_address"`
}

// QueryIdentityWalletResponse is the response for QueryIdentityWallet
// Returns only non-sensitive metadata
type QueryIdentityWalletResponse struct {
	// Wallet contains non-sensitive wallet information
	Wallet *PublicWalletInfo `json:"wallet"`

	// Found indicates if the wallet was found
	Found bool `json:"found"`
}

// QueryWalletScopesRequest is the request for QueryWalletScopes
type QueryWalletScopesRequest struct {
	// AccountAddress is the account to query scopes for
	AccountAddress string `json:"account_address"`

	// ScopeType is an optional filter by scope type
	ScopeType string `json:"scope_type,omitempty"`

	// StatusFilter is an optional filter by status
	StatusFilter string `json:"status_filter,omitempty"`

	// ActiveOnly filters to only active scopes
	ActiveOnly bool `json:"active_only,omitempty"`
}

// QueryWalletScopesResponse is the response for QueryWalletScopes
type QueryWalletScopesResponse struct {
	// Scopes contains non-sensitive scope information
	Scopes []WalletScopeInfo `json:"scopes"`

	// TotalCount is the total number of scopes in the wallet
	TotalCount int `json:"total_count"`

	// ActiveCount is the number of active scopes
	ActiveCount int `json:"active_count"`
}

// QueryConsentSettingsRequest is the request for QueryConsentSettings
type QueryConsentSettingsRequest struct {
	// AccountAddress is the account to query consent for
	AccountAddress string `json:"account_address"`

	// ScopeID is an optional filter for specific scope consent
	ScopeID string `json:"scope_id,omitempty"`
}

// PublicConsentInfo represents non-sensitive consent information
type PublicConsentInfo struct {
	// ScopeID is the scope identifier
	ScopeID string `json:"scope_id"`

	// Granted indicates if consent is granted
	Granted bool `json:"granted"`

	// IsActive indicates if consent is currently active
	IsActive bool `json:"is_active"`

	// Purpose is the consent purpose
	Purpose string `json:"purpose,omitempty"`

	// ExpiresAt is when consent expires (if applicable)
	ExpiresAt *int64 `json:"expires_at,omitempty"`
}

// QueryConsentSettingsResponse is the response for QueryConsentSettings
type QueryConsentSettingsResponse struct {
	// GlobalSettings contains global consent settings
	GlobalSettings struct {
		ShareWithProviders         bool `json:"share_with_providers"`
		ShareForVerification       bool `json:"share_for_verification"`
		AllowReVerification        bool `json:"allow_re_verification"`
		AllowDerivedFeatureSharing bool `json:"allow_derived_feature_sharing"`
	} `json:"global_settings"`

	// ScopeConsents contains per-scope consent info
	ScopeConsents []PublicConsentInfo `json:"scope_consents"`

	// ConsentVersion is the current consent version
	ConsentVersion uint32 `json:"consent_version"`

	// LastUpdatedAt is when consent was last updated (Unix timestamp)
	LastUpdatedAt int64 `json:"last_updated_at"`
}

// QueryDerivedFeaturesRequest is the request for QueryDerivedFeatures
type QueryDerivedFeaturesRequest struct {
	// AccountAddress is the account to query features for
	AccountAddress string `json:"account_address"`
}

// QueryDerivedFeaturesResponse is the response for QueryDerivedFeatures
type QueryDerivedFeaturesResponse struct {
	// Features contains non-sensitive derived features information
	Features *PublicDerivedFeaturesInfo `json:"features"`

	// Found indicates if features were found
	Found bool `json:"found"`
}

// QueryDerivedFeatureHashesRequest is the request for QueryDerivedFeatureHashes
// This is used for verification matching by authorized parties
type QueryDerivedFeatureHashesRequest struct {
	// AccountAddress is the account to query hashes for
	AccountAddress string `json:"account_address"`

	// Requester is the address requesting the hashes
	Requester string `json:"requester"`

	// Purpose describes why the hashes are being requested
	Purpose string `json:"purpose"`
}

// QueryDerivedFeatureHashesResponse is the response for QueryDerivedFeatureHashes
// Only returned if consent allows sharing
type QueryDerivedFeatureHashesResponse struct {
	// Allowed indicates if the request was allowed based on consent
	Allowed bool `json:"allowed"`

	// DenialReason is set if Allowed is false
	DenialReason string `json:"denial_reason,omitempty"`

	// FaceEmbeddingHash is the face embedding hash (if consented)
	FaceEmbeddingHash []byte `json:"face_embedding_hash,omitempty"`

	// DocFieldHashes are document field hashes (if consented)
	DocFieldHashes map[string][]byte `json:"doc_field_hashes,omitempty"`

	// BiometricHash is the biometric hash (if consented)
	BiometricHash []byte `json:"biometric_hash,omitempty"`

	// ModelVersion is the model version used
	ModelVersion string `json:"model_version,omitempty"`
}

// QueryVerificationHistoryRequest is the request for QueryVerificationHistory
type QueryVerificationHistoryRequest struct {
	// AccountAddress is the account to query history for
	AccountAddress string `json:"account_address"`

	// Limit is the maximum number of entries to return
	Limit uint32 `json:"limit,omitempty"`

	// Offset is the number of entries to skip
	Offset uint32 `json:"offset,omitempty"`
}

// PublicVerificationHistoryEntry represents non-sensitive verification history
type PublicVerificationHistoryEntry struct {
	// EntryID is the entry identifier
	EntryID string `json:"entry_id"`

	// Timestamp is when this verification occurred (Unix timestamp)
	Timestamp int64 `json:"timestamp"`

	// BlockHeight is the block height
	BlockHeight int64 `json:"block_height"`

	// PreviousScore is the score before verification
	PreviousScore uint32 `json:"previous_score"`

	// NewScore is the score after verification
	NewScore uint32 `json:"new_score"`

	// PreviousStatus is the status before verification
	PreviousStatus string `json:"previous_status"`

	// NewStatus is the status after verification
	NewStatus string `json:"new_status"`

	// ScopeCount is the number of scopes evaluated
	ScopeCount int `json:"scope_count"`

	// ModelVersion is the model version used
	ModelVersion string `json:"model_version"`
}

// QueryVerificationHistoryResponse is the response for QueryVerificationHistory
type QueryVerificationHistoryResponse struct {
	// Entries contains verification history entries
	Entries []PublicVerificationHistoryEntry `json:"entries"`

	// TotalCount is the total number of entries
	TotalCount int `json:"total_count"`
}

// ============================================================================
// proto.Message interface implementations for Wallet Query types
// Required for Cosmos SDK gRPC router registration
// ============================================================================

func (m *QueryIdentityWalletRequest) Reset()         { *m = QueryIdentityWalletRequest{} }
func (m *QueryIdentityWalletRequest) String() string { return "" }
func (*QueryIdentityWalletRequest) ProtoMessage()    {}

func (m *QueryIdentityWalletResponse) Reset()         { *m = QueryIdentityWalletResponse{} }
func (m *QueryIdentityWalletResponse) String() string { return "" }
func (*QueryIdentityWalletResponse) ProtoMessage()    {}

func (m *QueryWalletScopesRequest) Reset()         { *m = QueryWalletScopesRequest{} }
func (m *QueryWalletScopesRequest) String() string { return "" }
func (*QueryWalletScopesRequest) ProtoMessage()    {}

func (m *QueryWalletScopesResponse) Reset()         { *m = QueryWalletScopesResponse{} }
func (m *QueryWalletScopesResponse) String() string { return "" }
func (*QueryWalletScopesResponse) ProtoMessage()    {}

func (m *QueryConsentSettingsRequest) Reset()         { *m = QueryConsentSettingsRequest{} }
func (m *QueryConsentSettingsRequest) String() string { return "" }
func (*QueryConsentSettingsRequest) ProtoMessage()    {}

func (m *QueryConsentSettingsResponse) Reset()         { *m = QueryConsentSettingsResponse{} }
func (m *QueryConsentSettingsResponse) String() string { return "" }
func (*QueryConsentSettingsResponse) ProtoMessage()    {}

func (m *QueryDerivedFeaturesRequest) Reset()         { *m = QueryDerivedFeaturesRequest{} }
func (m *QueryDerivedFeaturesRequest) String() string { return "" }
func (*QueryDerivedFeaturesRequest) ProtoMessage()    {}

func (m *QueryDerivedFeaturesResponse) Reset()         { *m = QueryDerivedFeaturesResponse{} }
func (m *QueryDerivedFeaturesResponse) String() string { return "" }
func (*QueryDerivedFeaturesResponse) ProtoMessage()    {}

func (m *QueryDerivedFeatureHashesRequest) Reset()         { *m = QueryDerivedFeatureHashesRequest{} }
func (m *QueryDerivedFeatureHashesRequest) String() string { return "" }
func (*QueryDerivedFeatureHashesRequest) ProtoMessage()    {}

func (m *QueryDerivedFeatureHashesResponse) Reset()         { *m = QueryDerivedFeatureHashesResponse{} }
func (m *QueryDerivedFeatureHashesResponse) String() string { return "" }
func (*QueryDerivedFeatureHashesResponse) ProtoMessage()    {}

func (m *QueryVerificationHistoryRequest) Reset()         { *m = QueryVerificationHistoryRequest{} }
func (m *QueryVerificationHistoryRequest) String() string { return "" }
func (*QueryVerificationHistoryRequest) ProtoMessage()    {}

func (m *QueryVerificationHistoryResponse) Reset()         { *m = QueryVerificationHistoryResponse{} }
func (m *QueryVerificationHistoryResponse) String() string { return "" }
func (*QueryVerificationHistoryResponse) ProtoMessage()    {}
