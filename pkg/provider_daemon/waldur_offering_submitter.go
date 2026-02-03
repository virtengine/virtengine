// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-3D: Chain offering submitter for Waldur ingestion.
package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
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
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	"github.com/virtengine/virtengine/x/marketplace"
)

// ChainOfferingSubmitterConfig configures on-chain offering submission.
type ChainOfferingSubmitterConfig struct {
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
	IngestStateFile   string
}

// ChainOfferingSubmitter submits offerings to the chain.
type ChainOfferingSubmitter struct {
	client          clientv1beta3.Client
	sender          string
	opts            []clientv1beta3.BroadcastOption
	ingestStateFile string
	veidQuery       veidv1.QueryClient
}

// NewChainOfferingSubmitter creates a submitter for offering ingestion.
func NewChainOfferingSubmitter(ctx context.Context, cfg ChainOfferingSubmitterConfig) (*ChainOfferingSubmitter, error) {
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
		WithBroadcastMode(clientv1beta3.BroadcastSync)

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

	return &ChainOfferingSubmitter{
		client:          client,
		sender:          addr.String(),
		opts:            opts,
		ingestStateFile: cfg.IngestStateFile,
		veidQuery:       veidv1.NewQueryClient(grpcConn),
	}, nil
}

// CreateOffering creates a new on-chain offering.
func (s *ChainOfferingSubmitter) CreateOffering(ctx context.Context, offering *marketplacetypes.Offering) (string, error) {
	if offering == nil {
		return "", fmt.Errorf("offering is nil")
	}
	if offering.ID.ProviderAddress != s.sender {
		return "", fmt.Errorf("offering provider mismatch: %s != %s", offering.ID.ProviderAddress, s.sender)
	}

	payload, err := json.Marshal(offering)
	if err != nil {
		return "", fmt.Errorf("marshal offering: %w", err)
	}

	msg := marketplacetypes.NewMsgCreateOffering(s.sender, payload)
	resp, err := s.client.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg}, s.opts...)
	if err != nil {
		return "", err
	}
	if err := checkTxResponse(resp); err != nil {
		return "", err
	}

	return offering.ID.String(), nil
}

// UpdateOffering updates an existing on-chain offering.
func (s *ChainOfferingSubmitter) UpdateOffering(ctx context.Context, offeringID string, offering *marketplacetypes.Offering) error {
	if offeringID == "" {
		return fmt.Errorf("offering ID is required")
	}
	if offering == nil {
		return fmt.Errorf("offering is nil")
	}
	if offering.ID.ProviderAddress != s.sender {
		return fmt.Errorf("offering provider mismatch: %s != %s", offering.ID.ProviderAddress, s.sender)
	}

	payload, err := json.Marshal(offering)
	if err != nil {
		return fmt.Errorf("marshal offering: %w", err)
	}

	msg := marketplacetypes.NewMsgUpdateOffering(s.sender, offeringID, payload)
	resp, err := s.client.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg}, s.opts...)
	if err != nil {
		return err
	}
	return checkTxResponse(resp)
}

// DeprecateOffering deprecates an offering on-chain.
func (s *ChainOfferingSubmitter) DeprecateOffering(ctx context.Context, offeringID string) error {
	if offeringID == "" {
		return fmt.Errorf("offering ID is required")
	}

	msg := marketplacetypes.NewMsgDeprecateOffering(s.sender, offeringID, "")
	resp, err := s.client.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg}, s.opts...)
	if err != nil {
		return err
	}
	return checkTxResponse(resp)
}

// GetNextOfferingSequence returns the next offering sequence number.
func (s *ChainOfferingSubmitter) GetNextOfferingSequence(ctx context.Context, providerAddress string) (uint64, error) {
	if providerAddress == "" {
		return 0, fmt.Errorf("provider address is required")
	}
	if s.ingestStateFile == "" {
		return 1, nil
	}

	stateStore := NewWaldurIngestStateStore(s.ingestStateFile)
	state, err := stateStore.Load("", providerAddress)
	if err != nil {
		return 0, fmt.Errorf("load ingest state: %w", err)
	}

	maxSeq := uint64(0)
	for _, record := range state.Records {
		if record == nil || record.ChainOfferingID == "" {
			continue
		}
		offeringID, err := marketplacetypes.ParseOfferingID(record.ChainOfferingID)
		if err != nil {
			continue
		}
		if offeringID.ProviderAddress != providerAddress {
			continue
		}
		if offeringID.Sequence > maxSeq {
			maxSeq = offeringID.Sequence
		}
	}

	return maxSeq + 1, nil
}

// ValidateProviderVEID checks the provider VEID score.
func (s *ChainOfferingSubmitter) ValidateProviderVEID(ctx context.Context, providerAddress string, minScore uint32) error {
	if minScore == 0 {
		return nil
	}
	if providerAddress == "" {
		return fmt.Errorf("provider address is required")
	}
	if s.veidQuery == nil {
		return fmt.Errorf("veid query client unavailable")
	}

	resp, err := s.veidQuery.Identity(ctx, &veidv1.QueryIdentityRequest{AccountAddress: providerAddress})
	if err != nil {
		return fmt.Errorf("query identity: %w", err)
	}
	if resp == nil || !resp.Found || resp.Identity == nil {
		return fmt.Errorf("provider identity not found")
	}
	if resp.Identity.CurrentScore < minScore {
		return fmt.Errorf("provider identity score %d below minimum %d", resp.Identity.CurrentScore, minScore)
	}

	return nil
}

func checkTxResponse(resp interface{}) error {
	switch res := resp.(type) {
	case *sdk.TxResponse:
		if res.Code != 0 {
			return fmt.Errorf("broadcast failed code=%d log=%s", res.Code, res.RawLog)
		}
	}
	return nil
}
