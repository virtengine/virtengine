package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EncryptionHooks defines hooks for encryption key lifecycle events.
type EncryptionHooks interface {
	AfterKeyRevoked(ctx sdk.Context, address sdk.AccAddress, fingerprint string) error
	AfterKeyRotated(ctx sdk.Context, address sdk.AccAddress, oldFingerprint, newFingerprint string) error
	AfterKeyExpired(ctx sdk.Context, address sdk.AccAddress, fingerprint string) error
}

// MultiEncryptionHooks combines multiple hook implementations.
type MultiEncryptionHooks []EncryptionHooks

// NewMultiEncryptionHooks creates a new multi hook wrapper.
func NewMultiEncryptionHooks(hooks ...EncryptionHooks) MultiEncryptionHooks {
	return MultiEncryptionHooks(hooks)
}

// AfterKeyRevoked triggers AfterKeyRevoked on all hooks.
func (h MultiEncryptionHooks) AfterKeyRevoked(ctx sdk.Context, address sdk.AccAddress, fingerprint string) error {
	for _, hook := range h {
		if err := hook.AfterKeyRevoked(ctx, address, fingerprint); err != nil {
			return err
		}
	}
	return nil
}

// AfterKeyRotated triggers AfterKeyRotated on all hooks.
func (h MultiEncryptionHooks) AfterKeyRotated(ctx sdk.Context, address sdk.AccAddress, oldFingerprint, newFingerprint string) error {
	for _, hook := range h {
		if err := hook.AfterKeyRotated(ctx, address, oldFingerprint, newFingerprint); err != nil {
			return err
		}
	}
	return nil
}

// AfterKeyExpired triggers AfterKeyExpired on all hooks.
func (h MultiEncryptionHooks) AfterKeyExpired(ctx sdk.Context, address sdk.AccAddress, fingerprint string) error {
	for _, hook := range h {
		if err := hook.AfterKeyExpired(ctx, address, fingerprint); err != nil {
			return err
		}
	}
	return nil
}
