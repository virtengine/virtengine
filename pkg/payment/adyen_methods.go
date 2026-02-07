// Package payment provides payment gateway integration for Visa/Mastercard.
//
// VE-3059: Adyen payment method configuration helpers.
package payment

import "strings"

// AdyenPaymentMethodConfig defines allowed payment methods by country.
type AdyenPaymentMethodConfig struct {
	Country        string
	Currency       Currency
	AllowedMethods []string
}

// DefaultAdyenMethodConfigs lists default allowed payment methods per country.
var DefaultAdyenMethodConfigs = []AdyenPaymentMethodConfig{
	{
		Country:  "NL",
		Currency: CurrencyEUR,
		AllowedMethods: []string{
			"ideal",
			"scheme",
			"sepadirectdebit",
		},
	},
	{
		Country:  "BE",
		Currency: CurrencyEUR,
		AllowedMethods: []string{
			"bancontact",
			"scheme",
			"sepadirectdebit",
		},
	},
	{
		Country:  "DE",
		Currency: CurrencyEUR,
		AllowedMethods: []string{
			"scheme",
			"sepadirectdebit",
			"giropay",
			"klarna",
		},
	},
	{
		Country:  "CN",
		Currency: Currency("CNY"),
		AllowedMethods: []string{
			"alipay",
			"wechatpayWeb",
			"unionpay",
		},
	},
	{
		Country:  "GB",
		Currency: CurrencyGBP,
		AllowedMethods: []string{
			"scheme",
			"applepay",
			"googlepay",
			"paypal",
		},
	},
	{
		Country:  "US",
		Currency: CurrencyUSD,
		AllowedMethods: []string{
			"scheme",
			"applepay",
			"googlepay",
			"paypal",
		},
	},
}

// GetAdyenAllowedMethods returns allowed payment methods for a country.
// Defaults to card scheme only when no match is found.
func GetAdyenAllowedMethods(country string, currency Currency) []string {
	normalizedCountry := strings.ToUpper(strings.TrimSpace(country))
	normalizedCurrency := strings.ToUpper(strings.TrimSpace(string(currency)))

	for _, cfg := range DefaultAdyenMethodConfigs {
		if strings.ToUpper(cfg.Country) != normalizedCountry {
			continue
		}
		if cfg.Currency != "" && strings.ToUpper(string(cfg.Currency)) != normalizedCurrency {
			continue
		}
		methods := make([]string, len(cfg.AllowedMethods))
		copy(methods, cfg.AllowedMethods)
		return methods
	}

	return []string{"scheme"}
}
