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

	"cosmossdk.io/log"
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/virtengine/virtengine/pkg/security"
	"github.com/virtengine/virtengine/sdk/go/node/client/types"
	clientv1beta3 "github.com/virtengine/virtengine/sdk/go/node/client/v1beta3"
	supportv1 "github.com/virtengine/virtengine/sdk/go/node/support/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/x/support"
	supporttypes "github.com/virtengine/virtengine/x/support/types"
)

type supportTxClient interface {
	BroadcastMsgs(context.Context, []sdk.Msg, ...clientv1beta3.BroadcastOption) (interface{}, error)
}

type supportExternalRefStore struct {
	ResourceID       string `json:"resource_id"`
	ResourceType     string `json:"resource_type"`
	ExternalSystem   string `json:"external_system"`
	ExternalTicketID string `json:"external_ticket_id"`
	ExternalURL      string `json:"external_url,omitempty"`
	CreatedAt        int64  `json:"created_at"`
	CreatedBy        string `json:"created_by"`
	UpdatedAt        int64  `json:"updated_at"`
}

type rpcSupportChainWriter struct {
	sender     string
	storeQuery providerStoreQueryClient
	txClient   supportTxClient
	txOpts     []clientv1beta3.BroadcastOption
	timeout    time.Duration
	logger     log.Logger
	grpcConn   *grpc.ClientConn
}

func newSupportChainWriter(ctx context.Context, cfg SupportServiceConfig, logger log.Logger) (SupportChainWriter, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	if cfg.ChainID == "" {
		return nil, fmt.Errorf("support chain ID is required")
	}
	if cfg.CometRPC == "" {
		return nil, fmt.Errorf("support comet RPC endpoint is required")
	}
	if cfg.GRPCEndpoint == "" {
		return nil, fmt.Errorf("support gRPC endpoint is required")
	}
	if strings.TrimSpace(cfg.SignerKeyName) == "" {
		return nil, fmt.Errorf("support signer key name is required")
	}

	storeQuery, err := newProviderStoreQueryClient(cfg.CometRPC)
	if err != nil {
		return nil, err
	}

	txClient, sender, txOpts, grpcConn, err := newSupportTxClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &rpcSupportChainWriter{
		sender:     sender,
		storeQuery: storeQuery,
		txClient:   txClient,
		txOpts:     txOpts,
		timeout:    timeout,
		logger:     logger,
		grpcConn:   grpcConn,
	}, nil
}

func newSupportTxClient(
	ctx context.Context,
	cfg SupportServiceConfig,
) (supportTxClient, string, []clientv1beta3.BroadcastOption, *grpc.ClientConn, error) {
	encCfg := sdkutil.MakeEncodingConfig(support.AppModuleBasic{})

	keyringDir := cfg.SignerKeyringDir
	if keyringDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, "", nil, nil, fmt.Errorf("resolve home dir: %w", err)
		}
		keyringDir = filepath.Join(home, ".virtengine")
	}

	backend := cfg.SignerKeyringBackend
	if backend == "" {
		backend = "test"
	}

	in := strings.NewReader(cfg.SignerKeyringPassphrase + "\n" + cfg.SignerKeyringPassphrase + "\n")
	kr, err := keyring.New(sdk.KeyringServiceName(), backend, keyringDir, in, encCfg.Codec)
	if err != nil {
		return nil, "", nil, nil, fmt.Errorf("init support keyring: %w", err)
	}

	record, err := kr.Key(cfg.SignerKeyName)
	if err != nil {
		return nil, "", nil, nil, fmt.Errorf("resolve support signer key %q: %w", cfg.SignerKeyName, err)
	}

	addr, err := record.GetAddress()
	if err != nil {
		return nil, "", nil, nil, fmt.Errorf("get support signer address: %w", err)
	}

	rpc, err := sdkclient.NewClientFromNode(cfg.CometRPC)
	if err != nil {
		return nil, "", nil, nil, fmt.Errorf("connect support comet rpc: %w", err)
	}

	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	//nolint:staticcheck // grpc.DialContext is retained for compatibility with sdk client wiring.
	grpcConn, err := grpc.DialContext(
		dialCtx,
		cfg.GRPCEndpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(security.SecureTLSConfig())),
	)
	if err != nil {
		return nil, "", nil, nil, fmt.Errorf("dial support grpc: %w", err)
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
		_ = grpcConn.Close()
		return nil, "", nil, nil, fmt.Errorf("init support tx client: %w", err)
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

	return client.Tx(), addr.String(), opts, grpcConn, nil
}

func (w *rpcSupportChainWriter) SenderAddress() string {
	if w == nil {
		return ""
	}
	return w.sender
}

func (w *rpcSupportChainWriter) Close() error {
	if w == nil || w.grpcConn == nil {
		return nil
	}
	return w.grpcConn.Close()
}

func (w *rpcSupportChainWriter) UpdateSupportRequest(ctx context.Context, msg *SupportUpdateRequest) error {
	if msg == nil {
		return nil
	}
	request := &supportv1.MsgUpdateSupportRequest{
		Sender:        w.sender,
		TicketId:      strings.TrimSpace(msg.TicketID),
		Status:        strings.TrimSpace(msg.Status),
		AssignedAgent: strings.TrimSpace(msg.AssignedAgent),
	}
	if len(msg.Metadata) > 0 {
		request.PublicMetadata = cloneSupportMetadata(msg.Metadata)
	}
	return w.broadcast(ctx, request)
}

func (w *rpcSupportChainWriter) AddSupportResponse(ctx context.Context, msg *SupportAddResponse) error {
	if msg == nil {
		return nil
	}
	if msg.Payload == nil {
		return fmt.Errorf("support response payload is required")
	}
	msg.Payload.EnsureEnvelopeHash()
	request := &supportv1.MsgAddSupportResponse{
		Sender:   w.sender,
		TicketId: strings.TrimSpace(msg.TicketID),
		Payload:  supporttypes.ToProtoEncryptedSupportPayload(*msg.Payload),
	}
	return w.broadcast(ctx, request)
}

func (w *rpcSupportChainWriter) RegisterExternalTicket(ctx context.Context, msg *SupportRegisterExternal) error {
	if msg == nil {
		return nil
	}

	resourceType := supporttypes.ResourceType(strings.TrimSpace(msg.ResourceType))
	resourceID := strings.TrimSpace(msg.ResourceID)
	externalSystem := supporttypes.ExternalSystem(strings.TrimSpace(msg.ExternalSystem))
	externalTicketID := strings.TrimSpace(msg.ExternalTicketID)
	externalURL := strings.TrimSpace(msg.ExternalURL)

	existing, err := w.queryExternalTicketRef(ctx, resourceType, resourceID)
	if err != nil {
		return err
	}
	if existing == nil {
		request := &supportv1.MsgRegisterExternalTicket{
			Sender:           w.sender,
			ResourceId:       resourceID,
			ResourceType:     string(resourceType),
			ExternalSystem:   string(externalSystem),
			ExternalTicketId: externalTicketID,
			ExternalUrl:      externalURL,
		}
		return w.broadcast(ctx, request)
	}
	if existing.ExternalSystem != externalSystem {
		return fmt.Errorf(
			"external ticket %s/%s already registered with system %s",
			resourceType,
			resourceID,
			existing.ExternalSystem,
		)
	}
	if existing.CreatedBy != "" && existing.CreatedBy != w.sender {
		return fmt.Errorf(
			"external ticket %s/%s owned by %s, signer %s cannot update it",
			resourceType,
			resourceID,
			existing.CreatedBy,
			w.sender,
		)
	}
	if existing.ExternalTicketID == externalTicketID && existing.ExternalURL == externalURL {
		return nil
	}

	request := &supportv1.MsgUpdateExternalTicket{
		Sender:           w.sender,
		ResourceId:       resourceID,
		ResourceType:     string(resourceType),
		ExternalTicketId: externalTicketID,
		ExternalUrl:      externalURL,
	}
	return w.broadcast(ctx, request)
}

func (w *rpcSupportChainWriter) queryExternalTicketRef(
	ctx context.Context,
	resourceType supporttypes.ResourceType,
	resourceID string,
) (*supporttypes.ExternalTicketRef, error) {
	value, err := querySupportStoreValue(ctx, w.storeQuery, w.timeout, supporttypes.ExternalRefKey(resourceType, resourceID))
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, nil
	}

	var stored supportExternalRefStore
	if err := json.Unmarshal(value, &stored); err != nil {
		return nil, fmt.Errorf("decode support external ticket ref: %w", err)
	}

	ref := &supporttypes.ExternalTicketRef{
		ResourceID:       stored.ResourceID,
		ResourceType:     supporttypes.ResourceType(stored.ResourceType),
		ExternalSystem:   supporttypes.ExternalSystem(stored.ExternalSystem),
		ExternalTicketID: stored.ExternalTicketID,
		ExternalURL:      stored.ExternalURL,
		CreatedAt:        time.Unix(stored.CreatedAt, 0).UTC(),
		CreatedBy:        stored.CreatedBy,
		UpdatedAt:        time.Unix(stored.UpdatedAt, 0).UTC(),
	}
	return ref, nil
}

func (w *rpcSupportChainWriter) broadcast(ctx context.Context, msg sdk.Msg) error {
	resp, err := w.txClient.BroadcastMsgs(ctx, []sdk.Msg{msg}, w.txOpts...)
	if err != nil {
		return err
	}

	switch res := resp.(type) {
	case *sdk.TxResponse:
		if res.Code != 0 {
			return fmt.Errorf("support broadcast failed code=%d log=%s", res.Code, res.RawLog)
		}
	}

	return nil
}

func querySupportStoreValue(
	ctx context.Context,
	client providerStoreQueryClient,
	timeout time.Duration,
	key []byte,
) ([]byte, error) {
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := client.ABCIQueryWithOptions(
		reqCtx,
		fmt.Sprintf("/store/%s/key", supporttypes.StoreKey),
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

func cloneSupportMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}
