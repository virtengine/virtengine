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
	"sync"
	"time"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// OrderStatusCallbackHandler processes Waldur order status callbacks.
type OrderStatusCallbackHandler struct {
	keyManager   *KeyManager
	callbackSink CallbackSink
	orderStore   *WaldurOrderStore
	state        *WaldurOrderState
	mu           sync.Mutex
}

// NewOrderStatusCallbackHandler creates a new handler.
func NewOrderStatusCallbackHandler(
	keyManager *KeyManager,
	callbackSink CallbackSink,
	orderStore *WaldurOrderStore,
) (*OrderStatusCallbackHandler, error) {
	if keyManager == nil {
		return nil, fmt.Errorf("key manager is required")
	}
	if orderStore == nil {
		return nil, fmt.Errorf("order store is required")
	}
	state, err := orderStore.Load()
	if err != nil {
		return nil, err
	}
	return &OrderStatusCallbackHandler{
		keyManager:   keyManager,
		callbackSink: callbackSink,
		orderStore:   orderStore,
		state:        state,
	}, nil
}

// ProcessWaldurCallback handles a status update callback for orders.
func (h *OrderStatusCallbackHandler) ProcessWaldurCallback(ctx context.Context, callback *marketplace.WaldurCallback) error {
	if callback == nil {
		return fmt.Errorf("callback is nil")
	}
	if callback.ChainEntityType != marketplace.SyncTypeOrder {
		return nil
	}
	if callback.Payload == nil {
		callback.Payload = map[string]string{}
	}

	orderID := callback.ChainEntityID
	if orderID == "" {
		orderID = callback.Payload["ve_order_id"]
	}
	if orderID == "" {
		orderID = callback.Payload["order_id"]
	}
	if orderID == "" {
		orderID = callback.Payload["backend_id"]
	}
	if orderID == "" {
		return fmt.Errorf("order id not provided in callback")
	}

	state := callback.Payload["state"]
	if state == "" {
		state = callback.Payload["status"]
	}
	chainState := mapWaldurOrderStateToChain(state)
	if chainState == "" {
		chainState = "open"
	}

	h.updateState(orderID, callback.WaldurID, state, chainState, callback.Payload)

	chainCallback := marketplace.NewWaldurCallbackAt(
		marketplace.ActionTypeStatusUpdate,
		callback.WaldurID,
		marketplace.SyncTypeOrder,
		orderID,
		time.Now().UTC(),
	)
	signerAddr, err := h.signerAddress()
	if err != nil {
		return err
	}
	chainCallback.SignerID = signerAddr
	chainCallback.ExpiresAt = chainCallback.Timestamp.Add(time.Hour)
	chainCallback.Payload["state"] = chainState
	chainCallback.Payload["waldur_state"] = state
	if errMsg := callback.Payload["error_message"]; errMsg != "" {
		chainCallback.Payload["error_message"] = errMsg
	}
	if res := callback.Payload["resource_uuid"]; res != "" {
		chainCallback.Payload["resource_uuid"] = res
	}

	sig, err := h.keyManager.Sign(chainCallback.SigningPayload())
	if err != nil {
		return fmt.Errorf("sign chain callback: %w", err)
	}
	sigBytes, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	chainCallback.Signature = sigBytes

	if h.callbackSink == nil {
		log.Printf("[order-status] callback sink not configured for order %s", orderID)
		return nil
	}
	return h.callbackSink.Submit(ctx, chainCallback)
}

func (h *OrderStatusCallbackHandler) updateState(orderID, waldurID, waldurState, chainState string, payload map[string]string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	record := h.state.Get(orderID)
	if record == nil {
		record = &WaldurOrderRecord{OrderID: orderID}
	}
	record.WaldurOrderUUID = waldurID
	record.WaldurState = waldurState
	record.ChainState = chainState
	if payload != nil {
		if record.Attributes == nil {
			record.Attributes = map[string]string{}
		}
		for k, v := range payload {
			record.Attributes[k] = v
		}
	}
	record.UpdatedAt = time.Now().UTC()
	h.state.Upsert(record)
	_ = h.orderStore.Save(h.state)
}

func mapWaldurOrderStateToChain(state string) string {
	switch strings.ToLower(state) {
	case "pending-consumer", "pending_payment":
		return "pending_payment"
	case "pending-provider", "pending_provider":
		return "open"
	case "executing", "provisioning":
		return "provisioning"
	case "done", "active":
		return "active"
	case "terminating", "pending_termination":
		return "pending_termination"
	case "terminated":
		return "terminated"
	case "erred", "failed", "error":
		return "failed"
	case "canceled", "cancelled":
		return "cancelled"
	default:
		return ""
	}
}

func (h *OrderStatusCallbackHandler) signerAddress() (string, error) {
	key, err := h.keyManager.GetActiveKey()
	if err != nil {
		return "", fmt.Errorf("get active key: %w", err)
	}
	if key.ProviderAddress == "" {
		return "", fmt.Errorf("provider address missing on active key")
	}
	return key.ProviderAddress, nil
}

// OrderStatusWebhookConfig configures the order status webhook server.
type OrderStatusWebhookConfig struct {
	ListenAddr      string
	CallbackPath    string
	MaxPayloadBytes int64
}

// DefaultOrderStatusWebhookConfig returns defaults for the webhook server.
func DefaultOrderStatusWebhookConfig() OrderStatusWebhookConfig {
	return OrderStatusWebhookConfig{
		ListenAddr:      ":8444",
		CallbackPath:    "/v1/callbacks/waldur/orders",
		MaxPayloadBytes: 1 << 20,
	}
}

// OrderStatusWebhookServer handles Waldur order status webhooks.
type OrderStatusWebhookServer struct {
	cfg     OrderStatusWebhookConfig
	handler *OrderStatusCallbackHandler
	server  *http.Server
}

// NewOrderStatusWebhookServer creates a new webhook server.
func NewOrderStatusWebhookServer(cfg OrderStatusWebhookConfig, handler *OrderStatusCallbackHandler) (*OrderStatusWebhookServer, error) {
	if handler == nil {
		return nil, fmt.Errorf("order status handler is required")
	}
	if cfg.CallbackPath == "" {
		cfg.CallbackPath = DefaultOrderStatusWebhookConfig().CallbackPath
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = DefaultOrderStatusWebhookConfig().ListenAddr
	}
	if cfg.MaxPayloadBytes <= 0 {
		cfg.MaxPayloadBytes = DefaultOrderStatusWebhookConfig().MaxPayloadBytes
	}
	return &OrderStatusWebhookServer{
		cfg:     cfg,
		handler: handler,
	}, nil
}

// Start runs the webhook server.
func (s *OrderStatusWebhookServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.cfg.CallbackPath, s.handleOrderStatus)
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{
		Addr:         s.cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = s.server.Shutdown(context.Background())
	}()

	log.Printf("[order-status] webhook listening on %s%s", s.cfg.ListenAddr, s.cfg.CallbackPath)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop stops the webhook server.
func (s *OrderStatusWebhookServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *OrderStatusWebhookServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *OrderStatusWebhookServer) handleOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := readRequestBody(r, s.cfg.MaxPayloadBytes)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	callback := &marketplace.WaldurCallback{}
	if err := json.Unmarshal(body, callback); err == nil && callback.ActionType != "" {
		if callback.ChainEntityType == "" {
			callback.ChainEntityType = marketplace.SyncTypeOrder
		}
		if err := s.handler.ProcessWaldurCallback(r.Context(), callback); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
		return
	}

	var payload orderStatusPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	cb := payload.toCallback()
	if err := s.handler.ProcessWaldurCallback(r.Context(), cb); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"accepted"}`))
}

type orderStatusPayload struct {
	OrderID         string            `json:"order_id"`
	BackendID       string            `json:"backend_id"`
	WaldurOrderUUID string            `json:"uuid"`
	State           string            `json:"state"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	ResourceUUID    string            `json:"resource_uuid,omitempty"`
	Attributes      map[string]string `json:"attributes,omitempty"`
}

func (p orderStatusPayload) toCallback() *marketplace.WaldurCallback {
	orderID := p.OrderID
	if orderID == "" {
		orderID = p.BackendID
	}
	callback := marketplace.NewWaldurCallbackAt(
		marketplace.ActionTypeStatusUpdate,
		p.WaldurOrderUUID,
		marketplace.SyncTypeOrder,
		orderID,
		time.Now().UTC(),
	)
	if p.State != "" {
		callback.Payload["state"] = p.State
	}
	if p.ErrorMessage != "" {
		callback.Payload["error_message"] = p.ErrorMessage
	}
	if p.ResourceUUID != "" {
		callback.Payload["resource_uuid"] = p.ResourceUUID
	}
	for k, v := range p.Attributes {
		callback.Payload[k] = v
	}
	return callback
}

func readRequestBody(r *http.Request, max int64) ([]byte, error) {
	if r.Body == nil {
		return nil, fmt.Errorf("request body is empty")
	}
	limit := max
	if limit <= 0 {
		limit = 1 << 20
	}
	return io.ReadAll(io.LimitReader(r.Body, limit))
}
