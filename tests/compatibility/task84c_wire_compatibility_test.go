package compatibility_test

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
)

func TestTask84COldWireFixturesDecode(t *testing.T) {
	fixtures := []struct {
		name    string
		hexWire string
		message proto.Message
		assert  func(*testing.T, proto.Message)
	}{
		{"market order", "2007", &marketv1beta5.Order{}, func(t *testing.T, message proto.Message) {
			order := message.(*marketv1beta5.Order)
			require.Equal(t, int64(7), order.CreatedAt)
			require.Empty(t, order.ReservationId)
		}},
		{"market bid", "2008", &marketv1beta5.Bid{}, func(t *testing.T, message proto.Message) {
			bid := message.(*marketv1beta5.Bid)
			require.Equal(t, int64(8), bid.CreatedAt)
			require.Empty(t, bid.ReservationId)
		}},
		{"market lease", "2009", &marketv1.Lease{}, func(t *testing.T, message proto.Message) {
			lease := message.(*marketv1.Lease)
			require.Equal(t, int64(9), lease.CreatedAt)
			require.Empty(t, lease.ReservationId)
		}},
		{"resource allocation", "780a", &resourcesv1.ResourceAllocation{}, func(t *testing.T, message proto.Message) {
			allocation := message.(*resourcesv1.ResourceAllocation)
			require.Equal(t, int64(10), allocation.BlockHeight)
			require.Empty(t, allocation.ReservationId)
		}},
		{"hpc job", "c0010b", &hpcv1.HPCJob{}, func(t *testing.T, message proto.Message) {
			job := message.(*hpcv1.HPCJob)
			require.Equal(t, int64(11), job.BlockHeight)
			require.Empty(t, job.ReservationId)
		}},
		{"resources genesis", "", &resourcesv1.GenesisState{}, func(t *testing.T, message proto.Message) {
			genesis := message.(*resourcesv1.GenesisState)
			require.False(t, genesis.CanonicalReservationsActive)
		}},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			bz, err := hex.DecodeString(fixture.hexWire)
			require.NoError(t, err)
			require.NoError(t, proto.Unmarshal(bz, fixture.message))
			fixture.assert(t, fixture.message)
		})
	}
}

func TestTask84CAdditiveFieldNumbers(t *testing.T) {
	require.Equal(t, 5, reservationFieldNumber(t, reflect.TypeOf(marketv1beta5.Order{})))
	require.Equal(t, 6, reservationFieldNumber(t, reflect.TypeOf(marketv1beta5.Bid{})))
	require.Equal(t, 7, reservationFieldNumber(t, reflect.TypeOf(marketv1.Lease{})))
	require.Equal(t, 16, reservationFieldNumber(t, reflect.TypeOf(resourcesv1.ResourceAllocation{})))
	require.Equal(t, 25, reservationFieldNumber(t, reflect.TypeOf(hpcv1.HPCJob{})))
	require.Equal(t, 5, protobufFieldNumber(t, reflect.TypeOf(resourcesv1.GenesisState{}), "CanonicalReservationsActive"))
	require.Equal(t, 6, protobufFieldNumber(t, reflect.TypeOf(resourcesv1.GenesisState{}), "ReservationEvents"))
	require.Equal(t, 2, protobufFieldNumber(t, reflect.TypeOf(resourcesv1.QueryReservationLineageRequest{}), "Pagination"))
	require.Equal(t, 3, protobufFieldNumber(t, reflect.TypeOf(resourcesv1.QueryReservationLineageResponse{}), "Pagination"))
}

func reservationFieldNumber(t *testing.T, typ reflect.Type) int {
	t.Helper()
	return protobufFieldNumber(t, typ, "ReservationId")
}

func protobufFieldNumber(t *testing.T, typ reflect.Type, fieldName string) int {
	t.Helper()
	field, ok := typ.FieldByName(fieldName)
	require.True(t, ok)
	match := regexp.MustCompile(`^[^,]+,(\d+),`).FindStringSubmatch(field.Tag.Get("protobuf"))
	require.Len(t, match, 2, fmt.Sprintf("missing reservation protobuf field tag for %s", typ.Name()))
	number, err := strconv.Atoi(match[1])
	require.NoError(t, err)
	return number
}

func TestTask84CTypeURLsRemainStable(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	marketv1beta5.RegisterInterfaces(registry)
	resolved, err := registry.Resolve("/virtengine.market.v1beta5.MsgCreateBid")
	require.NoError(t, err)
	require.IsType(t, &marketv1beta5.MsgCreateBid{}, resolved)

	marketplacetypes.RegisterInterfaces(registry)
	deprecated, err := registry.Resolve("/virtengine.marketplace.v1.MsgAcceptBid")
	require.NoError(t, err)
	require.IsType(t, &marketplacetypes.MsgAcceptBid{}, deprecated)
}
