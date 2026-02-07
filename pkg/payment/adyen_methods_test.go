// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-3059: Adyen payment method configuration tests.
package payment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAdyenAllowedMethods(t *testing.T) {
	methods := GetAdyenAllowedMethods("NL", CurrencyEUR)
	assert.Contains(t, methods, "ideal")
	assert.Contains(t, methods, "scheme")

	fallback := GetAdyenAllowedMethods("ZZ", CurrencyUSD)
	assert.Equal(t, []string{"scheme"}, fallback)
}
