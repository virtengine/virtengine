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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/virtengine/virtengine/sdk/go/node/client/types"
	clientv1beta3 "github.com/virtengine/virtengine/sdk/go/node/client/v1beta3"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	"github.com/virtengine/virtengine/x/marketplace"
)

// GasSetting describes the gas configuration for on-chain submissions.
type GasSetting = types.GasSetting

// ChainCallbackSinkConfig configures on-chain callback submission.
type ChainCallbackSinkConfig struct {
	ChainID           string
	NodeURI           string
	GRPCEndpoint      string
	KeyName           string
	KeyringBackend    string
	KeyringDir        string
	KeyringPassphrase string
	GasSetting        GasSetting
	GasPrices         string
	Fees              string
	GasAdjustment     float64
	BroadcastTimeout  time.Duration
}

// ChainCallbackSink submits Waldur callbacks to the chain.
type ChainCallbackSink struct {
	client clientv1beta3.Client
	sender string
	opts   []clientv1beta3.BroadcastOption
}

// NewChainCallbackSink creates a chain-backed callback sink.
func NewChainCallbackSink(ctx context.Context, cfg ChainCallbackSinkConfig) (*ChainCallbackSink, error) {
	if cfg.ChainID == "" {
		return nil, fmt.Errorf("chain ID is required")
	}
	if cfg.NodeURI == "" {
		return nil, fmt.Errorf("node URI is required")
	}
	if cfg.GRPCEndpoint == "" {
		return nil, fmt.Errorf("gRPC endpoint is required")
	}
	if cfg.KeyName == "" {
		return nil, fmt.Errorf("key name is required")
	}

	encCfg := sdkutil.MakeEncodingConfig(marketplace.AppModuleBasic{})
	keyringDir := cfg.KeyringDir
	if keyringDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve home dir: %w", err)
		}
		keyringDir = filepath.Join(home, ".virtengine")
	}

	backend := cfg.KeyringBackend
	if backend == "" {
		backend = "test"
	}

	in := strings.NewReader(cfg.KeyringPassphrase + "\n" + cfg.KeyringPassphrase + "\n")
	kr, err := keyring.New(sdk.KeyringServiceName(), backend, keyringDir, in, encCfg.Codec)
	if err != nil {
		return nil, fmt.Errorf("init keyring: %w", err)
	}

	record, err := kr.Key(cfg.KeyName)
	if err != nil {
		return nil, fmt.Errorf("resolve key %q: %w", cfg.KeyName, err)
	}

	addr, err := record.GetAddress()
	if err != nil {
		return nil, fmt.Errorf("get key address: %w", err)
	}

	rpc, err := sdkclient.NewClientFromNode(cfg.NodeURI)
	if err != nil {
		return nil, fmt.Errorf("connect comet rpc: %w", err)
	}

	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	//nolint:staticcheck // grpc.DialContext kept for compatibility with existing connection flow.
	grpcConn, err := grpc.DialContext(dialCtx, cfg.GRPCEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial grpc: %w", err)
	}

	cctx := sdkclient.Context{}.
		WithChainID(cfg.ChainID).
		WithNodeURI(cfg.NodeURI).
		WithClient(rpc).
		WithGRPCClient(grpcConn).
		WithKeyring(kr).
		WithFromName(cfg.KeyName).
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
		return nil, fmt.Errorf("init tx client: %w", err)
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

	return &ChainCallbackSink{
		client: client,
		sender: addr.String(),
		opts:   opts,
	}, nil
}

// SenderAddress returns the bech32 sender address used for callbacks.
func (s *ChainCallbackSink) SenderAddress() string {
	return s.sender
}

// Submit submits the callback to the chain via MsgWaldurCallback.
func (s *ChainCallbackSink) Submit(ctx context.Context, callback *marketplacetypes.WaldurCallback) error {
	if callback == nil {
		return fmt.Errorf("callback is nil")
	}
	if callback.SignerID != s.sender {
		return fmt.Errorf("callback signer mismatch: %s != %s", callback.SignerID, s.sender)
	}

	// Serialize full callback to JSON so the chain can validate fields.
	payloadBytes, err := json.Marshal(callback)
	if err != nil {
		return fmt.Errorf("marshal callback: %w", err)
	}

	msg := marketplacetypes.NewMsgWaldurCallback(
		s.sender,
		string(callback.ActionType),
		callback.ChainEntityID,
		string(callback.ChainEntityType),
		string(payloadBytes),
		callback.Signature,
	)
	resp, err := s.client.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg}, s.opts...)
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
