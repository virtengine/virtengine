// Package payment provides payment gateway integration for Visa/Mastercard.
package payment

// NewPayPalGateway creates a PayPal gateway adapter.
func NewPayPalGateway(config PayPalConfig) (Gateway, error) {
	return NewPayPalAdapter(config)
}

// NewACHGateway creates an ACH gateway adapter.
func NewACHGateway(config ACHConfig) (Gateway, error) {
	return NewACHAdapter(config)
}
