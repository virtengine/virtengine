package rest

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/virtengine/virtengine/provider"
	gwutils "github.com/virtengine/virtengine/provider/gateway/utils"
	ctypes "github.com/virtengine/virtengine/x/cert/types"
)

func NewServer(
	ctx context.Context,
	log log.Logger,
	pclient provider.Client,
	cquery ctypes.QueryClient,
	address string,
	pid sdk.Address,
	certs []tls.Certificate) (*http.Server, error) {

	srv := &http.Server{
		Addr:    address,
		Handler: newRouter(log, pid, pclient),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	var err error

	srv.TLSConfig, err = gwutils.NewServerTLSConfig(context.Background(), certs, cquery)
	if err != nil {
		return nil, err
	}

	return srv, nil
}
