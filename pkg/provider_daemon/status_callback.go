package provider_daemon

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// OrderStatusCallbackConfig configures Waldur order status callbacks.
type OrderStatusCallbackConfig struct {
	ListenAddr      string
	CallbackPath    string
	ProviderAddress string
	MaxPayloadBytes int64
	CallbackTTL     time.Duration
	AuthHeader      string
	AuthToken       string
	StateFile       string
}

// DefaultOrderStatusCallbackConfig returns default callback config.
func DefaultOrderStatusCallbackConfig() OrderStatusCallbackConfig {
	return OrderStatusCallbackConfig{
		ListenAddr:      ":8450",
		CallbackPath:    "/v1/callbacks/waldur/order",
		MaxPayloadBytes: 1 << 20,
		CallbackTTL:     time.Hour,
		AuthHeader:      "X-Waldur-Token",
	}
}

// OrderStatusCallbackHandler handles Waldur order status webhooks.
type OrderStatusCallbackHandler struct {
	cfg          OrderStatusCallbackConfig
	keyManager   *KeyManager
	callbackSink CallbackSink
	stateStore   *OrderRoutingStateStore
	server       *http.Server
}

// NewOrderStatusCallbackHandler creates a new handler.
func NewOrderStatusCallbackHandler(
	cfg OrderStatusCallbackConfig,
	keyManager *KeyManager,
	callbackSink CallbackSink,
) (*OrderStatusCallbackHandler, error) {
	if cfg.ProviderAddress == "" {
		return nil, fmt.Errorf("provider address is required")
	}
	if keyManager == nil {
		return nil, fmt.Errorf("key manager is required")
	}
	if callbackSink == nil {
		return nil, fmt.Errorf("callback sink is required")
	}
	if cfg.CallbackPath == "" {
		cfg.CallbackPath = "/v1/callbacks/waldur/order"
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8450"
	}
	if cfg.MaxPayloadBytes == 0 {
		cfg.MaxPayloadBytes = 1 << 20
	}
	if cfg.CallbackTTL == 0 {
		cfg.CallbackTTL = time.Hour
	}
	if cfg.StateFile == "" {
		cfg.StateFile = "data/waldur_order_state.json"
	}

	store := NewOrderRoutingStateStore(cfg.StateFile)

	return &OrderStatusCallbackHandler{
		cfg:          cfg,
		keyManager:   keyManager,
		callbackSink: callbackSink,
		stateStore:   store,
	}, nil
}

// Start starts the HTTP callback server.
func (h *OrderStatusCallbackHandler) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(h.cfg.CallbackPath, h.handleCallback)
	mux.HandleFunc("/health", h.handleHealth)

	h.server = &http.Server{
		Addr:         h.cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		if h.server != nil {
			_ = h.server.Shutdown(context.Background())
		}
	}()

	log.Printf("[order-callback] listening on %s%s", h.cfg.ListenAddr, h.cfg.CallbackPath)
	if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop stops the server.
func (h *OrderStatusCallbackHandler) Stop(ctx context.Context) error {
	if h.server == nil {
		return nil
	}
	return h.server.Shutdown(ctx)
}

func (h *OrderStatusCallbackHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (h *OrderStatusCallbackHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.cfg.AuthToken != "" {
		if token := r.Header.Get(h.cfg.AuthHeader); token != h.cfg.AuthToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, h.cfg.MaxPayloadBytes))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	payload, err := parseOrderStatusPayload(body)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	orderID, err := h.resolveOrderID(payload)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	chainState := mapWaldurOrderState(payload.State)
	if chainState == "" {
		http.Error(w, "unknown state", http.StatusBadRequest)
		return
	}

	callback := marketplace.NewWaldurCallbackAt(
		marketplace.ActionTypeStatusUpdate,
		payload.OrderUUID,
		marketplace.SyncTypeOrder,
		orderID,
		time.Now().UTC(),
	)
	callback.SignerID = h.cfg.ProviderAddress
	callback.ExpiresAt = callback.Timestamp.Add(h.cfg.CallbackTTL)
	callback.Payload["state"] = chainState
	callback.Payload["waldur_state"] = payload.State
	if payload.ResourceUUID != "" {
		callback.Payload["resource_uuid"] = payload.ResourceUUID
	}
	if payload.ErrorMessage != "" {
		callback.Payload["error"] = payload.ErrorMessage
	}
	if payload.BackendID != "" {
		callback.Payload["backend_id"] = payload.BackendID
	}

	if err := h.signAndSubmitCallback(r.Context(), callback); err != nil {
		log.Printf("[order-callback] submit failed for %s: %v", orderID, err)
		http.Error(w, "callback submission failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"accepted"}`))
}

func (h *OrderStatusCallbackHandler) signAndSubmitCallback(ctx context.Context, callback *marketplace.WaldurCallback) error {
	sig, err := h.keyManager.Sign(callback.SigningPayload())
	if err != nil {
		return fmt.Errorf("sign callback: %w", err)
	}
	sigBytes, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	callback.Signature = sigBytes
	return h.callbackSink.Submit(ctx, callback)
}

func (h *OrderStatusCallbackHandler) resolveOrderID(payload OrderStatusPayload) (string, error) {
	if payload.OrderID != "" {
		return payload.OrderID, nil
	}
	if payload.BackendID != "" {
		return payload.BackendID, nil
	}
	if payload.OrderUUID == "" {
		return "", fmt.Errorf("order uuid missing")
	}

	state, err := h.stateStore.Load()
	if err != nil {
		return "", err
	}
	for _, record := range state.Records {
		if strings.EqualFold(record.WaldurOrderUUID, payload.OrderUUID) {
			return record.OrderID, nil
		}
	}
	return "", fmt.Errorf("order not found for waldur uuid")
}

// OrderStatusPayload captures a Waldur webhook payload.
type OrderStatusPayload struct {
	OrderUUID    string
	OrderID      string
	BackendID    string
	State        string
	ResourceUUID string
	ErrorMessage string
	Attributes   map[string]interface{}
}

func parseOrderStatusPayload(body []byte) (OrderStatusPayload, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return OrderStatusPayload{}, err
	}

	payload := OrderStatusPayload{
		OrderUUID:    extractString(raw, "order_uuid", "order", "uuid"),
		OrderID:      extractString(raw, "order_id"),
		BackendID:    extractString(raw, "backend_id"),
		State:        extractString(raw, "state", "order_state"),
		ResourceUUID: extractString(raw, "resource_uuid", "resource"),
		ErrorMessage: extractString(raw, "error", "error_message"),
	}
	if attrs, ok := raw["attributes"].(map[string]interface{}); ok {
		payload.Attributes = attrs
		if payload.OrderID == "" {
			payload.OrderID = extractString(attrs, "order_id")
		}
		if payload.BackendID == "" {
			payload.BackendID = extractString(attrs, "backend_id")
		}
	}
	if payload.State == "" {
		payload.State = extractString(payload.Attributes, "state")
	}
	return payload, nil
}

func extractString(raw map[string]interface{}, keys ...string) string {
	if raw == nil {
		return ""
	}
	for _, key := range keys {
		if val, ok := raw[key]; ok {
			if s, ok := val.(string); ok {
				return s
			}
		}
	}
	return ""
}

func mapWaldurOrderState(state string) string {
	switch strings.ToLower(state) {
	case "pending-consumer", "pending_consumer":
		return marketplace.OrderStatePendingPayment.String()
	case "pending-provider", "pending_provider":
		return marketplace.OrderStateOpen.String()
	case "executing":
		return marketplace.OrderStateProvisioning.String()
	case "done":
		return marketplace.OrderStateActive.String()
	case "terminating":
		return marketplace.OrderStatePendingTermination.String()
	case "terminated":
		return marketplace.OrderStateTerminated.String()
	case "erred", "error", string(HPCJobStateFailed):
		return marketplace.OrderStateFailed.String()
	case "canceled", "cancelled":
		return marketplace.OrderStateCancelled.String()
	default:
		return ""
	}
}
