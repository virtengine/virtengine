// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
)

func TestTransferParams_Validate(t *testing.T) {
	validParams := TransferParams{
		SourcePort:       "transfer",
		SourceChannel:    "channel-0",
		Token:            sdk.NewInt64Coin("uve", 1000000),
		Sender:           "virtengine1sender",
		Receiver:         "cosmos1receiver",
		TimeoutHeight:    clienttypes.NewHeight(0, 100),
		TimeoutTimestamp: 0,
	}

	tests := []struct {
		name    string
		modify  func(*TransferParams)
		wantErr bool
	}{
		{
			name:    "valid params",
			modify:  func(p *TransferParams) {},
			wantErr: false,
		},
		{
			name:    "empty source port",
			modify:  func(p *TransferParams) { p.SourcePort = "" },
			wantErr: true,
		},
		{
			name:    "empty source channel",
			modify:  func(p *TransferParams) { p.SourceChannel = "" },
			wantErr: true,
		},
		{
			name:    "zero token",
			modify:  func(p *TransferParams) { p.Token = sdk.NewInt64Coin("uve", 0) },
			wantErr: true,
		},
		{
			name:    "empty sender",
			modify:  func(p *TransferParams) { p.Sender = "" },
			wantErr: true,
		},
		{
			name:    "empty receiver",
			modify:  func(p *TransferParams) { p.Receiver = "" },
			wantErr: true,
		},
		{
			name: "no timeout",
			modify: func(p *TransferParams) {
				p.TimeoutHeight = clienttypes.ZeroHeight()
				p.TimeoutTimestamp = 0
			},
			wantErr: true,
		},
		{
			name: "timestamp timeout instead of height",
			modify: func(p *TransferParams) {
				p.TimeoutHeight = clienttypes.ZeroHeight()
				p.TimeoutTimestamp = 1000000000
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validParams
			tt.modify(&p)
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewTransferMsg(t *testing.T) {
	params := TransferParams{
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
		Token:         sdk.NewInt64Coin("uve", 1000000),
		Sender:        "virtengine1sender",
		Receiver:      "cosmos1receiver",
		TimeoutHeight: clienttypes.NewHeight(0, 100),
		Memo:          "test transfer",
	}

	msg, err := NewTransferMsg(params)
	if err != nil {
		t.Fatalf("NewTransferMsg() error = %v", err)
	}
	if msg == nil {
		t.Fatal("NewTransferMsg() returned nil")
	}
	if msg.SourcePort != "transfer" {
		t.Errorf("SourcePort = %s, want transfer", msg.SourcePort)
	}
	if msg.SourceChannel != "channel-0" {
		t.Errorf("SourceChannel = %s, want channel-0", msg.SourceChannel)
	}
	if msg.Memo != "test transfer" {
		t.Errorf("Memo = %s, want 'test transfer'", msg.Memo)
	}
}

func TestNewTransferMsg_Invalid(t *testing.T) {
	params := TransferParams{} // empty, should fail validation
	_, err := NewTransferMsg(params)
	if err == nil {
		t.Error("NewTransferMsg() should fail for invalid params")
	}
}

func TestMultiHopRoute_Validate(t *testing.T) {
	tests := []struct {
		name    string
		route   MultiHopRoute
		wantErr bool
	}{
		{
			name: "valid single hop",
			route: MultiHopRoute{
				Hops: []RouteHop{
					{SourcePort: "transfer", SourceChannel: "channel-0", DestChainID: "cosmoshub-4"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid multi hop",
			route: MultiHopRoute{
				Hops: []RouteHop{
					{SourcePort: "transfer", SourceChannel: "channel-0", DestChainID: "cosmoshub-4"},
					{SourcePort: "transfer", SourceChannel: "channel-1", DestChainID: "osmosis-1"},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty hops",
			route:   MultiHopRoute{},
			wantErr: true,
		},
		{
			name: "empty source port in hop",
			route: MultiHopRoute{
				Hops: []RouteHop{
					{SourcePort: "", SourceChannel: "channel-0", DestChainID: "cosmoshub-4"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty source channel in hop",
			route: MultiHopRoute{
				Hops: []RouteHop{
					{SourcePort: "transfer", SourceChannel: "", DestChainID: "cosmoshub-4"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty dest chain ID in hop",
			route: MultiHopRoute{
				Hops: []RouteHop{
					{SourcePort: "transfer", SourceChannel: "channel-0", DestChainID: ""},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.route.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPendingTransfer_Validate(t *testing.T) {
	tests := []struct {
		name    string
		t       PendingTransfer
		wantErr bool
	}{
		{
			name: "valid transfer",
			t: PendingTransfer{
				TransferID:    "transfer-1",
				SourceChannel: "channel-0",
				Sender:        "virtengine1sender",
				Receiver:      "cosmos1receiver",
				Amount:        sdk.NewCoins(sdk.NewInt64Coin("uve", 1000000)),
				TimeoutAction: TimeoutActionRefund,
				Status:        TransferStatusPending,
			},
			wantErr: false,
		},
		{
			name: "empty transfer ID",
			t: PendingTransfer{
				SourceChannel: "channel-0",
				Sender:        "virtengine1sender",
				Receiver:      "cosmos1receiver",
				Amount:        sdk.NewCoins(sdk.NewInt64Coin("uve", 1000000)),
			},
			wantErr: true,
		},
		{
			name: "empty amount",
			t: PendingTransfer{
				TransferID:    "transfer-1",
				SourceChannel: "channel-0",
				Sender:        "virtengine1sender",
				Receiver:      "cosmos1receiver",
				Amount:        sdk.Coins{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.t.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
