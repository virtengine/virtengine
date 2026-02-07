// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TimeoutAction defines what happens when an IBC transfer times out.
type TimeoutAction string

const (
	// TimeoutActionRefund automatically refunds the sender
	TimeoutActionRefund TimeoutAction = "refund"

	// TimeoutActionRetry automatically retries the transfer
	TimeoutActionRetry TimeoutAction = "retry"

	// TimeoutActionEscrowHold keeps funds in escrow for manual resolution
	TimeoutActionEscrowHold TimeoutAction = "escrow_hold"
)

// MaxRetryAttempts is the maximum number of retry attempts for timed-out transfers.
const MaxRetryAttempts = 3

// Store key prefixes for timeout tracking.
var (
	PrefixPendingTransfer  = []byte{0x30}
	PrefixTimedOutTransfer = []byte{0x31}
	PrefixTransferSequence = []byte{0x32}
)

// PendingTransfer tracks an in-flight IBC transfer for timeout handling.
type PendingTransfer struct {
	// TransferID is the unique identifier
	TransferID string `json:"transfer_id"`

	// SourceChannel is the IBC channel used
	SourceChannel string `json:"source_channel"`

	// Sender is the sender address
	Sender string `json:"sender"`

	// Receiver is the receiver address on the destination chain
	Receiver string `json:"receiver"`

	// Amount is the transfer amount
	Amount sdk.Coins `json:"amount"`

	// SettlementID is the linked settlement (if applicable)
	SettlementID string `json:"settlement_id,omitempty"`

	// CreatedAt is when the transfer was initiated
	CreatedAt time.Time `json:"created_at"`

	// TimeoutAction determines what happens on timeout
	TimeoutAction TimeoutAction `json:"timeout_action"`

	// RetryCount is the number of retry attempts made
	RetryCount int `json:"retry_count"`

	// Status is the current transfer status
	Status TransferStatus `json:"status"`
}

// TransferStatus represents the status of a pending transfer.
type TransferStatus string

const (
	TransferStatusPending  TransferStatus = "pending"
	TransferStatusComplete TransferStatus = "complete"
	TransferStatusTimedOut TransferStatus = "timed_out"
	TransferStatusRefunded TransferStatus = "refunded"
	TransferStatusRetrying TransferStatus = "retrying"
)

// Validate validates a pending transfer.
func (t PendingTransfer) Validate() error {
	if t.TransferID == "" {
		return fmt.Errorf("transfer ID cannot be empty")
	}
	if t.SourceChannel == "" {
		return fmt.Errorf("source channel cannot be empty")
	}
	if t.Sender == "" {
		return fmt.Errorf("sender cannot be empty")
	}
	if t.Receiver == "" {
		return fmt.Errorf("receiver cannot be empty")
	}
	if t.Amount.IsZero() {
		return fmt.Errorf("amount cannot be zero")
	}
	return nil
}

// TimeoutKeeper manages IBC transfer timeout handling.
type TimeoutKeeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewTimeoutKeeper creates a new TimeoutKeeper.
func NewTimeoutKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) TimeoutKeeper {
	return TimeoutKeeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// StorePendingTransfer stores a pending transfer for timeout tracking.
func (k TimeoutKeeper) StorePendingTransfer(ctx sdk.Context, transfer PendingTransfer) error {
	if err := transfer.Validate(); err != nil {
		return fmt.Errorf("invalid pending transfer: %w", err)
	}

	bz, err := json.Marshal(transfer)
	if err != nil {
		return fmt.Errorf("failed to marshal pending transfer: %w", err)
	}

	store := ctx.KVStore(k.storeKey)
	key := pendingTransferKey(transfer.TransferID)
	store.Set(key, bz)
	return nil
}

// GetPendingTransfer retrieves a pending transfer by ID.
func (k TimeoutKeeper) GetPendingTransfer(ctx sdk.Context, transferID string) (PendingTransfer, bool) {
	store := ctx.KVStore(k.storeKey)
	key := pendingTransferKey(transferID)
	bz := store.Get(key)
	if bz == nil {
		return PendingTransfer{}, false
	}

	var transfer PendingTransfer
	if err := json.Unmarshal(bz, &transfer); err != nil {
		return PendingTransfer{}, false
	}
	return transfer, true
}

// UpdateTransferStatus updates the status of a pending transfer.
func (k TimeoutKeeper) UpdateTransferStatus(ctx sdk.Context, transferID string, status TransferStatus) error {
	transfer, found := k.GetPendingTransfer(ctx, transferID)
	if !found {
		return fmt.Errorf("transfer %s not found", transferID)
	}

	transfer.Status = status
	return k.StorePendingTransfer(ctx, transfer)
}

// HandleTimeout processes a timeout event for a pending transfer.
func (k TimeoutKeeper) HandleTimeout(ctx sdk.Context, transferID string) (*PendingTransfer, error) {
	transfer, found := k.GetPendingTransfer(ctx, transferID)
	if !found {
		return nil, fmt.Errorf("transfer %s not found", transferID)
	}

	switch transfer.TimeoutAction {
	case TimeoutActionRefund:
		transfer.Status = TransferStatusRefunded
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			"ibc_transfer_refunded",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("sender", transfer.Sender),
			sdk.NewAttribute("amount", transfer.Amount.String()),
		))

	case TimeoutActionRetry:
		if transfer.RetryCount >= MaxRetryAttempts {
			transfer.Status = TransferStatusTimedOut
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				"ibc_transfer_max_retries",
				sdk.NewAttribute("transfer_id", transferID),
				sdk.NewAttribute("retry_count", fmt.Sprintf("%d", transfer.RetryCount)),
			))
		} else {
			transfer.RetryCount++
			transfer.Status = TransferStatusRetrying
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				"ibc_transfer_retry",
				sdk.NewAttribute("transfer_id", transferID),
				sdk.NewAttribute("retry_count", fmt.Sprintf("%d", transfer.RetryCount)),
			))
		}

	case TimeoutActionEscrowHold:
		transfer.Status = TransferStatusTimedOut
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			"ibc_transfer_escrow_hold",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("settlement_id", transfer.SettlementID),
		))
	}

	if err := k.StorePendingTransfer(ctx, transfer); err != nil {
		return nil, fmt.Errorf("failed to update transfer: %w", err)
	}

	return &transfer, nil
}

// GetTimedOutTransfers returns all transfers in timed-out state.
func (k TimeoutKeeper) GetTimedOutTransfers(ctx sdk.Context) []PendingTransfer {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, PrefixPendingTransfer)
	defer iter.Close()

	var timedOut []PendingTransfer
	for ; iter.Valid(); iter.Next() {
		var transfer PendingTransfer
		if err := json.Unmarshal(iter.Value(), &transfer); err != nil {
			continue
		}
		if transfer.Status == TransferStatusTimedOut {
			timedOut = append(timedOut, transfer)
		}
	}
	return timedOut
}

// DeletePendingTransfer removes a pending transfer.
func (k TimeoutKeeper) DeletePendingTransfer(ctx sdk.Context, transferID string) {
	store := ctx.KVStore(k.storeKey)
	key := pendingTransferKey(transferID)
	store.Delete(key)
}

func pendingTransferKey(transferID string) []byte {
	key := make([]byte, 0, len(PrefixPendingTransfer)+len(transferID))
	key = append(key, PrefixPendingTransfer...)
	key = append(key, []byte(transferID)...)
	return key
}
