package veid_scopes

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockWebhookProcessor struct {
	failuresLeft int32
	calls        int32
}

func (m *mockWebhookProcessor) ProcessWebhook(_ context.Context, _ string, _ []byte) error {
	atomic.AddInt32(&m.calls, 1)
	if atomic.LoadInt32(&m.failuresLeft) > 0 {
		atomic.AddInt32(&m.failuresLeft, -1)
		return errors.New("temporary error")
	}
	return nil
}

func TestWebScopeWebhookHandler_RetrySuccess(t *testing.T) {
	processor := &mockWebhookProcessor{failuresLeft: 2}
	handler := NewWebScopeWebhookHandler(processor, processor, WebhookConfig{
		MaxRetries: 3,
		RetryDelay: 1 * time.Millisecond,
		Async:      false,
	})

	err := handler.HandleEmailWebhook(context.Background(), "mock", []byte("payload"))
	require.NoError(t, err)
	require.Equal(t, int32(3), atomic.LoadInt32(&processor.calls))
}

func TestWebScopeWebhookHandler_RetryExhausted(t *testing.T) {
	processor := &mockWebhookProcessor{failuresLeft: 5}
	handler := NewWebScopeWebhookHandler(processor, processor, WebhookConfig{
		MaxRetries: 2,
		RetryDelay: 1 * time.Millisecond,
		Async:      false,
	})

	err := handler.HandleSMSWebhook(context.Background(), "mock", []byte("payload"))
	require.Error(t, err)
	require.Equal(t, int32(2), atomic.LoadInt32(&processor.calls))
}
