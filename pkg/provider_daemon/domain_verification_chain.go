package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	types "github.com/virtengine/virtengine/sdk/go/node/client/types"
	clientv1beta3 "github.com/virtengine/virtengine/sdk/go/node/client/v1beta3"
	providertypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/x/provider"
	"github.com/virtengine/virtengine/x/provider/keeper"

	"github.com/virtengine/virtengine/pkg/security"
)

const testLiteral = "test"

type domainVerificationChainBackend interface {
	QueryDomainVerificationRecord(context.Context, sdk.AccAddress) (*keeper.DomainVerificationRecord, error)
	ConfirmDomainVerification(context.Context, sdk.AccAddress, string) error
}

type domainVerificationTxClient interface {
	BroadcastMsgs(context.Context, []sdk.Msg, ...clientv1beta3.BroadcastOption) (interface{}, error)
}

type rpcDomainVerificationBackend struct {
	storeQuery providerStoreQueryClient
	txClient   domainVerificationTxClient
	txOpts     []clientv1beta3.BroadcastOption
	timeout    time.Duration
}

func newRPCDomainVerificationBackend(ctx context.Context, cfg DomainVerificationCheckerConfig) (*rpcDomainVerificationBackend, error) {
	if cfg.CometRPC == "" {
		return nil, fmt.Errorf("comet RPC endpoint is required")
	}
	if cfg.GRPCEndpoint == "" {
		return nil, fmt.Errorf("gRPC endpoint is required")
	}
	if cfg.ChainID == "" {
		return nil, fmt.Errorf("chain ID is required")
	}
	if cfg.SignerKeyName == "" {
		return nil, fmt.Errorf("domain verification signer key name is required")
	}

	rpcClient, err := newProviderStoreQueryClient(cfg.CometRPC)
	if err != nil {
		return nil, err
	}

	txClient, txOpts, err := newDomainVerificationTxClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	timeout := cfg.VerificationTimeout
	if timeout <= 0 {
		timeout = DefaultVerificationTimeout
	}

	return &rpcDomainVerificationBackend{
		storeQuery: rpcClient,
		txClient:   txClient,
		txOpts:     txOpts,
		timeout:    timeout,
	}, nil
}

func newProviderStoreQueryClient(endpoint string) (providerStoreQueryClient, error) {
	rpcClient, err := rpchttp.New(endpoint, "/websocket")
	if err != nil {
		return nil, fmt.Errorf("failed to create provider store query client: %w", err)
	}
	return rpcClient, nil
}

func newDomainVerificationTxClient(
	ctx context.Context,
	cfg DomainVerificationCheckerConfig,
) (domainVerificationTxClient, []clientv1beta3.BroadcastOption, error) {
	encCfg := sdkutil.MakeEncodingConfig(provider.AppModuleBasic{})

	keyringDir := cfg.SignerKeyringDir
	if keyringDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, nil, fmt.Errorf("resolve home dir: %w", err)
		}
		keyringDir = filepath.Join(home, ".virtengine")
	}

	backend := cfg.SignerKeyringBackend
	if backend == "" {
		backend = testLiteral
	}

	in := strings.NewReader(cfg.SignerKeyringPassphrase + "\n" + cfg.SignerKeyringPassphrase + "\n")
	kr, err := keyring.New(sdk.KeyringServiceName(), backend, keyringDir, in, encCfg.Codec)
	if err != nil {
		return nil, nil, fmt.Errorf("init keyring: %w", err)
	}

	record, err := kr.Key(cfg.SignerKeyName)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve signer key %q: %w", cfg.SignerKeyName, err)
	}

	addr, err := record.GetAddress()
	if err != nil {
		return nil, nil, fmt.Errorf("get signer address: %w", err)
	}

	if cfg.ProviderAddress != "" && cfg.ProviderAddress != addr.String() {
		return nil, nil, fmt.Errorf(
			"domain verification signer address %s does not match provider address %s",
			addr.String(),
			cfg.ProviderAddress,
		)
	}

	rpc, err := sdkclient.NewClientFromNode(cfg.CometRPC)
	if err != nil {
		return nil, nil, fmt.Errorf("connect comet rpc: %w", err)
	}

	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	//nolint:staticcheck // grpc.DialContext kept for compatibility with existing provider-daemon connection flow.
	grpcConn, err := grpc.DialContext(
		dialCtx,
		cfg.GRPCEndpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(security.SecureTLSConfig())),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("dial grpc: %w", err)
	}

	cctx := sdkclient.Context{}.
		WithChainID(cfg.ChainID).
		WithNodeURI(cfg.CometRPC).
		WithClient(rpc).
		WithGRPCClient(grpcConn).
		WithKeyring(kr).
		WithFromName(cfg.SignerKeyName).
		WithFromAddress(addr).
		WithTxConfig(encCfg.TxConfig).
		WithCodec(encCfg.Codec).
		WithLegacyAmino(encCfg.Amino).
		WithInterfaceRegistry(encCfg.InterfaceRegistry).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithSignModeStr(types.SignModeDirect).
		WithBroadcastMode(clientv1beta3.BroadcastSync).
		WithOutput(io.Discard)

	client, err := clientv1beta3.NewClient(ctx, cctx)
	if err != nil {
		return nil, nil, fmt.Errorf("init provider tx client: %w", err)
	}

	opts := []clientv1beta3.BroadcastOption{
		clientv1beta3.WithSkipConfirm(true),
		clientv1beta3.WithBroadcastMode(clientv1beta3.BroadcastSync),
		clientv1beta3.WithResultCodeAsError(),
	}

	gasSetting := cfg.GasSetting
	if gasSetting.Gas == 0 && !gasSetting.Simulate {
		gasSetting = GasSetting{Simulate: true}
	}
	opts = append(opts, clientv1beta3.WithGas(gasSetting))

	if cfg.GasPrices != "" {
		opts = append(opts, clientv1beta3.WithGasPrices(cfg.GasPrices))
	}
	if cfg.Fees != "" {
		opts = append(opts, clientv1beta3.WithFees(cfg.Fees))
	}
	if cfg.GasAdjustment > 0 {
		opts = append(opts, clientv1beta3.WithGasAdjustment(cfg.GasAdjustment))
	}
	if cfg.BroadcastTimeout > 0 {
		opts = append(opts, clientv1beta3.WithBroadcastTimeout(cfg.BroadcastTimeout))
	}

	return client.Tx(), opts, nil
}

func (b *rpcDomainVerificationBackend) QueryDomainVerificationRecord(
	ctx context.Context,
	providerAddr sdk.AccAddress,
) (*keeper.DomainVerificationRecord, error) {
	value, err := queryProviderStoreValue(ctx, b.storeQuery, b.timeout, keeper.DomainVerificationKey(providerAddr))
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, nil
	}

	var record keeper.DomainVerificationRecord
	if err := json.Unmarshal(value, &record); err != nil {
		return nil, fmt.Errorf("decode domain verification record: %w", err)
	}

	return &record, nil
}

func (b *rpcDomainVerificationBackend) ConfirmDomainVerification(
	ctx context.Context,
	providerAddr sdk.AccAddress,
	proof string,
) error {
	msg := providertypes.NewMsgConfirmDomainVerification(providerAddr, proof)
	resp, err := b.txClient.BroadcastMsgs(ctx, []sdk.Msg{msg}, b.txOpts...)
	if err != nil {
		return err
	}

	switch res := resp.(type) {
	case *sdk.TxResponse:
		if res.Code != 0 {
			return fmt.Errorf("broadcast failed code=%d log=%s", res.Code, res.RawLog)
		}
	}

	return nil
}

func queryProviderStoreValue(
	ctx context.Context,
	client providerStoreQueryClient,
	timeout time.Duration,
	key []byte,
) ([]byte, error) {
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := client.ABCIQueryWithOptions(
		reqCtx,
		fmt.Sprintf("/store/%s/key", providertypes.StoreKey),
		tmbytes.HexBytes(key),
		rpcclient.ABCIQueryOptions{},
	)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("empty abci query response")
	}
	if result.Response.IsErr() {
		return nil, fmt.Errorf(
			"abci query failed with code %d: %s",
			result.Response.GetCode(),
			result.Response.GetLog(),
		)
	}
	return result.Response.GetValue(), nil
}
