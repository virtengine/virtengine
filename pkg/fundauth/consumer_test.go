package fundauth

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

var errFixtureReplay = errors.New("fixture authorization replay")

// memoryConsumer is a test fixture only; T5-18 owns keeper persistence.
type memoryConsumer struct {
	mu       sync.Mutex
	consumed map[string]Digest
	calls    atomic.Int64
}

func newMemoryConsumer() *memoryConsumer {
	return &memoryConsumer{consumed: make(map[string]Digest)}
}

func (consumer *memoryConsumer) KeeperRequired() bool { return true }

type nonKeeperConsumer struct{ *memoryConsumer }

func (consumer nonKeeperConsumer) KeeperRequired() bool { return false }

func (consumer *memoryConsumer) WithAuthorization(ctx context.Context, accountID string, nonceDigest, authDigest Digest, protected func(context.Context) error) error {
	consumer.calls.Add(1)
	consumer.mu.Lock()
	defer consumer.mu.Unlock()
	key := accountID + string(nonceDigest[:])
	if _, exists := consumer.consumed[key]; exists {
		return errFixtureReplay
	}
	if err := protected(ctx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	consumer.consumed[key] = authDigest
	return nil
}

func TestVerifyAndConsumeProofReplayAndRollback(t *testing.T) {
	signed, resolver, _ := signedFixture(t)
	opts := verifyOptions(signed.Authorization)

	t.Run("proof before store", func(t *testing.T) {
		consumer := newMemoryConsumer()
		invalid := signed
		invalid.Signature = nil
		if err := VerifyAndConsume(context.Background(), invalid, DefaultRegistry(), resolver, opts, consumer, func(context.Context) error { return nil }); err == nil {
			t.Fatal("invalid proof accepted")
		}
		if consumer.calls.Load() != 0 {
			t.Fatal("consumer called before proof completed")
		}
	})

	t.Run("replay", func(t *testing.T) {
		consumer := newMemoryConsumer()
		var protectedCalls atomic.Int64
		callback := func(context.Context) error { protectedCalls.Add(1); return nil }
		if err := VerifyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, opts, consumer, callback); err != nil {
			t.Fatal(err)
		}
		if err := VerifyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, opts, consumer, callback); !errors.Is(err, errFixtureReplay) {
			t.Fatalf("replay error = %v", err)
		}
		if protectedCalls.Load() != 1 {
			t.Fatalf("protected calls = %d", protectedCalls.Load())
		}
	})

	t.Run("callback rollback", func(t *testing.T) {
		consumer := newMemoryConsumer()
		callbackFailure := errors.New("protected operation failed")
		if err := VerifyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, opts, consumer, func(context.Context) error { return callbackFailure }); !errors.Is(err, callbackFailure) {
			t.Fatalf("callback error = %v", err)
		}
		if err := VerifyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, opts, consumer, func(context.Context) error { return nil }); err != nil {
			t.Fatalf("retry after rollback failed: %v", err)
		}
	})
}

func TestVerifyAndConsumeConcurrentExactlyOnce(t *testing.T) {
	signed, resolver, _ := signedFixture(t)
	opts := verifyOptions(signed.Authorization)
	consumer := newMemoryConsumer()
	var successes atomic.Int64
	var protectedCalls atomic.Int64
	start := make(chan struct{})
	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			err := VerifyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, opts, consumer, func(context.Context) error {
				protectedCalls.Add(1)
				return nil
			})
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, errFixtureReplay) {
				t.Errorf("unexpected consume error: %v", err)
			}
		}()
	}
	close(start)
	wait.Wait()
	if successes.Load() != 1 || protectedCalls.Load() != 1 {
		t.Fatalf("successes=%d protected=%d, want exactly one", successes.Load(), protectedCalls.Load())
	}
}

func TestConsumerContractAndCancellationRollback(t *testing.T) {
	signed, resolver, _ := signedFixture(t)
	binding := verifyOptions(signed.Authorization)
	fixture := newMemoryConsumer()
	if err := VerifyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, binding, nonKeeperConsumer{fixture}, func(context.Context) error { return nil }); err == nil {
		t.Fatal("non-keeper consumer accepted")
	}

	canceled, cancel := context.WithCancel(context.Background())
	if err := VerifyAndConsume(canceled, signed, DefaultRegistry(), resolver, binding, fixture, func(context.Context) error {
		cancel()
		return nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
	if err := VerifyAndConsume(context.Background(), signed, DefaultRegistry(), resolver, binding, fixture, func(context.Context) error { return nil }); err != nil {
		t.Fatalf("canceled callback committed replay state: %v", err)
	}
}
