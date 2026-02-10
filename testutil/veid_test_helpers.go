//go:build testing || e2e.integration

// Package testutil provides VEID testing utilities for setting up test identities.
//
// This file implements test helpers for:
// - Creating verified identity records with specific scores
// - Setting up approved test capture clients
// - Bypassing validator authorization for score updates
// - MFA test mode configuration
//
// Task Reference: VEID Testing Infrastructure Setup
package testutil

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/app"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Constants
// ============================================================================

const (
	// TestModelVersion is the ML model version for test scores
	TestModelVersion = "veid-test-v1.0.0"

	// DefaultProviderScore is the default score for test providers (meets 70 requirement)
	DefaultProviderScore = uint32(80)

	// DefaultCustomerScore is the default score for test customers
	DefaultCustomerScore = uint32(60)

	// DefaultValidatorScore is the default score for test validators (meets 85 requirement)
	DefaultValidatorScore = uint32(90)
)

// ============================================================================
// Identity Record Setup
// ============================================================================

// VEIDTestIdentity represents a test identity configuration
type VEIDTestIdentity struct {
	Address   sdk.AccAddress
	Score     uint32
	Tier      veidtypes.IdentityTier
	Status    veidtypes.AccountStatus
	Scopes    []veidtypes.ScopeType
	HasMFA    bool
	MFAFactor string // "fido2", "totp", etc.
}

// DefaultProviderIdentity returns a default identity configuration for a provider
func DefaultProviderIdentity(address sdk.AccAddress) VEIDTestIdentity {
	return VEIDTestIdentity{
		Address:   address,
		Score:     DefaultProviderScore,
		Tier:      veidtypes.IdentityTierVerified,
		Status:    veidtypes.AccountStatusVerified,
		Scopes:    []veidtypes.ScopeType{veidtypes.ScopeTypeSelfie, veidtypes.ScopeTypeIDDocument},
		HasMFA:    true,
		MFAFactor: "fido2",
	}
}

// DefaultCustomerIdentity returns a default identity configuration for a customer
func DefaultCustomerIdentity(address sdk.AccAddress) VEIDTestIdentity {
	return VEIDTestIdentity{
		Address: address,
		Score:   DefaultCustomerScore,
		Tier:    veidtypes.IdentityTierVerified,
		Status:  veidtypes.AccountStatusVerified,
		Scopes:  []veidtypes.ScopeType{veidtypes.ScopeTypeSelfie},
		HasMFA:  false,
	}
}

// DefaultValidatorIdentity returns a default identity configuration for a validator
func DefaultValidatorIdentity(address sdk.AccAddress) VEIDTestIdentity {
	return VEIDTestIdentity{
		Address:   address,
		Score:     DefaultValidatorScore,
		Tier:      veidtypes.IdentityTierTrusted,
		Status:    veidtypes.AccountStatusVerified,
		Scopes:    []veidtypes.ScopeType{veidtypes.ScopeTypeSelfie, veidtypes.ScopeTypeIDDocument, veidtypes.ScopeTypeFaceVideo},
		HasMFA:    true,
		MFAFactor: "fido2",
	}
}

// ============================================================================
// Setup Functions
// ============================================================================

// SetupVEIDIdentity sets up a VEID identity record for testing.
// This bypasses the validator authorization check by directly calling the keeper.
func SetupVEIDIdentity(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	identity VEIDTestIdentity,
) {
	t.Helper()

	// Create identity record if it doesn't exist
	_, found := a.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, identity.Address)
	if !found {
		_, err := a.Keepers.VirtEngine.VEID.CreateIdentityRecord(ctx, identity.Address)
		require.NoError(t, err, "failed to create identity record")
	}

	// Set the score directly (bypasses validator check)
	err := a.Keepers.VirtEngine.VEID.SetScore(
		ctx,
		identity.Address.String(),
		identity.Score,
		TestModelVersion,
	)
	require.NoError(t, err, "failed to set identity score")

	// Verify the score was set
	score, found := a.Keepers.VirtEngine.VEID.GetVEIDScore(ctx, identity.Address)
	require.True(t, found, "identity score should be found after setting")
	require.Equal(t, identity.Score, score, "score should match")
}

// SetupVEIDIdentityWithScore is a convenience function to set up an identity with just a score.
func SetupVEIDIdentityWithScore(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	address sdk.AccAddress,
	score uint32,
) {
	t.Helper()

	identity := VEIDTestIdentity{
		Address: address,
		Score:   score,
		Tier:    getTierForScore(score),
		Status:  veidtypes.AccountStatusVerified,
	}

	SetupVEIDIdentity(t, a, ctx, identity)
}

// SetupProviderVEID sets up a VEID identity suitable for provider registration (score >= 70).
func SetupProviderVEID(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	providerAddress sdk.AccAddress,
) {
	t.Helper()
	SetupVEIDIdentity(t, a, ctx, DefaultProviderIdentity(providerAddress))
}

// SetupCustomerVEID sets up a VEID identity suitable for customer operations.
func SetupCustomerVEID(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	customerAddress sdk.AccAddress,
) {
	t.Helper()
	SetupVEIDIdentity(t, a, ctx, DefaultCustomerIdentity(customerAddress))
}

// SetupValidatorVEID sets up a VEID identity suitable for validator operations (score >= 85).
func SetupValidatorVEID(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	validatorAddress sdk.AccAddress,
) {
	t.Helper()
	SetupVEIDIdentity(t, a, ctx, DefaultValidatorIdentity(validatorAddress))
}

// ============================================================================
// Verification Functions
// ============================================================================

// RequireVEIDScore asserts that an account has the expected VEID score.
func RequireVEIDScore(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	address sdk.AccAddress,
	expectedScore uint32,
) {
	t.Helper()

	score, found := a.Keepers.VirtEngine.VEID.GetVEIDScore(ctx, address)
	require.True(t, found, "VEID score should be found for address %s", address)
	require.Equal(t, expectedScore, score, "VEID score mismatch")
}

// RequireVEIDTier asserts that an account has the expected VEID tier.
func RequireVEIDTier(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	address sdk.AccAddress,
	expectedTier veidtypes.IdentityTier,
) {
	t.Helper()

	record, found := a.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, address)
	require.True(t, found, "identity record should be found for address %s", address)
	require.Equal(t, expectedTier, record.Tier, "VEID tier mismatch")
}

// RequireProviderEligible asserts that an account meets provider VEID requirements.
func RequireProviderEligible(
	t *testing.T,
	a *app.VirtEngineApp,
	ctx sdk.Context,
	address sdk.AccAddress,
) {
	t.Helper()

	score, found := a.Keepers.VirtEngine.VEID.GetVEIDScore(ctx, address)
	require.True(t, found, "VEID score should be found for address %s", address)
	require.GreaterOrEqual(t, score, uint32(70), "provider requires VEID score >= 70, got %d", score)
}

// ============================================================================
// Helper Functions
// ============================================================================

// getTierForScore returns the appropriate tier for a given score.
func getTierForScore(score uint32) veidtypes.IdentityTier {
	switch {
	case score >= 85:
		return veidtypes.IdentityTierTrusted
	case score >= 60:
		return veidtypes.IdentityTierVerified
	case score >= 30:
		return veidtypes.IdentityTierStandard
	case score >= 1:
		return veidtypes.IdentityTierBasic
	default:
		return veidtypes.IdentityTierUnverified
	}
}

// ============================================================================
// Mock Staking Keeper for Validator Authorization
// ============================================================================

// MockStakingKeeper is a mock implementation of the staking keeper interface
// that can be used to make addresses appear as bonded validators.
type MockStakingKeeper struct {
	validators map[string]bool
}

// NewMockStakingKeeper creates a new mock staking keeper.
func NewMockStakingKeeper() *MockStakingKeeper {
	return &MockStakingKeeper{
		validators: make(map[string]bool),
	}
}

// AddValidator marks an address as a bonded validator.
func (m *MockStakingKeeper) AddValidator(valAddr sdk.ValAddress) {
	m.validators[valAddr.String()] = true
}

// RemoveValidator removes an address from the validator set.
func (m *MockStakingKeeper) RemoveValidator(valAddr sdk.ValAddress) {
	delete(m.validators, valAddr.String())
}

// IsValidator checks if an address is a bonded validator.
func (m *MockStakingKeeper) IsValidator(addr sdk.ValAddress) bool {
	return m.validators[addr.String()]
}

// ============================================================================
// Approved Client Setup
// ============================================================================

// TestCaptureClient represents a test capture client configuration.
type TestCaptureClient struct {
	ClientID  string
	Name      string
	PublicKey []byte
	Algorithm string
}

// DefaultTestCaptureClient returns a default test capture client.
func DefaultTestCaptureClient() TestCaptureClient {
	return TestCaptureClient{
		ClientID:  "ve-test-capture-app",
		Name:      "VirtEngine Test Capture App",
		PublicKey: make([]byte, 32), // Zero-filled for testing
		Algorithm: "Ed25519",
	}
}

// ToApprovedClient converts to a veidtypes.ApprovedClient.
func (c TestCaptureClient) ToApprovedClient() veidtypes.ApprovedClient {
	return veidtypes.ApprovedClient{
		ClientID:     c.ClientID,
		Name:         c.Name,
		PublicKey:    c.PublicKey,
		Algorithm:    c.Algorithm,
		Active:       true,
		RegisteredAt: time.Now().Unix(),
	}
}

// ============================================================================
// Genesis State Helpers
// ============================================================================

// VEIDGenesisConfig holds configuration for VEID genesis state.
type VEIDGenesisConfig struct {
	Identities      []VEIDTestIdentity
	ApprovedClients []TestCaptureClient
	DisableGating   bool
}

// DefaultVEIDGenesisConfig returns a default VEID genesis configuration for testing.
func DefaultVEIDGenesisConfig() VEIDGenesisConfig {
	return VEIDGenesisConfig{
		Identities:      []VEIDTestIdentity{},
		ApprovedClients: []TestCaptureClient{DefaultTestCaptureClient()},
		DisableGating:   true,
	}
}

// WithProviderIdentity adds a provider identity to the genesis config.
func (c VEIDGenesisConfig) WithProviderIdentity(address sdk.AccAddress) VEIDGenesisConfig {
	c.Identities = append(c.Identities, DefaultProviderIdentity(address))
	return c
}

// WithCustomerIdentity adds a customer identity to the genesis config.
func (c VEIDGenesisConfig) WithCustomerIdentity(address sdk.AccAddress) VEIDGenesisConfig {
	c.Identities = append(c.Identities, DefaultCustomerIdentity(address))
	return c
}

// WithValidatorIdentity adds a validator identity to the genesis config.
func (c VEIDGenesisConfig) WithValidatorIdentity(address sdk.AccAddress) VEIDGenesisConfig {
	c.Identities = append(c.Identities, DefaultValidatorIdentity(address))
	return c
}

// WithIdentity adds a custom identity to the genesis config.
func (c VEIDGenesisConfig) WithIdentity(identity VEIDTestIdentity) VEIDGenesisConfig {
	c.Identities = append(c.Identities, identity)
	return c
}

// BuildGenesisState builds a VEID genesis state from the configuration.
func (c VEIDGenesisConfig) BuildGenesisState() *veidtypes.GenesisState {
	genesis := veidtypes.DefaultGenesisState()

	// Add identity records
	for _, identity := range c.Identities {
		record := veidtypes.IdentityRecord{
			AccountAddress: identity.Address.String(),
			CurrentScore:   identity.Score,
			Tier:           identity.Tier,
		}
		genesis.IdentityRecords = append(genesis.IdentityRecords, record)
	}

	// Add approved clients
	for _, client := range c.ApprovedClients {
		genesis.ApprovedClients = append(genesis.ApprovedClients, client.ToApprovedClient())
	}

	return genesis
}
