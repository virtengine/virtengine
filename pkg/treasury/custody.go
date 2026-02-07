package treasury

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

type WalletType string

type WalletStatus string

type WithdrawalStatus string

const (
	WalletTypeHot  WalletType = "hot"
	WalletTypeCold WalletType = "cold"

	WalletStatusActive  WalletStatus = "active"
	WalletStatusRetired WalletStatus = "retired"

	WithdrawalStatusPendingApproval WithdrawalStatus = "pending_approval"
	WithdrawalStatusApproved        WithdrawalStatus = "approved"
	WithdrawalStatusRejected        WithdrawalStatus = "rejected"
	WithdrawalStatusBlocked         WithdrawalStatus = "blocked"
	WithdrawalStatusExecuted        WithdrawalStatus = "executed"
)

type Wallet struct {
	ID               string
	Address          string
	Type             WalletType
	Status           WalletStatus
	LastRotatedAt    time.Time
	RotationInterval time.Duration
	Balances         map[string]sdkmath.Int
}

type RotationEvent struct {
	WalletID  string
	FromAddr  string
	ToAddr    string
	RotatedAt time.Time
	Reason    string
}

type WithdrawalPolicy struct {
	MaxHotPerTx          sdkmath.Int
	MaxHotDaily          sdkmath.Int
	RequireMultiSigAbove sdkmath.Int
	RequiredApprovals    int
	BlockedDestinations  map[string]struct{}
	AllowedDestinations  map[string]struct{}
	RequireAllowlist     bool
}

func DefaultWithdrawalPolicy() WithdrawalPolicy {
	return WithdrawalPolicy{
		MaxHotPerTx:          sdkmath.NewInt(2500000),
		MaxHotDaily:          sdkmath.NewInt(8000000),
		RequireMultiSigAbove: sdkmath.NewInt(1000000),
		RequiredApprovals:    2,
		BlockedDestinations:  make(map[string]struct{}),
		AllowedDestinations:  make(map[string]struct{}),
		RequireAllowlist:     false,
	}
}

type WithdrawalRequest struct {
	ID          string
	WalletID    string
	Asset       string
	Amount      sdkmath.Int
	Destination string
	RequestedAt time.Time
	Status      WithdrawalStatus
	Approvals   map[string]struct{}
	Reason      string
}

type CustodyManager struct {
	mu           sync.Mutex
	wallets      map[string]*Wallet
	policy       WithdrawalPolicy
	withdrawals  map[string]*WithdrawalRequest
	rotationLogs []RotationEvent
	dailyTotals  map[string]sdkmath.Int
}

func NewCustodyManager(policy WithdrawalPolicy) *CustodyManager {
	return &CustodyManager{
		wallets:     make(map[string]*Wallet),
		policy:      policy,
		withdrawals: make(map[string]*WithdrawalRequest),
		dailyTotals: make(map[string]sdkmath.Int),
	}
}

func (c *CustodyManager) AddWallet(wallet *Wallet) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if wallet.Balances == nil {
		wallet.Balances = make(map[string]sdkmath.Int)
	}
	c.wallets[wallet.ID] = wallet
}

func (c *CustodyManager) ListWallets() []*Wallet {
	c.mu.Lock()
	defer c.mu.Unlock()
	wallets := make([]*Wallet, 0, len(c.wallets))
	for _, wallet := range c.wallets {
		wallets = append(wallets, wallet)
	}
	sort.Slice(wallets, func(i, j int) bool {
		return strings.Compare(wallets[i].ID, wallets[j].ID) < 0
	})
	return wallets
}

func (c *CustodyManager) RequestWithdrawal(asset string, amount sdkmath.Int, destination string) (*WithdrawalRequest, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	wallet := c.selectWallet(amount)
	if wallet == nil {
		return nil, fmt.Errorf("no wallet available")
	}

	now := time.Now().UTC()
	req := &WithdrawalRequest{
		ID:          fmt.Sprintf("wd-%d", now.UnixNano()),
		WalletID:    wallet.ID,
		Asset:       asset,
		Amount:      amount,
		Destination: destination,
		RequestedAt: now,
		Status:      WithdrawalStatusPendingApproval,
		Approvals:   make(map[string]struct{}),
	}

	if blocked, reason := c.policyBlocks(destination); blocked {
		req.Status = WithdrawalStatusBlocked
		req.Reason = reason
		c.withdrawals[req.ID] = req
		return req, ErrPolicyBlocked
	}

	if c.policy.RequireAllowlist && !c.isAllowedDestination(destination) {
		req.Status = WithdrawalStatusBlocked
		req.Reason = "destination not allowlisted"
		c.withdrawals[req.ID] = req
		return req, ErrPolicyBlocked
	}

	if wallet.Type == WalletTypeHot {
		dateKey := now.Format("2006-01-02")
		used, ok := c.dailyTotals[dateKey]
		if !ok {
			used = sdkmath.ZeroInt()
		}
		if !c.policy.MaxHotDaily.IsZero() && used.Add(amount).GT(c.policy.MaxHotDaily) {
			req.Status = WithdrawalStatusBlocked
			req.Reason = "daily hot wallet limit exceeded"
			c.withdrawals[req.ID] = req
			return req, ErrPolicyBlocked
		}
		if !c.policy.MaxHotPerTx.IsZero() && amount.GT(c.policy.MaxHotPerTx) {
			req.Status = WithdrawalStatusBlocked
			req.Reason = "hot wallet per-transaction limit exceeded"
			c.withdrawals[req.ID] = req
			return req, ErrPolicyBlocked
		}
		c.dailyTotals[dateKey] = used.Add(amount)
	}

	if c.requiresApproval(amount, wallet.Type) {
		req.Status = WithdrawalStatusPendingApproval
		c.withdrawals[req.ID] = req
		return req, ErrApprovalRequired
	}

	req.Status = WithdrawalStatusApproved
	c.withdrawals[req.ID] = req
	return req, nil
}

func (c *CustodyManager) ApproveWithdrawal(id string, approver string) (*WithdrawalRequest, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	req, ok := c.withdrawals[id]
	if !ok {
		return nil, ErrWithdrawalNotFound
	}

	if req.Status != WithdrawalStatusPendingApproval {
		return req, nil
	}

	req.Approvals[approver] = struct{}{}
	if len(req.Approvals) >= c.policy.RequiredApprovals {
		req.Status = WithdrawalStatusApproved
	}
	return req, nil
}

func (c *CustodyManager) ExecuteWithdrawal(id string) (*WithdrawalRequest, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	req, ok := c.withdrawals[id]
	if !ok {
		return nil, ErrWithdrawalNotFound
	}

	if req.Status != WithdrawalStatusApproved {
		return req, ErrApprovalRequired
	}

	wallet := c.wallets[req.WalletID]
	if wallet == nil {
		return req, fmt.Errorf("wallet not found")
	}

	balance, ok := wallet.Balances[req.Asset]
	if !ok {
		balance = sdkmath.ZeroInt()
	}
	if balance.LT(req.Amount) {
		req.Status = WithdrawalStatusRejected
		req.Reason = "insufficient balance"
		return req, ErrPolicyBlocked
	}

	wallet.Balances[req.Asset] = balance.Sub(req.Amount)
	req.Status = WithdrawalStatusExecuted
	return req, nil
}

func (c *CustodyManager) RotateWalletIfDue(id string, reason string) (*Wallet, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	wallet, ok := c.wallets[id]
	if !ok {
		return nil, fmt.Errorf("wallet not found")
	}
	if wallet.RotationInterval == 0 {
		return wallet, nil
	}

	now := time.Now().UTC()
	if now.Sub(wallet.LastRotatedAt) < wallet.RotationInterval {
		return wallet, nil
	}

	oldAddr := wallet.Address
	wallet.Address = fmt.Sprintf("%s-rotated-%d", oldAddr, now.Unix())
	wallet.LastRotatedAt = now

	c.rotationLogs = append(c.rotationLogs, RotationEvent{
		WalletID:  wallet.ID,
		FromAddr:  oldAddr,
		ToAddr:    wallet.Address,
		RotatedAt: now,
		Reason:    reason,
	})

	return wallet, nil
}

func (c *CustodyManager) RotationLogs() []RotationEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	logs := make([]RotationEvent, len(c.rotationLogs))
	copy(logs, c.rotationLogs)
	return logs
}

func (c *CustodyManager) requiresApproval(amount sdkmath.Int, walletType WalletType) bool {
	if c.policy.RequiredApprovals == 0 {
		return false
	}
	if walletType == WalletTypeCold {
		return true
	}
	if c.policy.RequireMultiSigAbove.IsZero() {
		return false
	}
	return amount.GT(c.policy.RequireMultiSigAbove)
}

func (c *CustodyManager) selectWallet(amount sdkmath.Int) *Wallet {
	var hot *Wallet
	var cold *Wallet
	for _, wallet := range c.wallets {
		if wallet.Status != WalletStatusActive {
			continue
		}
		switch wallet.Type {
		case WalletTypeHot:
			if hot == nil {
				hot = wallet
			}
		case WalletTypeCold:
			if cold == nil {
				cold = wallet
			}
		}
	}

	if hot != nil && (c.policy.MaxHotPerTx.IsZero() || amount.LTE(c.policy.MaxHotPerTx)) {
		return hot
	}
	if cold != nil {
		return cold
	}
	return hot
}

func (c *CustodyManager) policyBlocks(destination string) (bool, string) {
	if _, ok := c.policy.BlockedDestinations[destination]; ok {
		return true, "destination blocked"
	}
	return false, ""
}

func (c *CustodyManager) isAllowedDestination(destination string) bool {
	if len(c.policy.AllowedDestinations) == 0 {
		return true
	}
	_, ok := c.policy.AllowedDestinations[destination]
	return ok
}
