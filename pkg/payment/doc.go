// Package payment provides payment gateway integration for fiat-to-crypto onramp
// in the VirtEngine marketplace, supporting card, PayPal, and ACH payments.
//
// VE-906: Payment gateway integration for Visa/Mastercard
//
// This package implements:
//   - Payment gateway client interface (Stripe, Adyen, PayPal, ACH backends)
//   - Card tokenization (PCI-DSS compliant - never stores actual card numbers)
//   - Payment intent creation and processing
//   - 3D Secure / Strong Customer Authentication (SCA) handling
//   - Webhook handlers for asynchronous payment events
//   - Refund processing and dispute handling
//   - Multi-currency support with automatic conversion
//   - Real-time price feed aggregation (CoinGecko/Chainlink/Pyth)
//   - Cached conversion rates with retry/fallback strategies
//   - Price feed health monitoring and source attribution
//
// Architecture:
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│                       Payment Service                               │
//	├─────────────────────────────────────────────────────────────────────┤
//	│  TokenManager        │  PaymentProcessor   │  WebhookHandler        │
//	│  - Card tokenization │  - Intent creation  │  - Event validation    │
//	│  - Token lifecycle   │  - 3DS/SCA flow     │  - Signature verify    │
//	│  - PCI compliance    │  - Multi-currency   │  - Idempotency         │
//	├─────────────────────────────────────────────────────────────────────┤
//	│                       Gateway Adapters                              │
//	├─────────────────────┬─────────────────────┬─────────────────────────┤
//	│       Stripe        │       Adyen         │       PayPal/ACH         │
//	└─────────────────────┴─────────────────────┴─────────────────────────┘
//
// PCI-DSS Compliance:
//   - Card numbers are NEVER stored - only gateway-provided tokens
//   - All card data flows directly to payment gateway (client-side tokenization)
//   - This package only handles tokens, never raw PANs
//   - Sensitive data is redacted from all logs
//   - All connections use TLS 1.2+ minimum
//
// 3D Secure / SCA Flow:
//  1. Create PaymentIntent with customer data
//  2. Gateway determines if 3DS is required
//  3. If required, return redirect URL for customer authentication
//  4. Customer completes 3DS challenge on issuer's page
//  5. Gateway webhook notifies of authentication result
//  6. Payment completes or fails based on 3DS result
//
// Security Considerations:
//   - All webhook payloads are cryptographically verified
//   - Idempotency keys prevent duplicate charges
//   - Rate limiting protects against abuse
//   - Webhook secrets are stored securely and rotated regularly
//   - All monetary amounts use decimal precision to avoid floating-point errors
//
// Price Feed Fallback Strategy:
//   - The price feed layer uses prioritized sources, cache, and stale cache
//   - See pkg/pricefeed/FALLBACK.md for detailed fallback behavior
//   - If all sources fail, conversion requests are rejected for safety
//
// Usage:
//
//	cfg := payment.DefaultConfig()
//	cfg.Gateway = payment.GatewayStripe
//	cfg.StripeConfig.SecretKey = os.Getenv("STRIPE_SECRET_KEY")
//
//	service, err := payment.NewService(cfg)
//	if err != nil {
//	    return err
//	}
//
//	// Create payment intent
//	intent, err := service.CreatePaymentIntent(ctx, payment.PaymentIntentRequest{
//	    Amount:      payment.NewAmount(10000, "USD"), // $100.00
//	    CustomerID:  "cus_xxx",
//	    Description: "VirtEngine credits",
//	    Metadata:    map[string]string{"order_id": "order_123"},
//	})
//
//	// Handle 3DS if required
//	if intent.RequiresSCA {
//	    // Redirect customer to intent.SCARedirectURL
//	}
//
//	// Process refund
//	refund, err := service.CreateRefund(ctx, payment.RefundRequest{
//	    PaymentIntentID: intent.ID,
//	    Amount:          payment.NewAmount(5000, "USD"), // Partial refund
//	    Reason:          payment.RefundReasonRequestedByCustomer,
//	})
package payment
