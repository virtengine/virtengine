// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package scenarios

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

type scenarioBackend interface {
	SubmitVEID(ctx context.Context, account string, payload []byte) (scenarioReceipt, error)
	CreateOrder(ctx context.Context, owner string, quantity int, config map[string]string) (scenarioReceipt, error)
	SubmitBid(ctx context.Context, orderID, provider string, price int64) (scenarioReceipt, error)
	SettleOrder(ctx context.Context, orderID string) (scenarioReceipt, error)
}

type scenarioReceipt struct {
	ID       string
	Metadata map[string]interface{}
}

type deterministicBackend struct {
	mu sync.Mutex

	veidSeq       uint64
	orderSeq      uint64
	bidSeq        uint64
	settlementSeq uint64

	orders map[string]*backendOrder
}

type backendOrder struct {
	ID         string
	Owner      string
	Quantity   int
	Config     map[string]string
	Settled    bool
	BestBidID  string
	BestPrice  int64
	Winner     string
	Bids       []*backendBid
	Settlement *backendSettlement
}

type backendBid struct {
	ID       string
	Provider string
	Price    int64
}

type backendSettlement struct {
	ID       string
	Winner   string
	Amount   int64
	Attempts int
}

func newDeterministicBackend() *deterministicBackend {
	return &deterministicBackend{
		orders: make(map[string]*backendOrder),
	}
}

func (b *deterministicBackend) SubmitVEID(_ context.Context, account string, payload []byte) (scenarioReceipt, error) {
	if account == "" {
		return scenarioReceipt{}, fmt.Errorf("account is required")
	}
	if len(payload) == 0 {
		return scenarioReceipt{}, fmt.Errorf("payload is required")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.veidSeq++
	hash := sha256.Sum256(payload)
	return scenarioReceipt{
		ID: fmt.Sprintf("veid-%06d", b.veidSeq),
		Metadata: map[string]interface{}{
			"account":      account,
			"payload_hash": hex.EncodeToString(hash[:]),
			"payload_size": len(payload),
		},
	}, nil
}

func (b *deterministicBackend) CreateOrder(_ context.Context, owner string, quantity int, config map[string]string) (scenarioReceipt, error) {
	if owner == "" {
		return scenarioReceipt{}, fmt.Errorf("owner is required")
	}
	if quantity <= 0 {
		return scenarioReceipt{}, fmt.Errorf("quantity must be positive")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.orderSeq++
	orderID := fmt.Sprintf("order-%06d", b.orderSeq)
	b.orders[orderID] = &backendOrder{
		ID:        orderID,
		Owner:     owner,
		Quantity:  quantity,
		Config:    cloneStringMap(config),
		BestPrice: 0,
	}

	return scenarioReceipt{
		ID: orderID,
		Metadata: map[string]interface{}{
			"owner":    owner,
			"quantity": quantity,
			"config":   cloneStringMap(config),
		},
	}, nil
}

func (b *deterministicBackend) SubmitBid(_ context.Context, orderID, provider string, price int64) (scenarioReceipt, error) {
	if provider == "" {
		return scenarioReceipt{}, fmt.Errorf("provider is required")
	}
	if price <= 0 {
		return scenarioReceipt{}, fmt.Errorf("price must be positive")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	order, ok := b.orders[orderID]
	if !ok {
		return scenarioReceipt{}, fmt.Errorf("order %s not found", orderID)
	}
	if order.Settled {
		return scenarioReceipt{}, fmt.Errorf("order %s already settled", orderID)
	}

	b.bidSeq++
	bidID := fmt.Sprintf("bid-%06d", b.bidSeq)
	bid := &backendBid{
		ID:       bidID,
		Provider: provider,
		Price:    price,
	}
	order.Bids = append(order.Bids, bid)

	isBestBid := order.BestBidID == "" || price < order.BestPrice
	if isBestBid {
		order.BestBidID = bidID
		order.BestPrice = price
		order.Winner = provider
	}

	return scenarioReceipt{
		ID: bidID,
		Metadata: map[string]interface{}{
			"order_id":    orderID,
			"provider":    provider,
			"price":       price,
			"is_best_bid": isBestBid,
		},
	}, nil
}

func (b *deterministicBackend) SettleOrder(_ context.Context, orderID string) (scenarioReceipt, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	order, ok := b.orders[orderID]
	if !ok {
		return scenarioReceipt{}, fmt.Errorf("order %s not found", orderID)
	}
	if len(order.Bids) == 0 {
		return scenarioReceipt{}, fmt.Errorf("order %s has no bids", orderID)
	}

	if order.Settlement != nil {
		order.Settlement.Attempts++
		return scenarioReceipt{
			ID: order.Settlement.ID,
			Metadata: map[string]interface{}{
				"order_id":          orderID,
				"winner_provider":   order.Settlement.Winner,
				"settlement_amount": order.Settlement.Amount,
				"attempts":          order.Settlement.Attempts,
			},
		}, nil
	}

	b.settlementSeq++
	settlementID := fmt.Sprintf("settlement-%06d", b.settlementSeq)
	order.Settled = true
	order.Settlement = &backendSettlement{
		ID:       settlementID,
		Winner:   order.Winner,
		Amount:   order.BestPrice * int64(order.Quantity),
		Attempts: 1,
	}

	return scenarioReceipt{
		ID: settlementID,
		Metadata: map[string]interface{}{
			"order_id":          orderID,
			"winner_provider":   order.Settlement.Winner,
			"settlement_amount": order.Settlement.Amount,
			"attempts":          order.Settlement.Attempts,
			"bid_count":         len(order.Bids),
		},
	}, nil
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}

	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
