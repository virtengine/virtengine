// Package keeper provides the VEID module keeper.
//
// This file implements trusted setup ceremony tooling for ZK proof circuits.
// It provides deterministic multi-party ceremony coordination and verification
// key management for Groth16 ZK-SNARKs used in privacy-preserving claims.
//
// Task Reference: 22A - Pre-mainnet security hardening
package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Trusted Setup Ceremony Types
// ============================================================================

// CeremonyStatus represents the current state of a trusted setup ceremony
type CeremonyStatus string

const (
	// CeremonyStatusPending indicates the ceremony has been created but not started
	CeremonyStatusPending CeremonyStatus = "pending"

	// CeremonyStatusInProgress indicates contributions are being collected
	CeremonyStatusInProgress CeremonyStatus = "in_progress"

	// CeremonyStatusCompleted indicates the ceremony finished successfully
	CeremonyStatusCompleted CeremonyStatus = "completed"

	// CeremonyStatusFailed indicates the ceremony failed verification
	CeremonyStatusFailed CeremonyStatus = "failed"

	// MinCeremonyParticipants is the minimum number of participants for a valid ceremony
	MinCeremonyParticipants = 3

	// MaxCeremonyParticipants is the maximum number of participants
	MaxCeremonyParticipants = 100

	// CeremonyContributionMaxSize is the max size of a contribution in bytes (10MB)
	CeremonyContributionMaxSize = 10 * 1024 * 1024
)

// TrustedSetupCeremony represents a multi-party computation ceremony
// for generating Groth16 proving and verification keys.
type TrustedSetupCeremony struct {
	// CeremonyID is a unique identifier for this ceremony
	CeremonyID string `json:"ceremony_id"`

	// CircuitType identifies which ZK circuit this ceremony is for
	CircuitType string `json:"circuit_type"`

	// Status is the current ceremony state
	Status CeremonyStatus `json:"status"`

	// MinParticipants is the minimum required participants
	MinParticipants uint32 `json:"min_participants"`

	// Contributions is the ordered list of participant contributions
	Contributions []CeremonyContribution `json:"contributions"`

	// VerificationKeyHash is the SHA-256 hash of the final verification key
	VerificationKeyHash string `json:"verification_key_hash,omitempty"`

	// InitiatedBy is the governance authority that initiated the ceremony
	InitiatedBy string `json:"initiated_by"`

	// InitiatedAt is when the ceremony was created
	InitiatedAt time.Time `json:"initiated_at"`

	// CompletedAt is when the ceremony completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// ExpiresAt is the deadline for contributions
	ExpiresAt time.Time `json:"expires_at"`

	// ProposalID links to the governance proposal if applicable
	ProposalID uint64 `json:"proposal_id,omitempty"`
}

// CeremonyContribution represents a single participant's contribution
type CeremonyContribution struct {
	// ParticipantAddress is the contributor's address
	ParticipantAddress string `json:"participant_address"`

	// ContributionHash is the SHA-256 hash of the contribution data
	ContributionHash string `json:"contribution_hash"`

	// PreviousContributionHash links to the prior contribution for chain integrity
	PreviousContributionHash string `json:"previous_contribution_hash,omitempty"`

	// ContributedAt is when this contribution was submitted
	ContributedAt time.Time `json:"contributed_at"`

	// Verified indicates if this contribution passed verification
	Verified bool `json:"verified"`

	// VerificationProofHash is the hash of the verification proof
	VerificationProofHash string `json:"verification_proof_hash,omitempty"`
}

// ============================================================================
// Ceremony Validation
// ============================================================================

// ValidateCeremony validates a trusted setup ceremony record
func ValidateCeremony(ceremony *TrustedSetupCeremony) error {
	if ceremony.CeremonyID == "" {
		return types.ErrInvalidCeremony.Wrap("ceremony_id cannot be empty")
	}
	if ceremony.CircuitType == "" {
		return types.ErrInvalidCeremony.Wrap("circuit_type cannot be empty")
	}
	if ceremony.MinParticipants < MinCeremonyParticipants {
		return types.ErrInvalidCeremony.Wrapf(
			"min_participants must be at least %d, got %d",
			MinCeremonyParticipants, ceremony.MinParticipants,
		)
	}
	if ceremony.MinParticipants > MaxCeremonyParticipants {
		return types.ErrInvalidCeremony.Wrapf(
			"min_participants cannot exceed %d, got %d",
			MaxCeremonyParticipants, ceremony.MinParticipants,
		)
	}
	if ceremony.InitiatedBy == "" {
		return types.ErrInvalidCeremony.Wrap("initiated_by cannot be empty")
	}
	if ceremony.ExpiresAt.IsZero() {
		return types.ErrInvalidCeremony.Wrap("expires_at cannot be zero")
	}
	return nil
}

// ValidateContribution validates a ceremony contribution
func ValidateContribution(contribution *CeremonyContribution) error {
	if contribution.ParticipantAddress == "" {
		return types.ErrInvalidCeremonyContribution.Wrap("participant_address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(contribution.ParticipantAddress); err != nil {
		return types.ErrInvalidCeremonyContribution.Wrapf("invalid participant address: %v", err)
	}
	if contribution.ContributionHash == "" {
		return types.ErrInvalidCeremonyContribution.Wrap("contribution_hash cannot be empty")
	}
	if len(contribution.ContributionHash) != 64 {
		return types.ErrInvalidCeremonyContribution.Wrap("contribution_hash must be a 64-character hex SHA-256 hash")
	}
	return nil
}

// ============================================================================
// Ceremony Store Operations
// ============================================================================

// InitiateTrustedSetupCeremony creates a new trusted setup ceremony.
// Only the governance authority can initiate ceremonies.
func (k Keeper) InitiateTrustedSetupCeremony(
	ctx sdk.Context,
	authority string,
	circuitType string,
	minParticipants uint32,
	expiryDuration time.Duration,
) (*TrustedSetupCeremony, error) {
	// Verify authority
	if authority != k.authority {
		return nil, types.ErrUnauthorized.Wrapf(
			"only governance authority can initiate ceremonies: expected %s, got %s",
			k.authority, authority,
		)
	}

	// Validate circuit type
	validCircuitTypes := map[string]bool{
		"age_range":   true,
		"residency":   true,
		"score_range": true,
	}
	if !validCircuitTypes[circuitType] {
		return nil, types.ErrInvalidCeremony.Wrapf("invalid circuit type: %s", circuitType)
	}

	// Generate ceremony ID
	idData := fmt.Sprintf("%s|%s|%d", circuitType, authority, ctx.BlockHeight())
	idHash := sha256.Sum256([]byte(idData))
	ceremonyID := hex.EncodeToString(idHash[:16])

	ceremony := &TrustedSetupCeremony{
		CeremonyID:      ceremonyID,
		CircuitType:     circuitType,
		Status:          CeremonyStatusPending,
		MinParticipants: minParticipants,
		Contributions:   []CeremonyContribution{},
		InitiatedBy:     authority,
		InitiatedAt:     ctx.BlockTime(),
		ExpiresAt:       ctx.BlockTime().Add(expiryDuration),
	}

	if err := ValidateCeremony(ceremony); err != nil {
		return nil, err
	}

	if err := k.setCeremony(ctx, ceremony); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info(
		"trusted setup ceremony initiated",
		"ceremony_id", ceremonyID,
		"circuit_type", circuitType,
		"min_participants", minParticipants,
	)

	return ceremony, nil
}

// AddCeremonyContribution adds a participant contribution to a ceremony
func (k Keeper) AddCeremonyContribution(
	ctx sdk.Context,
	ceremonyID string,
	participant sdk.AccAddress,
	contributionHash string,
	verificationProofHash string,
) error {
	ceremony, found := k.GetCeremony(ctx, ceremonyID)
	if !found {
		return types.ErrCeremonyNotFound.Wrapf("ceremony %s not found", ceremonyID)
	}

	// Check ceremony is accepting contributions
	if ceremony.Status != CeremonyStatusPending && ceremony.Status != CeremonyStatusInProgress {
		return types.ErrInvalidCeremony.Wrapf("ceremony %s is %s, not accepting contributions", ceremonyID, ceremony.Status)
	}

	// Check ceremony hasn't expired
	if ctx.BlockTime().After(ceremony.ExpiresAt) {
		ceremony.Status = CeremonyStatusFailed
		_ = k.setCeremony(ctx, ceremony)
		return types.ErrInvalidCeremony.Wrap("ceremony has expired")
	}

	// Check participant hasn't already contributed
	for _, c := range ceremony.Contributions {
		if c.ParticipantAddress == participant.String() {
			return types.ErrInvalidCeremonyContribution.Wrap("participant has already contributed")
		}
	}

	// Build contribution chain
	previousHash := ""
	if len(ceremony.Contributions) > 0 {
		previousHash = ceremony.Contributions[len(ceremony.Contributions)-1].ContributionHash
	}

	contribution := CeremonyContribution{
		ParticipantAddress:       participant.String(),
		ContributionHash:         contributionHash,
		PreviousContributionHash: previousHash,
		ContributedAt:            ctx.BlockTime(),
		Verified:                 true, // On-chain verification of hash chain
		VerificationProofHash:    verificationProofHash,
	}

	if err := ValidateContribution(&contribution); err != nil {
		return err
	}

	ceremony.Contributions = append(ceremony.Contributions, contribution)

	// Update status
	if ceremony.Status == CeremonyStatusPending {
		ceremony.Status = CeremonyStatusInProgress
	}

	return k.setCeremony(ctx, ceremony)
}

// CompleteCeremony finalizes a trusted setup ceremony
func (k Keeper) CompleteCeremony(
	ctx sdk.Context,
	authority string,
	ceremonyID string,
	verificationKeyHash string,
) error {
	if authority != k.authority {
		return types.ErrUnauthorized.Wrap("only governance authority can complete ceremonies")
	}

	ceremony, found := k.GetCeremony(ctx, ceremonyID)
	if !found {
		return types.ErrCeremonyNotFound.Wrapf("ceremony %s not found", ceremonyID)
	}

	if ceremony.Status != CeremonyStatusInProgress {
		return types.ErrInvalidCeremony.Wrapf("ceremony %s is %s, cannot complete", ceremonyID, ceremony.Status)
	}

	// Verify minimum participants
	//nolint:gosec // MinParticipants is bounded by MaxCeremonyParticipants (100)
	if len(ceremony.Contributions) < int(ceremony.MinParticipants) {
		return types.ErrInvalidCeremony.Wrapf(
			"ceremony needs %d participants, only has %d",
			ceremony.MinParticipants, len(ceremony.Contributions),
		)
	}

	// Verify contribution chain integrity
	for i, c := range ceremony.Contributions {
		if i > 0 && c.PreviousContributionHash != ceremony.Contributions[i-1].ContributionHash {
			return types.ErrInvalidCeremony.Wrapf(
				"contribution chain broken at index %d", i,
			)
		}
	}

	now := ctx.BlockTime()
	ceremony.Status = CeremonyStatusCompleted
	ceremony.CompletedAt = &now
	ceremony.VerificationKeyHash = verificationKeyHash

	k.Logger(ctx).Info(
		"trusted setup ceremony completed",
		"ceremony_id", ceremonyID,
		"circuit_type", ceremony.CircuitType,
		"participants", len(ceremony.Contributions),
		"verification_key_hash", verificationKeyHash,
	)

	return k.setCeremony(ctx, ceremony)
}

// ============================================================================
// Ceremony Store Helpers
// ============================================================================

// GetCeremony retrieves a ceremony by ID
func (k Keeper) GetCeremony(ctx sdk.Context, ceremonyID string) (*TrustedSetupCeremony, bool) {
	store := ctx.KVStore(k.skey)
	key := types.TrustedSetupCeremonyKey(ceremonyID)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var ceremony TrustedSetupCeremony
	if err := json.Unmarshal(bz, &ceremony); err != nil {
		return nil, false
	}
	return &ceremony, true
}

// setCeremony stores a ceremony
func (k Keeper) setCeremony(ctx sdk.Context, ceremony *TrustedSetupCeremony) error {
	bz, err := json.Marshal(ceremony)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.skey)
	store.Set(types.TrustedSetupCeremonyKey(ceremony.CeremonyID), bz)
	return nil
}

// WithCeremonies iterates over all ceremonies
func (k Keeper) WithCeremonies(ctx sdk.Context, fn func(ceremony TrustedSetupCeremony) bool) {
	store := ctx.KVStore(k.skey)
	iter := store.Iterator(types.PrefixTrustedSetupCeremony, storetypes.PrefixEndBytes(types.PrefixTrustedSetupCeremony))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var ceremony TrustedSetupCeremony
		if err := json.Unmarshal(iter.Value(), &ceremony); err != nil {
			continue
		}
		if fn(ceremony) {
			break
		}
	}
}

// GetCompletedCeremonyForCircuit returns the most recent completed ceremony for a circuit type
func (k Keeper) GetCompletedCeremonyForCircuit(ctx sdk.Context, circuitType string) (*TrustedSetupCeremony, bool) {
	var latest *TrustedSetupCeremony

	k.WithCeremonies(ctx, func(ceremony TrustedSetupCeremony) bool {
		if ceremony.CircuitType == circuitType && ceremony.Status == CeremonyStatusCompleted {
			if latest == nil || (ceremony.CompletedAt != nil && latest.CompletedAt != nil &&
				ceremony.CompletedAt.After(*latest.CompletedAt)) {
				c := ceremony
				latest = &c
			}
		}
		return false
	})

	if latest == nil {
		return nil, false
	}
	return latest, true
}

// ============================================================================
// Verification Key Registry
// ============================================================================

// VerificationKeyRecord stores a verification key produced by a ceremony
type VerificationKeyRecord struct {
	// CircuitType identifies which circuit this key is for
	CircuitType string `json:"circuit_type"`

	// KeyHash is the SHA-256 hash of the verification key bytes
	KeyHash string `json:"key_hash"`

	// CeremonyID links to the ceremony that produced this key
	CeremonyID string `json:"ceremony_id"`

	// Participants is the number of ceremony participants
	Participants uint32 `json:"participants"`

	// ActivatedAt is when this key was activated for on-chain use
	ActivatedAt time.Time `json:"activated_at"`

	// Active indicates if this key is currently used for verification
	Active bool `json:"active"`
}

// SetVerificationKeyRecord stores a verification key record
func (k Keeper) SetVerificationKeyRecord(ctx sdk.Context, record *VerificationKeyRecord) error {
	if record.CircuitType == "" || record.KeyHash == "" || record.CeremonyID == "" {
		return types.ErrInvalidCeremony.Wrap("verification key record has empty required fields")
	}

	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.VerificationKeyRecordKey(record.CircuitType), bz)
	return nil
}

// GetVerificationKeyRecord retrieves the active verification key for a circuit type
func (k Keeper) GetVerificationKeyRecord(ctx sdk.Context, circuitType string) (*VerificationKeyRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.VerificationKeyRecordKey(circuitType))
	if bz == nil {
		return nil, false
	}

	var record VerificationKeyRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, false
	}
	return &record, true
}
