package provider_daemon

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

type mockCallbackSink struct {
	mu        sync.Mutex
	callbacks []*marketplace.WaldurCallback
}

func (m *mockCallbackSink) Submit(_ context.Context, callback *marketplace.WaldurCallback) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callbacks = append(m.callbacks, callback)
	return nil
}

func (m *mockCallbackSink) last() *marketplace.WaldurCallback {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.callbacks) == 0 {
		return nil
	}
	return m.callbacks[len(m.callbacks)-1]
}

func TestOrderStatusCallbackHandler_MapsState(t *testing.T) {
	kmCfg := DefaultKeyManagerConfig()
	kmCfg.StorageType = KeyStorageTypeMemory
	keyManager, err := NewKeyManager(kmCfg)
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	if err := keyManager.Unlock(""); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if _, err := keyManager.GenerateKey("cosmos1provider"); err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	statePath := filepath.Join(t.TempDir(), "order_state.json")
	_ = os.WriteFile(statePath, []byte("{}"), 0o600)
	store, err := NewWaldurOrderStore(statePath)
	if err != nil {
		t.Fatalf("NewWaldurOrderStore: %v", err)
	}

	sink := &mockCallbackSink{}
	handler, err := NewOrderStatusCallbackHandler(keyManager, sink, store)
	if err != nil {
		t.Fatalf("NewOrderStatusCallbackHandler: %v", err)
	}

	cb := marketplace.NewWaldurCallbackAt(
		marketplace.ActionTypeStatusUpdate,
		"waldur-order-1",
		marketplace.SyncTypeOrder,
		"customer1/1",
		time.Now().UTC(),
	)
	cb.Payload["state"] = "done"

	if err := handler.ProcessWaldurCallback(context.Background(), cb); err != nil {
		t.Fatalf("ProcessWaldurCallback: %v", err)
	}

	last := sink.last()
	if last == nil {
		t.Fatal("expected callback to be submitted")
	}
	if got := last.Payload["state"]; got != "active" {
		t.Fatalf("payload state = %s, want active", got)
	}
	if last.ChainEntityType != marketplace.SyncTypeOrder {
		t.Fatalf("chain entity type = %s, want order", last.ChainEntityType)
	}
}
