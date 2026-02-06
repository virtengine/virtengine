package economics

import "math/big"

// EscrowModel tracks escrow balances and underfunding risk.
type EscrowModel struct {
	Locked *big.Int
}

// NewEscrowModel creates an escrow model.
func NewEscrowModel() *EscrowModel {
	return &EscrowModel{Locked: big.NewInt(0)}
}

// Lock places funds into escrow.
func (e *EscrowModel) Lock(amount *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	e.Locked.Add(e.Locked, amount)
}

// Release removes funds from escrow.
func (e *EscrowModel) Release(amount *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	e.Locked.Sub(e.Locked, amount)
	if e.Locked.Sign() < 0 {
		e.Locked.SetInt64(0)
	}
}

// IsUnderfunded checks if escrow is insufficient for a settlement.
func (e *EscrowModel) IsUnderfunded(amount *big.Int) bool {
	if amount == nil {
		return false
	}
	return e.Locked.Cmp(amount) < 0
}
