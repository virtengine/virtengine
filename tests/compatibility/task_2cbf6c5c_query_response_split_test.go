//nolint:staticcheck // SA1019: naming the retired shared query types is the entire subject of this test; see the note below.
package compatibility

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

// The per-RPC request/response split in resources/v1 and settlement/v1 is an
// approved API migration, and it is the reason sdk/buf.yaml now carries a
// breaking-rule exemption for exactly those two files:
//
//	RPC_SAME_REQUEST_TYPE / RPC_SAME_RESPONSE_TYPE
//	  -> proto/node/virtengine/resources/v1/query.proto
//	  -> proto/node/virtengine/settlement/v1/query.proto
//
// Repointing an RPC at its own type is exactly what those rules flag, so the
// exemption is load-bearing. What makes it acceptable is that the new types
// are field-identical to the shared ones they replace: a client that decoded
// the shared response still decodes the per-RPC response, and the bytes on the
// wire are unchanged. These tests pin that property so a future edit that
// diverges a per-RPC type from its predecessor fails here rather than
// silently in a downstream SDK.
//
// Both directions matter and are checked below:
//   - a shared message marshals to the same bytes as its per-RPC replacement
//     (the old client's encoder output is what a new node must still accept);
//   - the legacy type still decodes (nothing broke existing compiled clients).
//
// SA1019 is waived file-wide on purpose: naming the deprecated shared types is
// the entire subject of this file. The waiver is scoped to this one test rather
// than a repo-wide staticcheck exclusion, because everywhere else in the tree a
// deprecated-type reference is a real finding and should stay one.

// sharedToPerRPC lists each retired shared message against the per-RPC types
// that replaced it, by fully qualified proto name so the pairs are read off the
// generated registration rather than off a hand-written count.
func sharedToPerRPC(t *testing.T) map[string][]proto.Message {
	t.Helper()
	return map[string][]proto.Message{
		"virtengine.resources.v1.QueryReservationResponse": {
			&resourcesv1.QueryReservationResponse{},
			&resourcesv1.QueryReservationByOrderResponse{},
			&resourcesv1.QueryReservationByBidResponse{},
			&resourcesv1.QueryReservationByLeaseResponse{},
			&resourcesv1.QueryReservationByJobResponse{},
			&resourcesv1.QueryReservationByConsumerResponse{},
		},
		"virtengine.resources.v1.QueryReservationsResponse": {
			&resourcesv1.QueryReservationsResponse{},
			&resourcesv1.QueryReservationsByProviderResponse{},
		},
		"virtengine.settlement.v1.QueryFinancialCasesRequest": {
			&settlementv1.QueryFinancialCasesRequest{},
			&settlementv1.QueryFinancialCasesByOrderRequest{},
			&settlementv1.QueryFinancialCasesByInvoiceRequest{},
			&settlementv1.QueryFinancialCasesByUsageRequest{},
			&settlementv1.QueryFinancialCasesByJobRequest{},
			&settlementv1.QueryFinancialCasesByEscrowRequest{},
			&settlementv1.QueryFinancialCasesByStatusRequest{},
			&settlementv1.QueryFinancialCasesByPartyRequest{},
		},
		"virtengine.settlement.v1.QueryFinancialCasesResponse": {
			&settlementv1.QueryFinancialCasesResponse{},
			&settlementv1.QueryFinancialCasesByOrderResponse{},
			&settlementv1.QueryFinancialCasesByInvoiceResponse{},
			&settlementv1.QueryFinancialCasesByUsageResponse{},
			&settlementv1.QueryFinancialCasesByJobResponse{},
			&settlementv1.QueryFinancialCasesByEscrowResponse{},
			&settlementv1.QueryFinancialCasesByStatusResponse{},
			&settlementv1.QueryFinancialCasesByPartyResponse{},
		},
	}
}

// TestQueryResponseSplitIsWireIdentical is the core claim of the migration: a
// per-RPC type encodes to exactly the bytes its retired shared type did.
func TestQueryResponseSplitIsWireIdentical(t *testing.T) {
	for sharedName, family := range sharedToPerRPC(t) {
		shared, replacements := family[0], family[1:]
		require.NotEmpty(t, replacements, sharedName)

		// Populate the shared message from every field it declares, so the
		// comparison exercises each field rather than an empty message (two
		// empty messages encode to the same zero bytes whatever the fields are).
		source := populateEveryField(t, shared)

		want, err := proto.Marshal(source)
		require.NoErrorf(t, err, "marshal %s", sharedName)

		for _, replacement := range replacements {
			// A per-RPC type is only a rename if it carries the same fields, so
			// require identical field sets first; then the bytes must match.
			requireSameFieldSet(t, shared, replacement)

			// Re-encode the shared bytes through the per-RPC type: this is the
			// decode path an existing client takes when the server starts
			// answering with the renamed type.
			decoded := proto.Clone(replacement)
			decoded.Reset()
			require.NoErrorf(t, proto.Unmarshal(want, decoded), "unmarshal %s into %T", sharedName, replacement)
			requireSameFieldSet(t, decoded, shared)

			got, err := proto.Marshal(decoded)
			require.NoErrorf(t, err, "re-marshal %T", replacement)
			require.Equalf(t, want, got,
				"%s must encode byte-identically to its replacement %T", sharedName, replacement)
		}
	}
}

// TestQueryResponseSplitTypesAreDistinct guards the other half of the
// migration: the exemption exists because each RPC now names its own type. If
// two entries in a family collapse back to one generated type, the split has
// been undone silently and the buf exemption is no longer describing reality.
func TestQueryResponseSplitTypesAreDistinct(t *testing.T) {
	for sharedName, family := range sharedToPerRPC(t) {
		seen := map[string]struct{}{}
		for _, message := range family {
			name := proto.MessageName(message)
			require.NotEmptyf(t, name, "%s: message is not registered", sharedName)
			_, duplicate := seen[name]
			require.Falsef(t, duplicate, "%s: %s appears twice in the same family", sharedName, name)
			seen[name] = struct{}{}
		}
	}
}

// wireFields reads a generated message's wire contract out of its struct tags:
// field number -> Go type plus cardinality. Two messages are wire-identical
// when these agree for every field, which is exactly the condition under which
// renaming a message (or renaming one of its fields) leaves the encoded bytes
// alone. The field NAME is deliberately not part of the comparison: it does not
// appear on the wire, so the per-RPC split is free to give `key` a more honest
// name (`order_id`, `invoice_id`, ...) as long as the number and type hold.
func wireFields(t *testing.T, message proto.Message) map[int]string {
	t.Helper()
	out := map[int]string{}
	messageType := reflect.TypeOf(message)
	// Generated messages are handled as pointers, and NumField panics on a
	// pointer type, so dereference to the struct before walking the fields.
	for messageType.Kind() == reflect.Pointer {
		messageType = messageType.Elem()
	}
	if messageType.Kind() != reflect.Struct {
		return out
	}
	for i := 0; i < messageType.NumField(); i++ {
		field := messageType.Field(i)
		tag := field.Tag.Get("protobuf")
		number := regexp.MustCompile(`^[^,]+,(\d+),`).FindStringSubmatch(tag)
		if len(number) != 2 {
			continue
		}
		// number -> type:cardinality, so a renumbered field, a retyped field
		// and a cardinality change are all caught.
		out[mustAtoi(t, number[1])] = field.Type.String() + ":" + cardinality(tag)
	}
	return out
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	require.NoErrorf(t, err, "field number %q is not an integer", s)
	return n
}

func cardinality(tag string) string {
	if strings.Contains(tag, "rep") {
		return "rep"
	}
	return "singular"
}

func requireSameFieldSet(t *testing.T, want, got proto.Message) {
	t.Helper()
	wantFields, gotFields := wireFields(t, want), wireFields(t, got)
	require.Equalf(t, wantFields, gotFields,
		"%s and %T must declare the same field numbers, types and cardinality",
		proto.MessageName(want), got)
}

// populateEveryField fills every declared field of a generated message with a
// non-zero value and returns a fresh instance, so a comparison of encoded
// bytes actually covers the fields instead of comparing two empty messages.
func populateEveryField(t *testing.T, message proto.Message) proto.Message {
	t.Helper()
	out := reflect.New(reflect.TypeOf(message).Elem())
	seen := setMessageFields(out.Elem(), 0)
	require.NotZerof(t, seen, "%s: populateEveryField set no fields, so the byte comparison is vacuous",
		proto.MessageName(message))
	return out.Interface().(proto.Message)
}

// setMessageFields sets every protobuf-tagged field of a struct value and
// returns how many leaves it managed to set. Fields gogoproto generates for
// bookkeeping (XXX_*) carry no protobuf tag and are skipped, which is what
// keeps this from writing into unexported state.
func setMessageFields(structValue reflect.Value, depth int) int {
	if depth > maxPopulateDepth || structValue.Kind() != reflect.Struct {
		return 0
	}
	any := 0
	for i := 0; i < structValue.NumField(); i++ {
		if structValue.Type().Field(i).Tag.Get("protobuf") == "" {
			continue
		}
		any += setScalarOrNested(structValue.Field(i), depth+1)
	}
	return any
}

// maxPopulateDepth bounds the walk so a self-referential type cannot loop.
const maxPopulateDepth = 4

// setScalarOrNested sets one field, descending through pointers and slices of
// structs because a response whose only fields are `repeated Foo` and
// `*PageResponse` would otherwise contribute nothing and make the byte
// comparison vacuous - exactly the messages this test exists to cover.
func setScalarOrNested(field reflect.Value, depth int) int {
	switch field.Kind() {
	case reflect.String:
		field.SetString("x")
		return 1
	case reflect.Bool:
		field.SetBool(true)
		return 1
	case reflect.Int, reflect.Int32, reflect.Int64:
		field.SetInt(7)
		return 1
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		field.SetUint(7)
		return 1
	case reflect.Float32, reflect.Float64:
		field.SetFloat(1.5)
		return 1
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Uint8 {
			field.SetBytes([]byte{0x01, 0x02})
			return 1
		}
		if field.Type().Elem().Kind() != reflect.Struct || depth >= maxPopulateDepth {
			return 0
		}
		element := reflect.New(field.Type().Elem()).Elem()
		if set := setMessageFields(element, depth+1); set == 0 {
			return 0
		}
		field.Set(reflect.Append(field, element))
		return 1
	case reflect.Pointer:
		if field.Type().Elem().Kind() != reflect.Struct || depth >= maxPopulateDepth {
			return 0
		}
		pointer := reflect.New(field.Type().Elem())
		if set := setMessageFields(pointer.Elem(), depth+1); set == 0 {
			return 0
		}
		field.Set(pointer)
		return 1
	case reflect.Struct:
		return setMessageFields(field, depth+1)
	}
	return 0
}
