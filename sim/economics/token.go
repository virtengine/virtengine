package economics

import "math/big"

// TokenModel tracks supply state.
type TokenModel struct {
	Supply      *big.Int
	Circulating *big.Int
	Burned      *big.Int
}

// NewTokenModel creates a token model with initial supply.
func NewTokenModel(supply, circulating, burned *big.Int) *TokenModel {
	return &TokenModel{
		Supply:      new(big.Int).Set(supply),
		Circulating: new(big.Int).Set(circulating),
		Burned:      new(big.Int).Set(burned),
	}
}

// Mint increases total supply and circulating supply.
func (m *TokenModel) Mint(amount *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	m.Supply.Add(m.Supply, amount)
	m.Circulating.Add(m.Circulating, amount)
}

// Burn reduces circulating and total supply.
func (m *TokenModel) Burn(amount *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	m.Supply.Sub(m.Supply, amount)
	m.Circulating.Sub(m.Circulating, amount)
	m.Burned.Add(m.Burned, amount)
	if m.Supply.Sign() < 0 {
		m.Supply.SetInt64(0)
	}
	if m.Circulating.Sign() < 0 {
		m.Circulating.SetInt64(0)
	}
}

// Lock removes tokens from circulation (escrow, vesting).
func (m *TokenModel) Lock(amount *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	m.Circulating.Sub(m.Circulating, amount)
	if m.Circulating.Sign() < 0 {
		m.Circulating.SetInt64(0)
	}
}

// Unlock returns tokens to circulation.
func (m *TokenModel) Unlock(amount *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	m.Circulating.Add(m.Circulating, amount)
}
