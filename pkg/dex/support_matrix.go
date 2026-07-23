// Package dex provides DEX integration for crypto-to-fiat conversions.
package dex

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
)

// RouteSupportState is the externally auditable lifecycle of a DEX route.
type RouteSupportState string

const (
	dexOsmosis = "osmosis"

	RouteUnsupported                        RouteSupportState = "unsupported"
	RouteEngineeringIncomplete              RouteSupportState = "engineering_incomplete"
	RouteEngineeringCompleteExternalBlocked RouteSupportState = "engineering_complete_external_blocked"
	RouteCertifiedEnabled                   RouteSupportState = "certified_enabled"
	RoutePaused                             RouteSupportState = "paused"
)

// RouteEnvironment identifies the chain environment without inferring certification.
type RouteEnvironment string

const (
	EnvironmentDevelopment RouteEnvironment = "development"
	EnvironmentTestnet     RouteEnvironment = "testnet"
	EnvironmentMainnet     RouteEnvironment = "mainnet"
)

// RouteToken binds a denom to the exact decimals expected by quote math.
type RouteToken struct {
	Symbol   string `json:"symbol"`
	Denom    string `json:"denom"`
	Decimals uint8  `json:"decimals"`
}

// RouteEvidence records non-secret certification references and accountable owners.
type RouteEvidence struct {
	EngineeringEvidence string   `json:"engineering_evidence"`
	NetworkEvidence     string   `json:"network_evidence"`
	LiquidityEvidence   string   `json:"liquidity_evidence"`
	OracleEvidence      string   `json:"oracle_evidence"`
	CustodyEvidence     string   `json:"custody_evidence"`
	GovernanceEvidence  string   `json:"governance_evidence"`
	EngineeringOwners   []string `json:"engineering_owners"`
	OperationsOwners    []string `json:"operations_owners"`
	SecurityOwners      []string `json:"security_owners"`
}

// DEXRouteProfile is one exact, versioned route support-matrix row. All ratios
// are decimal strings to keep production value checks exact.
type DEXRouteProfile struct {
	ID                  string            `json:"id"`
	State               RouteSupportState `json:"state"`
	Network             string            `json:"network"`
	ChainID             string            `json:"chain_id"`
	Environment         RouteEnvironment  `json:"environment"`
	DEX                 string            `json:"dex"`
	Version             string            `json:"version"`
	AllowedPoolIDs      []string          `json:"allowed_pool_ids"`
	Tokens              []RouteToken      `json:"tokens"`
	FinalityBlocks      uint64            `json:"finality_blocks"`
	MaxObservationAge   time.Duration     `json:"max_observation_age"`
	MaxHeightLag        uint64            `json:"max_height_lag"`
	MaxHops             uint32            `json:"max_hops"`
	MinLiquidity        sdkmath.Int       `json:"min_liquidity"`
	MinReserve          sdkmath.Int       `json:"min_reserve"`
	MaxAmount           sdkmath.Int       `json:"max_amount"`
	MaxPriceImpact      sdkmath.LegacyDec `json:"max_price_impact"`
	MaxOracleDeviation  sdkmath.LegacyDec `json:"max_oracle_deviation"`
	QuoteTTL            time.Duration     `json:"quote_ttl"`
	CustodyMode         string            `json:"custody_mode"`
	OracleSource        string            `json:"oracle_source"`
	Evidence            RouteEvidence     `json:"evidence"`
	EngineeringTestOnly bool              `json:"engineering_test_only,omitempty"`
}

// RouteValidationMode controls whether an externally blocked engineering row
// may be used. Production never opts in implicitly.
type RouteValidationMode string

const (
	RouteValidationRuntime     RouteValidationMode = "runtime"
	RouteValidationEngineering RouteValidationMode = "engineering_test"
)

// RouteProfileAuthorizer verifies that a certified profile came from the
// deployment's trusted, versioned support matrix. Profile evidence strings are
// descriptive references and are never sufficient authorization by themselves.
type RouteProfileAuthorizer interface {
	AuthorizeDEXRoute(profile DEXRouteProfile) error
}

// Executable returns true only for the production runtime state.
func (p DEXRouteProfile) Executable() bool { return p.State == RouteCertifiedEnabled }

// Validate validates a support-matrix row and its requested execution mode.
func (p DEXRouteProfile) Validate(mode RouteValidationMode) error {
	switch p.State {
	case RouteUnsupported, RouteEngineeringIncomplete, RouteEngineeringCompleteExternalBlocked, RouteCertifiedEnabled, RoutePaused:
	default:
		return fmt.Errorf("invalid DEX route state %q", p.State)
	}

	executable := p.State == RouteCertifiedEnabled
	engineeringBlocked := p.State == RouteEngineeringCompleteExternalBlocked && mode == RouteValidationEngineering && p.EngineeringTestOnly
	if !executable && !engineeringBlocked {
		return fmt.Errorf("%w: route %q in %s mode has state %s", ErrRouteNotCertified, p.ID, mode, p.State)
	}

	mandatory := map[string]string{
		"id": p.ID, "network": p.Network, "chain_id": p.ChainID, "environment": string(p.Environment),
		"dex": p.DEX, "version": p.Version, "custody_mode": p.CustodyMode, "oracle_source": p.OracleSource,
	}
	for name, value := range mandatory {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || strings.EqualFold(trimmed, "tbd") {
			return fmt.Errorf("DEX route %q has invalid mandatory %s", p.ID, name)
		}
	}
	if p.Environment != EnvironmentDevelopment && p.Environment != EnvironmentTestnet && p.Environment != EnvironmentMainnet {
		return fmt.Errorf("DEX route %q has invalid environment %q", p.ID, p.Environment)
	}
	if p.DEX != dexOsmosis || p.Version != OsmosisAdapterVersion {
		return fmt.Errorf("DEX route %q uses unsupported adapter version %s/%s", p.ID, p.DEX, p.Version)
	}
	if len(p.AllowedPoolIDs) == 0 || len(p.Tokens) < 2 || p.MaxHops == 0 || p.MaxObservationAge <= 0 || p.QuoteTTL <= 0 || p.FinalityBlocks == 0 {
		return fmt.Errorf("DEX route %q has incomplete route/finality bounds", p.ID)
	}
	if p.MinLiquidity.IsNil() || !p.MinLiquidity.IsPositive() || p.MinReserve.IsNil() || !p.MinReserve.IsPositive() || p.MaxAmount.IsNil() || !p.MaxAmount.IsPositive() {
		return fmt.Errorf("DEX route %q has invalid liquidity or amount limits", p.ID)
	}
	if p.MaxPriceImpact.IsNil() || !p.MaxPriceImpact.IsPositive() || !p.MaxPriceImpact.LT(sdkmath.LegacyOneDec()) ||
		p.MaxOracleDeviation.IsNil() || !p.MaxOracleDeviation.IsPositive() || !p.MaxOracleDeviation.LT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("DEX route %q has invalid price safety limits", p.ID)
	}
	seenPools := make(map[string]struct{}, len(p.AllowedPoolIDs))
	for _, poolID := range p.AllowedPoolIDs {
		poolID = strings.TrimSpace(poolID)
		if poolID == "" || strings.EqualFold(poolID, "tbd") {
			return fmt.Errorf("DEX route %q contains an invalid pool ID", p.ID)
		}
		parsedPoolID, err := strconv.ParseUint(poolID, 10, 64)
		if err != nil || parsedPoolID == 0 || strconv.FormatUint(parsedPoolID, 10) != poolID {
			return fmt.Errorf("DEX route %q contains a non-canonical Osmosis pool ID %q", p.ID, poolID)
		}
		if _, exists := seenPools[poolID]; exists {
			return fmt.Errorf("DEX route %q contains duplicate pool ID %s", p.ID, poolID)
		}
		seenPools[poolID] = struct{}{}
	}
	seenDenoms := make(map[string]struct{}, len(p.Tokens))
	for _, token := range p.Tokens {
		if strings.TrimSpace(token.Symbol) == "" || strings.TrimSpace(token.Denom) == "" || strings.EqualFold(strings.TrimSpace(token.Denom), "tbd") || token.Decimals > 18 {
			return fmt.Errorf("DEX route %q contains an invalid token definition", p.ID)
		}
		if _, exists := seenDenoms[token.Denom]; exists {
			return fmt.Errorf("DEX route %q contains duplicate denom %s", p.ID, token.Denom)
		}
		seenDenoms[token.Denom] = struct{}{}
	}
	if executable {
		if p.Environment != EnvironmentMainnet {
			return fmt.Errorf("certified route %q must identify a mainnet environment", p.ID)
		}
		if err := p.Evidence.validateCertified(); err != nil {
			return fmt.Errorf("DEX route %q certification evidence: %w", p.ID, err)
		}
	}
	return nil
}

func (e RouteEvidence) validateCertified() error {
	mandatory := []string{e.EngineeringEvidence, e.NetworkEvidence, e.LiquidityEvidence, e.OracleEvidence, e.CustodyEvidence, e.GovernanceEvidence}
	for _, value := range mandatory {
		if strings.TrimSpace(value) == "" || strings.EqualFold(strings.TrimSpace(value), "tbd") {
			return fmt.Errorf("all evidence references are required")
		}
	}
	if len(e.EngineeringOwners) == 0 || len(e.OperationsOwners) == 0 || len(e.SecurityOwners) == 0 {
		return fmt.Errorf("engineering, operations, and security owners are required")
	}
	return nil
}

func (p DEXRouteProfile) poolAllowed(poolID string) bool {
	for _, allowed := range p.AllowedPoolIDs {
		if allowed == poolID {
			return true
		}
	}
	return false
}

func (p DEXRouteProfile) token(denom string) (RouteToken, bool) {
	for _, token := range p.Tokens {
		if token.Denom == denom {
			return token, true
		}
	}
	return RouteToken{}, false
}
