package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RateLimitModuleName is the module name for rate limiting
const RateLimitModuleName = "ratelimit"

// RateLimitTStoreKey is the transient store key for rate limit counters
const RateLimitTStoreKey = "transient_ratelimit"

// Rate limit default parameters
const (
	// DefaultMaxTxPerBlockPerAccount is the default maximum transactions per block per account
	DefaultMaxTxPerBlockPerAccount uint64 = 10

	// DefaultMaxVEIDTxPerBlockGlobal is the default maximum VEID transactions per block globally
	DefaultMaxVEIDTxPerBlockGlobal uint64 = 100

	// DefaultMaxTotalTxPerBlock is the default maximum total transactions per block
	DefaultMaxTotalTxPerBlock uint64 = 5000

	// DefaultRateLimitEnabled is whether rate limiting is enabled by default
	DefaultRateLimitEnabled bool = true
)

// RateLimitParams defines the parameters for chain-level rate limiting
type RateLimitParams struct {
	// Enabled determines whether rate limiting is active
	Enabled bool `json:"enabled"`

	// MaxTxPerBlockPerAccount is the maximum number of transactions
	// a single account can submit per block
	MaxTxPerBlockPerAccount uint64 `json:"max_tx_per_block_per_account"`

	// MaxVEIDTxPerBlockGlobal is the maximum number of VEID verification
	// transactions allowed per block globally (these are expensive)
	MaxVEIDTxPerBlockGlobal uint64 `json:"max_veid_tx_per_block_global"`

	// MaxTotalTxPerBlock is the maximum total transactions per block
	MaxTotalTxPerBlock uint64 `json:"max_total_tx_per_block"`

	// ExemptAddresses is a list of addresses exempt from rate limiting
	// (e.g., genesis accounts, validators, system addresses)
	ExemptAddresses []string `json:"exempt_addresses"`
}

// DefaultRateLimitParams returns the default rate limit parameters
func DefaultRateLimitParams() RateLimitParams {
	return RateLimitParams{
		Enabled:                 DefaultRateLimitEnabled,
		MaxTxPerBlockPerAccount: DefaultMaxTxPerBlockPerAccount,
		MaxVEIDTxPerBlockGlobal: DefaultMaxVEIDTxPerBlockGlobal,
		MaxTotalTxPerBlock:      DefaultMaxTotalTxPerBlock,
		ExemptAddresses:         []string{},
	}
}

// Validate validates the rate limit parameters
func (p RateLimitParams) Validate() error {
	if p.MaxTxPerBlockPerAccount == 0 {
		return ErrInvalidRateLimitParams.Wrap("max_tx_per_block_per_account must be greater than 0")
	}
	if p.MaxVEIDTxPerBlockGlobal == 0 {
		return ErrInvalidRateLimitParams.Wrap("max_veid_tx_per_block_global must be greater than 0")
	}
	if p.MaxTotalTxPerBlock == 0 {
		return ErrInvalidRateLimitParams.Wrap("max_total_tx_per_block must be greater than 0")
	}
	// Validate exempt addresses
	for _, addr := range p.ExemptAddresses {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return ErrInvalidRateLimitParams.Wrapf("invalid exempt address: %s", addr)
		}
	}
	return nil
}

// IsExempt checks if an address is exempt from rate limiting
func (p RateLimitParams) IsExempt(addr sdk.AccAddress) bool {
	addrStr := addr.String()
	for _, exempt := range p.ExemptAddresses {
		if exempt == addrStr {
			return true
		}
	}
	return false
}

// RateLimitKeeper defines the interface for rate limit state management
type RateLimitKeeper interface {
	// GetParams returns the current rate limit parameters
	GetParams(ctx sdk.Context) RateLimitParams

	// GetAccountTxCount returns the current transaction count for an account in this block
	GetAccountTxCount(ctx sdk.Context, addr sdk.AccAddress) uint64

	// IncrementAccountTxCount increments the transaction count for an account
	IncrementAccountTxCount(ctx sdk.Context, addr sdk.AccAddress)

	// GetVEIDTxCount returns the current global VEID transaction count for this block
	GetVEIDTxCount(ctx sdk.Context) uint64

	// IncrementVEIDTxCount increments the global VEID transaction count
	IncrementVEIDTxCount(ctx sdk.Context)

	// GetTotalTxCount returns the current total transaction count for this block
	GetTotalTxCount(ctx sdk.Context) uint64

	// IncrementTotalTxCount increments the total transaction count
	IncrementTotalTxCount(ctx sdk.Context)
}

// RateLimitMetrics tracks rate limiting statistics for monitoring
type RateLimitMetrics struct {
	// TotalBlocked is the total number of transactions blocked by rate limiting
	TotalBlocked uint64

	// AccountBlocked is the number of transactions blocked due to per-account limits
	AccountBlocked uint64

	// VEIDBlocked is the number of VEID transactions blocked due to global limit
	VEIDBlocked uint64

	// TotalTxBlocked is the number of transactions blocked due to total block limit
	TotalTxBlocked uint64
}

// RateLimitEvent represents a rate limit event for telemetry
type RateLimitEvent struct {
	BlockHeight  int64
	Account      string
	TxType       string
	Blocked      bool
	Reason       string
	CurrentCount uint64
	Limit        uint64
}

// TransientRateLimitStore provides transient storage for per-block rate limit counters
type TransientRateLimitStore struct {
	params          RateLimitParams
	accountCounts   map[string]uint64
	veidCount       uint64
	totalCount      uint64
	currentBlockHeight int64
}

// NewTransientRateLimitStore creates a new transient rate limit store
func NewTransientRateLimitStore(params RateLimitParams) *TransientRateLimitStore {
	return &TransientRateLimitStore{
		params:        params,
		accountCounts: make(map[string]uint64),
		veidCount:     0,
		totalCount:    0,
		currentBlockHeight: 0,
	}
}

// ResetForBlock resets counters for a new block
func (s *TransientRateLimitStore) ResetForBlock(blockHeight int64) {
	if blockHeight != s.currentBlockHeight {
		s.accountCounts = make(map[string]uint64)
		s.veidCount = 0
		s.totalCount = 0
		s.currentBlockHeight = blockHeight
	}
}

// GetAccountTxCount returns the tx count for an account in the current block
func (s *TransientRateLimitStore) GetAccountTxCount(addr sdk.AccAddress) uint64 {
	return s.accountCounts[addr.String()]
}

// IncrementAccountTxCount increments the tx count for an account
func (s *TransientRateLimitStore) IncrementAccountTxCount(addr sdk.AccAddress) {
	s.accountCounts[addr.String()]++
}

// GetVEIDTxCount returns the global VEID tx count for the current block
func (s *TransientRateLimitStore) GetVEIDTxCount() uint64 {
	return s.veidCount
}

// IncrementVEIDTxCount increments the global VEID tx count
func (s *TransientRateLimitStore) IncrementVEIDTxCount() {
	s.veidCount++
}

// GetTotalTxCount returns the total tx count for the current block
func (s *TransientRateLimitStore) GetTotalTxCount() uint64 {
	return s.totalCount
}

// IncrementTotalTxCount increments the total tx count
func (s *TransientRateLimitStore) IncrementTotalTxCount() {
	s.totalCount++
}

// GetParams returns the rate limit params
func (s *TransientRateLimitStore) GetParams() RateLimitParams {
	return s.params
}

// SetParams updates the rate limit params
func (s *TransientRateLimitStore) SetParams(params RateLimitParams) {
	s.params = params
}

// Custom errors for rate limiting
var (
	// ErrAccountRateLimited is returned when an account exceeds its per-block transaction limit
	ErrAccountRateLimited = sdkerrors.Register(RateLimitModuleName, 1, "account rate limited: too many transactions in this block")

	// ErrVEIDRateLimited is returned when the global VEID transaction limit is exceeded
	ErrVEIDRateLimited = sdkerrors.Register(RateLimitModuleName, 2, "VEID rate limited: too many VEID transactions in this block")

	// ErrBlockRateLimited is returned when the total block transaction limit is exceeded
	ErrBlockRateLimited = sdkerrors.Register(RateLimitModuleName, 3, "block rate limited: too many transactions in this block")

	// ErrInvalidRateLimitParams is returned when rate limit parameters are invalid
	ErrInvalidRateLimitParams = sdkerrors.Register(RateLimitModuleName, 4, "invalid rate limit parameters")
)
