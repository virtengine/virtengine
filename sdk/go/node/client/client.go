package client

import (
	"context"
	"errors"
	"fmt"

	cmtrpc "github.com/cometbft/cometbft/rpc/core"
	cmjclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	cmtrpcsrv "github.com/cometbft/cometbft/rpc/jsonrpc/server"
	cmtrpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"

	sdkclient "github.com/cosmos/cosmos-sdk/client"

	cltypes "github.com/virtengine/virtengine/sdk/go/node/client/types"
	"github.com/virtengine/virtengine/sdk/go/node/client/v1beta3"
)

var (
	ErrDetectClientVersion  = errors.New("chain-sdk: unable detect client version")
	ErrUnknownClientVersion = errors.New("chain-sdk: unknown client version")
)

const (
	VersionV1beta3 = "v1beta3"
)

func init() {
	// register akash api routes
	cmtrpc.Routes["akash"] = cmtrpcsrv.NewRPCFunc(RPCAkash, "")
}

// SetupFn defines a function that takes a parameter, ideally a Client or QueryClient.
// These functions must validate the client and make it accessible.
type SetupFn func(interface{}) error

func queryClientInfo(ctx context.Context, cctx sdkclient.Context) (*Akash, error) {
	result := new(Akash)
	var err error

	if !cctx.Offline {
		if cctx.Client != nil {
			switch rpc := cctx.Client.(type) {
			case RPCClient:
				if result, err = rpc.Akash(ctx); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unsupported RPC client [%T]", rpc)
			}
		} else {
			rpc, err := cmjclient.New(cctx.NodeURI)
			if err != nil {
				return nil, err
			}

			params := make(map[string]interface{})
			if _, err = rpc.Call(ctx, "akash", params, result); err != nil {
				return nil, err
			}
		}

		if result.ClientInfo.ApiVersion == "" {
			return nil, ErrDetectClientVersion
		}
	} else {
		result.ClientInfo = ClientInfo{ApiVersion: VersionV1beta3}
	}

	return result, nil
}

// DiscoverClient queries an RPC node to get the version of the client and executes a SetupFn function
// passing a new versioned Client instance as parameter.
// If any error occurs when calling the RPC node or the Cosmos SDK client Context is set to offline the default value of
// DefaultClientAPIVersion will be used.
// An error is returned if client discovery is not successful.
func DiscoverClient(ctx context.Context, cctx sdkclient.Context, setup SetupFn, opts ...cltypes.ClientOption) error {
	result, err := queryClientInfo(ctx, cctx)
	if err != nil {
		return err
	}

	var cl interface{}

	switch result.ClientInfo.ApiVersion {
	case VersionV1beta3:
		cl, err = v1beta3.NewClient(ctx, cctx, opts...)
	default:
		err = ErrUnknownClientVersion
	}

	if err != nil {
		return err
	}

	if err = setup(cl); err != nil {
		return err
	}

	return nil
}

func DiscoverLightClient(ctx context.Context, cctx sdkclient.Context, setup SetupFn) error {
	result, err := queryClientInfo(ctx, cctx)
	if err != nil {
		return err
	}

	var cl interface{}

	switch result.ClientInfo.ApiVersion {
	case VersionV1beta3:
		cl, err = v1beta3.NewLightClient(cctx)
	default:
		err = ErrUnknownClientVersion
	}

	if err != nil {
		return err
	}

	if err = setup(cl); err != nil {
		return err
	}

	return nil
}

// DiscoverQueryClient queries an RPC node to get the version of the client and executes a SetupFn function
// passing a new versioned QueryClient instance as parameter.
// If any error occurs when calling the RPC node or the Cosmos SDK client Context is set to offline the default value of
// DefaultClientAPIVersion will be used.
// An error is returned if client discovery is not successful.
func DiscoverQueryClient(ctx context.Context, cctx sdkclient.Context, setup SetupFn) error {
	result, err := queryClientInfo(ctx, cctx)

	var cl interface{}

	switch result.ClientInfo.ApiVersion {
	case VersionV1beta3:
		cl = v1beta3.NewQueryClient(cctx)
	default:
		err = ErrUnknownClientVersion
	}

	if err != nil {
		return err
	}

	if err = setup(cl); err != nil {
		return err
	}

	return nil
}

func RPCAkash(_ *cmtrpctypes.Context) (*Akash, error) {
	result := &Akash{
		ClientInfo: ClientInfo{
			ApiVersion: "v1beta3",
		},
	}

	return result, nil
}

