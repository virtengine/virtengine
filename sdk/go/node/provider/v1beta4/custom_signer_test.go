package v1beta4

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/sdkutil"
)

func TestProviderCustomSignerFieldExtraction(t *testing.T) {
	msg := &MsgCreateProvider{
		Owner: sdk.AccAddress(bytes.Repeat([]byte{0x01}, 20)).String(),
	}

	v2 := protoadapt.MessageV2Of(msg)
	require.NotNil(t, v2)
	t.Logf("v2 type: %T", v2)
	t.Logf("v1 type: %T", protoadapt.MessageV1Of(v2))
	t.Logf("message name: %s", proto.MessageName(v2))

	var found bool
	for _, signer := range sdkutil.BuildCustomSigners() {
		if signer.MsgType != proto.MessageName(v2) {
			continue
		}
		found = true
		signers, err := signer.Fn(v2)
		require.NoError(t, err)
		require.Len(t, signers, 1)
		require.Equal(t, sdk.MustAccAddressFromBech32(msg.Owner).Bytes(), signers[0])
	}

	require.True(t, found, "expected custom signer for %s", proto.MessageName(v2))
}

func TestProviderSignerFallbackStructShape(t *testing.T) {
	msg := &MsgCreateProvider{
		Owner: sdk.AccAddress(bytes.Repeat([]byte{0x02}, 20)).String(),
	}

	v1 := protoadapt.MessageV1Of(protoadapt.MessageV2Of(msg))
	require.NotNil(t, v1)

	value := reflect.ValueOf(v1)
	require.Equal(t, reflect.Ptr, value.Kind())
	require.Equal(t, "*v1beta4.MsgCreateProvider", value.Type().String())

	structValue := value.Elem()
	require.Equal(t, reflect.Struct, structValue.Kind())

	var sawOwner bool
	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Type().Field(i)
		if field.Name == "Owner" {
			sawOwner = true
			require.Equal(t, msg.Owner, structValue.Field(i).String())
			require.Contains(t, field.Tag.Get("protobuf"), "name=owner")
		}
	}

	require.True(t, sawOwner)
}
