package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	sdkclient "github.com/cosmos/cosmos-sdk/client"

	cmtrpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"

	aclient "github.com/virtengine/virtengine/sdk/go/node/client"
	cltypes "github.com/virtengine/virtengine/sdk/go/node/client/types"
	"github.com/virtengine/virtengine/sdk/go/node/client/v1beta3"
)

var (
	ErrInvalidClient = errors.New("invalid client")
)

func DiscoverQueryClient(ctx context.Context, cctx sdkclient.Context) (aclient.QueryClient, error) {
	var cl v1beta3.QueryClient
	err := aclient.DiscoverQueryClient(ctx, cctx, func(i interface{}) error {
		var valid bool

		if cl, valid = i.(v1beta3.QueryClient); !valid {
			return fmt.Errorf("%w: expected %s, actual %T", ErrInvalidClient, reflect.TypeOf((*v1beta3.QueryClient)(nil)).Elem(), i)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return cl, nil
}

func DiscoverLightClient(ctx context.Context, cctx sdkclient.Context) (aclient.LightClient, error) {
	var cl v1beta3.LightClient
	err := aclient.DiscoverLightClient(ctx, cctx, func(i interface{}) error {
		var valid bool

		if cl, valid = i.(v1beta3.LightClient); !valid {
			return fmt.Errorf("%w: expected %s, actual %T", ErrInvalidClient, reflect.TypeOf((*v1beta3.LightClient)(nil)).Elem(), i)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return cl, nil
}

func DiscoverClient(ctx context.Context, cctx sdkclient.Context, opts ...cltypes.ClientOption) (aclient.Client, error) {
	var cl v1beta3.Client

	setupFn := func(i interface{}) error {
		var valid bool

		if cl, valid = i.(v1beta3.Client); !valid {
			return fmt.Errorf("%w: expected %s, actual %T", ErrInvalidClient, reflect.TypeOf((*v1beta3.Client)(nil)).Elem(), i)
		}

		return nil
	}

	err := aclient.DiscoverClient(ctx, cctx, setupFn, opts...)

	if err != nil {
		return nil, err
	}

	return cl, nil
}

func RPCAkash(_ *cmtrpctypes.Context) (*aclient.Akash, error) {
	result := &aclient.Akash{
		ClientInfo: aclient.ClientInfo{
			ApiVersion: "v1beta3",
		},
	}

	return result, nil
}

