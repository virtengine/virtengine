package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/log"
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"

	supportv1 "github.com/virtengine/virtengine/sdk/go/node/support/v1"
	supporttypes "github.com/virtengine/virtengine/x/support/types"
)

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
	submitter  *ProviderMutationSubmitter
	storeQuery providerStoreQueryClient
	timeout    time.Duration
	logger     log.Logger
	grpcConn   *grpc.ClientConn
}

func newSupportChainWriter(_ context.Context, cfg SupportServiceConfig, logger log.Logger) (SupportChainWriter, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	if cfg.MutationSubmitter != nil {
		timeout := cfg.RequestTimeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		return &rpcSupportChainWriter{
			sender:     cfg.ProviderAddress,
			submitter:  cfg.MutationSubmitter,
			storeQuery: cfg.StoreQuery,
			timeout:    timeout,
			logger:     logger,
		}, nil
	}
	return nil, fmt.Errorf("%w: support writer requires generalized mutation submitter", ErrProviderMutationUnavailable)
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
	if w.storeQuery == nil {
		return nil, ErrProviderMutationUnavailable
	}
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
	if w.submitter != nil {
		kind, err := supportMutationKind(msg)
		if err != nil {
			return err
		}
		_, err = w.submitter.Submit(ctx, kind, msg)
		return err
	}
	return ErrProviderMutationUnavailable
}

func supportMutationKind(msg sdk.Msg) (ProviderMutationKind, error) {
	switch msg.(type) {
	case *supportv1.MsgUpdateSupportRequest:
		return MutationSupportUpdateRequest, nil
	case *supportv1.MsgAddSupportResponse:
		return MutationSupportAddResponse, nil
	case *supportv1.MsgRegisterExternalTicket:
		return MutationSupportRegisterExternal, nil
	case *supportv1.MsgUpdateExternalTicket:
		return MutationSupportUpdateExternal, nil
	default:
		return "", fmt.Errorf("unsupported support mutation %T", msg)
	}
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
