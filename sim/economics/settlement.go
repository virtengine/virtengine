package economics

import "math/big"

// SettlementModel handles settlement outcomes.
type SettlementModel struct {
	FailureRate float64
}

// Settle returns whether settlement succeeds and the settled amount.
func (s SettlementModel) Settle(amount *big.Int, escrow *EscrowModel) (bool, *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return true, big.NewInt(0)
	}
	if escrow != nil && escrow.IsUnderfunded(amount) {
		return false, big.NewInt(0)
	}
	return true, new(big.Int).Set(amount)
}
