package dex

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
)

var (
	ErrRouteNotCertified    = errors.New("DEX route is not certified for execution")
	ErrWrongChain           = errors.New("DEX evidence chain ID mismatch")
	ErrPoolNotAllowed       = errors.New("DEX pool is not allowlisted")
	ErrTokenDecimals        = errors.New("DEX token decimals mismatch")
	ErrPoolStateStale       = errors.New("DEX pool state is stale")
	ErrPoolStateEvidence    = errors.New("DEX pool state evidence is invalid")
	ErrRouteCycle           = errors.New("DEX route contains a cycle")
	ErrRouteHopsExceeded    = errors.New("DEX route exceeds maximum hops")
	ErrPriceImpactExceeded  = errors.New("DEX quote price impact exceeds limit")
	ErrOracleDeviation      = errors.New("DEX quote deviates from oracle")
	ErrExecutionPayload     = errors.New("DEX execution payload does not match quote")
	ErrMinimumOutput        = errors.New("DEX output is below minimum")
	ErrRemoteResponseTooBig = errors.New("DEX remote response exceeds body limit")
)

const (
	quoteCanonicalVersion   = uint32(1)
	executionPayloadVersion = uint32(1)
)

// PoolStateEvidence binds one hop to one finalized observed pool state.
type PoolStateEvidence struct {
	ChainID      string            `json:"chain_id"`
	SourceID     string            `json:"source_id"`
	ProfileID    string            `json:"profile_id"`
	PoolID       string            `json:"pool_id"`
	Height       uint64            `json:"height"`
	BlockHash    string            `json:"block_hash"`
	ObservedAt   time.Time         `json:"observed_at"`
	FromDenom    string            `json:"from_denom"`
	ToDenom      string            `json:"to_denom"`
	FromDecimals uint8             `json:"from_decimals"`
	ToDecimals   uint8             `json:"to_decimals"`
	ReserveIn    sdkmath.Int       `json:"reserve_in"`
	ReserveOut   sdkmath.Int       `json:"reserve_out"`
	SwapFee      sdkmath.LegacyDec `json:"swap_fee"`
	StateDigest  string            `json:"state_digest"`
}

// ChainObservation identifies one authenticated block observation. SourceID is
// an opaque identity for the node or evidence authority and prevents evidence
// from one source being attached to pool bytes returned by another.
type ChainObservation struct {
	ChainID    string
	SourceID   string
	Height     uint64
	BlockHash  string
	ObservedAt time.Time
}

// ChainEvidenceProvider resolves canonical chain/block evidence at the process boundary.
type ChainEvidenceProvider interface {
	SourceID() string
	LatestObservation() (ChainObservation, error)
	BlockHash(height uint64) (string, error)
}

// BoundOsmosisPoolResponse carries the exact response bytes together with the
// authenticated block observation from which they were acquired.
//
// ResponseHeight and ResponseSourceID are independently verified response
// metadata. They must match Observation. A provider that cannot prove these
// bindings (for example, an unheighted current-state REST response) must return
// an error rather than populate this structure from request timing.
type BoundOsmosisPoolResponse struct {
	Payload           []byte
	ResponseHeight    uint64
	ResponseBlockHash string
	ResponseSourceID  string
	Observation       ChainObservation
}

// OsmosisPoolStateProvider is the mandatory acquisition boundary for Osmosis
// reserve data. Implementations may use a single authenticated API or a
// height-pinned archival query, but must return pool bytes and their verified
// height/source/block observation together.
type OsmosisPoolStateProvider interface {
	PoolState(ctx context.Context, poolID string) (BoundOsmosisPoolResponse, error)
	PoolStates(ctx context.Context, limit int) (BoundOsmosisPoolResponse, error)
}

// OraclePriceProvider supplies an exact quote/base price for deviation checks.
type OraclePriceProvider interface {
	Price(baseDenom, quoteDenom string, atHeight uint64) (sdkmath.LegacyDec, error)
}

// ExecutionEnvelopeVerifier proves that signedTx cryptographically authorizes
// the exact canonical payload. Production construction fails without it.
type ExecutionEnvelopeVerifier interface {
	VerifySignedExecution(ctx context.Context, payload, signedTx []byte) error
}

// ExecutionPayload is the only payload accepted by RealOsmosisAdapter.ExecuteSwap.
type ExecutionPayload struct {
	Version     uint32           `json:"version"`
	ProfileID   string           `json:"profile_id"`
	ChainID     string           `json:"chain_id"`
	QuoteID     string           `json:"quote_id"`
	QuoteDigest string           `json:"quote_digest"`
	Sender      string           `json:"sender"`
	Recipient   string           `json:"recipient"`
	InputDenom  string           `json:"input_denom"`
	OutputDenom string           `json:"output_denom"`
	InputAmount string           `json:"input_amount"`
	MinOutput   string           `json:"min_output"`
	ExpiresAt   int64            `json:"expires_at_unix_nano"`
	Routes      []ExecutionRoute `json:"routes"`
}

// ExecutionRoute is one exact Osmosis pool-manager route.
type ExecutionRoute struct {
	PoolID        string `json:"pool_id"`
	TokenOutDenom string `json:"token_out_denom"`
}

// SignedExecutionEnvelope carries the exact unsigned payload and external
// signer output. The adapter verifies payload binding; custody verifies signature
// policy before constructing this envelope.
type SignedExecutionEnvelope struct {
	Payload       []byte          `json:"payload"`
	PayloadDigest string          `json:"payload_digest"`
	SignedTx      json.RawMessage `json:"signed_tx"`
}

type canonicalQuote struct {
	Version              uint32              `json:"version"`
	ProfileID            string              `json:"profile_id"`
	ChainID              string              `json:"chain_id"`
	DEX                  string              `json:"dex"`
	DEXVersion           string              `json:"dex_version"`
	InputDenom           string              `json:"input_denom"`
	InputDecimals        uint8               `json:"input_decimals"`
	OutputDenom          string              `json:"output_denom"`
	OutputDecimals       uint8               `json:"output_decimals"`
	SwapType             SwapType            `json:"swap_type"`
	RequestedAmount      string              `json:"requested_amount"`
	Sender               string              `json:"sender"`
	Recipient            string              `json:"recipient"`
	Deadline             int64               `json:"deadline_unix_nano"`
	Slippage             string              `json:"slippage"`
	InputAmount          string              `json:"input_amount"`
	OutputAmount         string              `json:"output_amount"`
	MinOutputAmount      string              `json:"min_output_amount"`
	PriceImpact          string              `json:"price_impact"`
	OraclePrice          string              `json:"oracle_price"`
	OracleDeviation      string              `json:"oracle_deviation"`
	CreatedAt            int64               `json:"created_at_unix_nano"`
	ExpiresAt            int64               `json:"expires_at_unix_nano"`
	ObservationHeight    uint64              `json:"observation_height"`
	ObservationBlockHash string              `json:"observation_block_hash"`
	Hops                 []canonicalQuoteHop `json:"hops"`
}

type canonicalQuoteHop struct {
	PoolID      string `json:"pool_id"`
	FromDenom   string `json:"from_denom"`
	ToDenom     string `json:"to_denom"`
	AmountIn    string `json:"amount_in"`
	AmountOut   string `json:"amount_out"`
	Fee         string `json:"fee"`
	StateDigest string `json:"state_digest"`
	Height      uint64 `json:"height"`
	BlockHash   string `json:"block_hash"`
	ObservedAt  int64  `json:"observed_at_unix_nano"`
}

// QuoteDigest derives a deterministic quote identity from canonical economic
// content and exact pool-state evidence. No random input is used.
func QuoteDigest(quote SwapQuote) (string, error) {
	canonical, err := quote.canonical()
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func genericQuoteDigest(quote SwapQuote) (string, error) {
	type generic struct {
		From    string   `json:"from"`
		To      string   `json:"to"`
		Amount  string   `json:"amount"`
		Type    SwapType `json:"type"`
		Input   string   `json:"input"`
		Output  string   `json:"output"`
		Minimum string   `json:"minimum"`
		Route   string   `json:"route"`
	}
	encoded, err := json.Marshal(generic{
		From:   quote.Request.FromToken.Denom + "\x00" + quote.Request.FromToken.Symbol,
		To:     quote.Request.ToToken.Denom + "\x00" + quote.Request.ToToken.Symbol,
		Amount: intString(quote.Request.Amount), Type: quote.Request.Type,
		Input: intString(quote.InputAmount), Output: intString(quote.OutputAmount), Minimum: intString(quote.MinOutputAmount), Route: routeSignature(quote.Route),
	})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return "quote_" + hex.EncodeToString(digest[:]), nil
}

func (q SwapQuote) canonical() (canonicalQuote, error) {
	if len(q.Route.Hops) == 0 || len(q.PoolStateEvidence) != len(q.Route.Hops) {
		return canonicalQuote{}, ErrPoolStateEvidence
	}
	result := canonicalQuote{
		Version: quoteCanonicalVersion, ProfileID: q.ProfileID, ChainID: q.ChainID, DEX: q.DEX,
		DEXVersion: q.DEXVersion, InputDenom: q.Request.FromToken.Denom, InputDecimals: q.Request.FromToken.Decimals,
		OutputDenom: q.Request.ToToken.Denom, OutputDecimals: q.Request.ToToken.Decimals,
		SwapType: q.Request.Type, RequestedAmount: intString(q.Request.Amount), Sender: q.Request.Sender,
		Recipient: q.Request.Recipient, Deadline: q.Request.Deadline.UTC().UnixNano(), Slippage: decString(q.Request.SlippageToleranceExact),
		InputAmount: intString(q.InputAmount), OutputAmount: intString(q.OutputAmount), MinOutputAmount: intString(q.MinOutputAmount),
		PriceImpact: decString(q.PriceImpactExact), OraclePrice: decString(q.OraclePrice), OracleDeviation: decString(q.OracleDeviation),
		CreatedAt: q.CreatedAt.UTC().UnixNano(), ExpiresAt: q.ExpiresAt.UTC().UnixNano(),
		ObservationHeight: q.ObservationHeight, ObservationBlockHash: q.ObservationBlockHash,
		Hops: make([]canonicalQuoteHop, 0, len(q.Route.Hops)),
	}
	for i, hop := range q.Route.Hops {
		evidence := q.PoolStateEvidence[i]
		result.Hops = append(result.Hops, canonicalQuoteHop{
			PoolID: hop.PoolID, FromDenom: hop.FromToken.Denom, ToDenom: hop.ToToken.Denom,
			AmountIn: intString(hop.AmountIn), AmountOut: intString(hop.AmountOut), Fee: decString(hop.Fee),
			StateDigest: evidence.StateDigest, Height: evidence.Height, BlockHash: evidence.BlockHash,
			ObservedAt: evidence.ObservedAt.UTC().UnixNano(),
		})
	}
	return result, nil
}

func intString(value sdkmath.Int) string {
	if value.IsNil() {
		return ""
	}
	return value.String()
}

func decString(value sdkmath.LegacyDec) string {
	if value.IsNil() {
		return ""
	}
	return value.String()
}

// BuildExecutionPayload builds deterministic unsigned execution bytes bound to
// the complete quote digest.
func BuildExecutionPayload(quote SwapQuote) ([]byte, error) {
	digest, err := QuoteDigest(quote)
	if err != nil {
		return nil, err
	}
	if quote.ID != digest || quote.QuoteDigest != digest {
		return nil, fmt.Errorf("%w: quote digest mismatch", ErrExecutionPayload)
	}
	recipient := quote.Request.Recipient
	if recipient == "" {
		recipient = quote.Request.Sender
	}
	payload := ExecutionPayload{
		Version: executionPayloadVersion, ProfileID: quote.ProfileID, ChainID: quote.ChainID,
		QuoteID: quote.ID, QuoteDigest: digest, Sender: quote.Request.Sender, Recipient: recipient,
		InputDenom: quote.Request.FromToken.Denom, OutputDenom: quote.Request.ToToken.Denom,
		InputAmount: quote.InputAmount.String(), MinOutput: quote.MinOutputAmount.String(), ExpiresAt: quote.ExpiresAt.UTC().UnixNano(),
		Routes: make([]ExecutionRoute, 0, len(quote.Route.Hops)),
	}
	for _, hop := range quote.Route.Hops {
		payload.Routes = append(payload.Routes, ExecutionRoute{PoolID: hop.PoolID, TokenOutDenom: hop.ToToken.Denom})
	}
	return json.Marshal(payload)
}

// MarshalSignedExecutionEnvelope creates transport bytes after a custody signer
// has signed the exact execution payload.
func MarshalSignedExecutionEnvelope(payload, signedTx []byte) ([]byte, error) {
	if len(payload) == 0 || len(signedTx) == 0 {
		return nil, ErrExecutionPayload
	}
	digest := sha256.Sum256(payload)
	rawSignedTx, err := encodeRawSignedTx(signedTx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(SignedExecutionEnvelope{Payload: payload, PayloadDigest: hex.EncodeToString(digest[:]), SignedTx: rawSignedTx})
}

func verifySignedExecutionEnvelope(quote SwapQuote, encoded []byte) ([]byte, error) {
	var envelope SignedExecutionEnvelope
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExecutionPayload, err)
	}
	expected, err := BuildExecutionPayload(quote)
	if err != nil {
		return nil, err
	}
	if string(envelope.Payload) != string(expected) || len(envelope.SignedTx) == 0 {
		return nil, ErrExecutionPayload
	}
	digest := sha256.Sum256(envelope.Payload)
	if envelope.PayloadDigest != hex.EncodeToString(digest[:]) {
		return nil, ErrExecutionPayload
	}
	return decodeRawSignedTx(envelope.SignedTx)
}

func encodeRawSignedTx(value []byte) (json.RawMessage, error) {
	if json.Valid(value) {
		return append(json.RawMessage(nil), value...), nil
	}
	encoded, err := json.Marshal(value)
	return encoded, err
}

func decodeRawSignedTx(value json.RawMessage) ([]byte, error) {
	if len(value) == 0 || string(value) == "null" {
		return nil, ErrExecutionPayload
	}
	if value[0] == '"' {
		var decoded []byte
		if err := json.Unmarshal(value, &decoded); err != nil || len(decoded) == 0 {
			return nil, ErrExecutionPayload
		}
		return decoded, nil
	}
	if !json.Valid(value) {
		return nil, ErrExecutionPayload
	}
	return append([]byte(nil), value...), nil
}

func canonicalPoolStateDigest(evidence PoolStateEvidence) (string, error) {
	type canonicalState struct {
		ChainID      string `json:"chain_id"`
		SourceID     string `json:"source_id"`
		ProfileID    string `json:"profile_id"`
		PoolID       string `json:"pool_id"`
		Height       uint64 `json:"height"`
		BlockHash    string `json:"block_hash"`
		FromDenom    string `json:"from_denom"`
		ToDenom      string `json:"to_denom"`
		FromDecimals uint8  `json:"from_decimals"`
		ToDecimals   uint8  `json:"to_decimals"`
		ReserveIn    string `json:"reserve_in"`
		ReserveOut   string `json:"reserve_out"`
		SwapFee      string `json:"swap_fee"`
	}
	encoded, err := json.Marshal(canonicalState{
		ChainID: evidence.ChainID, SourceID: evidence.SourceID, ProfileID: evidence.ProfileID, PoolID: evidence.PoolID,
		Height: evidence.Height, BlockHash: evidence.BlockHash,
		FromDenom: evidence.FromDenom, ToDenom: evidence.ToDenom, FromDecimals: evidence.FromDecimals, ToDecimals: evidence.ToDecimals,
		ReserveIn: intString(evidence.ReserveIn), ReserveOut: intString(evidence.ReserveOut), SwapFee: decString(evidence.SwapFee),
	})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

// ConstantProductExactIn calculates floor(reserveOut * amountInAfterFee /
// (reserveIn + amountInAfterFee)) with exact integers and 18-digit fee scale.
func ConstantProductExactIn(reserveIn, reserveOut, amountIn sdkmath.Int, fee sdkmath.LegacyDec) (sdkmath.Int, error) {
	if reserveIn.IsNil() || reserveOut.IsNil() || amountIn.IsNil() || !reserveIn.IsPositive() || !reserveOut.IsPositive() || !amountIn.IsPositive() {
		return sdkmath.Int{}, ErrInsufficientLiquidity
	}
	if fee.IsNil() || fee.IsNegative() || !fee.LT(sdkmath.LegacyOneDec()) {
		return sdkmath.Int{}, fmt.Errorf("invalid swap fee")
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(sdkmath.LegacyPrecision), nil)
	feeUnits := fee.BigInt()
	multiplier := new(big.Int).Sub(scale, feeUnits)
	amountAfterFeeNumerator := new(big.Int).Mul(amountIn.BigInt(), multiplier)
	numerator := new(big.Int).Mul(reserveOut.BigInt(), amountAfterFeeNumerator)
	denominator := new(big.Int).Add(new(big.Int).Mul(reserveIn.BigInt(), scale), amountAfterFeeNumerator)
	if denominator.Sign() <= 0 {
		return sdkmath.Int{}, ErrInsufficientLiquidity
	}
	output := new(big.Int).Quo(numerator, denominator)
	if output.Sign() <= 0 || output.Cmp(reserveOut.BigInt()) >= 0 {
		return sdkmath.Int{}, ErrInsufficientLiquidity
	}
	return sdkmath.NewIntFromBigInt(output), nil
}

func exactPriceImpact(reserveIn, reserveOut, amountIn, amountOut sdkmath.Int) (sdkmath.LegacyDec, error) {
	if !reserveIn.IsPositive() || !reserveOut.IsPositive() || !amountIn.IsPositive() || !amountOut.IsPositive() {
		return sdkmath.LegacyDec{}, ErrInsufficientLiquidity
	}
	spot := sdkmath.LegacyNewDecFromInt(reserveOut).Quo(sdkmath.LegacyNewDecFromInt(reserveIn))
	effective := sdkmath.LegacyNewDecFromInt(amountOut).Quo(sdkmath.LegacyNewDecFromInt(amountIn))
	if effective.GTE(spot) {
		return sdkmath.LegacyZeroDec(), nil
	}
	return spot.Sub(effective).Quo(spot), nil
}

func exactExchangeRate(input, output sdkmath.Int, inputDecimals, outputDecimals uint8) (sdkmath.LegacyDec, error) {
	if input.IsNil() || output.IsNil() || !input.IsPositive() || !output.IsPositive() || inputDecimals > 18 || outputDecimals > 18 {
		return sdkmath.LegacyDec{}, ErrTokenDecimals
	}
	rate := sdkmath.LegacyNewDecFromInt(output).Quo(sdkmath.LegacyNewDecFromInt(input))
	if inputDecimals > outputDecimals {
		rate = rate.Mul(sdkmath.LegacyNewDec(10).Power(uint64(inputDecimals - outputDecimals)))
	} else if outputDecimals > inputDecimals {
		rate = rate.Quo(sdkmath.LegacyNewDec(10).Power(uint64(outputDecimals - inputDecimals)))
	}
	return rate, nil
}

func exactDeviation(actual, reference sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	if actual.IsNil() || reference.IsNil() || !actual.IsPositive() || !reference.IsPositive() {
		return sdkmath.LegacyDec{}, fmt.Errorf("invalid oracle price")
	}
	return actual.Sub(reference).Abs().Quo(reference), nil
}
