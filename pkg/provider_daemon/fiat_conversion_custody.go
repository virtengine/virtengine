// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/virtengine/virtengine/pkg/dex"
)

var ErrDEXCustodyUnavailable = errors.New("DEX custody signer unavailable")

// DEXCustodySignRequest is sent to an external target-chain custody backend.
// ProviderAuthorization proves provider-daemon approval, but is not itself a
// target-chain transaction signature.
type DEXCustodySignRequest struct {
	Quote                 dex.SwapQuote
	ExecutionPayload      []byte
	ProviderAuthorization Signature
}

// DEXCustodyBackend creates and verifies actual target-chain Cosmos TxRaw
// bytes. It must never reinterpret a provider-chain key as a DEX-chain key.
type DEXCustodyBackend interface {
	Address(ctx context.Context, chainID string) (string, error)
	SignTxRaw(ctx context.Context, request DEXCustodySignRequest) ([]byte, error)
	RecoverTxRaw(ctx context.Context, request DEXCustodySignRequest, expectedTxHash string) ([]byte, error)
	VerifyTxRaw(ctx context.Context, payload, txRaw []byte) error
	ProductionReady(ctx context.Context) error
	TestOnly() bool
}

// DEXCustodySigner is the orchestrator's narrow custody boundary.
type DEXCustodySigner interface {
	Address(ctx context.Context, chainID string) (string, error)
	SignExecution(ctx context.Context, quote dex.SwapQuote, payload []byte) ([]byte, error)
	RecoverSignedExecution(ctx context.Context, quote dex.SwapQuote, payload []byte, expectedTxHash string) ([]byte, error)
	VerifySignedExecution(ctx context.Context, payload, txRaw []byte) error
	ProductionReady(ctx context.Context) error
	TestOnly() bool
}

// KeyManagerDEXCustodySigner binds provider approval to an external custody
// backend. The KeyManager signature authorizes the exact payload; only the
// backend can construct the target-chain TxRaw.
type KeyManagerDEXCustodySigner struct {
	keys    *KeyManager
	backend DEXCustodyBackend
}

func NewKeyManagerDEXCustodySigner(keys *KeyManager, backend DEXCustodyBackend) (*KeyManagerDEXCustodySigner, error) {
	if keys == nil || keys.IsLocked() || backend == nil {
		return nil, ErrDEXCustodyUnavailable
	}
	return &KeyManagerDEXCustodySigner{keys: keys, backend: backend}, nil
}

func (s *KeyManagerDEXCustodySigner) Address(ctx context.Context, chainID string) (string, error) {
	if s == nil || s.backend == nil {
		return "", ErrDEXCustodyUnavailable
	}
	return s.backend.Address(ctx, chainID)
}

func (s *KeyManagerDEXCustodySigner) SignExecution(ctx context.Context, quote dex.SwapQuote, payload []byte) ([]byte, error) {
	if s == nil || s.keys == nil || s.backend == nil || len(payload) == 0 {
		return nil, ErrDEXCustodyUnavailable
	}
	expected, err := dex.BuildExecutionPayload(quote)
	if err != nil || !equalBytes(expected, payload) {
		return nil, dex.ErrExecutionPayload
	}
	authorization, err := s.keys.Sign(payload)
	if err != nil {
		return nil, fmt.Errorf("provider custody authorization: %w", err)
	}
	txRaw, err := s.backend.SignTxRaw(ctx, DEXCustodySignRequest{
		Quote: quote, ExecutionPayload: append([]byte(nil), payload...), ProviderAuthorization: *authorization,
	})
	if err != nil || len(txRaw) == 0 {
		return nil, fmt.Errorf("%w: target-chain signer failed", ErrDEXCustodyUnavailable)
	}
	if err := s.backend.VerifyTxRaw(ctx, payload, txRaw); err != nil {
		return nil, fmt.Errorf("verify target-chain TxRaw: %w", err)
	}
	return txRaw, nil
}

func (s *KeyManagerDEXCustodySigner) VerifySignedExecution(ctx context.Context, payload, txRaw []byte) error {
	if s == nil || s.backend == nil {
		return ErrDEXCustodyUnavailable
	}
	return s.backend.VerifyTxRaw(ctx, payload, txRaw)
}

func (s *KeyManagerDEXCustodySigner) RecoverSignedExecution(ctx context.Context, quote dex.SwapQuote, payload []byte, expectedTxHash string) ([]byte, error) {
	if s == nil || s.keys == nil || s.backend == nil || expectedTxHash == "" {
		return nil, ErrDEXCustodyUnavailable
	}
	expected, err := dex.BuildExecutionPayload(quote)
	if err != nil || !equalBytes(expected, payload) {
		return nil, dex.ErrExecutionPayload
	}
	authorization, err := s.keys.Sign(payload)
	if err != nil {
		return nil, fmt.Errorf("provider custody recovery authorization: %w", err)
	}
	txRaw, err := s.backend.RecoverTxRaw(ctx, DEXCustodySignRequest{Quote: quote, ExecutionPayload: append([]byte(nil), payload...), ProviderAuthorization: *authorization}, expectedTxHash)
	if err != nil || len(txRaw) == 0 || !equalHexDigest(expectedTxHash, txRaw) {
		return nil, fmt.Errorf("%w: target-chain transaction recovery mismatch", ErrDEXCustodyUnavailable)
	}
	if err := s.backend.VerifyTxRaw(ctx, payload, txRaw); err != nil {
		return nil, fmt.Errorf("verify recovered target-chain TxRaw: %w", err)
	}
	return txRaw, nil
}

func (s *KeyManagerDEXCustodySigner) ProductionReady(ctx context.Context) error {
	if s == nil || s.keys == nil || s.keys.IsLocked() || s.backend == nil || s.backend.TestOnly() {
		return ErrDEXCustodyUnavailable
	}
	return s.backend.ProductionReady(ctx)
}

func (s *KeyManagerDEXCustodySigner) TestOnly() bool {
	return s == nil || s.backend == nil || s.backend.TestOnly()
}

// DEXExecutionVerifierAdapter connects the custody boundary to the hardened DEX adapter.
type DEXExecutionVerifierAdapter struct{ Signer DEXCustodySigner }

func (a DEXExecutionVerifierAdapter) VerifySignedExecution(ctx context.Context, payload, signedTx []byte) error {
	if a.Signer == nil {
		return ErrDEXCustodyUnavailable
	}
	return a.Signer.VerifySignedExecution(ctx, payload, signedTx)
}

func equalBytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	leftDigest, rightDigest := sha256.Sum256(left), sha256.Sum256(right)
	return leftDigest == rightDigest
}

func equalHexDigest(expected string, value []byte) bool {
	digest := sha256.Sum256(value)
	return strings.EqualFold(expected, hex.EncodeToString(digest[:]))
}

var _ DEXCustodySigner = (*KeyManagerDEXCustodySigner)(nil)
var _ dex.ExecutionEnvelopeVerifier = DEXExecutionVerifierAdapter{}
