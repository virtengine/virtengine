package cli

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aclient "github.com/virtengine/virtengine/sdk/go/node/client"
)

const (
	ClientContextKey = sdk.ContextKey("client.context")
	ServerContextKey = sdk.ContextKey("server.context")
)

type ContextType string

const (
	ContextTypeClient         = ContextType("context-client")
	ContextTypeQueryClient    = ContextType("context-query-client")
	ContextTypeAddressCodec   = ContextType("address-codec")
	ContextTypeValidatorCodec = ContextType("validator-codec")
	ContextTypeRPCURI         = ContextType("rpc-uri")
	ContextTypeRPCClient      = ContextType("rpc-client")
)

var ErrContextValueNotSet = errors.New("context does not have value set")

func ClientFromContext(ctx context.Context) (aclient.Client, error) {
	val := ctx.Value(ContextTypeClient)
	if val == nil {
		return nil, fmt.Errorf("%w: %s", ErrContextValueNotSet, ContextTypeClient)
	}

	res, valid := val.(aclient.Client)
	if !valid {
		return nil, fmt.Errorf("invalid context value, expected \"aclient.Client\", actual \"%s\"", reflect.TypeOf(val))
	}

	return res, nil
}

func MustClientFromContext(ctx context.Context) aclient.Client {
	cl, err := ClientFromContext(ctx)
	if err != nil {
		panic(err.Error())
	}

	return cl
}

func LightClientFromContext(ctx context.Context) (aclient.LightClient, error) {
	val := ctx.Value(ContextTypeQueryClient)
	if val == nil {
		val = ctx.Value(ContextTypeClient)
		if val == nil {
			return nil, fmt.Errorf("%w: %s", ErrContextValueNotSet, ContextTypeClient)
		}
	}

	switch cl := val.(type) {
	case aclient.Client:
		return cl, nil
	case aclient.LightClient:
		return cl, nil
	default:
		return nil, fmt.Errorf("invalid context value. expected \"aclient.Client|aclient.LightClient\" actual %s", reflect.TypeOf(val).String())
	}
}

func MustLightClientFromContext(ctx context.Context) aclient.LightClient {
	cl, err := LightClientFromContext(ctx)
	if err != nil {
		panic(err.Error())
	}

	return cl
}

func MustAddressCodecFromContext(ctx context.Context) address.Codec {
	val := ctx.Value(ContextTypeAddressCodec)
	if val == nil {
		panic(fmt.Sprintf("%s: %s", ErrContextValueNotSet, ContextTypeAddressCodec))
	}

	res, valid := val.(address.Codec)
	if !valid {
		panic("invalid context value")
	}

	return res
}

func MustValidatorCodecFromContext(ctx context.Context) address.Codec {
	val := ctx.Value(ContextTypeValidatorCodec)
	if val == nil {
		panic(fmt.Sprintf("%s: %s", ErrContextValueNotSet, ContextTypeValidatorCodec))
	}

	res, valid := val.(address.Codec)
	if !valid {
		panic("invalid context value")
	}

	return res
}

