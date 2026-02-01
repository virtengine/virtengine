package v1beta5_test

// import (
// 	"testing"
//
// 	"github.com/stretchr/testify/require"
//
// 	testutil "github.com/virtengine/virtengine/sdk/go/node/client/testutil/v1beta3"
//
// 	"github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
// )
//
// func TestBid_GSpecMatch_Valid(t *testing.T) {
// 	gspec := testutil.GroupSpec(t)
//
// 	rOffer := v1beta5.ResourceOfferFromRU(gspec.Resources)
//
// 	require.True(t, rOffer.MatchGSpec(gspec))
// }
//
// func TestBid_GSpecMatch_Valid2(t *testing.T) {
// 	gspec := testutil.GroupSpec(t)
//
// 	if len(gspec.Resources) == 1 {
// 		rl := testutil.ResourcesList(t, 2)
// 		rl[0].Count = 4
// 		gspec.Resources = append(gspec.Resources, rl...)
// 	}
//
// 	rOffer := v1beta5.ResourceOfferFromRU(gspec.Resources)
//
// 	require.True(t, rOffer.MatchGSpec(gspec))
// }
//
// func TestBid_GSpecMatch_InvalidCount(t *testing.T) {
// 	gspec := testutil.GroupSpec(t)
//
// 	if len(gspec.Resources) == 1 {
// 		rl := testutil.ResourcesList(t, 2)
// 		gspec.Resources = append(gspec.Resources, rl...)
// 	}
//
// 	rOffer := v1beta5.ResourceOfferFromRU(gspec.Resources)
//
// 	gspec.Resources[0].Count = 2
//
// 	require.False(t, rOffer.MatchGSpec(gspec))
// }

