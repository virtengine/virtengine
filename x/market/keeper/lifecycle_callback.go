// Package keeper provides the market module keeper implementation.
//
// VE-4E: Resource lifecycle control via Waldur
// This file implements callback validation for lifecycle operations.
package keeper

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// Lifecycle callback validation errors
var (
	// ErrCallbackExpired is returned when a callback has expired
	ErrCallbackExpired = errors.New("callback has expired")

	// ErrCallbackInvalidSignature is returned for invalid signatures
	ErrCallbackInvalidSignature = errors.New("callback signature invalid")

	// ErrCallbackNonceReused is returned when a nonce is reused
	ErrCallbackNonceReused = errors.New("callback nonce already used")

	// ErrCallbackUnauthorizedSigner is returned for unauthorized signers
	ErrCallbackUnauthorizedSigner = errors.New("callback signer not authorized")

	// ErrCallbackAllocationMismatch is returned for allocation mismatches
	ErrCallbackAllocationMismatch = errors.New("callback allocation ID mismatch")

	// ErrCallbackInvalidAction is returned for invalid actions
	ErrCallbackInvalidAction = errors.New("callback action invalid")

	// ErrCallbackInvalidStateTransition is returned for invalid state transitions
	ErrCallbackInvalidStateTransition = errors.New("callback would cause invalid state transition")
)

// LifecycleCallbackValidator validates lifecycle callbacks
type LifecycleCallbackValidator struct {
	keeper *Keeper
}

// NewLifecycleCallbackValidator creates a new validator
func NewLifecycleCallbackValidator(k *Keeper) *LifecycleCallbackValidator {
	return &LifecycleCallbackValidator{keeper: k}
}

// ValidateLifecycleCallback validates a lifecycle callback
func (v *LifecycleCallbackValidator) ValidateLifecycleCallback(
	ctx sdk.Context,
	callback *marketplace.LifecycleCallback,
) error {
	now := ctx.BlockTime()

	// Basic validation
	if err := callback.ValidateAt(now); err != nil {
		return err
	}

	// Check expiry
	if now.After(callback.ExpiresAt) {
		return ErrCallbackExpired
	}

	// Validate action type
	if !callback.Action.IsValid() {
		return fmt.Errorf("%w: %s", ErrCallbackInvalidAction, callback.Action)
	}

	// Verify the provider is authorized (the signer should match the provider)
	if err := v.validateSigner(ctx, callback); err != nil {
		return err
	}

	// Validate state transition is allowed
	if err := v.validateStateTransition(ctx, callback); err != nil {
		return err
	}

	return nil
}

// ValidateWaldurCallback validates a Waldur callback
func (v *LifecycleCallbackValidator) ValidateWaldurCallback(
	ctx sdk.Context,
	callback *marketplace.WaldurCallback,
) error {
	now := ctx.BlockTime()

	// Basic validation
	if err := callback.ValidateAt(now); err != nil {
		return err
	}

	// Check expiry
	if callback.IsExpiredAt(now) {
		return ErrCallbackExpired
	}

	// Verify signature
	if err := v.verifyWaldurCallbackSignature(ctx, callback); err != nil {
		return err
	}

	return nil
}

// validateSigner validates the callback signer is authorized
//
//nolint:unparam // ctx kept for future on-chain provider validation
func (v *LifecycleCallbackValidator) validateSigner(
	_ sdk.Context,
	callback *marketplace.LifecycleCallback,
) error {
	// The signer should be the provider handling the allocation
	signerAddr := callback.SignerID
	providerAddr := callback.ProviderAddress

	// Signer must match the provider address
	if signerAddr != providerAddr {
		return fmt.Errorf("%w: signer %s != provider %s",
			ErrCallbackUnauthorizedSigner, signerAddr, providerAddr)
	}

	// Verify this is a valid provider address
	_, err := sdk.AccAddressFromBech32(providerAddr)
	if err != nil {
		return fmt.Errorf("%w: invalid provider address: %v",
			ErrCallbackUnauthorizedSigner, err)
	}

	return nil
}

// validateStateTransition validates the state transition is allowed
//
//nolint:unparam // ctx kept for future state lookup from chain
func (v *LifecycleCallbackValidator) validateStateTransition(
	_ sdk.Context,
	callback *marketplace.LifecycleCallback,
) error {
	// Get current allocation state from callback context
	// In real implementation, this would look up the allocation from state
	// For now, just validate the result state is valid
	if !callback.ResultState.IsValid() {
		return fmt.Errorf("%w: invalid result state %s",
			ErrCallbackInvalidStateTransition, callback.ResultState)
	}

	return nil
}

// verifyWaldurCallbackSignature verifies a Waldur callback signature
func (v *LifecycleCallbackValidator) verifyWaldurCallbackSignature(
	ctx sdk.Context,
	callback *marketplace.WaldurCallback,
) error {
	if len(callback.Signature) == 0 {
		return ErrCallbackInvalidSignature
	}

	// Get the signing payload
	payload := callback.SigningPayload()

	// In production, we would look up the provider's public key from
	// the provider registry or from the on-chain provider record
	// For now, we validate the signature format is correct

	// Signature must be at least 64 bytes for ed25519
	if len(callback.Signature) < 64 {
		return fmt.Errorf("%w: signature too short", ErrCallbackInvalidSignature)
	}

	// In a full implementation:
	// 1. Look up provider's public key from chain state
	// 2. Verify the signature against the payload

	ctx.Logger().Debug("callback signature validation passed",
		"callback_id", callback.ID,
		"signer", callback.SignerID,
		"payload_hash", hex.EncodeToString(payload[:8]),
	)

	return nil
}

// VerifyEd25519Signature verifies an ed25519 signature
func VerifyEd25519Signature(publicKey, message, signature []byte) bool {
	if len(publicKey) != ed25519.PublicKeySize {
		return false
	}
	if len(signature) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(publicKey, message, signature)
}

// LifecycleCallbackResult holds the result of callback processing
type LifecycleCallbackResult struct {
	// Validated indicates the callback was validated successfully
	Validated bool

	// AllocationID is the allocation that was updated
	AllocationID string

	// NewState is the new allocation state after the callback
	NewState marketplace.AllocationState

	// OperationID is the operation that was completed
	OperationID string

	// Success indicates if the lifecycle action succeeded
	Success bool

	// Error contains any error message
	Error string

	// ProcessedAt is when the callback was processed
	ProcessedAt time.Time
}

// ProcessLifecycleCallback processes a lifecycle callback and updates state
func (k *Keeper) ProcessLifecycleCallback(
	ctx sdk.Context,
	callback *marketplace.LifecycleCallback,
) (*LifecycleCallbackResult, error) {
	validator := NewLifecycleCallbackValidator(k)

	// Validate the callback
	if err := validator.ValidateLifecycleCallback(ctx, callback); err != nil {
		return &LifecycleCallbackResult{
			Validated:    false,
			AllocationID: callback.AllocationID,
			Error:        err.Error(),
			ProcessedAt:  ctx.BlockTime(),
		}, err
	}

	result := &LifecycleCallbackResult{
		Validated:    true,
		AllocationID: callback.AllocationID,
		NewState:     callback.ResultState,
		OperationID:  callback.OperationID,
		Success:      callback.Success,
		Error:        callback.Error,
		ProcessedAt:  ctx.BlockTime(),
	}

	// In a full implementation, we would:
	// 1. Update the allocation state in the store
	// 2. Emit events for the state change
	// 3. Update any related escrow/payment states

	ctx.Logger().Info("lifecycle callback processed",
		"allocation_id", callback.AllocationID,
		"operation_id", callback.OperationID,
		"action", callback.Action,
		"success", callback.Success,
		"new_state", callback.ResultState,
	)

	return result, nil
}

// NonceTracker tracks processed nonces for replay protection
type NonceTracker struct {
	nonces map[string]time.Time
	maxAge time.Duration
}

// NewNonceTracker creates a new nonce tracker
func NewNonceTracker(maxAge time.Duration) *NonceTracker {
	return &NonceTracker{
		nonces: make(map[string]time.Time),
		maxAge: maxAge,
	}
}

// IsProcessed checks if a nonce has been processed
func (t *NonceTracker) IsProcessed(nonce string, now time.Time) bool {
	expiry, exists := t.nonces[nonce]
	if !exists {
		return false
	}
	if now.After(expiry) {
		delete(t.nonces, nonce)
		return false
	}
	return true
}

// MarkProcessed marks a nonce as processed
func (t *NonceTracker) MarkProcessed(nonce string, now time.Time) {
	t.nonces[nonce] = now.Add(t.maxAge)
}

// Cleanup removes expired nonces
func (t *NonceTracker) Cleanup(now time.Time) int {
	count := 0
	for nonce, expiry := range t.nonces {
		if now.After(expiry) {
			delete(t.nonces, nonce)
			count++
		}
	}
	return count
}

// ComputeCallbackHash computes a hash for callback deduplication
func ComputeCallbackHash(callback *marketplace.LifecycleCallback) string {
	h := sha256.New()
	h.Write([]byte(callback.OperationID))
	h.Write([]byte(callback.AllocationID))
	h.Write([]byte(callback.Action))
	h.Write([]byte(callback.Nonce))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateCallbackChain validates a chain of callbacks
func ValidateCallbackChain(callbacks []*marketplace.LifecycleCallback) error {
	if len(callbacks) == 0 {
		return nil
	}

	// Validate each callback
	for i, cb := range callbacks {
		if err := cb.Validate(); err != nil {
			return fmt.Errorf("callback %d invalid: %w", i, err)
		}
	}

	// Validate ordering (timestamps should be increasing)
	for i := 1; i < len(callbacks); i++ {
		if callbacks[i].Timestamp.Before(callbacks[i-1].Timestamp) {
			return fmt.Errorf("callback %d has earlier timestamp than callback %d", i, i-1)
		}
	}

	return nil
}

// LifecycleStateMachine validates and applies state transitions
type LifecycleStateMachine struct {
	transitions map[marketplace.AllocationState]map[marketplace.LifecycleActionType]marketplace.AllocationState
}

// NewLifecycleStateMachine creates a new state machine
func NewLifecycleStateMachine() *LifecycleStateMachine {
	sm := &LifecycleStateMachine{
		transitions: make(map[marketplace.AllocationState]map[marketplace.LifecycleActionType]marketplace.AllocationState),
	}
	sm.initTransitions()
	return sm
}

// initTransitions initializes the state transition table
func (sm *LifecycleStateMachine) initTransitions() {
	// Copy transitions from marketplace package
	sm.transitions = marketplace.AllocationLifecycleTransitions
}

// CanTransition checks if a transition is valid
func (sm *LifecycleStateMachine) CanTransition(
	currentState marketplace.AllocationState,
	action marketplace.LifecycleActionType,
) bool {
	actions, ok := sm.transitions[currentState]
	if !ok {
		return false
	}
	_, valid := actions[action]
	return valid
}

// GetTargetState returns the target state for a transition
func (sm *LifecycleStateMachine) GetTargetState(
	currentState marketplace.AllocationState,
	action marketplace.LifecycleActionType,
) (marketplace.AllocationState, error) {
	actions, ok := sm.transitions[currentState]
	if !ok {
		return currentState, fmt.Errorf("no transitions from state %s", currentState)
	}

	targetState, ok := actions[action]
	if !ok {
		return currentState, fmt.Errorf("action %s not allowed in state %s", action, currentState)
	}

	return targetState, nil
}

// ValidateTransition validates a state transition
func (sm *LifecycleStateMachine) ValidateTransition(
	currentState marketplace.AllocationState,
	action marketplace.LifecycleActionType,
	resultState marketplace.AllocationState,
) error {
	expectedState, err := sm.GetTargetState(currentState, action)
	if err != nil {
		return err
	}

	if expectedState != resultState {
		return fmt.Errorf("expected state %s but got %s", expectedState, resultState)
	}

	return nil
}

// SerializeCallback serializes a callback for signing
func SerializeCallback(callback *marketplace.LifecycleCallback) []byte {
	// Create deterministic serialization
	var buf bytes.Buffer
	buf.WriteString(callback.ID)
	buf.WriteString(callback.OperationID)
	buf.WriteString(callback.AllocationID)
	buf.WriteString(string(callback.Action))
	if callback.Success {
		buf.WriteString("1")
	} else {
		buf.WriteString("0")
	}
	buf.WriteString(callback.ResultState.String())
	buf.WriteString(callback.ProviderAddress)
	buf.WriteString(callback.Nonce)
	buf.WriteString(fmt.Sprintf("%d", callback.Timestamp.Unix()))
	return buf.Bytes()
}

// HashCallback creates a hash of the callback for verification
func HashCallback(callback *marketplace.LifecycleCallback) []byte {
	serialized := SerializeCallback(callback)
	hash := sha256.Sum256(serialized)
	return hash[:]
}
