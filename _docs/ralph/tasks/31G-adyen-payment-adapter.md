# Task 31G: Adyen Payment Adapter

**vibe-kanban ID:** `17b51b93-6e11-44ae-a5b9-fc540fe0ddec`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31G |
| **Title** | feat(payments): Adyen payment adapter |
| **Priority** | P2 |
| **Wave** | 4 |
| **Estimated LOC** | 2500 |
| **Duration** | 2-3 weeks |
| **Dependencies** | Stripe adapter exists as reference |
| **Blocking** | None |

---

## Problem Statement

Currently only Stripe is supported for fiat payment processing. For international adoption, Adyen support is needed because:

- Better European payment method support (iDEAL, Bancontact, SEPA)
- Asian payment methods (Alipay, WeChat Pay)
- Local acquiring for better acceptance rates
- Enterprise customers prefer Adyen for compliance reasons
- Multi-currency settlement support

### Current State Analysis

```
pkg/payments/stripe/            ✅ Stripe adapter exists
pkg/payments/adyen/             ❌ Does not exist
pkg/payments/interface.go       ✅ PaymentProvider interface exists
```

---

## Acceptance Criteria

### AC-1: Core Payment Processing
- [ ] One-time payments
- [ ] Subscription/recurring payments
- [ ] Payment authorization and capture
- [ ] Refund processing
- [ ] Partial refunds

### AC-2: Payment Methods
- [ ] Card payments (Visa, Mastercard, Amex)
- [ ] iDEAL (Netherlands)
- [ ] Bancontact (Belgium)
- [ ] SEPA Direct Debit
- [ ] Alipay
- [ ] WeChat Pay

### AC-3: Webhook Integration
- [ ] Payment status updates
- [ ] Chargeback notifications
- [ ] Refund confirmations
- [ ] Fraud alerts

### AC-4: Admin Operations
- [ ] Payment search and lookup
- [ ] Manual capture/cancel
- [ ] Dispute management
- [ ] Settlement reports

---

## Technical Requirements

### Adyen Client

```go
// pkg/payments/adyen/client.go

package adyen

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/virtengine/virtengine/pkg/payments"
)

type Config struct {
    APIKey           string
    MerchantAccount  string
    Environment      string  // "test" or "live"
    LiveURLPrefix    string  // Your live URL prefix
    WebhookHMACKey   string
}

type Client struct {
    config     Config
    httpClient *http.Client
    baseURL    string
}

func NewClient(cfg Config) (*Client, error) {
    baseURL := "https://checkout-test.adyen.com/v71"
    if cfg.Environment == "live" {
        baseURL = fmt.Sprintf("https://%s-checkout-live.adyenpayments.com/checkout/v71", cfg.LiveURLPrefix)
    }
    
    return &Client{
        config:     cfg,
        httpClient: &http.Client{Timeout: 30 * time.Second},
        baseURL:    baseURL,
    }, nil
}

// Implement payments.Provider interface
var _ payments.Provider = (*Client)(nil)

func (c *Client) Name() string {
    return "adyen"
}

func (c *Client) CreatePaymentIntent(ctx context.Context, req payments.PaymentIntentRequest) (*payments.PaymentIntent, error) {
    adyenReq := &PaymentsRequest{
        Amount: Amount{
            Currency: req.Currency,
            Value:    req.Amount, // Minor units (cents)
        },
        Reference:        req.Reference,
        MerchantAccount:  c.config.MerchantAccount,
        ReturnURL:        req.ReturnURL,
        Channel:          "Web",
        ShopperReference: req.CustomerID,
        RecurringProcessingModel: req.RecurringModel,
    }
    
    if req.PaymentMethodID != "" {
        adyenReq.PaymentMethod = map[string]string{
            "storedPaymentMethodId": req.PaymentMethodID,
        }
    }
    
    resp, err := c.doRequest(ctx, "POST", "/payments", adyenReq)
    if err != nil {
        return nil, err
    }
    
    var paymentResp PaymentsResponse
    if err := json.Unmarshal(resp, &paymentResp); err != nil {
        return nil, err
    }
    
    return &payments.PaymentIntent{
        ID:            paymentResp.PSPReference,
        ClientSecret:  paymentResp.Action.PaymentData,
        Status:        mapResultCode(paymentResp.ResultCode),
        Amount:        req.Amount,
        Currency:      req.Currency,
        CustomerID:    req.CustomerID,
        PaymentMethod: paymentResp.PaymentMethod.Type,
        ActionData:    paymentResp.Action,
    }, nil
}

func (c *Client) CreateCheckoutSession(ctx context.Context, req payments.CheckoutRequest) (*payments.CheckoutSession, error) {
    sessionReq := &SessionsRequest{
        Amount: Amount{
            Currency: req.Currency,
            Value:    req.Amount,
        },
        Reference:           req.Reference,
        MerchantAccount:     c.config.MerchantAccount,
        ReturnURL:           req.ReturnURL,
        Channel:             "Web",
        ShopperReference:    req.CustomerID,
        CountryCode:         req.CountryCode,
        ShopperLocale:       req.Locale,
        AllowedPaymentMethods: req.AllowedMethods,
        
        // Enable storing payment for subscriptions
        StorePaymentMethod:   req.SetupFutureUsage != "",
        RecurringProcessingModel: req.SetupFutureUsage,
    }
    
    // Add line items if provided
    if len(req.LineItems) > 0 {
        sessionReq.LineItems = make([]LineItem, len(req.LineItems))
        for i, item := range req.LineItems {
            sessionReq.LineItems[i] = LineItem{
                Description:        item.Description,
                AmountIncludingTax: item.Amount,
                Quantity:           item.Quantity,
            }
        }
    }
    
    resp, err := c.doRequest(ctx, "POST", "/sessions", sessionReq)
    if err != nil {
        return nil, err
    }
    
    var sessionResp SessionsResponse
    if err := json.Unmarshal(resp, &sessionResp); err != nil {
        return nil, err
    }
    
    return &payments.CheckoutSession{
        ID:           sessionResp.ID,
        URL:          "", // Adyen uses drop-in, not hosted page
        ClientSecret: sessionResp.SessionData,
        ExpiresAt:    sessionResp.ExpiresAt,
    }, nil
}

func (c *Client) RefundPayment(ctx context.Context, paymentID string, amount int64, reason string) (*payments.Refund, error) {
    refundReq := &RefundsRequest{
        MerchantAccount: c.config.MerchantAccount,
        Amount: Amount{
            Value: amount,
        },
        Reference: fmt.Sprintf("refund-%s-%d", paymentID, time.Now().Unix()),
    }
    
    endpoint := fmt.Sprintf("/payments/%s/refunds", paymentID)
    resp, err := c.doRequest(ctx, "POST", endpoint, refundReq)
    if err != nil {
        return nil, err
    }
    
    var refundResp RefundsResponse
    if err := json.Unmarshal(resp, &refundResp); err != nil {
        return nil, err
    }
    
    return &payments.Refund{
        ID:        refundResp.PSPReference,
        PaymentID: paymentID,
        Amount:    amount,
        Status:    mapRefundStatus(refundResp.Status),
        Reason:    reason,
    }, nil
}
```

### Webhook Handler

```go
// pkg/payments/adyen/webhook.go

package adyen

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
)

type WebhookHandler struct {
    client    *Client
    processor WebhookProcessor
}

type WebhookProcessor interface {
    ProcessPaymentAuthorized(ctx context.Context, event PaymentAuthorizedEvent) error
    ProcessPaymentCaptured(ctx context.Context, event PaymentCapturedEvent) error
    ProcessRefundCompleted(ctx context.Context, event RefundCompletedEvent) error
    ProcessChargeback(ctx context.Context, event ChargebackEvent) error
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Verify HMAC signature
    signature := r.Header.Get("HmacSignature")
    if !h.verifySignature(r, signature) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    var notification NotificationRequest
    if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    ctx := r.Context()
    
    for _, item := range notification.NotificationItems {
        notif := item.NotificationRequestItem
        
        switch notif.EventCode {
        case "AUTHORISATION":
            if notif.Success {
                err := h.processor.ProcessPaymentAuthorized(ctx, PaymentAuthorizedEvent{
                    PSPReference:     notif.PSPReference,
                    MerchantReference: notif.MerchantReference,
                    Amount:           notif.Amount.Value,
                    Currency:         notif.Amount.Currency,
                    PaymentMethod:    notif.PaymentMethod,
                })
                if err != nil {
                    // Log but don't fail - Adyen will retry
                    fmt.Printf("Error processing authorization: %v\n", err)
                }
            }
            
        case "CAPTURE":
            err := h.processor.ProcessPaymentCaptured(ctx, PaymentCapturedEvent{
                PSPReference:         notif.PSPReference,
                OriginalReference:    notif.OriginalReference,
                Amount:               notif.Amount.Value,
            })
            if err != nil {
                fmt.Printf("Error processing capture: %v\n", err)
            }
            
        case "REFUND":
            err := h.processor.ProcessRefundCompleted(ctx, RefundCompletedEvent{
                PSPReference:      notif.PSPReference,
                OriginalReference: notif.OriginalReference,
                Amount:            notif.Amount.Value,
                Success:           notif.Success,
            })
            if err != nil {
                fmt.Printf("Error processing refund: %v\n", err)
            }
            
        case "CHARGEBACK":
            err := h.processor.ProcessChargeback(ctx, ChargebackEvent{
                PSPReference:      notif.PSPReference,
                OriginalReference: notif.OriginalReference,
                Amount:            notif.Amount.Value,
                Reason:            notif.Reason,
            })
            if err != nil {
                fmt.Printf("Error processing chargeback: %v\n", err)
            }
        }
    }
    
    // Always respond with [accepted] for Adyen
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("[accepted]"))
}

func (h *WebhookHandler) verifySignature(r *http.Request, signature string) bool {
    if h.client.config.WebhookHMACKey == "" {
        return true // Skip verification if no key configured
    }
    
    // Read body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        return false
    }
    r.Body = io.NopCloser(bytes.NewBuffer(body))
    
    // Calculate expected signature
    key, _ := base64.StdEncoding.DecodeString(h.client.config.WebhookHMACKey)
    mac := hmac.New(sha256.New, key)
    mac.Write(body)
    expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
    
    return hmac.Equal([]byte(signature), []byte(expectedSig))
}
```

### Payment Methods Configuration

```go
// pkg/payments/adyen/methods.go

package adyen

// PaymentMethodConfig defines available payment methods per country
type PaymentMethodConfig struct {
    Country         string
    Currency        string
    AllowedMethods  []string
}

var DefaultMethodConfigs = []PaymentMethodConfig{
    {
        Country:  "NL",
        Currency: "EUR",
        AllowedMethods: []string{
            "ideal",
            "scheme",  // Cards
            "sepadirectdebit",
        },
    },
    {
        Country:  "BE",
        Currency: "EUR",
        AllowedMethods: []string{
            "bancontact",
            "scheme",
            "sepadirectdebit",
        },
    },
    {
        Country:  "DE",
        Currency: "EUR",
        AllowedMethods: []string{
            "scheme",
            "sepadirectdebit",
            "giropay",
            "klarna",
        },
    },
    {
        Country:  "CN",
        Currency: "CNY",
        AllowedMethods: []string{
            "alipay",
            "wechatpayWeb",
            "unionpay",
        },
    },
    {
        Country:  "GB",
        Currency: "GBP",
        AllowedMethods: []string{
            "scheme",
            "applepay",
            "googlepay",
            "paypal",
        },
    },
    {
        Country:  "US",
        Currency: "USD",
        AllowedMethods: []string{
            "scheme",
            "applepay",
            "googlepay",
            "paypal",
        },
    },
}

func GetAllowedMethods(country, currency string) []string {
    for _, cfg := range DefaultMethodConfigs {
        if cfg.Country == country {
            return cfg.AllowedMethods
        }
    }
    // Default to cards only
    return []string{"scheme"}
}
```

### Frontend Integration

```tsx
// portal/src/components/payments/AdyenCheckout.tsx

'use client';

import { useEffect, useRef, useState } from 'react';
import AdyenCheckout from '@adyen/adyen-web';
import '@adyen/adyen-web/dist/adyen.css';

interface AdyenCheckoutProps {
  sessionId: string;
  sessionData: string;
  onSuccess: (result: any) => void;
  onError: (error: any) => void;
  amount: { value: number; currency: string };
  countryCode: string;
  locale?: string;
}

export function AdyenPaymentForm({
  sessionId,
  sessionData,
  onSuccess,
  onError,
  amount,
  countryCode,
  locale = 'en-US',
}: AdyenCheckoutProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [checkout, setCheckout] = useState<any>(null);

  useEffect(() => {
    const initCheckout = async () => {
      const adyenCheckout = await AdyenCheckout({
        environment: process.env.NEXT_PUBLIC_ADYEN_ENV || 'test',
        clientKey: process.env.NEXT_PUBLIC_ADYEN_CLIENT_KEY!,
        session: {
          id: sessionId,
          sessionData: sessionData,
        },
        onPaymentCompleted: (result: any, component: any) => {
          if (result.resultCode === 'Authorised') {
            onSuccess(result);
          } else {
            onError({ message: `Payment ${result.resultCode}` });
          }
        },
        onError: (error: any, component: any) => {
          onError(error);
        },
        locale,
        countryCode,
        amount,
        analytics: {
          enabled: true,
        },
      });

      setCheckout(adyenCheckout);

      // Mount drop-in component
      if (containerRef.current) {
        adyenCheckout.create('dropin').mount(containerRef.current);
      }
    };

    initCheckout();

    return () => {
      if (checkout) {
        checkout.unmount();
      }
    };
  }, [sessionId, sessionData]);

  return (
    <div className="adyen-checkout-container">
      <div ref={containerRef} />
    </div>
  );
}
```

---

## Directory Structure

```
pkg/payments/
├── interface.go          # Payment provider interface
├── types.go              # Common types
├── registry.go           # Provider registry
├── stripe/
│   ├── client.go         # Existing Stripe client
│   └── webhook.go
└── adyen/
    ├── client.go         # Adyen API client
    ├── webhook.go        # Webhook handler
    ├── methods.go        # Payment methods config
    ├── types.go          # Request/response types
    └── errors.go         # Error handling

portal/src/components/payments/
├── PaymentSelector.tsx   # Provider selection
├── StripeCheckout.tsx    # Existing Stripe component
└── AdyenCheckout.tsx     # New Adyen component
```

---

## Testing Requirements

### Unit Tests
- Request/response serialization
- Signature verification
- Error handling

### Integration Tests
- Test environment payments
- Webhook processing
- Multiple payment methods

### Manual Testing
- iDEAL flow (test cards)
- Alipay (sandbox)
- Card payments 3DS

---

## Security Considerations

1. **API Keys**: Store in secrets manager, never in code
2. **Webhook Verification**: Always verify HMAC signature
3. **PCI Compliance**: Never log full card numbers
4. **Client Key**: Use environment-specific client keys
5. **HTTPS Only**: All callbacks must be HTTPS
