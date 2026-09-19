// Package marketplace provides types for the marketplace on-chain module.
//
// This file implements the deterministic resolution engine defined by
// _docs/adr/ADR-010-unified-market-resolution-and-waldur-supply.md. A supply
// listing is treated as a standing ask and a provider bid as a standing bid;
// the engine resolves a demand order against the union of both using an
// integer-only, fully deterministic ranking.
package marketplace

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CandidateKind identifies the origin of a resolution candidate.
type CandidateKind string

const (
	// CandidateKindListing is a standing ask published as an offering.
	CandidateKindListing CandidateKind = "listing"

	// CandidateKindBid is a standing bid placed by a provider on an order.
	CandidateKindBid CandidateKind = "bid"
)

// ResolutionPolicy controls deterministic candidate selection.
type ResolutionPolicy struct {
	// AllowPartialFill permits satisfying an order across multiple candidates.
	AllowPartialFill bool `json:"allow_partial_fill"`

	// PreferNative breaks exact price ties in favour of native listings.
	PreferNative bool `json:"prefer_native"`

	// PreferListings breaks exact price ties in favour of standing asks over bids.
	PreferListings bool `json:"prefer_listings"`
}

// DefaultResolutionPolicy returns the conservative default policy.
func DefaultResolutionPolicy() ResolutionPolicy {
	return ResolutionPolicy{
		AllowPartialFill: false,
		PreferNative:     true,
		PreferListings:   true,
	}
}

// Candidate is a single resolvable supply quote. All fields are committed state,
// so ranking is deterministic.
type Candidate struct {
	// Kind is the candidate origin.
	Kind CandidateKind `json:"kind"`

	// OfferingID is set for listing candidates.
	OfferingID *OfferingID `json:"offering_id,omitempty"`

	// BidID is set for bid candidates.
	BidID *BidID `json:"bid_id,omitempty"`

	// ProviderAddress is the supplying provider.
	ProviderAddress string `json:"provider_address"`

	// Price is the total price for the requested order quantity from this candidate.
	Price sdk.Coin `json:"price"`

	// Capacity is the quantity the candidate can supply.
	Capacity uint64 `json:"capacity"`

	// ProviderScore is an integer reputation/benchmark score (higher is better).
	ProviderScore uint32 `json:"provider_score"`

	// CapacityFit is an integer fit score (higher is better).
	CapacityFit uint64 `json:"capacity_fit"`

	// Sequence is the deterministic candidate sequence (offering or bid sequence).
	Sequence uint64 `json:"sequence"`

	// Source identifies the supply origin for tie-breaking.
	Source OfferingSource `json:"source"`
}

// Validate validates a candidate.
func (c Candidate) Validate() error {
	switch c.Kind {
	case CandidateKindListing:
		if c.OfferingID == nil {
			return fmt.Errorf("listing candidate requires an offering id")
		}
	case CandidateKindBid:
		if c.BidID == nil {
			return fmt.Errorf("bid candidate requires a bid id")
		}
	default:
		return fmt.Errorf("invalid candidate kind: %s", c.Kind)
	}
	if c.ProviderAddress == "" {
		return fmt.Errorf("candidate provider address is required")
	}
	if !c.Price.IsValid() || !c.Price.Amount.IsPositive() {
		return fmt.Errorf("candidate price is invalid")
	}
	if c.Capacity == 0 {
		return fmt.Errorf("candidate capacity must be positive")
	}
	return nil
}

// ResolutionResult is the outcome of resolving an order.
type ResolutionResult struct {
	// Matches are the selected candidates in deterministic priority order.
	Matches []Candidate `json:"matches"`

	// Matched is true when the order was fully or partially satisfied.
	Matched bool `json:"matched"`

	// FullyMatched is true when the requested quantity was fully satisfied.
	FullyMatched bool `json:"fully_matched"`

	// Filled is the quantity satisfied by the matches.
	Filled uint64 `json:"filled"`

	// Remaining is the unsatisfied quantity.
	Remaining uint64 `json:"remaining"`
}

// FilterCandidatesByDenom keeps only candidates priced in the given denomination.
func FilterCandidatesByDenom(candidates []Candidate, denom string) []Candidate {
	if denom == "" {
		return candidates
	}
	out := make([]Candidate, 0, len(candidates))
	for _, c := range candidates {
		if c.Price.Denom == denom {
			out = append(out, c)
		}
	}
	return out
}

// RankCandidates returns candidates in deterministic priority order.
//
// Ordering (ascending unless noted):
//  1. price (lower first)
//  2. capacity fit (higher first)
//  3. provider score (higher first)
//  4. source priority per policy (native preferred)
//  5. kind priority per policy (listing preferred)
//  6. sequence (lower first)
//  7. provider address (lexicographic)
func RankCandidates(candidates []Candidate, policy ResolutionPolicy) []Candidate {
	ranked := make([]Candidate, 0, len(candidates))
	for _, c := range candidates {
		if err := c.Validate(); err != nil {
			continue
		}
		ranked = append(ranked, c)
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		return lessCandidate(ranked[i], ranked[j], policy)
	})
	return ranked
}

func lessCandidate(a, b Candidate, policy ResolutionPolicy) bool {
	if cmp := a.Price.Amount.BigInt().Cmp(b.Price.Amount.BigInt()); cmp != 0 {
		if a.Price.Denom == b.Price.Denom {
			return cmp < 0
		}
		return a.Price.Denom < b.Price.Denom
	}
	if a.CapacityFit != b.CapacityFit {
		return a.CapacityFit > b.CapacityFit
	}
	if a.ProviderScore != b.ProviderScore {
		return a.ProviderScore > b.ProviderScore
	}
	if policy.PreferNative {
		as, bs := sourceRank(a.Source), sourceRank(b.Source)
		if as != bs {
			return as < bs
		}
	}
	if policy.PreferListings && a.Kind != b.Kind {
		return a.Kind == CandidateKindListing
	}
	if a.Sequence != b.Sequence {
		return a.Sequence < b.Sequence
	}
	return a.ProviderAddress < b.ProviderAddress
}

func sourceRank(source OfferingSource) int {
	switch source.Effective() {
	case OfferingSourceNative:
		return 0
	case OfferingSourceWaldur:
		return 1
	default:
		return 2
	}
}

// SelectCandidates resolves a requested quantity against ranked candidates.
//
// When partial fills are disabled, the first candidate whose capacity covers the
// full quantity wins; if none can, the result is unmatched. When enabled, the
// highest-priority candidates are consumed until the quantity is filled.
func SelectCandidates(candidates []Candidate, quantity uint64, policy ResolutionPolicy) ResolutionResult {
	result := ResolutionResult{Remaining: quantity}
	if quantity == 0 {
		return result
	}

	ranked := RankCandidates(candidates, policy)
	if !policy.AllowPartialFill {
		for _, candidate := range ranked {
			if candidate.Capacity >= quantity {
				result.Matches = []Candidate{candidate}
				result.Matched = true
				result.FullyMatched = true
				result.Filled = quantity
				result.Remaining = 0
				return result
			}
		}
		return result
	}

	remaining := quantity
	for _, candidate := range ranked {
		if remaining == 0 {
			break
		}
		take := candidate.Capacity
		if take > remaining {
			take = remaining
		}
		if take == 0 {
			continue
		}
		match := candidate
		match.Capacity = take
		result.Matches = append(result.Matches, match)
		remaining -= take
	}
	result.Matched = len(result.Matches) > 0
	result.Filled = quantity - remaining
	result.Remaining = remaining
	result.FullyMatched = remaining == 0
	return result
}
