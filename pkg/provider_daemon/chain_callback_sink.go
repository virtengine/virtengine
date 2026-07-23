// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/virtengine/virtengine/sdk/go/node/client/types"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
)

type GasSetting = types.GasSetting

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
	Sender            string
	MutationSubmitter *ProviderMutationSubmitter
}

type ChainCallbackSink struct {
	sender    string
	submitter *ProviderMutationSubmitter
}

func NewChainCallbackSink(_ context.Context, cfg ChainCallbackSinkConfig) (*ChainCallbackSink, error) {
	if cfg.MutationSubmitter == nil {
		return nil, fmt.Errorf("%w: callback sink requires generalized mutation submitter", ErrProviderMutationUnavailable)
	}
	if cfg.Sender == "" {
		return nil, fmt.Errorf("callback sender is required")
	}
	return &ChainCallbackSink{sender: cfg.Sender, submitter: cfg.MutationSubmitter}, nil
}

func NewDurableChainCallbackSink(sender string, submitter *ProviderMutationSubmitter) (*ChainCallbackSink, error) {
	if submitter == nil {
		return nil, fmt.Errorf("%w: callback sink requires generalized mutation submitter", ErrProviderMutationUnavailable)
	}
	if strings.TrimSpace(sender) == "" {
		return nil, fmt.Errorf("callback sender is required")
	}
	return &ChainCallbackSink{sender: sender, submitter: submitter}, nil
}

func (s *ChainCallbackSink) SenderAddress() string {
	if s == nil {
		return ""
	}
	return s.sender
}

func (s *ChainCallbackSink) Submit(ctx context.Context, callback *marketplacetypes.WaldurCallback) error {
	if s == nil || s.submitter == nil {
		return ErrProviderMutationUnavailable
	}
	if callback == nil {
		return fmt.Errorf("callback is nil")
	}
	if callback.SignerID != s.sender {
		return fmt.Errorf("callback signer mismatch: %s != %s", callback.SignerID, s.sender)
	}
	payloadBytes, err := json.Marshal(callback)
	if err != nil {
		return fmt.Errorf("marshal callback: %w", err)
	}
	msg := marketplacetypes.NewMsgWaldurCallback(s.sender, string(callback.ActionType), callback.ChainEntityID, string(callback.ChainEntityType), string(payloadBytes), callback.Signature)
	_, err = s.submitter.Submit(ctx, MutationMarketplaceCallback, msg)
	return err
}
