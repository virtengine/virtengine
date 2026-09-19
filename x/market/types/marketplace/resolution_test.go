package marketplace

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func listingCandidate(t *testing.T, provider string, seq uint64, price int64, capacity uint64, score uint32, fit uint64, source OfferingSource) Candidate {
	t.Helper()
	id := OfferingID{ProviderAddress: provider, Sequence: seq}
	return Candidate{
		Kind:            CandidateKindListing,
		OfferingID:      &id,
		ProviderAddress: provider,
		Price:           sdk.NewInt64Coin("uvirt", price),
		Capacity:        capacity,
		ProviderScore:   score,
		CapacityFit:     fit,
		Sequence:        seq,
		Source:          source,
	}
}

func bidCandidate(t *testing.T, provider string, seq uint64, price int64, capacity uint64, score uint32, fit uint64) Candidate {
	t.Helper()
	orderID := OrderID{CustomerAddress: "ve1cust", Sequence: 1}
	id := BidID{OrderID: orderID, ProviderAddress: provider, Sequence: seq}
	return Candidate{
		Kind:            CandidateKindBid,
		BidID:           &id,
		ProviderAddress: provider,
		Price:           sdk.NewInt64Coin("uvirt", price),
		Capacity:        capacity,
		ProviderScore:   score,
		CapacityFit:     fit,
		Sequence:        seq,
	}
}

func TestRankCandidatesByPrice(t *testing.T) {
	candidates := []Candidate{
		listingCandidate(t, "ve1b", 1, 300, 10, 50, 1, OfferingSourceNative),
		listingCandidate(t, "ve1a", 2, 100, 10, 50, 1, OfferingSourceNative),
		listingCandidate(t, "ve1c", 3, 200, 10, 50, 1, OfferingSourceNative),
	}

	ranked := RankCandidates(candidates, DefaultResolutionPolicy())
	require.Len(t, ranked, 3)
	require.Equal(t, int64(100), ranked[0].Price.Amount.Int64())
	require.Equal(t, int64(200), ranked[1].Price.Amount.Int64())
	require.Equal(t, int64(300), ranked[2].Price.Amount.Int64())
}

func TestRankCandidatesTieBreakers(t *testing.T) {
	candidates := []Candidate{
		listingCandidate(t, "ve1waldur", 1, 100, 10, 50, 1, OfferingSourceWaldur),
		listingCandidate(t, "ve1native", 2, 100, 10, 50, 1, OfferingSourceNative),
	}

	ranked := RankCandidates(candidates, DefaultResolutionPolicy())
	require.Equal(t, "ve1native", ranked[0].ProviderAddress, "native preferred on price tie")
}

func TestRankCandidatesProviderScore(t *testing.T) {
	candidates := []Candidate{
		listingCandidate(t, "ve1low", 1, 100, 10, 10, 1, OfferingSourceNative),
		listingCandidate(t, "ve1high", 2, 100, 10, 90, 1, OfferingSourceNative),
	}

	ranked := RankCandidates(candidates, ResolutionPolicy{PreferNative: true, PreferListings: true})
	require.Equal(t, "ve1high", ranked[0].ProviderAddress, "higher provider score wins")
}

func TestRankCandidatesRejectsInvalid(t *testing.T) {
	invalid := Candidate{Kind: CandidateKindListing, ProviderAddress: "ve1x", Price: sdk.NewInt64Coin("uvirt", 1)}
	candidates := []Candidate{
		invalid,
		listingCandidate(t, "ve1a", 1, 100, 10, 50, 1, OfferingSourceNative),
	}
	ranked := RankCandidates(candidates, DefaultResolutionPolicy())
	require.Len(t, ranked, 1)
}

func TestSelectCandidatesAllOrNothing(t *testing.T) {
	candidates := []Candidate{
		listingCandidate(t, "ve1small", 1, 50, 1, 50, 1, OfferingSourceNative),
		listingCandidate(t, "ve1big", 2, 100, 5, 50, 1, OfferingSourceNative),
	}

	// Quantity 3 cannot be served by the cheaper capacity-1 listing; with partial
	// fills disabled the engine must either find one listing that fits or match none.
	result := SelectCandidates(candidates, 3, ResolutionPolicy{})
	require.True(t, result.Matched)
	require.True(t, result.FullyMatched)
	require.Len(t, result.Matches, 1)
	require.Equal(t, "ve1big", result.Matches[0].ProviderAddress)
}

func TestSelectCandidatesUnmatchedWhenInsufficient(t *testing.T) {
	candidates := []Candidate{
		listingCandidate(t, "ve1small", 1, 50, 1, 50, 1, OfferingSourceNative),
	}
	result := SelectCandidates(candidates, 5, ResolutionPolicy{})
	require.False(t, result.Matched)
	require.Equal(t, uint64(5), result.Remaining)
}

func TestSelectCandidatesPartialFill(t *testing.T) {
	candidates := []Candidate{
		listingCandidate(t, "ve1a", 1, 50, 2, 50, 1, OfferingSourceNative),
		listingCandidate(t, "ve1b", 2, 60, 2, 50, 1, OfferingSourceNative),
		listingCandidate(t, "ve1c", 3, 70, 2, 50, 1, OfferingSourceNative),
	}
	policy := ResolutionPolicy{AllowPartialFill: true}

	result := SelectCandidates(candidates, 5, policy)
	require.True(t, result.FullyMatched)
	require.Equal(t, uint64(5), result.Filled)
	require.Equal(t, uint64(0), result.Remaining)
	require.Len(t, result.Matches, 3)
	require.Equal(t, uint64(2), result.Matches[0].Capacity)
	require.Equal(t, uint64(2), result.Matches[1].Capacity)
	require.Equal(t, uint64(1), result.Matches[2].Capacity, "last fill consumes only the remainder")
}

func TestSelectCandidatesMixBidAndListing(t *testing.T) {
	candidates := []Candidate{
		listingCandidate(t, "ve1listing", 1, 100, 1, 50, 1, OfferingSourceNative),
		bidCandidate(t, "ve1bid", 2, 100, 1, 50, 1),
	}
	// Listing preferred on an exact tie under default policy.
	result := SelectCandidates(candidates, 1, DefaultResolutionPolicy())
	require.True(t, result.Matched)
	require.Len(t, result.Matches, 1)
	require.Equal(t, CandidateKindListing, result.Matches[0].Kind)
}

func TestFilterCandidatesByDenom(t *testing.T) {
	candidates := []Candidate{
		listingCandidate(t, "ve1a", 1, 100, 1, 50, 1, OfferingSourceNative),
		{
			Kind:            CandidateKindListing,
			OfferingID:      &OfferingID{ProviderAddress: "ve1b", Sequence: 2},
			ProviderAddress: "ve1b",
			Price:           sdk.NewInt64Coin("uatom", 100),
			Capacity:        1,
		},
	}
	filtered := FilterCandidatesByDenom(candidates, "uvirt")
	require.Len(t, filtered, 1)
	require.Equal(t, "ve1a", filtered[0].ProviderAddress)
}
