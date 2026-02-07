// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
)

func TestCrossChainDeposit_Validate(t *testing.T) {
	validDeposit := CrossChainDeposit{
		SourceChain:     "cosmoshub-4",
		SourceChannel:   "channel-0",
		OriginalDenom:   "uatom",
		IBCDenom:        "ibc/transfer/channel-0/uatom",
		Amount:          sdkmath.NewInt(1000000),
		Sender:          "cosmos1sender",
		DepositorOnDest: sdk.AccAddress([]byte("depositor-acct")).String(),
		TimeoutHeight:   clienttypes.NewHeight(0, 100),
	}

	tests := []struct {
		name    string
		modify  func(*CrossChainDeposit)
		wantErr bool
	}{
		{
			name:    "valid deposit",
			modify:  func(d *CrossChainDeposit) {},
			wantErr: false,
		},
		{
			name:    "empty source chain",
			modify:  func(d *CrossChainDeposit) { d.SourceChain = "" },
			wantErr: true,
		},
		{
			name:    "empty source channel",
			modify:  func(d *CrossChainDeposit) { d.SourceChannel = "" },
			wantErr: true,
		},
		{
			name:    "empty original denom",
			modify:  func(d *CrossChainDeposit) { d.OriginalDenom = "" },
			wantErr: true,
		},
		{
			name:    "empty IBC denom",
			modify:  func(d *CrossChainDeposit) { d.IBCDenom = "" },
			wantErr: true,
		},
		{
			name:    "zero amount",
			modify:  func(d *CrossChainDeposit) { d.Amount = sdkmath.ZeroInt() },
			wantErr: true,
		},
		{
			name:    "negative amount",
			modify:  func(d *CrossChainDeposit) { d.Amount = sdkmath.NewInt(-100) },
			wantErr: true,
		},
		{
			name:    "empty sender",
			modify:  func(d *CrossChainDeposit) { d.Sender = "" },
			wantErr: true,
		},
		{
			name:    "empty depositor",
			modify:  func(d *CrossChainDeposit) { d.DepositorOnDest = "" },
			wantErr: true,
		},
		{
			name:    "invalid depositor address",
			modify:  func(d *CrossChainDeposit) { d.DepositorOnDest = "invalid" },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := validDeposit
			tt.modify(&d)
			err := d.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCrossChainSettlement_Validate(t *testing.T) {
	tests := []struct {
		name    string
		s       CrossChainSettlement
		wantErr bool
	}{
		{
			name: "valid settlement",
			s: CrossChainSettlement{
				SettlementID:    "settlement-1",
				SourceChain:     "virtengine-1",
				DestChain:       "cosmoshub-4",
				OrderID:         "order-1",
				ProviderAddress: "cosmos1...",
				CustomerAddress: "virtengine1...",
				Status:          SettlementStatusPending,
			},
			wantErr: false,
		},
		{
			name: "empty settlement ID",
			s: CrossChainSettlement{
				SourceChain:     "virtengine-1",
				DestChain:       "cosmoshub-4",
				OrderID:         "order-1",
				ProviderAddress: "cosmos1...",
				CustomerAddress: "virtengine1...",
				Status:          SettlementStatusPending,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			s: CrossChainSettlement{
				SettlementID:    "settlement-1",
				SourceChain:     "virtengine-1",
				DestChain:       "cosmoshub-4",
				OrderID:         "order-1",
				ProviderAddress: "cosmos1...",
				CustomerAddress: "virtengine1...",
				Status:          "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.s.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCrossChainSettlementStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   CrossChainSettlementStatus
		terminal bool
	}{
		{SettlementStatusPending, false},
		{SettlementStatusConfirmed, true},
		{SettlementStatusFailed, true},
		{SettlementStatusTimedOut, false},
		{SettlementStatusRefunded, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.terminal {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.terminal)
			}
		})
	}
}

func TestSupportedIBCDenom(t *testing.T) {
	tests := []struct {
		denom     string
		supported bool
	}{
		{"uve", true},
		{"uatom", true},
		{"uosmo", true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.denom, func(t *testing.T) {
			if got := SupportedIBCDenom(tt.denom); got != tt.supported {
				t.Errorf("SupportedIBCDenom(%s) = %v, want %v", tt.denom, got, tt.supported)
			}
		})
	}
}

func TestEstimateTransferFee(t *testing.T) {
	tests := []struct {
		name    string
		amount  sdkmath.Int
		numHops int
		want    sdk.Coin
	}{
		{
			name:    "single hop normal amount",
			amount:  sdkmath.NewInt(1000000),
			numHops: 1,
			want:    sdk.NewCoin("uve", sdkmath.NewInt(1000)),
		},
		{
			name:    "two hops",
			amount:  sdkmath.NewInt(1000000),
			numHops: 2,
			want:    sdk.NewCoin("uve", sdkmath.NewInt(2000)),
		},
		{
			name:    "small amount uses minimum fee",
			amount:  sdkmath.NewInt(100),
			numHops: 1,
			want:    sdk.NewCoin("uve", sdkmath.NewInt(1)), // min 1 per hop
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTransferFee(tt.amount, tt.numHops)
			if !got.IsEqual(tt.want) {
				t.Errorf("EstimateTransferFee() = %s, want %s", got, tt.want)
			}
		})
	}
}
