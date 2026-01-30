# VirtEngine Go SDK Examples

This directory contains Go examples for interacting with the VirtEngine API.

## Prerequisites

```bash
# Install the VirtEngine SDK
go get github.com/virtengine/virtengine/sdk/go@latest
```

## Examples

### Basic Query Client

```go
package main

import (
    "context"
    "fmt"
    "log"

    market "github.com/virtengine/virtengine/sdk/go/node/market/v2beta1"
    veid "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // Connect to gRPC endpoint
    conn, err := grpc.Dial(
        "api.virtengine.com:9090",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    ctx := context.Background()

    // Query market orders
    marketClient := market.NewQueryClient(conn)
    ordersResp, err := marketClient.Orders(ctx, &market.QueryOrdersRequest{
        Filters: market.OrderFilters{
            State: "open",
        },
    })
    if err != nil {
        log.Fatalf("Failed to query orders: %v", err)
    }
    fmt.Printf("Found %d open orders\n", len(ordersResp.Orders))

    // Query identity
    veidClient := veid.NewQueryClient(conn)
    identityResp, err := veidClient.Identity(ctx, &veid.QueryIdentityRequest{
        AccountAddress: "virtengine1abc...",
    })
    if err != nil {
        log.Printf("Identity not found: %v", err)
    } else {
        fmt.Printf("Identity score: %d\n", identityResp.Identity.Score.Overall)
    }
}
```

### Transaction Signing

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/client/tx"
    "github.com/cosmos/cosmos-sdk/crypto/keyring"
    sdk "github.com/cosmos/cosmos-sdk/types"
    market "github.com/virtengine/virtengine/sdk/go/node/market/v2beta1"
)

func main() {
    // Setup client context (simplified)
    clientCtx := client.Context{}.
        WithChainID("virtengine-1").
        WithFromAddress(sdk.AccAddress{}).
        WithKeyring(keyring.Keyring{})

    // Create bid message
    msg := &market.MsgCreateBid{
        OrderId: market.OrderID{
            Owner: "virtengine1owner...",
            Dseq:  12345,
            Gseq:  1,
            Oseq:  1,
        },
        Provider: "virtengine1provider...",
        Price:    sdk.NewDecCoin("uvirt", sdk.NewInt(950)),
        Deposit:  sdk.NewCoin("uvirt", sdk.NewInt(500000)),
    }

    // Build and sign transaction
    txBuilder := clientCtx.TxConfig.NewTxBuilder()
    if err := txBuilder.SetMsgs(msg); err != nil {
        log.Fatalf("Failed to set msgs: %v", err)
    }

    // Sign and broadcast
    txFactory := tx.Factory{}.
        WithChainID("virtengine-1").
        WithGas(200000).
        WithGasAdjustment(1.2)

    // ... complete signing and broadcast
    fmt.Println("Transaction built successfully")
}
```

### Query with Pagination

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/cosmos/cosmos-sdk/types/query"
    market "github.com/virtengine/virtengine/sdk/go/node/market/v2beta1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    conn, _ := grpc.Dial(
        "api.virtengine.com:9090",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    defer conn.Close()

    client := market.NewQueryClient(conn)
    ctx := context.Background()

    var allOrders []market.Order
    var nextKey []byte

    for {
        resp, err := client.Orders(ctx, &market.QueryOrdersRequest{
            Pagination: &query.PageRequest{
                Key:   nextKey,
                Limit: 100,
            },
        })
        if err != nil {
            log.Fatalf("Query failed: %v", err)
        }

        allOrders = append(allOrders, resp.Orders...)

        if resp.Pagination == nil || len(resp.Pagination.NextKey) == 0 {
            break
        }
        nextKey = resp.Pagination.NextKey
    }

    fmt.Printf("Total orders: %d\n", len(allOrders))
}
```

### VEID Scope Upload

```go
package main

import (
    "context"
    "crypto/rand"
    "fmt"
    "log"

    "golang.org/x/crypto/nacl/box"
    veid "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// EncryptPayload encrypts data for a validator
func EncryptPayload(payload []byte, recipientPubKey *[32]byte) (*veid.EncryptedPayloadEnvelope, error) {
    // Generate ephemeral keypair
    ephemeralPub, ephemeralPriv, err := box.GenerateKey(rand.Reader)
    if err != nil {
        return nil, err
    }

    // Generate nonce
    var nonce [24]byte
    if _, err := rand.Read(nonce[:]); err != nil {
        return nil, err
    }

    // Encrypt
    ciphertext := box.Seal(nil, payload, &nonce, recipientPubKey, ephemeralPriv)

    return &veid.EncryptedPayloadEnvelope{
        RecipientFingerprint: fmt.Sprintf("%x", recipientPubKey[:8]),
        Algorithm:            "X25519-XSalsa20-Poly1305",
        Ciphertext:           ciphertext,
        Nonce:                nonce[:],
        EphemeralPublicKey:   ephemeralPub[:],
    }, nil
}

func main() {
    // Example payload (would be actual identity data)
    payload := []byte(`{"type": "facial_biometric", "data": "..."}`)

    // Validator public key (would be fetched from chain)
    var validatorPubKey [32]byte
    // ... populate from validator's registered key

    envelope, err := EncryptPayload(payload, &validatorPubKey)
    if err != nil {
        log.Fatalf("Encryption failed: %v", err)
    }

    // Create upload message
    msg := &veid.MsgUploadScope{
        Sender:           "virtengine1sender...",
        ScopeId:          "scope_xyz123",
        ScopeType:        veid.ScopeType_FACIAL_BIOMETRIC,
        EncryptedPayload: envelope,
    }

    fmt.Printf("Scope message created: %+v\n", msg)
}
```

### MFA Flow

```go
package main

import (
    "context"
    "fmt"
    "log"

    mfa "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    conn, _ := grpc.Dial(
        "api.virtengine.com:9090",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    defer conn.Close()

    client := mfa.NewQueryClient(conn)
    ctx := context.Background()

    address := "virtengine1abc..."
    txType := "veid.MsgSubmitScope"

    // Check if MFA is required
    required, err := client.MFARequired(ctx, &mfa.QueryMFARequiredRequest{
        Address:         address,
        TransactionType: txType,
    })
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }

    if required.Required {
        fmt.Printf("MFA required: %d factor(s) needed\n", required.FactorsNeeded)
        fmt.Printf("Allowed factors: %v\n", required.AllowedFactors)

        // Get enrolled factors
        enrollments, err := client.FactorEnrollments(ctx, &mfa.QueryFactorEnrollmentsRequest{
            Address: address,
        })
        if err != nil {
            log.Fatalf("Failed to get enrollments: %v", err)
        }

        for _, factor := range enrollments.Enrollments {
            fmt.Printf("Enrolled: %s (%s)\n", factor.Label, factor.FactorType)
        }
    } else {
        fmt.Println("MFA not required for this transaction")
    }
}
```

### WebSocket Subscriptions

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gorilla/websocket"
)

type SubscribeRequest struct {
    JSONRPC string                 `json:"jsonrpc"`
    Method  string                 `json:"method"`
    Params  map[string]interface{} `json:"params"`
    ID      int                    `json:"id"`
}

func main() {
    // Connect to WebSocket
    conn, _, err := websocket.DefaultDialer.Dial("wss://api.virtengine.com/websocket", nil)
    if err != nil {
        log.Fatalf("Connection failed: %v", err)
    }
    defer conn.Close()

    // Subscribe to new blocks
    subscribeReq := SubscribeRequest{
        JSONRPC: "2.0",
        Method:  "subscribe",
        Params: map[string]interface{}{
            "query": "tm.event='NewBlock'",
        },
        ID: 1,
    }

    if err := conn.WriteJSON(subscribeReq); err != nil {
        log.Fatalf("Subscribe failed: %v", err)
    }

    // Read events
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Read error: %v", err)
            break
        }
        fmt.Printf("Event: %s\n", message)
    }
}
```

### Error Handling

```go
package main

import (
    "context"
    "fmt"
    "log"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    market "github.com/virtengine/virtengine/sdk/go/node/market/v2beta1"
)

func queryOrderWithRetry(client market.QueryClient, orderID market.OrderID) (*market.Order, error) {
    ctx := context.Background()

    for attempt := 0; attempt < 3; attempt++ {
        resp, err := client.Order(ctx, &market.QueryOrderRequest{
            Id: orderID,
        })

        if err == nil {
            return &resp.Order, nil
        }

        // Check error type
        st, ok := status.FromError(err)
        if !ok {
            return nil, fmt.Errorf("unknown error: %v", err)
        }

        switch st.Code() {
        case codes.NotFound:
            return nil, fmt.Errorf("order not found")
        case codes.ResourceExhausted:
            // Rate limited - exponential backoff
            wait := time.Second * time.Duration(1<<attempt)
            log.Printf("Rate limited, waiting %v", wait)
            time.Sleep(wait)
            continue
        case codes.Unavailable:
            // Service unavailable - retry
            time.Sleep(time.Second)
            continue
        default:
            return nil, fmt.Errorf("error: %s", st.Message())
        }
    }

    return nil, fmt.Errorf("max retries exceeded")
}
```

## Running Examples

```bash
# Run any example
go run example.go

# With specific endpoint
VIRTENGINE_GRPC=localhost:9090 go run example.go
```

## See Also

- [API Reference](../../reference/)
- [Authentication Guide](../../guides/authentication.md)
- [TypeScript Examples](../typescript/)
