// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"

	"github.com/virtengine/virtengine/pkg/dex"
)

// NewBoundOsmosisAdapter constructs the provider-daemon DEX boundary. A plain
// LCD endpoint is never used as reserve evidence: the caller must inject a
// provider that returns pool payload bytes already bound to authenticated
// height, block hash, and node identity.
func NewBoundOsmosisAdapter(
	cfg dex.AdapterConfig,
	poolState dex.OsmosisPoolStateProvider,
	chainEvidence dex.ChainEvidenceProvider,
	oracle dex.OraclePriceProvider,
	verifier dex.ExecutionEnvelopeVerifier,
) (dex.Adapter, error) {
	if poolState == nil || chainEvidence == nil || oracle == nil || verifier == nil {
		return nil, fmt.Errorf("bound Osmosis pool, chain, oracle, and execution evidence are required")
	}
	cfg.PoolState = poolState
	cfg.ChainEvidence = chainEvidence
	cfg.Oracle = oracle
	cfg.ExecutionVerifier = verifier
	return dex.CreateAdapter(cfg)
}

// CometDEXChainEvidence adapts a target-chain Comet RPC endpoint to the exact
// block evidence required by RealOsmosisAdapter.
type CometDEXChainEvidence struct {
	chainID  string
	sourceID string
	rpc      *rpchttp.HTTP
	now      func() time.Time
}

func NewCometDEXChainEvidence(chainID, endpoint string, now func() time.Time) (*CometDEXChainEvidence, error) {
	if strings.TrimSpace(chainID) == "" || strings.TrimSpace(endpoint) == "" {
		return nil, dex.ErrWrongChain
	}
	client, err := rpchttp.New(endpoint, "/websocket")
	if err != nil {
		return nil, err
	}
	status, err := client.Status(context.Background())
	if err != nil {
		return nil, err
	}
	if status.NodeInfo.Network != chainID {
		return nil, dex.ErrWrongChain
	}
	sourceID := strings.TrimSpace(string(status.NodeInfo.ID()))
	if sourceID == "" {
		return nil, errors.New("target chain node identity unavailable")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &CometDEXChainEvidence{chainID: chainID, sourceID: sourceID, rpc: client, now: now}, nil
}

func (e *CometDEXChainEvidence) SourceID() string {
	if e == nil {
		return ""
	}
	return e.sourceID
}

func (e *CometDEXChainEvidence) LatestObservation() (dex.ChainObservation, error) {
	if e == nil || e.rpc == nil {
		return dex.ChainObservation{}, ErrFiatConversionQueryUnavailable
	}
	result, err := e.rpc.Status(context.Background())
	if err != nil {
		return dex.ChainObservation{}, err
	}
	if result.NodeInfo.Network != e.chainID {
		return dex.ChainObservation{}, dex.ErrWrongChain
	}
	sourceID := strings.TrimSpace(string(result.NodeInfo.ID()))
	if sourceID == "" || sourceID != e.sourceID || result.SyncInfo.LatestBlockHeight <= 0 {
		return dex.ChainObservation{}, errors.New("target chain has no committed height")
	}
	return dex.ChainObservation{ChainID: e.chainID, SourceID: sourceID, Height: uint64(result.SyncInfo.LatestBlockHeight), BlockHash: strings.ToUpper(hex.EncodeToString(result.SyncInfo.LatestBlockHash)), ObservedAt: e.now()}, nil //nolint:gosec // positive int64 fits uint64.
}
func (e *CometDEXChainEvidence) BlockHash(height uint64) (string, error) {
	if e == nil || e.rpc == nil || height == 0 {
		return "", ErrFiatConversionQueryUnavailable
	}
	if height > uint64(^uint64(0)>>1) {
		return "", errors.New("height overflow")
	}
	value := int64(height) //nolint:gosec // bounded to MaxInt64 above.
	result, err := e.rpc.Block(context.Background(), &value)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(result.BlockID.Hash)), nil
}

// OraclePriceFunc adapts an externally authenticated oracle query.
type OraclePriceFunc func(baseDenom, quoteDenom string, atHeight uint64) (sdkmath.LegacyDec, error)

func (f OraclePriceFunc) Price(baseDenom, quoteDenom string, atHeight uint64) (sdkmath.LegacyDec, error) {
	if f == nil {
		return sdkmath.LegacyDec{}, errors.New("oracle unavailable")
	}
	return f(baseDenom, quoteDenom, atHeight)
}

// FinalityEvidenceFunc is the target-chain transaction reconciliation seam.
// correlationID is deterministic and may be used only when the target exposes
// an authenticated client/custody correlation query.
type FinalityEvidenceFunc func(context.Context, dex.SwapQuote, string, string) (DEXSwapReconciliation, error)

func (f FinalityEvidenceFunc) ReconcileSwap(ctx context.Context, quote dex.SwapQuote, hash, correlationID string) (DEXSwapReconciliation, error) {
	if f == nil {
		return DEXSwapReconciliation{}, ErrFiatConversionQueryUnavailable
	}
	return f(ctx, quote, hash, correlationID)
}

// CanonicalDEXFinalityHash derives privacy-safe finality evidence from stable
// chain fields without including event payloads.
func CanonicalDEXFinalityHash(chainID, txHash string, height int64, blockHash []byte, confirmations uint32, output sdkmath.Int) ([]byte, error) {
	if chainID == "" || txHash == "" || height <= 0 || len(blockHash) != sha256.Size || confirmations == 0 || output.IsNil() || !output.IsPositive() {
		return nil, errors.New("invalid DEX finality evidence")
	}
	canonical := strings.Join([]string{chainID, strings.ToUpper(txHash), strconv.FormatInt(height, 10), hex.EncodeToString(blockHash), strconv.FormatUint(uint64(confirmations), 10), output.String()}, "\x00")
	digest := sha256.Sum256([]byte(canonical))
	return digest[:], nil
}

var _ dex.ChainEvidenceProvider = (*CometDEXChainEvidence)(nil)
var _ dex.OraclePriceProvider = OraclePriceFunc(nil)
var _ = fmt.Sprintf
