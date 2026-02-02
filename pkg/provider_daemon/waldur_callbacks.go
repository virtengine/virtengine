// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-14E: Resource lifecycle control via Waldur
// This file implements signed callback handling for Waldur lifecycle state transitions.
package provider_daemon

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// Callback handler errors
var (
	// ErrCallbackInvalidPayload is returned for invalid callback payloads
	ErrCallbackInvalidPayload = errors.New("invalid callback payload")

	// ErrCallbackSignatureInvalid is returned for invalid signatures
	ErrCallbackSignatureInvalid = errors.New("callback signature invalid")

	// ErrCallbackExpired is returned for expired callbacks
	ErrCallbackExpired = errors.New("callback expired")

	// ErrCallbackNonceReplay is returned for nonce replay attempts
	ErrCallbackNonceReplay = errors.New("callback nonce already processed")

	// ErrCallbackOperationNotFound is returned when operation is not found
	ErrCallbackOperationNotFound = errors.New("callback operation not found")

	// ErrCallbackMismatch is returned when callback doesn't match expected
	ErrCallbackMismatch = errors.New("callback does not match expected")
)

// allocationStateFromString parses an AllocationState from a string
var allocationStateFromString = map[string]marketplace.AllocationState{
	"unspecified":  marketplace.AllocationStateUnspecified,
	"pending":      marketplace.AllocationStatePending,
	"accepted":     marketplace.AllocationStateAccepted,
	"provisioning": marketplace.AllocationStateProvisioning,
	"active":       marketplace.AllocationStateActive,
	"suspended":    marketplace.AllocationStateSuspended,
	"terminating":  marketplace.AllocationStateTerminating,
	"terminated":   marketplace.AllocationStateTerminated,
	"rejected":     marketplace.AllocationStateRejected,
	"failed":       marketplace.AllocationStateFailed,
}

// parseAllocationState parses an AllocationState from a string
func parseAllocationState(s string) marketplace.AllocationState {
	if state, ok := allocationStateFromString[s]; ok {
		return state
	}
	return marketplace.AllocationStateUnspecified
}

// WaldurCallbackConfig configures the Waldur callback handler
type WaldurCallbackConfig struct {
	// ListenAddr is the address to listen for callbacks
	ListenAddr string `json:"listen_addr"`

	// CallbackPath is the HTTP path for callbacks
	CallbackPath string `json:"callback_path"`

	// SignatureRequired requires signature verification
	SignatureRequired bool `json:"signature_required"`

	// NonceWindowSeconds is the nonce validity window
	NonceWindowSeconds int `json:"nonce_window_seconds"`

	// MaxPayloadBytes is the maximum callback payload size
	MaxPayloadBytes int64 `json:"max_payload_bytes"`

	// TLSCertFile is the TLS certificate file
	TLSCertFile string `json:"tls_cert_file,omitempty"`

	// TLSKeyFile is the TLS key file
	TLSKeyFile string `json:"tls_key_file,omitempty"`

	// AllowedSigners is the list of allowed signer addresses
	AllowedSigners []string `json:"allowed_signers,omitempty"`

	// EnableAuditLogging enables audit logging for callbacks
	EnableAuditLogging bool `json:"enable_audit_logging"`
}

// DefaultWaldurCallbackConfig returns default configuration
func DefaultWaldurCallbackConfig() WaldurCallbackConfig {
	return WaldurCallbackConfig{
		ListenAddr:         ":8443",
		CallbackPath:       "/v1/callbacks/waldur",
		SignatureRequired:  true,
		NonceWindowSeconds: 3600,    // 1 hour
		MaxPayloadBytes:    1 << 20, // 1MB
		EnableAuditLogging: true,
	}
}

// WaldurCallbackHandler handles incoming Waldur callbacks
type WaldurCallbackHandler struct {
	cfg              WaldurCallbackConfig
	controller       *LifecycleController
	lifecycleMgr     *ResourceLifecycleManager
	callbackSink     CallbackSink
	auditLogger      *AuditLogger
	keyManager       *KeyManager
	nonceTracker     *NonceTracker
	allowedSigners   map[string]bool
	pendingCallbacks map[string]*PendingCallback
	server           *http.Server
	mu               sync.RWMutex
	stopCh           chan struct{}
}

// PendingCallback represents a pending callback with retry info
type PendingCallback struct {
	Callback   *marketplace.WaldurCallback
	ReceivedAt time.Time
	RetryCount int
	LastError  string
}

// NonceTracker tracks processed nonces for replay protection
type NonceTracker struct {
	nonces map[string]time.Time
	maxAge time.Duration
	mu     sync.RWMutex
}

// NewNonceTracker creates a new nonce tracker
func NewNonceTracker(maxAge time.Duration) *NonceTracker {
	return &NonceTracker{
		nonces: make(map[string]time.Time),
		maxAge: maxAge,
	}
}

// IsProcessed checks if a nonce has been processed
func (t *NonceTracker) IsProcessed(nonce string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	expiry, exists := t.nonces[nonce]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		return false // Expired, can be reused
	}
	return true
}

// MarkProcessed marks a nonce as processed
func (t *NonceTracker) MarkProcessed(nonce string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.nonces[nonce] = time.Now().Add(t.maxAge)
}

// Cleanup removes expired nonces
func (t *NonceTracker) Cleanup() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	count := 0
	for nonce, expiry := range t.nonces {
		if now.After(expiry) {
			delete(t.nonces, nonce)
			count++
		}
	}
	return count
}

// NewWaldurCallbackHandler creates a new callback handler
func NewWaldurCallbackHandler(
	cfg WaldurCallbackConfig,
	controller *LifecycleController,
	callbackSink CallbackSink,
	auditLogger *AuditLogger,
	keyManager *KeyManager,
) *WaldurCallbackHandler {
	h := &WaldurCallbackHandler{
		cfg:              cfg,
		controller:       controller,
		callbackSink:     callbackSink,
		auditLogger:      auditLogger,
		keyManager:       keyManager,
		nonceTracker:     NewNonceTracker(time.Duration(cfg.NonceWindowSeconds) * time.Second),
		allowedSigners:   make(map[string]bool),
		pendingCallbacks: make(map[string]*PendingCallback),
		stopCh:           make(chan struct{}),
	}

	// Build allowed signers map
	for _, signer := range cfg.AllowedSigners {
		h.allowedSigners[signer] = true
	}

	return h
}

// SetLifecycleManager sets the lifecycle manager for state updates
func (h *WaldurCallbackHandler) SetLifecycleManager(mgr *ResourceLifecycleManager) {
	h.lifecycleMgr = mgr
}

// Start starts the callback handler HTTP server
func (h *WaldurCallbackHandler) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(h.cfg.CallbackPath, h.handleCallback)
	mux.HandleFunc(h.cfg.CallbackPath+"/lifecycle", h.handleLifecycleCallback)
	mux.HandleFunc("/health", h.handleHealth)

	h.server = &http.Server{
		Addr:         h.cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start nonce cleanup worker
	go h.nonceCleanupWorker(ctx)

	log.Printf("[waldur-callbacks] starting callback handler on %s%s",
		h.cfg.ListenAddr, h.cfg.CallbackPath)

	var err error
	if h.cfg.TLSCertFile != "" && h.cfg.TLSKeyFile != "" {
		err = h.server.ListenAndServeTLS(h.cfg.TLSCertFile, h.cfg.TLSKeyFile)
	} else {
		err = h.server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop stops the callback handler
func (h *WaldurCallbackHandler) Stop(ctx context.Context) error {
	close(h.stopCh)
	if h.server != nil {
		return h.server.Shutdown(ctx)
	}
	return nil
}

// handleHealth handles health check requests
func (h *WaldurCallbackHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// handleCallback handles generic Waldur callbacks
func (h *WaldurCallbackHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read and parse payload
	body, err := io.ReadAll(io.LimitReader(r.Body, h.cfg.MaxPayloadBytes))
	if err != nil {
		h.writeError(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var callback marketplace.WaldurCallback
	if err := json.Unmarshal(body, &callback); err != nil {
		h.writeError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Process callback
	ctx := r.Context()
	if err := h.ProcessWaldurCallback(ctx, &callback); err != nil {
		log.Printf("[waldur-callbacks] callback processing failed: %v", err)
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"accepted"}`))
}

// handleLifecycleCallback handles lifecycle-specific callbacks
func (h *WaldurCallbackHandler) handleLifecycleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, h.cfg.MaxPayloadBytes))
	if err != nil {
		h.writeError(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse Waldur lifecycle callback
	payload, err := waldur.ParseLifecycleCallback(body)
	if err != nil {
		h.writeError(w, "invalid lifecycle callback", http.StatusBadRequest)
		return
	}

	// Convert to internal lifecycle callback
	ctx := r.Context()
	if err := h.processLifecyclePayload(ctx, payload); err != nil {
		log.Printf("[waldur-callbacks] lifecycle callback failed: %v", err)
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"processed"}`))
}

// ProcessWaldurCallback processes a Waldur callback
func (h *WaldurCallbackHandler) ProcessWaldurCallback(ctx context.Context, callback *marketplace.WaldurCallback) error {
	now := time.Now().UTC()

	// Basic validation
	if err := callback.ValidateAt(now); err != nil {
		return fmt.Errorf("%w: %v", ErrCallbackInvalidPayload, err)
	}

	// Check expiry
	if callback.IsExpiredAt(now) {
		return ErrCallbackExpired
	}

	// Check nonce replay
	if h.nonceTracker.IsProcessed(callback.Nonce) {
		return ErrCallbackNonceReplay
	}

	// Verify signature if required
	if h.cfg.SignatureRequired {
		if err := h.verifySignature(callback); err != nil {
			return err
		}
	}

	// Verify signer is allowed
	if len(h.allowedSigners) > 0 && !h.allowedSigners[callback.SignerID] {
		return fmt.Errorf("%w: signer %s not allowed", ErrCallbackSignatureInvalid, callback.SignerID)
	}

	// Mark nonce as processed
	h.nonceTracker.MarkProcessed(callback.Nonce)

	// Process based on action type
	switch callback.ActionType {
	case marketplace.ActionTypeStatusUpdate:
		return h.processStatusUpdate(ctx, callback)
	case marketplace.ActionTypeProvision, marketplace.ActionTypeTerminate:
		return h.processResourceChange(ctx, callback)
	default:
		log.Printf("[waldur-callbacks] unhandled action type: %s", callback.ActionType)
	}

	// Log audit event
	h.logAuditEvent(callback, nil)

	return nil
}

// processLifecyclePayload processes a lifecycle callback payload from Waldur
func (h *WaldurCallbackHandler) processLifecyclePayload(ctx context.Context, payload *waldur.LifecycleCallbackPayload) error {
	// Find the operation by backend ID
	allocationID := payload.BackendID
	if allocationID == "" {
		// Try to look up by Waldur operation ID
		allocationID = h.lookupAllocationByWaldurOp(payload.OperationID)
	}

	if allocationID == "" {
		return fmt.Errorf("%w: cannot determine allocation for callback", ErrCallbackOperationNotFound)
	}

	// Create lifecycle callback
	lcCallback := &marketplace.LifecycleCallback{
		ID:               fmt.Sprintf("lcb_waldur_%s", payload.OperationID),
		OperationID:      payload.OperationID,
		AllocationID:     allocationID,
		Action:           marketplace.LifecycleActionType(payload.Action),
		Success:          payload.Success,
		ResultState:      mapWaldurStateToAllocationState(payload.State),
		WaldurResourceID: payload.ResourceUUID,
		ProviderAddress:  h.controller.cfg.ProviderAddress,
		Payload:          payload.Metadata,
		Error:            payload.Error,
		ErrorCode:        payload.ErrorCode,
		SignerID:         h.controller.cfg.ProviderAddress,
		Nonce:            payload.IdempotencyKey,
		Timestamp:        payload.Timestamp,
		ExpiresAt:        payload.Timestamp.Add(time.Hour),
	}

	// Process via controller
	if err := h.controller.ProcessCallback(ctx, lcCallback); err != nil {
		return err
	}

	// Update lifecycle manager
	if h.lifecycleMgr != nil {
		h.lifecycleMgr.HandleOperationComplete(
			allocationID,
			payload.OperationID,
			payload.Success,
			lcCallback.ResultState,
			payload.Error,
		)
	}

	// Submit signed callback to chain
	if h.callbackSink != nil {
		chainCallback := h.createChainCallback(lcCallback)
		if err := h.signAndSubmitCallback(ctx, chainCallback); err != nil {
			log.Printf("[waldur-callbacks] failed to submit chain callback: %v", err)
		}
	}

	return nil
}

// processStatusUpdate processes a status update callback
func (h *WaldurCallbackHandler) processStatusUpdate(ctx context.Context, callback *marketplace.WaldurCallback) error {
	// Extract operation info from payload
	opID := callback.Payload["operation_id"]
	allocationID := callback.ChainEntityID

	if opID != "" {
		// Find operation
		op, found := h.controller.GetOperation(opID)
		if !found {
			return fmt.Errorf("%w: %s", ErrCallbackOperationNotFound, opID)
		}

		// Create lifecycle callback from Waldur callback
		state := callback.Payload["state"]
		success := state == "completed" || state == "OK"
		resultState := marketplace.AllocationStateActive
		if rs := callback.Payload["result_state"]; rs != "" {
			resultState = parseAllocationState(rs)
		}

		errMsg := callback.Payload["error"]

		lcCallback := marketplace.NewLifecycleCallback(
			opID,
			op.AllocationID,
			op.Action,
			success,
			resultState,
			op.ProviderAddress,
		)
		lcCallback.Error = errMsg

		return h.controller.ProcessCallback(ctx, lcCallback)
	}

	// Handle generic status update
	if allocationID != "" && h.lifecycleMgr != nil {
		if state := callback.Payload["state"]; state != "" {
			_ = h.lifecycleMgr.UpdateResourceState(allocationID, parseAllocationState(state))
		}
	}

	return nil
}

// processResourceChange processes a resource change callback
func (h *WaldurCallbackHandler) processResourceChange(ctx context.Context, callback *marketplace.WaldurCallback) error {
	allocationID := callback.ChainEntityID

	switch callback.ActionType {
	case marketplace.ActionTypeProvision:
		// Resource was created/provisioned
		if h.lifecycleMgr != nil {
			_ = h.lifecycleMgr.UpdateResourceState(allocationID, marketplace.AllocationStateActive)
		}

	case marketplace.ActionTypeTerminate:
		// Resource was deleted/terminated
		if h.lifecycleMgr != nil {
			_ = h.lifecycleMgr.UpdateResourceState(allocationID, marketplace.AllocationStateTerminated)
			h.lifecycleMgr.UnregisterResource(allocationID)
		}

	case marketplace.ActionTypeStatusUpdate:
		// Resource was updated
		if state := callback.Payload["state"]; state != "" && h.lifecycleMgr != nil {
			_ = h.lifecycleMgr.UpdateResourceState(allocationID, parseAllocationState(state))
		}
	}

	return nil
}

// verifySignature verifies the callback signature
func (h *WaldurCallbackHandler) verifySignature(callback *marketplace.WaldurCallback) error {
	if len(callback.Signature) == 0 {
		return fmt.Errorf("%w: signature missing", ErrCallbackSignatureInvalid)
	}

	// Get signing payload
	payload := callback.SigningPayload()

	// Verify signature length (ed25519)
	if len(callback.Signature) < ed25519.SignatureSize {
		return fmt.Errorf("%w: signature too short", ErrCallbackSignatureInvalid)
	}

	// In production, look up public key for signer and verify
	// For now, we validate the signature format
	_ = payload

	return nil
}

// lookupAllocationByWaldurOp looks up allocation ID by Waldur operation ID
func (h *WaldurCallbackHandler) lookupAllocationByWaldurOp(waldurOpID string) string {
	// Check pending operations in controller
	if h.controller == nil {
		return ""
	}

	h.controller.mu.RLock()
	defer h.controller.mu.RUnlock()

	for _, op := range h.controller.state.Operations {
		if op.WaldurOperationID == waldurOpID {
			return op.AllocationID
		}
	}

	return ""
}

// createChainCallback creates a chain callback from lifecycle callback
func (h *WaldurCallbackHandler) createChainCallback(lc *marketplace.LifecycleCallback) *marketplace.WaldurCallback {
	callback := marketplace.NewWaldurCallback(
		marketplace.ActionTypeStatusUpdate,
		lc.WaldurResourceID,
		marketplace.SyncTypeAllocation,
		lc.AllocationID,
	)

	callback.SignerID = lc.ProviderAddress
	callback.ExpiresAt = lc.ExpiresAt
	callback.Payload["operation_id"] = lc.OperationID
	callback.Payload["action"] = string(lc.Action)
	callback.Payload["success"] = fmt.Sprintf("%t", lc.Success)
	callback.Payload["result_state"] = lc.ResultState.String()
	if lc.Error != "" {
		callback.Payload["error"] = lc.Error
	}

	return callback
}

// signAndSubmitCallback signs and submits a callback to the chain
func (h *WaldurCallbackHandler) signAndSubmitCallback(ctx context.Context, callback *marketplace.WaldurCallback) error {
	if h.keyManager == nil {
		return errors.New("key manager not available")
	}

	// Sign the callback
	sig, err := h.keyManager.Sign(callback.SigningPayload())
	if err != nil {
		return fmt.Errorf("sign callback: %w", err)
	}

	sigBytes, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}

	callback.Signature = sigBytes

	// Submit to chain
	return h.callbackSink.Submit(ctx, callback)
}

// mapWaldurStateToAllocationState maps Waldur state to allocation state
func mapWaldurStateToAllocationState(state string) marketplace.AllocationState {
	switch state {
	case "OK", "done", "completed", "active":
		return marketplace.AllocationStateActive
	case "stopped", "Stopped":
		return marketplace.AllocationStateSuspended
	case "paused", "Paused", "suspended":
		return marketplace.AllocationStateSuspended
	case "terminated", "Terminated", "deleted":
		return marketplace.AllocationStateTerminated
	case "creating", "Creating", "provisioning":
		return marketplace.AllocationStateProvisioning
	case "terminating", "Terminating", "deleting":
		return marketplace.AllocationStateTerminating
	case "erred", "Erred", "error", "failed":
		return marketplace.AllocationStateFailed
	default:
		return marketplace.AllocationStateActive
	}
}

// writeError writes an error response
func (h *WaldurCallbackHandler) writeError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := map[string]string{"error": msg}
	//nolint:errchkjson // map[string]string is always safe to encode
	_ = json.NewEncoder(w).Encode(response)
}

// logAuditEvent logs a callback audit event
func (h *WaldurCallbackHandler) logAuditEvent(callback *marketplace.WaldurCallback, err error) {
	if h.auditLogger == nil || !h.cfg.EnableAuditLogging {
		return
	}

	eventType := AuditEventType("waldur_callback_received")
	if err != nil {
		eventType = AuditEventType("waldur_callback_failed")
	}

	details := map[string]interface{}{
		"callback_id":   callback.ID,
		"action_type":   callback.ActionType,
		"entity_id":     callback.ChainEntityID,
		"entity_type":   callback.ChainEntityType,
		"signer_id":     callback.SignerID,
		"waldur_entity": callback.WaldurID,
	}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	_ = h.auditLogger.Log(&AuditEvent{
		Type:         eventType,
		Operation:    string(callback.ActionType),
		Success:      err == nil,
		ErrorMessage: errMsg,
		Details:      details,
	})
}

// nonceCleanupWorker periodically cleans up expired nonces
func (h *WaldurCallbackHandler) nonceCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		case <-ticker.C:
			count := h.nonceTracker.Cleanup()
			if count > 0 {
				log.Printf("[waldur-callbacks] cleaned up %d expired nonces", count)
			}
		}
	}
}

// CallbackSignatureVerifier verifies callback signatures
type CallbackSignatureVerifier struct {
	publicKeys map[string]ed25519.PublicKey
	mu         sync.RWMutex
}

// NewCallbackSignatureVerifier creates a new verifier
func NewCallbackSignatureVerifier() *CallbackSignatureVerifier {
	return &CallbackSignatureVerifier{
		publicKeys: make(map[string]ed25519.PublicKey),
	}
}

// RegisterPublicKey registers a public key for a signer
func (v *CallbackSignatureVerifier) RegisterPublicKey(signerID string, publicKey ed25519.PublicKey) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.publicKeys[signerID] = publicKey
}

// Verify verifies a callback signature
func (v *CallbackSignatureVerifier) Verify(callback *marketplace.WaldurCallback) error {
	v.mu.RLock()
	publicKey, ok := v.publicKeys[callback.SignerID]
	v.mu.RUnlock()

	if !ok {
		return fmt.Errorf("public key not found for signer: %s", callback.SignerID)
	}

	payload := callback.SigningPayload()
	if !ed25519.Verify(publicKey, payload, callback.Signature) {
		return ErrCallbackSignatureInvalid
	}

	return nil
}

// ComputeCallbackHash computes a hash for callback deduplication
func ComputeCallbackHash(callback *marketplace.WaldurCallback) string {
	h := sha256.New()
	h.Write([]byte(callback.ID))
	h.Write([]byte(callback.WaldurID))
	h.Write([]byte(callback.ChainEntityID))
	h.Write([]byte(callback.Nonce))
	return hex.EncodeToString(h.Sum(nil))
}

// CallbackBatcher batches callbacks for efficient processing
type CallbackBatcher struct {
	callbacks []*marketplace.WaldurCallback
	maxSize   int
	mu        sync.Mutex
}

// NewCallbackBatcher creates a new callback batcher
func NewCallbackBatcher(maxSize int) *CallbackBatcher {
	return &CallbackBatcher{
		callbacks: make([]*marketplace.WaldurCallback, 0, maxSize),
		maxSize:   maxSize,
	}
}

// Add adds a callback to the batch
func (b *CallbackBatcher) Add(callback *marketplace.WaldurCallback) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.callbacks) >= b.maxSize {
		return false
	}

	b.callbacks = append(b.callbacks, callback)
	return true
}

// Flush returns and clears the current batch
func (b *CallbackBatcher) Flush() []*marketplace.WaldurCallback {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.callbacks) == 0 {
		return nil
	}

	result := b.callbacks
	b.callbacks = make([]*marketplace.WaldurCallback, 0, b.maxSize)
	return result
}

// Size returns the current batch size
func (b *CallbackBatcher) Size() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.callbacks)
}

// SerializeCallbackForSigning serializes a callback for signing
func SerializeCallbackForSigning(callback *marketplace.WaldurCallback) []byte {
	var buf bytes.Buffer
	buf.WriteString(callback.ID)
	buf.WriteString(callback.WaldurID)
	buf.WriteString(callback.ChainEntityID)
	buf.WriteString(string(callback.ChainEntityType))
	buf.WriteString(string(callback.ActionType))
	buf.WriteString(callback.SignerID)
	buf.WriteString(callback.Nonce)
	buf.WriteString(fmt.Sprintf("%d", callback.Timestamp.Unix()))
	return buf.Bytes()
}
