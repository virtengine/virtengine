// Package types contains types for the delegation module.
//
// VE-922: Delegation types for delegated staking
package types

import (
	"fmt"
	"math/big"
	"time"
)

// DelegationStatus represents the status of a delegation
type DelegationStatus string

const (
	// DelegationStatusActive means the delegation is active
	DelegationStatusActive DelegationStatus = "active"

	// DelegationStatusUnbonding means the delegation is unbonding
	DelegationStatusUnbonding DelegationStatus = "unbonding"

	// SharePrecision is the precision for share calculations (18 decimals)
	SharePrecision = 18

	// MaxSharesPerValidator is the maximum total shares per validator
	// Using a large number to prevent overflow while allowing sufficient precision
	MaxSharesPerValidator = "1000000000000000000000000000" // 10^27
)

// Delegation represents a delegation from a delegator to a validator
type Delegation struct {
	// DelegatorAddress is the delegator's address
	DelegatorAddress string `json:"delegator_address"`

	// ValidatorAddress is the validator's address
	ValidatorAddress string `json:"validator_address"`

	// Shares is the delegation shares (fixed-point, 18 decimals)
	// Shares represent the delegator's proportional ownership of the validator's stake
	Shares string `json:"shares"`

	// InitialAmount is the initial delegation amount in base units
	InitialAmount string `json:"initial_amount"`

	// CreatedAt is when the delegation was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the delegation was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Height is the block height when delegation was created
	Height int64 `json:"height"`
}

// NewDelegation creates a new delegation
func NewDelegation(delegatorAddr, validatorAddr string, shares, amount string, blockTime time.Time, height int64) *Delegation {
	return &Delegation{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Shares:           shares,
		InitialAmount:    amount,
		CreatedAt:        blockTime,
		UpdatedAt:        blockTime,
		Height:           height,
	}
}

// Validate validates the delegation
func (d *Delegation) Validate() error {
	if d.DelegatorAddress == "" {
		return fmt.Errorf("delegator_address cannot be empty")
	}
	if d.ValidatorAddress == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}

	shares, ok := new(big.Int).SetString(d.Shares, 10)
	if !ok || shares.Sign() < 0 {
		return fmt.Errorf("invalid shares: %s", d.Shares)
	}

	amount, ok := new(big.Int).SetString(d.InitialAmount, 10)
	if !ok || amount.Sign() < 0 {
		return fmt.Errorf("invalid initial_amount: %s", d.InitialAmount)
	}

	return nil
}

// GetSharesBigInt returns shares as big.Int
func (d *Delegation) GetSharesBigInt() *big.Int {
	shares, _ := new(big.Int).SetString(d.Shares, 10)
	if shares == nil {
		return big.NewInt(0)
	}
	return shares
}

// AddShares adds shares to the delegation
func (d *Delegation) AddShares(amount string, updateTime time.Time) error {
	add, ok := new(big.Int).SetString(amount, 10)
	if !ok || add.Sign() < 0 {
		return fmt.Errorf("invalid shares amount: %s", amount)
	}

	current := d.GetSharesBigInt()
	current.Add(current, add)
	d.Shares = current.String()
	d.UpdatedAt = updateTime
	return nil
}

// SubtractShares subtracts shares from the delegation
func (d *Delegation) SubtractShares(amount string, updateTime time.Time) error {
	sub, ok := new(big.Int).SetString(amount, 10)
	if !ok || sub.Sign() < 0 {
		return fmt.Errorf("invalid shares amount: %s", amount)
	}

	current := d.GetSharesBigInt()
	if current.Cmp(sub) < 0 {
		return fmt.Errorf("insufficient shares: have %s, need %s", d.Shares, amount)
	}

	current.Sub(current, sub)
	d.Shares = current.String()
	d.UpdatedAt = updateTime
	return nil
}

// UnbondingDelegation represents a delegation that is unbonding
type UnbondingDelegation struct {
	// ID is the unique unbonding delegation ID
	ID string `json:"id"`

	// DelegatorAddress is the delegator's address
	DelegatorAddress string `json:"delegator_address"`

	// ValidatorAddress is the validator's address
	ValidatorAddress string `json:"validator_address"`

	// Entries are the unbonding entries
	Entries []UnbondingDelegationEntry `json:"entries"`

	// CreatedAt is when the unbonding started
	CreatedAt time.Time `json:"created_at"`

	// Height is the block height when unbonding was initiated
	Height int64 `json:"height"`
}

// UnbondingDelegationEntry represents a single unbonding entry
type UnbondingDelegationEntry struct {
	// CreationHeight is the height at which the unbonding was created
	CreationHeight int64 `json:"creation_height"`

	// CompletionTime is when the unbonding will complete
	CompletionTime time.Time `json:"completion_time"`

	// InitialBalance is the initial balance to undelegate
	InitialBalance string `json:"initial_balance"`

	// Balance is the remaining balance to return
	Balance string `json:"balance"`

	// UnbondingShares is the shares being unbonded
	UnbondingShares string `json:"unbonding_shares"`
}

// NewUnbondingDelegation creates a new unbonding delegation
func NewUnbondingDelegation(id, delegatorAddr, validatorAddr string, height int64, completionTime, creationTime time.Time, balance, shares string) *UnbondingDelegation {
	return &UnbondingDelegation{
		ID:               id,
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Entries: []UnbondingDelegationEntry{
			{
				CreationHeight:  height,
				CompletionTime:  completionTime,
				InitialBalance:  balance,
				Balance:         balance,
				UnbondingShares: shares,
			},
		},
		CreatedAt: creationTime,
		Height:    height,
	}
}

// Validate validates the unbonding delegation
func (u *UnbondingDelegation) Validate() error {
	if u.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}
	if u.DelegatorAddress == "" {
		return fmt.Errorf("delegator_address cannot be empty")
	}
	if u.ValidatorAddress == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}
	if len(u.Entries) == 0 {
		return fmt.Errorf("entries cannot be empty")
	}

	for i, entry := range u.Entries {
		balance, ok := new(big.Int).SetString(entry.Balance, 10)
		if !ok || balance.Sign() < 0 {
			return fmt.Errorf("invalid balance in entry %d: %s", i, entry.Balance)
		}
	}

	return nil
}

// TotalBalance returns the total balance of all entries
func (u *UnbondingDelegation) TotalBalance() *big.Int {
	total := big.NewInt(0)
	for _, entry := range u.Entries {
		balance, ok := new(big.Int).SetString(entry.Balance, 10)
		if ok {
			total.Add(total, balance)
		}
	}
	return total
}

// Redelegation represents a redelegation from one validator to another
type Redelegation struct {
	// ID is the unique redelegation ID
	ID string `json:"id"`

	// DelegatorAddress is the delegator's address
	DelegatorAddress string `json:"delegator_address"`

	// ValidatorSrcAddress is the source validator's address
	ValidatorSrcAddress string `json:"validator_src_address"`

	// ValidatorDstAddress is the destination validator's address
	ValidatorDstAddress string `json:"validator_dst_address"`

	// Entries are the redelegation entries
	Entries []RedelegationEntry `json:"entries"`

	// CreatedAt is when the redelegation started
	CreatedAt time.Time `json:"created_at"`

	// Height is the block height when redelegation was initiated
	Height int64 `json:"height"`
}

// RedelegationEntry represents a single redelegation entry
type RedelegationEntry struct {
	// CreationHeight is the height at which the redelegation was created
	CreationHeight int64 `json:"creation_height"`

	// CompletionTime is when the redelegation matures
	CompletionTime time.Time `json:"completion_time"`

	// InitialBalance is the initial balance being redelegated
	InitialBalance string `json:"initial_balance"`

	// SharesDst is the shares received at the destination validator
	SharesDst string `json:"shares_dst"`
}

// NewRedelegation creates a new redelegation
func NewRedelegation(id, delegatorAddr, srcValidator, dstValidator string, height int64, completionTime, creationTime time.Time, balance, sharesDst string) *Redelegation {
	return &Redelegation{
		ID:                  id,
		DelegatorAddress:    delegatorAddr,
		ValidatorSrcAddress: srcValidator,
		ValidatorDstAddress: dstValidator,
		Entries: []RedelegationEntry{
			{
				CreationHeight: height,
				CompletionTime: completionTime,
				InitialBalance: balance,
				SharesDst:      sharesDst,
			},
		},
		CreatedAt: creationTime,
		Height:    height,
	}
}

// Validate validates the redelegation
func (r *Redelegation) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}
	if r.DelegatorAddress == "" {
		return fmt.Errorf("delegator_address cannot be empty")
	}
	if r.ValidatorSrcAddress == "" {
		return fmt.Errorf("validator_src_address cannot be empty")
	}
	if r.ValidatorDstAddress == "" {
		return fmt.Errorf("validator_dst_address cannot be empty")
	}
	if r.ValidatorSrcAddress == r.ValidatorDstAddress {
		return fmt.Errorf("source and destination validators cannot be the same")
	}
	if len(r.Entries) == 0 {
		return fmt.Errorf("entries cannot be empty")
	}
	return nil
}

// ValidatorShares represents the total shares for a validator
type ValidatorShares struct {
	// ValidatorAddress is the validator's address
	ValidatorAddress string `json:"validator_address"`

	// TotalShares is the total delegation shares (fixed-point, 18 decimals)
	TotalShares string `json:"total_shares"`

	// TotalStake is the total stake amount in base units
	TotalStake string `json:"total_stake"`

	// UpdatedAt is when the record was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// NewValidatorShares creates a new validator shares record
func NewValidatorShares(validatorAddr string, updateTime time.Time) *ValidatorShares {
	return &ValidatorShares{
		ValidatorAddress: validatorAddr,
		TotalShares:      "0",
		TotalStake:       "0",
		UpdatedAt:        updateTime,
	}
}

// Validate validates the validator shares
func (v *ValidatorShares) Validate() error {
	if v.ValidatorAddress == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}

	shares, ok := new(big.Int).SetString(v.TotalShares, 10)
	if !ok || shares.Sign() < 0 {
		return fmt.Errorf("invalid total_shares: %s", v.TotalShares)
	}

	stake, ok := new(big.Int).SetString(v.TotalStake, 10)
	if !ok || stake.Sign() < 0 {
		return fmt.Errorf("invalid total_stake: %s", v.TotalStake)
	}

	return nil
}

// GetTotalSharesBigInt returns total shares as big.Int
func (v *ValidatorShares) GetTotalSharesBigInt() *big.Int {
	shares, _ := new(big.Int).SetString(v.TotalShares, 10)
	if shares == nil {
		return big.NewInt(0)
	}
	return shares
}

// GetTotalStakeBigInt returns total stake as big.Int
func (v *ValidatorShares) GetTotalStakeBigInt() *big.Int {
	stake, _ := new(big.Int).SetString(v.TotalStake, 10)
	if stake == nil {
		return big.NewInt(0)
	}
	return stake
}

// AddShares adds shares and stake to the validator
func (v *ValidatorShares) AddShares(shares, stake string, updateTime time.Time) error {
	addShares, ok := new(big.Int).SetString(shares, 10)
	if !ok || addShares.Sign() < 0 {
		return fmt.Errorf("invalid shares: %s", shares)
	}

	addStake, ok := new(big.Int).SetString(stake, 10)
	if !ok || addStake.Sign() < 0 {
		return fmt.Errorf("invalid stake: %s", stake)
	}

	currentShares := v.GetTotalSharesBigInt()
	currentStake := v.GetTotalStakeBigInt()

	currentShares.Add(currentShares, addShares)
	currentStake.Add(currentStake, addStake)

	v.TotalShares = currentShares.String()
	v.TotalStake = currentStake.String()
	v.UpdatedAt = updateTime
	return nil
}

// SubtractShares subtracts shares from the validator
func (v *ValidatorShares) SubtractShares(shares, stake string, updateTime time.Time) error {
	subShares, ok := new(big.Int).SetString(shares, 10)
	if !ok || subShares.Sign() < 0 {
		return fmt.Errorf("invalid shares: %s", shares)
	}

	subStake, ok := new(big.Int).SetString(stake, 10)
	if !ok || subStake.Sign() < 0 {
		return fmt.Errorf("invalid stake: %s", stake)
	}

	currentShares := v.GetTotalSharesBigInt()
	currentStake := v.GetTotalStakeBigInt()

	if currentShares.Cmp(subShares) < 0 {
		return fmt.Errorf("insufficient shares: have %s, need %s", v.TotalShares, shares)
	}

	currentShares.Sub(currentShares, subShares)
	currentStake.Sub(currentStake, subStake)

	v.TotalShares = currentShares.String()
	v.TotalStake = currentStake.String()
	v.UpdatedAt = updateTime
	return nil
}

// CalculateSharesForAmount calculates the shares for a given token amount
// Uses the formula: shares = amount * totalShares / totalStake
// If no shares exist yet, shares = amount * 10^SharePrecision
func (v *ValidatorShares) CalculateSharesForAmount(amount string) (string, error) {
	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok || amountBig.Sign() <= 0 {
		return "", fmt.Errorf("invalid amount: %s", amount)
	}

	totalShares := v.GetTotalSharesBigInt()
	totalStake := v.GetTotalStakeBigInt()

	// If no shares exist, initialize with precision multiplier
	if totalShares.Sign() == 0 || totalStake.Sign() == 0 {
		// shares = amount * 10^SharePrecision
		precision := new(big.Int).Exp(big.NewInt(10), big.NewInt(SharePrecision), nil)
		shares := new(big.Int).Mul(amountBig, precision)
		return shares.String(), nil
	}

	// shares = amount * totalShares / totalStake
	shares := new(big.Int).Mul(amountBig, totalShares)
	shares.Div(shares, totalStake)

	return shares.String(), nil
}

// CalculateAmountForShares calculates the token amount for given shares
// Uses the formula: amount = shares * totalStake / totalShares
func (v *ValidatorShares) CalculateAmountForShares(shares string) (string, error) {
	sharesBig, ok := new(big.Int).SetString(shares, 10)
	if !ok || sharesBig.Sign() <= 0 {
		return "", fmt.Errorf("invalid shares: %s", shares)
	}

	totalShares := v.GetTotalSharesBigInt()
	totalStake := v.GetTotalStakeBigInt()

	if totalShares.Sign() == 0 {
		return "0", nil
	}

	// amount = shares * totalStake / totalShares
	amount := new(big.Int).Mul(sharesBig, totalStake)
	amount.Div(amount, totalShares)

	return amount.String(), nil
}

// DelegatorReward represents rewards for a delegator from a specific validator
type DelegatorReward struct {
	// DelegatorAddress is the delegator's address
	DelegatorAddress string `json:"delegator_address"`

	// ValidatorAddress is the validator's address
	ValidatorAddress string `json:"validator_address"`

	// EpochNumber is the epoch this reward belongs to
	EpochNumber uint64 `json:"epoch_number"`

	// Reward is the reward amount in base units
	Reward string `json:"reward"`

	// SharesAtEpoch is the delegator's shares at the epoch
	SharesAtEpoch string `json:"shares_at_epoch"`

	// ValidatorTotalSharesAtEpoch is the validator's total shares at the epoch
	ValidatorTotalSharesAtEpoch string `json:"validator_total_shares_at_epoch"`

	// CalculatedAt is when the reward was calculated
	CalculatedAt time.Time `json:"calculated_at"`

	// Claimed indicates if the reward has been claimed
	Claimed bool `json:"claimed"`

	// ClaimedAt is when the reward was claimed
	ClaimedAt *time.Time `json:"claimed_at,omitempty"`
}

// NewDelegatorReward creates a new delegator reward
func NewDelegatorReward(delegatorAddr, validatorAddr string, epoch uint64, reward, shares, totalShares string, calcTime time.Time) *DelegatorReward {
	return &DelegatorReward{
		DelegatorAddress:            delegatorAddr,
		ValidatorAddress:            validatorAddr,
		EpochNumber:                 epoch,
		Reward:                      reward,
		SharesAtEpoch:               shares,
		ValidatorTotalSharesAtEpoch: totalShares,
		CalculatedAt:                calcTime,
		Claimed:                     false,
	}
}

// Validate validates the delegator reward
func (r *DelegatorReward) Validate() error {
	if r.DelegatorAddress == "" {
		return fmt.Errorf("delegator_address cannot be empty")
	}
	if r.ValidatorAddress == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}

	reward, ok := new(big.Int).SetString(r.Reward, 10)
	if !ok || reward.Sign() < 0 {
		return fmt.Errorf("invalid reward: %s", r.Reward)
	}

	return nil
}

// GetRewardBigInt returns reward as big.Int
func (r *DelegatorReward) GetRewardBigInt() *big.Int {
	reward, _ := new(big.Int).SetString(r.Reward, 10)
	if reward == nil {
		return big.NewInt(0)
	}
	return reward
}
