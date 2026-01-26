package types

// GenesisState is the genesis state for the veid module
type GenesisState struct {
	// IdentityRecords are the initial identity records
	IdentityRecords []IdentityRecord `json:"identity_records"`

	// Scopes are the initial identity scopes
	Scopes []IdentityScope `json:"scopes"`

	// ApprovedClients are the initially approved clients
	ApprovedClients []ApprovedClient `json:"approved_clients"`

	// Params are the module parameters
	Params Params `json:"params"`
}

// Params defines the parameters for the veid module
type Params struct {
	// MaxScopesPerAccount is the maximum number of scopes per account
	MaxScopesPerAccount uint32 `json:"max_scopes_per_account"`

	// MaxScopesPerType is the maximum number of scopes per type per account
	MaxScopesPerType uint32 `json:"max_scopes_per_type"`

	// SaltMinBytes is the minimum salt size in bytes
	SaltMinBytes uint32 `json:"salt_min_bytes"`

	// SaltMaxBytes is the maximum salt size in bytes
	SaltMaxBytes uint32 `json:"salt_max_bytes"`

	// RequireClientSignature determines if client signatures are mandatory
	RequireClientSignature bool `json:"require_client_signature"`

	// RequireUserSignature determines if user signatures are mandatory
	RequireUserSignature bool `json:"require_user_signature"`

	// VerificationExpiryDays is how long a verification is valid (in days)
	VerificationExpiryDays uint32 `json:"verification_expiry_days"`

	// MinScoreForTier contains the minimum scores for each tier
	MinScoreForTier map[string]uint32 `json:"min_score_for_tier"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		IdentityRecords: []IdentityRecord{},
		Scopes:          []IdentityScope{},
		ApprovedClients: []ApprovedClient{},
		Params:          DefaultParams(),
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		MaxScopesPerAccount:    50,
		MaxScopesPerType:       5,
		SaltMinBytes:           16,
		SaltMaxBytes:           64,
		RequireClientSignature: true,
		RequireUserSignature:   true,
		VerificationExpiryDays: 365,
		MinScoreForTier: map[string]uint32{
			string(IdentityTierUnverified): 0,
			string(IdentityTierBasic):      1,
			string(IdentityTierStandard):   30,
			string(IdentityTierVerified):   60,
			string(IdentityTierTrusted):    85,
		},
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	// Validate identity records
	seenRecords := make(map[string]bool)
	for _, record := range gs.IdentityRecords {
		if err := record.Validate(); err != nil {
			return err
		}
		if seenRecords[record.AccountAddress] {
			return ErrInvalidIdentityRecord.Wrapf("duplicate identity record: %s", record.AccountAddress)
		}
		seenRecords[record.AccountAddress] = true
	}

	// Validate scopes
	seenScopes := make(map[string]bool)
	for _, scope := range gs.Scopes {
		if err := scope.Validate(); err != nil {
			return err
		}
		if seenScopes[scope.ScopeID] {
			return ErrInvalidScope.Wrapf("duplicate scope: %s", scope.ScopeID)
		}
		seenScopes[scope.ScopeID] = true
	}

	// Validate approved clients
	seenClients := make(map[string]bool)
	for _, client := range gs.ApprovedClients {
		if err := client.Validate(); err != nil {
			return err
		}
		if seenClients[client.ClientID] {
			return ErrInvalidClientID.Wrapf("duplicate client: %s", client.ClientID)
		}
		seenClients[client.ClientID] = true
	}

	// Validate params
	return gs.Params.Validate()
}

// Validate validates the params
func (p Params) Validate() error {
	if p.MaxScopesPerAccount == 0 {
		return ErrInvalidParams.Wrap("max_scopes_per_account must be greater than 0")
	}

	if p.MaxScopesPerType == 0 {
		return ErrInvalidParams.Wrap("max_scopes_per_type must be greater than 0")
	}

	if p.SaltMinBytes == 0 {
		return ErrInvalidParams.Wrap("salt_min_bytes must be greater than 0")
	}

	if p.SaltMaxBytes < p.SaltMinBytes {
		return ErrInvalidParams.Wrap("salt_max_bytes must be >= salt_min_bytes")
	}

	if p.VerificationExpiryDays == 0 {
		return ErrInvalidParams.Wrap("verification_expiry_days must be greater than 0")
	}

	return nil
}

// GetMinScoreForTier returns the minimum score for a tier
func (p Params) GetMinScoreForTier(tier IdentityTier) uint32 {
	if score, ok := p.MinScoreForTier[string(tier)]; ok {
		return score
	}
	return TierMinimumScore(tier)
}
